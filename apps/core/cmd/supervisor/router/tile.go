package router

import (
	"context"
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

// 계정당 타일 트리 문서 1개 — ⑪(타일 트리 서버 저장), REF-node-ui-save.md
func (r *supervisorRouter) mountTiles(e *echo.Echo) {
	g := e.Group("/tiles")
	g.GET("", r.getTiles)
	g.PUT("", r.putTiles, middleware.BodyLimit("256K"))
}

// 프론트 tileTree.store의 Tile 미러. 서버는 모양만 검사하고 순서·닫기 규칙은 프론트 몫
type tileSize struct {
	W float64 `json:"w"`
	H float64 `json:"h"`
}

type tileItem struct {
	ID      string   `json:"id"`
	Type    string   `json:"type"` // folder | terminal
	Parent  *string  `json:"parent"`
	Kids    []string `json:"kids"`
	Size    tileSize `json:"size"`
	NodeID  *int64   `json:"nodeId"` // folder: null = 루트 목록 / terminal: 필수
	UID     string   `json:"uid,omitempty"`
	Adopted bool     `json:"adopted,omitempty"`
}

type tileDoc struct {
	RootID string              `json:"rootId"`
	Tiles  map[string]tileItem `json:"tiles"`
}

type tileTreeResponse struct {
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
func (r *supervisorRouter) publishTiles(userID int64, tree []byte, version int64, tabID string) {
	ev := protocol.TilesUpdateEvent{Tree: tree, Version: version, TabID: tabID}
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
			if t.UID != "" || t.Adopted {
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
	rec, err = q.CreateTileTree(ctx, superdb.CreateTileTreeParams{UserID: userID, Tree: tree})
	if errors.Is(err, pgx.ErrNoRows) {
		return q.GetTileTree(ctx, userID)
	}
	return rec, err
}

func ensureTileTree(c echo.Context, userID int64) superdb.TileTree {
	rec, err := loadOrCreateTileTree(c.Request().Context(), TxQueries(c), userID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return rec
}

// appendTerminalTile: exec 성공 직후 터미널 타일을 서버 트리에 넣는다 — ⑪-4(타일 없는 실행 중 process).
// 요청 트랜잭션이 아니라 별도 트랜잭션으로 핸들러 안에서 커밋까지 끝낸다: 요청 트랜잭션은 응답을 보낸 뒤
// 커밋되므로 그 실패를 핸들러가 알 수 없고, 그러면 실패 시 process kill(보상)을 할 수 없다 (REF-node-ui-save-impl.md "3-b'(exec 연동) 구현")
func appendTerminalTile(ctx context.Context, userID int64, parentTileID string, size tileSize, nodeID int64, uid string) (tileItem, superdb.TileTree, error) {
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
		rec, err := q.GetTileTreeForUpdate(ctx, userID)
		if err != nil {
			return tileItem{}, superdb.TileTree{}, err
		}
		var doc tileDoc
		if err := json.Unmarshal(rec.Tree, &doc); err != nil {
			return tileItem{}, superdb.TileTree{}, err
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
		upd, err := q.UpdateTileTree(ctx, superdb.UpdateTileTreeParams{UserID: userID, Version: rec.Version, Tree: tree})
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

func (r *supervisorRouter) getTiles(c echo.Context) error {
	sess := r.requireSession(c)
	rec := ensureTileTree(c, sess.Data.ID)
	return c.JSON(http.StatusOK, tileTreeResponse{Tree: rec.Tree, Version: rec.Version})
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
	tree := doc.mustMarshal()
	rec, err := TxQueries(c).UpdateTileTree(c.Request().Context(), superdb.UpdateTileTreeParams{
		UserID: userID, Version: body.Version, Tree: tree,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		cur := ensureTileTree(c, userID)
		return c.JSON(http.StatusConflict, tileTreeResponse{Tree: cur.Tree, Version: cur.Version})
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	tabID := r.requestTab(c, sess.Name())
	AfterCommit(c, func() { r.publishTiles(userID, tree, rec.Version, tabID) })
	return c.JSON(http.StatusOK, tilePutResponse{Version: rec.Version})
}
