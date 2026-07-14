package router

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"nexus/cmd/supervisor/process"
	"nexus/internal/execute"
	"nexus/internal/protocol"
	"nexus/internal/supervisor/bind"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/transfer"
	"nexus/internal/transport"

	"github.com/jackc/pgx/v5/pgtype"
)

// ===== frontend → supervisor: 실행 트리거 (wire 오케스트레이션) =====
//
// 상태 생성(UID·spec·pool·AgentInteractive)은 ProcessManager에 위임하고, 여기선 "누구(who)"에
// 대한 라우팅/전송만 한다: worker 조회 → manager 호출 → (content 선배치) → MsgExec 전송 →
// bind.Relay 기동. folder-open은 worker 무접촉이라 presence만 남기고 조기 반환한다.

// Exec은 frontend의 "이 노드 실행/편집" 요청을 받아 worker에 명령하는 진입점이다.
// kind로 manager.Exec(EXEC/folder) vs manager.ExecEdit(EDIT)를 고른다.
func (r *supervisorRouter) Exec(owner superdb.User, authKey string, kind protocol.ExecType, node superdb.Node) error {
	worker, _ := r.workers.Get(authKey)

	// 1. 상태 등록(manager). UID·spec·Inter는 manager가 authoritative하게 만든다.
	//    method value로 EXEC/EDIT 분기(entry 타입 추론 → process 패키지 직접 import 회피).
	register := r.processManager.Exec
	if kind == protocol.ExecTypeEdit {
		register = r.processManager.ExecEdit
	}
	entry, err := register(worker, authKey, node, owner)
	if err != nil {
		return err
	}

	// 2. folder-open: worker 무접촉 → presence만 남기고 조기 종료(전송/relay 없음).
	if !entry.HasProcess() {
		return nil
	}
	uid := entry.Record.Uid

	// 3. content 선배치: node 본문을 실행 경로에 파일로 먼저 깐다. DestPath는 manager가 spec에
	//    쓴 것과 동일한 WorkerNodePath라 실행 대상과 정확히 일치한다(worker가 양쪽 동일 치환).
	//    EXEC=실행 가능(0755) / EDIT=편집 대상(0644). 전송 실패 시 등록 롤백.
	destPath := process.WorkerNodePath(node, *entry.Record)
	var content []byte
	if node.Content.Valid {
		content = []byte(node.Content.String)
	}
	perm := os.FileMode(0755)
	if kind == protocol.ExecTypeEdit {
		perm = 0644
	}
	if _, err := r.SendBuffer(authKey, transfer.NewReadBuffer(content, 0), destPath, perm); err != nil {
		r.processManager.Remove(uid)
		return fmt.Errorf("content 선배치 실패: %w", err)
	}

	// 4. fan-out relay 기동: Output()/Status() 드레인 → PROC:<uid> 토픽으로 Publish.
	//    Hub 제네릭 결합을 피하려 토픽 키를 클로저로 감싸 주입한다.
	bind.NewRelay(uid, entry.Inter, func(k protocol.MsgType, p any) error {
		return r.subscribeHub.Publish(processTopic(uid), k, p)
	}).Start()

	// 5. worker에 실행 명령. spec은 Record에서 파생(단일 진실의 출처).
	ctx, cancel := context.WithTimeout(context.Background(), sendCallTimeout)
	defer cancel()
	res, err := worker.Call(ctx, protocol.MsgExec, entry.Spec())
	if err != nil {
		r.processManager.Remove(uid)
		return fmt.Errorf("MsgExec 전송 실패: %w", err)
	}
	var execRes protocol.ExecResponse
	if err := res.Bind(&execRes); err != nil {
		r.processManager.Remove(uid)
		return err
	}
	if !execRes.Accept {
		r.processManager.Remove(uid)
		return fmt.Errorf("worker가 실행을 거부했습니다: %s", execRes.Reason)
	}
	return nil
}

// processTopic은 한 process의 fan-out 토픽 키다(구독/발행 단일 출처).
func processTopic(uid string) string { return "PROC:" + uid }

// TODO(frontend → supervisor 제어): 실행 중 process에 대한 입력/리사이즈/종료 핸들러.
//   input(MsgData)  → r.process.Get(uid).Inter.Write(data)
//   resize(MsgResize) → .Inter.Layout(cols, rows)
//   kill(MsgKill)     → .Inter.Kill()
// frontend 평면 어휘 확정 후 추가(worker용 MsgExec와 별개 타입일 수 있음).

// ===== worker → supervisor: process 이벤트 수신 =====

// output: MsgData(EVENT). worker가 보낸 출력 바이트를 받아 해당 process의 AgentInteractive로
// 밀어넣는다 → bind.Relay가 드레인해 frontend로 fan-out.
func (r *supervisorRouter) output(ev protocol.Frame) {
	var body protocol.DataEvent
	if err := ev.Bind(&body); err != nil {
		log.Printf("[process] output 디코드 실패: %v", err)
		return
	}
	entry, ok := r.processManager.Get(body.UID)
	if !ok || entry.Inter == nil {
		return // 이미 정리됐거나 folder-only 엔트리 → 버림
	}
	entry.Inter.PushOutput(body.Data)
}

// status: MsgStatus(EVENT). RUNNING(+PID) / 종료(Completed|Failed +ExitCode).
// worker 보고와 worker 끊김 합성이 공유하는 applyStatus 깔때기로 위임한다.
func (r *supervisorRouter) status(ev protocol.Frame) {
	var body protocol.StatusEvent
	if err := ev.Bind(&body); err != nil {
		log.Printf("[process] status 디코드 실패: %v", err)
		return
	}
	r.applyStatus(body.UID, body.Status, body.PID, body.ExitCode)
}

// applyStatus는 모든 상태전이의 유일 수렴점이다(REF-process "status 단일 깔때기").
// 진입: ① worker On(MsgStatus) ② worker 끊김 시 supervisor 합성 호출(CommandPending)
// ③ 재접속 3-way 대조 결과(CommandProcess=재바인딩 성공 / CommandFailed=worker측 소실).
// PID를 아는 여기서 pool 상태(MarkProcessRunning/Pending/Done)를 갱신하고, Inter에도 반영한다.
//
// ⚠️ memory entry 없어도(=이미 정리됨) DB 반영은 계속한다 — 재접속 "supervisor만 앎" 분기는
// 끊김 처리 때 이미 processManager.Remove된 uid에 대해 DB만 Failed로 닫아야 하기 때문이다.
// entry가 필요한 동작(Inter 반영·memory 정리)만 존재 여부로 가드한다.
//
// ⚠️ EDIT는 Completed에서 즉시 teardown 금지 — editResult(read-back) 처리가 UID→NodeID
// 매핑을 위해 엔트리를 필요로 한다. 따라서 EDIT 엔트리 제거는 editResult가 맡는다.
func (r *supervisorRouter) applyStatus(uid string, status execute.CommandStatus, pid, exit int) {
	entry, ok := r.processManager.Get(uid)

	q := store.GetStorePool().Queries()
	ctx := context.Background()

	switch {
	case status == execute.CommandProcess:
		if !ok {
			return // 이미 정리된 uid의 스퓨리어스 이벤트 → 무시
		}
		if _, err := q.MarkProcessRunning(ctx, superdb.MarkProcessRunningParams{
			Uid: uid,
			Pid: pgtype.Int4{Int32: int32(pid), Valid: pid > 0},
		}); err != nil {
			log.Printf("[process] MarkProcessRunning uid=%s: %v", uid, err)
		}
		if entry.Inter != nil {
			entry.Inter.PushStatus(status)
		}

	case status.IsCompleted():
		// entry 유무와 무관하게 DB는 항상 닫는다(주석 참조).
		if _, err := q.MarkProcessDone(ctx, superdb.MarkProcessDoneParams{
			Uid:      uid,
			Status:   status.String(),
			ExitCode: pgtype.Int4{Int32: int32(exit), Valid: true},
		}); err != nil {
			log.Printf("[process] MarkProcessDone uid=%s: %v", uid, err)
		}
		if !ok {
			return
		}
		if entry.Inter != nil {
			entry.Inter.Done(exit) // output/status 채널 close → relay 드레인 종료
		}
		// EDIT는 editResult 처리 후 제거(위 주석). EXEC 등은 여기서 memory 정리.
		if entry.Record == nil || protocol.ExecType(entry.Record.Type) != protocol.ExecTypeEdit {
			r.processManager.Remove(uid)
		}

	case status == execute.CommandPending:
		// worker 끊김 합성 호출 전용(REF-process "worker 끊김→PENDING"). 정상 경로에선 끊김
		// 직후 entry가 아직 memory에 있을 때만 호출된다 — 없으면 이미 정리된 것.
		if !ok {
			return
		}
		if _, err := q.MarkProcessPending(ctx, uid); err != nil {
			log.Printf("[process] MarkProcessPending uid=%s: %v", uid, err)
		}
		r.processManager.Remove(uid) // Inter.Done(502) 포함(안전망, sync.Once) → relay 드레인 종료

	default: // 기타 확정 live 아닌 상태
		if !ok {
			return
		}
		if entry.Inter != nil {
			entry.Inter.PushStatus(status)
		}
	}
}

// reconcileDisconnect는 worker 연결이 끊겼을 때 그 device 소유 활성 process를 전부
// PENDING으로 낙관적 전이한다(REF-process "worker 끊김→PENDING"). deviceKey=InstanceKey
// (workers 레지스트리 키와 동일). FOLDER는 pool 미저장이라 ListActiveByDevice에 안 잡혀
// 자동 제외된다.
func (r *supervisorRouter) reconcileDisconnect(deviceKey string) {
	q := store.GetStorePool().Queries()
	ctx := context.Background()

	rows, err := q.ListActiveByDevice(ctx, deviceKey)
	if err != nil {
		log.Printf("[process] ListActiveByDevice(disconnect) deviceKey=%s: %v", deviceKey, err)
		return
	}
	for _, row := range rows {
		r.applyStatus(row.Uid, execute.CommandPending, 0, 0)
	}
}

// sync: MsgSync(EVENT, worker→sup). 재접속 직후 worker가 보고하는 procs 스냅샷을 받아
// reconcileReconnect로 넘긴다. conn/auth는 register 핸들러와 동일하게 클로저로 캡처한다
// (register 완료 후에만 MsgSync가 오므로 auth는 이 시점에 항상 채워져 있다).
func (r *supervisorRouter) sync(conn *transport.Conn, auth *protocol.RegisterRequest) transport.EventHandler {
	return func(ev protocol.Frame) {
		var body protocol.SyncEvent
		if err := ev.Bind(&body); err != nil {
			log.Printf("[process] sync 디코드 실패: %v", err)
			return
		}
		deviceKey := auth.InstanceKey()
		if deviceKey == "" {
			return // register 전(비정상 순서) — 무시
		}
		r.reconcileReconnect(deviceKey, conn, body.Procs)
	}
}

// reconcileReconnect는 재접속한 worker의 보고(reported)와 DB의 활성(PENDING/PROCESS) uid
// 목록을 3-way 대조한다(REF-process "재접속 재바인딩"):
//   - 교집합: Rebind로 새 Inter 장착 + relay 재기동 + applyStatus로 보고 상태 동기화.
//   - DB만 앎(worker 재부팅 소실): CommandFailed로 종결(성공/실패 알 길 없어 Failed로 닫음).
//   - worker만 앎(고아): 로그만 남기고 무시(YAGNI, kill 지시 안 함).
func (r *supervisorRouter) reconcileReconnect(deviceKey string, worker *transport.Conn, reported []protocol.SyncEntry) {
	q := store.GetStorePool().Queries()
	ctx := context.Background()

	dbRows, err := q.ListActiveByDevice(ctx, deviceKey)
	if err != nil {
		log.Printf("[process] ListActiveByDevice(reconnect) deviceKey=%s: %v", deviceKey, err)
		return
	}

	dbUIDs := make(map[string]struct{}, len(dbRows))
	for _, row := range dbRows {
		dbUIDs[row.Uid] = struct{}{}
	}
	reportedByUID := make(map[string]protocol.SyncEntry, len(reported))
	for _, e := range reported {
		reportedByUID[e.UID] = e
	}

	for uid := range dbUIDs {
		entry, ok := reportedByUID[uid]
		if !ok {
			// supervisor만 앎: worker 재부팅으로 소실 → 성공/실패 알 길 없어 Failed로 종결.
			r.applyStatus(uid, execute.CommandFailed, 0, 502)
			continue
		}

		newEntry, err := r.processManager.Rebind(uid, worker)
		if err != nil {
			log.Printf("[process] Rebind uid=%s: %v", uid, err)
			continue
		}
		bind.NewRelay(uid, newEntry.Inter, func(k protocol.MsgType, p any) error {
			return r.subscribeHub.Publish(processTopic(uid), k, p)
		}).Start()
		r.applyStatus(uid, entry.Status, entry.PID, 0)
	}

	for uid := range reportedByUID {
		if _, ok := dbUIDs[uid]; !ok {
			log.Printf("[process] worker만 아는 고아 process 무시 uid=%s deviceKey=%s", uid, deviceKey)
		}
	}
}

// editResult: MsgEditResult(REQ, EDIT 전용). worker가 편집 종료 후 회수한 최종 파일 내용.
// EditResult엔 nodeId가 없으므로 UID→entry→NodeID로 매핑해 nodes.content를 diff 후 UPDATE한다
// (같으면 no-op). 처리 후 EDIT 엔트리를 제거한다(teardown 담당).
func (r *supervisorRouter) editResult(req protocol.Frame) (any, error) {
	var body protocol.EditResult
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	entry, ok := r.processManager.Get(body.UID)
	if !ok {
		return nil, fmt.Errorf("알 수 없는 편집 세션: %s", body.UID)
	}
	nodeID := entry.NodeID()
	if nodeID == 0 || entry.Record == nil {
		return nil, fmt.Errorf("편집 대상 노드를 찾을 수 없습니다: %s", body.UID)
	}
	owner := entry.Record.OwnerUserID

	q := store.GetStorePool().Queries()
	ctx := context.Background()

	// 저장판별 = read-back & diff. 현재 content와 같으면 no-op(:wq 안 했으면 파일 불변).
	cur, err := q.GetNode(ctx, superdb.GetNodeParams{ID: nodeID, OwnerUserID: owner})
	if err != nil {
		return nil, err
	}
	if !bytes.Equal([]byte(cur.Content.String), body.Content) {
		if _, err := q.UpdateNodeContent(ctx, superdb.UpdateNodeContentParams{
			ID:          nodeID,
			OwnerUserID: owner,
			Content:     pgtype.Text{String: string(body.Content), Valid: true},
		}); err != nil {
			return nil, err
		}
	}

	r.processManager.Remove(body.UID) // EDIT 세션 teardown(엔트리 제거)
	return nil, nil
}
