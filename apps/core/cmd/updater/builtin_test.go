package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nexus/internal/updproto"
)

// 보완-1(자기 폴더 보호): TARGET_DIR 안에 updater 폴더가 있을 때 빌드의 updater/... 파일은 아무것도 안 쓰고 실패
func TestCopyRefusesUpdaterDir(t *testing.T) {
	target := t.TempDir()
	udir := filepath.Join(target, "updater")
	stage := filepath.Join(udir, "staging", "v1-x")
	for _, p := range []string{filepath.Join(stage, "app.conf"), filepath.Join(stage, "updater", "updater.json")} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("new"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := &configFile{path: filepath.Join(udir, "updater.json"), c: config{TargetDir: target}}
	a := newAgent(udir, cfg)
	logf, _ := os.CreateTemp(t.TempDir(), "log")
	defer logf.Close()
	j := &job{a: a, ctx: context.Background(), stage: stage, logf: logf, run: updproto.Run{
		Files: []updproto.StageFile{{Path: "app.conf", Mode: 0o644}, {Path: "updater/updater.json", Mode: 0o644}},
	}}
	err := j.copyFiles()
	if err == nil || !strings.Contains(err.Error(), "updater가 쓰는 파일") {
		t.Fatalf("err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "app.conf")); !os.IsNotExist(err) {
		t.Fatalf("앞 파일이 써짐: %v", err)
	}
	if err := j.guardSelf(target); err != nil {
		t.Fatalf("TARGET_DIR 자체는 허용: %v", err)
	}
	for _, p := range []string{filepath.Join(udir, "target", "app.conf"), filepath.Join(udir, "logs", "1.log"), filepath.Join(udir, "state.json")} {
		err := j.guardSelf(p)
		if want := !strings.Contains(p, "target"); (err != nil) != want {
			t.Errorf("guardSelf(%s) = %v", p, err)
		}
	}
}

// 단위-다(갈아치우기): backup -> copy가 web/를 비우고 채움(conf/는 합치기) -> restore가 지운 파일까지 되돌림
func TestReplaceRoundTrip(t *testing.T) {
	root := t.TempDir()
	target, udir := filepath.Join(root, "target"), filepath.Join(root, "updater")
	stage := filepath.Join(udir, "staging", "v2")
	put := func(dir, rel, body string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	read := func(rel string) string {
		b, err := os.ReadFile(filepath.Join(target, filepath.FromSlash(rel)))
		if err != nil {
			return "<없음>"
		}
		return string(b)
	}
	for rel, body := range map[string]string{"web/index.html": "old", "web/old.js": "old", "web/sub/x.txt": "old", "web/app.exe.old-20260101-010101": "old", "conf/a.env": "old", "conf/b.env": "old"} {
		put(target, rel, body)
	}
	files := []updproto.StageFile{}
	for _, rel := range []string{"web/index.html", "web/new.js", "conf/a.env"} {
		put(stage, rel, "new")
		files = append(files, updproto.StageFile{Path: rel, Mode: 0o644})
	}
	cfg := &configFile{path: filepath.Join(udir, "updater.json"), c: config{TargetDir: target}}
	a := newAgent(udir, cfg)
	logf, _ := os.CreateTemp(t.TempDir(), "log")
	defer logf.Close()
	j := &job{a: a, ctx: context.Background(), stage: stage, logf: logf, run: updproto.Run{DeployID: "d1", Files: files, Replace: []string{"web"}}}
	for _, step := range []string{"backup", "copy"} {
		if err := j.builtin(step); err != nil {
			t.Fatalf("%s: %v", step, err)
		}
	}
	for rel, want := range map[string]string{"web/index.html": "new", "web/new.js": "new", "web/old.js": "<없음>", "web/sub/x.txt": "<없음>", "web/app.exe.old-20260101-010101": "old", "conf/a.env": "new", "conf/b.env": "old"} {
		if got := read(rel); got != want {
			t.Errorf("copy 뒤 %s = %s, want %s", rel, got, want)
		}
	}
	if err := j.builtin("restore"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	for rel, want := range map[string]string{"web/index.html": "old", "web/new.js": "<없음>", "web/old.js": "old", "web/sub/x.txt": "old", "conf/a.env": "old", "conf/b.env": "old"} {
		if got := read(rel); got != want {
			t.Errorf("restore 뒤 %s = %s, want %s", rel, got, want)
		}
	}
}

// 결정-13(안전장치): 설치 경로 밖 · updater 폴더를 품은 폴더는 아무것도 안 지우고 실패
func TestReplaceRefuses(t *testing.T) {
	target := t.TempDir()
	udir := filepath.Join(target, "tools", "updater")
	if err := os.MkdirAll(udir, 0o755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(target, "tools", "x.txt")
	if err := os.WriteFile(keep, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &configFile{path: filepath.Join(udir, "updater.json"), c: config{TargetDir: target}}
	a := newAgent(udir, cfg)
	logf, _ := os.CreateTemp(t.TempDir(), "log")
	defer logf.Close()
	for _, r := range []string{"../x", "tools", "tools/updater/backups"} {
		j := &job{a: a, ctx: context.Background(), stage: t.TempDir(), logf: logf, run: updproto.Run{Replace: []string{r}}}
		if err := j.copyFiles(); err == nil {
			t.Errorf("replace %q 통과", r)
		}
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("거절인데 지워짐: %v", err)
	}
}
