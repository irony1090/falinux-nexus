# CURRENT

## 현재 날짜
2026-06-26

## 오늘(2026-06-26) — Node/Label 모듈 설계 확정 (frontend 카탈로그)
- frontend에서 worker의 "앱/셸"을 나열·실행하는 **supervisor PG 카탈로그** 설계 합의. 상세 전부 MEMORY "Node/Label 모듈 설계"
- **핵심**: 단일 `nodes` 테이블(kind=folder/script, 한 트리) / worker 귀속=폴더 단위(device_id, 상속) / position_x·y=NUMERIC(5,2) 실수 0~100(캔버스 % 배치) / 스크립트=작은 텍스트라 실행 시 **인라인 전송**(청크 모듈 불필요) / 공유·분류는 추후 **Label 모듈**(직교 트리, M:N)로 분리
- **다음**: 이 설계를 supervisor PG 스키마 + sqlc 쿼리 + CRUD로 구현. 단 "실행" 버튼 동작은 아래 process 모듈 MsgExec 배선이 있어야 완성됨

## 오늘(2026-06-26) 추가 작업 — user 모듈 착수 + agents 폐기
- **세션 모듈 이식**: 구 PortBridge `internal/manager/session/sessionManager.go` 가져옴(`echo/v4`+`gorilla/sessions`, `go mod tidy`로 의존성 채움). 로직 버그 3건 수정(saveData 비공개필드 패닉 / T 비-struct 패닉 / GetSession 디코드실패 잠금) + `Name()`을 `NameFunc[T]` 주입 전략으로 전환
- **user 모듈 sqlc**: 구 프로젝트 그대로 이식. `00001_users.sql`(users 테이블: identification UNIQUE/password/nickname) + `query/users.sql`(CreateUser/GetUser/GetUserByIdentification) → `superdb` gen. build 통과
- **agents 테이블 폐기**: worker는 메모리 전용이라 불필요 → 예시 `agents` 마이그/쿼리/gen 전부 삭제, users를 `00001`로 재정렬. ★node의 `device_id→agents` FK는 `device_key TEXT`(InstanceKey 문자열, FK 아님)로 변경 — MEMORY 반영 완료
- **supervisor PG 스토어 계층 이식 완료**: 구 PortBridge `internal/store/connectManager.go`(pgx) → `internal/supervisor/store/connectManager.go`(package `store`, `superdb` 참조). worker(SQLite)판이 고친 두 수정 동일 반영(afterUnsfae→afterUnsafe+결과 err 전달 / InitStorePool 재호출 시 실패연결 정리). `pgxpool`→`puddle/v2` indirect 추가. build/vet 통과
- **supervisor DB 배선 + 마이그레이션 자동적용 완료(검증됨)**: `cmd/supervisor/constants/env.go`에 DB env 추가(DB_USER/PASS/NAME/HOST 필수, DB_PORT 기본 5432) → `main.go init()`에 `mountStore`(worker 패턴: InitStorePool→Migrate, 실패=Fatalf) 배선. `store/migrate.go`(goose, dialect=postgres) + `db/migrations/embed.go` 신설. **실제 nexus DB에 00001_users 적용 확인**(users+goose_db_version 생성, goose v1). 상세 HISTORY 2026-06-26
- **다음 후보**: 요청 경계 트랜잭션 미들웨어(아래) + echo 핸들러(가입·로그인). password는 해시 저장 전제 확인 필요

## 트랜잭션 미들웨어 (구 RequestScope 대체) — 구현 완료 (2026-06-26)
- `cmd/supervisor/router/tx.go` 신설. 구 PortBridge `web/requestContext.go`(전역 map+RequestScope+도메인 *Context 4종) → **단일 `txScope` + `txMiddleware` + `Tx(c)`/`TxQueries(c)`**(둘 다 `(…,error)`). 전역 map 폐기. 상세 MEMORY "트랜잭션 미들웨어"
- error 반환·panic **둘 다 롤백**(주 경로=error 반환[return-style]). lazy Begin. `e.Use(txMiddleware(store.GetStorePool()))`(PanicMiddleware 안쪽)

## 에러 처리 = panic-style로 회귀 (PanicMiddleware) — 2026-06-26 번복
- **사용자가 return-style을 되돌림**. `util.go`를 panic 기반으로 재반입(`ClientError.Error`=필드라 error 미구현, `Err`→ClientError, `PanicMiddleware`가 렌더, `HTTPErrorHandler` 없음) → tx.go/user.go/supervisorRouter.go를 다시 panic-style로 맞춤(빌드/vet 통과)
- 핸들러=`panic(web.Err(...))`, `ensureTx/Tx/TxQueries/requireSession` error 반환 제거(TxQueries=QueriesPanic), DB·Save 에러도 `panic(web.Err(500,...))` 통일, `supervisorRouter`에서 `e.HTTPErrorHandler=...` 줄 삭제. 상세 MEMORY "에러 처리 규약" + HISTORY 2026-06-26
- (구 return-style 전환 근거·내용은 폐기됨 — MEMORY/HISTORY에 보존)

## user 가입/로그인 핸들러 — 구현 + e2e 검증 완료 (2026-06-26)
- `cmd/supervisor/router/user.go`: **echo 기본 핸들러**(func(c)error, return web.Err + `q,err:=TxQueries(c)`). 라우트 POST `/users`(가입)·POST `/users/session`(로그인)·GET(세션확인)·DELETE(로그아웃)
- **별도 User 타입 불필요**: `session.SessionManager[superdb.User]` 직접 사용(세션 키 `"irony"/"sid"` 구 동일, nameFn=nil). supervisorRouter에 `sessions` 필드 추가
- `requireSession`(구 getSessionOrPanic): flag 사슬 → 단락 OR + `web.Err(401)`
- **e2e 9시나리오 통과**(return-style 재검증): 가입200/중복400/검증실패400/미로그인401/로그인200/세션확인200/오답400/로그아웃200/로그아웃후401. 상세 HISTORY 2026-06-26
- ★관찰/잔여: ①checkSession은 세션 복원값이라 **createdAt=0**(pgtype.Timestamptz는 gob 직렬화 대상 아님 → 세션에 안 실림. 구 SQLite int64는 실렸음). 타임스탬프 필요하면 sess.Data.ID로 DB 재조회 ②응답 identification=sha256 해시(구 설계) ③Password 해시가 세션 쿠키에 실림(CookieStore 암호화, 구 동일) — 정리 여지
- **다음**: SESSION_KEY 등 env화(현재 "irony" 하드코딩) / Node 모듈 CRUD or process 모듈

## 오늘(2026-06-25) 완료 — 상세는 각 HISTORY 2026-06-25
- **EVENT 평면**(transport): REQ/RES와 대칭인 단방향 `Emit`/`On` + 전용 dispatch goroutine(순서보존). `-race` 통과. 상세 MEMORY "EVENT 평면"
- **subscribe 리네임**: `Manager`→`Hub`, `NewManager`→`New`, 내부 `subscriber`→`topic`, 파일 `hub.go`
- **worker 재연결 루프 + backoff**: 단일 `Result{Reached,Err}`(sync.Once finish, conn 1회 close), `reached`(register 성공) 기준 `backoff.Reset()`, dial 실패는 error 반환(루프 재시도). `internal/util/retry.util.go` Backoff 헬퍼(jitter 없음=보류)

## 다음 작업(착수): process 실행 모듈
- worker 핵심 = 자기 프로세스 실행/모니터링/종료(PTY)
- **설계 확정 전부 MEMORY "process 모듈 설계"에 기록** (UID/PID 분리, fan-out 허브=IInteractive 위 재사용, frontend 다리=transport.Conn+subscribe.Hub, 느린소비자 격리, frontend 결정 3건)
- **착수점**: `execute` 패키지 골격 — `IInteractive` + `ProcessSpec`/`Status` + 양쪽 manager UID 키잉
- 메시지 어휘 확장: `MsgExec`/`MsgData`/`MsgResize`/`MsgKill`/`MsgStatus`
- EVENT 평면 골격 완료 → 남은 건 도메인 배선: MsgData/MsgStatus payload(streamID=UID) → worker `Emit`/supervisor `On` → `AgentInteractive.PushOutput`/`Done` 연결

## 이전 완료: 파일 전송 모듈 (supervisor → worker) — 구현 완료 / e2e 미검증
- 송신/수신 한 바퀴 + 이어받기(resume) + sha256 무결성 검증 + 재시도(3회) + abort 전부 구현
- `go build ./...` / `go vet` 통과. **2프로세스 실제 전송·이어받기·abort 스모크는 아직 안 돌림.**
- 상세는 HISTORY 2026-06-24 참조. 관련 파일:
  - `internal/transfer/{readFile,saveFile}.go` — Hash()/Sync()
  - `cmd/supervisor/router/supervisorRouter.go` — SendFile/sendOnce/AbortFile + fileSend
  - `cmd/worker/router/workerRouter.go` — fileInit/fileChunk/fileResult/fileAbort + fileRecv

## 전송 모듈 마무리 잔여 (process 작업 전/중 정리)
1. **e2e 스모크 테스트** — supervisor+worker 띄워 SHA256SUMS 전송 → 이어받기(중간 끊고 재접속) → abort 확인
2. register 안의 임시 테스트 전송을 **밖으로 분리**(등록 완료 후 트리거) → 타이밍 레이스(subKey 미설정 시 폴더명 mainKey만) 해소
3. abort로 끝난 SendFile 에러를 정상중단으로 구분할 sentinel 필요해지면 추가(현재는 과해서 보류)

## 미해결 이슈 (이월)
- **supervisorRouter register**: `Append` 반환 무시(원자적 점유 미확인) — cleanup은 conn 기준 FindAll로 이미 전환됨
- **서브키 충돌/위조**: seq→RandomKey로 충돌 완화, key↔subkey 결속 검증은 미구현
- **supervisor 영속성**: registry 메모리 → PG 미착수 (전송/실행 안정화 후)
- (이후) 도메인 구독 계층, apps/web
