package transport

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Pipe는 외부 WS 없이 Conn을 테스트하기 위한 인메모리 양방향 전송면 한 쌍을 만든다.
// 한쪽의 WriteMessage가 다른 쪽의 ReadMessage로 그대로 배달된다.
// 어느 한쪽을 Close하면 공유 closed가 닫혀 양쪽 Read/Write가 ErrClosed로 깨어난다
// (WS 연결 끊김과 동일한 의미).
func Pipe() (MessageRW, MessageRW) {
	a2b := make(chan []byte, 64)
	b2a := make(chan []byte, 64)
	closed := make(chan struct{})
	var once sync.Once
	closeFn := func() { once.Do(func() { close(closed) }) }

	a := &pipeEnd{in: b2a, out: a2b, closed: closed, closeFn: closeFn}
	b := &pipeEnd{in: a2b, out: b2a, closed: closed, closeFn: closeFn}
	return a, b
}

type pipeEnd struct {
	in      chan []byte // 이 끝으로 들어오는 메시지(상대의 out)
	out     chan []byte // 이 끝에서 나가는 메시지(상대의 in)
	closed  chan struct{}
	closeFn func()
}

func (p *pipeEnd) ReadMessage() (int, []byte, error) {
	select {
	case msg := <-p.in:
		return websocket.BinaryMessage, msg, nil
	case <-p.closed:
		return 0, nil, ErrClosed
	}
}

func (p *pipeEnd) WriteMessage(_ int, data []byte) error {
	// 호출자 버퍼 재사용에 대비해 복사한다(aliasing 방지).
	b := make([]byte, len(data))
	copy(b, data)
	select {
	case p.out <- b:
		return nil
	case <-p.closed:
		return ErrClosed
	}
}

func (p *pipeEnd) Close() error {
	p.closeFn()
	return nil
}
