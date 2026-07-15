# history/process-reconnect — 종료/재접속 모델 + 세션 복원

> 설계·재사용 지식 → `REF-process-reconnect.md` / 계약·원칙 → `REF-process.md` / 현재 진행 → `CURRENT.md`
> supervisor 아키텍처·worker 실행부 배선 이력은 → `history/process-wiring.md`

## 2026-07-14 (3) — 세션→uid 원장 설계 착수: sid 추출 배선(코드 완료) + process_subscribers 테이블 설계(미구현)

**코드 변경** (사용자 직접 작업, 대화로 검토)
- `internal/manager/session/sessionManager.go`: `NameFunc[T]`에 `req *http.Request` 파라미터 추가, `Name()`이 전달하도록 변경.
- `cmd/supervisor/router/supervisorRouter.go`: `getSessionKey` nameFn 구현 배선 — `req.Cookie(key)`로 원본 sid 값 추출해 `NewSessionManager`의 세 번째 인자로 주입(기존 `nil`).
- 확인된 미해결 버그(코드리뷰로 발견, 아직 미수정): `getSessionKey`가 쿠키를 못 찾을 때(`err != nil`) 가드 없이 `c.Value`에 접근 → nil pointer panic 가능(PanicMiddleware가 잡아 500으로는 렌더됨).

**설계 대화 흐름**
1. "UserID 제외, 새로고침에도 안 바뀌는 브라우저별 식별자가 있는가" 질문 → 세션 저장 구조(`gorilla/sessions.CookieStore`, `sid` 쿠키) 조사. 처음엔 `Identification`(로그인 아이디 해시) 후보로 나왔으나 UserID와 사실상 1:1이라 기각.
2. "다른 브라우저에서 같은 User로 로그인해도 같은 사람으로 묶이는 걸 피하고 싶다"로 요구 명확화 → `sid` 쿠키 **원본 문자열**이 로그인(`Save()`) 시점 timestamp 때문에 브라우저(로그인 인스턴스)마다 다르고, 같은 브라우저 내에선 로그아웃 전까지 고정됨을 확인(MAC 서명 구조 분석 + `Save()` 호출부가 `signIn`/`signOut` 두 곳뿐임을 grep으로 확인).
3. "`NewSessionManager` 3번째 인자(`nameFn`)로 sid를 뺄 수 있는가" 질문 → 당시 시그니처(`func(data T, session *sessions.Session, key string) string`)로는 `req`가 없어 불가 판정, 대안(직접 `c.Request().Cookie` 읽기 or `SessionElement.req` 노출)을 제시.
4. 사용자가 직접 `NameFunc`에 `req *http.Request` 추가 → 가능해짐 확인, 빌드/vet 통과 확인.
5. 사용자가 `supervisorRouter.go`에 `getSessionKey` 배선 완료 → 코드 리뷰 중 nil-cookie panic 버그 발견해 알림(아직 미수정).
6. 이어서 `process_subscribers`(userId, sid, createdAt) 릴레이션 테이블 설계 요청 — 기존 스키마/sqlc 컨벤션(마이그레이션+query+gen 3분리, `processes` 테이블 FK/CHECK 패턴) 조사 후 DDL·쿼리 3종 초안.
7. DELETE 트리거 시점을 먼저 확정(소켓 끊김 vs 명시적 구독해제) → **명시적 구독해제만**(새로고침=소켓끊김과 서버 입장에서 구분 불가라서, 소켓끊김에 걸면 복원 기능 자체가 무너짐).
8. CREATE 트리거 시점 — 1안(소켓 끊길 때 스냅샷) vs 2안(구독 요청/해지마다 즉시 반영) 비교 제시 → **2안 채택**(DELETE와 트리거 대칭 + supervisor 자체 재시작에도 데이터 유실 없음).

상세 설계(DDL·쿼리·트리거 근거) → `REF-process-reconnect.md` "세션→uid 원장 구체화" 절.

**남은 것 (다음 세션)**
- 마이그레이션(`00004_process_subscribers.sql`) + query 파일 생성 + `sqlc generate`.
- CREATE 실제 호출 지점(동적 구독 배선과 합류) / DELETE 실제 호출 지점(UNSUBSCRIBE 프로토콜 신설 여부).
- `getSessionKey` nil-cookie panic 가드.
- sid 원본 vs 해시 저장 여부 최종 결정.

## 2026-07-14 (2) — worker 끊김→PENDING→재접속 재바인딩 구현 완료 + e2e 검증

아래 "설계 확정" 절 그대로 구현. 상세 = `REF-process-reconnect.md` "worker 끊김 → PENDING → 재접속 재바인딩" 절(구현 목록 표 + `applyStatus` 가드 완화 + `WorkerState` 두 하위 절 포함).

**변경 파일**
- `internal/supervisor/db/query/processes.sql` — `MarkProcessPending` (+sqlc generate)
- `internal/protocol/messages.go` — `MsgSync`/`SyncEntry`/`SyncEvent`
- `cmd/supervisor/process/manager.go` — `ProcessManager.Rebind`
- `cmd/supervisor/router/process.go` — `applyStatus` 가드 완화+`CommandPending` 분기, `reconcileDisconnect`, `sync` 핸들러, `reconcileReconnect`
- `cmd/supervisor/router/supervisorRouter.go` — 끊김 시 `reconcileDisconnect` 호출 + `MsgSync` 핸들러 등록
- `cmd/worker/router/workerRouter.go` — `WorkerState`(재접속 넘어 사는 procs+conn) 신설
- `cmd/worker/router/process.go` — pump/teardown이 `state.currentConn()` 참조하도록 변경
- `cmd/worker/router/register.go` — 등록 성공 직후 `sendSync()`
- `cmd/worker/main.go` — `WorkerState`를 재접속 루프 밖에서 1회 생성해 전달

**구현 중 발견해 원 설계에 없던 것 두 가지** (상세는 REF 두 하위 절)
1. `applyStatus`가 memory entry 없으면 조기 리턴하던 가드가 "재접속·supervisor만 앎" 분기(entry가 끊김 처리 때 이미 Remove된 상태)를 막고 있어 DB가 Failed로 안 닫히는 공백 발견 → 가드를 case별로 완화(사용자에게 두 대안 제시 후 "가드 완화" 채택).
2. `cmd/worker/main.go`가 재접속마다 `NewWorkerRouter`를 새로 호출하며 `procs`(PTY 핸들 맵)를 매번 새로 만들던 기존 구조가 "재접속해도 procs가 산다"는 설계 전제와 충돌 발견(구현 도중 발견, 사용자에게 "지금 같이 고칠지" 확인 후 진행) → `WorkerState` 도입으로 해소.

**e2e 검증(수동, 실제 프로세스 kill/restart)**: 로컬 supervisor+worker 기동(postgres15 컨테이너, node id=2 스크립트를 `sleep N`으로 임시 교체) 후
- worker 프로세스 kill → supervisor 로그에 `E N D`, DB `PROCESS→PENDING` 확인.
- (PENDING 상태에서) 빈 procs로 새 worker 재접속 → DB `PENDING→FAILED(exit_code=502)` 확인("supervisor만 앎" 분기 = 가드 완화가 실제로 살리는 경로).
- worker는 살려둔 채 supervisor만 kill/재기동 → 재접속 시 DB가 `PROCESS`로 재차 갱신(`updated_at` 확인, Rebind+relay 재기동 성공) → 그 프로세스가 자연 종료(sleep 만료)할 때까지 대기해 **새 conn을 통해** `COMPLETED, exit_code=0`까지 정상 보고됨을 확인(교집합/Rebind 경로 + worker측 conn 스왑 둘 다 실증).
- 테스트 후 node content 원복, 테스트 프로세스 정리. `go build`/`go vet`/`gofmt -l .` 클린.

## 2026-07-14 (1) — worker 끊김→PENDING→재접속 재바인딩 설계 확정 (코드 변경 없음, 순수 설계)

CURRENT.md "다음 배선" 2번 착수를 위한 설계 대화. 결과는 `REF-process-reconnect.md` "worker 끊김 → PENDING → 재접속 재바인딩" 절에 반영(이 절이 최신, 구 "⚠️ 재접속 conn 재바인딩(난점)" 대체).

**핵심 결정**
- 재접속 시 죽은 conn을 캡처한 `AgentInteractive` 콜백을 "교체"하려던 기존 난제를 폐기 — 대신 끊기면 **버리고**(`Inter.Done(sentinel)`로 `bind.Relay` 고루틴 정지 후 `processManager.Remove`), 재접속하면 **새로 만든다**(`newWorkerInteractive` 재호출). 훨씬 단순.
- 3-way 재동기화는 memory가 아니라 **DB(`ListActiveByDevice`, 기존 존재하나 미사용이던 쿼리)를 진실로 삼음** — memory에 Record-only placeholder를 남겨둘 필요 없음이 드러남(사용자 질문으로 발견).
- **FOLDER 타입은 원천적으로 안전**: `openFolder()`가 애초에 `CreateProcess`(DB persist)를 안 부름 → `ListActiveByDevice`에 안 걸림 → 끊김 처리 루프가 자동으로 건드리지 않음(별도 필터 불필요). 근거는 `entry.go` `HasProcess()` 기존 주석("folder는 제외 — frontend 끊김 기준")과 일치.
- worker→sup 재접속 보고는 `RegisterRequest`에 얹지 않고 별도 `MsgSync{uid,status,pid}[]`로 분리(식별자 핸드셰이크와 도메인 상태 동기화는 관심사 분리).
- 3-way 대조 처리: 교집합=`Rebind`(신규 메서드)로 재장착, supervisor만 앎=`Failed`+sentinel로 종결(LOST 상태 신설 안 함), worker만 앎(고아)=로그만(YAGNI).

**구현·검증은 위 2026-07-14 (2) 참조.**
