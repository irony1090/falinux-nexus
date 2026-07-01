package syncProcess

import (
	"io"
	"sync"
)

type SyncData[T any] struct {
	data   []T
	mutex  sync.Mutex
	signal *sync.Cond
	last   T
	closed bool
}

func NewSyncData[T any](data []T) *SyncData[T] {
	sd := &SyncData[T]{
		data: data,
	}
	sd.signal = sync.NewCond(&sd.mutex)
	return sd
}

// Push 데이터 추가. 대기 중인 고루틴이 있으면 깨움
func (q *SyncData[T]) Push(data T) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if q.closed {
		return io.EOF
	}

	q.data = append(q.data, data)
	q.last = data
	q.signal.Signal() // 대기 중인 고루틴 하나를 깨움
	return nil
}

// func (q *SyncData[T]) Push2(data T) error {
// 	if v, ok := any(data).([]byte); ok {
// 		log.Printf("\t[PUSH] %d (%d)", len(v), len(q.data))
// 		defer log.Printf("\t[PUSH] %d END", len(v))
// 	}
// 	return q.Push(data)
// }

// Shift 첫 번째 요소를 꺼내 반환. 데이터가 없으면 데이터가 올 때까지 대기
func (q *SyncData[T]) Shift() (T, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// data가 없으면 데이터가 올 때까지 대기
	for len(q.data) == 0 {
		if q.closed {
			return q.last, io.EOF
		}
		q.signal.Wait()
	}

	// 첫 번째 요소를 꺼냄
	result := q.data[0]
	q.data = q.data[1:]

	return result, nil
}

// ShiftAll 대기 중인 모든 요소를 한 번에 꺼내 반환. 데이터가 없으면 최소 1건이 올 때까지 대기.
// Shift와 종료 의미 동일: 닫혔고 남은 데이터가 없으면 (nil, io.EOF), 닫혔더라도 남은 데이터가
// 있으면 그것을 먼저 반환(err=nil)한다. 고빈도 스트림에서 항목당 락을 잡는 Shift 대신
// 한 번의 락으로 배치 드레인해 락 경합을 줄인다.
func (q *SyncData[T]) ShiftAll() ([]T, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	// data가 없으면 최소 1건이 올 때까지 대기
	for len(q.data) == 0 {
		if q.closed {
			return nil, io.EOF
		}
		q.signal.Wait()
	}

	// 대기분 전체를 넘기고 백킹 배열을 새로 분리(q.data[1:] 방식의 head 잔존 회피)
	result := q.data
	q.data = make([]T, 0)
	return result, nil
}

// Close 큐를 닫음. 이후 Push/Shift는 오류 반환
func (q *SyncData[T]) Close() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.closed = true
	q.signal.Broadcast() // 모든 대기 중인 고루틴을 깨움
}

func (q *SyncData[T]) Last() T {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.last
}

func (q *SyncData[T]) Init() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.closed = false
	q.data = q.data[:0]
}
