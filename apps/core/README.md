# nexus / apps/core

Nexus의 Go 코어. 단일 모듈(`module nexus`), 멀티 cmd.

- **supervisor** — 다수 worker agent를 관리하는 총괄 서버 (PostgreSQL / pgx/v5)
- **worker** — 장치에 배포되어 자기 process를 관리하는 에이전트 (SQLite / modernc.org/sqlite)

## 디렉토리

```
cmd/
  supervisor/       총괄 서버 진입점 (main.go, router/, constants/)
  worker/           장치 에이전트 진입점 (main.go, router/, constants/)
internal/
  protocol/         공유: 메시지/프레임 정의 (단일 진실)
  transport/        공유: WebSocket 위 REQ/RES + EVENT (Conn, 양 끝 대칭)
  call/             공유: 요청/응답 상관관계 (Correlator)
  execute/          공유: IInteractive / AgentInteractive
    pty/            실제 PTY 상호작용
  subscribe/        재사용: 토픽→구독자 Hub (실시간 push fan-out)
  transfer/         재사용: 파일/버퍼 전송
  manager/          재사용: KeyVal·세션 매니저
  patch/            재사용: 3-state PATCH 필드 (patch.Field[T])
  web/              재사용: 에러·미들웨어 등 HTTP 유틸
  util/             재사용: 공용 유틸
  worker/
    manager/        worker process 휘발성 관리
    store/          worker 스토어
    db/{migrations,query,gen}   goose(SQLite) / sqlc 쿼리 / 생성코드(workerdb)
  supervisor/
    registry/       연결된 worker 관리 (메모리, 라우팅 authority)
    bind/           process ↔ client 중계
    subscribe/      supervisor 전용 구독
    store/          supervisor 스토어 (PG, tx)
    db/{migrations,query,gen}   goose(PostgreSQL) / sqlc 쿼리 / 생성코드(superdb)
```

## 도구 설치

```sh
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## sqlc — 쿼리 코드 생성

`sqlc.yaml`에 두 엔진이 정의돼 있다. 스키마 입력은 goose 마이그레이션 디렉토리를 그대로 사용한다(스키마 단일 진실).

```sh
sqlc generate
```

## goose — 마이그레이션

```sh
# worker (SQLite)
goose -dir internal/worker/db/migrations sqlite3 ./worker.db up

# supervisor (PostgreSQL)
goose -dir internal/supervisor/db/migrations postgres "postgres://user:pass@localhost:5432/nexus?sslmode=disable" up

# 새 마이그레이션 생성
goose -dir internal/worker/db/migrations create add_something sql
```

> 코드 안에서 자동 적용하려면 `goose.Up(db, dir)` 사용 (worker 장치 배포에 유용).

## 빌드

```sh
go build ./...
go run ./cmd/supervisor
go run ./cmd/worker
```

## 현황 / 다음 작업

- ✅ 통신 인프라 — `protocol` / `transport`(REQ/RES + EVENT) / `call` / `subscribe`(Hub)
- ✅ DB·스토어 — sqlc·goose, worker(SQLite) / supervisor(PG, tx)
- ✅ supervisor↔웹 socket 전송계층 — `/subscribe` 엔드포인트 + 3모드 e2e 검증
- ⬜ 실시간 push 도메인 배선 — 동적 SUBSCRIBE + DB 인가, node/process CRUD commit 후 Publish
- ⬜ process 도메인 — EXEC/EDIT 배선 (worker `pty.Interactive` / supervisor `AgentInteractive`)

> 설계·진행 상세는 `obsidian-vault/`(REF-*, CURRENT.md) 참고.
