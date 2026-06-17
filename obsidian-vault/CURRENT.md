# CURRENT

## 현재 날짜
2026-06-17

## 현재 작업 중인 폴더/프로젝트
- `/home/jh-bae/irony/nexus/apps/core` — 재사용 구독 매니저 완료, 공유 패키지 채우는 단계

## 진행 상황
- 프로젝트명 **Nexus** 확정
- 이전 프로젝트(PortBridge / test-jig)에서 신규 방향으로 분리
- vault 신규 구성 완료 (PortBridge UI/그룹/아이콘 로직 폐기, agent 통신 아키텍처만 계승)
- **레이아웃 확정**: `apps/core`(Go) + `apps/web`(프론트, 미착수)
- **git 레포 `nexus` 생성 + 초기 커밋 완료**
- **apps/core 스캐폴딩 완료** (`go build ./...` OK, 두 바이너리 실행 확인)
  - 단일 모듈 `nexus`, 멀티 cmd: `cmd/supervisor`, `cmd/worker`
  - 공유 패키지: `internal/protocol`, `transport`, `execute` (현재 doc.go 스텁만)
  - worker: SQLite(modernc 예정) / supervisor: PostgreSQL(pgx/v5)
  - sqlc 두 엔진 설정 + 생성 확인 (workerdb / superdb)
  - goose 마이그레이션 디렉토리 + 00001_init 스타터(settings / agents)
  - 마이그레이션 도구 **goose** 채택
- **재사용 구독 매니저 완료** (`internal/subscribe/manager.go`)
  - 기존 PortBridge subscribeManager를 리팩터링해서 도메인 제거 + 일반화
  - `Manager[C comparable]` — 키는 문자열만, 클라이언트는 제네릭
  - 직렬화(marshal)/전송(send) 전략을 `NewManager`에서 1회 주입
  - `Publish(key, payload)`: marshal 1번 → 락 밖 N번 전송, send 실패는 `errors.Join`으로 모아 반환(중단 안 함)
  - `Subscribers(key) []C` 스냅샷 탈출구 제공
  - worker main.go로 동작 테스트 완료 후 테스트 코드 제거

## 결정 사항 (확정)
- Go 모듈: 단일 `nexus`, 멀티 cmd
- PG 드라이버: pgx/v5 (native 타입)
- worker SQLite: 설정/메타데이터만 영속 (process는 메모리)
- 마이그레이션: goose (스키마 단일 진실 = goose 마이그레이션 파일, sqlc가 이를 schema로 읽음)

## 컨셉 정리
- **Worker Agent**: 본인 process 관리 (실행/모니터링/종료)
- **Supervisor Agent**: 다수 worker agent 관리 (연결/상태/명령 분배)
- 계승 자산: `IInteractive`/`AgentInteractive` 추상화, agent↔server 프로토콜, 구독/토픽 모델, SQLite/sqlc 학습

## 다음 할 일
1. supervisor 도메인 구독 계층: 키 구조체(`Key() string`) + `*transport.Client` 확정 + `NewManager` 전략 주입
2. `internal/protocol` 메시지 타입 정의 (PLAN-agent-comm.md 기반)
3. `internal/transport` WebSocket 연결/패킷 송수신 구현
4. `internal/execute` IInteractive / AgentInteractive 구현
5. worker DB 코드 작성 시 `modernc.org/sqlite` 의존성 추가 + goose 자동 적용
6. 최소 동작 골격: supervisor 1 + worker 1 연결 → process 실행/스트리밍 PoC
7. (이후) `apps/web` 프론트 착수 여부 결정

## 미해결 이슈 / 결정 필요
- supervisor-worker 계층이 2단계 고정인지, N단계 트리인지
- agent 식별자 체계 (이전엔 MAC 주소 — Nexus에선?)
- process 영속성 범위 (worker 휘발 / supervisor 영속 유지할지)
