package router

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/web"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// 작업 공간마다 타일 트리 문서 1개 — ⑪(타일 트리 서버 저장) REF-node-ui-save.md, 작업 공간 REF-workspace-impl.md.
// `/tiles` = 기본 작업 공간, `/tiles/:hash` = 그 외(WS1(트리 주소))
func (r *supervisorRouter) mountTiles(e *echo.Echo) {
	g := e.Group("/tiles")
	g.GET("", r.getTiles)
	g.PUT("", r.putTiles, middleware.BodyLimit("256K"))
	g.GET("/:hash", r.getTiles)
	g.PUT("/:hash", r.putTiles, middleware.BodyLimit("256K"))
}

// 프론트 tileTree.store의 Tile 미러. 서버는 모양만 검사하고 순서·닫기 규칙은 프론트 몫
type tileSize struct {
	W float64 `json:"w"`
	H float64 `json:"h"`
}

type tileItem struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"` // folder | terminal
	Parent    *string  `json:"parent"`
	Kids      []string `json:"kids"`
	Size      tileSize `json:"size"`
	NodeID    *int64   `json:"nodeId"` // folder: null = 루트 목록 / terminal: 필수
	UID       string   `json:"uid,omitempty"`
	Adopted   bool     `json:"adopted,omitempty"`
	Minimized bool     `json:"minimized,omitempty"` // 배치에서만 빠짐(트리·uid 유지) — SD3(실행 중 터미널 치우기)
}

type tileDoc struct {
	RootID string              `json:"rootId"`
	Tiles  map[string]tileItem `json:"tiles"`
}

// Hash = 그 트리의 보드 hash — 프론트가 TILES:UPDATE의 hash와 비교해 자기 보드 것만 받는다(WS6(TILES 이벤트 구분))
type tileTreeResponse struct {
	Hash    string          `json:"hash"`
	Tree    json.RawMessage `json:"tree"`
	Version int64           `json:"version"`
}

type tilePutRequest struct {
	Tree    json.RawMessage `json:"tree" validate:"required"`
	Version int64           `json:"version" validate:"required"`
}

type tilePutResponse struct {
	Version int64 `json:"version"`
}

func tilesTopic(userID int64) string { return fmt.Sprintf("TILES:%d", userID) }

// publishTiles는 계정의 모든 탭에 저장된 트리 전체를 보낸다 — 4①(push 내용). 커밋 뒤에만 부를 것
func (r *supervisorRouter) publishTiles(userID int64, rec superdb.TileTree, tabID string) {
	ev := protocol.TilesUpdateEvent{Hash: rec.Hash, Tree: rec.Tree, Version: rec.Version, TabID: tabID}
	if err := r.subscribeHub.Publish(tilesTopic(userID), protocol.MsgTilesUpdate, ev); err != nil {
		log.Printf("[tiles] 발행 실패 user=%d: %v", userID, err)
	}
}

func newRootDoc() tileDoc {
	id := uuid.NewString()
	return tileDoc{RootID: id, Tiles: map[string]tileItem{
		id: {ID: id, Type: "folder", Kids: []string{}, Size: tileSize{W: 0.5, H: 0.5}},
	}}
}

func validSpan(v float64) bool { return v == 0.5 || v == 1 }

// validate: 루트 1개 · parent/kids 쌍 일치 · 루트에서 전부 닿음(고아·순환 없음) · type/size/terminal 필드
func (d *tileDoc) validate() error {
	root, ok := d.Tiles[d.RootID]
	if !ok || root.Parent != nil || root.Type != "folder" {
		return errors.New("루트 타일이 없거나 올바르지 않습니다")
	}
	for key, t := range d.Tiles {
		if t.ID == "" || t.ID != key {
			return fmt.Errorf("타일 id가 키와 다릅니다: %q", key)
		}
		if !validSpan(t.Size.W) || !validSpan(t.Size.H) {
			return fmt.Errorf("타일 크기가 올바르지 않습니다: %s", key)
		}
		switch t.Type {
		case "folder":
			if t.UID != "" || t.Adopted || t.Minimized {
				return fmt.Errorf("폴더 타일에 터미널 필드가 있습니다: %s", key)
			}
		case "terminal":
			if t.UID == "" || t.NodeID == nil || len(t.Kids) > 0 {
				return fmt.Errorf("터미널 타일이 올바르지 않습니다: %s", key)
			}
		default:
			return fmt.Errorf("타일 type이 올바르지 않습니다: %s", key)
		}
		if t.Parent == nil {
			if key != d.RootID {
				return fmt.Errorf("루트가 아닌 타일에 parent가 없습니다: %s", key)
			}
			continue
		}
		p, ok := d.Tiles[*t.Parent]
		if !ok || p.Type != "folder" {
			return fmt.Errorf("parent가 없거나 폴더가 아닙니다: %s", key)
		}
		n := 0
		for _, k := range p.Kids {
			if k == key {
				n++
			}
		}
		if n != 1 {
			return fmt.Errorf("parent의 kids에 정확히 한 번 있어야 합니다: %s", key)
		}
	}
	// 위에서 모든 parent->kid 쌍을 확인했으니, kids 쪽만 보면 루트에서 닿는지 알 수 있다
	seen := map[string]bool{d.RootID: true}
	queue := []string{d.RootID}
	for len(queue) > 0 {
		t := d.Tiles[queue[0]]
		queue = queue[1:]
		for _, k := range t.Kids {
			kid, ok := d.Tiles[k]
			if !ok || kid.Parent == nil || *kid.Parent != t.ID || seen[k] {
				return fmt.Errorf("kids가 올바르지 않습니다: %s -> %s", t.ID, k)
			}
			seen[k] = true
			queue = append(queue, k)
		}
	}
	if len(seen) != len(d.Tiles) {
		return errors.New("루트에서 닿지 않는 타일이 있습니다")
	}
	return nil
}

// marshal: 모르는 필드는 버리고 서버 기준 모양으로 저장. kids는 null 대신 []
func (d *tileDoc) marshal() ([]byte, error) {
	for key, t := range d.Tiles {
		if t.Kids == nil {
			t.Kids = []string{}
			d.Tiles[key] = t
		}
	}
	return json.Marshal(d)
}

func (d *tileDoc) mustMarshal() []byte {
	b, err := d.marshal()
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return b
}

// loadOrCreateTileTree: 없으면 루트만 있는 트리를 만든다. 다른 탭이 먼저 만들었으면(행 없음) 그걸 다시 읽는다
func loadOrCreateTileTree(ctx context.Context, q *superdb.Queries, userID int64) (superdb.TileTree, error) {
	rec, err := q.GetTileTree(ctx, userID)
	if !errors.Is(err, pgx.ErrNoRows) {
		return rec, err
	}
	root := newRootDoc()
	tree, err := root.marshal()
	if err != nil {
		return rec, err
	}
	rec, err = q.CreateTileTree(ctx, superdb.CreateTileTreeParams{UserID: userID, Hash: newTreeHash(), Tree: tree})
	if errors.Is(err, pgx.ErrNoRows) {
		return q.GetTileTree(ctx, userID)
	}
	return rec, err
}

// newTreeHash: 작업 공간 URL·API 주소용 소문자 hex 12자 — 00008_workspaces.sql의 기존 행 채우기와 같은 형식
func newTreeHash() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func ensureTileTree(c echo.Context, userID int64) superdb.TileTree {
	rec, err := loadOrCreateTileTree(c.Request().Context(), TxQueries(c), userID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return rec
}

// requestTileTree: hash 없음 = 기본(없으면 만듦) / 있으면 그 작업 공간, 다른 계정 것이면 404
func requestTileTree(c echo.Context, userID int64) superdb.TileTree {
	hash := c.Param("hash")
	if hash == "" {
		return ensureTileTree(c, userID)
	}
	return requestWorkspace(c, userID, hash)
}

func requestWorkspace(c echo.Context, userID int64, hash string) superdb.TileTree {
	rec, err := TxQueries(c).GetTileTreeByHash(c.Request().Context(), superdb.GetTileTreeByHashParams{UserID: userID, Hash: hash})
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "보드를 찾을 수 없습니다"))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return rec
}

func mustParseDoc(tree []byte) tileDoc {
	var doc tileDoc
	if err := json.Unmarshal(tree, &doc); err != nil {
		panic(web.Err(500, "저장된 타일 트리를 읽을 수 없습니다: %v", err))
	}
	return doc
}

// checkTerminalUIDs: 저장본에 없던 터미널 uid 추가·한 트리 안 uid 중복을 막는다 — WS5(터미널 타일 추가 주체).
// 터미널 타일은 서버만 넣으므로(exec·고아 안전망·작업 공간 동작) 이것만으로 계정 전체 트리에서 uid가 하나로 유지된다
func checkTerminalUIDs(next, stored *tileDoc) error {
	had := map[string]bool{}
	for _, t := range stored.Tiles {
		if t.Type == "terminal" {
			had[t.UID] = true
		}
	}
	seen := map[string]bool{}
	for _, t := range next.Tiles {
		if t.Type != "terminal" {
			continue
		}
		if !had[t.UID] {
			return fmt.Errorf("터미널 타일은 서버만 추가할 수 있습니다: %s", t.UID)
		}
		if seen[t.UID] {
			return fmt.Errorf("같은 터미널이 두 번 있습니다: %s", t.UID)
		}
		seen[t.UID] = true
	}
	return nil
}

// appendTerminalTile: exec 성공 직후 터미널 타일을 서버 트리에 넣는다 — ⑪-4(타일 없는 실행 중 process).
// 요청 트랜잭션이 아니라 별도 트랜잭션으로 핸들러 안에서 커밋까지 끝낸다: 요청 트랜잭션은 응답을 보낸 뒤
// 커밋되므로 그 실패를 핸들러가 알 수 없고, 그러면 실패 시 process kill(보상)을 할 수 없다 (REF-node-ui-save-impl.md "3-b'(exec 연동) 구현").
// 계정 트리 전부를 잠그고 uid를 찾는다: 실행과 타일 추가 사이에 고아 안전망(adoptOrphans)이 먼저 붙였으면 그 타일을 돌려준다(uid 중복 방지, WS5(터미널 타일 추가 주체))
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

// attachViewerTerminal: exec 때 실행자가 아닌 계정의 트리에 터미널 타일을 넣는다 — SD2(실행 때 터미널 타일을 넣을 계정).
// 어느 보드에든 이미 있으면 nil / 위치 = 스크립트의 가장 가까운 조상 폴더 타일 아래, 없으면 기본 보드 루트 끝.
// 계정마다 별도 트랜잭션(한 번에 한 계정 트리만 잠금 → 교착 없음)
func attachViewerTerminal(ctx context.Context, userID, nodeID int64, uid string) (*superdb.TileTree, error) {
	tx := store.GetStorePool().Transaction()
	if err := tx.Begin(ctx); err != nil {
		return nil, err
	}
	upd, err := func() (*superdb.TileTree, error) {
		q, err := tx.Queries()
		if err != nil {
			return nil, err
		}
		if _, err := loadOrCreateTileTree(ctx, q, userID); err != nil {
			return nil, err
		}
		recs, err := q.ListTileTreesForUpdate(ctx, userID)
		if err != nil {
			return nil, err
		}
		docs := make([]tileDoc, len(recs))
		var def int
		for i, rec := range recs {
			if err := json.Unmarshal(rec.Tree, &docs[i]); err != nil {
				return nil, err
			}
			if _, ok := docs[i].terminalOf(uid); ok {
				return nil, nil
			}
			if rec.IsDefault {
				def = i
			}
		}
		ancestors, err := q.ListNodeAncestors(ctx, nodeID)
		if err != nil {
			return nil, err
		}
		at, parentID := def, docs[def].RootID
	find:
		for _, anc := range ancestors {
			for i := range docs {
				for _, t := range docs[i].Tiles {
					if t.Type == "folder" && t.NodeID != nil && *t.NodeID == anc {
						at, parentID = i, t.ID
						break find
					}
				}
			}
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

// terminalOf: uid의 터미널 타일
func (d *tileDoc) terminalOf(uid string) (tileItem, bool) {
	for _, t := range d.Tiles {
		if t.Type == "terminal" && t.UID == uid {
			return t, true
		}
	}
	return tileItem{}, false
}

// terminalUIDs: 문서 안 터미널 uid 전부
func (d *tileDoc) terminalUIDs() []string {
	var uids []string
	for _, t := range d.Tiles {
		if t.Type == "terminal" {
			uids = append(uids, t.UID)
		}
	}
	return uids
}

// adoptOrphans: 살아 있는데 계정의 어느 트리에도 타일이 없는 process를 기본 작업 공간 루트 끝에 붙인다 — WS3(고아 안전망 위치).
// 대상 = 내가 소유하거나 실행한 것만(공유받은 것은 공유-d가 SH2(받는 쪽 화면 위치)로 넣음). 먼저 잠그지 않고 보고, 고아가 있을 때만 잠근 뒤 다시 본다
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

// getTiles: 어느 작업 공간을 불러오든 고아 안전망을 먼저 돈다(/w/{hash}로만 접속하는 탭도 있어서)
func (r *supervisorRouter) getTiles(c echo.Context) error {
	sess := r.requireSession(c)
	r.adoptOrphans(c, sess.Data.ID)
	rec := requestTileTree(c, sess.Data.ID)
	return c.JSON(http.StatusOK, tileTreeResponse{Hash: rec.Hash, Tree: rec.Tree, Version: rec.Version})
}

// putTiles: 읽은 version일 때만 저장. 다르면 409 + 서버 트리 — ⑪-2(충돌 처리), 프론트가 그 위에 자기 동작을 다시 적용
func (r *supervisorRouter) putTiles(c echo.Context) error {
	sess := r.requireSession(c)
	var body tilePutRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	var doc tileDoc
	if err := json.Unmarshal(body.Tree, &doc); err != nil {
		panic(web.Err(400, "타일 트리를 읽을 수 없습니다: %v", err))
	}
	if err := doc.validate(); err != nil {
		panic(web.Err(400, "%v", err))
	}

	userID := sess.Data.ID
	cur := requestTileTree(c, userID)
	if cur.Version != body.Version {
		return c.JSON(http.StatusConflict, tileTreeResponse{Hash: cur.Hash, Tree: cur.Tree, Version: cur.Version})
	}
	stored := mustParseDoc(cur.Tree)
	if err := checkTerminalUIDs(&doc, &stored); err != nil {
		panic(web.Err(400, "%v", err))
	}

	// 읽은 뒤 다른 저장이 끼었으면 version 조건으로 걸린다 — 위 검사의 기준이 바뀐 경우도 여기서 409
	rec, err := TxQueries(c).UpdateTileTreeByHash(c.Request().Context(), superdb.UpdateTileTreeByHashParams{
		UserID: userID, Hash: cur.Hash, Version: body.Version, Tree: doc.mustMarshal(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		cur = requestTileTree(c, userID)
		return c.JSON(http.StatusConflict, tileTreeResponse{Hash: cur.Hash, Tree: cur.Tree, Version: cur.Version})
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	tabID := r.requestTab(c, sess.Name())
	AfterCommit(c, func() { r.publishTiles(userID, rec, tabID) })
	return c.JSON(http.StatusOK, tilePutResponse{Version: rec.Version})
}
