package router

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/transport"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// GET /workers = 접속 중만(레지스트리) / /workers/instances = 명부(DB, online은 레지스트리로 계산) — RS2(API 모양)
func (r *supervisorRouter) mountWorkers(e *echo.Echo) {
	g := e.Group("/workers")
	g.GET("", r.listWorkers)
	g.GET("/instances", r.listWorkerInstances)
	g.DELETE("/instances/:mainKey/:subKey", r.deleteWorkerInstance)
	g.GET("/:mainKey/:subKey/dir", r.listWorkerDir)             // workerDir.go
	g.POST("/:mainKey/:subKey/dir/exists", r.workerDirExists)   // workerDir.go
	g.GET("/:mainKey/:subKey/dir/suggest", r.suggestWorkerDir)  // workerDir.go
	g.POST("/:mainKey/:subKey/uploads", r.startUpload)          // upload.go
	g.POST("/:mainKey/:subKey/downloads", r.startDownload)      // download.go
	g.POST("/:mainKey/:subKey/exec", r.execWorkerFile)          // workerExec.go
	g.GET("/:mainKey/runs", r.listDeviceRuns)                   // nodeRuns.go
	g.GET("/:mainKey/:subKey/addrs", r.listWorkerAddrs)         // workerMove.go
	g.POST("/:mainKey/:subKey/move", r.moveWorker)              // workerMove.go
	g.POST("/move-all", r.moveAllWorkers)                       // workerMove.go
	g.GET("/:mainKey/:subKey/unregistered", r.listUnregistered) // workerMove.go
	g.POST("/:mainKey/:subKey/unregistered/:uid/kill", r.killUnregistered)
}

type workerResponse struct {
	MainKey     string `json:"mainKey"`
	SubKey      string `json:"subKey"`
	InstanceKey string `json:"instanceKey"`
	OS          string `json:"os"` // linux | windows ... — WW5(OS 표시)
}

type workerInstanceResponse struct {
	workerResponse
	Online    bool  `json:"online"`
	FirstSeen int64 `json:"firstSeen"` // Unix 초
	LastSeen  int64 `json:"lastSeen"`
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
// ownerID = 노드 소유자(판정 뒤 호출). 귀속 폴더가 없으면 404
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

// listWorkers: 접속 중인 인스턴스 목록. ?nodeId= 면 그 스크립트의 실행 대상 main_key 인스턴스만(조작(20) 이상)
func (r *supervisorRouter) listWorkers(c echo.Context) error {
	sess := r.requireSession(c)

	mainKey := ""
	if s := c.QueryParam("nodeId"); s != "" {
		nodeID, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic(web.Err(400, "잘못된 nodeId입니다"))
		}
		node, _ := requireNode(c, sess.Data.ID, nodeID, superdb.PermOperate)
		mainKey = resolveMainKey(c.Request().Context(), TxQueries(c), nodeID, node.OwnerUserID)
	}

	found := r.workers.FindAll(func(key string, _ *transport.Conn) bool {
		main, _ := splitInstanceKey(key)
		return mainKey == "" || main == mainKey
	})
	rst := make([]workerResponse, 0, len(found))
	for _, f := range found {
		main, sub := splitInstanceKey(f.Key)
		os, _ := r.workerOS.Get(f.Key)
		rst = append(rst, workerResponse{MainKey: main, SubKey: sub, InstanceKey: f.Key, OS: protocol.NormalizeOS(os)})
	}
	sort.Slice(rst, func(i, j int) bool { return rst[i].InstanceKey < rst[j].InstanceKey })
	return c.JSON(200, rst)
}

// listWorkerInstances: 명부 전체(?mainKey= 면 그 장비만). 장비는 계정 소유가 아니라 로그인만 확인 — RS3(보이는 범위)
func (r *supervisorRouter) listWorkerInstances(c echo.Context) error {
	r.requireSession(c)
	rows, err := TxQueries(c).ListWorkerInstances(c.Request().Context(), c.QueryParam("mainKey"))
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	rst := make([]workerInstanceResponse, 0, len(rows))
	for _, row := range rows {
		key := row.MainKey + "#" + row.SubKey
		_, online := r.workers.Get(key)
		rst = append(rst, workerInstanceResponse{
			workerResponse: workerResponse{MainKey: row.MainKey, SubKey: row.SubKey, InstanceKey: key, OS: protocol.NormalizeOS(row.Os)},
			Online:         online,
			FirstSeen:      row.FirstSeen.Time.Unix(),
			LastSeen:       row.LastSeen.Time.Unix(),
		})
	}
	return c.JSON(200, rst)
}

// deleteWorkerInstance: 끊긴 인스턴스만 명부에서 지운다(접속 중 409) — RS4(정리)
func (r *supervisorRouter) deleteWorkerInstance(c echo.Context) error {
	r.requireSession(c)
	mainKey, subKey := pathParam(c, "mainKey"), pathParam(c, "subKey")
	if _, online := r.workers.Get(mainKey + "#" + subKey); online {
		panic(web.Err(409, "접속 중인 인스턴스는 지울 수 없습니다"))
	}
	n, err := TxQueries(c).DeleteWorkerInstance(c.Request().Context(), superdb.DeleteWorkerInstanceParams{MainKey: mainKey, SubKey: subKey})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if n == 0 {
		panic(web.Err(404, "명부에 없는 인스턴스입니다"))
	}
	return c.JSON(200, map[string]any{})
}

// pathParam: echo는 RawPath로 라우팅하면(인코딩된 '/' 등) 파라미터를 디코딩하지 않으므로 그때만 디코딩한다
func pathParam(c echo.Context, name string) string {
	v := c.Param(name)
	if c.Request().URL.RawPath == "" {
		return v
	}
	d, err := url.PathUnescape(v)
	if err != nil {
		panic(web.Err(400, "잘못된 경로 값입니다: %s", name))
	}
	return d
}
