package router

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/supervisor/updstore"
	"nexus/internal/transport"
	"nexus/internal/updproto"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// 배포 — 단계 기록을 미리 만들고(wait) 목록을 updater에 한 번에 보낸다. updater가 차례로 실행·보고(설계-10(자기 교체) 10-2(순서 주체)).
// 되돌리기 = 같은 배포 안 rollback 목록(rollback_result). 잠그기 판정 = DB 함수 update_lockable. REF-update-p3.md 단위-4(supervisor API)

const updTopic = "UPDATE" // Updater 페이지를 연 관리자 탭(UPD:WATCH)

// ===== 실시간 알림(단위-5) =====

// publishUpd: 바뀜 신호 — 받은 쪽이 그 메인키의 장비·배포를 다시 조회. runID != 0 = 그 실행 로그 전체가 바뀜(재전송)
func (r *supervisorRouter) publishUpd(main string, runID int64) {
	if err := r.subscribeHub.Publish(updTopic, protocol.MsgUpdChange, protocol.UpdChangeEvent{MainKey: main, RunID: runID}); err != nil {
		log.Printf("[push] updater 알림 실패 %s: %v", main, err)
	}
}

func (r *supervisorRouter) publishUpdOutput(_ string, o updproto.Output, off int64) {
	_ = r.subscribeHub.Publish(updTopic, protocol.MsgUpdOutput, protocol.UpdOutputEvent{RunID: o.RunID, Data: o.Data, Off: off})
}

// onUpdWatch: 탭이 Updater 페이지를 열고 닫을 때. 관리자만 구독(권한은 그때 DB로 한 번)
func (r *supervisorRouter) onUpdWatch(conn *transport.Conn, userID int64) {
	conn.On(protocol.MsgUpdWatch, func(f protocol.Frame) {
		var w protocol.UpdWatchEvent
		if f.Bind(&w) != nil {
			return
		}
		if !w.On {
			r.subscribeHub.Unsubscribe(updTopic, conn)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		u, err := store.GetStorePool().Queries().GetUser(ctx, userID)
		if err != nil || !slices.Contains(u.Perms, permAdmin) {
			return
		}
		r.subscribeHub.Subscribe(updTopic, conn)
	})
}

// ===== 계획 =====

func protoStep(s updstore.Step) updproto.Step {
	ps := updproto.Step{ID: s.ID, Kind: updproto.KindBuiltin, Builtin: s.Builtin, Interactive: s.Interactive, TimeoutSec: int64(s.Timeout / time.Second)}
	if s.Script != "" {
		ps.Kind, ps.Script, ps.Builtin = updproto.KindScript, s.Script, ""
	}
	return ps
}

// runnable: 배포 흐름에 끼는 단계(manual 뺌)
func runnable(steps []updstore.Step) []updstore.Step {
	var out []updstore.Step
	for _, s := range steps {
		if !s.Manual {
			out = append(out, s)
		}
	}
	return out
}

// buildSteps: 빌드의 steps.json(오류가 있으면 422)
func (r *supervisorRouter) buildSteps(c echo.Context, ref updstore.Ref, b superdb.GetBundleRow) updstore.Steps {
	meta := requireBundleFile(c, b.ID, updstore.ScopeMain, "", updstore.StepsFile)
	src, err := r.readStoreFile(ref, meta)
	if err != nil {
		panic(web.Err(500, "steps.json을 읽을 수 없습니다: %v", err))
	}
	st, issues := updstore.Check(src)
	if hasError(issues) {
		panic(web.Err(422, "steps.json에 오류가 있습니다"))
	}
	return st
}

func stageSha(files []mergedFile) string {
	es := make([]updstore.Entry, len(files))
	for i, f := range files {
		es[i] = updstore.Entry{Scope: "main", Path: f.path, Sha: f.sha, Mode: f.mode}
	}
	return updstore.FolderSha(es)
}

// waitRuns: 목록 단계마다 기록(wait)을 만들고 updater에 보낼 순서를 돌려준다
func (r *supervisorRouter) waitRuns(ctx context.Context, dep superdb.UpdateDeployment, list string, steps []updstore.Step, by int64) ([]updproto.DeployStep, error) {
	q := store.GetStorePool().Queries()
	out := make([]updproto.DeployStep, 0, len(steps))
	for _, s := range steps {
		ps := protoStep(s)
		row, err := q.CreateWaitStepRun(ctx, superdb.CreateWaitStepRunParams{
			DeploymentID: pgtype.Int8{Int64: dep.ID, Valid: true}, BundleID: dep.BundleID, MainKey: dep.MainKey, SubKey: dep.SubKey,
			Version: dep.Version, List: list, StepID: s.ID, StepKind: ps.Kind, StartedBy: pgtype.Int8{Int64: by, Valid: true},
		})
		if err != nil {
			return nil, err
		}
		if s.Interactive {
			r.updRuns.setInteractive(row.ID)
		}
		out = append(out, updproto.DeployStep{RunID: row.ID, Step: ps})
	}
	return out, nil
}

// abortDeploy: updater에 보내지 못한 배포·되돌리기 — 실패로 닫고 남은 wait는 skip
func (r *supervisorRouter) abortDeploy(l *updLink, id int64, list, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	q := store.GetStorePool().Queries()
	if list == updproto.ListDeploy {
		_, _ = q.FinishDeployment(ctx, superdb.FinishDeploymentParams{ID: id, Result: updproto.ResultFail, Msg: msg})
	} else {
		_, _ = q.SetRollbackResult(ctx, superdb.SetRollbackResultParams{ID: id, RollbackResult: updproto.ResultFail, Msg: msg})
	}
	_ = q.SkipWaitRunsOf(ctx, []int64{id})
	l.deploying.CompareAndSwap(id, 0)
	r.publishUpd(l.main, 0)
}

// ===== API =====

type updDeployRequest struct {
	SubKeys []string `json:"subKeys"`
	Cols    uint16   `json:"cols"`
	Rows    uint16   `json:"rows"`
}

type updDeployResult struct {
	SubKey   string `json:"subKey"`
	DeployID int64  `json:"deployId,omitempty"`
	Error    string `json:"error,omitempty"`
}

// updDeploy: 빌드 타일 배포 팝업 — 고른 인스턴스마다 배포 하나(동시에, 서로 기다리지 않음). 오프라인·바쁨은 그 인스턴스만 실패
func (r *supervisorRouter) updDeploy(c echo.Context) error {
	me := r.requireAdmin(c)
	r.requireUpdStore()
	ref := updRef(c)
	if ref.Kind != updstore.KindBuild {
		panic(web.Err(400, "배포는 빌드만 합니다"))
	}
	var body updDeployRequest
	bindJSON(c, &body)
	if len(body.SubKeys) == 0 {
		panic(web.Err(400, "배포할 인스턴스를 고릅니다"))
	}
	b := requireBundle(c, ref)
	st := r.buildSteps(c, ref, b)
	steps := runnable(st.Deploy)
	if len(steps) == 0 {
		panic(web.Err(422, "deploy 목록에 배포 흐름 단계가 없습니다(manual만 있음)"))
	}
	type plan struct {
		sub   string
		files []mergedFile
	}
	var plans []plan
	ctx := c.Request().Context()
	for _, sub := range body.SubKeys {
		updName(sub, "서브키")
		if _, err := TxQueries(c).GetUpdater(ctx, superdb.GetUpdaterParams{MainKey: ref.Owner, SubKey: sub}); err != nil {
			panic(web.Err(404, "인스턴스 %s#%s가 없습니다", ref.Owner, sub))
		}
		plans = append(plans, plan{sub: sub, files: mergedFiles(c, b, sub)})
	}
	EndTx(c)

	out := make([]updDeployResult, 0, len(plans))
	for _, p := range plans {
		id, err := r.startDeploy(ctx, me, b, st, p.sub, p.files, body.Cols, body.Rows)
		res := updDeployResult{SubKey: p.sub, DeployID: id}
		if err != nil {
			res.Error = err.Error()
		}
		out = append(out, res)
	}
	return c.JSON(200, map[string]any{"items": out})
}

func (r *supervisorRouter) startDeploy(ctx context.Context, me int64, b superdb.GetBundleRow, st updstore.Steps, sub string, files []mergedFile, cols, rows uint16) (int64, error) {
	l := r.updLinks.get(b.Owner, sub)
	if l == nil {
		return 0, errors.New("updater가 접속해 있지 않습니다")
	}
	if l.busy() {
		return 0, errors.New("다른 단계·배포가 실행 중입니다")
	}
	q := store.GetStorePool().Queries()
	dep, err := q.CreateDeployment(ctx, superdb.CreateDeploymentParams{
		BundleID: pgtype.Int8{Int64: b.ID, Valid: true}, MainKey: b.Owner, SubKey: sub, Version: b.Version, Sha: b.Sha,
		AutoRollback: st.AutoRollback, StartedBy: pgtype.Int8{Int64: me, Valid: true}, SigBundle: b.ID,
	})
	if err != nil {
		return 0, err
	}
	// 바쁨 판정과 표시 사이에 개별 실행이 끼면 배포를 접는다
	if !l.deploying.CompareAndSwap(0, dep.ID) {
		_, _ = q.FinishDeployment(ctx, superdb.FinishDeploymentParams{ID: dep.ID, Result: updproto.ResultCancel, Msg: "다른 배포가 먼저 시작됨"})
		return dep.ID, errors.New("다른 배포가 먼저 시작됐습니다")
	}
	if l.running.Load() != 0 {
		r.abortDeploy(l, dep.ID, updproto.ListDeploy, "다른 단계가 먼저 시작됨")
		return dep.ID, errors.New("다른 단계가 먼저 시작됐습니다")
	}
	req := updproto.Deploy{DeployID: dep.ID, List: updproto.ListDeploy, Version: b.Version, Stage: stageSha(files), Files: stageOf(files), Cols: cols, Rows: rows}
	if req.Steps, err = r.waitRuns(ctx, dep, updproto.ListDeploy, runnable(st.Deploy), me); err == nil && st.AutoRollback {
		req.Rollback, err = r.waitRuns(ctx, dep, updproto.ListRollback, runnable(st.Rollback), me)
	}
	if err != nil {
		r.abortDeploy(l, dep.ID, updproto.ListDeploy, "단계 기록을 만들지 못함: "+err.Error())
		return dep.ID, err
	}
	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if _, err := l.conn.Call(cctx, updproto.MsgDeploy, req); err != nil {
		r.abortDeploy(l, dep.ID, updproto.ListDeploy, "updater가 받지 않음: "+err.Error())
		return dep.ID, fmt.Errorf("updater가 배포를 받지 않았습니다: %w", err)
	}
	r.publishUpd(b.Owner, 0)
	return dep.ID, nil
}

func requireDeployment(c echo.Context) superdb.GetDeploymentRow {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 배포 id입니다"))
	}
	d, err := TxQueries(c).GetDeployment(c.Request().Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "배포 %d가 없습니다", id))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return d
}

// depBuild: 배포한 빌드가 아직 있는가(같은 이름으로 다시 만든 빌드는 다른 것)
func depBuild(c echo.Context, d superdb.GetDeploymentRow) (superdb.GetBundleRow, bool) {
	if !d.BundleID.Valid {
		return superdb.GetBundleRow{}, false
	}
	b, err := TxQueries(c).GetBundle(c.Request().Context(), superdb.GetBundleParams{Kind: updstore.KindBuild, Owner: d.MainKey, Version: d.Version})
	if err != nil || b.ID != d.BundleID.Int64 {
		return superdb.GetBundleRow{}, false
	}
	return b, true
}

// updRollback: Run·Log 타일 되돌리기 = 같은 배포의 rollback 목록. 배포가 끝난 뒤(성공이어도)만
func (r *supervisorRouter) updRollback(c echo.Context) error {
	me := r.requireAdmin(c)
	r.requireUpdStore()
	d := requireDeployment(c)
	if d.Result == "running" || d.RollbackResult == "running" {
		panic(web.Err(409, "배포·되돌리기가 아직 끝나지 않았습니다"))
	}
	b, ok := depBuild(c, d)
	if !ok {
		panic(web.Err(409, "빌드가 지워져 되돌리기 단계(스크립트)를 받을 수 없습니다"))
	}
	ref := updstore.Ref{Kind: updstore.KindBuild, Owner: d.MainKey, Version: d.Version}
	st := r.buildSteps(c, ref, b)
	steps := runnable(st.Rollback)
	if len(steps) == 0 {
		panic(web.Err(422, "rollback 목록에 단계가 없습니다"))
	}
	files := mergedFiles(c, b, d.SubKey)
	EndTx(c)

	l := r.updLinks.get(d.MainKey, d.SubKey)
	if l == nil {
		panic(web.Err(409, "updater가 접속해 있지 않습니다"))
	}
	if l.busy() || !l.deploying.CompareAndSwap(0, d.ID) {
		panic(web.Err(409, "이 인스턴스에서 다른 단계·배포가 실행 중입니다"))
	}
	ctx := c.Request().Context()
	q := store.GetStorePool().Queries()
	_ = q.DeleteWaitRuns(ctx, superdb.DeleteWaitRunsParams{DeploymentID: pgtype.Int8{Int64: d.ID, Valid: true}, List: updproto.ListRollback})
	if _, err := q.SetRollbackResult(ctx, superdb.SetRollbackResultParams{ID: d.ID, RollbackResult: "running"}); err != nil {
		l.deploying.CompareAndSwap(d.ID, 0)
		panic(web.Err(500, "%v", err))
	}
	dep := superdb.UpdateDeployment{ID: d.ID, BundleID: d.BundleID, MainKey: d.MainKey, SubKey: d.SubKey, Version: d.Version}
	run, err := r.waitRuns(ctx, dep, updproto.ListRollback, steps, me)
	if err != nil {
		r.abortDeploy(l, d.ID, updproto.ListRollback, "단계 기록을 만들지 못함: "+err.Error())
		panic(web.Err(500, "%v", err))
	}
	var body updDeployRequest
	_ = c.Bind(&body)
	req := updproto.Deploy{DeployID: d.ID, List: updproto.ListRollback, Version: d.Version, Stage: stageSha(files), Files: stageOf(files), Steps: run, Cols: body.Cols, Rows: body.Rows}
	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if _, err := l.conn.Call(cctx, updproto.MsgDeploy, req); err != nil {
		r.abortDeploy(l, d.ID, updproto.ListRollback, "updater가 받지 않음: "+err.Error())
		panic(web.Err(409, "updater가 되돌리기를 받지 않았습니다: %v", err))
	}
	r.publishUpd(d.MainKey, 0)
	return c.JSON(200, map[string]any{"deployId": d.ID})
}

// updDeployCancel: 지금 단계를 멈추고 다음 단계로 넘어가지 않게
func (r *supervisorRouter) updDeployCancel(c echo.Context) error {
	r.requireAdmin(c)
	d := requireDeployment(c)
	if d.Result != "running" && d.RollbackResult != "running" {
		panic(web.Err(409, "이미 끝난 배포입니다"))
	}
	l := r.updLinks.get(d.MainKey, d.SubKey)
	if l == nil || l.deploying.Load() != d.ID {
		panic(web.Err(409, "updater가 접속해 있지 않거나 이 배포를 돌리고 있지 않습니다"))
	}
	if err := l.conn.Emit(updproto.MsgCancel, updproto.Cancel{RunID: l.running.Load(), DeployID: d.ID}); err != nil {
		panic(web.Err(409, "보내지 못했습니다: %v", err))
	}
	return c.JSON(200, map[string]any{})
}

type updStepDef struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Name        string `json:"name"` // script 파일 또는 builtin 이름
	Manual      bool   `json:"manual"`
	Interactive bool   `json:"interactive"`
	TimeoutSec  int64  `json:"timeoutSec"`
}

type updDeployDetail struct {
	ID             int64        `json:"id"`
	MainKey        string       `json:"mainKey"`
	SubKey         string       `json:"subKey"`
	Version        string       `json:"version"`
	Sha            string       `json:"sha"`
	Result         string       `json:"result"`
	RollbackResult string       `json:"rollbackResult"`
	AutoRollback   bool         `json:"autoRollback"`
	Msg            string       `json:"msg"`
	StartedByName  string       `json:"startedByName"`
	StartedAt      int64        `json:"startedAt"`
	EndedAt        *int64       `json:"endedAt"`
	Build          bool         `json:"build"`       // 빌드가 아직 있음(없으면 되돌리기·단계 실행 불가)
	BuildSha       string       `json:"buildSha"`    // 지금 빌드 sha(배포 뒤 고쳤으면 다름)
	BuildLocked    bool         `json:"buildLocked"` // 이미 잠김
	Lockable       bool         `json:"lockable"`    // 잠그기 판정(DB 함수)
	Online         bool         `json:"online"`
	Busy           bool         `json:"busy"`
	Deploy         []updStepDef `json:"deploy"`   // 지금 steps.json 목록(빌드가 있고 오류 없을 때)
	Rollback       []updStepDef `json:"rollback"` // 같은 방식
	Runs           []updRunItem `json:"runs"`     // 이 배포의 단계 기록(목록·순서대로)
	Manual         []updRunItem `json:"manual"`   // 이 배포 뒤 같은 인스턴스의 개별 실행(최근 순)
}

func stepDefs(steps []updstore.Step) []updStepDef {
	out := make([]updStepDef, len(steps))
	for i, s := range steps {
		ps := protoStep(s)
		out[i] = updStepDef{ID: s.ID, Kind: ps.Kind, Name: ps.Script + ps.Builtin, Manual: s.Manual, Interactive: s.Interactive, TimeoutSec: ps.TimeoutSec}
	}
	return out
}

// updGetDeployment: Run·Log 타일 — 배포 하나 × 인스턴스 하나(REF-update-p3.md "합의한 해석")
func (r *supervisorRouter) updGetDeployment(c echo.Context) error {
	r.requireAdmin(c)
	d := requireDeployment(c)
	ctx := c.Request().Context()
	q := TxQueries(c)
	out := updDeployDetail{ID: d.ID, MainKey: d.MainKey, SubKey: d.SubKey, Version: d.Version, Sha: d.Sha, Result: d.Result, RollbackResult: d.RollbackResult,
		AutoRollback: d.AutoRollback, Msg: d.Msg, StartedByName: d.StartedByName, StartedAt: d.StartedAt.Time.Unix(), EndedAt: tsPtr(d.EndedAt),
		Deploy: []updStepDef{}, Rollback: []updStepDef{}, Runs: []updRunItem{}, Manual: []updRunItem{}}
	if b, ok := depBuild(c, d); ok {
		out.Build, out.BuildSha, out.BuildLocked, out.Lockable = true, b.Sha, b.Locked, b.Lockable
		if r.upd.Enabled() {
			if meta, err := q.GetBundleFile(ctx, superdb.GetBundleFileParams{BundleID: b.ID, Scope: updstore.ScopeMain, SubKey: "", Path: updstore.StepsFile}); err == nil {
				if src, err := r.readStoreFile(updstore.Ref{Kind: updstore.KindBuild, Owner: d.MainKey, Version: d.Version}, meta); err == nil {
					if st, issues := updstore.Check(src); !hasError(issues) {
						out.Deploy, out.Rollback = stepDefs(st.Deploy), stepDefs(st.Rollback)
					}
				}
			}
		}
	}
	if l := r.updLinks.get(d.MainKey, d.SubKey); l != nil {
		out.Online, out.Busy = true, l.busy()
	}
	runs, err := q.ListDeploymentRuns(ctx, pgtype.Int8{Int64: d.ID, Valid: true})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	for _, row := range runs {
		out.Runs = append(out.Runs, runItemOf(row))
	}
	recent, err := q.ListStepRuns(ctx, superdb.ListStepRunsParams{MainKey: d.MainKey, SubKey: d.SubKey, Limit: 50})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	for _, row := range recent {
		if !row.DeploymentID.Valid && row.Version == d.Version && !row.StartedAt.Time.Before(d.StartedAt.Time) {
			out.Manual = append(out.Manual, runItemOf(row))
		}
	}
	return c.JSON(200, out)
}

// updLockBuild: Run·Log 타일 잠그기 — 판정은 DB 함수 한 곳(REF-update-rules.md "잠금"). 빌드 + 쓰는 공용 묶음 같이
func (r *supervisorRouter) updLockBuild(c echo.Context) error {
	me := r.requireAdmin(c)
	ref := updRef(c)
	if ref.Kind != updstore.KindBuild {
		panic(web.Err(400, "잠그기는 빌드만 합니다(공용 묶음은 쓰는 빌드와 같이 잠김)"))
	}
	ctx := c.Request().Context()
	q := TxQueries(c)
	b, err := q.GetBundleForUpdate(ctx, superdb.GetBundleForUpdateParams{Kind: ref.Kind, Owner: ref.Owner, Version: ref.Version})
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "%s가 없습니다", ref))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if b.Locked {
		panic(web.Err(409, "%s는 이미 잠겨 있습니다", ref))
	}
	ok, err := q.IsLockable(ctx, b.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if !ok {
		panic(web.Err(409, "지금 내용(sha)으로 끝까지 성공한 배포가 없어 잠글 수 없습니다"))
	}
	if err := q.LockBuildWithUses(ctx, superdb.LockBuildWithUsesParams{BuildID: b.ID, LockedBy: me}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	AfterCommit(c, func() { r.publishUpd(ref.Owner, 0) })
	return c.JSON(200, map[string]any{})
}

// ===== updater -> supervisor =====

// updStepStart: 배포 단계 시작 — 기록 running · 로그 열기. 끊겨서 못 받으면 HELLO Running의 adopt가 대신
func (r *supervisorRouter) updStepStart(l *updLink, s updproto.StepStart) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	q := store.GetStorePool().Queries()
	row, err := q.GetStepRun(ctx, s.RunID)
	if err != nil || row.MainKey != l.main || row.SubKey != l.sub || !row.DeploymentID.Valid || row.DeploymentID.Int64 != s.DeployID {
		return
	}
	if n, err := q.StartStepRun(ctx, s.RunID); err != nil || n == 0 {
		return
	}
	if row.List == updproto.ListRollback { // 자동 되돌리기가 시작됨
		_, _ = q.SetRollbackResult(ctx, superdb.SetRollbackResultParams{ID: s.DeployID, RollbackResult: "running"})
	}
	u := r.updRuns.open(s.RunID, l.main, l.sub, r.updRuns.isInteractive(s.RunID))
	r.updRuns.emit(u, []byte(fmt.Sprintf("[%s %s / %s · 배포 %d · %s]\r\n", row.List, row.StepID, row.Version, s.DeployID, time.Now().Format("2006-01-02 15:04:05"))))
	l.running.Store(s.RunID)
	r.publishUpd(l.main, 0)
}

// updDeployEnd: 배포·되돌리기 끝 — 결과 · 남은 wait 정리 · 설치본 이름. updater가 다시 보내도 같은 결과
func (r *supervisorRouter) updDeployEnd(l *updLink, e updproto.DeployEnd) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	q := store.GetStorePool().Queries()
	d, err := q.GetDeployment(ctx, e.DeployID)
	if err != nil || d.MainKey != l.main || d.SubKey != l.sub {
		l.deploying.CompareAndSwap(e.DeployID, 0)
		return
	}
	id := pgtype.Int8{Int64: d.ID, Valid: true}
	if e.List == updproto.ListDeploy {
		_, _ = q.FinishDeployment(ctx, superdb.FinishDeploymentParams{ID: d.ID, Result: e.Result, Msg: e.Msg})
		_ = q.SkipWaitRuns(ctx, superdb.SkipWaitRunsParams{DeploymentID: id, List: updproto.ListDeploy})
		if e.Rollback != "" {
			_, _ = q.SetRollbackResult(ctx, superdb.SetRollbackResultParams{ID: d.ID, RollbackResult: e.Rollback})
			_ = q.SkipWaitRuns(ctx, superdb.SkipWaitRunsParams{DeploymentID: id, List: updproto.ListRollback})
		} else {
			_ = q.DeleteWaitRuns(ctx, superdb.DeleteWaitRunsParams{DeploymentID: id, List: updproto.ListRollback}) // 자동 되돌리기를 안 함
		}
	} else {
		_, _ = q.SetRollbackResult(ctx, superdb.SetRollbackResultParams{ID: d.ID, RollbackResult: e.Result, Msg: e.Msg})
		_ = q.SkipWaitRuns(ctx, superdb.SkipWaitRunsParams{DeploymentID: id, List: updproto.ListRollback})
	}
	_ = q.UpdateUpdaterInstalled(ctx, superdb.UpdateUpdaterInstalledParams{MainKey: l.main, SubKey: l.sub, InstalledVersion: e.Installed})
	l.deploying.CompareAndSwap(e.DeployID, 0)
	r.publishUpd(l.main, 0)
}
