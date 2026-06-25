package router

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"nexus/internal/manager"
	"nexus/internal/protocol"
	"nexus/internal/transfer"
	"nexus/internal/transport"
	workerdb "nexus/internal/worker/db/gen"
	"nexus/internal/worker/store"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// recvExpire는 청크 사이 idle 허용 시간이다. 초과하면 SaveFile이 닫히지만
// .part는 디스크에 남아(이어받기 가능) saves 맵에서만 빠진다.
const recvExpire = 60 * time.Second

type Result struct {
	Reached bool
	Err     error
}

// fileRecv는 수신 중인 한 전송 세션의 상태다.
// FileResult 검증은 FileInit에서 받은 hash가 필요한데, FileResultRequest엔
// transferId만 실리므로 기대 hash와 최종 경로를 여기 보관한다.
type fileRecv struct {
	save      *transfer.SaveFile
	hash      string // FileInit에서 받은 기대 sha256(hex). 빈 값이면 hash 검증 생략
	finalPath string // 성공 시 .part를 rename할 최종 경로
}

type workerRouter struct {
	conn      *transport.Conn
	store     *store.StorePool
	uniqueKey string
	baseDir   string // 수신 파일 저장 루트(인스턴스 폴더의 상위)
	saves     *manager.KeyValManager[string, *fileRecv]

	mu     sync.RWMutex // subKey 보호(register 고루틴이 쓰고 핸들러가 읽음)
	subKey string

	reached atomic.Bool // register 성공(정상 가동) 여부. register가 쓰고 Serve 종료 경로가 읽음
}

func NewWorkerRouter(supervisor url.URL, uniqueKey, baseDir string, store *store.StorePool) (<-chan Result, error) {

	ws, _, err := websocket.DefaultDialer.Dial(supervisor.String(), nil)
	if err != nil {
		// log.Fatalf("worker: 연결 실패 %v", err)
		return nil, err
	}
	conn := transport.New(ws)

	done := make(chan Result, 1)
	var once sync.Once
	finish := func(r Result) {
		once.Do(func() {
			done <- r
			conn.Close(r.Err)
		})
	}

	router := &workerRouter{
		conn:      conn,
		store:     store,
		uniqueKey: uniqueKey,
		baseDir:   baseDir,
		saves:     manager.NewKeyValManager[string, *fileRecv](),
	}

	// 핸들러는 Serve(수신 루프)보다 먼저 등록한다(등록 전 REQ 수신 창 제거).
	conn.Handle(protocol.MsgFileInit, router.fileInit)
	conn.Handle(protocol.MsgFileChunk, router.fileChunk)
	conn.Handle(protocol.MsgFileResult, router.fileResult)
	conn.Handle(protocol.MsgFileAbort, router.fileAbort)

	go func() { // 수신 루프: 끝나면 세션 종료. reached로 정상 가동 여부 보고.
		err := conn.Serve()
		finish(Result{Reached: router.reached.Load(), Err: err})
	}()

	go func() { // 등록 실패는 세션 종료. 성공 시엔 reached만 세팅하고 종료는 Serve가 담당.
		if err := router.register(); err != nil {
			finish(Result{Reached: false, Err: err})
		}
	}()

	return done, nil
}

// instanceKey는 저장 경로(인스턴스 폴더)에 쓰는 메인키#서브키다.
func (r *workerRouter) instanceKey() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.subKey == "" {
		return r.uniqueKey
	}
	return r.uniqueKey + "#" + r.subKey
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

func (r *workerRouter) register() error {

	q := r.store.Queries()
	ctx, selectCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer selectCancel()
	identity, err := q.GetIdentity(ctx, r.uniqueKey)
	var subKey string
	switch {
	case err == nil:
		subKey = identity.SubKey
	case errors.Is(err, sql.ErrNoRows):
		// 최초 접속: 저장된 서브키 없음 (정상)
	default:
		log.Printf("[WORKER-REGISTER] GetIdentity 실패 %v", err)
	}

	req := protocol.NewRegisterRequest(r.uniqueKey, subKey)

	// REGISTER 요청 → 응답이 이 한 줄에 묶여 돌아온다.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := r.conn.Call(ctx, protocol.MsgRegister, req)
	if err != nil {
		// 서브키 중복 등으로 supervisor가 차단하면 여기로 온다.
		log.Printf("[WORKER-REGISTER] ERR %v", err)
		return err
	}

	var rr protocol.RegisterResponse
	if err := res.Bind(&rr); err != nil {
		return err
	}

	// 저장 경로에 쓸 서브키 확정(파일 수신 핸들러가 읽음).
	r.mu.Lock()
	r.subKey = rr.SubKey
	r.mu.Unlock()

	r.reached.Store(true) // supervisor가 받아줌 = 정상 가동 도달(backoff reset 기준)

	if rr.SubKey != subKey {
		ctx, updateCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer updateCancel()
		if err := q.UpsertIdentity(ctx, workerdb.UpsertIdentityParams{
			MainKey: r.uniqueKey,
			SubKey:  rr.SubKey,
		}); err != nil {
			log.Printf("[WORKER-REGISTER] UpsertIdentity 실패 %v", err)
		}
	}

	return nil
}
