package hold

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const readyWait = 10 * time.Second

// logPath는 보관 프로세스 stderr 파일이다 — 정상 종료하면 보관 프로세스가 지운다(남아 있으면 비정상 종료 흔적)
func logPath(sock string) string { return strings.TrimSuffix(sock, ".sock") + ".log" }

// Cleanup은 끝났는데 스스로 지우지 못한(강제 종료된) 보관 프로세스의 소켓·로그를 지운다
func Cleanup(sock string) {
	os.Remove(sock)
	os.Remove(logPath(sock))
}

// Spawn은 exe를 `-hold <sock>` 모드로 띄워 spec을 실행시킨다. 반환 = 보관 프로세스 pid, 스크립트 pid
func Spawn(exe, sock string, spec Spec) (holderPid, pid int, err error) {
	body, err := json.Marshal(spec)
	if err != nil {
		return 0, 0, err
	}
	inR, inW, err := os.Pipe()
	if err != nil {
		return 0, 0, err
	}
	defer inR.Close()
	outR, outW, err := os.Pipe()
	if err != nil {
		inW.Close()
		return 0, 0, err
	}
	defer outR.Close()
	logf, err := os.OpenFile(logPath(sock), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		inW.Close()
		outW.Close()
		return 0, 0, err
	}
	defer logf.Close()

	cmd := exec.Command(exe, "-hold", sock)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = inR, outW, logf
	cmd.SysProcAttr = detachAttr()
	err = cmd.Start()
	outW.Close()
	if err != nil {
		inW.Close()
		return 0, 0, err
	}
	go cmd.Wait() // worker가 살아 있는 동안 끝난 보관 프로세스를 거둔다(worker가 꺼지면 init이 거둠)

	_, err = inW.Write(body)
	inW.Close()
	if err != nil {
		cmd.Process.Kill()
		return 0, 0, err
	}

	line := make(chan []byte, 1)
	go func() {
		b, _ := bufio.NewReader(outR).ReadBytes('\n')
		line <- b
	}()
	var r ready
	select {
	case b := <-line:
		if err := json.Unmarshal(b, &r); err != nil {
			cmd.Process.Kill()
			return 0, 0, fmt.Errorf("보관 프로세스 시작 실패(응답 %q)", b)
		}
	case <-time.After(readyWait):
		cmd.Process.Kill()
		return 0, 0, errors.New("보관 프로세스 시작 시간 초과")
	}
	if r.Err != "" {
		os.Remove(logPath(sock))
		return 0, 0, errors.New(r.Err)
	}
	return cmd.Process.Pid, r.Pid, nil
}
