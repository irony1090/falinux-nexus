package execute

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type CommandStatus uint

const (
	CommandPending CommandStatus = iota
	CommandProcess
	CommandCompleted
	CommandFailed
)

type OutputType string

const (
	OutputTypeStdout OutputType = "OUT"
	OutputTypeStderr OutputType = "ERR"
)

type TaggedOutput struct {
	Type    OutputType
	Content string
}

func (s CommandStatus) IsCompleted() bool {
	return s == CommandCompleted || s == CommandFailed
}

func (s CommandStatus) String() string {
	switch s {
	case CommandPending:
		return "PENDING"
	case CommandProcess:
		return "PROCESS"
	case CommandCompleted:
		return "COMPLETED"
	case CommandFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// OutputStreamer는 출력 스트림을 관리합니다
type OutputStreamer struct {
	Output   chan TaggedOutput // 표준 출력/에러를 수신하는 단일 채널
	streamWg sync.WaitGroup    // 스트림 고루틴 완료 대기를 위한 WaitGroup
}

type Command struct {
	Error    error
	ExitCode int
	// Status   CommandStatus
	Status chan CommandStatus
}

func ExecCommand(ctx context.Context, command string, args ...string) (*Command, *OutputStreamer, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	// devNull, err := os.Open(os.DevNull)

	// cmd.Stdin = devNull
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stdout 파이프 생성 실패: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stderr 파이프 생성 실패: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("명령어 시작 실패: %w", err)
	}

	result := &Command{
		Error:    nil,
		ExitCode: -1,
		Status:   make(chan CommandStatus, 8),
	}
	result.Status <- CommandPending

	output := make(chan TaggedOutput)
	streamer := &OutputStreamer{
		Output: output,
	}
	streamer.streamWg.Add(2)

	mutex := &sync.Mutex{}
	go streamOutput(stdoutPipe, output, OutputTypeStdout, &streamer.streamWg, mutex)
	go streamOutput(stderrPipe, output, OutputTypeStderr, &streamer.streamWg, mutex)

	go func() {
		result.Status <- CommandProcess
		defer close(result.Status)
		defer close(output)
		err := cmd.Wait()

		// streamer.Out = nil
		// streamer.Err = nil

		if err != nil {
			result.Status <- CommandFailed
			if exitErr, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitErr.ExitCode()
			}
		} else {
			result.ExitCode = 0
			result.Status <- CommandCompleted
		}

		streamer.streamWg.Wait()
		// fmt.Println("종료!!!")

	}()

	// command := &
	return result, streamer, nil
}

func streamOutput(reader io.ReadCloser, ch chan<- TaggedOutput, outputType OutputType, wg *sync.WaitGroup, mutex *sync.Mutex) {
	defer wg.Done()
	defer reader.Close()

	buffer := make([]byte, 1024)
	line := make([]byte, 0, 1024)
	// writeMutex := &sync.Mutex{}

	for {
		n, err := reader.Read(buffer)
		// mutex.Lock()

		if n > 0 {
			line = append(line, buffer[:n]...)

			ok, newLine := pubLine(line, ch, outputType, mutex)
			if ok {
				line = newLine
			}

		} else if err != nil {
			// fmt.Printf("종료 %v\n", line)
			if len(line) > 0 {
				_, newLine := pubLine(line, ch, outputType, mutex)
				if len(newLine) > 0 {
					mutex.Lock()
					ch <- TaggedOutput{
						Type:    outputType,
						Content: string(newLine),
					}
					mutex.Unlock()
				}
			}
			// mutex.Unlock()
			break
		}
		// mutex.Unlock()
	}

}

func newLineContent(c []byte) (int, string, string) {
	// 1. \r\n 또는 \n\r 먼저 처리
	rnIndex := bytes.Index(c, rn)
	nrIndex := bytes.Index(c, nr)

	if rnIndex >= 0 || nrIndex >= 0 {
		idx := rnIndex
		if nrIndex >= 0 && (rnIndex < 0 || nrIndex < rnIndex) {
			idx = nrIndex
		}
		return idx + 2, string(c[:idx]), "\n"
	}

	// 2. 단일 \r 또는 \n 중 먼저 나오는 것 찾기
	rIndex := bytes.IndexByte(c, '\r')
	nIndex := bytes.IndexByte(c, '\n')

	// 둘 다 없으면
	if rIndex < 0 && nIndex < 0 {
		return -1, "", ""
	}

	// \n만 있거나, 둘 다 있는데 \n이 더 앞에 있으면
	if rIndex < 0 || (nIndex >= 0 && nIndex < rIndex) {
		return nIndex + 1, string(c[:nIndex]), "\n"
	}

	// \r이 더 앞에 있으면 (개행 아님, 그대로 유지)
	return rIndex + 1, string(c[:rIndex]), "\r"

}

func pubLine(content []byte, ch chan<- TaggedOutput, outputType OutputType, writeMutex *sync.Mutex) (bool, []byte) {
	c := bytes.Clone(content)
	prefix := ""
	// fmt.Printf("[pubLine-%s] START\n", outputType)
	for {
		writeMutex.Lock()

		newLineIndex, str, last := newLineContent(c)
		if newLineIndex < 0 {
			writeMutex.Unlock()
			break
		}

		var content string
		if last == "\r" {
			content = prefix + str
			prefix = last
		} else {
			content = prefix + str + last
			prefix = ""
		}
		// fmt.Printf("[%s] %s [??]\n", outputType, content)

		ch <- TaggedOutput{
			Type:    outputType,
			Content: content,
		}

		c = c[newLineIndex:]
		writeMutex.Unlock()

	}
	// fmt.Printf("[pubLine-%s] E N D\n", outputType)

	remaining := c
	if !bytes.Equal(remaining, content) {
		return true, remaining
	} else {
		return false, nil
	}
}
