package syncProcess

import (
	"math"
	"sync"
)

type LimitArray[T any] struct {
	arr []T
	max int
	mu  sync.Mutex
}

func NewLimitArray[T any](max int) *LimitArray[T] {
	return &LimitArray[T]{
		arr: make([]T, 0),
		max: max,
	}
}

func (l *LimitArray[T]) Push(data T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// log.Printf("[FRAME] PUSH %d", len(l.arr))
	if len(l.arr) >= l.max {
		start := len(l.arr) - l.max + 1
		start = int(math.Max(float64(start), float64(0)))
		l.arr = l.arr[start:]
	}
	l.arr = append(l.arr, data)
}

func (l *LimitArray[T]) SetMax(max int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.max = max
}

func (l *LimitArray[T]) Get() []T {
	l.mu.Lock()
	defer l.mu.Unlock()
	arr := make([]T, len(l.arr))
	for _, v := range l.arr {
		arr = append(arr, v)
	}
	return arr
}
