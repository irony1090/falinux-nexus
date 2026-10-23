package router

import (
	"errors"
	"fmt"
	"log"
	"nexus/cmd/worker/constants"
	"nexus/internal/execute/hold"
	"nexus/internal/protocol"
	"os"
	"path/filepath"
	"strings"
)

// ===== process 실행 엔드포인트 (supervisor → worker) =====

// procEntry는 이 worker가 붙어 있는 process 하나다. PTY는 보관 프로세스가 쥐고 h로 붙는다(REF-process-keep.md).
// kind/editPath는 종료 시 EDIT read-back 여부를 finish가 판별하는 데 쓴다
type procEntry struct {
	uid      string
	kind     protocol.ExecType
	editPath string
	pid      int // 스크립트 pid
	h        *hold.Client
}

// workerBase는 이 worker의 실행/저장 루트다({WORKER_BASE} 치환 값). exec·resolveDest가
// 같은 값을 써야 선배치 위치와 실행 대상 경로가 일치한다.
func workerBase() string { return constants.GetEnv().ProcessRoot }

// resolveEditor는 {WORKER_EDITOR} 치환 값을 정한다: $VISUAL > EDITOR(env) > OS 기본값(defaultEditor).
// (편집기 선택은 worker 책임 — supervisor는 토큰만 보낸다.) 인자 분리는 prepareEdit에서.
func resolveEditor() string {
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	if v := constants.GetEnv().Editor; v != "" {
		return v
	}
	return defaultEditor()
}

// prepareEdit는 에디터 값을 공백으로 나눠 Cmd·Args로 펼친다("code --wait" -> code, [--wait, 파일]).
// 공백이 든 에디터 경로는 지원하지 않는다(PATH에 넣도록 안내) → REF-windows-worker.md WW3(에디터 기본값)
func prepareEdit(body protocol.ProcessSpec) (protocol.ProcessSpec, string, error) {
	fields := strings.Fields(body.Cmd)
	if len(fields) == 0 {
		return body, "", errors.New("에디터를 찾을 수 없습니다 — worker env에 EDITOR를 설정하세요")
	}
	var editPath string
	if len(body.Args) > 0 {
		body.Args[0] = filepath.Clean(body.Args[0])
		editPath = body.Args[0]
	}
	body.Cmd = fields[0]
	body.Args = append(fields[1:], body.Args...)
	return body, editPath, nil
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

// exec: MsgExec(REQ). ProcessSpec을 받아 보관 프로세스로 실행하고 forward를 기동한다. → ExecResponse
func (r *workerRouter) exec(req protocol.Frame) (any, error) {
	var body protocol.ProcessSpec
	if err := req.Bind(&body); err != nil {
		return protocol.ExecResponse{Accept: false, Reason: err.Error()}, nil
	}

	// placeholder 지역화: Cmd/Args/Cwd/Env 전부. Cmd는 {WORKER_EDITOR}(EDIT)도 치환된다.
	body = allLocalize(body)
	// log.Printf("[WORKER-EXEC] %v", body)
	kind := body.Kind()
	var editPath string
	var err error
	if kind == protocol.ExecTypeEdit {
		body, editPath, err = prepareEdit(body)
	} else if body.Direct {
		body, err = prepareDirect(body)
	} else {
		body, err = prepareExec(body)
	}
	if err != nil {
		return protocol.ExecResponse{Accept: false, Reason: err.Error()}, nil
	}

	// env는 완전 대체 정책 → 베이스(os.Environ())를 여기서 조립해 넘긴다.
	// TERM은 TUI 에디터(vi 등)에 사실상 필수라 강제한다.
	env := append(append([]string{}, os.Environ()...), body.Env...)
	env = append(env, "TERM=xterm-256color")

	spec := hold.Spec{UID: body.UID, Cmd: body.Cmd, Args: body.Args, Env: env, Cwd: body.Cwd}
	entry, err := r.state.spawn(spec, kind, editPath)
	if err != nil {
		return protocol.ExecResponse{Accept: false, Reason: err.Error()}, nil
	}
	go r.state.forward(entry)

	return protocol.ExecResponse{Accept: true}, nil
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
	if err := entry.h.Layout(body.Cols, body.Rows); err != nil {
		return nil, fmt.Errorf("resize 실패: %w", err)
	}
	return nil, nil
}

// kill: MsgKill(REQ). 지정 UID의 process를 종료한다. procs에서 제거하지 않는다 — 종료 감지 후
// forward의 finish가 단일 정리 경로를 맡는다(이중정리 방지).
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
	if err := entry.h.Kill(body.Force); err != nil {
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
	if err := entry.h.Write(body.Data); err != nil {
		log.Printf("[worker input] write uid=%s: %v", body.UID, err)
	}
}
