package main

import (
	"io"
	"os"
	"os/exec"
	"syscall"
)

// scriptCmd: Linux = 파일을 바로 실행(shebang). 변환 없음
func scriptCmd(path string, _ int64) (string, []string, func(), error) {
	return path, nil, func() {}, nil
}

// outText: Linux 출력은 그대로
func outText(w io.Writer) io.WriteCloser { return nopFlush{w} }

// procGroup: 새 프로세스 그룹 — 시간 제한·취소 때 자식까지 같이 종료
type procGroup struct{ cmd *exec.Cmd }

func newGroup(cmd *exec.Cmd) (*procGroup, error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return &procGroup{cmd}, nil
}

func (g *procGroup) started() error { return nil }
func (g *procGroup) close()         {}

func (g *procGroup) kill(force bool) {
	if g.cmd.Process == nil {
		return
	}
	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}
	_ = syscall.Kill(-g.cmd.Process.Pid, sig)
}

// exitCode: 신호로 죽었으면 128+신호(셸 관례)
func exitCode(st *os.ProcessState) int {
	if ws, ok := st.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return st.ExitCode()
}
