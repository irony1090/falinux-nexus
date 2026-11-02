package main

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"time"

	"nexus/internal/execute/pty"
)

// runPTY: interactive 단계 — PTY에서 실행, 입력·크기는 procCtl로(시간 제한 없음, 정-6(콘솔))
func runPTY(j *job, path string, env []string) (int, error) {
	it, err := pty.ExecInteractive(context.Background(), path, env, filepath.Dir(path))
	if err != nil {
		return -1, err
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			b, err := it.OutputAll()
			if err != nil {
				return
			}
			j.out(b)
		}
	}()
	j.setProc(ptyProc{it})
	code := -1
	for {
		st, c, err := it.Status()
		if err != nil {
			break
		}
		if st.IsCompleted() {
			code = c
			break
		}
	}
	<-done
	return code, nil
}

type ptyProc struct{ it *pty.Interactive }

func (p ptyProc) write(b []byte)           { _ = p.it.Write(b) }
func (p ptyProc) resize(cols, rows uint16) { _ = p.it.Layout(cols, rows) }
func (p ptyProc) kill(force bool)          { _ = p.it.Kill(force) }

// runPipe: 비대화형 단계 — stdin 없음(/dev/null), stdout·stderr 합쳐서. 화면용으로 LF -> CRLF
func runPipe(j *job, path string, env []string) (int, error) {
	cmd := exec.Command(path)
	cmd.Dir = filepath.Dir(path)
	cmd.Env = env
	w := crlfWriter{outWriter{j}}
	cmd.Stdout, cmd.Stderr = w, w
	// 스크립트가 띄운 데몬이 출력 파이프를 쥐고 있어도 Wait가 끝나게
	cmd.WaitDelay = 2 * time.Second
	setGroup(cmd)
	if err := cmd.Start(); err != nil {
		return -1, err
	}
	j.setProc(pipeProc{cmd})
	err := cmd.Wait()
	if cmd.ProcessState == nil {
		return -1, err
	}
	return exitCode(cmd.ProcessState), nil
}

type pipeProc struct{ cmd *exec.Cmd }

func (p pipeProc) write([]byte)       {}
func (p pipeProc) resize(_, _ uint16) {}
func (p pipeProc) kill(force bool)    { killGroup(p.cmd, force) }

type crlfWriter struct{ w outWriter }

func (c crlfWriter) Write(b []byte) (int, error) {
	_, err := c.w.Write(bytes.ReplaceAll(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")), []byte("\n"), []byte("\r\n")))
	return len(b), err
}
