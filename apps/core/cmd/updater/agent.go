package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/transport"
	"nexus/internal/updproto"

	"github.com/gorilla/websocket"
)

// agent: 접속 한 번(session)이 끊겨도 실행 중인 단계·보내지 못한 결과는 남는다 — 다음 접속에서 HELLO Running · MsgEnd 재전송
type agent struct {
	dir string
	cfg *configFile

	mu   sync.Mutex
	conn *transport.Conn // 지금 연결(nil = 끊김)
	job  *job            // 실행 중인 단계(하나만)
	ends []updproto.End  // supervisor가 받았다고 답하지 않은 결과
}

func newAgent(dir string, cfg *configFile) *agent {
	return &agent{dir: dir, cfg: cfg}
}

func (a *agent) statePath() string { return filepath.Join(a.dir, "state.json") }

// session: 접속 -> HELLO -> 끊길 때까지. reached = HELLO까지 갔음(재시도 간격 초기화)
func (a *agent) session() (reached bool) {
	c := a.cfg.get()
	wsBase, _, _ := endpoints(c.Supervisor)
	d := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	ws, _, err := d.Dial(wsBase+updproto.WSPath, nil)
	if err != nil {
		log.Printf("접속 실패 %s: %v", wsBase, err)
		return false
	}
	transport.KeepAlive(ws)
	conn := transport.New(ws)
	conn.Handle(updproto.MsgSetDir, a.onSetDir)
	conn.Handle(updproto.MsgRun, a.onRun)
	conn.On(updproto.MsgInput, a.onInput)
	conn.On(updproto.MsgResize, a.onResize)
	conn.On(updproto.MsgCancel, a.onCancel)
	served := make(chan error, 1)
	go func() { served <- conn.Serve() }()

	st := loadState(a.statePath())
	h := updproto.Hello{MainKey: c.MainKey, SubKey: st.SubKey, APIKey: c.APIKey, OS: runtime.GOOS, Arch: runtime.GOARCH,
		TargetDir: c.TargetDir, Installed: st.Installed, Version: buildTag(), Running: a.running()}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	res, err := conn.Call(ctx, updproto.MsgHello, h)
	cancel()
	if err != nil {
		log.Printf("HELLO 거절: %v", err)
		conn.Close(err)
		<-served
		return false
	}
	var hr updproto.HelloResponse
	if err := res.Bind(&hr); err != nil || hr.SubKey == "" {
		log.Printf("HELLO 응답이 이상합니다: %v", err)
		conn.Close(err)
		<-served
		return false
	}
	if hr.SubKey != st.SubKey {
		st.SubKey = hr.SubKey
		if err := saveState(a.statePath(), st); err != nil {
			log.Printf("state.json 저장 실패: %v", err)
		}
	}
	log.Printf("접속됨 %s#%s", c.MainKey, hr.SubKey)
	a.mu.Lock()
	a.conn = conn
	a.mu.Unlock()
	go a.flushEnds(conn)

	err = <-served
	a.mu.Lock()
	if a.conn == conn {
		a.conn = nil
	}
	a.mu.Unlock()
	conn.Close(err)
	log.Printf("끊김: %v", err)
	return true
}

func (a *agent) running() []int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	var ids []int64
	if a.job != nil {
		ids = append(ids, a.job.run.RunID)
	}
	for _, e := range a.ends {
		ids = append(ids, e.RunID) // 결과를 아직 못 보낸 것도 supervisor가 취소로 덮지 않게
	}
	return ids
}

// ===== supervisor -> updater =====

func (a *agent) onSetDir(f protocol.Frame) (any, error) {
	var b updproto.SetDir
	if err := f.Bind(&b); err != nil {
		return nil, err
	}
	a.mu.Lock()
	busy := a.job != nil
	a.mu.Unlock()
	if busy {
		return nil, errors.New("단계가 실행 중입니다")
	}
	if err := a.cfg.setTargetDir(b.TargetDir); err != nil {
		return nil, err
	}
	log.Printf("설치 경로 바뀜 -> %q", b.TargetDir)
	return nil, nil
}

func (a *agent) onRun(f protocol.Frame) (any, error) {
	var r updproto.Run
	if err := f.Bind(&r); err != nil {
		return nil, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.job != nil {
		return nil, fmt.Errorf("다른 단계(기록 %d)가 실행 중입니다", a.job.run.RunID)
	}
	j, err := newJob(a, r)
	if err != nil {
		return nil, err
	}
	a.job = j
	go j.main()
	return nil, nil
}

func (a *agent) current(id int64) *job {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.job != nil && a.job.run.RunID == id {
		return a.job
	}
	return nil
}

func (a *agent) onInput(f protocol.Frame) {
	var in updproto.Input
	if f.Bind(&in) == nil {
		if j := a.current(in.RunID); j != nil {
			j.input(in.Data)
		}
	}
}

func (a *agent) onResize(f protocol.Frame) {
	var rs updproto.Resize
	if f.Bind(&rs) == nil {
		if j := a.current(rs.RunID); j != nil {
			j.resize(rs.Cols, rs.Rows)
		}
	}
}

func (a *agent) onCancel(f protocol.Frame) {
	var cc updproto.Cancel
	if f.Bind(&cc) == nil {
		if j := a.current(cc.RunID); j != nil {
			j.cancel(updproto.ResultCancel)
		}
	}
}

// ===== updater -> supervisor =====

// output: 끊겨 있으면 supervisor 쪽에는 빠지고 updater 로그 파일에만 남는다
func (a *agent) output(id int64, b []byte) {
	a.mu.Lock()
	conn := a.conn
	a.mu.Unlock()
	if conn != nil {
		_ = conn.Emit(updproto.MsgOutput, updproto.Output{RunID: id, Data: b})
	}
}

func (a *agent) finish(j *job, e updproto.End) {
	a.mu.Lock()
	if a.job == j {
		a.job = nil
	}
	a.ends = append(a.ends, e)
	conn := a.conn
	a.mu.Unlock()
	log.Printf("단계 끝 기록 %d: %s %s", e.RunID, e.Result, e.Msg)
	if conn != nil {
		go a.flushEnds(conn)
	}
}

// flushEnds: 받았다는 답이 온 결과만 지운다
func (a *agent) flushEnds(conn *transport.Conn) {
	a.mu.Lock()
	pending := append([]updproto.End(nil), a.ends...)
	a.mu.Unlock()
	for _, e := range pending {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err := conn.Call(ctx, updproto.MsgEnd, e)
		cancel()
		if err != nil {
			log.Printf("결과 보내기 실패(다음 접속 때 다시) 기록 %d: %v", e.RunID, err)
			return
		}
		a.mu.Lock()
		for i, x := range a.ends {
			if x.RunID == e.RunID {
				a.ends = append(a.ends[:i], a.ends[i+1:]...)
				break
			}
		}
		a.mu.Unlock()
	}
}

func (a *agent) logFile(id int64) (*os.File, error) {
	return os.OpenFile(filepath.Join(a.dir, "logs", strconv.FormatInt(id, 10)+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}
