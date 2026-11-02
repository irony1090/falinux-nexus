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
