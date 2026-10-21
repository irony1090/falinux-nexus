package router

import (
	"context"
	"log"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/transport"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 원격 active 배지 — 다른 계정 탭의 active 타일을 볼 수 있는 계정에 알린다(메모리만, REF-node-ui-active.md "원격 active 구조안")

// activeTarget: 둘 다 비면 active 없음
type activeTarget struct {
	uid    string
	nodeID int64
}

func (a activeTarget) empty() bool { return a.uid == "" && a.nodeID == 0 }

func (r *supervisorRouter) mountActive(e *echo.Echo) {
	e.GET("/tabs/active", r.listActive)
}

func (r *supervisorRouter) onTabActive(conn *transport.Conn) {
	conn.On(protocol.MsgTabActive, func(ev protocol.Frame) {
		var body protocol.TabActiveEvent
		if err := ev.Bind(&body); err != nil {
			return
		}
		target := activeTarget{uid: body.UID}
		if target.uid == "" && body.NodeID != nil {
			target.nodeID = *body.NodeID
		}
		tab, ok := r.tabs.setActive(conn, target)
		if !ok {
			return
		}
		r.publishActive(tab)
	})
}

// publishActive: 접속 중인 다른 계정 전부에 — 볼 수 없는 계정엔 대상을 비워서(= 지움) 보내 옛 배지가 남지 않게
func (r *supervisorRouter) publishActive(tab browserTab) {
	var others []int64
	for _, u := range r.tabs.onlineUsers() {
		if u != tab.userID {
			others = append(others, u)
		}
	}
	if len(others) == 0 {
		return
	}
	var visible map[int64]bool
	if !tab.active.empty() {
		// 보내는 계정도 판정 — 볼 수 없는 대상으로 남의 화면에 배지를 띄우지 못하게
		visible = r.activeVisibility(append([]int64{tab.userID}, others...), tab.active)
		if !visible[tab.userID] {
			visible = nil
		}
	}
	for _, u := range others {
		ev := protocol.ActiveUpdateEvent{Key: sidKey(tab.tabID), UserID: tab.userID, Nickname: tab.nickname}
		if visible[u] {
			ev.UID, ev.NodeID = activeFields(tab.active)
		}
		if err := r.subscribeHub.Publish(userTopic(u), protocol.MsgActiveUpdate, ev); err != nil {
			log.Printf("[active] ACTIVE:UPDATE 발행 실패 user=%d: %v", u, err)
		}
	}
}

func activeFields(a activeTarget) (string, *int64) {
	if a.uid != "" {
		return a.uid, nil
	}
	id := a.nodeID
	return "", &id
}

// activeVisibility: 계정마다 대상을 볼 수 있는지(보기 이상). 조회 실패 = 아무도 못 봄
func (r *supervisorRouter) activeVisibility(users []int64, target activeTarget) map[int64]bool {
	visible := map[int64]bool{}
	q, ctx := store.GetStorePool().Queries(), context.Background()
	if target.uid != "" {
		rows, err := q.ListProcessPermissionMatrix(ctx, superdb.ListProcessPermissionMatrixParams{UserIds: users, Uids: []string{target.uid}})
		if err != nil {
			log.Printf("[active] process 단계 조회 실패 uid=%s: %v", target.uid, err)
			return visible
		}
		for _, row := range rows {
			visible[row.UserID] = row.Permission > superdb.PermBlock
		}
		return visible
	}
	rows, err := q.ListNodePermissionMatrix(ctx, superdb.ListNodePermissionMatrixParams{UserIds: users, NodeIds: []int64{target.nodeID}})
	if err != nil {
		log.Printf("[active] 노드 단계 조회 실패 node=%d: %v", target.nodeID, err)
		return visible
	}
	for _, row := range rows {
		visible[row.UserID] = row.Permission > superdb.PermBlock
	}
	return visible
}

// listActive: 다른 계정 탭의 active 중 볼 수 있는 것 — 새 탭·재접속·공유 변경 때 처음 받기
func (r *supervisorRouter) listActive(c echo.Context) error {
	me := r.requireSession(c).Data.ID
	var tabs []browserTab
	var uids []string
	var nodeIDs []int64
	for _, tab := range r.tabs.activeTabs() {
		if tab.userID == me {
			continue
		}
		tabs = append(tabs, tab)
		if tab.active.uid != "" {
			uids = append(uids, tab.active.uid)
		} else {
			nodeIDs = append(nodeIDs, tab.active.nodeID)
		}
	}
	out := []protocol.ActiveUpdateEvent{}
	if len(tabs) == 0 {
		return c.JSON(200, out)
	}

	// 보내는 계정도 볼 수 있어야 한다(publishActive와 같은 규칙)
	users := []int64{me}
	for _, tab := range tabs {
		users = append(users, tab.userID)
	}
	q, ctx := TxQueries(c), c.Request().Context()
	procPerm := map[int64]map[string]bool{}
	nodePerm := map[int64]map[int64]bool{}
	if len(uids) > 0 {
		rows, err := q.ListProcessPermissionMatrix(ctx, superdb.ListProcessPermissionMatrixParams{UserIds: users, Uids: uids})
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		for _, row := range rows {
			if procPerm[row.UserID] == nil {
				procPerm[row.UserID] = map[string]bool{}
			}
			procPerm[row.UserID][row.Uid] = row.Permission > superdb.PermBlock
		}
	}
	if len(nodeIDs) > 0 {
		rows, err := q.ListNodePermissionMatrix(ctx, superdb.ListNodePermissionMatrixParams{UserIds: users, NodeIds: nodeIDs})
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		for _, row := range rows {
			if nodePerm[row.UserID] == nil {
				nodePerm[row.UserID] = map[int64]bool{}
			}
			nodePerm[row.UserID][row.NodeID] = row.Permission > superdb.PermBlock
		}
	}
	sees := func(user int64, a activeTarget) bool {
		if a.uid != "" {
			return procPerm[user][a.uid]
		}
		return nodePerm[user][a.nodeID]
	}
	for _, tab := range tabs {
		if !sees(me, tab.active) || !sees(tab.userID, tab.active) {
			continue
		}
		ev := protocol.ActiveUpdateEvent{Key: sidKey(tab.tabID), UserID: tab.userID, Nickname: tab.nickname}
		ev.UID, ev.NodeID = activeFields(tab.active)
		out = append(out, ev)
	}
	return c.JSON(200, out)
}
