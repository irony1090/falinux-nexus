package router

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/transfer"
	"nexus/internal/transport"
	"nexus/internal/util"
)

const (
	sendChunkSize   = 256 * 1024 // 청크 크기(base64로 ~33% 부풀어 와이어에 실림), 청크마다 왕복 1회 — F2(청크 왕복)
	maxSendAttempts = 3          // 전송 재시도 횟수
	sendCallTimeout = 10 * time.Second
	retryBackoff    = 500 * time.Millisecond
	reconnectWait   = 10 * time.Second       // 재시도 전 worker 재접속을 기다리는 최대 시간 — F1(재시도 창)
	reconnectPoll   = 200 * time.Millisecond // 재접속 조회 간격
)

// errRejected는 수신측이 FileInit을 거부한 영구 실패다(재시도해도 의미 없음).
var errRejected = errors.New("수신측이 전송을 거부함")

// rejectedError: FileInit 거절 사유. errors.Is(err, errRejected) = true, 코드는 errors.As로 꺼낸다
type rejectedError struct {
	Code   string // protocol.FileErr*
	Reason string
}

func (e *rejectedError) Error() string        { return fmt.Sprintf("%v: %s", errRejected, e.Reason) }
func (e *rejectedError) Is(target error) bool { return target == errRejected }

// sendJob은 송신 중인 한 전송 작업이다. authKey는 AbortFile이 worker에 통보할 대상이다.
// name(destPath)·perm은 reader에서 캐내지 않고 작업이 직접 들고 있다 — reader는
// 메타 없는 순수 바이트 스트림(transfer.Reader)이라 파일/메모리 어느 쪽이든 받는다.
// cancel은 진행 중인 send 루프를 중단시키는 손잡이다(AbortFile에서 호출).
type sendJob struct {
	authKey   string
	name      string // 수신측 저장명(destPath). dir이 있으면 그 안의 파일 이름
	dir       string // 비어 있지 않으면 장비 디렉토리 업로드(Abs): worker 디스크 절대 폴더
	perm      fs.FileMode
	exclusive bool // 이미 있으면 거절 — Q23(같은 이름 파일)
	reader    transfer.Reader
	cancel    context.CancelFunc

	id         string           // 비어 있으면 send가 발급(전송 목록은 자기 id를 그대로 씀)
	ctx        context.Context  // 비어 있으면 Background. 전송 목록의 중단이 시작 전에도 먹게
	onProgress func(done int64) // 청크마다(worker가 받은 바이트), 전송 목록 진행률
}

// AbortFile은 진행 중인 전송을 중단한다.
// ① cancel()로 SendFile 루프를 깨우고(진행 중 Call 실패 → ctx.Err()로 재시도 차단)
// ② worker에 MsgFileAbort를 보내 .part를 지우게 한다(best-effort).
// reader 정리는 SendFile의 defer Close가 맡으므로 여기서 따로 닫지 않는다.
func (r *supervisorRouter) AbortFile(transferId, reason string) error {
	job, ok := r.readers.Get(transferId)
	if !ok {
		return fmt.Errorf("진행 중인 전송 없음: %s", transferId)
	}
	job.cancel()

	if conn, ok := r.workers.Get(job.authKey); ok {
		ctx, cancel := context.WithTimeout(context.Background(), sendCallTimeout)
		defer cancel()
		if _, err := conn.Call(ctx, protocol.MsgFileAbort, protocol.FileAbortRequest{
			TransferID: transferId,
			Reason:     reason,
		}); err != nil {
			log.Printf("[AbortFile] %s worker 통보 실패(무시): %v", transferId, err)
		}
	}
	return nil
}

// SendFile은 authKey 대상 worker에게 디스크 파일(reader)을 전송한다.
// 저장명(destPath)·권한은 파일 메타(Info().Name()/Perm())에서 도출한다.
func (r *supervisorRouter) SendFile(authKey string, reader *transfer.ReadFile) (string, error) {
	if reader == nil {
		return "", fmt.Errorf("파일이 존재하지 않습니다")
	}
	return r.send(&sendJob{authKey: authKey, name: reader.Info().Name(), perm: reader.Perm(), reader: reader})
}

// SendBuffer는 authKey 대상 worker에게 메모리 버퍼(reader)를 전송한다.
// 저장명·권한은 버퍼 생성 시 호출자가 지정한 값(reader.Name()/Perm())을 쓴다.
// (EDIT seed처럼 디스크 파일이 아닌 내용을 그대로 보낼 때.)
func (r *supervisorRouter) SendBuffer(authKey string, reader *transfer.ReadBuffer, name string, perm os.FileMode) (string, error) {
	if reader == nil {
		return "", fmt.Errorf("전송할 데이터가 없습니다")
	}
	return r.send(&sendJob{authKey: authKey, name: name, perm: perm, reader: reader})
}

// send는 reader 종류(파일/메모리)와 무관한 전송 코어다: transferId 발급 → 등록 →
// hash → FileInit→청크→FileResult 한 바퀴를 maxSendAttempts회까지 시도하며, 실패
// 시 이어받기(같은 transferId·destPath라 수신측 .part 보존)로 재개한다.
// 저장명(name=destPath)·권한(perm)·받는 방식은 호출 래퍼가 job에 채워 넘긴다(cancel은 여기서 채움).
// 전 과정을 블로킹으로 수행하므로 호출자가 필요하면 고루틴으로 돌린다.
//
// 주의: reader는 expire 없이(0) 생성해야 한다 — 청크 사이 왕복이 길어도
// idle 타이머에 닫히지 않도록.
func (r *supervisorRouter) send(job *sendJob) (string, error) {
	authKey, reader := job.authKey, job.reader
	if _, exist := r.workers.Get(authKey); !exist {
		return "", fmt.Errorf("전송할 대상이 존재하지 않습니다: %s", authKey)
	}

	transferId := job.id
	if transferId == "" {
		var err error
		if transferId, err = util.RandomKey(16, "", ""); err != nil {
			return "", fmt.Errorf("transferId 발급 실패: %w", err)
		}
	}

	// 취소 가능 ctx: AbortFile이 cancel()하면 진행 중 Call이 깨지고 루프가 종료된다.
	// 각 conn.Call timeout은 이 ctx에서 파생되므로(아래 sendOnce) abort가 곧장 전파됨.
	parent := job.ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	defer cancel() // 정상 종료 시 ctx 자원 정리

	job.cancel = cancel
	if !r.readers.Append(transferId, job) {
		return "", fmt.Errorf("transferId 충돌: %s", transferId)
	}
	reader.SetOnClose(func() { r.readers.Remove(transferId) })
	defer reader.Close() // 전송 종료 시 reader 정리(+readers에서 제거)

	hash, err := reader.Hash()
	if err != nil {
		return "", fmt.Errorf("해시 계산 실패: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= maxSendAttempts; attempt++ {
		// 재접속 대비 매 시도마다 최신 연결을 다시 조회한다(없으면 재접속을 기다림).
		conn, exist := r.waitWorker(ctx, authKey)
		if !exist {
			lastErr = fmt.Errorf("대상 연결 없음: %s", authKey)
		} else if err := r.sendOnce(ctx, conn, transferId, job, hash); err == nil {
			log.Printf("[send] %s 전송 완료 → %s", transferId, job.name)
			return transferId, nil // 성공
		} else if errors.Is(err, errRejected) {
			return transferId, err // 영구 실패 → 재시도 안 함
		} else {
			lastErr = err
		}

		// abort로 취소됐으면 재시도하지 않고 즉시 종료한다.
		if ctx.Err() != nil {
			return transferId, fmt.Errorf("전송 중단됨: %w", ctx.Err())
		}

		log.Printf("[send] %s 시도 %d/%d 실패: %v", transferId, attempt, maxSendAttempts, lastErr)
		if attempt < maxSendAttempts {
			time.Sleep(retryBackoff)
		}
	}
	return transferId, fmt.Errorf("%d회 시도 실패: %w", maxSendAttempts, lastErr)
}

// waitWorker는 authKey 연결을 reconnectWait까지 기다린다. abort(ctx 취소)면 즉시 빠진다.
func (r *supervisorRouter) waitWorker(ctx context.Context, authKey string) (*transport.Conn, bool) {
	deadline := time.Now().Add(reconnectWait)
	for {
		if conn, ok := r.workers.Get(authKey); ok {
			return conn, true
		}
		if time.Now().After(deadline) {
			return nil, false
		}
		select {
		case <-ctx.Done():
			return nil, false
		case <-time.After(reconnectPoll):
		}
	}
}

// sendOnce는 FileInit → 청크 루프 → FileResult 한 바퀴를 수행한다.
// name(destPath)·perm은 reader가 아니라 작업 메타(job)로 넘어온다(reader는 스트림 전용).
func (r *supervisorRouter) sendOnce(ctx context.Context, conn *transport.Conn, transferId string, job *sendJob, hash string) error {
	reader := job.reader
	dest := job.name
	if job.dir != "" {
		dest = job.dir
	}
	// 1) FileInit — 수락 여부 + 이어받기 시작 위치.
	initCtx, cancel := context.WithTimeout(ctx, sendCallTimeout)
	res, err := conn.Call(initCtx, protocol.MsgFileInit, protocol.FileInitRequest{
		TransferID: transferId,
		Name:       job.name,
		DestPath:   dest,
		Size:       reader.Size(),
		Mode:       uint32(job.perm),
		Hash:       hash,
		Resume:     true,
		Abs:        job.dir != "",
		Exclusive:  job.exclusive,
	})
	cancel()
	if err != nil {
		return fmt.Errorf("FileInit: %w", err)
	}
	var initRes protocol.FileInitResponse
	if err := res.Bind(&initRes); err != nil {
		return err
	}
	if !initRes.Accept {
		return &rejectedError{Code: initRes.Code, Reason: initRes.Reason}
	}

	// 2) 청크 루프 — ResumeOffset부터 EOF까지.
	offset := initRes.ResumeOffset
	if offset < 0 || offset > reader.Size() {
		return fmt.Errorf("잘못된 resumeOffset %d (size=%d)", offset, reader.Size())
	}
	if err := reader.SeekTo(offset); err != nil {
		return err
	}
	for {
		data, err := reader.Read(sendChunkSize)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read@%d: %w", offset, err)
		}
		if len(data) == 0 {
			break
		}
		cctx, ccancel := context.WithTimeout(ctx, sendCallTimeout)
		_, err = conn.Call(cctx, protocol.MsgFileChunk, protocol.FileChunkRequest{
			TransferID: transferId,
			Offset:     offset,
			Data:       data,
		})
		ccancel()
		if err != nil {
			return fmt.Errorf("chunk@%d: %w", offset, err)
		}
		offset += int64(len(data))
		if job.onProgress != nil {
			job.onProgress(offset)
		}
	}

	// 3) FileResult — 수신측 검증(크기+hash) 결과.
	rctx, rcancel := context.WithTimeout(ctx, sendCallTimeout)
	rres, err := conn.Call(rctx, protocol.MsgFileResult, protocol.FileResultRequest{TransferID: transferId})
	rcancel()
	if err != nil {
		return fmt.Errorf("FileResult: %w", err)
	}
	var result protocol.FileResult
	if err := rres.Bind(&result); err != nil {
		return err
	}
	if !result.OK {
		return fmt.Errorf("수신 검증 실패: %s (resumable=%v)", result.Reason, result.Resumable)
	}
	return nil
}
