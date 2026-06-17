package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"nexus/internal/syncProcess"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pingInterval = 30 * time.Second
	pongTimeout  = 10 * time.Second
)

type Status uint

const (
	StatusPending Status = iota
	StatusProcess
	StatusCompleted
	StatusFailed
)

func (s Status) String() string {
	switch s {
	case 0:
		return "PENDING"
	case 1:
		return "PROCESS"
	case 2:
		return "COMPLETED"
	case 3:
		return "FAILED"
	default:
		return "UNLKNOWN"

	}
}

func (s Status) IsDone() bool {
	switch s {
	case 2:
		fallthrough
	case 3:
		return true
	default:
		return false
	}
}

type WsClientConn struct {
	conn     *websocket.Conn
	response *http.Response

	closedOnce sync.Once
}

func (c *WsClientConn) close() {
	c.closedOnce.Do(func() {
		c.conn.Close()
	})
}

type WsClient struct {
	header  *http.Header
	receive *syncProcess.SyncData[[]byte]
	status  *syncProcess.SyncData[Status]
	err     error
	url     url.URL
	cleint  *WsClientConn
	mutex   sync.Mutex
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	wMutex  sync.Mutex
}

func NewWsClient(url url.URL, header *http.Header) *WsClient {
	client := &WsClient{
		url:     url,
		header:  header,
		receive: syncProcess.NewSyncData([][]byte{}),
		status:  syncProcess.NewSyncData([]Status{}),
	}
	client.status.Push(StatusPending)

	return client
}

func (c *WsClient) Connect() error {
	// 1. 기존 goroutine에 종료 신호
	c.mutex.Lock()
	c.initUnsafe()
	c.mutex.Unlock()

	// 2. goroutine 완전 종료 대기 (mutex 없이)
	c.wg.Wait()

	// 3. 상태 초기화 후 새 연결
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.err = nil
	c.status.Init()
	c.receive.Init()

	conn, resp, err := websocket.DefaultDialer.Dial(c.url.String(), *c.header)
	if err != nil {
		defer c.status.Push(StatusFailed)
		c.err = err
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.cleint = &WsClientConn{
		conn:     conn,
		response: resp,
	}

	c.wg.Add(2)
	go c.readPump(ctx, c.cleint)
	go c.pingPump(ctx, c.cleint)

	defer c.status.Push(StatusProcess)
	return nil
}

func (c *WsClient) readPump(ctx context.Context, client *WsClientConn) {
	defer c.wg.Done()

	client.conn.SetPongHandler(func(string) error {
		return client.conn.SetReadDeadline(time.Now().Add(pingInterval + pongTimeout))
	})
	client.conn.SetReadDeadline(time.Now().Add(pingInterval + pongTimeout))

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			select {
			case <-ctx.Done():
			default:
				c.cancel()
				client.close()
				c.status.Push(StatusFailed)
			}
			return
		}

		c.receive.Push(msg)
	}
}

func (c *WsClient) pingPump(ctx context.Context, client *WsClientConn) {
	defer c.wg.Done()

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.wMutex.Lock()
			err := client.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(pongTimeout))
			c.wMutex.Unlock()

			if err != nil {
				select {
				case <-ctx.Done():
				default:
					c.mutex.Lock()
					c.err = err
					c.mutex.Unlock()
					c.cancel()
					client.close()
					c.status.Push(StatusFailed)
				}
				return
			}
		}
	}
}

// initUnsafe는 종료 신호만 보냄. 상태 초기화(Init)는 wg.Wait() 이후 호출자가 직접 수행해야 함.
func (c *WsClient) initUnsafe() {
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}

	if c.cleint != nil {
		c.cleint.close()
		c.cleint = nil
	}
}

func (c *WsClient) Close() {
	// 1. 종료 신호
	c.mutex.Lock()
	c.initUnsafe()
	c.status.Push(StatusCompleted)
	c.mutex.Unlock()

	// 2. goroutine 완전 종료 대기 (mutex 없이)
	c.wg.Wait()

	// 3. 채널 정리
	c.receive.Close()
	c.status.Close()
}

func (c *WsClient) Status() (Status, error) {
	return c.status.Shift()
}

func (c *WsClient) Receive() ([]byte, error) {
	return c.receive.Shift()
}

func (c *WsClient) WriteData(websocketMessageType int, data []byte) error {

	c.mutex.Lock()
	client := c.cleint
	c.mutex.Unlock()

	if client == nil {
		return fmt.Errorf("연결되지 않은 상태")
	}

	c.wMutex.Lock()
	defer c.wMutex.Unlock()
	return client.conn.WriteMessage(websocketMessageType, data)
}
