package router

import (
	"bytes"
	"context"
	"errors"
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

// ===== frontend -> supervisor: 실행 =====
// 상태 생성은 ProcessManager, 여기는 라우팅·전송만: worker 조회 -> 등록 -> 본문 선배치 -> 구독 -> relay -> MsgExec

// Exec: 실행·편집 진입점. 성공하면 tabID가 크기 소유자. 소유 = 노드 소유자, startedBy = 실행 계정 — SH9(받은 쪽이 실행한 process의 소유)
func (r *supervisorRouter) Exec(startedBy int64, authKey string, kind protocol.ExecType, node superdb.Node, tabID string) (string, error) {
	worker, _ := r.workers.Get(authKey)

	// 1. 상태 등록 — UID·spec·Inter는 manager가 만든다
	register := r.processManager.Exec
	if kind == protocol.ExecTypeEdit {
		register = r.processManager.ExecEdit
	}
	entry, err := register(worker, authKey, node, node.OwnerUserID, startedBy)
	if err != nil {
		return "", err
	}
	uid := entry.Record.Uid

	// 2. 본문 선배치 — spec과 같은 WorkerNodePath. EXEC = 0755 / EDIT = 0644, 실패하면 등록 롤백
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
		r.failExec(uid)
		return "", fmt.Errorf("content 선배치 실패: %w", err)
	}
	return uid, r.launch(worker, entry, tabID, node.Name)
}

// ExecFile: 장비 디렉토리 파일 실행(폴더-4(실행)) — 본문 선배치 없이 Exec의 3~5단계만. 소유·실행 = userID
func (r *supervisorRouter) ExecFile(userID int64, authKey, cmd, cwd string, dirNodeID int64, tabID string) (string, error) {
	worker, _ := r.workers.Get(authKey)
	entry, err := r.processManager.ExecFile(worker, authKey, cmd, cwd, dirNodeID, userID)
	if err != nil {
		return "", err
	}
	uid := entry.Record.Uid
	return uid, r.launch(worker, entry, tabID, "")
}

// launch: 등록된 process를 worker에서 시작한다. 실패하면 FAILED로 닫는다. label = 스크립트 노드 이름(미등록 process 표시용)
func (r *supervisorRouter) launch(worker *transport.Conn, entry *process.ProcessEntry, tabID, label string) error {
	uid := entry.Record.Uid

	// 3. 순서 의존: 구독은 relay Start() 전에 — 늦으면 RUNNING 등 첫 이벤트를 놓친다
	r.subscribeViewers(uid)

	// 4. relay 기동(끝나면 구독 정리까지 — startRelay)
	r.startRelay(uid, entry.Inter)

	// 5. worker에 실행 명령
	ctx, cancel := context.WithTimeout(context.Background(), sendCallTimeout)
	defer cancel()
	spec := entry.Spec()
	spec.Label = label
	res, err := worker.Call(ctx, protocol.MsgExec, spec)
	if err != nil {
		r.failExec(uid)
		return fmt.Errorf("MsgExec 전송 실패: %w", err)
	}
	var execRes protocol.ExecResponse
	if err := res.Bind(&execRes); err != nil {
		r.failExec(uid)
		return err
	}
	if !execRes.Accept {
		r.failExec(uid)
		return fmt.Errorf("%w: %s", errExecRejected, execRes.Reason)
	}
	r.setSizeOwner(uid, tabID) // 실행한 탭 = 크기 소유자
	return nil
}

var errExecRejected = errors.New("worker가 실행을 거부했습니다") // 파일 없음·실행 불가 등 = 요청 쪽 문제

const exitExecFailed = 126 // 실행이 시작되지 못함(셸 관례 "실행 불가")

// failExec: 시작 못 한 process를 DB까지 FAILED로 — 메모리만 지우면 PENDING 행이 "실행 중"으로 남는다(2026-10-12). EDIT 엔트리는 여기서 지움
func (r *supervisorRouter) failExec(uid string) {
	r.applyStatus(uid, execute.CommandFailed, 0, exitExecFailed)
	if _, ok := r.processManager.Get(uid); ok {
		r.processManager.Remove(uid)
	}
}

// publishProcess: PROCESS:UPDATE(전체 process 구조체) — 크기 변경·종료 요청
func (r *supervisorRouter) publishProcess(rec superdb.Process) {
	if err := r.subscribeHub.Publish(processTopic(rec.Uid), protocol.MsgProcessUpdate, newProcessResponse(rec)); err != nil {
		log.Printf("[process] Publish 실패 topic=%s: %v", processTopic(rec.Uid), err)
	}
}

// startRelay: relay를 띄우고, 드레인이 끝나면 구독을 정리한다. Exec·reconcileReconnect 둘 다 여기로만 띄울 것
// 재발 방지: worker 끊김(Detach)으로 끝난 relay는 정리하지 않는다 — process는 살아 있어, 정리하면 재접속 뒤
// 새 relay의 출력을 받을 소켓이 없어진다(2026-10-01 "Rebind 뒤 출력 끊김")
func (r *supervisorRouter) startRelay(uid string, inter *execute.AgentInteractive) {
	relay := bind.NewRelay(uid, inter, r.screens.open(uid), func(k protocol.MsgType, p any) error {
		return r.subscribeHub.Publish(processTopic(uid), k, p)
	})
	relay.Start()
	go func() {
		relay.Wait() // 마지막 Completed/Failed 발행까지 드레인한 뒤 반환
		if inter.Detached() {
			return
		}
		r.cleanupProcessTopic(uid)
	}()
}

// frontend → supervisor 제어: 입력 = 소켓 PROCESS:INPUT(input.go) / 리사이즈·종료 = REST(processApi.go)

// ===== worker → supervisor: process 이벤트 수신 =====

// output: MsgData — 출력 바이트를 Inter로(relay가 브라우저로 보냄)
func (r *supervisorRouter) output(ev protocol.Frame) {
	var body protocol.DataEvent
	if err := ev.Bind(&body); err != nil {
		log.Printf("[process] output 디코드 실패: %v", err)
		return
	}
	entry, ok := r.processManager.Get(body.UID)
	if !ok || entry.Inter == nil {
		return // 이미 정리된 엔트리 → 버림
	}
	entry.Inter.PushOutput(execute.Chunk{Data: body.Data, Off: body.Off})
}

// screen: MsgScreen — 보관 프로세스 화면 스냅샷을 출력과 같은 줄로(relay가 순서대로 ring을 채움). 재접속 뒤 sync가 Rebind한 다음에 온다
func (r *supervisorRouter) screen(ev protocol.Frame) {
	var body protocol.ScreenEvent
	if err := ev.Bind(&body); err != nil {
		log.Printf("[process] screen 디코드 실패: %v", err)
		return
	}
	entry, ok := r.processManager.Get(body.UID)
	if !ok || entry.Inter == nil {
		return
	}
	entry.Inter.PushOutput(execute.Chunk{Data: body.Data, Off: body.Off, Seed: true, Alt: body.Alt})
}

// status: MsgStatus — applyStatus로
func (r *supervisorRouter) status(ev protocol.Frame) {
	var body protocol.StatusEvent
	if err := ev.Bind(&body); err != nil {
		log.Printf("[process] status 디코드 실패: %v", err)
		return
	}
	log.Printf("[PROCESS.GO] status %v", body)
	if body.Status.IsCompleted() {
		r.unregistered.removeUID(body.UID)
	}
	r.applyStatus(body.UID, body.Status, body.PID, body.ExitCode)
}

// applyStatus: 모든 상태 전이의 단일 깔때기 — worker 보고 / 끊김 합성(PENDING) / 재접속 대조(REF-process "status 단일 깔때기")
// 메모리 entry가 없어도 DB는 닫는다(재접속 "supervisor만 앎" = 이미 Remove된 uid). entry가 필요한 동작만 존재로 가드
// EDIT는 Completed에서 지우지 않는다 — editResult가 UID -> NodeID 매핑에 엔트리를 쓴다
func (r *supervisorRouter) applyStatus(uid string, status execute.CommandStatus, pid, exit int) {
	entry, ok := r.processManager.Get(uid)

	q := store.GetStorePool().Queries()
	ctx := context.Background()
	log.Printf("[PROCESS.GO] applyStatus:%s:%s", uid, status.String())
	if ok && entry.Record != nil && entry.Record.Status == status.String() { // 현재와 똑같은 상태일 경우 무시
		return
	}
	switch {
	case status == execute.CommandProcess:
		if !ok {
			return // 이미 정리된 uid의 스퓨리어스 이벤트 → 무시
		}
		row, err := q.MarkProcessRunning(ctx, superdb.MarkProcessRunningParams{
			Uid: uid,
			Pid: pgtype.Int4{Int32: int32(pid), Valid: pid > 0},
		})
		if err != nil {
			log.Printf("[process] MarkProcessRunning uid=%s: %v", uid, err)
		} else {
			entry.SetRecord(&row) // DB RETURNING 결과로 memory record 전체 교체(Pid·StartedAt 포함 항상 동기화)
		}
		if entry.Inter != nil {
			entry.Inter.PushStatus(status)
		}

	case status.IsCompleted():
		// entry 유무와 무관하게 DB는 닫는다
		row, err := q.MarkProcessDone(ctx, superdb.MarkProcessDoneParams{
			Uid:      uid,
			Status:   status.String(),
			ExitCode: pgtype.Int4{Int32: int32(exit), Valid: true},
		})
		if err != nil {
			log.Printf("[process] MarkProcessDone uid=%s: %v", uid, err)
		}
		r.releaseSizeOwner(uid)
		if !ok {
			return
		}
		if err == nil {
			entry.SetRecord(&row) // ExitCode·FinishedAt 포함 항상 동기화
		}
		if entry.Inter != nil {
			entry.Inter.Done(exit) // output/status 채널 close → relay 드레인 종료
		}
		// EDIT는 editResult가 제거
		if entry.Record == nil || protocol.ExecType(entry.Record.Type) != protocol.ExecTypeEdit {
			r.processManager.Remove(uid)
		}

	case status == execute.CommandPending:
		// worker 끊김 합성 전용 — entry가 없으면 이미 정리된 것
		if !ok {
			return
		}
		row, err := q.MarkProcessPending(ctx, uid)
		if err != nil {
			log.Printf("[process] MarkProcessPending uid=%s: %v", uid, err)
		} else {
			entry.SetRecord(&row)
		}
		// X(끊김 표시): PENDING을 발행하고 종료 STATUS 없이 relay만 끝낸다(Remove의 Done(502)는 FAILED를 발행함)
		if entry.Inter != nil {
			entry.Inter.PushStatus(status)
		}
		r.processManager.Detach(uid)

	default: // 기타 확정 live 아닌 상태(DB에 대응 컬럼 없음 — memory Status만 반영)
		if !ok {
			return
		}
		if entry.Record != nil {
			entry.Record.Status = status.String()
		}
		if entry.Inter != nil {
			entry.Inter.PushStatus(status)
		}
	}
}

// reconcileDisconnect: 끊긴 worker(deviceKey = InstanceKey)의 활성 process를 전부 PENDING으로(낙관적)
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

// sync: MsgSync — 재접속 직후 worker의 procs 스냅샷을 reconcileReconnect로(register 뒤에만 오므로 auth는 채워져 있음)
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

// reconcileReconnect: worker 보고와 DB 활성 uid를 대조한다(REF-process "재접속 재바인딩")
//   - 둘 다: Rebind + 재구독 + relay 재기동 + 상태 동기화 + 기록된 종료 요청 전달(종료-3(재접속 전달))
//   - DB만: FAILED로 닫고 직접 발행(relay 없음) / worker만: 미등록 process 목록(unregistered)
func (r *supervisorRouter) reconcileReconnect(deviceKey string, worker *transport.Conn, reported []protocol.SyncEntry) {
	q := store.GetStorePool().Queries()
	ctx := context.Background()

	dbRows, err := q.ListActiveByDevice(ctx, deviceKey)
	if err != nil {
		log.Printf("[process] ListActiveByDevice(reconnect) deviceKey=%s: %v", deviceKey, err)
		return
	}

	dbUIDs := make(map[string]superdb.Process, len(dbRows))
	for _, row := range dbRows {
		dbUIDs[row.Uid] = row
	}
	reportedByUID := make(map[string]protocol.SyncEntry, len(reported))
	for _, e := range reported {
		reportedByUID[e.UID] = e
	}

	for uid, row := range dbUIDs {
		entry, ok := reportedByUID[uid]
		if !ok {
			// supervisor만 앎: worker 재부팅으로 소실 → 성공/실패 알 길 없어 Failed로 종결.
			r.applyStatus(uid, execute.CommandFailed, 0, 502)
			r.publishLost(uid)
			continue
		}

		newEntry, err := r.processManager.Rebind(uid, worker)
		if err != nil {
			log.Printf("[process] Rebind uid=%s: %v", uid, err)
			continue
		}
		r.subscribeViewers(uid) // W(재접속 재구독): 끊긴 사이 연결된 소켓·supervisor 재시작 대비
		r.startRelay(uid, newEntry.Inter)
		r.applyStatus(uid, entry.Status, entry.PID, 0)
		if row.KillRequestedAt.Valid { // 종료-3(재접속 전달): EVENT 배달 고루틴이라 Call 응답을 기다리지 않게 따로 띄운다
			go func() {
				if err := newEntry.Inter.Kill(row.KillForced); err != nil {
					log.Printf("[process] 재접속 종료 전달 실패 uid=%s force=%v: %v", uid, row.KillForced, err)
				}
			}()
		}
	}

	// worker만 앎 = 미등록 process: 죽이지 않고 목록에만 둔다(장비 타일에서 보고 끔) — 주소-Q12, REF-supervisor-move.md
	var unregistered []protocol.SyncEntry
	for uid, e := range reportedByUID {
		if _, ok := dbUIDs[uid]; !ok {
			log.Printf("[process] 미등록 process uid=%s deviceKey=%s", uid, deviceKey)
			unregistered = append(unregistered, e)
		}
	}
	r.unregistered.set(deviceKey, unregistered)
}

// publishLost: relay 없이 끝난 process의 FAILED를 직접 발행 — 구독이 이미 정리됐을 수 있어 먼저 다시 구독한다
func (r *supervisorRouter) publishLost(uid string) {
	r.subscribeViewers(uid)
	ev := protocol.StatusEvent{UID: uid, Status: execute.CommandFailed, ExitCode: 502}
	if err := r.subscribeHub.Publish(processTopic(uid), protocol.MsgStatus, ev); err != nil {
		log.Printf("[process] Publish 실패 topic=%s: %v", processTopic(uid), err)
	}
	r.cleanupProcessTopic(uid)
}

// editResult: EDIT 종료 후 파일 내용 — UID -> entry -> NodeID로 content를 비교해 다르면 UPDATE, 끝나면 엔트리 제거
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
	owner := entry.Record.OwnerUserID // SH9로 process 소유자 = 노드 소유자

	q := store.GetStorePool().Queries()
	ctx := context.Background()

	// 같으면 저장 안 함(:wq 안 했으면 파일 그대로)
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
