package router

import (
	"slices"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// 칸 순서 바꾸기 — 칸-1(node 순서 API), REF-tile-drag.md 칸Q2(저장 방식)·칸Q3(루트 목록)
type nodeOrderRequest struct {
	ParentID *int64  `json:"parentId"` // nil = 내 루트(공유받은 루트는 못 옮김)
	IDs      []int64 `json:"ids"`      // 옮길 형제(보내는 순서 무관, 지금 형제 순서를 지킨다)
	After    *int64  `json:"after"`    // 이 형제 바로 뒤로. nil = 맨 앞, 형제에 없음(그새 지워짐) = 맨 뒤
}

// POST /nodes/order: 부모 편집 이상. 안 보이는 형제까지 1부터 다시 매겨 그 형제들의 상대 위치도 지킨다
func (router *supervisorRouter) orderNodes(c echo.Context) error {
	sess := router.requireSession(c)
	var body nodeOrderRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if len(body.IDs) == 0 {
		panic(web.Err(400, "옮길 노드가 없습니다"))
	}
	q, ctx := TxQueries(c), c.Request().Context()

	owner, parent := sess.Data.ID, pgtype.Int8{}
	if body.ParentID != nil {
		p, _ := requireFolder(c, sess.Data.ID, *body.ParentID, superdb.PermEdit)
		owner, parent = p.OwnerUserID, pgtype.Int8{Int64: p.ID, Valid: true}
	}
	rows, err := q.ListSiblingOrds(ctx, superdb.ListSiblingOrdsParams{OwnerUserID: owner, ParentID: parent})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	sibs := make([]int64, len(rows))
	for i, r := range rows {
		sibs[i] = r.ID
	}
	next, bad := placeSiblings(sibs, body.IDs, body.After)
	if bad != "" {
		panic(web.Err(400, "%s", bad))
	}

	var ids, ords []int64
	for i, id := range next {
		if ord := int64(i + 1); rows[slices.Index(sibs, id)].Ord != ord {
			ids, ords = append(ids, id), append(ords, ord)
		}
	}
	if len(ids) == 0 {
		return c.JSON(200, map[string]any{})
	}
	if err := q.SetNodeOrds(ctx, superdb.SetNodeOrdsParams{Ids: ids, Ords: ords, OwnerUserID: owner}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	// 목록 다시 읽기 신호는 하나면 된다(받는 쪽은 parentId로 그 목록을 무효화)
	moved, err := q.GetNodeByID(ctx, body.IDs[0])
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	AfterCommit(c, func() { router.publishNode(protocol.MsgNodeUpdate, moved) })
	return c.JSON(200, map[string]any{})
}

// placeSiblings: 형제 순서 sibs에서 moving을 빼서 after 바로 뒤에 넣은 새 순서. moving끼리는 sibs 순서. 못 하면 이유
func placeSiblings(sibs, moving []int64, after *int64) ([]int64, string) {
	for _, id := range moving {
		if !slices.Contains(sibs, id) {
			return nil, "같은 폴더에 있는 노드만 순서를 바꿀 수 있습니다"
		}
	}
	if after != nil && slices.Contains(moving, *after) {
		return nil, "옮기는 노드 뒤로는 놓을 수 없습니다"
	}
	var rest, pick []int64
	for _, id := range sibs {
		if slices.Contains(moving, id) {
			pick = append(pick, id)
		} else {
			rest = append(rest, id)
		}
	}
	at := 0
	if after != nil {
		at = len(rest)
		if i := slices.Index(rest, *after); i >= 0 {
			at = i + 1
		}
	}
	return slices.Concat(rest[:at], pick, rest[at:]), ""
}
