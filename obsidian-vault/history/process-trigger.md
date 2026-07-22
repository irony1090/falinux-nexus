# history/process-trigger — frontend REST 트리거(exec/kill) + 상태동기화 버그 수정

> 설계·재사용 지식 → `REF-process-trigger.md` / 배선 아키텍처 → `REF-process-wiring.md` / 계약·원칙 → `REF-process.md` / 현재 진행 → `CURRENT.md`
> 재접속 모델 이력 → `history/process-reconnect.md` / 세션원장·REST구독 이력 → `history/process-subscription.md`

## 2026-07-22 — entry.Record memory 동기화(SetRecord) + kill 종료 이벤트 유실 버그 수정 (build/vet 통과, kill 실사용 검증 완료)

사용자가 frontend에서 실행→kill을 직접 테스트하며 순서대로 발견한 3건. 대화로 원인 추적 후 사용자 승인 하에 구현. 상세 → `REF-process-trigger.md` "entry.Record memory 동기화 + kill 종료 이벤트 유실 버그 수정" 절. PENDING 오삭제(별개, 순서상 먼저 발견)는 `history/process-reconnect.md` "2026-07-22" 참조.

**① `entry.Record` memory 동기화** — `applyStatus`(`process.go`) `CommandProcess`/`Completed`/`Pending` 3분기가 `q.MarkProcess*` 반환값(`RETURNING *`)을 버리던 걸 캡처해 `entry.Record`를 통째로 교체하도록 변경. 처음엔 `apply()`(Status 필드만 손복사)로 시작했다가, 사용자가 "Pid/ExitCode 등 다른 필드도 항상 DB와 같아야 한다"고 지적 → `Mark*`가 이미 `RETURNING *`임을 확인하고 반환 row 통째 대입으로 전환. 이어서 "쓰기 경로에 setter가 필요하지 않냐"는 사용자 질문에 `subscribers` 필드의 기존 캡슐화 패턴(뮤텍스+메서드)과의 비일관성·동시쓰기 레이스 가능성을 근거로 동의 → `ProcessEntry.SetRecord()`(`entry.go`, 전용 뮤텍스) 신설, `process.go` 3곳 + `manager.go`(`execScript`) 1곳 교체. 읽기는 캡슐화 범위 밖으로 명시적으로 남김(사용자 선택, "쓰기만 setter로 좁혀서").

**② kill exit code 유닉스 관례화** — 사용자 질문("사용자에 의한 종료도 FAILED로 나오는데 실제 OS process 정책상 그게 맞아?")으로 시작. `cmd.Wait()`이 시그널 종료 시 `ExitCode()=-1`을 주는 Go 동작과, bash/docker/systemd가 쓰는 `128+시그널` 관례를 대조해 두 범위(작은 수정=exit code만 관례화 / 큰 수정=별도 터미널 상태 신설)를 제시 → 사용자가 작은 쪽 선택. `exitCodeOf()` 헬퍼 신설(`execInteractive.go`).

**③ `pty.Interactive.Status()` 에러 계약 버그(근본 원인)** — 사용자가 "killProcess로 종료했을 때 applyStatus까지 전달이 안 되는 듯"이라고 보고. `Status()`가 큐 종료 신호 자리에 `cmd.Wait()`의 프로세스 종료 에러(`i.Error`)를 반환하고 있어서, 소비자 `pumpStatus`(`cmd/worker/router/process.go`)의 "err != nil = 그만 읽어라" 관례와 충돌 — kill이든 0 아닌 코드의 자연 종료든 마지막 상태 이벤트가 `MsgStatus` 발송 전에 조기 리턴으로 유실되고 worker `procs` teardown도 누락되던 버그. `AgentInteractive.Status()`(정상 구현)와 대조해 원인 특정 → `Shift()` 자신의 에러를 반환하도록 한 줄 수정. `REF-process.md`에 `Status()` 계약을 명문화해 재발 방지.

**검증**: 세 수정 모두 `go build`/`go vet` 클린. kill 실사용 테스트로 최종 확인(`MsgStatus`가 `applyStatus`까지 도달, `FAILED`+`exit_code=137` DB 반영).

## 2026-07-21(2) — `processDto.go` 신설: `listSubscriptions` 응답 DTO 정정 (build/vet 통과)

프론트 `process.api.ts` 작성 중 발견: `listSubscriptions`가 DTO 없이 `superdb.Process`(sqlc raw, json 태그 없음)를 그대로 `c.JSON`하고 있어서 실제 와이어가 PascalCase(`Uid`,`NodeID`,...) + `Timestamptz`는 unix sec가 아니라 RFC3339 문자열로 나가던 상태(node/user의 DTO 컨벤션과 어긋남). `nodeDto.go`와 대칭으로 `processDto.go` 신설: `processResponse`(camelCase json 태그, nullable은 포인터, `ownerUserId`는 nodeResponse처럼 제외) + `newProcessResponse(s)`. `processApi.go`의 `listSubscriptions`가 이걸 쓰도록 교체. 프론트 작업 → `history/frontend.md` "2026-07-21".

## 2026-07-21(1) — 토픽 접두사 `PROC:`→`PROCESS:` 정정 (`process.go:101`, 사용자 직접 수정)

`processTopic()`이 쓰던 접두사를 `PROC:`→`PROCESS:`로 변경. 실사용처는 이 함수 하나뿐이라 로직 영향 없음(단일 출처). `bind/relay.go`·`process.go` 안의 예시/설명 주석 2곳만 옛 문자열 남아있어 같이 갱신 대상으로 확인(코드 주석이라 사용자 수정 대기). REF 반영 → `REF-process-wiring.md` "bind.Relay / 배치 드레인" 절.

## 2026-07-16 — frontend 트리거(exec/kill) REST 배선 + 종료 후 구독 정리 (build/vet 통과)

**exec/kill REST 핸들러 (`processApi.go`)**
- `POST /processes/exec`(`execProcess`): `{nodeId, authKey, type?}` 바인딩 → `TxQueries(c).GetNode` → `router.Exec(owner, authKey, kind, node, sub)` 위임 → `{uid}` 응답.
- `POST /processes/kill/:processId`(`killProcess`): `processManager.Get(uid)` → `entry.Inter.Kill()` → 요청자 자동구독.
- `subscribeSid(uid, sub)` 헬퍼 신설: `SubscribeProcess`(memory+DB) + `browsersForSid`→`subscribeHub.Subscribe`를 묶음. 기존 `subscribeProcess`(수동 구독)도 이걸 쓰도록 리팩터 — 수동/실행/종료 세 경로가 구독 로직 단일 지점 공유.

**`router.Exec` 시그니처 변경 (`process.go`)**
- `(owner, authKey, kind, node) error` → `(owner, authKey, kind, node, sub process.Subscriber) (string, error)`. content 선배치 직후·relay 기동(`startRelay`) 이전에 `subscribeSid` 호출 — relay가 돌기 전에 구독을 걸어야 `RUNNING` 등 초기 이벤트를 놓치지 않는다(순서가 핵심).

**종료 후 Hub 구독 정리 — 사용자 질문("killProcess에선 구독해지 안 해?")으로 발견**
- 기존엔 process가 끝나도(`applyStatus` `IsCompleted()`) Hub 토픽 구독·DB subscriber row가 전혀 정리 안 돼, 브라우저가 지켜본 process 수만큼 Hub 메모리가 계속 쌓이는 누수가 있었음(kill 전용이 아니라 정상 종료도 동일).
- `killProcess` 안에서 즉시 구독해지하면 `Kill()`(신호 전송, 동기)과 실제 종료 확인(`MsgStatus`, 비동기) 사이 race로 마지막 상태 이벤트를 놓칠 수 있어 그 자리에서 처리 불가 — 대신 `bind.NewRelay(...).Start()`를 직접 부르던 두 지점(`Exec`, `reconcileReconnect`)을 `startRelay(uid, inter)`로 통합하고, `Start()` 직후 `go{ relay.Wait(); cleanupProcessTopic(uid) }`를 붙여 **relay가 완전히 드레인된 뒤에만**(=마지막 Completed/Failed 발행까지 끝난 뒤) Hub 구독을 해제하도록 함.
- `process_subscribers` DB row는 이력으로 남김(정리 안 함) — 재접속 Hub 재구독은 이미 `processManager.Get` 가드가 죽은 process를 걸러내 기능상 문제 없음.

상세 → `REF-process-trigger.md` "frontend 트리거(exec/kill)" 절.

