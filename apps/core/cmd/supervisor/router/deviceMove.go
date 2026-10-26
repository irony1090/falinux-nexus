package router

import (
	"strings"

	"nexus/internal/protocol"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// ===== 열린 장비 타일의 supervisor 주소 일괄 변경 — 내 장비 = 장비 주인 / 관리자 = 전부. REF-device-move.md =====
// 장비 타일 메뉴의 단일 옮기기(workerMove.go)는 폴더 인가라 폴더가 없는 관리 페이지용으로 따로 둔다

type deviceMoveRequest struct {
	Targets []string `json:"targets" validate:"required,min=1,max=100"` // instanceKey(메인키#서브키)
	URL     string   `json:"url" validate:"required"`
	Now     bool     `json:"now"`
}

// moveMyDevices: POST /devices/move — 남의 장비가 하나라도 있으면 404
func (r *supervisorRouter) moveMyDevices(c echo.Context) error {
	sess := r.requireSession(c)
	return r.moveDevices(c, func(mainKey string) { requireMyDevice(c, sess.Data.ID, mainKey) })
}

// adminMoveDevices: POST /admin/devices/move
func (r *supervisorRouter) adminMoveDevices(c echo.Context) error {
	r.requireAdmin(c)
	return r.moveDevices(c, func(mainKey string) { requireDevice(c, mainKey) })
}

func (r *supervisorRouter) moveDevices(c echo.Context, check func(mainKey string)) error {
	var body deviceMoveRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	url := normalizeMoveURL(body.URL)
	checked := map[string]bool{}
	targets := make([]moveTarget, 0, len(body.Targets))
	for _, key := range body.Targets {
		mainKey, subKey, ok := strings.Cut(key, "#")
		if !ok || mainKey == "" || subKey == "" {
			panic(web.Err(400, "잘못된 instance입니다: %q", key))
		}
		if !checked[mainKey] {
			check(mainKey)
			checked[mainKey] = true
		}
		conn, _ := r.workers.Get(key)
		targets = append(targets, moveTarget{key: key, conn: conn})
	}
	EndTx(c)
	return c.JSON(200, moveConns(c.Request().Context(), targets, protocol.MoveRequest{URL: url, Now: body.Now}))
}

// myInstanceAddrs: GET /devices/:mainKey/instances/:subKey/addrs — worker에 저장된 주소 목록(1순위부터)
func (r *supervisorRouter) myInstanceAddrs(c echo.Context) error {
	sess := r.requireSession(c)
	requireMyDevice(c, sess.Data.ID, pathParam(c, "mainKey"))
	return r.instanceAddrs(c)
}

// adminInstanceAddrs: GET /admin/devices/:mainKey/instances/:subKey/addrs
func (r *supervisorRouter) adminInstanceAddrs(c echo.Context) error {
	r.requireAdmin(c)
	requireDevice(c, pathParam(c, "mainKey"))
	return r.instanceAddrs(c)
}

func (r *supervisorRouter) instanceAddrs(c echo.Context) error {
	conn, ok := r.workers.Get(pathParam(c, "mainKey") + "#" + pathParam(c, "subKey"))
	if !ok {
		panic(web.Err(409, "접속 중이 아닌 instance입니다"))
	}
	EndTx(c)
	var out protocol.AddrsResponse
	callWorker(c.Request().Context(), conn, protocol.MsgAddrs, struct{}{}, &out)
	return c.JSON(200, addrsResponse{Addrs: nonNil(out.Addrs)})
}
