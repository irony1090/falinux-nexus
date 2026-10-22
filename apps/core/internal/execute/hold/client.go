package hold

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// Client는 worker가 보관 프로세스 하나에 붙은 연결이다. 출력·스냅샷·상태는 Events로 순서대로 온다
type Client struct {
	Hello Msg // 붙은 순간의 상태(Status·Pid·Exit)

	conn   net.Conn
	wmu    sync.Mutex
	enc    *json.Encoder
	events chan Msg
}

// Dial은 소켓에 붙어 형식 버전을 맞춘다. 보관 프로세스가 없거나 버전이 다르면 에러
func Dial(sock string) (*Client, error) {
	conn, err := net.DialTimeout("unix", sock, 3*time.Second)
	if err != nil {
		return nil, err
	}
	c := &Client{conn: conn, enc: json.NewEncoder(conn), events: make(chan Msg, 64)}
	dec := json.NewDecoder(conn)
	conn.SetDeadline(time.Now().Add(writeWait))
	if err := c.enc.Encode(Msg{T: MsgHello, V: Version}); err != nil {
		conn.Close()
		return nil, err
	}
	if err := dec.Decode(&c.Hello); err != nil {
		conn.Close()
		return nil, err
	}
	conn.SetDeadline(time.Time{})
	if c.Hello.T != MsgHello || c.Hello.V != Version {
		conn.Close()
		return nil, fmt.Errorf("보관 프로세스 형식 버전 %d, worker는 %d", c.Hello.V, Version)
	}
	go c.read(dec)
	return c, nil
}

func (c *Client) read(dec *json.Decoder) {
	defer close(c.events)
	for {
		var m Msg
		if err := dec.Decode(&m); err != nil {
			return
		}
		c.events <- m
	}
}

// Events는 hello 뒤의 메시지(data·snap·status)다. 연결이 끊기면 닫힌다
func (c *Client) Events() <-chan Msg { return c.events }

func (c *Client) send(m Msg) error {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.enc.Encode(m)
}

func (c *Client) Write(data []byte) error { return c.send(Msg{T: MsgInput, Data: data}) }

// Layout은 보내기만 한다(ioctl 결과는 보관 프로세스 로그)
func (c *Client) Layout(cols, rows uint16) error {
	return c.send(Msg{T: MsgResize, Cols: cols, Rows: rows})
}

func (c *Client) Kill(force bool) error { return c.send(Msg{T: MsgKill, Force: force}) }

// RequestSnap은 스냅샷을 요청한다 — 응답은 Events에 출력과 순서를 지켜 온다
func (c *Client) RequestSnap() error { return c.send(Msg{T: MsgSnap}) }

// Release는 종료 코드를 받았다고 알린다(보관 프로세스가 끝남)
func (c *Client) Release() error { return c.send(Msg{T: MsgRelease}) }

func (c *Client) Close() error { return c.conn.Close() }
