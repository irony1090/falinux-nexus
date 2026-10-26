package router

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strconv"

	"nexus/internal/execute"
	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// ===== 미등록 process 가져오기 — 가져오기-1(화면 다시 요청)~가져오기-3(등록 API), REF-supervisor-move.md "가져오기" =====
// supervisor를 옮긴 뒤 worker에서 돌고 있는 process를 이 supervisor의 장비 주인이 자기 process로 등록하고 터미널 타일로 띄운다.
// 주인 = 가져온 계정. worker가 옛 supervisor로 돌아가면 거기 행으로 이어진다(붙은 쪽이 주인)

type adoptRequest struct {
	FolderID     int64     `json:"folderId" validate:"required"` // 장비 타일을 연 폴더(인가·node 없을 때 dir_node_id)
	Workspace    string    `json:"workspace"`
	ParentTileID string    `json:"parentTileId"` // 장비 타일. 없거나 닫혔으면 루트
	Size         *tileSize `json:"size"`
}

// POST /workers/:mainKey/:subKey/unregistered/:uid/adopt — EXEC만(EDIT는 결과를 저장할 노드를 확정할 수 없어 끄기만)
func (r *supervisorRouter) adoptUnregistered(c echo.Context) error {
	sess := r.requireSession(c)
	tabID := r.requireTab(c, sess.Name())
	mainKey, subKey, uid := pathParam(c, "mainKey"), pathParam(c, "subKey"), pathParam(c, "uid")
	var body adoptRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	size := execTileSize(c, sess.Data.ID, body.Workspace, body.Size)
	worker := r.requireDeviceDir(c, sess.Data.ID, body.FolderID, mainKey, subKey)
	key := mainKey + "#" + subKey

	e, ok := r.unregistered.get(key, uid)
	if !ok {
		panic(web.Err(404, "미등록 process가 아닙니다(이미 가져갔거나 끝났을 수 있습니다)"))
	}
	if e.Kind == protocol.ExecTypeEdit {
		panic(web.Err(400, "편집 process는 가져올 수 없습니다"))
	}
	nodeID := adoptNode(c, sess.Data.ID, mainKey, e)
	dirNodeID := pgtype.Int8{}
	if !nodeID.Valid {
		dirNodeID = pgtype.Int8{Int64: body.FolderID, Valid: true}
	}
	EndTx(c) // 인가 끝 — 아래는 process 등록(자체 연결)·worker 호출

	// 꺼내기는 등록 직전에 — 가져오기가 둘 겹쳐도 하나만 통과
	if _, ok := r.unregistered.take(key, uid); !ok {
		panic(web.Err(404, "미등록 process가 아닙니다(이미 가져갔거나 끝났을 수 있습니다)"))
	}
	entry, err := r.processManager.Adopt(worker, key, uid, e.Cmd, e.Cwd, nodeID, dirNodeID, sess.Data.ID)
	if err != nil {
		r.unregistered.put(key, e)
		panic(web.Err(500, "%v", err))
	}
	// 순서 의존: 재접속 Rebind와 같은 순서 — 구독 -> relay -> 상태. 화면 요청은 relay가 뜬 뒤(MsgScreen이 entry를 찾아야 ring에 들어감)
	r.subscribeViewers(uid)
	r.startRelay(uid, entry.Inter)
	r.applyStatus(uid, execute.CommandProcess, e.PID, 0)
	r.setSizeOwner(uid, tabID)
	ctx, cancel := context.WithTimeout(context.Background(), sendCallTimeout)
	defer cancel()
	if _, err := worker.Call(ctx, protocol.MsgScreenReq, protocol.ScreenRequest{UID: uid}); err != nil {
		log.Printf("[adopt] 화면 다시 요청 실패 uid=%s: %v", uid, err) // 옛 worker = 가져온 뒤 출력만 보임
	}
	return c.JSON(200, r.execTile(c, sess.Data.ID, tabID, uid, superdb.PermOwner, body.Workspace, body.ParentTileID, size))
}

// 스크립트 실행 경로 = {WORKER_BASE}/<node.ID>/<uid><확장자>(process/path.go) — worker 보관 목록의 명령에서 옛 node id를 읽는다
func nodeIDInCmd(cmd, uid string) (int64, bool) {
	m := regexp.MustCompile(`[/\\](\d+)[/\\]` + regexp.QuoteMeta(uid)).FindStringSubmatch(cmd)
	if m == nil {
		return 0, false
	}
	id, err := strconv.ParseInt(m[1], 10, 64)
	return id, err == nil
}

// adoptNode: 가져오기-2(node 찾기). 명령의 node id가 이 DB에서도 같은 스크립트일 때만 연결 —
// 내 노드 + SCRIPT + 이름 = 보관 목록 이름 + 들어 있는 폴더의 장비 = 이 장비. DB를 통째로 옮긴 경우만 맞고, 아니면 node 없이(장비 파일 실행처럼)
func adoptNode(c echo.Context, userID int64, mainKey string, e protocol.SyncEntry) pgtype.Int8 {
	id, ok := nodeIDInCmd(e.Cmd, e.UID)
	if !ok || e.Label == "" {
		return pgtype.Int8{}
	}
	q, ctx := TxQueries(c), c.Request().Context()
	node, err := q.GetNode(ctx, superdb.GetNodeParams{ID: id, OwnerUserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return pgtype.Int8{}
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if node.Kind != superdb.NodeKindScript || node.Name != e.Label || !node.ParentID.Valid {
		return pgtype.Int8{}
	}
	key, err := q.ResolveDeviceKey(ctx, superdb.ResolveDeviceKeyParams{ID: node.ParentID.Int64, OwnerUserID: userID})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(500, "%v", err))
	}
	if !key.Valid || key.String != mainKey {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: node.ID, Valid: true}
}
