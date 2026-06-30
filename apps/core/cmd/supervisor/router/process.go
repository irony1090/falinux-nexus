package router

import (
	"context"
	"fmt"
	"nexus/internal/protocol"
)

func (r *supervisorRouter) ExecEdit(authKey string, spec protocol.ProcessSpec) error {
	conn, exist := r.workers.Get(authKey)
	if !exist {
		return fmt.Errorf("전송할 대상이 존재하지 않습니다: %s", authKey)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 정상 종료 시 ctx 자원 정리

	if spec.Type == protocol.ExecTypeEdit {

	}

	conn.Call(ctx, protocol.MsgExec, spec)

	return nil
}

// ===== process 실행 엔드포인트 (worker → supervisor, 수신) =====

// output: MsgData(EVENT). worker가 보낸 process 출력 바이트를 받는다(이후 fan-out/적재).
func (r *supervisorRouter) output(ev protocol.Frame) {
}

// status: MsgStatus(EVENT). worker가 보낸 상태 변화(RUNNING+PID / 종료+ExitCode)를 받는다.
func (r *supervisorRouter) status(ev protocol.Frame) {
}
