// Package execute는 로컬/원격 process를 동일하게 다루는 실행 추상화의 **이식 가능한** 부분이다(공유).
//
// 보유: IInteractive(인터페이스) / AgentInteractive(원격 래퍼) / CommandStatus / ExecCommand / Tmux.
// 로컬 PTY 실행(Linux 전용 syscall)은 하위 패키지 execute/pty로 분리되어 있다
// → supervisor 등 비-Linux 빌드가 execute만 import할 수 있다.
// (PLAN-agent-comm.md 참고)
package execute
