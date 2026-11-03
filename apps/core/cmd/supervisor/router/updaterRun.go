package router

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/supervisor/updstore"
	"nexus/internal/updproto"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// 단계 하나 실행(개별 실행) — 기록 = update_step_runs(deployment_id NULL), 전체 로그 = 보관소 logs/<id>.log, 메모리 = 끝 64KB.
// P2(updater Linux)는 API로만(Run·Log 화면은 P3) — REF-update-p2.md 단위-6(script 실행)

const (
	runTailMax   = 64 << 10 // 메모리·조회 응답
	runDBTailMax = 4 << 10  // DB log_tail
)

type updRun struct {
	id          int64
	main, sub   string
	interactive bool

	mu    sync.Mutex
	f     *os.File // nil = 열지 못함(로그는 메모리만)
	tail  []byte
	total int64 // 지금까지 받은 바이트(재전송이면 그 길이부터)
	end   *updproto.End
}

// write: 누적 바이트를 돌려준다(화면 이음매 — UPD:OUTPUT off)
func (u *updRun) write(b []byte) int64 {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.f != nil {
		_, _ = u.f.Write(b)
	}
	u.total += int64(len(b))
	u.tail = append(u.tail, b...)
	if n := len(u.tail); n > runTailMax {
		u.tail = append([]byte(nil), u.tail[n-runTailMax:]...)
	}
	return u.total
}

type updRunSet struct {
	dir   string // 로그 폴더(빈 값 = 파일 없음)
	mu    sync.Mutex
	m     map[int64]*updRun
	inter map[int64]bool // 배포가 미리 만든 단계 중 interactive(시작 알림 때 입력 허용) — supervisor 재시작이면 잊음
	pub   func(main string, data updproto.Output, off int64)
}

func newUpdRunSet(storeDir string) *updRunSet {
	s := &updRunSet{m: map[int64]*updRun{}, inter: map[int64]bool{}}
	if storeDir != "" {
		s.dir = filepath.Join(storeDir, "logs")
		if err := os.MkdirAll(s.dir, 0o755); err != nil {
			log.Printf("[updater] 로그 폴더를 만들 수 없음 — 로그는 메모리만: %v", err)
			s.dir = ""
		}
	}
	return s
}

func (s *updRunSet) logPath(id int64) string {
	if s.dir == "" {
		return ""
	}
	return filepath.Join(s.dir, strconv.FormatInt(id, 10)+".log")
}

func (s *updRunSet) open(id int64, main, sub string, interactive bool) *updRun {
	u := &updRun{id: id, main: main, sub: sub, interactive: interactive}
	if p := s.logPath(id); p != "" {
		f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Printf("[updater] 로그 파일 %s: %v", p, err)
		} else {
			u.f = f
		}
	}
	s.mu.Lock()
	s.m[id] = u
	s.mu.Unlock()
	return u
}

func (s *updRunSet) get(id int64) *updRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.m[id]
}

func (s *updRunSet) drop(id int64) {
	s.mu.Lock()
	u := s.m[id]
	delete(s.m, id)
	s.mu.Unlock()
	if u != nil {
		u.mu.Lock()
		if u.f != nil {
			u.f.Close()
			u.f = nil
		}
		u.mu.Unlock()
	}
}

// adopt: 재접속한 updater가 아직 돌린다고 알린 실행 — supervisor가 재시작해 메모리에 없으면 DB 행으로 다시 연다
func (s *updRunSet) adopt(l *updLink, id int64) bool {
	if u := s.get(id); u != nil {
		return u.main == l.main && u.sub == l.sub
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	q := store.GetStorePool().Queries()
	row, err := q.GetStepRun(ctx, id)
	if err != nil || row.MainKey != l.main || row.SubKey != l.sub || (row.Result != "running" && row.Result != "wait") {
		return false
	}
	if row.Result == "wait" { // 배포 단계 — 시작 알림(StepStart)을 끊긴 동안 놓침
		_, _ = q.StartStepRun(ctx, id)
	}
	// 끊긴 사이 출력은 updater가 재접속 때 로그 전체로 다시 올린다(MsgLog -> replaceLog)
	s.open(id, l.main, l.sub, s.isInteractive(id))
	return true
}

func (s *updRunSet) setInteractive(id int64) {
	s.mu.Lock()
	s.inter[id] = true
	s.mu.Unlock()
}

func (s *updRunSet) isInteractive(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.inter[id]
}

func (s *updRunSet) output(l *updLink, o updproto.Output) {
	if u := s.get(o.RunID); u != nil && u.main == l.main && u.sub == l.sub {
		s.emit(u, o.Data)
	}
}

// emit: 로그에 쓰고 화면에도(머리 줄·끝 메시지 포함 — 누적 바이트가 화면과 어긋나지 않게)
func (s *updRunSet) emit(u *updRun, b []byte) {
	off := u.write(b)
	if s.pub != nil {
		s.pub(u.main, updproto.Output{RunID: u.id, Data: b}, off)
	}
}

// replaceLog: 재접속한 updater가 다시 올린 로그 전체로 덮어쓴다(10-1(로그 끊김)). 이미 끝난 실행이면 DB 요약도
func (s *updRunSet) replaceLog(l *updLink, g updproto.Log) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	q := store.GetStorePool().Queries()
	row, err := q.GetStepRun(ctx, g.RunID)
	if err != nil || row.MainKey != l.main || row.SubKey != l.sub {
		return
	}
	tail := g.Data
	if n := len(tail); n > runTailMax {
		tail = tail[n-runTailMax:]
	}
	u := s.get(g.RunID)
	if u != nil {
		u.mu.Lock()
		defer u.mu.Unlock()
	}
	// 순서 의존: 파일을 덮어쓰고 열린 파일을 다시 여는 동안 u.mu를 쥔다 — 그 사이 출력(write)이 옛 파일 위치에 쓰면 덮어쓴 내용이 깨짐
	if p := s.logPath(g.RunID); p != "" {
		if err := os.WriteFile(p, g.Data, 0o644); err != nil {
			log.Printf("[updater] 로그 덮어쓰기 실패 %s: %v", p, err)
		}
		if u != nil && u.f != nil {
			u.f.Close()
			u.f, _ = os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		}
	}
	if u != nil {
		u.tail = append([]byte(nil), tail...)
		u.total = int64(len(g.Data))
	}
	if row.Result != "running" && row.Result != "wait" {
		db := tail
		if len(db) > runDBTailMax {
			db = db[len(db)-runDBTailMax:]
		}
		_ = q.SetStepRunLogTail(ctx, superdb.SetStepRunLogTailParams{ID: g.RunID, LogTail: string(db)})
	}
}

// finish: 결과 기록 — 한 번만(updater가 응답을 못 받아 다시 보내도 같은 결과)
func (s *updRunSet) finish(l *updLink, e updproto.End) {
	l.running.CompareAndSwap(e.RunID, 0)
	u := s.get(e.RunID)
	var tail []byte
	if u != nil {
		if u.main != l.main || u.sub != l.sub {
			return
		}
		if e.Msg != "" {
			s.emit(u, []byte("\r\n["+e.Msg+"]\r\n"))
		}
		u.mu.Lock()
		u.end = &e
		tail = u.tail
		u.mu.Unlock()
	}
	if len(tail) > runDBTailMax {
		tail = tail[len(tail)-runDBTailMax:]
	}
	var code pgtype.Int4
	if e.ExitCode != nil {
		code = pgtype.Int4{Int32: int32(*e.ExitCode), Valid: true}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	q := store.GetStorePool().Queries()
	if row, err := q.GetStepRun(ctx, e.RunID); err != nil || row.MainKey != l.main || row.SubKey != l.sub || (row.Result != "running" && row.Result != "wait") {
		return
	}
	if err := q.FinishStepRun(ctx, superdb.FinishStepRunParams{ID: e.RunID, Result: e.Result, ExitCode: code, LogTail: string(tail)}); err != nil {
		log.Printf("[updater] 실행 %d 결과 저장 실패: %v", e.RunID, err)
	}
	// 끝난 실행은 파일만 닫고 메모리는 잠깐 둔다(바로 조회하는 화면·확인 스크립트) — 10분 뒤 비움
	if u != nil {
		u.mu.Lock()
		if u.f != nil {
			u.f.Close()
			u.f = nil
		}
		u.mu.Unlock()
		time.AfterFunc(10*time.Minute, func() { s.drop(e.RunID) })
	}
}

// ===== 합친 파일 목록(공용 uses 순서 -> main -> 그 서브키, 뒤가 이김 — updateZip.go와 같은 순서) =====

type mergedFile struct {
	bundleID         int64
	scope, sub, path string
	size             int64
	sha              string
	mode             int32
}

func mergedFiles(c echo.Context, b superdb.GetBundleRow, sub string) []mergedFile {
	ctx := c.Request().Context()
	q := TxQueries(c)
	var all []mergedFile
	add := func(id int64, files []superdb.UpdateFile) {
		for _, f := range files {
			if f.Scope == updstore.ScopeSub && f.SubKey != sub {
				continue
			}
			all = append(all, mergedFile{bundleID: id, scope: f.Scope, sub: f.SubKey, path: f.Path, size: f.Size, sha: f.Sha, mode: f.Mode})
		}
	}
	uses, err := q.ListBundleUses(ctx, b.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	for _, u := range uses {
		files, err := q.ListBundleFiles(ctx, u.ID)
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		add(u.ID, files)
	}
	files, err := q.ListBundleFiles(ctx, b.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	// ListBundleFiles = scope 순(main 먼저, sub 나중) — 서브키가 main을 이긴다
	add(b.ID, files)
	last := map[string]int{}
	for i, f := range all {
		last[f.path] = i
	}
	var out []mergedFile
	for i, f := range all {
		if last[f.path] == i {
			out = append(out, f)
		}
	}
	return out
}

func stageOf(files []mergedFile) []updproto.StageFile {
	out := make([]updproto.StageFile, len(files))
	for i, f := range files {
		v := url.Values{"b": {strconv.FormatInt(f.bundleID, 10)}, "scope": {f.scope}, "sub": {f.sub}, "path": {f.path}}
		out[i] = updproto.StageFile{Path: f.path, Size: f.size, Sha: f.sha, Mode: uint32(f.mode), URL: updproto.FilePath + "?" + v.Encode()}
	}
	return out
}

// ===== API =====

type updRunRequest struct {
	Version  string `json:"version"`
	List     string `json:"list"` // deploy | rollback
	StepID   string `json:"stepId"`
	DeployID string `json:"deployId"` // 빈 값 = 개별 실행(backup은 새 묶음, restore는 가장 최근 백업)
	Cols     uint16 `json:"cols"`
	Rows     uint16 `json:"rows"`
}

func (r *supervisorRouter) updStartRun(c echo.Context) error {
	me := r.requireAdmin(c)
	r.requireUpdStore()
	u := r.requireUpdater(c)
	var body updRunRequest
	bindJSON(c, &body)
	if body.List == "" {
		body.List = "deploy"
	}
	if body.List != "deploy" && body.List != "rollback" {
		panic(web.Err(400, "목록은 deploy · rollback 입니다"))
	}
	if body.DeployID != "" {
		updName(body.DeployID, "배포 id")
	}
	l := r.updLinks.get(u.MainKey, u.SubKey)
	if l == nil {
		panic(web.Err(409, "updater가 접속해 있지 않습니다"))
	}
	if l.busy() {
		panic(web.Err(409, "이 인스턴스에서 다른 단계·배포가 실행 중입니다"))
	}
	ref := updstore.Ref{Kind: updstore.KindBuild, Owner: u.MainKey, Version: updName(body.Version, "버전")}
	b := requireBundle(c, ref)
	meta := requireBundleFile(c, b.ID, updstore.ScopeMain, "", updstore.StepsFile)
	src, err := r.readStoreFile(ref, meta)
	if err != nil {
		panic(web.Err(500, "steps.json을 읽을 수 없습니다: %v", err))
	}
	st, issues := updstore.Check(src)
	if hasError(issues) {
		panic(web.Err(422, "steps.json에 오류가 있습니다"))
	}
	list := st.Deploy
	if body.List == "rollback" {
		list = st.Rollback
	}
	var step *updstore.Step
	for i := range list {
		if list[i].ID == body.StepID {
			step = &list[i]
		}
	}
	if step == nil {
		panic(web.Err(404, "%s 목록에 단계 %q가 없습니다", body.List, body.StepID))
	}
	files := mergedFiles(c, b, u.SubKey)
	ps := protoStep(*step)
	EndTx(c)

	// 기록은 요청 트랜잭션 밖에서 바로 커밋 — updater의 출력·끝이 커밋 전에 올 수 있다
	ctx := c.Request().Context()
	q := store.GetStorePool().Queries()
	row, err := q.CreateStepRun(ctx, superdb.CreateStepRunParams{
		BundleID: pgtype.Int8{Int64: b.ID, Valid: true}, MainKey: u.MainKey, SubKey: u.SubKey, Version: ref.Version,
		List: body.List, StepID: step.ID, StepKind: ps.Kind, StartedBy: pgtype.Int8{Int64: me, Valid: true},
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if l.deploying.Load() != 0 || !l.running.CompareAndSwap(0, row.ID) {
		_ = q.FinishStepRun(ctx, superdb.FinishStepRunParams{ID: row.ID, Result: updproto.ResultCancel, LogTail: "다른 단계가 먼저 시작됨"})
		panic(web.Err(409, "이 인스턴스에서 다른 단계가 실행 중입니다"))
	}
	run := r.updRuns.open(row.ID, u.MainKey, u.SubKey, step.Interactive)
	r.updRuns.emit(run, []byte(fmt.Sprintf("[%s %s / %s · %s]\r\n", body.List, step.ID, ref.Version, time.Now().Format("2006-01-02 15:04:05"))))
	req := updproto.Run{RunID: row.ID, DeployID: body.DeployID, Version: ref.Version, Stage: stageSha(files), Files: stageOf(files), Step: ps, Cols: body.Cols, Rows: body.Rows}
	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if _, err := l.conn.Call(cctx, updproto.MsgRun, req); err != nil {
		r.updRuns.finish(l, updproto.End{RunID: row.ID, Result: updproto.ResultFail, Msg: "updater가 받지 않음: " + err.Error()})
		panic(web.Err(409, "updater가 실행을 받지 않았습니다: %v", err))
	}
	r.publishUpd(u.MainKey, 0)
	return c.JSON(200, map[string]any{"runId": row.ID})
}

func (r *supervisorRouter) readStoreFile(ref updstore.Ref, meta superdb.UpdateFile) ([]byte, error) {
	f, err := r.upd.OpenFile(ref, meta.Scope, meta.SubKey, meta.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, updstore.TextMax))
}

type updRunItem struct {
	ID        int64  `json:"id"`
	MainKey   string `json:"mainKey"`
	SubKey    string `json:"subKey"`
	Version   string `json:"version"`
	List      string `json:"list"`
	StepID    string `json:"stepId"`
	Kind      string `json:"kind"`
	Result    string `json:"result"`
	ExitCode  *int32 `json:"exitCode"`
	StartedAt int64  `json:"startedAt"`
	EndedAt   *int64 `json:"endedAt"`
	Log       string `json:"log,omitempty"`
	LogEnd    int64  `json:"logEnd"` // 실시간이면 로그 끝의 누적 바이트(UPD:OUTPUT off와 비교), 0 = 실시간 아님
}

func runItemOf(row superdb.UpdateStepRun) updRunItem {
	it := updRunItem{ID: row.ID, MainKey: row.MainKey, SubKey: row.SubKey, Version: row.Version, List: row.List, StepID: row.StepID, Kind: row.StepKind,
		Result: row.Result, StartedAt: row.StartedAt.Time.Unix(), EndedAt: tsPtr(row.EndedAt)}
	if row.ExitCode.Valid {
		v := row.ExitCode.Int32
		it.ExitCode = &v
	}
	return it
}

func (r *supervisorRouter) requireRun(c echo.Context) superdb.UpdateStepRun {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 실행 id입니다"))
	}
	row, err := TxQueries(c).GetStepRun(c.Request().Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "실행 기록 %d가 없습니다", id))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return row
}

// updGetRun: 기록 + 로그 끝(메모리, 없으면 파일 끝 64KB, 그것도 없으면 DB 요약)
func (r *supervisorRouter) updGetRun(c echo.Context) error {
	r.requireAdmin(c)
	row := r.requireRun(c)
	it := runItemOf(row)
	if u := r.updRuns.get(row.ID); u != nil {
		u.mu.Lock()
		it.Log, it.LogEnd = string(u.tail), u.total
		u.mu.Unlock()
	} else if b, err := readTail(r.updRuns.logPath(row.ID), runTailMax); err == nil {
		it.Log = string(b)
	} else {
		it.Log = row.LogTail
	}
	return c.JSON(200, it)
}

func readTail(name string, n int64) ([]byte, error) {
	if name == "" {
		return nil, os.ErrNotExist
	}
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if st.Size() > n {
		if _, err := f.Seek(st.Size()-n, io.SeekStart); err != nil {
			return nil, err
		}
	}
	return io.ReadAll(f)
}

func (r *supervisorRouter) updListRuns(c echo.Context) error {
	r.requireAdmin(c)
	u := r.requireUpdater(c)
	rows, err := TxQueries(c).ListStepRuns(c.Request().Context(), superdb.ListStepRunsParams{MainKey: u.MainKey, SubKey: u.SubKey, Limit: 30})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := make([]updRunItem, len(rows))
	for i, row := range rows {
		out[i] = runItemOf(row)
	}
	return c.JSON(200, map[string]any{"items": out})
}

// liveRun: 실행 중이고 그 updater가 붙어 있어야 입력·취소를 보낼 수 있다
func (r *supervisorRouter) liveRun(c echo.Context) (*updLink, superdb.UpdateStepRun) {
	row := r.requireRun(c)
	if row.Result != "running" {
		panic(web.Err(409, "이미 끝난 실행입니다"))
	}
	l := r.updLinks.get(row.MainKey, row.SubKey)
	if l == nil {
		panic(web.Err(409, "updater가 접속해 있지 않습니다"))
	}
	return l, row
}

type updInputRequest struct {
	Data string `json:"data"`
}

// updRunInput: interactive 단계에만(정-6(콘솔)) — 비대화형 단계는 stdin이 닫혀 있다
func (r *supervisorRouter) updRunInput(c echo.Context) error {
	r.requireAdmin(c)
	l, row := r.liveRun(c)
	if u := r.updRuns.get(row.ID); u == nil || !u.interactive {
		panic(web.Err(409, "입력을 받는 단계(interactive)가 아닙니다"))
	}
	var body updInputRequest
	bindJSON(c, &body)
	if err := l.conn.Emit(updproto.MsgInput, updproto.Input{RunID: row.ID, Data: []byte(body.Data)}); err != nil {
		panic(web.Err(409, "보내지 못했습니다: %v", err))
	}
	return c.JSON(200, map[string]any{})
}

func (r *supervisorRouter) updRunCancel(c echo.Context) error {
	r.requireAdmin(c)
	l, row := r.liveRun(c)
	if err := l.conn.Emit(updproto.MsgCancel, updproto.Cancel{RunID: row.ID}); err != nil {
		panic(web.Err(409, "보내지 못했습니다: %v", err))
	}
	return c.JSON(200, map[string]any{})
}
