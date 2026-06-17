package execute

import (
	"fmt"
	"hash/crc32"
	"nexus/internal/syncProcess"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

type TmuxHook string

const (
	HookSessionCreated = "session-created"
	HookSessionClosed  = "session-closed"
	HookClientAttached = "client-attached"
	HookClientDetached = "client-detached"
)

type Box struct {
	Cols int
	Rows int
}
type TmuxSession struct {
	id  string
	box Box

	error    error
	exitCode int
	status   *syncProcess.SyncData[CommandStatus]
	mu       sync.Mutex
	// ouput    *syncProcess.SyncData[[]byte]
	// input    *syncProcess.SyncData[[]byte]
}

func (t *TmuxSession) SetHookRunShell(hookType TmuxHook, command string) error {
	hookCmd := fmt.Sprintf("run-shell \"%s\"", command)
	// err := exec.Command("tmux", "set-hook", "-u", "-t", t.id, string(hookType)).Run()
	// log.Printf("[삭제 에러] %v", err)
	// cmd := exec.Command("tmux", "set-hook", "-g", string(hookType), hookCmd)
	// cmd := exec.Command("tmux", "set-hook", "-t", t.id, string(hookType), hookCmd)
	// log.Printf("[HOOK CMD] %s", cmd)
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	log.Printf("[HOOK ERR] %v, output: %s", err, string(output))
	// }

	err := exec.Command("tmux", "set-hook", "-t", t.id, string(hookType), hookCmd).Run()
	return err
}

func (t *TmuxSession) Kill() error {
	if t.exitCode != -1 { // 이미 종료됨
		return nil
	}
	if t.IsAlive() {
		cmd := exec.Command("tmux", "kill-session", "-t", t.id)
		err := cmd.Run()
		t.setStatusCompleted(cmd.ProcessState.ExitCode(), err)
		return err
	} else {
		t.setStatusCompleted(0, nil)
		return nil
	}
}

// IsAlive - 세션 생존 확인
func (t *TmuxSession) IsAlive() bool {
	return exec.Command("tmux", "has-session", "-t", t.id).Run() == nil
}

//	func (t *TmuxSession) setStatusPending() {
//		t.status.Push(CommandPending)
//	}
func (t *TmuxSession) setStatusProcess() {
	t.status.Push(CommandProcess)
}
func (t *TmuxSession) setStatusCompleted(exitCode int, err error) {
	defer t.status.Close()
	// defer t.ouput.Close()
	// defer t.input.Close()

	t.error = err
	t.exitCode = exitCode
	if exitCode == 0 {
		t.status.Push(CommandCompleted)
	} else {
		t.status.Push(CommandFailed)
	}

}

// Resize - 세션 크기 변경 (모든 클라이언트에 적용)
func (t *TmuxSession) Resize(box Box) error {
	t.box = box
	return exec.Command("tmux", "resize-window", "-t", t.id,
		"-x", strconv.Itoa(t.box.Cols),
		"-y", strconv.Itoa(t.box.Rows),
	).Run()
}

func (t *TmuxSession) CreateClient() (*TmuxClient, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return nil, nil
}
func (t *TmuxSession) RemoveClient(client *TmuxClient) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return nil
}

func (t *TmuxSession) Id() string {
	return t.id
}

func NewTmuxSession(id string, box Box, command string, args ...string) (*TmuxSession, error) {

	fullCmd := command
	// sessionId := fmt.Sprintf("tmux_sess-%s", id)
	// sessionId += "-" + command
	for _, arg := range args {
		fullCmd += " " + arg
		// sessionId += "_" + arg
	}
	// sessionId := fmt.Sprintf("tmux_sess-%s-%x", id, md5.Sum([]byte(fullCmd)))
	hash := crc32.ChecksumIEEE([]byte(fullCmd))
	sessionId := fmt.Sprintf("tmux_sess-%s-%x", id, hash)

	if hasSessionName(sessionId) {
		deleteSessionName(sessionId)
	}

	cmd := exec.Command("tmux", "new-session", "-d",
		"-s", sessionId,
		"-x", strconv.Itoa(box.Cols),
		"-y", strconv.Itoa(box.Rows),
		fullCmd,
	)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tmux 세션 생성 실패: %w", err)
	}
	sess := &TmuxSession{
		id:       sessionId,
		box:      box,
		exitCode: -1,
		status:   syncProcess.NewSyncData([]CommandStatus{}),
	}
	sess.setStatusProcess()

	// 세션 종료 감지 (polling 방식)
	// go func() {
	// 	for {
	// 		time.Sleep(5 * time.Second)
	// 		if !sess.IsAlive() {
	// 			sess.setStatusCompleted(0, nil)
	// 			break
	// 		}
	// 	}
	// }()

	return sess, nil
}

type TmuxClient struct {
	master *os.File
	cmd    *exec.Cmd
}

func hasSessionName(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func deleteSessionName(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil

}
