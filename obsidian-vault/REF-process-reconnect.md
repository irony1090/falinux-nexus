# REF — process 종료/재접속 모델 (worker 끊김→PENDING→재바인딩)

> 배선(구현) 상세는 `REF-process-wiring.md` / frontend 트리거·상태동기화 버그수정은 `REF-process-trigger.md`. 세션→uid 원장·REST 구독 배선은 `REF-process-subscription.md`. 계약/설계 원칙은 `REF-process.md`. 작업 이력은 `history/process-reconnect.md`.

## 종료 / 재접속 모델 (2026-07-01 확정) — status 단일 깔때기
> MEMORY의 "끊김→Done(502)" 폐기. 이 절이 최신.

- **`status` = 모든 상태전이의 유일 수렴점.** 어떤 종료든 process 라우터 `status`로. 두 진입: ① worker `On(MsgStatus)` ② worker 끊김 시 supervisor **합성 호출**. → `status(ev Frame)` 안쪽에 **`applyStatus(uid,status,pid,exit)` 코어** 분리(양쪽 재사용).
- **종료 시나리오**: 자연종료 / **터미널입력**(Ctrl+C=입력 0x03, kill 아님—프로그램이 결정, 입력경로) / **명시 kill**(종료버튼=MsgKill, 제어경로) / worker끊김(→PENDING) / EDIT취소(:cq→Failed→content UPDATE 안 함) / worker외부kill(OOM). **Ctrl+C≠kill 반드시 구분**.
- **frontend 끊김 ≠ 종료**: 명시적 종료 시그널 전까진 계속 실행(브라우저 종료 무시). **같은 세션 재접속=화면 그대로 복원** → 필요: process별 **ring buffer(SNAPSHOT, 이제 필수)** + **세션→보던 uid 원장**(conn 단위 Hub 구독과 별개) + 재접속 시 SNAPSHOT→live. **세션→uid 원장의 구체 설계는 아래 "세션→uid 원장" 절 참조(2026-07-14 착수).**
- **worker 끊김 = PENDING(낙관적)**: 죽이지 않고 관련 process 전부 PENDING. 재접속 시 worker가 live 상태 보고→재동기화(보고=PENDING→PROCESS / 미보고=끊긴 사이 사망→합성 종료). 매칭=worker가 **같은 InstanceKey(subkey resume)** 복귀 전제(`RegisterRequest.SubKey`).
- **⚠️ 재접속 conn 재바인딩(난점)**: 메모리 entry의 AgentInteractive `onWrite/onKill/onLayout`이 죽은 옛 conn을 클로저로 잡음 → 재접속 시 **출력/상태 SyncData 채널 유지**(relay·frontend 구독 보존)한 채 **콜백만 새 conn으로 교체**(또는 entry가 swappable conn 참조 보유).
- **PENDING 의미 확장**: "RUNNING前" + "끊김·상태미상" 둘 다 = "확정 live 아님".
- **미해결**: 끊긴 창 입력/kill(죽은 conn→"재연결 대기" 거절 or 큐잉) / 공유 process kill 인가(소유자/write공유자?) / kill 에스컬레이션(worker측 SIGTERM→SIGKILL) / supervisor 재시작 시 ring 소실→DB 출력sink(후순위).

## worker 끊김 → PENDING → 재접속 재바인딩 (2026-07-14 설계 확정 + 구현 완료 + e2e 검증)
> 위 절의 "⚠️ 재접속 conn 재바인딩(난점)"(콜백을 새 conn으로 교체) 대체. **교체 대신 폐기 후 재생성**으로 단순화 — `AgentInteractive`는 특정 conn을 캡처한 일회성 핸들이라 제자리 교체 시도가 오히려 복잡했음.
> **구현 완료(2026-07-14)**: 아래 설계 그대로 구현 + 실제 worker/supervisor 프로세스를 kill/restart해가며 3경로(끊김→PENDING / 재접속·소실→FAILED / 재접속·교집합→Rebind) e2e 검증(수동, DB 상태로 확인). `go build`/`go vet`/`gofmt` 클린.

- **끊김** (`handleWorkerWS`, `conn.Serve()` 리턴 직후, `workers.Remove(key)` 부근): `ListActiveByDevice(deviceKey)`(기존 존재·미사용 쿼리, `device_key`=InstanceKey) → 해당 worker 소유 PENDING/PROCESS 레코드 uid 목록 조회.
  - **FOLDER는 여기 자동 제외**됨 — `openFolder()`가 애초에 `CreateProcess`(DB persist)를 호출하지 않음(`entry.go` "pool 미저장" 주석 그대로). DB 조회 기반이라 별도 타입 필터 불필요, folder entry는 worker 연결 여부와 무관하게 memory에 영구 잔존(frontend 끊김 기준은 별개 이슈).
  - 각 uid에 대해 3단계:
    1. `entry, ok := processManager.Get(uid)`; `ok && entry.Inter != nil`이면 **`entry.Inter.Done(sentinel)`** 먼저 호출 — 안 하면 `bind.Relay`의 `pumpOutput`/`pumpStatus` 고루틴이 죽은 conn 채널을 계속 블로킹 읽으며 leak.
    2. `MarkProcessPending(ctx, uid)`(신규 쿼리, DB status→PENDING. `MarkProcessRunning`/`MarkProcessDone`과 대칭 네이밍).
    3. `processManager.Remove(uid)` — memory에서 **완전 제거**(Record-only로 남겨두지 않음). 재접속 시 3-way 대조는 memory가 아니라 DB(`ListActiveByDevice`)를 다시 읽으므로 memory에 잔존시킬 이유가 없음.
  - `applyStatus`에 **`case status == execute.CommandPending`을 `default`가 아닌 명시적 분기로 승격**해 위 3단계를 구현. 실사용상 Pending은 worker가 자발적으로 보고하는 경우가 사실상 없어(RUNNING 진입이 즉시) 이 분기는 사실상 이 합성 호출 전용.
  - EDIT/EXEC 구분 없이 동일 경로(합의됨) — EDIT 도중 끊겨도 그냥 PENDING, teardown은 안 함(editResult 못 받은 채로 대기).

- **재접속** (register 완료 직후): worker가 자기 `procs`(`cmd/worker/router/process.go`의 `procEntry` map)를 훑어 **`MsgSync{uid,status,pid}[]`을 자발적으로 전송**. `RegisterRequest`에 얹지 않고 별도 메시지로 분리(합의됨) — 식별자 핸드셰이크와 도메인 상태 동기화는 관심사가 다름.
  - supervisor: `ListActiveByDevice(deviceKey)`로 자기 쪽 PENDING/PROCESS uid 목록(위 끊김 처리로 memory에선 이미 제거된 상태 — DB만 진실)과 worker 보고를 3-way 대조:
    - **교집합**(양쪽 다 살아있음): `newWorkerInteractive(신규conn, uid)`로 **새 Inter 생성**(재사용/교체 안 함) → memory에 재등록(`ProcessManager`에 신규 메서드 필요, 가칭 `Rebind(uid, worker) (*ProcessEntry, error)` — DB에서 Record 다시 읽어 entry 재구성 + Inter 장착) → `applyStatus(uid, 보고된 status, pid, 0)`로 현재 상태 동기화(RUNNING 보고 시 PENDING→PROCESS 복귀).
    - **supervisor만 앎**(worker 보고에 없음 = worker 재부팅으로 소실): `applyStatus(uid, CommandFailed, 0, sentinel)`로 종결. 성공/실패 알 길 없어 Failed로 닫음(합의됨 — 별도 "LOST" 상태 신설 안 함).
    - **worker만 앎**(고아, supervisor 기록 소실 등 극단 케이스): 로그만 남기고 무시(YAGNI, 합의됨). kill 지시 안 함.

**구현 목록** (완료, 위치):

| 항목 | 위치 |
|---|---|
| `MarkProcessPending` 쿼리 신설 | `internal/supervisor/db/query/processes.sql` (+sqlc generate) |
| `applyStatus`에 `CommandPending` 명시 분기 + **가드 완화**(아래 별도 절) | `cmd/supervisor/router/process.go` |
| `reconcileDisconnect(deviceKey)`: `ListActiveByDevice` 순회 → `applyStatus(uid, CommandPending, 0, 0)` | `cmd/supervisor/router/process.go`, 호출은 `supervisorRouter.go` `handleWorkerWS`(`workers.Remove(key)` 직후) |
| `MsgSync`(worker→sup EVENT, `SyncEntry{UID,Status,PID}[]`) 프로토콜 신설 | `internal/protocol/messages.go` |
| worker: register 완료 직후 `sendSync()`(procs 스냅샷 전송) | `cmd/worker/router/register.go` |
| supervisor: `sync()` 핸들러(conn/auth 클로저) + `reconcileReconnect`(3-way partition + Rebind/applyStatus) | `cmd/supervisor/router/process.go`, 등록은 `supervisorRouter.go` `handleWorkerWS` |
| `ProcessManager.Rebind(uid, worker) (*ProcessEntry, error)`(DB에서 Record 재조회+새 Inter 장착) | `cmd/supervisor/process/manager.go` |

### `applyStatus` 가드 완화 (구현 중 발견한 설계 공백 해소, 2026-07-14)
원래 `applyStatus`는 진입 시 `entry, ok := processManager.Get(uid); if !ok { return }`로 memory entry 없으면 즉시 리턴했다. 그런데 재접속의 "supervisor만 앎(worker 소실)" 분기는 `applyStatus(uid, CommandFailed, ...)`를 호출하는 시점에 그 uid가 **이미 끊김 처리(`CommandPending` 분기)에서 `processManager.Remove`된 뒤**라 entry가 없다 — 원 설계 그대로면 DB가 Failed로 안 닫힌다.
**해소**: 상단 가드 제거, `switch` 각 `case` 내부에서 개별적으로 `ok` 처리.
- `CommandProcess`: entry 없으면 스킵(스퓨리어스 이벤트 방어, 기존과 동일 동작).
- `IsCompleted()`: **entry 유무와 무관하게 `MarkProcessDone` 먼저 실행** → entry 있을 때만 `Inter.Done`/memory 정리.
- `CommandPending`: entry 없으면 스킵(정상 경로는 항상 entry 있을 때 호출됨).
- "모든 상태전이는 `applyStatus` 단일 깔때기" 원칙은 그대로 유지(별도 경로로 우회하지 않음).

### worker측 기반 문제 발견 + `WorkerState` (구현 중 발견, 2026-07-14)
REF 설계는 "재접속 시 worker가 자기 procs 스냅샷을 보고한다"를 전제하지만, 기존 구현은 `cmd/worker/main.go`의 재접속 루프가 매 시도마다 `NewWorkerRouter`를 새로 호출하고 그 안에서 `procs`(살아있는 PTY 핸들 맵)를 **매번 새로 생성**했다 — ws가 끊겼다 재접속하면 새 `workerRouter`가 빈 `procs`로 시작해 끊기기 전 PTY를 전부 잃는다. 게다가 `pumpOutput`/`pumpStatus`가 옛 `workerRouter`의 `r.conn`(죽은 conn)을 메서드 리시버로 영구 캡처해, 재접속해도 새 conn으로 못 옮겨가고 죽은 conn에 계속 씀.

**해소 = `WorkerState`**(`cmd/worker/router/workerRouter.go`): `procs *manager.KeyValManager[string,*procEntry]` + `conn atomic.Pointer[transport.Conn]`을 묶어 **`main.go`의 재접속 루프 밖에서 1회 생성**, 매 `NewWorkerRouter` 호출에 그대로 넘긴다.
- `procs`는 identity를 그대로 재사용(새 workerRouter도 `router.procs = state.procs`) → 재접속해도 이전 PTY 엔트리가 살아있음.
- `NewWorkerRouter`가 dial 성공 직후 `state.conn.Store(conn)`으로 "현재 conn"을 교체. `pumpOutput`/`pumpStatus`/`teardown`은 `r.conn` 대신 `r.state.currentConn()`을 매 전송 시점에 읽는다 — 옛 세대의 pump 고루틴이라도 공유된 `state`를 통해 최신 conn을 즉시 얻는다(nil이면 그 순간만 전송 스킵, 드레인은 계속).
- worker의 `sendSync()`는 `r.procs`에 남아있는 엔트리를 전부 `CommandProcess`로 보고한다 — `teardown()`이 종료된 uid를 즉시 `procs.Remove`하므로 "맵에 남아있다=아직 실행 중"이 불변조건이라, `entry.inter.Status()`를 다시 읽지 않는다(그 채널은 `pumpStatus` 고루틴 전용 단일 소비자라 여기서 또 읽으면 이벤트를 가로채 레이스가 난다).

### PENDING 오삭제 버그 — "합성 호출 전용" 가정이 틀렸음 (2026-07-22 발견·수정)
> 위 "구현 목록" 표의 `applyStatus` `CommandPending` 분기 주석("실사용상 Pending은 worker가 자발적으로 보고하는 경우가 사실상 없어... 이 분기는 사실상 이 합성 호출 전용")이 **틀린 가정**이었음이 사용자 실사용 테스트로 드러남.

- **실제로는 매 실행마다 발생**: worker의 `pty.ExecInteractive`는 실행 시작 시 `setStatusPending()`을 항상 먼저 push하고 곧이어 `setStatusProcess()`를 push한다(`execInteractive.go`) — PENDING→RUNNING이 정상 시퀀스. worker `pump`가 이 PENDING을 그대로 `MsgStatus`로 supervisor에 보내, `reconcileDisconnect`(끊김 합성 호출)와 **완전히 같은 `applyStatus`의 `CommandPending` 분기**를 정상 실행 직후에도 타게 됨.
- **증상**: frontend에서 process를 실행하면 worker가 정상 PENDING을 보고하자마자 `applyStatus`가 `processManager.Remove(uid)`를 호출 → `Inter.Done(502)`로 FAILED 합성 → 화면엔 실행하자마자 정지된 것처럼 보임(worker의 실제 프로세스는 안 죽고 계속 살아있어 좀비화).
- **수정**: `applyStatus` 진입부에 `if ok && entry.Record.Status == status.String() { return }` 가드 추가 — "지금 memory가 이미 알고 있는 상태와 똑같은 상태 보고면 무시". 정상 exec 흐름에서 entry 생성 시 DB 기본값이 이미 `'PENDING'`이라(`CreateProcess`), 처음 들어오는 진짜 PENDING 보고는 이 가드에 걸려 조용히 무시되고, 그다음 PROCESS(RUNNING) 보고는 값이 달라(`PENDING`≠`PROCESS`) 정상 처리된다.
- **끊김 케이스가 안 깨지는 이유**: 이 가드가 정확히 동작하려면 memory `entry.Record.Status`가 실시간이어야 한다(`REF-process-trigger.md` "entry.Record memory 동기화" 참조 — `SetRecord`로 해결). 실행 중이던 process는 이미 `Record.Status="PROCESS"`로 갱신돼 있어, 끊김 시 합성 PENDING 호출이 들어오면 `"PROCESS"≠"PENDING"`이라 가드를 통과해 원래대로 `Remove`된다. `Rebind` 시 새 entry는 DB에서 다시 읽어오므로(`GetProcess`) 항상 최신 status로 시작 — 재접속 흐름과 충돌 없음.
- **검증**: 정상 exec 후 process가 안 죽는 것 확인 + kill 테스트로 재검증 완료.

