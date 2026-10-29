// Command supervisor는 worker agent의 연결을 받는 총괄 서버다.
//
// 등록(REGISTER) 핸드셰이크:
//   - 최초 접속: worker가 고유키만 보냄 → supervisor가 새 서브키를 부여
//   - 재접속:   worker가 고유키+서브키를 보냄 → 인식 (단, 그 서브키가 이미
//     접속 중이면 차단)
package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"nexus/cmd/supervisor/constants"
	"nexus/cmd/supervisor/router"
	"nexus/internal/supervisor/db/migrations"
	"nexus/internal/supervisor/store"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed env
var envContent string

// 빌드본은 실행 파일 옆 env를 읽는다(-env로 바꿈). 개발은 -env dev.env 필수
var envFile = flag.String("env", "", "env 파일 (개발 = 필수·상대 경로는 소스 폴더 / 빌드본 = 기본 env·상대 경로는 실행 파일 폴더)")

func init() {
	initEnv()
	mountStore(constants.GetEnv())
}
func main() {
	env := constants.GetEnv()
	// 바인드 주소 = env IP(없으면 localhost). 외부 프록시(test*.falinux.com)가 받으려면 0.0.0.0
	host := env.Ip
	if host == "" {
		host = "localhost"
	}
	addr := fmt.Sprintf("%s:%d", host, env.Port)
	e, _ := router.NewSupervisorRouter(env.WorkerPath, env.SessionKey)
	if env.WebDir != "" {
		// Pre = 라우팅·tx 앞. 파일이 없으면 API로 넘기고, 404면 index.html(SPA 경로 /w/:hash 새로고침)
		// /api·worker 경로는 건너뜀 — 없는 API가 index.html(200)이 아니라 JSON 404로 나가게
		e.Pre(middleware.StaticWithConfig(middleware.StaticConfig{Root: env.WebDir, HTML5: true, Skipper: func(c echo.Context) bool {
			p := c.Request().URL.Path
			return p == router.APIPrefix || strings.HasPrefix(p, router.APIPrefix+"/") || p == env.WorkerPath
		}}))
		log.Printf("supervisor: 웹 서빙 %s", env.WebDir)
	}
	log.Fatal(e.Start(addr))
}

// mountStore는 supervisor PostgreSQL 풀을 열고 goose 마이그레이션을 적용한다.
// 마이그레이션 적용은 InitStorePool과 분리된 별도 단계(GetStorePool().Migrate).
func mountStore(env constants.EnvVars) {
	if err := store.InitStorePool(env.DBUser, env.DBPass, env.DBHost, env.DBPort, env.DBName); err != nil {
		log.Fatalf("supervisor: DB 연결 실패 %v", err)
	}
	if err := store.GetStorePool().Migrate(migrations.FS, "."); err != nil {
		log.Fatalf("supervisor: 마이그레이션 실패 %v", err)
	}
	if err := router.MigrateTileOrder(context.Background()); err != nil {
		log.Fatalf("supervisor: 타일 문서 변환 실패 %v", err)
	}
	log.Printf("supervisor: DB 준비됨 (%s@%s:%d/%s)", env.DBUser, env.DBHost, env.DBPort, env.DBName)
}

// embed env = 배포 값 -> 빌드본에서 파일이 없을 때 내뱉는 용도로만 쓴다 (REF-build.md "env 내뱉기 계획")
func initEnv() {
	projectDir := getProjectRoot()
	flag.Parse()
	envPath := *envFile
	if constants.IsDev() {
		// 개발은 -env 필수: 빼먹으면 embed(배포 값)로 떠서 개발 마이그레이션이 배포 DB(nexus_build)에 적용된다
		if envPath == "" {
			panic("개발 실행은 -env 필수 (예: go run ./cmd/supervisor -env dev.env)")
		}
	} else if envPath == "" {
		envPath = "env"
	}
	if !filepath.IsAbs(envPath) {
		envPath = filepath.Join(projectDir, envPath)
	}

	if _, err := os.Stat(envPath); errors.Is(err, os.ErrNotExist) && !constants.IsDev() {
		if err := os.WriteFile(envPath, []byte(envContent), 0o600); err != nil {
			panic(fmt.Sprintf("env 파일 만들기 실패: %v", err))
		}
		log.Printf("supervisor: env 파일이 없어 배포 기본값으로 만듦 -> %s (값 확인 후 다시 실행)", envPath)
		os.Exit(1)
	}
	if _, err := constants.LoadEnv(envPath, projectDir); err != nil {
		panic(err.Error())
	}
}

func getProjectRoot() string {
	// 개발 환경 감지: go run으로 실행되면 임시 디렉토리에서 실행됨
	ex, _ := os.Executable()
	if constants.IsDev() {
		// 개발 환경: 소스 코드 기준 경로
		_, filename, _, _ := runtime.Caller(0)
		return filepath.Dir(filename)
	}

	return filepath.Dir(ex)
}
