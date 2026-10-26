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

func (r *supervisorRouter) mountActive(e *echo.Group) {
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
	var sight activeSight
	senderSees := false
	if !tab.active.empty() {
		// 보내는 계정도 판정 — 볼 수 없는 대상으로 남의 화면에 배지를 띄우지 못하게
		var err error
		sight, err = loadActiveSight(context.Background(), store.GetStorePool().Queries(), append([]int64{tab.userID}, others...), []activeTarget{tab.active})
		if err != nil {
			log.Printf("[active] 대상 판정 실패 target=%+v: %v", tab.active, err)
		}
		senderSees = sight.sees(tab.userID, tab.active)
	}
	for _, u := range others {
		ev := protocol.ActiveUpdateEvent{Key: sidKey(tab.tabID), UserID: tab.userID, Nickname: tab.nickname}
		if senderSees && sight.sees(u, tab.active) {
			ev.UID, ev.NodeID = activeFields(tab.active)
		}
		r.publishUser(u, protocol.MsgActiveUpdate, ev)
	}
}

func activeFields(a activeTarget) (string, *int64) {
	if a.uid != "" {
		return a.uid, nil
	}
	id := a.nodeID
	return "", &id
}

// activeSight: 계정마다 active 대상을 볼 수 있는지(process·노드 두 표)
type activeSight struct {
	procs visibility[string]
	nodes visibility[int64]
}

func (s activeSight) sees(user int64, a activeTarget) bool {
	if a.uid != "" {
		return s.procs[user][a.uid]
	}
	return s.nodes[user][a.nodeID]
}

// loadActiveSight: users × targets 판정. 조회 실패 = 빈 표(아무도 못 봄) + err
func loadActiveSight(ctx context.Context, q *superdb.Queries, users []int64, targets []activeTarget) (activeSight, error) {
	var uids []string
	var nodeIDs []int64
	for _, a := range targets {
		if a.uid != "" {
			uids = append(uids, a.uid)
		} else {
			nodeIDs = append(nodeIDs, a.nodeID)
		}
	}
	var s activeSight
	var err error
	if len(uids) > 0 {
		if s.procs, err = processVisibility(ctx, q, users, uids); err != nil {
			return activeSight{}, err
		}
	}
	if len(nodeIDs) > 0 {
		if s.nodes, err = nodeVisibility(ctx, q, users, nodeIDs); err != nil {
			return activeSight{}, err
		}
	}
	return s, nil
}

// listActive: 다른 계정 탭의 active 중 볼 수 있는 것 — 새 탭·재접속·공유 변경 때 처음 받기
func (r *supervisorRouter) listActive(c echo.Context) error {
	me := r.requireSession(c).Data.ID
	var tabs []browserTab
	for _, tab := range r.tabs.activeTabs() {
		if tab.userID != me {
			tabs = append(tabs, tab)
		}
	}
	out := []protocol.ActiveUpdateEvent{}
	if len(tabs) == 0 {
		return c.JSON(200, out)
	}

	// 보내는 계정도 볼 수 있어야 한다(publishActive와 같은 규칙)
	users := []int64{me}
	targets := make([]activeTarget, 0, len(tabs))
	for _, tab := range tabs {
		users = append(users, tab.userID)
		targets = append(targets, tab.active)
	}
	sight, err := loadActiveSight(c.Request().Context(), TxQueries(c), users, targets)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	for _, tab := range tabs {
		if !sight.sees(me, tab.active) || !sight.sees(tab.userID, tab.active) {
			continue
		}
		ev := protocol.ActiveUpdateEvent{Key: sidKey(tab.tabID), UserID: tab.userID, Nickname: tab.nickname}
		ev.UID, ev.NodeID = activeFields(tab.active)
		out = append(out, ev)
	}
	return c.JSON(200, out)
}
