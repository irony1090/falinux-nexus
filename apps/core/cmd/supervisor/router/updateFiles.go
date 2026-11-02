package router

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/updstore"
	"nexus/internal/web"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// 업데이트 app 파일 — 텍스트 저장·검증·읽기·삭제·조각 업로드. REF-update-deploy.md "P1 착수 계획" 단위-3(API)·단위-6(편집 타일)

// updFileTarget: 범위·경로 검사. P1은 서브키 파일을 쓰지 않는다(자리만)
func updFileTarget(ref updstore.Ref, scope, sub, p string, write bool) (string, string, string) {
	if scope == "" {
		scope = updstore.ScopeMain
	}
	switch scope {
	case updstore.ScopeMain:
		sub = ""
	case updstore.ScopeSub:
		if ref.Kind != updstore.KindBuild {
			panic(web.Err(400, "공용 묶음에는 서브키 파일이 없습니다"))
		}
		if write {
			panic(web.Err(400, "서브키 파일은 아직 만들 수 없습니다(자리만 있음)"))
		}
		updName(sub, "서브키")
	default:
		panic(web.Err(400, "범위는 main · sub 입니다"))
	}
	rel, err := updstore.CleanRel(p)
	if err != nil {
		panic(web.Err(400, "%v", err))
	}
	return scope, sub, rel
}

func isStepsFile(ref updstore.Ref, scope, rel string) bool {
	return ref.Kind == updstore.KindBuild && scope == updstore.ScopeMain && rel == updstore.StepsFile
}

// checkStepsRefs: 문법·구조 + 공용 묶음 존재 + 스크립트 파일 존재. 빌드 안 파일 = 이번에 저장하는 것까지 포함(extra)
func checkStepsRefs(c echo.Context, bundleID int64, src []byte, extra string) (updstore.Steps, []updstore.Issue) {
	st, issues := updstore.Check(src) // 문법 오류면 st가 비어 아래 참조 검사는 할 것이 없다
	ctx := c.Request().Context()
	q := TxQueries(c)
	have := map[string]bool{extra: true}
	if bundleID != 0 {
		files, err := q.ListBundleFiles(ctx, bundleID)
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		for _, f := range files {
			if f.Scope == updstore.ScopeMain {
				have[f.Path] = true
			}
		}
	}
	commons := map[string]int64{}
	exists := func(name, ver string) bool {
		b, err := q.GetBundle(ctx, superdb.GetBundleParams{Kind: updstore.KindCommon, Owner: name, Version: ver})
		if err != nil {
			return false
		}
		commons[name+"@"+ver] = b.ID
		files, err := q.ListBundleFiles(ctx, b.ID)
		if err == nil {
			for _, f := range files {
				have[f.Path] = true
			}
		}
		return true
	}
	// 공용 존재 검사가 먼저 돌아야 공용 파일이 have에 들어간다 — CheckRefs가 uses부터 본다
	issues = append(issues, updstore.CheckRefs(src, st, exists, func(rel string) bool { return have[rel] })...)
	return st, issues
}

func hasError(is []updstore.Issue) bool {
	for _, i := range is {
		if i.Level == "error" {
			return true
		}
	}
	return false
}

// saveUses: steps.json의 uses로 update_uses를 다시 만든다(깔 순서 = 배열 순서)
func saveUses(c echo.Context, bundleID int64, st updstore.Steps) {
	ctx := c.Request().Context()
	q := TxQueries(c)
	if err := q.DeleteBundleUses(ctx, bundleID); err != nil {
		panic(web.Err(500, "%v", err))
	}
	for i, u := range st.Uses {
		b, err := q.GetBundle(ctx, superdb.GetBundleParams{Kind: updstore.KindCommon, Owner: u.Name, Version: u.Version})
		if err != nil {
			panic(web.Err(500, "공용 묶음 %s@%s: %v", u.Name, u.Version, err))
		}
		if err := q.InsertBundleUse(ctx, superdb.InsertBundleUseParams{BundleID: bundleID, CommonID: b.ID, Ord: int32(i)}); err != nil {
			panic(web.Err(500, "%v", err))
		}
	}
}

type updIssuesResponse struct {
	Issues []updstore.Issue `json:"issues"`
	File   *updFileItem     `json:"file,omitempty"`
}

func nonNilIssues(is []updstore.Issue) []updstore.Issue {
	if is == nil {
		return []updstore.Issue{}
	}
	return is
}

// ===== 텍스트 =====

type updTextRequest struct {
	Scope   string `json:"scope"`
	SubKey  string `json:"subKey"`
	Path    string `json:"path"`
	Content string `json:"content"`
	Create  bool   `json:"create"` // 새 파일 — 같은 이름이 있으면 409
}

// updSaveText: 웹 편집 저장. `.sh`는 LF로(REF-update-redesign.md "웹 편집"), steps.json은 오류가 있으면 저장하지 않고 422 + 줄 목록
func (r *supervisorRouter) updSaveText(c echo.Context) error {
	me := r.requireAdmin(c)
	r.requireUpdStore()
	ref := updRef(c)
	var body updTextRequest
	bindJSON(c, &body)
	scope, sub, rel := updFileTarget(ref, body.Scope, body.SubKey, body.Path, true)
	content := body.Content
	if strings.HasSuffix(strings.ToLower(rel), ".sh") {
		content = strings.ReplaceAll(content, "\r\n", "\n")
	}
	data := []byte(content)
	if len(data) > updstore.TextMax {
		panic(web.Err(413, "텍스트 파일은 %dKB까지입니다", updstore.TextMax>>10))
	}
	id := lockBundleForWrite(c, ref)
	ctx := c.Request().Context()
	q := TxQueries(c)
	if body.Create {
		if _, err := q.GetBundleFile(ctx, superdb.GetBundleFileParams{BundleID: id, Scope: scope, SubKey: sub, Path: rel}); err == nil {
			panic(web.Err(409, "%s 파일이 이미 있습니다", rel))
		}
	}
	var issues []updstore.Issue
	steps := isStepsFile(ref, scope, rel)
	var st updstore.Steps
	if steps {
		st, issues = checkStepsRefs(c, id, data, rel)
		if hasError(issues) {
			return c.JSON(422, updIssuesResponse{Issues: issues})
		}
	}
	sum := sha256.Sum256(data)
	f, err := q.UpsertBundleFile(ctx, superdb.UpsertBundleFileParams{BundleID: id, Scope: scope, SubKey: sub, Path: rel, Size: int64(len(data)), Sha: hex.EncodeToString(sum[:]), UpdatedBy: pgtype.Int8{Int64: me, Valid: true}})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if steps {
		saveUses(c, id, st)
	}
	refreshSha(c, id)
	if err := r.upd.WriteBytes(ref, scope, sub, rel, data); err != nil {
		panic(web.Err(500, "파일 저장 실패: %v", err))
	}
	item := updFileItem{Scope: f.Scope, SubKey: f.SubKey, Path: f.Path, Size: f.Size, Sha: f.Sha, Text: true, UpdatedAt: f.UpdatedAt.Time.Unix()}
	return c.JSON(200, updIssuesResponse{Issues: nonNilIssues(issues), File: &item})
}

type updCheckRequest struct {
	Content string `json:"content"`
}

// updCheckSteps: 저장 없이 검증만(편집기의 `검증` 버튼)
func (r *supervisorRouter) updCheckSteps(c echo.Context) error {
	r.requireAdmin(c)
	ref := updRef(c)
	var body updCheckRequest
	bindJSON(c, &body)
	b := requireBundle(c, ref)
	_, issues := checkStepsRefs(c, b.ID, []byte(body.Content), updstore.StepsFile)
	return c.JSON(200, updIssuesResponse{Issues: nonNilIssues(issues)})
}

// ===== 읽기·삭제 =====

func requireBundleFile(c echo.Context, bundleID int64, scope, sub, rel string) superdb.UpdateFile {
	f, err := TxQueries(c).GetBundleFile(c.Request().Context(), superdb.GetBundleFileParams{BundleID: bundleID, Scope: scope, SubKey: sub, Path: rel})
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(404, "%s 파일이 없습니다", rel))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return f
}

// updReadFile: ?scope=&sub=&path=&dl=1 — 편집기 읽기(텍스트) 또는 내려받기(dl)
func (r *supervisorRouter) updReadFile(c echo.Context) error {
	r.requireAdmin(c)
	r.requireUpdStore()
	ref := updRef(c)
	scope, sub, rel := updFileTarget(ref, c.QueryParam("scope"), c.QueryParam("sub"), c.QueryParam("path"), false)
	b := requireBundle(c, ref)
	meta := requireBundleFile(c, b.ID, scope, sub, rel)
	EndTx(c) // 큰 파일을 보내는 동안 DB 연결을 쥐지 않는다
	f, err := r.upd.OpenFile(ref, scope, sub, rel)
	if err != nil {
		panic(web.Err(500, "보관소에 파일이 없습니다(메타와 어긋남): %v", err))
	}
	defer f.Close()
	h := c.Response().Header()
	h.Set("Cache-Control", "no-store")
	h.Set("Content-Length", strconv.FormatInt(meta.Size, 10))
	h.Set("X-Content-Sha", meta.Sha)
	ctype := "application/octet-stream"
	if c.QueryParam("dl") == "1" {
		name := path.Base(rel)
		h.Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(name)))
	} else if r.upd.IsText(ref, scope, sub, rel, meta.Size) {
		ctype = "text/plain; charset=utf-8"
	}
	return c.Stream(http.StatusOK, ctype, f)
}

func (r *supervisorRouter) updDeleteFile(c echo.Context) error {
	r.requireAdmin(c)
	r.requireUpdStore()
	ref := updRef(c)
	scope, sub, rel := updFileTarget(ref, c.QueryParam("scope"), c.QueryParam("sub"), c.QueryParam("path"), true)
	id := lockBundleForWrite(c, ref)
	ctx := c.Request().Context()
	q := TxQueries(c)
	n, err := q.DeleteBundleFile(ctx, superdb.DeleteBundleFileParams{BundleID: id, Scope: scope, SubKey: sub, Path: rel})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if n == 0 {
		panic(web.Err(404, "%s 파일이 없습니다", rel))
	}
	if isStepsFile(ref, scope, rel) {
		if err := q.DeleteBundleUses(ctx, id); err != nil {
			panic(web.Err(500, "%v", err))
		}
	}
	refreshSha(c, id)
	if err := r.upd.Remove(ref, scope, sub, rel); err != nil {
		panic(web.Err(500, "파일 삭제 실패: %v", err))
	}
	return c.JSON(200, map[string]any{})
}

// ===== 조각 업로드 =====
// 시작(번들·경로 검사) -> PUT 조각(512KB, 앞단 nginx 1MB 아래 — upload.go와 같은 이유) -> 마지막 조각에서 sha 계산·DB·제자리 이동

type updUploadRequest struct {
	Scope     string `json:"scope"`
	SubKey    string `json:"subKey"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Overwrite bool   `json:"overwrite"` // 같은 이름이 있으면 확인 창에서 덮어쓰기를 골랐을 때만
}

func (r *supervisorRouter) updStartUpload(c echo.Context) error {
	me := r.requireAdmin(c)
	r.requireUpdStore()
	ref := updRef(c)
	var body updUploadRequest
	bindJSON(c, &body)
	scope, sub, rel := updFileTarget(ref, body.Scope, body.SubKey, body.Path, true)
	if body.Size < 0 || body.Size > updFileMax {
		panic(web.Err(413, "파일은 %dGB까지입니다", updFileMax>>30))
	}
	b := requireBundle(c, ref)
	if b.Locked {
		panic(web.Err(409, "%s는 잠겨 있어 파일을 고칠 수 없습니다", ref))
	}
	if !body.Overwrite {
		if _, err := TxQueries(c).GetBundleFile(c.Request().Context(), superdb.GetBundleFileParams{BundleID: b.ID, Scope: scope, SubKey: sub, Path: rel}); err == nil {
			return c.JSON(409, map[string]any{"error": fmt.Sprintf("%s 파일이 이미 있습니다", rel), "exists": true})
		}
	}
	tmp, err := r.upd.Temp()
	if err != nil {
		panic(web.Err(500, "임시 파일을 만들 수 없습니다: %v", err))
	}
	id := uuid.NewString()
	r.upd.mu.Lock()
	r.upd.uploads[id] = &updUpload{ref: ref, scope: scope, sub: sub, path: rel, size: body.Size, tmp: tmp, userID: me, touched: time.Now()}
	r.upd.mu.Unlock()
	return c.JSON(200, map[string]any{"id": id})
}

func (r *supervisorRouter) takeUpload(id string, remove bool) *updUpload {
	r.upd.mu.Lock()
	defer r.upd.mu.Unlock()
	up := r.upd.uploads[id]
	if up == nil {
		panic(web.Err(404, "업로드가 없거나 오래되어 정리되었습니다. 다시 올리세요"))
	}
	if remove {
		delete(r.upd.uploads, id)
	}
	return up
}

type updChunkResponse struct {
	Received int64            `json:"received"`
	Done     bool             `json:"done"`
	File     *updFileItem     `json:"file,omitempty"`
	Issues   []updstore.Issue `json:"issues,omitempty"`
}

// updUploadChunk: PUT /uploads/:id?offset= — offset이 받은 양보다 크면 409(받은 양), 작으면 덮어씀(재시도 멱등)
func (r *supervisorRouter) updUploadChunk(c echo.Context) error {
	r.requireAdmin(c)
	id := c.Param("id")
	up := r.takeUpload(id, false)
	offset, err := strconv.ParseInt(c.QueryParam("offset"), 10, 64)
	if err != nil || offset < 0 {
		panic(web.Err(400, "잘못된 offset입니다"))
	}
	data, err := io.ReadAll(http.MaxBytesReader(c.Response(), c.Request().Body, uploadChunkMax))
	var tooBig *http.MaxBytesError
	if errors.As(err, &tooBig) {
		panic(web.Err(413, "조각이 너무 큽니다(최대 %d바이트)", uploadChunkMax))
	}
	if err != nil {
		panic(web.Err(400, "%v", err))
	}

	up.mu.Lock()
	defer up.mu.Unlock()
	up.touched = time.Now()
	if offset > up.received {
		return c.JSON(409, updChunkResponse{Received: up.received})
	}
	end := offset + int64(len(data))
	if end > up.size {
		panic(web.Err(400, "올린 양이 처음 알린 크기(%d)를 넘습니다", up.size))
	}
	if _, err := up.tmp.WriteAt(data, offset); err != nil {
		panic(web.Err(500, "임시 파일 쓰기 실패: %v", err))
	}
	up.received = max(up.received, end)
	if up.received < up.size {
		return c.JSON(200, updChunkResponse{Received: up.received})
	}
	return r.updFinishUpload(c, id, up)
}

// updFinishUpload: up.mu를 쥔 채로 부른다. 실패하면 업로드를 버린다(다시 올림)
func (r *supervisorRouter) updFinishUpload(c echo.Context, id string, up *updUpload) error {
	r.takeUpload(id, true)
	name := up.tmp.Name()
	done := false
	defer func() {
		if !done {
			updstore.Discard(up.tmp)
		}
	}()
	if err := up.tmp.Close(); err != nil {
		panic(web.Err(500, "%v", err))
	}
	size, sha, err := updstore.HashFile(name)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	bid := lockBundleForWrite(c, up.ref)
	var issues []updstore.Issue
	var st updstore.Steps
	steps := isStepsFile(up.ref, up.scope, up.path)
	if steps {
		src, err := readSmall(name)
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		st, issues = checkStepsRefs(c, bid, src, up.path)
		if hasError(issues) {
			return c.JSON(422, updChunkResponse{Received: up.received, Issues: issues})
		}
	}
	q := TxQueries(c)
	ctx := c.Request().Context()
	f, err := q.UpsertBundleFile(ctx, superdb.UpsertBundleFileParams{BundleID: bid, Scope: up.scope, SubKey: up.sub, Path: up.path, Size: size, Sha: sha, UpdatedBy: pgtype.Int8{Int64: up.userID, Valid: true}})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if steps {
		saveUses(c, bid, st)
	}
	refreshSha(c, bid)
	if err := r.upd.Commit(name, up.ref, up.scope, up.sub, up.path); err != nil {
		panic(web.Err(500, "파일 저장 실패: %v", err))
	}
	done = true
	item := updFileItem{Scope: f.Scope, SubKey: f.SubKey, Path: f.Path, Size: f.Size, Sha: f.Sha, UpdatedAt: f.UpdatedAt.Time.Unix(),
		Text: r.upd.IsText(up.ref, f.Scope, f.SubKey, f.Path, f.Size)}
	return c.JSON(200, updChunkResponse{Received: up.received, Done: true, File: &item, Issues: issues})
}

func readSmall(name string) ([]byte, error) {
	st, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	if st.Size() > updstore.TextMax {
		return nil, fmt.Errorf("steps.json이 너무 큽니다")
	}
	return os.ReadFile(name)
}

func (r *supervisorRouter) updCancelUpload(c echo.Context) error {
	r.requireAdmin(c)
	up := r.takeUpload(c.Param("id"), true)
	up.mu.Lock()
	updstore.Discard(up.tmp)
	up.mu.Unlock()
	return c.JSON(200, map[string]any{})
}
