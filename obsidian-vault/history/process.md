# history/process — process 실행 모듈 (supervisor 측)

> 설계·재사용 지식 → `REF-process.md` / 현재 진행 → `CURRENT.md`
> 통신 인프라 → `REF-infra.md` / 카탈로그(노드) → `REF-node-label.md` / 실시간 push → `REF-realtime.md`

## 2026-07-01 (2) — process 도메인 배선 + 경로 조립/치환 책임 분리 (build/vet 통과)

골격을 실동작 가능한 배선으로 채웠다. 핵심은 **경로 조립=supervisor / 지역화(placeholder 치환)=worker** 책임 분리로, 이전에 막혀 있던 "선배치 위치 ≠ 실행 위치" blocker를 닫은 것.

**경로 계약 확정 — `{WORKER_BASE}/<node.ID>/<proc.Uid>`**
- 헬퍼 위치 이동: `protocol.WorkerNodePath`(id-only) **폐기** → `cmd/supervisor/process/path.go`의 `WorkerNodePath(node superdb.Node, proc superdb.Process)`. superdb 구조체 의존이라 protocol엔 못 둠(이식성). process pkg가 적격.
- 형식에 `proc.Uid` 포함 = **실행 인스턴스별 격리**(동시/재실행 충돌 방지). process PK=uid라 "process.Id"=Uid(numeric id 없음).
- 이 문자열을 **전송 DestPath(router)와 실행 Cmd/Args(manager)가 공유** → 단일 출처.
- manager `execScript`: Record 생성 전이라 발급된 `uid`로 `WorkerNodePath(node, superdb.Process{Uid: uid})` 조립(= persist 후 Record.Uid와 동일). EXEC `Cmd=path,Args=[]` / EDIT `Cmd={WORKER_EDITOR},Args=[path]`.

**router `process.go` 본문 완성**
- `Exec`: worker 조회 → manager 등록(method-value로 EXEC/EDIT 분기) → folder 조기반환 → **content 선배치**(`SendBuffer`, DestPath=WorkerNodePath, EXEC 0755/EDIT 0644) → relay 기동 → `MsgExec` Call + ExecResponse 검증(실패 시 `Remove` 롤백).
- `output`: DataEvent→`Inter.PushOutput`. `status`→`applyStatus(uid,status,pid,exit)` 코어 분리(단일 깔때기): PROCESS=`MarkProcessRunning`+PushStatus / 종료=`MarkProcessDone`+`Done(exit)`, **EXEC만 Remove, EDIT는 editResult가 teardown**. `editResult`: UID→NodeID→`GetNode` diff→변경 시 `UpdateNodeContent`→Remove.

**worker 측 치환 (transfer.go / process.go)**
- `resolveDest` 재설계: `baseDir/instanceKey` 루트 폐기 → `{WORKER_BASE}→ProcessRoot` 치환 후 ProcessRoot 하위 traversal 검증. exec와 동일 규칙이라 선배치=실행 경로 일치. (부작용: `baseDir` 필드·`instanceKey()` 메서드가 dead code — instanceKey별 격리 포기, 컴파일은 OK)
- `exec`: `localize()`로 Cmd/Args/Cwd/Env의 `{WORKER_BASE}`+`{WORKER_EDITOR}` 둘 다 치환. `resolveEditor()`=$VISUAL>$EDITOR>vi. `workerBase()` 공유 헬퍼. debug `log.Printf` 제거. 실제 PTY 실행은 TODO(now `ExecResponse{Accept:true}` 반환).

**supervisorRouter 배선**: `process *process.ProcessManager` 필드+생성, `handleWorkerWS`에 `On(MsgData)=output`/`On(MsgStatus)=status`/`Handle(MsgEditResult)=editResult` 등록.

**남은 TODO(컴파일 무관)**: worker 실제 PTY 실행(PID보고/출력스트리밍/종료status/EDIT read-back/$TERM·TUI강제) / router frontend 제어(input/resize/kill)·worker끊김→PENDING 합성·재바인딩·ring buffer.

## 2026-07-01 — supervisor ProcessManager·bind·router 골격 + 종료/재접속 모델 확정

세션 내내 설계 대화로 역할분리와 종료/재접속 모델을 확정하고, supervisor 측 골격을 구현했다(전부 빌드·vet 통과).

**구현(골격)**
- `cmd/supervisor/process/manager.go` — `ProcessManager`(memory=`KeyValManager[uid,*ProcessEntry]`+pool). `Exec`(folder=memory only / script=memory+pool+AgentInteractive), `ExecEdit`(SCRIPT 검증→EDIT), `execScript`(공통 코어), `openFolder`, `newWorkerInteractive`(onWrite=`Emit(MsgData)` / onLayout·onKill=`Call`), `Get`, `Remove`(`Done(502)` 안전망), `createKey`. **spec 파라미터 제거 → node로부터 내부작성**.
- `cmd/supervisor/process/entry.go` — `ProcessEntry{Record,Inter}` + `HasProcess()/NodeID()/Spec()`.
- `internal/supervisor/bind/relay.go` — `Relay`(`NewRelay/Start/Wait/pumpOutput/pumpStatus`), `Output` 배치 드레인.
- `internal/syncProcess/syncData.go` — `ShiftAll()` 배치 드레인 추가.
- `internal/execute/iInteractive.go`(+구현 3: AgentInteractive/Interactive/Fifo) — `OutputAll()`(연결바이트) 추가. relay가 이걸 사용.
- `cmd/supervisor/router/process.go` — 골격: `Exec`(오케스트레이션 트리거) / `output`·`status`(EVENT 수신) / `editResult`(REQ, EDIT read-back). **구 `exec`·`execEdit`(spec 받던 것) 삭제**.
- `internal/protocol/messages.go` — `PlaceholderWorkerEditor = "{WORKER_EDITOR}"` 추가.

**확정 설계 결정** (상세 → `REF-process.md` 2026-07-01 절)
- 역할분리: manager=상태 / router=wire·트리거 / bind.Relay=fan-out. "누구=router, 무엇=manager".
- spec은 Record 파생(`entry.Spec()`), 별도 상주 금지. content=파일 선배치 + 공유 헬퍼 `WorkerNodePath`. EDIT `Cmd={WORKER_EDITOR}`.
- **종료/재접속 개정**: `status` 단일 깔때기 / worker 끊김→**PENDING**(구 `Done(502)` 폐기) / frontend 끊김≠종료 + 같은 세션 화면복원(ring+세션원장) / 재접속 conn 재바인딩.

**미배선(→ CURRENT.md 다음 배선 목록)**: router `process` 필드, `status` applyStatus 분리, On/Handle 등록, worker 끊김 PENDING, 재바인딩, ring/세션원장, `WorkerNodePath` 헬퍼, EXEC content→실행 정책.
