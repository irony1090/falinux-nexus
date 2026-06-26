// Command supervisor는 worker agent의 연결을 받는 총괄 서버다.
//
// 등록(REGISTER) 핸드셰이크:
//   - 최초 접속: worker가 고유키만 보냄 → supervisor가 새 서브키를 부여
//   - 재접속:   worker가 고유키+서브키를 보냄 → 인식 (단, 그 서브키가 이미
//     접속 중이면 차단)
package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"nexus/cmd/supervisor/constants"
	"nexus/cmd/supervisor/router"
	"nexus/internal/supervisor/db/migrations"
	"nexus/internal/supervisor/store"
)

//go:embed env
var envContent string

func init() {
	initEnv()
	mountStore(constants.GetEnv())
}
func main() {
	env := constants.GetEnv()
	addr := fmt.Sprintf("localhost:%d", env.Port)
	e, _ := router.NewSupervisorRouter(env.WorkerPath)
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
	log.Printf("supervisor: DB 준비됨 (%s@%s:%d/%s)", env.DBUser, env.DBHost, env.DBPort, env.DBName)
}

func initEnv() {
	projectDir := getProjectRoot()
	_, err := constants.LoadEnvFromString(envContent, projectDir)
	if err != nil {
		panic(err.Error())
	}

	if !constants.IsDev() {
		envPath := filepath.Join(projectDir, "env")
		_, err = constants.LoadEnv(envPath, projectDir)
		if err != nil {
			panic(err.Error())
		}
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
