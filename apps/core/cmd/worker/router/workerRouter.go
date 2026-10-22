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

// WorkerState는 재접속 루프 동안 살아남는 worker 상태(main.go가 1회 생성)다. workerRouter는 접속마다 새로 만들어진다.
// 이게 없으면 재접속 때 procs가 빈 채로 시작돼 보관 프로세스 연결을 잃고, forward 고루틴은 죽은 conn만 물게 된다(REF-process "재접속 재바인딩")
type WorkerState struct {
	procs   *manager.KeyValManager[string, *procEntry]
	conn    atomic.Pointer[transport.Conn]
	ready   atomic.Pointer[transport.Conn] // register·sync 보고를 마친 conn — 종료 보고(finish)는 이것으로만
	store   *store.StorePool
	holdDir string
}

func NewWorkerState(store *store.StorePool, holdDir string) *WorkerState {
	return &WorkerState{procs: manager.NewKeyValManager[string, *procEntry](), store: store, holdDir: holdDir}
}

// currentConn: forward가 전송마다 읽는 "지금" conn. 끊긴 동안엔 죽은 conn이라 실패해도 로그만(재접속 때 교체됨)
func (s *WorkerState) currentConn() *transport.Conn { return s.conn.Load() }

// Result: 한 연결 세션의 종료 결과. Reached = register 성공까지 갔는지(main의 backoff reset 기준)
type Result struct {
	Reached bool
	Err     error
}

type workerRouter struct {
	conn      *transport.Conn
	state     *WorkerState // 재접속 넘어 살아남는 공유 상태(procs identity + 현재 conn 참조)
	store     *store.StorePool
	uniqueKey string
	saves     *manager.KeyValManager[string, *fileRecv]
	procs     *manager.KeyValManager[string, *procEntry] // = state.procs(동일 identity, 재접속 간 유지)

	mu     sync.RWMutex // subKey 보호(register 고루틴이 쓰고 핸들러가 읽음)
	subKey string

	reached atomic.Bool // register 성공(정상 가동) 여부. register가 쓰고 Serve 종료 경로가 읽음
}

// onReady = register 성공 때 1회(nil 가능), state = 재접속을 넘어 공유(WorkerState)
func NewWorkerRouter(supervisor url.URL, uniqueKey string, store *store.StorePool, state *WorkerState, onReady func()) (<-chan Result, error) {

	ws, _, err := websocket.DefaultDialer.Dial(supervisor.String(), nil)
	if err != nil {
		return nil, err
	}
	transport.KeepAlive(ws)
	conn := transport.New(ws)
	state.conn.Store(conn) // forward가 곧바로 이 conn을 참조하도록 먼저 교체

	done := make(chan Result, 1)
	var once sync.Once
	finish := func(r Result) {
		once.Do(func() {
			state.ready.CompareAndSwap(conn, nil)
			done <- r
			conn.Close(r.Err)
		})
	}

	router := &workerRouter{
		conn:      conn,
		state:     state,
		store:     store,
		uniqueKey: uniqueKey,
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
