package router

import (
	"errors"
	"math"
	"strings"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// 영상 보기 타일 — 같은 파일(instance+path)은 계정 전체에서 하나라 서버만 넣는다(MEMORY "타일 추가 주체", REF-view-tile.md)

type videoTileRequest struct {
	FolderID     int64     `json:"folderId" validate:"required"` // 장비 타일을 연 폴더(인가·보기 타일 nodeId)
	Instance     string    `json:"instance" validate:"required"`
	Path         string    `json:"path" validate:"required"`
	Workspace    string    `json:"workspace"`
	ParentTileID string    `json:"parentTileId"` // 연 장비 타일. 없거나 닫혔으면 루트
	Size         *tileSize `json:"size"`
}

// Tile = 새로 넣었을 때만. 없으면 이미 있던 타일(Hash 보드의 TileID)로 이동
type videoTileResponse struct {
	Hash        string    `json:"hash"`
	TileID      string    `json:"tileId"`
	Tile        *tileItem `json:"tile,omitempty"`
	TileVersion int64     `json:"tileVersion,omitempty"`
}

// POST /tiles/viewers/video — 계정 트리 전부 잠금 -> 같은 파일이 있으면 그것, 없으면 요청한 보드의 장비 타일 아래에 넣음
func (r *supervisorRouter) openVideoTile(c echo.Context) error {
	sess := r.requireSession(c)
	userID := sess.Data.ID
	body := bindBody[videoTileRequest](c)
	if !isVideoPath(body.Path) {
		panic(web.Err(415, "영상 형식이 아닙니다"))
	}
	if !strings.Contains(body.Instance, "#") {
		panic(web.Err(400, "잘못된 instance입니다"))
	}
	mainKey, _ := splitInstanceKey(body.Instance)
	requireDeviceFolder(c, userID, body.FolderID, mainKey)
	size := execTileSize(c, userID, body.Workspace, body.Size)

	ensureTileTree(c, userID)
	locked := lockWorkspaces(c, userID)
	var target superdb.TileTree
	for _, rec := range locked {
		doc := mustParseDoc(rec.Tree)
		for _, t := range doc.Tiles {
			if t.Type == tileViewer && t.Instance == body.Instance && t.Path == body.Path {
				return c.JSON(200, videoTileResponse{Hash: rec.Hash, TileID: t.ID})
			}
		}
		// 요청한 보드가 그 사이 삭제됐으면 기본으로(appendTerminalTile과 같음)
		if rec.Hash == body.Workspace || (target.Hash == "" && rec.IsDefault) {
			target = rec
		}
	}

	doc := mustParseDoc(target.Tree)
	parent, ok := doc.Tiles[body.ParentTileID]
	if !ok || !parent.canParent() {
		parent = doc.Tiles[doc.RootID]
	}
	tile := newViewerTile(&parent.ID, body.FolderID, body.Instance, body.Path, size)
	parent.Kids = append(parent.Kids, tile.ID)
	doc.Tiles[parent.ID] = parent
	doc.Tiles[tile.ID] = tile
	upd := saveLockedTree(c, target, &doc)

	tabID := r.requestTab(c, sess.Name())
	AfterCommit(c, func() { r.publishTiles(userID, upd, tabID) })
	return c.JSON(200, videoTileResponse{Hash: upd.Hash, TileID: tile.ID, Tile: &tile, TileVersion: upd.Version})
}

// ---- S1(이어보기): 재생 위치 — 키 = 파일(instance+path), 계정마다. 인가 = 내 행만 읽고 쓰므로 세션만 ----

type videoPosResponse struct {
	Pos *float64 `json:"pos"` // null = 저장된 위치 없음
}

type videoPosRequest struct {
	Instance string  `json:"instance" validate:"required"`
	Path     string  `json:"path" validate:"required"`
	Pos      float64 `json:"pos"`
}

// GET /tiles/viewers/video/position?instance&path
func (r *supervisorRouter) getVideoPosition(c echo.Context) error {
	sess := r.requireSession(c)
	pos, err := TxQueries(c).GetVideoPosition(c.Request().Context(), superdb.GetVideoPositionParams{
		UserID: sess.Data.ID, Instance: c.QueryParam("instance"), Path: c.QueryParam("path"),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return c.JSON(200, videoPosResponse{})
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return c.JSON(200, videoPosResponse{Pos: &pos})
}

// PUT /tiles/viewers/video/position — 그 파일의 보기 타일이 없으면 저장 안 함(204, 닫힌 뒤 늦게 온 저장)
func (r *supervisorRouter) saveVideoPosition(c echo.Context) error {
	sess := r.requireSession(c)
	body := bindBody[videoPosRequest](c)
	if !isVideoPath(body.Path) {
		panic(web.Err(415, "영상 형식이 아닙니다"))
	}
	if body.Pos < 0 || math.IsNaN(body.Pos) || math.IsInf(body.Pos, 0) {
		panic(web.Err(400, "잘못된 위치입니다"))
	}
	n, err := TxQueries(c).SaveVideoPosition(c.Request().Context(), superdb.SaveVideoPositionParams{
		UserID: sess.Data.ID, Instance: body.Instance, Path: body.Path, Pos: body.Pos,
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if n == 0 {
		return c.NoContent(204)
	}
	return c.NoContent(200)
}
