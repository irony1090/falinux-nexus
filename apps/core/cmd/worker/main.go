// Command worker는 supervisor 서버에 접속하는 에이전트다.

package main

import (
	_ "embed"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"nexus/cmd/worker/constants"
	"nexus/cmd/worker/router"
	"nexus/internal/util"
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
	backoff := util.NewBackoff(0, time.Second*2, time.Second*60)
	for {
		log.Printf("접속 시도 worker -> supervisor")
		done, err := router.NewWorkerRouter(
			u,
			env.Name,
			env.ProcessRoot,
			store.GetStorePool(),
		)
		if err != nil {
			log.Printf("접속 실패(dial) %v", err)
			time.Sleep(backoff.Next())
			continue
		}

		res := <-done // 세션 종료까지 대기(끊김/등록실패)
		if res.Reached {
			backoff.Reset() // 정상 가동했었으면 다음 재시도는 짧게
		}
		log.Printf("세션 종료(reached=%v) %v", res.Reached, res.Err)
		time.Sleep(backoff.Next())
	}
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
