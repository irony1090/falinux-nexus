package transport

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"nexus/internal/call"
	"nexus/internal/protocol"

	"github.com/gorilla/websocket"
)

// ErrClosed는 Conn이 닫힌 뒤의 동작에 반환된다.
var ErrClosed = errors.New("transport: conn closed")

// MessageRW: Conn이 의존하는 최소 전송면(gorilla *websocket.Conn이 그대로 만족)
type MessageRW interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// Handler는 들어온 REQ를 처리해 응답 데이터(또는 에러)를 돌려준다.
type Handler func(req protocol.Frame) (any, error)

// EventHandler는 들어온 EVENT를 소비한다(응답 없음).
type EventHandler func(ev protocol.Frame)

// eventBufferSize: 수신 EVENT 큐. 가득 차면 Serve가 잠깐 블록(backpressure)
const eventBufferSize = 256

// Conn: WS 연결 하나 위의 REQ/RES·EVENT(supervisor·worker 대칭)
type Conn struct {
	rw   MessageRW
	corr *call.Correlator[protocol.Frame]

	mu       sync.RWMutex
	handlers map[protocol.MsgType]Handler      // REQ 핸들러
	on       map[protocol.MsgType]EventHandler // EVENT 핸들러

	events chan protocol.Frame // 수신 EVENT 큐(디스패처가 순차 소비)
	done   chan struct{}       // 디스패처 종료 신호(Close에서 close)

	wmu       sync.Mutex // 쓰기 직렬화
	closeOnce sync.Once
}

// outReq는 Correlator.Call에 넘기는 요청 묶음이다(코어는 들여다보지 않음).
type outReq struct {
	t    protocol.MsgType
	data any
}

// New는 전송면을 감싸 Conn을 만든다. 수신을 돌리려면 Serve를 호출한다.
func New(rw MessageRW) *Conn {
	c := &Conn{
		rw:       rw,
		handlers: make(map[protocol.MsgType]Handler),
		on:       make(map[protocol.MsgType]EventHandler),
		events:   make(chan protocol.Frame, eventBufferSize),
		done:     make(chan struct{}),
	}
	// send 전략 주입: 채번된 id를 REQ 프레임에 찍어 전송.
	c.corr = call.NewCorrelator[protocol.Frame](func(id uint64, payload any) error {
		r := payload.(outReq)
		f, err := protocol.NewRequest(id, r.t, r.data)
		if err != nil {
			return err
		}
		return c.write(f)
	})
	go c.dispatch() // EVENT 수신 배달기(단일 goroutine → 순서 보존). Close까지 생존.
	return c
}

// Handle은 REQ 핸들러를 등록한다(요청을 받는 쪽).
func (c *Conn) Handle(t protocol.MsgType, h Handler) {
	c.mu.Lock()
	c.handlers[t] = h
	c.mu.Unlock()
}

// On은 EVENT 핸들러를 등록한다(단방향 알림을 받는 쪽). Handle과 대칭.
func (c *Conn) On(t protocol.MsgType, h EventHandler) {
	c.mu.Lock()
	c.on[t] = h
	c.mu.Unlock()
}

// Emit: 응답 없는 EVENT(블록 안 함, Correlator 안 거침)
func (c *Conn) Emit(t protocol.MsgType, data any) error {
	f, err := protocol.NewEvent(t, data)
	if err != nil {
		return err
	}
	return c.write(f)
}

// Call: 응답까지 블록. 응답의 Err도 error로 반환(실패 통로 하나)
func (c *Conn) Call(ctx context.Context, t protocol.MsgType, data any) (protocol.Frame, error) {
	return c.corr.Call(ctx, outReq{t: t, data: data})
}

// Serve는 수신 루프를 돈다(블로킹). 연결이 끊기면 매달린 Call을 정리하고 반환한다.
func (c *Conn) Serve() error {
	for {
		_, raw, err := c.rw.ReadMessage()
		if err != nil {
			c.corr.Close(err) // 끊김 → 모든 대기 Call 실패
			return err
		}

		f, derr := protocol.Decode(raw)
		if derr != nil {
			continue // 깨진 프레임은 버린다
		}

		switch f.Kind {
		case protocol.RES:
			var rerr error
			if f.Err != "" {
				rerr = errors.New(f.Err)
			}
			c.corr.Resolve(f.ID, f, rerr) // 기다리던 Call로 배달

		case protocol.REQ:
			go c.dispatchReq(f) // 핸들러가 느려도 수신 루프를 막지 않게

		case protocol.EVENT:
			// 디스패처에 순서대로 넘긴다. done이 닫혔으면(Close) 버리고 진행.
			select {
			case c.events <- f:
			case <-c.done:
			}
		}
	}
}

// dispatch: 수신 EVENT를 goroutine 하나로 순차 배달 — 수신 순서 = 출력 순서(New에서 시작, Close에서 끝)
func (c *Conn) dispatch() {
	for {
		select {
		case <-c.done:
			return
		case f := <-c.events:
			c.mu.RLock()
			h, ok := c.on[f.Type]
			c.mu.RUnlock()
			if ok {
				h(f) // 미등록 타입은 조용히 버린다(깨진 프레임 정책과 동일)
			}
		}
	}
}

func (c *Conn) dispatchReq(req protocol.Frame) {
	c.mu.RLock()
	h, ok := c.handlers[req.Type]
	c.mu.RUnlock()

	if !ok {
		c.write(protocol.NewError(req.ID, req.Type, fmt.Errorf("핸들러 없음: %s", req.Type)))
		return
	}

	data, err := h(req)
	if err != nil {
		c.write(protocol.NewError(req.ID, req.Type, err))
		return
	}

	reply, merr := protocol.NewResponse(req.ID, req.Type, data)
	if merr != nil {
		c.write(protocol.NewError(req.ID, req.Type, merr))
		return
	}
	c.write(reply)
}

func (c *Conn) write(f protocol.Frame) error {
	raw, err := protocol.Encode(f)
	if err != nil {
		return err
	}
	c.wmu.Lock()
	defer c.wmu.Unlock()
	return c.rw.WriteMessage(websocket.BinaryMessage, raw)
}

// Close는 매달린 Call을 err로 정리하고 전송면을 닫는다.
func (c *Conn) Close(err error) error {
	if err == nil {
		err = ErrClosed
	}
	var closeErr error
	c.closeOnce.Do(func() {
		close(c.done) // 디스패처 종료 + Serve의 EVENT push 깨움
		c.corr.Close(err)
		closeErr = c.rw.Close()
	})
	return closeErr
}
