package router

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/transport"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// 접속 중인 worker 인스턴스 조회 — 메모리 레지스트리(r.workers)만 본다(roster/DB 아님).
func (r *supervisorRouter) mountWorkers(e *echo.Echo) {
	g := e.Group("/workers")
	g.GET("", r.listWorkers)
}

type workerResponse struct {
	MainKey     string `json:"mainKey"`
	SubKey      string `json:"subKey"`
	InstanceKey string `json:"instanceKey"`
}

// splitInstanceKey: "main#sub" → (main, sub). 서브키는 랜덤 영숫자라 마지막 '#'로 자른다
func splitInstanceKey(key string) (string, string) {
	i := strings.LastIndex(key, "#")
	if i < 0 {
		return key, ""
	}
	return key[:i], key[i+1:]
}

// resolveMainKey: 스크립트 node의 실행 대상 main_key(가장 가까운 폴더의 device_key 상속).
// 없거나(귀속 폴더 없음) 타인 소유 node면 404
func resolveMainKey(ctx context.Context, q *superdb.Queries, nodeID, ownerID int64) string {
	key, err := q.ResolveDeviceKey(ctx, superdb.ResolveDeviceKeyParams{ID: nodeID, OwnerUserID: ownerID})
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !key.Valid) {
		panic(web.Err(404, "실행 대상 장비(device_key)가 지정된 폴더가 없습니다"))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return key.String
}

// listWorkers: 접속 중인 인스턴스 목록. ?nodeId= 면 그 스크립트의 실행 대상 main_key 인스턴스만
func (r *supervisorRouter) listWorkers(c echo.Context) error {
	sess := r.requireSession(c)

	mainKey := ""
	if s := c.QueryParam("nodeId"); s != "" {
		nodeID, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic(web.Err(400, "잘못된 nodeId입니다"))
		}
		mainKey = resolveMainKey(c.Request().Context(), TxQueries(c), nodeID, sess.Data.ID)
	}

	found := r.workers.FindAll(func(key string, _ *transport.Conn) bool {
		main, _ := splitInstanceKey(key)
		return mainKey == "" || main == mainKey
	})
	rst := make([]workerResponse, 0, len(found))
	for _, f := range found {
		main, sub := splitInstanceKey(f.Key)
		rst = append(rst, workerResponse{MainKey: main, SubKey: sub, InstanceKey: f.Key})
	}
	sort.Slice(rst, func(i, j int) bool { return rst[i].InstanceKey < rst[j].InstanceKey })
	return c.JSON(200, rst)
}
