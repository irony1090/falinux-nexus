// Package pty는 로컬 PTY 기반 process 실행(Linux 전용)을 제공한다.
// 이식 가능한 execute 패키지(IInteractive·AgentInteractive·CommandStatus)와 분리되어,
// supervisor 등 비-Linux 빌드가 execute만 import할 수 있게 한다.
package pty

import (
	"bytes"
	"context"
	"fmt"
	"nexus/internal/execute"
	"nexus/internal/syncProcess"
	"nexus/internal/util"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

// 컴파일 시 *Interactive가 execute.IInteractive를 만족하는지 확인.
var _ execute.IInteractive = (*Interactive)(nil)

// openPty는 Linux에서 PTY master/slave를 생성합니다
func openPty() (master *os.File, slave *os.File, err error) {
	master, err = os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("ptmx 열기 실패: %w", err)
	}

	// unlock the slave pty
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		master.Fd(),
		syscall.TIOCSPTLCK,
		uintptr(unsafe.Pointer(new(int))),
	)
	if errno != 0 {
		master.Close()
		return nil, nil, fmt.Errorf("unlockpt 실패: %v", errno)
	}

	var n uint32
	// get the slave pty number
	_, _, errno = syscall.Syscall(
		syscall.SYS_IOCTL,
		master.Fd(),
		syscall.TIOCGPTN,
		uintptr(unsafe.Pointer(&n)),
	)
	if errno != 0 {
		master.Close()
		return nil, nil, fmt.Errorf("ptsname 실패: %v", errno)
	}

	slavePath := fmt.Sprintf("/dev/pts/%d", n)
	slave, err = os.OpenFile(slavePath, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("slave 열기 실패: %w", err)
	}

	return master, slave, nil
}

type Interactive struct {
	Error    error
	exitCode int
	status   *syncProcess.SyncData[execute.CommandStatus]
	ouput    *syncProcess.SyncData[[]byte]
	input    *syncProcess.SyncData[[]byte]
	master   *os.File
	cmd      *exec.Cmd
}

func (i *Interactive) Kill() error {
	if i.cmd == nil || i.cmd.Process == nil {
		return nil
	}
	return syscall.Kill(-i.cmd.Process.Pid, syscall.SIGKILL)
}

// exitCodeOf는 종료된 프로세스의 exit code를 유닉스 관례로 계산한다. 시그널로 죽었으면(Kill()의
// SIGKILL 포함 — 외부에서 kill된 경우도 동일) Go의 ProcessState.ExitCode()는 -1을 줘서 "정상
// 종료했지만 코드가 0이 아님(진짜 실패)"과 구분이 안 된다. bash/docker/systemd와 동일하게
// 128+시그널번호(SIGKILL=137)로 인코딩해 두 케이스를 구분 가능하게 한다.
func exitCodeOf(state *os.ProcessState) int {
	if ws, ok := state.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return state.ExitCode()
}

func (i *Interactive) setStatusPending() {
	i.status.Push(execute.CommandPending)
}
func (i *Interactive) setStatusProcess() {
	i.status.Push(execute.CommandProcess)
}
func (i *Interactive) setStatusCompleted(exitCode int, err error) {
	defer i.status.Close()
	defer i.ouput.Close()
	defer i.input.Close()

	i.Error = err
	i.exitCode = exitCode
	if exitCode == 0 {
		i.status.Push(execute.CommandCompleted)
	} else {
		i.status.Push(execute.CommandFailed)
	}

}

// Status 명령어 상태 조회. 데이터 없으면 상태가 올 때까지 대기.
// 세 번째 반환값은 큐 종료 신호(Shift 자신의 err)다 — i.Error(cmd.Wait()의 종료 에러, 0이
// 아닌 코드/시그널 종료 시 항상 non-nil)를 여기 얹으면 호출자(pumpStatus 등)의 "err != nil =
// 더 이상 없음" 관례와 충돌해 마지막 상태(Failed)를 못 보내고 조기 리턴하게 된다.
func (i *Interactive) Status() (execute.CommandStatus, int, error) {
	sts, err := i.status.Shift()
	return sts, i.exitCode, err
}

// Output 대화형 출력 읽기. 데이터가 없으면 데이터가 올 때까지 대기
func (i *Interactive) Output() ([]byte, error) {
	return i.ouput.Shift()
}
func (i *Interactive) OutputAll() ([]byte, error) {
	batch, err := i.ouput.ShiftAll()
	if err != nil {
		return nil, err
	}
	return bytes.Join(batch, nil), nil
}
func (i *Interactive) Write(data []byte) error {
	return i.input.Push(data)
}

type Winsize struct {
	Row    uint16 // 행 수 (터미널 세로 크기)
	Col    uint16 // 열 수 (터미널 가로 크기)
	Xpixel uint16 // 픽셀 단위 가로 (선택사항, 보통 0)
	Ypixel uint16 // 픽셀 단위 세로 (선택사항, 보통 0)
}

func (i *Interactive) Layout(cols, rows uint16) syscall.Errno {
	// 터미널 크기 설정
	_, _, errNo := syscall.Syscall(
		syscall.SYS_IOCTL,
		i.master.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&Winsize{
			Row: rows,
			Col: cols,
		})),
	)
	return errNo
}

func (i *Interactive) Refresh() error {
	i.Write([]byte{0x0c})
	// err := syscall.Kill(-i.cmd.Process.Pid, syscall.SIGWINCH)
	// log.Printf("REFRESH~ %v", err)
	return nil
}
func (i *Interactive) ExitCode() int {
	return i.exitCode
}

// Pid는 worker 로컬 OS 프로세스 ID를 반환한다(미기동/종료 시 -1).
// RUNNING 보고(MsgStatus PROCESS)에 얹어 supervisor로 올린다. 와이어 라우팅 키는 UID이고
// PID는 관측/제어 참고값이다(REF-process "UID vs PID 분리").
func (i *Interactive) Pid() int {
	if i.cmd == nil || i.cmd.Process == nil {
		return -1
	}
	return i.cmd.Process.Pid
}

// ExecInteractive 명령어를 PTY에서 대화형으로 실행한다.
// env는 자식 프로세스 환경변수("KEY=VALUE" 목록)다. nil이면 worker 프로세스 환경(os.Environ())을
// 그대로 상속한다. 값을 넘기면 그 목록으로 완전히 대체되므로(상속 안 함), 호출자가 필요한 베이스
// (os.Environ())·TERM 등을 미리 조립해 넘긴다. TUI 에디터(vi 등)엔 TERM 지정이 사실상 필수.
// dir은 작업 디렉토리다. 빈 문자열이면 worker 프로세스의 cwd를 상속한다(exec.Cmd.Dir 기본값과 동일).
func ExecInteractive(ctx context.Context, command string, env []string, dir string, args ...string) (*Interactive, error) {
	master, slave, err := openPty()
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdin = slave
	cmd.Stdout = slave
	cmd.Stderr = slave
	if env != nil {
		cmd.Env = env
	}
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("명령어 시작 실패: %v", err)
	}
	slave.Close()

	// oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	// if err != nil {
	// 	return nil, fmt.Errorf("터미널 raw mode 실패: %v", err)
	// }

	interactive := &Interactive{
		exitCode: -1,
		status:   syncProcess.NewSyncData([]execute.CommandStatus{}),
		ouput:    syncProcess.NewSyncData([][]byte{}),
		input:    syncProcess.NewSyncData([][]byte{}),
		master:   master,
		cmd:      cmd,
	}
	interactive.setStatusPending()
	go func() {
		interactive.setStatusProcess()
		err = cmd.Wait()
		master.Close()
		interactive.setStatusCompleted(exitCodeOf(cmd.ProcessState), err)
		// term.Restore(int(os.Stdin.Fd()), oldState)
	}()

	go streamOutputLogic(master, interactive.ouput)
	go streamInputLogic(master, interactive.input)

	return interactive, nil
}

func streamInputLogic(master *os.File, input *syncProcess.SyncData[[]byte]) {
	for {
		data, err := input.Shift()
		if err != nil {
			break
		}
		master.Write(data)
	}
}

func streamOutputLogic(master *os.File, output *syncProcess.SyncData[[]byte]) {
	buffer := make([]byte, 1024)
	lines := make([]byte, 0, 1024)
	for {
		// log.Printf("[INTERACTIVE] READED %d bytes", len(lines))
		n, err := master.Read(buffer)
		if n > 0 {
			lines = append(lines, buffer[:n]...)
			holdback := util.IncompleteTailLen(lines)
			safe := lines[:len(lines)-holdback]
			publishLine(safe, output)
			lines = append([]byte{}, lines[len(lines)-holdback:]...)

		} else if err != nil {
			// log.Printf("[INTERACTIVE] ERR - %v", err)
			if len(lines) > 0 {
				ok, newLine := publishLine(lines, output)
				if ok {
					lines = newLine
				}
				output.Push(bytes.Clone(lines))
			}
			break
		}

	}
	// log.Printf("[INTERACTIVE] EXIT")

}

func publishLine(data []byte, output *syncProcess.SyncData[[]byte]) (bool, []byte) {
	if lastIdx, _ := util.IsCompletedUTF8(data); lastIdx > -1 {
		// log.Printf("[PUBLISH_LINE] ok - %v \n", lastIdx)
		c := bytes.Clone(data)
		prefix := empty
		for {
			newLineIndex, str, last := newLineBytes(c)
			line := make([]byte, 0)
			if newLineIndex < 0 {

				if prefix != empty {
					line = append(line, prefix)
					prefix = empty
				}
				line = append(line, c...)
				c = c[len(c):]
				if !bytes.Equal(line, blank) {
					// log.Printf("\t[FOR] noNewLine %d\n", len(line))
					output.Push(line)
				}

				break
			}

			if prefix != empty { // 이전 라인에서 캐리지리턴이 있었음
				line = append(line, prefix)
			}

			// 캐리지리턴인 경우. 다음 라인에서 처리
			if last == r {
				line = append(line, str...)
				prefix = last
			} else { // 개행인 경우
				line = append(line, str...)
				line = append(line, rn...)
				// content = prefix + str + last
				prefix = empty
			}
			c = c[newLineIndex:]
			if !bytes.Equal(line, blank) {
				// log.Printf("\t[FOR] newLine %d\n", len(line))
				output.Push(line)
			}
		}
		return true, c
	} else {
		// log.Printf("[PUBLISH_LINE] fail - %v \n", lastIdx)

		return false, nil
	}
}

var (
	empty = byte(0)
	n     = byte('\n')
	r     = byte('\r')
	rn    = []byte("\r\n")
	nr    = []byte("\n\r")
	blank = []byte("")
)

// newLineBytes 바이트 배열에서 첫 번째 개행(\n, \r, \r\n, \n\r)까지의 인덱스와 내용을 반환
func newLineBytes(c []byte) (int, []byte, byte) {
	// 1. \r\n 또는 \n\r 먼저 처리
	rnIndex := bytes.Index(c, rn)
	nrIndex := bytes.Index(c, nr)

	if rnIndex >= 0 || nrIndex >= 0 {
		idx := rnIndex
		if nrIndex >= 0 && (rnIndex < 0 || nrIndex < rnIndex) {
			idx = nrIndex
		}
		return idx + 2, bytes.Clone(c[:idx]), n
	}

	// 2. 단일 \r 또는 \n 중 먼저 나오는 것 찾기
	rIndex := bytes.IndexByte(c, r)
	nIndex := bytes.IndexByte(c, n)

	// 둘 다 없으면
	if rIndex < 0 && nIndex < 0 {
		return -1, nil, empty
	}

	// \n만 있거나, 둘 다 있는데 \n이 더 앞에 있으면
	if rIndex < 0 || (nIndex >= 0 && nIndex < rIndex) {
		return nIndex + 1, bytes.Clone(c[:nIndex]), n
	}

	// \r이 더 앞에 있으면 (개행 아님, 그대로 유지)
	return rIndex + 1, bytes.Clone(c[:rIndex]), r

}
