// Package registry는 supervisor에 연결된 worker agent들의 연결/상태를 관리한다.
//
// 계획: OnCreated/OnRemoved 콜백으로 상태 갱신 + 브로드캐스트 (PLAN-agent-comm.md 참고)
package registry
