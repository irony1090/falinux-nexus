package main

import (
	"context"
	"errors"
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

	mu         sync.Mutex
	conn       *transport.Conn      // 지금 연결(nil = 끊김)
	job        *job                 // 실행 중인 단계(하나만)
	dep        *deployRun           // 실행 중인 배포(하나만, 그 단계가 job)
	ends       []updproto.End       // supervisor가 받았다고 답하지 않은 결과
	deployEnds []updproto.DeployEnd // 같은 방식(배포 끝)

	// 순서 의존: 출력(out)과 재접속 때 로그 재전송(resendLogs)을 같은 잠금으로 줄 세운다.
	// 재전송은 파일을 읽고 보내는 동안 출력이 끼면 그 조각이 빠지거나 겹친다(10-1(로그 끊김)).
	// 잠금 순서 = outMu -> mu(반대로 잡지 말 것)
	outMu  sync.Mutex
	missed map[int64]bool // 끊긴 동안 출력이 있던 실행 — 재접속 때 로그 전체를 다시 올림

	stMu sync.Mutex // state.json 읽고 고쳐 쓰기
}

func newAgent(dir string, cfg *configFile) *agent {
	return &agent{dir: dir, cfg: cfg, missed: map[int64]bool{}}
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
	conn.Handle(updproto.MsgDeploy, a.onDeploy)
	conn.On(updproto.MsgInput, a.onInput)
	conn.On(updproto.MsgResize, a.onResize)
	conn.On(updproto.MsgCancel, a.onCancel)
	served := make(chan error, 1)
	go func() { served <- conn.Serve() }()

	st := a.state()
	h := updproto.Hello{MainKey: c.MainKey, SubKey: st.SubKey, APIKey: c.APIKey, OS: runtime.GOOS, Arch: runtime.GOARCH,
		TargetDir: c.TargetDir, Installed: st.Installed, Version: buildTag(), Running: a.running(), Deploying: a.deploying()}
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
		a.editState(func(s *state) { s.SubKey = hr.SubKey })
	}
	log.Printf("접속됨 %s#%s", c.MainKey, hr.SubKey)
	a.outMu.Lock()
	a.mu.Lock()
	a.conn = conn
	a.mu.Unlock()
	a.resendLogs(conn)
	a.outMu.Unlock()
	go a.flush(conn)

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
	busy := a.job != nil || a.dep != nil
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
	if a.job != nil || a.dep != nil {
		return nil, a.busyErr()
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
	if f.Bind(&cc) != nil {
		return
	}
	if cc.DeployID != 0 { // 배포 취소 = 다음 단계로 안 넘어감 + 지금 단계(id가 달라도) 멈춤
		a.mu.Lock()
		d, j := a.dep, a.job
		a.mu.Unlock()
		if d != nil && d.req.DeployID == cc.DeployID {
			d.stopped.Store(true)
			if j != nil {
				j.cancel(updproto.ResultCancel)
			}
		}
		return
	}
	if j := a.current(cc.RunID); j != nil {
		j.cancel(updproto.ResultCancel)
	}
}

// ===== updater -> supervisor =====

// out: 로그 파일 + supervisor(접속 중일 때). 끊겨 있으면 파일에만 남기고 재접속 때 전체를 다시 올린다
func (a *agent) out(f *os.File, id int64, b []byte) {
	a.outMu.Lock()
	defer a.outMu.Unlock()
	_, _ = f.Write(b)
	a.mu.Lock()
	conn := a.conn
	a.mu.Unlock()
	if conn == nil || conn.Emit(updproto.MsgOutput, updproto.Output{RunID: id, Data: b}) != nil {
		a.missed[id] = true
	}
}

const resendMax = 1 << 20 // 로그 재전송 상한(끝 1MB)

// resendLogs: outMu를 쥔 채로 부른다(session)
func (a *agent) resendLogs(conn *transport.Conn) {
	for id := range a.missed {
		b, err := readTail(filepath.Join(a.dir, "logs", strconv.FormatInt(id, 10)+".log"), resendMax)
		if err != nil {
			delete(a.missed, id)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err = conn.Call(ctx, updproto.MsgLog, updproto.Log{RunID: id, Data: b})
		cancel()
		if err != nil {
			log.Printf("로그 다시 보내기 실패(다음 접속 때) 기록 %d: %v", id, err)
			return
		}
		delete(a.missed, id)
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
		go a.flush(conn)
	}
}

// flush: 단계 결과 -> 배포 결과 순서(supervisor는 배포 끝에서 남은 wait 단계를 skip으로 바꾼다)
func (a *agent) flush(conn *transport.Conn) {
	if a.flushEnds(conn) {
		a.flushDeployEnds(conn)
	}
}

// flushEnds: 받았다는 답이 온 결과만 지운다. false = 보내다 실패
func (a *agent) flushEnds(conn *transport.Conn) bool {
	a.mu.Lock()
	pending := append([]updproto.End(nil), a.ends...)
	a.mu.Unlock()
	for _, e := range pending {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err := conn.Call(ctx, updproto.MsgEnd, e)
		cancel()
		if err != nil {
			log.Printf("결과 보내기 실패(다음 접속 때 다시) 기록 %d: %v", e.RunID, err)
			return false
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
	return true
}

func (a *agent) logFile(id int64) (*os.File, error) {
	return os.OpenFile(filepath.Join(a.dir, "logs", strconv.FormatInt(id, 10)+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}
