# REF — process 실행 모듈 (execute / pty / agent comm)

> worker 프로세스 실행/모니터링/종료(PTY) 설계와 계승 자산. **설계 확정 + supervisor 측 도메인 배선 완료(build/vet 통과). 남은 것=worker 실제 PTY 실행 + router 제어/재접속 배선.**
> 출력 스트리밍 토대는 `REF-infra.md` EVENT 평면. 통신 상세는 `PLAN-agent-comm.md`.

## Agent↔Server 통신 프로토콜 (계승)
- `agentId` = 에이전트 식별자 (PortBridge에선 MAC 주소, HTTP 헤더로 전달 → 현재는 InstanceKey)
- `SessionId` = 사용자/세션 식별자
- key 형식: `agentId:processId` (또는 `agentId:userId:processId`)
- **Agent → Server**: `MsgStatus`(RUNNING/STOPPED + ExitCode), `MsgData`(output)
- **Server → Agent**: `MsgExec`, `MsgData`(input), `MsgResize`, `MsgKill`, `MsgUploadInit/Chunk`
- 파일 전송 흐름: `UploadInit → Ready → Chunk → Status → Result`

## Process 실행 추상화 (worker agent 핵심)
- **`IInteractive` 인터페이스**: `Output()`, `Write()`, `Status()`, `Kill()`, `Layout() syscall.Errno`, `ExitCode() int`
  - `*Interactive` (로컬 PTY 구체 타입), `*AgentInteractive` (원격 agent 래퍼) 모두 구현
- **`AgentInteractive`**: agent WebSocket을 통한 원격 process 래퍼
  - `onWrite`/`onLayout`/`onKill` 콜백으로 agent에 명령 전달
  - `PushOutput()`/`PushStatus()`/`Done(exitCode)` 외부 주입 메서드
  - `sync.Once`로 `Done()` 중복 호출 방지
  - 생성 시 초기 status push 없음 → RUNNING 수신 시 첫 push

## Process 실행 흐름 (동기화 패턴)
```
CreateProcess → process 생성 + AgentInteractive 생성 + (필요시 파일 전송)
  → ExecPayload 전송 → agent 실행
  → MsgStatus(RUNNING) → PushStatus → Status() 블로킹 해제 → RUNNING 반영
  → bindManager에 등록 → goroutine 대기
  → MsgData → PushOutput → UI/상위 스트리밍
  → MsgStatus(STOPPED) → Done(exitCode) → goroutine 종료
```

## Supervisor가 agent 관리하는 패턴
- agent 연결/해제 시 콜백(`OnCreated`/`OnRemoved`)으로 상태 갱신 + 브로드캐스트
- agent 연결 끊김 시 해당 agent 소속 모든 interactive를 `Done(502)`로 정리 (defer cleanup)
- agent 상태를 `agentId(string)` 기반 싱글턴으로 관리 → DB 조회 없이 매핑

---

## process 모듈 설계 (2026-06-25 확정, Step 1 구현 시작)
- **★발견: `execute`는 스텁 아님** — PortBridge 이식 완료/미배선. 보유: `IInteractive`(Output/Write/Status/Kill/Layout/ExitCode), `*Interactive`(로컬 PTY — **직접 만든 `/dev/ptmx` ioctl, creack/pty 불필요, Linux 전용**), `*AgentInteractive`(원격 래퍼 onWrite/onLayout/onKill+Push*/Done), 기타 `ExecCommand`/`Fifo`/`TmuxSession`, `syncProcess.SyncData[T]`. 빠진 것=신원(UID/PID)·매니저·배선
- **★패키지 분리(2026-06-25)**: `execute`가 통째 Linux 전용이면 supervisor(AgentInteractive만 필요)까지 비-Linux 빌드 불가 → **OS syscall 의존 파일을 `execute/pty`(package pty)로 분리**. `execInteractive.go`(PTY)·`execFIFO.go`(Mkfifo) 이동. `execute`(이식 가능)=IInteractive/AgentInteractive/CommandStatus/ExecCommand/Tmux 유지. pty는 execute를 import(CommandStatus 등), `var _ execute.IInteractive = (*Interactive)(nil)` 어서션. 공유 변수 `rn`/`nr`는 분리 후 양쪽 각자 보유
- **상태 어휘**: `execute.CommandStatus`(Pending=0/Process=1/Completed=2/Failed=3)를 **와이어에서 그대로 사용**(`StatusEvent.Status`). execute가 이식 가능해져 `protocol`이 import 가능 → **ProcStatus 미러는 폐기**(단일 소스). transport→protocol→execute 그래프에 execute 이식부가 딸려오나 전부 portable이라 무방
- **구조체 공유 = 데이터 서술자만**: `ProcessSpec{UID,Cmd,Args,Cwd,Env,Rows,Cols}`는 **`protocol`에 둠**(execute 아님 — 이식성 이유, 순수 데이터라 protocol이 적격, execute가 필요시 protocol을 import=사이클 없음). 런타임 핸들은 각자 → 행위는 `IInteractive`로 통일. **통째 동일 구조체 박지 말 것**(런타임 핸들 다름)
- **Step 1 완료(2026-06-25)**: `protocol/messages.go`에 MsgExec/Data/Resize/Kill/Status + ProcessSpec/ExecResponse/DataEvent/ResizeRequest/KillRequest/ProcStatus/StatusEvent 추가. 제어=REQ/RES, 스트림=EVENT(streamID=UID). MsgData는 양방향(입력 sup→worker/출력 worker→sup). StatusEvent.PID=RUNNING 시, ExitCode=Done 시(omitempty 금지)
- **UID vs PID 분리**: UID=시스템 전체 정체/라우팅 키(**initiator 발급**=supervisor top-down 또는 worker 자기 프론트, RandomKey 전역유일, **모든 메시지에 실림**) / PID=worker 로컬 OS 핸들(worker 발급, RUNNING 시 status에 얹어 위로 보고만, 와이어 라우팅 안 함). **양쪽 manager 모두 UID로 키잉**
- **평면 매핑**: 제어(Exec/Kill/Resize)=REQ/RES, 스트림(Data/Status)=EVENT(streamID=UID). 입력 키스트로크=EVENT
- **fan-out 허브 = `IInteractive` 위에서 재사용**: 출력 → ring buffer(SNAPSHOT용) + subscribe.Hub(라이브) + DB sink(영속). 허브는 sink 모름(bind 계층이 배선). **공유 위치 + IInteractive 파라미터화**(supervisor 전용 금지) → worker 자기 프론트도 나중에 mount만으로 붙음(YAGNI: 지금은 안 막기만)
- **frontend 다리**: 브라우저도 `transport.Conn`(gorilla=MessageRW). 명령=`Call`/`Handle`(REQ/RES), 입력=`Emit`(EVENT), 출력 fan-out=`subscribe.Hub.Publish`(send 전략=`conn.Emit`). 구독=**WS Call로 통일**(REST 안 섞음). SNAPSHOT(스크롤백)+UPDATE(라이브)
- **느린 소비자 격리(frontend 한정)**: worker→sup `On(Data)` 핸들러가 Publish 동기호출 시 느린 브라우저 1개가 디스패처까지 막음 → **per-client 바운드 큐+writer goroutine**, Emit은 넣기만. 큐 차면 **그 클라만 끊고 재구독 유도**(터미널은 중간유실=화면깨짐이라 끊고 SNAPSHOT 재전송이 깔끔). 저빈도 list 토픽은 동기 OK
- **인코딩**: 일단 Frame(JSON) 통일. 출력 바이트 base64 ~33% 부풀음 문제되면 **제어=Frame/출력=raw WS 바이너리** 분리(추후 여지만 남김)
- **process 영속성**: worker 휘발(메모리) / supervisor 영속(PG, 재시작 복구)
- **재연결 reconciliation**: 끊김→`Done(502)` 비관적 정리 / 재연결→worker가 live 스냅샷 보내 재동기화. 미구현(supervisor 본격 작업 시)

## 다음 (도메인 배선)
- MsgData(stdout/stderr)·MsgStatus(RUNNING/STOPPED+ExitCode) payload(streamID=UID) → worker `Emit`/supervisor `On` → `AgentInteractive.PushOutput`/`Done(exitCode)` 연결

## 실행 타입: EXEC vs EDIT (2026-06-26 확정) — script 편집 = worker PTY vi 왕복
> node script 편집 = frontend→supervisor→worker로 worker의 실제 `vi`($EDITOR)를 **PTY로 띄워** 편집, 종료 시 내용 회수. **PTY 엔진의 특수 사례** — 새 메커니즘 아님. 카탈로그(`REF-node-label.md`) "무엇"에 "어떻게(편집)"를 먹이는 동작.

- **단일 `MsgExec{ type, spec }` + 단일 결과채널**에 `type` 디스크리미네이터. 제어/스트림(Data·Resize·Kill·Status) 공유라 메시지 안 가르고 type만 추가(separate MsgEditExec보다 깔끔)
- **두 타입(닫힌 집합)**:
  | type | 출력 스트림 | DB sink | 산출물 | PTY |
  |------|------------|---------|--------|-----|
  | `EXEC` | 라이브 fan-out | **ON**(스크롤백 영속) | 없음(exit code) | O |
  | `EDIT` | 라이브 fan-out | **OFF**(vi 화면=버림) | **파일 read-back 내용** | O |
- **공유 엔진 / type 무지**: `IInteractive`/PTY/MsgData·Resize·Status 100% 재사용, 런타임 핸들은 type 모름. type 소비자 = ① **worker bracket 핸들러**(seed 깔까/read-back 할까) ② **supervisor bind 배선**(sink on·off / 결과 라우팅). 엔진은 한 줄도 안 갈라짐
- **EDIT 흐름**: supervisor가 노드 content + 대상 worker(device_key 상속) 해석 → `MsgExec{EDIT, content 인라인}` → worker가 tmp에 content 기록 → `vi <tmp>` PTY 실행 → 종료(STOPPED) 시 tmp 재읽기 → `MsgEditResult{UID, content}`(인라인 REQ) → tmp 정리 → supervisor `nodes.content` UPDATE
- **저장판별 = read-back & diff**: vi는 `:wq`·`:q!` 모두 exit 0 → 종료코드로 저장여부 못 가림. **종료 후 tmp 무조건 재읽기→원본과 비교→바뀌면 UPDATE/같으면 no-op**(`:w` 안 했으면 파일 불변→자연히 "저장 안 함"). `:cq`(non-zero)=명시 취소는 선택적
- **구조체**: base `ProcessSpec`(UID/Cmd/Env/…) 공유 + **type별 addendum**(EDIT만 seedContent/tmp 정책). 3개 평행 구조체 금지(REF L44)
- **editor 선택 = worker 책임**($EDITOR/OS 기본). 1차 **Linux worker 한정**(Windows PTY=ConPTY 별도 이슈)
- **이름**: `EXEC`/`EDIT` 둘 다 **동사로 일관**. `EDITOR`(도구명—레이어 샘) 기각, `RETURN`(모호—exit code도 return) 기각. RETURN은 EDIT의 후보명이었을 뿐 동의어
- **의존성**: process 도메인 배선 선행 필수(현재 Step1만). frontend xterm.js 필요(apps/web 미착수) — e2e는 테스트 클라 conn으로 PTY 왕복 검증 가능
- **YAGNI**: "인터랙티브 세션이 아티팩트 반환" 거창한 프레임워크 금지. **EDIT 한 동작만**. 비인터랙티브 출력 캡처(CAPTURE류)는 실수요 나올 때 별 type으로
- **프로토콜 어휘 구현됨(2026-06-26)** `protocol/messages.go`: `ExecType`(EXEC|EDIT) + `ProcessSpec.Type`(빈값=EXEC, `Kind()` 정규화) + `MsgEditResult`(worker→sup REQ) + `EditResult{UID,Content}`. 와이어 = 공통 ProcessSpec + Type 구분. 저장판별=supervisor diff. ※구 `ProcessSpec.Seed []byte`/"인라인 content" 서술 폐기 → 아래 2026-07-01 절 참조(파일 선배치 + WorkerNodePath).

## supervisor 측 아키텍처 (2026-07-01) — manager/entry/bind/router 역할분리
> 구현(빌드·vet 통과): `cmd/supervisor/process/{manager.go,entry.go}`, `internal/supervisor/bind/relay.go`, `cmd/supervisor/router/process.go`(골격). 다이어그램=스크래치 `process-exec-structure.md`.

- **역할 경계** — 규칙: **"누구(who)를 찾고 전송" = router / "무엇을 상태로" = manager**.
  - **manager = 상태**: UID 발급·spec 내부작성·pool 영속·AgentInteractive 생성. 레지스트리·hub·세션 **모름**(worker conn은 받되 찾지 않음).
  - **router = wire**: authKey→worker conn 조회·MsgExec 전송·bind.Relay 기동·소켓 핸들러. **트리거는 router**(소켓 Handle→`router.Exec(owner,authKey,kind,node)`→manager 호출). 소켓이 manager를 직접 부르면 manager가 router화됨.
  - **bind.Relay = fan-out**: `Output()/Status()` 드레인→`hub.Publish`. `IInteractive` 위 재사용(AgentInteractive/로컬PTY 무관).
- **ProcessEntry**(memory 값, 키=UID) = `{Record *superdb.Process, Inter *AgentInteractive}`. **folder=Inter nil**(레코드-only, worker 무접촉). `HasProcess()`=Inter有 / `NodeID()` / `Spec()`=Record→MsgExec용 ProcessSpec 투영.
- **folder-open은 worker 연결 여부와 무관**(2026-07-02 발견 → 2026-07-03 수정 완료): folder(`openFolder`)는 worker 무접촉이 설계 의도. 구 `router.Exec`가 `node.Kind` 분기 전에 `r.workers.Get(authKey)` 존재를 선검증해 오프라인 device의 폴더 열람까지 막던 버그. **수정 = router 사전 게이트 제거**(`worker, _ := r.workers.Get(authKey)`, nil 허용) → worker 필요 검증을 manager로 단일 위임. manager 측 검증 위치 = `Exec`(folder 분기 뒤 `worker==nil`) + `ExecEdit` + `execScript`(기존). device 오프라인이어도 folder 열림 / SCRIPT·EDIT은 worker 없으면 에러.
- **spec은 파라미터 아님 → node로부터 내부작성**: manager가 UID 발급 후 node로 구성→CreateProcessParams로 persist. **바인딩=Record 하나(진실의 출처)**, 와이어 spec은 `entry.Spec()`으로 필요할 때 파생(별도 상주 금지—resize 등 drift 방지). 초기 Rows/Cols=0(**resize-on-attach**).
- **content 전달 = 파일 선배치**(인라인 폐기, 2026-07-01(2) 구현 완료): EXEC·EDIT 둘 다 worker에 파일 먼저 깔고 경로를 가리킴.
  - **경로 조립=supervisor / 지역화=worker** 책임 분리(핵심). 조립 헬퍼 = `cmd/supervisor/process.WorkerNodePath(node superdb.Node, proc superdb.Process)` → **`{WORKER_BASE}/<node.ID>/<proc.Uid>`**. superdb 구조체 의존이라 **protocol엔 못 둠**(이식성) → process pkg. 구 `protocol.WorkerNodePath(id)` 폐기.
  - 형식에 **`proc.Uid` 포함 = 실행 인스턴스별 격리**(동시/재실행 충돌 방지). process PK=uid(numeric id 없음)라 "process.Id"=Uid. node.ID = 어떤 스크립트인지.
  - 이 문자열을 **router 전송(DestPath)과 manager Cmd/Args가 함께** 사용(불일치=버그). manager는 Record 생성 전이라 발급된 uid로 `superdb.Process{Uid: uid}` 넘겨 조립.
  - **worker 치환(양 채널 동일 규칙)**: `resolveDest`(수신)·`exec`(실행) 둘 다 `{WORKER_BASE}→ProcessRoot` 치환 후 ProcessRoot 하위 traversal 검증 → 선배치=실행 경로 일치. ⚠️ resolveDest의 구 `baseDir/instanceKey` 루트 폐기(instanceKey별 격리 포기, `baseDir`·`instanceKey()`는 dead code화).
  - **EXEC**: `Cmd=path`, `Args=[]`, 전송 perm 0755. (content→실행 세부정책=TODO: 직접실행 vs `sh -c`)
  - **EDIT**: `Cmd={WORKER_EDITOR}`(치환 토큰 — vi 하드코딩 폐기), `Args=[path]`, perm 0644. worker `localize()`가 `{WORKER_EDITOR}`도 치환=`resolveEditor()`=`$VISUAL>$EDITOR>vi`. **TUI 에디터 강제 + PTY `$TERM` 세팅**은 실제 PTY 실행부(TODO)에서.
- **bind.Relay / 배치 드레인**: `NewRelay(uid, IInteractive, publish).Start()`→`pumpOutput`(MsgData)/`pumpStatus`(MsgStatus). `SyncData.ShiftAll()`+`IInteractive.OutputAll()`(연결바이트) 추가로 **락경합·프레임 수 감소**. publish=클로저(`hub.Publish("PROC:"+uid,…)`)로 Hub 제네릭 결합 회피.
- **EDIT 계약 2개(프로토콜 밖, 배선에서 보장)**: ① 전송 `DestPath`==MsgExec `Args` 경로 일치 ② `MsgEditResult`(UID→NodeID→`UpdateNodeContent`) **처리 전 엔트리 teardown 금지**(EditResult엔 nodeId 없어 UID 매핑 필요).

## worker 실행부 (2026-07-13 본체 구현 완료)
> supervisor 측 대칭. supervisor=원격 래퍼(AgentInteractive)·pool 영속 / worker=로컬 PTY(`*pty.Interactive`)·**휘발**(DB 없음).

**선행작업 완료 (2026-07-03)**
- **pty primitive 확장**: `ExecInteractive(ctx, command, env []string, args...)` — **nil=os.Environ() 상속 / 목록=완전 대체**(cmd.Env 세팅). `Interactive.Pid() int`(미기동/종료 -1). `Pid()`는 **`IInteractive` 밖**(worker 로컬 개념, 구체타입 전용).
- **env 조립 정책**: 완전 대체이므로 호출자가 베이스 조립 책임 — `os.Environ()` + `spec.Env`(치환됨) + **`TERM=xterm-256color` 강제**(TUI/vi 필수, REF 구 "$TERM 세팅" 요구 충족). PATH/HOME/LANG은 os.Environ() 상속으로 자동.
- **EDITOR 출처**: `resolveEditor()` = `constants.GetEnv().Editor > "vi"`. env 필드 = `EnvVars.Editor`(`iniName:"EDITOR"`, 선택값).

**본체 구현 완료 (2026-07-13, build/vet 통과 + e2e 스모크 확인)**: 관리 객체 = **supervisor `ProcessManager` 패턴 미러링**. REF-process:46 "양쪽 manager 모두 UID로 키잉" 준수.
- **관리 객체 = `workerRouter.procs manager.KeyValManager[string, *procEntry]`**(`saves` 옆 필드, `cmd/worker/router/process.go`). `procEntry{uid, kind protocol.ExecType, editPath string, inter *pty.Interactive}`.
- **exec()**: localize(기존) → `env := os.Environ()+body.Env+"TERM=xterm-256color"` → `pty.ExecInteractive(ctx,cmd,env,dir,args...)` → `procs.Append(uid,entry)`(중복 시 `inter.Kill()`+거부) → `go pump(uid,entry)` → `ExecResponse{Accept}`.
- **pump = 서브 고루틴 2개**(`bind.Relay`의 `pumpOutput`/`pumpStatus` 쌍을 그대로 미러링): `go pumpOutput`(별도 고루틴) + `pumpStatus`(호출된 고루틴 자신이 실행, `pump()` 자체가 블로킹 리턴 없이 여기서 상주).
  - `pumpOutput`: `inter.OutputAll()` 배치 드레인 → `conn.Emit(MsgData)`. 채널 close(err) 시 종료.
  - `pumpStatus`: `inter.Status()` 드레인 → `conn.Emit(MsgStatus)`(PROCESS 전이 시만 `Pid()` 얹음). **`IsCompleted()` 감지 시 `teardown()` 호출 후 자기 자신 종료** — 종료 감지·teardown이 **단일 지점**(pumpStatus)으로 수렴, pumpOutput은 채널 close로 뒤따라 자연 종료.
  - `teardown(uid,entry)`: EDIT면 `os.ReadFile(editPath)` → `conn.Call(MsgEditResult)`(실패해도 계속 진행) → `procs.Remove(uid)`. 저장판별(diff)은 여기서 안 함 — supervisor `editResult` 책임(REF 상단 EDIT 계약 그대로).
- **input/resize/kill**: `procs.Get(uid).inter.Write/Layout/Kill`. **kill은 Remove 안 함** — kill→프로세스 종료→`pumpStatus`가 `IsCompleted()`로 감지→`teardown`이 **단일 teardown 경로**(이중정리 방지). `resize`/`kill`은 없는 uid면 에러 RES, `input`은 EVENT라 짝 응답 없어 조용히 무시.
- **핸들러 등록**(`workerRouter.go`): `On(MsgData)=input`, `Handle(MsgResize)=resize`, `Handle(MsgKill)=kill` 추가(`exec`는 기존 등록 유지).

**Cwd 배선 (2026-07-13, 본체 구현과 함께)**: `pty.ExecInteractive`에 `dir string` 파라미터 추가(`env` 다음, variadic `args` 앞) — **빈 문자열=worker 프로세스 cwd 상속**(`cmd.Dir` 미설정, env의 nil-상속 정책과 결 맞춤), 비면 `cmd.Dir=dir`. `body.Cwd`(localize 완료)를 그대로 전달. 콜사이트가 worker `exec()` 하나뿐이라 파급 없음. **traversal 검증은 안 함** — 파일 전송 쪽 `resolveDest`와 달리 지금 `Cmd`/`Args`도 검증 없이 localize만 거치는 기존 정책과 대칭 유지(비대칭 방지).

**⚠️ 발견 — PROC 토픽 무구독 (미해결, 다음 단계 선행조건)**: worker `pumpOutput`/`pumpStatus` → sup `On(MsgData)/On(MsgStatus)` → `entry.Inter.Push*` → `bind.Relay` 드레인까지는 정상 도달하지만, `subscribeHub.Publish("PROC:"+uid, ...)`가 **구독자 0명이면 에러·로그 없이 조용히 폐기**(`internal/subscribe` Publish: `len(clients)==0 → return nil`). 현재 프론트는 `NODE:0` 고정구독뿐이라 `PROC:<uid>` 동적 구독 경로가 없어 출력이 여기서 끊김. e2e 확인은 worker 로그(Emit 직전)로 했음 — supervisor~프론트 구간은 "frontend 트리거·제어"(동적 구독 포함, 아래 순번 2) 배선 전까진 검증 불가. → `REF-realtime.md`(동적 구독 설계)와 연결 지점.

## 종료 / 재접속 모델 (2026-07-01 확정) — status 단일 깔때기
> MEMORY의 "끊김→Done(502)" 폐기. 이 절이 최신.

- **`status` = 모든 상태전이의 유일 수렴점.** 어떤 종료든 process 라우터 `status`로. 두 진입: ① worker `On(MsgStatus)` ② worker 끊김 시 supervisor **합성 호출**. → `status(ev Frame)` 안쪽에 **`applyStatus(uid,status,pid,exit)` 코어** 분리(양쪽 재사용).
- **종료 시나리오**: 자연종료 / **터미널입력**(Ctrl+C=입력 0x03, kill 아님—프로그램이 결정, 입력경로) / **명시 kill**(종료버튼=MsgKill, 제어경로) / worker끊김(→PENDING) / EDIT취소(:cq→Failed→content UPDATE 안 함) / worker외부kill(OOM). **Ctrl+C≠kill 반드시 구분**.
- **frontend 끊김 ≠ 종료**: 명시적 종료 시그널 전까진 계속 실행(브라우저 종료 무시). **같은 세션 재접속=화면 그대로 복원** → 필요: process별 **ring buffer(SNAPSHOT, 이제 필수)** + **세션→보던 uid 원장**(conn 단위 Hub 구독과 별개) + 재접속 시 SNAPSHOT→live.
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
원래 `applyStatus`는 진입 시 `entry, ok := processManager.Get(uid); if !ok { return }`로 memory entry 없으면 즉시 리턴했다. 그런데 재접속의 "supervisor만 앎(worker 소실)" 분기는 `applyStatus(uid, CommandFailed, ...)`를 호출하는 시점에 그 uid가 **이미 끊김 처리(`CommandPending` 분기)에서 `processManager.Remove`된 뒤**라 entry가 없다 — 원 설계 그대로면 DB가 Failed로 안 닫힘.
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
