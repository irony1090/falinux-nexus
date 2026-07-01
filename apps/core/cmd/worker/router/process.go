package router

import (
	"nexus/cmd/worker/constants"
	"nexus/internal/protocol"
	"os"
	"strings"
)

// ===== process 실행 엔드포인트 (supervisor → worker) =====

// workerBase는 이 worker의 실행/저장 루트다({WORKER_BASE} 치환 값). exec·resolveDest가
// 같은 값을 써야 선배치 위치와 실행 대상 경로가 일치한다.
func workerBase() string { return constants.GetEnv().ProcessRoot }

// resolveEditor는 {WORKER_EDITOR} 치환 값을 정한다: $VISUAL > $EDITOR > "vi" 폴백.
// (편집기 선택은 worker 책임 — supervisor는 토큰만 보낸다.)
func resolveEditor() string {
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	if v := os.Getenv("EDITOR"); v != "" {
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

// exec: MsgExec(REQ). ProcessSpec을 받아 PTY로 process를 실행한다. → ExecResponse
func (r *workerRouter) exec(req protocol.Frame) (any, error) {
	var body protocol.ProcessSpec
	if err := req.Bind(&body); err != nil {
		return protocol.ExecResponse{Accept: false, Reason: err.Error()}, nil
	}

	// placeholder 지역화: Cmd/Args/Cwd/Env 전부. Cmd는 {WORKER_EDITOR}(EDIT)도 치환된다.
	body.Cmd = localize(body.Cmd)
	for i, v := range body.Args {
		body.Args[i] = localize(v)
	}
	body.Cwd = localize(body.Cwd)
	for i, v := range body.Env {
		body.Env[i] = localize(v)
	}

	// TODO(도메인 배선): 치환된 spec으로 PTY 실행(execute/pty) → PID를 MsgStatus(PROCESS)로 보고,
	//   출력 MsgData(EVENT) 스트리밍, 종료 시 MsgStatus(COMPLETED/FAILED). EDIT는 종료 후
	//   Args[0] 파일 read-back → MsgEditResult. $TERM 세팅·TUI 에디터 강제는 실행부에서.
	return protocol.ExecResponse{Accept: true}, nil
}

// resize: MsgResize(REQ). 터미널 창 크기를 변경한다(Layout).
func (r *workerRouter) resize(req protocol.Frame) (any, error) {

	return nil, nil
}

// kill: MsgKill(REQ). 지정 UID의 process를 종료한다.
func (r *workerRouter) kill(req protocol.Frame) (any, error) {
	return nil, nil
}

// input: MsgData(EVENT). 입력 바이트를 해당 process에 써넣는다(짝 응답 없음).
func (r *workerRouter) input(ev protocol.Frame) {
}
