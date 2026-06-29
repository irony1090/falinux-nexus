package router

import (
	"strconv"

	"nexus/internal/patch"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// node 요청/응답 DTO + DTO↔pgtype 변환. 핸들러는 node.go, PATCH 래퍼는 internal/patch.
// 변환 헬퍼가 여기 남는 이유: pgtype 전용(worker는 SQLite Null* 라 공유 불가).

// point는 캔버스 절대좌표 한 쌍. position은 2D 한 개념이라 x/y를 한 필드로 묶는다.
type point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type nodeCreateRequest struct {
	Kind      string  `json:"kind" validate:"required,oneof=FOLDER SCRIPT"`
	ParentID  *int64  `json:"parentId"` // null=루트
	Name      string  `json:"name" validate:"required"`
	Ord       *int64  `json:"ord"`       // null=형제 끝에 append(핸들러 계산)
	Position  *point  `json:"position"`  // null=캔버스 미배치
	DeviceKey *string `json:"deviceKey"` // FOLDER 전용
	Content   *string `json:"content"`   // SCRIPT 전용
}

type nodePatchRequest struct {
	Name      patch.Field[string] `json:"name"`
	Ord       patch.Field[int64]  `json:"ord"`
	ParentID  patch.Field[int64]  `json:"parentId"`
	Position  patch.Field[point]  `json:"position"`
	DeviceKey patch.Field[string] `json:"deviceKey"`
	Content   patch.Field[string] `json:"content"`
}

// bindCreateRequest는 요청을 nodeCreateRequest로 바인딩·검증한다(실패=panic 400).
// 라우터가 panic-style이라 error 대신 구조체만 돌려줘 핸들러를 한 줄로 만든다.
func bindCreateRequest(c echo.Context) nodeCreateRequest {
	var body nodeCreateRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	return body
}

// bindPatchRequest는 요청을 nodePatchRequest로 바인딩한다(실패=panic 400).
// patch 필드는 전부 옵셔널이라 validate 태그가 없어, NOT NULL 컬럼 가드만 직접 건다.
func bindPatchRequest(c echo.Context) nodePatchRequest {
	var body nodePatchRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if body.Name.Set && body.Name.Value == nil {
		panic(web.Err(400, "name은 null이 될 수 없습니다"))
	}
	if body.Ord.Set && body.Ord.Value == nil {
		panic(web.Err(400, "ord는 null이 될 수 없습니다"))
	}
	return body
}

// toParams는 PATCH 요청을 PatchNodeParams로 변환한다.
//   - name/ord(NOT NULL): Set+값일 때만 채우고 아니면 skip(narg NULL → SQL COALESCE가 기존값 유지)
//   - nullable 필드: set_* 플래그 + 값(NULL 포함). 플래그로 skip/갱신 구분
func (body nodePatchRequest) toParams(id, ownerID int64) superdb.PatchNodeParams {
	p := superdb.PatchNodeParams{ID: id, OwnerUserID: ownerID}
	if body.Name.Set {
		p.Name = pgtype.Text{String: *body.Name.Value, Valid: true}
	}
	if body.Ord.Set {
		p.Ord = pgtype.Int8{Int64: *body.Ord.Value, Valid: true}
	}
	p.SetParent, p.ParentID = patchInt8(body.ParentID)
	p.SetPosition, p.PositionX, p.PositionY = patchPosition(body.Position)
	p.SetDevice, p.DeviceKey = patchText(body.DeviceKey)
	p.SetContent, p.Content = patchText(body.Content)
	return p
}

type nodeResponse struct {
	ID        int64   `json:"id"`
	ParentID  *int64  `json:"parentId"`
	Kind      string  `json:"kind"`
	Name      string  `json:"name"`
	Ord       int64   `json:"ord"`
	Position  *point  `json:"position"` // 미배치면 null
	DeviceKey *string `json:"deviceKey"`
	Content   *string `json:"content"`
	CreatedAt int64   `json:"createdAt"`
	UpdatedAt *int64  `json:"updatedAt"`
}

func newNodeResponse(n superdb.Node) nodeResponse {
	r := nodeResponse{ID: n.ID, Kind: n.Kind, Name: n.Name}
	if n.ParentID.Valid {
		v := n.ParentID.Int64
		r.ParentID = &v
	}
	r.Ord = n.Ord
	if n.PositionX.Valid && n.PositionY.Valid {
		x, _ := floatFromNum(n.PositionX)
		y, _ := floatFromNum(n.PositionY)
		r.Position = &point{X: x, Y: y}
	}
	if n.DeviceKey.Valid {
		v := n.DeviceKey.String
		r.DeviceKey = &v
	}
	if n.Content.Valid {
		v := n.Content.String
		r.Content = &v
	}
	if n.CreatedAt.Valid {
		r.CreatedAt = n.CreatedAt.Time.Unix()
	}
	if n.UpdatedAt.Valid {
		t := n.UpdatedAt.Time.Unix()
		r.UpdatedAt = &t
	}
	return r
}

func newNodeResponses(ns []superdb.Node) []nodeResponse {
	out := make([]nodeResponse, len(ns))
	for i, n := range ns {
		out[i] = newNodeResponse(n)
	}
	return out
}

// ===== DTO ↔ pgtype 변환 =====

func int8Ptr(p *int64) pgtype.Int8 {
	if p == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *p, Valid: true}
}

func textPtr(p *string) pgtype.Text {
	if p == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *p, Valid: true}
}

func numFromFloat(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.ScanScientific(strconv.FormatFloat(f, 'f', -1, 64))
	return n
}

func floatFromNum(n pgtype.Numeric) (float64, bool) {
	if !n.Valid {
		return 0, false
	}
	fv, err := n.Float64Value()
	if err != nil || !fv.Valid {
		return 0, false
	}
	return fv.Float64, true
}

// patch.Field[T] → (set_* 플래그, pgtype 값). 플래그 false면 값은 무시된다.
func patchText(p patch.Field[string]) (bool, pgtype.Text) {
	if !p.Set {
		return false, pgtype.Text{}
	}
	if p.Value == nil {
		return true, pgtype.Text{} // Valid:false = NULL
	}
	return true, pgtype.Text{String: *p.Value, Valid: true}
}

func patchInt8(p patch.Field[int64]) (bool, pgtype.Int8) {
	if !p.Set {
		return false, pgtype.Int8{}
	}
	if p.Value == nil {
		return true, pgtype.Int8{}
	}
	return true, pgtype.Int8{Int64: *p.Value, Valid: true}
}

func patchPosition(p patch.Field[point]) (bool, pgtype.Numeric, pgtype.Numeric) {
	if !p.Set {
		return false, pgtype.Numeric{}, pgtype.Numeric{}
	}
	if p.Value == nil { // 캔버스 미배치로(둘 다 NULL)
		return true, pgtype.Numeric{}, pgtype.Numeric{}
	}
	return true, numFromFloat(p.Value.X), numFromFloat(p.Value.Y)
}
