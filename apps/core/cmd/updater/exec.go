package main

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"path/filepath"
	"time"

	"nexus/internal/execute/pty"
)

// runPTY: interactive 단계 — PTY에서 실행, 입력·크기는 procCtl로(시간 제한 없음, 정-6(콘솔))
func runPTY(j *job, path string, env []string) (int, error) {
	name, args, cleanup, err := scriptCmd(path, j.run.RunID)
	if err != nil {
		return -1, err
	}
	defer cleanup()
	it, err := pty.ExecInteractive(context.Background(), name, env, filepath.Dir(path), args...)
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
	name, args, cleanup, err := scriptCmd(path, j.run.RunID)
	if err != nil {
		return -1, err
	}
	defer cleanup()
	cmd := exec.Command(name, args...)
	cmd.Dir = filepath.Dir(path)
	cmd.Env = env
	w := outText(crlfWriter{outWriter{j}})
	cmd.Stdout, cmd.Stderr = w, w // 같은 값이면 Write는 한 번에 하나(os/exec)
	// 스크립트가 띄운 데몬이 출력 파이프를 쥐고 있어도 Wait가 끝나게
	cmd.WaitDelay = 2 * time.Second
	g, err := newGroup(cmd)
	if err != nil {
		return -1, err
	}
	defer g.close()
	if err := cmd.Start(); err != nil {
		return -1, err
	}
	if err := g.started(); err != nil {
		j.say("주의: %v", err)
	}
	j.setProc(pipeProc{g})
	err = cmd.Wait()
	_ = w.Close()
	if cmd.ProcessState == nil {
		return -1, err
	}
	return exitCode(cmd.ProcessState), nil
}

type pipeProc struct{ g *procGroup }

func (p pipeProc) write([]byte)       {}
func (p pipeProc) resize(_, _ uint16) {}
func (p pipeProc) kill(force bool)    { p.g.kill(force) }

// nopFlush: 버퍼 없는 출력(Linux)
type nopFlush struct{ io.Writer }

func (nopFlush) Close() error { return nil }

type crlfWriter struct{ w outWriter }

func (c crlfWriter) Write(b []byte) (int, error) {
	_, err := c.w.Write(bytes.ReplaceAll(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")), []byte("\n"), []byte("\r\n")))
	return len(b), err
}
