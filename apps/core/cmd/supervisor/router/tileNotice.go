package router

import (
	"context"
	"log"
	"sync"
	"time"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
)

// 타일 위치 알림 — 개편-5(알림 이벤트), REF-tile-binding-rework.md "위치 알림"

// pinState: 계정별 마지막으로 보낸 고정 지도 — PIN_MOVED(M5(고정 이동)) 비교용. 서버 재시작 뒤 첫 계산은 비교 없이 기록만
type pinState struct {
	mu    sync.Mutex
	users map[int64]*userPins
}

// userPins.mu: 한 계정의 "읽기 -> 비교 -> 발행"을 한 줄로 — 동시에 두 저장이 끝나면 옛 지도를 늦게 읽은 쪽이 고정을 되돌린 것처럼 알리지 않게
type userPins struct {
	mu   sync.Mutex
	last map[int64]protocol.TilePin
	seen bool
}

func newPinState() *pinState { return &pinState{users: map[int64]*userPins{}} }

func (s *pinState) of(userID int64) *userPins {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[userID]
	if !ok {
		u = &userPins{}
		s.users[userID] = u
	}
	return u
}

// publishPins: 트리가 바뀔 때마다 계정 보드 전체로 고정 지도를 다시 계산해 보내고, 고정 타일이 바뀐 노드는 PIN_MOVED. 커밋 뒤 호출(publishTiles 끝)
func (r *supervisorRouter) publishPins(userID int64) {
	u := r.pins.of(userID)
	u.mu.Lock()
	defer u.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := store.GetStorePool().Queries()
	pins, err := loadPins(ctx, q, userID)
	if err != nil {
		log.Printf("[tiles] 고정 지도 계산 실패 user=%d: %v", userID, err)
		return
	}
	if err := r.subscribeHub.Publish(tilesTopic(userID), protocol.MsgTilesPins, protocol.TilesPinsEvent{Pins: pins}); err != nil {
		log.Printf("[tiles] 고정 지도 발행 실패 user=%d: %v", userID, err)
	}
	next := make(map[int64]protocol.TilePin, len(pins))
	for _, p := range pins {
		next[p.NodeID] = p
		// 핀이 두 지도 모두에 있을 때만 이동 — 새로 생김(타일 2개째)·사라짐(1개 남음)은 알리지 않음
		if old, ok := u.last[p.NodeID]; u.seen && ok && (old.Hash != p.Hash || old.TileID != p.TileID) {
			r.publishNotice(userID, protocol.TilesNoticeEvent{Kind: protocol.NoticePinMoved, NodeID: p.NodeID, Name: nodeName(ctx, q, p.NodeID),
				From: &protocol.TileSpot{Hash: old.Hash, TileID: old.TileID}, To: &protocol.TileSpot{Hash: p.Hash, TileID: p.TileID}})
		}
	}
	u.last, u.seen = next, true
}

func (r *supervisorRouter) publishNotice(userID int64, ev protocol.TilesNoticeEvent) {
	ev.At = time.Now().UnixMilli()
	if err := r.subscribeHub.Publish(tilesTopic(userID), protocol.MsgTilesNotice, ev); err != nil {
		log.Printf("[tiles] 위치 알림 발행 실패 user=%d: %v", userID, err)
	}
}

// nodeName: 알림 문구용 — 권한 판정은 알림을 만드는 쪽에서 이미 끝남
func nodeName(ctx context.Context, q *superdb.Queries, nodeID int64) string {
	n, err := q.GetNodeByID(ctx, nodeID)
	if err != nil {
		return ""
	}
	return n.Name
}

// terminalSpot: 저장된 트리에서 그 uid 터미널 타일 위치
func terminalSpot(rec superdb.TileTree, uid string) *protocol.TileSpot {
	doc := mustParseDoc(rec.Tree)
	if t, ok := doc.terminalOf(uid); ok {
		return &protocol.TileSpot{Hash: rec.Hash, TileID: t.ID}
	}
	return nil
}
