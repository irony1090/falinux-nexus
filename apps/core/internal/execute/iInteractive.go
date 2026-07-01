package execute

import "syscall"

type IInteractive interface {
	Output() ([]byte, error)
	// OutputAll은 대기 중인 출력을 한 번에(연결된 바이트로) 반환한다. 고빈도 스트림에서
	// 항목당 락을 잡는 Output 대신 배치 드레인해 락 경합과 프레임 수를 줄인다.
	OutputAll() ([]byte, error)
	Write(data []byte) error
	Status() (CommandStatus, int, error)
	Kill() error
	Layout(cols, rows uint16) syscall.Errno
	ExitCode() int
}
