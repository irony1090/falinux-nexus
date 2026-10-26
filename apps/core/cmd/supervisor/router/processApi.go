package router

import (
	"context"
	"errors"
	"log"
	"slices"
	"time"

	"nexus/cmd/supervisor/process"
	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

const tabIDHeader = "X-Tab-Id" // REST 요청이 어느 탭에서 왔는지(TAB:ID로 받은 값)

func (r *supervisorRouter) mountProcesses(e *echo.Group) {
	g := e.Group("/processes")
	g.GET("", r.listProcesses)
	g.GET("/lookup", r.lookupProcesses)
	g.POST("/exec", r.execProcess)
	g.POST("/kill/:processId", r.killProcess)
	g.POST("/resize/:processId", r.resizeProcess)
	g.POST("/size-owner/:processId", r.takeProcessSizeOwner)
	g.GET("/snapshot/:processId", r.snapshotProcess)
}

// requestTab: 요청 헤더의 탭 id. 없거나 요청 세션의 탭이 아니면 ""
func (r *supervisorRouter) requestTab(c echo.Context, sid string) string {
	tabID := c.Request().Header.Get(tabIDHeader)
	if tabID == "" || !r.tabs.belongsTo(tabID, sid) {
		return ""
	}
	return tabID
}

func (r *supervisorRouter) requireTab(c echo.Context, sid string) string {
	tabID := r.requestTab(c, sid)
	if tabID == "" {
		panic(web.Err(400, "탭 id가 없거나 이 세션의 탭이 아닙니다"))
	}
	return tabID
}

// listProcesses: 볼 수 있는 살아 있는 process + 요청 탭 기준 sizeOwner(U(목록 API))
func (r *supervisorRouter) listProcesses(c echo.Context) error {
	sess := r.requireSession(c)
	tabID := r.requestTab(c, sess.Name())

	processes, err := r.processManager.ListVisibleLive(sess.Data.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}

	rows := liveRows(processes)
	out := withProcessNames(c, sess.Data.ID, newProcessResponses(rows), processesOf(rows))
	for i := range out {
		out[i] = withSizeOwner(out[i], tabID != "" && r.sizeOwnerOf(out[i].Uid) == tabID)
	}
	return c.JSON(200, out)
}

const lookupMax = 256 // 한 번에 조회할 uid 개수 상한(타일 수보다 넉넉히)

// lookupProcesses: ?uids=의 process를 상태와 무관하게(볼 수 없는 것은 빠짐) — 3-d①(끝난 process 조회), 새로고침 뒤 타일 복원용
func (r *supervisorRouter) lookupProcesses(c echo.Context) error {
	sess := r.requireSession(c)
	tabID := r.requestTab(c, sess.Name())

	uids := c.QueryParams()["uids"]
	if len(uids) > lookupMax {
		panic(web.Err(400, "uid는 한 번에 %d개까지 조회할 수 있습니다", lookupMax))
	}
	if len(uids) == 0 {
		return c.JSON(200, []processResponse{})
	}
	processes, err := TxQueries(c).ListVisibleProcessesByUids(c.Request().Context(), superdb.ListVisibleProcessesByUidsParams{
		UserID: sess.Data.ID,
		Uids:   uids,
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}

	rows := uidRows(processes)
	out := withProcessNames(c, sess.Data.ID, newProcessResponses(rows), processesOf(rows))
	for i := range out {
		out[i] = withSizeOwner(out[i], tabID != "" && r.sizeOwnerOf(out[i].Uid) == tabID)
	}
	return c.JSON(200, out)
}

// execRequest: POST /processes/exec
type execRequest struct {
	NodeID  int64  `json:"nodeId" validate:"required"`
	AuthKey string `json:"authKey" validate:"required"`               // worker 인스턴스 키(main#sub). 후보 = GET /workers?nodeId=
	Type    string `json:"type" validate:"omitempty,oneof=EXEC EDIT"` // 비우면 EXEC

	Workspace    string    `json:"workspace"`    // 터미널 타일을 넣을 작업 공간 hash. 비우면 기본
	ParentTileID string    `json:"parentTileId"` // 터미널 타일을 붙일 폴더 타일. 없거나 닫혔으면 루트
	Size         *tileSize `json:"size"`         // 없으면 0.5×0.5
}

// execResponse: process 응답 필드 그대로 + 서버 트리에 넣은 터미널 타일(process 없는 실행이면 생략)
type execResponse struct {
	processResponse
	Tile        *tileItem `json:"tile,omitempty"`
	TileVersion int64     `json:"tileVersion,omitempty"`
}

// execProcess: EXEC = 조작(20) / EDIT = 편집(30) 이상. 요청 탭이 크기 소유자(router.Exec)
func (r *supervisorRouter) execProcess(c echo.Context) error {
	sess := r.requireSession(c)
	tabID := r.requireTab(c, sess.Name())
	var body execRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	size := execTileSize(c, sess.Data.ID, body.Workspace, body.Size)

	min := superdb.PermOperate
	if protocol.ExecType(body.Type) == protocol.ExecTypeEdit {
		min = superdb.PermEdit
	}
	node, perm := requireNode(c, sess.Data.ID, body.NodeID, min)
	if node.Kind != superdb.NodeKindScript {
		panic(web.Err(400, "스크립트만 실행할 수 있습니다"))
	}
	// 다른 장비로 잘못 실행되는 것 방지: authKey의 main_key = 스크립트의 상속 device_key (서브키 위조 검증은 별도)
	if main, _ := splitInstanceKey(body.AuthKey); main != resolveMainKey(c.Request().Context(), TxQueries(c), body.NodeID, node.OwnerUserID) {
		panic(web.Err(400, "이 스크립트의 실행 대상 장비가 아닙니다"))
	}
	// Windows는 확장자로 실행기를 고른다(worker prepareExec와 같은 NodeExt 판정) — WW1(실행 규칙)·WW5(OS 표시)
	if os, _ := r.workerOS.Get(body.AuthKey); os == "windows" && protocol.ExecType(body.Type) != protocol.ExecTypeEdit && process.NodeExt(node.Name) == "" {
		panic(web.Err(400, "Windows에서는 노드 이름에 확장자(.ps1·.bat·.py 등)가 필요합니다"))
	}

	uid, err := r.Exec(sess.Data.ID, body.AuthKey, protocol.ExecType(body.Type), node, tabID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	// 새 process엔 process 공유 행이 없어 단계 = 노드 단계
	res := r.execTile(c, sess.Data.ID, tabID, uid, perm, body.Workspace, body.ParentTileID, size)
	r.attachViewerTiles(sess.Data.ID, node.ID, uid)
	return c.JSON(200, res)
}

// execTileSize: 실행 요청의 타일 크기(없으면 0.5×0.5). 작업 공간은 실행 전에 확인 — 없는 작업 공간이면 실행하지 않음
func execTileSize(c echo.Context, userID int64, workspace string, req *tileSize) tileSize {
	size := tileSize{W: 0.5, H: 0.5}
	if req != nil {
		size = *req
	}
	if !validSpan(size.W) || !validSpan(size.H) {
		panic(web.Err(400, "타일 크기가 올바르지 않습니다"))
	}
	if workspace != "" {
		requestWorkspace(c, userID, workspace)
	}
	return size
}

// execTile: 시작한 process의 응답 + 실행한 계정의 트리에 터미널 타일. 스크립트 실행·장비 파일 실행(workerExec.go) 공용
func (r *supervisorRouter) execTile(c echo.Context, userID int64, tabID, uid string, perm int16, workspace, parentTileID string, size tileSize) execResponse {
	entry, _ := r.processManager.Get(uid)
	rec := *entry.Record
	res := execResponse{processResponse: withPermission(withSizeOwner(newProcessResponse(rec), r.sizeOwnerOf(uid) == tabID), perm)}
	res.processResponse = withProcessNames(c, userID, []processResponse{res.processResponse}, []superdb.Process{rec})[0]
	// 타일을 못 넣으면 실행을 되돌린다 — 보이지도 kill되지도 않는 process를 남기지 않기 위해. kill도 실패하면 3-d 안전망이 받음
	// 요청 컨텍스트를 쓰면 브라우저가 응답 전에 끊을 때 저장이 취소돼 멀쩡한 process가 kill된다
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tile, tree, err := appendTerminalTile(ctx, userID, workspace, parentTileID, size, nullID(rec.NodeID), uid)
	if err != nil {
		if kErr := entry.Inter.Kill(true); kErr != nil {
			log.Printf("[process] 타일 저장 실패 후 kill 실패 uid=%s: %v", uid, kErr)
		}
		panic(web.Err(500, "터미널 타일 저장에 실패해 실행을 중단했습니다: %v", err))
	}
	r.publishTiles(userID, tree, tabID) // 별도 트랜잭션이 이미 커밋됨
	res.Tile, res.TileVersion = &tile, tree.Version
	return res
}

// killRequest: POST /processes/kill/:processId — force = SIGKILL, 아니면 SIGTERM
type killRequest struct {
	Force bool `json:"force"`
}

// killProcess: 종료 요청을 DB에 기록하고 커밋 뒤 worker에 — 끊겨 있으면 재접속 때 reconcileReconnect가 보낸다(REF-process-kill.md)
func (r *supervisorRouter) killProcess(c echo.Context) error {
	sess := r.requireSession(c)
	uid := c.Param("processId")
	var body killRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	perm := r.processPerm(uid, sess.Data.ID)
	checkPerm(perm, superdb.PermOperate, "존재하지 않는 process입니다")

	row, err := TxQueries(c).RequestProcessKill(c.Request().Context(), superdb.RequestProcessKillParams{Uid: uid, Force: body.Force})
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "실행 중인 process가 아닙니다"))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	AfterCommit(c, func() { r.sendKill(row) })

	res := withPermission(newProcessResponse(row), perm)
	return c.JSON(200, withProcessNames(c, sess.Data.ID, []processResponse{res}, []superdb.Process{row})[0])
}

// sendKill: 커밋된 종료 요청을 발행하고 접속 중인 worker에 보낸다(끊겨 있으면 안 보냄)
func (r *supervisorRouter) sendKill(row superdb.Process) {
	r.publishProcess(row)
	entry, ok := r.processManager.Get(row.Uid)
	if !ok || entry.Inter == nil {
		return
	}
	entry.SetRecord(&row)
	if err := entry.Inter.Kill(row.KillForced); err != nil {
		log.Printf("[process] 종료 전달 실패 uid=%s force=%v: %v", row.Uid, row.KillForced, err)
	}
}

// forceAbandoned: 종료 요청됐는데 조작 이상 계정의 보드 어디에도 타일이 없으면 강제 종료로 — 종료-4(모두 닫음). 타일 닫기·공유 변경 때만(종료 요청 때는 안 봄)
func (r *supervisorRouter) forceAbandoned(uids []string) {
	if len(uids) == 0 {
		return
	}
	rows, err := store.GetStorePool().Queries().ForceAbandonedKills(context.Background(), superdb.ForceAbandonedKillsParams{Uids: uids, MinPerm: superdb.PermOperate})
	if err != nil {
		log.Printf("[process] 모두 닫음 판정 실패 uids=%v: %v", uids, err)
		return
	}
	for _, row := range rows {
		r.sendKill(row)
	}
}

// resizeRequest: POST /processes/resize/:processId
type resizeRequest struct {
	Rows uint16 `json:"rows" validate:"required"`
	Cols uint16 `json:"cols" validate:"required"`
}

// resizeProcess: worker가 확인한 크기만 DB·메모리에 쓰고 커밋 뒤 PROCESS:UPDATE(뒤따르는 STATUS가 없어 여기서 확정, REF-process-trigger.md)
func (r *supervisorRouter) resizeProcess(c echo.Context) error {
	sess := r.requireSession(c)
	uid := c.Param("processId")
	tabID := r.requireTab(c, sess.Name())

	var body resizeRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}

	entry := r.requireProcess(uid, sess.Data.ID, superdb.PermOperate)
	if !r.claimSizeOwner(uid, tabID) {
		panic(web.Err(403, "크기를 바꿀 수 있는 탭이 아닙니다"))
	}
	// 재발 방지 — 밀림-가(먼저 발행): worker에 보내기 전에 고정 탭들에 새 크기를 먼저 알린다.
	// Publish는 동기 쓰기라 SIGWINCH 뒤 다시 그린 출력(DATA)보다 반드시 앞선다. 커밋 뒤에만 알리면
	// 고정 탭이 새 크기용 출력을 옛 크기 xterm에 그려 화면이 밀린다(REF-size-take.md "실측-3 결과")
	prev := *entry.Record
	next := prev
	next.Rows, next.Cols = int16(body.Rows), int16(body.Cols)
	r.publishProcess(next)
	if err := entry.Inter.Layout(body.Cols, body.Rows); err != nil {
		r.publishProcess(prev) // 고정 탭을 옛 크기로 되돌린다
		panic(web.Err(500, "resize 실패: %v", err))
	}

	row, err := TxQueries(c).UpdateProcessLayout(c.Request().Context(), superdb.UpdateProcessLayoutParams{
		Uid:  uid,
		Rows: int16(body.Rows),
		Cols: int16(body.Cols),
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	entry.SetRecord(&row)
	AfterCommit(c, func() { r.publishProcess(row) })

	res := withPermission(withSizeOwner(newProcessResponse(row), true), r.processPerm(uid, sess.Data.ID))
	return c.JSON(200, withProcessNames(c, sess.Data.ID, []processResponse{res}, []superdb.Process{row})[0])
}

// takeProcessSizeOwner: 요청 탭이 크기 우선권을 바로 가져온다 — 크기-1(가져오기 API). PTY 크기는 가져온 탭이 resize로 맞춘다
func (r *supervisorRouter) takeProcessSizeOwner(c echo.Context) error {
	sess := r.requireSession(c)
	uid := c.Param("processId")
	tabID := r.requireTab(c, sess.Name())
	r.requireProcess(uid, sess.Data.ID, superdb.PermOperate)
	owner := r.processPerm(uid, sess.Data.ID) == superdb.PermOwner // 스크립트 소유 계정(SH9) — K4(소유 계정 우선)

	switch err := r.takeSizeOwner(uid, tabID, sess.Data.ID, owner); {
	case errors.Is(err, errSizeOtherAccount):
		panic(web.Err(403, "다른 계정이 이 터미널의 크기를 쓰는 중입니다"))
	case errors.Is(err, errSizeNotCandidate):
		panic(web.Err(403, "이 화면에 없는 터미널입니다"))
	}
	return c.JSON(200, map[string]any{})
}

// attachViewerTiles: 실행자 외 볼 수 있는 계정 전부에 터미널 타일 + push(계정 id 순, SD2(실행 때 터미널 타일을 넣을 계정)). 실패·자리 없음 = 안 붙음(위치 없음)
func (r *supervisorRouter) attachViewerTiles(starter, nodeID int64, uid string) {
	viewers, err := store.GetStorePool().Queries().ListProcessViewers(context.Background(), uid)
	if err != nil {
		log.Printf("[process] 터미널 타일 대상 조회 실패 uid=%s: %v", uid, err)
		return
	}
	users := make([]int64, 0, len(viewers))
	for _, v := range viewers {
		if v.UserID != starter {
			users = append(users, v.UserID)
		}
	}
	slices.Sort(users)
	name := ""
	if len(users) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		name = nodeName(ctx, store.GetStorePool().Queries(), nodeID)
		cancel()
	}
	for _, userID := range users {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		tree, err := attachViewerTerminal(ctx, userID, nodeID, uid, r.boundPlaceOf(userID))
		cancel()
		if err != nil {
			log.Printf("[process] 터미널 타일 추가 실패 user=%d uid=%s: %v", userID, uid, err)
			continue
		}
		// M2(자동으로 붙음)·M4(위치 없음) — 새 실행이라 이미 타일이 있어 nil인 경우는 없음
		ev := protocol.TilesNoticeEvent{Kind: protocol.NoticeNoPlace, UID: uid, NodeID: nodeID, Name: name}
		if tree != nil {
			r.publishTiles(userID, *tree, "")
			ev.Kind, ev.To = protocol.NoticeAttach, terminalSpot(*tree, uid)
		}
		r.publishNotice(userID, ev)
	}
}
