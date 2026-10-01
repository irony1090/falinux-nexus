package transport

import (
	"errors"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

const (
	PingPeriod = 10 * time.Second // ping 간격
	PongWait   = 25 * time.Second // 이 안에 pong(또는 아무 프레임)이 없으면 연결을 끊는다
)

// KeepAlive는 ws에 ping/pong 생존 확인을 건다. transport.New 전에 호출하고, Serve가 읽기를 돌려야 pong이 처리된다.
//
// 재발 방지(REF-realtime.md "발견: 응답 없는 소켓"): 상대가 읽지 않아 데이터 쓰기가 막히면 그 쓰기가
// gorilla 쓰기 잠금을 쥐고 있어 ping도 못 나간다 -> pong 없음 -> read deadline 초과로 Serve가 반환 ->
// 호출자의 Conn.Close가 net.Conn을 닫아 막힌 쓰기를 에러로 풀어 준다. 이 경로가 Hub Publish 막힘을 PongWait로 제한한다.
func KeepAlive(ws *websocket.Conn) {
	ws.SetReadDeadline(time.Now().Add(PongWait))
	ws.SetPongHandler(func(string) error {
		return ws.SetReadDeadline(time.Now().Add(PongWait))
	})
	go func() {
		ticker := time.NewTicker(PingPeriod)
		defer ticker.Stop()
		for range ticker.C {
			err := ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(PingPeriod))
			var ne net.Error
			if err != nil && !(errors.As(err, &ne) && ne.Timeout()) {
				return // 닫힌 연결 — 쓰기 잠금 대기 timeout은 다음 tick에 다시 시도
			}
		}
	}()
}
