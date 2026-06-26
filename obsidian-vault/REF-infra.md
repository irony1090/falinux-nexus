# REF — 통신 인프라 (transport / subscribe / call / EVENT / util)

> 재사용 통신 코어와 인프라 자산·교훈. 상세 작업 이력은 `history/transport.md`.

## 구독/토픽 모델 (다수 agent 상태 감시)
- topic 기반 구독/해제로 selective push (전체 broadcast 회피)
- 구독 등록은 REST, push는 WS 단방향 권장 (PLAN 참고)
- 구독 시 즉시 `SNAPSHOT`, 이후 변경분만 `UPDATE`
- WS 연결 해제 시 `UnsubscribeAll` 필수 (재연결 중복/dead client publish 방지)
- 동시성: Lock 구간 분리 (구독자 수집 → Unlock 후 json.Marshal/WriteMessage)

## 재사용 구독 허브 (`internal/subscribe/hub.go`) — 구현 완료
- ※ 명명 최신화(2026-06-25): 타입 **`Hub[C]`**(구 `Manager`), 생성자 **`New`**(구 `NewManager`), 키별 집합 = 비공개 **`topic[C]`**(구 `subscriber`), 파일 **`hub.go`**(구 manager/subscriber.go)
- PortBridge subscribeManager를 도메인 제거하고 일반화한 코어
- `Hub[C comparable]`: 키는 **문자열만**, 클라이언트는 **제네릭** (`*Client`/인터페이스 모두 가능)
- 키 생성·도메인 타입은 매니저가 모름 → 프로젝트가 `Key() string`으로 생성해 넘김
- `New(marshal, send)`: 직렬화/전송 전략 1회 주입 → 호출부는 `Publish(key, payload)` 한 줄
- `Publish`: marshal 1번 → 락 밖에서 N번 send, send 실패는 `errors.Join`으로 모아 반환(브로드캐스트 중단 안 함)
- `Subscribers(key) []C`: 특수 전송용 스냅샷 탈출구
- 설계 원칙: **매니저 책임 = "문자열 키 → 구독 클라이언트 집합" 장부 + 동시성**, 그 외(키/직렬화/전송)는 프로젝트 몫

## 요청/응답 상관관계 프레임 (`internal/call` + `protocol` + `transport`) — 구현 완료
- 구 PortBridge `agentRouter.go`의 통증 해결: 요청·응답이 두 파일 switch에 흩어지고 짝 ID가 없어 추적 곤란 → 도메인 매니저 조회로 역추적하던 문제
- **`call.Correlator[R]`**: "요청 ID → 응답 기다리는 채널" 장부 + 동시성만. send 주입, 응답 타입 R 자유. `Call(ctx,payload)`(블록)/`Resolve(id,val,err)`/`Close(err)`(전부 깨움). subscribe.Hub와 동일 철학(코어는 도메인/전송 모름)
- **`protocol.Frame{Kind,ID,Type,Err,Data}`**: Kind=**REQ/RES** 로 요청/응답 프로토콜 레벨 구분. Data=json.RawMessage(자유 payload). REQ가 ID 채번 → RES가 반사
- **`transport.Conn`**: `Call`(REQ→RES 블록)/`Handle`(REQ 핸들러)/`Serve`(수신 루프). supervisor·worker 양쪽 대칭 API. `MessageRW` 인터페이스에만 의존(gorilla `*websocket.Conn`이 그대로 만족)
- 연결 끊김 시 `corr.Close(err)`로 매달린 Call 일괄 실패 = Done(502) reconciliation과 같은 정신
- **확장 완료**: EVENT 평면(아래) + `transport.Pipe()` 재구현됨(2026-06-25)

## EVENT 평면 (단방향 1:N, `protocol`+`transport`) — 구현 완료 (2026-06-25)
- REQ/RES(제어, 1:1, Call 블록)와 대칭인 **데이터 평면**: 짝 없는 단방향(`Emit` 즉시리턴 / `On` 핸들러). 용도=출력 스트리밍(MsgData), STATUS 푸시. **Correlator 안 거침**
- `protocol`: `EVENT` Kind + `NewEvent(t,data)`(ID 미채번). Frame 구조 불변(Kind 분기만)
- `transport.Conn`: `On`(핸들러 등록)/`Emit`(=`write`만)/`dispatch()`(전용 goroutine). streamID는 **payload**에 (transport 도메인 무지 유지, Frame.ID는 REQ 전용)
- **순서 보존 체인**: 송신 `wmu` 직렬화 → Serve 순차 push → **단일 dispatch goroutine** 순차 호출 = stdout 순서 end-to-end 보존. (REQ는 `go dispatchReq`라 순서 무관, EVENT는 순차 필수)
- **결정 1(안전종료)**: `events` 채널 **절대 close 안 함** → `Close`에서 `close(done)`만. Serve push는 `select{events<-f; <-done}` 가드 → 끊김 순간 send-to-closed 패닉 차단. transfer의 `done` 패턴 재사용
- **결정 2(수명계약)**: dispatch는 `New`서 1회 시작, `Close`(owner=라우터)서 종료. 안 부르면 goroutine 누수
- backpressure: `events` 버퍼(256) 차면 Serve 블록(=느린 구독자 신호). On(DATA) 핸들러는 "버퍼 push/forward"라 빨라야 정상
- `transport.Pipe()`: 인메모리 양방향 MessageRW(한쪽 Close→공유 closed로 양쪽 깨움). 테스트: Emit×1000 순서+개수 / REQ·EVENT 혼류 무간섭, 둘 다 `-race` 통과

## worker 재연결 + 신호 설계 교훈 (2026-06-25)
- **재연결 정책은 router 바깥(main for문)**: router 필드(conn/saves/subKey)=연결 귀속 상태 → 새 router로 자동 폐기. 장수 상태(store/향후 process매니저)는 main서 만들어 주입. (router=연결 하나, transport.Conn이 스스로 재연결 안 하는 철학과 동일)
- **★신호 교훈(재사용)**: "끝날 수도 안 끝날 수도 있는 신호 2개(Ready+Done)" 대신 **"반드시 한 번 끝나는 신호 1개에 결과를 실어라"**(`Result{Reached,Err}` 단일 채널). 그러면 "에러 때 Ready도 닫나" 문제가 *풀리는 게 아니라 사라짐*. `sync.Once finish`로 다송신자(Serve/register) 1회 송신+conn 1회 Close까지 동시 해결
- backoff: `internal/util/retry.util.go` `Backoff{cur,base,max}` `Next()`(2배·상한)/`Reset()`. `reached`(register 성공) 기준 Reset → 정상가동 후 끊김=짧게, 그 전 실패=증가. **jitter 미적용**(worker 다수·supervisor 동시재접속 부하 시 full jitter 추가)

## transport.Conn 수명 상태
- `ConnState`(StateActive=0/StateClosed) + `State()` + `String()`. `atomic.Int32` 필드(zero=Active). Serve 수신루프 종료·Close에서 StateClosed 전이(둘 다 idempotent)
- 도메인 무지·대칭 유지 — "연결이 살아있나"는 conn 속성, **신원(메인키#서브키)은 라우터 몫**(conn 아님)

## `internal/util/string.util.go` — RandomKey
- `RandomKey(length int, prefix, suffix string) (string, error)`: base62(`0-9A-Za-z`), length=prefix/suffix **포함** 전체 길이(결과 정확히 length), `crypto/rand`
- 62는 256 약수 아님 → **거부 샘플링**(>=248 바이트 버림)으로 modulo 편향 제거 (hex 16이면 `b&0x0f`로 충분했음). 타입: `maxByte byte`, 인덱스 `b%byte(len(alphabet))`
- 용도: 서브키 발급, transferId. `*.util.go`가 util 패키지 컨벤션

## `internal/manager/KeyValManager` (PortBridge 이식)
- 제네릭 `KeyValManager[K comparable, V any]`: 키→**단일 값** 장부 + `OnCreated`/`OnRemoved` 수명 콜백. `subscribe.Hub`(키→**집합**)와 역할 다름
- **함정 4개**: ① `FindAll` predicate는 **락 안**에서 실행 → predicate서 매니저 재호출 금지(데드락) ② 콜백은 락 바깥(LIFO defer)이라 TOCTOU 가능 ③ `OnCreated/OnRemoved`는 동시사용 전 1회 세팅(동기화 없음) ④ `Append`는 키 존재 시 false → **반환값 꼭 확인**(무시하면 중복/경합이 조용히 묻힘)

## Go 교훈: 포인터도 "값으로" 복사된다
- 핸들러가 채운 값을 호출자가 보려면 **호출자가 실제 구조체를 들고 그 주소(`&val`)를 넘겨야** 함. `var p *T`(nil) 넘기고 핸들러가 `p`를 새 구조체로 재할당해도 **호출자 변수엔 안 보임**(포인터 변수 자체가 복사됐으므로) → supervisorRouter cleanup이 안 돌던 원인
- 응용: 연결별 cleanup은 공유 변수 대신 **스코프의 `conn` 값으로** `FindAll(c==conn)` 하면 레이스·오삭제 둘 다 없앰

## http 서버 구성
- 패키지 전역 `http.HandleFunc`/`ListenAndServe(_,nil)`는 전역 `DefaultServeMux` 공유 → 라우팅 섞임·서버별 타임아웃/shutdown 불가. **명시적 `http.NewServeMux()` + `&http.Server{}`** 권장. `NewSupervisorRouter`가 mux 내부 생성 후 `*http.Server` 반환하는 형태로 정리됨

## 상세 PLAN
- `PLAN-agent-comm.md` (agent↔server 통신/PTY 실행 상세)
- `PLAN-subscription.md` (구독/토픽 모델 상세)
