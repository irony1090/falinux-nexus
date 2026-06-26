# HISTORY — supervisor web/HTTP 계층 (store·tx·error·user)

> supervisor의 PG 스토어 배선, 트랜잭션 미들웨어, 에러 처리 규약, user/세션 핸들러 관련 상세 기록.
> 요약·재사용 지식은 `REF-supervisor-web.md` / `REF-db.md`, 현재 상황은 `CURRENT.md`.

---

### 2026-06-26 - 에러 처리 panic→return 전환 (echo HTTPErrorHandler, e2e 재검증)

> 사용자 지적: `web.Panic`을 쓰면 컴파일러가 "정말 panic하는지" 몰라 경고가 나지 않냐. 맞음 → 핸들러+헬퍼를 return-style로 전환.
> ※ 이후 2026-06-26 같은 날 사용자가 **panic-style로 다시 번복**(아래 별도 기록). 이 항목은 중간 단계 기록.

#### 근거 (Go 종료문 분석)
- Go 컴파일러/분석기는 **빌트인 `panic(...)`/`return`/무한 for{}** 등 문법적으로 정해진 "종료문"만 "이후 실행 안 됨"으로 인식. `web.Panic()`은 래퍼 함수라 종료문이 아님 → ① 값 반환 분기의 유일 탈출구면 `missing return` ② gopls nilness가 `getSessionOrPanic` 뒤 `return sess`를 "nil 반환 가능"으로 오판 → 호출부 `sess.Data`에 nil 역참조 경고
- 선택지: `panic(web.Err(...))`(빌트인이라 경고 해소되나 비관용·untyped·recover 필요) vs **`return web.Err(...)`(타입안전·echo 정석)**. 코드베이스가 이미 `QueriesPanic`/`web.Panic`으로 panic-leaning이라 헬퍼까지 묶어 전환해야 일관 → 사용자가 return 채택
- 트레이드오프 정리: panic=깊은 헬퍼서 자동 전파(실 꿰기 불필요)·untyped·recover 필수 / return=타입안전·관용·계층마다 `if err!=nil return err`. 핸들러가 얕아 후자 비용 작음

#### 전환 (4파일)
- `internal/web/util.go`: `ClientError{Status,Msg}`+`Error()`(error 구현). `Err(status,format,...)`. **`Panic`·`HttpProcess` 삭제**. 신규 `HTTPErrorHandler(err,c)`=ClientError→상태+메시지/echo.HTTPError→코드/그외→500, type=SERVER(>=500)·CLIENT. `PanicMiddleware`=안전망으로 단순화(recover→`err=rec`(error면)/`Err(500)` → echo가 HTTPErrorHandler로). LogMiddelware 유지
- `supervisorRouter.go`: `e.HTTPErrorHandler = web.HTTPErrorHandler`. PanicMiddleware는 안전망으로 유지(예기치 못한 panic + txMiddleware 재-panic 환원)
- `tx.go`: `ensureTx()`/`Tx(c)`/`TxQueries(c)` 전부 `(…, error)` 반환. Begin 실패=`web.Err(500)` 반환(구 web.Panic), Queries는 `t.Queries()`(비-panic, 구 QueriesPanic). 미들웨어 defer는 그대로(error 반환·panic 둘 다 롤백)
- `user.go`: 4핸들러 `return web.Err(...)`. `getSessionOrPanic`→`requireSession(c) (*SessionElement, error)`. 검증실패도 `web.Err(400,"%v",err)`

#### e2e 재검증 (go run + postgres15 + curl)
- 9시나리오 통과(동작 불변, 경고만 소멸): 가입200/중복400/검증실패400(validator 메시지)/미로그인401/로그인200/세션확인200(createdAt=0 동일)/오답400/로그아웃200/로그아웃후401
- 트러블: scratchpad 바이너리 실행·`pkill`이 샌드박스에서 exit144로 죽음 → `go run` 백그라운드(run_in_background)로 띄우고 `ss`로 pid 찾아 `kill`(dangerouslyDisableSandbox)로 종료

---

### 2026-06-26 - 트랜잭션 미들웨어 + user 가입/로그인 핸들러 (구현 완료, e2e 검증)

> 구 RequestScope를 echo 미들웨어로 대체하고, 그 위에 가입/로그인/세션 핸들러를 echo 기본 레이아웃으로 이식. 실 서버+postgres로 8시나리오 e2e 통과.

#### 트랜잭션 미들웨어 (`cmd/supervisor/router/tx.go`)
- 구 `web/requestContext.go`(전역 `map[*http.Request]any`+mutex+RequestScope 인터페이스+도메인 *Context 4종=transaction+Err 복붙)의 실체=요청 경계 커밋/롤백 → echo.Context(c.Set/c.Get)로 전역 map 폐기, 4 Context→단일 `txScope`
- `txScope{pool,tx,err}`(요청당 1개=락 불필요) + `ensureTx()`(lazy Begin, 실패 web.Panic500) + `release()`(err면 Rollback/아니면 Commit/tx==nil no-op, 실패는 로그만)
- `txMiddleware(pool)`: scope→c.Set→defer{recover면 err+release+재panic / 정상이면 err=핸들러반환+release}. **panic·error 둘 다 롤백**(코드베이스가 web.Panic으로 에러 던짐). 등록=PanicMiddleware 안쪽(재panic 전파)
- 핸들러 접근: `Tx(c)`/`TxQueries(c)`(=`GetRequestScopePanic[Ctx](req).Transaction()` 자리)

#### user 핸들러 (`cmd/supervisor/router/user.go`)
- 구 cmd/agent_v2 user 핸들러 이식하되 **web.HttpProcess 폐기, echo 기본 핸들러**(func(c)error). 에러=web.Panic→PanicMiddleware, tx=TxQueries(c)
- **별도 User 타입 불필요 발견**: 세션 모듈이 `superdb.User`의 export 기본형 필드(ID/Identification/Password/Nickname)를 직접 직렬화 → `session.SessionManager[superdb.User]` 그대로 사용. 키 `"irony"/"sid"`(구 동일), nameFn=nil. supervisorRouter에 `sessions` 필드 + `mountUsers`
- 로직: createUser(같은 tx서 중복검사+INSERT) / signIn(해시 대조→sess.Data=user→Save) / checkSession(getSessionOrPanic) / signOut(MaxAge=-1)
- `getSessionOrPanic`: 구 flag if/else 4단 → `sess==nil||IsNew||Data.ID==0` 단락 OR + 401. 구체 User 타입 고정
- 식별자·비번 모두 sha256 hex 저장/조회(`toHash`, 구 동일). 계정없음/비번불일치 동일 메시지(존재여부 비노출)

#### e2e 검증 (실 서버 + postgres15 + curl 쿠키jar)
- 8시나리오 전부 기대대로: 가입200 / 중복400 / 미로그인401 / 로그인200 / 세션확인200 / 오답비번400 / 로그아웃200 / 로그아웃후401
- **관찰 1**: checkSession 응답 `createdAt:0` — 세션 복원 user엔 타임스탬프 없음(`pgtype.Timestamptz`는 `isGobSerializable` 대상 아님 → 세션에 안 실림. 구 SQLite는 CreatedAt이 int64라 실렸음). 필요 시 `sess.Data.ID`로 DB 재조회
- **관찰 2**: 응답 identification=해시(구 설계 그대로), Password 해시가 세션 쿠키에 적재(CookieStore 암호화)
- 트러블: `go run` 백그라운드가 subshell에서 종료(exit 144) → `nohup`+run_in_background로 해결. postgres15는 사용자가 supervisor만 닫아서 계속 떠 있었음

---

### 2026-06-26 - supervisor PG 스토어 이식 + DB 배선/마이그레이션 자동적용 (구현 완료, 검증됨)

> 구 PortBridge pgx connectManager를 supervisor로 이식하고, supervisor main에 DB 연결+goose 마이그레이션을 worker 패턴으로 배선. 실제 nexus DB에 users 마이그레이션 적용까지 확인.

#### 발단 — RequestScope 제거 논의에서 트랜잭션 발견
- user 모듈 핸들러 작업 준비 중 구 `GetSessionOrPanic`/`GetRequestScopePanic` 이식 검토. `GetRequestScopePanic`은 전역 `map[*http.Request]any`+mutex라 "echo 쓰면 통째로 불필요(c.Get/Set이 대체)"로 결론냈으나, 사용자가 **"Release()에 DB commit이 걸려 있어 못 뺀다"** 지적 → `user.model.go` 확인 결과 `*Context.Release()`=요청 경계 커밋/롤백. 결론: 전역 map만 버리고 **트랜잭션 수명은 echo 미들웨어로 이전**. 그 전제로 supervisor에 트랜잭션 계층(pgx StorePool)이 먼저 필요 → 이번 작업

#### pgx connectManager 이식 (`internal/supervisor/store/connectManager.go`)
- 구 `internal/store/connectManager.go`(pgx) 이식. worker(SQLite)판과 **API 대칭**(InitStorePool/GetStorePool/Queries/Transaction + Begin/Commit/Rollback/Queries/QueriesPanic/AddAfterRelease). 차이=타입(`*pgxpool.Pool`/`pgx.Tx`)·시그니처(`user,pass,host,name,port`→DSN 조립+Ping)·gen(`superdb`)
- worker판이 이미 고친 2건 동일 반영: `afterUnsfae`→`afterUnsafe`+결과 err 전달(구 버그=store.err 넘김) / InitStorePool 재호출 시 실패연결 정리 후 재시도
- `pgxpool`이 `puddle/v2`를 indirect로 끌어옴(go mod tidy)

#### 마이그레이션 (`store/migrate.go` + `db/migrations/embed.go`)
- **핵심 난점**: goose는 `database/sql(*sql.DB)` 요구인데 풀은 pgxpool → `*sql.DB` 직접 안 나옴. 해결: `StorePool`에 `dsn` 필드 추가(InitStorePool서 저장) → Migrate가 `sql.Open("pgx", dsn)`로 **마이그레이션 1회용 throwaway `*sql.DB`** 열고 `defer Close`(풀과 완전 무관). `_ "github.com/jackc/pgx/v5/stdlib"`로 "pgx" 드라이버 등록, dialect=`postgres`
- (대안 `stdlib.OpenDBFromPool`은 Close 시 풀 영향 불확실해서 회피 — 독립 throwaway가 예측가능)
- `embed.go`=`//go:embed *.sql`(worker판 대칭, 배포 바이너리 단독 적용)

#### env + main 배선
- `cmd/supervisor/constants/env.go`: `EnvVars`에 DBUser/Pass/Name/Host(string)+DBPort(int). LoadEnv에 기존 패턴대로(필수 문자열=workerPath식 빈값에러, DBPort=port식 기본5432) 로드
- `cmd/supervisor/main.go init()`: `mountStore(GetEnv())`=InitStorePool(user,pass,host,name,port)→Migrate, 둘 다 실패 Fatalf. worker `mountStore` 패턴 그대로

#### 로컬 DB 준비 + 검증
- supervisor PG = docker `postgres15`(이미지 postgres:15). `docker start postgres15`(※ `docker start postgres`=그런 컨테이너 없음). user/pass/포트 일치, 기본 DB는 `workspace`라 **`nexus` DB 수동 생성**(`CREATE DATABASE nexus OWNER irony`)
- `go run ./cmd/supervisor`로 검증: 로그 `OK 00001_users.sql`→`migrated to version: 1`→`DB 준비됨`. `\dt`로 users+goose_db_version 확인. (서버는 잔류 프로세스가 5050 점유해 bind 실패했으나 마이그레이션은 init서 완료 — 사용자가 본인 프로세스라 종료)
