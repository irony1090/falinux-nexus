// Package pty는 로컬 PTY 기반 process 실행을 제공한다(OS별 구현 = execInteractive_{linux,windows}.go).
// 이식 가능한 execute 패키지(IInteractive·AgentInteractive·CommandStatus)와 분리되어,
// supervisor 등 PTY가 필요 없는 빌드가 execute만 import할 수 있게 한다.
package pty

import (
	"bytes"
	"io"
	"nexus/internal/execute"
	"nexus/internal/syncProcess"
	"nexus/internal/util"
)

// 컴파일 시 *Interactive가 execute.IInteractive를 만족하는지 확인.
var _ execute.IInteractive = (*Interactive)(nil)

// Interactive의 Kill·Layout·Pid·ExecInteractive는 OS별 파일에 있다.
type Interactive struct {
	Error    error
	exitCode int
	status   *syncProcess.SyncData[execute.CommandStatus]
	ouput    *syncProcess.SyncData[[]byte]
	input    *syncProcess.SyncData[[]byte]
	sys      sysProc
}

func newInteractive(sys sysProc) *Interactive {
	return &Interactive{
		exitCode: -1,
		status:   syncProcess.NewSyncData([]execute.CommandStatus{}),
		ouput:    syncProcess.NewSyncData([][]byte{}),
		input:    syncProcess.NewSyncData([][]byte{}),
		sys:      sys,
	}
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

func (i *Interactive) Refresh() error {
	i.Write([]byte{0x0c})
	// err := syscall.Kill(-i.cmd.Process.Pid, syscall.SIGWINCH)
	// log.Printf("REFRESH~ %v", err)
	return nil
}
func (i *Interactive) ExitCode() int {
	return i.exitCode
}

func streamInputLogic(master io.Writer, input *syncProcess.SyncData[[]byte]) {
	for {
		data, err := input.Shift()
		if err != nil {
			break
		}
		master.Write(data)
	}
}

func streamOutputLogic(master io.Reader, output *syncProcess.SyncData[[]byte]) {
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
