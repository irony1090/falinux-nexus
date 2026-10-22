package ring

import "testing"

func TestScreenAlt(t *testing.T) {
	cases := []struct {
		name   string
		writes []string
		want   bool
	}{
		{"none", []string{"hello\r\n"}, false},
		{"enter", []string{"\x1b[?1049h\x1b[H"}, true},
		{"enter then leave", []string{"\x1b[?1049hdraw", "\x1b[?1049l$ "}, false},
		{"leave then enter in one chunk", []string{"\x1b[?1049l\x1b[?1049h"}, true},
		{"split across chunks", []string{"abc\x1b[?10", "49hdraw"}, true},
		{"split byte by byte", []string{"\x1b", "[", "?", "1", "0", "4", "9", "h"}, true},
		{"split after long chunk", []string{"0123456789\x1b[?1", "049h"}, true},
		{"old form 47", []string{"\x1b[?47h"}, true},
		{"old form 1047 leave", []string{"\x1b[?1047h", "x\x1b[?1047l"}, false},
		{"stays after plain output", []string{"\x1b[?1049h", "0123456789abcdef"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewScreen(64)
			for _, w := range tc.writes {
				s.Write([]byte(w))
			}
			if _, _, alt := s.Snapshot(); alt != tc.want {
				t.Fatalf("alt = %v, want %v", alt, tc.want)
			}
		})
	}
}

func TestScreenOffset(t *testing.T) {
	s := NewScreen(4)
	s.Write([]byte("ab"))
	if off := s.Write([]byte("cdef")); off != 6 {
		t.Fatalf("off = %d, want 6", off)
	}
	data, off, _ := s.Snapshot()
	if string(data) != "cdef" || off != 6 {
		t.Fatalf("got (%q, %d)", data, off)
	}
}

func TestScreenWriteAt(t *testing.T) {
	s := NewScreen(8)
	if !s.WriteAt([]byte("abc"), 3) {
		t.Fatal("first write rejected")
	}
	if s.WriteAt([]byte("bc"), 3) {
		t.Fatal("duplicate accepted")
	}
	s.WriteAt([]byte("cde"), 5) // 겹친 앞부분 c는 버림
	if data, off, _ := s.Snapshot(); string(data) != "abcde" || off != 5 {
		t.Fatalf("overlap: got (%q, %d)", data, off)
	}
	s.WriteAt([]byte("xy"), 10) // 5..8 빠짐 -> 비우고 이어 씀
	if data, off, _ := s.Snapshot(); string(data) != "xy" || off != 10 {
		t.Fatalf("gap: got (%q, %d)", data, off)
	}
}

func TestScreenSeed(t *testing.T) {
	s := NewScreen(8)
	s.Write([]byte("ab"))
	if !s.Seed([]byte("cdef"), 6, true) {
		t.Fatal("seed rejected")
	}
	if data, off, alt := s.Snapshot(); string(data) != "cdef" || off != 6 || !alt {
		t.Fatalf("got (%q, %d, %v)", data, off, alt)
	}
	if s.Seed([]byte("ef"), 6, false) {
		t.Fatal("stale seed accepted")
	}
	s.WriteAt([]byte("g"), 7)
	if data, off, _ := s.Snapshot(); string(data) != "cdefg" || off != 7 {
		t.Fatalf("after seed: got (%q, %d)", data, off)
	}
}
