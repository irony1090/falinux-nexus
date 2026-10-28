package process

import (
	"context"
	"fmt"
	"log"
	"nexus/internal/execute"
	"nexus/internal/manager"
	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/transport"
	"nexus/internal/util"

	"github.com/jackc/pgx/v5/pgtype"
)

type ProcessManager struct {
	memory *manager.KeyValManager[string, *ProcessEntry]
	pool   *store.StorePool
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		memory: manager.NewKeyValManager[string, *ProcessEntry](),
		pool:   store.GetStorePool(),
	}
}

// Get은 uid로 실행 엔트리를 조회한다. worker→sup의 MsgData/MsgStatus 핸들러가
// 해당 process의 AgentInteractive를 되찾을 때 쓴다.
func (p *ProcessManager) Get(uid string) (*ProcessEntry, bool) {
	// log.Printf("[MANAGER] memory.Get")
	// p.memory.FindAll(func(s string, pe *ProcessEntry) bool {
	// 	log.Printf("[MANAGER] manager.items: %s", s)
	// 	return false
	// })
	return p.memory.Get(uid)
}

// Exec은 SCRIPT node를 실행 상태로 등록한다(memory + pool + AgentInteractive(EXEC)). spec은 node로부터 내부 작성한다.
//
// UID는 매니저가 발급해 entry.Record.Uid로 authoritative하다. 실제 MsgExec 전송은
// 호출자(router)가 entry.Spec()으로 ProcessSpec을 얻어 worker에 전송한다.
func (p *ProcessManager) Exec(
	worker *transport.Conn,
	workerKey string,
	node superdb.Node,
	ownerID, startedBy int64,
) (*ProcessEntry, error) {
	if worker == nil {
		return nil, fmt.Errorf("전송할 대상이 존재하지 않습니다: %s", workerKey)
	}
	if node.Kind != superdb.NodeKindScript {
		return nil, fmt.Errorf("실행할 수 없는 대상입니다")
	}

	// Cmd/Args는 파라미터가 아니라 execScript가 node로부터 WorkerNodePath로 내부작성한다
	// (router 선배치 DestPath와 단일 출처 공유). content는 router가 파일로 먼저 깔아 둔다.
	return p.execScript(worker, workerKey, protocol.ExecTypeExec, node, ownerID, startedBy)
}

// ExecEdit은 SCRIPT 노드를 편집 세션(EDIT)으로 등록한다. EXEC과 동일하게 memory+pool+Inter를
// 만들되 Type=EDIT다. 실제 흐름(현재 content seed 전송=SendBuffer, MsgExec 전송, 종료 시
// MsgEditResult read-back → nodes.content UPDATE)은 호출자(router)가 비동기로 잇는다.
// 즉 이 함수는 content를 직접 수정하지 않고 "편집 세션을 개시"할 준비만 한다.
func (p *ProcessManager) ExecEdit(
	worker *transport.Conn,
	workerKey string,
	node superdb.Node,
	ownerID, startedBy int64,
) (*ProcessEntry, error) {
	if worker == nil {
		return nil, fmt.Errorf("전송할 대상이 존재하지 않습니다: %s", workerKey)
	}
	if node.Kind != superdb.NodeKindScript {
		return nil, fmt.Errorf("수정할 수 없는 대상입니다")
	}
	// EDIT도 Cmd/Args를 execScript가 내부작성한다: Cmd={WORKER_EDITOR}, Args=[WorkerNodePath].
	// 종료 시 router가 MsgEditResult를 받아 nodes.content를 UPDATE한다(이 함수는 개시만).
	return p.execScript(worker, workerKey, protocol.ExecTypeEdit, node, ownerID, startedBy)
}

// execScript는 SCRIPT의 EXEC/EDIT 공통 경로다. UID 발급 → 명세 작성 → register.
// ownerID = 노드 소유자, startedBy = 실행한 계정 — SH9(받은 쪽이 실행한 process의 소유)
func (p *ProcessManager) execScript(
	worker *transport.Conn,
	workerKey string,
	kind protocol.ExecType,
	node superdb.Node,
	ownerID, startedBy int64,
) (*ProcessEntry, error) {
	if worker == nil {
		return nil, fmt.Errorf("전송할 대상이 존재하지 않습니다: %s", workerKey)
	}

	uid, err := p.createKey()
	if err != nil {
		return nil, err
	}
	// spec 필드는 node로부터 내부 작성한다(파라미터 아님). content 자체는 router가 파일로
	// 선배치(WorkerNodePath)하므로 여기선 그 경로만 가리킨다. 초기 Rows/Cols는 요청시점
	// 정보라 0으로 두고 attach 후 MsgResize로 맞춘다(resize-on-attach).
	//   EXEC: 파일을 직접 실행 → Cmd=WorkerNodePath, Args=[].
	//   EDIT: 에디터로 그 파일을 연다 → Cmd={WORKER_EDITOR}(worker가 $VISUAL>$EDITOR>OS 기본값 치환),
	//         Args=[WorkerNodePath]. 종료 시 read-back은 router가 처리.
	// 경로는 node+process로 조립한다. Record는 아직 없으나 uid는 발급됐으므로 그걸로 만든다
	// (persist 후 entry.Record.Uid와 동일 → router의 선배치 DestPath와 정확히 일치).
	nodePath := WorkerNodePath(node, superdb.Process{Uid: uid})
	var cmd string
	var args []string
	if kind == protocol.ExecTypeEdit {
		cmd = protocol.PlaceholderWorkerEditor
		args = []string{nodePath}
	} else {
		cmd = nodePath
		args = []string{}
	}

	params := superdb.CreateProcessParams{
		Uid:         uid,
		Type:        string(kind),
		OwnerUserID: ownerID,
		DeviceKey:   workerKey,
		Cmd:         cmd,
		Args:        args,
		Env:         []string{},
		Cwd:         "",
		Rows:        0,
		Cols:        0,
		StartedBy:   startedBy,
	}
	params.NodeID.Valid, params.NodeID.Int64 = true, node.ID
	return p.register(worker, params)
}

// ExecFile은 장비 디렉토리의 파일을 그대로 실행하도록 등록한다(폴더-4(실행)) — 노드 없음, 본문 선배치 없음.
// cmd = 파일 전체 경로(편집 = 편집기 토큰 + args[0] = 파일), cwd = 보고 있던 폴더, dirNodeID = 장비 타일을 연 폴더(Q15(장비 디렉토리 id)). 소유·실행 = 같은 계정(Q5)
func (p *ProcessManager) ExecFile(worker *transport.Conn, workerKey, cmd string, args []string, cwd string, dirNodeID, userID int64) (*ProcessEntry, error) {
	if worker == nil {
		return nil, fmt.Errorf("전송할 대상이 존재하지 않습니다: %s", workerKey)
	}
	uid, err := p.createKey()
	if err != nil {
		return nil, err
	}
	params := superdb.CreateProcessParams{
		Uid:         uid,
		Type:        string(protocol.ExecTypeExec),
		OwnerUserID: userID,
		DeviceKey:   workerKey,
		Cmd:         cmd,
		Args:        append([]string{}, args...),
		Env:         []string{},
		Cwd:         cwd,
		StartedBy:   userID,
		DirNodeID:   pgtype.Int8{Int64: dirNodeID, Valid: true},
	}
	return p.register(worker, params)
}

// Adopt는 worker에서 이미 돌고 있는 미등록 process를 이 supervisor에 등록한다(가져오기-3(등록 API)) — uid는 worker가 쓰는 값 그대로.
// nodeID가 있으면 그 스크립트 실행으로, 없으면 장비 파일 실행처럼 dirNodeID(장비 타일 폴더)에 붙인다. 소유·실행 = 가져온 계정(장비 주인)
func (p *ProcessManager) Adopt(worker *transport.Conn, workerKey, uid, cmd, cwd string, nodeID, dirNodeID pgtype.Int8, userID int64) (*ProcessEntry, error) {
	if worker == nil {
		return nil, fmt.Errorf("전송할 대상이 존재하지 않습니다: %s", workerKey)
	}
	params := superdb.CreateProcessParams{
		Uid:         uid,
		Type:        string(protocol.ExecTypeExec),
		OwnerUserID: userID,
		NodeID:      nodeID,
		DeviceKey:   workerKey,
		Cmd:         cmd,
		Args:        []string{},
		Env:         []string{},
		Cwd:         cwd,
		StartedBy:   userID,
		DirNodeID:   dirNodeID,
	}
	return p.register(worker, params)
}

// register: memory 등록 -> pool 영속(재시작 복구·프론트 목록). pool 실패 시 memory 롤백
func (p *ProcessManager) register(worker *transport.Conn, params superdb.CreateProcessParams) (*ProcessEntry, error) {
	entry := &ProcessEntry{Inter: newWorkerInteractive(worker, params.Uid)}
	if !p.memory.Append(params.Uid, entry) {
		return nil, fmt.Errorf("중복되는 키가 존재합니다: %s", params.Uid)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rec, err := p.pool.Queries().CreateProcess(ctx, params)
	if err != nil {
		p.memory.Remove(params.Uid) // 영속 실패 → 메모리 등록 롤백(고아 엔트리 방지)
		return nil, err
	}
	entry.SetRecord(&rec)
	return entry, nil
}

// Rebind는 재접속한 worker 위에 uid의 process를 재바인딩한다("폐기 후 재생성" —
// REF-process 2026-07-14). DB Record는 그대로 두고(진실의 출처), 죽은 conn을 캡처했던 옛
// AgentInteractive 대신 새 conn 위에 새 AgentInteractive를 만들어 memory에 재등록한다.
// 끊김 처리(applyStatus의 CommandPending 분기)가 이미 memory에서 Detach했으므로 충돌 없이
// Append된다.
func (p *ProcessManager) Rebind(uid string, worker *transport.Conn) (*ProcessEntry, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rec, err := p.pool.Queries().GetProcess(ctx, uid)
	if err != nil {
		return nil, err
	}

	inter := newWorkerInteractive(worker, uid)
	entry := &ProcessEntry{Record: &rec, Inter: inter}
	if !p.memory.Append(uid, entry) {
		return nil, fmt.Errorf("이미 재바인딩된 process: %s", uid)
	}
	return entry, nil
}

// newWorkerInteractive는 worker conn 위에 원격 process 프록시를 만든다. 콜백은 uid를
// 클로저로 잡아 이 process를 가리킨다: 입력=EVENT(MsgData), 리사이즈/킬=REQ(MsgResize/MsgKill).
func newWorkerInteractive(worker *transport.Conn, uid string) *execute.AgentInteractive {
	return execute.NewAgentInteractive(
		func(data []byte) error { // onWrite: 입력 키스트로크
			// log.Printf("NEW_INTER ARGS1")
			return worker.Emit(protocol.MsgData, protocol.DataEvent{UID: uid, Data: data})
		},
		func(cols, rows uint16) error { // onLayout
			_, err := worker.Call(context.Background(), protocol.MsgResize, protocol.ResizeRequest{UID: uid, Rows: rows, Cols: cols})
			return err
		},
		func(force bool) error { // onKill
			_, err := worker.Call(context.Background(), protocol.MsgKill, protocol.KillRequest{UID: uid, Force: force})
			return err
		},
	)
}

// ListVisibleLive는 계정이 볼 수 있는(소유 + 공유받은) 살아 있는(PENDING/PROCESS) process를 단계와 함께 실행 순서대로 반환한다(DB — PENDING은 memory에 없어서).
func (p *ProcessManager) ListVisibleLive(userID int64) ([]superdb.ListVisibleLiveProcessesRow, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return p.pool.Queries().ListVisibleLiveProcesses(ctx, userID)
}

// Detach는 worker 끊김(PENDING) 전용 정리다: 종료 STATUS 없이 relay만 끝내고 memory에서 뺀다.
// Remove를 쓰면 Done(502) 안전망이 FAILED를 발행해 브라우저가 종료로 본다 — REF-process-sync.md X(끊김 표시)
func (p *ProcessManager) Detach(uid string) {
	if entry, ok := p.memory.Get(uid); ok && entry.Inter != nil {
		entry.Inter.Detach()
	}
	p.memory.Remove(uid)
}

// Remove는 실행 엔트리를 정리한다. Inter가 있으면 Done(502)로 안전망을 건다(이미 Done이면
// sync.Once로 no-op) — bind.Relay 드레인 고루틴을 확실히 해제한다. pool 행은 남긴다(이력).
func (p *ProcessManager) Remove(uid string) {
	log.Printf("[MANAGER.go] Remove: %s", uid)
	if entry, ok := p.memory.Get(uid); ok && entry.Inter != nil {
		entry.Inter.Done(502)
	} else {

		log.Printf("[MANAGER.go] Remove entry: %v", entry)
	}

	p.memory.Remove(uid)
}

func (p *ProcessManager) createKey() (string, error) {
	newKey, err := util.RandomKey(16, "", "")
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := p.pool.Queries().GetProcess(ctx, newKey); err == nil {
		return "", fmt.Errorf("중복되는 키가 존재합니다")
	} else if _, exist := p.memory.Get(newKey); exist {
		return "", fmt.Errorf("중복되는 키가 존재합니다")
	}

	return newKey, nil
}
