# history/process-subscription — 세션→uid 원장 + REST 구독/해지 배선

> 설계·재사용 지식 → `REF-process-subscription.md` / 재접속 모델 → `REF-process-reconnect.md` / 계약·원칙 → `REF-process.md` / 현재 진행 → `CURRENT.md`
> supervisor 아키텍처·worker 실행부 배선 이력은 → `history/process-wiring.md` / frontend 트리거·버그수정 이력은 → `history/process-trigger.md`

## 2026-07-16 — REST 구독/해지(GET/POST/DELETE /processes/*) + browsers(conn→sid) registry 배선 완료

세션→uid 원장(2026-07-14 (3) 설계, 이후 2026-07-15에 마이그레이션/query/ProcessManager 메서드까지 구현됨)의 마지막 조각 — **실제 호출 지점**을 배선. 대화로 설계 후 사용자가 직접 구현.

**설계 대화 흐름**
1. 재접속 시 구독 목록을 프론트에 전달할 채널을 REST vs 소켓 push 중 결정 — 프론트 `on()`이 버퍼링 없는 구조라 소켓 push는 "연결 직후 프론트가 핸들러 등록하기 전에 도착해 유실"되는 레이스가 있음을 근거로 **REST 채택**.
2. `mountProcesses`(node.go/user.go와 동일 관례) 위치를 `process.go`(worker-wire 전용, 이미 밀도 높음) 대신 **신규 `processApi.go`로 분리** — node.go/nodeDto.go의 "핸들러 vs DTO" 분리와 같은 원리로 "wire orchestration vs REST API" 축 분리.
3. `subscribeProcess`/`unSubscribeProcess`(REST) 구현 중 "이미 열려있는 소켓에 즉시 반영이 안 됨" gap 발견 → sid로 살아있는 conn을 찾을 registry 필요성 도출.
4. `KeyValManager.Append`가 key당 값 1개만 허용함을 코드로 확인(같은 sid 여러 탭이면 충돌) → **conn을 key로 쓰는 역방향 registry**(`browsers *manager.KeyValManager[*transport.Conn, string]`)로 설계, "이 sid의 모든 conn" 조회는 `FindAll`(기존 `workers.FindAll`/`ListInstances(main_key)`와 동일 패턴)로 해결.
5. 두 대안 검토 후 기각: ① `subscribeHub` 내부에 sid 인식 심기(Hub의 도메인-무관 설계 원칙 위배 + `keyRecords`가 "이미 구독 중인 client만" 추적해 애초에 질문 방향이 안 맞음), ② `sess.Name()`(sid)에 port 붙여 유일화(재접속마다 port가 바뀌어 DB PK로 쓰는 sid의 "재접속 전후 불변" 특성이 깨짐 → 복원 기능 붕괴).
6. Hub 안에 browsers를 nesting하는 안도 검토 후 기각(같은 도메인-무관 원칙) — 대신 router 레벨 `browsersForSid` glue 헬퍼로 결합도 낮춤.

**코드 변경** (사용자 직접 작업, 대화로 검토·버그 1건 발견)
- `cmd/supervisor/router/processApi.go`(신규) — `mountProcesses`, `listSubscriptions`(GET), `subscribeProcess`(POST), `unSubscribeProcess`(DELETE).
- `cmd/supervisor/router/supervisorRouter.go` — `browsers` 필드 추가 + `browsersForSid` 헬퍼, `mountProcesses(e)` 등록.
- `cmd/supervisor/router/subscribe.go` — handshake에서 `browsers.Append`/`Remove`, 죽은 `processSubscribeKey`("PROCESS:" prefix, 실제 발행 prefix "PROC:"와 불일치하던 버그) 주석처리 후 기존 `processTopic` 재사용으로 정리.
- `cmd/supervisor/router/process.go` — `processTopic` 파라미터명 `uid`→`processId`(REST 라우트 파라미터명과 통일).
- **버그 발견·수정**: `browsers` 필드가 생성자에서 초기화 안 돼 nil → `/subscribe` 첫 연결에서 `KeyValManager.Append`가 nil 포인터 역참조로 panic. `manager.NewKeyValManager[*transport.Conn, string]()` 추가해 해소.

`go build`/`go vet` 클린 확인. 상세 설계 근거 → `REF-process-subscription.md` "REST 구독/해지 배선" 절.

**남은 것**: 프론트 쪽 전부(REST 클라이언트 함수·UI 트리거·`on()` 수신 핸들러) 미착수. `NODE:<parentId>` 동적 구독은 별개(REST로 갈지 미정). exec/input/kill 트리거도 미착수.

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

상세 설계(DDL·쿼리·트리거 근거) → `REF-process-subscription.md` "세션→uid 원장 구체화" 절.

**남은 것 (다음 세션)**
- 마이그레이션(`00004_process_subscribers.sql`) + query 파일 생성 + `sqlc generate`.
- CREATE 실제 호출 지점(동적 구독 배선과 합류) / DELETE 실제 호출 지점(UNSUBSCRIBE 프로토콜 신설 여부).
- `getSessionKey` nil-cookie panic 가드.
- sid 원본 vs 해시 저장 여부 최종 결정.

