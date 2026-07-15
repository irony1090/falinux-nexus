# REF — process 종료/재접속 모델 + 세션 복원

> 배선(구현) 상세는 `REF-process-wiring.md`. 계약/설계 원칙은 `REF-process.md`. 작업 이력은 `history/process-reconnect.md`.

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

## 세션 → uid 원장 구체화: process_subscribers 테이블 + sid 추출 (2026-07-14, 설계 진행 중 — 미구현)
> 위 "종료/재접속 모델"의 "세션→보던 uid 원장" 항목의 구체 설계. **아직 마이그레이션/쿼리 파일 미생성** — 스키마·쿼리·트리거 시점은 합의됐고, CREATE/DELETE 실제 호출 지점 배선만 남음.

**sid(브라우저 세션) 추출 배선 — 코드 완료**
- 문제의식: 기존엔 `owner_user_id`(UserID)만 있어 **같은 유저가 다른 브라우저로 로그인해도 구분이 안 됨**. "같은 브라우저=같음, 다른 브라우저(같은 유저라도)=다름"을 만족하는 키가 필요.
- `sid` 쿠키(gorilla `CookieStore`)는 **로그인(`Save()`) 시점의 timestamp가 MAC 서명에 포함**되어 로그인마다 다른 문자열이 되고, `requireSession`/새로고침은 `Save()`를 다시 안 불러 **같은 브라우저 내에서는 로그아웃 전까지 고정** — 정확히 필요한 특성. `Save()` 호출부는 `signIn`/`signOut`(`user.go`) 딱 두 곳뿐(grep으로 확인).
- `internal/manager/session/sessionManager.go`: `NameFunc[T]` 시그니처에 `req *http.Request` 추가(`func(data T, session *sessions.Session, req *http.Request, key string) string`) → `Name()`이 `req`까지 넘겨줌. (→ `REF-supervisor-web.md` "세션 매니저" 절도 갱신됨)
- `cmd/supervisor/router/supervisorRouter.go`: `getSessionKey` nameFn 구현 배선 — `req.Cookie(key)`로 원본 `sid` 쿠키 값 추출해 `NewSessionManager("irony","sid", getSessionKey)`로 주입.
- **⚠️ 미해결 버그**: `getSessionKey`가 `c, _ := req.Cookie(key)`로 에러를 버려서, 쿠키가 없는 요청에서 `.Name()`을 호출하면 `c`가 `nil`이라 **nil pointer panic**(PanicMiddleware가 잡아 500으로 렌더는 됨, 그러나 실제 요청은 실패). `err != nil` 가드 필요.
- 쿠키에 `Password`(해시) 필드가 함께 직렬화되어 있음(`REF-supervisor-web.md` 기존 기록) — `sid` 원본 문자열을 그대로 DB 컬럼/로그에 남기지 말고 **해시(sha256 등)해서 저장 권장**(아직 미적용, 아래 DDL·쿼리는 raw 저장으로 초안 작성됨 — 실제 적용 시 재검토).

**process_subscribers 테이블 설계 (DDL 초안, 마이그레이션 미생성)**
```sql
CREATE TABLE process_subscribers (
    process_uid   TEXT        NOT NULL REFERENCES processes(uid) ON DELETE CASCADE,
    owner_user_id BIGINT      NOT NULL REFERENCES users(id),
    sid           TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (process_uid, sid)
);
CREATE INDEX idx_process_subscribers_sid ON process_subscribers(sid);
```
- **PK=(process_uid, sid)** 자연키 — 같은 브라우저가 같은 process를 중복 구독해도 1행만 유지(재구독은 `ON CONFLICT DO NOTHING`으로 무시).
- **`process_uid` FK `ON DELETE CASCADE`** — process 자체가 삭제되면 구독 정보도 자동 정리(고아 row 방지 안전망. 아래 DELETE 트리거 정책과는 별개, 병행).
- `owner_user_id`는 FK 걸어 조회/감사 용도로 유지(스키마상 sid만으로도 유저 특정 가능하나 컬럼 분리 유지).

**쿼리 3종 (`internal/supervisor/db/query/processSubscribers.sql`, 미생성)**
```sql
-- name: CreateProcessSubscriber :exec
INSERT INTO process_subscribers (process_uid, owner_user_id, sid)
VALUES ($1, $2, $3)
ON CONFLICT (process_uid, sid) DO NOTHING;

-- name: DeleteProcessSubscriber :exec
DELETE FROM process_subscribers WHERE process_uid = $1 AND sid = $2;

-- name: ListProcessesBySid :many
SELECT p.* FROM processes p
JOIN process_subscribers ps ON ps.process_uid = p.uid
WHERE ps.sid = $1
ORDER BY ps.created_at DESC;
```
- CREATE/DELETE만으론 "재접속 시 복원"이 실제로 동작 안 함(조회 수단이 없어서) → `ListProcessesBySid`(sid로 process 목록 join) 추가. 재접속 시 프론트가 부를 엔드포인트가 이걸 사용하게 됨.

**CREATE/DELETE 트리거 시점 — 2안(요청/해지마다 즉시 반영) 채택**
- 비교한 두 안: ① 프론트 소켓이 끊어질 때 그 순간의 구독 목록을 스냅샷으로 일괄 write / ② 구독 요청·해지마다 그때그때 메모리+DB 반영.
- DELETE는 이미 "프론트가 명시적으로 구독 해제 요청할 때만"으로 정해져 있음(①처럼 소켓 끊김에 걸면 새로고침도 소켓이 끊기는 거라 **매 새로고침마다 구독 정보가 삭제**돼 복원 기능 자체가 성립 안 함 — 새로고침과 완전 종료를 서버가 구분 못 하기 때문).
- **② 채택 이유**: DELETE와 트리거 지점이 대칭(둘 다 명시적 프로토콜 이벤트, 소켓 라이프사이클에 안 얹음) + **supervisor 프로세스 자체가 재시작해도 안전**(①은 close 핸들러가 못 돌면 그 세션의 구독 정보가 DB에 아예 안 남음 — supervisor 재시작 복구 시나리오와 충돌).

**남은 것 (미착수)**
- 마이그레이션 파일(`00004_process_subscribers.sql`, goose Up/Down) + query 파일 생성 + `sqlc generate`(로컬 실행 필요, 저장소에 committed 스크립트/Makefile 타깃 없음 — 수동 실행). ⚠️ 번호 충돌 주의: MEMORY의 node roster(`worker_instances`)도 `00004` 후보 — 이 작업이 먼저 착수되면 `worker_instances`는 `00005`로 밀려야 함(MEMORY 갱신 필요).
- **CREATE 실제 호출 지점**: 클라이언트가 `PROC:<uid>` 토픽을 동적 구독할 때 함께 INSERT. `REF-process-wiring.md` "PROC 토픽 무구독" 발견 + `REF-realtime.md` "동적 구독 어휘"와 합류 지점 — 애초에 동적 구독 자체가 아직 미배선이라 이 배선과 같이 가야 함.
- **DELETE 실제 호출 지점**: 프로토콜에 UNSUBSCRIBE류 메시지가 아직 없으면 신설 필요(현재 `subscribe.go`는 구독만 있고 명시적 해지 메시지 없음).
- `getSessionKey`의 nil-cookie panic 가드: **작업트리에 이미 적용됨**(`err != nil` → `""` 반환, 커밋 전). 위 "미해결 버그" 서술은 커밋 시점 기준 stale.
- `sid` 저장 방식: **원본 그대로 저장 확정**(2026-07-15, 해시 대안 기각 — 단순 구현 우선).

**`ProcessEntry` 인메모리 구독자 필드 — 구현 완료 (2026-07-15)**
- `ProcessEntry`(`cmd/supervisor/process/entry.go`)에 구독자 목록 필드 추가. DB row(`process_subscribers`) 그대로 미러링 **안 함** — `Subscriber{Sid, OwnerUserID}`만 담는 단순 구조.
- **이유**: `process_uid`는 엔트리 자체가 그 uid로 관리돼 중복(엔트리가 곧 process_uid 스코프) / `created_at`은 DB 감사·원장 목적이지 런타임 라우팅엔 불필요 — "DB≠라우팅 authority"(MEMORY 확정 결정, 실시간 push 절)와 동일한 원칙: 메모리 객체는 DB row 미러가 아니라 그 순간 필요한 도메인 필드만 갖는 lean 구조.
- API: `IsSubscribed(sid) bool` / `AddSubscriber(Subscriber)`(같은 sid 중복 무시, DB `ON CONFLICT DO NOTHING`과 동일 시맨틱) / `RemoveSubscriber(sid)`. 엔트리별 `sync.RWMutex`로 슬라이스 보호(`KeyValManager`의 맵 mutex는 map 자체만 지키고 entry 내부 필드 동시접근은 안 막아줘서 별도 필요).
- `go build`/`go vet` 클린. **아직 미배선**: DB CREATE/DELETE(위 "남은 것" 절)와 이 메모리 메서드를 같이 호출하는 지점(동적 구독 핸들러) — 이 필드 자체는 아직 아무 데서도 호출 안 됨(dead code 상태).

**`ProcessManager.SubscribeProcess`/`UnsubscribeProcess` — memory+DB write-through 구현 완료 (2026-07-15)**
- 위치는 `cmd/supervisor/process/manager.go`(라우터/핸들러가 아니라 매니저) — `execScript`(memory Append→DB CreateProcess, 실패 시 memory 롤백)/`Rebind`/`Remove`가 이미 "memory+DB를 한 지점에서 같이 맞추는" 책임을 매니저가 지는 패턴이라 그대로 따름(사용자 확정).
- `SubscribeProcess(uid, Subscriber)`: **memory 먼저**(`entry.AddSubscriber`, 저렴·되돌리기 쉬움) → `CreateProcessSubscriber` DB insert, **실패 시 memory 롤백**(`entry.RemoveSubscriber`) — `execScript`와 동일 순서/롤백 대칭. entry 없으면(uid 존재 안 함) 에러 반환.
- `UnsubscribeProcess(uid, sid)`: entry 있으면 memory 먼저 제거, **DB delete는 entry 유무와 무관하게 항상 시도** — process가 이미 완료돼 `Remove()`로 memory에서 정리된 뒤라도 `process_subscribers` DB 원장은 남아있을 수 있어서(구독 해제는 프론트 명시 요청 시에만, 완료와 무관) 비대칭으로 뒀음. DB delete 실패 시 memory 롤백은 안 함(이미 지웠고 재추가할 정보 없음 — 다음 재구독 때 자연 회복).
- `go build`/`go vet` 클린. **아직 미배선**: 이 두 메서드를 실제로 호출하는 동적 구독 핸들러(`MsgSubscribe`/`MsgUnsubscribe`, `handleSubscribeWS`) — 이번 커밋 스코프는 CREATE/DELETE 로직 자체만, 호출 지점은 후속.

**`ProcessManager.ListSubscriptions(sid)` — 조회는 memory 아니라 DB 직접 (2026-07-15)**
- `ListProcessesBySid` DB 쿼리를 그대로 감싸는 얇은 래퍼. **memory-first 아님** — memory엔 sid→uid 역인덱스가 없어 전체 entry 스캔이 필요한데 이 조회는 재접속/화면복원 시 1회성이라 인프라 투자 실익이 없고, memory는 휘발성(supervisor 재시작 시 소실)이라 "재시작 복구" 목적과도 안 맞음. `status` 컬럼이 write-through로 최신화되어 있어 DB 단독 조회로 충분(realtime push의 "라우팅=메모리/인가·조회=DB" 분리 원칙과 동일 결).
- `go build`/`go vet` 클린. 호출 지점(재접속 시 프론트가 부를 엔드포인트)은 아직 미배선.
