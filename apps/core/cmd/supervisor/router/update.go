package router

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"nexus/cmd/supervisor/constants"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/updstore"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// 업데이트 app 관리자 API — 장비 루트·빌드·공용 묶음 보관(P1(보관·화면)). REF-update-deploy.md "P1 착수 계획" 단위-3(API), 규칙 → REF-update-rules.md
// 쓰기 순서: 검사 -> DB(번들 행 FOR UPDATE) -> 파일 이름 변경 -> 반환(커밋). 이름 변경을 행 잠금 안에서 해야 같은 파일 동시 저장이 DB와 어긋나지 않는다

func (r *supervisorRouter) mountUpdates(e *echo.Group) {
	g := e.Group("/admin/updates")
	g.GET("/devices", r.updListDevices)
	g.POST("/devices", r.updCreateDevice)
	g.GET("/devices/:owner", r.updGetDevice)
	g.PATCH("/devices/:owner", r.updPatchDevice)
	g.DELETE("/devices/:owner", r.updDeleteDevice)
	g.GET("/devices/:owner/key", r.updShowKey)                      // updaterKey.go
	g.POST("/devices/:owner/key", r.updReissueKey)                  // updaterKey.go
	g.PATCH("/devices/:owner/instances/:sub", r.updPatchInstance)   // updaterConn.go
	g.DELETE("/devices/:owner/instances/:sub", r.updDeleteInstance) // updaterConn.go
	g.POST("/devices/:owner/instances/:sub/runs", r.updStartRun)    // updaterRun.go
	g.GET("/devices/:owner/instances/:sub/runs", r.updListRuns)     // updaterRun.go
	g.GET("/runs/:id", r.updGetRun)                                 // updaterRun.go
	g.POST("/runs/:id/input", r.updRunInput)                        // updaterRun.go
	g.POST("/runs/:id/cancel", r.updRunCancel)                      // updaterRun.go
	g.GET("/deployments/:id", r.updGetDeployment)                   // updaterDeploy.go
	g.POST("/deployments/:id/rollback", r.updRollback)              // updaterDeploy.go
	g.POST("/deployments/:id/cancel", r.updDeployCancel)            // updaterDeploy.go
	g.GET("/commons", r.updListCommons)

	g.POST("/bundles/:kind/:owner", r.updCreateBundle)
	b := g.Group("/bundles/:kind/:owner/:version")
	b.GET("", r.updGetBundle)
	b.PATCH("", r.updPatchBundle)
	b.DELETE("", r.updDeleteBundle)
	b.GET("/file", r.updReadFile)      // updateFiles.go
	b.DELETE("/file", r.updDeleteFile) // updateFiles.go
	b.PATCH("/mode", r.updSetMode)     // updateFiles.go
	b.PUT("/text", r.updSaveText)      // updateFiles.go
	b.PUT("/replace", r.updSetReplace) // updateFiles.go
	b.POST("/check", r.updCheckSteps)  // updateFiles.go
	b.POST("/uploads", r.updStartUpload)
	b.GET("/zip", r.updZip)         // updateZip.go
	b.POST("/deploy", r.updDeploy)  // updaterDeploy.go
	b.POST("/lock", r.updLockBuild) // updaterDeploy.go
	b.GET("/trial", r.updGetTrial)  // updaterTrial.go
	g.PUT("/uploads/:id", r.updUploadChunk)
	g.DELETE("/uploads/:id", r.updCancelUpload)
}

// ===== 보관소 + 받는 중 파일 =====

const (
	updUploadIdle = 30 * time.Minute
	updFileMax    = 4 << 30
)

type updUpload struct {
	mu       sync.Mutex
	ref      updstore.Ref
	scope    string
	sub      string
	path     string
	size     int64
	received int64
	tmp      *os.File
	userID   int64
	touched  time.Time
}

type updStore struct {
	*updstore.Store
	mu      sync.Mutex
	uploads map[string]*updUpload
}

func newUpdStore(env constants.EnvVars) *updStore {
	owner := fmt.Sprintf("%s:%d/%s", env.DBHost, env.DBPort, env.DBName)
	return &updStore{Store: updstore.Open(env.UpdateStore, owner), uploads: map[string]*updUpload{}}
}

// runSweep: 멈춘 업로드의 임시 파일 정리
func (u *updStore) runSweep() {
	for range time.Tick(5 * time.Minute) {
		u.mu.Lock()
		for id, up := range u.uploads {
			if up.mu.TryLock() {
				if time.Since(up.touched) > updUploadIdle {
					updstore.Discard(up.tmp)
					delete(u.uploads, id)
				}
				up.mu.Unlock()
			}
		}
		u.mu.Unlock()
	}
}

func (r *supervisorRouter) requireUpdStore() {
	if !r.upd.Enabled() {
		panic(web.Err(503, "업데이트 보관소가 꺼져 있습니다(UPDATE_STORE_DIR — 서버 로그 확인)"))
	}
}

// ===== 응답 =====

type updDeviceItem struct {
	MainKey       string `json:"mainKey"`
	Memo          string `json:"memo"`
	BuildCount    int64  `json:"buildCount"`
	InstanceCount int64  `json:"instanceCount"`
	LastBuildAt   *int64 `json:"lastBuildAt"`
	CreatedAt     int64  `json:"createdAt"`
	HasKey        bool   `json:"hasKey"`
	KeyIssuedAt   *int64 `json:"keyIssuedAt,omitempty"`
}

type updaterItem struct {
	SubKey           string `json:"subKey"`
	Label            string `json:"label"`
	Memo             string `json:"memo"`
	OS               string `json:"os"`
	Arch             string `json:"arch"`
	InstallDir       string `json:"installDir"`
	InstalledVersion string `json:"installedVersion"`
	LastSeenAt       *int64 `json:"lastSeenAt"`
	Online           bool   `json:"online"`    // 메모리 연결(updaterConn.go)
	Running          *int64 `json:"running"`   // 지금 실행 중인 단계 기록 id
	Deploying        *int64 `json:"deploying"` // 지금 실행 중인 배포 id(P3)
	Version          string `json:"version"`   // updater 빌드 태그(접속 중일 때)
}

type updBundleItem struct {
	Kind          string `json:"kind"`
	Owner         string `json:"owner"`
	Version       string `json:"version"`
	Note          string `json:"note"`
	Sha           string `json:"sha"`
	Locked        bool   `json:"locked"`
	UsedBy        int64  `json:"usedBy"`
	FileCount     int64  `json:"fileCount"`
	TotalSize     int64  `json:"totalSize"`
	CreatedByName string `json:"createdByName"`
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
	LastDeployID  int64  `json:"lastDeployId"` // 0 = 배포 없음(장비 타일 빌드 행 Run · Log = 이 배포)
}

type updDeviceDetail struct {
	updDeviceItem
	Instances []updaterItem   `json:"instances"`
	Builds    []updBundleItem `json:"builds"`
}

func tsPtr(t pgtype.Timestamptz) *int64 {
	if !t.Valid {
		return nil
	}
	v := t.Time.Unix()
	return &v
}

func bundleItemOf(row superdb.ListBundlesByOwnerRow) updBundleItem {
	return updBundleItem{
		Kind: row.Kind, Owner: row.Owner, Version: row.Version, Note: row.Note, Sha: row.Sha,
		Locked: row.Locked, UsedBy: row.UsedBy, FileCount: row.FileCount, TotalSize: row.TotalSize,
		CreatedByName: row.CreatedByName, CreatedAt: row.CreatedAt.Time.Unix(), UpdatedAt: row.UpdatedAt.Time.Unix(), LastDeployID: row.LastDeployID,
	}
}

// ===== 파라미터 =====

func updName(v, what string) string {
	if !updstore.NamePattern.MatchString(v) {
		panic(web.Err(400, "%s은(는) 영문·숫자·._- 1~64자입니다", what))
	}
	return v
}

func updRef(c echo.Context) updstore.Ref {
	kind := c.Param("kind")
	if kind != updstore.KindBuild && kind != updstore.KindCommon {
		panic(web.Err(400, "번들 종류는 build · common 입니다"))
	}
	return updstore.Ref{Kind: kind, Owner: updName(c.Param("owner"), "이름"), Version: updName(c.Param("version"), "버전")}
}

func bindJSON(c echo.Context, v any) {
	if err := c.Bind(v); err != nil {
		panic(web.Err(400, "%v", err))
	}
}

func requireUpdDevice(c echo.Context, mainKey string) superdb.UpdateDevice {
	d, err := TxQueries(c).GetUpdateDevice(c.Request().Context(), mainKey)
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "장비 %s가 없습니다", mainKey))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return d
}

func requireBundle(c echo.Context, ref updstore.Ref) superdb.GetBundleRow {
	b, err := TxQueries(c).GetBundle(c.Request().Context(), superdb.GetBundleParams{Kind: ref.Kind, Owner: ref.Owner, Version: ref.Version})
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "%s가 없습니다", ref))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return b
}

// lockBundleForWrite: 번들 행을 잠그고(FOR UPDATE) 잠긴 번들이면 409 — REF-update-rules.md "잠금"(잠긴 뒤 파일·이름 수정 불가)
func lockBundleForWrite(c echo.Context, ref updstore.Ref) int64 {
	b, err := TxQueries(c).GetBundleForUpdate(c.Request().Context(), superdb.GetBundleForUpdateParams{Kind: ref.Kind, Owner: ref.Owner, Version: ref.Version})
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "%s가 없습니다", ref))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if b.Locked {
		panic(web.Err(409, "%s는 잠겨 있어 파일을 고칠 수 없습니다", ref))
	}
	return b.ID
}

// refreshSha: 파일 메타에서 폴더 sha를 다시 계산
func refreshSha(c echo.Context, bundleID int64) {
	ctx := c.Request().Context()
	q := TxQueries(c)
	files, err := q.ListBundleFiles(ctx, bundleID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	es := make([]updstore.Entry, len(files))
	for i, f := range files {
		es[i] = updstore.Entry{Scope: f.Scope, Sub: f.SubKey, Path: f.Path, Sha: f.Sha, Mode: f.Mode}
	}
	if err := q.UpdateBundleSha(ctx, superdb.UpdateBundleShaParams{ID: bundleID, Sha: updstore.FolderSha(es)}); err != nil {
		panic(web.Err(500, "%v", err))
	}
}

// ===== 장비 루트 =====

func (r *supervisorRouter) updListDevices(c echo.Context) error {
	r.requireAdmin(c)
	rows, err := TxQueries(c).ListUpdateDevices(c.Request().Context())
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	all := make([]updDeviceItem, len(rows))
	for i, d := range rows {
		all[i] = updDeviceItem{MainKey: d.MainKey, Memo: d.Memo, BuildCount: d.BuildCount, InstanceCount: d.InstanceCount, LastBuildAt: tsPtr(d.LastBuildAt), CreatedAt: d.CreatedAt.Time.Unix(), HasKey: d.HasKey}
	}
	lq := parseListQuery(c)
	match := func(d updDeviceItem) bool { return containsFold(d.MainKey, lq.Q) || containsFold(d.Memo, lq.Q) }
	filters := map[string]func(updDeviceItem) bool{
		"nobuild": func(d updDeviceItem) bool { return d.BuildCount == 0 },
		"nokey":   func(d updDeviceItem) bool { return !d.HasKey },
	}
	return c.JSON(200, pageList(all, lq, match, filters))
}

type updDeviceRequest struct {
	MainKey string  `json:"mainKey"`
	Memo    *string `json:"memo"`
}

// updCreateDevice: 장비 루트 + updater 키 발급(정-1(접속 인증) — 첫 접속 자동 생성 없음, 관리자가 먼저 만든다). 키는 응답으로 한 번, 이후 키 보기로
func (r *supervisorRouter) updCreateDevice(c echo.Context) error {
	r.requireAdmin(c)
	var body updDeviceRequest
	bindJSON(c, &body)
	key := updName(body.MainKey, "메인키")
	ctx := c.Request().Context()
	q := TxQueries(c)
	if _, err := q.GetUpdateDevice(ctx, key); err == nil {
		panic(web.Err(409, "장비 %s가 이미 있습니다", key))
	}
	memo := ""
	if body.Memo != nil {
		memo = *body.Memo
	}
	apiKey := newUpdaterKey()
	if _, err := q.CreateUpdateDevice(ctx, superdb.CreateUpdateDeviceParams{MainKey: key, Memo: memo, ApiKeyHash: apiKeyHash(apiKey), ApiKeyEnc: sealAPIKey(apiKey)}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	return c.JSON(200, updKeyResponse{MainKey: key, APIKey: apiKey})
}

func (r *supervisorRouter) updGetDevice(c echo.Context) error {
	r.requireAdmin(c)
	key := updName(c.Param("owner"), "메인키")
	d := requireUpdDevice(c, key)
	ctx := c.Request().Context()
	q := TxQueries(c)
	ups, err := q.ListUpdatersByMainKey(ctx, key)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	builds, err := q.ListBundlesByOwner(ctx, superdb.ListBundlesByOwnerParams{Kind: updstore.KindBuild, Owner: key})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	out := updDeviceDetail{
		updDeviceItem: updDeviceItem{MainKey: d.MainKey, Memo: d.Memo, BuildCount: int64(len(builds)), InstanceCount: int64(len(ups)), CreatedAt: d.CreatedAt.Time.Unix(),
			HasKey: d.ApiKeyHash != "", KeyIssuedAt: tsPtr(d.KeyIssuedAt)},
		Instances: make([]updaterItem, len(ups)),
		Builds:    make([]updBundleItem, len(builds)),
	}
	for i, u := range ups {
		out.Instances[i] = r.updaterItemOf(u)
	}
	for i, b := range builds {
		out.Builds[i] = bundleItemOf(b)
		if t := b.UpdatedAt.Time.Unix(); out.LastBuildAt == nil || t > *out.LastBuildAt {
			out.LastBuildAt = &t
		}
	}
	return c.JSON(200, out)
}

func (r *supervisorRouter) updPatchDevice(c echo.Context) error {
	r.requireAdmin(c)
	key := updName(c.Param("owner"), "메인키")
	var body updDeviceRequest
	bindJSON(c, &body)
	if body.Memo == nil {
		return c.JSON(200, map[string]any{})
	}
	n, err := TxQueries(c).UpdateUpdateDeviceMemo(c.Request().Context(), superdb.UpdateUpdateDeviceMemoParams{MainKey: key, Memo: *body.Memo})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if n == 0 {
		panic(web.Err(404, "장비 %s가 없습니다", key))
	}
	return c.JSON(200, map[string]any{})
}

// updDeleteDevice: 빌드가 남아 있으면 거절(빌드부터 지운다). 인스턴스 명부는 같이 지워지고, 다시 접속하면 새로 생긴다
func (r *supervisorRouter) updDeleteDevice(c echo.Context) error {
	r.requireAdmin(c)
	key := updName(c.Param("owner"), "메인키")
	ctx := c.Request().Context()
	q := TxQueries(c)
	requireUpdDevice(c, key)
	builds, err := q.ListBundlesByOwner(ctx, superdb.ListBundlesByOwnerParams{Kind: updstore.KindBuild, Owner: key})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if len(builds) > 0 {
		panic(web.Err(409, "빌드 %d개가 남아 있습니다. 빌드를 먼저 지우세요", len(builds)))
	}
	if _, err := q.DeleteUpdateDevice(ctx, key); err != nil {
		panic(web.Err(500, "%v", err))
	}
	AfterCommit(c, func() { r.dropUpdaters(key) })
	if r.upd.Enabled() {
		AfterCommit(c, func() {
			if err := r.upd.RemoveOwner(updstore.KindBuild, key); err != nil {
				log.Printf("[update] 장비 폴더 삭제 실패 %s: %v", key, err)
			}
		})
	}
	return c.JSON(200, map[string]any{})
}

// ===== 공용 묶음 목록 =====

type updCommonGroup struct {
	Name     string          `json:"name"`
	Versions []updBundleItem `json:"versions"`
}

func (r *supervisorRouter) updListCommons(c echo.Context) error {
	r.requireAdmin(c)
	rows, err := TxQueries(c).ListCommonBundles(c.Request().Context())
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	var all []updCommonGroup
	for _, b := range rows {
		item := bundleItemOf(superdb.ListBundlesByOwnerRow(b))
		if n := len(all); n > 0 && all[n-1].Name == b.Owner {
			all[n-1].Versions = append(all[n-1].Versions, item)
			continue
		}
		all = append(all, updCommonGroup{Name: b.Owner, Versions: []updBundleItem{item}})
	}
	lq := parseListQuery(c)
	match := func(g updCommonGroup) bool { return containsFold(g.Name, lq.Q) }
	filters := map[string]func(updCommonGroup) bool{
		"unused": func(g updCommonGroup) bool {
			for _, v := range g.Versions {
				if v.UsedBy > 0 {
					return false
				}
			}
			return true
		},
	}
	return c.JSON(200, pageList(all, lq, match, filters))
}

// ===== 번들 =====

type updBundleCreateRequest struct {
	Version string `json:"version"`
	Note    string `json:"note"`
	From    *struct {
		Owner   string `json:"owner"`
		Version string `json:"version"`
	} `json:"from"` // 있으면 복사로 새 버전(같은 종류, 다른 장비의 빌드도 됨)
}

func (r *supervisorRouter) updCreateBundle(c echo.Context) error {
	me := r.requireAdmin(c)
	r.requireUpdStore()
	kind := c.Param("kind")
	if kind != updstore.KindBuild && kind != updstore.KindCommon {
		panic(web.Err(400, "번들 종류는 build · common 입니다"))
	}
	var body updBundleCreateRequest
	bindJSON(c, &body)
	ref := updstore.Ref{Kind: kind, Owner: updName(c.Param("owner"), "이름"), Version: updName(body.Version, "버전")}
	ctx := c.Request().Context()
	q := TxQueries(c)
	if kind == updstore.KindBuild {
		requireUpdDevice(c, ref.Owner)
	}
	if _, err := q.GetBundle(ctx, superdb.GetBundleParams{Kind: ref.Kind, Owner: ref.Owner, Version: ref.Version}); err == nil {
		panic(web.Err(409, "%s가 이미 있습니다", ref))
	}
	var src *superdb.GetBundleRow
	var srcRef updstore.Ref
	if body.From != nil {
		srcRef = updstore.Ref{Kind: kind, Owner: updName(body.From.Owner, "복사할 이름"), Version: updName(body.From.Version, "복사할 버전")}
		b := requireBundle(c, srcRef)
		src = &b
	}
	nb, err := q.CreateBundle(ctx, superdb.CreateBundleParams{Kind: ref.Kind, Owner: ref.Owner, Version: ref.Version, Note: body.Note, CreatedBy: pgtype.Int8{Int64: me, Valid: true}})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	// DB에 행이 없던 이름이니 폴더가 있으면 지난 실패의 찌꺼기
	if err := r.upd.RemoveBundle(ref); err != nil {
		panic(web.Err(500, "보관소 정리 실패: %v", err))
	}
	if src == nil {
		if err := r.upd.MkBundle(ref); err != nil {
			panic(web.Err(500, "보관소 폴더를 만들 수 없습니다: %v", err))
		}
		return c.JSON(200, map[string]any{"version": ref.Version})
	}
	if err := r.upd.CopyBundle(srcRef, ref); err != nil {
		panic(web.Err(500, "파일 복사 실패: %v", err))
	}
	if err := q.CopyBundleFiles(ctx, superdb.CopyBundleFilesParams{DstID: nb.ID, SrcID: src.ID, UserID: me}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	if err := q.CopyBundleUses(ctx, superdb.CopyBundleUsesParams{DstID: nb.ID, SrcID: src.ID}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	if body.Note == "" && src.Note != "" {
		_, _ = q.UpdateBundleNote(ctx, superdb.UpdateBundleNoteParams{ID: nb.ID, Note: src.Note})
	}
	refreshSha(c, nb.ID)
	return c.JSON(200, map[string]any{"version": ref.Version})
}

type updFileItem struct {
	Scope     string `json:"scope"`
	SubKey    string `json:"subKey"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Sha       string `json:"sha"`
	Text      bool   `json:"text"`
	Mode      int32  `json:"mode"`
	UpdatedAt int64  `json:"updatedAt"`
}

type updUseItem struct {
	Name    string        `json:"name"`
	Version string        `json:"version"`
	Sha     string        `json:"sha"`
	Locked  bool          `json:"locked"`
	Files   []updFileItem `json:"files"`
}

type updUserItem struct {
	MainKey string `json:"mainKey"`
	Version string `json:"version"`
	Locked  bool   `json:"locked"`
}

type updBundleDetail struct {
	updBundleItem
	Files []updFileItem `json:"files"`
	Uses  []updUseItem  `json:"uses"`  // 빌드만 — 깔 순서
	Users []updUserItem `json:"users"` // 공용만 — 쓰는 빌드
	// Lockable: 잠그기 판정(DB 함수 update_lockable) — 지금 내용으로 끝까지 성공한 배포가 있음
	Lockable bool `json:"lockable"`
	// Replace: 빌드 steps.json의 replace 그대로(폴더 행 토글 표시용). steps.json이 없거나 오류면 빈 배열
	Replace []string `json:"replace"`
}

func (r *supervisorRouter) fileItems(ref updstore.Ref, files []superdb.UpdateFile) []updFileItem {
	out := make([]updFileItem, len(files))
	for i, f := range files {
		out[i] = updFileItem{Scope: f.Scope, SubKey: f.SubKey, Path: f.Path, Size: f.Size, Sha: f.Sha, Mode: f.Mode, UpdatedAt: f.UpdatedAt.Time.Unix(),
			Text: r.upd.Enabled() && r.upd.IsText(ref, f.Scope, f.SubKey, f.Path, f.Size)}
	}
	return out
}

func (r *supervisorRouter) updGetBundle(c echo.Context) error {
	r.requireAdmin(c)
	ref := updRef(c)
	b := requireBundle(c, ref)
	ctx := c.Request().Context()
	q := TxQueries(c)
	files, err := q.ListBundleFiles(ctx, b.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	var size int64
	for _, f := range files {
		size += f.Size
	}
	out := updBundleDetail{
		updBundleItem: updBundleItem{Kind: b.Kind, Owner: b.Owner, Version: b.Version, Note: b.Note, Sha: b.Sha, Locked: b.Locked, UsedBy: b.UsedBy,
			FileCount: int64(len(files)), TotalSize: size, CreatedByName: b.CreatedByName, CreatedAt: b.CreatedAt.Time.Unix(), UpdatedAt: b.UpdatedAt.Time.Unix(), LastDeployID: b.LastDeployID},
		Files: r.fileItems(ref, files), Uses: []updUseItem{}, Users: []updUserItem{}, Lockable: b.Lockable, Replace: []string{},
	}
	if b.Kind == updstore.KindBuild {
		out.Replace = r.stepsReplace(ref, files)
		uses, err := q.ListBundleUses(ctx, b.ID)
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		for _, u := range uses {
			cf, err := q.ListBundleFiles(ctx, u.ID)
			if err != nil {
				panic(web.Err(500, "%v", err))
			}
			cref := updstore.Ref{Kind: updstore.KindCommon, Owner: u.Owner, Version: u.Version}
			out.Uses = append(out.Uses, updUseItem{Name: u.Owner, Version: u.Version, Sha: u.Sha, Locked: u.Locked, Files: r.fileItems(cref, cf)})
		}
	} else {
		users, err := q.ListBundleUsers(ctx, b.ID)
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		for _, u := range users {
			out.Users = append(out.Users, updUserItem{MainKey: u.Owner, Version: u.Version, Locked: u.Locked})
		}
	}
	return c.JSON(200, out)
}

func (r *supervisorRouter) stepsReplace(ref updstore.Ref, files []superdb.UpdateFile) []string {
	out := []string{}
	for _, f := range files {
		if f.Scope != updstore.ScopeMain || f.Path != updstore.StepsFile || !r.upd.Enabled() {
			continue
		}
		src, err := r.readStoreFile(ref, f)
		if err != nil {
			return out
		}
		if st, issues := updstore.Check(src); !hasError(issues) {
			for _, x := range st.Replace {
				out = append(out, x.Path)
			}
		}
	}
	return out
}

type updBundlePatchRequest struct {
	Note *string `json:"note"`
}

// updPatchBundle: 설명 글 — 잠긴 뒤에도 고칠 수 있음(REF-update-rules.md "잠금")
func (r *supervisorRouter) updPatchBundle(c echo.Context) error {
	r.requireAdmin(c)
	ref := updRef(c)
	var body updBundlePatchRequest
	bindJSON(c, &body)
	b := requireBundle(c, ref)
	if body.Note != nil {
		if _, err := TxQueries(c).UpdateBundleNote(c.Request().Context(), superdb.UpdateBundleNoteParams{ID: b.ID, Note: *body.Note}); err != nil {
			panic(web.Err(500, "%v", err))
		}
	}
	return c.JSON(200, map[string]any{})
}

// updDeleteBundle: 빌드 = 잠겨도 삭제 가능(배포 이력은 사본으로 남음). 공용 = 쓰는 빌드가 있으면 거절 — REF-update-rules.md "삭제"
func (r *supervisorRouter) updDeleteBundle(c echo.Context) error {
	r.requireAdmin(c)
	ref := updRef(c)
	b := requireBundle(c, ref)
	ctx := c.Request().Context()
	q := TxQueries(c)
	if b.Kind == updstore.KindCommon && b.UsedBy > 0 {
		panic(web.Err(409, "이 공용 묶음을 쓰는 빌드가 %d개 있어 지울 수 없습니다", b.UsedBy))
	}
	if _, err := q.DeleteBundle(ctx, b.ID); err != nil {
		panic(web.Err(500, "%v", err))
	}
	if r.upd.Enabled() {
		AfterCommit(c, func() {
			if err := r.upd.RemoveBundle(ref); err != nil {
				log.Printf("[update] 번들 폴더 삭제 실패 %s: %v", ref, err)
			}
		})
	}
	return c.JSON(200, map[string]any{})
}
