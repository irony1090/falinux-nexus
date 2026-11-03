package router

import (
	"log"
	"nexus/cmd/worker/constants"
	"nexus/internal/scriptfmt"
	"os"
	"path/filepath"
	"strings"
)

// EXEC 직전 스크립트 파일 형식 — 변환 본체는 internal/scriptfmt(updater와 공용) → REF-windows-worker.md WW1(실행 규칙)

// scriptFormats는 OS 기본값(defaultScriptFormats) 위에 env SCRIPT_FORMAT을 덮어쓴 표를 돌려준다.
// 형식: ".ps1:utf8bom,crlf;.bat:cp949,crlf" — 잘못된 항목은 로그만 남기고 건너뛴다.
func scriptFormats() map[string]scriptfmt.Format {
	formats := map[string]scriptfmt.Format{}
	for ext, f := range defaultScriptFormats {
		formats[ext] = f
	}
	for _, item := range strings.Split(constants.GetEnv().ScriptFormat, ";") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		ext, f, err := scriptfmt.Parse(item)
		if err != nil {
			log.Printf("[SCRIPT_FORMAT] 무시: %q - %v", item, err)
			continue
		}
		formats[ext] = f
	}
	return formats
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
	out, lossy := scriptfmt.Convert(data, f)
	if lossy {
		log.Printf("[SCRIPT_FORMAT] %s: %s로 바꿀 수 없는 문자를 '?'로 바꿈", path, f.Encoding)
	}
	return os.WriteFile(path, out, 0755)
}
