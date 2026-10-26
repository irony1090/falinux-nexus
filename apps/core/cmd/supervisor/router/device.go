package router

import (
	"context"
	"errors"
	"slices"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/transport"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// 내 장비 페이지 — 장비 이름(메인키) 등록·API 키·인스턴스 라벨. 설계 → REF-member.md 회원-3(워커 등록·키 발급)
func (r *supervisorRouter) mountDevices(e *echo.Group) {
	g := e.Group("/devices")
	g.GET("", r.listMyDevices)
	g.GET("/:mainKey", r.getMyDevice)
	g.POST("", r.createMyDevice)
	g.GET("/:mainKey/key", r.showMyDeviceKey)
	g.POST("/:mainKey/key", r.reissueMyDeviceKey)
	g.DELETE("/:mainKey", r.deleteMyDevice)
	g.PATCH("/:mainKey/instances/:subKey", r.labelMyInstance)
	g.GET("/:mainKey/instances/:subKey/addrs", r.myInstanceAddrs) // deviceMove.go
	g.POST("/move", r.moveMyDevices)                              // deviceMove.go
}

type deviceResponse struct {
	MainKey       string                   `json:"mainKey"`
	OwnerID       int64                    `json:"ownerId"`
	OwnerNickname string                   `json:"ownerNickname,omitempty"` // 관리자 목록에서만
	CreatedAt     int64                    `json:"createdAt"`
	KeyIssuedAt   int64                    `json:"keyIssuedAt"`
	Instances     []workerInstanceResponse `json:"instances"`
}

// apiKey = 빈 값이면 새로 발급, 있으면 붙여넣기(다른 supervisor에서 받은 키, 회원-Q14)
type deviceCreateRequest struct {
	MainKey string `json:"mainKey"`
	APIKey  string `json:"apiKey"`
}

type deviceKeyRequest struct {
	APIKey string `json:"apiKey"`
}

type deviceKeyResponse struct {
	MainKey string `json:"mainKey"`
	APIKey  string `json:"apiKey"`
}

type instanceLabelRequest struct {
	Label string `json:"label" validate:"max=64"`
}

func (r *supervisorRouter) newDeviceResponse(d superdb.Device, rows []superdb.WorkerInstance) deviceResponse {
	out := deviceResponse{
		MainKey:     d.MainKey,
		OwnerID:     d.OwnerUserID,
		CreatedAt:   d.CreatedAt.Time.Unix(),
		KeyIssuedAt: d.KeyIssuedAt.Time.Unix(),
		Instances:   []workerInstanceResponse{},
	}
	for _, row := range rows {
		if row.MainKey == d.MainKey {
			out.Instances = append(out.Instances, r.newWorkerInstanceResponse(row))
		}
	}
	return out
}

// pickAPIKey: 붙여넣은 키면 형식 검사, 비었으면 새로 발급
func pickAPIKey(pasted string) string {
	if pasted == "" {
		return newAPIKey()
	}
	if !validAPIKey(pasted) {
		panic(web.Err(400, "API 키 형식이 아닙니다(%s로 시작하는 발급 키만 등록할 수 있습니다)", apiKeyPrefix))
	}
	return pasted
}

// requireMyDevice: 내 장비만(없거나 남의 것 = 404, 존재 여부 노출 방지)
func requireMyDevice(c echo.Context, userID int64, mainKey string) superdb.Device {
	d, err := TxQueries(c).GetDevice(c.Request().Context(), mainKey)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && d.OwnerUserID != userID) {
		panic(web.Err(404, "장비를 찾을 수 없습니다"))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return d
}

func (r *supervisorRouter) myDevices(c echo.Context, userID int64) []deviceResponse {
	q := TxQueries(c)
	ctx := c.Request().Context()
	devices, err := q.ListDevicesByOwner(ctx, userID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	rows, err := q.ListOwnedWorkerInstances(ctx, userID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := make([]deviceResponse, 0, len(devices))
	for _, d := range devices {
		out = append(out, r.newDeviceResponse(d, rows))
	}
	return out
}

// listMyDevices: ?q=이름·라벨&f=all|on|off|none&page=&size=(0 = 전부) — 화-4(서버 목록)
func (r *supervisorRouter) listMyDevices(c echo.Context) error {
	sess := r.requireSession(c)
	lq := parseListQuery(c)
	return c.JSON(200, pageList(r.myDevices(c, sess.Data.ID), lq, deviceMatch(lq.Q), deviceFilters))
}

// getMyDevice: 장비 타일 하나(남의 것 = 404)
func (r *supervisorRouter) getMyDevice(c echo.Context) error {
	sess := r.requireSession(c)
	d := requireMyDevice(c, sess.Data.ID, pathParam(c, "mainKey"))
	for _, res := range r.myDevices(c, sess.Data.ID) {
		if res.MainKey == d.MainKey {
			return c.JSON(200, res)
		}
	}
	panic(web.Err(404, "장비를 찾을 수 없습니다"))
}

// createDevice: 등록 공통(본인·관리자). checkLimit = false면 상한 무시(관리자, 회원-Q17)
func createDevice(c echo.Context, ownerID int64, body deviceCreateRequest, checkLimit bool) deviceKeyResponse {
	if !mainKeyPattern.MatchString(body.MainKey) {
		panic(web.Err(400, "장비 이름은 영문·숫자·. _ - 로 64자까지입니다"))
	}
	key := pickAPIKey(body.APIKey)
	q := TxQueries(c)
	ctx := c.Request().Context()

	owner, err := q.GetUser(ctx, ownerID)
	if err != nil {
		panic(web.Err(404, "회원을 찾을 수 없습니다"))
	}
	if checkLimit {
		if !slices.Contains(owner.Perms, permWorkerRegister) {
			panic(web.Err(403, "장비 등록 권한이 없습니다"))
		}
		n, err := q.CountDevicesByOwner(ctx, ownerID)
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		if n >= int64(owner.WorkerLimit) {
			panic(web.Err(409, "등록할 수 있는 장비는 %d개까지입니다", owner.WorkerLimit))
		}
	}
	if _, err := q.GetDevice(ctx, body.MainKey); err == nil {
		panic(web.Err(409, "이미 사용 중인 장비 이름입니다"))
	}
	if _, err := q.CreateDevice(ctx, superdb.CreateDeviceParams{
		MainKey:     body.MainKey,
		OwnerUserID: ownerID,
		ApiKeyHash:  apiKeyHash(key),
		ApiKeyEnc:   sealAPIKey(key),
	}); err != nil {
		panic(web.Err(409, "이미 사용 중인 장비 이름입니다")) // 동시 등록 = PK 충돌
	}
	return deviceKeyResponse{MainKey: body.MainKey, APIKey: key}
}

func (r *supervisorRouter) createMyDevice(c echo.Context) error {
	sess := r.requireSession(c)
	var body deviceCreateRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	return c.JSON(200, createDevice(c, sess.Data.ID, body, true))
}

func showDeviceKey(d superdb.Device) deviceKeyResponse {
	key, err := openAPIKey(d.ApiKeyEnc)
	if err != nil {
		panic(web.Err(500, "API 키를 풀 수 없습니다(KEY_SECRET이 바뀌었으면 재발급하세요)"))
	}
	return deviceKeyResponse{MainKey: d.MainKey, APIKey: key}
}

func (r *supervisorRouter) showMyDeviceKey(c echo.Context) error {
	sess := r.requireSession(c)
	return c.JSON(200, showDeviceKey(requireMyDevice(c, sess.Data.ID, pathParam(c, "mainKey"))))
}

// reissueDeviceKey: 재발급·붙여넣기. 이미 붙은 연결은 그대로, 다음 접속부터 새 키
func reissueDeviceKey(c echo.Context, mainKey string, pasted string) deviceKeyResponse {
	key := pickAPIKey(pasted)
	if _, err := TxQueries(c).UpdateDeviceKey(c.Request().Context(), superdb.UpdateDeviceKeyParams{
		MainKey:    mainKey,
		ApiKeyHash: apiKeyHash(key),
		ApiKeyEnc:  sealAPIKey(key),
	}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	return deviceKeyResponse{MainKey: mainKey, APIKey: key}
}

func (r *supervisorRouter) reissueMyDeviceKey(c echo.Context) error {
	sess := r.requireSession(c)
	var body deviceKeyRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	d := requireMyDevice(c, sess.Data.ID, pathParam(c, "mainKey"))
	return c.JSON(200, reissueDeviceKey(c, d.MainKey, body.APIKey))
}

// deleteDevice: 접속 중 인스턴스가 있으면 409. 명부(worker_instances)·폴더 device_key는 남긴다
func (r *supervisorRouter) deleteDevice(c echo.Context, mainKey string) {
	if online := r.workers.FindAll(func(key string, _ *transport.Conn) bool {
		main, _ := splitInstanceKey(key)
		return main == mainKey
	}); len(online) > 0 {
		panic(web.Err(409, "접속 중인 인스턴스가 있어 지울 수 없습니다"))
	}
	if _, err := TxQueries(c).DeleteDevice(c.Request().Context(), mainKey); err != nil {
		panic(web.Err(500, "%v", err))
	}
}

func (r *supervisorRouter) deleteMyDevice(c echo.Context) error {
	sess := r.requireSession(c)
	d := requireMyDevice(c, sess.Data.ID, pathParam(c, "mainKey"))
	r.deleteDevice(c, d.MainKey)
	return c.JSON(200, map[string]any{})
}

func labelInstance(c echo.Context, mainKey, subKey string) {
	var body instanceLabelRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "라벨은 64자까지입니다"))
	}
	n, err := TxQueries(c).UpdateWorkerInstanceLabel(c.Request().Context(), superdb.UpdateWorkerInstanceLabelParams{
		MainKey: mainKey, SubKey: subKey, Label: body.Label,
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if n == 0 {
		panic(web.Err(404, "명부에 없는 인스턴스입니다"))
	}
}

func (r *supervisorRouter) labelMyInstance(c echo.Context) error {
	sess := r.requireSession(c)
	d := requireMyDevice(c, sess.Data.ID, pathParam(c, "mainKey"))
	labelInstance(c, d.MainKey, pathParam(c, "subKey"))
	return c.JSON(200, map[string]any{})
}

// requireDeviceOwner: 장비 이름의 주인 == ownerID(폴더·노드 소유자) — 회원-6(실행 권한).
// 공유받은 사람은 소유자 폴더를 거쳐 실행하므로 ownerID는 항상 폴더 소유자를 넘긴다(실행 계정 아님)
func requireDeviceOwner(ctx context.Context, q *superdb.Queries, mainKey string, ownerID int64) {
	d, err := q.GetDevice(ctx, mainKey)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(500, "%v", err))
	}
	if err != nil || d.OwnerUserID != ownerID {
		panic(web.Err(403, "%s 는 이 폴더 소유자가 등록한 장비가 아닙니다", mainKey))
	}
}

// newWorkerInstanceResponse: 명부 행 + 접속 여부(레지스트리)
func (r *supervisorRouter) newWorkerInstanceResponse(row superdb.WorkerInstance) workerInstanceResponse {
	key := row.MainKey + "#" + row.SubKey
	_, online := r.workers.Get(key)
	return workerInstanceResponse{
		workerResponse: workerResponse{MainKey: row.MainKey, SubKey: row.SubKey, InstanceKey: key, OS: protocol.NormalizeOS(row.Os), Label: row.Label},
		Online:         online,
		FirstSeen:      row.FirstSeen.Time.Unix(),
		LastSeen:       row.LastSeen.Time.Unix(),
	}
}
