# HISTORY — transport / 통신 인프라 (subscribe·call·protocol·EVENT·재연결)

> 재사용 통신 코어(구독 허브, 요청/응답 상관관계, EVENT 평면, worker 재연결) 상세 기록.
> 요약·재사용 지식은 `REF-infra.md`, 현재 상황은 `CURRENT.md`.

---

### 2026-06-25 - worker 재연결 루프 + backoff + subscribe 리네임

#### subscribe 리네임 (refactor)
- `subscribe.Manager`→**`Hub`**, `NewManager`→**`New`**, 내부 키별집합 `subscriber[C]`→**`topic[C]`**, 파일 `subscriber.go`→**`hub.go`**. 외부 API 의미 동일, 호출부 없어서 churn 최소. doc 주석(call/correlator, supervisor/subscribe) 갱신

#### worker 재연결 (feat)
- **배경**: 연결 끊기면 N초 쉬고 재접속. "재연결 로직을 router 안 vs 바깥 for문" 고민 → **바깥 for문(main)** 채택
- **이유**: router 필드(conn/saves/subKey)는 전부 **연결 귀속 상태** → 새 router 만들면 자동 폐기. 장수 상태(store, 향후 process매니저)는 main에서 만들어 주입. router=연결 하나, 재연결 정책=바깥(transport.Conn이 스스로 재연결 안 하는 것과 동일 철학)
- **신호 설계 — 핵심 교훈**: 처음엔 `Ready`(접속)+`Done`(종료) 두 채널로 갔다가 "에러 때 Ready도 닫아야 하나"로 복잡해짐 → **"끝날 수도 안 끝날 수도 있는 신호 2개" 대신 "반드시 한 번 끝나는 신호 1개에 결과를 실어라"**. `Result{Reached bool, Err error}` 단일 채널로 수렴 → 복잡함이 *풀리는 게 아니라 사라짐*
- **구현**(`workerRouter`): `done chan Result`(cap1) + `sync.Once finish`(누가 먼저 끝내든 1회 송신 + conn 1회 Close). `reached atomic.Bool`=register 성공 시 set, Serve 종료 경로가 읽어 보고. dial 실패는 `Fatalf`→error 반환. 핸들러를 Serve보다 먼저 등록(등록 전 REQ 수신 창 제거)
- **`go Serve`를 router 생성 뒤로 이동**(finish가 router.reached 읽어야 하므로). sync.Once가 "cap1 채널 두 송신자(Serve/register) 누수"도 같이 해결
- **main**: for 루프. `res.Reached`면 `backoff.Reset()`(정상 가동 후 끊김=짧게 재시도), 그 전 실패는 backoff 증가. Ready goroutine 제거로 누수 경로 소멸
- **backoff**(`internal/util/retry.util.go`): `Backoff{cur,base,max}`, `Next()`(2배·상한)/`Reset()`. **jitter 없음** — worker 수 적어 보류, supervisor 동시재접속이 부하될 때 full jitter 추가 예정
- **버그(잡음)**: 처음 구현서 `Reached`가 영영 false(필드·set 누락, finish가 false 하드코딩) → reset 무력화. reached 필드+register set+Serve 보고로 수정
- **하우스키핑**: `.gitignore`에 `/worker`,`/supervisor` 추가(*.exe만 잡혀 Linux ELF 누락). `go build ./cmd/worker/`가 산출물 떨군 것 삭제 → 컴파일확인은 `go build ./...` 사용

---

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

### 2026-06-17 - 재사용 구독 매니저 리팩터링 (subscribe.Manager)

#### 배경
- PortBridge `internal/manager/subscribeManager.go`를 Nexus로 이식하려는데 확장성/가독성 문제 발견 → 도메인 제거 + 일반화하기로 결정

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
- `internal/subscribe/manager.go` 생성 (제네릭 Manager + subscriber, 죽은 주석 전부 제거) ※ 2026-06-25 `hub.go`/`Hub`로 리네임됨
  - `Subscribe/Unsubscribe/UnsubscribeAll/Subscribers/Publish`
  - 락 최적화 유지: Subscribers가 스냅샷 반환 → 전송은 락 밖
- `internal/supervisor/subscribe/doc.go`: 도메인 전용 계층임을 명시 (코어 사용)
- worker main.go에 임시 테스트(Test1→[]byte{1}, Test2→[]byte{2}, 구독/Publish/Unsubscribe) 작성 → 검증 후 제거
- `go build ./...` OK

---

### 2026-06-18 - 요청/응답 상관관계 프레임 + 등록(REGISTER) 핸드셰이크

> 상태: **구현 동작 확인 완료. 사용자가 직접 리팩토링 예정** (현 코드를 베이스로).

#### 배경 / 문제의식
- 구 PortBridge `agentRouter.go`(test-jig) 검토: 요청·응답이 **두 파일 switch에 흩어져** 흐름 추적 곤란. `Packet{Type,SessionId,Payload}`에 짝지을 ID가 없어 도메인 매니저(`GetInteractive(payload.Id)`)로 역추적. 요청/응답/스트리밍 3성격이 한 평면에 뒤섞임
- → "요청↔응답을 한 묶음으로 추적 가능한, 도메인 자유로운(재사용) 프레임"이 필요하다는 결론

#### 설계 결정
- 핵심 통찰: 요청↔응답을 묶는 데 필요한 건 "ID → 기다리는 채널" 맵 하나. 나머지(데이터 모양/인코딩/소켓)는 전부 주입 → subscribe.Manager와 같은 철학
- 3층 분담: **correlator(엔진, 고정) / protocol(어휘, 채움) / conn(둘을 묶어 동작 얹음)**
- 제어평면(REQ/RES 1:1) vs 데이터평면(EVENT 스트림 1:N) 분리 원칙 (단 EVENT는 예제 최소화로 일단 제거 → 2026-06-25 재구현)

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
