package execute

import "syscall"

type IInteractive interface {
	Output() ([]byte, error)
	Write(data []byte) error
	Status() (CommandStatus, int, error)
	Kill() error
	Layout(cols, rows uint16) syscall.Errno
	ExitCode() int
}
