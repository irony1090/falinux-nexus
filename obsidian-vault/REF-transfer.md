# REF — 파일 전송 모듈 (internal/transfer)

> supervisor→worker 단일 파일 전송. FileInit→FileChunk×N→FileResult(+FileAbort). 송신/수신 대칭.
> **구현 완료, e2e 미검증.** 상세 작업 이력은 `history/transfer.md`.

## 무결성 해시 (`transfer.Hash()`)
- Hash = **원본 전체 sha256(hex 64자)**. 출력은 입력 크기 무관 고정(32B) → 와이어 부풀음 無, 비용은 선읽기 I/O뿐
- 둘 다 **스트리밍 fd와 별개의 새 fd(`os.Open(name)`)로 풀읽기** → 커서·카운터·타이머 불간섭
- `ReadFile.Hash()`: 원본 불변 → `sync.Once` 캐시, 락 불필요 / `SaveFile.Hash()`: 수신중 계속 변함 → **캐시 금지**, "완료 후 1회", `s.mu` 잡음. WriteAt 비순차/재전송/이어받기는 running hash로 못 다룸 → 결과 풀읽기가 정답
- `SaveFile.Sync()` = rename 직전 fsync

## 송신(SendFile) / 수신(핸들러) 패턴
- 송신 래퍼 `fileSend{authKey,reader,cancel}` / 수신 래퍼 `fileRecv{save,hash,finalPath}` (FileResult 검증에 init의 hash·경로 필요한데 ResultReq엔 transferId만 실림)
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
- `internal/transfer/{readFile,saveFile}.go` — Hash()/Sync()
- `cmd/supervisor/router/supervisorRouter.go` — SendFile/sendOnce/AbortFile + fileSend
- `cmd/worker/router/workerRouter.go` — fileInit/fileChunk/fileResult/fileAbort + fileRecv

## 잔여 (마무리)
1. **e2e 스모크 테스트** — supervisor+worker 띄워 SHA256SUMS 전송 → 이어받기(중간 끊고 재접속) → abort 확인
2. register 안의 임시 테스트 전송을 **밖으로 분리**(등록 완료 후 트리거) → 타이밍 레이스(subKey 미설정 시 폴더명 mainKey만) 해소
3. abort로 끝난 SendFile 에러를 정상중단으로 구분할 sentinel 필요해지면 추가(현재는 과해서 보류)
