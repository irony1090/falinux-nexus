# REF — supervisor web/HTTP 계층 (tx 미들웨어 / 에러 규약 / user·세션)

> echo 기반 supervisor HTTP 핸들러 규약. 상세 작업 이력은 `history/supervisor-web.md`.
> DB 스토어 계층은 `REF-db.md`.

## 트랜잭션 미들웨어 (`cmd/supervisor/router/tx.go`) — 구현 완료 (2026-06-26)
- 구 PortBridge `web/requestContext.go`(전역 map+RequestScope+도메인 *Context 4종) 대체. **전역 map 폐기**(echo.Context의 c.Set/c.Get이 요청 수명 저장소) + 4개 중복 Context → **단일 `txScope` + 미들웨어 1개**
- `txScope{pool,tx,err}`: 요청 1건 트랜잭션 수명. 요청마다 새로 만들어 echo.Context에 실림 → **단일 goroutine 전용, 락 불필요**
- **lazy Begin**: `ensureTx() (*Transaction, error)` 첫 호출 때만 `pool.Transaction()+Begin` → 읽기전용 핸들러는 트랜잭션 안 엶. 핸들러에서 `TxQueries(c)`로 사용
- `release()`(요청당 1회): err 있으면 Rollback, 없으면 Commit. tx==nil이면 no-op. 커밋/롤백 실패는 로그만(응답 이미 나간 뒤)
- **`txMiddleware(pool)`**: scope 생성→c.Set→defer{recover면 err기록+release+재panic / 정상이면 err=핸들러반환값+release}. **error 반환·panic 둘 다 롤백**(panic은 안전망)
- **등록 위치 중요**: `e.Use` 순서 = `PanicMiddleware → Log → txMiddleware → CORS`. tx는 **PanicMiddleware 안쪽**이라야 재-panic이 PanicMiddleware로 전파돼 JSON 응답됨. WS 업그레이드 라우트도 전역 래핑되나 Tx() 안 부르면 무해(no-op)
- **commit-in-defer는 c.JSON 이후 실행**(응답 먼저, 커밋 나중) — 구 HttpProcess와 동일 한계, 현재 허용

## ★에러 처리 규약: panic-style (PanicMiddleware) — 2026-06-26 재확정
- **결론 번복 이력**: 한때 return-style(echo HTTPErrorHandler)로 전환했으나 사용자가 "맘에 안 든다"고 **다시 panic-style**로 되돌림
- **현재 규약(panic-style)**:
  - `internal/web/util.go` = panic 기반. `ClientError{Status int, Error error}` = `Error`가 **필드**라 error 인터페이스 **미구현** / `Err(status, fmt, ...)`는 `ClientError` 반환 / `PanicMiddleware`가 recover→ClientError면 `{message,type}` JSON, 아니면 500. **`HTTPErrorHandler` 없음**
  - 핸들러는 `panic(web.Err(status, fmt, ...))`. tx 헬퍼 `ensureTx()/Tx(c)/TxQueries(c)`는 **error 반환 제거**(panic; TxQueries=`Tx(c).QueriesPanic()`). `requireSession`도 단일 반환+401 panic. DB/Save 에러도 `panic(web.Err(500,"%v",err))`로 통일. 성공만 `return c.JSON(...)`
  - txMiddleware는 불변(recover→롤백→재-panic→PanicMiddleware 렌더). 빌드/vet 통과
- **여전히 필요한 panic 인프라**: 진짜 버그(nil 역참조)·txMiddleware 재-panic 때문에 PanicMiddleware 유지. tx 롤백은 error 반환·panic 둘 다 처리
- ※ 폐기된 return-style 근거(Go 종료문 분석: `web.Panic` 래퍼는 종료문 아니라 `missing return`/gopls nilness 오경고)는 `history/supervisor-web.md`에 보존
- **향후 모든 supervisor echo 핸들러 이 규약 따를 것**

## user 가입/로그인 핸들러 (`cmd/supervisor/router/user.go`) — 구현+e2e 완료 (2026-06-26)
- **echo 기본 핸들러**(func(c)error, panic-style + `TxQueries(c)`). 라우트: POST `/users`(가입)·POST `/users/session`(로그인)·GET(세션확인)·DELETE(로그아웃). `supervisorRouter.mountUsers(e)`
- **★별도 User 타입 안 만듦**: 세션이 `superdb.User`의 export 기본형 필드 직접 직렬화 → `session.SessionManager[superdb.User]` 그대로(키 `"irony"/"sid"`, nameFn=nil). supervisorRouter에 `sessions` 필드
- `requireSession(c) (*SessionElement, error)`(구 getSessionOrPanic) = `sess==nil||IsNew||Data.ID==0` 단락 OR → `web.Err(401)`. `toHash`(sha256 hex)로 식별자·비번 저장/조회
- **함정**: checkSession은 세션 복원값이라 `pgtype.Timestamptz`(CreatedAt/UpdatedAt) **0/누락**(gob 직렬화 대상 아님). 타임스탬프 필요하면 `sess.Data.ID`로 DB 재조회. 응답 identification=해시. Password 해시가 세션 쿠키에 적재됨
- e2e 8~9시나리오 통과(가입/중복/미로그인/로그인/세션확인/오답/로그아웃/로그아웃후)
- **잔여**: 세션키 `"irony"` 하드코딩 → `SESSION_KEY` env화

## 세션 매니저 (`internal/manager/session/sessionManager.go`)
- 구 PortBridge 이식(`echo/v4`+`gorilla/sessions`). 제네릭 `SessionManager[T]`. 로직 버그 3건 수정(saveData 비공개필드 패닉 / T 비-struct 패닉 / GetSession 디코드실패 잠금) + `Name()`을 `NameFunc[T]` 주입 전략으로 전환
