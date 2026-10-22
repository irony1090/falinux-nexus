//go:build windows

package hold

import (
	"os"
	"syscall"
)

// detachAttr: 새 프로세스 그룹 + 콘솔 없음 — worker 콘솔의 Ctrl+C·닫기가 닿지 않게(미확인, K6(Windows) Linux 먼저)
func detachAttr() *syscall.SysProcAttr {
	const detachedProcess = 0x00000008
	return &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcess}
}

// IsHolder: Windows는 확인 수단이 아직 없어 정리하지 않는다
func IsHolder(pid int, sock string) bool { return false }

func KillHolder(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
