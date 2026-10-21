package router

import (
	"errors"
	"maps"
	"net/http"
	"strings"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// 작업 공간 목록·만들기·폴더 이동·이름·삭제 — 타일 이동은 한 트랜잭션에서 두 트리를 함께 고친다(WS2(변경 방식), REF-workspace-impl.md)

func (r *supervisorRouter) mountWorkspaces(e *echo.Echo) {
	g := e.Group("/workspaces")
	g.GET("", r.listWorkspaces)
	g.POST("", r.createWorkspace)
	g.PATCH("/:hash", r.renameWorkspace)
	g.DELETE("/:hash", r.deleteWorkspace)
	g.POST("/:hash/move", r.moveWorkspaceFolder)
}

type workspaceCreateRequest struct {
	FolderTileID string `json:"folderTileId" validate:"required"`
	Name         string `json:"name" validate:"max=64"`
	Version      int64  `json:"version" validate:"required"`
}

type workspaceMoveRequest struct {
	FolderTileID string `json:"folderTileId" validate:"required"`
	To           string `json:"to" validate:"required"`
	Version      int64  `json:"version" validate:"required"`
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

// detachFolder: 폴더 타일과 그 하위를 d에서 떼어 새 문서로 돌려준다. 떼어낸 폴더 타일이 새 문서의 루트
func (d *tileDoc) detachFolder(id string) tileDoc {
	t, ok := d.Tiles[id]
	if !ok {
		panic(web.Err(404, "타일을 찾을 수 없습니다"))
	}
	if t.Type != "folder" {
		panic(web.Err(400, "폴더 타일만 옮길 수 있습니다"))
	}
	if id == d.RootID {
		panic(web.Err(400, "보드의 루트 폴더 타일은 옮길 수 없습니다"))
	}
	parent := d.Tiles[*t.Parent]
	kids := make([]string, 0, len(parent.Kids))
	for _, k := range parent.Kids {
		if k != id {
			kids = append(kids, k)
		}
	}
	parent.Kids = kids
	d.Tiles[parent.ID] = parent

	out := tileDoc{RootID: id, Tiles: map[string]tileItem{}}
	stack := []string{id}
	for len(stack) > 0 {
		k := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		out.Tiles[k] = d.Tiles[k]
		stack = append(stack, d.Tiles[k].Kids...)
		delete(d.Tiles, k)
	}
	t.Parent = nil
	out.Tiles[id] = t
	return out
}

// attachUnderRoot: sub 문서 전체를 d의 루트 끝에 붙인다(sub의 루트가 d 루트의 마지막 자식)
func (d *tileDoc) attachUnderRoot(sub tileDoc) {
	maps.Copy(d.Tiles, sub.Tiles)
	rootID := d.RootID
	top := d.Tiles[sub.RootID]
	top.Parent = &rootID
	d.Tiles[top.ID] = top
	root := d.Tiles[rootID]
	root.Kids = append(root.Kids, sub.RootID)
	d.Tiles[rootID] = root
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

// createWorkspace: 기본 작업 공간의 폴더 타일 하나를 하위째 떼어 새 작업 공간의 루트로 — W1(분리 범위), `/`에서만 만듦
func (r *supervisorRouter) createWorkspace(c echo.Context) error {
	sess := r.requireSession(c)
	body := bindBody[workspaceCreateRequest](c)
	userID := sess.Data.ID
	ctx := c.Request().Context()
	q := TxQueries(c)

	// 행 하나만 잠그므로 잠금 순서 규칙 대상 아님
	ensureTileTree(c, userID)
	src, err := q.GetTileTreeForUpdate(ctx, userID)
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

// moveWorkspaceFolder: 폴더 타일을 하위째 다른 작업 공간 루트 끝으로 — W3(이동 범위). 도착 쪽은 version 검사 없음(WS2(변경 방식))
func (r *supervisorRouter) moveWorkspaceFolder(c echo.Context) error {
	sess := r.requireSession(c)
	body := bindBody[workspaceMoveRequest](c)
	userID := sess.Data.ID

	locked := lockWorkspaces(c, userID)
	src := mustWorkspace(locked, c.Param("hash"))
	dst := mustWorkspace(locked, body.To)
	if src.Hash == dst.Hash {
		panic(web.Err(400, "같은 보드로는 옮길 수 없습니다"))
	}
	if src.Version != body.Version {
		return c.JSON(http.StatusConflict, tileTreeResponse{Hash: src.Hash, Tree: src.Tree, Version: src.Version})
	}
	srcDoc := mustParseDoc(src.Tree)
	dstDoc := mustParseDoc(dst.Tree)
	sub := srcDoc.detachFolder(body.FolderTileID)
	dstDoc.attachUnderRoot(sub)
	srcUpd := saveLockedTree(c, src, &srcDoc)
	dstUpd := saveLockedTree(c, dst, &dstDoc)

	AfterCommit(c, func() {
		r.publishTiles(userID, srcUpd, "")
		r.publishTiles(userID, dstUpd, "")
		r.resyncSizeOwners(sub.terminalUIDs())
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

// deleteWorkspace: 트리 전체(루트 폴더 타일째)를 기본 작업 공간 루트 끝으로 옮기고 삭제, kill 없음 — W2(삭제 처리)
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
	defDoc.attachUnderRoot(moved)
	defUpd := saveLockedTree(c, def, &defDoc)
	n, err := TxQueries(c).DeleteTileTree(c.Request().Context(), superdb.DeleteTileTreeParams{UserID: userID, Hash: target.Hash})
	if err != nil || n != 1 {
		panic(web.Err(500, "보드 삭제 실패: %v", err))
	}

	AfterCommit(c, func() {
		r.publishTiles(userID, defUpd, "")
		r.resyncSizeOwners(moved.terminalUIDs())
	})
	r.publishWorkspacesAfterCommit(c, userID)
	return c.JSON(http.StatusOK, workspaceChangeResponse{tileTreeResponse: tileTreeResponse{Hash: defUpd.Hash, Tree: defUpd.Tree, Version: defUpd.Version}})
}
