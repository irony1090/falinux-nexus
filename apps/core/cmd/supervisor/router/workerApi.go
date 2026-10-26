package router

import (
	"context"
	"errors"
	"net/url"
	"slices"
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
func (r *supervisorRouter) mountWorkers(e *echo.Group) {
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
	g.POST("/:mainKey/:subKey/unregistered/:uid/adopt", r.adoptUnregistered) // adopt.go
}

type workerResponse struct {
	MainKey     string `json:"mainKey"`
	SubKey      string `json:"subKey"`
	InstanceKey string `json:"instanceKey"`
	OS          string `json:"os"`    // linux | windows ... — WW5(OS 표시)
	Label       string `json:"label"` // 인스턴스 라벨(회원-Q11), 명부에서만 채움
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
	requireDeviceOwner(ctx, q, key.String, ownerID)
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

	// nodeId 없음 = 폴더에 지정할 후보 = 내 장비 이름만(회원-6(실행 권한))
	var mine []string
	if mainKey == "" {
		var err error
		if mine, err = TxQueries(c).ListOwnedMainKeys(c.Request().Context(), sess.Data.ID); err != nil {
			panic(web.Err(500, "%v", err))
		}
	}
	found := r.workers.FindAll(func(key string, _ *transport.Conn) bool {
		main, _ := splitInstanceKey(key)
		if mainKey == "" {
			return slices.Contains(mine, main)
		}
		return main == mainKey
	})
	labels := instanceLabels(c, sess.Data.ID, mainKey)
	rst := make([]workerResponse, 0, len(found))
	for _, f := range found {
		main, sub := splitInstanceKey(f.Key)
		os, _ := r.workerOS.Get(f.Key)
		rst = append(rst, workerResponse{MainKey: main, SubKey: sub, InstanceKey: f.Key, OS: protocol.NormalizeOS(os), Label: labels[f.Key]})
	}
	sort.Slice(rst, func(i, j int) bool { return rst[i].InstanceKey < rst[j].InstanceKey })
	return c.JSON(200, rst)
}

// instanceLabels: InstanceKey -> 라벨(회원-Q11). mainKey 없음 = 내 장비 전부
func instanceLabels(c echo.Context, userID int64, mainKey string) map[string]string {
	q, ctx := TxQueries(c), c.Request().Context()
	var rows []superdb.WorkerInstance
	var err error
	if mainKey == "" {
		rows, err = q.ListOwnedWorkerInstances(ctx, userID)
	} else {
		rows, err = q.ListWorkerInstances(ctx, mainKey)
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := make(map[string]string, len(rows))
	for _, row := range rows {
		if row.Label != "" {
			out[row.MainKey+"#"+row.SubKey] = row.Label
		}
	}
	return out
}

// listWorkerInstances: 명부(?mainKey= 면 그 장비만). 보이는 범위 = 내 장비 + 볼 수 있는 폴더의 장비 — 회원-6(실행 권한), 옛 RS3(로그인만) 대체
func (r *supervisorRouter) listWorkerInstances(c echo.Context) error {
	sess := r.requireSession(c)
	rows, err := TxQueries(c).ListVisibleWorkerInstances(c.Request().Context(), superdb.ListVisibleWorkerInstancesParams{
		MainKey: c.QueryParam("mainKey"), UserID: sess.Data.ID,
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	rst := make([]workerInstanceResponse, 0, len(rows))
	for _, row := range rows {
		rst = append(rst, r.newWorkerInstanceResponse(row))
	}
	return c.JSON(200, rst)
}

// deleteWorkerInstance: 끊긴 인스턴스만 명부에서 지운다(접속 중 409) — RS4(정리)
func (r *supervisorRouter) deleteWorkerInstance(c echo.Context) error {
	sess := r.requireSession(c)
	mainKey, subKey := pathParam(c, "mainKey"), pathParam(c, "subKey")
	requireMyDevice(c, sess.Data.ID, mainKey)
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
