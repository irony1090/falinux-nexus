# history/process-reconnect — 종료/재접속 모델 (worker 끊김→PENDING→재바인딩)

> 설계·재사용 지식 → `REF-process-reconnect.md` / 계약·원칙 → `REF-process.md` / 현재 진행 → `CURRENT.md`
> supervisor 아키텍처·worker 실행부 배선 이력은 → `history/process-wiring.md` / frontend 트리거·버그수정 이력은 → `history/process-trigger.md`
> 세션→uid 원장·REST 구독 배선 이력은 → `history/process-subscription.md`

## 2026-07-22 — PENDING 오삭제 버그(정상 실행 vs 끊김 합성 혼동) 수정

사용자가 frontend에서 process를 실행하면 worker가 PENDING을 보고하자마자 supervisor가 즉시 `processManager`에서 지워버려 정지된 것처럼 보인다고 보고 → 조사 착수(fork 조사 에이전트로 `applyStatus`/`pty.ExecInteractive` 흐름 추적).

**원인**: `worker의 pty.ExecInteractive`가 실행 시작 시 `CommandPending`을 항상 먼저 push(정상 시퀀스의 일부)하는데, `applyStatus`의 `CommandPending` 분기는 "끊김 시 supervisor 합성 호출 전용"이라는 주석상 가정(2026-07-14 설계, `REF-process-reconnect.md` 구현 목록 표 하단)에 기대 무조건 `Remove(uid)`를 호출하고 있었음 — 정상 실행 직후 보고와 끊김 합성 호출이 같은 분기를 공유해 구분이 안 됨.

**논의·구현 흐름**
1. 사용자가 "entry의 현재 상태가 PENDING인데 또 PENDING을 보내면 무시" 아이디어 제시 → memory `entry.Record.Status`가 생성 이후 갱신된 적 없는(항상 `PENDING`) 박제 필드라 이대로 적용하면 **진짜 끊김 케이스까지 다 무시**돼 재접속 `Rebind`가 "이미 재바인딩된 process" 에러를 내는 부작용을 짚어줌.
2. 사용자가 일반화된 대안("모든 상태에 대해 현재==들어옴이면 무시") 제시 → 두 가지 다 memory `Record.Status`가 실시간이어야 성립함을 재확인.
3. 사용자가 직접 가드(`applyStatus` 진입부, `if ok && entry.Record.Status == status.String() { return }`)를 구현 → 코드 리뷰로 정상 exec/끊김/재접속 세 경로를 트레이스해 검증(가드 자체는 올바르게 동작).
4. 이어서 "memory entry가 항상 DB 값과 동일해야 한다"는 사용자 요구로 `entry.Record` 전체 동기화 작업(→ `history/process-trigger.md` "2026-07-22" 참조)까지 같은 세션에서 이어짐 — 이 가드가 정확히 동작하기 위한 전제조건이었기 때문.

상세 설계 근거 → `REF-process-reconnect.md` "PENDING 오삭제 버그" 절. `go build`/`go vet` 클린, 정상 exec 후 process 유지 확인 + kill 재테스트로 최종 검증.

## 2026-07-14 (2) — worker 끊김→PENDING→재접속 재바인딩 구현 완료 + e2e 검증

아래 "설계 확정" 절 그대로 구현. 상세 = `REF-process-reconnect.md` "worker 끊김 → PENDING → 재접속 재바인딩" 절(구현 목록 표 + `applyStatus` 가드 완화 + `WorkerState` 두 하위 절 포함).

**변경 파일**
- `internal/supervisor/db/query/processes.sql` — `MarkProcessPending` (+sqlc generate)
- `internal/protocol/messages.go` — `MsgSync`/`SyncEntry`/`SyncEvent`
- `cmd/supervisor/process/manager.go` — `ProcessManager.Rebind`
- `cmd/supervisor/router/process.go` — `applyStatus` 가드 완화+`CommandPending` 분기, `reconcileDisconnect`, `sync` 핸들러, `reconcileReconnect`
- `cmd/supervisor/router/supervisorRouter.go` — 끊김 시 `reconcileDisconnect` 호출 + `MsgSync` 핸들러 등록
- `cmd/worker/router/workerRouter.go` — `WorkerState`(재접속 넘어 사는 procs+conn) 신설
- `cmd/worker/router/process.go` — pump/teardown이 `state.currentConn()` 참조하도록 변경
- `cmd/worker/router/register.go` — 등록 성공 직후 `sendSync()`
- `cmd/worker/main.go` — `WorkerState`를 재접속 루프 밖에서 1회 생성해 전달

**구현 중 발견해 원 설계에 없던 것 두 가지** (상세는 REF 두 하위 절)
1. `applyStatus`가 memory entry 없으면 조기 리턴하던 가드가 "재접속·supervisor만 앎" 분기(entry가 끊김 처리 때 이미 Remove된 상태)를 막고 있어 DB가 Failed로 안 닫히는 공백 발견 → 가드를 case별로 완화(사용자에게 두 대안 제시 후 "가드 완화" 채택).
2. `cmd/worker/main.go`가 재접속마다 `NewWorkerRouter`를 새로 호출하며 `procs`(PTY 핸들 맵)를 매번 새로 만들던 기존 구조가 "재접속해도 procs가 산다"는 설계 전제와 충돌 발견(구현 도중 발견, 사용자에게 "지금 같이 고칠지" 확인 후 진행) → `WorkerState` 도입으로 해소.

**e2e 검증(수동, 실제 프로세스 kill/restart)**: 로컬 supervisor+worker 기동(postgres15 컨테이너, node id=2 스크립트를 `sleep N`으로 임시 교체) 후
- worker 프로세스 kill → supervisor 로그에 `E N D`, DB `PROCESS→PENDING` 확인.
- (PENDING 상태에서) 빈 procs로 새 worker 재접속 → DB `PENDING→FAILED(exit_code=502)` 확인("supervisor만 앎" 분기 = 가드 완화가 실제로 살리는 경로).
- worker는 살려둔 채 supervisor만 kill/재기동 → 재접속 시 DB가 `PROCESS`로 재차 갱신(`updated_at` 확인, Rebind+relay 재기동 성공) → 그 프로세스가 자연 종료(sleep 만료)할 때까지 대기해 **새 conn을 통해** `COMPLETED, exit_code=0`까지 정상 보고됨을 확인(교집합/Rebind 경로 + worker측 conn 스왑 둘 다 실증).
- 테스트 후 node content 원복, 테스트 프로세스 정리. `go build`/`go vet`/`gofmt -l .` 클린.

## 2026-07-14 (1) — worker 끊김→PENDING→재접속 재바인딩 설계 확정 (코드 변경 없음, 순수 설계)

CURRENT.md "다음 배선" 2번 착수를 위한 설계 대화. 결과는 `REF-process-reconnect.md` "worker 끊김 → PENDING → 재접속 재바인딩" 절에 반영(이 절이 최신, 구 "⚠️ 재접속 conn 재바인딩(난점)" 대체).

**핵심 결정**
- 재접속 시 죽은 conn을 캡처한 `AgentInteractive` 콜백을 "교체"하려던 기존 난제를 폐기 — 대신 끊기면 **버리고**(`Inter.Done(sentinel)`로 `bind.Relay` 고루틴 정지 후 `processManager.Remove`), 재접속하면 **새로 만든다**(`newWorkerInteractive` 재호출). 훨씬 단순.
- 3-way 재동기화는 memory가 아니라 **DB(`ListActiveByDevice`, 기존 존재하나 미사용이던 쿼리)를 진실로 삼음** — memory에 Record-only placeholder를 남겨둘 필요 없음이 드러남(사용자 질문으로 발견).
- **FOLDER 타입은 원천적으로 안전**: `openFolder()`가 애초에 `CreateProcess`(DB persist)를 안 부름 → `ListActiveByDevice`에 안 걸림 → 끊김 처리 루프가 자동으로 건드리지 않음(별도 필터 불필요). 근거는 `entry.go` `HasProcess()` 기존 주석("folder는 제외 — frontend 끊김 기준")과 일치.
- worker→sup 재접속 보고는 `RegisterRequest`에 얹지 않고 별도 `MsgSync{uid,status,pid}[]`로 분리(식별자 핸드셰이크와 도메인 상태 동기화는 관심사 분리).
- 3-way 대조 처리: 교집합=`Rebind`(신규 메서드)로 재장착, supervisor만 앎=`Failed`+sentinel로 종결(LOST 상태 신설 안 함), worker만 앎(고아)=로그만(YAGNI).

**구현·검증은 위 2026-07-14 (2) 참조.**
