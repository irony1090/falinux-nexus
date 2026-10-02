package router

import (
	"fmt"
	"log"
	"nexus/internal/protocol"
	"nexus/internal/transport"

	"github.com/labstack/echo/v4"
)

func nodeSubscribeTopic(parentId int64) string {
	return fmt.Sprintf("NODE:%d", parentId)
}

// handleSubscribeWS: 브라우저 소켓 = 탭 하나. URL `?tabId=`는 sessionStorage에 있던 이전 탭 id(없으면 빈 값).
func (r *supervisorRouter) handleSubscribeWS(c echo.Context) error {
	req, res := c.Request(), c.Response()
	sess := r.requireSession(c)
	sid, userID := sess.Name(), sess.Data.ID
	ws, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Printf("supervisor: upgrade 실패 %v", err)
		return nil
	}

	transport.KeepAlive(ws)
	conn := transport.New(ws)

	tab, err := r.tabs.attach(conn, sid, userID, c.QueryParam("tabId"))
	if err != nil {
		log.Printf("supervisor: 탭 id 발급 실패 %v", err)
		conn.Close(err)
		return nil
	}
	if err := conn.Emit(protocol.MsgTabID, protocol.TabIDEvent{TabID: tab.tabID}); err != nil {
		log.Printf("supervisor: TAB:ID 전송 실패 %v", err)
	}

	live := r.subscribeLiveProcesses(userID, conn)
	r.sizeOwnerOnConnect(tab.tabID, live)
	//node:{parentId} - 0이면 nil인 node들을 구독한다
	r.subscribeHub.Subscribe(nodeSubscribeTopic(0), conn)
	r.subscribeHub.Subscribe(tilesTopic(userID), conn)
	r.onProcessInput(conn, userID)

	err = conn.Serve()
	conn.Close(err)
	// 모든 구독 해제
	r.subscribeHub.UnsubscribeAll(conn)
	r.tabs.detach(conn)
	r.sizeOwnerOnDisconnect(tab.tabID)

	return nil
}

// subscribeLiveProcesses는 계정의 살아 있는(PENDING/PROCESS) process 토픽에 conn을 구독시키고 그 uid들을 반환한다.
// PENDING(worker 끊김)도 포함 — 재접속 Rebind 뒤 출력이 바로 이 구독으로 온다.
func (r *supervisorRouter) subscribeLiveProcesses(userID int64, conn *transport.Conn) []string {
	procs, err := r.processManager.ListLive(userID)
	if err != nil {
		log.Printf("[process] 살아 있는 process 조회 실패 user=%d: %v", userID, err)
		return nil
	}
	uids := make([]string, 0, len(procs))
	for _, p := range procs {
		r.subscribeHub.Subscribe(processTopic(p.Uid), conn)
		uids = append(uids, p.Uid)
	}
	return uids
}

// subscribeUser는 계정의 연결된 소켓 전부를 uid 토픽에 구독시킨다(exec·worker 재접속 — 이미 구독돼 있으면 그대로).
func (r *supervisorRouter) subscribeUser(uid string, userID int64) {
	for _, conn := range r.tabs.connsOfUser(userID) {
		r.subscribeHub.Subscribe(processTopic(uid), conn)
	}
}
