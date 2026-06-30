# Nexus

계층형 에이전트·프로세스 관리 플랫폼.

- **Worker Agent** — 장치에 배포되어 **자기 자신의 process를 관리**한다 (실행/모니터링/종료).
- **Supervisor Agent** — **다수의 worker를 총괄 관리**한다 (연결/상태/명령 분배).

## 구조

```
nexus/
├── apps/
│   ├── core/       Go 코어 — supervisor / worker (단일 모듈, 멀티 cmd)
│   └── frontend/   프론트엔드 — Vue3 + TS + Vuetify (supervisor용 UI)
└── obsidian-vault/   설계·진행 기록
```

## 기술 스택

| 영역 | 선택 |
|------|------|
| 언어 | Go (단일 모듈 `nexus`) |
| 프론트 | Vue 3 + TypeScript + Vuetify 4 (Vite, Pinia 미사용) |
| worker DB | SQLite (`modernc.org/sqlite`) |
| supervisor DB | PostgreSQL (`pgx/v5`) |
| 쿼리 코드 생성 | sqlc |
| 마이그레이션 | goose |
| 통신 | WebSocket — REQ/RES + EVENT 프레임. supervisor↔worker, supervisor↔웹이 동일 transport 공용 |

## 통신 모델

한 WebSocket 위에 세 가지 모드를 얹는다(`internal/transport`, 양 끝 대칭):

- **REQ → RES** — 요청/응답 (id 상관). 예: 웹 → supervisor 명령.
- **EVENT** — 짝 없는 단방향 알림. 예: process 출력 스트리밍, node 변경 push.

웹 클라이언트로의 실시간 push는 `internal/subscribe.Hub`(토픽→구독자)로 fan-out하며,
**인가는 구독 시점 DB 1회 / 라우팅은 메모리**로 분리한다(상세 → `obsidian-vault/REF-realtime.md`).

## 시작하기

개발 안내(빌드·sqlc·goose)는 [`apps/core/README.md`](apps/core/README.md) 참고.

```sh
# 백엔드
cd apps/core
go build ./...
go run ./cmd/supervisor
go run ./cmd/worker

# 프론트엔드
cd apps/frontend
npm install
npm run dev
```
