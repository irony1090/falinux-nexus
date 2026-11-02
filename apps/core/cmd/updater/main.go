// Command updater: 장비에 상주하며 supervisor가 보낸 빌드 파일을 받고 steps.json 단계를 실행한다(worker와 독립) — REF-update-deploy.md · REF-update-p2.md
//
// 폴더 = 설정 파일이 있는 곳: updater.json(설정) · state.json(서브키) · staging/ · backups/ · logs/
package main

import (
	_ "embed"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nexus/internal/util"
)

//go:embed env
var envContent string

// buildTag: env 첫 줄 "# 태그"(worker·supervisor와 같은 규칙, REF-build.md "빌드 태그")
func buildTag() string {
	first, _, _ := strings.Cut(envContent, "\n")
	return strings.TrimSpace(strings.TrimPrefix(first, "#"))
}

var configPath = flag.String("c", "", "설정 파일(기본 = 실행 파일 옆 updater.json)")

func main() {
	flag.Parse()
	path := *configPath
	if path == "" {
		exe, err := os.Executable()
		if err != nil {
			log.Fatalf("updater: 실행 파일 위치 %v", err)
		}
		path = filepath.Join(filepath.Dir(exe), "updater.json")
	}
	path, _ = filepath.Abs(path)
	cfg, err := loadConfig(path)
	if err != nil {
		log.Fatalf("updater: 설정 %v", err)
	}
	dir := filepath.Dir(path)
	for _, d := range []string{"staging", "backups", "logs"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			log.Fatalf("updater: 폴더 %v", err)
		}
	}
	c := cfg.get()
	log.Printf("updater %s 시작 — 메인키 %s · supervisor %s · 설치 경로 %q · 폴더 %s", buildTag(), c.MainKey, c.Supervisor, c.TargetDir, dir)

	a := newAgent(dir, cfg)
	backoff := util.NewBackoff(0, 2*time.Second, 60*time.Second)
	for {
		if a.session() {
			backoff.Reset()
		}
		time.Sleep(backoff.Next())
	}
}
