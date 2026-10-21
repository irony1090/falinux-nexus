package router

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

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
	r.publishUser(userID, protocol.MsgTilesUpdate, ev)
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
