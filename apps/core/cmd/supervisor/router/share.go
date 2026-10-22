package router

import (
	"errors"
	"strconv"

	"nexus/internal/execute"
	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// 공유 걸기·풀기·목록 — 대상 소유자만(P(공유 권한), 재공유 불가). 받는 쪽 구독·타일 반영 = shareAccess.go (REF-process-share-api.md)

func (r *supervisorRouter) mountShares(e *echo.Echo) {
	e.GET("/shares/mine", r.listMyShares)
	e.GET("/nodes/:id/shares", r.listNodeShares)
	e.PUT("/nodes/:id/shares/:userId", r.putNodeShare)
	e.DELETE("/nodes/:id/shares/:userId", r.deleteNodeShare)
	e.POST("/processes/:processId/tile", r.bindProcessTile)
	e.GET("/processes/:processId/shares", r.listProcessShares)
	e.PUT("/processes/:processId/shares/:userId", r.putProcessShare)
	e.DELETE("/processes/:processId/shares/:userId", r.deleteProcessShare)
}

type shareRequest struct {
	Permission *int16 `json:"permission" validate:"required"`
	TileID     string `json:"tileId"` // 노드 공유: 공유 버튼을 누른 폴더 타일 — 받는 쪽 폴더 타일이 그 uid를 받음(FU2)
}

// shareEntry: 대상에 실제로 적용되는 한 계정의 공유
type shareEntry struct {
	UserID       int64   `json:"userId"`
	Nickname     string  `json:"nickname"`
	Permission   int16   `json:"permission"`
	Direct       bool    `json:"direct"` // 노드 = 출처가 이 노드 / process = process 단독 공유 행
	SourceNodeID *int64  `json:"sourceNodeId,omitempty"`
	SourceName   *string `json:"sourceName,omitempty"`
}

type shareChangeResponse struct {
	Entry    *shareEntry `json:"entry"`              // 바뀐 뒤 그 계정에 적용되는 공유(없으면 null)
	Previous *int16      `json:"previous,omitempty"` // PUT: 대상에 있던 직접 단계(없었으면 생략)
}

func paramUserID(c echo.Context) int64 {
	id, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 userId입니다"))
	}
	return id
}

// bindShareRequest: 저장할 수 있는 단계 = 0 차단·10 보기·20 조작·30 편집(40 소유자는 저장 안 함)
func bindShareRequest(c echo.Context) (int16, string) {
	var body shareRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	switch *body.Permission {
	case superdb.PermBlock, superdb.PermView, superdb.PermOperate, superdb.PermEdit:
		return *body.Permission, body.TileID
	}
	panic(web.Err(400, "공유 단계는 0, 10, 20, 30 중 하나입니다"))
}

// requireShareTarget: 없으면 404, 대상 소유자 자신이면 400(SH4(공유 대상 찾기))
func requireShareTarget(c echo.Context, ownerID int64) int64 {
	userID := paramUserID(c)
	if _, err := TxQueries(c).GetUser(c.Request().Context(), userID); err != nil {
		panic(web.Err(404, "계정을 찾을 수 없습니다"))
	}
	if userID == ownerID {
		panic(web.Err(400, "자기 자신에게는 공유할 수 없습니다"))
	}
	return userID
}

// ===== 내가 공유한 것 — O5 =====

// myShare: 내가 직접 건 공유 한 건(해제 = kind에 맞는 DELETE)
type myShare struct {
	Kind       string  `json:"kind"` // NODE | PROCESS
	NodeID     *int64  `json:"nodeId,omitempty"`
	ProcessUID *string `json:"processUid,omitempty"`
	Name       string  `json:"name"`
	UserID     int64   `json:"userId"`
	Nickname   string  `json:"nickname"`
	Permission int16   `json:"permission"`
}

// listMyShares: 내가 건 공유 전부(노드 + 살아 있는 process 단독)
func (r *supervisorRouter) listMyShares(c echo.Context) error {
	sess := r.requireSession(c)
	q, ctx := TxQueries(c), c.Request().Context()
	nodes, err := q.ListMyNodeShares(ctx, sess.Data.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	procs, err := q.ListMyProcessShares(ctx, sess.Data.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := make([]myShare, 0, len(nodes)+len(procs))
	for _, row := range nodes {
		id := row.NodeID
		out = append(out, myShare{Kind: "NODE", NodeID: &id, Name: row.NodeName, UserID: row.UserID, Nickname: row.Nickname, Permission: row.Permission})
	}
	for _, row := range procs {
		uid := row.ProcessUid
		out = append(out, myShare{Kind: "PROCESS", ProcessUID: &uid, Name: row.ProcessName, UserID: row.UserID, Nickname: row.Nickname, Permission: row.Permission})
	}
	return c.JSON(200, out)
}

// ===== 노드 공유 =====

func (r *supervisorRouter) nodeShareEntries(c echo.Context, nodeID int64) []shareEntry {
	rows, err := TxQueries(c).ListNodeShareEntries(c.Request().Context(), nodeID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := make([]shareEntry, len(rows))
	for i, row := range rows {
		src, name := row.SourceNodeID, row.SourceName
		out[i] = shareEntry{
			UserID: row.UserID, Nickname: row.Nickname, Permission: row.Permission,
			Direct: src == nodeID, SourceNodeID: &src, SourceName: &name,
		}
	}
	return out
}

func findShareEntry(entries []shareEntry, userID int64) *shareEntry {
	for i := range entries {
		if entries[i].UserID == userID {
			return &entries[i]
		}
	}
	return nil
}

// listNodeShares: 노드에 적용되는 공유 전부(직접 + 상위에서 이어진 것) — O2(적용 공유 목록)
func (r *supervisorRouter) listNodeShares(c echo.Context) error {
	sess := r.requireSession(c)
	node, _ := requireNode(c, sess.Data.ID, paramID(c), superdb.PermOwner)
	return c.JSON(200, r.nodeShareEntries(c, node.ID))
}

// putNodeShare: upsert, 폴더만 — SH1(공유 단위)
func (r *supervisorRouter) putNodeShare(c echo.Context) error {
	sess := r.requireSession(c)
	node, _ := requireNode(c, sess.Data.ID, paramID(c), superdb.PermOwner)
	if node.Kind != "FOLDER" {
		panic(web.Err(400, "폴더만 공유할 수 있습니다"))
	}
	userID := requireShareTarget(c, node.OwnerUserID)
	perm, tileID := bindShareRequest(c)
	q, ctx := TxQueries(c), c.Request().Context()

	var previous *int16
	old, err := q.GetNodeShare(ctx, superdb.GetNodeShareParams{NodeID: node.ID, UserID: userID})
	switch {
	case err == nil:
		previous = &old.Permission
	case !errors.Is(err, pgx.ErrNoRows):
		panic(web.Err(500, "%v", err))
	}
	before := mustNodePerm(c, userID, node.ID)
	if _, err := q.UpsertNodeShare(ctx, superdb.UpsertNodeShareParams{NodeID: node.ID, UserID: userID, Permission: perm}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	// 못 보다가 보게 될 때만 받는 쪽에 폴더 타일 — SD7(받는 쪽 타일 추가 시점)
	saved := savedTrees{}
	if before <= superdb.PermBlock && mustNodePerm(c, userID, node.ID) > superdb.PermBlock {
		saved.add(attachSharedFolder(c, userID, node.ID, sharedFromUID(c, sess.Data.ID, tileID, node.ID)))
		r.attachSharedTerminals(c, userID, liveUnder(c, node.ID), saved) // 이미 실행 중인 터미널도 — 폴더 타일을 먼저 넣어야 받는 타일이 됨
	}
	r.reconcileShareAfterCommit(c, userID, liveUnder(c, node.ID), saved)
	return c.JSON(200, shareChangeResponse{Entry: findShareEntry(r.nodeShareEntries(c, node.ID), userID), Previous: previous})
}

// deleteNodeShare: 이 노드의 공유 행만(상위에서 이어지는 공유는 그대로)
func (r *supervisorRouter) deleteNodeShare(c echo.Context) error {
	sess := r.requireSession(c)
	node, _ := requireNode(c, sess.Data.ID, paramID(c), superdb.PermOwner)
	userID := paramUserID(c)
	n, err := TxQueries(c).DeleteNodeShare(c.Request().Context(), superdb.DeleteNodeShareParams{NodeID: node.ID, UserID: userID})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if n == 0 {
		panic(web.Err(404, "이 노드에 건 공유가 없습니다"))
	}
	r.reconcileShareAfterCommit(c, userID, liveUnder(c, node.ID), nil)
	return c.JSON(200, shareChangeResponse{Entry: findShareEntry(r.nodeShareEntries(c, node.ID), userID)})
}

// ===== process 단독 공유 =====

// requireProcessOwner: 소유자 확인 + DB 행(끝난 process 포함)
func (r *supervisorRouter) requireProcessOwner(c echo.Context, userID int64) superdb.Process {
	uid := c.Param("processId")
	checkPerm(r.processPerm(uid, userID), superdb.PermOwner, "존재하지 않는 process입니다")
	rec, err := TxQueries(c).GetProcess(c.Request().Context(), uid)
	if err != nil {
		panic(web.Err(404, "존재하지 않는 process입니다"))
	}
	return rec
}

func (r *supervisorRouter) processShareEntries(c echo.Context, uid string) []shareEntry {
	rows, err := TxQueries(c).ListProcessShareEntries(c.Request().Context(), uid)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := make([]shareEntry, len(rows))
	for i, row := range rows {
		out[i] = shareEntry{UserID: row.UserID, Nickname: row.Nickname, Permission: row.Permission, Direct: row.Direct}
	}
	return out
}

func (r *supervisorRouter) listProcessShares(c echo.Context) error {
	sess := r.requireSession(c)
	rec := r.requireProcessOwner(c, sess.Data.ID)
	return c.JSON(200, r.processShareEntries(c, rec.Uid))
}

// putProcessShare: upsert, 살아 있는 process만 — SC6(process 단독 공유 대상)
func (r *supervisorRouter) putProcessShare(c echo.Context) error {
	sess := r.requireSession(c)
	rec := r.requireProcessOwner(c, sess.Data.ID)
	if rec.Status != execute.CommandPending.String() && rec.Status != execute.CommandProcess.String() {
		panic(web.Err(400, "끝난 process는 공유할 수 없습니다"))
	}
	userID := requireShareTarget(c, rec.OwnerUserID)
	perm, _ := bindShareRequest(c)
	q, ctx := TxQueries(c), c.Request().Context()

	var previous *int16
	old, err := q.GetProcessShare(ctx, superdb.GetProcessShareParams{ProcessUid: rec.Uid, UserID: userID})
	switch {
	case err == nil:
		previous = &old.Permission
	case !errors.Is(err, pgx.ErrNoRows):
		panic(web.Err(500, "%v", err))
	}
	before := mustProcessPerm(c, userID, rec.Uid)
	if _, err := q.UpsertProcessShare(ctx, superdb.UpsertProcessShareParams{ProcessUid: rec.Uid, UserID: userID, Permission: perm}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	// 못 보다가 보게 될 때만 받는 쪽에 터미널 타일 — SD7(받는 쪽 타일 추가 시점)
	saved := savedTrees{}
	if before <= superdb.PermBlock && mustProcessPerm(c, userID, rec.Uid) > superdb.PermBlock && rec.NodeID.Valid && !rec.KillRequestedAt.Valid {
		saved.add(attachTerminal(c, userID, rec.NodeID.Int64, rec.Uid, r.boundPlaceOf(userID)))
	}
	r.reconcileShareAfterCommit(c, userID, []string{rec.Uid}, saved)
	return c.JSON(200, shareChangeResponse{Entry: findShareEntry(r.processShareEntries(c, rec.Uid), userID), Previous: previous})
}

// deleteProcessShare: 단독 공유 행만(노드에서 이어지는 공유는 그대로)
func (r *supervisorRouter) deleteProcessShare(c echo.Context) error {
	sess := r.requireSession(c)
	rec := r.requireProcessOwner(c, sess.Data.ID)
	userID := paramUserID(c)
	n, err := TxQueries(c).DeleteProcessShare(c.Request().Context(), superdb.DeleteProcessShareParams{ProcessUid: rec.Uid, UserID: userID})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if n == 0 {
		panic(web.Err(404, "이 process에 건 공유가 없습니다"))
	}
	r.reconcileShareAfterCommit(c, userID, []string{rec.Uid}, nil)
	return c.JSON(200, shareChangeResponse{Entry: findShareEntry(r.processShareEntries(c, rec.Uid), userID)})
}

// bindProcessTile: "현재 타일에 바인딩"(개편-4) — 터미널 타일을 {hash} 보드의 {tileId} 아래로 옮김(없으면 새로, 닫혔으면 루트). 보기(10) 이상
// 타일 고정은 안 건드림(R6(바인딩 단어 겹침)). 이미 그 아래면 added=false
func (r *supervisorRouter) bindProcessTile(c echo.Context) error {
	sess := r.requireSession(c)
	uid := c.Param("processId")
	checkPerm(r.processPerm(uid, sess.Data.ID), superdb.PermView, "존재하지 않는 process입니다")
	q, ctx := TxQueries(c), c.Request().Context()
	rec, err := q.GetProcess(ctx, uid)
	if err != nil {
		panic(web.Err(404, "존재하지 않는 process입니다"))
	}
	if !rec.NodeID.Valid {
		panic(web.Err(400, "타일을 만들 수 없는 process입니다"))
	}
	var body struct {
		Hash   string `json:"hash"` // "" = 기본 보드
		TileID string `json:"tileId"`
	}
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}

	userID := sess.Data.ID
	def := ensureTileTree(c, userID)
	locked := lockWorkspaces(c, userID)
	hash := body.Hash
	if hash == "" {
		hash = def.Hash
	}
	mustWorkspace(locked, hash)
	docs := make(map[string]*tileDoc, len(locked))
	for h, wr := range locked {
		d := mustParseDoc(wr.Tree)
		docs[h] = &d
	}
	dst := docs[hash]
	parentID := body.TileID
	if p, ok := dst.Tiles[parentID]; !ok || !p.canParent() {
		parentID = dst.RootID
	}

	var from *protocol.TileSpot
	var tile tileItem
	for h, d := range docs {
		t, ok := d.terminalOf(uid)
		if !ok {
			continue
		}
		if h == hash && t.Parent != nil && *t.Parent == parentID {
			return c.JSON(200, map[string]bool{"added": false})
		}
		from, tile = &protocol.TileSpot{Hash: h, TileID: t.ID}, t
		d.removeLeaf(t.ID)
		break
	}
	if from == nil {
		if rec.Status != execute.CommandPending.String() && rec.Status != execute.CommandProcess.String() {
			panic(web.Err(400, "끝난 process는 열 수 없습니다"))
		}
		nodeID := rec.NodeID.Int64
		tile = newTerminalTile(nil, nodeID, uid, defaultTileSize)
	}
	tile.Parent, tile.Minimized = &parentID, false
	parent := dst.Tiles[parentID]
	parent.Kids = append(parent.Kids, tile.ID)
	dst.Tiles[parentID] = parent
	dst.Tiles[tile.ID] = tile

	saved := savedTrees{}
	upd := saveLockedTree(c, locked[hash], dst)
	saved.add(&upd)
	if from != nil && from.Hash != hash {
		src := saveLockedTree(c, locked[from.Hash], docs[from.Hash])
		saved.add(&src)
	}
	ev := protocol.TilesNoticeEvent{Kind: protocol.NoticeBind, UID: uid, NodeID: rec.NodeID.Int64, Name: nodeName(ctx, q, rec.NodeID.Int64),
		From: from, To: &protocol.TileSpot{Hash: hash, TileID: tile.ID}}
	AfterCommit(c, func() {
		for _, t := range saved {
			r.publishTiles(userID, t, "")
		}
		r.publishNotice(userID, ev)
	})
	return c.JSON(200, map[string]bool{"added": true})
}
