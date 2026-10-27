//go:build !windows

package router

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"nexus/internal/protocol"
)

// topListing: 최상위 = `/`(Q6(처음 여는 위치))
func topListing() (*dirListing, error) { return readListing("/") }

// rootParent: `/`는 위가 없다
func rootParent() *string { return nil }

// execOK: 실행 비트가 있는 일반 파일 — Q24(실행 대상)
func execOK(_ string, m fs.FileMode) bool { return m.IsRegular() && m.Perm()&0o111 != 0 }

// directRunner: 파일을 그대로 실행(shebang·ELF)
func directRunner(body protocol.ProcessSpec) protocol.ProcessSpec { return body }

// shellPath: "여기서 셸"의 셸 — $SHELL > PATH의 bash > /bin/sh. 절대 경로여야 prepareDirect를 지난다
func shellPath() string {
	if s := os.Getenv("SHELL"); filepath.IsAbs(s) {
		return s
	}
	if p, err := exec.LookPath("bash"); err == nil && filepath.IsAbs(p) {
		return p
	}
	return "/bin/sh"
}
