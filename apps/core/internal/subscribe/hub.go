// Package subscribe는 "문자열 키 → 구독 클라이언트 집합" 장부를 관리하는 재사용 코어다.
//
// 허브(Hub)는 키 생성 로직과 도메인 타입을 모른다. 키는 호출자가 문자열로 만들어 넘기고,
// 직렬화(marshal)와 전송(send) 전략은 생성 시 1회 주입한다.
// 클라이언트 타입은 제네릭 [C comparable]로, *Client 포인터든 인터페이스든 받을 수 있다.
package subscribe

import (
	"errors"
	"sync"
)

// topic은 하나의 키(토픽)에 구독 중인 클라이언트 집합이다.
type topic[C comparable] struct {
	mu      sync.Mutex
	clients []C
}

func newTopic[C comparable]() *topic[C] {
	return &topic[C]{clients: make([]C, 0)}
}

func (s *topic[C]) hasUnsafe(client C) bool {
	for _, c := range s.clients {
		if c == client {
			return true
		}
	}
	return false
}

// append는 새로 추가됐으면 true, 이미 있으면 false를 반환한다.
func (s *topic[C]) append(client C) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.hasUnsafe(client) {
		return false
	}
	s.clients = append(s.clients, client)
	return true
}

// delete는 삭제 후 남은 구독자 수를 반환한다. 없던 클라이언트면 -1.
func (s *topic[C]) delete(client C) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, c := range s.clients {
		if c == client {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			return len(s.clients)
		}
	}
	return -1
}

// snapshot은 전송용 복사본을 반환한다(락 밖 전송을 위해).
func (s *topic[C]) snapshot() []C {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]C, len(s.clients))
	copy(out, s.clients)
	return out
}

// Hub는 키별 구독자(토픽)를 관리한다. 동시성 안전.
type Hub[C comparable, T comparable] struct {
	mu         sync.Mutex
	records    map[string]*topic[C] // key → 토픽(구독자 집합)
	keyRecords map[C][]string       // client → 구독 중인 키들

	marshal func(any) ([]byte, error) // 직렬화 전략 (Publish에서 사용)
	send    func(C, T, []byte) error  // 전송 전략   (Publish에서 사용)
}

// New는 직렬화/전송 전략을 주입받아 허브를 만든다.
// Subscribers만 쓰고 Publish를 쓰지 않을 거라면 marshal/send에 nil을 줘도 된다.
func New[C comparable, T comparable](
	marshal func(any) ([]byte, error),
	send func(C, T, []byte) error,
) *Hub[C, T] {
	return &Hub[C, T]{
		records:    make(map[string]*topic[C]),
		keyRecords: make(map[C][]string),
		marshal:    marshal,
		send:       send,
	}
}

// getOrCreateUnsafe는 m.mu 보유 상태에서 호출해야 한다.
func (m *Hub[C, T]) getOrCreateUnsafe(key string, create bool) *topic[C] {
	if sc, ok := m.records[key]; ok {
		return sc
	}
	if !create {
		return nil
	}
	sc := newTopic[C]()
	m.records[key] = sc
	return sc
}

// Subscribe는 client를 key에 구독시킨다.
func (m *Hub[C, T]) Subscribe(key string, client C) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sc := m.getOrCreateUnsafe(key, true)
	if sc.append(client) {
		m.keyRecords[client] = append(m.keyRecords[client], key)
	}
}

// Unsubscribe는 client를 key 구독에서 해제한다.
func (m *Hub[C, T]) Unsubscribe(key string, client C) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.unsubscribeUnsafe(key, client)
}

// UnsubscribeAll은 client의 모든 구독을 해제한다(연결 종료 시 호출).
func (m *Hub[C, T]) UnsubscribeAll(client C) {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := append([]string(nil), m.keyRecords[client]...) // 순회 중 변경되므로 복사
	for _, k := range keys {
		m.unsubscribeUnsafe(k, client)
	}
}

// unsubscribeUnsafe는 m.mu 보유 상태에서 호출해야 한다.
func (m *Hub[C, T]) unsubscribeUnsafe(key string, client C) {
	sc := m.getOrCreateUnsafe(key, false)
	if sc == nil {
		return
	}
	if sc.delete(client) == 0 { // 구독자가 0이면 키 자체 제거
		delete(m.records, key)
	}

	keys, ok := m.keyRecords[client]
	if !ok {
		return
	}
	for i, k := range keys {
		if k == key {
			m.keyRecords[client] = append(keys[:i], keys[i+1:]...)
			if len(m.keyRecords[client]) == 0 {
				delete(m.keyRecords, client)
			}
			break
		}
	}
}

// Subscribers는 key 구독자들의 스냅샷(복사본)을 반환한다.
// 직접 전송 루프를 짜고 싶은 특수 케이스용 탈출구.
func (m *Hub[C, T]) Subscribers(key string) []C {
	m.mu.Lock()
	defer m.mu.Unlock()
	sc, ok := m.records[key]
	if !ok {
		return nil
	}
	return sc.snapshot()
}

// Publish는 payload를 1번 marshal한 뒤 key 구독자들에게 전송한다.
// 전송은 락 밖에서 수행하며, 특정 클라이언트 전송 실패가 나머지를 막지 않는다.
// 발생한 send 에러들은 모아서 함께 반환한다.
func (m *Hub[C, T]) Publish(key string, Kind T, payload any) error {
	clients := m.Subscribers(key)
	if len(clients) == 0 {
		return nil
	}
	data, err := m.marshal(payload)
	if err != nil {
		return err
	}
	var errs []error
	for _, c := range clients {
		if e := m.send(c, Kind, data); e != nil {
			errs = append(errs, e)
		}
	}
	return errors.Join(errs...)
}
