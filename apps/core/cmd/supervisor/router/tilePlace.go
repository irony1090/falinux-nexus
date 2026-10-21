package router

import (
	"context"
	"encoding/json"
	"log"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/web"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// 터미널 타일 넣기 — 실행한 계정·보는 계정 보드에 자리 찾기, 고아 process 입양

// appendTerminalTile: exec 성공 직후 터미널 타일을 넣는다 — ⑪-4(타일 없는 실행 중 process)
// 별도 트랜잭션으로 핸들러 안에서 커밋까지: 요청 트랜잭션은 응답 뒤 커밋돼 실패를 알 수 없고, 그러면 kill(보상)을 못 한다
// 고아 안전망(adoptOrphans)이 먼저 붙였으면 그 타일을 돌려준다(uid 중복 방지)
func appendTerminalTile(ctx context.Context, userID int64, hash, parentTileID string, size tileSize, nodeID int64, uid string) (tileItem, superdb.TileTree, error) {
	tx := store.GetStorePool().Transaction()
	if err := tx.Begin(ctx); err != nil {
		return tileItem{}, superdb.TileTree{}, err
	}
	tile, upd, err := func() (tileItem, superdb.TileTree, error) {
		q, err := tx.Queries()
		if err != nil {
			return tileItem{}, superdb.TileTree{}, err
		}
		if _, err := loadOrCreateTileTree(ctx, q, userID); err != nil {
			return tileItem{}, superdb.TileTree{}, err
		}
		recs, err := q.ListTileTreesForUpdate(ctx, userID)
		if err != nil {
			return tileItem{}, superdb.TileTree{}, err
		}
		var rec, def superdb.TileTree
		var doc, defDoc tileDoc
		for _, cur := range recs {
			var d tileDoc
			if err := json.Unmarshal(cur.Tree, &d); err != nil {
				return tileItem{}, superdb.TileTree{}, err
			}
			if t, ok := d.terminalOf(uid); ok {
				return t, cur, nil
			}
			if cur.IsDefault {
				def, defDoc = cur, d
			}
			if hash != "" && cur.Hash == hash {
				rec, doc = cur, d
			}
		}
		// 요청한 작업 공간이 그 사이 삭제됐으면 기본으로(삭제된 트리는 기본 루트 끝에 붙어 있음)
		if rec.Hash == "" {
			rec, doc = def, defDoc
		}
		// 부모 타일을 그 사이 다른 탭이 닫았으면 루트에 붙인다
		parent, ok := doc.Tiles[parentTileID]
		if !ok || parent.Type != "folder" {
			parent = doc.Tiles[doc.RootID]
		}
		tile := tileItem{ID: uuid.NewString(), Type: "terminal", Parent: &parent.ID, Kids: []string{}, Size: size, NodeID: &nodeID, UID: uid}
		parent.Kids = append(parent.Kids, tile.ID)
		doc.Tiles[parent.ID] = parent
		doc.Tiles[tile.ID] = tile
		tree, err := doc.marshal()
		if err != nil {
			return tileItem{}, superdb.TileTree{}, err
		}
		upd, err := q.UpdateTileTreeByHash(ctx, superdb.UpdateTileTreeByHashParams{UserID: userID, Hash: rec.Hash, Version: rec.Version, Tree: tree})
		if err != nil {
			return tileItem{}, superdb.TileTree{}, err
		}
		return tile, upd, nil
	}()
	if err != nil {
		_ = tx.Rollback(ctx, err)
		return tileItem{}, superdb.TileTree{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return tileItem{}, superdb.TileTree{}, err
	}
	return tile, upd, nil
}

// attachViewerTerminal: exec 때 실행자가 아닌 계정에 터미널 타일 — SD2(실행 때 터미널 타일을 넣을 계정). 계정마다 별도 트랜잭션(교착 없음)
func attachViewerTerminal(ctx context.Context, userID, nodeID int64, uid string, place placeFunc) (*superdb.TileTree, error) {
	tx := store.GetStorePool().Transaction()
	if err := tx.Begin(ctx); err != nil {
		return nil, err
	}
	upd, err := func() (*superdb.TileTree, error) {
		q, err := tx.Queries()
		if err != nil {
			return nil, err
		}
		return insertTerminal(ctx, q, userID, nodeID, uid, place)
	}()
	if err != nil {
		_ = tx.Rollback(ctx, err)
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return upd, nil
}

// insertTerminal: place가 고른 자리에 터미널 타일. 이미 있거나 자리가 없으면 nil. q = 호출부 트랜잭션, 계정 트리 전부 잠금(잠금 순서)
func insertTerminal(ctx context.Context, q *superdb.Queries, userID, nodeID int64, uid string, place placeFunc) (*superdb.TileTree, error) {
	if _, err := loadOrCreateTileTree(ctx, q, userID); err != nil {
		return nil, err
	}
	recs, err := q.ListTileTreesForUpdate(ctx, userID)
	if err != nil {
		return nil, err
	}
	docs := make([]tileDoc, len(recs))
	for i, rec := range recs {
		if err := json.Unmarshal(rec.Tree, &docs[i]); err != nil {
			return nil, err
		}
		if _, ok := docs[i].terminalOf(uid); ok {
			return nil, nil
		}
	}
	ancestors, err := visibleAncestors(ctx, q, userID, nodeID)
	if err != nil {
		return nil, err
	}
	at, parentID, ok := place(recs, docs, ancestors)
	if !ok {
		return nil, nil
	}
	doc := &docs[at]
	parent := doc.Tiles[parentID]
	tile := tileItem{ID: uuid.NewString(), Type: "terminal", Parent: &parent.ID, Kids: []string{}, Size: tileSize{W: 0.5, H: 0.5}, NodeID: &nodeID, UID: uid}
	parent.Kids = append(parent.Kids, tile.ID)
	doc.Tiles[parent.ID] = parent
	doc.Tiles[tile.ID] = tile
	if err := doc.validate(); err != nil {
		return nil, err
	}
	tree, err := doc.marshal()
	if err != nil {
		return nil, err
	}
	saved, err := q.UpdateTileTreeByHash(ctx, superdb.UpdateTileTreeByHashParams{UserID: userID, Hash: recs[at].Hash, Version: recs[at].Version, Tree: tree})
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

// visibleAncestors: 조상 중 userID가 볼 수 있는 것(가까운 순) — 공유가 풀려 배너만 남은 폴더 타일 아래로 넣지 않게
func visibleAncestors(ctx context.Context, q *superdb.Queries, userID, nodeID int64) ([]int64, error) {
	ancestors, err := q.ListNodeAncestors(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	out := ancestors[:0]
	for _, anc := range ancestors {
		perm, err := q.NodePermission(ctx, superdb.NodePermissionParams{UserID: userID, NodeID: anc})
		if err != nil {
			return nil, err
		}
		if perm > superdb.PermBlock {
			out = append(out, anc)
		}
	}
	return out, nil
}

// placeFunc: 터미널 타일을 넣을 보드(recs·docs 같은 순서)와 부모 타일. ok=false = 넣지 않음(위치 없음)
type placeFunc func(recs []superdb.TileTree, docs []tileDoc, ancestors []int64) (int, string, bool)

// boundPlace: 다른 계정이 실행한 터미널 — 받는 타일이 있고 그 보드를 접속 중 탭이 볼 때만(BD9(보는 보드 판정)). viewed = tabs.viewedBoards
func boundPlace(viewed map[string]bool) placeFunc {
	return func(recs []superdb.TileTree, docs []tileDoc, ancestors []int64) (int, string, bool) {
		at, id, ok := bindingTile(docs, ancestors)
		if !ok || !(viewed[recs[at].Hash] || recs[at].IsDefault && viewed[""]) {
			return 0, "", false
		}
		return at, id, true
	}
}

// bindingTile: 터미널을 받을 폴더 타일 — 폴더 타일이 있는 가장 가까운 조상에서 멈추고, 그중 betterBinding(BD1(우선순위)·BD2(바인딩 단위))
func bindingTile(docs []tileDoc, ancestors []int64) (int, string, bool) {
	for _, anc := range ancestors {
		at, best, found := 0, tileItem{}, false
		for i := range docs {
			for _, id := range docs[i].folderTilesOf(anc) {
				t := docs[i].Tiles[id]
				if !found || betterBinding(t, best) {
					at, best, found = i, t, true
				}
			}
		}
		if found {
			return at, best.ID, true
		}
	}
	return 0, "", false
}

// betterBinding: 받는 타일 비교 — 고정 > shownAt 이른 것. 같으면 먼저 본 것 유지(보드 순·너비 우선 순). bindingTile·pinsOf 공용
func betterBinding(t, best tileItem) bool {
	return (t.Bound && !best.Bound) || (t.Bound == best.Bound && t.ShownAt < best.ShownAt)
}

// adoptOrphans: 타일 없는 살아 있는 process(내가 실행한 것만, BD8(고아 안전망))를 기본 루트 끝에 — WS3(고아 안전망 위치)
// 잠그지 않고 먼저 보고, 고아가 있을 때만 잠근 뒤 다시 본다
func (r *supervisorRouter) adoptOrphans(c echo.Context, userID int64) {
	ctx := c.Request().Context()
	q := TxQueries(c)
	procs, err := q.ListOwnLiveProcesses(ctx, userID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if len(procs) == 0 {
		return
	}
	docs, err := q.ListTileTreeDocs(ctx, userID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	trees := make([][]byte, 0, len(docs))
	for _, d := range docs {
		trees = append(trees, d.Tree)
	}
	if len(findOrphans(procs, trees)) == 0 {
		return
	}

	ensureTileTree(c, userID)
	locked := lockWorkspaces(c, userID)
	var def superdb.TileTree
	trees = trees[:0]
	for _, rec := range locked {
		trees = append(trees, rec.Tree)
		if rec.IsDefault {
			def = rec
		}
	}
	orphans := findOrphans(procs, trees)
	if len(orphans) == 0 {
		return
	}
	doc := mustParseDoc(def.Tree)
	root := doc.Tiles[doc.RootID]
	uids := make([]string, 0, len(orphans))
	for _, p := range orphans {
		nodeID := p.NodeID.Int64
		t := tileItem{ID: uuid.NewString(), Type: "terminal", Parent: &root.ID, Kids: []string{}, Size: tileSize{W: 0.5, H: 0.5}, NodeID: &nodeID, UID: p.Uid, Adopted: true}
		doc.Tiles[t.ID] = t
		root.Kids = append(root.Kids, t.ID)
		uids = append(uids, p.Uid)
	}
	doc.Tiles[root.ID] = root
	upd := saveLockedTree(c, def, &doc)
	log.Printf("[tiles] 고아 터미널 %d개를 기본 작업 공간에 붙임 user=%d", len(orphans), userID)
	AfterCommit(c, func() {
		r.publishTiles(userID, upd, "")
		r.resyncSizeOwners(uids)
	})
}

func findOrphans(procs []superdb.ListOwnLiveProcessesRow, trees [][]byte) []superdb.ListOwnLiveProcessesRow {
	has := map[string]bool{}
	for _, tree := range trees {
		doc := mustParseDoc(tree)
		for _, uid := range doc.terminalUIDs() {
			has[uid] = true
		}
	}
	var out []superdb.ListOwnLiveProcessesRow
	for _, p := range procs {
		if !has[p.Uid] {
			out = append(out, p)
		}
	}
	return out
}
