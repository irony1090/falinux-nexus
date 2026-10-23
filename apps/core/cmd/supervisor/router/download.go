package router

import (
	"context"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/util"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 내려받기 = 전송-8(파일 하나)·전송-9(zip), REF-transfer-down.md. POST로 등록(인가·worker에서 열기) -> 브라우저 기본 다운로드가 GET으로 받음(임시 파일 없음)
const downloadStartWait = time.Minute // 등록 뒤 GET이 이만큼 안 오면 중단

type downloadStartRequest struct {
	FolderID int64    `json:"folderId" validate:"required"` // 장비 타일을 연 폴더(인가)
	Dir      string   `json:"dir" validate:"required"`      // 지금 폴더(worker 절대 경로)
	Names    []string `json:"names" validate:"required,min=1"`
	Path     string   `json:"path"`   // 파일 하나면 그 전체 경로(worker가 이은 값). 비면 Dir 안의 Names를 zip으로 — Q33(내려받기 범위)
	TileID   string   `json:"tileId"` // 시작한 장비 타일 — Q27(타일 막대 범위)
}

// POST /workers/:mainKey/:subKey/downloads: worker에서 열어 두고 항목을 돌려준다. 받기 = GET /transfers/:id/data
func (r *supervisorRouter) startDownload(c echo.Context) error {
	sess := r.requireSession(c)
	mainKey, subKey := pathParam(c, "mainKey"), pathParam(c, "subKey")
	var body downloadStartRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	r.requireDeviceDir(c, sess.Data.ID, body.FolderID, mainKey, subKey)
	EndTx(c)

	id, err := util.RandomKey(16, "", "")
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	authKey := mainKey + "#" + subKey
	ctx, cancel := context.WithCancel(context.Background())
	var src *workerReader
	if body.Path != "" {
		src, err = r.openWorkerFile(ctx, authKey, body.Path, nil)
	} else {
		src, err = r.openWorkerFile(ctx, authKey, body.Dir, body.Names)
	}
	if err != nil {
		cancel()
		panicOpenError(err)
	}

	it := &transferItem{
		info: protocol.TransferInfo{
			ID:          id,
			Kind:        protocol.TransferKindDownload,
			TileID:      body.TileID,
			InstanceKey: authKey,
			Dir:         body.Dir,
			Name:        src.Name(),
			Size:        src.Size(),
			Phase:       protocol.TransferPhaseDownload,
			State:       protocol.TransferRunning,
			StartedAt:   time.Now().UnixMilli(),
		},
		userID: sess.Data.ID,
		ctx:    ctx,
		cancel: cancel,
		src:    src,
	}
	r.transfers.Append(id, it)
	r.pushTransfer(it, true)
	time.AfterFunc(downloadStartWait, func() {
		it.mu.Lock()
		started := it.started
		it.mu.Unlock()
		if !started {
			r.endTransfer(it, protocol.TransferAborted, "", "내려받기가 시작되지 않음")
		}
	})
	return c.JSON(200, it.snapshot())
}

// panicOpenError: FILE_OPEN 거절 = DirErr*에 맞는 상태, 그 밖(연결·응답 없음) = 502
func panicOpenError(err error) {
	var oe *workerOpenError
	if errors.As(err, &oe) {
		status, ok := dirErrStatus[oe.Code]
		if !ok {
			status = 500
		}
		panic(web.Err(status, "%s", oe.Reason))
	}
	panic(web.Err(502, "%v", err))
}

// GET /transfers/:id/data: 내려받기 본문(한 번만). 도중에 실패하면 연결을 끊어 브라우저가 실패로 보게 한다
func (r *supervisorRouter) downloadData(c echo.Context) error {
	sess := r.requireSession(c)
	EndTx(c)
	it := r.requireTransfer(c, sess.Data.ID)

	it.mu.Lock()
	src := it.src
	ok := it.info.Kind == protocol.TransferKindDownload && it.info.State == protocol.TransferRunning && !it.started && src != nil
	if ok {
		it.started = true
	}
	info := it.info
	it.mu.Unlock()
	if !ok {
		panic(web.Err(409, "이미 시작했거나 끝난 내려받기입니다"))
	}
	defer src.Close()

	h := c.Response().Header()
	h.Set(echo.HeaderContentType, "application/octet-stream")
	h.Set(echo.HeaderContentDisposition, mime.FormatMediaType("attachment", map[string]string{"filename": info.Name}))
	h.Set("Cache-Control", "no-store")
	if info.Size >= 0 {
		h.Set(echo.HeaderContentLength, strconv.FormatInt(info.Size, 10))
	}
	c.Response().WriteHeader(200)

	var done int64
	for {
		data, err := src.Read(sendChunkSize)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			if errors.Is(err, context.Canceled) {
				r.endTransfer(it, protocol.TransferAborted, "", "사용자가 중단함")
			} else {
				r.endTransfer(it, protocol.TransferFailed, "", err.Error())
			}
			abortConn(c)
			return nil
		}
		if _, err := c.Response().Write(data); err != nil {
			r.endTransfer(it, protocol.TransferAborted, "", "브라우저 연결 끊김")
			return nil
		}
		done += int64(len(data))
		it.mu.Lock()
		it.info.Done = done
		it.mu.Unlock()
		r.pushTransfer(it, false)
	}
	if info.Size >= 0 && done != info.Size {
		r.endTransfer(it, protocol.TransferFailed, "", "받은 크기가 다름(파일이 바뀜)")
		abortConn(c)
		return nil
	}
	c.Response().Flush()
	r.endTransfer(it, protocol.TransferDone, "", "")
	return nil
}

// abortConn: 응답을 정상 종료로 닫지 않고 연결을 끊는다(Content-Length 없는 zip은 정상 종료하면 잘린 파일이 완료로 보임)
func abortConn(c echo.Context) {
	conn, _, err := http.NewResponseController(c.Response().Writer).Hijack()
	if err != nil {
		log.Printf("[download] 연결 끊기 실패: %v", err)
		return
	}
	conn.Close()
}
