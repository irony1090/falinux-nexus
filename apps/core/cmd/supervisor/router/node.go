package router

import (
	"context"
	"log"
	"slices"
	"strconv"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// nodes 카탈로그 핸들러 — 인가 = 공유 단계 판정(perm.go), 쓰기 쿼리의 owner 인자 = 대상 노드 소유자(SB4(노드 쓰기 쿼리의 owner 인자)). DTO = nodeDto.go

func (router *supervisorRouter) mountNodes(e *echo.Echo) {
	g := e.Group("/nodes")
	g.POST("", router.createNode)
	g.GET("", router.listChildren)
	g.POST("/suggest", router.suggestNodes) // nodeSuggest.go
	g.GET("/:id", router.getNode)
	g.GET("/:id/runs", router.listNodeRuns)
	g.PATCH("/:id", router.patchNode)
	g.DELETE("/:id", router.deleteNode)
	g.POST("/:id/copy", router.copyNode)
}

// paramID: 경로 :id(실패 = 400)
func paramID(c echo.Context) int64 {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 id입니다"))
	}
	return id
}

// node 변경 전달 = 볼 수 있는 접속 계정의 계정 토픽 — REF-node-list-sync.md NS2(받는 계정)·NS3(판정 시점)

// publishNode: 받는 계정을 커밋 뒤 판정해서 보낸다(노드 자신 + 부모 + parents)
func (router *supervisorRouter) publishNode(kind protocol.MsgType, node superdb.Node, parents ...pgtype.Int8) {
	users := router.nodeRecipients(context.Background(), store.GetStorePool().Queries(), nodeJudgeIDs(node, parents...))
	router.publishUsers(users, kind, newNodeResponse(node))
}

// nodeRecipients: 접속 중 계정 중 ids 하나라도 보기 이상인 계정. 조회 실패 = 아무에게도 안 보냄
func (router *supervisorRouter) nodeRecipients(ctx context.Context, q *superdb.Queries, ids []int64) []int64 {
	users := router.tabs.onlineUsers()
	if len(users) == 0 {
		return nil
	}
	vis, err := nodeVisibility(ctx, q, users, ids)
	if err != nil {
		log.Printf("[node] 받는 계정 조회 실패 nodes=%v: %v", ids, err)
		return nil
	}
	var out []int64
	for _, u := range users {
		for _, visible := range vis[u] {
			if visible {
				out = append(out, u)
				break
			}
		}
	}
	return out
}

// nodeJudgeIDs: 노드 자신(공유 루트면 부모를 못 보는 받은 쪽이 있음) + 부모 + 이동 전 부모
func nodeJudgeIDs(node superdb.Node, parents ...pgtype.Int8) []int64 {
	ids := []int64{node.ID}
	for _, p := range append([]pgtype.Int8{node.ParentID}, parents...) {
		if p.Valid && !slices.Contains(ids, p.Int64) {
			ids = append(ids, p.Int64)
		}
	}
	return ids
}

// createNode: 부모 편집(30) 이상, 소유 = 부모 소유자, created_by = 요청 계정 — SH8(받은 쪽이 만든 노드의 소유). 루트 = 내 것
func (router *supervisorRouter) createNode(c echo.Context) error {
	sess := router.requireSession(c)
	body := bindCreateRequest(c)

	q := TxQueries(c)
	ctx := c.Request().Context()
	owner, perm := sess.Data.ID, superdb.PermOwner
	if body.ParentID != nil {
		var parentNode superdb.Node
		parentNode, perm = requireFolder(c, sess.Data.ID, *body.ParentID, superdb.PermEdit)
		owner = parentNode.OwnerUserID
	}
	parent := int8Ptr(body.ParentID)

	// ord 없으면 형제 끝
	var ord int64
	if body.Ord != nil {
		ord = *body.Ord
	} else {
		mx, err := q.MaxChildOrd(ctx, superdb.MaxChildOrdParams{OwnerUserID: owner, ParentID: parent})
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		ord = mx + 1
	}

	var px, py pgtype.Numeric
	if body.Position != nil {
		px, py = numFromFloat(body.Position.X), numFromFloat(body.Position.Y)
	}

	if body.DeviceKey != nil && *body.DeviceKey != "" {
		requireDeviceOwner(ctx, q, *body.DeviceKey, owner)
	}

	var node superdb.Node
	var err error
	switch body.Kind {
	case "FOLDER":
		node, err = q.CreateFolder(ctx, superdb.CreateFolderParams{
			OwnerUserID: owner, ParentID: parent, Name: body.Name, Ord: ord,
			PositionX: px, PositionY: py, DeviceKey: textPtr(body.DeviceKey), CreatedBy: sess.Data.ID,
		})
	case "SCRIPT":
		node, err = q.CreateScript(ctx, superdb.CreateScriptParams{
			OwnerUserID: owner, ParentID: parent, Name: body.Name, Ord: ord,
			PositionX: px, PositionY: py, Content: textPtr(body.Content), CreatedBy: sess.Data.ID,
		})
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	AfterCommit(c, func() { router.publishNode(protocol.MsgNodeCreate, node) })
	// 새 노드엔 공유 행이 없어 단계 = 부모 단계
	return c.JSON(200, withNodePermission(newNodeResponse(node), perm))
}

// listChildren: 부모 보기(10) 이상, 차단 자식 제외. parentId 없음 = 내 루트 + 공유받은 루트
func (router *supervisorRouter) listChildren(c echo.Context) error {
	sess := router.requireSession(c)
	q, ctx := TxQueries(c), c.Request().Context()

	s := c.QueryParam("parentId")
	if s == "" {
		nodes, err := q.ListChildren(ctx, superdb.ListChildrenParams{OwnerUserID: sess.Data.ID})
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		shared, err := q.ListSharedRoots(ctx, sess.Data.ID)
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		out := append(newOwnedNodeResponses(nodes), newSharedRootResponses(shared)...) // 내 루트 뒤에 공유받은 루트
		fillShareCounts(c, out)
		return c.JSON(200, out)
	}

	parentID, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 parentId입니다"))
	}
	perm := requireNodePerm(c, sess.Data.ID, parentID, superdb.PermView)
	rows, err := q.ListChildrenVisible(ctx, superdb.ListChildrenVisibleParams{
		ParentPerm: perm,
		UserID:     sess.Data.ID,
		ParentID:   parentID,
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := newVisibleNodeResponses(rows)
	fillShareCounts(c, out)
	return c.JSON(200, out)
}

// getNode: 보기(10) 이상(볼 수 없음·없음 = 404)
func (router *supervisorRouter) getNode(c echo.Context) error {
	sess := router.requireSession(c)
	node, perm := requireNode(c, sess.Data.ID, paramID(c), superdb.PermView)
	res := withNodePermission(newNodeResponse(node), perm)
	// 소유자의 부모는 항상 내 것 / 아니면 부모 단계로 판정
	visible := node.ParentID.Valid && (perm == superdb.PermOwner || mustNodePerm(c, sess.Data.ID, node.ParentID.Int64) > 0)
	res.ParentVisible = &visible
	if node.OwnerUserID != sess.Data.ID { // 공유받은 노드면 소유자 닉네임 — 공유-e-1
		owner, err := TxQueries(c).GetUser(c.Request().Context(), node.OwnerUserID)
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		res.OwnerNickname = &owner.Nickname
	}
	return c.JSON(200, res)
}

// patchNode: 보낸 필드만 수정, 편집(30) 이상, 이동은 checkMove. device_key = FOLDER만·content = SCRIPT만(어기면 CHECK 위반 500)
func (router *supervisorRouter) patchNode(c echo.Context) error {
	sess := router.requireSession(c)
	id := paramID(c)
	body := bindPatchRequest(c)
	q, ctx := TxQueries(c), c.Request().Context()

	target, perm := requireNode(c, sess.Data.ID, id, superdb.PermEdit)
	oldParent := target.ParentID // 이동이면 옛 부모를 보는 계정에도 보낸다(그 목록에서 빠져야 함)
	moving := body.ParentID.Set
	if moving {
		checkMove(c, sess.Data.ID, target, perm, body.ParentID.Value)
	}
	// 부모가 실제로 바뀐 이동만 공유 재계산 대상(순서만 바꾼 건 단계 불변)
	reparent := moving && !sameParent(oldParent, body.ParentID.Value)
	var users []int64
	var uids []string
	var before map[accessKey]int16
	if reparent {
		users, uids = collectAround(c, id)
		if dest := body.ParentID.Value; dest != nil { // 이동해 들어갈 쪽 계정도 이동 전 단계가 있어야 "새로 보게 됨"을 가린다
			users = mergeUsers(c, users, *dest)
		}
		before = permIndex(c, users, uids)
	}

	if body.DeviceKey.Set && body.DeviceKey.Value != nil && *body.DeviceKey.Value != "" {
		requireDeviceOwner(ctx, q, *body.DeviceKey.Value, target.OwnerUserID)
	}

	node, err := q.PatchNode(ctx, body.toParams(id, target.OwnerUserID))
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if moving { // 위치가 바뀌면 가까운 공유도 바뀜
		if perm, err = q.NodePermission(ctx, superdb.NodePermissionParams{UserID: sess.Data.ID, NodeID: id}); err != nil {
			panic(web.Err(500, "%v", err))
		}
	}

	AfterCommit(c, func() { router.publishNode(protocol.MsgNodeUpdate, node, oldParent) })
	if reparent {
		users = mergeUsers(c, users, id)
		added := router.attachNewlyVisible(c, before, users, uids)
		AfterCommit(c, func() {
			router.reconcileAccess(users, uids)
			router.publishSaved(added)
		})
	}
	return c.JSON(200, withNodePermission(newNodeResponse(node), perm))
}

// checkMove: 아래 편집 미만 공유 없음 + 목적지 편집 + 같은 소유자 + 루트로는 소유자만 + SB6(이동 사이클 검사). 같은 부모 안 순서 변경은 통과
func checkMove(c echo.Context, userID int64, target superdb.Node, perm int16, dest *int64) {
	if dest != nil && target.ParentID.Valid && target.ParentID.Int64 == *dest {
		return
	}
	if perm < superdb.PermOwner {
		requireNoLowerShare(c, userID, target.ID)
	}
	if dest == nil {
		if perm < superdb.PermOwner {
			panic(web.Err(403, "루트로는 소유자만 옮길 수 있습니다"))
		}
		return
	}
	parent, _ := requireFolder(c, userID, *dest, superdb.PermEdit)
	if parent.OwnerUserID != target.OwnerUserID {
		panic(web.Err(403, "소유자가 다른 폴더로 옮길 수 없습니다"))
	}
	cycle, err := TxQueries(c).IsDescendant(c.Request().Context(), superdb.IsDescendantParams{AncestorID: target.ID, NodeID: *dest})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if cycle {
		panic(web.Err(400, "자기 자신이나 그 아래로 옮길 수 없습니다"))
	}
}

// requireNoLowerShare: 지우거나 옮기는 노드 아래에 그 계정의 편집 미만 공유가 있으면 403 — 차단·보기 하위가 같이 지워지거나 옮겨지지 않게(SH8)
func requireNoLowerShare(c echo.Context, userID, nodeID int64) {
	found, err := TxQueries(c).HasLowerShareBelow(c.Request().Context(), superdb.HasLowerShareBelowParams{
		UserID: userID, MinPerm: superdb.PermEdit, NodeID: nodeID,
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if found {
		panic(web.Err(403, "아래에 편집 권한이 없는 노드가 있습니다"))
	}
}

// deleteNode: 편집(30) 이상, FOLDER면 하위도 CASCADE. 발행 payload = 삭제 직전 스냅샷이라 미리 조회
func (router *supervisorRouter) deleteNode(c echo.Context) error {
	sess := router.requireSession(c)
	id := paramID(c)

	victim, perm := requireNode(c, sess.Data.ID, id, superdb.PermEdit)
	if perm < superdb.PermOwner {
		requireNoLowerShare(c, sess.Data.ID, id)
	}
	// 둘 다 삭제 전에만 구할 수 있음 — NS3(판정 시점)
	users, uids := collectAround(c, id)
	recipients := router.nodeRecipients(c.Request().Context(), TxQueries(c), nodeJudgeIDs(victim)) // 커밋 뒤엔 노드가 없어 판정 불가
	if err := TxQueries(c).DeleteNode(c.Request().Context(), superdb.DeleteNodeParams{
		ID:          id,
		OwnerUserID: victim.OwnerUserID,
	}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	AfterCommit(c, func() { router.publishUsers(recipients, protocol.MsgNodeDelete, newNodeResponse(victim)) })
	AfterCommit(c, func() { router.reconcileAccess(users, uids) })
	return c.JSON(200, map[string]any{})
}

// sameParent: 목적지(dest, nil = 루트)가 지금 부모와 같은지
func sameParent(cur pgtype.Int8, dest *int64) bool {
	if dest == nil {
		return !cur.Valid
	}
	return cur.Valid && cur.Int64 == *dest
}

// collectAround: 노드 주변 공유 계정 + 하위 살아 있는 process uid(요청 트랜잭션 안)
func collectAround(c echo.Context, id int64) ([]int64, []string) {
	q, ctx := TxQueries(c), c.Request().Context()
	users, err := q.ListShareUsersAround(ctx, id)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	uids, err := q.ListLiveProcessesUnder(ctx, id)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return users, uids
}

// mergeUsers: 이동 뒤 공유 계정을 이동 전 목록에 합친다
func mergeUsers(c echo.Context, before []int64, id int64) []int64 {
	after, err := TxQueries(c).ListShareUsersAround(c.Request().Context(), id)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return uniq(append(before, after...))
}
