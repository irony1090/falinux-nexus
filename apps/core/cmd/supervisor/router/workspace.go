package router

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// 작업 공간 목록·만들기·폴더 이동·이름·삭제 — 타일 이동은 한 트랜잭션에서 두 트리를 함께 고친다(WS2(변경 방식), REF-workspace-impl.md)

func (r *supervisorRouter) mountWorkspaces(e *echo.Group) {
	g := e.Group("/workspaces")
	g.GET("", r.listWorkspaces)
	g.POST("", r.createWorkspace)
	g.PATCH("/:hash", r.renameWorkspace)
	g.DELETE("/:hash", r.deleteWorkspace)
	g.POST("/:hash/tiles/move", r.moveWorkspaceTile)
}

type workspaceCreateRequest struct {
	From         string `json:"from"` // 출발 보드 hash, 비면 기본 보드
	FolderTileID string `json:"folderTileId" validate:"required"`
	Name         string `json:"name" validate:"max=64"`
	Version      int64  `json:"version" validate:"required"`
}

type workspaceTileMoveRequest struct {
	TileID  string `json:"tileId" validate:"required"`
	To      string `json:"to" validate:"required"`
	After   string `json:"after"`   // 도착 보드에서 이 타일 바로 뒤. "" = 루트 바로 뒤, 없는 타일 = 맨 뒤
	Version *int64 `json:"version"` // 출발 보드 version — 보고 있는 보드일 때만 보냄, 없으면 검사 안 함(전체보기의 다른 보드 줄)
}

type workspaceRenameRequest struct {
	Name string `json:"name" validate:"max=64"`
}

// workspaceChangeResponse: 바뀐 작업 공간 + 요청한 탭이 보던 트리(만들기·이동 = 출발, 삭제 = 기본)
type workspaceChangeResponse struct {
	Workspace *protocol.WorkspaceItem `json:"workspace,omitempty"`
	tileTreeResponse
}

func toWorkspaceItem(rec superdb.TileTree) protocol.WorkspaceItem {
	return protocol.WorkspaceItem{Hash: rec.Hash, Name: rec.Name, IsDefault: rec.IsDefault, Ord: rec.Ord}
}

func bindBody[T any](c echo.Context) T {
	var body T
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	return body
}

// detachFolder: 루트가 될 수 있는 타일(능력-1(루트 가능))과 그 하위를 d에서 떼어 새 문서로 돌려준다. 떼어낸 타일이 새 문서의 루트
func (d *tileDoc) detachFolder(id string) tileDoc {
	t, ok := d.Tiles[id]
	if !ok {
		panic(web.Err(404, "타일을 찾을 수 없습니다"))
	}
	if !t.canRoot() {
		panic(web.Err(400, "보드 루트가 될 수 없는 타일은 옮길 수 없습니다"))
	}
	if id == d.RootID {
		panic(web.Err(400, "보드의 루트 폴더 타일은 옮길 수 없습니다"))
	}
	// 떼어낸 타일이 새 루트라 맨 앞, 나머지는 옛 보드에서의 상대 순서
	out := tileDoc{RootID: id, Order: []string{id}, Tiles: map[string]tileItem{}}
	for _, k := range d.subtree(id) {
		out.Tiles[k] = d.Tiles[k]
		if k != id {
			out.Order = append(out.Order, k)
		}
		delete(d.Tiles, k)
	}
	d.Order = slices.DeleteFunc(d.Order, func(k string) bool { _, moved := out.Tiles[k]; return moved })
	t.Parent = nil
	out.Tiles[id] = t
	return out
}

// placeAfter: ids(order에 없는 상태)를 그 순서대로 after 바로 뒤에. "" = 루트 바로 뒤, 없는 타일 = 맨 뒤(낡은 화면에서 온 요청)
func (d *tileDoc) placeAfter(after string, ids []string) {
	at := len(d.Order)
	if after == "" {
		after = d.RootID
	}
	if i := slices.Index(d.Order, after); i >= 0 {
		at = i + 1
	}
	d.Order = slices.Insert(d.Order, at, ids...)
}

// takeOpened: id와 그 타일이 연 타일 전부를 떼어 order 순으로 — 보드 간 이동 묶음(보드-1, REF-tile-drag.md). 루트는 호출부가 막는다
func (d *tileDoc) takeOpened(id string) []tileItem {
	ids := d.subtree(id)
	out := make([]tileItem, 0, len(ids))
	for _, k := range ids {
		out = append(out, d.Tiles[k])
		delete(d.Tiles, k)
	}
	d.Order = slices.DeleteFunc(d.Order, func(k string) bool { _, ok := d.Tiles[k]; return !ok })
	return out
}

// putOpened: takeOpened 묶음을 after 바로 뒤에. 맨 위 타일(연 곳이 묶음 밖)의 연 곳 = 이 보드 루트(parent 없음 = 루트뿐).
// 맨 위가 묶음 첫 타일이라는 보장은 없다 — order는 연 곳과 무관해서 연 터미널이 폴더보다 앞일 수 있음(10-29 실측)
func (d *tileDoc) putOpened(after string, items []tileItem) {
	rootID := d.RootID
	in := make(map[string]bool, len(items))
	for _, t := range items {
		in[t.ID] = true
	}
	ids := make([]string, 0, len(items))
	for _, t := range items {
		if t.Parent == nil || !in[*t.Parent] {
			t.Parent = &rootID
		}
		d.Tiles[t.ID] = t
		ids = append(ids, t.ID)
	}
	d.placeAfter(after, ids)
}

// lockWorkspaces: 계정의 트리 전부를 잠그고 hash로 찾는다. 여러 트리를 고치는 동작은 이것만 쓴다(ListTileTreesForUpdate 주석의 잠금 순서)
func lockWorkspaces(c echo.Context, userID int64) map[string]superdb.TileTree {
	recs, err := TxQueries(c).ListTileTreesForUpdate(c.Request().Context(), userID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := make(map[string]superdb.TileTree, len(recs))
	for _, rec := range recs {
		out[rec.Hash] = rec
	}
	return out
}

func mustWorkspace(locked map[string]superdb.TileTree, hash string) superdb.TileTree {
	rec, ok := locked[hash]
	if !ok {
		panic(web.Err(404, "보드를 찾을 수 없습니다"))
	}
	return rec
}

// saveLockedTree: 잠근 행이라 version이 바뀌었을 리 없다 — 행 없음이면 500
func saveLockedTree(c echo.Context, rec superdb.TileTree, doc *tileDoc) superdb.TileTree {
	if err := doc.validate(); err != nil {
		panic(web.Err(500, "타일 트리 변경 결과가 올바르지 않습니다: %v", err))
	}
	upd, err := TxQueries(c).UpdateTileTreeByHash(c.Request().Context(), superdb.UpdateTileTreeByHashParams{
		UserID: rec.UserID, Hash: rec.Hash, Version: rec.Version, Tree: doc.mustMarshal(),
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return upd
}

// publishWorkspacesAfterCommit: 커밋 전에 목록을 읽어 두고 커밋 뒤에 보낸다
func (r *supervisorRouter) publishWorkspacesAfterCommit(c echo.Context, userID int64) {
	recs, err := TxQueries(c).ListTileTrees(c.Request().Context(), userID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	ev := protocol.WorkspacesUpdateEvent{Workspaces: make([]protocol.WorkspaceItem, 0, len(recs))}
	for _, rec := range recs {
		ev.Workspaces = append(ev.Workspaces, protocol.WorkspaceItem{Hash: rec.Hash, Name: rec.Name, IsDefault: rec.IsDefault, Ord: rec.Ord})
	}
	AfterCommit(c, func() { r.publishUser(userID, protocol.MsgWorkspacesUpdate, ev) })
}

// listWorkspaces: 기본이 없으면 만든다
func (r *supervisorRouter) listWorkspaces(c echo.Context) error {
	sess := r.requireSession(c)
	ensureTileTree(c, sess.Data.ID)
	recs, err := TxQueries(c).ListTileTrees(c.Request().Context(), sess.Data.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := make([]protocol.WorkspaceItem, 0, len(recs))
	for _, rec := range recs {
		out = append(out, protocol.WorkspaceItem{Hash: rec.Hash, Name: rec.Name, IsDefault: rec.IsDefault, Ord: rec.Ord})
	}
	return c.JSON(http.StatusOK, out)
}

// createWorkspace: 출발 보드의 폴더 타일 하나를 하위째 떼어 새 작업 공간의 루트로 — W1(분리 범위), 어느 보드에서든(2026-10-22)
func (r *supervisorRouter) createWorkspace(c echo.Context) error {
	sess := r.requireSession(c)
	body := bindBody[workspaceCreateRequest](c)
	userID := sess.Data.ID
	ctx := c.Request().Context()
	q := TxQueries(c)

	// 행 하나만 잠그므로 잠금 순서 규칙 대상 아님
	var src superdb.TileTree
	var err error
	if body.From == "" {
		ensureTileTree(c, userID)
		src, err = q.GetTileTreeForUpdate(ctx, userID)
	} else {
		src, err = q.GetTileTreeByHashForUpdate(ctx, superdb.GetTileTreeByHashForUpdateParams{UserID: userID, Hash: body.From})
		if errors.Is(err, pgx.ErrNoRows) {
			panic(web.Err(404, "보드를 찾을 수 없습니다"))
		}
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if src.Version != body.Version {
		return c.JSON(http.StatusConflict, tileTreeResponse{Hash: src.Hash, Tree: src.Tree, Version: src.Version})
	}
	doc := mustParseDoc(src.Tree)
	sub := doc.detachFolder(body.FolderTileID)
	if err := sub.validate(); err != nil {
		panic(web.Err(500, "떼어낸 타일 트리가 올바르지 않습니다: %v", err))
	}
	srcUpd := saveLockedTree(c, src, &doc)
	ws, err := q.CreateWorkspaceTree(ctx, superdb.CreateWorkspaceTreeParams{
		UserID: userID, Hash: newTreeHash(), Name: strings.TrimSpace(body.Name), Tree: sub.mustMarshal(),
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}

	// 요청한 탭도 이벤트로 받도록 tabID를 비운다 — 서버가 만든 변경이라 그 탭에도 없는 내용
	AfterCommit(c, func() {
		r.publishTiles(userID, srcUpd, "")
		r.publishTiles(userID, ws, "")
		r.resyncSizeOwners(sub.terminalUIDs())
	})
	r.publishWorkspacesAfterCommit(c, userID)
	item := toWorkspaceItem(ws)
	return c.JSON(http.StatusOK, workspaceChangeResponse{Workspace: &item, tileTreeResponse: tileTreeResponse{Hash: srcUpd.Hash, Tree: srcUpd.Tree, Version: srcUpd.Version}})
}

// moveWorkspaceTile: 타일을 도착 보드의 after 뒤로 — 보드-1(보드 간 이동 API). 같은 보드 = 그 타일 하나(연 곳 그대로),
// 다른 보드 = 그 타일이 연 타일까지 묶음째(가(같은 보드 안끼리 통일), 2026-10-29). 도착 쪽은 version 검사 없음(WS2(변경 방식))
func (r *supervisorRouter) moveWorkspaceTile(c echo.Context) error {
	sess := r.requireSession(c)
	body := bindBody[workspaceTileMoveRequest](c)
	userID := sess.Data.ID

	locked := lockWorkspaces(c, userID)
	src := mustWorkspace(locked, c.Param("hash"))
	dst := mustWorkspace(locked, body.To)
	if body.Version != nil && src.Version != *body.Version {
		return c.JSON(http.StatusConflict, tileTreeResponse{Hash: src.Hash, Tree: src.Tree, Version: src.Version})
	}
	srcDoc := mustParseDoc(src.Tree)
	if _, ok := srcDoc.Tiles[body.TileID]; !ok {
		panic(web.Err(404, "타일을 찾을 수 없습니다"))
	}
	if body.TileID == srcDoc.RootID {
		panic(web.Err(400, "보드의 루트 타일은 옮길 수 없습니다"))
	}

	if src.Hash == dst.Hash {
		if body.After == body.TileID {
			panic(web.Err(400, "자기 뒤로는 옮길 수 없습니다"))
		}
		srcDoc.Order = slices.DeleteFunc(srcDoc.Order, func(k string) bool { return k == body.TileID })
		srcDoc.placeAfter(body.After, []string{body.TileID})
		srcUpd := saveLockedTree(c, src, &srcDoc)
		AfterCommit(c, func() { r.publishTiles(userID, srcUpd, "") })
		return c.JSON(http.StatusOK, workspaceChangeResponse{tileTreeResponse: tileTreeResponse{Hash: srcUpd.Hash, Tree: srcUpd.Tree, Version: srcUpd.Version}})
	}

	dstDoc := mustParseDoc(dst.Tree)
	moved := srcDoc.takeOpened(body.TileID)
	dstDoc.putOpened(body.After, moved)
	srcUpd := saveLockedTree(c, src, &srcDoc)
	dstUpd := saveLockedTree(c, dst, &dstDoc)

	var uids []string
	for _, t := range moved {
		if uid := t.procUID(); uid != "" {
			uids = append(uids, uid)
		}
	}
	AfterCommit(c, func() {
		r.publishTiles(userID, srcUpd, "")
		r.publishTiles(userID, dstUpd, "")
		r.resyncSizeOwners(uids)
	})
	return c.JSON(http.StatusOK, workspaceChangeResponse{tileTreeResponse: tileTreeResponse{Hash: srcUpd.Hash, Tree: srcUpd.Tree, Version: srcUpd.Version}})
}

func (r *supervisorRouter) renameWorkspace(c echo.Context) error {
	sess := r.requireSession(c)
	body := bindBody[workspaceRenameRequest](c)
	rec, err := TxQueries(c).RenameTileTree(c.Request().Context(), superdb.RenameTileTreeParams{
		UserID: sess.Data.ID, Hash: c.Param("hash"), Name: strings.TrimSpace(body.Name),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "보드를 찾을 수 없습니다"))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	r.publishWorkspacesAfterCommit(c, sess.Data.ID)
	item := toWorkspaceItem(rec)
	return c.JSON(http.StatusOK, workspaceChangeResponse{Workspace: &item})
}

// deleteWorkspace: 트리 전체(루트 폴더 타일째)를 기본 작업 공간 루트 바로 뒤로 옮기고 삭제, kill 없음 — W2(삭제 처리)
func (r *supervisorRouter) deleteWorkspace(c echo.Context) error {
	sess := r.requireSession(c)
	userID := sess.Data.ID

	locked := lockWorkspaces(c, userID)
	target := mustWorkspace(locked, c.Param("hash"))
	if target.IsDefault {
		panic(web.Err(400, "기본 보드는 삭제할 수 없습니다"))
	}
	var def superdb.TileTree
	for _, rec := range locked {
		if rec.IsDefault {
			def = rec
		}
	}
	if def.Hash == "" {
		panic(web.Err(500, "기본 보드가 없습니다"))
	}
	defDoc := mustParseDoc(def.Tree)
	moved := mustParseDoc(target.Tree)
	uids := moved.terminalUIDs()
	defDoc.putOpened("", moved.takeOpened(moved.RootID))
	defUpd := saveLockedTree(c, def, &defDoc)
	n, err := TxQueries(c).DeleteTileTree(c.Request().Context(), superdb.DeleteTileTreeParams{UserID: userID, Hash: target.Hash})
	if err != nil || n != 1 {
		panic(web.Err(500, "보드 삭제 실패: %v", err))
	}

	AfterCommit(c, func() {
		r.publishTiles(userID, defUpd, "")
		r.resyncSizeOwners(uids)
	})
	r.publishWorkspacesAfterCommit(c, userID)
	return c.JSON(http.StatusOK, workspaceChangeResponse{tileTreeResponse: tileTreeResponse{Hash: defUpd.Hash, Tree: defUpd.Tree, Version: defUpd.Version}})
}
