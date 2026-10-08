package router

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/transport"
)

// 크기 우선권: PTY rows/cols를 바꿀 수 있는 탭은 process당 하나 — REF-process-sync.md Q(크기 우선권 단위)

// sizeOwners는 ProcessEntry 밖에 둔다 — worker 재접속(Rebind) 때 entry가 새로 만들어져도 유지돼야 해서.
// 메모리 전용이라 supervisor 재시작 후엔 비어 있고, 다음에 연결되거나 resize하는 탭이 가져간다.
type sizeOwners struct {
	mu     sync.Mutex
	owner  map[string]string      // uid -> tabId, "" = 비어 있음
	timers map[string]*time.Timer // uid -> 넘김 대기
}

func newSizeOwners() *sizeOwners {
	return &sizeOwners{owner: map[string]string{}, timers: map[string]*time.Timer{}}
}

func (o *sizeOwners) stopTimerLocked(uid string) {
	if t := o.timers[uid]; t != nil {
		t.Stop()
		delete(o.timers, uid)
	}
}

// setSizeOwner는 알림 없이 소유자를 정한다(exec 직후 — 실행한 탭은 REST 응답으로 안다).
func (r *supervisorRouter) setSizeOwner(uid, tabID string) {
	o := r.sizeOwners
	o.mu.Lock()
	defer o.mu.Unlock()
	o.stopTimerLocked(uid)
	o.owner[uid] = tabID
}

func (r *supervisorRouter) sizeOwnerOf(uid string) string {
	o := r.sizeOwners
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.owner[uid]
}

// claimSizeOwner는 tabID가 소유자인지 반환한다. 비어 있으면 tabID가 가져간다(resize 요청 시).
func (r *supervisorRouter) claimSizeOwner(uid, tabID string) bool {
	o := r.sizeOwners
	o.mu.Lock()
	if cur := o.owner[uid]; cur != "" {
		o.mu.Unlock()
		return cur == tabID
	}
	o.owner[uid] = tabID
	o.mu.Unlock()
	r.notifySizeOwner(uid, tabID)
	return true
}

// releaseSizeOwner는 process 종료 시 기록을 지운다(PENDING은 재바인딩될 수 있어 지우지 않음).
func (r *supervisorRouter) releaseSizeOwner(uid string) {
	o := r.sizeOwners
	o.mu.Lock()
	defer o.mu.Unlock()
	o.stopTimerLocked(uid)
	delete(o.owner, uid)
}

// notifySizeOwner는 그 탭의 소켓에만 보낸다(끊겨 있으면 재연결 후 GET /processes로 안다).
func (r *supervisorRouter) notifySizeOwner(uid, tabID string) {
	conn := r.tabs.connOf(tabID)
	if conn == nil {
		return
	}
	if err := conn.Emit(protocol.MsgProcessSizeOwner, protocol.SizeOwnerEvent{UID: uid}); err != nil {
		log.Printf("[sizeOwner] 알림 실패 uid=%s: %v", uid, err)
	}
}

// sizeOwnerOnConnect는 탭이 붙은 직후(tabs.attach 뒤) 부른다: 대기 중인 넘김 취소 + 빈 소유권은 보는 탭이 가져감(resync).
// uids = 그 계정이 조작(20) 이상인 것만(subscribeLiveProcesses) — SB3(크기 우선권 후보)
func (r *supervisorRouter) sizeOwnerOnConnect(tabID string, uids []string) {
	o := r.sizeOwners
	o.mu.Lock()
	for _, uid := range uids {
		if o.owner[uid] == tabID {
			o.stopTimerLocked(uid)
		}
	}
	o.mu.Unlock()
	r.resyncSizeOwners(uids)
}

// onTabView: 탭이 보는 작업 공간이 바뀌면 그 계정이 조작할 수 있는 process의 크기 우선권을 다시 맞춘다 — WS7(보는 작업 공간 전달)
func (r *supervisorRouter) onTabView(conn *transport.Conn, userID int64) {
	conn.On(protocol.MsgTabView, func(ev protocol.Frame) {
		var body protocol.TabViewEvent
		if err := ev.Bind(&body); err != nil {
			return
		}
		if _, ok := r.tabs.setView(conn, body.Hash); !ok {
			return
		}
		procs, err := r.processManager.ListVisibleLive(userID)
		if err != nil {
			log.Printf("[sizeOwner] 살아 있는 process 조회 실패 user=%d: %v", userID, err)
			return
		}
		uids := make([]string, 0, len(procs))
		for _, p := range procs {
			if p.Permission >= superdb.PermOperate {
				uids = append(uids, p.Process.Uid)
			}
		}
		r.resyncSizeOwners(uids)
	})
}

func (r *supervisorRouter) resyncSizeOwners(uids []string) {
	for _, uid := range uids {
		r.resyncSizeOwner(uid)
	}
}

// resyncSizeOwner는 소유자가 그 터미널을 화면에 둔 탭인지 맞춘다 — WS7(보는 작업 공간 전달):
// 비어 있으면 첫 후보가 가져가고, 연결된 소유자가 후보가 아니면(다른 작업 공간을 봄·터미널이 옮겨짐) 첫 후보로 넘긴다(없으면 비움).
// 끊긴 소유자(tabGrace 대기)는 sizeOwnerOnDisconnect 규칙에 맡긴다. 끝난 process(소유 기록 없음 + memory에 없음)는 건드리지 않는다
func (r *supervisorRouter) resyncSizeOwner(uid string) {
	cands := r.sizeOwnerCandidates(uid)
	_, running := r.processManager.Get(uid)

	o := r.sizeOwners
	o.mu.Lock()
	cur, known := o.owner[uid]
	if !running && !known {
		o.mu.Unlock()
		return
	}
	if cur != "" && (o.timers[uid] != nil || r.tabs.connOf(cur) == nil) {
		o.mu.Unlock()
		return
	}
	next := ""
	for _, t := range cands {
		if t.tabID == cur {
			o.mu.Unlock()
			return
		}
		if next == "" {
			next = t.tabID
		}
	}
	if next == cur {
		o.mu.Unlock()
		return
	}
	o.owner[uid] = next
	o.mu.Unlock()

	if next != "" {
		r.notifySizeOwner(uid, next)
	}
}

// sizeOwnerOnDisconnect는 탭이 빠진 직후(tabs.detach 뒤) 부른다. tabGrace 안에 같은 탭 id로 돌아오면 유지.
func (r *supervisorRouter) sizeOwnerOnDisconnect(tabID string) {
	o := r.sizeOwners
	o.mu.Lock()
	defer o.mu.Unlock()
	for uid, cur := range o.owner {
		if cur != tabID || o.timers[uid] != nil {
			continue
		}
		o.timers[uid] = time.AfterFunc(tabGrace, func() { r.handoffSizeOwner(uid, tabID, "", false) })
	}
}

// sizeOwnerOnSignOut은 로그아웃한 sid의 탭들이 가진 소유권을 즉시 넘긴다(그 sid의 다른 탭은 후보에서 뺌).
func (r *supervisorRouter) sizeOwnerOnSignOut(sid string) {
	mine := map[string]bool{}
	for _, id := range r.tabs.tabIDsOfSid(sid) {
		mine[id] = true
	}
	o := r.sizeOwners
	owned := map[string]string{} // uid -> tabId
	o.mu.Lock()
	for uid, cur := range o.owner {
		if mine[cur] {
			owned[uid] = cur
		}
	}
	o.mu.Unlock()
	for uid, tabID := range owned {
		r.handoffSizeOwner(uid, tabID, sid, true)
	}
}

// handoffSizeOwner는 leaving의 소유권을 다음 후보 탭(sizeOwnerCandidates)에 넘긴다. 없으면 비워 둔다.
// 재접속과의 경합: connect는 tabs.attach -> o.mu 순서라, 아래 잠금 안에서 leaving의 소켓이 보이면
// 재접속이 이긴 것이므로 넘기지 않는다(force = 로그아웃은 소켓이 남아 있어도 넘김).
func (r *supervisorRouter) handoffSizeOwner(uid, leaving, excludeSid string, force bool) {
	next := ""
	for _, t := range r.sizeOwnerCandidates(uid) {
		if t.tabID != leaving && (excludeSid == "" || t.sid != excludeSid) {
			next = t.tabID
			break
		}
	}

	o := r.sizeOwners
	o.mu.Lock()
	if o.owner[uid] != leaving { // 그 사이 이미 바뀜 — 타이머도 새 소유자 것일 수 있어 건드리지 않음
		o.mu.Unlock()
		return
	}
	o.stopTimerLocked(uid)
	if !force && r.tabs.connOf(leaving) != nil {
		o.mu.Unlock()
		return
	}
	o.owner[uid] = next
	o.mu.Unlock()

	if next != "" {
		r.notifySizeOwner(uid, next)
	}
}

// sizeOwnerCandidates는 크기 우선권 후보 탭을 넘길 순서대로 반환한다 — SB3(크기 우선권 후보):
// 조작(20) 이상 계정의 탭만, process 소유 계정 탭 먼저 -> 다른 계정 탭, 각각 연결 순. PENDING은 memory에 없어 DB로 본다.
// 그중 그 uid의 터미널 타일이 있는 작업 공간을 보는 탭만 — WS7(보는 작업 공간 전달)
func (r *supervisorRouter) sizeOwnerCandidates(uid string) []browserTab {
	q := store.GetStorePool().Queries()
	viewers, err := q.ListProcessViewers(context.Background(), uid)
	if err != nil {
		log.Printf("[sizeOwner] 후보 계정 조회 실패 uid=%s: %v", uid, err)
		return nil
	}
	trees, err := q.ListTreesWithTerminal(context.Background(), uid)
	if err != nil {
		log.Printf("[sizeOwner] 터미널 위치 조회 실패 uid=%s: %v", uid, err)
		return nil
	}
	viewing := func(tabs []browserTab) []browserTab {
		var out []browserTab
		for _, t := range tabs {
			if tabViews(t, trees) {
				out = append(out, t)
			}
		}
		return out
	}
	var owner, others []browserTab
	for _, v := range viewers {
		switch {
		case v.Permission == superdb.PermOwner:
			owner = viewing(r.tabs.tabsOfUser(v.UserID))
		case v.Permission >= superdb.PermOperate:
			others = append(others, viewing(r.tabs.tabsOfUser(v.UserID))...)
		}
	}
	sort.SliceStable(others, func(i, j int) bool { return others[i].connectedAt.Before(others[j].connectedAt) })
	return append(owner, others...)
}

// tabViews: 탭이 trees 중 자기 계정의 작업 공간을 보는 중인지(viewing "" = 기본)
func tabViews(tab browserTab, trees []superdb.ListTreesWithTerminalRow) bool {
	for _, t := range trees {
		if t.UserID == tab.userID && (tab.viewing == t.Hash || (tab.viewing == "" && t.IsDefault)) {
			return true
		}
	}
	return false
}
