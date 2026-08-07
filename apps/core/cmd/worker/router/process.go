package router

import (
	"context"
	"fmt"
	"log"
	"nexus/cmd/worker/constants"
	"nexus/internal/execute"
	"nexus/internal/execute/pty"
	"nexus/internal/protocol"
	"os"
	"strings"
)

// ===== process 실행 엔드포인트 (supervisor → worker) =====

// procEntry는 이 worker에서 실행 중인 process 하나의 로컬 핸들이다(휘발성 — DB 없음).
// supervisor의 ProcessEntry에 대응하지만 Record/pool이 없다(REF-process: worker는 spec
// 내부작성·영속 책임 없음). kind/editPath는 종료 시 EDIT read-back 여부를 pump가 판별하는 데 쓴다.
type procEntry struct {
	uid      string
	kind     protocol.ExecType
	editPath string
	inter    *pty.Interactive
}

// workerBase는 이 worker의 실행/저장 루트다({WORKER_BASE} 치환 값). exec·resolveDest가
// 같은 값을 써야 선배치 위치와 실행 대상 경로가 일치한다.
func workerBase() string { return constants.GetEnv().ProcessRoot }

// resolveEditor는 {WORKER_EDITOR} 치환 값을 정한다: $VISUAL > $EDITOR > env(EDITOR) > "vi" 폴백.
// (편집기 선택은 worker 책임 — supervisor는 토큰만 보낸다.)
func resolveEditor() string {
	if v := constants.GetEnv().Editor; v != "" {
		return v
	}
	return "vi"
}

// localize는 supervisor가 박아 보낸 치환 토큰을 이 worker의 실제 값으로 바꾼다.
// 매칭 안 되는 {..}는 그대로 둔다(supervisor 규약).
func localize(s string) string {
	s = strings.ReplaceAll(s, protocol.PlaceholderWorkerBase, workerBase())
	s = strings.ReplaceAll(s, protocol.PlaceholderWorkerEditor, resolveEditor())
	return s
}

func allLocalize(body protocol.ProcessSpec) protocol.ProcessSpec {
	body.Cmd = localize(body.Cmd)
	for i, v := range body.Args {
		body.Args[i] = localize(v)
	}
	body.Cwd = localize(body.Cwd)
	for i, v := range body.Env {
		body.Env[i] = localize(v)
	}
	return body
}

// exec: MsgExec(REQ). ProcessSpec을 받아 PTY로 process를 실행하고 pump를 기동한다. → ExecResponse
func (r *workerRouter) exec(req protocol.Frame) (any, error) {
	var body protocol.ProcessSpec
	if err := req.Bind(&body); err != nil {
		return protocol.ExecResponse{Accept: false, Reason: err.Error()}, nil
	}

	// placeholder 지역화: Cmd/Args/Cwd/Env 전부. Cmd는 {WORKER_EDITOR}(EDIT)도 치환된다.
	body = allLocalize(body)
	// log.Printf("[WORKER-EXEC] %v", body)

	// env는 완전 대체 정책(ExecInteractive nil=상속) → 베이스(os.Environ())를 여기서 조립해 넘긴다.
	// TERM은 TUI 에디터(vi 등)에 사실상 필수라 강제한다.
	env := append(append([]string{}, os.Environ()...), body.Env...)
	env = append(env, "TERM=xterm-256color")

	inter, err := pty.ExecInteractive(context.Background(), body.Cmd, env, body.Cwd, body.Args...)
	if err != nil {
		return protocol.ExecResponse{Accept: false, Reason: err.Error()}, nil
	}

	kind := body.Kind()
	var editPath string
	if kind == protocol.ExecTypeEdit && len(body.Args) > 0 {
		editPath = body.Args[0]
	}

	entry := &procEntry{uid: body.UID, kind: kind, editPath: editPath, inter: inter}
	if !r.procs.Append(body.UID, entry) {
		_ = inter.Kill()
		return protocol.ExecResponse{Accept: false, Reason: "중복되는 uid"}, nil
	}

	go r.pump(body.UID, entry)

	return protocol.ExecResponse{Accept: true}, nil
}

// pump는 한 process의 출력/상태를 드레인해 supervisor로 중계한다(supervisor bind.Relay 미러).
// 출력은 별도 goroutine, 상태는 호출된 goroutine 자신이 맡는다(둘 다 Done() 시 채널 close로 자연 종료).
func (r *workerRouter) pump(uid string, entry *procEntry) {
	go r.pumpOutput(uid, entry)
	r.pumpStatus(uid, entry)
}

// pumpOutput은 OutputAll()로 대기분을 배치 드레인해 MsgData(EVENT ↑)로 발행한다.
// 전송은 r.state.currentConn()(재접속 시 자동 교체)으로 한다 — 이 고루틴은 process 수명
// 내내(=워크라우터 세대를 넘어) 살아있으므로 r.conn(생성 시점 고정)을 쓰면 재접속 후에도
// 죽은 옛 conn만 계속 문다.
func (r *workerRouter) pumpOutput(uid string, entry *procEntry) {
	for {
		data, err := entry.inter.OutputAll()
		if err != nil {
			return // Done() 시 output 채널 close → 정상 종료
		}
		conn := r.state.currentConn()
		if conn == nil {
			continue // 끊긴 동안: 드레인은 계속하되 전송만 건너뜀(재접속하면 회복)
		}
		if err := conn.Emit(protocol.MsgData, protocol.DataEvent{UID: uid, Data: data}); err != nil {
			log.Printf("[worker pump] emit data uid=%s: %v", uid, err)
		}
	}
}

// pumpStatus는 Status()를 드레인해 MsgStatus(EVENT ↑)로 발행한다. PROCESS 전이 시 PID를 얹는다.
// 종료(IsCompleted)를 감지하면 teardown(EDIT read-back + procs.Remove)까지 여기서 맡는다(단일 지점).
func (r *workerRouter) pumpStatus(uid string, entry *procEntry) {
	for {
		st, code, err := entry.inter.Status()
		// log.Printf("[PROCESS.GO] pumpStatus %s", st.String())
		if err != nil {
			return
		}

		pid := 0
		if st == execute.CommandProcess {
			pid = entry.inter.Pid()
		}
		if conn := r.state.currentConn(); conn != nil {
			if err := conn.Emit(protocol.MsgStatus, protocol.StatusEvent{
				UID:      uid,
				Status:   st,
				PID:      pid,
				ExitCode: code,
			}); err != nil {
				log.Printf("[worker pump] emit status uid=%s: %v", uid, err)
			}
		}

		if st.IsCompleted() {
			r.teardown(uid, entry)
			return
		}
	}
}

// teardown은 process 종료 후 뒷정리다. EDIT면 편집 파일을 read-back해 MsgEditResult로 회수를
// 보고한다(저장 여부 판별=supervisor의 diff 책임, 여기선 무조건 읽어 보낸다). 그 뒤 procs에서 제거한다.
func (r *workerRouter) teardown(uid string, entry *procEntry) {
	if entry.kind == protocol.ExecTypeEdit {
		content, err := os.ReadFile(entry.editPath)
		if err != nil {
			log.Printf("[worker pump] edit read-back uid=%s: %v", uid, err)
		} else if conn := r.state.currentConn(); conn != nil {
			ctx, cancel := context.WithCancel(context.Background())
			if _, err := conn.Call(ctx, protocol.MsgEditResult, protocol.EditResult{UID: uid, Content: content}); err != nil {
				log.Printf("[worker pump] MsgEditResult uid=%s: %v", uid, err)
			}
			cancel()
		}
	}
	r.procs.Remove(uid)
}

// resize: MsgResize(REQ). 터미널 창 크기를 변경한다(Layout).
func (r *workerRouter) resize(req protocol.Frame) (any, error) {
	var body protocol.ResizeRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	entry, ok := r.procs.Get(body.UID)
	if !ok {
		return nil, fmt.Errorf("존재하지 않는 process: %s", body.UID)
	}
	if err := entry.inter.Layout(body.Cols, body.Rows); err != nil {
		return nil, fmt.Errorf("resize 실패: %w", err)
	}
	return nil, nil
}

// kill: MsgKill(REQ). 지정 UID의 process를 종료한다. procs에서 제거하지 않는다 — 종료 감지 후
// pumpStatus의 teardown이 단일 정리 경로를 맡는다(이중정리 방지).
func (r *workerRouter) kill(req protocol.Frame) (any, error) {
	var body protocol.KillRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	// log.Printf("[PROCESS.go] kill-req:%v", body)
	entry, ok := r.procs.Get(body.UID)
	if !ok {
		return nil, fmt.Errorf("존재하지 않는 process: %s", body.UID)
	}
	if err := entry.inter.Kill(); err != nil {
		return nil, err
	}
	return nil, nil
}

// input: MsgData(EVENT). 입력 바이트를 해당 process에 써넣는다(짝 응답 없음 — 없는 uid는 무시).
func (r *workerRouter) input(ev protocol.Frame) {
	var body protocol.DataEvent
	if err := ev.Bind(&body); err != nil {
		log.Printf("[worker input] bind: %v", err)
		return
	}
	entry, ok := r.procs.Get(body.UID)
	if !ok {
		return
	}
	if err := entry.inter.Write(body.Data); err != nil {
		log.Printf("[worker input] write uid=%s: %v", body.UID, err)
	}
}
