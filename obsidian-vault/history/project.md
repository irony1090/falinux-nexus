# HISTORY — 프로젝트 셋업 / worker 기반 (분리·스캐폴딩·SQLite 정체성)

> Nexus 프로젝트 분리, apps/core 스캐폴딩, worker router 패턴 + SQLite 정체성 영속 상세 기록.
> 요약은 `MEMORY.md`(기술 스택·디렉토리) / `REF-db.md`(store·sqlc), 현재 상황은 `CURRENT.md`.

---

### 2026-06-17 - Nexus 프로젝트 분리 & vault 신규 구성
- 이전 프로젝트(PortBridge / test-jig)의 방향성이 "테스트 장비 데스크탑 UI"에서 "2-tier 에이전트 관리 플랫폼"으로 전환됨에 따라 신규 프로젝트 **Nexus**로 분리
- 새 경로: `/home/jh-bae/irony/nexus`
- vault 구성: PortBridge 전용 자산(데스크탑 메타포 UI, 아이콘/그룹 계층, Dashboard, 제안서/와이어프레임)은 폐기. agent↔server 통신 아키텍처·PTY 실행 추상화·구독 모델·SQLite/sqlc 학습만 MEMORY로 계승
- PortBridge의 상세 HISTORY는 가져오지 않음 (test-jig 폴더에 원본 유지)

---

### 2026-06-17 - apps/core 스캐폴딩 & 레포 초기 커밋

#### 결정 사항
- 레이아웃: `apps/core`(Go) + `apps/web`(프론트)
- Go: 단일 모듈 `nexus`, 멀티 cmd (`cmd/supervisor`, `cmd/worker`)
- worker → SQLite(modernc), supervisor → PostgreSQL(pgx/v5)
- sqlc 두 엔진 (`workerdb`/`superdb`), 마이그레이션 도구 goose
- goose 마이그레이션 디렉토리를 sqlc의 schema 입력으로 사용 (스키마 단일 진실)

#### 작업 완료
- `apps/core` 디렉토리 트리 + go.mod(module nexus) 생성
- `cmd/supervisor`, `cmd/worker` 스텁 main.go (go run 검증)
- 공유 패키지 스텁: protocol / transport / execute (doc.go)
- supervisor 패키지 스텁: registry / bind / subscribe
- `sqlc.yaml` 두 엔진 정의 → `sqlc generate` 정상 (workerdb, superdb)
- goose 스타터 마이그레이션 00001_init (worker: settings / supervisor: agents)
- sqlc 스타터 쿼리 (settings.sql / agents.sql)
- README.md(도구 설치·sqlc·goose·빌드 안내), .gitignore
- `go build ./...` OK, pgx/v5 의존성 추가
- git 레포 `nexus` 생성 + 초기 커밋 완료

#### 메모
- 00001_init의 settings/agents 테이블은 sqlc 생성을 위한 스타터 → 실제 도메인 정해지면 교체 (※ agents는 2026-06-26 worker 메모리 전용 확정으로 폐기, users가 00001로 재정렬)
- modernc.org/sqlite, pressly/goose는 코드에서 import 후 `go mod tidy` 시 추가됨

---

### 2026-06-18 - worker 리팩토링(router 패턴) + SQLite 정체성 영속

> 상태: **worker 정체성 영속 한 바퀴 완료** (연결→마이그레이션→등록→저장→재접속). 동작 확인됨.

#### worker 리팩토링 (사용자 주도)
- `cmd/worker/main.go`를 얇은 진입점으로: env 로드 + `mountStore()` + `router.NewWorkerRouter`
- `cmd/worker/router/workerRouter.go` 신설: `workerRouter{conn, store, uniqueKey}` 가 dial+Serve+register 소유 (구 PortBridge agentRouter 패턴)
- `cmd/worker/constants/env.go`: env 기반 설정 (NAME / WS_HOST / WS_SCHEME / SQLITE). `pathToAbsolutePath(projectDir, ...)`로 SQLITE 경로 해석
- 서브키 채번을 **메인키·서브키 분리**로 변경: supervisor가 내부 조회/active 판정은 합친 키 `메인키#서브키`(`req.InstanceKey()`)로, worker엔 **서브키(숫자)만** 반환. `protocol.RegisterRequest.InstanceKey()` 추가

#### sqlc + identity 테이블
- 단일 `sqlc.yaml` 두 엔진(sqlite=workerdb / postgres=superdb). 엔진은 블록에 지정 → `sqlc generate` 1회로 둘 다 생성. sqlc는 DB 접속 없이 마이그레이션 .sql을 정적 파싱
- worker `identity` 테이블: `main_key TEXT PRIMARY KEY, sub_key TEXT NOT NULL, updated_at, UNIQUE(main_key,sub_key)` (복합유니크는 PK라 사실상 중복 — 사용자 요청대로 둠). 스타터 `settings` 교체
- `query/identity.sql`: `GetIdentity(:one)`, `UpsertIdentity(:exec, ON CONFLICT(main_key) DO UPDATE)`

#### connectManager 이식 (`internal/worker/store/connectManager.go`)
- 구 PortBridge `agentStore`에서 `StorePool` 싱글턴 + `Transaction` 레이어 전체 이식 (트랜잭션은 단일문장엔 불필요하나 곧 쓸 인프라라 통째로)
- sqlc 참조를 `workerdb.New`/`*workerdb.Queries`로 교체 (gen이 별도 패키지)
- **PRAGMA 옵션화**: `InitStorePool(dbFile, pragmas map[string]string)` → DSN `_pragma=key(val)`로 실음. **핵심**: `foreign_keys` 등은 커넥션별 설정인데 `db.Exec("PRAGMA")`는 풀의 커넥션 1개에만 먹음(구 코드 잠재버그). DSN 방식은 매 커넥션 적용 → 다중 커넥션 검증 통과
- 버그 수정: `afterUnsfae`→`afterUnsafe` + 인자 무시(커밋=nil/롤백=err 결과를 리스너에 제대로 전달), init "실패시 return nil" 모호함 제거

#### goose 자동 마이그레이션 (InitStorePool 바깥, 별도 단계)
- `internal/worker/db/migrations/embed.go`: `//go:embed *.sql` (배포 바이너리 단독 실행)
- `internal/worker/store/migrate.go`: `(s *StorePool) Migrate(fsys, dir)` = SetBaseFS+SetDialect("sqlite3")+goose.Up
- `cmd/worker/main.go` `mountStore()`: InitStorePool(PRAGMA) → Migrate
- 학습: goose는 `goose_db_version`로 버전 추적 → `Up`은 멱등(미적용분만, 데이터 보존). 적용된 마이그레이션 파일 수정은 반영 안 됨(새 번호 추가해야)
- `modernc.org/sqlite v1.52.0`, `github.com/pressly/goose/v3 v3.27.1` 의존성 추가

#### 정체성 저장 배선 + 버그
- `workerRouter.register()`: 시작 시 `GetIdentity`→재접속 서브키 로드, `Call(REGISTER)` 응답 서브키가 기존과 다르면 `UpsertIdentity` 저장. DB 호출마다 독립 ctx
- 학습: `context.WithTimeout`은 **생성 시점부터 절대 데드라인** → 한 ctx 공유 시 전체가 그 예산. 작업별 독립 ctx로 분리
- **버그(잡음)**: supervisor `register()`를 `InstanceKey()`로 리팩터링하며 새 서브키를 `req.SubKey`에 넣었는데 반환은 옛 `sub`("")를 함 → 빈 서브키 응답 → worker 저장 안 됨 → 매 실행 최초접속. 수정: `return sub`→`return req.SubKey`
- housekeeping: GetIdentity 에러 `sql.ErrNoRows`(정상) vs 실제 에러 구분 로깅, UpsertIdentity 실패 로깅, stale 주석 제거

#### supervisor 상태 (당시)
- supervisor는 현재 전부 **임시 로직** — worker 흐름 굴리기용 최소 골격. **본격 작업은 파일 전송/이후 모듈 작업 시 진입** → registry/구독/PG영속 제대로 구현
