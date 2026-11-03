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
