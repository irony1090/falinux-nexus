# REF — process 실행 모듈 (execute / pty / agent comm) — 계약·설계 원칙

> worker 프로세스 실행/모니터링/종료(PTY) 설계와 계승 자산. **이 파일 = 상위 계약/원칙(자주 안 바뀜)**.
> 실제 배선(누가 무엇을 호출)은 `REF-process-wiring.md`, 종료/재접속·세션 복원은 `REF-process-reconnect.md`.
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
- **재연결 reconciliation**: 끊김→`Done(502)` 비관적 정리 / 재연결→worker가 live 스냅샷 보내 재동기화. **2026-07-01 개정, 2026-07-14 구현+e2e 완료**(구 `Done(502)` 폐기 → PENDING 모델) → `REF-process-reconnect.md`

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
- **프로토콜 어휘 구현됨(2026-06-26)** `protocol/messages.go`: `ExecType`(EXEC|EDIT) + `ProcessSpec.Type`(빈값=EXEC, `Kind()` 정규화) + `MsgEditResult`(worker→sup REQ) + `EditResult{UID,Content}`. 와이어 = 공통 ProcessSpec + Type 구분. 저장판별=supervisor diff. ※구 `ProcessSpec.Seed []byte`/"인라인 content" 서술 폐기 → `REF-process-wiring.md` "경로 조립" 참조.

## 하위 문서
- **배선(구현)**: supervisor 아키텍처(manager/entry/bind/router 역할분리) + worker 실행부(procs/exec/pump/teardown) → `REF-process-wiring.md`
- **재접속·세션 복원**: 종료/재접속 모델, worker 끊김→PENDING→재바인딩, 세션→uid 원장(`process_subscribers`) → `REF-process-reconnect.md`
