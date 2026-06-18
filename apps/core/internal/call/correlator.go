// Package call은 "보낸 요청 ↔ 그 응답"을 채번한 ID로 묶는 재사용 코어다.
//
// subscribe.Manager가 "문자열 키 → 구독 클라이언트 집합" 장부라면,
// 이 패키지는 "요청 ID → 응답을 기다리는 한 명" 장부다.
// 둘 다 장부 + 동시성만 책임지고, 와이어 포맷·전송·데이터 모양은 모른다(전부 주입/제네릭).
package call

import (
	"context"
	"errors"
	"sync"
)

// ErrClosed는 Close된 코릴레이터에 Call을 시도했을 때 반환된다.
var ErrClosed = errors.New("call: correlator closed")

// Correlator는 보낸 요청 하나하나를 채번한 ID로 그 응답과 묶는다.
// R = 응답 데이터 타입(자유). 와이어 포맷도 전송 방식도 모르며,
// send 전략만 주입받아 "ID를 찍어 내보내는 일"은 호출자에게 맡긴다.
type Correlator[R any] struct {
	mu      sync.Mutex
	nextID  uint64
	waiters map[uint64]chan res[R]
	closed  bool

	// send는 채번된 id를 요청에 찍어 전송한다.
	// payload는 코어가 들여다보지 않는다(= 어떤 요청 데이터든 자유).
	send func(id uint64, payload any) error
}

type res[R any] struct {
	val R
	err error
}

// NewCorrelator는 전송 전략을 주입받아 코릴레이터를 만든다.
func NewCorrelator[R any](send func(id uint64, payload any) error) *Correlator[R] {
	return &Correlator[R]{
		waiters: make(map[uint64]chan res[R]),
		send:    send,
	}
}

// Call은 payload를 요청으로 보내고, 짝 응답이 올 때까지 블록한다.
// ctx 만료 또는 Close 시 에러로 깨어난다.
// 요청·대기·응답이 이 한 함수에 모이므로 호출부가 위에서 아래로 읽힌다.
func (c *Correlator[R]) Call(ctx context.Context, payload any) (R, error) {
	var zero R

	// 1) 새 id로 대기표 등록
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return zero, ErrClosed
	}
	c.nextID++
	id := c.nextID
	ch := make(chan res[R], 1) // 버퍼 1: Resolve/Close가 절대 막히지 않게
	c.waiters[id] = ch
	c.mu.Unlock()

	defer c.discard(id) // 타임아웃 등으로 빠져나갈 때 대기표 청소

	// 2) 전송 (락 밖)
	if err := c.send(id, payload); err != nil {
		return zero, err
	}

	// 3) 짝 응답 대기
	select {
	case r := <-ch:
		return r.val, r.err
	case <-ctx.Done():
		return zero, ctx.Err()
	}
}

// Resolve는 수신 루프가 응답을 받았을 때 호출한다 → id로 기다리던 Call에 배달.
// 모르는 id(이미 타임아웃됐거나 없는 id)는 조용히 무시한다.
func (c *Correlator[R]) Resolve(id uint64, val R, err error) {
	c.mu.Lock()
	ch, ok := c.waiters[id]
	if ok {
		delete(c.waiters, id)
	}
	c.mu.Unlock()
	if ok {
		ch <- res[R]{val, err}
	}
}

func (c *Correlator[R]) discard(id uint64) {
	c.mu.Lock()
	delete(c.waiters, id)
	c.mu.Unlock()
}

// Close는 진행 중인 모든 Call을 err로 깨우고 이후 Call을 막는다.
// 연결 끊김 시 호출하면 매달려 있던 요청들이 일괄 정리된다(Done(502) 정신).
func (c *Correlator[R]) Close(err error) {
	c.mu.Lock()
	c.closed = true
	waiters := c.waiters
	c.waiters = make(map[uint64]chan res[R])
	c.mu.Unlock()

	for _, ch := range waiters {
		ch <- res[R]{err: err}
	}
}
