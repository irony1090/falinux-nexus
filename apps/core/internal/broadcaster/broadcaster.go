package broadcaster

import (
	"errors"
	"sync"
	"time"
)

var ErrClosed = errors.New("Broadcaster: closed")

// statusQueueCap은 status 갱신을 data 경로에서 분리하는 내부 큐 용량이다.
// status는 "현재 상태 스냅샷"이라 꽉 찼을 때 오래된 값을 버려도 무해하다(drop-oldest).
const statusQueueCap = 4

type Broadcaster[C any] struct {
	dests      []chan C
	mutex      sync.RWMutex
	pushMu     sync.Mutex
	status     BroadcasterStatus
	timeout    time.Duration
	bufferSize int

	statusBC *Broadcaster[BroadcasterStatus]
	statusCh chan BroadcasterStatus
	statusWg sync.WaitGroup
}

// newBroadcaster는 statusBC를 연결하지 않은 순수한 인스턴스를 만든다.
// 내부 status broadcaster 자신을 만들 때도 재사용해서 무한 재귀를 막는다.
func newBroadcaster[C any](bufferSize int, timeout time.Duration) *Broadcaster[C] {
	return &Broadcaster[C]{
		dests:      make([]chan C, 0),
		timeout:    timeout,
		bufferSize: bufferSize,
	}
}

func NewBroadcaster[C any](bufferSize int, timeout time.Duration) *Broadcaster[C] {
	b := newBroadcaster[C](bufferSize, timeout)
	b.statusBC = newBroadcaster[BroadcasterStatus](bufferSize, timeout)
	b.statusCh = make(chan BroadcasterStatus, statusQueueCap)

	// dispatcher: 실제(블로킹 가능한) statusBC.Push 호출을 data 경로 밖에서 순서대로 처리한다.
	b.statusWg.Go(func() {
		for s := range b.statusCh {
			b.statusBC.Push(s)
		}
	})

	return b
}

// enqueueStatus는 절대 블로킹하지 않는다: 큐가 차 있으면 가장 오래된 값을 버리고 넣는다.
// 단일 생산자(pushMu로 직렬화된 Push 호출)·단일 소비자(dispatcher) 전제로 락 없이 안전하다.
func (b *Broadcaster[C]) enqueueStatus(s BroadcasterStatus) {
	if b.statusCh == nil {
		return
	}
	for {
		select {
		case b.statusCh <- s:
			return
		default:
		}
		select {
		case <-b.statusCh:
		default:
		}
	}
}

// WatchStatus는 상태가 전이될 때마다 발생 순서 그대로 구독한다.
func (b *Broadcaster[C]) WatchStatus() (<-chan BroadcasterStatus, func(), error) {
	return b.statusBC.Subscribe()
}

func (b *Broadcaster[C]) Subscribe() (<-chan C, func(), error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.status == StatusClosed {
		return nil, nil, ErrClosed
	}

	ch := make(chan C, b.bufferSize)
	b.dests = append(b.dests, ch)

	// cancel은 블로킹하지 않는다: dests 제거(불변식을 즉시 성립시킴)는 그 자리에서 끝내고,
	// 실제 채널 close만 진행 중일 수 있는 Push와 안전하게 순서를 맞추기 위해 백그라운드로 미룬다.
	cancel := func() {
		if b.unsubscribe(ch) {
			go b.closeDest(ch)
		}
	}

	return ch, cancel, nil
}

// unsubscribe는 dest를 dests 목록에서 즉시 제거한다(mutex만 사용, pushMu 불필요 → 논블로킹).
// 이 순간부터 이후에 시작되는 모든 Push는 이 채널을 절대 fan-out 대상으로 보지 않는다.
// 단, 이미 스냅샷을 뜬 진행 중인 Push는 여전히 이 채널에 send를 시도하고 있을 수 있으므로
// 실제 close는 여기서 하지 않는다(closeDest 참고).
func (b *Broadcaster[C]) unsubscribe(dest chan C) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for i, d := range b.dests {
		if d == dest {
			b.dests = append(b.dests[:i], b.dests[i+1:]...)
			return true
		}
	}
	return false
}

// closeDest는 pushMu를 잡은 뒤 채널을 닫는다 — 진행 중일 수 있는 Push의 fan-out
// goroutine과 절대 경쟁하지 않기 위함이다. cancel() 호출자를 막지 않도록 항상
// 별도 goroutine에서 호출한다(Subscribe의 cancel 참고). Push 내부(이미 pushMu 보유)에서는
// 이 함수 대신 직접 close(dest)를 호출한다 — 재진입 락은 데드락을 일으킨다.
func (b *Broadcaster[C]) closeDest(dest chan C) {
	b.pushMu.Lock()
	defer b.pushMu.Unlock()
	close(dest)
}

func (b *Broadcaster[C]) Push(data C) {
	b.pushMu.Lock()
	defer b.pushMu.Unlock()

	b.mutex.Lock()
	if b.status == StatusClosed {
		b.mutex.Unlock()
		return
	}
	b.status = StatusDistributing
	dests := make([]chan C, len(b.dests))
	copy(dests, b.dests)
	b.mutex.Unlock()

	b.enqueueStatus(StatusDistributing)

	var wg sync.WaitGroup
	var staleMu sync.Mutex
	stale := make([]chan C, 0)

	for _, dest := range dests {
		select {
		case dest <- data:
			continue // 버퍼 여유(또는 대기 중인 수신자) → goroutine/timer 없이 즉시 완료
		default:
		}
		wg.Add(1)
		go func(d chan C) {
			defer wg.Done()
			select {
			case d <- data:
			case <-time.After(b.timeout):
				staleMu.Lock()
				stale = append(stale, d)
				staleMu.Unlock()
			}
		}(dest)
	}
	wg.Wait()

	for _, d := range stale {
		if b.unsubscribe(d) {
			close(d) // Push가 이미 pushMu를 쥐고 있으므로 바로 닫아도 안전
		}
	}

	b.mutex.Lock()
	b.status = StatusPending
	b.mutex.Unlock()

	b.enqueueStatus(StatusPending)
}

func (b *Broadcaster[C]) Status() BroadcasterStatus {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.status
}

func (b *Broadcaster[C]) Close() {
	b.pushMu.Lock()
	defer b.pushMu.Unlock()

	b.mutex.Lock()
	if b.status == StatusClosed {
		b.mutex.Unlock()
		return
	}
	for _, d := range b.dests {
		close(d)
	}
	b.dests = nil
	b.status = StatusClosed
	b.mutex.Unlock()

	if b.statusBC != nil {
		close(b.statusCh) // dispatcher가 남은 큐를 비우고 종료할 때까지 대기
		b.statusWg.Wait()
		b.statusBC.Push(StatusClosed)
		b.statusBC.Close()
	}
}

func (b *Broadcaster[C]) SubscriberCount() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return len(b.dests)
}
