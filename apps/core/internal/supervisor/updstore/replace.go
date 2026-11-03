package updstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// steps.json "replace" — 배포 때 장비 대상 폴더에서 비우고 채울 폴더(REF-update-files-tile.md 단위-다(갈아치우기))

type Replace struct {
	Path string
	at   int
}

func checkReplace(v *jnode, add func(int, string, string, ...any)) []Replace {
	if v.kind != jArr {
		add(v.at, "error", `"replace"는 폴더 경로 배열이어야 합니다`)
		return nil
	}
	var out []Replace
	seen := map[string]bool{}
	for _, it := range v.items {
		if it.kind != jStr {
			add(it.at, "error", "replace 항목은 폴더 경로 문자열이어야 합니다")
			continue
		}
		c, err := CleanRel(it.str)
		if err != nil {
			add(it.at, "error", "replace 경로: %v", err)
			continue
		}
		if seen[c] {
			add(it.at, "error", "replace에 %q를 두 번 썼습니다", c)
			continue
		}
		seen[c] = true
		out = append(out, Replace{Path: c, at: it.at})
	}
	for _, r := range out {
		for _, o := range out {
			if under(r.Path, o.Path) {
				add(r.at, "warn", "부모 폴더 %s/가 이미 갈아치우기라 %s/는 따로 적지 않아도 됩니다(배포 때 부모만 보냄)", o.Path, r.Path)
				break
			}
		}
	}
	return out
}

// under: child가 parent 폴더 아래(같은 경로는 아님)
func under(child, parent string) bool { return strings.HasPrefix(child, parent+"/") }

func dirHas(paths []string, dir string) bool {
	for _, p := range paths {
		if under(p, dir) {
			return true
		}
	}
	return false
}

// EffectiveReplace: 배포 메시지에 넣을 폴더 — 메인키 파일이 없는 폴더(확인-3(사라진 폴더))와 부모가 이미 있는 자식은 뺀다
func EffectiveReplace(rs []Replace, mainPaths []string) []string {
	var live []string
	for _, r := range rs {
		if dirHas(mainPaths, r.Path) {
			live = append(live, r.Path)
		}
	}
	var out []string
	for _, d := range live {
		child := false
		for _, o := range live {
			if under(d, o) {
				child = true
				break
			}
		}
		if !child {
			out = append(out, d)
		}
	}
	return out
}

// SetReplace: steps.json 원문에서 replace 값 자리만 바꾼다(확인-2(steps.json 고쳐 쓰기)). 없으면 uses 다음에 넣고, 빈 목록이면 항목째 지운다
func SetReplace(src []byte, dirs []string) ([]byte, error) {
	p := &jparser{src: src}
	root, perr := p.parse()
	if perr != nil {
		return nil, fmt.Errorf("steps.json %d줄 문법 오류: %s", p.issue(perr.at, "", "").Line, perr.msg)
	}
	if root.kind != jObj {
		return nil, errors.New("steps.json 맨 바깥이 { } 객체가 아닙니다")
	}
	splice := func(a, b int, mid string) []byte {
		out := make([]byte, 0, len(src)+len(mid))
		out = append(append(append(out, src[:a]...), mid...), src[b:]...)
		return out
	}
	val := replaceText(dirs)
	ri, ui := -1, -1
	for i, kv := range root.keys {
		switch kv.key {
		case "replace":
			ri = i
		case "uses":
			ui = i
		}
	}
	keys := root.keys
	switch {
	case ri >= 0 && len(dirs) > 0:
		return splice(keys[ri].val.at, keys[ri].val.end, val), nil
	case ri >= 0 && ri+1 < len(keys): // 지우기 — 다음 키가 이 자리로 당겨짐
		return splice(keys[ri].at, keys[ri+1].at, ""), nil
	case ri > 0: // 지우기 — 마지막 키면 앞 값 뒤 쉼표부터
		return splice(keys[ri-1].val.end, keys[ri].val.end, ""), nil
	case ri == 0:
		return splice(keys[0].at, keys[0].val.end, ""), nil
	case len(dirs) == 0:
		return src, nil
	}
	entry := `"replace": ` + val
	switch {
	case ui >= 0 && ui+1 < len(keys):
		at := keys[ui+1].at
		return splice(at, at, entry+","+lineSep(src, at)), nil
	case ui >= 0:
		at := keys[ui].val.end
		return splice(at, at, ","+lineSep(src, keys[ui].at)+entry), nil
	case len(keys) > 0:
		return splice(keys[0].at, keys[0].at, entry+","+lineSep(src, keys[0].at)), nil
	}
	return splice(root.at+1, root.at+1, entry), nil
}

func replaceText(dirs []string) string {
	q := make([]string, len(dirs))
	for i, d := range dirs {
		b, _ := json.Marshal(d)
		q[i] = string(b)
	}
	return "[" + strings.Join(q, ", ") + "]"
}

// lineSep: at의 키가 줄 맨 앞(들여쓰기 뒤)에 있으면 "\n" + 그 들여쓰기, 한 줄 JSON이면 " "
func lineSep(src []byte, at int) string {
	i := at
	for i > 0 && (src[i-1] == ' ' || src[i-1] == '\t') {
		i--
	}
	if i == 0 || src[i-1] == '\n' {
		nl := "\n"
		if i >= 2 && src[i-2] == '\r' {
			nl = "\r\n"
		}
		return nl + string(src[i:at])
	}
	return " "
}
