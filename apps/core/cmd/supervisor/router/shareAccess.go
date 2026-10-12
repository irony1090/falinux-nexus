package router

import (
	"context"
	"log"
	"slices"

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

// attachSharedFolder는 받는 쪽 기본 보드 루트 끝에 폴더 타일을 넣는다 — SD7(받는 쪽 타일 추가 시점)·SH2(받는 쪽 화면 위치).
// 받는 쪽 어느 보드에든 그 노드의 폴더 타일이 이미 있으면 안 넣음(nil 반환). 커밋 뒤 TILES:UPDATE는 호출부가.
// uid = 공유한 타일의 uid(FU2). 받는 쪽 계정에 이미 같은 uid가 있으면(그 타일이 다른 폴더로 들어간 뒤 다시 공유 등) 새로 발급 — 계정 안 uid는 하나
func attachSharedFolder(c echo.Context, userID, nodeID int64, uid string) *superdb.TileTree {
	ensureTileTree(c, userID)
	var def superdb.TileTree
	for _, rec := range lockWorkspaces(c, userID) {
		doc := mustParseDoc(rec.Tree)
		if _, ok := doc.folderTileOf(nodeID); ok {
			return nil
		}
		if uid != "" && doc.hasFolderUID(uid) {
			uid = ""
		}
		if rec.IsDefault {
			def = rec
		}
	}
	doc := mustParseDoc(def.Tree)
	root := doc.Tiles[doc.RootID]
	tile := tileItem{ID: uuid.NewString(), Type: "folder", Parent: &root.ID, Kids: []string{}, Size: tileSize{W: 0.5, H: 0.5}, NodeID: &nodeID, UID: uid}
	doc.Tiles[tile.ID] = tile
	root.Kids = append(root.Kids, tile.ID)
	doc.Tiles[root.ID] = root
	upd := saveLockedTree(c, def, &doc)
	return &upd
}

// sharedFromUID: 공유 버튼을 누른 내 폴더 타일의 uid(지금 그 노드를 보여 주는 타일만, 아직 저장 전이라 uid가 없으면 "")
func sharedFromUID(c echo.Context, userID int64, tileID string, nodeID int64) string {
	if tileID == "" {
		return ""
	}
	docs, err := TxQueries(c).ListTileTreeDocs(c.Request().Context(), userID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	for _, d := range docs {
		doc := mustParseDoc(d.Tree)
		if t, ok := doc.Tiles[tileID]; ok && t.Type == "folder" && t.NodeID != nil && *t.NodeID == nodeID {
			return t.UID
		}
	}
	return ""
}

// attachSharedTerminals는 uids 중 userID가 볼 수 있는 살아 있는 process의 터미널 타일을 넣는다(이미 있으면 건너뜀) — 폴더 공유를 건 시점에
// 이미 실행 중이던 터미널·노드 이동으로 새로 보게 된 터미널(exec 때는 그 시점의 볼 수 있는 계정에만 들어가므로 빈틈). 바뀐 보드는 saved에
func attachSharedTerminals(c echo.Context, userID int64, uids []string, saved savedTrees) {
	for _, uid := range uids {
		if mustProcessPerm(c, userID, uid) <= superdb.PermBlock {
			continue
		}
		rec, err := TxQueries(c).GetProcess(c.Request().Context(), uid)
		if err != nil || !rec.NodeID.Valid {
			continue
		}
		saved.add(attachTerminal(c, userID, rec.NodeID.Int64, uid))
	}
}

// attachTerminal: insertTerminal(placeTerminal 위치)을 요청 트랜잭션으로 — 공유 시 터미널·내 보드에 열기·process 단독 공유·노드 이동
func attachTerminal(c echo.Context, userID, nodeID int64, uid string) *superdb.TileTree {
	t, err := insertTerminal(c.Request().Context(), TxQueries(c), userID, nodeID, uid)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return t
}

// savedTrees: 한 계정에서 바뀐 보드(hash별 마지막 저장본) — 폴더 타일과 터미널이 다른 보드에 들어갈 수 있어 보드마다 push
type savedTrees map[string]superdb.TileTree

func (s savedTrees) add(t *superdb.TileTree) {
	if t != nil {
		s[t.Hash] = *t
	}
}

// afterShareChange: 커밋 뒤 재계산 -> 받는 쪽 타일 push. 구독을 먼저 넣어야 타일을 받은 탭의 복원 조회 뒤 출력이 빠지지 않는다
func (r *supervisorRouter) afterShareChange(c echo.Context, userID int64, uids []string, saved savedTrees) {
	AfterCommit(c, func() {
		r.reconcileAccess([]int64{userID}, uids)
		for _, t := range saved {
			r.publishTiles(userID, t, "")
		}
	})
}

// ===== 노드 이동으로 새로 보게 된 터미널 =====

type accessKey struct {
	user int64
	uid  string
}

// permIndex: users × uids 단계(요청 트랜잭션 — 같은 요청의 이동이 보임)
func permIndex(c echo.Context, users []int64, uids []string) map[accessKey]int16 {
	out := map[accessKey]int16{}
	if len(users) == 0 || len(uids) == 0 {
		return out
	}
	rows, err := TxQueries(c).ListProcessPermissionMatrix(c.Request().Context(), superdb.ListProcessPermissionMatrixParams{UserIds: users, Uids: uids})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	for _, row := range rows {
		out[accessKey{row.UserID, row.Uid}] = row.Permission
	}
	return out
}

// attachNewlyVisible: 이동 전(before) 못 보다가 이동 뒤 보게 된 계정 × process에 터미널 타일 — SD7(받는 쪽 타일 추가 시점)과 같은 규칙.
// 순서 의존: 계정 id 순으로 넣는다. 다른 계정 트리를 여럿 잠그므로, 두 이동이 서로 반대 순서로 잠가 교착되지 않게
func attachNewlyVisible(c echo.Context, before map[accessKey]int16, users []int64, uids []string) map[int64]savedTrees {
	after := permIndex(c, users, uids)
	sorted := slices.Sorted(slices.Values(users))
	out := map[int64]savedTrees{}
	for _, u := range sorted {
		var fresh []string
		for _, uid := range uids {
			k := accessKey{u, uid}
			if after[k] > superdb.PermBlock && before[k] <= superdb.PermBlock {
				fresh = append(fresh, uid)
			}
		}
		if len(fresh) > 0 {
			out[u] = savedTrees{}
			attachSharedTerminals(c, u, fresh, out[u])
		}
	}
	return out
}

// publishSaved: 계정별로 바뀐 보드 push(커밋 뒤)
func (r *supervisorRouter) publishSaved(saved map[int64]savedTrees) {
	for u, trees := range saved {
		for _, t := range trees {
			r.publishTiles(u, t, "")
		}
	}
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
