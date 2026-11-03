package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/transport"
	"nexus/internal/updproto"
)

// 배포 — 목록을 updater가 차례로 실행(설계-10(자기 교체) 10-2(순서 주체)). REF-update-p3.md 단위-3(updater 배포 실행)
// 단계 기록은 supervisor가 미리 만든다(wait). 실패·취소면 멈추고, autoRollback이면 rollback 목록을 이어서.
// updater가 재시작되면 진행 중이던 배포는 잊는다 — supervisor가 HELLO Deploying에 없는 배포를 중단으로 기록

type deployRun struct {
	req     updproto.Deploy
	stopped atomic.Bool // 취소 — 지금 단계가 끝나면 다음으로 넘어가지 않음
}

func (a *agent) busyErr() error {
	if a.dep != nil {
		return fmt.Errorf("배포 %d가 실행 중입니다", a.dep.req.DeployID)
	}
	return fmt.Errorf("다른 단계(기록 %d)가 실행 중입니다", a.job.run.RunID)
}

func (a *agent) onDeploy(f protocol.Frame) (any, error) {
	var d updproto.Deploy
	if err := f.Bind(&d); err != nil {
		return nil, err
	}
	if d.List != updproto.ListDeploy && d.List != updproto.ListRollback {
		return nil, fmt.Errorf("모르는 목록 %q", d.List)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.job != nil || a.dep != nil {
		return nil, a.busyErr()
	}
	r := &deployRun{req: d}
	a.dep = r
	go a.runDeploy(r)
	return nil, nil
}

func (a *agent) deploying() []int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	var ids []int64
	if a.dep != nil {
		ids = append(ids, a.dep.req.DeployID)
	}
	for _, e := range a.deployEnds {
		ids = append(ids, e.DeployID)
	}
	return ids
}

func (a *agent) runDeploy(r *deployRun) {
	d := r.req
	end := updproto.DeployEnd{DeployID: d.DeployID, List: d.List}
	defer func() {
		if v := recover(); v != nil {
			end.Result, end.Msg = updproto.ResultFail, fmt.Sprintf("updater 내부 오류: %v", v)
		}
		end.Installed = a.state().Installed
		a.finishDeploy(r, end)
	}()
	end.Result, end.Msg = a.runList(r, d.Steps)
	if end.Result == updproto.ResultFail && len(d.Rollback) > 0 && !r.stopped.Load() {
		log.Printf("배포 %d 실패 — 자동 되돌리기", d.DeployID)
		end.Rollback, _ = a.runList(r, d.Rollback)
	}
	// 설치본 이름(설계-6(버전 이름 체계)): 끝까지 성공한 배포 = 그 버전, 되돌리기 성공 = 백업 때 이름
	switch {
	case d.List == updproto.ListDeploy && end.Result == updproto.ResultOK:
		a.editState(func(s *state) { s.Installed = d.Version })
	case (d.List == updproto.ListRollback && end.Result == updproto.ResultOK) || end.Rollback == updproto.ResultOK:
		if m, err := readManifest(filepath.Join(a.dir, "backups", updproto.DeployTag(d.DeployID))); err == nil {
			a.editState(func(s *state) { s.Installed = m.PrevInstalled })
		}
	}
}

// runList: 단계 결과 = 목록 결과(timeout은 fail). 멈추면 남은 단계는 supervisor가 skip으로
func (a *agent) runList(r *deployRun, steps []updproto.DeployStep) (result, msg string) {
	d := r.req
	for _, s := range steps {
		if r.stopped.Load() {
			return updproto.ResultCancel, "취소됨"
		}
		a.emit(updproto.MsgStepStart, updproto.StepStart{DeployID: d.DeployID, RunID: s.RunID})
		run := updproto.Run{RunID: s.RunID, DeployID: updproto.DeployTag(d.DeployID), Version: d.Version, Stage: d.Stage, Files: d.Files, Step: s.Step, Replace: d.Replace, Cols: d.Cols, Rows: d.Rows}
		a.mu.Lock()
		j, err := newJob(a, run)
		if err == nil {
			a.job = j
		}
		a.mu.Unlock()
		if err != nil {
			a.finish(nil, updproto.End{RunID: s.RunID, Result: updproto.ResultFail, Msg: err.Error()})
			return updproto.ResultFail, s.Step.ID + ": " + err.Error()
		}
		e := j.main()
		switch e.Result {
		case updproto.ResultOK:
		case updproto.ResultCancel:
			return updproto.ResultCancel, s.Step.ID + " 취소됨"
		default:
			return updproto.ResultFail, s.Step.ID + " " + e.Result + stepMsg(e.Msg)
		}
	}
	return updproto.ResultOK, ""
}

func stepMsg(m string) string {
	if m == "" {
		return ""
	}
	return ": " + m
}

func (a *agent) emit(kind protocol.MsgType, v any) {
	a.mu.Lock()
	conn := a.conn
	a.mu.Unlock()
	if conn != nil {
		_ = conn.Emit(kind, v)
	}
}

func (a *agent) finishDeploy(r *deployRun, e updproto.DeployEnd) {
	a.mu.Lock()
	if a.dep == r {
		a.dep = nil
	}
	a.deployEnds = append(a.deployEnds, e)
	conn := a.conn
	a.mu.Unlock()
	log.Printf("배포 끝 %d %s: %s 되돌리기=%q %s", e.DeployID, e.List, e.Result, e.Rollback, e.Msg)
	if conn != nil {
		go a.flush(conn)
	}
}

func (a *agent) flushDeployEnds(conn *transport.Conn) {
	a.mu.Lock()
	pending := append([]updproto.DeployEnd(nil), a.deployEnds...)
	a.mu.Unlock()
	for _, e := range pending {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err := conn.Call(ctx, updproto.MsgDeployEnd, e)
		cancel()
		if err != nil {
			log.Printf("배포 결과 보내기 실패(다음 접속 때 다시) 배포 %d: %v", e.DeployID, err)
			return
		}
		a.mu.Lock()
		for i, x := range a.deployEnds {
			if x.DeployID == e.DeployID && x.List == e.List {
				a.deployEnds = append(a.deployEnds[:i], a.deployEnds[i+1:]...)
				break
			}
		}
		a.mu.Unlock()
	}
}

// ===== state.json =====

func (a *agent) state() state {
	a.stMu.Lock()
	defer a.stMu.Unlock()
	return loadState(a.statePath())
}

func (a *agent) editState(fn func(*state)) {
	a.stMu.Lock()
	defer a.stMu.Unlock()
	st := loadState(a.statePath())
	fn(&st)
	if err := saveState(a.statePath(), st); err != nil {
		log.Printf("state.json 저장 실패: %v", err)
	}
}

// readTail: 파일 끝 n바이트(잘렸으면 앞에 표시)
func readTail(name string, n int64) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	var head []byte
	if st.Size() > n {
		if _, err := f.Seek(st.Size()-n, io.SeekStart); err != nil {
			return nil, err
		}
		head = []byte(fmt.Sprintf("[앞 %d바이트 생략 — 전체는 updater 로그 %s]\r\n", st.Size()-n, name))
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(head) == 0 && len(b) == 0 {
		return nil, errors.New("빈 로그")
	}
	return append(head, b...), nil
}
