// Package ring은 고정 크기 바이트 ring buffer다. 넘치면 오래된 바이트부터 밀려난다.
// 누적 offset(지금까지 쓴 총 바이트)을 함께 들고 있어, 스냅샷과 그 뒤 실시간 스트림의 이음매를 offset으로 맞출 수 있다.
package ring

// Ring은 동시성 안전이 아니다 — 호출자가 잠근다.
type Ring struct {
	buf  []byte
	head int   // 다음에 쓸 위치
	size int   // 채워진 양(<= len(buf))
	off  int64 // 지금까지 쓴 총 바이트
}

func New(capacity int) *Ring {
	return &Ring{buf: make([]byte, capacity)}
}

// Write는 p를 쓰고 쓴 뒤의 누적 offset을 반환한다.
func (r *Ring) Write(p []byte) int64 {
	r.off += int64(len(p))
	n := len(r.buf)
	if len(p) >= n {
		copy(r.buf, p[len(p)-n:])
		r.head, r.size = 0, n
		return r.off
	}
	c := copy(r.buf[r.head:], p)
	copy(r.buf, p[c:])
	r.head = (r.head + len(p)) % n
	r.size = min(r.size+len(p), n)
	return r.off
}

// Snapshot은 남아 있는 바이트(오래된 것부터, 복사본)와 그 끝의 누적 offset을 반환한다.
// off > len(data)면 앞부분이 밀려난 것이다.
func (r *Ring) Snapshot() ([]byte, int64) {
	out := make([]byte, r.size)
	start := (r.head - r.size + len(r.buf)) % len(r.buf)
	c := copy(out, r.buf[start:min(start+r.size, len(r.buf))])
	copy(out[c:], r.buf[:r.size-c])
	return out, r.off
}

// Off는 지금까지 쓴 누적 바이트를 반환한다.
func (r *Ring) Off() int64 { return r.off }
