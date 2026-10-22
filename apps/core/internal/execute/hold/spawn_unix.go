//go:build !windows

package hold

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
)

// detachAttr: worker와 다른 세션 — worker의 터미널이 닫혀도(SIGHUP) 보관 프로세스는 남는다
func detachAttr() *syscall.SysProcAttr { return &syscall.SysProcAttr{Setsid: true} }

// IsHolder는 pid가 sock을 맡은 보관 프로세스인지 cmdline으로 확인한다(pid 재사용 대비)
func IsHolder(pid int, sock string) bool {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return false
	}
	args := bytes.Split(bytes.TrimRight(b, "\x00"), []byte{0})
	return len(args) >= 3 && string(args[len(args)-2]) == "-hold" && string(args[len(args)-1]) == sock
}

// KillHolder는 소켓이 망가진 보관 프로세스를 끝낸다 — PTY master가 닫혀 스크립트도 SIGHUP을 받는다
func KillHolder(pid int) error { return syscall.Kill(pid, syscall.SIGKILL) }
