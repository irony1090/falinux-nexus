package main

import (
	"os"
	"os/exec"
	"syscall"
)

// setGroup: 새 프로세스 그룹 — 시간 제한·취소 때 자식까지 같이 종료
func setGroup(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} }

func killGroup(cmd *exec.Cmd, force bool) {
	if cmd.Process == nil {
		return
	}
	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}
	_ = syscall.Kill(-cmd.Process.Pid, sig)
}

// exitCode: 신호로 죽었으면 128+신호(셸 관례)
func exitCode(st *os.ProcessState) int {
	if ws, ok := st.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return st.ExitCode()
}
