# Nexus

계층형 에이전트·프로세스 관리 플랫폼.

- **Worker Agent** — 장치에 배포되어 **자기 자신의 process를 관리**한다 (실행/모니터링/종료).
- **Supervisor Agent** — **다수의 worker를 총괄 관리**한다 (연결/상태/명령 분배).

## 구조

```
nexus/
├── apps/
│   ├── core/    Go 코어 — supervisor / worker (단일 모듈, 멀티 cmd)
│   └── web/     프론트엔드 (미착수)
└── obsidian-vault/   설계·진행 기록
```

## 기술 스택

| 영역 | 선택 |
|------|------|
| 언어 | Go (단일 모듈 `nexus`) |
| worker DB | SQLite (`modernc.org/sqlite`) |
| supervisor DB | PostgreSQL (`pgx/v5`) |
| 쿼리 코드 생성 | sqlc |
| 마이그레이션 | goose |
| 통신 | WebSocket |

## 시작하기

개발 안내(빌드·sqlc·goose)는 [`apps/core/README.md`](apps/core/README.md) 참고.

```sh
cd apps/core
go build ./...
go run ./cmd/supervisor
go run ./cmd/worker
```
