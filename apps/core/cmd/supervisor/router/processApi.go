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

func (r *supervisorRouter) mountProcesses(e *echo.Echo) {
	g := e.Group("/processes")
	g.GET("", r.listProcesses)
	g.GET("/lookup", r.lookupProcesses)
	g.POST("/exec", r.execProcess)
	g.POST("/kill/:processId", r.killProcess)
	g.POST("/resize/:processId", r.resizeProcess)
	g.POST("/size-owner/:processId", r.takeProcessSizeOwner)
	g.GET("/snapshot/:processId", r.snapshotProcess)
}

// requestTab은 요청 헤더의 탭 id를 반환한다. 없거나 요청 세션의 탭이 아니면 "".
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

// listProcesses는 그 계정이 볼 수 있는 살아 있는 process를 요청 탭 기준 sizeOwner와 함께 반환한다(U(목록 API)).
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

// lookupProcesses는 요청한 uid(?uids=a&uids=b)의 process를 상태와 무관하게 반환한다. 볼 수 없는 것·없는 uid는 빠짐
// — 3-d①(끝난 process 조회): 새로고침 뒤 터미널 타일 복원용, 살아 있는 것 전체는 listProcesses
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

// execRequest: 노드 실행 요청. (POST /processes/exec)
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

// execProcess는 frontend의 "이 노드 실행" 요청을 받아 worker에 명령한다(router.Exec 위임).
// EXEC = 조작(20) / EDIT(vi 편집) = 편집(30). 볼 수 있는 계정의 연결된 소켓은 relay 기동 전에 전부 구독되고, 요청 탭이 크기 소유자가 된다.
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
	size := tileSize{W: 0.5, H: 0.5}
	if body.Size != nil {
		size = *body.Size
	}
	if !validSpan(size.W) || !validSpan(size.H) {
		panic(web.Err(400, "타일 크기가 올바르지 않습니다"))
	}

	if body.Workspace != "" {
		requestWorkspace(c, sess.Data.ID, body.Workspace) // 실행 전에 확인 — 없는 작업 공간이면 실행하지 않음
	}

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
	entry, _ := r.processManager.Get(uid)
	// 새 process엔 process 공유 행이 없어 단계 = 노드 단계
	res := execResponse{processResponse: withPermission(withSizeOwner(newProcessResponse(*entry.Record), r.sizeOwnerOf(uid) == tabID), perm)}
	res.processResponse = withProcessNames(c, sess.Data.ID, []processResponse{res.processResponse}, []superdb.Process{*entry.Record})[0]
	// 타일을 못 넣으면 실행을 되돌린다 — 보이지도 kill되지도 않는 process를 남기지 않기 위해. kill도 실패하면 3-d 안전망이 받음
	// 요청 컨텍스트를 쓰면 브라우저가 응답 전에 끊을 때 저장이 취소돼 멀쩡한 process가 kill된다
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tile, tree, err := appendTerminalTile(ctx, sess.Data.ID, body.Workspace, body.ParentTileID, size, node.ID, uid)
	if err != nil {
		if kErr := entry.Inter.Kill(true); kErr != nil {
			log.Printf("[process] 타일 저장 실패 후 kill 실패 uid=%s: %v", uid, kErr)
		}
		panic(web.Err(500, "터미널 타일 저장에 실패해 실행을 중단했습니다: %v", err))
	}
	r.publishTiles(sess.Data.ID, tree, tabID) // 별도 트랜잭션이 이미 커밋됨
	r.attachViewerTiles(sess.Data.ID, node.ID, uid)
	res.Tile, res.TileVersion = &tile, tree.Version
	return c.JSON(200, res)
}

// killRequest: 종료 요청. (POST /processes/kill/:processId) force = 강제 종료(SIGKILL), 아니면 SIGTERM
type killRequest struct {
	Force bool `json:"force"`
}

// killProcess는 종료 요청을 DB에 기록하고 커밋 뒤 worker에 보낸다. worker가 끊겨 있어도(PENDING) 기록만 하고
// 재접속 때 reconcileReconnect가 보낸다 — REF-process-kill.md. 종료 STATUS는 relay가 구독 소켓 전부에 발행한다.
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

// sendKill은 커밋된 종료 요청을 발행하고 접속 중인 worker에 보낸다(force = row.KillForced). 커밋 뒤에만 부를 것.
// 끊겨 있으면(memory에 없음) 아무것도 안 보낸다 — 재접속 때 reconcileReconnect가 DB 기록을 보고 보낸다
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

// forceAbandoned는 uids 중 종료 요청됐는데 조작 이상 계정의 보드 어디에도 터미널 타일이 없는 것을 강제 종료로 바꿔 보낸다 — 종료-4(모두 닫음).
// 타일 닫기·공유 변경의 커밋 뒤에만 부른다. 종료 요청 때는 판정하지 않는다(그때 이미 타일이 없어도 SIGTERM만 — 2026-10-12 결정)
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

// resizeRequest: 터미널 창 크기 변경 요청. (POST /processes/resize/:processId)
type resizeRequest struct {
	Rows uint16 `json:"rows" validate:"required"`
	Cols uint16 `json:"cols" validate:"required"`
}

// resizeProcess는 실행 중인 process의 PTY 창 크기를 바꾼다. entry.Inter.Layout이 worker에
// MsgResize(REQ)를 전달하고(process/manager.go newWorkerInteractive의 onLayout) 그 응답을
// 그대로 기다린다 — 이미 종료된 uid는 404, worker가 거부/끊김이면
// 500이라 DB/memory엔 손대지 않는다. worker가 실제로 확인해 준 값만 "진짜 결과"로 취급해
// DB(UpdateProcessLayout)+memory(entry.SetRecord)를 갱신하고, 커밋 성공 후
// PROCESS:<uid> 토픽에 MsgProcessUpdate로 알린다(kill/exec와 달리 뒤이어 오는 STATUS 이벤트가
// 없어 이 REST 안에서 동기로 확정 짓는다 — REF-process-trigger.md 참조).
// 단 새 크기는 worker에 보내기 전에 한 번 먼저 알린다(아래 밀림-가(먼저 발행)).
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
	// 재발 방지 — 밀림-가(먼저 발행): 새 크기를 worker에 보내기 전에 고정 탭들에 먼저 알린다.
	// Publish는 구독 소켓에 동기로 쓰므로, 이 뒤 SIGWINCH로 앱이 다시 그린 출력(DATA)보다 반드시 앞선다.
	// 응답·커밋 뒤에만 알리면 고정 탭이 새 크기용 출력을 옛 크기 xterm에 그려 밀린 화면이 남는다(REF-size-take.md "실측-3 결과").
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

// attachViewerTiles: 실행자 외에 볼 수 있는 계정 전부의 트리에 터미널 타일을 넣고 push한다(계정 id 순) — SD2.
// 실패·자리 없음은 그 계정 보드에 안 붙을 뿐(위치 없음 — 스크립트 행 "내 보드에 열기"). 소유자도 고아 안전망 대상이 아님 — BD8(고아 안전망)
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
