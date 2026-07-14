# history/process — process 실행 모듈 (supervisor 측)

> 설계·재사용 지식 → `REF-process.md` / 현재 진행 → `CURRENT.md`
> 통신 인프라 → `REF-infra.md` / 카탈로그(노드) → `REF-node-label.md` / 실시간 push → `REF-realtime.md`

## 2026-07-14 (2) — worker 끊김→PENDING→재접속 재바인딩 구현 완료 + e2e 검증

아래 "설계 확정" 절 그대로 구현. 상세 = `REF-process.md` "worker 끊김 → PENDING → 재접속 재바인딩" 절(구현 목록 표 + `applyStatus` 가드 완화 + `WorkerState` 두 하위 절 포함).

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

CURRENT.md "다음 배선" 2번 착수를 위한 설계 대화. 결과는 `REF-process.md` "worker 끊김 → PENDING → 재접속 재바인딩" 절에 반영(이 절이 최신, 구 "⚠️ 재접속 conn 재바인딩(난점)" 대체).

**핵심 결정**
- 재접속 시 죽은 conn을 캡처한 `AgentInteractive` 콜백을 "교체"하려던 기존 난제를 폐기 — 대신 끊기면 **버리고**(`Inter.Done(sentinel)`로 `bind.Relay` 고루틴 정지 후 `processManager.Remove`), 재접속하면 **새로 만든다**(`newWorkerInteractive` 재호출). 훨씬 단순.
- 3-way 재동기화는 memory가 아니라 **DB(`ListActiveByDevice`, 기존 존재하나 미사용이던 쿼리)를 진실로 삼음** — memory에 Record-only placeholder를 남겨둘 필요 없음이 드러남(사용자 질문으로 발견).
- **FOLDER 타입은 원천적으로 안전**: `openFolder()`가 애초에 `CreateProcess`(DB persist)를 안 부름 → `ListActiveByDevice`에 안 걸림 → 끊김 처리 루프가 자동으로 건드리지 않음(별도 필터 불필요). 근거는 `entry.go` `HasProcess()` 기존 주석("folder는 제외 — frontend 끊김 기준")과 일치.
- worker→sup 재접속 보고는 `RegisterRequest`에 얹지 않고 별도 `MsgSync{uid,status,pid}[]`로 분리(식별자 핸드셰이크와 도메인 상태 동기화는 관심사 분리).
- 3-way 대조 처리: 교집합=`Rebind`(신규 메서드)로 재장착, supervisor만 앎=`Failed`+sentinel로 종결(LOST 상태 신설 안 함), worker만 앎(고아)=로그만(YAGNI).

**구현·검증은 위 2026-07-14 (2) 참조.**

## 2026-07-13 — worker 실행부 본체 구현 완료 + Cwd 배선 (build/vet 통과, e2e 스모크 확인)

`cmd/worker/router/process.go` 실체화 + `workerRouter.go` 배선. 계획 → REF-process.md 2026-07-03절 "본체 계획" 그대로 따르되 세부는 완료 절로 이전 반영(→ `REF-process.md` "worker 실행부" 절 최신).

**구현**
- `procEntry{uid, kind, editPath, inter *pty.Interactive}` + `workerRouter.procs KeyValManager[string,*procEntry]`(`saves` 옆).
- `exec()`: localize(기존 유지) → `env=os.Environ()+body.Env+TERM=xterm-256color` → `pty.ExecInteractive(ctx,cmd,env,dir,args...)` → `procs.Append` → `go pump` → `ExecResponse`.
- `pump` = `go pumpOutput`(별도 고루틴, `OutputAll()`→`Emit(MsgData)`) + `pumpStatus`(호출 고루틴 자신, `Status()`→`Emit(MsgStatus)`, PROCESS 전이 시 `Pid()` 포함). `pumpStatus`가 `IsCompleted()` 감지 시 `teardown` 호출 후 종료 — **teardown 단일 지점**.
- `teardown`: EDIT면 `os.ReadFile(editPath)`→`conn.Call(MsgEditResult)` → `procs.Remove(uid)`(EXEC/EDIT 공통 마지막 단계).
- `resize`/`kill`/`input` 실구현: `procs.Get(uid).inter.Layout/Kill/Write`. **kill은 Remove 안 함**(이중정리 방지, pumpStatus가 감지해 teardown).
- `workerRouter.go`: `conn.On(MsgData)=input` / `Handle(MsgResize)=resize` / `Handle(MsgKill)=kill` 등록 추가.

**Cwd 배선(부수 작업)**
- `pty.ExecInteractive`에 `dir string` 파라미터 추가(`env` 다음·`args` 앞). 빈 값=cwd 상속(기존 env nil-상속 정책과 결 맞춤). 콜사이트 worker `exec()` 유일이라 파급 없음. traversal 검증은 안 함(Cmd/Args와 대칭 유지 — 기존에도 안 하던 것).

**e2e 스모크 확인**: SCRIPT 노드 content `#!/bin/bash\nhtop` 실행 → worker PTY 기동, `htop` 정상 구동 확인(출력 스트리밍 = worker 쪽 Emit 직전 로그로 확인).

**발견 1 — `newWorkerInteractive`(supervisor manager.go) 콜백 로그와 exec는 별개 경로**: `onWrite`/`onLayout`/`onKill`(NEW_INTER ARGS1/2/3)은 `Inter.Write/Layout/Kill` 호출 시에만 찍힘 — exec 성공 자체와 무관. 지금 frontend 트리거·제어(input/resize/kill 발신)가 미배선이라 저 콜백은 아직 아무도 안 부름. exec 확인은 worker `[WORKER-EXEC]` 로그나 MsgData/MsgStatus 수신으로 해야 함.

**발견 2 — PROC 토픽 무구독 (미해결, 다음 선행조건)**: worker→sup 출력은 `bind.Relay`까지 정상 도달하지만 `subscribeHub.Publish("PROC:"+uid,...)`가 **구독자 0명이면 조용히 폐기**(에러·로그 없음, `internal/subscribe` Publish 구현). 프론트가 `PROC:<uid>`를 동적 구독하는 경로가 아직 없어(현재 `NODE:0` 고정구독뿐) 출력이 여기서 끊김 — supervisor~프론트 구간은 실제로 검증 못 함. → REF-realtime.md "동적 구독 어휘" 선행 필요.

**참고**: 세션 시작 시점에 `cmd/supervisor/router/subscribe.go`/`apps/frontend/src/pages/index.vue`의 `TEST`/`TTTT` 스모크 핸들러가 이미 worktree에서 제거돼 있었음(미커밋, 이번 세션 작업 아님) — REF-realtime.md 갱신해 반영.

## 2026-07-03 — folder-open 버그 수정 + worker EDITOR env + PTY 실행부 선행작업 (build/vet 통과)

worker 실제 PTY 실행부 착수 전 준비 작업. 세 갈래.

**① folder-open 버그 수정 (2026-07-02 발견분 클로즈)**
- `router.Exec`의 사전 게이트(`worker, exist := r.workers.Get(authKey); if !exist { return err }`) 제거 → `worker, _ := r.workers.Get(authKey)`(nil 허용). worker 필요 검증은 manager로 단일 위임.
- `manager.Exec`: folder 분기(`openFolder`, worker 무접촉) **뒤에** `if worker == nil` 검사 추가. `ExecEdit`도 동일 검사 선행 추가. `execScript`엔 기존부터 있던 검사 유지(3중이지만 createKey 전 조기차단+에러위치 명확 이점, 무해).
- 결과: device 오프라인이어도 카탈로그 폴더 열람 가능. SCRIPT/EDIT은 worker 없으면 종전대로 에러.

**② worker EDITOR env 필드**
- `cmd/worker/constants/env.go`: `EnvVars`에 `Editor string \`iniName:"EDITOR"\`` 추가 + `LoadEnv`에서 `EDITOR`를 **선택값**으로 로드(비면 리로드 시 기존값 유지, 최종 폴백은 resolveEditor). `.env`의 godotenv.Overload가 OS env도 세팅.
- `resolveEditor`(worker/router/process.go): 폴백 체인 재정의 → **`env.Editor > "vi"`**(사용자가 구 `$VISUAL>$EDITOR` OS 조회 제거, `os` import도 제거). `{WORKER_EDITOR}` 치환값 출처.

**③ pty 실행 primitive 확장 (worker 실행부 선행)**
- `execute/pty/execInteractive.go` `ExecInteractive`에 **`env []string`** 파라미터 추가(`command` 뒤, `args` 앞). **nil=os.Environ() 상속**(cmd.Env 미설정) / **목록=완전 대체**. → 호출자(worker exec)가 `os.Environ()+spec.Env+TERM`을 조립해 넘기는 정책. TUI(vi)엔 `TERM=xterm-256color` 사실상 필수.
- `pty.Interactive`에 **`Pid() int`** 접근자 추가(미기동/종료 시 -1). MsgStatus(PROCESS)로 PID 보고용. **`IInteractive` 인터페이스엔 넣지 않음** — PID는 worker 로컬 개념, 원격 래퍼(AgentInteractive)엔 실 PID 없음. worker는 `*pty.Interactive` 구체타입 보유라 직접 호출.

**남은 것(다음)**: worker 실행부 본체 = `workerRouter.procs`(uid→*procEntry) 매니저 + `exec()` 실체화 + 출력/상태 **pump goroutine** + input/resize/kill + EDIT read-back. 계획 → CURRENT.md / REF-process.md "worker 실행부" 절.

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
