# HISTORY — 파일 전송 모듈 (internal/transfer)

> ReadFile/SaveFile 검토·수정, 전송 본체+abort 구현, 착수 준비(인프라) 상세 기록.
> 요약·재사용 지식은 `REF-transfer.md`, 현재 상황은 `CURRENT.md`.

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

> worker 파일 전송 모듈 본격 착수 직전. 인프라(conn 상태·RandomKey·KeyValManager) 갖추고 supervisorRouter 정비 + 전송 프로토콜 설계 확정. **전송 본체는 미구현**.
> ※ conn 수명상태·RandomKey·KeyValManager 재사용 지식은 `REF-infra.md`에도 정리됨.

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
- PortBridge `KeyValManager[K,V]` 그대로(순수 stdlib·도메인무지). 검토 결론 = "버그 없음 + 함정 4개"(REF-infra 기록): FindAll predicate 락 안 실행 / 콜백 락 바깥 TOCTOU / 콜백필드 동기화없음 / Append 반환 확인 필요
- `subscribe.Manager`(키→집합)와 역할 다름(키→단일값+수명콜백) → 중복 아님

#### supervisorRouter 정비 (`cmd/supervisor/router/supervisorRouter.go` + main.go)
- `NewSupervisorRouter(addr,path)` → 내부 `http.NewServeMux()` 생성 후 `(*http.Server, *supervisorRouter)` 반환. main은 명시적 `http.Server.ListenAndServe()`. **전역 DefaultServeMux 탈피**(라우팅 격리·타임아웃·graceful shutdown 가능)
- `connectors`(authKey→Conn) + `readers`(transferId→reader) 두 KeyValManager. register를 라우터 메서드로(클로저가 conn 캡처)
- **컴파일 막던 것**: 커스텀 `registerHandler`는 `transport.Handler`와 underlying 같아도 **named끼리 대입 불가** → 반환타입을 `transport.Handler`로
- **포인터 값전달 버그(교훈)**: `var auth *T`(nil) 값으로 넘기면 핸들러가 Bind로 채워도 호출자 `auth`엔 안 보임(별개 변수) → cleanup `if auth!=nil` 영원히 거짓(린터도 "impossible condition" 경고). 해결: `var auth T`+`&auth`로 실제 구조체 주소 전달. (REF-infra 교훈)
- **남은 권고(미적용)**: register `Append` 반환 확인(원자적 점유) / cleanup을 요청키 아닌 conn 기준 `FindAll(c==conn)`로 (거부된 재접속이 정상접속자 evict 방지 + 공유변수 레이스 제거)

#### `internal/transfer/readFile.go` — Info 추가 + size 버그
- `info os.FileInfo` 필드 + `Info()`/`Perm()`/`Size()`. Info/Size는 New에서 1회 세팅 후 불변 → 락 불필요
- **버그**: `size`를 `info.Size()`에서 안 채워 0 → `Read`가 즉시 `io.EOF`(전송 통째로 안 됨). `size = s.Size()`로 수정
- 교훈: info·size 두 출처 분리가 드리프트 원인 → 가능하면 `Size()`가 `info.Size()` 반환하게 합치는 게 안전(현재는 일관)

#### 파일 전송 프로토콜 설계 (`internal/protocol/messages.go`)
- 3단계(통보 FileInit / 전송 FileChunk / 결과 FileResult) + FileAbort. `MsgFileTransfer` 단일 → 단계별 MsgType 분리 권장
- 보완 확정 4가지: **transferId**(전송 식별, Frame.ID와 별개) / **init 응답 resumeOffset**(받는쪽 결정) / **무결성 hash** / **offset 기반 chunk**(short read로 chunk 가변이라 index 불가)
- transferId = `authKey/filename` 아님(충돌·resume혼선) → `RandomKey` 발급, readers 값에 `{authKey, reader}` 묶어 끊김 시 FindAll 정리
- 수신 저장경로 `/${base}/${mainKey}#${subKey}/${DestPath}` + traversal검증 / MkdirAll / `.part`→hash→fsync→rename / resume=.part stat
- 인코딩: Frame.Data가 json이라 chunk []byte base64 ~33% 부풀음(대용량시 raw WS 프레임 분리 검토)
