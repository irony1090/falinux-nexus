package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unicode/utf8"

	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/korean"

	"nexus/internal/scriptfmt"
)

// P4(Windows) 단위-1(스크립트 실행)·단위-2(자식까지 종료) — REF-update-p4.md

// scriptCmd: 확장자별 해석기(worker WW1과 같음). .ps1·.bat·.cmd는 형식을 바꾼 사본을 같은 폴더에 만들어 실행(원본은 copy 대상이라 그대로 둠)
func scriptCmd(path string, runID int64) (string, []string, func(), error) {
	ext := strings.ToLower(filepath.Ext(path))
	run, done := path, func() {}
	if f, ok := scriptfmt.Windows[ext]; ok {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", nil, nil, err
		}
		out, _ := scriptfmt.Convert(b, f)
		run = filepath.Join(filepath.Dir(path), fmt.Sprintf(".nexus-run-%d-%s", runID, filepath.Base(path)))
		if err := os.WriteFile(run, out, 0o644); err != nil {
			return "", nil, nil, err
		}
		done = func() { _ = os.Remove(run) }
	}
	switch ext {
	case ".ps1":
		return "powershell", []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", run}, done, nil
	case ".exe", ".com":
		return run, nil, done, nil
	}
	return "cmd", []string{"/c", run}, done, nil // .bat·.cmd·그 밖 = 파일 연결
}

// outText: 파이프 출력은 프로그램마다 인코딩이 달라(PowerShell·옛 프로그램 = CP949) 줄 단위로 UTF-8이 아니면 CP949로 읽는다
func outText(w io.Writer) io.WriteCloser { return &cp949Lines{w: w} }

const lineMax = 4096 // 줄바꿈 없이 이만큼 쌓이면 그대로 내보냄

type cp949Lines struct {
	mu  sync.Mutex
	w   io.Writer
	buf []byte
}

func (c *cp949Lines) Write(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.buf = append(c.buf, b...)
	for {
		i := bytes.IndexAny(c.buf, "\r\n")
		if i < 0 || (c.buf[i] == '\r' && i+1 == len(c.buf)) { // \r 뒤에 \n이 올 수 있어 기다림
			break
		}
		end := i + 1
		if c.buf[i] == '\r' && c.buf[end] == '\n' {
			end++
		}
		c.emit(c.buf[:end])
		c.buf = c.buf[end:]
	}
	if len(c.buf) > lineMax {
		c.emit(c.buf)
		c.buf = nil
	}
	return len(b), nil
}

func (c *cp949Lines) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.buf) > 0 {
		c.emit(c.buf)
		c.buf = nil
	}
	return nil
}

func (c *cp949Lines) emit(line []byte) {
	if !utf8.Valid(line) {
		if d, err := korean.EUCKR.NewDecoder().Bytes(line); err == nil {
			line = d
		}
	}
	_, _ = c.w.Write(line)
}

// procGroup: Job 하나에 넣어 시간 제한·취소 때 Job째 종료.
// KILL_ON_JOB_CLOSE는 쓰지 않는다 — 정상으로 끝난 단계가 띄운 프로세스(서비스 등)는 남아야 하므로.
// 순서 의존: 멈춘 채(CREATE_SUSPENDED) 시작 -> Job에 넣음 -> 재개. 먼저 재개하면 Job에 넣기 전에 띄운 자식이 Job 밖에 남는다
type procGroup struct {
	cmd *exec.Cmd
	job windows.Handle
}

func newGroup(cmd *exec.Cmd) (*procGroup, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("job 생성 실패: %w", err)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_SUSPENDED}
	return &procGroup{cmd: cmd, job: job}, nil
}

var ntResumeProcess = windows.NewLazySystemDLL("ntdll.dll").NewProc("NtResumeProcess")

// started: Start 직후 — Job에 넣고 재개. Job 등록이 실패해도 재개는 한다(안 하면 멈춘 채 남음)
func (g *procGroup) started() error {
	h, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE|windows.PROCESS_SUSPEND_RESUME, false, uint32(g.cmd.Process.Pid))
	if err != nil {
		_ = g.cmd.Process.Kill()
		return fmt.Errorf("프로세스 열기 실패: %w", err)
	}
	defer windows.CloseHandle(h)
	aerr := windows.AssignProcessToJobObject(g.job, h)
	if r, _, _ := ntResumeProcess.Call(uintptr(h)); r != 0 {
		_ = g.cmd.Process.Kill()
		return fmt.Errorf("프로세스 재개 실패: NTSTATUS 0x%x", r)
	}
	if aerr != nil {
		return fmt.Errorf("job 등록 실패(자식은 같이 안 꺼짐): %w", aerr)
	}
	return nil
}

func (g *procGroup) close() { windows.CloseHandle(g.job) }

// kill: Windows엔 SIGTERM이 없어 force와 상관없이 Job째 종료
func (g *procGroup) kill(bool) {
	if err := windows.TerminateJobObject(g.job, 1); err != nil && g.cmd.Process != nil {
		_ = g.cmd.Process.Kill()
	}
}

func exitCode(st *os.ProcessState) int { return st.ExitCode() }
