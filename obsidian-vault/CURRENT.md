# CURRENT

## 현재 날짜
2026-06-24

## 직전 완료: 파일 전송 모듈 (supervisor → worker) — 구현 완료 / e2e 미검증
- 송신/수신 한 바퀴 + 이어받기(resume) + sha256 무결성 검증 + 재시도(3회) + abort 전부 구현
- `go build ./...` / `go vet` 통과. **2프로세스 실제 전송·이어받기·abort 스모크는 아직 안 돌림.**
- 상세는 HISTORY 2026-06-24 참조. 관련 파일:
  - `internal/transfer/{readFile,saveFile}.go` — Hash()/Sync()
  - `cmd/supervisor/router/supervisorRouter.go` — SendFile/sendOnce/AbortFile + fileSend
  - `cmd/worker/router/workerRouter.go` — fileInit/fileChunk/fileResult/fileAbort + fileRecv

## 다음 작업 (내일~): process 실행 모듈
- worker agent의 핵심 = 자기 프로세스 실행/모니터링/종료 (PTY)
- 계승 자산(MEMORY): `IInteractive` 인터페이스, `*Interactive`(로컬 PTY)/`*AgentInteractive`(원격 래퍼), 실행 흐름(ExecPayload→RUNNING→Done(exitCode))
- 메시지 어휘 확장 필요: `MsgExec`/`MsgData`(input·output)/`MsgResize`/`MsgKill`/`MsgStatus`(RUNNING/STOPPED+ExitCode)
- **출력 스트리밍은 응답 없는 단방향** → 현재 REQ/RES 평면으로는 부적합. **EVENT(Kind 추가)+On/Emit 별도 평면 재도입** 검토 (전에 만들었다 예제 최소화로 제거, git 히스토리에 있음)

## 전송 모듈 마무리 잔여 (process 작업 전/중 정리)
1. **e2e 스모크 테스트** — supervisor+worker 띄워 SHA256SUMS 전송 → 이어받기(중간 끊고 재접속) → abort 확인
2. register 안의 임시 테스트 전송을 **밖으로 분리**(등록 완료 후 트리거) → 타이밍 레이스(subKey 미설정 시 폴더명 mainKey만) 해소
3. abort로 끝난 SendFile 에러를 정상중단으로 구분할 sentinel 필요해지면 추가(현재는 과해서 보류)

## 미해결 이슈 (이월)
- **supervisorRouter register**: `Append` 반환 무시(원자적 점유 미확인) — cleanup은 conn 기준 FindAll로 이미 전환됨
- **서브키 충돌/위조**: seq→RandomKey로 충돌 완화, key↔subkey 결속 검증은 미구현
- **supervisor 영속성**: registry 메모리 → PG 미착수 (전송/실행 안정화 후)
- (이후) 도메인 구독 계층, apps/web
