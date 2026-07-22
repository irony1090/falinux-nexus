# REF — process frontend 트리거(exec/kill REST) + 상태동기화 버그 수정

> 배선 아키텍처는 `REF-process-wiring.md`. 계약/설계 원칙은 `REF-process.md`. 재접속 모델은 `REF-process-reconnect.md`. 세션→uid 원장·REST 구독 배선은 `REF-process-subscription.md`. 작업 이력은 `history/process-trigger.md`.

## frontend 트리거(exec/kill) — REST 배선 구현 완료 (2026-07-16)
> `cmd/supervisor/router/process.go`(`Exec`, `startRelay`, `cleanupProcessTopic`) + `cmd/supervisor/router/processApi.go`(`execProcess`, `killProcess`, `subscribeSid`).

- **REST로 결정**: process 동적 구독을 REST로 확정한 결정(`REF-realtime.md`)과 일관되게, 실행·종료도 소켓 `Handle`이 아니라 `POST /processes/exec` / `POST /processes/kill/:processId`로 배선. 당초 TODO 주석(`register.go`, `process.go`)에 있던 "소켓 Handle(실행요청)" 가정은 폐기.
- **`router.Exec` 시그니처 변경**: `(owner, authKey, kind, node) error` → `(owner, authKey, kind, node, sub process.Subscriber) (uid string, error)`. `sub`는 실행 요청자를 가리키며, 함수 내부에서 자동 구독까지 처리한다(호출자가 uid를 몰라도 됨).
- **자동 구독 = `subscribeSid(uid, sub)` 단일 지점 재사용**: `ProcessManager.SubscribeProcess`(memory+DB write-through) + `browsersForSid`→`subscribeHub.Subscribe`를 묶은 헬퍼(`processApi.go`). 수동 구독(`subscribeProcess` REST)·실행 자동구독(`Exec`)·종료 자동구독(`killProcess`) 셋이 전부 이 지점 하나를 공유 — 세 경로 어디로 들어와도 결과가 동일하다.
- **자동 구독 타이밍이 실행/종료에서 다르다**:
  - **exec**: content 선배치(`SendBuffer`) 직후, **`bind.Relay` 기동(`startRelay`) 이전**에 구독을 건다. relay가 돌기 시작한 뒤에 구독하면 그 사이의 `RUNNING` STATUS나 초기 출력을 놓칠 수 있어서 — 순서가 핵심이다.
  - **kill**: relay는 실행 시점부터 이미 돌고 있으므로 순서 상관없이 `Kill()` 호출 뒤 아무 때나 구독해도 안전. 단, **`killProcess`는 구독해지를 하지 않는다** — `Kill()`은 worker에 신호만 보낼 뿐 실제 종료는 비동기(`MsgStatus`)로 나중에 오므로, 그 자리에서 바로 구독해지하면 아직 안 온 최종 상태를 놓칠 race가 생긴다(아래 종료 후 정리 절 참조).
- **종료 후 Hub 구독 정리 — `startRelay`/`cleanupProcessTopic`**: 기존엔 process가 끝나도(`applyStatus`의 `IsCompleted()` → `Remove(uid)`) `subscribeHub`의 토픽 구독과 `process_subscribers` DB row가 전혀 정리되지 않아, 브라우저 conn이 살아있는 동안 지켜본 process 수만큼 Hub 메모리가 계속 쌓이는 누수가 있었다(정상 종료·kill 공통, kill 전용 문제가 아니었음).
  - **해결**: `bind.NewRelay(...).Start()`를 직접 부르던 두 지점(`Exec`, `reconcileReconnect`)을 `startRelay(uid, inter)` 헬퍼로 통합. `Start()` 직후 `go func(){ relay.Wait(); cleanupProcessTopic(uid) }()`를 붙여, **relay가 완전히 드레인(=마지막 Completed/Failed 발행까지 끝)된 뒤에만** 그 토픽의 Hub 구독자를 전부 `Unsubscribe`.
  - **왜 `applyStatus` 안에서 바로 안 하는가**: `entry.Inter.Done(exit)`는 상태를 큐에 push하고 채널을 닫을 뿐이고, 실제 `Publish`는 별도 `pumpStatus` 고루틴이 비동기로 드레인한다. `applyStatus`가 `Remove(uid)` 직후 바로 Hub를 정리하면 아직 발행 안 된 마지막 이벤트가 지워질 race가 생긴다 — `relay.Wait()`(두 pump가 채널 close까지 완전히 드레인해야 반환)로 "더 이상 발행될 일 없음"을 보장받은 뒤에만 정리해야 안전하다.
  - **DB row는 정리 안 함(의도적)**: `process_subscribers`는 이력으로 남긴다. `ListSubscriptions(sid)`가 상태 무관 전체 반환이지만, 재접속 시 Hub 재구독은 이미 `processManager.Get(uid)` 가드로 죽은 process를 걸러내고 있어(`subscribe.go` `handleSubscribeWS`) 기능상 문제가 없다. DB까지 정리하려면 "구독 이력을 남길 필요가 있는가"부터 별도 결정 필요(미정).

## entry.Record memory 동기화 + kill 종료 이벤트 유실 버그 수정 (2026-07-22)
> 사용자가 frontend exec/kill을 실사용 테스트하며 발견한 세 가지 문제. 상세 이력 → `history/process-trigger.md`. PENDING 오삭제(별개 버그, 이 절 ①의 전제조건) → `REF-process-reconnect.md`.

**① memory `entry.Record`가 생성 시점에 박제되던 문제**
- `applyStatus`의 `MarkProcessRunning`/`Pending`/`Done` 호출이 DB는 갱신하면서도 memory `entry.Record`는 그대로 둬서, `Status`뿐 아니라 `Pid`/`ExitCode`/`StartedAt`/`FinishedAt`까지 전부 최초 생성 시점 값에 박제돼 있었음.
- **해결**: 세 `Mark*` 쿼리가 전부 `RETURNING *`라는 점을 이용, 반환된 row를 필드별로 옮기지 않고 `entry.Record`를 통째로 교체. 교체는 `ProcessEntry.SetRecord(rec *superdb.Process)`(`entry.go`, 전용 `recordMu sync.Mutex`)로 캡슐화 — `subscribers` 필드가 이미 `subMu`+메서드로 캡슐화된 것과 결을 맞춤. DB 쓰기 실패 시엔 memory도 안 건드림(이전엔 에러여도 무조건 Status를 덮어썼음).
  - 콜사이트: `applyStatus`의 `CommandProcess`/`Completed`/`Pending` 3분기(`process.go`) + `execScript`(`manager.go`, `memory.Append` **이후** 시점에 `entry.Record`를 채우던 지점이라 동시쓰기 위험이 있었음).
  - **쓰기만 캡슐화(합의)** — 읽기(`entry.Record.X` 직접 접근)는 그대로 둠. 완전한 레이스 봉쇄는 아니지만(동시 읽기 vs `SetRecord` 교체), `subscribers`만큼 광범위한 동시 읽기 패턴은 아직 없어 우선순위를 낮췄다. 필요해지면 getter까지 확장.
- **왜 지금 필요했나**: `REF-process-reconnect.md`의 PENDING 오삭제 수정("현재 status와 들어온 status가 같으면 무시")이 정확히 동작하려면 memory `Record.Status`가 항상 최신이어야 한다 — 이 SetRecord 작업이 그 전제조건.

**② kill 시 exit code가 유닉스 관례를 안 따르던 문제**
- `Interactive.Kill()`은 프로세스 그룹에 SIGKILL을 보내는데, Go `cmd.Wait()`은 시그널 종료 시 `ProcessState.ExitCode()`가 `-1`을 줘서 "정상 종료했지만 코드가 0이 아닌 진짜 실패"와 구분이 안 됐음(둘 다 `!=0`→`CommandFailed`로만 뭉개짐).
- **해결**: `exitCodeOf(state *os.ProcessState) int`(`execInteractive.go`) 신설 — `WaitStatus.Signaled()`면 `128+시그널번호`(SIGKILL=137, bash/docker/systemd와 동일 관례), 아니면 기존 `ExitCode()`. 상태값(`CommandStatus`, DB `status` 컬럼)은 안 건드림 — `exit_code`만 이 관례를 따르게 됨. 나중에 "kill=별도 터미널 상태" 신설(더 큰 스코프, 보류 중)로 갈 때 `exitCode>=128`을 프론트가 판별 기준으로 바로 쓸 수 있음.

**③ [근본 원인] `pty.Interactive.Status()`가 큐-종료 신호 대신 프로세스 종료 에러를 반환 — kill 시 `applyStatus`까지 이벤트가 아예 안 감**
- 사용자가 kill 후 "`applyStatus`까지 전달이 안 되는 듯"이라고 보고해 추적. `Status()`(`execInteractive.go`)가 세 번째 반환값으로 `i.Error`(`cmd.Wait()`의 종료 에러 — 0 아닌 코드/시그널 종료 시 항상 non-nil)를 주고 있었는데, 소비자인 `pumpStatus`(`cmd/worker/router/process.go`)는 `REF-process.md`의 `Status()` 계약대로 "세 번째 반환값 `!= nil` = 큐 닫힘, 그만 읽어라"로 해석함.
- 정상 종료(코드 0)는 `cmd.Wait()`도 nil이라 우연히 안 걸렸지만, **kill이든 0 아닌 코드로 끝나는 모든 exec든** 마지막 상태를 큐에서 막 꺼내온 순간 조기 `return`에 걸려 **`MsgStatus`를 supervisor에 보내기도 전에, `teardown`도 안 하고 함수가 끝남** — worker `procs`에 고아 엔트리도 남았음.
- **해결**: `Status()`가 `i.Error` 대신 `Shift()` 자신의 에러를 반환하도록 수정(`AgentInteractive.Status()`와 동일 패턴). `IInteractive.Status()` 계약 자체를 `REF-process.md`에 명시해 재발 방지.
- kill 재테스트로 정상 동작(`MsgStatus`→`applyStatus` 도달, DB `FAILED`+`exit_code=137` 반영) 확인 완료.
