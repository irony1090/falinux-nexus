package router

import "nexus/internal/protocol"

// ===== process 실행 엔드포인트 (supervisor → worker) =====

// exec: MsgExec(REQ). ProcessSpec을 받아 PTY로 process를 실행한다. → ExecResponse
func (r *workerRouter) exec(req protocol.Frame) (any, error) {
	return nil, nil
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
