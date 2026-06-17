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

## 상세 reference
→ `PLAN-agent-comm.md` (agent↔server 통신/PTY 실행 상세)
→ `PLAN-subscription.md` (구독/토픽 모델 상세)
