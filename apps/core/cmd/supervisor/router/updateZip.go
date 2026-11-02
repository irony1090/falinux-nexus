package router

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/url"
	"sort"
	"time"

	"nexus/internal/supervisor/updstore"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 빌드 파일 한 번에 받기 — 기본 = 범위별 그대로(common/<이름>@<버전>/ · main/ · sub/<서브키>/), ?merged=1 = 합친 결과(공용 -> 메인키 -> 서브키, 뒤가 이김)
// 서버가 만들면서 바로 내려보낸다(큰 바이너리를 메모리에 올리지 않음) — REF-update-deploy.md "릴리스 보관 구조"

type zipEntry struct {
	name             string
	ref              updstore.Ref
	scope, sub, path string
	modified         time.Time
}

// updZip: ?merged=1&sub=<서브키>(합친 결과에 넣을 인스턴스, 없으면 서브키 파일 제외)
func (r *supervisorRouter) updZip(c echo.Context) error {
	r.requireAdmin(c)
	r.requireUpdStore()
	ref := updRef(c)
	merged := c.QueryParam("merged") == "1"
	sub := c.QueryParam("sub")
	if sub != "" {
		updName(sub, "서브키")
	}
	b := requireBundle(c, ref)
	ctx := c.Request().Context()
	q := TxQueries(c)

	var entries []zipEntry
	add := func(name string, er updstore.Ref, scope, s, p string, t time.Time) {
		entries = append(entries, zipEntry{name: name, ref: er, scope: scope, sub: s, path: p, modified: t})
	}
	if ref.Kind == updstore.KindBuild {
		uses, err := q.ListBundleUses(ctx, b.ID)
		if err != nil {
			panic(web.Err(500, "%v", err))
		}
		for _, u := range uses {
			files, err := q.ListBundleFiles(ctx, u.ID)
			if err != nil {
				panic(web.Err(500, "%v", err))
			}
			cref := updstore.Ref{Kind: updstore.KindCommon, Owner: u.Owner, Version: u.Version}
			for _, f := range files {
				add(fmt.Sprintf("common/%s@%s/%s", u.Owner, u.Version, f.Path), cref, f.Scope, "", f.Path, f.UpdatedAt.Time)
			}
		}
	}
	files, err := q.ListBundleFiles(ctx, b.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	for _, f := range files {
		switch {
		case ref.Kind == updstore.KindCommon:
			add(f.Path, ref, f.Scope, "", f.Path, f.UpdatedAt.Time)
		case f.Scope == updstore.ScopeMain:
			add("main/"+f.Path, ref, f.Scope, "", f.Path, f.UpdatedAt.Time)
		case !merged || f.SubKey == sub:
			add(fmt.Sprintf("sub/%s/%s", f.SubKey, f.Path), ref, f.Scope, f.SubKey, f.Path, f.UpdatedAt.Time)
		}
	}
	EndTx(c)

	if merged && ref.Kind == updstore.KindBuild {
		// 뒤가 이긴다 — entries는 이미 공용(uses 순서) -> main -> sub 순서
		last := map[string]int{}
		for i, e := range entries {
			last[e.path] = i
		}
		var out []zipEntry
		for i, e := range entries {
			if last[e.path] == i {
				e.name = e.path
				out = append(out, e)
			}
		}
		sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
		entries = out
	}

	fname := fmt.Sprintf("%s_%s.zip", ref.Owner, ref.Version)
	if merged {
		fname = fmt.Sprintf("%s_%s_merged.zip", ref.Owner, ref.Version)
	}
	h := c.Response().Header()
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(fname)))
	h.Set("Cache-Control", "no-store")
	c.Response().WriteHeader(200)

	zw := zip.NewWriter(c.Response())
	for _, e := range entries {
		if err := r.zipOne(zw, e); err != nil {
			// 머리를 이미 보냈으니 상태 코드로 알릴 수 없다 — 끝 목록(central directory)을 안 써서 받는 쪽이 깨진 zip으로 알게 한다
			log.Printf("[update] zip 중단 %s %s: %v", ref, e.name, err)
			return nil
		}
	}
	if err := zw.Close(); err != nil {
		log.Printf("[update] zip 마무리 실패 %s: %v", ref, err)
	}
	return nil
}

func (r *supervisorRouter) zipOne(zw *zip.Writer, e zipEntry) error {
	f, err := r.upd.OpenFile(e.ref, e.scope, e.sub, e.path)
	if err != nil {
		return err
	}
	defer f.Close()
	w, err := zw.CreateHeader(&zip.FileHeader{Name: e.name, Method: zip.Deflate, Modified: e.modified})
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}
