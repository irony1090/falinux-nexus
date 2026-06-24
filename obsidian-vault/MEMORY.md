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

### 재사용 구독 매니저 (`internal/subscribe/manager.go`) — 구현 완료
- PortBridge subscribeManager를 도메인 제거하고 일반화한 코어
- `Manager[C comparable]`: 키는 **문자열만**, 클라이언트는 **제네릭** (`*Client`/인터페이스 모두 가능)
- 키 생성·도메인 타입은 매니저가 모름 → 프로젝트가 `Key() string`으로 생성해 넘김
- `NewManager(marshal, send)`: 직렬화/전송 전략 1회 주입 → 호출부는 `Publish(key, payload)` 한 줄
- `Publish`: marshal 1번 → 락 밖에서 N번 send, send 실패는 `errors.Join`으로 모아 반환(브로드캐스트 중단 안 함)
- `Subscribers(key) []C`: 특수 전송용 스냅샷 탈출구
- 설계 원칙: **매니저 책임 = "문자열 키 → 구독 클라이언트 집합" 장부 + 동시성**, 그 외(키/직렬화/전송)는 프로젝트 몫

### 요청/응답 상관관계 프레임 (`internal/call` + `protocol` + `transport`) — 구현 완료
- 구 PortBridge `agentRouter.go`의 통증 해결: 요청·응답이 두 파일 switch에 흩어지고 짝 ID가 없어 추적 곤란 → 도메인 매니저 조회로 역추적하던 문제
- **`call.Correlator[R]`**: "요청 ID → 응답 기다리는 채널" 장부 + 동시성만. send 주입, 응답 타입 R 자유. `Call(ctx,payload)`(블록)/`Resolve(id,val,err)`/`Close(err)`(전부 깨움). subscribe.Manager와 동일 철학(코어는 도메인/전송 모름)
- **`protocol.Frame{Kind,ID,Type,Err,Data}`**: Kind=**REQ/RES** 로 요청/응답 프로토콜 레벨 구분. Data=json.RawMessage(자유 payload). REQ가 ID 채번 → RES가 반사
- **`transport.Conn`**: `Call`(REQ→RES 블록)/`Handle`(REQ 핸들러)/`Serve`(수신 루프). supervisor·worker 양쪽 대칭 API. `MessageRW` 인터페이스에만 의존(gorilla `*websocket.Conn`이 그대로 만족)
- 연결 끊김 시 `corr.Close(err)`로 매달린 Call 일괄 실패 = 앞서 정한 Done(502) reconciliation과 같은 정신
- **예제**: 각 main에 supervisor(WS서버) ↔ worker(WS클라) 1요청→6응답. 동작 확인됨
- **향후 확장**: 출력 스트리밍은 응답 없는 단방향이라 Call로 하면 안 됨 → EVENT(Kind 추가) + On/Emit 별도 평면으로 재도입 예정. 인메모리 테스트는 `transport.Pipe()` 패턴. (둘 다 이번에 만들었다가 예제 최소화로 제거, git 히스토리에 있음)

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
- **agent 식별자: 사전 지정 고유키(메인키)** + supervisor가 접속 시 부여하는 **서브키**. 저장/조회는 `메인키#서브키`(`InstanceKey()`). MAC 폐기
- **process 영속성**: worker 휘발(메모리) / supervisor 영속(PG, 재시작 복구)
- **재연결 reconciliation**: 끊김→`Done(502)` 비관적 정리 / 재연결→worker가 live 스냅샷 보내 재동기화. 미구현(supervisor 본격 작업 시)

## 이번 세션 재사용 자산/교훈 (2026-06-19)

### transport.Conn 수명 상태
- `ConnState`(StateActive=0/StateClosed) + `State()` + `String()`. `atomic.Int32` 필드(zero=Active). Serve 수신루프 종료·Close에서 StateClosed 전이(둘 다 idempotent)
- 도메인 무지·대칭 유지 — "연결이 살아있나"는 conn 속성, **신원(메인키#서브키)은 라우터 몫**(conn 아님)

### `internal/util/string.util.go` — RandomKey
- `RandomKey(length int, prefix, suffix string) (string, error)`: base62(`0-9A-Za-z`), length=prefix/suffix **포함** 전체 길이(결과 정확히 length), `crypto/rand`
- 62는 256 약수 아님 → **거부 샘플링**(>=248 바이트 버림)으로 modulo 편향 제거 (hex 16이면 `b&0x0f`로 충분했음). 타입: `maxByte byte`, 인덱스 `b%byte(len(alphabet))`
- 용도: 서브키 발급, transferId. `*.util.go`가 util 패키지 컨벤션

### `internal/manager/KeyValManager` (PortBridge 이식)
- 제네릭 `KeyValManager[K comparable, V any]`: 키→**단일 값** 장부 + `OnCreated`/`OnRemoved` 수명 콜백. `subscribe.Manager`(키→**집합**)와 역할 다름
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

## 상세 reference
→ `PLAN-agent-comm.md` (agent↔server 통신/PTY 실행 상세)
→ `PLAN-subscription.md` (구독/토픽 모델 상세)
