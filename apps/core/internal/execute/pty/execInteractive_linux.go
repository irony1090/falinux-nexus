package pty

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

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

	if err := setUTF8(slave); err != nil {
		slave.Close()
		master.Close()
		return nil, nil, err
	}
	return master, slave, nil
}

// setUTF8은 줄 편집(canonical 모드)의 백스페이스가 UTF-8 한 글자를 지우게 IUTF8을 켠다(없으면 1바이트만 지움)
func setUTF8(slave *os.File) error {
	var t syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, slave.Fd(), syscall.TCGETS, uintptr(unsafe.Pointer(&t))); errno != 0 {
		return fmt.Errorf("termios 읽기 실패: %v", errno)
	}
	t.Iflag |= syscall.IUTF8
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, slave.Fd(), syscall.TCSETS, uintptr(unsafe.Pointer(&t))); errno != 0 {
		return fmt.Errorf("termios 쓰기 실패: %v", errno)
	}
	return nil
}

// sysProc는 Interactive의 Linux 전용 핸들이다.
type sysProc struct {
	master *os.File
	cmd    *exec.Cmd
}

// Kill은 프로세스 그룹 전체에 신호를 보낸다(Setsid라 pgid = pid).
func (i *Interactive) Kill(force bool) error {
	if i.sys.cmd == nil || i.sys.cmd.Process == nil {
		return nil
	}
	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}
	return syscall.Kill(-i.sys.cmd.Process.Pid, sig)
}

// exitCodeOf는 종료된 프로세스의 exit code를 유닉스 관례로 계산한다. 시그널로 죽었으면(Kill()의
// SIGTERM·SIGKILL 포함 — 외부에서 kill된 경우도 동일) Go의 ProcessState.ExitCode()는 -1을 줘서 "정상
// 종료했지만 코드가 0이 아님(진짜 실패)"과 구분이 안 된다. bash/docker/systemd와 동일하게
// 128+시그널번호(SIGKILL=137)로 인코딩해 두 케이스를 구분 가능하게 한다.
func exitCodeOf(state *os.ProcessState) int {
	if ws, ok := state.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return state.ExitCode()
}

type Winsize struct {
	Row    uint16 // 행 수 (터미널 세로 크기)
	Col    uint16 // 열 수 (터미널 가로 크기)
	Xpixel uint16 // 픽셀 단위 가로 (선택사항, 보통 0)
	Ypixel uint16 // 픽셀 단위 세로 (선택사항, 보통 0)
}

func (i *Interactive) Layout(cols, rows uint16) error {
	// 터미널 크기 설정
	_, _, errNo := syscall.Syscall(
		syscall.SYS_IOCTL,
		i.sys.master.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&Winsize{
			Row: rows,
			Col: cols,
		})),
	)
	if errNo != 0 {
		return errNo
	}
	return nil
}

// Pid는 worker 로컬 OS 프로세스 ID를 반환한다(미기동/종료 시 -1).
// RUNNING 보고(MsgStatus PROCESS)에 얹어 supervisor로 올린다. 와이어 라우팅 키는 UID이고
// PID는 관측/제어 참고값이다(REF-process "UID vs PID 분리").
func (i *Interactive) Pid() int {
	if i.sys.cmd == nil || i.sys.cmd.Process == nil {
		return -1
	}
	return i.sys.cmd.Process.Pid
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

	interactive := newInteractive(sysProc{master: master, cmd: cmd})
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
