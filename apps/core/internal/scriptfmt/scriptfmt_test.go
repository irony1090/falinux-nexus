package scriptfmt

import (
	"bytes"
	"testing"

	"golang.org/x/text/encoding/korean"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		ext  string
		f    Format
		fail bool
	}{
		{in: ".PS1:utf8bom,crlf", ext: ".ps1", f: Format{"utf8bom", "crlf"}},
		{in: ".bat:cp949", ext: ".bat", f: Format{"cp949", "keep"}},
		{in: "ps1:utf8", fail: true},
		{in: ".ps1:latin1", fail: true},
		{in: ".ps1:utf8,cr", fail: true},
	}
	for _, c := range cases {
		ext, f, err := Parse(c.in)
		if (err != nil) != c.fail || (!c.fail && (ext != c.ext || f != c.f)) {
			t.Errorf("Parse(%q) = %q %+v %v", c.in, ext, f, err)
		}
	}
}

func TestConvertBOMCRLFTwice(t *testing.T) {
	src := []byte("echo 한글\nexit 0\n")
	once, _ := Convert(src, Windows[".ps1"])
	twice, _ := Convert(once, Windows[".ps1"])
	want := append([]byte{0xEF, 0xBB, 0xBF}, []byte("echo 한글\r\nexit 0\r\n")...)
	if !bytes.Equal(once, want) || !bytes.Equal(twice, want) {
		t.Fatalf("once=%q twice=%q", once, twice)
	}
}

func TestConvertCP949(t *testing.T) {
	out, lossy := Convert([]byte("echo 한글\n"), Windows[".bat"])
	back, err := korean.EUCKR.NewDecoder().Bytes(out)
	if lossy || err != nil || string(back) != "echo 한글\r\n" {
		t.Fatalf("lossy=%v err=%v back=%q", lossy, err, back)
	}
	out, lossy = Convert([]byte("a😀b"), Windows[".bat"])
	if !lossy || string(out) != "a?b" {
		t.Fatalf("emoji: lossy=%v out=%q", lossy, out)
	}
}
