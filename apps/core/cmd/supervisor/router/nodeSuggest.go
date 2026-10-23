package router

import (
	"sort"
	"strings"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// 폴더 타일 경로 입력 — 경로-6(폴더 후보 REST), REF-path-input.md. 단계마다 보이는 자식 폴더를 읽어 이름으로 고른다
const nodeSuggestMax = 200

type nodeSuggestRequest struct {
	BaseID int64    `json:"baseId"` // 0 = 기본 목록(내 루트 + 공유받은 루트). 자동완성 목록에서 고른 단계(Q43(같은 이름 폴더))
	Names  []string `json:"names"`  // BaseID 아래로 내려갈 이름들
	Prefix string   `json:"prefix"`
	Max    int      `json:"max"` // 0 = 후보 없이 Names 끝 폴더만 찾기(Enter)
}

type nodeSuggestItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type nodeSuggestResponse struct {
	ID    *int64            `json:"id"` // Names 끝 폴더(nil = 기본 목록)
	Items []nodeSuggestItem `json:"items"`
	More  bool              `json:"more,omitempty"`
}

type nodeLevel struct {
	node superdb.Node
	perm int16
}

// POST /nodes/suggest: 같은 이름이면 목록 순서대로 첫 번째(대소문자까지 같은 것 먼저, 없으면 대소문자 무시) — Q43(같은 이름 폴더)
func (router *supervisorRouter) suggestNodes(c echo.Context) error {
	sess := router.requireSession(c)
	var body nodeSuggestRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}

	var cur *nodeLevel
	if body.BaseID != 0 {
		node, perm := requireFolder(c, sess.Data.ID, body.BaseID, superdb.PermView)
		cur = &nodeLevel{node, perm}
	}
	for _, name := range body.Names {
		kids := folderKids(c, sess.Data.ID, cur)
		next := -1
		for i, k := range kids {
			if k.node.Name == name {
				next = i
				break
			}
		}
		if next < 0 {
			for i, k := range kids {
				if strings.EqualFold(k.node.Name, name) {
					next = i
					break
				}
			}
		}
		if next < 0 {
			panic(web.Err(404, "'%s' 폴더가 없습니다", name))
		}
		cur = &kids[next]
	}

	out := nodeSuggestResponse{Items: []nodeSuggestItem{}}
	if cur != nil {
		out.ID = &cur.node.ID
	}
	max := min(body.Max, nodeSuggestMax)
	if max <= 0 {
		return c.JSON(200, out)
	}
	for _, k := range folderKids(c, sess.Data.ID, cur) {
		if strings.HasPrefix(strings.ToLower(k.node.Name), strings.ToLower(body.Prefix)) {
			out.Items = append(out.Items, nodeSuggestItem{ID: k.node.ID, Name: k.node.Name})
		}
	}
	// 대소문자까지 같은 앞글자 먼저, 그 안은 목록 순서 그대로(같은 이름의 첫 번째가 앞에 오게)
	sort.SliceStable(out.Items, func(i, j int) bool {
		return strings.HasPrefix(out.Items[i].Name, body.Prefix) && !strings.HasPrefix(out.Items[j].Name, body.Prefix)
	})
	if len(out.Items) > max {
		out.Items, out.More = out.Items[:max], true
	}
	return c.JSON(200, out)
}

// folderKids: parent nil = 기본 목록(listChildren과 같은 순서: 내 루트 뒤에 공유받은 루트). 폴더만
func folderKids(c echo.Context, userID int64, parent *nodeLevel) []nodeLevel {
	q, ctx := TxQueries(c), c.Request().Context()
	var out []nodeLevel
	if parent == nil {
		own, err := q.ListChildren(ctx, superdb.ListChildrenParams{OwnerUserID: userID, ParentID: pgtype.Int8{}})
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		for _, n := range own {
			if n.Kind == superdb.NodeKindFolder {
				out = append(out, nodeLevel{n, superdb.PermOwner})
			}
		}
		shared, err := q.ListSharedRoots(ctx, userID)
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		for _, r := range shared {
			out = append(out, nodeLevel{r.Node, r.Permission})
		}
		return out
	}
	rows, err := q.ListChildrenVisible(ctx, superdb.ListChildrenVisibleParams{ParentPerm: parent.perm, UserID: userID, ParentID: parent.node.ID})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	for _, r := range rows {
		if r.Node.Kind == superdb.NodeKindFolder {
			out = append(out, nodeLevel{r.Node, r.Permission})
		}
	}
	return out
}
