package process

import (
	"sync"

	"nexus/internal/execute"
	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
)

// Subscriber는 이 process를 보고 있는 브라우저 세션 하나를 가리킨다.
// DB(process_subscribers) row를 그대로 미러링하지 않는다 — process_uid는 엔트리 자신의
// 스코프라 중복이고 created_at은 DB 감사 목적이라 런타임 라우팅엔 불필요(2026-07-15 확정,
// REF-process-reconnect.md "ProcessEntry 인메모리 구독자 필드").
type Subscriber struct {
	Sid         string
	OwnerUserID int64
}

// ProcessEntry는 memory 레지스트리(KeyValManager)의 값이다.
// DB 스냅샷(Record)과 런타임 I/O 핸들(Inter)을 한데 묶는다.
//
//	folder-open : worker에서 아무것도 실행되지 않음 → Inter=nil (레코드-only, frontend presence).
//	script EXEC : 실 프로세스 → Inter 有, pool 영속.
//	script EDIT : vi PTY      → Inter 有, pool 영속(출력 sink만 off), 종료 시 read-back.
type ProcessEntry struct {
	Record *superdb.Process          // 상태/영속 스냅샷 (folder도 메모리엔 존재)
	Inter  *execute.AgentInteractive // I/O 라우팅 핸들. folder면 nil

	recordMu sync.Mutex // SetRecord 동시 교체 직렬화(쓰기만 보호 — 읽기는 entry.Record.X 직접 접근 유지)

	subMu       sync.RWMutex
	subscribers []Subscriber
}

// SetRecord는 DB round-trip(RETURNING 등)으로 얻은 최신 스냅샷으로 Record를 통째로 교체한다.
// worker 보고·끊김 합성·재접속 등 여러 경로가 겹쳐 들어와도 교체끼리 순서가 뒤섞이지 않도록
// 직렬화한다(REF-process-wiring.md "memory record 동기화"). entry.Record 직접 대입은 이 메서드로
// 통일한다 — 단, 최초 생성 시점(Append 이전, 아직 아무도 못 보는 entry)의 구성은 예외.
func (e *ProcessEntry) SetRecord(rec *superdb.Process) {
	e.recordMu.Lock()
	defer e.recordMu.Unlock()
	e.Record = rec
}

// IsSubscribed는 sid가 이 process를 구독 중인지 반환한다.
func (e *ProcessEntry) IsSubscribed(sid string) bool {
	e.subMu.RLock()
	defer e.subMu.RUnlock()
	for _, s := range e.subscribers {
		if s.Sid == sid {
			return true
		}
	}
	return false
}

// AddSubscriber는 s를 구독자 목록에 더한다. 이미 같은 sid가 있으면 중복 추가하지 않는다
// (DB 쪽 ON CONFLICT DO NOTHING과 동일 시맨틱).
func (e *ProcessEntry) AddSubscriber(s Subscriber) {
	e.subMu.Lock()
	defer e.subMu.Unlock()
	for _, existing := range e.subscribers {
		if existing.Sid == s.Sid {
			return
		}
	}
	e.subscribers = append(e.subscribers, s)
}

// RemoveSubscriber는 sid를 구독자 목록에서 제거한다. 없으면 no-op.
func (e *ProcessEntry) RemoveSubscriber(sid string) {
	e.subMu.Lock()
	defer e.subMu.Unlock()
	for i, s := range e.subscribers {
		if s.Sid == sid {
			e.subscribers = append(e.subscribers[:i], e.subscribers[i+1:]...)
			return
		}
	}
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
