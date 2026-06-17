// Command supervisor는 다수의 worker agent를 관리하는 총괄 서버 진입점이다.
package main

import (
	_ "embed"
	"log"
	"nexus/cmd/supervisor/constants"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed env
var envContent string

func init() {
	initEnv()

}

func main() {
	log.Println("nexus supervisor: starting (scaffold)")
	// TODO: PostgreSQL 연결 (pgx/v5) + goose 마이그레이션
	// TODO: transport 서버 기동 (worker WebSocket 수신)
	// TODO: registry / bind / subscribe 매니저 초기화
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
