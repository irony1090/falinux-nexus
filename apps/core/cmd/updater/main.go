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

var (
	configPath = flag.String("c", "", "설정 파일(기본 = 실행 파일 옆 updater.json)")
	logPath    = flag.String("log", "", "로그 파일(상대 경로 = 설정 폴더 기준, 빈 값 = 표준 출력) — Windows 예약 작업엔 journal이 없어서")
)

const logMax = 5 << 20 // 켤 때 이보다 크면 <이름>.1로 넘김

// openLog: 켤 때 한 번 — 크면 하나만 넘기고 이어 씀
func openLog(dir, p string) {
	if !filepath.IsAbs(p) {
		p = filepath.Join(dir, p)
	}
	if st, err := os.Stat(p); err == nil && st.Size() > logMax {
		_ = os.Rename(p, p+".1")
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Printf("updater: 로그 파일 %v — 표준 출력에 씀", err)
		return
	}
	log.SetOutput(f)
}

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
	dir := filepath.Dir(path)
	if *logPath != "" {
		openLog(dir, *logPath)
	}
	cfg, err := loadConfig(path)
	if err != nil {
		log.Fatalf("updater: 설정 %v", err)
	}
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
