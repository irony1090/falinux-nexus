package router

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/transfer"
	"nexus/internal/util"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 업로드 = 폴더-5(업로드), REF-transfer-ext.md. 브라우저가 조각(512KB)으로 supervisor 임시 파일에 올리고 끝나면 worker로 send — Q30(조각 업로드)
// 조각 상한: 앞단 nginx client_max_body_size 기본 1MB 아래로 보내야 해서 브라우저가 나누고, 서버는 이 이상 거절
const uploadChunkMax = 1 << 20

type uploadStartRequest struct {
	FolderID  int64  `json:"folderId" validate:"required"` // 장비 타일을 연 폴더(인가)
	Path      string `json:"path" validate:"required"`     // 받을 폴더(worker 절대 경로)
	Name      string `json:"name" validate:"required"`
	Size      int64  `json:"size" validate:"min=0"`
	Overwrite bool   `json:"overwrite"` // Q23(같은 이름 파일): 확인 창에서 덮어쓰기를 골랐을 때만
	TileID    string `json:"tileId"`    // 시작한 장비 타일 — Q27(타일 막대 범위)
}

type uploadDataResponse struct {
	Received int64 `json:"received"`
}

// POST /workers/:mainKey/:subKey/uploads: 전송 목록에 등록(단계 = 올리는 중)하고 항목을 돌려준다
func (r *supervisorRouter) startUpload(c echo.Context) error {
	sess := r.requireSession(c)
	mainKey, subKey := pathParam(c, "mainKey"), pathParam(c, "subKey")
	var body uploadStartRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if !protocol.ValidFileName(body.Name) {
		panic(web.Err(400, "쓸 수 없는 파일 이름입니다: %s", body.Name))
	}
	r.requireDeviceDir(c, sess.Data.ID, body.FolderID, mainKey, subKey)
	EndTx(c)

	id, err := util.RandomKey(16, "", "")
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	os.MkdirAll(r.uploadDir, 0700) // 다른 실행이 하루 지난 폴더로 보고 지웠을 수도 있음(newUploadDir)
	tmpPath := filepath.Join(r.uploadDir, id)
	tmp, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		panic(web.Err(500, "임시 파일 생성 실패: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	it := &transferItem{
		info: protocol.TransferInfo{
			ID:          id,
			Kind:        protocol.TransferKindUpload,
			TileID:      body.TileID,
			InstanceKey: mainKey + "#" + subKey,
			Dir:         body.Path,
			Name:        body.Name,
			Size:        body.Size,
			Phase:       protocol.TransferPhaseUpload,
			State:       protocol.TransferRunning,
			StartedAt:   time.Now().UnixMilli(),
		},
		userID:    sess.Data.ID,
		overwrite: body.Overwrite,
		ctx:       ctx,
		cancel:    cancel,
		tmp:       tmp,
		tmpPath:   tmpPath,
		touched:   time.Now(),
	}
	r.transfers.Append(id, it)
	r.pushTransfer(it, true)
	return c.JSON(200, it.snapshot())
}

// PUT /transfers/:id/data?offset=: 조각 하나. offset이 받은 양보다 크면 409(받은 양을 알려 줌), 작으면 덮어씀(재시도 멱등)
func (r *supervisorRouter) uploadData(c echo.Context) error {
	sess := r.requireSession(c)
	EndTx(c)
	it := r.requireTransfer(c, sess.Data.ID)
	offset, err := strconv.ParseInt(c.QueryParam("offset"), 10, 64)
	if err != nil || offset < 0 {
		panic(web.Err(400, "잘못된 offset입니다"))
	}
	data, err := io.ReadAll(http.MaxBytesReader(c.Response(), c.Request().Body, uploadChunkMax))
	var tooBig *http.MaxBytesError
	if errors.As(err, &tooBig) {
		panic(web.Err(413, "조각이 너무 큽니다(최대 %d바이트)", uploadChunkMax))
	}
	if err != nil {
		panic(web.Err(400, "조각 읽기 실패: %v", err))
	}

	received := it.writeChunk(offset, data)
	r.pushTransfer(it, false)
	return c.JSON(200, uploadDataResponse{Received: received})
}

// writeChunk: 임시 파일 offset 위치에 쓰고 받은 양을 돌려준다. 실패 = panic(web.Err)
func (it *transferItem) writeChunk(offset int64, data []byte) int64 {
	it.mu.Lock()
	defer it.mu.Unlock()
	if it.info.State != protocol.TransferRunning || it.info.Phase != protocol.TransferPhaseUpload || it.tmp == nil {
		panic(web.Err(409, "올리는 중인 전송이 아닙니다"))
	}
	if offset > it.info.Done {
		panic(web.Err(409, "받은 양(%d)보다 뒤입니다", it.info.Done))
	}
	end := offset + int64(len(data))
	if end > it.info.Size {
		panic(web.Err(400, "파일 크기를 넘습니다"))
	}
	if _, err := it.tmp.WriteAt(data, offset); err != nil {
		panic(web.Err(500, "임시 파일 쓰기 실패: %v", err))
	}
	it.info.Done = max(it.info.Done, end)
	it.touched = time.Now()
	return it.info.Done
}

// POST /transfers/:id/finish: 다 받았으면 단계 = 보내는 중으로 바꾸고 뒤에서 worker로 send
func (r *supervisorRouter) finishUpload(c echo.Context) error {
	sess := r.requireSession(c)
	EndTx(c)
	it := r.requireTransfer(c, sess.Data.ID)

	it.mu.Lock()
	if it.info.State != protocol.TransferRunning || it.info.Phase != protocol.TransferPhaseUpload || it.tmp == nil {
		it.mu.Unlock()
		panic(web.Err(409, "올리는 중인 전송이 아닙니다"))
	}
	if it.info.Done != it.info.Size {
		done := it.info.Done
		it.mu.Unlock()
		panic(web.Err(409, "아직 다 받지 않았습니다(%d/%d)", done, it.info.Size))
	}
	it.tmp.Close()
	it.tmp = nil
	it.info.Phase, it.info.Done = protocol.TransferPhaseSend, 0
	it.mu.Unlock()

	r.pushTransfer(it, true)
	go r.runUpload(it)
	return c.JSON(200, it.snapshot())
}

// runUpload: 보내는 중 단계 — 임시 파일을 worker 받을 폴더로(Abs), 결과를 항목에 적는다
func (r *supervisorRouter) runUpload(it *transferItem) {
	info := it.snapshot()
	reader, err := transfer.NewReadFile(it.tmpPath, 0)
	if err != nil {
		r.endTransfer(it, protocol.TransferFailed, "", "임시 파일 열기 실패: "+err.Error())
		return
	}
	defer reader.Close() // send가 일찍 실패하면 send의 defer Close까지 못 감

	_, err = r.send(&sendJob{
		id:        info.ID,
		ctx:       it.ctx,
		authKey:   info.InstanceKey,
		name:      info.Name,
		dir:       info.Dir,
		perm:      0644,
		exclusive: !it.overwrite,
		reader:    reader,
		onProgress: func(done int64) {
			it.mu.Lock()
			it.info.Done = done
			it.mu.Unlock()
			r.pushTransfer(it, false)
		},
	})
	var rej *rejectedError
	switch {
	case err == nil:
		it.mu.Lock()
		it.info.Done = it.info.Size
		it.mu.Unlock()
		r.endTransfer(it, protocol.TransferDone, "", "")
	case errors.Is(err, context.Canceled):
		r.endTransfer(it, protocol.TransferAborted, "", "사용자가 중단함")
	case errors.As(err, &rej):
		r.endTransfer(it, protocol.TransferFailed, rej.Code, rej.Reason)
	default:
		r.endTransfer(it, protocol.TransferFailed, "", err.Error())
	}
}
