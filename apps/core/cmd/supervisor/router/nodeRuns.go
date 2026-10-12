package router

import (
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// 스크립트 행 "실행 중 N · 최소화 n"(실시간-6a): 폴더의 자식 스크립트별 살아 있는 process + 내 보드에서의 위치.

type runResponse struct {
	Uid         string `json:"uid"`
	NodeID      int64  `json:"nodeId"`
	Status      string `json:"status"`
	Permission  int16  `json:"permission"`
	StartedBy   string `json:"startedBy"`
	StartedByMe bool   `json:"startedByMe,omitempty"` // 내가 실행했으면 프론트가 닉네임 대신 "나"

	Hash      string `json:"hash,omitempty"` // 내 어느 보드에도 타일이 없으면 비어 있음
	TileID    string `json:"tileId,omitempty"`
	Minimized bool   `json:"minimized,omitempty"`
}

type runLocation struct {
	hash, tileID string
	minimized    bool
}

// listNodeRuns는 폴더(보기 이상, id 0 = 루트 목록)의 직접 자식 스크립트에서 실행 중인 process를 평평하게 돌려준다(프론트가 nodeId로 묶음).
func (r *supervisorRouter) listNodeRuns(c echo.Context) error {
	sess := r.requireSession(c)
	folderID := paramID(c)
	parent := pgtype.Int8{} // id 0 = 루트 목록
	if folderID != 0 {
		requireNodePerm(c, sess.Data.ID, folderID, superdb.PermView)
		parent = pgtype.Int8{Int64: folderID, Valid: true}
	}
	q, ctx := TxQueries(c), c.Request().Context()

	rows, err := q.ListVisibleLiveProcessesInFolder(ctx, superdb.ListVisibleLiveProcessesInFolderParams{UserID: sess.Data.ID, FolderID: parent})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := make([]runResponse, len(rows))
	if len(rows) == 0 {
		return c.JSON(200, out)
	}

	docs, err := q.ListTileTreeDocs(ctx, sess.Data.ID) // 잠그지 않는 읽기
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	where := map[string]runLocation{}
	for _, rec := range docs {
		doc := mustParseDoc(rec.Tree)
		for _, t := range doc.Tiles {
			if t.Type == "terminal" {
				where[t.UID] = runLocation{hash: rec.Hash, tileID: t.ID, minimized: t.Minimized}
			}
		}
	}
	for i, row := range rows {
		loc := where[row.Uid]
		out[i] = runResponse{Uid: row.Uid, NodeID: row.NodeID.Int64, Status: row.Status, Permission: row.Permission, StartedBy: row.StartedByNickname, StartedByMe: row.StartedBy == sess.Data.ID,
			Hash: loc.hash, TileID: loc.tileID, Minimized: loc.minimized}
	}
	return c.JSON(200, out)
}
