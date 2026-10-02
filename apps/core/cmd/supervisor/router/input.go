package router

import (
	"log"

	"nexus/internal/execute"
	"nexus/internal/protocol"
	"nexus/internal/transport"
)

// onProcessInput은 브라우저 소켓에 키 입력 핸들러를 건다(Serve 전에 호출). userID = 소켓 연결 시점의 계정.
func (r *supervisorRouter) onProcessInput(conn *transport.Conn, userID int64) {
	conn.On(protocol.MsgProcessInput, func(ev protocol.Frame) {
		var body protocol.DataEvent
		if err := ev.Bind(&body); err != nil {
			return
		}
		r.input(userID, body)
	})
}

// input: 버림 조건은 조용히 넘긴다(키마다 로그가 쌓이므로) — REF-process-input.md I2(입력 권한)·I3(PENDING 중 입력)
func (r *supervisorRouter) input(userID int64, body protocol.DataEvent) {
	entry, ok := r.processManager.Get(body.UID)
	if !ok || entry.Inter == nil || entry.Record == nil {
		return // 종료됨 / worker 끊김(Detach로 memory에서 빠짐) / folder
	}
	if entry.Record.OwnerUserID != userID {
		log.Printf("[process] 다른 계정의 입력 버림 uid=%s user=%d", body.UID, userID)
		return
	}
	if entry.Record.Status != execute.CommandProcess.String() {
		return // 실행 직후 RUNNING 전 PENDING 포함
	}
	if err := entry.Inter.Write(body.Data); err != nil {
		log.Printf("[process] 입력 전달 실패 uid=%s: %v", body.UID, err)
	}
}
