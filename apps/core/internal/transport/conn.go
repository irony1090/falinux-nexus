package transport

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"nexus/internal/call"
	"nexus/internal/protocol"

	"github.com/gorilla/websocket"
)

// ErrClosed는 Conn이 닫힌 뒤의 동작에 반환된다.
var ErrClosed = errors.New("transport: conn closed")

// ConnState는 Conn의 수명 상태다(도메인 무관, supervisor·worker 대칭).
type ConnState int32

const (
	StateActive ConnState = iota // 수신/송신 가능
	StateClosed                  // 닫힘: Serve 종료 또는 Close 호출
)

func (s ConnState) String() string {
	switch s {
	case StateActive:
		return "active"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// MessageRW는 Conn이 의존하는 최소 전송면이다.
// gorilla *websocket.Conn이 그대로 만족한다.
type MessageRW interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// Handler는 들어온 REQ를 처리해 응답 데이터(또는 에러)를 돌려준다.
type Handler func(req protocol.Frame) (any, error)

// Conn은 한 WS 연결 위에 요청/응답(REQ/RES)을 얹는다.
// supervisor·worker 양쪽이 동일한 API(Call/Handle)를 쓴다(대칭).
type Conn struct {
	rw   MessageRW
	corr *call.Correlator[protocol.Frame]

	mu       sync.RWMutex
	handlers map[protocol.MsgType]Handler // REQ 핸들러

	state atomic.Int32 // ConnState: 연결 수명 상태(zero=StateActive)

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
	return c
}

// Handle은 REQ 핸들러를 등록한다(요청을 받는 쪽).
func (c *Conn) Handle(t protocol.MsgType, h Handler) {
	c.mu.Lock()
	c.handlers[t] = h
	c.mu.Unlock()
}

// State는 현재 연결 수명 상태를 반환한다.
func (c *Conn) State() ConnState {
	return ConnState(c.state.Load())
}

// Call은 요청을 보내고 짝 응답이 올 때까지 블록한다.
// 응답이 Err를 담고 있으면 그 에러를 반환한다(모든 실패가 같은 통로).
func (c *Conn) Call(ctx context.Context, t protocol.MsgType, data any) (protocol.Frame, error) {
	return c.corr.Call(ctx, outReq{t: t, data: data})
}

// Serve는 수신 루프를 돈다(블로킹). 연결이 끊기면 매달린 Call을 정리하고 반환한다.
func (c *Conn) Serve() error {
	for {
		_, raw, err := c.rw.ReadMessage()
		if err != nil {
			c.state.Store(int32(StateClosed)) // 수신 루프 종료 = 더는 활성 아님
			c.corr.Close(err)                 // 끊김 → 모든 대기 Call 실패
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
		c.state.Store(int32(StateClosed))
		c.corr.Close(err)
		closeErr = c.rw.Close()
	})
	return closeErr
}
