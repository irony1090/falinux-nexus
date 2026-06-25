package router

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/transfer"
)

// recvExpire는 청크 사이 idle 허용 시간이다. 초과하면 SaveFile이 닫히지만
// .part는 디스크에 남아(이어받기 가능) saves 맵에서만 빠진다.
const recvExpire = 60 * time.Second

// fileRecv는 수신 중인 한 전송 세션의 상태다.
// FileResult 검증은 FileInit에서 받은 hash가 필요한데, FileResultRequest엔
// transferId만 실리므로 기대 hash와 최종 경로를 여기 보관한다.
type fileRecv struct {
	save      *transfer.SaveFile
	hash      string // FileInit에서 받은 기대 sha256(hex). 빈 값이면 hash 검증 생략
	finalPath string // 성공 시 .part를 rename할 최종 경로
}

// resolveDest는 수신 상대경로를 인스턴스 루트 하위의 절대경로로 푼다.
// traversal(절대경로/`..`)을 차단하고, 최종 경로가 루트 하위인지 재확인한다.
func (r *workerRouter) resolveDest(destPath string) (string, error) {
	root := filepath.Join(r.baseDir, r.instanceKey())

	// "/" 접두 후 Clean → 루트 위로 빠지는 선행 ".."를 무력화한 뒤 상대경로화.
	clean := filepath.Clean(string(filepath.Separator) + destPath)
	clean = strings.TrimPrefix(clean, string(filepath.Separator))
	if clean == "" {
		return "", fmt.Errorf("destPath 비어있음")
	}

	final := filepath.Join(root, clean)
	rel, err := filepath.Rel(root, final)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("destPath 경로 이탈: %q", destPath)
	}
	return final, nil
}

// fileInit: 전송 통보를 받아 수락 여부와 이어받기 시작 위치를 정한다.
func (r *workerRouter) fileInit(req protocol.Frame) (any, error) {
	var body protocol.FileInitRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}

	final, err := r.resolveDest(body.DestPath)
	if err != nil {
		return protocol.FileInitResponse{Accept: false, Reason: err.Error()}, nil
	}
	part := final + ".part"

	// 같은 transferId가 이미 등록돼 있으면(직전 시도의 잔재) 닫아 정리한다.
	// .part는 Close로 지워지지 않으므로 이어받기 바이트는 보존된다.
	if old, ok := r.saves.Get(body.TransferID); ok {
		old.save.Close()
	}

	// 이어받기 판단은 받는 쪽 몫: .part 크기를 본다.
	var resumeOffset int64
	if body.Resume {
		if s, e := os.Stat(part); e == nil {
			if s.Size() <= body.Size {
				resumeOffset = s.Size()
			} else {
				os.Remove(part) // 원본보다 큰 잔재 → 폐기하고 처음부터
			}
		}
	} else {
		os.Remove(part) // 이어받기 의향 없음 → 처음부터
	}

	perm := os.FileMode(body.Mode)
	if perm == 0 {
		perm = 0644 // 0이면 read-back(Hash)·후속 접근 불가 → 기본값
	}

	save, err := transfer.NewSaveFile(part, body.Size, recvExpire, perm)
	if err != nil {
		return protocol.FileInitResponse{Accept: false, Reason: err.Error()}, nil
	}

	rec := &fileRecv{save: save, hash: body.Hash, finalPath: final}
	if !r.saves.Append(body.TransferID, rec) {
		save.Close()
		return protocol.FileInitResponse{Accept: false, Reason: "transferId 중복"}, nil
	}
	save.OnClose = func() { r.saves.Remove(body.TransferID) }

	return protocol.FileInitResponse{Accept: true, ResumeOffset: resumeOffset}, nil
}

// fileChunk: 한 조각을 offset 위치에 멱등 쓰기(WriteAt)한다.
func (r *workerRouter) fileChunk(req protocol.Frame) (any, error) {
	var body protocol.FileChunkRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	rec, ok := r.saves.Get(body.TransferID)
	if !ok {
		return nil, fmt.Errorf("알 수 없는 전송: %s", body.TransferID)
	}
	if err := rec.save.WriteAt(body.Offset, body.Data); err != nil {
		return nil, err
	}
	return protocol.FileChunkResponse{Received: rec.save.Written()}, nil
}

// fileResult: 크기·hash를 검증하고 성공 시 .part를 최종 경로로 원자적 완성한다.
func (r *workerRouter) fileResult(req protocol.Frame) (any, error) {
	var body protocol.FileResultRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	rec, ok := r.saves.Get(body.TransferID)
	if !ok {
		return nil, fmt.Errorf("알 수 없는 전송: %s", body.TransferID)
	}

	// 1) 크기 검증 — 모자라면 partial 보존, 이어받기 가능.
	if !rec.save.Completed() {
		return protocol.FileResult{OK: false, Reason: "전송 미완료", Resumable: true}, nil
	}

	// 2) 무결성 hash 검증.
	if rec.hash != "" {
		got, err := rec.save.Hash()
		if err != nil {
			return protocol.FileResult{OK: false, Reason: "hash 계산 실패: " + err.Error(), Resumable: true}, nil
		}
		if got != rec.hash {
			// 내용 손상 → 이어받아도 해결 안 됨. .part 폐기 후 처음부터 재전송.
			rec.save.Remove()
			return protocol.FileResult{OK: false, Reason: "hash 불일치", Resumable: false}, nil
		}
	}

	// 3) 완성: fsync → 닫기 → 원자적 rename.
	if err := rec.save.Sync(); err != nil {
		return protocol.FileResult{OK: false, Reason: "sync 실패: " + err.Error(), Resumable: true}, nil
	}
	rec.save.Close() // fd 닫기(+OnClose로 saves에서 제거)
	if err := os.Rename(rec.finalPath+".part", rec.finalPath); err != nil {
		return protocol.FileResult{OK: false, Reason: "rename 실패: " + err.Error(), Resumable: true}, nil
	}

	log.Printf("[WORKER-RECV] 완료 %s → %s", body.TransferID, rec.finalPath)
	return protocol.FileResult{OK: true}, nil
}

// fileAbort: 진행 중 전송을 취소하고 .part를 지운다.
func (r *workerRouter) fileAbort(req protocol.Frame) (any, error) {
	var body protocol.FileAbortRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	if rec, ok := r.saves.Get(body.TransferID); ok {
		rec.save.Remove() // .part 삭제(+OnClose로 saves에서 제거)
		log.Printf("[WORKER-RECV] 중단 %s: %s", body.TransferID, body.Reason)
	}
	return nil, nil
}
