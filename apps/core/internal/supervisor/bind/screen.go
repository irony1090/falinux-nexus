package bind

import (
	"bytes"
	"sync"

	"nexus/internal/ring"
)

// alt screen 진입·해제 시퀀스 — worker가 TERM=xterm-256color를 강제하므로 TUI는 smcup/rmcup로 이것을 낸다.
// 파라미터를 합친 형태(`ESC[?1049;25h` 등)는 보지 않는다 — REF-process-snapshot.md S6(화면 종류 판정)
var altSeqs = []struct {
	seq []byte
	alt bool
}{
	{[]byte("\x1b[?1049h"), true}, {[]byte("\x1b[?1049l"), false},
	{[]byte("\x1b[?1047h"), true}, {[]byte("\x1b[?1047l"), false},
	{[]byte("\x1b[?47h"), true}, {[]byte("\x1b[?47l"), false},
}

const altCarry = 7 // 가장 긴 시퀀스 길이 - 1: 묶음 경계에 걸친 시퀀스용

// Screen은 process 하나의 화면 복원 버퍼다(최근 출력 ring + 지금 alt screen인지).
// relay가 Write, 스냅샷 API가 Snapshot — 한 락 안에서 셋(data·off·alt)을 같이 읽어야 서로 어긋나지 않는다.
type Screen struct {
	mu   sync.Mutex
	ring *ring.Ring
	alt  bool
	tail []byte // 직전 묶음 끝 altCarry 바이트
}

func NewScreen(capacity int) *Screen {
	return &Screen{ring: ring.New(capacity)}
}

// Write는 p를 쌓고 쓴 뒤의 누적 offset을 반환한다(DATA 이벤트의 off).
func (s *Screen) Write(p []byte) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	edge := append(append([]byte(nil), s.tail...), p[:min(altCarry, len(p))]...)
	if alt, ok := lastAlt(edge); ok {
		s.alt = alt
	}
	if alt, ok := lastAlt(p); ok { // p 안의 것이 더 나중
		s.alt = alt
	}
	src := edge // p가 짧으면 edge = tail + p 전체
	if len(p) > altCarry {
		src = p
	}
	s.tail = append(s.tail[:0], src[max(0, len(src)-altCarry):]...)
	return s.ring.Write(p)
}

// Snapshot은 남은 출력·그 끝 offset·alt screen 여부를 반환한다.
func (s *Screen) Snapshot() (data []byte, off int64, alt bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, off = s.ring.Snapshot()
	return data, off, s.alt
}

// Off는 지금까지 쓴 누적 바이트를 반환한다(데이터 복사 없이 출력 진행만 볼 때).
func (s *Screen) Off() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ring.Off()
}

// lastAlt는 b에서 가장 나중에 나온 진입·해제 시퀀스의 상태를 반환한다. 없으면 ok=false.
func lastAlt(b []byte) (alt, ok bool) {
	at := -1
	for _, a := range altSeqs {
		if i := bytes.LastIndex(b, a.seq); i > at {
			at, alt = i, a.alt
		}
	}
	return alt, at >= 0
}
