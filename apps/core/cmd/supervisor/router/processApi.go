package router

import (
	"context"
	"log"
	"time"

	"nexus/cmd/supervisor/process"
	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

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

// requireOwnedEntry는 실행 중인 uid의 entry를 반환한다. 없거나 요청 계정 것이 아니면 404(존재를 드러내지 않음).
func (r *supervisorRouter) requireOwnedEntry(uid string, userID int64) *process.ProcessEntry {
	entry, ok := r.processManager.Get(uid)
	if !ok || entry.Inter == nil || entry.Record == nil || entry.Record.OwnerUserID != userID {
		panic(web.Err(404, "존재하지 않는 process입니다"))
	}
	return entry
}

// listProcesses는 그 계정의 살아 있는 process를 요청 탭 기준 sizeOwner와 함께 반환한다(U(목록 API)).
func (r *supervisorRouter) listProcesses(c echo.Context) error {
	sess := r.requireSession(c)
	tabID := r.requestTab(c, sess.Name())

	processes, err := r.processManager.ListLive(sess.Data.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}

	out := newProcessResponses(processes)
	for i := range out {
		out[i] = withSizeOwner(out[i], tabID != "" && r.sizeOwnerOf(out[i].Uid) == tabID)
	}
	return c.JSON(200, out)
}

const lookupMax = 256 // 한 번에 조회할 uid 개수 상한(타일 수보다 넉넉히)

// lookupProcesses는 요청한 uid(?uids=a&uids=b)의 process를 상태와 무관하게 반환한다. 남의 것·없는 uid는 빠짐
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
	processes, err := TxQueries(c).ListProcessesByUids(c.Request().Context(), superdb.ListProcessesByUidsParams{
		OwnerUserID: sess.Data.ID,
		Uids:        uids,
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}

	out := newProcessResponses(processes)
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
// 그 계정의 연결된 소켓은 relay 기동 전에 전부 구독되고, 요청 탭이 크기 소유자가 된다.
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

	node, err := TxQueries(c).GetNode(c.Request().Context(), superdb.GetNodeParams{
		ID:          body.NodeID,
		OwnerUserID: sess.Data.ID,
	})
	if err != nil {
		panic(web.Err(404, "노드를 찾을 수 없습니다"))
	}
	// 다른 장비로 잘못 실행되는 것 방지: authKey의 main_key = 스크립트의 상속 device_key (서브키 위조 검증은 별도)
	if main, _ := splitInstanceKey(body.AuthKey); main != resolveMainKey(c.Request().Context(), TxQueries(c), body.NodeID, sess.Data.ID) {
		panic(web.Err(400, "이 스크립트의 실행 대상 장비가 아닙니다"))
	}

	uid, err := r.Exec(*sess.Data, body.AuthKey, protocol.ExecType(body.Type), node, tabID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	entry, _ := r.processManager.Get(uid)
	res := execResponse{processResponse: withSizeOwner(newProcessResponse(*entry.Record), r.sizeOwnerOf(uid) == tabID)}
	if !entry.HasProcess() {
		return c.JSON(200, res)
	}
	// 타일을 못 넣으면 실행을 되돌린다 — 보이지도 kill되지도 않는 process를 남기지 않기 위해. kill도 실패하면 3-d 안전망이 받음
	// 요청 컨텍스트를 쓰면 브라우저가 응답 전에 끊을 때 저장이 취소돼 멀쩡한 process가 kill된다
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tile, tree, err := appendTerminalTile(ctx, sess.Data.ID, body.ParentTileID, size, node.ID, uid)
	if err != nil {
		if kErr := entry.Inter.Kill(); kErr != nil {
			log.Printf("[process] 타일 저장 실패 후 kill 실패 uid=%s: %v", uid, kErr)
		}
		panic(web.Err(500, "터미널 타일 저장에 실패해 실행을 중단했습니다: %v", err))
	}
	r.publishTiles(sess.Data.ID, tree.Tree, tree.Version, tabID) // 별도 트랜잭션이 이미 커밋됨
	res.Tile, res.TileVersion = &tile, tree.Version
	return c.JSON(200, res)
}

// killProcess는 실행 중인 process를 종료한다(worker에 MsgKill 전달, entry.Inter.Kill()).
// 종료 STATUS는 relay가 그 계정의 구독 소켓 전부에 발행한다.
func (r *supervisorRouter) killProcess(c echo.Context) error {
	sess := r.requireSession(c)
	entry := r.requireOwnedEntry(c.Param("processId"), sess.Data.ID)
	if err := entry.Inter.Kill(); err != nil {
		panic(web.Err(500, "%v", err))
	}
	return c.JSON(200, map[string]any{})
}

// resizeRequest: 터미널 창 크기 변경 요청. (POST /processes/resize/:processId)
type resizeRequest struct {
	Rows uint16 `json:"rows" validate:"required"`
	Cols uint16 `json:"cols" validate:"required"`
}

// resizeProcess는 실행 중인 process의 PTY 창 크기를 바꾼다. entry.Inter.Layout이 worker에
// MsgResize(REQ)를 전달하고(process/manager.go newWorkerInteractive의 onLayout) 그 응답을
// 그대로 기다린다 — folder-open(Inter=nil)이거나 이미 종료된 uid는 404, worker가 거부/끊김이면
// 500이라 DB/memory엔 손대지 않는다. worker가 실제로 확인해 준 값만 "진짜 결과"로 취급해
// DB(UpdateProcessLayout)+memory(entry.SetRecord)를 갱신하고, 커밋 성공 후에만
// PROCESS:<uid> 토픽에 MsgProcessUpdate로 알린다(kill/exec와 달리 뒤이어 오는 STATUS 이벤트가
// 없어 이 REST 안에서 동기로 확정 짓는다 — REF-process-trigger.md 참조).
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

	entry := r.requireOwnedEntry(uid, sess.Data.ID)
	if !r.claimSizeOwner(uid, tabID) {
		panic(web.Err(403, "크기를 바꿀 수 있는 탭이 아닙니다"))
	}
	if err := entry.Inter.Layout(body.Cols, body.Rows); err != nil {
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

	return c.JSON(200, withSizeOwner(newProcessResponse(row), true))
}
