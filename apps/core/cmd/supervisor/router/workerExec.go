package router

import (
	"errors"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// workerExecRequest: POST /workers/:mainKey/:subKey/exec
type workerExecRequest struct {
	FolderID int64  `json:"folderId" validate:"required"` // 장비 타일을 연 폴더(Q15(장비 디렉토리 id))
	Cmd      string `json:"cmd" validate:"required"`      // 파일 전체 경로(worker 기준) / "여기서 셸" = protocol.PlaceholderWorkerShell
	Cwd      string `json:"cwd" validate:"required"`      // 보고 있던 폴더

	Workspace    string    `json:"workspace"`
	ParentTileID string    `json:"parentTileId"` // 장비 타일. 없거나 닫혔으면 루트
	Size         *tileSize `json:"size"`
}

// execWorkerFile: 장비 디렉토리 파일을 새 터미널 타일로 실행 — 폴더-4(실행), REF-worker-dir-exec.md
// 인가 = 목록과 같음(requireDeviceDir). 실행 가능 여부는 worker가 다시 본다(Q24(실행 대상))
func (r *supervisorRouter) execWorkerFile(c echo.Context) error {
	sess := r.requireSession(c)
	tabID := r.requireTab(c, sess.Name())
	mainKey, subKey := pathParam(c, "mainKey"), pathParam(c, "subKey")
	var body workerExecRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	size := execTileSize(c, sess.Data.ID, body.Workspace, body.Size)
	r.requireDeviceDir(c, sess.Data.ID, body.FolderID, mainKey, subKey)
	EndTx(c) // 인가 끝 — worker 실행 응답을 기다리는 동안 DB 연결을 잡지 않게

	uid, err := r.ExecFile(sess.Data.ID, mainKey+"#"+subKey, body.Cmd, body.Cwd, body.FolderID, tabID)
	if errors.Is(err, errExecRejected) {
		panic(web.Err(400, "%v", err))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return c.JSON(200, r.execTile(c, sess.Data.ID, tabID, uid, superdb.PermOwner, body.Workspace, body.ParentTileID, size))
}
