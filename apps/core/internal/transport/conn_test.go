package transport

import (
	"context"
	"testing"
	"time"

	"nexus/internal/protocol"
)

const (
	msgData protocol.MsgType = "DATA"
	msgAdd  protocol.MsgType = "ADD"
)

// newPair는 Pipe로 묶인 두 Conn을 만들고 각자의 Serve를 띄운다.
func newPair(t *testing.T) (*Conn, *Conn) {
	t.Helper()
	rwA, rwB := Pipe()
	a, b := New(rwA), New(rwB)
	go a.Serve()
	go b.Serve()
	t.Cleanup(func() {
		a.Close(nil)
		b.Close(nil)
	})
	return a, b
}

// EVENT가 보낸 순서 그대로, 개수 누락 없이 배달되는지 확인한다.
func TestEmitOrderAndCount(t *testing.T) {
	a, b := newPair(t)

	const n = 1000
	got := make(chan int, n)
	b.On(msgData, func(ev protocol.Frame) {
		var i int
		if err := ev.Bind(&i); err != nil {
			t.Errorf("bind: %v", err)
			return
		}
		got <- i
	})

	for i := range n {
		if err := a.Emit(msgData, i); err != nil {
			t.Fatalf("emit %d: %v", i, err)
		}
	}

	deadline := time.After(2 * time.Second)
	for want := range n {
		select {
		case g := <-got:
			if g != want {
				t.Fatalf("순서 깨짐: want %d, got %d", want, g)
			}
		case <-deadline:
			t.Fatalf("타임아웃: %d/%d 수신", want, n)
		}
	}
}

// REQ/RES와 EVENT가 같은 연결에 섞여 흘러도 서로 간섭하지 않는지 확인한다.
// (EVENT 폭주 중에도 Call이 정상 응답을 받아야 한다)
func TestEventAndCallNoInterference(t *testing.T) {
	a, b := newPair(t)

	b.Handle(msgAdd, func(req protocol.Frame) (any, error) {
		var v int
		if err := req.Bind(&v); err != nil {
			return nil, err
		}
		return v + 1, nil
	})

	const n = 500
	got := make(chan struct{}, n)
	b.On(msgData, func(ev protocol.Frame) { got <- struct{}{} })

	// EVENT를 흘리는 와중에 Call을 끼워 넣는다.
	for i := range n {
		if err := a.Emit(msgData, i); err != nil {
			t.Fatalf("emit: %v", err)
		}
		if i%100 == 0 {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			res, err := a.Call(ctx, msgAdd, i)
			cancel()
			if err != nil {
				t.Fatalf("call: %v", err)
			}
			var out int
			if err := res.Bind(&out); err != nil {
				t.Fatalf("bind res: %v", err)
			}
			if out != i+1 {
				t.Fatalf("call 응답 오류: want %d, got %d", i+1, out)
			}
		}
	}

	deadline := time.After(2 * time.Second)
	for c := range n {
		select {
		case <-got:
		case <-deadline:
			t.Fatalf("타임아웃: EVENT %d/%d 수신", c, n)
		}
	}
}
