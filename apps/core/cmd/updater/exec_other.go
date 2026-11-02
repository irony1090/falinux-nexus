//go:build !linux

package main

import (
	"os"
	"os/exec"
)

// P4(Windows) 전 자리 — 프로세스 트리 종료는 그때(REF-update-deploy.md "단계 계획")
func setGroup(*exec.Cmd) {}

func killGroup(cmd *exec.Cmd, _ bool) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func exitCode(st *os.ProcessState) int { return st.ExitCode() }
