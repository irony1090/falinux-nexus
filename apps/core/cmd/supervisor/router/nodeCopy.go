package router

import (
	"context"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// 노드 복사 — SH10(노드 복사), REF-process-share-api.md "SH10(노드 복사) 구현 모양"

const copyMax = 1000 // 한 번에 복사할 수 있는 노드 수(원본 포함)

type nodeCopyRequest struct {
	ParentID *int64 `json:"parentId"` // null = 내 루트
}

type nodeCopyResponse struct {
	Nodes []nodeResponse `json:"nodes"` // 만든 최상위 노드
}

// copyItem은 복사할 원본 하위 트리(넣기 전에 전부 읽어 둠).
type copyItem struct {
	node     superdb.Node
	children []*copyItem
}

// copyNode는 노드(폴더면 하위 전체)를 목적지 아래에 복사한다. 원본 보기(10) + 목적지 편집(30), 목적지 없음 = 내 루트.
// 복사본 소유 = 목적지 소유자, created_by = 요청 계정(SH8과 같은 규칙). 차단 하위는 빼고, 뺀 개수도 알리지 않음 — SC7(복사 제외 개수)
func (router *supervisorRouter) copyNode(c echo.Context) error {
	sess := router.requireSession(c)
	src, srcPerm := requireNode(c, sess.Data.ID, paramID(c), superdb.PermView)
	var body nodeCopyRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}

	owner, perm := sess.Data.ID, superdb.PermOwner
	if body.ParentID != nil {
		var dest superdb.Node
		dest, perm = requireFolder(c, sess.Data.ID, *body.ParentID, superdb.PermEdit)
		owner = dest.OwnerUserID
	}
	q, ctx := TxQueries(c), c.Request().Context()
	parent := int8Ptr(body.ParentID)

	// 원본을 전부 읽은 뒤 넣는다 — 폴더를 자기 아래로 복사해도 새로 만든 노드를 다시 읽지 않음
	count := 0
	root := collectCopy(ctx, q, src, srcPerm, sess.Data.ID, &count)

	mx, err := q.MaxChildOrd(ctx, superdb.MaxChildOrdParams{OwnerUserID: owner, ParentID: parent})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	top := insertCopy(ctx, q, root, parent, mx+1, owner, sess.Data.ID)

	AfterCommit(c, func() { router.publishNode(protocol.MsgNodeCreate, top) })
	// 새 노드엔 공유 행이 없어 단계 = 목적지 단계
	return c.JSON(200, nodeCopyResponse{Nodes: []nodeResponse{withNodePermission(newNodeResponse(top), perm)}})
}

// collectCopy는 원본 하위를 읽는다. 자식 단계는 ListChildrenVisible이 판정(차단 자식과 그 아래는 빠짐).
func collectCopy(ctx context.Context, q *superdb.Queries, n superdb.Node, perm int16, userID int64, count *int) *copyItem {
	if *count++; *count > copyMax {
		panic(web.Err(400, "한 번에 %d개까지 복사할 수 있습니다", copyMax))
	}
	item := &copyItem{node: n}
	if n.Kind != superdb.NodeKindFolder {
		return item
	}
	rows, err := q.ListChildrenVisible(ctx, superdb.ListChildrenVisibleParams{ParentPerm: perm, UserID: userID, ParentID: n.ID})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	for _, row := range rows {
		item.children = append(item.children, collectCopy(ctx, q, row.Node, row.Permission, userID, count))
	}
	return item
}

// insertCopy는 복사본을 넣는다. 최상위는 목적지 형제 끝(ord), 하위는 원본 ord 유지. device_key는 노드 자신의 값만(상속값 안 따라감).
func insertCopy(ctx context.Context, q *superdb.Queries, item *copyItem, parent pgtype.Int8, ord, owner, createdBy int64) superdb.Node {
	n := item.node
	var made superdb.Node
	var err error
	if n.Kind == superdb.NodeKindFolder {
		made, err = q.CreateFolder(ctx, superdb.CreateFolderParams{
			OwnerUserID: owner, ParentID: parent, Name: n.Name, Ord: ord, DeviceKey: n.DeviceKey, CreatedBy: createdBy,
		})
	} else {
		made, err = q.CreateScript(ctx, superdb.CreateScriptParams{
			OwnerUserID: owner, ParentID: parent, Name: n.Name, Ord: ord, Content: n.Content, CreatedBy: createdBy,
		})
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	for _, child := range item.children {
		insertCopy(ctx, q, child, pgtype.Int8{Int64: made.ID, Valid: true}, child.node.Ord, owner, createdBy)
	}
	return made
}
