package process

import (
	"nexus/internal/execute"
	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
)

// ProcessEntry는 memory 레지스트리(KeyValManager)의 값이다.
// DB 스냅샷(Record)과 런타임 I/O 핸들(Inter)을 한데 묶는다.
//
//	folder-open : worker에서 아무것도 실행되지 않음 → Inter=nil (레코드-only, frontend presence).
//	script EXEC : 실 프로세스 → Inter 有, pool 영속.
//	script EDIT : vi PTY      → Inter 有, pool 영속(출력 sink만 off), 종료 시 read-back.
type ProcessEntry struct {
	Record *superdb.Process          // 상태/영속 스냅샷 (folder도 메모리엔 존재)
	Inter  *execute.AgentInteractive // I/O 라우팅 핸들. folder면 nil
}

// HasProcess는 worker에 실제 프로세스가 물려 있는지(=Inter 존재) 반환한다.
// worker 끊김 정리(Done(502))는 이 엔트리들만 대상으로 한다(folder는 제외 — frontend 끊김 기준).
func (e *ProcessEntry) HasProcess() bool {
	return e.Inter != nil
}

// NodeID는 기원 노드 id를 반환한다(EDIT read-back 대상 조회 등). 없으면 0.
func (e *ProcessEntry) NodeID() int64 {
	if e.Record == nil || !e.Record.NodeID.Valid {
		return 0
	}
	return e.Record.NodeID.Int64
}

// Spec은 Record로부터 MsgExec용 ProcessSpec을 복원한다. 매니저가 내부 작성해 영속한 값
// 그대로이며(UID 포함), router가 worker에 MsgExec를 보낼 때 쓴다.
func (e *ProcessEntry) Spec() protocol.ProcessSpec {
	r := e.Record
	return protocol.ProcessSpec{
		UID:  r.Uid,
		Type: protocol.ExecType(r.Type),
		Cmd:  r.Cmd,
		Args: r.Args,
		Cwd:  r.Cwd,
		Env:  r.Env,
		Rows: uint16(r.Rows),
		Cols: uint16(r.Cols),
	}
}
