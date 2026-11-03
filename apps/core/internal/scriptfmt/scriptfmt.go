// Package scriptfmt: 스크립트 파일을 해석기가 기대하는 인코딩·줄 끝으로 바꾼다(worker EXEC·updater script 단계 공용).
// 보관 원본은 UTF-8(BOM 없음)·LF — REF-windows-exec.md
package scriptfmt

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/korean"
)

// Format: 인코딩 utf8 | utf8bom | cp949 / 줄 끝 keep | lf | crlf
type Format struct {
	Encoding string
	EOL      string
}

// Windows: Windows 기본값 — PowerShell 5.1은 BOM 없는 파일을 시스템 코드 페이지로 읽음, .bat은 CP949·CRLF여야 한글·goto가 맞음
var Windows = map[string]Format{
	".ps1": {Encoding: "utf8bom", EOL: "crlf"},
	".bat": {Encoding: "cp949", EOL: "crlf"},
	".cmd": {Encoding: "cp949", EOL: "crlf"},
}

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Parse: ".ps1:utf8bom,crlf" 한 항목. 줄 끝 생략 = keep
func Parse(item string) (string, Format, error) {
	ext, spec, ok := strings.Cut(item, ":")
	ext = strings.ToLower(strings.TrimSpace(ext))
	if !ok || !strings.HasPrefix(ext, ".") || len(ext) < 2 {
		return "", Format{}, fmt.Errorf("'.확장자:인코딩,줄끝' 형식이 아님")
	}
	enc, eol, _ := strings.Cut(spec, ",")
	f := Format{Encoding: strings.TrimSpace(enc), EOL: strings.TrimSpace(eol)}
	if f.EOL == "" {
		f.EOL = "keep"
	}
	switch f.Encoding {
	case "utf8", "utf8bom", "cp949":
	default:
		return "", Format{}, fmt.Errorf("인코딩 %q 모름(utf8|utf8bom|cp949)", f.Encoding)
	}
	switch f.EOL {
	case "keep", "lf", "crlf":
	default:
		return "", Format{}, fmt.Errorf("줄 끝 %q 모름(keep|lf|crlf)", f.EOL)
	}
	return ext, f, nil
}

// Convert: data(UTF-8)를 f 형식으로. lossy = 대상 인코딩에 없는 문자를 '?'로 바꿈
func Convert(data []byte, f Format) ([]byte, bool) {
	if f.Encoding != "utf8" {
		data = bytes.TrimPrefix(data, utf8BOM)
	}
	switch f.EOL {
	case "lf":
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	case "crlf":
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		data = bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n"))
	}
	switch f.Encoding {
	case "utf8bom":
		return append(append([]byte{}, utf8BOM...), data...), false
	case "cp949":
		return encodeCP949(data)
	}
	return data, false
}

func encodeCP949(data []byte) ([]byte, bool) {
	if out, err := korean.EUCKR.NewEncoder().Bytes(data); err == nil {
		return out, false
	}
	// 한 글자씩 바꿔 실패한 글자만 '?'로(이모지 등 CP949 밖 문자)
	enc := korean.EUCKR.NewEncoder()
	out := make([]byte, 0, len(data))
	var buf [utf8.UTFMax]byte
	for _, r := range string(data) {
		n := utf8.EncodeRune(buf[:], r)
		b, err := enc.Bytes(buf[:n])
		if err != nil {
			out = append(out, '?')
			continue
		}
		out = append(out, b...)
	}
	return out, true
}
