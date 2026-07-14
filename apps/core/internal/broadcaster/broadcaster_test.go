package broadcaster

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestSubscribePushReceive(t *testing.T) {
	b := NewBroadcaster[int](0, 200*time.Millisecond)
	ch, cancel, err := b.Subscribe()
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer cancel()

	go b.Push(42)

	select {
	case v := <-ch:
		if v != 42 {
			t.Fatalf("got %d, want 42", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for value")
	}
}

func TestBroadcastToAllSubscribers(t *testing.T) {
	b := NewBroadcaster[int](0, 200*time.Millisecond)

	const n = 5
	chans := make([]<-chan int, n)
	for i := range n {
		ch, cancel, err := b.Subscribe()
		if err != nil {
			t.Fatalf("subscribe %d: %v", i, err)
		}
		defer cancel()
		chans[i] = ch
	}

	go b.Push(7)

	var wg sync.WaitGroup
	wg.Add(n)
	for _, ch := range chans {
		go func() {
			defer wg.Done()
			select {
			case v := <-ch:
				if v != 7 {
					t.Errorf("got %d, want 7", v)
				}
			case <-time.After(time.Second):
				t.Error("timed out waiting for value")
			}
		}()
	}
	wg.Wait()
}

func TestPushOrderPerSubscriber(t *testing.T) {
	b := NewBroadcaster[int](0, 200*time.Millisecond)
	ch, cancel, err := b.Subscribe()
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer cancel()

	const n = 50
	go func() {
		for i := range n {
			b.Push(i)
		}
	}()

	for i := range n {
		select {
		case v := <-ch:
			if v != i {
				t.Fatalf("out of order: got %d, want %d", v, i)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for value %d", i)
		}
	}
}

func TestCancelClosesChannel(t *testing.T) {
	b := NewBroadcaster[int](1, 200*time.Millisecond)
	ch, cancel, err := b.Subscribe()
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	cancel()

	select {
	case v, ok := <-ch:
		if ok {
			t.Fatalf("expected closed channel, got value %d", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}

// cancel()이 진행 중인(이 채널로의 send를 아직 시도하고 있는) Push를 기다리지 않고
// 즉시 리턴하는지 확인한다. 예전 구현(remove가 pushMu부터 잡음)이었다면 최소 timeout만큼 블로킹했을 것.
func TestCancelDoesNotBlockDuringInFlightPush(t *testing.T) {
	const timeout = 150 * time.Millisecond
	b := NewBroadcaster[int](0, timeout)

	// 절대 읽지 않는 구독자 → Push의 fan-out goroutine이 이 채널에 timeout까지 블로킹된다.
	_, cancelSlow, err := b.Subscribe()
	if err != nil {
		t.Fatalf("subscribe slow: %v", err)
	}

	pushDone := make(chan struct{})
	go func() {
		b.Push(1)
		close(pushDone)
	}()

	// Push가 slow 구독자에 대한 send 시도를 시작할 시간을 준다.
	time.Sleep(10 * time.Millisecond)

	start := time.Now()
	cancelSlow()
	elapsed := time.Since(start)

	if elapsed >= timeout {
		t.Fatalf("cancel took %v, want well under %v (must not block on in-flight Push)", elapsed, timeout)
	}

	// unsubscribe는 mutex만 쓰므로, 실제 close 완료 여부와 무관하게 즉시 반영돼야 한다.
	if got := b.SubscriberCount(); got != 0 {
		t.Fatalf("subscriber count after cancel = %d, want 0", got)
	}

	select {
	case <-pushDone:
	case <-time.After(time.Second):
		t.Fatal("Push did not finish in time")
	}
}

func TestSlowSubscriberRemovedAfterTimeout(t *testing.T) {
	b := NewBroadcaster[int](0, 30*time.Millisecond)

	goodCh, cancelGood, err := b.Subscribe()
	if err != nil {
		t.Fatalf("subscribe good: %v", err)
	}
	defer cancelGood()

	_, cancelBad, err := b.Subscribe() // never read from this one
	if err != nil {
		t.Fatalf("subscribe bad: %v", err)
	}
	defer cancelBad()

	if got := b.SubscriberCount(); got != 2 {
		t.Fatalf("subscriber count before push = %d, want 2", got)
	}

	done := make(chan struct{})
	go func() {
		b.Push(1)
		close(done)
	}()

	select {
	case v := <-goodCh:
		if v != 1 {
			t.Fatalf("got %d, want 1", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for good subscriber to receive")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Push did not return in time")
	}

	if got := b.SubscriberCount(); got != 1 {
		t.Fatalf("subscriber count after timeout = %d, want 1", got)
	}
}

func TestSubscriberCount(t *testing.T) {
	b := NewBroadcaster[int](0, 200*time.Millisecond)

	if got := b.SubscriberCount(); got != 0 {
		t.Fatalf("initial count = %d, want 0", got)
	}

	_, cancel1, _ := b.Subscribe()
	_, cancel2, _ := b.Subscribe()

	if got := b.SubscriberCount(); got != 2 {
		t.Fatalf("count after subscribe = %d, want 2", got)
	}

	cancel1()

	if got := b.SubscriberCount(); got != 1 {
		t.Fatalf("count after cancel = %d, want 1", got)
	}

	cancel2()

	if got := b.SubscriberCount(); got != 0 {
		t.Fatalf("count after cancel all = %d, want 0", got)
	}
}

func TestClose(t *testing.T) {
	b := NewBroadcaster[int](1, 200*time.Millisecond)

	ch1, _, err := b.Subscribe()
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	ch2, _, err := b.Subscribe()
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	b.Close()

	for i, ch := range []<-chan int{ch1, ch2} {
		select {
		case _, ok := <-ch:
			if ok {
				t.Fatalf("subscriber %d: expected closed channel", i)
			}
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d: timed out waiting for close", i)
		}
	}

	if _, _, err := b.Subscribe(); !errors.Is(err, ErrClosed) {
		t.Fatalf("subscribe after close: got err %v, want ErrClosed", err)
	}

	if got := b.Status(); got != StatusClosed {
		t.Fatalf("status after close = %v, want StatusClosed", got)
	}

	done := make(chan struct{})
	go func() {
		b.Push(99) // must be a no-op: no panic, no block
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Push after close did not return")
	}
}

func TestPushNotBlockedBySlowStatusSubscriber(t *testing.T) {
	const timeout = 50 * time.Millisecond
	b := NewBroadcaster[int](0, timeout)

	dataCh, cancelData, err := b.Subscribe()
	if err != nil {
		t.Fatalf("subscribe data: %v", err)
	}
	defer cancelData()
	go func() {
		for range dataCh {
		}
	}()

	// status 구독자는 만들되 절대 읽지 않는다: 내부 statusBC.Push가 매번 timeout을 타야 하는 상황을 만든다.
	_, cancelStatus, err := b.WatchStatus()
	if err != nil {
		t.Fatalf("watch status: %v", err)
	}
	defer cancelStatus()

	start := time.Now()
	b.Push(1)
	elapsed := time.Since(start)

	// 구조 결합이 남아있었다면(Distributing+Pending 동기 fan-out) 최소 2*timeout은 걸렸을 것.
	// 분리됐다면 data 구독자가 즉시 읽으므로 timeout 근처에도 못 미쳐야 한다.
	if elapsed >= timeout {
		t.Fatalf("Push took %v, want well under %v (status subscriber must not block data path)", elapsed, timeout)
	}
}

// consumer가 매 Push의 두 status 이벤트를 즉시 다 소비하고 나서야 다음 Push로 넘어간다.
// dispatcher가 backpressure 없이 항상 따라잡는 이 조건에서는 drop 없이 정확히 교대돼야 한다.
// (반대로 producer가 dispatcher보다 훨씬 빠르면 drop-oldest가 일부러 일부를 버린다 — 별개의 설계 의도.)
func TestStatusOrderPreservedWhenConsumerKeepsUp(t *testing.T) {
	b := NewBroadcaster[int](1, 200*time.Millisecond)

	statusCh, cancelStatus, err := b.WatchStatus()
	if err != nil {
		t.Fatalf("watch status: %v", err)
	}
	defer cancelStatus()

	dataCh, cancelData, err := b.Subscribe()
	if err != nil {
		t.Fatalf("subscribe data: %v", err)
	}
	defer cancelData()
	go func() {
		for range dataCh {
		}
	}()

	recv := func(want BroadcasterStatus) {
		t.Helper()
		select {
		case s := <-statusCh:
			if s != want {
				t.Fatalf("got %v, want %v (order broken)", s, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for %v", want)
		}
	}

	const n = 20
	for range n {
		b.Push(1)
		recv(StatusDistributing)
		recv(StatusPending)
	}
}
