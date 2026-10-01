package router

import (
	"sort"
	"sync"
	"time"

	"nexus/internal/transport"
	"nexus/internal/util"
)

// 탭 등록부: 브라우저 소켓 연결 하나 = 탭 하나. 탭 id는 서버가 발급한다 — REF-process-sync.md "탭 id 수명"

const tabGrace = 60 * time.Second // 끊긴 탭 id(와 그 탭의 크기 우선권)를 보관하는 시간

type browserTab struct {
	sid         string
	userID      int64
	tabID       string
	connectedAt time.Time // 처음 연결된 시각(같은 id로 다시 붙어도 유지) — 크기 우선권 다음 후보 순서
}

type releasedTab struct {
	tab   browserTab
	timer *time.Timer
}

type tabRegistry struct {
	mu       sync.Mutex
	conns    map[*transport.Conn]*browserTab
	byTab    map[string]*transport.Conn
	released map[string]*releasedTab // 끊긴 뒤 tabGrace 안의 탭 id
}

func newTabRegistry() *tabRegistry {
	return &tabRegistry{
		conns:    map[*transport.Conn]*browserTab{},
		byTab:    map[string]*transport.Conn{},
		released: map[string]*releasedTab{},
	}
}

// attach는 conn을 탭으로 등록한다. want가 같은 sid의 끊긴 탭 id면 재사용, 아니면 새로 발급한다
// (탭 복제·세션 복원으로 sessionStorage가 복사돼 같은 id가 동시에 오는 경우 = 지금 쓰는 소켓이 있으니 새로 발급).
func (t *tabRegistry) attach(conn *transport.Conn, sid string, userID int64, want string) (browserTab, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if rel, ok := t.released[want]; ok && rel.tab.sid == sid {
		rel.timer.Stop()
		delete(t.released, want)
		tab := rel.tab
		t.conns[conn] = &tab
		t.byTab[want] = conn
		return tab, nil
	}

	id, err := t.newIDLocked()
	if err != nil {
		return browserTab{}, err
	}
	tab := browserTab{sid: sid, userID: userID, tabID: id, connectedAt: time.Now()}
	t.conns[conn] = &tab
	t.byTab[id] = conn
	return tab, nil
}

func (t *tabRegistry) newIDLocked() (string, error) {
	for {
		id, err := util.RandomKey(16, "", "")
		if err != nil {
			return "", err
		}
		if _, used := t.byTab[id]; used {
			continue
		}
		if _, used := t.released[id]; used {
			continue
		}
		return id, nil
	}
}

// detach는 conn을 빼고 그 탭 id를 tabGrace 동안 보관한다.
func (t *tabRegistry) detach(conn *transport.Conn) (browserTab, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	tab, ok := t.conns[conn]
	if !ok {
		return browserTab{}, false
	}
	delete(t.conns, conn)
	delete(t.byTab, tab.tabID)

	rel := &releasedTab{tab: *tab}
	rel.timer = time.AfterFunc(tabGrace, func() {
		t.mu.Lock()
		defer t.mu.Unlock()
		if t.released[tab.tabID] == rel {
			delete(t.released, tab.tabID)
		}
	})
	t.released[tab.tabID] = rel
	return *tab, true
}

// connOf는 탭 id의 연결된 소켓을 반환한다(끊겨 있으면 nil).
func (t *tabRegistry) connOf(tabID string) *transport.Conn {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.byTab[tabID]
}

// belongsTo는 탭 id가 sid의 것인지 반환한다(끊긴 지 tabGrace 안이면 포함 — REST가 소켓 재연결 중에 올 수 있어서).
func (t *tabRegistry) belongsTo(tabID, sid string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if conn, ok := t.byTab[tabID]; ok {
		return t.conns[conn].sid == sid
	}
	if rel, ok := t.released[tabID]; ok {
		return rel.tab.sid == sid
	}
	return false
}

// tabIDsOfSid는 sid의 탭 id를 연결된 것·보관 중인 것 모두 반환한다.
func (t *tabRegistry) tabIDsOfSid(sid string) []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	var ids []string
	for _, tab := range t.conns {
		if tab.sid == sid {
			ids = append(ids, tab.tabID)
		}
	}
	for id, rel := range t.released {
		if rel.tab.sid == sid {
			ids = append(ids, id)
		}
	}
	return ids
}

// connsOfUser는 계정의 연결된 소켓 전부를 반환한다.
func (t *tabRegistry) connsOfUser(userID int64) []*transport.Conn {
	t.mu.Lock()
	defer t.mu.Unlock()
	var conns []*transport.Conn
	for conn, tab := range t.conns {
		if tab.userID == userID {
			conns = append(conns, conn)
		}
	}
	return conns
}

// tabsOfUser는 계정의 연결된 탭을 먼저 연결된 순서로 반환한다.
func (t *tabRegistry) tabsOfUser(userID int64) []browserTab {
	t.mu.Lock()
	defer t.mu.Unlock()
	var tabs []browserTab
	for _, tab := range t.conns {
		if tab.userID == userID {
			tabs = append(tabs, *tab)
		}
	}
	sort.Slice(tabs, func(i, j int) bool { return tabs[i].connectedAt.Before(tabs[j].connectedAt) })
	return tabs
}
