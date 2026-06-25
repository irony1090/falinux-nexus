# HISTORY

> CURRENT/MEMORY가 비대해지는 것을 막기 위한 상세 기록 저장소.
> 작업 상태가 바뀌면 기존 CURRENT 내용을 여기로 정리해 내린다.

## 과거 작업 기록

### 2026-06-25 - EVENT 평면(단방향 데이터 평면) 재구현 (구현 완료, `-race` 통과)

> 상태: **transport 레벨 EVENT(Emit/On) 평면 골격 완료. `go build ./...`/`go vet ./...`/`go test -race` 전부 통과.** process 출력 스트리밍 토대. 도메인 배선(MsgData/MsgStatus)은 process 모듈에서.

#### 배경 / 정정
- process 출력 스트리밍은 **응답 없는 단방향(1:N)** → REQ/RES(Correlator=ID→대기채널 1:1) 평면에 안 맞음 → 별도 EVENT 평면 필요
- vault에 "전에 만들었다 제거, git 히스토리에 있음"이라 적혀 있었으나 **확인 결과 커밋엔 없음**(작업트리에서만 만들었다 커밋 전 삭제). `-S Emit/EVENT`·reflog·dangling·stash 전부 빈 결과 → 복구 불가라 **새로 작성**

#### 설계 결정 (사용자와 C안 합의)
- 디스패치 전략 3안 중 **C(전용 dispatch goroutine + 버퍼채널)** 채택: A(go마다)=순서깨짐 / B(Serve 동기)=느린핸들러가 수신루프(RES포함) 막음·콜백 데드락 / **C=순서보존+비차단**
- **결정 1(안전종료)**: `events` 채널 절대 close 안 함 → `Close`에서 `close(done)`만. 닫는 주체(Close=다른 goroutine)와 보내는 주체(Serve)가 달라 `close(events)`면 send-to-closed 패닉. Serve push를 `select{events<-f; <-done}`로 가드. transfer 모듈 `done` 패턴 재사용
- **결정 2(수명계약)**: dispatch goroutine은 `New`서 1회 시작(단일 시작점=race 없음), `Close`(owner=라우터)서 종료. rw.Close 때문에 owner는 어차피 Close 호출 → 새 부담 아님. 안 부르면 누수
- streamID는 **payload**에 둠(Frame.ID는 REQ 채번 전용, transport 도메인 무지 유지)

#### 작업 완료
- `internal/protocol/protocol.go`: `EVENT` Kind + `NewEvent(t,data)`(ID 미채번). Frame 구조 불변
- `internal/transport/conn.go`: `EventHandler` 타입, `on` 맵·`events`/`done` 채널·`eventBufferSize=256`, `New`서 `go c.dispatch()`, `On`/`Emit`(=write만), `dispatch()`(select{done; events}→핸들러 순차호출, 미등록 무시), `Serve` EVENT 분기(done 가드), `Close` `close(done)`
- 신규 `internal/transport/pipe.go`: 인메모리 양방향 MessageRW. 공유 `closed`+`sync.Once`로 한쪽 Close→양쪽 Read/Write 깨움. WriteMessage는 버퍼 복사(aliasing 방지)
- 신규 `internal/transport/conn_test.go`: ①Emit×1000 순서+개수 ②REQ/RES·EVENT 혼류 무간섭(EVENT 폭주 중 Call 정상응답). 둘 다 `-race` 통과
- Go 1.26 idiom: `for i := range n`로 모던화(lint 반영)

#### 순서 보존 체인 (핵심)
- 송신 `wmu` 직렬화 → 와이어 순서 = Serve 수신 순서 → Serve가 순서대로 `events` push → **단일 dispatch goroutine**이 하나씩 호출 = stdout 청크 순서 end-to-end 보존. (REQ가 `go dispatchReq`인 건 순서 무관해서, EVENT는 순차 필수라 대비)

#### 의도적 트레이드오프
- Close 시 `events` 버퍼 잔여분 **버림**(drain 안 함) = drop-on-disconnect = Done(502) 비관적 정리 정신
- backpressure가 RES 배달에 영향: `events` 꽉 차 Serve 블록 시 RES도 잠시 못 읽음. 단 On(DATA) 핸들러는 버퍼push/forward라 빨라야 정상(느리면 설계결함 신호), 버퍼256이면 현실적 안 닿음

#### 다음 (process 모듈)
- MsgData(stdout/stderr)·MsgStatus(RUNNING/STOPPED+ExitCode) payload 정의(streamID 포함) → worker `Emit`/supervisor `On` → `AgentInteractive.PushOutput`/`Done(exitCode)` 배선

---

### 2026-06-24 - 파일 전송 모듈 본체 + abort 구현 (구현 완료, e2e 미검증)

> 상태: **SendFile→완료까지 한 바퀴 + 이어받기 + hash검증 + 재시도 + abort 전부 구현. `go build ./...`/`go vet` 통과. 2프로세스 실제 전송 e2e는 아직 안 돌림.**

#### 해시 메서드 (transfer)
- 설계 갈림길 정리: Hash는 **원본 파일 전체 sha256(hex 64자)**. 출력은 입력 크기 무관 고정(32B)이라 와이어 부풀음 걱정 無 — 유일 비용은 "선읽기 I/O/CPU"
- resume를 지원하면 전체 해시 선계산이 어차피 강제됨 → **A안(FileInit에 전체 해시)** 채택
- `ReadFile.Hash()`: **별도 fd(`os.Open(source.Name())`)로 풀읽기** → r.source 커서·r.read·expire 타이머와 완전 분리. `sync.Once`로 1회 계산·캐시(hashVal/hashErr), 락 불필요(Once가 동기화)
- `SaveFile.Hash()`: 동일 방식이나 **캐시 안 함** — 수신 파일은 전송 중 계속 변함, "완료(Completed) 후 1회" 호출용. `s.mu` 잡아 in-flight 쓰기와 일관성. WriteAt 비순차/재전송/이어받기를 running hash로 못 다루니 결과 파일 풀읽기가 정답
- `SaveFile.Sync()` 추가: rename 직전 fsync(크래시 시 버퍼 유실 방지)

#### 전송 본체 (supervisor 송신)
- `SendFile(authKey, reader)`: transferId 발급(RandomKey 16) → `fileSend{authKey,reader,cancel}` 등록 → `reader.Hash()` → **재시도 루프(maxSendAttempts=3)**. 거부(errRejected)는 즉시 중단, 그 외 실패는 retryBackoff(500ms) 후 재시도, N회 초과 시 error. destPath=원본 파일명
- `sendOnce(ctx,...)`: FileInit(resume=true) → resumeOffset 검증 → `reader.SeekTo` → 청크 루프(32KB) → FileResult 검증. 3개 conn.Call timeout(10s)은 **부모 ctx에서 파생**
- 연결 끊김 시 그 conn 소속 reader만 `FindAll(authKey==key)`로 정리(누수 해소). **이중 Append 버그 제거**

#### 전송 본체 (worker 수신, `cmd/worker/router`)
- `fileRecv{save,hash,finalPath}` 래퍼: FileResult 검증에 FileInit의 hash·최종경로 필요한데 FileResultRequest엔 transferId만 실리므로 보관
- 4핸들러: `fileInit`(traversal검증+`.part` stat으로 resumeOffset+SaveFile생성) / `fileChunk`(WriteAt 멱등) / `fileResult`(**Completed→Hash일치→Sync→Close→Rename(.part→최종)**) / `fileAbort`(`.part` 삭제)
- 검증 분기: 크기미달=`.part`보존+Resumable:true / hash불일치=`.part`삭제+Resumable:false(이어받아도 손상은 해결 안 됨) / 성공=원자적 rename
- traversal: `Clean("/"+destPath)`로 선행 `..` 무력화 후 `filepath.Rel(root,final)`로 루트 하위 재확인
- perm=0이면 0644 보정(read-back/Hash 위해), subKey는 register에서 락으로 확정(핸들러가 저장경로에 사용)
- baseDir=`env.ProcessRoot` 전달(main.go), 저장경로 `${baseDir}/${mainKey}#${subKey}/${destPath}`

#### abort (송신측, worker fileRecv와 대칭)
- `fileSend`에 `cancel context.CancelFunc` 추가. `SendFile`이 `context.WithCancel(Background)`로 ctx 생성→저장, 루프에 **`ctx.Err()` 체크**(abort 시 재시도 차단)
- `AbortFile(transferId,reason)`: ①`fs.cancel()`로 진행 중 Call 깨움→루프 종료 ②worker에 `MsgFileAbort`(best-effort). reader 정리는 SendFile의 defer Close가 전담(abort/정상/끊김 한 경로 수렴)

#### context 교훈 (재확인)
- `cancel()`은 "연결 닫기"가 아니라 **그 ctx의 타이머·parent children 등록 해제**. 안 불러도 Background 자식은 timeout 시 자동 정리되나 **그때까지 쌓임** → 청크 루프에선 `defer` 금지, Call 직후 즉시 `cancel()` (defer는 함수 끝까지 안 풀려 타이머 수천 개 누적). parent가 취소가능 ctx면 안 부르면 진짜 누수
- abort용 per-call timeout은 부모를 Background가 아닌 취소가능 ctx로 파생해야 cancel이 전파됨

#### 미검증/임시
- **e2e 미검증**: 빌드/vet만 통과, 2프로세스 실제 전송·이어받기·abort 스모크 안 돌림
- register 안에 임시 테스트 전송(`SHA256SUMS`) 존재 — **블로킹 SendFile이라 `go`로 감쌈**(안 그러면 register 응답 지연→worker 타임아웃), reader expire=0
- 타이밍 레이스: 등록 직후 subKey 미설정 상태로 FileInit 도착 시 저장 폴더명이 mainKey만 될 수 있음(전송 자체는 정상) → 테스트 전송을 register 밖으로 빼면 해소

---

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
- (한때 만들었다 제거: `transport/pipe.go` 인메모리, EVENT On/Emit, `*_test.go` 2개 — 예제 최소화 요청. ※ **커밋 전 작업트리에서만 삭제라 git엔 없었음** — 2026-06-25 새로 재구현)

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

---

### 2026-06-18 - `internal/transfer` 검토 (readFile.go / saveFile.go, 미수정)

> PortBridge에서 이식해온 파일 전송용 객체(`ReadFile`/`SaveFile`)의 로직 검토. 아직 미배선, 수정 안 함. 향후 파일 전송 모듈 착수 시 반영.

#### 공통 구조
- `expire>0`이면 idle 타임아웃 타이머 + goroutine(`<-expireTimer.C` 1회 받고 `Close()` 후 종료). 작업(Read/Write)마다 `Stop()`→작업→`Reset()` 패턴. expire 동안 작업 없으면 발사→Close→`OnClose()`
- 정상 흐름(작업 반복 중 Stop→Reset)은 의도대로 동작 — goroutine은 그동안 계속 `<-C`에 블록돼 살아있음

#### 🔴 버그
1. **타이머 goroutine 영구 누수** (readFile+saveFile)
   - `Stop()`이 true(정지 성공) 반환 후 **`Reset` 없이 리턴하는 경로**:
     - readFile: `source.Read` 에러 리턴 / EOF 경로(`read>=size`, 80행서 Stop 먼저 후 Reset 없이 io.EOF)
     - saveFile: `dest.Write` 에러 리턴(initChunkUnsafe 안 탐)
   - → 타이머 정지된 채 남고 goroutine은 `<-C`에 **영원히 블록**. 그 경로에선 타이머발 `OnClose()` 자동정리도 안 됨(호출부가 명시 Close 해야 정리)
   - **해결안**: goroutine을 `select { case <-timer.C: case <-done: }`로, `Close()`에서 `done` 닫기(+`timer.Stop()`) → 확실히 깨워 종료. 정상 완료 시에도 즉시 정리됨
2. **데이터 레이스** (readFile `Current()`): `r.read`를 락 없이 읽음(`Read()`는 `mu` 안에서 `r.read += n`). 바로 옆 `Completed()`는 락 잡음. saveFile `Written()`도 락 잡음 → **readFile `Current()`만 누락**

#### 🟡 설계상 알아둘 점
- `Close()`가 타이머 미정리 → 1번 고쳐도 정상 완료 후 goroutine이 최대 expire만큼 생존(지연 정리)
- partial write/read 시 카운터(`written`/`read`)와 실제 파일 크기 불일치 가능(객체 수명 내). 재시작 시 `os.Stat`으로 자가보정
- **fsync 없음** — 완료 선언(`Validate`/`Completed`) 전 `dest.Sync()` 없어 크래시 시 버퍼 유실 가능. 전송 무결성 모듈이면 완료 직전 Sync 권장
- `Validate()`는 바이트 수만 비교(`written==size`) — 체크섬 아님, 이름이 과함. 무결성 보장하려면 해시 비교 별도
- `Read`가 `io.ReadFull`이 아니라 short read 가능 → chunk가 chunkSize보다 작을 수 있음(소비자가 가변 크기 처리하면 무방)

---

### 2026-06-19 - `internal/transfer` 🔴 누수·레이스 수정 (A 방식)

> 위 2026-06-18 검토에서 발견한 🔴 2건(타이머 goroutine 영구 누수, readFile `Current()` 데이터 레이스)을 수정. `go build`/`go vet ./internal/transfer/` 통과.

#### 🔴 1. 타이머 goroutine 영구 누수 → `done` 채널 방식(A)
- **원인**: goroutine이 `<-timer.C` **하나만** 대기 → `Read`/`Write`의 EOF·에러 경로가 `Stop()`(타이머 끔) 후 `Reset` 없이 리턴하면 타이머가 꺼진 채 남아 goroutine이 `<-timer.C`에 영원히 블록. `Stop()`은 "체크"가 아니라 **정지(부수효과)**라는 점이 핵심
- **수정** (readFile.go / saveFile.go 대칭):
  1. 구조체에 `done chan struct{}` 추가
  2. 생성 시 goroutine을 `select { case <-timer.C: Close(); case <-done: }`로 — 둘 중 하나라도 오면 종료
  3. `Close()`(onceClose.Do 안)에서 `timer.Stop()` + `close(done)` → goroutine 확실히 깨움. onceClose라 close 중복 panic 없음
- **효과**: EOF·에러 경로의 `Reset` 누락이 무해해짐(누수 소멸). 정상 완료 후 `Close()` 시 즉시 종료(지연 정리 없음). 타이머 발화→Close 경로는 goroutine이 이미 select 탈출 후라 close(done) 무해
- **택일 메모**: A(=done 채널만, EOF 즉시정리는 호출부가 `Completed()` 확인 후 Close/Remove). B(=Read 안에서 명시 Close)는 락 재진입 주의 필요 → transfer 객체 단순 유지 위해 A 채택

#### 🔴 2. 데이터 레이스 (readFile `Current()`)
- `r.read`를 락 없이 읽던 것(`Read()`는 `mu` 안에서 씀)에 `r.mu.Lock()/Unlock()` 추가. saveFile `Written()`은 이미 락 잡고 있어 변경 불필요

#### 동시성 주의 (의도된 동작)
- 타이머 발화 시 goroutine이 `Close()` 호출 → `mu` 잡는데, `Read`/`Write`가 `mu` 보유 중이면 잠깐 대기(데드락 아님 — Read/Write는 goroutine을 안 기다림)

#### 남은 🟡 (미수정, 파일 전송 모듈 본격 착수 시 정책 결정)
- fsync 없음(완료 선언 전 `Sync()` 권장) / `Validate()`는 바이트수만 비교(체크섬 아님) / `Read` short read 가능 / partial 카운터 불일치(재시작 `os.Stat` 자가보정)

---

### 2026-06-19 - 전송 모듈 착수 준비: conn 수명상태 / util·manager 도입 / supervisorRouter 정비 / 전송 프로토콜 설계

> worker 파일 전송 모듈 본격 착수 직전. 인프라(conn 상태·RandomKey·KeyValManager) 갖추고 supervisorRouter 정비 + 전송 프로토콜 설계 확정. **전송 본체는 미구현** — 진행 중 설계는 CURRENT 참조. 재사용 자산/교훈은 MEMORY 기록.

#### transport.Conn 수명 상태 (`internal/transport/conn.go`)
- `ConnState`(StateActive=0/StateClosed) + `String()` + `State()`. `atomic.Int32` 필드(zero=Active라 New 초기화 불필요)
- 전이 2곳: Serve의 ReadMessage 에러 경로 + Close(`closeOnce.Do` 안), 둘 다 idempotent (Serve가 죽었지만 Close 전 짧은 구간도 정확히 closed)
- "현재 상태" 의미는 사용자 선택 = **연결 수명(enum)**. (대안: 연결별 신원 보관함 any — 기각. 신원은 라우터 몫)
- `go build`/`vet` 통과

#### `internal/util/string.util.go` — RandomKey
- base62(`0-9A-Za-z`) 랜덤키. `RandomKey(length, prefix, suffix)`, length=prefix/suffix 포함 전체 길이. crypto/rand + **거부샘플링**(>=248 버림)으로 편향 제거
- 빌드 중 타입 불일치(byte vs int) 잡음: `maxByte`를 `byte`로, 인덱스 `b%byte(len(alphabet))`
- util 패키지 컨벤션은 `*.util.go`(기존 path/char/byte.util.go). 용도: 서브키·transferId

#### KeyValManager 이식 (`internal/manager/KeyValManager.go`)
- PortBridge `KeyValManager[K,V]` 그대로(순수 stdlib·도메인무지). 검토 결론 = "버그 없음 + 함정 4개"(MEMORY 기록): FindAll predicate 락 안 실행 / 콜백 락 바깥 TOCTOU / 콜백필드 동기화없음 / Append 반환 확인 필요
- `subscribe.Manager`(키→집합)와 역할 다름(키→단일값+수명콜백) → 중복 아님

#### supervisorRouter 정비 (`cmd/supervisor/router/supervisorRouter.go` + main.go)
- `NewSupervisorRouter(addr,path)` → 내부 `http.NewServeMux()` 생성 후 `(*http.Server, *supervisorRouter)` 반환. main은 명시적 `http.Server.ListenAndServe()`. **전역 DefaultServeMux 탈피**(라우팅 격리·타임아웃·graceful shutdown 가능)
- `connectors`(authKey→Conn) + `readers`(transferId→reader) 두 KeyValManager. register를 라우터 메서드로(클로저가 conn 캡처)
- **컴파일 막던 것**: 커스텀 `registerHandler`는 `transport.Handler`와 underlying 같아도 **named끼리 대입 불가** → 반환타입을 `transport.Handler`로
- **포인터 값전달 버그(교훈)**: `var auth *T`(nil) 값으로 넘기면 핸들러가 Bind로 채워도 호출자 `auth`엔 안 보임(별개 변수) → cleanup `if auth!=nil` 영원히 거짓(린터도 "impossible condition" 경고). 해결: `var auth T`+`&auth`로 실제 구조체 주소 전달. (MEMORY 교훈)
- **남은 권고(미적용)**: register `Append` 반환 확인(원자적 점유) / cleanup을 요청키 아닌 conn 기준 `FindAll(c==conn)`로 (거부된 재접속이 정상접속자 evict 방지 + 공유변수 레이스 제거)

#### `internal/transfer/readFile.go` — Info 추가 + size 버그
- `info os.FileInfo` 필드 + `Info()`/`Perm()`/`Size()`. Info/Size는 New에서 1회 세팅 후 불변 → 락 불필요
- **버그**: `size`를 `info.Size()`에서 안 채워 0 → `Read`가 즉시 `io.EOF`(전송 통째로 안 됨). `size = s.Size()`로 수정
- 교훈: info·size 두 출처 분리가 드리프트 원인 → 가능하면 `Size()`가 `info.Size()` 반환하게 합치는 게 안전(현재는 일관)

#### 파일 전송 프로토콜 설계 (`internal/protocol/messages.go`, 미구현)
- 3단계(통보 FileInit / 전송 FileChunk / 결과 FileResult) + FileAbort. `MsgFileTransfer` 단일 → 단계별 MsgType 분리 권장
- 보완 확정 4가지: **transferId**(전송 식별, Frame.ID와 별개) / **init 응답 resumeOffset**(받는쪽 결정) / **무결성 hash** / **offset 기반 chunk**(short read로 chunk 가변이라 index 불가)
- transferId = `authKey/filename` 아님(충돌·resume혼선) → `RandomKey` 발급, readers 값에 `{authKey, reader}` 묶어 끊김 시 FindAll 정리
- 수신 저장경로 `/${base}/${mainKey}#${subKey}/${DestPath}` + traversal검증 / MkdirAll / `.part`→hash→fsync→rename / resume=.part stat
- 인코딩: Frame.Data가 json이라 chunk []byte base64 ~33% 부풀음(대용량시 raw WS 프레임 분리 검토)
- 상세 설계·다음 단계는 CURRENT
