// Package protocol은 supervisor와 worker가 공유하는 메시지/패킷 정의의 단일 진실이다.
//
// 양쪽이 이 패키지를 import 하므로 메시지 타입은 여기에만 정의한다.
// 계획: MsgExec / MsgData / MsgStatus / MsgKill / MsgResize (PLAN-agent-comm.md 참고)
package protocol
