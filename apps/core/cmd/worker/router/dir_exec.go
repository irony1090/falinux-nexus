package router

import (
	"errors"
	"os"
	"path/filepath"

	"nexus/internal/protocol"
)

// prepareDirect: 장비 파일 그대로 실행(폴더-4(실행), REF-worker-dir-exec.md) — 사용자 파일이라 형식 변환(applyScriptFormat)은 하지 않는다.
// 목록에서 "실행"을 보인 판정(execOK)을 여기서 다시 본다(REST로 아무 경로나 올 수 있음)
func prepareDirect(body protocol.ProcessSpec) (protocol.ProcessSpec, error) {
	if !filepath.IsAbs(body.Cmd) || !filepath.IsAbs(body.Cwd) {
		return body, errors.New("절대 경로가 아닙니다")
	}
	body.Cmd, body.Cwd = filepath.Clean(body.Cmd), filepath.Clean(body.Cwd)
	st, err := os.Stat(body.Cmd) // 링크면 원본을 본다
	if err != nil {
		return body, err
	}
	if !execOK(body.Cmd, st.Mode()) {
		return body, errors.New("실행할 수 있는 파일이 아닙니다: " + body.Cmd)
	}
	if st, err := os.Stat(body.Cwd); err != nil || !st.IsDir() {
		return body, errors.New("실행할 폴더가 없습니다: " + body.Cwd)
	}
	return directRunner(body), nil
}
