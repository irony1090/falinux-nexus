package router

import (
	"context"
	"errors"
	"strconv"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/transport"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// worker Err 코드 -> HTTP 상태
var dirErrStatus = map[string]int{
	protocol.DirErrNotFound: 404,
	protocol.DirErrDenied:   403,
	protocol.DirErrNotDir:   400,
	protocol.DirErrBadPath:  400,
	protocol.DirErrExpired:  410,
	protocol.DirErrIsDir:    400,
}

// listWorkerDir: worker 디렉토리 한 페이지 — 폴더-1(목록 요청), REF-worker-dir.md
// ?folderId=(연 폴더) &path= 첫 페이지 / &cursor=&offset= 다음 페이지
func (r *supervisorRouter) listWorkerDir(c echo.Context) error {
	sess := r.requireSession(c)
	mainKey, subKey := pathParam(c, "mainKey"), pathParam(c, "subKey")
	folderID, err := strconv.ParseInt(c.QueryParam("folderId"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 folderId입니다"))
	}
	offset := 0
	if s := c.QueryParam("offset"); s != "" {
		if offset, err = strconv.Atoi(s); err != nil {
			panic(web.Err(400, "잘못된 offset입니다"))
		}
	}

	conn := r.requireDeviceDir(c, sess.Data.ID, folderID, mainKey, subKey)
	EndTx(c) // 인가 끝 — worker 응답(최대 10초)을 기다리는 동안 DB 연결을 잡지 않게

	ctx, cancel := context.WithTimeout(c.Request().Context(), sendCallTimeout)
	defer cancel()
	res, err := conn.Call(ctx, protocol.MsgDirList, protocol.DirListRequest{
		Path:   c.QueryParam("path"),
		Cursor: c.QueryParam("cursor"),
		Offset: offset,
	})
	if errors.Is(err, context.DeadlineExceeded) {
		panic(web.Err(504, "worker 응답이 없습니다"))
	}
	if err != nil {
		panic(web.Err(502, "%v", err))
	}
	var body protocol.DirListResponse
	if err := res.Bind(&body); err != nil {
		panic(web.Err(502, "%v", err))
	}
	if body.Err != "" {
		status, ok := dirErrStatus[body.Err]
		if !ok {
			status = 500
		}
		panic(web.Err(status, "%s", body.Reason))
	}
	return c.JSON(200, body)
}

// suggestWorkerDir: 경로 자동완성 후보 — 경로-2(후보 REST), REF-path-input.md. ?folderId &dir &prefix &max(0 = 폴더인지만)
func (r *supervisorRouter) suggestWorkerDir(c echo.Context) error {
	sess := r.requireSession(c)
	mainKey, subKey := pathParam(c, "mainKey"), pathParam(c, "subKey")
	folderID, err := strconv.ParseInt(c.QueryParam("folderId"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 folderId입니다"))
	}
	max := 0
	if s := c.QueryParam("max"); s != "" {
		if max, err = strconv.Atoi(s); err != nil {
			panic(web.Err(400, "잘못된 max입니다"))
		}
	}
	conn := r.requireDeviceDir(c, sess.Data.ID, folderID, mainKey, subKey)
	EndTx(c)

	ctx, cancel := context.WithTimeout(c.Request().Context(), sendCallTimeout)
	defer cancel()
	res, err := conn.Call(ctx, protocol.MsgDirSuggest, protocol.DirSuggestRequest{
		Dir:    c.QueryParam("dir"),
		Prefix: c.QueryParam("prefix"),
		Max:    max,
	})
	if errors.Is(err, context.DeadlineExceeded) {
		panic(web.Err(504, "worker 응답이 없습니다"))
	}
	if err != nil {
		panic(web.Err(502, "%v", err)) // 옛 worker(DIR_SUGGEST 모름)도 여기
	}
	var out protocol.DirSuggestResponse
	if err := res.Bind(&out); err != nil {
		panic(web.Err(502, "%v", err))
	}
	if out.Err != "" {
		status, ok := dirErrStatus[out.Err]
		if !ok {
			status = 500
		}
		panic(web.Err(status, "%s", out.Reason))
	}
	return c.JSON(200, out)
}

type dirExistsRequest struct {
	FolderID int64    `json:"folderId" validate:"required"`
	Path     string   `json:"path" validate:"required"`
	Names    []string `json:"names" validate:"required"`
}

// workerDirExists: 받을 폴더에 이미 있는 이름 — Q28(여러 파일 겹침) 확인 창의 근거
func (r *supervisorRouter) workerDirExists(c echo.Context) error {
	sess := r.requireSession(c)
	mainKey, subKey := pathParam(c, "mainKey"), pathParam(c, "subKey")
	var body dirExistsRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	conn := r.requireDeviceDir(c, sess.Data.ID, body.FolderID, mainKey, subKey)
	EndTx(c)

	ctx, cancel := context.WithTimeout(c.Request().Context(), sendCallTimeout)
	defer cancel()
	res, err := conn.Call(ctx, protocol.MsgDirExists, protocol.DirExistsRequest{Path: body.Path, Names: body.Names})
	if errors.Is(err, context.DeadlineExceeded) {
		panic(web.Err(504, "worker 응답이 없습니다"))
	}
	if err != nil {
		panic(web.Err(502, "%v", err))
	}
	var out protocol.DirExistsResponse
	if err := res.Bind(&out); err != nil {
		panic(web.Err(502, "%v", err))
	}
	if out.Err != "" {
		status, ok := dirErrStatus[out.Err]
		if !ok {
			status = 500
		}
		panic(web.Err(status, "%s", out.Reason))
	}
	return c.JSON(200, out.Found)
}

// requireDeviceDir: 장비 디렉토리 인가 = requireDeviceFolder + 인스턴스 접속 중
func (r *supervisorRouter) requireDeviceDir(c echo.Context, userID, folderID int64, mainKey, subKey string) *transport.Conn {
	requireDeviceFolder(c, userID, folderID, mainKey)
	conn, ok := r.workers.Get(mainKey + "#" + subKey)
	if !ok {
		panic(web.Err(409, "접속 중인 인스턴스가 아닙니다"))
	}
	return conn
}

// requireDeviceFolder: 연 폴더 소유자(Q5) + 그 폴더에 적용되는 device_key(상속 포함, Q12) == mainKey
func requireDeviceFolder(c echo.Context, userID, folderID int64, mainKey string) {
	folder, _ := requireNode(c, userID, folderID, superdb.PermOwner)
	if folder.Kind != superdb.NodeKindFolder {
		panic(web.Err(400, "폴더가 아닙니다"))
	}
	key, err := TxQueries(c).ResolveDeviceKey(c.Request().Context(), superdb.ResolveDeviceKeyParams{ID: folder.ID, OwnerUserID: folder.OwnerUserID})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(500, "%v", err))
	}
	if !key.Valid || key.String != mainKey {
		panic(web.Err(403, "이 폴더에 지정된 장비가 아닙니다"))
	}
	requireDeviceOwner(c.Request().Context(), TxQueries(c), mainKey, folder.OwnerUserID)
}
