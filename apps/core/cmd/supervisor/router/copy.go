package router

import (
	"context"
	"io/fs"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/util"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// worker 간 전송 = 전송-10(worker 간 중계), REF-transfer-down.md. A에서 FILE_READ로 끌어온 조각을 B로 send(Abs·Exclusive). 파일 하나 = 요청 하나, 복사만(Q35·Q36)
const copyParallel = 3 // 동시에 중계하는 수(나머지는 진행 0으로 기다림)

type copyFrom struct {
	FolderID int64  `json:"folderId" validate:"required"`
	MainKey  string `json:"mainKey" validate:"required"`
	SubKey   string `json:"subKey" validate:"required"`
	Path     string `json:"path" validate:"required"` // 보낼 파일 전체 경로(worker가 이은 값)
}

type copyTo struct {
	FolderID int64  `json:"folderId" validate:"required"`
	MainKey  string `json:"mainKey" validate:"required"`
	SubKey   string `json:"subKey" validate:"required"`
	Dir      string `json:"dir" validate:"required"` // 받을 폴더
}

type copyStartRequest struct {
	From      copyFrom `json:"from"`
	To        copyTo   `json:"to"`
	Overwrite bool     `json:"overwrite"` // Q23(같은 이름 파일)
	TileID    string   `json:"tileId"`    // 받는 장비 타일 — Q27(타일 막대 범위)
}

// POST /transfers/copies: 두 장비 폴더 모두 인가 -> A에서 열어 이름·크기 확인 -> 항목 등록 -> 뒤에서 중계
func (r *supervisorRouter) startCopy(c echo.Context) error {
	sess := r.requireSession(c)
	var body copyStartRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	r.requireDeviceDir(c, sess.Data.ID, body.From.FolderID, body.From.MainKey, body.From.SubKey)
	r.requireDeviceDir(c, sess.Data.ID, body.To.FolderID, body.To.MainKey, body.To.SubKey)
	EndTx(c)

	id, err := util.RandomKey(16, "", "")
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	fromKey := body.From.MainKey + "#" + body.From.SubKey
	ctx, cancel := context.WithCancel(context.Background())
	src, err := r.openWorkerFile(ctx, fromKey, body.From.Path, nil)
	if err != nil {
		cancel()
		panicOpenError(err)
	}
	if !protocol.ValidFileName(src.Name()) {
		src.Close()
		cancel()
		panic(web.Err(400, "받는 쪽에 쓸 수 없는 파일 이름입니다: %s", src.Name()))
	}

	it := &transferItem{
		info: protocol.TransferInfo{
			ID:          id,
			Kind:        protocol.TransferKindCopy,
			TileID:      body.TileID,
			InstanceKey: body.To.MainKey + "#" + body.To.SubKey,
			Dir:         body.To.Dir,
			Name:        src.Name(),
			Size:        src.Size(),
			FromKey:     fromKey,
			FromPath:    body.From.Path,
			Phase:       protocol.TransferPhaseSend,
			State:       protocol.TransferRunning,
			StartedAt:   time.Now().UnixMilli(),
		},
		userID:    sess.Data.ID,
		overwrite: body.Overwrite,
		ctx:       ctx,
		cancel:    cancel,
	}
	r.transfers.Append(id, it)
	r.pushTransfer(it, true)
	go r.runCopy(it, src)
	return c.JSON(200, it.snapshot())
}

// runCopy: 자리가 나면 send. 기다리는 중 중단돼도 A의 읽기를 닫는다
func (r *supervisorRouter) runCopy(it *transferItem, src *workerReader) {
	defer src.Close() // send가 일찍 실패하면 send의 defer Close까지 못 감
	select {
	case r.copySlots <- struct{}{}:
		defer func() { <-r.copySlots }()
	case <-it.ctx.Done():
		r.endSend(it, context.Canceled)
		return
	}

	info := it.snapshot()
	perm := fs.FileMode(src.Mode())
	if perm == 0 {
		perm = 0644
	}
	_, err := r.send(&sendJob{
		id:        info.ID,
		ctx:       it.ctx,
		authKey:   info.InstanceKey,
		name:      info.Name,
		dir:       info.Dir,
		perm:      perm,
		exclusive: !it.overwrite,
		reader:    src,
		onProgress: func(done int64) {
			it.mu.Lock()
			it.info.Done = done
			it.mu.Unlock()
			r.pushTransfer(it, false)
		},
	})
	r.endSend(it, err)
}
