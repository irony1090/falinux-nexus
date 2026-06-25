// Package subscribe(supervisor)는 supervisor 도메인 전용 구독 계층이다.
//
// 재사용 코어(nexus/internal/subscribe.Hub[C])를 사용해, 도메인 키 구조체와
// 그 Key() string, 클라이언트 타입 확정(예: *transport.Client), 직렬화/전송 전략
// 주입을 담당한다. (PLAN-subscription.md 참고)
package subscribe
