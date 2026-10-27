package router

import (
	"context"
	"errors"
	"log"
	"slices"
	"strconv"
	"sync"
	"time"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/transport"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// 관리자 페이지 — 모든 회원의 장비·키·라벨 + 회원 상한·권한·정지·삭제. 설계 → REF-member.md 회원-5(관리자 화면)·회원-Q15~Q21
func (r *supervisorRouter) mountAdmin(e *echo.Group) {
	g := e.Group("/admin")
	g.GET("/users", r.adminListUsers)
	g.GET("/users/:id", r.adminGetUser)
	g.PATCH("/users/:id", r.adminPatchUser)
	g.POST("/users/:id/suspend", r.adminSuspendUser)
	g.DELETE("/users/:id", r.adminDeleteUser)
	g.GET("/devices", r.adminListDevices)
	g.GET("/devices/:mainKey", r.adminGetDevice)
	g.POST("/devices", r.adminCreateDevice)
	g.PATCH("/devices/:mainKey", r.adminChangeOwner)
	g.DELETE("/devices/:mainKey", r.adminDeleteDevice)
	g.GET("/devices/:mainKey/key", r.adminShowKey)
	g.POST("/devices/:mainKey/key", r.adminReissueKey)
	g.PATCH("/devices/:mainKey/instances/:subKey", r.adminLabelInstance)
	g.GET("/devices/:mainKey/instances/:subKey/addrs", r.adminInstanceAddrs) // deviceMove.go
	g.POST("/devices/move", r.adminMoveDevices)                              // deviceMove.go
}

// ===== 정지·삭제된 회원 (메모리) =====

// blockedUsers: requireSession이 매 요청 DB를 읽지 않게 정지·삭제된 회원 id를 들고 있다. 시작 때 DB에서 채운다
type blockedUsers struct {
	mu  sync.RWMutex
	ids map[int64]bool
}

func loadBlockedUsers() *blockedUsers {
	b := &blockedUsers{ids: map[int64]bool{}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ids, err := store.GetStorePool().Queries().ListSuspendedUserIDs(ctx)
	if err != nil {
		log.Printf("[admin] 정지 회원 목록 조회 실패: %v", err)
	}
	for _, id := range ids {
		b.ids[id] = true
	}
	return b
}

func (b *blockedUsers) has(id int64) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.ids[id]
}

func (b *blockedUsers) set(id int64, blocked bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if blocked {
		b.ids[id] = true
	} else {
		delete(b.ids, id)
	}
}

// cutOff: 그 회원의 브라우저 소켓과 그 회원 장비의 worker 연결을 끊는다(정지·삭제 커밋 뒤). worker process는 PENDING으로 남는다
func (r *supervisorRouter) cutOff(userID int64, mainKeys []string) {
	for _, conn := range r.tabs.connsOfUser(userID) {
		conn.Close(errors.New("계정 정지·삭제"))
	}
	for _, f := range r.workers.FindAll(func(key string, _ *transport.Conn) bool {
		main, _ := splitInstanceKey(key)
		return slices.Contains(mainKeys, main)
	}) {
		f.Val.Close(errors.New("장비 주인 계정 정지·삭제"))
	}
}

// ===== 회원 =====

type adminUserResponse struct {
	ID          int64    `json:"id"`
	Nickname    string   `json:"nickname"`
	Perms       []string `json:"perms"`
	WorkerLimit int32    `json:"workerLimit"`
	DeviceCount int32    `json:"deviceCount"`
	CreatedAt   int64    `json:"createdAt"`
	SuspendedAt *int64   `json:"suspendedAt"`
}

type adminUserPatchRequest struct {
	WorkerLimit *int32 `json:"workerLimit" validate:"omitempty,min=0,max=1000"`
	CanRegister *bool  `json:"canRegister"`
	Admin       *bool  `json:"admin"`
}

type adminSuspendRequest struct {
	Suspended bool `json:"suspended"`
}

// requireAdmin: 관리자 아니면 403. 세션 id를 돌려준다
func (r *supervisorRouter) requireAdmin(c echo.Context) int64 {
	sess := r.requireSession(c)
	requirePerm(c, sess.Data.ID, permAdmin)
	return sess.Data.ID
}

func userIDParam(c echo.Context) int64 {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		panic(web.Err(400, "잘못된 회원 id입니다"))
	}
	return id
}

func requireUser(c echo.Context, id int64) superdb.User {
	u, err := TxQueries(c).GetUser(c.Request().Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "회원을 찾을 수 없습니다"))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return u
}

// allAdminUsers: 회원 전부 + 장비 수
func allAdminUsers(c echo.Context) []adminUserResponse {
	rows, err := TxQueries(c).ListUsersWithDevices(c.Request().Context())
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := make([]adminUserResponse, len(rows))
	for i, u := range rows {
		out[i] = adminUserResponse{ID: u.ID, Nickname: u.Nickname, Perms: u.Perms, WorkerLimit: u.WorkerLimit, DeviceCount: u.DeviceCount, CreatedAt: u.CreatedAt.Time.Unix()}
		if out[i].Perms == nil {
			out[i].Perms = []string{}
		}
		if u.SuspendedAt.Valid {
			t := u.SuspendedAt.Time.Unix()
			out[i].SuspendedAt = &t
		}
	}
	return out
}

// adminListUsers: ?q=닉네임&f=all|admin|suspended|noreg&page=&size=(0 = 전부) — 화-4(서버 목록)
func (r *supervisorRouter) adminListUsers(c echo.Context) error {
	r.requireAdmin(c)
	lq := parseListQuery(c)
	match := func(u adminUserResponse) bool { return containsFold(u.Nickname, lq.Q) }
	return c.JSON(200, pageList(allAdminUsers(c), lq, match, userFilters))
}

// adminGetUser: 회원 타일 하나
func (r *supervisorRouter) adminGetUser(c echo.Context) error {
	r.requireAdmin(c)
	id := userIDParam(c)
	for _, u := range allAdminUsers(c) {
		if u.ID == id {
			return c.JSON(200, u)
		}
	}
	panic(web.Err(404, "회원을 찾을 수 없습니다"))
}

// togglePerm: perms에서 name을 켜거나 끈다(다른 권한은 그대로 — worker_move 등)
func togglePerm(perms []string, name string, on bool) []string {
	out := slices.DeleteFunc(slices.Clone(perms), func(p string) bool { return p == name })
	if on {
		out = append(out, name)
	}
	return out
}

func (r *supervisorRouter) adminPatchUser(c echo.Context) error {
	me := r.requireAdmin(c)
	id := userIDParam(c)
	var body adminUserPatchRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "상한은 0~1000입니다"))
	}
	u := requireUser(c, id)
	limit, perms := u.WorkerLimit, u.Perms
	if body.WorkerLimit != nil {
		limit = *body.WorkerLimit
	}
	if body.CanRegister != nil {
		perms = togglePerm(perms, permWorkerRegister, *body.CanRegister)
	}
	if body.Admin != nil {
		if id == me && !*body.Admin {
			panic(web.Err(400, "자기 자신의 관리자 권한은 뺄 수 없습니다"))
		}
		perms = togglePerm(perms, permAdmin, *body.Admin)
	}
	if perms == nil {
		perms = []string{}
	}
	if _, err := TxQueries(c).UpdateUserMember(c.Request().Context(), superdb.UpdateUserMemberParams{ID: id, WorkerLimit: limit, Perms: perms}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	return c.JSON(200, map[string]any{})
}

func ownedMainKeys(c echo.Context, userID int64) []string {
	keys, err := TxQueries(c).ListOwnedMainKeys(c.Request().Context(), userID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return keys
}

// adminSuspendUser: 정지 = 로그인·세션·장비 worker 접속 차단, 데이터 유지(회원-Q21). 해제하면 그대로 돌아온다
func (r *supervisorRouter) adminSuspendUser(c echo.Context) error {
	me := r.requireAdmin(c)
	id := userIDParam(c)
	var body adminSuspendRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if id == me {
		panic(web.Err(400, "자기 자신은 정지할 수 없습니다"))
	}
	requireUser(c, id)
	if _, err := TxQueries(c).SetUserSuspended(c.Request().Context(), superdb.SetUserSuspendedParams{ID: id, Suspend: body.Suspended}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	keys := ownedMainKeys(c, id)
	AfterCommit(c, func() {
		r.blocked.set(id, body.Suspended)
		if body.Suspended {
			r.cutOff(id, keys)
		}
	})
	return c.JSON(200, map[string]any{})
}

// adminDeleteUser: 회원 + 그 회원의 폴더·스크립트·실행 기록·장비·공유 전부(회원-Q21). 살아 있는 process가 있으면 409.
// 순서 의존: nodes·processes가 users를 CASCADE 없이 참조하므로 process -> created_by를 노드 주인으로 -> nodes -> users 순서로 지운다
func (r *supervisorRouter) adminDeleteUser(c echo.Context) error {
	me := r.requireAdmin(c)
	id := userIDParam(c)
	if id == me {
		panic(web.Err(400, "자기 자신은 삭제할 수 없습니다"))
	}
	requireUser(c, id)
	q, ctx := TxQueries(c), c.Request().Context()
	n, err := q.CountActiveProcessesOfUser(ctx, id)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if n > 0 {
		panic(web.Err(409, "실행 중인 process가 %d개 있습니다. 먼저 종료하세요", n))
	}
	keys := ownedMainKeys(c, id)
	for _, step := range []func(context.Context, int64) error{q.DeleteProcessesOfUser, q.ReassignNodeCreatedBy, q.DeleteNodesOfUser} {
		if err := step(ctx, id); err != nil {
			panic(web.Err(500, "%v", err))
		}
	}
	if _, err := q.DeleteUser(ctx, id); err != nil {
		panic(web.Err(500, "%v", err))
	}
	AfterCommit(c, func() {
		r.blocked.set(id, true)
		r.cutOff(id, keys)
	})
	return c.JSON(200, map[string]any{})
}

// ===== 장비 =====

type adminDeviceCreateRequest struct {
	OwnerID int64  `json:"ownerId"`
	MainKey string `json:"mainKey"`
	APIKey  string `json:"apiKey"`
}

type adminOwnerRequest struct {
	OwnerID int64 `json:"ownerId"`
}

// allAdminDevices: 장비 전부(주인 닉네임 포함). ownerID > 0 이면 그 회원 것만
func (r *supervisorRouter) allAdminDevices(c echo.Context, ownerID int64) []deviceResponse {
	q, ctx := TxQueries(c), c.Request().Context()
	devices, err := q.ListAllDevices(ctx)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	rows, err := q.ListWorkerInstances(ctx, "")
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := make([]deviceResponse, 0, len(devices))
	for _, d := range devices {
		if ownerID > 0 && d.OwnerUserID != ownerID {
			continue
		}
		res := r.newDeviceResponse(superdb.Device{
			MainKey: d.MainKey, OwnerUserID: d.OwnerUserID, CreatedAt: d.CreatedAt, KeyIssuedAt: d.KeyIssuedAt,
		}, rows)
		res.OwnerNickname = d.OwnerNickname
		out = append(out, res)
	}
	return out
}

// adminListDevices: ?q=이름·주인·라벨&f=all|on|off|none&page=&size=(0 = 전부)&owner=회원 id(회원 타일의 장비 줄) — 화-4(서버 목록)
func (r *supervisorRouter) adminListDevices(c echo.Context) error {
	r.requireAdmin(c)
	lq := parseListQuery(c)
	owner, _ := strconv.ParseInt(c.QueryParam("owner"), 10, 64)
	return c.JSON(200, pageList(r.allAdminDevices(c, owner), lq, deviceMatch(lq.Q), deviceFilters))
}

// adminGetDevice: 장비 타일 하나
func (r *supervisorRouter) adminGetDevice(c echo.Context) error {
	r.requireAdmin(c)
	mainKey := pathParam(c, "mainKey")
	for _, d := range r.allAdminDevices(c, 0) {
		if d.MainKey == mainKey {
			return c.JSON(200, d)
		}
	}
	panic(web.Err(404, "장비를 찾을 수 없습니다"))
}

// requireDevice: 관리자용 — 주인 상관없이
func requireDevice(c echo.Context, mainKey string) superdb.Device {
	d, err := TxQueries(c).GetDevice(c.Request().Context(), mainKey)
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "장비를 찾을 수 없습니다"))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return d
}

// adminCreateDevice: 회원을 골라 대신 등록, 상한·등록 권한 무시(회원-Q17)
func (r *supervisorRouter) adminCreateDevice(c echo.Context) error {
	r.requireAdmin(c)
	var body adminDeviceCreateRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	res := createDevice(c, body.OwnerID, deviceCreateRequest{MainKey: body.MainKey, APIKey: body.APIKey}, false)
	AfterCommit(c, func() { r.publishDevice(res.MainKey) })
	return c.JSON(200, res)
}

// adminChangeOwner: 주인 바꾸기(회원-Q18). 옛 주인 폴더에 지정된 이 장비는 실행이 막힌다(requireDeviceOwner)
func (r *supervisorRouter) adminChangeOwner(c echo.Context) error {
	r.requireAdmin(c)
	var body adminOwnerRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	d := requireDevice(c, pathParam(c, "mainKey"))
	requireUser(c, body.OwnerID)
	if _, err := TxQueries(c).UpdateDeviceOwner(c.Request().Context(), superdb.UpdateDeviceOwnerParams{MainKey: d.MainKey, OwnerUserID: body.OwnerID}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	AfterCommit(c, func() { r.publishDevice(d.MainKey, d.OwnerUserID) }) // 옛 주인 목록에서도 빠지게
	return c.JSON(200, map[string]any{})
}

func (r *supervisorRouter) adminDeleteDevice(c echo.Context) error {
	r.requireAdmin(c)
	d := requireDevice(c, pathParam(c, "mainKey"))
	r.deleteDevice(c, d)
	return c.JSON(200, map[string]any{})
}

func (r *supervisorRouter) adminShowKey(c echo.Context) error {
	r.requireAdmin(c)
	return c.JSON(200, showDeviceKey(requireDevice(c, pathParam(c, "mainKey"))))
}

func (r *supervisorRouter) adminReissueKey(c echo.Context) error {
	r.requireAdmin(c)
	var body deviceKeyRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	d := requireDevice(c, pathParam(c, "mainKey"))
	return c.JSON(200, reissueDeviceKey(c, d.MainKey, body.APIKey))
}

func (r *supervisorRouter) adminLabelInstance(c echo.Context) error {
	r.requireAdmin(c)
	d := requireDevice(c, pathParam(c, "mainKey"))
	labelInstance(c, d.MainKey, pathParam(c, "subKey"))
	AfterCommit(c, func() { r.publishDevice(d.MainKey) })
	return c.JSON(200, map[string]any{})
}
