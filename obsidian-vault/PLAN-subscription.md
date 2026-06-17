# 구독 / 토픽 모델 설계 (계승 reference)

> 출처: PortBridge WebSocket 구독 모델. Nexus에서 supervisor가 다수 worker agent의
> 상태/출력을 selective하게 감시할 때 재활용 후보.

## 배경
- 소켓으로 모든 데이터를 broadcast하면 동시성/불필요 전송 문제
- agent 상태, process 출력 등을 topic 기반 구독/해제로 관리

## 구독 등록/해제
- **REST**로 구독 등록/해제, **WS**는 서버→클라이언트 단방향 push 전용 (권장안)
- (대안: WS 메시지로 SUBSCRIBE/UNSUBSCRIBE — PortBridge 실제 구현은 이 방식)
```
POST   /subscriptions        — 구독 등록   { "topic": "agents:{id}" }
DELETE /subscriptions/:topic — 구독 해제
GET    /subscriptions        — 구독 목록 (디버깅)
```

## Topic 구조 (Nexus 예시)
```
agents               — 전체 agent 목록/연결 변경
agents:{agentId}     — 특정 agent 상태
processes            — 전체 process 목록
processes:{id}       — 특정 process 출력/상태
```

## WS 메시지 포맷 (서버 → 클라이언트)
```json
{ "type": "SNAPSHOT", "topic": "agents", "data": [...] }
{ "type": "UPDATE",   "topic": "agents:42", "data": {...} }
```
- 구독 시 즉시 현재 상태 `SNAPSHOT`, 이후 변경분만 `UPDATE`

## Go 서버 구조
```go
type TopicManager struct {
    mu     sync.RWMutex
    topics map[string]map[int64]struct{} // topic → subscriberIdSet
}
func (tm *TopicManager) Subscribe(topic string, id int64)
func (tm *TopicManager) Unsubscribe(topic string, id int64)
func (tm *TopicManager) UnsubscribeAll(id int64)
func (tm *TopicManager) Publish(topic string, msg any)
```
- DB 변경 후 서비스 레이어에서 `Publish(topic, data)` 한 줄
- Publish 시 해당 구독자의 WS 클라이언트로 전송 (wsManager 재활용)

## 필수 주의사항
- **WS 해제 시 `UnsubscribeAll` 필수** (재연결 중복 구독 / dead client publish 방지)
- **SNAPSHOT 타이밍**: 구독 등록 → DB 조회 순서 보장 (등록 전 변경 누락 방지)
- **동시성 최적화**: Lock 구간 분리 — Lock에서 구독자/클라이언트 포인터만 수집,
  Unlock 후 json.Marshal + WriteMessage(느린 I/O)
- **topic 세분화 폭발 주의**: `processes:{id}` 가 수백 개면 구독 수 폭증
- 메시지는 BinaryMessage로 통일 (TextMessage 혼용 금지)
