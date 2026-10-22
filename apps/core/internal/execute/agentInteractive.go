package execute

import (
	"nexus/internal/syncProcess"
	"sync"
)

// Chunk는 worker에서 온 출력 한 묶음이다. Off = 끝 offset(0 = 모름), Seed = 화면 스냅샷(MsgScreen)
type Chunk struct {
	Data []byte
	Off  int64
	Seed bool
	Alt  bool
}

type AgentInteractive struct {
	exitCode  int
	output    *syncProcess.SyncData[Chunk]
	status    *syncProcess.SyncData[CommandStatus]
	onWrite   func(data []byte) error
	onKill    func(force bool) error
	onLayout  func(cols, rows uint16) error
	onceClose sync.Once
	detached  bool // Done 대신 Detach로 닫힘(종료 아님)
}

func NewAgentInteractive(
	onWrite func(data []byte) error,
	onLayout func(cols, rows uint16) error,
	onKill func(force bool) error,
) *AgentInteractive {
	ai := &AgentInteractive{
		exitCode: -1,
		output:   syncProcess.NewSyncData([]Chunk{}),
		status:   syncProcess.NewSyncData([]CommandStatus{}),
		onWrite:  onWrite,
		onKill:   onKill,
		onLayout: onLayout,
	}
	// ai.status.Push(CommandPending)
	return ai
}

func (a *AgentInteractive) OutputAll() ([]byte, error) {
	batch, err := a.output.ShiftAll()
	if err != nil {
		return nil, err
	}
	var out []byte
	for _, c := range batch {
		if !c.Seed {
			out = append(out, c.Data...)
		}
	}
	return out, nil
}

// OutputChunks는 대기 중인 묶음을 offset·스냅샷 구분과 함께 꺼낸다(relay용 — 순서 유지)
func (a *AgentInteractive) OutputChunks() ([]Chunk, error) {
	return a.output.ShiftAll()
}

func (a *AgentInteractive) Write(data []byte) error { return a.onWrite(data) }
func (a *AgentInteractive) Status() (CommandStatus, int, error) {
	sts, err := a.status.Shift()
	return sts, a.exitCode, err
}
func (a *AgentInteractive) Kill(force bool) error { return a.onKill(force) }
func (a *AgentInteractive) Layout(cols, rows uint16) error {
	return a.onLayout(cols, rows)
}

func (a *AgentInteractive) PushOutput(c Chunk) {
	a.output.Push(c)
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
