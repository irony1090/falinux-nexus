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

func (p *ProcessManager) Test(predicate func(string, *ProcessEntry) bool) []manager.FindResult[string, *ProcessEntry] {
	return p.memory.FindAll(predicate)
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

	if worker == nil {
		return nil, fmt.Errorf("전송할 대상이 존재하지 않습니다: %s", workerKey)
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
	if worker == nil {
		return nil, fmt.Errorf("전송할 대상이 존재하지 않습니다: %s", workerKey)
	}
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
	entry.SetRecord(&rec)
	return entry, nil
}

// Rebind는 재접속한 worker 위에 uid의 process를 재바인딩한다("폐기 후 재생성" —
// REF-process 2026-07-14). DB Record는 그대로 두고(진실의 출처), 죽은 conn을 캡처했던 옛
// AgentInteractive 대신 새 conn 위에 새 AgentInteractive를 만들어 memory에 재등록한다.
// 끊김 처리(applyStatus의 CommandPending 분기)가 이미 memory에서 Remove했으므로 충돌 없이
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
		func() error { // onKill
			// log.Printf("NEW_INTER ARGS3")
			_, err := worker.Call(context.Background(), protocol.MsgKill, protocol.KillRequest{UID: uid})
			return err
		},
	)
}

// SubscribeProcess는 sid를 uid의 구독자로 등록한다. entry.AddSubscriber(memory, 저렴·되돌리기
// 쉬움)를 먼저 반영하고 나서 pool에 write-through한다(execScript와 동일 순서) — DB insert가
// 실패하면 memory를 롤백해 두 저장소가 어긋나지 않게 한다.
func (p *ProcessManager) SubscribeProcess(uid string, sub Subscriber) error {
	entry, ok := p.memory.Get(uid)
	if !ok {
		return fmt.Errorf("존재하지 않는 process입니다: %s", uid)
	}
	entry.AddSubscriber(sub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := p.pool.Queries().CreateProcessSubscriber(ctx, superdb.CreateProcessSubscriberParams{
		ProcessUid:  uid,
		OwnerUserID: sub.OwnerUserID,
		Sid:         sub.Sid,
	}); err != nil {
		entry.RemoveSubscriber(sub.Sid) // 롤백
		return err
	}
	return nil
}

// UnsubscribeProcess는 sid를 uid의 구독자에서 뺀다. entry가 이미 memory에서 정리된 뒤(process
// 완료 등)라도 DB 원장은 남아있을 수 있어 entry 유무와 무관하게 DB delete는 항상 시도한다.
func (p *ProcessManager) UnsubscribeProcess(uid string, sid string) error {
	if entry, ok := p.memory.Get(uid); ok {
		entry.RemoveSubscriber(sid)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return p.pool.Queries().DeleteProcessSubscriber(ctx, superdb.DeleteProcessSubscriberParams{
		ProcessUid: uid,
		Sid:        sid,
	})
}

// ListSubscriptions는 sid가 구독 중인 process 목록을 반환한다. memory가 아니라 DB에서 직접
// 조회한다 — memory엔 sid→uid 역인덱스가 없고(전체 entry 스캔 필요), 이 조회는 재접속/화면복원
// 시 1회성이라 그 인프라를 둘 실익이 없음. memory는 휘발성(재시작 시 소실)이라 재시작 직후
// 복원 목적과도 맞지 않음 — status 컬럼도 write-through로 최신화되어 DB 단독으로 충분하다.
func (p *ProcessManager) ListSubscriptions(sid string) ([]superdb.Process, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return p.pool.Queries().ListProcessesBySid(ctx, sid)
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
