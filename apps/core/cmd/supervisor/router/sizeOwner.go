package router

import (
	"context"
	"log"
	"sync"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/supervisor/store"
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

// sizeOwnerOnConnect는 탭이 붙은 직후(tabs.attach 뒤) 부른다: 대기 중인 넘김 취소 + 빈 소유권 가져가기.
func (r *supervisorRouter) sizeOwnerOnConnect(tabID string, uids []string) {
	o := r.sizeOwners
	var claimed []string
	o.mu.Lock()
	for _, uid := range uids {
		switch o.owner[uid] {
		case tabID:
			o.stopTimerLocked(uid)
		case "":
			o.owner[uid] = tabID
			claimed = append(claimed, uid)
		}
	}
	o.mu.Unlock()
	for _, uid := range claimed {
		r.notifySizeOwner(uid, tabID)
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

// handoffSizeOwner는 leaving의 소유권을 process 소유 계정의 연결된 탭 중 먼저 연결된 탭에 넘긴다. 없으면 비워 둔다.
// 재접속과의 경합: connect는 tabs.attach -> o.mu 순서라, 아래 잠금 안에서 leaving의 소켓이 보이면
// 재접속이 이긴 것이므로 넘기지 않는다(force = 로그아웃은 소켓이 남아 있어도 넘김).
func (r *supervisorRouter) handoffSizeOwner(uid, leaving, excludeSid string, force bool) {
	next := ""
	if userID, ok := r.processOwner(uid); ok {
		for _, t := range r.tabs.tabsOfUser(userID) {
			if t.tabID != leaving && (excludeSid == "" || t.sid != excludeSid) {
				next = t.tabID
				break
			}
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

// processOwner는 uid의 소유 계정을 반환한다(PENDING은 memory에 없어 DB로).
func (r *supervisorRouter) processOwner(uid string) (int64, bool) {
	if entry, ok := r.processManager.Get(uid); ok && entry.Record != nil {
		return entry.Record.OwnerUserID, true
	}
	rec, err := store.GetStorePool().Queries().GetProcess(context.Background(), uid)
	if err != nil {
		log.Printf("[sizeOwner] 소유 계정 조회 실패 uid=%s: %v", uid, err)
		return 0, false
	}
	return rec.OwnerUserID, true
}
