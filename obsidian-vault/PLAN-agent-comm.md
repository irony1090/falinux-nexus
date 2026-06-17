# Agent ↔ Server 통신 / Process 실행 설계 (계승 reference)

> 출처: PortBridge의 "Agent Process → BindManager 연동". Nexus의 worker↔supervisor
> 통신 골격으로 일반화. UI/아이콘/경로 분기 등 PortBridge 전용 로직은 제거함.

## 핵심 추상화: IInteractive
하나의 인터페이스로 로컬 PTY와 원격 agent process를 동일하게 다룬다.
```go
type IInteractive interface {
    Output() ([]byte, error)
    Write(data []byte) error
    Status() (CommandStatus, int, error)
    Kill() error
    Layout(cols, rows uint16) syscall.Errno
    ExitCode() int  // 필드가 아닌 메서드
}
```
- `*Interactive` — 로컬 PTY 구체 타입 (worker가 자기 머신에서 실행)
- `*AgentInteractive` — 원격 agent WebSocket 래퍼 (supervisor가 worker를 원격 구동)

## AgentInteractive (원격 process 래퍼)
- 콜백 주입: `onWrite(data)`, `onLayout(cols,rows)`, `onKill()` → 원격에 명령 전달
- 외부 주입: `PushOutput(data)`, `PushStatus(status)`, `Done(exitCode)`
- `sync.Once`로 `Done()` 중복 방지
- 생성 시 status push 없음 → RUNNING 수신 시 첫 push (블로킹 동기화에 활용)
- `SignalDone(exitCode)`: exitCode==0 → Completed, 그 외 → Failed, 채널 Close

## 프로토콜 메시지
| 방향 | 메시지 | 용도 |
|------|--------|------|
| Server→Agent | `MsgExec` | 실행 명령 (id, cmd) |
| Server→Agent | `MsgData` | 입력(stdin) 전달 |
| Server→Agent | `MsgResize` | 터미널 크기 변경 |
| Server→Agent | `MsgKill` | 종료 |
| Server→Agent | `MsgUploadInit/Chunk` | 파일 전송 |
| Agent→Server | `MsgStatus` | RUNNING/STOPPED + ExitCode |
| Agent→Server | `MsgData` | 출력(stdout) 스트리밍 |

- key 형식: `agentId:processId` (멀티유저 시 `agentId:userId:processId`)
- 파일 전송: `UploadInit → Ready → Chunk → Status → Result`

## 실행 흐름
```
CreateProcess(agentId, cmd, processId)
  → AgentInteractive 생성 + 등록(key)
  → MsgExec 전송 → 원격 agent 실행
  → MsgStatus(RUNNING) → PushStatus → Status() 블로킹 해제 → RUNNING 반영
  → bindManager 등록 → status 감시 goroutine
  → MsgData(output) → PushOutput → 상위/UI 스트리밍
  → MsgStatus(STOPPED, exitCode) → SignalDone → goroutine 종료 + key 제거
```

## Supervisor 측 관리
- agent 연결/해제 콜백(`OnCreated`/`OnRemoved`) → 상태 갱신 + 브로드캐스트
- agent 끊김 시 defer cleanup: 해당 agent 소속 모든 interactive `Done(502)` 처리
- agent 상태는 `agentId(string)` 기반 싱글턴 (DB 조회 없이 매핑)

## 주의사항 (이전 프로젝트에서 겪은 것)
- `/agents` 라우트 그룹을 `/processes`보다 먼저 선언 (executor 주입 순서)
- `onKill` 콜백 nil 체크
- `Status()` 반환 시 `err != nil` 체크 누락 주의 (PortBridge 미해결 버그였음)
- ExitCode는 필드 직접 접근이 아니라 메서드로 (인터페이스 일관성)

## Nexus에서 재결정할 것
- worker 식별자 체계 (PortBridge는 MAC 주소를 agentId로 사용)
- 2단계 고정 vs N단계 트리 (supervisor가 supervisor를 관리하는 재귀 구조?)
- 명령 분배/라우팅 정책 (어느 worker에 process를 띄울지)
