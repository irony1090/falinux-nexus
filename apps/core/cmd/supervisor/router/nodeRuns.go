package router

import (
	"strconv"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// 노드의 상태 리스트·터미널 탭(실시간-6a, 개편-3(목록 API)): 살아 있는 process + 내 보드에 타일이 있는 끝난 process, 내 보드에서의 위치.

type runResponse struct {
	Uid         string `json:"uid"`
	NodeID      *int64 `json:"nodeId"`        // null = 장비 파일 실행(폴더-4(실행))
	Name        string `json:"name"`          // 스크립트 노드 이름 — 단독 공유 process는 노드를 못 봐도 이걸로 표시. 장비 실행 = 파일 이름
	Cwd         string `json:"cwd,omitempty"` // 장비 실행만 — 경로 표시용
	Status      string `json:"status"`
	Permission  int16  `json:"permission"`
	StartedBy   string `json:"startedBy"`
	StartedByMe bool   `json:"startedByMe,omitempty"` // 내가 실행했으면 프론트가 닉네임 대신 "나"
	Terminating bool   `json:"terminating,omitempty"` // 종료 요청됨 — processResponse와 같은 뜻
	Forced      bool   `json:"forced,omitempty"`

	Hash      string `json:"hash,omitempty"` // 내 어느 보드에도 타일이 없으면 비어 있음
	TileID    string `json:"tileId,omitempty"`
	Minimized bool   `json:"minimized,omitempty"`
}

type runLocation struct {
	hash, tileID string
	minimized    bool
}

// listNodeRuns: 폴더(보기 이상, id 0 = 루트 목록)의 직접 자식 스크립트 process를 평평하게 돌려준다(프론트가 nodeId로 묶음).
// ?deep=1 = 그 폴더 하위 전체(터미널 탭), id 0이면 볼 수 있는 process 전부
func (r *supervisorRouter) listNodeRuns(c echo.Context) error {
	sess := r.requireSession(c)
	folderID := paramID(c)
	parent := pgtype.Int8{} // id 0 = 루트 목록
	if folderID != 0 {
		requireNodePerm(c, sess.Data.ID, folderID, superdb.PermView)
		parent = pgtype.Int8{Int64: folderID, Valid: true}
	}
	q, ctx := TxQueries(c), c.Request().Context()
	where, uids := runLocations(c, sess.Data.ID)

	var rows []superdb.ListVisibleRunsInFolderRow
	var err error
	if c.QueryParam("deep") == "1" {
		sub, err := q.ListVisibleRunsInSubtree(ctx, superdb.ListVisibleRunsInSubtreeParams{UserID: sess.Data.ID, FolderID: parent, TileUids: uids})
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		rows = make([]superdb.ListVisibleRunsInFolderRow, len(sub))
		for i, row := range sub {
			rows[i] = superdb.ListVisibleRunsInFolderRow(row)
		}
	} else {
		rows, err = q.ListVisibleRunsInFolder(ctx, superdb.ListVisibleRunsInFolderParams{UserID: sess.Data.ID, FolderID: parent, TileUids: uids})
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
	}
	return c.JSON(200, toRunResponses(rows, where, sess.Data.ID))
}

// listDeviceRuns: 루트 장비 타일 터미널 탭 = 그 장비(main_key)의 볼 수 있는 process 전부 — 이동-3(장비 목록 API), REF-worker-dir-board.md.
// ?folderId= 연 폴더(장비 디렉토리와 같은 인가, 인스턴스 접속 여부는 안 봄)
func (r *supervisorRouter) listDeviceRuns(c echo.Context) error {
	sess := r.requireSession(c)
	mainKey := pathParam(c, "mainKey")
	folderID, err := strconv.ParseInt(c.QueryParam("folderId"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 folderId입니다"))
	}
	requireDeviceFolder(c, sess.Data.ID, folderID, mainKey)
	where, uids := runLocations(c, sess.Data.ID)

	sub, err := TxQueries(c).ListVisibleRunsOnDevice(c.Request().Context(), superdb.ListVisibleRunsOnDeviceParams{UserID: sess.Data.ID, MainKey: mainKey, TileUids: uids})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	rows := make([]superdb.ListVisibleRunsInFolderRow, len(sub))
	for i, row := range sub {
		rows[i] = superdb.ListVisibleRunsInFolderRow(row)
	}
	return c.JSON(200, toRunResponses(rows, where, sess.Data.ID))
}

// runLocations: 내 보드들의 process 타일 위치(uid -> 보드·타일) + 그 uid 목록(끝난 process를 목록에 남기는 근거)
func runLocations(c echo.Context, userID int64) (map[string]runLocation, []string) {
	docs, err := TxQueries(c).ListTileTreeDocs(c.Request().Context(), userID) // 잠그지 않는 읽기
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	where := map[string]runLocation{}
	uids := []string{}
	for _, rec := range docs {
		doc := mustParseDoc(rec.Tree)
		for _, t := range doc.Tiles {
			if t.holdsProc() {
				where[t.UID] = runLocation{hash: rec.Hash, tileID: t.ID, minimized: t.Minimized}
				uids = append(uids, t.UID)
			}
		}
	}
	return where, uids
}

func toRunResponses(rows []superdb.ListVisibleRunsInFolderRow, where map[string]runLocation, userID int64) []runResponse {
	out := make([]runResponse, len(rows))
	for i, row := range rows {
		loc := where[row.Uid]
		name, cwd := row.Name, ""
		if !row.NodeID.Valid {
			name, cwd = fileBase(row.Cmd), row.Cwd
		}
		out[i] = runResponse{Uid: row.Uid, NodeID: nullID(row.NodeID), Name: name, Cwd: cwd, Status: row.Status, Permission: row.Permission,
			StartedBy: row.StartedByNickname, StartedByMe: row.StartedBy == userID,
			Terminating: row.Terminating, Forced: row.KillForced,
			Hash: loc.hash, TileID: loc.tileID, Minimized: loc.minimized}
	}
	return out
}
