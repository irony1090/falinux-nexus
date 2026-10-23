package router

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"nexus/internal/protocol"
)

// 경로-1(후보 요청) — REF-path-input.md. 큰 폴더도 빠짐없이 거르려고 worker가 앞글자로 거른다(Q42(자동완성 원천))
const (
	suggestMax  = 200 // 요청 Max 상한
	suggestRead = 512 // ReadDir 한 번에 읽는 수(큰 폴더를 한꺼번에 메모리에 올리지 않음)
)

func (r *workerRouter) dirSuggest(req protocol.Frame) (any, error) {
	var body protocol.DirSuggestRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	return suggestDir(body), nil
}

func suggestDir(body protocol.DirSuggestRequest) protocol.DirSuggestResponse {
	max := min(body.Max, suggestMax)
	if body.Dir == "" {
		l, err := topListing()
		if err != nil {
			return suggestFail(err)
		}
		if l.path != "" { // Linux = `/` 목록
			return suggestIn(l.path, body.Prefix, max)
		}
		out := protocol.DirSuggestResponse{Items: []protocol.DirCrumb{}}
		for _, it := range l.items {
			if hasPrefixFold(it.entry.Name, body.Prefix) {
				out.Items = append(out.Items, protocol.DirCrumb{Name: it.entry.Name, Path: it.entry.Path})
			}
		}
		out.Items, out.More = cutSuggest(out.Items, body.Prefix, max)
		return out
	}
	if v := filepath.VolumeName(body.Dir); v != "" && v == body.Dir { // Windows `C:`만 = 그 드라이브 루트(`C:`는 그 드라이브의 현재 폴더라 상대 경로)
		body.Dir += string(filepath.Separator)
	}
	if !filepath.IsAbs(body.Dir) {
		return protocol.DirSuggestResponse{Err: protocol.DirErrBadPath, Reason: "절대 경로가 아닙니다: " + body.Dir}
	}
	return suggestIn(filepath.Clean(body.Dir), body.Prefix, max)
}

// suggestIn: 숨은 폴더(.으로 시작)는 앞글자가 .일 때만(셸 자동완성과 같음)
func suggestIn(dir, prefix string, max int) protocol.DirSuggestResponse {
	st, err := os.Stat(dir)
	if err != nil {
		return suggestFail(err)
	}
	if !st.IsDir() {
		return protocol.DirSuggestResponse{Err: protocol.DirErrNotDir, Reason: "폴더가 아닙니다: " + dir}
	}
	out := protocol.DirSuggestResponse{Dir: dir, Items: []protocol.DirCrumb{}}
	if max <= 0 {
		return out
	}
	f, err := os.Open(dir)
	if err != nil {
		return suggestFail(err)
	}
	defer f.Close()
	hidden := strings.HasPrefix(prefix, ".")
	for {
		ents, err := f.ReadDir(suggestRead)
		for _, e := range ents {
			name := e.Name()
			if (!hidden && strings.HasPrefix(name, ".")) || !hasPrefixFold(name, prefix) {
				continue
			}
			full := filepath.Join(dir, name)
			if !e.IsDir() {
				if e.Type().IsRegular() {
					continue
				}
				if st, err := os.Stat(full); err != nil || !st.IsDir() { // 링크·junction = 원본이 폴더면 후보
					continue
				}
			}
			out.Items = append(out.Items, protocol.DirCrumb{Name: name, Path: full})
		}
		if errors.Is(err, io.EOF) || len(ents) == 0 {
			break
		}
		if err != nil {
			return suggestFail(err)
		}
	}
	out.Items, out.More = cutSuggest(out.Items, prefix, max)
	return out
}

// cutSuggest: 대소문자까지 같은 앞글자 먼저, 그다음 이름순(대소문자 무시)
func cutSuggest(items []protocol.DirCrumb, prefix string, max int) ([]protocol.DirCrumb, bool) {
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i].Name, items[j].Name
		if ea, eb := strings.HasPrefix(a, prefix), strings.HasPrefix(b, prefix); ea != eb {
			return ea
		}
		if la, lb := strings.ToLower(a), strings.ToLower(b); la != lb {
			return la < lb
		}
		return a < b
	})
	if len(items) > max {
		return items[:max], true
	}
	return items, false
}

func hasPrefixFold(name, prefix string) bool {
	return len(name) >= len(prefix) && strings.EqualFold(name[:len(prefix)], prefix)
}

func suggestFail(err error) protocol.DirSuggestResponse {
	f := dirFail(err)
	return protocol.DirSuggestResponse{Err: f.Err, Reason: f.Reason}
}
