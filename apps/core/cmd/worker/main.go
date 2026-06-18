// Command worker는 supervisor 서버에 접속하는 에이전트다.
//
// 등록 흐름:
//   - 저장된 서브키가 없으면 → 고유키만 보내 최초 등록, 받은 서브키를 파일에 저장
//   - 저장된 서브키가 있으면 → 고유키+서브키를 보내 재접속(재부팅 시나리오)
//
// 같은 고유키로 여러 worker를 띄워보려면 -store 를 다르게 주면 된다:
//
//	go run ./cmd/worker -key worker-A -store /tmp/w1.subkey
//	go run ./cmd/worker -key worker-A -store /tmp/w2.subkey
package main

import (
	_ "embed"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"nexus/cmd/worker/constants"
	"nexus/cmd/worker/router"
	"nexus/internal/worker/db/migrations"
	"nexus/internal/worker/store"
)

//go:embed env
var envContent string

func init() {
	initEnv()
	env := constants.GetEnv()

	mountStore(env.Sqlite)

}

func main() {
	env := constants.GetEnv()

	u := url.URL{Scheme: env.WsScheme, Host: env.WsHost, Path: "/agent"}
	_, errChan := router.NewWorkerRouter(u, env.Name, store.GetStorePool())

	err := <-errChan
	log.Printf("접속 종료 %v", err)
}

// mountStore는 worker SQLite를 열고(PRAGMA 옵션) goose 마이그레이션을 적용한다.
// 마이그레이션 적용은 InitStorePool과 분리된 별도 단계(GetStorePool().Migrate).
func mountStore(dbPath string) {
	if err := store.InitStorePool(dbPath, map[string]string{
		"journal_mode": "WAL",
		"foreign_keys": "1",
		"busy_timeout": "5000",
	}); err != nil {
		log.Fatalf("worker: DB 연결 실패 %v", err)
	}
	if err := store.GetStorePool().Migrate(migrations.FS, "."); err != nil {
		log.Fatalf("worker: 마이그레이션 실패 %v", err)
	}
	log.Printf("worker: DB 준비됨 (%s)", dbPath)
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
