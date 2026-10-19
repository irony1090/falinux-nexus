package pty

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// 설계·순서 근거 → REF-windows-conpty.md
const (
	forceExitCode      = 137 // 강제 종료 exit code — Linux SIGKILL(128+9)과 화면 표시를 맞춘다
	outputDrainTimeout = 3 * time.Second
)

// sysProc는 Interactive의 Windows 전용 핸들이다(뮤텍스가 있어 포인터로 둔다).
type sysProc struct{ c *conpty }

type conpty struct {
	mu      sync.Mutex
	closed  bool // 대기 고루틴이 핸들을 닫기 시작함 — 이후 hpc·job을 쓰지 않는다
	hpc     windows.Handle
	job     windows.Handle
	process windows.Handle
	pid     int
}

// Kill: 요청 = 입력에 Ctrl+C(ConPTY가 CTRL_C_EVENT로 바꿈) / 강제 = job 통째 종료.
func (i *Interactive) Kill(force bool) error {
	if !force {
		return i.Write([]byte{0x03})
	}
	c := i.sys.c
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	return windows.TerminateJobObject(c.job, forceExitCode)
}

func (i *Interactive) Layout(cols, rows uint16) error {
	c := i.sys.c
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	return windows.ResizePseudoConsole(c.hpc, windows.Coord{X: int16(cols), Y: int16(rows)})
}

// Pid는 worker 로컬 OS 프로세스 ID를 반환한다(REF-process "UID vs PID 분리").
func (i *Interactive) Pid() int {
	return i.sys.c.pid
}

var restoreCtrlC sync.Once

// 재발 방지(REF-windows-conpty.md "윈도-6(확인) 결과", Linux판 = REF-worker-orphan.md SIGHUP 무시 상속):
// WMI·서비스 등으로 띄운 worker는 "Ctrl+C 무시" 속성을 갖고 시작할 수 있고, 이 속성은 자식에 상속된다.
// 그러면 Kill(false)의 Ctrl+C를 스크립트가 무시하므로 첫 실행 전에 worker 자신의 속성을 되돌린다.
// x/sys에 SetConsoleCtrlHandler가 없어 kernel32에서 직접 부른다(NULL, FALSE = 무시 해제).
func allowCtrlCInChildren() {
	restoreCtrlC.Do(func() {
		windows.NewLazySystemDLL("kernel32.dll").NewProc("SetConsoleCtrlHandler").Call(0, 0)
	})
}

// ExecInteractive 명령어를 ConPTY에서 대화형으로 실행한다. env·dir 의미는 Linux판과 같다.
// 초기 크기는 80x24이고 이후 supervisor의 resize(Layout)로 맞춰진다.
func ExecInteractive(ctx context.Context, command string, env []string, dir string, args ...string) (_ *Interactive, err error) {
	var cleanup []func()
	defer func() {
		if err != nil {
			for _, f := range slices.Backward(cleanup) {
				f()
			}
		}
	}()

	allowCtrlCInChildren()

	path, err := exec.LookPath(command)
	if err != nil {
		return nil, fmt.Errorf("명령어 찾기 실패: %w", err)
	}
	appName, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	cmdLine, err := windows.UTF16PtrFromString(windows.ComposeCommandLine(append([]string{path}, args...)))
	if err != nil {
		return nil, err
	}
	var curDir *uint16
	if dir != "" {
		if curDir, err = windows.UTF16PtrFromString(dir); err != nil {
			return nil, err
		}
	}
	var envBlock *uint16
	flags := uint32(windows.EXTENDED_STARTUPINFO_PRESENT | windows.CREATE_SUSPENDED)
	if env != nil {
		if envBlock, err = createEnvBlock(env); err != nil {
			return nil, err
		}
		flags |= windows.CREATE_UNICODE_ENVIRONMENT
	}

	var ptyIn, inW, outR, ptyOut windows.Handle
	if err = windows.CreatePipe(&ptyIn, &inW, nil, 0); err != nil {
		return nil, fmt.Errorf("입력 파이프 생성 실패: %w", err)
	}
	if err = windows.CreatePipe(&outR, &ptyOut, nil, 0); err != nil {
		windows.CloseHandle(ptyIn)
		windows.CloseHandle(inW)
		return nil, fmt.Errorf("출력 파이프 생성 실패: %w", err)
	}
	var hpc windows.Handle
	// flags=0: INHERIT_CURSOR를 주면 ConPTY가 ESC[6n 응답을 기다리며 멈출 수 있다(REF-term-query.md).
	err = windows.CreatePseudoConsole(windows.Coord{X: 80, Y: 24}, ptyIn, ptyOut, 0, &hpc)
	// ConPTY가 복제해 가졌으므로 worker 쪽 끝은 바로 닫는다 — 안 닫으면 출력 읽기에 EOF가 오지 않는다.
	windows.CloseHandle(ptyIn)
	windows.CloseHandle(ptyOut)
	inFile := os.NewFile(uintptr(inW), "conpty-in")
	outFile := os.NewFile(uintptr(outR), "conpty-out")
	cleanup = append(cleanup, func() { inFile.Close(); outFile.Close() })
	if err != nil {
		return nil, fmt.Errorf("ConPTY 생성 실패(Windows 10 1809 이상 필요): %w", err)
	}
	cleanup = append(cleanup, func() { windows.ClosePseudoConsole(hpc) })

	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("job 생성 실패: %w", err)
	}
	cleanup = append(cleanup, func() { windows.CloseHandle(job) })
	// worker가 어떻게 죽든(crash 포함) job 핸들이 닫히면 OS가 안의 프로세스를 전부 끈다.
	limit := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	if _, err = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&limit)), uint32(unsafe.Sizeof(limit))); err != nil {
		return nil, fmt.Errorf("job 설정 실패: %w", err)
	}

	attrs, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		return nil, err
	}
	defer attrs.Delete()
	// PSEUDOCONSOLE 속성 값은 HPCON 자체(포인터가 가리키는 값이 아님). 변환은 vet unsafeptr 회피용.
	if err = attrs.Update(windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE,
		*(*unsafe.Pointer)(unsafe.Pointer(&hpc)), unsafe.Sizeof(hpc)); err != nil {
		return nil, fmt.Errorf("ConPTY 속성 설정 실패: %w", err)
	}

	si := windows.StartupInfoEx{ProcThreadAttributeList: attrs.List()}
	si.Cb = uint32(unsafe.Sizeof(si))
	// worker 자신의 표준 입출력(리다이렉트된 로그 등)이 자식에 넘어가지 않게 — 자식은 ConPTY만 쓴다.
	si.Flags = windows.STARTF_USESTDHANDLES
	var pi windows.ProcessInformation
	// inheritHandles=false: worker 핸들(소켓 등)이 스크립트로 새지 않게 한다.
	if err = windows.CreateProcess(appName, cmdLine, nil, nil, false, flags, envBlock, curDir,
		&si.StartupInfo, &pi); err != nil {
		return nil, fmt.Errorf("명령어 시작 실패: %w", err)
	}
	cleanup = append(cleanup, func() {
		windows.TerminateProcess(pi.Process, 1)
		windows.CloseHandle(pi.Process)
		windows.CloseHandle(pi.Thread)
	})

	// 순서 의존: 멈춘 채(CREATE_SUSPENDED) job에 넣은 뒤 재개한다.
	// 먼저 재개하면 job에 넣기 전에 스크립트가 띄운 자식이 job 밖에 남아 Kill(true)·worker 종료로 안 꺼진다.
	if err = windows.AssignProcessToJobObject(job, pi.Process); err != nil {
		return nil, fmt.Errorf("job 등록 실패: %w", err)
	}
	if _, err = windows.ResumeThread(pi.Thread); err != nil {
		return nil, fmt.Errorf("프로세스 재개 실패: %w", err)
	}
	windows.CloseHandle(pi.Thread)

	c := &conpty{hpc: hpc, job: job, process: pi.Process, pid: int(pi.ProcessId)}
	interactive := newInteractive(sysProc{c: c})
	interactive.setStatusPending()

	outDone := make(chan struct{})
	go func() {
		streamOutputLogic(outFile, interactive.ouput)
		close(outDone)
	}()
	go streamInputLogic(inFile, interactive.input)

	exited := make(chan struct{})
	if ctx.Done() != nil {
		go func() {
			select {
			case <-ctx.Done():
				interactive.Kill(true)
			case <-exited:
			}
		}()
	}

	// 끝날 때 순서 의존(REF-windows-conpty.md "끝날 때 순서"):
	// closed 표시 -> ClosePseudoConsole -> 출력 EOF 대기 -> 핸들 닫기 -> Completed.
	// closed를 먼저 세워야 닫힌 핸들 값(OS가 재사용)으로 Layout·Kill하지 않고,
	// ClosePseudoConsole은 출력 읽기가 도는 중에 불러야 한다(옛 Windows 10은 출력이 비워질 때까지 막힘).
	// 출력 EOF 전에 Completed(큐 닫기)하면 마지막 출력이 유실된다.
	go func() {
		interactive.setStatusProcess()
		windows.WaitForSingleObject(pi.Process, windows.INFINITE)
		close(exited)
		var code uint32
		waitErr := windows.GetExitCodeProcess(pi.Process, &code)

		c.mu.Lock()
		c.closed = true
		c.mu.Unlock()
		windows.ClosePseudoConsole(hpc)

		select {
		case <-outDone:
			outFile.Close()
		case <-time.After(outputDrainTimeout):
			go outFile.Close() // 막힌 ReadFile이 끝나야 Close가 반환되므로 기다리지 않는다
		}
		inFile.Close()
		windows.CloseHandle(pi.Process)
		windows.CloseHandle(job) // KILL_ON_JOB_CLOSE — 남은 자손 정리

		if waitErr == nil && code != 0 {
			waitErr = fmt.Errorf("exit status %d", code)
		}
		interactive.setStatusCompleted(int(code), waitErr)
	}()

	return interactive, nil
}

// createEnvBlock은 "KEY=VALUE" 목록을 CreateProcess용 UTF-16 블록으로 만든다(같은 키는 뒤의 값, 대소문자 무시).
func createEnvBlock(env []string) (*uint16, error) {
	idx := map[string]int{}
	var kept []string
	for _, kv := range env {
		if strings.IndexByte(kv, 0) >= 0 {
			return nil, fmt.Errorf("환경변수에 NUL 포함: %q", kv)
		}
		k, _, _ := strings.Cut(kv, "=")
		if kv != "" && kv[0] == '=' { // "=C:=C:\dir" 같은 드라이브 항목
			k, _, _ = strings.Cut(kv[1:], "=")
			k = "=" + k
		}
		key := strings.ToUpper(k)
		if j, ok := idx[key]; ok {
			kept[j] = kv
			continue
		}
		idx[key] = len(kept)
		kept = append(kept, kv)
	}
	block := utf16.Encode([]rune(strings.Join(kept, "\x00") + "\x00\x00"))
	return &block[0], nil
}
