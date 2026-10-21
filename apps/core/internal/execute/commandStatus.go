package execute

type CommandStatus uint

const (
	CommandPending CommandStatus = iota
	CommandProcess
	CommandCompleted
	CommandFailed
)

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
