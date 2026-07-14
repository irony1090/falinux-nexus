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

// WorkerState는 main.go의 재접속 루프 동안 살아남는 worker 상태다. workerRouter 자체는
// 재접속 시도마다 새로 만들어지지만, 실행 중이던 PTY(procs)와 그 pump/teardown 고루틴이
// 참조할 "현재 conn"은 이 구조체를 통해 세대(generation)를 넘어 공유된다 — 안 하면 재접속
// 시 procs가 빈 채로 시작돼 끊기기 전 PTY를 잃어버리고, pump 고루틴은 죽은 conn만 계속
// 물게 된다(REF-process "worker 끊김→PENDING→재접속 재바인딩"의 worker측 대응).
// main.go가 재접속 루프 시작 전 1회 생성해 매 NewWorkerRouter 호출에 그대로 넘긴다.
type WorkerState struct {
	procs *manager.KeyValManager[string, *procEntry]
	conn  atomic.Pointer[transport.Conn]
}

func NewWorkerState() *WorkerState {
	return &WorkerState{procs: manager.NewKeyValManager[string, *procEntry]()}
}

// currentConn은 pump/teardown이 매 전송 시점에 읽는 "지금" conn이다. 연결이 끊긴 동안엔
// 죽은 conn을 반환할 수 있어 Emit/Call이 실패할 수 있다(로그만 남기고 계속 — 재접속하면
// NewWorkerRouter가 이 값을 새 conn으로 교체해 자동 회복된다).
func (s *WorkerState) currentConn() *transport.Conn { return s.conn.Load() }

// Result는 한 연결 세션의 종료 결과다.
// Reached=정상 가동(register 성공)까지 갔는지 → main의 backoff reset 기준.
type Result struct {
	Reached bool
	Err     error
}

type workerRouter struct {
	conn      *transport.Conn
	state     *WorkerState // 재접속 넘어 살아남는 공유 상태(procs identity + 현재 conn 참조)
	store     *store.StorePool
	uniqueKey string
	baseDir   string // 수신 파일 저장 루트(인스턴스 폴더의 상위)
	saves     *manager.KeyValManager[string, *fileRecv]
	procs     *manager.KeyValManager[string, *procEntry] // = state.procs(동일 identity, 재접속 간 유지)

	mu     sync.RWMutex // subKey 보호(register 고루틴이 쓰고 핸들러가 읽음)
	subKey string

	reached atomic.Bool // register 성공(정상 가동) 여부. register가 쓰고 Serve 종료 경로가 읽음
}

// onReady는 register 성공(정상 가동) 시 1회 호출된다(주로 로그용). nil이면 무시.
// state는 main.go가 재접속 루프 시작 전 1회 만들어 매 호출에 그대로 넘기는 공유 상태다
// (NewWorkerState) — procs와 pump의 "현재 conn"이 재접속을 넘어 살아남게 한다.
func NewWorkerRouter(supervisor url.URL, uniqueKey, baseDir string, store *store.StorePool, state *WorkerState, onReady func()) (<-chan Result, error) {

	ws, _, err := websocket.DefaultDialer.Dial(supervisor.String(), nil)
	if err != nil {
		return nil, err
	}
	conn := transport.New(ws)
	state.conn.Store(conn) // pump/teardown이 곧바로 이 conn을 참조하도록 먼저 교체

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
		state:     state,
		store:     store,
		uniqueKey: uniqueKey,
		baseDir:   baseDir,
		saves:     manager.NewKeyValManager[string, *fileRecv](),
		procs:     state.procs,
	}

	// 핸들러는 Serve(수신 루프)보다 먼저 등록한다(등록 전 REQ 수신 창 제거).
	conn.Handle(protocol.MsgFileInit, router.fileInit)
	conn.Handle(protocol.MsgFileChunk, router.fileChunk)
	conn.Handle(protocol.MsgFileResult, router.fileResult)
	conn.Handle(protocol.MsgFileAbort, router.fileAbort)

	conn.Handle(protocol.MsgExec, router.exec)
	conn.On(protocol.MsgData, router.input)
	conn.Handle(protocol.MsgResize, router.resize)
	conn.Handle(protocol.MsgKill, router.kill)

	go func() { // 수신 루프: 끝나면 세션 종료. reached로 정상 가동 여부 보고.
		err := conn.Serve()
		finish(Result{Reached: router.reached.Load(), Err: err})
	}()

	go func() { // 등록 실패는 세션 종료. 성공 시엔 reached만 세팅하고 종료는 Serve가 담당.
		if err := router.register(); err != nil {
			finish(Result{Reached: false, Err: err})
			return
		}
		if onReady != nil {
			onReady()
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
