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
// TxQueries(c)(txMiddleware가 깐 트랜잭션). owner_user_id = 세션 user(requireSession).
// DTO·pgtype 변환은 nodeDto.go, PATCH {valid,value} 래퍼는 internal/patch.

func (router *supervisorRouter) mountNodes(e *echo.Echo) {
	g := e.Group("/nodes")
	g.POST("", router.createNode)
	g.GET("", router.listChildren)
	g.GET("/:id", router.getNode)
	g.PATCH("/:id", router.patchNode)
	g.DELETE("/:id", router.deleteNode)
}

// paramID는 경로 :id를 int64로 파싱한다(실패=400).
func paramID(c echo.Context) int64 {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 id입니다"))
	}
	return id
}

// parentOrRoot는 pgtype.Int8 parentId를 토픽 키로 정규화한다(NULL=루트=0, subscribe.go의
// "NODE:0 고정 구독" 규약과 동일).
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
// (핸들러 책임: parentId가 같은 owner인지는 별도 검증 — 현재는 미적용.)
func (router *supervisorRouter) createNode(c echo.Context) error {
	sess := router.requireSession(c)
	body := bindCreateRequest(c)

	q := TxQueries(c)
	ctx := c.Request().Context()
	owner := sess.Data.ID
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
			PositionX: px, PositionY: py, DeviceKey: textPtr(body.DeviceKey),
		})
	case "SCRIPT":
		node, err = q.CreateScript(ctx, superdb.CreateScriptParams{
			OwnerUserID: owner, ParentID: parent, Name: body.Name, Ord: ord,
			PositionX: px, PositionY: py, Content: textPtr(body.Content),
		})
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	AfterCommit(c, func() { router.publishNode(protocol.MsgNodeCreate, node) })
	return c.JSON(200, newNodeResponse(node))
}

// listChildren은 한 부모의 자식(쿼리 parentId 없으면 루트)을 그리드 순서로 반환한다.
func (router *supervisorRouter) listChildren(c echo.Context) error {
	sess := router.requireSession(c)

	var parent pgtype.Int8
	if s := c.QueryParam("parentId"); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic(web.Err(400, "잘못된 parentId입니다"))
		}
		parent = pgtype.Int8{Int64: v, Valid: true}
	}

	nodes, err := TxQueries(c).ListChildren(c.Request().Context(), superdb.ListChildrenParams{
		OwnerUserID: sess.Data.ID,
		ParentID:    parent,
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return c.JSON(200, newNodeResponses(nodes))
}

// getNode는 노드 1건을 반환한다(없음/타인 소유=404).
func (router *supervisorRouter) getNode(c echo.Context) error {
	sess := router.requireSession(c)
	node, err := TxQueries(c).GetNode(c.Request().Context(), superdb.GetNodeParams{
		ID:          paramID(c),
		OwnerUserID: sess.Data.ID,
	})
	if err != nil {
		panic(web.Err(404, "노드를 찾을 수 없습니다"))
	}
	return c.JSON(200, newNodeResponse(node))
}

// patchNode는 보낸 필드만 부분수정한다(patch.Field → set_* 플래그/pgtype).
// (DB 보호: device_key=FOLDER만/content=SCRIPT만 → 잘못된 kind면 CHECK 위반 500.
//
//	parentId를 자기 자손으로 옮기는 사이클 방지는 핸들러 책임 — 현재 미적용.)
func (router *supervisorRouter) patchNode(c echo.Context) error {
	sess := router.requireSession(c)
	id := paramID(c)
	body := bindPatchRequest(c)
	ctx := c.Request().Context()

	// parentId가 바뀌는(=이동) 요청이면 발행 시 old parent 토픽에도 알려야 하므로 미리 조회.
	var oldParent pgtype.Int8
	moving := body.ParentID.Set
	if moving {
		old, err := TxQueries(c).GetNode(ctx, superdb.GetNodeParams{ID: id, OwnerUserID: sess.Data.ID})
		if err != nil {
			panic(web.Err(404, "노드를 찾을 수 없습니다"))
		}
		oldParent = old.ParentID
	}

	node, err := TxQueries(c).PatchNode(ctx, body.toParams(id, sess.Data.ID))
	if err != nil {
		panic(web.Err(500, "%v", err))
	}

	AfterCommit(c, func() { router.publishNode(protocol.MsgNodeUpdate, node) })
	if moving {
		oldTopic := nodeSubscribeTopic(parentOrRoot(oldParent))
		newTopic := nodeSubscribeTopic(parentOrRoot(node.ParentID))
		if oldTopic != newTopic {
			AfterCommit(c, func() { router.publishNodeTopic(oldTopic, protocol.MsgNodeUpdate, node) })
		}
	}
	return c.JSON(200, newNodeResponse(node))
}

// deleteNode는 노드를 삭제한다. FOLDER면 하위 subtree가 FK ON DELETE CASCADE로 동반 삭제.
// 발행 payload는 "삭제 직전 스냅샷"이라(REF-realtime.md) 삭제 전에 미리 조회해 둔다.
func (router *supervisorRouter) deleteNode(c echo.Context) error {
	sess := router.requireSession(c)
	id := paramID(c)
	ctx := c.Request().Context()

	victim, err := TxQueries(c).GetNode(ctx, superdb.GetNodeParams{ID: id, OwnerUserID: sess.Data.ID})
	if err != nil {
		panic(web.Err(404, "노드를 찾을 수 없습니다"))
	}
	if err := TxQueries(c).DeleteNode(ctx, superdb.DeleteNodeParams{
		ID:          id,
		OwnerUserID: sess.Data.ID,
	}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	AfterCommit(c, func() { router.publishNode(protocol.MsgNodeDelete, victim) })
	return c.JSON(200, map[string]any{})
}
