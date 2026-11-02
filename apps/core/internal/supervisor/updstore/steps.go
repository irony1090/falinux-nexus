package updstore

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// steps.json 검증 — 줄 단위 오류(REF-update-deploy.md "steps.json"). encoding/json은 값 위치를 주지 않아 위치를 기억하는 작은 파서를 둔다

const StepsFile = "steps.json"

var builtins = map[string]bool{"backup": true, "copy": true, "restore": true}

type Issue struct {
	Line  int    `json:"line"`  // 1부터
	Col   int    `json:"col"`   // 1부터(글자 단위)
	Level string `json:"level"` // error | warn
	Msg   string `json:"msg"`
}

// Use: uses 한 항목(공용 이름@버전)과 그 위치
type Use struct {
	Name, Version string
	at            int
}

type Step struct {
	ID, Script, Builtin string
	Timeout             time.Duration // 0 = 기본값(updater가 정함)
	Manual, Interactive bool
	at                  int
}

type Steps struct {
	Uses     []Use
	Deploy   []Step
	Rollback []Step
}

// Check: 문법·구조 검사. 공용 묶음 존재·스크립트 파일 존재는 호출자가 CheckRefs로
func Check(src []byte) (Steps, []Issue) {
	p := &jparser{src: src}
	var st Steps
	root, err := p.parse()
	if err != nil {
		return st, []Issue{p.issue(err.at, "error", err.msg)}
	}
	var out []Issue
	add := func(at int, level, f string, a ...any) { out = append(out, p.issue(at, level, fmt.Sprintf(f, a...))) }

	if root.kind != jObj {
		add(root.at, "error", "맨 바깥은 { } 객체여야 합니다")
		return st, out
	}
	for _, kv := range root.keys {
		v := kv.val
		switch kv.key {
		case "uses":
			if v.kind != jArr {
				add(v.at, "error", `"uses"는 배열이어야 합니다`)
				continue
			}
			seen := map[string]bool{}
			for _, it := range v.items {
				name, ver, ok := strings.Cut(it.str, "@")
				if it.kind != jStr || !ok || !NamePattern.MatchString(name) || !NamePattern.MatchString(ver) {
					add(it.at, "error", `uses 항목은 "이름@버전" 문자열이어야 합니다`)
					continue
				}
				if seen[name] {
					add(it.at, "error", "공용 묶음 %q를 두 번 썼습니다", name)
					continue
				}
				seen[name] = true
				st.Uses = append(st.Uses, Use{Name: name, Version: ver, at: it.at})
			}
		case "deploy", "rollback":
			steps := checkSteps(v, kv.key, add)
			if kv.key == "deploy" {
				st.Deploy = steps
			} else {
				st.Rollback = steps
			}
		case "autoRollback":
			if v.kind != jBool {
				add(v.at, "error", `"autoRollback"은 true/false여야 합니다`)
			}
		default:
			add(kv.at, "warn", "모르는 항목 %q — 저장은 되지만 updater가 쓰지 않습니다", kv.key)
		}
	}
	if root.get("deploy") == nil {
		add(root.at, "error", `"deploy" 단계 목록이 없습니다`)
	} else if len(st.Deploy) == 0 && !hasLevel(out, "error") {
		add(root.get("deploy").at, "warn", "deploy 단계가 비어 있습니다")
	}
	return st, out
}

func hasLevel(is []Issue, level string) bool {
	for _, i := range is {
		if i.Level == level {
			return true
		}
	}
	return false
}

func checkSteps(v *jnode, list string, add func(int, string, string, ...any)) []Step {
	if v.kind != jArr {
		add(v.at, "error", "%q는 단계 배열이어야 합니다", list)
		return nil
	}
	var out []Step
	ids := map[string]bool{}
	for _, it := range v.items {
		if it.kind != jObj {
			add(it.at, "error", "단계는 { } 객체여야 합니다")
			continue
		}
		s := Step{at: it.at}
		ok := true
		for _, kv := range it.keys {
			f := kv.val
			switch kv.key {
			case "id":
				if f.kind != jStr || !NamePattern.MatchString(f.str) {
					add(f.at, "error", "id는 영문·숫자·._- 1~64자 문자열이어야 합니다")
					ok = false
				} else if ids[f.str] {
					add(f.at, "error", "%s 목록에 id %q가 두 번 있습니다", list, f.str)
					ok = false
				} else {
					ids[f.str] = true
					s.ID = f.str
				}
			case "script":
				if f.kind != jStr {
					add(f.at, "error", "script는 파일 경로 문자열이어야 합니다")
					ok = false
				} else if c, err := CleanRel(f.str); err != nil {
					add(f.at, "error", "script 경로: %v", err)
					ok = false
				} else {
					s.Script = c
				}
			case "builtin":
				if f.kind != jStr || !builtins[f.str] {
					add(f.at, "error", "builtin은 backup · copy · restore 중 하나입니다")
					ok = false
				} else {
					s.Builtin = f.str
				}
			case "manual", "interactive":
				if f.kind != jBool {
					add(f.at, "error", "%s는 true/false여야 합니다", kv.key)
				} else if kv.key == "manual" {
					s.Manual = f.str == "true"
				} else {
					s.Interactive = f.str == "true"
				}
			case "timeout":
				if d, err := time.ParseDuration(f.str); f.kind != jStr || err != nil || d <= 0 {
					add(f.at, "error", `timeout은 "30s" · "15m" 같은 시간 문자열이어야 합니다`)
				} else {
					s.Timeout = d
				}
			default:
				add(kv.at, "warn", "모르는 항목 %q — updater가 쓰지 않습니다", kv.key)
			}
		}
		if it.get("id") == nil {
			add(it.at, "error", "단계에 id가 없습니다")
			ok = false
		}
		if s.Interactive && s.Timeout > 0 {
			add(it.get("timeout").at, "warn", "interactive 단계는 시간 제한을 쓰지 않습니다(입력을 기다릴 수 있어서)")
		}
		hasScript, hasBuiltin := it.get("script") != nil, it.get("builtin") != nil
		if s.Interactive && hasBuiltin {
			add(it.get("interactive").at, "warn", "builtin 단계는 입력을 받지 않습니다")
		}
		if hasScript == hasBuiltin {
			add(it.at, "error", "단계에는 script와 builtin 중 하나만 있어야 합니다")
			ok = false
		}
		if ok {
			out = append(out, s)
		}
	}
	return out
}

// CheckRefs: uses의 공용 묶음이 있는지(error), script 파일이 빌드·공용 안에 있는지(warn — 나중에 올릴 수 있음)
func CheckRefs(src []byte, st Steps, commonExists func(name, ver string) bool, fileExists func(rel string) bool) []Issue {
	p := &jparser{src: src}
	var out []Issue
	for _, u := range st.Uses {
		if !commonExists(u.Name, u.Version) {
			out = append(out, p.issue(u.at, "error", fmt.Sprintf("공용 묶음 %s@%s가 없습니다", u.Name, u.Version)))
		}
	}
	for _, list := range [][]Step{st.Deploy, st.Rollback} {
		for _, s := range list {
			if s.Script != "" && !fileExists(s.Script) {
				out = append(out, p.issue(s.at, "warn", fmt.Sprintf("스크립트 %s가 빌드·공용 묶음에 아직 없습니다", s.Script)))
			}
		}
	}
	return out
}

// ===== 위치를 기억하는 JSON 파서 =====

type jkind int

const (
	jNull jkind = iota
	jBool
	jNum
	jStr
	jArr
	jObj
)

type jkv struct {
	key string
	at  int // 키 위치
	val *jnode
}

type jnode struct {
	kind  jkind
	at    int // 값 시작 바이트
	str   string
	items []*jnode
	keys  []jkv
}

func (n *jnode) get(k string) *jnode {
	for _, kv := range n.keys {
		if kv.key == k {
			return kv.val
		}
	}
	return nil
}

type jerr struct {
	at  int
	msg string
}

type jparser struct {
	src []byte
	i   int
}

func (p *jparser) fail(f string, a ...any) *jerr { return &jerr{p.i, fmt.Sprintf(f, a...)} }

// issue: 바이트 위치 -> 줄·칸(글자 단위)
func (p *jparser) issue(at int, level, msg string) Issue {
	if at > len(p.src) {
		at = len(p.src)
	}
	line, start := 1, 0
	for i := 0; i < at; i++ {
		if p.src[i] == '\n' {
			line++
			start = i + 1
		}
	}
	return Issue{Line: line, Col: utf8.RuneCount(p.src[start:at]) + 1, Level: level, Msg: msg}
}

func (p *jparser) ws() {
	for p.i < len(p.src) {
		switch p.src[p.i] {
		case ' ', '\t', '\n', '\r':
			p.i++
		default:
			return
		}
	}
}

func (p *jparser) parse() (*jnode, *jerr) {
	if len(p.src) >= 3 && string(p.src[:3]) == "\xef\xbb\xbf" {
		p.i = 3
	}
	n, err := p.value(0)
	if err != nil {
		return nil, err
	}
	p.ws()
	if p.i < len(p.src) {
		return nil, p.fail("값 뒤에 남은 글자가 있습니다")
	}
	return n, nil
}

func (p *jparser) value(depth int) (*jnode, *jerr) {
	if depth > 64 {
		return nil, p.fail("너무 깊게 겹쳤습니다")
	}
	p.ws()
	if p.i >= len(p.src) {
		return nil, p.fail("값이 와야 하는데 파일이 끝났습니다")
	}
	at := p.i
	switch c := p.src[p.i]; {
	case c == '{':
		p.i++
		n := &jnode{kind: jObj, at: at}
		p.ws()
		if p.i < len(p.src) && p.src[p.i] == '}' {
			p.i++
			return n, nil
		}
		for {
			p.ws()
			kat := p.i
			if p.i >= len(p.src) || p.src[p.i] != '"' {
				return nil, p.fail(`키는 "따옴표" 문자열이어야 합니다`)
			}
			k, err := p.str()
			if err != nil {
				return nil, err
			}
			p.ws()
			if p.i >= len(p.src) || p.src[p.i] != ':' {
				return nil, p.fail("키 뒤에 : 가 없습니다")
			}
			p.i++
			v, err := p.value(depth + 1)
			if err != nil {
				return nil, err
			}
			for _, kv := range n.keys {
				if kv.key == k {
					return nil, &jerr{kat, fmt.Sprintf("키 %q가 두 번 있습니다", k)}
				}
			}
			n.keys = append(n.keys, jkv{key: k, at: kat, val: v})
			p.ws()
			if p.i < len(p.src) && p.src[p.i] == ',' {
				p.i++
				p.ws()
				if p.i < len(p.src) && p.src[p.i] == '}' {
					return nil, p.fail("마지막 항목 뒤에 쉼표를 둘 수 없습니다")
				}
				continue
			}
			if p.i < len(p.src) && p.src[p.i] == '}' {
				p.i++
				return n, nil
			}
			return nil, p.fail("쉼표(,)나 } 가 와야 합니다")
		}
	case c == '[':
		p.i++
		n := &jnode{kind: jArr, at: at}
		p.ws()
		if p.i < len(p.src) && p.src[p.i] == ']' {
			p.i++
			return n, nil
		}
		for {
			v, err := p.value(depth + 1)
			if err != nil {
				return nil, err
			}
			n.items = append(n.items, v)
			p.ws()
			if p.i < len(p.src) && p.src[p.i] == ',' {
				p.i++
				p.ws()
				if p.i < len(p.src) && p.src[p.i] == ']' {
					return nil, p.fail("마지막 항목 뒤에 쉼표를 둘 수 없습니다")
				}
				continue
			}
			if p.i < len(p.src) && p.src[p.i] == ']' {
				p.i++
				return n, nil
			}
			return nil, p.fail("쉼표(,)나 ] 가 와야 합니다")
		}
	case c == '"':
		s, err := p.str()
		if err != nil {
			return nil, err
		}
		return &jnode{kind: jStr, at: at, str: s}, nil
	case c == 't' || c == 'f' || c == 'n':
		for _, w := range []string{"true", "false", "null"} {
			if strings.HasPrefix(string(p.src[p.i:min(len(p.src), p.i+5)]), w) {
				p.i += len(w)
				if w == "null" {
					return &jnode{kind: jNull, at: at}, nil
				}
				return &jnode{kind: jBool, at: at, str: w}, nil
			}
		}
		return nil, p.fail("알 수 없는 값입니다(true · false · null?)")
	case c == '-' || (c >= '0' && c <= '9'):
		j := p.i
		for j < len(p.src) && strings.IndexByte("+-0123456789.eE", p.src[j]) >= 0 {
			j++
		}
		if _, err := strconv.ParseFloat(string(p.src[p.i:j]), 64); err != nil {
			return nil, p.fail("숫자 형식이 틀렸습니다")
		}
		p.i = j
		return &jnode{kind: jNum, at: at, str: string(p.src[at:j])}, nil
	default:
		return nil, p.fail("값이 와야 합니다(문자열은 \"따옴표\")")
	}
}

// str: p.i = 여는 따옴표
func (p *jparser) str() (string, *jerr) {
	start := p.i
	p.i++
	for p.i < len(p.src) {
		c := p.src[p.i]
		switch {
		case c == '"':
			p.i++
			var s string
			if err := json.Unmarshal(p.src[start:p.i], &s); err != nil {
				return "", &jerr{start, "문자열 안 이스케이프가 틀렸습니다"}
			}
			return s, nil
		case c == '\\':
			p.i += 2
		case c == '\n':
			return "", p.fail("문자열이 닫히지 않았습니다(따옴표 \" 빠짐)")
		case c < 0x20:
			return "", p.fail("문자열 안에 제어 문자가 있습니다")
		default:
			p.i++
		}
	}
	return "", &jerr{start, "문자열이 닫히지 않았습니다(따옴표 \" 빠짐)"}
}
