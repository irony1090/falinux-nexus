# nexus / apps/core

Nexus의 Go 코어. 단일 모듈(`module nexus`), 멀티 cmd.

- **supervisor** — 다수 worker agent를 관리하는 총괄 서버 (PostgreSQL / pgx/v5)
- **worker** — 장치에 배포되어 자기 process를 관리하는 에이전트 (SQLite / modernc.org/sqlite)

## 디렉토리

```
cmd/
  supervisor/       총괄 서버 진입점
  worker/           장치 에이전트 진입점
internal/
  protocol/         공유: 메시지/패킷 정의 (단일 진실)
  transport/        공유: WebSocket 송수신
  execute/          공유: IInteractive / AgentInteractive
  worker/
    manager/        worker process 휘발성 관리
    db/migrations/  goose 마이그레이션 (SQLite)
    db/query/       sqlc 쿼리
    db/gen/         sqlc 생성 코드 (package workerdb)
  supervisor/
    registry/       연결된 worker 관리
    bind/           process ↔ client 중계
    subscribe/      topic 구독 매니저
    db/migrations/  goose 마이그레이션 (PostgreSQL)
    db/query/       sqlc 쿼리
    db/gen/         sqlc 생성 코드 (package superdb)
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

## 다음 작업 (스캐폴드 채우기)

- `internal/protocol` — 메시지 타입 정의 (vault `PLAN-agent-comm.md`)
- `internal/transport` — WebSocket 연결/패킷
- `internal/execute` — `IInteractive`, `AgentInteractive`
- 아직 채울 의존성: `modernc.org/sqlite`(worker DB 코드 작성 시), `github.com/pressly/goose/v3`(자동 마이그레이션 시)
  → 코드에서 import 후 `go mod tidy`
