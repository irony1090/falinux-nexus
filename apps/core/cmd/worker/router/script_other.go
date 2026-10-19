//go:build !windows

package router

import "nexus/internal/protocol"

// Linux 등은 기본값 없음(파일 그대로) — env SCRIPT_FORMAT을 주면 그것만 적용한다.
var defaultScriptFormats = map[string]scriptFormat{}

// defaultEditor는 EDITOR가 없을 때 쓸 에디터다.
func defaultEditor() string { return "vi" }

// prepareExec는 파일을 직접 실행하는 지금 방식(shebang)을 그대로 둔다.
func prepareExec(body protocol.ProcessSpec) (protocol.ProcessSpec, error) {
	if err := applyScriptFormat(body.Cmd); err != nil {
		return body, err
	}
	return body, nil
}
