package ring

import (
	"bytes"
	"testing"
)

func TestWriteSnapshot(t *testing.T) {
	cases := []struct {
		name   string
		writes []string
		want   string
		off    int64
	}{
		{"empty", nil, "", 0},
		{"under", []string{"ab", "cd"}, "abcd", 4},
		{"exact", []string{"abc", "de"}, "abcde", 5},
		{"wrap", []string{"abc", "def"}, "bcdef", 6},
		{"wrap twice", []string{"abc", "def", "ghij"}, "fghij", 10},
		{"larger than cap", []string{"ab", "0123456789"}, "56789", 12},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := New(5)
			for _, w := range tc.writes {
				r.Write([]byte(w))
			}
			data, off := r.Snapshot()
			if !bytes.Equal(data, []byte(tc.want)) || off != tc.off {
				t.Fatalf("got (%q, %d), want (%q, %d)", data, off, tc.want, tc.off)
			}
		})
	}
}

func TestWriteReturnsOffset(t *testing.T) {
	r := New(4)
	if off := r.Write([]byte("abc")); off != 3 {
		t.Fatalf("off = %d, want 3", off)
	}
	if off := r.Write([]byte("defgh")); off != 8 {
		t.Fatalf("off = %d, want 8", off)
	}
}
