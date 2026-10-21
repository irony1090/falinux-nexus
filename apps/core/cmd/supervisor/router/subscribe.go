package router

import (
	"context"
	"fmt"
	"log"
	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/transport"

	"github.com/labstack/echo/v4"
)

// 브라우저 소켓 구독 — 토픽 키와 process 토픽 구독·해지는 이 파일에만 둔다(REF-refactor-socket.md)

// processTopic: process 하나의 출력·상태 토픽
func processTopic(uid string) string { return "PROCESS:" + uid }

// userTopic: 계정 하나의 알림 토픽(타일·노드·active·공유·보드 목록)
func userTopic(userID int64) string { return fmt.Sprintf("USER:%d", userID) }

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

	tab, err := r.tabs.attach(conn, sid, userID, c.QueryParam("tabId"), req.UserAgent(), sess.Data.Nickname)
	if err != nil {
		log.Printf("supervisor: 탭 id 발급 실패 %v", err)
		conn.Close(err)
		return nil
	}
	if err := conn.Emit(protocol.MsgTabID, protocol.TabIDEvent{TabID: tab.tabID}); err != nil {
		log.Printf("supervisor: TAB:ID 전송 실패 %v", err)
	}

	operable := r.subscribeLiveProcesses(userID, conn)
	r.sizeOwnerOnConnect(tab.tabID, operable)
	r.subscribeHub.Subscribe(userTopic(userID), conn)
	r.onProcessInput(conn, userID)
	r.onTabView(conn, userID)
	r.onTabActive(conn)

	err = conn.Serve()
	conn.Close(err)
	r.subscribeHub.UnsubscribeAll(conn)
	if gone, ok := r.tabs.detach(conn); ok && !gone.active.empty() {
		gone.active = activeTarget{}
		r.publishActive(gone)
	}
	r.sizeOwnerOnDisconnect(tab.tabID)

	return nil
}

// ===== process 토픽 구독 =====
//
// 순서 의존: subscribeLiveProcesses·applyAccess는 DB 판정 읽기와 구독 적용을 accessMu 안에서 한 번에 한다.
// 밖에서 읽으면 "볼 수 없음"을 먼저 읽은 재계산이 그 뒤 공유를 다시 건 재계산보다 늦게 적용돼 구독을 빼 버린다.
// subscribeViewers(새 uid에 추가만)·cleanupProcessTopic(끝난 uid 전부 해제)은 잠그지 않는다.

// subscribeLiveProcesses: 소켓 연결 때 볼 수 있는 살아 있는(PENDING 포함) process 토픽을 구독하고, 조작(20) 이상 uid를 반환 — SB3(크기 우선권 후보)
func (r *supervisorRouter) subscribeLiveProcesses(userID int64, conn *transport.Conn) []string {
	r.accessMu.Lock()
	defer r.accessMu.Unlock()
	procs, err := r.processManager.ListVisibleLive(userID)
	if err != nil {
		log.Printf("[process] 살아 있는 process 조회 실패 user=%d: %v", userID, err)
		return nil
	}
	operable := make([]string, 0, len(procs))
	for _, p := range procs {
		r.subscribeHub.Subscribe(processTopic(p.Process.Uid), conn)
		if p.Permission >= superdb.PermOperate {
			operable = append(operable, p.Process.Uid)
		}
	}
	return operable
}

// applyAccess: 공유 변경 뒤 users × uids를 다시 판정해 구독(보기 이상)·해지(차단)
func (r *supervisorRouter) applyAccess(users []int64, uids []string) {
	r.accessMu.Lock()
	defer r.accessMu.Unlock()
	rows, err := store.GetStorePool().Queries().ListProcessPermissionMatrix(context.Background(), superdb.ListProcessPermissionMatrixParams{UserIds: users, Uids: uids})
	if err != nil {
		log.Printf("[access] 단계 조회 실패 users=%v uids=%v: %v", users, uids, err)
		return
	}
	for _, row := range rows {
		for _, conn := range r.tabs.connsOfUser(row.UserID) {
			if row.Permission > superdb.PermBlock {
				r.subscribeHub.Subscribe(processTopic(row.Uid), conn)
			} else {
				r.subscribeHub.Unsubscribe(processTopic(row.Uid), conn)
			}
		}
	}
}

// subscribeViewers: uid를 볼 수 있는 계정의 연결된 소켓 전부를 구독(exec·worker 재접속, 이미 구독돼 있으면 그대로)
func (r *supervisorRouter) subscribeViewers(uid string) {
	viewers, err := store.GetStorePool().Queries().ListProcessViewers(context.Background(), uid)
	if err != nil {
		log.Printf("[process] 볼 수 있는 계정 조회 실패 uid=%s: %v", uid, err)
		return
	}
	for _, v := range viewers {
		for _, conn := range r.tabs.connsOfUser(v.UserID) {
			r.subscribeHub.Subscribe(processTopic(uid), conn)
		}
	}
}

// cleanupProcessTopic: process가 완전히 끝난 뒤(relay.Wait() 반환 후)에만 구독을 전부 해제한다.
// Kill은 신호만 보내고 종료는 STATUS로 늦게 오므로, 더 일찍 해제하면 마지막 상태 이벤트를 놓친다
func (r *supervisorRouter) cleanupProcessTopic(uid string) {
	r.screens.drop(uid) // S4(끝난 process): 끝나면 버림
	topic := processTopic(uid)
	for _, conn := range r.subscribeHub.Subscribers(topic) {
		r.subscribeHub.Unsubscribe(topic, conn)
	}
}
