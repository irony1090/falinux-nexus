package router

import (
	"errors"
	"nexus/internal/protocol"
	"os/exec"
	"path/filepath"
	"strings"
)

// Windows 기본 스크립트 형식 — env SCRIPT_FORMAT으로 덮어쓴다.
var defaultScriptFormats = map[string]scriptFormat{
	".ps1": {encoding: "utf8bom", eol: "crlf"}, // Windows PowerShell 5.1은 BOM 없는 파일을 시스템 코드 페이지로 읽음
	".bat": {encoding: "cp949", eol: "crlf"},
	".cmd": {encoding: "cp949", eol: "crlf"},
}

// defaultEditor는 EDITOR가 없을 때 PATH에서 처음 찾은 터미널 에디터다(없으면 "" -> prepareEdit가 안내).
// notepad 같은 GUI 에디터는 ConPTY에 안 보여 후보에서 뺀다.
func defaultEditor() string {
	for _, name := range []string{"edit", "nvim", "vim", "nano"} {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}
	return ""
}

// prepareExec는 EXEC 명세를 Windows 실행 규칙으로 바꾼다 → REF-windows-worker.md WW1(실행 규칙)
// .ps1 = powershell(파일 연결 기본값이 메모장이라 고정) / 그 외 = cmd /c(Windows 파일 연결). #!는 쓰지 않는다.
func prepareExec(body protocol.ProcessSpec) (protocol.ProcessSpec, error) {
	path := filepath.Clean(body.Cmd) // 선배치(resolveDest)와 같은 \ 경로
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return body, errors.New("Windows에서는 노드 이름에 확장자(.ps1·.bat·.py 등)가 필요합니다")
	}
	if err := applyScriptFormat(path); err != nil {
		return body, err
	}
	return scriptRunner(body, path), nil
}

// scriptRunner: path를 .ps1이면 powershell로, 그 외는 cmd /c로 실행하는 명세
func scriptRunner(body protocol.ProcessSpec, path string) protocol.ProcessSpec {
	if strings.ToLower(filepath.Ext(path)) == ".ps1" {
		body.Cmd = "powershell"
		body.Args = append([]string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", path}, body.Args...)
	} else {
		body.Cmd = "cmd"
		body.Args = append([]string{"/c", path}, body.Args...)
	}
	return body
}
