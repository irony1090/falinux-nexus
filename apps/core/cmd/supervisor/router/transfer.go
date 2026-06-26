package router

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/transfer"
	"nexus/internal/transport"
	"nexus/internal/util"
)

const (
	sendChunkSize   = 32 * 1024 // 청크 크기(base64로 ~33% 부풀어 와이어에 실림)
	maxSendAttempts = 3         // 전송 재시도 횟수
	sendCallTimeout = 10 * time.Second
	retryBackoff    = 500 * time.Millisecond
)

// errRejected는 수신측이 FileInit을 거부한 영구 실패다(재시도해도 의미 없음).
var errRejected = errors.New("수신측이 전송을 거부함")

// fileSend는 송신 중인 한 전송 세션이다. authKey를 함께 들고 있어
// 연결이 끊기면 그 연결 소속 미완 전송만 골라 정리할 수 있다.
// cancel은 진행 중인 SendFile 루프를 중단시키는 손잡이다(AbortFile에서 호출).
type fileSend struct {
	authKey string
	reader  *transfer.ReadFile
	cancel  context.CancelFunc
}

// AbortFile은 진행 중인 전송을 중단한다.
// ① cancel()로 SendFile 루프를 깨우고(진행 중 Call 실패 → ctx.Err()로 재시도 차단)
// ② worker에 MsgFileAbort를 보내 .part를 지우게 한다(best-effort).
// reader 정리는 SendFile의 defer Close가 맡으므로 여기서 따로 닫지 않는다.
func (r *supervisorRouter) AbortFile(transferId, reason string) error {
	fs, ok := r.readers.Get(transferId)
	if !ok {
		return fmt.Errorf("진행 중인 전송 없음: %s", transferId)
	}
	fs.cancel()

	if conn, ok := r.workers.Get(fs.authKey); ok {
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

// SendFile은 authKey 대상 worker에게 reader가 가리키는 파일을 전송한다.
// 저장 위치(destPath)는 원본 파일명으로 정한다(수신측이 traversal 검증).
// 한 번 발급한 transferId로 최대 maxSendAttempts회까지 시도하며, 실패 시
// 이어받기(같은 transferId·destPath라 수신측 .part 보존)로 재개한다.
// 전 과정을 블로킹으로 수행하므로 호출자가 필요하면 고루틴으로 돌린다.
//
// 주의: reader는 expire 없이(0) 생성해야 한다 — 청크 사이 왕복이 길어도
// idle 타이머에 닫히지 않도록.
func (r *supervisorRouter) SendFile(authKey string, reader *transfer.ReadFile) (string, error) {
	if reader == nil {
		return "", fmt.Errorf("파일이 존재하지 않습니다")
	}
	if _, exist := r.workers.Get(authKey); !exist {
		return "", fmt.Errorf("전송할 대상이 존재하지 않습니다: %s", authKey)
	}

	transferId, err := util.RandomKey(16, "", "")
	if err != nil {
		return "", fmt.Errorf("transferId 발급 실패: %w", err)
	}

	// 취소 가능 ctx: AbortFile이 cancel()하면 진행 중 Call이 깨지고 루프가 종료된다.
	// 각 conn.Call timeout은 이 ctx에서 파생되므로(아래 sendOnce) abort가 곧장 전파됨.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 정상 종료 시 ctx 자원 정리

	if !r.readers.Append(transferId, &fileSend{authKey: authKey, reader: reader, cancel: cancel}) {
		return "", fmt.Errorf("transferId 충돌: %s", transferId)
	}
	reader.OnClose = func() { r.readers.Remove(transferId) }
	defer reader.Close() // 전송 종료 시 reader 정리(+readers에서 제거)

	hash, err := reader.Hash()
	if err != nil {
		return "", fmt.Errorf("해시 계산 실패: %w", err)
	}
	destPath := reader.Info().Name()

	var lastErr error
	for attempt := 1; attempt <= maxSendAttempts; attempt++ {
		// 재접속 대비 매 시도마다 최신 연결을 다시 조회한다.
		conn, exist := r.workers.Get(authKey)
		if !exist {
			lastErr = fmt.Errorf("대상 연결 없음: %s", authKey)
		} else if err := r.sendOnce(ctx, conn, transferId, destPath, reader, hash); err == nil {
			log.Printf("[SendFile] %s 전송 완료 → %s", transferId, destPath)
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

		log.Printf("[SendFile] %s 시도 %d/%d 실패: %v", transferId, attempt, maxSendAttempts, lastErr)
		if attempt < maxSendAttempts {
			time.Sleep(retryBackoff)
		}
	}
	return transferId, fmt.Errorf("%d회 시도 실패: %w", maxSendAttempts, lastErr)
}

// sendOnce는 FileInit → 청크 루프 → FileResult 한 바퀴를 수행한다.
func (r *supervisorRouter) sendOnce(ctx context.Context, conn *transport.Conn, transferId, destPath string, reader *transfer.ReadFile, hash string) error {
	info := reader.Info()

	// 1) FileInit — 수락 여부 + 이어받기 시작 위치.
	initCtx, cancel := context.WithTimeout(ctx, sendCallTimeout)
	res, err := conn.Call(initCtx, protocol.MsgFileInit, protocol.FileInitRequest{
		TransferID: transferId,
		Name:       info.Name(),
		DestPath:   destPath,
		Size:       reader.Size(),
		Mode:       uint32(reader.Perm()),
		Hash:       hash,
		Resume:     true,
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
		return fmt.Errorf("%w: %s", errRejected, initRes.Reason)
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
