package router

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/util"
)

// 폴더-1(목록 요청) — REF-worker-dir.md. 안 다(worker 캐시): 첫 요청에 전부 읽어 정렬해 두고 페이지로 잘라 준다
const (
	dirPageSize = 500
	dirKeepFor  = 60 * time.Second // 마지막 요청 뒤 보관 시간
	dirKeepMax  = 32               // 연결당 보관 목록 수(넘으면 오래된 것부터 버림)
)

type dirItem struct {
	entry  protocol.DirEntry
	src    os.DirEntry // 크기·시각은 페이지를 자를 때 채운다(nil = 드라이브)
	filled bool
}

type dirListing struct {
	path   string
	parent *string
	crumbs []protocol.DirCrumb
	items  []dirItem
	used   time.Time
}

type dirError struct{ code, reason string }

func (e dirError) Error() string { return e.reason }

type dirCache struct {
	mu   sync.Mutex
	list map[string]*dirListing
}

func newDirCache() *dirCache { return &dirCache{list: map[string]*dirListing{}} }

// dirList: DIR_LIST 핸들러. 실패는 RES 에러가 아니라 Err 코드로(supervisor가 HTTP 상태로 바꿈)
func (r *workerRouter) dirList(req protocol.Frame) (any, error) {
	var body protocol.DirListRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	if body.Cursor != "" {
		return r.dirs.page(body.Cursor, body.Offset), nil
	}
	l, err := readListing(body.Path)
	if err != nil {
		return dirFail(err), nil
	}
	cursor, err := r.dirs.put(l)
	if err != nil {
		return dirFail(err), nil
	}
	return r.dirs.page(cursor, 0), nil
}

func dirFail(err error) protocol.DirListResponse {
	var de dirError
	if errors.As(err, &de) {
		return protocol.DirListResponse{Err: de.code, Reason: de.reason}
	}
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return protocol.DirListResponse{Err: protocol.DirErrNotFound, Reason: err.Error()}
	case errors.Is(err, fs.ErrPermission):
		return protocol.DirListResponse{Err: protocol.DirErrDenied, Reason: err.Error()}
	}
	return protocol.DirListResponse{Err: protocol.DirErrOther, Reason: err.Error()}
}

// readListing: 빈 path = 최상위(topListing, OS별). 폴더 먼저 + 이름순
func readListing(path string) (*dirListing, error) {
	if path == "" {
		return topListing()
	}
	if !filepath.IsAbs(path) {
		return nil, dirError{protocol.DirErrBadPath, "절대 경로가 아닙니다: " + path}
	}
	path = filepath.Clean(path)
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return nil, dirError{protocol.DirErrNotDir, "폴더가 아닙니다: " + path}
	}
	ents, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	items := make([]dirItem, len(ents))
	for i, e := range ents {
		items[i] = dirItem{entry: dirEntryOf(path, e), src: e}
	}
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i].entry, items[j].entry
		if da, db := isDirLike(a), isDirLike(b); da != db {
			return da
		}
		if la, lb := strings.ToLower(a.Name), strings.ToLower(b.Name); la != lb {
			return la < lb
		}
		return a.Name < b.Name
	})
	return &dirListing{path: path, parent: parentOf(path), crumbs: crumbsOf(path), items: items}, nil
}

// crumbsOf: 루트(`/`·`C:\`)까지 올라가며 모은 뒤 뒤집는다. 루트 위가 있으면(Windows) 최상위 칸("")을 앞에
func crumbsOf(path string) []protocol.DirCrumb {
	var out []protocol.DirCrumb
	for p := path; ; {
		up := filepath.Dir(p)
		if up == p {
			out = append(out, protocol.DirCrumb{Name: p, Path: p})
			if rp := rootParent(); rp != nil {
				out = append(out, protocol.DirCrumb{Path: *rp})
			}
			break
		}
		out = append(out, protocol.DirCrumb{Name: filepath.Base(p), Path: p})
		p = up
	}
	slices.Reverse(out)
	return out
}

func isDirLike(e protocol.DirEntry) bool {
	return e.Kind == protocol.DirKindDir || (e.Kind == protocol.DirKindLink && e.TargetDir)
}

// parentOf: 루트(Dir가 자기 자신)의 부모는 OS별(rootParent)
func parentOf(path string) *string {
	p := filepath.Dir(path)
	if p == path {
		return rootParent()
	}
	return &p
}

// dirEntryOf: 종류는 ReadDir가 준 것(stat 없음). 링크만 원본을 따라가 본다(Q14(링크 폴더))
func dirEntryOf(dir string, e os.DirEntry) protocol.DirEntry {
	full := filepath.Join(dir, e.Name())
	d := protocol.DirEntry{Name: e.Name(), Path: full, Kind: dirKind(e.Type())}
	var err error
	if d.Kind == protocol.DirKindOther {
		// Windows junction은 Go가 링크가 아닌 irregular로 준다 — 링크로 읽히면 링크로(2026-10-23 VM 확인)
		if d.LinkTarget, err = os.Readlink(full); err == nil {
			d.Kind = protocol.DirKindLink
		}
	} else if d.Kind == protocol.DirKindLink {
		d.LinkTarget, _ = os.Readlink(full)
	}
	if d.Kind != protocol.DirKindLink {
		return d
	}
	t, err := filepath.EvalSymlinks(full)
	if err == nil && t == full && d.LinkTarget != "" { // Windows junction은 EvalSymlinks가 안 따라감 -> 적힌 원본으로
		t = d.LinkTarget
		if !filepath.IsAbs(t) {
			t = filepath.Join(dir, t)
		}
	}
	if err == nil {
		if st, err := os.Stat(t); err == nil {
			d.TargetPath, d.TargetDir = filepath.Clean(t), st.IsDir()
		}
	}
	return d
}

func dirKind(m fs.FileMode) string {
	switch {
	case m&fs.ModeSymlink != 0:
		return protocol.DirKindLink
	case m.IsDir():
		return protocol.DirKindDir
	case m.IsRegular():
		return protocol.DirKindFile
	}
	return protocol.DirKindOther
}

func (it *dirItem) fill() {
	if it.filled || it.src == nil {
		return
	}
	it.filled = true
	info, err := it.src.Info()
	if err != nil {
		return // 읽은 뒤 지워진 항목 = 크기·시각 없이
	}
	it.entry.Size = info.Size()
	it.entry.ModTime = info.ModTime().Unix()
	it.entry.Exec = execOK(it.entry.Name, info.Mode())
	if it.entry.Kind == protocol.DirKindLink && it.entry.TargetPath != "" && !it.entry.TargetDir {
		if st, err := os.Stat(it.entry.TargetPath); err == nil { // 파일 링크 = 원본 기준
			it.entry.Exec = execOK(it.entry.TargetPath, st.Mode())
		}
	}
}

// put: 만료된 것을 먼저 버리고, 그래도 가득이면 가장 오래 안 쓴 것을 버린다
func (c *dirCache) put(l *dirListing) (string, error) {
	cursor, err := util.RandomKey(16, "", "")
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, v := range c.list {
		if now.Sub(v.used) > dirKeepFor {
			delete(c.list, k)
		}
	}
	if len(c.list) >= dirKeepMax {
		oldest := ""
		for k, v := range c.list {
			if oldest == "" || v.used.Before(c.list[oldest].used) {
				oldest = k
			}
		}
		delete(c.list, oldest)
	}
	l.used = now
	c.list[cursor] = l
	return cursor, nil
}

func (c *dirCache) page(cursor string, offset int) protocol.DirListResponse {
	c.mu.Lock()
	defer c.mu.Unlock()
	l, ok := c.list[cursor]
	if !ok || time.Since(l.used) > dirKeepFor {
		delete(c.list, cursor)
		return protocol.DirListResponse{Err: protocol.DirErrExpired, Reason: "목록 보관 시간이 지났습니다"}
	}
	l.used = time.Now()
	offset = max(0, min(offset, len(l.items)))
	end := min(offset+dirPageSize, len(l.items))
	out := make([]protocol.DirEntry, 0, end-offset)
	for i := offset; i < end; i++ {
		l.items[i].fill()
		out = append(out, l.items[i].entry)
	}
	return protocol.DirListResponse{Path: l.path, Parent: l.parent, Crumbs: l.crumbs, Cursor: cursor, Total: len(l.items), Offset: offset, Entries: out}
}
