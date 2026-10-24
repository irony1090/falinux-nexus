package router

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"strconv"
	"sync"

	"nexus/internal/protocol"
	"nexus/internal/transport"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// ===== supervisor 주소 옮기기 + 미등록 process — 주소-2(옮기기 명령)·주소-4(미등록 process), REF-supervisor-move.md =====
// 장비 하나 = 그 장비를 등록한 계정(requireDeviceDir) / 전체 = worker_move 권한(requirePerm)

// unregisteredSet: 인스턴스별 미등록 process(이 supervisor DB에 행이 없는데 worker가 살아 있다고 보고한 것). 메모리만 —
// sync(재접속 보고)로 통째로 바뀌고, 끝남(status)·끊김에 빠진다
type unregisteredSet struct {
	mu    sync.Mutex
	byKey map[string]map[string]protocol.SyncEntry
}

func newUnregisteredSet() *unregisteredSet {
	return &unregisteredSet{byKey: map[string]map[string]protocol.SyncEntry{}}
}

func (s *unregisteredSet) set(key string, list []protocol.SyncEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(list) == 0 {
		delete(s.byKey, key)
		return
	}
	m := make(map[string]protocol.SyncEntry, len(list))
	for _, e := range list {
		m[e.UID] = e
	}
	s.byKey[key] = m
}

func (s *unregisteredSet) drop(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byKey, key)
}

func (s *unregisteredSet) removeUID(uid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, m := range s.byKey {
		if _, ok := m[uid]; ok {
			delete(m, uid)
			if len(m) == 0 {
				delete(s.byKey, key)
			}
			return
		}
	}
}

func (s *unregisteredSet) has(key, uid string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.byKey[key][uid]
	return ok
}

// list: 시작이 오래된 것부터
func (s *unregisteredSet) list(key string) []protocol.SyncEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]protocol.SyncEntry, 0, len(s.byKey[key]))
	for _, e := range s.byKey[key] {
		out = append(out, e)
	}
	slices.SortFunc(out, func(a, b protocol.SyncEntry) int {
		return cmp.Or(cmp.Compare(a.StartedAt, b.StartedAt), cmp.Compare(a.UID, b.UID))
	})
	return out
}

type moveRequest struct {
	FolderID int64  `json:"folderId"` // 장비 하나 옮기기만(장비 타일이 연 폴더)
	URL      string `json:"url" validate:"required"`
	Now      bool   `json:"now"` // 주소-Q9(즉시 옮김 여부)
}

type addrsResponse struct {
	Addrs []string `json:"addrs"`
}

// callWorker: worker REQ 한 번 — 응답 없음 504 / 실패 502
func callWorker(ctx context.Context, conn *transport.Conn, t protocol.MsgType, data any, out any) {
	ctx, cancel := context.WithTimeout(ctx, sendCallTimeout)
	defer cancel()
	res, err := conn.Call(ctx, t, data)
	if errors.Is(err, context.DeadlineExceeded) {
		panic(web.Err(504, "worker 응답이 없습니다"))
	}
	if err != nil {
		panic(web.Err(502, "%v", err))
	}
	if out != nil {
		if err := res.Bind(out); err != nil {
			panic(web.Err(502, "%v", err))
		}
	}
}

// listWorkerAddrs: GET /workers/:mainKey/:subKey/addrs?folderId= — worker에 저장된 주소 목록(1순위부터)
func (r *supervisorRouter) listWorkerAddrs(c echo.Context) error {
	sess := r.requireSession(c)
	folderID, err := strconv.ParseInt(c.QueryParam("folderId"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 folderId입니다"))
	}
	conn := r.requireDeviceDir(c, sess.Data.ID, folderID, pathParam(c, "mainKey"), pathParam(c, "subKey"))
	EndTx(c)
	var out protocol.AddrsResponse
	callWorker(c.Request().Context(), conn, protocol.MsgAddrs, struct{}{}, &out)
	return c.JSON(200, addrsResponse{Addrs: nonNil(out.Addrs)})
}

// moveWorker: POST /workers/:mainKey/:subKey/move — 장비 하나 옮기기
func (r *supervisorRouter) moveWorker(c echo.Context) error {
	sess := r.requireSession(c)
	body := bindMove(c)
	conn := r.requireDeviceDir(c, sess.Data.ID, body.FolderID, pathParam(c, "mainKey"), pathParam(c, "subKey"))
	EndTx(c)
	var out protocol.AddrsResponse
	callWorker(c.Request().Context(), conn, protocol.MsgMove, protocol.MoveRequest{URL: body.URL, Now: body.Now}, &out)
	return c.JSON(200, addrsResponse{Addrs: nonNil(out.Addrs)})
}

type moveAllResult struct {
	InstanceKey string   `json:"instanceKey"`
	Addrs       []string `json:"addrs"`
	Error       string   `json:"error,omitempty"`
}

// moveAllWorkers: POST /workers/move-all — 접속 중인 worker 전부(주소-Q7(대상 범위)), worker_move 권한만(주소-Q20(전체 명령 권한))
func (r *supervisorRouter) moveAllWorkers(c echo.Context) error {
	sess := r.requireSession(c)
	body := bindMove(c)
	requirePerm(c, sess.Data.ID, permWorkerMove)
	EndTx(c)

	conns := r.workers.FindAll(func(string, *transport.Conn) bool { return true })
	out := make([]moveAllResult, len(conns))
	var wg sync.WaitGroup
	for i, w := range conns {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := moveAllResult{InstanceKey: w.Key, Addrs: []string{}}
			ctx, cancel := context.WithTimeout(c.Request().Context(), sendCallTimeout)
			defer cancel()
			f, err := w.Val.Call(ctx, protocol.MsgMove, protocol.MoveRequest{URL: body.URL, Now: body.Now})
			var a protocol.AddrsResponse
			if err == nil {
				err = f.Bind(&a)
			}
			if err != nil {
				res.Error = err.Error()
			} else {
				res.Addrs = nonNil(a.Addrs)
			}
			out[i] = res
		}()
	}
	wg.Wait()
	slices.SortFunc(out, func(a, b moveAllResult) int { return cmp.Compare(a.InstanceKey, b.InstanceKey) })
	return c.JSON(200, out)
}

// bindMove: 주소 검사는 worker와 같은 함수 — 잘못된 주소가 worker 목록에 들어가지 않게 먼저 거른다
func bindMove(c echo.Context) moveRequest {
	var body moveRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	addr, err := protocol.NormalizeSupervisorAddr(body.URL)
	if err != nil {
		panic(web.Err(400, "%v", err))
	}
	body.URL = addr
	return body
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

type unregisteredResponse struct {
	UID       string `json:"uid"`
	Kind      string `json:"kind"`
	Label     string `json:"label"` // 스크립트 노드 이름(장비 파일 실행·이 기능 전 process는 빈 값)
	Cmd       string `json:"cmd"`
	Cwd       string `json:"cwd"`
	PID       int    `json:"pid"`
	StartedAt int64  `json:"startedAt"` // Unix 초, 0 = 모름
}

// listUnregistered: GET /workers/:mainKey/:subKey/unregistered?folderId= — 미등록 process 목록
func (r *supervisorRouter) listUnregistered(c echo.Context) error {
	sess := r.requireSession(c)
	mainKey, subKey := pathParam(c, "mainKey"), pathParam(c, "subKey")
	folderID, err := strconv.ParseInt(c.QueryParam("folderId"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 folderId입니다"))
	}
	r.requireDeviceDir(c, sess.Data.ID, folderID, mainKey, subKey)
	list := r.unregistered.list(mainKey + "#" + subKey)
	out := make([]unregisteredResponse, len(list))
	for i, e := range list {
		kind := string(e.Kind)
		if kind == "" {
			kind = string(protocol.ExecTypeExec)
		}
		out[i] = unregisteredResponse{UID: e.UID, Kind: kind, Label: e.Label, Cmd: e.Cmd, Cwd: e.Cwd, PID: e.PID, StartedAt: e.StartedAt}
	}
	return c.JSON(200, out)
}

type killUnregisteredRequest struct {
	FolderID int64 `json:"folderId"`
	Force    bool  `json:"force"`
}

// killUnregistered: POST /workers/:mainKey/:subKey/unregistered/:uid/kill — DB 행이 없으니 worker 연결로 바로 MsgKill.
// 목록에서는 끝남(status)이 올 때 빠진다
func (r *supervisorRouter) killUnregistered(c echo.Context) error {
	sess := r.requireSession(c)
	mainKey, subKey, uid := pathParam(c, "mainKey"), pathParam(c, "subKey"), pathParam(c, "uid")
	var body killUnregisteredRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	conn := r.requireDeviceDir(c, sess.Data.ID, body.FolderID, mainKey, subKey)
	EndTx(c)
	if !r.unregistered.has(mainKey+"#"+subKey, uid) {
		panic(web.Err(404, "미등록 process가 아닙니다"))
	}
	callWorker(c.Request().Context(), conn, protocol.MsgKill, protocol.KillRequest{UID: uid, Force: body.Force}, nil)
	return c.JSON(200, map[string]any{})
}
