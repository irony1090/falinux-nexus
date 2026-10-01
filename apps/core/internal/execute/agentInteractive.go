package execute

import (
	"bytes"
	"nexus/internal/syncProcess"
	"sync"
)

type AgentInteractive struct {
	exitCode  int
	output    *syncProcess.SyncData[[]byte]
	status    *syncProcess.SyncData[CommandStatus]
	onWrite   func(data []byte) error
	onKill    func() error
	onLayout  func(cols, rows uint16) error
	onceClose sync.Once
	detached  bool // Done 대신 Detach로 닫힘(종료 아님)
}

func NewAgentInteractive(
	onWrite func(data []byte) error,
	onLayout func(cols, rows uint16) error,
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

func (a *AgentInteractive) OutputAll() ([]byte, error) {
	batch, err := a.output.ShiftAll()
	if err != nil {
		return nil, err
	}
	return bytes.Join(batch, nil), nil
}

func (a *AgentInteractive) Write(data []byte) error { return a.onWrite(data) }
func (a *AgentInteractive) Status() (CommandStatus, int, error) {
	sts, err := a.status.Shift()
	return sts, a.exitCode, err
}
func (a *AgentInteractive) Kill() error { return a.onKill() }
func (a *AgentInteractive) Layout(cols, rows uint16) error {
	return a.onLayout(cols, rows)
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

// Detach는 종료 STATUS 없이 채널만 닫는다(worker 끊김 = 종료 아님). Done과 onceClose를 공유해 둘 중 먼저 온 것만 적용된다.
// 이미 Push된 출력·상태는 relay가 마저 드레인한다.
func (a *AgentInteractive) Detach() {
	a.onceClose.Do(func() {
		a.detached = true
		a.status.Close()
		a.output.Close()
	})
}

// Detached는 Detach로 닫혔는지 반환한다(relay 드레인 이후에 읽을 것).
func (a *AgentInteractive) Detached() bool {
	return a.detached
}

func (a *AgentInteractive) ExitCode() int {
	return a.exitCode
}
