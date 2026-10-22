package router

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"nexus/internal/execute"
	"nexus/internal/execute/hold"
	"nexus/internal/protocol"
	"nexus/internal/transport"
	workerdb "nexus/internal/worker/db/gen"
)

// ===== 보관 프로세스 (REF-process-keep.md) =====
// worker -> 보관 프로세스(worker -hold) -> 스크립트. 목록 = sqlite holds(살아 있는 것만), worker가 다시 켜지면 Recover로 다시 붙는다

const exitLost = 502 // 보관 프로세스와 연결이 끊겨 결과를 모름(supervisor "DB만" 종결과 같은 값)

var unsafeName = regexp.MustCompile(`[^A-Za-z0-9._-]`)

// unix 소켓 경로 상한(Linux 108바이트, NUL 포함)에서 "/<uid 16자>.sock" 자리를 뺀 폴더 길이
const maxHoldDir = 107 - len("/0123456789abcdef.sock")

// HoldDir은 보관 프로세스 소켓 폴더다(메인키별 — 개발·빌드본 worker가 섞이지 않게). 재부팅 때 비는 임시 폴더
func HoldDir(mainKey string) (string, error) {
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("nexus-hold-%d", os.Getuid()), unsafeName.ReplaceAllString(mainKey, "_"))
	if len(dir) > maxHoldDir {
		return "", fmt.Errorf("소켓 폴더 경로가 너무 김(%d > %d): %s — TMPDIR을 짧은 경로로", len(dir), maxHoldDir, dir)
	}
	return dir, os.MkdirAll(dir, 0o700)
}

// admit은 process를 procs에 넣는다(같은 uid가 있으면 거부)
func (s *WorkerState) admit(entry *procEntry) (ok bool, reason string) {
	if !s.procs.Append(entry.uid, entry) {
		return false, "중복되는 uid"
	}
	return true, ""
}

// spawn은 보관 프로세스를 띄워 붙고 목록(procs·sqlite)에 넣는다
func (s *WorkerState) spawn(spec hold.Spec, kind protocol.ExecType, editPath string) (*procEntry, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	sock := filepath.Join(s.holdDir, spec.UID+".sock")
	holderPid, pid, err := hold.Spawn(exe, sock, spec)
	if err != nil {
		return nil, err
	}
	cli, err := hold.Dial(sock)
	if err != nil {
		hold.KillHolder(holderPid)
		hold.Cleanup(sock)
		return nil, fmt.Errorf("보관 프로세스 연결 실패: %w", err)
	}
	entry := &procEntry{uid: spec.UID, kind: kind, editPath: editPath, pid: pid, h: cli}
	if ok, reason := s.admit(entry); !ok {
		cli.Kill(true)
		cli.Close()
		return nil, fmt.Errorf("%s", reason)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.store.Queries().InsertHold(ctx, workerdb.InsertHoldParams{
		Uid: spec.UID, Kind: string(kind), EditPath: editPath, Sock: sock, HolderPid: int64(holderPid), Pid: int64(pid),
	}); err != nil {
		log.Printf("[hold] 목록 저장 실패 uid=%s: %v", spec.UID, err) // 실행은 계속(worker가 다시 켜지면 못 붙을 뿐)
	}
	return entry, nil
}

// Recover는 worker 시작 때 sqlite 목록의 보관 프로세스에 다시 붙는다. 못 붙으면 남은 보관 프로세스를 정리하고 목록에서 뺀다
// (supervisor는 sync 보고에 없는 uid를 끝남으로 닫는다 — K7(끝난 뒤 기다림))
func (s *WorkerState) Recover() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := s.store.Queries().ListHolds(ctx)
	if err != nil {
		log.Printf("[hold] 목록 읽기 실패: %v", err)
		return
	}
	for _, row := range rows {
		cli, err := hold.Dial(row.Sock)
		if err != nil {
			if hold.IsHolder(int(row.HolderPid), row.Sock) {
				log.Printf("[hold] 소켓에 못 붙는 보관 프로세스 정리 uid=%s pid=%d: %v", row.Uid, row.HolderPid, err)
				hold.KillHolder(int(row.HolderPid))
			}
			hold.Cleanup(row.Sock)
			s.dropHold(row.Uid)
			continue
		}
		entry := &procEntry{uid: row.Uid, kind: protocol.ExecType(row.Kind), editPath: row.EditPath, pid: int(row.Pid), h: cli}
		if ok, _ := s.admit(entry); !ok {
			cli.Close()
			continue
		}
		log.Printf("[hold] 다시 붙음 uid=%s pid=%d", row.Uid, row.Pid)
		go s.forward(entry)
	}
}

func (s *WorkerState) dropHold(uid string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.store.Queries().DeleteHold(ctx, uid); err != nil {
		log.Printf("[hold] 목록 삭제 실패 uid=%s: %v", uid, err)
	}
}

// forward는 보관 프로세스의 출력·스냅샷·상태를 supervisor로 순서대로 보낸다(process당 고루틴 하나 — 순서가 곧 offset 순서).
// 전송은 currentConn(재접속 시 교체) — 끊긴 동안은 버리고, 다시 붙으면 register가 요청한 스냅샷으로 메운다
func (s *WorkerState) forward(e *procEntry) {
	if st := e.h.Hello; st.Status.IsCompleted() {
		s.finish(e, st.Status, st.Exit)
		return
	} else if st.Status == execute.CommandProcess {
		s.emit(protocol.MsgStatus, protocol.StatusEvent{UID: e.uid, Status: st.Status, PID: e.pid})
	}
	for m := range e.h.Events() {
		switch m.T {
		case hold.MsgData:
			s.emit(protocol.MsgData, protocol.DataEvent{UID: e.uid, Data: m.Data, Off: m.Off})
		case hold.MsgSnap:
			s.emit(protocol.MsgScreen, protocol.ScreenEvent{UID: e.uid, Data: m.Data, Off: m.Off, Alt: m.Alt})
		case hold.MsgStatus:
			if m.Status.IsCompleted() {
				s.finish(e, m.Status, m.Exit)
				return
			}
			s.emit(protocol.MsgStatus, protocol.StatusEvent{UID: e.uid, Status: m.Status, PID: e.pid})
		}
	}
	log.Printf("[hold] 보관 프로세스 연결 끊김 uid=%s", e.uid)
	s.finish(e, execute.CommandFailed, exitLost)
}

func (s *WorkerState) emit(t protocol.MsgType, ev any) {
	conn := s.currentConn()
	if conn == nil {
		return
	}
	if err := conn.Emit(t, ev); err != nil {
		log.Printf("[hold] emit %s: %v", t, err)
	}
}

// finish는 끝난 process의 종료를 보고하고 정리한다. 종료는 놓치면 안 돼서 register를 마친 연결이 생길 때까지 기다린다
// (worker 시작 직후 Recover한 끝난 process — sync 보고 뒤에 보내야 supervisor가 받는다). EDIT면 파일을 읽어 회수를 보고
func (s *WorkerState) finish(e *procEntry, st execute.CommandStatus, exit int) {
	conn := s.waitReady()
	if err := conn.Emit(protocol.MsgStatus, protocol.StatusEvent{UID: e.uid, Status: st, ExitCode: exit}); err != nil {
		log.Printf("[hold] emit status uid=%s: %v", e.uid, err)
	}
	if e.kind == protocol.ExecTypeEdit {
		if content, err := os.ReadFile(e.editPath); err != nil {
			log.Printf("[hold] edit read-back uid=%s: %v", e.uid, err)
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if _, err := conn.Call(ctx, protocol.MsgEditResult, protocol.EditResult{UID: e.uid, Content: content}); err != nil {
				log.Printf("[hold] MsgEditResult uid=%s: %v", e.uid, err)
			}
			cancel()
		}
	}
	s.procs.Remove(e.uid)
	s.dropHold(e.uid)
	e.h.Release()
	e.h.Close()
}

func (s *WorkerState) waitReady() *transport.Conn {
	for {
		if conn := s.ready.Load(); conn != nil {
			return conn
		}
		time.Sleep(300 * time.Millisecond)
	}
}
