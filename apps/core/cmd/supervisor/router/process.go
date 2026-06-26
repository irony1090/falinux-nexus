package router

import (
	"fmt"
	"nexus/internal/protocol"
)

func (r *supervisorRouter) Exec(authKey string, spec protocol.ProcessSpec) error {

	if _, exist := r.workers.Get(authKey); !exist {
		return fmt.Errorf("전송할 대상이 존재하지 않습니다: %s", authKey)
	}

	return nil
}

// ===== process 실행 엔드포인트 (worker → supervisor, 수신) =====

// output: MsgData(EVENT). worker가 보낸 process 출력 바이트를 받는다(이후 fan-out/적재).
func (r *supervisorRouter) output(ev protocol.Frame) {
}

// status: MsgStatus(EVENT). worker가 보낸 상태 변화(RUNNING+PID / 종료+ExitCode)를 받는다.
func (r *supervisorRouter) status(ev protocol.Frame) {
}
