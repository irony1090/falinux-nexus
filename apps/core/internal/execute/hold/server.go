package hold

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"nexus/internal/execute"
	"nexus/internal/execute/pty"
	"nexus/internal/ring"
)

const writeWait = 5 * time.Second // worker가 읽지 않으면 연결을 끊는다(PTY 출력이 막혀 스크립트가 멈추지 않게)

// holder는 보관 프로세스 하나의 상태다. 붙은 worker는 한 번에 하나(peer).
//
// 순서 의존: 출력 쓰기(screen.Write)와 peer 전송, 스냅샷 전송은 전부 mu 안에서 한다 — 스냅샷과 출력의 offset 순서가
// 소켓 위에서 뒤집히면 supervisor·브라우저가 이음매를 맞추지 못한다(S1(이음매 처리)).
type holder struct {
	inter  *pty.Interactive
	screen *ring.Screen

	mu     sync.Mutex
	peer   *peer
	status execute.CommandStatus
	exit   int
	done   bool

	outDone  chan struct{}
	quit     chan struct{}
	quitOnce sync.Once
}

type peer struct {
	conn net.Conn
	enc  *json.Encoder
}

// Main은 보관 프로세스 진입점이다(worker 실행 파일의 -hold 모드). stdin = Spec, stdout = 준비 결과 한 줄
func Main(sock string) {
	log.SetPrefix("[hold " + sock + "] ")
	// SIGHUP 처리기를 두면 무시 상태가 풀려 스크립트가 기본 처분으로 시작한다(REF-worker-orphan 안 나(SIGHUP 복원))
	signal.Notify(make(chan os.Signal, 1), syscall.SIGHUP)

	var spec Spec
	out := json.NewEncoder(os.Stdout)
	fail := func(err error) {
		_ = out.Encode(ready{Err: err.Error()})
		os.Exit(1)
	}
	if err := json.NewDecoder(os.Stdin).Decode(&spec); err != nil {
		fail(err)
	}
	_ = os.Remove(sock)
	ln, err := net.Listen("unix", sock)
	if err != nil {
		fail(err)
	}
	_ = os.Chmod(sock, 0o600)

	inter, err := pty.ExecInteractive(context.Background(), spec.Cmd, spec.Env, spec.Cwd, spec.Args...)
	if err != nil {
		ln.Close()
		os.Remove(sock)
		fail(err)
	}
	h := &holder{
		inter:   inter,
		screen:  ring.NewScreen(ScreenCap),
		outDone: make(chan struct{}),
		quit:    make(chan struct{}),
	}
	go h.pumpOutput()
	go h.pumpStatus()
	go h.accept(ln)

	_ = out.Encode(ready{Pid: inter.Pid()})
	os.Stdout.Close() // worker가 꺼져도 이 파이프에 쓰지 않게
	os.Stdin.Close()

	<-h.quit
	ln.Close()
	os.Remove(sock)
	os.Remove(logPath(sock))
	os.Exit(0)
}

func (h *holder) stop() { h.quitOnce.Do(func() { close(h.quit) }) }

func (h *holder) pumpOutput() {
	defer close(h.outDone)
	for {
		data, err := h.inter.OutputAll()
		if err != nil {
			return
		}
		h.mu.Lock()
		off := h.screen.Write(data)
		h.sendLocked(Msg{T: MsgData, Data: data, Off: off})
		h.mu.Unlock()
	}
}

// pumpStatus: 끝남은 남은 출력을 다 보낸 뒤에 알린다. 끝나면 DoneWait 안에 worker가 가져가지 않아도 스스로 끝낸다
func (h *holder) pumpStatus() {
	for {
		st, code, err := h.inter.Status()
		if err != nil {
			return
		}
		if st.IsCompleted() {
			<-h.outDone
		}
		h.mu.Lock()
		h.status = st
		if st.IsCompleted() {
			h.done, h.exit = true, code
		}
		h.sendLocked(h.statusMsg())
		h.mu.Unlock()
		if st.IsCompleted() {
			time.AfterFunc(DoneWait, h.stop)
		}
	}
}

func (h *holder) statusMsg() Msg {
	return Msg{T: MsgStatus, Status: h.status, Pid: h.inter.Pid(), Exit: h.exit}
}

// sendLocked는 붙은 worker에 보낸다(mu 안). 실패하면 그 연결을 버린다 — worker가 다시 붙으면 스냅샷으로 메운다
func (h *holder) sendLocked(m Msg) {
	if h.peer == nil {
		return
	}
	h.peer.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := h.peer.enc.Encode(m); err != nil {
		log.Printf("worker 연결 끊음: %v", err)
		h.peer.conn.Close()
		h.peer = nil
	}
}

func (h *holder) snapLocked() {
	data, off, alt := h.screen.Snapshot()
	h.sendLocked(Msg{T: MsgSnap, Data: data, Off: off, Alt: alt})
}

func (h *holder) accept(ln net.Listener) {
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		go h.serve(c)
	}
}

// serve: hello로 형식 버전을 맞춘 뒤 이 연결을 peer로(옛 peer는 끊음), 상태·스냅샷을 먼저 보내고 명령을 받는다
func (h *holder) serve(c net.Conn) {
	dec := json.NewDecoder(c)
	enc := json.NewEncoder(c)
	c.SetReadDeadline(time.Now().Add(writeWait))
	var hello Msg
	if err := dec.Decode(&hello); err != nil || hello.T != MsgHello {
		c.Close()
		return
	}
	c.SetReadDeadline(time.Time{})
	if hello.V != Version {
		enc.Encode(Msg{T: MsgHello, V: Version, Err: "형식 버전이 다름"})
		c.Close()
		return
	}

	p := &peer{conn: c, enc: enc}
	h.mu.Lock()
	if h.peer != nil {
		h.peer.conn.Close()
	}
	h.peer = p
	st := h.statusMsg()
	st.T, st.V = MsgHello, Version
	h.sendLocked(st)
	h.snapLocked()
	h.mu.Unlock()

	for {
		var m Msg
		if err := dec.Decode(&m); err != nil {
			break
		}
		h.handle(m)
	}
	h.mu.Lock()
	if h.peer == p {
		h.peer = nil
	}
	h.mu.Unlock()
	c.Close()
}

func (h *holder) handle(m Msg) {
	switch m.T {
	case MsgInput:
		_ = h.inter.Write(m.Data)
	case MsgResize:
		if err := h.inter.Layout(m.Cols, m.Rows); err != nil {
			log.Printf("resize: %v", err)
		}
	case MsgKill:
		h.mu.Lock()
		done := h.done
		h.mu.Unlock()
		if !done { // 끝난 뒤의 pid는 다른 프로세스일 수 있다
			if err := h.inter.Kill(m.Force); err != nil {
				log.Printf("kill: %v", err)
			}
		}
	case MsgSnap:
		h.mu.Lock()
		h.snapLocked()
		h.mu.Unlock()
	case MsgRelease:
		h.mu.Lock()
		done := h.done
		h.mu.Unlock()
		if done {
			h.stop()
		}
	}
}
