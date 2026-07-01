package process

import (
	"context"
	"fmt"
	"nexus/internal/execute"
	"nexus/internal/manager"
	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/transport"
	"nexus/internal/util"
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
	return p.memory.Get(uid)
}

// Exec은 node를 실행 상태로 등록한다. spec은 파라미터로 받지 않고 node로부터 내부 작성한다.
//   - FOLDER : worker에서 아무것도 실행되지 않음 → memory 레코드만(Inter=nil, pool 미저장).
//   - SCRIPT : 실 프로세스 → memory + pool + AgentInteractive(EXEC).
//
// UID는 매니저가 발급해 entry.Record.Uid로 authoritative하다. 실제 MsgExec 전송은
// 호출자(router)가 entry.Spec()으로 ProcessSpec을 얻어 worker에 전송한다.
func (p *ProcessManager) Exec(
	worker *transport.Conn,
	workerKey string,
	node superdb.Node,
	owner superdb.User,
) (*ProcessEntry, error) {
	if node.Kind == superdb.NodeKindFolder {
		return p.openFolder(workerKey, node, owner)
	}
	// Cmd/Args는 파라미터가 아니라 execScript가 node로부터 WorkerNodePath로 내부작성한다
	// (router 선배치 DestPath와 단일 출처 공유). content는 router가 파일로 먼저 깔아 둔다.
	return p.execScript(worker, workerKey, protocol.ExecTypeExec, node, owner)
}

// ExecEdit은 SCRIPT 노드를 편집 세션(EDIT)으로 등록한다. EXEC과 동일하게 memory+pool+Inter를
// 만들되 Type=EDIT다. 실제 흐름(현재 content seed 전송=SendBuffer, MsgExec 전송, 종료 시
// MsgEditResult read-back → nodes.content UPDATE)은 호출자(router)가 비동기로 잇는다.
// 즉 이 함수는 content를 직접 수정하지 않고 "편집 세션을 개시"할 준비만 한다.
func (p *ProcessManager) ExecEdit(
	worker *transport.Conn,
	workerKey string,
	node superdb.Node,
	owner superdb.User,
) (*ProcessEntry, error) {
	if node.Kind != superdb.NodeKindScript {
		return nil, fmt.Errorf("수정할 수 없는 대상입니다")
	}
	// EDIT도 Cmd/Args를 execScript가 내부작성한다: Cmd={WORKER_EDITOR}, Args=[WorkerNodePath].
	// 종료 시 router가 MsgEditResult를 받아 nodes.content를 UPDATE한다(이 함수는 개시만).
	return p.execScript(worker, workerKey, protocol.ExecTypeEdit, node, owner)
}

// openFolder는 folder-open을 memory 전용 레코드로 등록한다(worker 무접촉).
func (p *ProcessManager) openFolder(workerKey string, node superdb.Node, owner superdb.User) (*ProcessEntry, error) {
	uid, err := p.createKey()
	if err != nil {
		return nil, err
	}

	prc := &superdb.Process{
		Uid:         uid,
		Type:        string(protocol.ExecTypeExec),
		DeviceKey:   workerKey,
		OwnerUserID: owner.ID,
	}
	prc.NodeID.Valid, prc.NodeID.Int64 = true, node.ID
	prc.Pid.Valid, prc.Pid.Int32 = false, -1

	entry := &ProcessEntry{Record: prc, Inter: nil}
	if !p.memory.Append(uid, entry) {
		return nil, fmt.Errorf("중복되는 키가 존재합니다: %s", uid)
	}
	return entry, nil
}

// execScript는 SCRIPT의 EXEC/EDIT 공통 경로다. UID 발급 → AgentInteractive 생성 →
// memory 등록 → pool 영속(재시작 복구·프론트 목록). pool 실패 시 memory 롤백.
func (p *ProcessManager) execScript(
	worker *transport.Conn,
	workerKey string,
	kind protocol.ExecType,
	node superdb.Node,
	owner superdb.User,
) (*ProcessEntry, error) {
	if worker == nil {
		return nil, fmt.Errorf("전송할 대상이 존재하지 않습니다: %s", workerKey)
	}

	uid, err := p.createKey()
	if err != nil {
		return nil, err
	}

	inter := newWorkerInteractive(worker, uid)
	entry := &ProcessEntry{Inter: inter}

	if !p.memory.Append(uid, entry) {
		return nil, fmt.Errorf("중복되는 키가 존재합니다: %s", uid)
	}

	// spec 필드는 node로부터 내부 작성한다(파라미터 아님). content 자체는 router가 파일로
	// 선배치(WorkerNodePath)하므로 여기선 그 경로만 가리킨다. 초기 Rows/Cols는 요청시점
	// 정보라 0으로 두고 attach 후 MsgResize로 맞춘다(resize-on-attach).
	//   EXEC: 파일을 직접 실행 → Cmd=WorkerNodePath, Args=[].
	//   EDIT: 에디터로 그 파일을 연다 → Cmd={WORKER_EDITOR}(worker가 $VISUAL>$EDITOR>vi 치환),
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
		OwnerUserID: owner.ID,
		DeviceKey:   workerKey,
		Cmd:         cmd,
		Args:        args,
		Env:         []string{},
		Cwd:         "",
		Rows:        0,
		Cols:        0,
	}
	params.NodeID.Valid, params.NodeID.Int64 = true, node.ID

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rec, err := p.pool.Queries().CreateProcess(ctx, params)
	if err != nil {
		p.memory.Remove(uid) // 영속 실패 → 메모리 등록 롤백(고아 엔트리 방지)
		return nil, err
	}
	entry.Record = &rec
	return entry, nil
}

// newWorkerInteractive는 worker conn 위에 원격 process 프록시를 만든다. 콜백은 uid를
// 클로저로 잡아 이 process를 가리킨다: 입력=EVENT(MsgData), 리사이즈/킬=REQ(MsgResize/MsgKill).
func newWorkerInteractive(worker *transport.Conn, uid string) *execute.AgentInteractive {
	return execute.NewAgentInteractive(
		func(data []byte) error { // onWrite: 입력 키스트로크
			return worker.Emit(protocol.MsgData, protocol.DataEvent{UID: uid, Data: data})
		},
		func(cols, rows uint16) { // onLayout
			_, _ = worker.Call(context.Background(), protocol.MsgResize, protocol.ResizeRequest{UID: uid, Rows: rows, Cols: cols})
		},
		func() error { // onKill
			_, err := worker.Call(context.Background(), protocol.MsgKill, protocol.KillRequest{UID: uid})
			return err
		},
	)
}

// Remove는 실행 엔트리를 정리한다. Inter가 있으면 Done(502)로 안전망을 건다(이미 Done이면
// sync.Once로 no-op) — bind.Relay 드레인 고루틴을 확실히 해제한다. pool 행은 남긴다(이력).
func (p *ProcessManager) Remove(uid string) {
	if entry, ok := p.memory.Get(uid); ok && entry.Inter != nil {
		entry.Inter.Done(502)
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
