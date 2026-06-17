// Command worker는 장치에 배포되어 자기 자신의 process를 관리하는 에이전트 진입점이다.
package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"nexus/cmd/supervisor/constants"
	"nexus/internal/web"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

//go:embed env
var envContent string

func init() {
	initEnv()
}

func main() {
	log.Println("nexus worker: starting (scaffold)")
	// TODO: SQLite 연결 (modernc.org/sqlite, WAL) + goose 마이그레이션
	// TODO: supervisor에 WebSocket 연결 (transport)
	// TODO: process manager 초기화 (휘발성, 메모리 관리)

	// mountWsClient()
}

func mountWsClient() *web.WsClient {
	env := constants.GetEnv()
	host := env.ApiHost
	u := url.URL{Scheme: env.WsScheme, Host: host, Path: fmt.Sprintf("/agents/%s", env.Name)}

	log.Printf("[WSClient] Connecting to %s", u.String())
	header := &http.Header{}
	header.Set("MAC", env.Name)

	c := web.NewWsClient(u, header)

	c.Connect()

	go func() {
		for {
			s, err := c.Status()
			if err != nil {
				log.Printf("[WSClient-%s-ERR] Connect To %s - %v\n", s, host, err)
			} else {
				log.Printf("[WSClient-%s] Connect To %s\n", s, host)
			}

			if s.IsDone() {
				time.Sleep(5 * time.Second)
				c.Connect()
			}
		}
	}()

	return c
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
