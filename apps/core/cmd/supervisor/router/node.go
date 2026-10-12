package router

import (
	"log"
	"strconv"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// nodes 카탈로그 핸들러. user.go와 동일 규약: echo 기본 핸들러 + panic(web.Err(...)) +
// TxQueries(c)(txMiddleware가 깐 트랜잭션). 인가 = 공유 단계 판정(perm.go), 쓰기 쿼리의 owner 인자 = 대상 노드 소유자(SB4(노드 쓰기 쿼리의 owner 인자)).
// DTO·pgtype 변환은 nodeDto.go, PATCH {valid,value} 래퍼는 internal/patch.

func (router *supervisorRouter) mountNodes(e *echo.Echo) {
	g := e.Group("/nodes")
	g.POST("", router.createNode)
	g.GET("", router.listChildren)
	g.GET("/:id", router.getNode)
	g.GET("/:id/runs", router.listNodeRuns)
	g.PATCH("/:id", router.patchNode)
	g.DELETE("/:id", router.deleteNode)
	g.POST("/:id/copy", router.copyNode)
}

// paramID는 경로 :id를 int64로 파싱한다(실패=400).
func paramID(c echo.Context) int64 {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 id입니다"))
	}
	return id
}

// parentOrRoot는 pgtype.Int8 parentId를 토픽 키로 정규화한다(NULL=루트=0).
// NODE:0은 구독자가 없어 발행이 버려진다 — SB5(NODE:0 누설 수정). 루트 토픽 설계는 열린 질문 4(NODE 동적 구독)에서.
func parentOrRoot(p pgtype.Int8) int64 {
	if p.Valid {
		return p.Int64
	}
	return 0
}

// publishNode는 node의 (현재) 부모 폴더 토픽에 kind 이벤트를 발행한다. 커밋 성공 후에만
// 실행되도록 반드시 AfterCommit(c, func(){ router.publishNode(...) })로 예약해서 부른다
// (REF-realtime.md "commit 이후 Publish" — tx 안에서 바로 부르면 롤백 시 누설).
func (router *supervisorRouter) publishNode(kind protocol.MsgType, node superdb.Node) {
	router.publishNodeTopic(nodeSubscribeTopic(parentOrRoot(node.ParentID)), kind, node)
}

// publishNodeTopic은 명시적 토픽에 발행한다 — 이동(parentId 변경)이 old parent 쪽에도
// 알려야 할 때(REF-realtime.md "이동/삭제 = old parent + new parent 2개 토픽") publishNode와
// 함께 쓴다. 마찬가지로 AfterCommit 안에서만 호출한다.
func (router *supervisorRouter) publishNodeTopic(topic string, kind protocol.MsgType, node superdb.Node) {
	if err := router.subscribeHub.Publish(topic, kind, newNodeResponse(node)); err != nil {
		log.Printf("[node] Publish 실패 topic=%s kind=%s: %v", topic, kind, err)
	}
}

// createNode는 FOLDER/SCRIPT 노드를 만든다. ord 미지정 시 형제 끝에 append.
// 부모가 있으면 부모 편집(30), 소유 = 부모 소유자, created_by = 요청 계정 — SH8(받은 쪽이 만든 노드의 소유). 루트 생성 = 내 것.
func (router *supervisorRouter) createNode(c echo.Context) error {
	sess := router.requireSession(c)
	body := bindCreateRequest(c)

	q := TxQueries(c)
	ctx := c.Request().Context()
	owner, perm := sess.Data.ID, superdb.PermOwner
	if body.ParentID != nil {
		var parentNode superdb.Node
		parentNode, perm = requireNode(c, sess.Data.ID, *body.ParentID, superdb.PermEdit)
		owner = parentNode.OwnerUserID
	}
	parent := int8Ptr(body.ParentID)

	// ord: 지정 없으면 형제 max+1로 끝에 붙인다.
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

// listChildren은 한 부모의 자식을 그리드 순서로 반환한다(부모 보기(10), 차단 자식 제외).
// parentId 없음(루트) = 내 루트만 — 공유받은 루트는 공유-d(실시간).
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

// getNode는 노드 1건을 반환한다(보기(10) 이상, 볼 수 없음/없음=404).
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

// patchNode는 보낸 필드만 부분수정한다(patch.Field → set_* 플래그/pgtype). 편집(30) 이상, 이동은 checkMove.
// (DB 보호: device_key=FOLDER만/content=SCRIPT만 → 잘못된 kind면 CHECK 위반 500.)
func (router *supervisorRouter) patchNode(c echo.Context) error {
	sess := router.requireSession(c)
	id := paramID(c)
	body := bindPatchRequest(c)
	q, ctx := TxQueries(c), c.Request().Context()

	target, perm := requireNode(c, sess.Data.ID, id, superdb.PermEdit)
	// parentId가 바뀌는(=이동) 요청이면 발행 시 old parent 토픽에도 알려야 하므로 미리 잡아 둔다.
	oldParent := target.ParentID
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

	node, err := q.PatchNode(ctx, body.toParams(id, target.OwnerUserID))
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if moving { // 위치가 바뀌면 가까운 공유도 바뀜
		if perm, err = q.NodePermission(ctx, superdb.NodePermissionParams{UserID: sess.Data.ID, NodeID: id}); err != nil {
			panic(web.Err(500, "%v", err))
		}
	}

	AfterCommit(c, func() { router.publishNode(protocol.MsgNodeUpdate, node) })
	if reparent {
		users = mergeUsers(c, users, id)
		added := attachNewlyVisible(c, before, users, uids)
		AfterCommit(c, func() {
			router.reconcileAccess(users, uids)
			router.publishSaved(added)
		})
	}
	if moving {
		oldTopic := nodeSubscribeTopic(parentOrRoot(oldParent))
		newTopic := nodeSubscribeTopic(parentOrRoot(node.ParentID))
		if oldTopic != newTopic {
			AfterCommit(c, func() { router.publishNodeTopic(oldTopic, protocol.MsgNodeUpdate, node) })
		}
	}
	return c.JSON(200, withNodePermission(newNodeResponse(node), perm))
}

// checkMove는 이동 검사다: 대상 편집(호출자가 확인) + 대상 아래 편집 미만 공유 없음 + 목적지 편집 + 목적지 소유자 = 대상 소유자
// + 루트로는 소유자만 + 사이클 금지(SB6(이동 사이클 검사)). 같은 부모 안의 순서 변경은 이동이 아니라 통과.
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
	parent, _ := requireNode(c, userID, *dest, superdb.PermEdit)
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

// requireNoLowerShare는 받은 쪽이 지우거나 옮기는 노드 아래에 그 계정의 편집 미만 공유 행이 있으면 403이다
// — 차단·보기 하위가 ON DELETE CASCADE로 같이 지워지거나 같이 옮겨지는 것 방지(SH8)
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

// deleteNode는 노드를 삭제한다(편집(30) 이상). FOLDER면 하위 subtree가 FK ON DELETE CASCADE로 동반 삭제.
// 발행 payload는 "삭제 직전 스냅샷"이라(REF-realtime.md) 삭제 전에 미리 조회해 둔다.
func (router *supervisorRouter) deleteNode(c echo.Context) error {
	sess := router.requireSession(c)
	id := paramID(c)

	victim, perm := requireNode(c, sess.Data.ID, id, superdb.PermEdit)
	if perm < superdb.PermOwner {
		requireNoLowerShare(c, sess.Data.ID, id)
	}
	users, uids := collectAround(c, id) // 삭제 전에만 구할 수 있음
	if err := TxQueries(c).DeleteNode(c.Request().Context(), superdb.DeleteNodeParams{
		ID:          id,
		OwnerUserID: victim.OwnerUserID,
	}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	AfterCommit(c, func() { router.publishNode(protocol.MsgNodeDelete, victim) })
	AfterCommit(c, func() { router.reconcileAccess(users, uids) })
	return c.JSON(200, map[string]any{})
}

// sameParent는 이동 요청의 목적지(dest, nil=루트)가 지금 부모와 같은지 본다.
func sameParent(cur pgtype.Int8, dest *int64) bool {
	if dest == nil {
		return !cur.Valid
	}
	return cur.Valid && cur.Int64 == *dest
}

// collectAround는 노드 주변 공유 계정과 하위 살아 있는 process uid를 모은다(요청 트랜잭션 안).
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

// mergeUsers는 이동 뒤(조상이 바뀐 상태)의 공유 계정을 이동 전 목록에 합친다.
func mergeUsers(c echo.Context, before []int64, id int64) []int64 {
	after, err := TxQueries(c).ListShareUsersAround(c.Request().Context(), id)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return uniq(append(before, after...))
}
