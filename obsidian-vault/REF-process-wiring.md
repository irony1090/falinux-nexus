# REF — process 배선 (supervisor 아키텍처 + worker 실행부)

> 계약/설계 원칙은 `REF-process.md`. 종료/재접속·세션 복원은 `REF-process-reconnect.md`. 작업 이력은 `history/process-wiring.md`.

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

**⚠️ 발견 — PROC 토픽 무구독 (미해결, 다음 단계 선행조건)**: worker `pumpOutput`/`pumpStatus` → sup `On(MsgData)/On(MsgStatus)` → `entry.Inter.Push*` → `bind.Relay` 드레인까지는 정상 도달하지만, `subscribeHub.Publish("PROC:"+uid, ...)`가 **구독자 0명이면 에러·로그 없이 조용히 폐기**(`internal/subscribe` Publish: `len(clients)==0 → return nil`). 현재 프론트는 `NODE:0` 고정구독뿐이라 `PROC:<uid>` 동적 구독 경로가 없어 출력이 여기서 끊김. e2e 확인은 worker 로그(Emit 직전)로 했음 — supervisor~프론트 구간은 "frontend 트리거·제어"(동적 구독 포함, 아래 순번 2) 배선 전까진 검증 불가. → `REF-realtime.md`(동적 구독 설계)와 `REF-process-reconnect.md`(세션→uid 원장의 CREATE 훅 지점)와 연결 지점.
