package router

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"

	"github.com/jackc/pgx/v5/pgtype"
)

// 타일 위치 알림 — 개편-5(알림 이벤트), REF-tile-binding-rework.md "위치 알림"

// pinState: 계정별 마지막으로 보낸 고정 지도 — M5(고정 이동) 비교용. 서버 재시작 뒤 첫 계산은 기록만
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

// publishPins: 계정 보드 전체로 고정 지도를 다시 계산해 보내고, 고정 타일이 바뀐 노드는 PIN_MOVED(publishTiles 끝에서)
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
	r.publishUser(userID, protocol.MsgTilesPins, protocol.TilesPinsEvent{Pins: pins})
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
	r.publishUser(userID, protocol.MsgTilesNotice, ev)
}

// nodeName: 알림 문구용 — 권한 판정은 알림을 만드는 쪽에서 이미 끝남
func nodeName(ctx context.Context, q *superdb.Queries, nodeID int64) string {
	n, err := q.GetNodeByID(ctx, nodeID)
	if err != nil {
		return ""
	}
	return n.Name
}

// procName: 알림·목록에 쓸 process 이름 — 스크립트면 노드 이름, 장비 파일 실행이면 파일 이름(폴더-4(실행))
func procName(ctx context.Context, q *superdb.Queries, p superdb.Process) string {
	if p.NodeID.Valid {
		return nodeName(ctx, q, p.NodeID.Int64)
	}
	return execName(p.Cmd, p.Args)
}

// execName: 장비 실행 이름 — 편집 = "편집 <파일 이름>", 그 밖은 fileBase
func execName(cmd string, args []string) string {
	if cmd == protocol.PlaceholderWorkerEditor && len(args) > 0 {
		return "편집 " + fileBase(args[0])
	}
	return fileBase(cmd)
}

// fileBase: 경로 끝 이름 — worker가 Windows일 수 있어 `\`도 구분자로. "여기서 셸" = "셸", 편집 = "편집"
func fileBase(p string) string {
	switch p {
	case protocol.PlaceholderWorkerShell:
		return "셸"
	case protocol.PlaceholderWorkerEditor:
		return "편집" // 인자가 없을 때 — 파일 이름까지는 execName
	}
	return p[strings.LastIndexAny(p, `/\`)+1:]
}

// nullID: NULL 가능 id -> 포인터(NULL = nil)
func nullID(v pgtype.Int8) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

// terminalSpot: 저장된 트리에서 그 uid 터미널 타일 위치
func terminalSpot(rec superdb.TileTree, uid string) *protocol.TileSpot {
	doc := mustParseDoc(rec.Tree)
	if t, ok := doc.terminalOf(uid); ok {
		return &protocol.TileSpot{Hash: rec.Hash, TileID: t.ID}
	}
	return nil
}
