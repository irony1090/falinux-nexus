package execute

import (
	"nexus/internal/syncProcess"
	"sync"
	"syscall"
)

type AgentInteractive struct {
	exitCode  int
	output    *syncProcess.SyncData[[]byte]
	status    *syncProcess.SyncData[CommandStatus]
	onWrite   func(data []byte) error
	onKill    func() error
	onLayout  func(cols, rows uint16)
	onceClose sync.Once
}

func NewAgentInteractive(
	onWrite func(data []byte) error,
	onLayout func(cols, rows uint16),
	onKill func() error,
) *AgentInteractive {
	ai := &AgentInteractive{
		exitCode: -1,
		output:   syncProcess.NewSyncData([][]byte{}),
		status:   syncProcess.NewSyncData([]CommandStatus{}),
		onWrite:  onWrite,
		onKill:   onKill,
		onLayout: onLayout,
	}
	// ai.status.Push(CommandPending)
	return ai
}

func (a *AgentInteractive) Output() ([]byte, error) {
	return a.output.Shift()
}

func (a *AgentInteractive) Write(data []byte) error { return a.onWrite(data) }
func (a *AgentInteractive) Status() (CommandStatus, int, error) {
	sts, err := a.status.Shift()
	return sts, a.exitCode, err
}
func (a *AgentInteractive) Kill() error { return a.onKill() }
func (a *AgentInteractive) Layout(cols, rows uint16) syscall.Errno {
	a.onLayout(cols, rows)
	return 0
}

func (a *AgentInteractive) PushOutput(data []byte) {
	// log.Printf("[AGENT_PUSH] %s", data)
	a.output.Push(data)
}

func (a *AgentInteractive) PushStatus(status CommandStatus) {
	// log.Printf("[AGENT_STATUS_PUSH] %s", status)
	a.status.Push(status)
}

func (a *AgentInteractive) Done(exitCode int) {
	a.onceClose.Do(func() {
		a.exitCode = exitCode
		// log.Printf("[AGENT_STATUS] %d", exitCode)
		if exitCode == 0 {
			a.status.Push(CommandCompleted)
		} else {
			a.status.Push(CommandFailed)
		}
		a.status.Close()
		a.output.Close()
	})
}

func (a *AgentInteractive) ExitCode() int {
	return a.exitCode
}
