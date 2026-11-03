package router

import (
	"errors"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/updstore"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// 배포 전 시험 실행(시험 타일) 조회 — REF-update-trial.md. 실행 자체는 updStartRun 그대로

type updTrialDetail struct {
	Deploy     []updStepDef `json:"deploy"`
	Rollback   []updStepDef `json:"rollback"`
	StepsIssue string       `json:"stepsIssue"` // 빈 값 = steps.json 정상
	Runs       []updRunItem `json:"runs"`       // sub의 이 버전 시험 기록(최근 순), sub 없으면 빈 목록
}

const trialRunsMax = 50

// updGetTrial: ?sub= — 단계 정의 + 그 인스턴스의 시험 기록
func (r *supervisorRouter) updGetTrial(c echo.Context) error {
	r.requireAdmin(c)
	r.requireUpdStore()
	ref := updRef(c)
	if ref.Kind != updstore.KindBuild {
		panic(web.Err(400, "시험 실행은 빌드에서만 합니다"))
	}
	b := requireBundle(c, ref)
	ctx := c.Request().Context()
	q := TxQueries(c)
	out := updTrialDetail{Deploy: []updStepDef{}, Rollback: []updStepDef{}, Runs: []updRunItem{}}
	out.StepsIssue = r.trialSteps(c, ref, b.ID, &out)

	if sub := c.QueryParam("sub"); sub != "" {
		rows, err := q.ListTrialRuns(ctx, superdb.ListTrialRunsParams{MainKey: ref.Owner, SubKey: updName(sub, "서브키"), Version: ref.Version, Limit: trialRunsMax})
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		for _, row := range rows {
			out.Runs = append(out.Runs, runItemOf(row))
		}
	}
	return c.JSON(200, out)
}

// trialSteps: steps.json을 읽어 out에 단계를 채우고, 못 채우면 그 이유를 돌려준다
func (r *supervisorRouter) trialSteps(c echo.Context, ref updstore.Ref, bundleID int64, out *updTrialDetail) string {
	meta, err := TxQueries(c).GetBundleFile(c.Request().Context(), superdb.GetBundleFileParams{BundleID: bundleID, Scope: updstore.ScopeMain, SubKey: "", Path: updstore.StepsFile})
	if errors.Is(err, pgx.ErrNoRows) {
		return "steps.json이 없습니다"
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	src, err := r.readStoreFile(ref, meta)
	if err != nil {
		return "steps.json을 읽을 수 없습니다"
	}
	st, issues := updstore.Check(src)
	if hasError(issues) {
		return "steps.json에 오류가 있습니다"
	}
	out.Deploy, out.Rollback = stepDefs(st.Deploy), stepDefs(st.Rollback)
	return ""
}
