# REF — 파일 전송 모듈 (internal/transfer)

> supervisor→worker 단일 전송. FileInit→FileChunk×N→FileResult(+FileAbort). 송신/수신 대칭.
> **구현 완료, e2e 미검증.** 2026-06-29 reader 추상화 → 파일(`SendFile`)·메모리(`SendBuffer`) 양쪽 전송. 상세 이력 `history/transfer.md`.

## 무결성 해시 (`transfer.Hash()`)
- Hash = **원본 전체 sha256(hex 64자)**. 출력은 입력 크기 무관 고정(32B) → 와이어 부풀음 無, 비용은 선읽기 I/O뿐
- 둘 다 **스트리밍 fd와 별개의 새 fd(`os.Open(name)`)로 풀읽기** → 커서·카운터·타이머 불간섭
- `ReadFile.Hash()`: 원본 불변 → `sync.Once` 캐시, 락 불필요 / `SaveFile.Hash()`: 수신중 계속 변함 → **캐시 금지**, "완료 후 1회", `s.mu` 잡음. WriteAt 비순차/재전송/이어받기는 running hash로 못 다룸 → 결과 풀읽기가 정답
- `SaveFile.Sync()` = rename 직전 fsync

## reader 추상화 + 메모리 전송 (2026-06-29)
- **`transfer.Reader` 인터페이스**(`reader.go`) = 순수 바이트 스트림 계약: `Size/Hash/SeekTo/Read/Close/SetOnClose`. **메타(name/perm)는 일부러 제외** — 그건 전송 세션(`sendJob`)이 소유하고 각 래퍼가 구체 reader에서 뽑아 넘긴다
- **왜 `SetOnClose`만 메서드?** `OnClose`는 *필드*라 인터페이스로 노출 불가 + 정리 콜백은 `transferId`를 캡처해야 하는데 transferId는 `send()` 안에서 생성됨(구체 래퍼 시점엔 아직 없음) → 세터로 우회. 기존 OnClose 자기제거 메커니즘은 그대로(끊김정리 무변경)
- **`ReadFile`**: 무변경 + `SetOnClose` 추가. `Info()`/`Perm()` 유지(`SendFile`이 메타 도출에 사용)
- **`ReadBuffer`**(신규 `readBuffer.go`): 메모리 전용. `NewReadBuffer(data, expire)` — **name/perm 안 가짐**(순수 스트림). 해시는 `sha256.Sum256(data)` 한 방(IO에러 없어 err 항상 nil, 시그니처만 `(string,error)` 유지). `Read`는 슬라이스 새 버퍼 복사. `Remove()` 없음(파일 전용 개념)
- **송신부 분리**: `fileSend`→**`sendJob{authKey,name,perm,reader Reader,cancel}`**. `SendFile(*ReadFile)`=파일 메타에서 name/perm 도출 / `SendBuffer(*ReadBuffer, name, perm)`=호출자가 명시 → 둘 다 공유 코어 **`send(authKey,name,perm,reader)`**로 합류. `sendOnce`도 name/perm을 인자로 받음(reader는 스트림 전용)
- **용도**: EDIT seed처럼 디스크 파일이 아닌 내용을 파일 거치지 않고 그대로 전송. `go build`/`vet` 통과

## 송신(SendFile/SendBuffer→send) / 수신(핸들러) 패턴
- 송신 래퍼 `sendJob{authKey,name,perm,reader,cancel}`(name=destPath/perm는 세션 소유) / 수신 래퍼 `fileRecv{save,hash,finalPath}` (FileResult 검증에 init의 hash·경로 필요한데 ResultReq엔 transferId만 실림)
- 재시도: `maxSendAttempts=3` + backoff. 거부(errRejected)=즉시중단 / 그 외=재시도 / 초과=error
- 이어받기: 받는 쪽이 `.part` stat으로 resumeOffset 결정 → 송신자 `SeekTo` 후 재개. 같은 transferId·destPath라 재부팅 후도 결정적
- 수신 완성: **Completed(크기)→Hash일치→Sync→Close→Rename(.part→최종)**. 크기미달=part보존+resumable / hash불일치=part삭제+not resumable
- traversal 차단: `Clean("/"+destPath)`로 선행 `..` 무력화 → `filepath.Rel(root,final)`로 루트 하위 재확인
- 끊김 정리: conn 소속만 `FindAll(authKey==key)`로 reader Close (공유변수 X)

## abort (cancel ctx 전파)
- `SendFile`이 `WithCancel(Background)` ctx 생성→`fileSend.cancel` 저장, **per-call timeout을 이 ctx에서 파생**(Background 아님)해야 cancel이 진행 중 Call로 전파됨. 루프에 `ctx.Err()` 체크로 재시도 차단
- `AbortFile`: cancel() + worker에 MsgFileAbort(best-effort). reader 정리는 SendFile defer Close가 전담(abort/정상/끊김 한 경로)

## context.cancel 교훈
- `cancel()`=ctx 타이머·parent등록 해제(연결과 무관). 안 불러도 Background자식은 timeout시 자동정리되나 **그때까지 쌓임** → **루프에선 defer 금지, Call 직후 즉시 cancel**(defer는 함수 끝까지 안 풀려 타이머 수천 개 누적)

## 관련 파일
- `internal/transfer/reader.go` — `Reader` 인터페이스(스트림 계약)
- `internal/transfer/{readFile,readBuffer,saveFile}.go` — 파일/메모리 reader + Hash()/Sync()
- `cmd/supervisor/router/transfer.go` — SendFile/SendBuffer/send/sendOnce/AbortFile + sendJob
- `cmd/supervisor/router/supervisorRouter.go` — readers 맵(`*sendJob`) + 끊김정리
- `cmd/worker/router/workerRouter.go` — fileInit/fileChunk/fileResult/fileAbort + fileRecv

## 잔여 (마무리)
1. **e2e 스모크 테스트** — supervisor+worker 띄워 SHA256SUMS 전송 → 이어받기(중간 끊고 재접속) → abort 확인
2. register 안의 임시 테스트 전송을 **밖으로 분리**(등록 완료 후 트리거) → 타이밍 레이스(subKey 미설정 시 폴더명 mainKey만) 해소
3. abort로 끝난 SendFile 에러를 정상중단으로 구분할 sentinel 필요해지면 추가(현재는 과해서 보류)
