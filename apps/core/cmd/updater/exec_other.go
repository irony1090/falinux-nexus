//go:build !linux && !windows

package main

import (
	"io"
	"os"
	"os/exec"
)

// Linux·Windows 밖(macOS 등 개발용) — 파일 바로 실행, 부모만 종료
func scriptCmd(path string, _ int64) (string, []string, func(), error) {
	return path, nil, func() {}, nil
}

func outText(w io.Writer) io.WriteCloser { return nopFlush{w} }

type procGroup struct{ cmd *exec.Cmd }

func newGroup(cmd *exec.Cmd) (*procGroup, error) { return &procGroup{cmd}, nil }
func (g *procGroup) started() error              { return nil }
func (g *procGroup) close()                      {}

func (g *procGroup) kill(bool) {
	if g.cmd.Process != nil {
		_ = g.cmd.Process.Kill()
	}
}

func exitCode(st *os.ProcessState) int { return st.ExitCode() }
