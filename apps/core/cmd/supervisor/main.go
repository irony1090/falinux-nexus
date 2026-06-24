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
)

//go:embed env
var envContent string

func init() {
	initEnv()

}
func main() {
	env := constants.GetEnv()
	srv, _ := router.NewSupervisorRouter(fmt.Sprintf("localhost:%d", env.Port), env.AgentPath)
	log.Fatal(srv.ListenAndServe())
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
