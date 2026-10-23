package router

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/util"
)

// workerReader: worker 파일을 FILE_READ로 끌어오는 읽기 원천 — 전송-7(worker 송신), REF-transfer-down.md.
// transfer.Reader라 send(worker 간 중계)에 그대로 넣고, 내려받기는 Read를 EOF까지 돌린다
type workerReader struct {
	r       *supervisorRouter
	ctx     context.Context
	authKey string
	open    protocol.FileOpenRequest
	info    protocol.FileOpenResponse
	pos     int64
	eof     bool
	stale   bool // 파일 하나: 앞 읽기가 실패함 -> 다음 읽기 전에 다시 연다(worker 재접속이면 읽기 목록이 비어 있음)
	once    sync.Once
	onClose func()
}

// workerOpenError: FILE_OPEN 거절(DirErr*) — REST는 dirErrStatus로 HTTP 상태를 고른다
type workerOpenError struct{ Code, Reason string }

func (e *workerOpenError) Error() string { return e.Reason }

// openWorkerFile: names가 비면 path = 파일 하나, 있으면 path 폴더 안의 names를 zip으로. ctx = 전송의 수명(중단 = 취소)
func (r *supervisorRouter) openWorkerFile(ctx context.Context, authKey, path string, names []string) (*workerReader, error) {
	w := &workerReader{r: r, ctx: ctx, authKey: authKey, open: protocol.FileOpenRequest{Path: path, Names: names}}
	if err := w.reopen(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *workerReader) zip() bool { return len(w.open.Names) > 0 }

// reopen: 새 readId로 FILE_OPEN. 다시 열 때 크기가 바뀌었으면 실패(이어 붙이면 내용이 섞임)
func (w *workerReader) reopen() error {
	id, err := util.RandomKey(16, "", "")
	if err != nil {
		return err
	}
	req := w.open
	req.ReadID = id
	res, err := w.call(protocol.MsgFileOpen, req)
	if err != nil {
		return fmt.Errorf("FileOpen: %w", err)
	}
	var info protocol.FileOpenResponse
	if err := res.Bind(&info); err != nil {
		return err
	}
	if info.Err != "" {
		return &workerOpenError{Code: info.Err, Reason: info.Reason}
	}
	if w.open.ReadID != "" && info.Size != w.info.Size {
		return fmt.Errorf("읽는 중 파일 크기가 바뀜(%d -> %d)", w.info.Size, info.Size)
	}
	w.open.ReadID, w.info, w.stale = id, info, false
	return nil
}

// call: 매번 지금 연결을 다시 찾는다(재접속이면 reconnectWait까지 기다림)
func (w *workerReader) call(t protocol.MsgType, body any) (protocol.Frame, error) {
	conn, ok := w.r.waitWorker(w.ctx, w.authKey)
	if !ok {
		if w.ctx.Err() != nil {
			return protocol.Frame{}, w.ctx.Err()
		}
		return protocol.Frame{}, fmt.Errorf("대상 연결 없음: %s", w.authKey)
	}
	ctx, cancel := context.WithTimeout(w.ctx, sendCallTimeout)
	defer cancel()
	return conn.Call(ctx, t, body)
}

func (w *workerReader) Name() string   { return w.info.Name }
func (w *workerReader) Size() int64    { return w.info.Size }
func (w *workerReader) Mode() uint32   { return w.info.Mode }
func (w *workerReader) Current() int64 { return w.pos }

// Hash: 빈 값 = 받는 worker가 hash 검사를 건너뜀(크기는 검사). 미리 구하려면 보내는 쪽이 파일을 한 번 더 다 읽어야 해서 안 함
func (w *workerReader) Hash() (string, error) { return "", nil }

// SeekTo: send의 재시도(이어받기)용. zip은 지금 위치만
func (w *workerReader) SeekTo(offset int64) error {
	if w.zip() {
		if offset != w.pos {
			return fmt.Errorf("zip은 되감을 수 없음(%d -> %d)", w.pos, offset)
		}
		return nil
	}
	w.pos, w.eof = offset, false
	return nil
}

// Read: 조각 하나. 실패하면 maxSendAttempts까지 다시(같은 offset — worker가 zip 앞 조각을 다시 줌)
func (w *workerReader) Read(chunkSize int64) ([]byte, error) {
	if w.eof || (!w.zip() && w.pos >= w.info.Size) {
		return nil, io.EOF
	}
	var lastErr error
	for attempt := 1; attempt <= maxSendAttempts; attempt++ {
		data, err := w.readOnce(chunkSize)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if w.ctx.Err() != nil {
			return nil, w.ctx.Err()
		}
		var oe *workerOpenError
		if errors.As(err, &oe) {
			return nil, err
		}
		log.Printf("[workerReader] %s 읽기 %d/%d 실패: %v", w.open.Path, attempt, maxSendAttempts, err)
		if !w.zip() {
			w.stale = true
		}
		time.Sleep(retryBackoff)
	}
	return nil, lastErr
}

func (w *workerReader) readOnce(chunkSize int64) ([]byte, error) {
	if w.stale {
		if err := w.reopen(); err != nil {
			return nil, err
		}
	}
	res, err := w.call(protocol.MsgFileRead, protocol.FileReadRequest{ReadID: w.open.ReadID, Offset: w.pos, Max: int(chunkSize)})
	if err != nil {
		return nil, err
	}
	var body protocol.FileReadResponse
	if err := res.Bind(&body); err != nil {
		return nil, err
	}
	w.pos += int64(len(body.Data))
	w.eof = body.EOF
	if len(body.Data) == 0 && body.EOF {
		return nil, io.EOF
	}
	return body.Data, nil
}

// Close: worker 쪽 읽기를 닫는다(best-effort, 중단된 ctx와 무관하게)
func (w *workerReader) Close() {
	w.once.Do(func() {
		if conn, ok := w.r.workers.Get(w.authKey); ok {
			ctx, cancel := context.WithTimeout(context.Background(), sendCallTimeout)
			if _, err := conn.Call(ctx, protocol.MsgFileClose, protocol.FileCloseRequest{ReadID: w.open.ReadID}); err != nil {
				log.Printf("[workerReader] %s 닫기 통보 실패(무시): %v", w.open.ReadID, err)
			}
			cancel()
		}
		if w.onClose != nil {
			w.onClose()
		}
	})
}

func (w *workerReader) SetOnClose(fn func()) { w.onClose = fn }
