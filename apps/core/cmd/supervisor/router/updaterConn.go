package router

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/supervisor/updstore"
	"nexus/internal/transport"
	"nexus/internal/updproto"
	"nexus/internal/util"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// updater 접속 — 메모리 연결 목록(online은 여기서 파생, DB updaters = 관측용 명부) · HELLO · 인스턴스 고치기 · 파일 받기. REF-update-p2.md 단위-3(접속)

type updLink struct {
	conn      *transport.Conn
	main, sub string
	version   string
	running   atomic.Int64 // 지금 실행 중인 단계 기록 id(0 = 없음) — 인스턴스당 하나
	deploying atomic.Int64 // 지금 실행 중인 배포 id(0 = 없음) — 배포 중이면 그 단계가 running
}

// busy: 단계든 배포든 하나라도 돌고 있음
func (l *updLink) busy() bool { return l.running.Load() != 0 || l.deploying.Load() != 0 }

func (l *updLink) key() string { return l.main + "#" + l.sub }

type updLinks struct {
	mu sync.Mutex
	m  map[string]*updLink
}

func newUpdLinks() *updLinks { return &updLinks{m: map[string]*updLink{}} }

func (s *updLinks) get(main, sub string) *updLink {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.m[main+"#"+sub]
}

// add: 같은 서브키가 이미 붙어 있으면 거절(updater 두 개가 같은 state.json을 쓰는 경우)
func (s *updLinks) add(l *updLink) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.m[l.key()]; ok {
		return false
	}
	s.m[l.key()] = l
	return true
}

// remove: 그 연결이 아직 주인일 때만(다시 붙은 새 연결을 지우지 않게)
func (s *updLinks) remove(l *updLink) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.m[l.key()] == l {
		delete(s.m, l.key())
	}
}

func (s *updLinks) byMain(main string) []*updLink {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*updLink
	for _, l := range s.m {
		if l.main == main {
			out = append(out, l)
		}
	}
	return out
}

func (r *supervisorRouter) updaterItemOf(u superdb.Updater) updaterItem {
	it := updaterItem{SubKey: u.SubKey, Label: u.Label, Memo: u.Memo, OS: u.Os, Arch: u.Arch, InstallDir: u.InstallDir, InstalledVersion: u.InstalledVersion, LastSeenAt: tsPtr(u.LastSeenAt)}
	if l := r.updLinks.get(u.MainKey, u.SubKey); l != nil {
		it.Online = true
		it.Version = l.version
		if id := l.running.Load(); id != 0 {
			it.Running = &id
		}
		if id := l.deploying.Load(); id != 0 {
			it.Deploying = &id
		}
	}
	return it
}

// ===== WS =====

// handleUpdaterWS: worker WS(handleWorkerWS)와 같은 틀 — 업그레이드 뒤 hijack이라 항상 nil 반환
func (r *supervisorRouter) handleUpdaterWS(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("[updater] upgrade 실패 %v", err)
		return nil
	}
	transport.KeepAlive(ws)
	conn := transport.New(ws)

	// HELLO는 REQ 고루틴, 출력은 EVENT 디스패처 — 서로 다른 고루틴이라 atomic
	// 순서 의존: 단계 끝·배포 끝·로그 재전송은 HandleOrdered — 앞선 시작 알림·출력(EVENT)보다 먼저 처리되면
	// 실행 중 표시가 끝난 뒤에 다시 켜지고(인스턴스가 바쁨으로 남음) 마지막 출력이 로그에서 빠진다(P3 화면 확인에서 발견)
	var link atomic.Pointer[updLink]
	conn.Handle(updproto.MsgHello, func(f protocol.Frame) (any, error) {
		if link.Load() != nil {
			return nil, errors.New("이미 HELLO를 받았습니다")
		}
		var h updproto.Hello
		if err := f.Bind(&h); err != nil {
			return nil, err
		}
		l, err := r.updHello(conn, h)
		if err != nil {
			log.Printf("[updater] 거부 %q: %v", h.MainKey, err)
			return nil, err
		}
		link.Store(l)
		return updproto.HelloResponse{SubKey: l.sub}, nil
	})
	conn.On(updproto.MsgOutput, func(f protocol.Frame) {
		if l := link.Load(); l != nil {
			var o updproto.Output
			if f.Bind(&o) == nil {
				r.updRuns.output(l, o)
			}
		}
	})
	conn.HandleOrdered(updproto.MsgEnd, func(f protocol.Frame) (any, error) {
		l := link.Load()
		if l == nil {
			return nil, errors.New("HELLO 전입니다")
		}
		var e updproto.End
		if err := f.Bind(&e); err != nil {
			return nil, err
		}
		r.updRuns.finish(l, e)
		r.publishUpd(l.main, 0)
		return nil, nil
	})
	// P3(배포) — updaterDeploy.go
	conn.On(updproto.MsgStepStart, func(f protocol.Frame) {
		if l := link.Load(); l != nil {
			var s updproto.StepStart
			if f.Bind(&s) == nil {
				r.updStepStart(l, s)
			}
		}
	})
	conn.HandleOrdered(updproto.MsgDeployEnd, func(f protocol.Frame) (any, error) {
		l := link.Load()
		if l == nil {
			return nil, errors.New("HELLO 전입니다")
		}
		var e updproto.DeployEnd
		if err := f.Bind(&e); err != nil {
			return nil, err
		}
		r.updDeployEnd(l, e)
		return nil, nil
	})
	conn.HandleOrdered(updproto.MsgLog, func(f protocol.Frame) (any, error) {
		l := link.Load()
		if l == nil {
			return nil, errors.New("HELLO 전입니다")
		}
		var g updproto.Log
		if err := f.Bind(&g); err != nil {
			return nil, err
		}
		r.updRuns.replaceLog(l, g)
		r.publishUpd(l.main, g.RunID)
		return nil, nil
	})

	err = conn.Serve()
	if l := link.Load(); l != nil {
		r.updLinks.remove(l)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = store.GetStorePool().Queries().TouchUpdater(ctx, superdb.TouchUpdaterParams{MainKey: l.main, SubKey: l.sub})
		cancel()
		log.Printf("[updater] 끊김 %s (%v)", l.key(), err)
		r.publishUpd(l.main, 0)
	}
	conn.Close(err)
	return nil
}

func (r *supervisorRouter) updHello(conn *transport.Conn, h updproto.Hello) (*updLink, error) {
	if err := checkUpdaterKey(h.MainKey, h.APIKey); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := store.GetStorePool().Queries()
	sub := h.SubKey
	if sub == "" {
		for {
			sub, _ = util.RandomKey(8, "", "")
			if _, err := q.GetUpdater(ctx, superdb.GetUpdaterParams{MainKey: h.MainKey, SubKey: sub}); errors.Is(err, pgx.ErrNoRows) {
				break
			}
		}
	} else if !updstore.NamePattern.MatchString(sub) {
		return nil, fmt.Errorf("서브키 %q 형식이 아닙니다", sub)
	}
	l := &updLink{conn: conn, main: h.MainKey, sub: sub, version: h.Version}
	if !r.updLinks.add(l) {
		return nil, fmt.Errorf("서브키 %s#%s는 이미 접속 중입니다(state.json을 복사한 updater가 둘?)", h.MainKey, sub)
	}
	// 지운 인스턴스가 다시 붙으면 행이 새로 생긴다(별칭·메모는 비어 있음)
	if err := q.UpsertUpdater(ctx, superdb.UpsertUpdaterParams{MainKey: h.MainKey, SubKey: sub, Os: h.OS, Arch: h.Arch, InstallDir: h.TargetDir, InstalledVersion: h.Installed}); err != nil {
		r.updLinks.remove(l)
		return nil, err
	}
	if lost, err := q.CancelLostStepRuns(ctx, superdb.CancelLostStepRunsParams{MainKey: h.MainKey, SubKey: sub, Keep: append([]int64{}, h.Running...)}); err == nil {
		for _, id := range lost {
			r.updRuns.drop(id)
		}
	}
	for _, id := range h.Running {
		if r.updRuns.adopt(l, id) {
			l.running.Store(id)
		}
	}
	// 배포: updater가 모르는 배포(재시작 = 중단, REF-update-p3.md "합의한 해석") · 아는 배포는 DeployEnd까지 바쁨
	if lost, err := q.CancelLostDeployments(ctx, superdb.CancelLostDeploymentsParams{MainKey: h.MainKey, SubKey: sub, Keep: append([]int64{}, h.Deploying...)}); err == nil && len(lost) > 0 {
		_ = q.SkipWaitRunsOf(ctx, lost)
	}
	if len(h.Deploying) > 0 {
		l.deploying.Store(h.Deploying[0])
	}
	defer r.publishUpd(h.MainKey, 0)
	log.Printf("[updater] 접속 %s %s/%s %s 경로=%q", l.key(), h.OS, h.Arch, h.Version, h.TargetDir)
	return l, nil
}

// ===== 인스턴스 =====

type updInstancePatch struct {
	Label     *string `json:"label"`
	Memo      *string `json:"memo"`
	TargetDir *string `json:"targetDir"` // 정-5(설치 경로): updater가 설정 파일을 고쳐 씀 — 접속 중이고 실행 중이 아닐 때만
}

func (r *supervisorRouter) requireUpdater(c echo.Context) superdb.Updater {
	main, sub := updName(c.Param("owner"), "메인키"), updName(c.Param("sub"), "서브키")
	u, err := TxQueries(c).GetUpdater(c.Request().Context(), superdb.GetUpdaterParams{MainKey: main, SubKey: sub})
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "인스턴스 %s#%s가 없습니다", main, sub))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return u
}

func (r *supervisorRouter) updPatchInstance(c echo.Context) error {
	r.requireAdmin(c)
	u := r.requireUpdater(c)
	var body updInstancePatch
	bindJSON(c, &body)
	ctx := c.Request().Context()
	q := TxQueries(c)
	if body.Label != nil || body.Memo != nil {
		label, memo := u.Label, u.Memo
		if body.Label != nil {
			label = *body.Label
		}
		if body.Memo != nil {
			memo = *body.Memo
		}
		if _, err := q.UpdateUpdaterLabel(ctx, superdb.UpdateUpdaterLabelParams{MainKey: u.MainKey, SubKey: u.SubKey, Label: label, Memo: memo}); err != nil {
			panic(web.Err(500, "%v", err))
		}
	}
	if body.TargetDir != nil && *body.TargetDir != u.InstallDir {
		l := r.updLinks.get(u.MainKey, u.SubKey)
		if l == nil {
			panic(web.Err(409, "updater가 접속해 있지 않아 설치 경로를 바꿀 수 없습니다(설정 파일은 장비에 있음)"))
		}
		if l.busy() {
			panic(web.Err(409, "단계·배포가 실행 중이라 설치 경로를 바꿀 수 없습니다"))
		}
		cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if _, err := l.conn.Call(cctx, updproto.MsgSetDir, updproto.SetDir{TargetDir: *body.TargetDir}); err != nil {
			panic(web.Err(409, "updater가 거절했습니다: %v", err))
		}
		if _, err := q.UpdateUpdaterInstallDir(ctx, superdb.UpdateUpdaterInstallDirParams{MainKey: u.MainKey, SubKey: u.SubKey, InstallDir: *body.TargetDir}); err != nil {
			panic(web.Err(500, "%v", err))
		}
	}
	return c.JSON(200, map[string]any{})
}

// updDeleteInstance: 명부 행만 — 접속 중이면 거절(다시 붙으면 새 행이 생김)
func (r *supervisorRouter) updDeleteInstance(c echo.Context) error {
	r.requireAdmin(c)
	u := r.requireUpdater(c)
	if r.updLinks.get(u.MainKey, u.SubKey) != nil {
		panic(web.Err(409, "접속 중인 인스턴스는 지울 수 없습니다"))
	}
	if _, err := TxQueries(c).DeleteUpdater(c.Request().Context(), superdb.DeleteUpdaterParams{MainKey: u.MainKey, SubKey: u.SubKey}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	return c.JSON(200, map[string]any{})
}

// dropUpdaters: 장비 루트를 지우면 붙어 있던 updater 연결을 끊는다(다음 HELLO는 키 검사에서 거절)
func (r *supervisorRouter) dropUpdaters(main string) {
	for _, l := range r.updLinks.byMain(main) {
		l.conn.Close(errors.New("장비 루트가 지워졌습니다"))
	}
}

// ===== 파일 받기(updater 전용, HTTP) =====

// handleUpdaterFile: ?b=<번들 id>&scope=&sub=&path= — 그 메인키의 빌드 또는 공용 묶음만
func (r *supervisorRouter) handleUpdaterFile(c echo.Context) error {
	main := c.Request().Header.Get(updproto.HeaderMain)
	if err := checkUpdaterKey(main, c.Request().Header.Get(updproto.HeaderKey)); err != nil {
		panic(web.Err(401, "%v", err))
	}
	r.requireUpdStore()
	id, err := strconv.ParseInt(c.QueryParam("b"), 10, 64)
	if err != nil {
		panic(web.Err(400, "번들 id가 없습니다"))
	}
	ctx := c.Request().Context()
	q := TxQueries(c)
	b, err := q.GetBundleByID(ctx, id)
	if err != nil {
		panic(web.Err(404, "번들이 없습니다"))
	}
	if b.Kind == updstore.KindBuild && b.Owner != main {
		panic(web.Err(403, "다른 장비의 빌드입니다"))
	}
	ref := updstore.Ref{Kind: b.Kind, Owner: b.Owner, Version: b.Version}
	scope, sub, rel := updFileTarget(ref, c.QueryParam("scope"), c.QueryParam("sub"), c.QueryParam("path"), false)
	meta := requireBundleFile(c, b.ID, scope, sub, rel)
	EndTx(c)
	f, err := r.upd.OpenFile(ref, scope, sub, rel)
	if err != nil {
		panic(web.Err(500, "보관소에 파일이 없습니다: %v", err))
	}
	defer f.Close()
	h := c.Response().Header()
	h.Set("Cache-Control", "no-store")
	h.Set("Content-Length", strconv.FormatInt(meta.Size, 10))
	h.Set("X-Content-Sha", meta.Sha)
	return c.Stream(http.StatusOK, "application/octet-stream", f)
}
