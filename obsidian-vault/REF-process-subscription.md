# REF — process 세션→uid 원장 + REST 구독/해지 배선

> 재접속 모델은 `REF-process-reconnect.md`. 배선 아키텍처는 `REF-process-wiring.md` / frontend 트리거는 `REF-process-trigger.md`. 계약/설계 원칙은 `REF-process.md`. 작업 이력은 `history/process-subscription.md`.

## 세션 → uid 원장 구체화: process_subscribers 테이블 + sid 추출 (2026-07-14, 설계 진행 중 — 미구현)
> `REF-process-reconnect.md` "종료/재접속 모델"의 "세션→보던 uid 원장" 항목의 구체 설계. **아직 마이그레이션/쿼리 파일 미생성** — 스키마·쿼리·트리거 시점은 합의됐고, CREATE/DELETE 실제 호출 지점 배선만 남음.

**sid(브라우저 세션) 추출 배선 — 코드 완료**
- 문제의식: 기존엔 `owner_user_id`(UserID)만 있어 **같은 유저가 다른 브라우저로 로그인해도 구분이 안 됨**. "같은 브라우저=같음, 다른 브라우저(같은 유저라도)=다름"을 만족하는 키가 필요.
- `sid` 쿠키(gorilla `CookieStore`)는 **로그인(`Save()`) 시점의 timestamp가 MAC 서명에 포함**되어 로그인마다 다른 문자열이 되고, `requireSession`/새로고침은 `Save()`를 다시 안 불러 **같은 브라우저 내에서는 로그아웃 전까지 고정** — 정확히 필요한 특성. `Save()` 호출부는 `signIn`/`signOut`(`user.go`) 딱 두 곳뿐(grep으로 확인).
- `internal/manager/session/sessionManager.go`: `NameFunc[T]` 시그니처에 `req *http.Request` 추가(`func(data T, session *sessions.Session, req *http.Request, key string) string`) → `Name()`이 `req`까지 넘겨줌. (→ `REF-supervisor-web.md` "세션 매니저" 절도 갱신됨)
- `cmd/supervisor/router/supervisorRouter.go`: `getSessionKey` nameFn 구현 배선 — `req.Cookie(key)`로 원본 `sid` 쿠키 값 추출해 `NewSessionManager("irony","sid", getSessionKey)`로 주입.
- **⚠️ 미해결 버그**: `getSessionKey`가 `c, _ := req.Cookie(key)`로 에러를 버려서, 쿠키가 없는 요청에서 `.Name()`을 호출하면 `c`가 `nil`이라 **nil pointer panic**(PanicMiddleware가 잡아 500으로 렌더는 됨, 그러나 실제 요청은 실패). `err != nil` 가드 필요.
- 쿠키에 `Password`(해시) 필드가 함께 직렬화되어 있음(`REF-supervisor-web.md` 기존 기록) — `sid` 원본 문자열을 그대로 DB 컬럼/로그에 남기지 말고 **해시(sha256 등)해서 저장 권장**(아직 미적용, 아래 DDL·쿼리는 raw 저장으로 초안 작성됨 — 실제 적용 시 재검토).

**process_subscribers 테이블 설계 (DDL 초안, 마이그레이션 미생성)**
```sql
CREATE TABLE process_subscribers (
    process_uid   TEXT        NOT NULL REFERENCES processes(uid) ON DELETE CASCADE,
    owner_user_id BIGINT      NOT NULL REFERENCES users(id),
    sid           TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (process_uid, sid)
);
CREATE INDEX idx_process_subscribers_sid ON process_subscribers(sid);
```
- **PK=(process_uid, sid)** 자연키 — 같은 브라우저가 같은 process를 중복 구독해도 1행만 유지(재구독은 `ON CONFLICT DO NOTHING`으로 무시).
- **`process_uid` FK `ON DELETE CASCADE`** — process 자체가 삭제되면 구독 정보도 자동 정리(고아 row 방지 안전망. 아래 DELETE 트리거 정책과는 별개, 병행).
- `owner_user_id`는 FK 걸어 조회/감사 용도로 유지(스키마상 sid만으로도 유저 특정 가능하나 컬럼 분리 유지).

**쿼리 3종 (`internal/supervisor/db/query/processSubscribers.sql`, 미생성)**
```sql
-- name: CreateProcessSubscriber :exec
INSERT INTO process_subscribers (process_uid, owner_user_id, sid)
VALUES ($1, $2, $3)
ON CONFLICT (process_uid, sid) DO NOTHING;

-- name: DeleteProcessSubscriber :exec
DELETE FROM process_subscribers WHERE process_uid = $1 AND sid = $2;

-- name: ListProcessesBySid :many
SELECT p.* FROM processes p
JOIN process_subscribers ps ON ps.process_uid = p.uid
WHERE ps.sid = $1
ORDER BY ps.created_at DESC;
```
- CREATE/DELETE만으론 "재접속 시 복원"이 실제로 동작 안 함(조회 수단이 없어서) → `ListProcessesBySid`(sid로 process 목록 join) 추가. 재접속 시 프론트가 부를 엔드포인트가 이걸 사용하게 됨.

**CREATE/DELETE 트리거 시점 — 2안(요청/해지마다 즉시 반영) 채택**
- 비교한 두 안: ① 프론트 소켓이 끊어질 때 그 순간의 구독 목록을 스냅샷으로 일괄 write / ② 구독 요청·해지마다 그때그때 메모리+DB 반영.
- DELETE는 이미 "프론트가 명시적으로 구독 해제 요청할 때만"으로 정해져 있음(①처럼 소켓 끊김에 걸면 새로고침도 소켓이 끊기는 거라 **매 새로고침마다 구독 정보가 삭제**돼 복원 기능 자체가 성립 안 함 — 새로고침과 완전 종료를 서버가 구분 못 하기 때문).
- **② 채택 이유**: DELETE와 트리거 지점이 대칭(둘 다 명시적 프로토콜 이벤트, 소켓 라이프사이클에 안 얹음) + **supervisor 프로세스 자체가 재시작해도 안전**(①은 close 핸들러가 못 돌면 그 세션의 구독 정보가 DB에 아예 안 남음 — supervisor 재시작 복구 시나리오와 충돌).

**남은 것 (미착수)**
- ~~마이그레이션·query 파일·CREATE/DELETE 호출 지점~~ → 전부 완료(아래 두 절). 이제 남은 건 **프론트 쪽**(REST 클라이언트 호출·UI 트리거)과 **NODE:<parentId> 동적 구독**(아래 참조).
- `getSessionKey`의 nil-cookie panic 가드: **작업트리에 이미 적용됨**(`err != nil` → `""` 반환, 커밋 전). 위 "미해결 버그" 서술은 커밋 시점 기준 stale.
- `sid` 저장 방식: **원본 그대로 저장 확정**(2026-07-15, 해시 대안 기각 — 단순 구현 우선).

## REST 구독/해지 배선 + browsers(conn→sid) registry — 구현 완료 (2026-07-16)

> 위 "CREATE/DELETE 실제 호출 지점"의 구체 구현. `MsgSubscribe`/`MsgUnsubscribe` 소켓 메시지 대신 **REST로 확정**(process 도메인 한정 — NODE 쪽 동적 구독은 별개, 아래 "남은 것" 참조).

- **REST 채택 이유**: 프론트 `on()`이 버퍼링 없는 구조(`websocket.hook.ts`, 미등록 타입은 조용히 버림)라, 소켓 연결 직후 서버가 바로 push하면 프론트가 핸들러 등록 전에 메시지가 도착해 유실되는 레이스가 구조적으로 있음. REST는 프론트가 원하는 타이밍에 직접 호출해서 이 레이스가 없음 + 이 레포의 기존 관례("REST=조회/CRUD, socket=push 전용", `REF-node-label.md`)와도 일치.
- **`GET /processes/subscriptions`**(`listSubscriptions`) — `ListSubscriptions(sid)` 그대로 감싸기. 재접속 시 프론트가 이걸 불러 화면 복원.
- **`POST/DELETE /processes/subscribe/:processId`**(`subscribeProcess`/`unSubscribeProcess`) — `cmd/supervisor/router/processApi.go`(신규 파일. `process.go`는 worker-wire 오케스트레이션 전용으로 유지, REST는 분리 — node 도메인의 `node.go`/`nodeDto.go` 분리 관례와 같은 원리).
  - `SubscribeProcess`/`UnsubscribeProcess`(memory+DB) 호출 → entry 없으면(uid 미존재) 404.
  - **곧바로 `browsersForSid(sid)`로 그 sid의 살아있는 conn 전부를 찾아 `subscribeHub.Subscribe`/`Unsubscribe(processTopic(uid), conn)`** — REST 호출만으론 DB/memory 장부만 바뀌고 이미 열려있는 소켓의 실시간 라우팅엔 반영이 안 되는 gap을 메움(발견 2026-07-16, 즉시 해소).
- **`browsers *manager.KeyValManager[*transport.Conn, string]`(conn→sid, 역방향) 신설** — `supervisorRouter.go`, `workers`/`readers` 옆.
  - **왜 역방향인가**: `KeyValManager.Append`는 key당 값 1개만 허용(이미 있으면 조용히 실패, 안 덮어씀) — sid를 key로 쓰면 같은 브라우저의 두 번째 탭(같은 sid, 다른 conn)이 등록 실패함. conn 포인터는 연결마다 유일해 key로 쓰면 이 문제가 없음.
  - "이 sid의 모든 conn" 조회는 `browsersForSid(sid)`(`supervisorRouter.go`) 헬퍼가 `browsers.FindAll`로 역조회 — `workers.FindAll`로 `ListInstances(main_key)`가 main#sub 중 main_key만 걸러내는 것(`REF-node-label.md`)과 완전히 같은 패턴(굵은 키로 가는 키 여러 개를 스캔).
  - 등록/해제는 `handleSubscribeWS`(`subscribe.go`)에서 conn 생성 직후 `Append` / `Serve()` 리턴 후 `Remove`.
- **기각된 대안 두 개**(둘 다 이 세션에서 검토):
  - **`subscribeHub` 내부에 sid 인식 심기** — 기각. Hub는 "키 생성 로직과 도메인 타입을 모른다"는 게 스스로 못 박은 설계 원칙(`internal/subscribe/hub.go` 상단 주석)이고, 결정적으로 Hub의 `keyRecords`는 **이미 뭔가 구독 중인 client만** 추적함(구독 0개 되면 그 client 자체가 지워짐) — "아직 구독 안 한 topic에 대해 sid로 conn 찾기"라는, 우리가 필요한 질문 방향 자체를 Hub가 못 풂.
  - **`sess.Name()`(sid)에 포트 번호 등 붙여서 유일화** — 기각. `sid`는 `process_subscribers` DB PK에 쓰이는 **영속 식별자**라 재접속 전후로 절대 안 바뀌어야 하는데(로그인 시점 서명이라 새로고침엔 불변), 포트는 TCP 연결마다 새로 배정돼 재접속마다 값이 바뀜 — 섞으면 재접속 시 `ListSubscriptions`가 예전 sid를 못 찾아 복원 기능 자체가 깨짐.
- **버그 1건 발견·수정**: `browsers` 필드를 `supervisorRouter` 생성자에서 초기화 안 해 nil이었음(`workers`/`readers`는 초기화하면서 누락) → `/subscribe` 첫 연결에서 `KeyValManager.Append`가 nil 포인터 역참조로 panic. `manager.NewKeyValManager[*transport.Conn, string]()` 추가해 해소. `go build`/`go vet` 클린 확인.
- **남은 것**: 프론트 쪽 전부(REST 클라이언트 호출 함수, subscribe/unsubscribe UI 트리거, `on('NODE:CREATE'|...)` 등 수신 핸들러) 미착수. `NODE:<parentId>` 쪽 동적 구독/해지는 이번 스코프 밖 — 지금도 `NODE:0` 고정 구독만(`REF-realtime.md` 참조, REST로 갈지 소켓 메시지로 갈지도 미정).

**`ProcessEntry` 인메모리 구독자 필드 — 구현 완료 (2026-07-15)**
- `ProcessEntry`(`cmd/supervisor/process/entry.go`)에 구독자 목록 필드 추가. DB row(`process_subscribers`) 그대로 미러링 **안 함** — `Subscriber{Sid, OwnerUserID}`만 담는 단순 구조.
- **이유**: `process_uid`는 엔트리 자체가 그 uid로 관리돼 중복(엔트리가 곧 process_uid 스코프) / `created_at`은 DB 감사·원장 목적이지 런타임 라우팅엔 불필요 — "DB≠라우팅 authority"(MEMORY 확정 결정, 실시간 push 절)와 동일한 원칙: 메모리 객체는 DB row 미러가 아니라 그 순간 필요한 도메인 필드만 갖는 lean 구조.
- API: `IsSubscribed(sid) bool` / `AddSubscriber(Subscriber)`(같은 sid 중복 무시, DB `ON CONFLICT DO NOTHING`과 동일 시맨틱) / `RemoveSubscriber(sid)`. 엔트리별 `sync.RWMutex`로 슬라이스 보호(`KeyValManager`의 맵 mutex는 map 자체만 지키고 entry 내부 필드 동시접근은 안 막아줘서 별도 필요).
- `go build`/`go vet` 클린. **호출 지점 배선 완료**(2026-07-16, 위 "REST 구독/해지 배선" 절) — `SubscribeProcess`/`UnsubscribeProcess` 경유로 실사용됨.

**`ProcessManager.SubscribeProcess`/`UnsubscribeProcess` — memory+DB write-through 구현 완료 (2026-07-15)**
- 위치는 `cmd/supervisor/process/manager.go`(라우터/핸들러가 아니라 매니저) — `execScript`(memory Append→DB CreateProcess, 실패 시 memory 롤백)/`Rebind`/`Remove`가 이미 "memory+DB를 한 지점에서 같이 맞추는" 책임을 매니저가 지는 패턴이라 그대로 따름(사용자 확정).
- `SubscribeProcess(uid, Subscriber)`: **memory 먼저**(`entry.AddSubscriber`, 저렴·되돌리기 쉬움) → `CreateProcessSubscriber` DB insert, **실패 시 memory 롤백**(`entry.RemoveSubscriber`) — `execScript`와 동일 순서/롤백 대칭. entry 없으면(uid 존재 안 함) 에러 반환.
- `UnsubscribeProcess(uid, sid)`: entry 있으면 memory 먼저 제거, **DB delete는 entry 유무와 무관하게 항상 시도** — process가 이미 완료돼 `Remove()`로 memory에서 정리된 뒤라도 `process_subscribers` DB 원장은 남아있을 수 있어서(구독 해제는 프론트 명시 요청 시에만, 완료와 무관) 비대칭으로 뒀음. DB delete 실패 시 memory 롤백은 안 함(이미 지웠고 재추가할 정보 없음 — 다음 재구독 때 자연 회복).
- `go build`/`go vet` 클린. **호출 지점 배선 완료**(2026-07-16) — `MsgSubscribe`/`MsgUnsubscribe` 소켓 메시지가 아니라 `processApi.go`의 REST 핸들러가 호출(위 "REST 구독/해지 배선" 절, 채택 이유 포함).

**`ProcessManager.ListSubscriptions(sid)` — 조회는 memory 아니라 DB 직접 (2026-07-15)**
- `ListProcessesBySid` DB 쿼리를 그대로 감싸는 얇은 래퍼. **memory-first 아님** — memory엔 sid→uid 역인덱스가 없어 전체 entry 스캔이 필요한데 이 조회는 재접속/화면복원 시 1회성이라 인프라 투자 실익이 없고, memory는 휘발성(supervisor 재시작 시 소실)이라 "재시작 복구" 목적과도 안 맞음. `status` 컬럼이 write-through로 최신화되어 있어 DB 단독 조회로 충분(realtime push의 "라우팅=메모리/인가·조회=DB" 분리 원칙과 동일 결).
- `go build`/`go vet` 클린. **호출 지점 배선 완료**(2026-07-16) — `GET /processes/subscriptions`(`listSubscriptions`, 위 "REST 구독/해지 배선" 절). 프론트가 이 엔드포인트를 실제로 부르는 건 아직 미착수.
