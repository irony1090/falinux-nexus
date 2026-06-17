# HISTORY

> CURRENT/MEMORY가 비대해지는 것을 막기 위한 상세 기록 저장소.
> 작업 상태가 바뀌면 기존 CURRENT 내용을 여기로 정리해 내린다.

## 과거 작업 기록

### 2026-06-17 - Nexus 프로젝트 분리 & vault 신규 구성
- 이전 프로젝트(PortBridge / test-jig)의 방향성이 "테스트 장비 데스크탑 UI"에서
  "2-tier 에이전트 관리 플랫폼"으로 전환됨에 따라 신규 프로젝트 **Nexus**로 분리
- 새 경로: `/home/jh-bae/irony/nexus`
- vault 구성: PortBridge 전용 자산(데스크탑 메타포 UI, 아이콘/그룹 계층, Dashboard,
  제안서/와이어프레임)은 폐기. agent↔server 통신 아키텍처·PTY 실행 추상화·구독 모델·
  SQLite/sqlc 학습만 MEMORY로 계승
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
- 00001_init의 settings/agents 테이블은 sqlc 생성을 위한 스타터 → 실제 도메인 정해지면 교체
- modernc.org/sqlite, pressly/goose는 코드에서 import 후 `go mod tidy` 시 추가됨

---

### 2026-06-17 - 재사용 구독 매니저 리팩터링 (subscribe.Manager)

#### 배경
- PortBridge `internal/manager/subscribeManager.go`를 Nexus로 이식하려는데
  확장성/가독성 문제 발견 → 도메인 제거 + 일반화하기로 결정

#### 기존 코드의 문제
- 도메인이 매니저에 박힘 (SubKeyGroup/Icon/Process, SubscribeKeyXxx, R/C/U/D, SubscribeResponse)
- `*web.Client`에 강결합
- `WriteObj`가 key 해석 + json.Marshal + websocket 전송을 다 함 (책임 혼재)
- 죽은 주석 코드 대량

#### 설계 결정
- 매니저 책임을 "문자열 키 → 구독 클라이언트 집합 장부 + 동시성"으로 축소
- 키 생성·직렬화·전송은 전부 프로젝트(호출자) 몫으로 분리
- 클라이언트는 제네릭 `[C comparable]` (keyRecords 맵 키로 써야 해서 comparable 필요)
- marshal/send 전략을 `NewManager`에서 1회 주입 → 호출부 `Publish(key, payload)` 한 줄
- Publish 에러 처리: 첫 실패에서 멈추지 않고 계속 전송 후 `errors.Join`으로 모아 반환
- 위치: 재사용 코어라 `internal/subscribe` (supervisor 밑 아님)

#### 작업 완료
- `internal/subscribe/manager.go` 생성 (제네릭 Manager + subscriber, 죽은 주석 전부 제거)
  - `Subscribe/Unsubscribe/UnsubscribeAll/Subscribers/Publish`
  - 락 최적화 유지: Subscribers가 스냅샷 반환 → 전송은 락 밖
- `internal/supervisor/subscribe/doc.go`: 도메인 전용 계층임을 명시 (코어 사용)
- worker main.go에 임시 테스트(Test1→[]byte{1}, Test2→[]byte{2}, 구독/Publish/Unsubscribe) 작성 → 검증 후 제거
- `go build ./...` OK

#### 다음
- supervisor 도메인 구독 계층: 키 구조체(`Key() string`) + `*transport.Client` 확정 + 전략 주입
- 단, `transport.Client`가 아직 스텁 → transport 구현과 함께 연결
