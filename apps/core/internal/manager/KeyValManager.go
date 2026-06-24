package manager

import (
	"sync"
)

type KeyValManager[K comparable, V any] struct {
	record map[K]V

	OnCreated func(K, V)
	OnRemoved func(K, V)

	mu sync.Mutex
}

type FindResult[K comparable, V any] struct {
	Key K
	Val V
}

func NewKeyValManager[K comparable, V any]() *KeyValManager[K, V] {
	m := &KeyValManager[K, V]{
		record: make(map[K]V),
	}
	return m
}

func (m *KeyValManager[K, V]) Get(key K) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.record[key]
	return v, ok
}

func (m *KeyValManager[K, V]) FindAll(predicate func(K, V) bool) []FindResult[K, V] {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]FindResult[K, V], 0)

	for k, v := range m.record {
		if predicate(k, v) {
			result = append(result, FindResult[K, V]{
				Key: k,
				Val: v,
			})
		}
	}

	return result

}

func (m *KeyValManager[K, V]) Append(
	key K,
	val V,
) bool {
	flag := true
	defer func() {
		if flag && m.OnCreated != nil {
			m.OnCreated(key, val)
		}
	}()
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.record[key]; ok {
		flag = false
		return flag
	}
	m.record[key] = val

	return flag
}

func (m *KeyValManager[K, V]) Remove(key K) bool {
	flag := false
	var exist V
	defer func() {
		if flag && m.OnRemoved != nil {
			m.OnRemoved(key, exist)
		}
	}()
	m.mu.Lock()
	defer m.mu.Unlock()
	if el, ok := m.record[key]; ok {
		delete(m.record, key)
		flag = true
		exist = el
		return flag
	}
	return flag
}
