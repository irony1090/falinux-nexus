# CURRENT

## 현재 날짜
2026-06-18

## 현재 작업 중인 폴더/프로젝트
- `/home/jh-bae/irony/nexus/apps/core`
- **worker 정체성 영속 한 바퀴 완료** (연결→마이그레이션→등록→저장→재접속, 동작 확인)
- **supervisor는 임시 골격 상태** — 본격 작업은 "worker 파일 전송 모듈" 착수 시 (아래 supervisor 상태 참조)

## 진행 상황 (요약 — 상세는 HISTORY 2026-06-18)
- 프로젝트 **Nexus** 분리, 모노레포 `apps/core`(Go) + `apps/web`(미착수), git 초기 커밋
- apps/core 스캐폴딩 (단일 모듈, `cmd/supervisor`·`cmd/worker`), sqlc 두 엔진(workerdb/superdb), goose
- **재사용 구독 매니저** `internal/subscribe/manager.go` (`Manager[C]`, 도메인 제거·일반화)
- **요청/응답 프레임 (3층)**: `call.Correlator[R]`(엔진) + `protocol.Frame`(REQ/RES 봉투+어휘) + `transport.Conn`(Call/Handle/Serve). correlator=고정, protocol=어휘 추가, conn 위 Handle/Call로 동작
- **등록(REGISTER) 핸드셰이크**: worker=Call / supervisor=Handle. 메인키·서브키 분리 — 내부 조회/active는 `메인키#서브키`, worker엔 서브키(숫자)만 반환. 4시나리오 검증
- **worker SQLite 정체성 영속 완료**:
  - `identity` 테이블(main_key PK, sub_key) + sqlc(GetIdentity/UpsertIdentity)
  - `internal/worker/store/connectManager.go` — `StorePool`+`Transaction` 이식, **PRAGMA를 DSN `_pragma=` 옵션으로**(커넥션별 적용)
  - goose 자동 마이그레이션(`migrate.go` + 임베드, InitStorePool 바깥 별도 단계, 멱등)
  - `workerRouter.register()`가 GetIdentity→register→UpsertIdentity 배선
  - 의존성: `modernc.org/sqlite`, `pressly/goose/v3`

## 결정 사항 (확정)
- Go 모듈: 단일 `nexus`, 멀티 cmd / PG: pgx/v5 / worker SQLite: 설정·메타데이터만 영속(process는 메모리)
- 마이그레이션: goose (스키마 단일 진실 = 마이그레이션 파일, sqlc가 schema로 읽음)
- **계층: 2단계 고정** — supervisor → worker만, 중간 노드 없음. 식별자/라우팅/구독 1홉 평면
- **agent 식별자: 사전 지정 고유키(메인키)** — MAC 폐기. + supervisor가 접속 시 **서브키(숫자)** 부여해 인스턴스 구분. 저장/조회는 `메인키#서브키`
- **process 영속성**: worker 휘발 / supervisor 영속(PG, 재시작 복구)
- **worker 재연결 reconciliation**: 끊김→`Done(502)` 비관적 정리 / 재연결→worker가 live 스냅샷 보내 재동기화(정정). 미구현(supervisor 본격 작업 시)

## 컨셉 정리
- **Worker Agent**: 본인 process 관리 (실행/모니터링/종료)
- **Supervisor Agent**: 다수 worker 관리 (연결/상태/명령 분배)
- 계승 자산: `IInteractive`/`AgentInteractive` 추상화, agent↔server 프로토콜, 구독/토픽 모델, SQLite/sqlc

## supervisor 상태 (중요)
- **supervisor는 현재 전부 "임시 로직"** (registry 메모리, REGISTER 핸들러 등 최소 구현). worker 흐름을 굴리기 위한 임시 골격.
- **본격 작업은 "worker 파일 전송 모듈" 착수 시 시작.** (파일 전송 = PortBridge `UploadInit→Ready→Chunk→Status→Result` 계승 — PLAN-agent-comm.md). 그때 supervisor registry/구독/PG영속 제대로 구현.

## 다음 할 일
1. **worker 파일 전송 모듈** 착수 → 이와 함께 supervisor 본격 구현 시작
2. 도메인 MsgType + payload 확장 (STATUS/EXEC/KILL/OUTPUT/SNAPSHOT) — protocol에 어휘 추가
3. supervisor registry 메모리→PG 영속화 (재시작 복구)
4. supervisor 도메인 구독 계층: 키 구조체 + 클라이언트 타입(`*transport.Conn`?) + `subscribe.NewManager` 전략 주입
5. `internal/execute` IInteractive/AgentInteractive를 transport.Conn 위에 연결 — 스트리밍 필요 시 EVENT 재도입
6. 최소 동작 골격: supervisor 1 + worker 1 → process 실행/스트리밍 PoC
7. (이후) `apps/web` 프론트 착수 여부

## 미해결 이슈 / 결정 필요
- 서브키 생성/충돌: 현재 숫자(전역 메모리 seq), 저장 `메인키#서브키` → supervisor 재시작 시 seq 리셋되면 옛 서브키와 충돌 가능. PG 영속 + 충돌 안 나는 생성(UUID 등) 검토
- 발급 이력 미검증: 재접속은 active 아니면 통과(임의 서브키 위조 수락), 최초는 active 체크 건너뜀 → 선점된 미래 seq와 겹칠 수 있음. key↔subkey 매핑 저장·대조로 닫힘
- 스냅샷 재동기화 프로토콜 메시지 정의 (재연결 시 worker→supervisor `SNAPSHOT` 포맷)
- (참고) `NewWorkerRouter` defer/goroutine register 흐름, `serveErr` cap1 송신자 2개 누수 가능 — 미수정
