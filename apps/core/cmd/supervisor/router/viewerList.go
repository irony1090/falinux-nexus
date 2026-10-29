package router

import (
	"strconv"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 터미널 탭의 보기 줄 — 보기 타일은 process가 아니라 내 보드 트리에서 찾는다(MEMORY "process 범위", REF-view-tile.md)

type viewerResponse struct {
	Hash     string `json:"hash"`
	TileID   string `json:"tileId"`
	NodeID   int64  `json:"nodeId"`
	Instance string `json:"instance"`
	Path     string `json:"path"`
}

// GET /tiles/viewers?folderId= — 범위는 터미널 탭과 같게: 0 = 전부 / 폴더 = 연 폴더가 그 하위 / &mainKey= = 그 장비
func (r *supervisorRouter) listViewers(c echo.Context) error {
	sess := r.requireSession(c)
	folderID, err := strconv.ParseInt(c.QueryParam("folderId"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 folderId입니다"))
	}
	mainKey := c.QueryParam("mainKey")
	if mainKey != "" {
		requireDeviceFolder(c, sess.Data.ID, folderID, mainKey)
	} else if folderID != 0 {
		requireNodePerm(c, sess.Data.ID, folderID, superdb.PermView)
	}
	q, ctx := TxQueries(c), c.Request().Context()
	docs, err := q.ListTileTreeDocs(ctx, sess.Data.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := []viewerResponse{}
	for _, rec := range docs {
		doc := mustParseDoc(rec.Tree)
		// order 순 = 보드 안 화면 순서
		for _, id := range doc.Order {
			t := doc.Tiles[id]
			if t.Type != tileViewer || t.NodeID == nil {
				continue
			}
			if mainKey != "" {
				if main, _ := splitInstanceKey(t.Instance); main != mainKey {
					continue
				}
			} else if folderID != 0 {
				in, err := q.IsDescendant(ctx, superdb.IsDescendantParams{AncestorID: folderID, NodeID: *t.NodeID})
				if err != nil || !in {
					continue
				}
			}
			out = append(out, viewerResponse{Hash: rec.Hash, TileID: t.ID, NodeID: *t.NodeID, Instance: t.Instance, Path: t.Path})
		}
	}
	return c.JSON(200, out)
}
