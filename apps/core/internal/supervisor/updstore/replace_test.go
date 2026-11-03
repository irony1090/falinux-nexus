package updstore

import (
	"strings"
	"testing"
)

func TestCheckReplace(t *testing.T) {
	st, is := Check([]byte(`{"replace": ["web", "web/assets", "conf\\x"], "deploy": [{"id": "c", "builtin": "copy"}]}`))
	if len(st.Replace) != 3 || st.Replace[2].Path != "conf/x" {
		t.Fatalf("replace = %+v", st.Replace)
	}
	if len(is) != 1 || is[0].Level != "warn" || !strings.Contains(is[0].Msg, "web/assets") {
		t.Fatalf("부모·자식 경고 %+v", is)
	}
	for _, bad := range []string{`"replace": "web"`, `"replace": ["../x"]`, `"replace": ["/etc"]`, `"replace": [""]`, `"replace": ["a", "a/"]`, `"replace": [1]`} {
		_, is := Check([]byte(`{` + bad + `, "deploy": [{"id": "c", "builtin": "copy"}]}`))
		if !hasLevel(is, "error") {
			t.Errorf("%s 통과 %+v", bad, is)
		}
	}
}

func TestEffectiveReplace(t *testing.T) {
	rs := []Replace{{Path: "web"}, {Path: "web/assets"}, {Path: "gone"}, {Path: "conf"}}
	got := EffectiveReplace(rs, []string{"web/index.html", "web/assets/a.js", "conf/a.env", "steps.json", "gone"})
	if strings.Join(got, ",") != "web,conf" {
		t.Fatalf("got %v", got)
	}
}

func TestCheckRefsReplace(t *testing.T) {
	src := []byte("{\n  \"replace\": [\"web\", \"gone\"],\n  \"deploy\": [{\"id\": \"c\", \"builtin\": \"copy\"}]\n}")
	st, _ := Check(src)
	is := CheckRefs(src, st, nil, nil, []string{"web/a"})
	if len(is) != 1 || is[0].Line != 2 || !strings.Contains(is[0].Msg, "gone/") {
		t.Fatalf("got %+v", is)
	}
}

func TestSetReplace(t *testing.T) {
	pretty := "{\n  \"uses\": [\"w@1\"],\n  \"deploy\": [\n    { \"id\": \"c\", \"builtin\": \"copy\" }\n  ]\n}\n"
	cases := []struct {
		name, src string
		dirs      []string
		want      string
	}{
		{"uses 다음에 넣기", pretty, []string{"web"}, "{\n  \"uses\": [\"w@1\"],\n  \"replace\": [\"web\"],\n  \"deploy\": [\n    { \"id\": \"c\", \"builtin\": \"copy\" }\n  ]\n}\n"},
		{"값만 바꾸기", "{\n  \"replace\": [\"a\"],\n  \"deploy\": []\n}", []string{"a", "b"}, "{\n  \"replace\": [\"a\", \"b\"],\n  \"deploy\": []\n}"},
		{"빈 목록 = 항목째 지움", "{\n  \"uses\": [],\n  \"replace\": [\"a\"],\n  \"deploy\": []\n}", nil, "{\n  \"uses\": [],\n  \"deploy\": []\n}"},
		{"마지막 키 지움", "{\n  \"deploy\": [],\n  \"replace\": [\"a\"]\n}", nil, "{\n  \"deploy\": []\n}"},
		{"uses 없음 = 맨 앞", "{\n  \"deploy\": []\n}", []string{"a"}, "{\n  \"replace\": [\"a\"],\n  \"deploy\": []\n}"},
		{"uses가 마지막", "{\n  \"deploy\": [],\n  \"uses\": []\n}", []string{"a"}, "{\n  \"deploy\": [],\n  \"uses\": [],\n  \"replace\": [\"a\"]\n}"},
		{"한 줄 JSON", `{"uses": [], "deploy": []}`, []string{"a"}, `{"uses": [], "replace": ["a"], "deploy": []}`},
		{"CRLF", "{\r\n  \"deploy\": []\r\n}", []string{"a"}, "{\r\n  \"replace\": [\"a\"],\r\n  \"deploy\": []\r\n}"},
		{"없는데 빈 목록 = 그대로", pretty, nil, pretty},
	}
	for _, c := range cases {
		got, err := SetReplace([]byte(c.src), c.dirs)
		if err != nil || string(got) != c.want {
			t.Errorf("%s: err=%v\n got %q\nwant %q", c.name, err, got, c.want)
		}
	}
	if _, err := SetReplace([]byte("{\n  \"deploy\": [\n}"), []string{"a"}); err == nil || !strings.Contains(err.Error(), "3줄") {
		t.Errorf("문법 오류 err = %v", err)
	}
}
