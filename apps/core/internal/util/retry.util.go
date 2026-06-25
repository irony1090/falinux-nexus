package util

import "time"

type Backoff struct {
	cur, base, max time.Duration
}

func NewBackoff(cur, base, max time.Duration) *Backoff {
	return &Backoff{cur, base, max}
}

func (b *Backoff) Next() time.Duration {
	d := b.cur
	if d == 0 {
		d = b.base
	}

	b.cur = min(d*2, b.max)
	return d
}

func (b *Backoff) Reset() { b.cur = 0 }
