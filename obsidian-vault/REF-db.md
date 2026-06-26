# REF — DB / 스토어 계층 (SQLite·PG·sqlc·goose)

> worker SQLite / supervisor PostgreSQL 스토어 계층과 sqlc·goose 학습.
> 상세 작업 이력: worker store=`history/project.md`, supervisor store=`history/supervisor-web.md`.

## SQLite + sqlc 학습 (재사용)
- **modernc.org/sqlite**: CGO 불필요, 순수 Go, SQLite 번들 → 배포 장치 외부 종속성 없음 (v1.49.x = SQLite 3.49.x)
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

## worker store 자산 (`internal/worker/store/connectManager.go`) — 구현 완료
- PortBridge `agentStore` 이식: `StorePool` 싱글턴(DB 1개) + `Transaction`(Begin/Commit/Rollback + AfterRelease 리스너)
- `InitStorePool(dbFile, pragmas map[string]string)` (PRAGMA 옵션화) / `(s) Migrate(fsys, dir)`(goose, InitStorePool 바깥 별도 단계) / `Queries() *workerdb.Queries`
- sqlc gen은 별도 패키지 `workerdb`(`internal/worker/db/gen`) → 참조는 `workerdb.New(DBTX)`(*sql.DB·*sql.Tx 모두 가능)

## supervisor store 자산 (`internal/supervisor/store/connectManager.go`) — 구현 완료 (2026-06-26)
- 구 PortBridge `internal/store/connectManager.go`(pgx) 이식. worker판과 **API 대칭**: `InitStorePool`/`GetStorePool`/`Queries()`/`Transaction()` + Transaction `Begin/Commit/Rollback/Queries/QueriesPanic/AddAfterRelease/IsStart`
- 차이는 타입(`*pgxpool.Pool`/`pgx.Tx`)과 시그니처: `InitStorePool(user, pass, host string, port int, name string)`(DSN `postgres://...` 조립 + Ping). gen 패키지 `superdb`(`superdb.New(DBTX)`, DBTX는 `*pgxpool.Pool`·`pgx.Tx` 둘 다 만족). ★인자 순서: port가 name보다 앞(사용자 조정)
- worker판이 고친 두 수정 동일 반영: **afterUnsfae→afterUnsafe + 결과 err 전달**(구 버그=인자 무시하고 store.err 넘김) / **InitStorePool 재호출 시 실패연결 정리 후 재시도**(모호한 return nil 제거)
- `pgxpool` 의존성이 `puddle/v2`를 indirect로 끌어옴(go mod tidy)
- 용도: tx 미들웨어(`REF-supervisor-web.md`)의 풀/트랜잭션 공급원. `Transaction()`은 lazy Begin이라 읽기전용 핸들러는 트랜잭션 안 엶
- **마이그레이션**(`store/migrate.go`, worker판 대칭): `Migrate(fsys, dir)` = goose dialect `postgres`. **단, goose는 `*sql.DB` 요구 / 풀은 pgxpool** → `StorePool.dsn`(InitStorePool서 저장) 재사용해 `sql.Open("pgx", dsn)`로 **마이그레이션 1회용 `*sql.DB`**를 열고 닫음(풀과 무관). `_ "github.com/jackc/pgx/v5/stdlib"`로 "pgx" 드라이버 등록. `db/migrations/embed.go`=`//go:embed *.sql`
- **배선**: `cmd/supervisor/main.go init()` → `mountStore(GetEnv())`(InitStorePool→Migrate, 실패 Fatalf). env: `cmd/supervisor/constants/env.go`에 DBUser/Pass/Name/Host(필수)+DBPort(기본5432)

## processes 모델 (supervisor superdb) — 설계 확정 (2026-06-26)
> process **영속 = supervisor PG 전용**(worker 메모리 전용이라 worker DB엔 없음). in-memory 레지스트리의 **write-through 영속층**(재시작 복구·frontend 목록). 도메인 배선과 1:1. → process 배선 `REF-process.md`, 현재 `CURRENT.md`.
- **이번 범위 = `processes` 메타데이터 1테이블만**. 출력 스크롤백 영속(DB sink)=**나중**(frontend SNAPSHOT 붙일 때 `process_outputs` 별 테이블, YAGNI)
- **마이그레이션 = `00002_processes.sql`** (nodes/labels → 00003/00004로 밀림)
- **키/타입 결정(사용자 확정)**:
  - `uid TEXT PRIMARY KEY` — 와이어 라우팅 키(RandomKey) 그대로 PK(서로게이트 BIGSERIAL 아님). 외부 의미키라 users와 다름
  - `type VARCHAR(16)` CHECK `IN ('EXEC','EDIT')` / `status VARCHAR(16)` CHECK `IN ('PENDING','PROCESS','COMPLETED','FAILED')` — **상수 대문자**. ⚠️ status는 `execute.CommandStatus`(uint enum)와 **양방향 매핑 1개** 필요(String()을 이 대문자에 맞추거나 변환 맵)
  - `owner_user_id BIGINT NOT NULL REFERENCES users(id)` / `node_id BIGINT NULL` **FK 없음**(nodes 미생성·ad-hoc 셸도 있음 → nodes 생기면 FK 추가 검토) / `device_key TEXT NOT NULL` worker InstanceKey, **FK 아님**(worker 메모리 전용)
  - spec: `cmd TEXT`, `args/env TEXT[] DEFAULT '{}'`(pgx/v5 → []string), `cwd TEXT`, `rows/cols SMALLINT`(마지막 PTY 크기)
  - lifecycle: `pid INT NULL`(PROCESS 시), `exit_code INT NULL`(종료 시), `created_at`/`started_at NULL`/`finished_at NULL`/`updated_at NULL`
- **가변 레코드(사용자 확정)**: process 1개=행 1개, 상태전이=`UPDATE`로 덮어씀(전이 이력 안 남김). 감사 로그 필요 시 별 테이블(YAGNI). append-only 폐기
- 인덱스: `owner`, `device`, **부분 `active`(device_key) WHERE status IN ('PENDING','PROCESS')**(재시작 복구·미종료 조회)
- **sqlc 쿼리**(`query/processes.sql`): `CreateProcess :one`(status 기본 PENDING) / `GetProcess :one` / `MarkProcessRunning :one`(status=PROCESS,pid,started_at=NOW) / `MarkProcessDone :one`(status,exit_code,finished_at=NOW) / `UpdateProcessLayout :exec`(선택) / `ListProcessesByOwner :many` / `ListActiveByDevice :many`(WHERE device_key AND status IN PENDING,PROCESS)
- **write-through 접점**: `Exec()`→CreateProcess(PENDING)→Call(MsgExec) / status 핸들러 PROCESS→MarkRunning, COMPLETED·FAILED→MarkDone / worker 끊김→ListActiveByDevice→각 Done(502)+MarkDone

## 마이그레이션 현황
- supervisor: `00001_users.sql` 적용 완료(users + goose_db_version, 실 nexus DB 검증). 다음: **`00002_processes.sql`**(설계 확정, 위), 이후 `00003_nodes.sql`/`00004_labels.sql`
- worker: `identity` 테이블(main_key PK / sub_key / updated_at)
