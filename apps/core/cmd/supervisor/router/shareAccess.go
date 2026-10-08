package router

import (
	"context"
	"log"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/web"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// 공유가 바뀐 뒤 출력 구독·크기 우선권을 다시 맞춘다 — SD5(구독·권한 재계산 방식), REF-process-share-reconcile.md

// reconcileAccess는 users × uids를 다시 판정해 그 계정 소켓의 출력 구독을 맞추고(볼 수 있으면 구독, 0이면 해제)
// 크기 우선권을 다시 맞춘 뒤 users 전부에 SHARE:CHANGE를 보낸다. 공유 걸기·풀기·노드 이동·삭제의 커밋 뒤에만 부를 것.
// 순서 의존: 판정 읽기와 구독 적용을 accessMu 안에서 한 번에 한다. 밖에서 읽으면 "볼 수 없음"을 먼저 읽은 재계산이
// 그 뒤 공유를 다시 건 재계산보다 늦게 적용돼 구독을 빼 버린다. 소켓 연결 구독(subscribeLiveProcesses)도 같은 잠금.
// exec의 subscribeViewers는 새 uid에 추가만 하므로 잠그지 않는다
func (r *supervisorRouter) reconcileAccess(users []int64, uids []string) {
	users, uids = uniq(users), uniq(uids)
	if len(users) == 0 {
		return
	}
	if len(uids) > 0 {
		r.applyAccess(users, uids)
		r.resyncSizeOwners(uids)
	}
	for _, u := range users {
		r.publishShareChange(u)
	}
}

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

func (r *supervisorRouter) publishShareChange(userID int64) {
	if err := r.subscribeHub.Publish(tilesTopic(userID), protocol.MsgShareChange, protocol.ShareChangeEvent{}); err != nil {
		log.Printf("[access] SHARE:CHANGE 발행 실패 user=%d: %v", userID, err)
	}
}

// liveUnder: 노드 자신과 하위의 살아 있는 process uid(요청 트랜잭션 안에서 — 커밋 뒤 재계산 대상)
func liveUnder(c echo.Context, nodeID int64) []string {
	uids, err := TxQueries(c).ListLiveProcessesUnder(c.Request().Context(), nodeID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return uids
}

// attachSharedTile은 받는 쪽 기본 보드 루트 끝에 타일을 넣는다 — SD7(받는 쪽 타일 추가 시점).
// has = 받는 쪽 어느 보드에든 이미 있으면 안 넣음(nil 반환). 넣었으면 저장된 트리를 반환 — 커밋 뒤 TILES:UPDATE는 호출부가
func attachSharedTile(c echo.Context, userID int64, has func(*tileDoc) bool, tile tileItem) *superdb.TileTree {
	ensureTileTree(c, userID)
	var def superdb.TileTree
	for _, rec := range lockWorkspaces(c, userID) {
		doc := mustParseDoc(rec.Tree)
		if has(&doc) {
			return nil
		}
		if rec.IsDefault {
			def = rec
		}
	}
	doc := mustParseDoc(def.Tree)
	root := doc.Tiles[doc.RootID]
	tile.ID = uuid.NewString()
	tile.Parent = &root.ID
	tile.Kids = []string{}
	tile.Size = tileSize{W: 0.5, H: 0.5}
	doc.Tiles[tile.ID] = tile
	root.Kids = append(root.Kids, tile.ID)
	doc.Tiles[root.ID] = root
	upd := saveLockedTree(c, def, &doc)
	return &upd
}

// attachSharedTerminals는 uids 중 userID가 볼 수 있는 살아 있는 process의 터미널 타일을 받는 쪽 기본 보드 끝에 넣는다 — 폴더 공유를 건 시점에
// 이미 실행 중이던 터미널을 받는 쪽도 볼 수 있게(exec 때는 그 시점의 볼 수 있는 계정에만 들어가므로 빈틈). 이미 타일이 있는 uid는 건너뜀.
// 마지막으로 저장한 트리(앞선 타일을 모두 포함)를 반환, 넣은 게 없으면 nil
func attachSharedTerminals(c echo.Context, userID int64, uids []string) *superdb.TileTree {
	var last *superdb.TileTree
	for _, uid := range uids {
		if mustProcessPerm(c, userID, uid) <= superdb.PermBlock {
			continue
		}
		rec, err := TxQueries(c).GetProcess(c.Request().Context(), uid)
		if err != nil || !rec.NodeID.Valid {
			continue
		}
		nodeID := rec.NodeID.Int64
		if t := attachSharedTile(c, userID, hasTerminalOf(uid), tileItem{Type: "terminal", NodeID: &nodeID, UID: uid}); t != nil {
			last = t
		}
	}
	return last
}

// hasFolderOf: 그 노드를 보여 주는 폴더 타일이 있나(다른 폴더로 이동해 간 타일은 아님)
func hasFolderOf(nodeID int64) func(*tileDoc) bool {
	return func(d *tileDoc) bool {
		for _, t := range d.Tiles {
			if t.Type == "folder" && t.NodeID != nil && *t.NodeID == nodeID {
				return true
			}
		}
		return false
	}
}

func hasTerminalOf(uid string) func(*tileDoc) bool {
	return func(d *tileDoc) bool {
		_, ok := d.terminalOf(uid)
		return ok
	}
}

// afterShareChange: 커밋 뒤 재계산 -> 받는 쪽 타일 push. 구독을 먼저 넣어야 타일을 받은 탭의 복원 조회 뒤 출력이 빠지지 않는다
func (r *supervisorRouter) afterShareChange(c echo.Context, userID int64, uids []string, added *superdb.TileTree) {
	AfterCommit(c, func() {
		r.reconcileAccess([]int64{userID}, uids)
		if added != nil {
			r.publishTiles(userID, *added, "")
		}
	})
}

func uniq[T comparable](in []T) []T {
	seen := make(map[T]bool, len(in))
	out := make([]T, 0, len(in))
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
