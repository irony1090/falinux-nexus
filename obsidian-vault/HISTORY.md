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

---

### 2026-06-18 - 요청/응답 상관관계 프레임 + 등록(REGISTER) 핸드셰이크

> 상태: **구현 동작 확인 완료. 사용자가 직접 리팩토링 예정** (현 코드를 베이스로).

#### 배경 / 문제의식
- 구 PortBridge `agentRouter.go`(test-jig) 검토: 요청·응답이 **두 파일 switch에 흩어져** 흐름 추적 곤란. `Packet{Type,SessionId,Payload}`에 짝지을 ID가 없어 도메인 매니저(`GetInteractive(payload.Id)`)로 역추적. 요청/응답/스트리밍 3성격이 한 평면에 뒤섞임
- → "요청↔응답을 한 묶음으로 추적 가능한, 도메인 자유로운(재사용) 프레임"이 필요하다는 결론

#### 설계 결정
- 핵심 통찰: 요청↔응답을 묶는 데 필요한 건 "ID → 기다리는 채널" 맵 하나. 나머지(데이터 모양/인코딩/소켓)는 전부 주입 → subscribe.Manager와 같은 철학
- 3층 분담: **correlator(엔진, 고정) / protocol(어휘, 채움) / conn(둘을 묶어 동작 얹음)**
- 제어평면(REQ/RES 1:1) vs 데이터평면(EVENT 스트림 1:N) 분리 원칙 (단 EVENT는 예제 최소화로 일단 제거)

#### 작업 완료 — 프레임
- `internal/call/correlator.go` — 제네릭 `Correlator[R]`: nextID 채번 + `waiters[id]→chan` 장부 + 동시성. `Call(ctx,payload)`(블록)/`Resolve(id,val,err)`/`Close(err)`(전부 깨움). send 전략 주입. **uint64 nextID는 wrap 안전 → 별도 처리 안 함으로 결정**
- `internal/protocol/protocol.go` — `Frame{Kind,ID,Type,Err,Data}` 봉투. Kind=REQ/RES. Data=json.RawMessage(자유). `NewRequest/NewResponse/NewError/Encode/Decode/Bind`
- `internal/transport/conn.go` — `Conn`: `Call`/`Handle`/`Serve`. `MessageRW` 인터페이스(gorilla `*websocket.Conn`이 그대로 만족). 끊김 시 `corr.Close`로 매달린 Call 일괄 정리
- (한때 만들었다 제거: `transport/pipe.go` 인메모리, EVENT On/Emit, `*_test.go` 2개 — 예제 최소화 요청. git 히스토리에 있음)

#### 작업 완료 — 등록 핸드셰이크
- `internal/protocol/messages.go` — `MsgRegister` + `RegisterRequest{Key,SubKey}` / `RegisterResponse{SubKey}`
- `cmd/supervisor/main.go` — WS 서버 + `registry{active map[string]bool, seq uint64}`(메모리). `register(req,&owned)`: 최초→`고유키#seq` 부여 / 재접속인데 active면 차단 / 점유키를 같은 락 안에서 `owned`에 기록. `release(&owned)`로 종료 시 해제. `Handle(MsgRegister)`
- `cmd/worker/main.go` — WS 클라. `loadSubKey/saveSubKey`(파일 영속), `-key`/`-store` 플래그. `Call(MsgRegister)` → 최초면 받은 서브키 저장
- **검증된 4시나리오**: ①최초→`worker-A#1` 부여·저장 ②재부팅 후 저장된 서브키로 재접속 인식 ③동일 서브키 동시접속 차단 ④같은 고유키 다른 인스턴스→`worker-A#2`로 구분

#### 트러블슈팅 메모
- 테스트 중 옛 ADD 버전 supervisor 프로세스(`go run`)가 8080 점유 → worker가 옛 supervisor에 붙어 "핸들러 없음: REGISTER" 발생. 잔류 프로세스 kill 후 정상. **포트 점유 잔류 프로세스 주의**
- `/tmp`에서 `go build` 시 `go.mod` 못 찾음 → 빌드는 모듈 디렉토리(`apps/core`)에서, 실행만 다른 곳에서

#### 다음 (사용자 리팩토링 후)
- 영속화: worker 서브키 파일→SQLite settings, supervisor registry 메모리→PG
- 서브키 생성/충돌(재시작 seq 리셋), key↔subkey 결속 검증(위조 방지)
- EVENT(스트리밍) 재도입, 다음 메시지(STATUS/EXEC 등) 확장

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

#### supervisor 상태 (중요)
- supervisor는 현재 전부 **임시 로직** — worker 흐름 굴리기용 최소 골격. **본격 작업은 "worker 파일 전송 모듈" 작업 시작 시 진입** → 그때 registry/구독/PG영속 제대로 구현. (파일전송 = PortBridge UploadInit→Ready→Chunk→Status→Result 계승)
