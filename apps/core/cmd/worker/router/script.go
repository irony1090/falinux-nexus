package router

import (
	"bytes"
	"fmt"
	"log"
	"nexus/cmd/worker/constants"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/korean"
)

// scriptFormat은 EXEC 직전 스크립트 파일에 적용할 인코딩·줄 끝이다.
// DB 본문은 UTF-8(BOM 없음)·LF라 해석기가 기대하는 형식으로 바꿔 준다 → REF-windows-worker.md WW1(실행 규칙)
type scriptFormat struct {
	encoding string // utf8 | utf8bom | cp949
	eol      string // keep | lf | crlf
}

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// scriptFormats는 OS 기본값(defaultScriptFormats) 위에 env SCRIPT_FORMAT을 덮어쓴 표를 돌려준다.
// 형식: ".ps1:utf8bom,crlf;.bat:cp949,crlf" — 잘못된 항목은 로그만 남기고 건너뛴다.
func scriptFormats() map[string]scriptFormat {
	formats := map[string]scriptFormat{}
	for ext, f := range defaultScriptFormats {
		formats[ext] = f
	}
	for _, item := range strings.Split(constants.GetEnv().ScriptFormat, ";") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		ext, f, err := parseScriptFormat(item)
		if err != nil {
			log.Printf("[SCRIPT_FORMAT] 무시: %q - %v", item, err)
			continue
		}
		formats[ext] = f
	}
	return formats
}

func parseScriptFormat(item string) (string, scriptFormat, error) {
	ext, spec, ok := strings.Cut(item, ":")
	ext = strings.ToLower(strings.TrimSpace(ext))
	if !ok || !strings.HasPrefix(ext, ".") || len(ext) < 2 {
		return "", scriptFormat{}, fmt.Errorf("'.확장자:인코딩,줄끝' 형식이 아님")
	}
	enc, eol, _ := strings.Cut(spec, ",")
	f := scriptFormat{encoding: strings.TrimSpace(enc), eol: strings.TrimSpace(eol)}
	if f.eol == "" {
		f.eol = "keep"
	}
	switch f.encoding {
	case "utf8", "utf8bom", "cp949":
	default:
		return "", scriptFormat{}, fmt.Errorf("인코딩 %q 모름(utf8|utf8bom|cp949)", f.encoding)
	}
	switch f.eol {
	case "keep", "lf", "crlf":
	default:
		return "", scriptFormat{}, fmt.Errorf("줄 끝 %q 모름(keep|lf|crlf)", f.eol)
	}
	return ext, f, nil
}

// applyScriptFormat은 확장자에 해당하는 형식이 있으면 파일을 그 형식으로 다시 쓴다(실행마다 받는 파일이라 DB 무관).
func applyScriptFormat(path string) error {
	f, ok := scriptFormats()[strings.ToLower(filepath.Ext(path))]
	if !ok {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out, lossy := convertScript(data, f)
	if lossy {
		log.Printf("[SCRIPT_FORMAT] %s: %s로 바꿀 수 없는 문자를 '?'로 바꿈", path, f.encoding)
	}
	return os.WriteFile(path, out, 0755)
}

// convertScript는 data(UTF-8)를 f 형식으로 바꾼다. lossy = 대상 인코딩에 없는 문자를 '?'로 바꿈.
func convertScript(data []byte, f scriptFormat) ([]byte, bool) {
	if f.encoding != "utf8" {
		data = bytes.TrimPrefix(data, utf8BOM)
	}
	switch f.eol {
	case "lf":
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	case "crlf":
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		data = bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n"))
	}
	switch f.encoding {
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
