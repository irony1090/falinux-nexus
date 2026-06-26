# MEMORY

## 경로 참조 (중요)
- **현재 프로젝트 (Nexus)**: `/home/jh-bae/irony/nexus`
  - Go 코어: `/home/jh-bae/irony/nexus/apps/core`
- **이전 프로젝트 (PortBridge)**: `/home/jh-bae/irony/test-jig`
  - Go 코드: `/home/jh-bae/irony/test-jig/apps/test-jig-api`
  - 프론트: `/home/jh-bae/irony/test-jig/apps/test-jig-web`
  - vault 원본: `/home/jh-bae/irony/test-jig/obsidian-vault`

> 사용자가 **"test-jig" / "예전 프로젝트" / "구 프로젝트" / "이전 프로젝트" / "PortBridge"** 로
> 칭하면 위 `/home/jh-bae/irony/test-jig` 경로를 가리키는 것이다. 해당 코드를 참조할 것.

## 프로젝트 개요
- **프로젝트명**: Nexus
- **설명**: 2-tier 에이전트 관리 플랫폼
  - **Worker Agent**: 자기 자신의 process를 관리하는 에이전트 (실행/모니터링/종료)
  - **Supervisor(총괄) Agent**: 다수의 worker agent를 관리하는 상위 에이전트 (연결/상태/명령 분배)
- **출발점**: 이전 프로젝트(PortBridge / test-jig)의 agent↔server 통신 아키텍처를 계승.
  방향성이 "테스트 장비 데스크탑 UI"에서 "에이전트 계층 관리"로 전환되어 신규 프로젝트로 분리함.

## 이전 프로젝트에서 계승한 핵심 자산 (검증됨)

> Nexus의 worker/supervisor 구조는 PortBridge의 agent↔workspace-server 통신을 일반화한 것.
> 아래는 그대로 재사용 가능한 설계/구현 지식. (PortBridge 전용 UI·아이콘·그룹 로직은 폐기)

### Agent↔Server 통신 프로토콜
- `agentId` = 에이전트 식별자 (PortBridge에선 MAC 주소, HTTP 헤더로 전달)
- `SessionId` = 사용자/세션 식별자
- key 형식: `agentId:processId` (또는 `agentId:userId:processId`)
- **Agent → Server**: `MsgStatus`(RUNNING/STOPPED + ExitCode), `MsgData`(output)
- **Server → Agent**: `MsgExec`, `MsgData`(input), `MsgResize`, `MsgKill`, `MsgUploadInit/Chunk`
- 파일 전송 흐름: `UploadInit → Ready → Chunk → Status → Result`

### Process 실행 추상화 (worker agent 핵심)
- **`IInteractive` 인터페이스**: `Output()`, `Write()`, `Status()`, `Kill()`, `Layout() syscall.Errno`, `ExitCode() int`
  - `*Interactive` (로컬 PTY 구체 타입), `*AgentInteractive` (원격 agent 래퍼) 모두 구현
- **`AgentInteractive`**: agent WebSocket을 통한 원격 process 래퍼
  - `onWrite`/`onLayout`/`onKill` 콜백으로 agent에 명령 전달
  - `PushOutput()`/`PushStatus()`/`Done(exitCode)` 외부 주입 메서드
  - `sync.Once`로 `Done()` 중복 호출 방지
  - 생성 시 초기 status push 없음 → RUNNING 수신 시 첫 push

### Process 실행 흐름 (동기화 패턴)
```
CreateProcess → process 생성 + AgentInteractive 생성 + (필요시 파일 전송)
  → ExecPayload 전송 → agent 실행
  → MsgStatus(RUNNING) → PushStatus → Status() 블로킹 해제 → RUNNING 반영
  → bindManager에 등록 → goroutine 대기
  → MsgData → PushOutput → UI/상위 스트리밍
  → MsgStatus(STOPPED) → Done(exitCode) → goroutine 종료
```

### Supervisor가 agent 관리하는 패턴
- agent 연결/해제 시 콜백(`OnCreated`/`OnRemoved`)으로 상태 갱신 + 브로드캐스트
- agent 연결 끊김 시 해당 agent 소속 모든 interactive를 `Done(502)`로 정리 (defer cleanup)
- agent 상태를 `agentId(string)` 기반 싱글턴으로 관리 → DB 조회 없이 매핑

### 구독/토픽 모델 (다수 agent 상태 감시)
- topic 기반 구독/해제로 selective push (전체 broadcast 회피)
- 구독 등록은 REST, push는 WS 단방향 권장 (PLAN 참고)
- 구독 시 즉시 `SNAPSHOT`, 이후 변경분만 `UPDATE`
- WS 연결 해제 시 `UnsubscribeAll` 필수 (재연결 중복/dead client publish 방지)
- 동시성: Lock 구간 분리 (구독자 수집 → Unlock 후 json.Marshal/WriteMessage)

### 재사용 구독 허브 (`internal/subscribe/hub.go`) — 구현 완료
- ※ 명명 최신화(2026-06-25): 타입 **`Hub[C]`**(구 `Manager`), 생성자 **`New`**(구 `NewManager`), 키별 집합 = 비공개 **`topic[C]`**(구 `subscriber`), 파일 **`hub.go`**(구 manager/subscriber.go)
- PortBridge subscribeManager를 도메인 제거하고 일반화한 코어
- `Hub[C comparable]`: 키는 **문자열만**, 클라이언트는 **제네릭** (`*Client`/인터페이스 모두 가능)
- 키 생성·도메인 타입은 매니저가 모름 → 프로젝트가 `Key() string`으로 생성해 넘김
- `New(marshal, send)`: 직렬화/전송 전략 1회 주입 → 호출부는 `Publish(key, payload)` 한 줄
- `Publish`: marshal 1번 → 락 밖에서 N번 send, send 실패는 `errors.Join`으로 모아 반환(브로드캐스트 중단 안 함)
- `Subscribers(key) []C`: 특수 전송용 스냅샷 탈출구
- 설계 원칙: **매니저 책임 = "문자열 키 → 구독 클라이언트 집합" 장부 + 동시성**, 그 외(키/직렬화/전송)는 프로젝트 몫

### 요청/응답 상관관계 프레임 (`internal/call` + `protocol` + `transport`) — 구현 완료
- 구 PortBridge `agentRouter.go`의 통증 해결: 요청·응답이 두 파일 switch에 흩어지고 짝 ID가 없어 추적 곤란 → 도메인 매니저 조회로 역추적하던 문제
- **`call.Correlator[R]`**: "요청 ID → 응답 기다리는 채널" 장부 + 동시성만. send 주입, 응답 타입 R 자유. `Call(ctx,payload)`(블록)/`Resolve(id,val,err)`/`Close(err)`(전부 깨움). subscribe.Hub와 동일 철학(코어는 도메인/전송 모름)
- **`protocol.Frame{Kind,ID,Type,Err,Data}`**: Kind=**REQ/RES** 로 요청/응답 프로토콜 레벨 구분. Data=json.RawMessage(자유 payload). REQ가 ID 채번 → RES가 반사
- **`transport.Conn`**: `Call`(REQ→RES 블록)/`Handle`(REQ 핸들러)/`Serve`(수신 루프). supervisor·worker 양쪽 대칭 API. `MessageRW` 인터페이스에만 의존(gorilla `*websocket.Conn`이 그대로 만족)
- 연결 끊김 시 `corr.Close(err)`로 매달린 Call 일괄 실패 = 앞서 정한 Done(502) reconciliation과 같은 정신
- **예제**: 각 main에 supervisor(WS서버) ↔ worker(WS클라) 1요청→6응답. 동작 확인됨
- **확장 완료**: EVENT 평면(아래) + `transport.Pipe()` 재구현됨(2026-06-25). ※ 이전 메모의 "git 히스토리에 있음"은 **오류** — 작업 트리에서만 만들었다 커밋 전 삭제라 복구 불가였고, 새로 작성함

### EVENT 평면 (단방향 1:N, `protocol`+`transport`) — 구현 완료 (2026-06-25)
- REQ/RES(제어, 1:1, Call 블록)와 대칭인 **데이터 평면**: 짝 없는 단방향(`Emit` 즉시리턴 / `On` 핸들러). 용도=출력 스트리밍(MsgData), STATUS 푸시. **Correlator 안 거침**
- `protocol`: `EVENT` Kind + `NewEvent(t,data)`(ID 미채번). Frame 구조 불변(Kind 분기만)
- `transport.Conn`: `On`(핸들러 등록)/`Emit`(=`write`만)/`dispatch()`(전용 goroutine). streamID는 **payload**에 (transport 도메인 무지 유지, Frame.ID는 REQ 전용)
- **순서 보존 체인**: 송신 `wmu` 직렬화 → Serve 순차 push → **단일 dispatch goroutine** 순차 호출 = stdout 순서 end-to-end 보존. (REQ는 `go dispatchReq`라 순서 무관, EVENT는 순차 필수)
- **결정 1(안전종료)**: `events` 채널 **절대 close 안 함** → `Close`에서 `close(done)`만. Serve push는 `select{events<-f; <-done}` 가드 → 끊김 순간 send-to-closed 패닉 차단. transfer의 `done` 패턴 재사용
- **결정 2(수명계약)**: dispatch는 `New`서 1회 시작, `Close`(owner=라우터)서 종료. 새 부담 아님(rw.Close 때문에 어차피 Close 호출). 안 부르면 goroutine 누수
- backpressure: `events` 버퍼(256) 차면 Serve 블록(=느린 구독자 신호). On(DATA) 핸들러는 "버퍼 push/forward"라 빨라야 정상
- `transport.Pipe()`: 인메모리 양방향 MessageRW(한쪽 Close→공유 closed로 양쪽 깨움). 테스트: Emit×1000 순서+개수 / REQ·EVENT 혼류 무간섭, 둘 다 `-race` 통과

### worker 재연결 + 신호 설계 교훈 (2026-06-25)
- **재연결 정책은 router 바깥(main for문)**: router 필드(conn/saves/subKey)=연결 귀속 상태 → 새 router로 자동 폐기. 장수 상태(store/향후 process매니저)는 main서 만들어 주입. (router=연결 하나, transport.Conn이 스스로 재연결 안 하는 철학과 동일)
- **★신호 교훈(재사용)**: "끝날 수도 안 끝날 수도 있는 신호 2개(Ready+Done)" 대신 **"반드시 한 번 끝나는 신호 1개에 결과를 실어라"**(`Result{Reached,Err}` 단일 채널). 그러면 "에러 때 Ready도 닫나" 문제가 *풀리는 게 아니라 사라짐*. `sync.Once finish`로 다송신자(Serve/register) 1회 송신+conn 1회 Close까지 동시 해결
- backoff: `internal/util/retry.util.go` `Backoff{cur,base,max}` `Next()`(2배·상한)/`Reset()`. `reached`(register 성공) 기준 Reset → 정상가동 후 끊김=짧게, 그 전 실패=증가. **jitter 미적용**(worker 다수·supervisor 동시재접속 부하 시 full jitter 추가)

## 파일 전송 모듈 (구현 완료, e2e 미검증 — 2026-06-24)

> supervisor→worker 단일 파일 전송. FileInit→FileChunk×N→FileResult(+FileAbort). 송신/수신 대칭 구조. 상세 HISTORY 2026-06-24.

### 무결성 해시 (`transfer.Hash()`)
- Hash = **원본 전체 sha256(hex 64자)**. 출력은 입력 크기 무관 고정(32B) → 와이어 부풀음 無, 비용은 선읽기 I/O뿐
- 둘 다 **스트리밍 fd와 별개의 새 fd(`os.Open(name)`)로 풀읽기** → 커서·카운터·타이머 불간섭
- `ReadFile.Hash()`: 원본 불변 → `sync.Once` 캐시, 락 불필요 / `SaveFile.Hash()`: 수신중 계속 변함 → **캐시 금지**, "완료 후 1회", `s.mu` 잡음. WriteAt 비순차/재전송/이어받기는 running hash로 못 다룸 → 결과 풀읽기가 정답
- `SaveFile.Sync()` = rename 직전 fsync

### 송신(SendFile) / 수신(핸들러) 패턴
- 송신 래퍼 `fileSend{authKey,reader,cancel}` / 수신 래퍼 `fileRecv{save,hash,finalPath}` (FileResult 검증에 init의 hash·경로 필요한데 ResultReq엔 transferId만 실림)
- 재시도: `maxSendAttempts=3` + backoff. 거부(errRejected)=즉시중단 / 그 외=재시도 / 초과=error
- 이어받기: 받는 쪽이 `.part` stat으로 resumeOffset 결정 → 송신자 `SeekTo` 후 재개. 같은 transferId·destPath라 재부팅 후도 결정적
- 수신 완성: **Completed(크기)→Hash일치→Sync→Close→Rename(.part→최종)**. 크기미달=part보존+resumable / hash불일치=part삭제+not resumable
- traversal 차단: `Clean("/"+destPath)`로 선행 `..` 무력화 → `filepath.Rel(root,final)`로 루트 하위 재확인
- 끊김 정리: conn 소속만 `FindAll(authKey==key)`로 reader Close (공유변수 X)

### abort (cancel ctx 전파)
- `SendFile`이 `WithCancel(Background)` ctx 생성→`fileSend.cancel` 저장, **per-call timeout을 이 ctx에서 파생**(Background 아님)해야 cancel이 진행 중 Call로 전파됨. 루프에 `ctx.Err()` 체크로 재시도 차단
- `AbortFile`: cancel() + worker에 MsgFileAbort(best-effort). reader 정리는 SendFile defer Close가 전담(abort/정상/끊김 한 경로)

### context.cancel 교훈
- `cancel()`=ctx 타이머·parent등록 해제(연결과 무관). 안 불러도 Background자식은 timeout시 자동정리되나 **그때까지 쌓임** → **루프에선 defer 금지, Call 직후 즉시 cancel**(defer는 함수 끝까지 안 풀려 타이머 수천 개 누적)

## 확정 결정 (불변 — 장기 유지)
- **계층: 2단계 고정** (supervisor → worker, 중간 노드 없음. 식별자/라우팅/구독 1홉 평면)
- **worker는 메모리 전용 관리**(2026-06-26 확정) — supervisor DB에 worker/agent 테이블 **없음**. 스캐폴딩의 예시 `agents` 테이블은 미사용이라 폐기. worker 신원은 런타임 메모리 레지스트리(`InstanceKey`)로만 관리. → node의 device 결속은 FK 아닌 `device_key TEXT`(위 node 설계 참조)

### process 모듈 설계 (2026-06-25 확정, Step 1 구현 시작)
- **★발견: `execute`는 스텁 아님** — PortBridge 이식 완료/미배선. 보유: `IInteractive`(Output/Write/Status/Kill/Layout/ExitCode), `*Interactive`(로컬 PTY — **직접 만든 `/dev/ptmx` ioctl, creack/pty 불필요, Linux 전용**), `*AgentInteractive`(원격 래퍼 onWrite/onLayout/onKill+Push*/Done), 기타 `ExecCommand`/`Fifo`/`TmuxSession`, `syncProcess.SyncData[T]`. 빠진 것=신원(UID/PID)·매니저·배선
- **★패키지 분리(2026-06-25)**: `execute`가 통째 Linux 전용이면 supervisor(AgentInteractive만 필요)까지 비-Linux 빌드 불가 → **OS syscall 의존 파일을 `execute/pty`(package pty)로 분리**. `execInteractive.go`(PTY)·`execFIFO.go`(Mkfifo) 이동. `execute`(이식 가능)=IInteractive/AgentInteractive/CommandStatus/ExecCommand/Tmux 유지. pty는 execute를 import(CommandStatus 등), `var _ execute.IInteractive = (*Interactive)(nil)` 어서션. 공유 변수 `rn`/`nr`는 분리 후 양쪽 각자 보유
- **상태 어휘**: `execute.CommandStatus`(Pending=0/Process=1/Completed=2/Failed=3)를 **와이어에서 그대로 사용**(`StatusEvent.Status`). execute가 이식 가능해져 `protocol`이 import 가능 → **ProcStatus 미러는 폐기**(단일 소스). transport→protocol→execute 그래프에 execute 이식부가 딸려오나 전부 portable이라 무방
- **구조체 공유 = 데이터 서술자만**: `ProcessSpec{UID,Cmd,Args,Cwd,Env,Rows,Cols}`는 **`protocol`에 둠**(execute 아님 — 위 이식성 이유, 순수 데이터라 protocol이 적격, execute가 필요시 protocol을 import=사이클 없음). 런타임 핸들은 각자 → 행위는 `IInteractive`로 통일. **통째 동일 구조체 박지 말 것**(런타임 핸들 다름)
- **Step 1 완료(2026-06-25)**: `protocol/messages.go`에 MsgExec/Data/Resize/Kill/Status + ProcessSpec/ExecResponse/DataEvent/ResizeRequest/KillRequest/ProcStatus/StatusEvent 추가. 제어=REQ/RES, 스트림=EVENT(streamID=UID). MsgData는 양방향(입력 sup→worker/출력 worker→sup). StatusEvent.PID=RUNNING 시, ExitCode=Done 시(omitempty 금지)
- **UID vs PID 분리**: UID=시스템 전체 정체/라우팅 키(**initiator 발급**=supervisor top-down 또는 worker 자기 프론트, RandomKey 전역유일, **모든 메시지에 실림**) / PID=worker 로컬 OS 핸들(worker 발급, RUNNING 시 status에 얹어 위로 보고만, 와이어 라우팅 안 함). **양쪽 manager 모두 UID로 키잉**
- **평면 매핑**: 제어(Exec/Kill/Resize)=REQ/RES, 스트림(Data/Status)=EVENT(streamID=UID). 입력 키스트로크=EVENT
- **fan-out 허브 = `IInteractive` 위에서 재사용**: 출력 → ring buffer(SNAPSHOT용) + subscribe.Hub(라이브) + DB sink(영속). 허브는 sink 모름(bind 계층이 배선). **공유 위치 + IInteractive 파라미터화**(supervisor 전용 금지) → worker 자기 프론트도 나중에 mount만으로 붙음(YAGNI: 지금은 안 막기만)
- **frontend 다리**: 브라우저도 `transport.Conn`(gorilla=MessageRW). 명령=`Call`/`Handle`(REQ/RES), 입력=`Emit`(EVENT), 출력 fan-out=`subscribe.Hub.Publish`(send 전략=`conn.Emit`). 구독=**WS Call로 통일**(REST 안 섞음). SNAPSHOT(스크롤백)+UPDATE(라이브)
- **느린 소비자 격리(frontend 한정)**: worker→sup `On(Data)` 핸들러가 Publish 동기호출 시 느린 브라우저 1개가 디스패처까지 막음 → **per-client 바운드 큐+writer goroutine**, Emit은 넣기만. 큐 차면 **그 클라만 끊고 재구독 유도**(터미널은 중간유실=화면깨짐이라 끊고 SNAPSHOT 재전송이 깔끔). 저빈도 list 토픽은 동기 OK
- **인코딩**: 일단 Frame(JSON) 통일. 출력 바이트 base64 ~33% 부풀음 문제되면 **제어=Frame/출력=raw WS 바이너리** 분리(추후 여지만 남김)
- **agent 식별자: 사전 지정 고유키(메인키)** + supervisor가 접속 시 부여하는 **서브키**. 저장/조회는 `메인키#서브키`(`InstanceKey()`). MAC 폐기
- **process 영속성**: worker 휘발(메모리) / supervisor 영속(PG, 재시작 복구)
- **재연결 reconciliation**: 끊김→`Done(502)` 비관적 정리 / 재연결→worker가 live 스냅샷 보내 재동기화. 미구현(supervisor 본격 작업 시)

## Node/Label 모듈 설계 (frontend 카탈로그, 2026-06-26 확정 — 미구현)

> frontend에서 worker의 "앱/셸"을 나열·실행하는 **supervisor PG 영속 카탈로그**. process 모듈에 "무엇을 실행할지"를 먹이는 계층(**카탈로그=무엇 / process 모듈=어떻게**). DB+CRUD는 process 모듈과 독립 선행 가능, 실제 "실행"은 MsgExec 배선 필요.
> ★용어 합의 과정: 사용자가 "쉘/앱"이라 부른 것의 실체 = **디바이스(worker) 위 트리에 정리된 실행 정의**. 단일 엔티티가 아니라 folder+script 두 종류가 한 트리에 묶임.

### 통합 노드 트리 (`nodes` 단일 테이블)
- **두 종류(folder/script)를 한 트리에**: 폴더·스크립트를 frontend 한 목록에 섞어 나열+자유 재정렬해야 함 → 테이블 2개로 쪼개면 UNION+정렬좌표 분산으로 불편 → **단일 `nodes` + `kind` 구분**(파일시스템 inode 패턴). 한 폴더 자식 = `WHERE parent_id=$1 ORDER BY position_x,position_y` 한 줄
- **folder**=컨테이너(트리 노드, **worker 귀속 + 접속상태 창** 역할) / **script**=리프(셸 본문 텍스트 보관, 실행 시 worker로 인라인 전송 후 PTY)
- **worker 귀속 = 폴더 단위**(`device_key TEXT`, **FK 아님**). 스크립트 실행 대상 worker = 자기 폴더에서 위로 올라가며 **가장 가까운 device_key 상속**. device_key 없는(NULL) 폴더 = 순수 정리용
  - ★**worker는 메모리 전용**(DB `agents` 테이블 폐기, 2026-06-26) → `device_key`는 worker의 `InstanceKey`(`메인키#서브키`) **문자열**일 뿐 FK 무결성 없음. 폴더↔worker 결속 검증 = "그 키가 **런타임 메모리 레지스트리에 살아있나**"로 함(끊긴 worker = 키가 메모리에 없음). DB는 키 문자열만 보관
- **트리** = `parent_id` 자기참조(인접리스트), 루트=NULL
- **position_x/y = `NUMERIC(5,2)` 실수 0~100**(정수면 한 부모에 101칸뿐 + 두 노드 사이 끼워넣기 불가라 실수 채택). PC=캔버스 **% 절대배치** / 그외=`ORDER BY x,y` 리스트·그리드. **부모별 로컬좌표**, x·y 동률은 name/id 타이브레이크
- **전송=인라인**: 스크립트가 작은 텍스트라 청크 전송 모듈 불필요(실행 메시지에 `content` 실어 보냄). 큰 바이너리 생기면 그때 transfer 모듈 붙임(YAGNI). ※"전송 후 실행"은 실행 동작에 흡수 — 맨 처음 논의한 파일전송 모듈은 헛다리가 아니라 **순서가 빨랐던 것**
- **CHECK 정합성**: `content`=script 전용(folder면 NULL), `device_id`=folder 전용(script면 NULL), x/y 0~100

```
nodes( id, owner_user_id→users, parent_id→nodes NULL,
       kind 'folder'|'script', name,
       position_x NUMERIC(5,2), position_y NUMERIC(5,2),  -- 0~100 부모별 로컬
       device_key TEXT NULL,     -- folder 전용, worker InstanceKey 문자열(FK 아님)
       content    TEXT NULL,     -- script 전용
       created_at, updated_at,
       CHECK(kind='script' OR content    IS NULL),
       CHECK(kind='folder' OR device_key IS NULL),
       CHECK(position_x BETWEEN 0 AND 100 AND position_y BETWEEN 0 AND 100) )
```
- `users`=구현 완료(2026-06-26, 구 PortBridge 이식) / worker=메모리 전용(DB 테이블 없음)

### Label 모듈 (추후) — 직교하는 두 번째 트리
- **node 트리 = 구조/배치(부모 1개) vs label 트리 = 횡단 분류 + 공유(M:N)**. 서로 직교
- `labels(id, owner_user_id, parent_id→labels NULL, name)` 중첩 카테고리(**Gmail 라벨 패턴**) + `node_labels(node_id, label_id)` M:N
- **공유를 라벨이 떠안음** → 노드 단위 sharing 테이블 **지금 안 만듦**. "노드를 라벨에 담고 → 라벨 공유 → 그 노드 전부 접근". 라벨=순수 가산 join이라 나중에 무통증 추가(YAGNI). 소유는 노드별 `owner_user_id`, 공유는 라벨로

## 이번 세션 재사용 자산/교훈 (2026-06-19)

### transport.Conn 수명 상태
- `ConnState`(StateActive=0/StateClosed) + `State()` + `String()`. `atomic.Int32` 필드(zero=Active). Serve 수신루프 종료·Close에서 StateClosed 전이(둘 다 idempotent)
- 도메인 무지·대칭 유지 — "연결이 살아있나"는 conn 속성, **신원(메인키#서브키)은 라우터 몫**(conn 아님)

### `internal/util/string.util.go` — RandomKey
- `RandomKey(length int, prefix, suffix string) (string, error)`: base62(`0-9A-Za-z`), length=prefix/suffix **포함** 전체 길이(결과 정확히 length), `crypto/rand`
- 62는 256 약수 아님 → **거부 샘플링**(>=248 바이트 버림)으로 modulo 편향 제거 (hex 16이면 `b&0x0f`로 충분했음). 타입: `maxByte byte`, 인덱스 `b%byte(len(alphabet))`
- 용도: 서브키 발급, transferId. `*.util.go`가 util 패키지 컨벤션

### `internal/manager/KeyValManager` (PortBridge 이식)
- 제네릭 `KeyValManager[K comparable, V any]`: 키→**단일 값** 장부 + `OnCreated`/`OnRemoved` 수명 콜백. `subscribe.Hub`(키→**집합**)와 역할 다름
- **함정 4개**: ① `FindAll` predicate는 **락 안**에서 실행 → predicate서 매니저 재호출 금지(데드락) ② 콜백은 락 바깥(LIFO defer)이라 TOCTOU 가능 ③ `OnCreated/OnRemoved`는 동시사용 전 1회 세팅(동기화 없음) ④ `Append`는 키 존재 시 false → **반환값 꼭 확인**(무시하면 중복/경합이 조용히 묻힘)

### Go 교훈: 포인터도 "값으로" 복사된다
- 핸들러가 채운 값을 호출자가 보려면 **호출자가 실제 구조체를 들고 그 주소(`&val`)를 넘겨야** 함. `var p *T`(nil) 넘기고 핸들러가 `p`를 새 구조체로 재할당해도 **호출자 변수엔 안 보임**(포인터 변수 자체가 복사됐으므로) → supervisorRouter cleanup이 안 돌던 원인
- 응용: 연결별 cleanup은 공유 변수 대신 **스코프의 `conn` 값으로** `FindAll(c==conn)` 하면 레이스·오삭제 둘 다 없앰

### http 서버 구성
- 패키지 전역 `http.HandleFunc`/`ListenAndServe(_,nil)`는 전역 `DefaultServeMux` 공유 → 라우팅 섞임·서버별 타임아웃/shutdown 불가. **명시적 `http.NewServeMux()` + `&http.Server{}`** 권장. `NewSupervisorRouter`가 mux 내부 생성 후 `*http.Server` 반환하는 형태로 정리됨

## 기술 스택 (확정)
- **레포**: `nexus` (git 초기 커밋 완료)
- **모노레포**: `apps/core`(Go) + `apps/web`(프론트, 미착수)
- **Go**: 단일 모듈 `nexus`, 멀티 cmd (`cmd/supervisor`, `cmd/worker`), Go 1.26
- **DB**:
  - worker → SQLite (`modernc.org/sqlite`, CGO 없음, 설정/메타데이터만 영속)
  - supervisor → PostgreSQL (`pgx/v5` native 타입)
- **sqlc**: 두 엔진 한 `sqlc.yaml`에 정의 → `workerdb` / `superdb` 패키지로 분리 생성
- **마이그레이션**: goose (스키마 단일 진실 = goose 마이그레이션 파일, sqlc가 schema로 읽음)
- **터미널 출력**: xterm.js (프론트 착수 시)

## 로컬 개발 환경 (DB)
- **supervisor PostgreSQL = docker 컨테이너 `postgres15`**(이미지 `postgres:15`). 시작: `docker start postgres15` (※ `docker start postgres`는 그런 이름 컨테이너 없음 — 보여지는 건 이미지명)
- 컨테이너 env가 supervisor env와 일치: user `irony` / pass `!Fa1289` / 포트 `5432:5432`. **단 컨테이너 기본 DB는 `workspace`**(구 PortBridge) → nexus용 `nexus` DB는 **수동 생성함**(`CREATE DATABASE nexus OWNER irony;`, 2026-06-26 1회 완료)
- supervisor 띄우면 `init()`의 mountStore가 goose 마이그레이션 자동 적용(멱등)
- **함정 재확인**: 잔류 supervisor 프로세스가 5050 점유 시 `bind: address already in use`(마이그레이션은 init서 돌아 정상, 서버 바인드만 실패). `ss -ltnp | grep :5050`로 pid 찾아 kill

## apps/core 디렉토리 (스캐폴딩 완료)
```
cmd/{supervisor,worker}/main.go   진입점 (현재 스텁, go run OK)
internal/
  protocol/ transport/ execute/   공유 (현재 doc.go 스텁)
  worker/{manager, db/{migrations,query,gen}}        gen=workerdb
  supervisor/{registry,bind,subscribe, db/{...}}     gen=superdb
sqlc.yaml  README.md  .gitignore
```
- 빌드/실행/`sqlc generate` 검증됨. 의존성은 pgx/v5만 추가 (생성 코드가 import)
- modernc.org/sqlite, pressly/goose는 실제 코드 작성 시 추가 예정

## SQLite + sqlc 학습 (재사용)
- **modernc.org/sqlite**: CGO 불필요, 순수 Go, SQLite 번들 → 배포 장치 외부 종속성 없음 (v1.49.1 = SQLite 3.49.x)
- **WAL 모드**: `PRAGMA journal_mode = WAL` → 읽기/쓰기 동시 가능 (`-wal`, `-shm` 파일 생성)
- **foreign_keys**: `PRAGMA foreign_keys = ON` 연결마다 필요 (기본 OFF)
- **Unix timestamp**: `INTEGER DEFAULT (unixepoch())` (SQLite 3.38+)
- **nullable overrides**: `db_type + nullable: true` 조합은 SQLite에서 동작 안 함 → `column: "table.col"` 방식으로 각각 명시
- **nullable Go 타입**: `sql.NullInt64`/`sql.NullString` → `.Valid`/`.Int64`/`.String` 접근
- **sqlc 파라미터**: `sqlc.arg(name)` 필수, `sqlc.narg(name)` nullable. `?N` 번호와 unnamed `?` 혼용 금지
- **process 영속성 설계**: 휘발성(메모리 manager) vs 영속성(DB) — agent는 휘발, 서버는 DB로 재시작 복구
- **PRAGMA는 DSN `_pragma=`로 (중요)**: `foreign_keys`/`busy_timeout`은 비영구·커넥션별. 풀에서 `db.Exec("PRAGMA ...")`는 *커넥션 1개*에만 먹어 누락 발생 → modernc는 `sql.Open("sqlite","file?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")`로 **매 커넥션 적용**. (journal_mode는 영구라 어느 쪽이든 OK)
- **goose 멱등**: `goose_db_version` 테이블로 버전 추적 → `Up`은 미적용분만 적용(데이터 보존), 매 실행 호출 안전. **적용된 마이그레이션 파일 수정은 반영 안 됨** → 스키마 변경은 새 번호(00002_) 추가. 코드 자동적용은 `//go:embed`+`goose.SetBaseFS`+`goose.Up`
- **context.WithTimeout = 생성 시점부터 절대 데드라인**: 한 ctx를 여러 작업이 공유하면 "각자 N초"가 아니라 "전체 N초". 독립 타임아웃 원하면 작업별로 새 ctx

### worker store 자산 (`internal/worker/store/connectManager.go`) — 구현 완료
- PortBridge `agentStore` 이식: `StorePool` 싱글턴(DB 1개) + `Transaction`(Begin/Commit/Rollback + AfterRelease 리스너)
- `InitStorePool(dbFile, pragmas map[string]string)` (PRAGMA 옵션화) / `(s) Migrate(fsys, dir)`(goose, InitStorePool 바깥 별도 단계) / `Queries() *workerdb.Queries`
- sqlc gen은 별도 패키지 `workerdb`(`internal/worker/db/gen`) → 참조는 `workerdb.New(DBTX)`(*sql.DB·*sql.Tx 모두 가능)

### supervisor store 자산 (`internal/supervisor/store/connectManager.go`) — 구현 완료 (2026-06-26)
- 구 PortBridge `internal/store/connectManager.go`(pgx) 이식. worker판과 **API 대칭**: `InitStorePool`/`GetStorePool`/`Queries()`/`Transaction()` + Transaction `Begin/Commit/Rollback/Queries/QueriesPanic/AddAfterRelease/IsStart`
- 차이는 타입(`*pgxpool.Pool`/`pgx.Tx`)과 시그니처: `InitStorePool(user,pass,host,name string, port int)`(DSN `postgres://...` 조립 + Ping). gen 패키지 `superdb`(`superdb.New(DBTX)`, DBTX는 `*pgxpool.Pool`·`pgx.Tx` 둘 다 만족)
- worker판이 고친 두 수정 동일 반영: **afterUnsfae→afterUnsafe + 결과 err 전달**(구 버그=인자 무시하고 store.err 넘김) / **InitStorePool 재호출 시 실패연결 정리 후 재시도**(모호한 return nil 제거)
- `pgxpool` 의존성이 `puddle/v2`를 indirect로 끌어옴(go mod tidy)
- ★용도: 위 "트랜잭션 미들웨어"(요청 경계 커밋/롤백)의 풀/트랜잭션 공급원. `Transaction()`은 lazy Begin이라 읽기전용 핸들러는 트랜잭션 안 엶
- **마이그레이션**(`store/migrate.go`, worker판 대칭): `Migrate(fsys, dir)` = goose dialect `postgres`. **단, goose는 `*sql.DB` 요구 / 풀은 pgxpool** → `StorePool.dsn`(InitStorePool서 저장) 재사용해 `sql.Open("pgx", dsn)`로 **마이그레이션 1회용 `*sql.DB`**를 열고 닫음(풀과 무관). `_ "github.com/jackc/pgx/v5/stdlib"`로 "pgx" 드라이버 등록. `db/migrations/embed.go`=`//go:embed *.sql`
- **배선**: `cmd/supervisor/main.go init()` → `mountStore(GetEnv())`(InitStorePool→Migrate, 실패 Fatalf). env: `cmd/supervisor/constants/env.go`에 DBUser/Pass/Name/Host(필수)+DBPort(기본5432). **인자 순서**: `InitStorePool(user, pass, host string, port int, name string)`(port가 name보다 앞 — 사용자가 조정)

### 트랜잭션 미들웨어 (`cmd/supervisor/router/tx.go`) — 구현 완료 (2026-06-26)
- 구 PortBridge `web/requestContext.go`(전역 map+RequestScope+도메인 *Context 4종) 대체. **전역 map 폐기**(echo.Context의 c.Set/c.Get이 요청 수명 저장소) + 4개 중복 Context → **단일 `txScope` + 미들웨어 1개**
- `txScope{pool,tx,err}`: 요청 1건 트랜잭션 수명. 요청마다 새로 만들어 echo.Context에 실림 → **단일 goroutine 전용, 락 불필요**
- **lazy Begin**: `ensureTx() (*Transaction, error)` 첫 호출 때만 `pool.Transaction()+Begin`(실패 시 `web.Err(500)` **반환**) → 읽기전용 핸들러는 트랜잭션 안 엶. 핸들러에서 `q,err:=TxQueries(c)`(둘 다 `(…,error)`)로 사용
- `release()`(요청당 1회): err 있으면 Rollback, 없으면 Commit. tx==nil이면 no-op. 커밋/롤백 실패는 로그만(응답 이미 나간 뒤라 못 바꿈)
- **`txMiddleware(pool)`**: scope 생성→c.Set→defer{recover면 err기록+release+재panic / 정상이면 err=핸들러반환값+release}. **error 반환·panic 둘 다 롤백**(주 경로=error 반환[return-style], panic은 안전망). 등록 위치=PanicMiddleware 안쪽
- **등록 위치 중요**: `e.Use` 순서 = `PanicMiddleware → Log → txMiddleware → CORS`. tx는 **PanicMiddleware 안쪽**이라야 재-panic이 PanicMiddleware로 전파돼 JSON 응답됨. WS 업그레이드 라우트도 전역 래핑되나 Tx() 안 부르면 무해(no-op)
- **commit-in-defer는 c.JSON 이후 실행**(응답 먼저, 커밋 나중) — 구 HttpProcess와 동일 한계, 현재 허용

### ★에러 처리 규약: panic-style로 회귀 (PanicMiddleware) — 2026-06-26 재확정 (return-style 폐기)
- **결론 번복**: 아래 return-style은 사용자가 "맘에 안 든다"고 되돌림 → **다시 panic-style**. `util.go`를 panic 기반으로 재반입(`ClientError{Status int, Error error}` = `Error`가 **필드**라 error 인터페이스 미구현 / `Err(...)`는 `ClientError` 반환 / `PanicMiddleware`가 recover→ClientError면 `{message,type}` JSON, 아니면 500). **`HTTPErrorHandler` 없음**
- **현재 규약(panic-style)**: 핸들러는 `panic(web.Err(status, fmt, ...))`. tx 헬퍼 `ensureTx()/Tx(c)/TxQueries(c)`는 **error 반환 제거**(panic; TxQueries=`Tx(c).QueriesPanic()`). `requireSession`도 단일 반환+401 panic. DB/Save 에러도 `panic(web.Err(500,"%v",err))`로 통일. 성공만 `return c.JSON(...)`. txMiddleware는 불변(recover→롤백→재-panic→PanicMiddleware 렌더). **빌드/vet 통과**(2026-06-26)
- 아래는 **폐기된 과거 결정**(참고용 보존):
- ~~**핸들러는 panic 대신 `error`를 반환한다**. `web.Err(status, format, ...)` → `web.ClientError`(이제 `error` 구현) 반환 → `e.HTTPErrorHandler = web.HTTPErrorHandler`가 JSON `{message,type}`로 렌더. type=SERVER(>=500)/CLIENT~~
- **계기**: `web.Panic`(래퍼 함수)은 Go "종료문(terminating statement)"이 아니라 컴파일러/분석기가 "이후 실행 안 됨"을 모름 → `missing return` / gopls nilness 오경고(`getSessionOrPanic` 뒤 `return sess`가 nil 가능하다고 판단). 빌트인 `panic(...)`만 종료문. 대안 2: `panic(web.Err(...))`(빌트인이라 경고 해소되나 비관용·untyped) vs **return(타입안전·echo 정석)**. 코드베이스 일관성 위해 헬퍼도 함께 전환 → **return 채택**
- **전환 범위**: `web.Panic`·`web.HttpProcess` **삭제**. `ClientError{Status,Msg}`+`Error()`. `PanicMiddleware`는 **안전망으로 단순화**(예기치 못한 panic/ tx 재-panic을 recover→error 환원→HTTPErrorHandler로). `Tx(c)`/`TxQueries(c)`/`ensureTx`가 **`(…, error)` 반환**(구 `QueriesPanic`·`web.Panic(500)` 대신 `t.Queries()`·`web.Err`). 핸들러는 `q, err := TxQueries(c); if err!=nil { return err }`
- **여전히 필요한 panic 인프라**: 진짜 버그(nil 역참조)·`txMiddleware` 재-panic 때문에 PanicMiddleware(또는 recover)는 유지. tx 롤백은 error 반환·panic 둘 다 처리(미들웨어 defer)
- e2e 9시나리오 재검증 통과(동작 불변, 경고만 소멸). 향후 모든 supervisor echo 핸들러 이 규약 따를 것

### user 가입/로그인 핸들러 (`cmd/supervisor/router/user.go`) — 구현+e2e 완료 (2026-06-26)
- **web.HttpProcess 폐기, echo 기본 핸들러**(func(c)error, **return web.Err** + `q,err:=TxQueries(c)`). 라우트: POST `/users`(가입)·POST `/users/session`(로그인)·GET(세션확인)·DELETE(로그아웃). `supervisorRouter.mountUsers(e)`
- **★별도 User 타입 안 만듦**: 세션이 `superdb.User`의 export 기본형 필드 직접 직렬화 → `session.SessionManager[superdb.User]` 그대로(키 `"irony"/"sid"`, nameFn=nil). supervisorRouter에 `sessions` 필드
- `requireSession(c) (*SessionElement, error)`(구 getSessionOrPanic) = 구 flag 사슬 → `sess==nil||IsNew||Data.ID==0` 단락 OR → `web.Err(401)`. `toHash`(sha256 hex)로 식별자·비번 저장/조회
- **함정**: checkSession은 세션 복원값이라 `pgtype.Timestamptz`(CreatedAt/UpdatedAt) **0/누락**(gob 직렬화 대상 아님). 타임스탬프 필요하면 `sess.Data.ID`로 DB 재조회. 응답 identification=해시. Password 해시가 세션 쿠키에 적재됨
- e2e 8시나리오 통과(가입/중복/미로그인/로그인/세션확인/오답/로그아웃/로그아웃후). 세션키 `"irony"` 하드코딩 → env화 잔여

## 상세 reference
→ `PLAN-agent-comm.md` (agent↔server 통신/PTY 실행 상세)
→ `PLAN-subscription.md` (구독/토픽 모델 상세)
