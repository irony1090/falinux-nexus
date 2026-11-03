package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"nexus/internal/updproto"
)

// 단계 하나 — staging 받기(같은 Stage면 다시 씀) -> builtin 또는 script -> 결과. REF-update-p2.md 단위-5(받기·builtin)·단위-6(script 실행)

const (
	defaultTimeout = 10 * time.Minute // 비대화형 script 기본 시간 제한
	keepStages     = 3
)

type job struct {
	a     *agent
	run   updproto.Run
	ctx   context.Context
	stop  context.CancelFunc
	logf  *os.File
	stage string // staging 폴더

	mu   sync.Mutex
	why  string // 멈춘 이유(cancel·timeout)
	proc procCtl
	pend []byte // proc이 생기기 전에 온 입력
	cols uint16
	rows uint16
}

// procCtl: 실행 중인 script(입력·크기·종료)
type procCtl interface {
	write([]byte)
	resize(cols, rows uint16)
	kill(force bool)
}

func newJob(a *agent, r updproto.Run) (*job, error) {
	if r.Stage == "" || strings.ContainsAny(r.Stage, `/\.`) {
		return nil, errors.New("stage 값이 이상합니다")
	}
	for _, f := range r.Files {
		if !filepath.IsLocal(filepath.FromSlash(f.Path)) {
			return nil, fmt.Errorf("파일 경로 %q가 staging 밖을 가리킵니다", f.Path)
		}
	}
	lf, err := a.logFile(r.RunID)
	if err != nil {
		return nil, err
	}
	ctx, stop := context.WithCancel(context.Background())
	name := r.Stage
	if len(name) > 16 {
		name = name[:16]
	}
	return &job{a: a, run: r, ctx: ctx, stop: stop, logf: lf, stage: filepath.Join(a.dir, "staging", r.Version+"-"+name), cols: r.Cols, rows: r.Rows}, nil
}

func (j *job) out(b []byte) { j.a.out(j.logf, j.run.RunID, b) }

func (j *job) say(f string, args ...any) {
	j.out([]byte("[" + fmt.Sprintf(f, args...) + "]\r\n"))
}

func (j *job) cancel(why string) {
	j.mu.Lock()
	if j.why == "" {
		j.why = why
	}
	p := j.proc
	j.mu.Unlock()
	j.stop()
	if p != nil {
		p.kill(false)
		time.AfterFunc(5*time.Second, func() { p.kill(true) })
	}
}

func (j *job) input(b []byte) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.proc != nil {
		j.proc.write(b)
	} else if j.run.Step.Interactive {
		j.pend = append(j.pend, b...)
	}
}

func (j *job) resize(cols, rows uint16) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.cols, j.rows = cols, rows
	if j.proc != nil {
		j.proc.resize(cols, rows)
	}
}

func (j *job) setProc(p procCtl) {
	j.mu.Lock()
	j.proc = p
	pend := j.pend
	j.pend = nil
	cols, rows := j.cols, j.rows
	stopped := j.why != ""
	j.mu.Unlock()
	if cols > 0 && rows > 0 {
		p.resize(cols, rows)
	}
	if len(pend) > 0 {
		p.write(pend)
	}
	if stopped {
		p.kill(true)
	}
}

// main: 결과를 돌려주고 supervisor에도 보낸다(배포는 결과를 보고 다음 단계로)
func (j *job) main() (e updproto.End) {
	e = updproto.End{RunID: j.run.RunID}
	defer func() {
		if v := recover(); v != nil {
			e.Result, e.Msg = updproto.ResultFail, fmt.Sprintf("updater 내부 오류: %v", v)
		}
		j.logf.Close()
		j.stop()
		j.a.finish(j, e)
	}()
	e = j.do()
	j.mu.Lock()
	if j.why != "" && e.Result != updproto.ResultOK {
		e.Result = j.why
	}
	j.mu.Unlock()
	return e
}

func (j *job) do() updproto.End {
	e := updproto.End{RunID: j.run.RunID}
	if err := j.fetch(); err != nil {
		e.Result, e.Msg = updproto.ResultFail, "파일 받기 실패: "+err.Error()
		return e
	}
	s := j.run.Step
	var err error
	switch s.Kind {
	case updproto.KindBuiltin:
		err = j.builtin(s.Builtin)
	case updproto.KindScript:
		var code int
		code, err = j.script()
		e.ExitCode = &code
		if err == nil && code != 0 {
			err = fmt.Errorf("종료 코드 %d", code)
		}
	default:
		err = fmt.Errorf("모르는 단계 종류 %q", s.Kind)
	}
	if err != nil {
		e.Result, e.Msg = updproto.ResultFail, err.Error()
		return e
	}
	e.Result = updproto.ResultOK
	return e
}

// ===== staging =====

// fetch: 합친 파일을 staging/<버전>-<stage>/에 — sha가 같은 파일은 건너뜀, 받은 뒤 sha·크기 확인, 권한 적용
func (j *job) fetch() error {
	if err := os.MkdirAll(j.stage, 0o755); err != nil {
		return err
	}
	_ = os.Chtimes(j.stage, time.Now(), time.Now()) // 최근 사용 = 지우기 순서
	_, httpBase, _ := endpoints(j.a.cfg.get().Supervisor)
	got := 0
	for _, f := range j.run.Files {
		if err := j.ctx.Err(); err != nil {
			return err
		}
		dst := filepath.Join(j.stage, filepath.FromSlash(f.Path))
		if sum, err := fileSha(dst); err == nil && sum == f.Sha {
			_ = os.Chmod(dst, os.FileMode(f.Mode))
			continue
		}
		j.say("받기 %s (%s)", f.Path, sizeText(f.Size))
		if err := j.download(httpBase+f.URL, dst, f); err != nil {
			return fmt.Errorf("%s: %w", f.Path, err)
		}
		got++
	}
	j.say("staging %s · 파일 %d개(새로 받음 %d)", filepath.Base(j.stage), len(j.run.Files), got)
	pruneDirs(filepath.Dir(j.stage), keepStages, j.stage)
	return nil
}

func (j *job) download(url, dst string, f updproto.StageFile) error {
	req, err := http.NewRequestWithContext(j.ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	c := j.a.cfg.get()
	req.Header.Set(updproto.HeaderMain, c.MainKey)
	req.Header.Set(updproto.HeaderKey, c.APIKey)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return fmt.Errorf("HTTP %d %s", res.StatusCode, strings.TrimSpace(string(b)))
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".dl-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(tmp, h), res.Body)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	if n != f.Size || hex.EncodeToString(h.Sum(nil)) != f.Sha {
		return fmt.Errorf("받은 파일이 다릅니다(크기 %d/%d, sha 불일치)", n, f.Size)
	}
	if err := os.Chmod(tmp.Name(), os.FileMode(f.Mode)); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dst)
}

func fileSha(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// pruneDirs: 최근에 쓴 keep개만 남김(keepAlso는 항상 남김)
func pruneDirs(parent string, keep int, keepAlso string) {
	ents, err := os.ReadDir(parent)
	if err != nil {
		return
	}
	type d struct {
		path string
		t    time.Time
	}
	var ds []d
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		ds = append(ds, d{filepath.Join(parent, e.Name()), info.ModTime()})
	}
	sort.Slice(ds, func(a, b int) bool { return ds[a].t.After(ds[b].t) })
	for i, x := range ds {
		if i >= keep && x.path != keepAlso {
			_ = os.RemoveAll(x.path)
		}
	}
}

func sizeText(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1fMB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1fKB", float64(n)/(1<<10))
	}
	return fmt.Sprintf("%dB", n)
}

// ===== script =====

// script: 정-6(콘솔) — interactive면 PTY(입력 중계, 시간 제한 없음), 아니면 stdin 없음 + 시간 제한
func (j *job) script() (int, error) {
	s := j.run.Step
	path := filepath.Join(j.stage, filepath.FromSlash(s.Script))
	if !filepath.IsLocal(filepath.FromSlash(s.Script)) {
		return -1, fmt.Errorf("script 경로 %q가 staging 밖입니다", s.Script)
	}
	if _, err := os.Stat(path); err != nil {
		return -1, fmt.Errorf("script %s가 빌드 파일에 없습니다", s.Script)
	}
	env := append(os.Environ(), j.stepEnv()...)
	if !s.Interactive {
		limit := defaultTimeout
		if s.TimeoutSec > 0 {
			limit = time.Duration(s.TimeoutSec) * time.Second
		}
		t := time.AfterFunc(limit, func() {
			j.say("시간 제한 %s 넘음 — 종료", limit)
			j.cancel(updproto.ResultTimeout)
		})
		defer t.Stop()
	}
	j.say("실행 %s%s", s.Script, map[bool]string{true: " (입력 받음)", false: ""}[s.Interactive])
	if s.Interactive {
		return runPTY(j, path, append(env, "TERM=xterm-256color"))
	}
	return runPipe(j, path, env)
}

// stepEnv: 단계에 넘기는 환경변수 — REF-update-deploy.md "steps.json"
func (j *job) stepEnv() []string {
	env := []string{
		"TARGET_DIR=" + j.a.cfg.get().TargetDir,
		"VERSION=" + j.run.Version,
		"STEP_ID=" + j.run.Step.ID,
		"STAGE_DIR=" + j.stage,
		"DEPLOY_ID=" + j.run.DeployID,
	}
	if b, err := j.findBackup(); err == nil {
		env = append(env, "BACKUP_DIR="+b, "NEW_FILES="+filepath.Join(b, "new-files.txt"))
	}
	return env
}

// outWriter: script 출력 -> out
type outWriter struct{ j *job }

func (w outWriter) Write(b []byte) (int, error) {
	w.j.out(append([]byte(nil), b...))
	return len(b), nil
}
