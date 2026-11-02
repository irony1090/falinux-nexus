package updstore

import "testing"

func TestCheckPositions(t *testing.T) {
	src := []byte("{\n  \"uses\": [\"nexus-worker@v1\"],\n  \"deploy\": [\n    { \"id\": \"stop\", \"script\": \"00-stop.sh\" },\n    { \"id\": \"copy\", \"builtin\": \"cp\" },\n    { \"script\": \"x.sh\", \"builtin\": \"copy\" }\n  ]\n}\n")
	st, is := Check(src)
	if len(st.Uses) != 1 || len(st.Deploy) != 1 {
		t.Fatalf("uses/deploy = %+v", st)
	}
	got := map[int]bool{}
	for _, i := range is {
		if i.Level == "error" {
			got[i.Line] = true
		}
	}
	if len(got) != 2 || !got[5] || !got[6] {
		t.Errorf("오류 줄 = %v, 받은 것 %+v", got, is)
	}
}

func TestCheckSyntax(t *testing.T) {
	_, is := Check([]byte("{\n  \"deploy\": [\n    { \"id\": \"a\" \"script\": \"b\" }\n  ]\n}"))
	if len(is) != 1 || is[0].Line != 3 {
		t.Fatalf("got %+v", is)
	}
	_, is = Check([]byte("{\"deploy\": [{\"id\": \"a\\/b\", \"builtin\": \"copy\",}]}"))
	if len(is) != 1 || is[0].Line != 1 {
		t.Fatalf("trailing comma %+v", is)
	}
}

func TestCleanRel(t *testing.T) {
	for _, bad := range []string{"", "/etc/x", "../a", "a/../../b", "@sub/x", "a /b"} {
		if _, err := CleanRel(bad); err == nil {
			t.Errorf("%q 통과", bad)
		}
	}
	if c, err := CleanRel(`bin\\worker`); err != nil || c != "bin/worker" {
		t.Errorf("got %q %v", c, err)
	}
}

func TestStepFlags(t *testing.T) {
	st, is := Check([]byte(`{"deploy":[{"id":"a","script":"a.sh","interactive":true,"timeout":"5m"},{"id":"b","builtin":"copy","manual":true}]}`))
	if len(st.Deploy) != 2 || !st.Deploy[0].Interactive || st.Deploy[0].Timeout == 0 || !st.Deploy[1].Manual {
		t.Fatalf("got %+v", st.Deploy)
	}
	if len(is) != 1 || is[0].Level != "warn" {
		t.Fatalf("interactive+timeout 경고 %+v", is)
	}
}

// 기본 권한이면 P1 sha와 같아야 한다(잠금 sha 호환)
func TestFolderShaMode(t *testing.T) {
	a := FolderSha([]Entry{{Scope: "main", Path: "a.sh", Sha: "x"}})
	if b := FolderSha([]Entry{{Scope: "main", Path: "a.sh", Sha: "x", Mode: 0o755}}); a != b {
		t.Fatal("기본 권한인데 sha가 바뀜")
	}
	if c := FolderSha([]Entry{{Scope: "main", Path: "a.sh", Sha: "x", Mode: 0o700}}); a == c {
		t.Fatal("권한을 바꿨는데 sha가 같음")
	}
}
