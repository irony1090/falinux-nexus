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
	"slices"
	"time"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/util"
	"nexus/internal/web"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// 작업 공간마다 타일 트리 문서 1개 — `/tiles` = 기본, `/tiles/:hash` = 그 외(WS1(트리 주소), REF-node-ui-save.md)
func (r *supervisorRouter) mountTiles(e *echo.Echo) {
	g := e.Group("/tiles")
	g.GET("", r.getTiles)
	g.PUT("", r.putTiles, middleware.BodyLimit("256K"))
	g.GET("/:hash", r.getTiles)
	g.PUT("/:hash", r.putTiles, middleware.BodyLimit("256K"))
	g.POST("/bind", r.bindTile)
}

// 프론트 tileTree.store의 Tile 미러. 서버는 모양만 검사하고 순서·닫기 규칙은 프론트 몫
type tileSize struct {
	W float64 `json:"w"`
	H float64 `json:"h"`
}

type tileItem struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"` // folder | terminal
	Parent       *string  `json:"parent"`
	Kids         []string `json:"kids"`
	Size         tileSize `json:"size"`
	NodeID       *int64   `json:"nodeId"`        // folder: null = 루트 목록 / terminal: 필수
	UID          string   `json:"uid,omitempty"` // terminal: process uid / folder: 서버가 발급하는 타일 uid(FU1), 공유로 생긴 받는 쪽 타일은 공유한 타일과 같은 값(FU2)
	Adopted      bool     `json:"adopted,omitempty"`
	Minimized    bool     `json:"minimized,omitempty"`    // 배치에서만 빠짐(트리·uid 유지) — SD3(실행 중 터미널 치우기)
	ShortcutsOff bool     `json:"shortcutsOff,omitempty"` // terminal: 이 타일이 active인 동안 앱 단축키 끔 — 이동-3(타일별 단축키 끔)
	ShownAt      int64    `json:"shownAt,omitempty"`      // folder: 지금 노드를 보여 주기 시작한 시각(unix ms, 서버만 씀) — BD3(저장 위치)
	Bound        bool     `json:"bound,omitempty"`        // folder: 터미널 받는 타일로 고정(서버만 씀) — BD4(설정 방법)
}

type tileDoc struct {
	RootID string              `json:"rootId"`
	Tiles  map[string]tileItem `json:"tiles"`
}

// Hash = 그 트리의 보드 hash — 프론트가 TILES:UPDATE의 hash와 비교해 자기 보드 것만 받는다(WS6(TILES 이벤트 구분))
type tileTreeResponse struct {
	Hash    string              `json:"hash"`
	Tree    json.RawMessage     `json:"tree"`
	Version int64               `json:"version"`
	Pins    *[]protocol.TilePin `json:"pins,omitempty"` // GET만 — 개편-1(고정 지도)
}

type tilePutRequest struct {
	Tree    json.RawMessage `json:"tree" validate:"required"`
	Version int64           `json:"version" validate:"required"`
}

type tilePutResponse struct {
	Version int64 `json:"version"`
}

// publishTiles: 계정의 모든 탭에 저장된 트리 전체 — 4①(push 내용)
func (r *supervisorRouter) publishTiles(userID int64, rec superdb.TileTree, tabID string) {
	ev := protocol.TilesUpdateEvent{Hash: rec.Hash, Tree: rec.Tree, Version: rec.Version, TabID: tabID}
	if err := r.subscribeHub.Publish(userTopic(userID), protocol.MsgTilesUpdate, ev); err != nil {
		log.Printf("[tiles] 발행 실패 user=%d: %v", userID, err)
	}
	r.publishPins(userID)
}

// loadPins: 잠그지 않는 읽기(id 순)로 고정 지도 계산
func loadPins(ctx context.Context, q *superdb.Queries, userID int64) ([]protocol.TilePin, error) {
	recs, err := q.ListTileTreeDocs(ctx, userID)
	if err != nil {
		return nil, err
	}
	hashes, docs := make([]string, len(recs)), make([]tileDoc, len(recs))
	for i, rec := range recs {
		hashes[i] = rec.Hash
		if err := json.Unmarshal(rec.Tree, &docs[i]); err != nil {
			return nil, err
		}
	}
	return pinsOf(hashes, docs), nil
}

// pinsOf: 폴더 노드마다 받는 타일(bindingTile과 같은 비교)을 고르되, 그 노드 타일이 2개 이상인 것만 — R4(고정 표시 남발).
// hashes·docs = 보드 id 순(bindingTile의 recs 순서와 같아야 동률일 때 같은 타일을 고름)
func pinsOf(hashes []string, docs []tileDoc) []protocol.TilePin {
	type cand struct {
		at, n int
		best  tileItem
	}
	cands := map[int64]*cand{}
	var order []int64
	for i := range docs {
		for _, t := range docs[i].folderTiles() {
			c, ok := cands[*t.NodeID]
			if !ok {
				cands[*t.NodeID] = &cand{at: i, n: 1, best: t}
				order = append(order, *t.NodeID)
				continue
			}
			c.n++
			if betterBinding(t, c.best) {
				c.at, c.best = i, t
			}
		}
	}
	out := []protocol.TilePin{}
	for _, id := range order {
		if c := cands[id]; c.n >= 2 {
			out = append(out, protocol.TilePin{NodeID: id, Hash: hashes[c.at], TileID: c.best.ID})
		}
	}
	return out
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
			if t.Adopted || t.Minimized || t.ShortcutsOff {
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

// marshal: 모르는 필드는 버리고 서버 기준 모양으로 저장. kids는 null 대신 [], 폴더 타일의 빈 uid는 발급·빈 shownAt은 지금 — 모든 저장이 여기를 지난다
func (d *tileDoc) marshal() ([]byte, error) {
	now := time.Now().UnixMilli()
	for key, t := range d.Tiles {
		if t.Kids == nil {
			t.Kids = []string{}
		}
		if t.Type == "folder" && t.UID == "" {
			uid, err := util.RandomKey(16, "", "")
			if err != nil {
				return nil, err
			}
			t.UID = uid
		}
		if t.Type == "folder" && t.ShownAt == 0 {
			t.ShownAt = now
		}
		d.Tiles[key] = t
	}
	return json.Marshal(d)
}

// keepFolderMeta: 프론트가 보낸 폴더 uid·shownAt·bound는 안 믿고 저장본 값을 쓴다(없으면 marshal이 채움, FU2·FU3)
// shownAt·bound는 보여 주는 노드가 그대로일 때만 — 다른 폴더로 가면 "새로 연" 것(BD3(저장 위치))
func (d *tileDoc) keepFolderMeta(stored *tileDoc) {
	for key, t := range d.Tiles {
		t.ShownAt, t.Bound = 0, false
		if t.Type == "folder" {
			t.UID = ""
			if s, ok := stored.Tiles[key]; ok && s.Type == "folder" {
				t.UID = s.UID
				if sameNode(s.NodeID, t.NodeID) {
					t.ShownAt, t.Bound = s.ShownAt, s.Bound
				}
			}
		}
		d.Tiles[key] = t
	}
}

func sameNode(a, b *int64) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
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

// checkTerminalUIDs: 새 터미널 uid 추가·한 트리 안 중복을 막는다 — 터미널 타일은 서버만 넣으므로 계정 안 uid가 하나로 유지(WS5(터미널 타일 추가 주체))
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

// closedTerminals: 저장본에 있었는데 새 트리에서 빠진 터미널 uid(사용자 저장에서 빠짐 = 닫기, checkTerminalUIDs가 추가는 막음)
func closedTerminals(next, stored *tileDoc) []string {
	keep := map[string]bool{}
	for _, uid := range next.terminalUIDs() {
		keep[uid] = true
	}
	var out []string
	for _, uid := range stored.terminalUIDs() {
		if !keep[uid] {
			out = append(out, uid)
		}
	}
	return out
}

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

// removeLeaf: 자식 없는 타일(터미널)을 부모 kids와 트리에서 뺀다
func (d *tileDoc) removeLeaf(id string) {
	t := d.Tiles[id]
	if t.Parent != nil {
		p := d.Tiles[*t.Parent]
		p.Kids = slices.DeleteFunc(slices.Clone(p.Kids), func(k string) bool { return k == id })
		d.Tiles[p.ID] = p
	}
	delete(d.Tiles, id)
}

// betterBinding: 받는 타일 비교 — 고정 > shownAt 이른 것. 같으면 먼저 본 것 유지(보드 순·너비 우선 순). bindingTile·pinsOf 공용
func betterBinding(t, best tileItem) bool {
	return (t.Bound && !best.Bound) || (t.Bound == best.Bound && t.ShownAt < best.ShownAt)
}

func (d *tileDoc) hasFolderUID(uid string) bool {
	for _, t := range d.Tiles {
		if t.Type == "folder" && t.UID == uid {
			return true
		}
	}
	return false
}

// folderTileOf: 그 노드를 보여 주는 폴더 타일 중 루트에서 너비 우선으로 처음 닿는 것(map 순서에 기대지 않게)
func (d *tileDoc) folderTileOf(nodeID int64) (string, bool) {
	ids := d.folderTilesOf(nodeID)
	if len(ids) == 0 {
		return "", false
	}
	return ids[0], true
}

// folderTilesOf: 그 노드를 보여 주는 폴더 타일 전부(너비 우선 순)
func (d *tileDoc) folderTilesOf(nodeID int64) []string {
	var out []string
	for _, t := range d.folderTiles() {
		if *t.NodeID == nodeID {
			out = append(out, t.ID)
		}
	}
	return out
}

// folderTiles: 노드를 보여 주는 폴더 타일 전부(루트 목록 제외, 루트에서 너비 우선 순)
func (d *tileDoc) folderTiles() []tileItem {
	var out []tileItem
	queue := []string{d.RootID}
	for len(queue) > 0 {
		t := d.Tiles[queue[0]]
		queue = queue[1:]
		if t.Type == "folder" && t.NodeID != nil {
			out = append(out, t)
		}
		queue = append(queue, t.Kids...)
	}
	return out
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

// getTiles: 어느 작업 공간을 불러오든 고아 안전망을 먼저 돈다(/w/{hash}로만 접속하는 탭도 있어서)
func (r *supervisorRouter) getTiles(c echo.Context) error {
	sess := r.requireSession(c)
	r.adoptOrphans(c, sess.Data.ID)
	rec := requestTileTree(c, sess.Data.ID)
	pins, err := loadPins(c.Request().Context(), TxQueries(c), sess.Data.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return c.JSON(http.StatusOK, tileTreeResponse{Hash: rec.Hash, Tree: rec.Tree, Version: rec.Version, Pins: &pins})
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
	doc.keepFolderMeta(&stored)

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
	closed := closedTerminals(&doc, &stored)
	AfterCommit(c, func() {
		r.publishTiles(userID, rec, tabID)
		r.forceAbandoned(closed)
	})
	return c.JSON(http.StatusOK, tilePutResponse{Version: rec.Version})
}

type tileBindRequest struct {
	Hash   string `json:"hash"` // "" = 기본 보드
	TileID string `json:"tileId" validate:"required"`
}

// bindTile: "여기로 타일 고정" — BD4(설정 방법), 해제 없음. 같은 노드의 다른 고정을 푼다(노드마다 하나). 루트 목록 타일 = 400
func (r *supervisorRouter) bindTile(c echo.Context) error {
	sess := r.requireSession(c)
	var body tileBindRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	userID := sess.Data.ID
	def := ensureTileTree(c, userID)
	locked := lockWorkspaces(c, userID)
	hash := body.Hash
	if hash == "" {
		hash = def.Hash
	}
	target := mustParseDoc(mustWorkspace(locked, hash).Tree).Tiles[body.TileID]
	if target.Type != "folder" || target.NodeID == nil {
		panic(web.Err(400, "타일 고정할 수 없는 타일입니다"))
	}
	saved := savedTrees{}
	for h, rec := range locked {
		doc := mustParseDoc(rec.Tree)
		changed := false
		for key, t := range doc.Tiles {
			if t.Type != "folder" || !sameNode(t.NodeID, target.NodeID) {
				continue
			}
			if want := h == hash && key == body.TileID; t.Bound != want {
				t.Bound = want
				doc.Tiles[key] = t
				changed = true
			}
		}
		if changed {
			upd := saveLockedTree(c, rec, &doc)
			saved.add(&upd)
		}
	}
	AfterCommit(c, func() {
		for _, t := range saved {
			r.publishTiles(userID, t, "")
		}
	})
	return c.JSON(http.StatusOK, map[string]any{})
}
