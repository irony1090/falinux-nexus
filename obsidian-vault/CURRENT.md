# CURRENT

## 현재 날짜
2026-06-25

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
