// Package execute는 로컬 PTY와 원격 agent process를 동일하게 다루는 실행 추상화를 제공한다(공유).
//
// 계획: IInteractive 인터페이스 + Interactive(PTY) / AgentInteractive(원격) 구현
// (PLAN-agent-comm.md 참고)
package execute
