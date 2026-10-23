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

// resolveDest는 수신 경로를 worker 루트(ProcessRoot) 하위의 절대경로로 푼다.
// supervisor가 조립한 치환 토큰({WORKER_BASE} 등)을 실제 값으로 지역화한 뒤(placeholder
// 치환은 worker 책임), traversal(`..`)을 차단하고 최종 경로가 루트 하위인지 재확인한다.
// 실행 경로(exec 핸들러의 {WORKER_BASE} 치환)와 동일 규칙을 써서 선배치 위치와 실행 대상이
// 정확히 일치하게 한다.
func (r *workerRouter) resolveDest(destPath string) (string, error) {
	root := workerBase()

	// placeholder 치환(supervisor 조립 → worker 지역화). 매칭 안 되는 {..}는 그대로 둔다.
	dest := strings.ReplaceAll(destPath, protocol.PlaceholderWorkerBase, root)
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(root, dest) // 토큰 없는 순수 상대경로 하위호환
	}
	final := filepath.Clean(dest)

	rel, err := filepath.Rel(root, final)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("destPath 경로 이탈: %q", destPath)
	}
	return final, nil
}

// resolveAbsDest: 장비 디렉토리 업로드 = 디스크 전체(Q2(볼 수 있는 범위)), 인가는 supervisor가 끝냄. 이름에 경로가 섞이면 거절
func resolveAbsDest(dir, name string) (string, error) {
	if !filepath.IsAbs(dir) {
		return "", fmt.Errorf("절대 경로가 아님: %q", dir)
	}
	if !protocol.ValidFileName(name) {
		return "", fmt.Errorf("잘못된 파일 이름: %q", name)
	}
	return filepath.Join(dir, name), nil
}

// dirExists: DIR_EXISTS 핸들러 — 이미 있는 이름만 돌려준다(Q28(여러 파일 겹침))
func (r *workerRouter) dirExists(req protocol.Frame) (any, error) {
	var body protocol.DirExistsRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	if !filepath.IsAbs(body.Path) {
		return protocol.DirExistsResponse{Err: protocol.DirErrBadPath, Reason: "절대 경로가 아닙니다"}, nil
	}
	found := []protocol.DirEntry{}
	for _, name := range body.Names {
		if !protocol.ValidFileName(name) {
			continue
		}
		p := filepath.Join(body.Path, name)
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		kind := protocol.DirKindFile
		if st.IsDir() {
			kind = protocol.DirKindDir
		}
		found = append(found, protocol.DirEntry{Name: name, Path: p, Kind: kind, Size: st.Size(), ModTime: st.ModTime().Unix()})
	}
	return protocol.DirExistsResponse{Found: found}, nil
}

// fileInit: 전송 통보를 받아 수락 여부와 이어받기 시작 위치를 정한다.
func (r *workerRouter) fileInit(req protocol.Frame) (any, error) {
	var body protocol.FileInitRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}

	var final string
	var err error
	if body.Abs {
		final, err = resolveAbsDest(body.DestPath, body.Name)
	} else {
		final, err = r.resolveDest(body.DestPath)
	}
	if err != nil {
		return protocol.FileInitResponse{Accept: false, Reason: err.Error(), Code: protocol.FileErrBadPath}, nil
	}
	if res, ok := checkDest(final, body.Abs, body.Exclusive); !ok {
		return res, nil
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

// checkDest: 받을 자리 검사 — 같은 이름 폴더는 항상, 파일은 Exclusive일 때 거절. Abs면 받을 폴더가 있어야 함(MkdirAll로 아무 데나 만들지 않게)
func checkDest(final string, abs, exclusive bool) (protocol.FileInitResponse, bool) {
	if st, err := os.Stat(final); err == nil {
		if st.IsDir() {
			return protocol.FileInitResponse{Reason: "같은 이름의 폴더가 있습니다", Code: protocol.FileErrIsDir}, false
		}
		if exclusive {
			return protocol.FileInitResponse{Reason: "같은 이름의 파일이 있습니다", Code: protocol.FileErrExists}, false
		}
	}
	if abs {
		if st, err := os.Stat(filepath.Dir(final)); err != nil || !st.IsDir() {
			return protocol.FileInitResponse{Reason: "받을 폴더가 없습니다", Code: protocol.FileErrNoParent}, false
		}
	}
	return protocol.FileInitResponse{}, true
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
