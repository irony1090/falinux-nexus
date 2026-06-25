package router

import (
	"net/url"
	"sync"
	"sync/atomic"

	"nexus/internal/manager"
	"nexus/internal/protocol"
	"nexus/internal/transport"
	"nexus/internal/worker/store"

	"github.com/gorilla/websocket"
)

// Result는 한 연결 세션의 종료 결과다.
// Reached=정상 가동(register 성공)까지 갔는지 → main의 backoff reset 기준.
type Result struct {
	Reached bool
	Err     error
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
