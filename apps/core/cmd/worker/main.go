// Command worker는 supervisor 서버에 접속하는 에이전트다.

package main

import (
	_ "embed"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
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

	u := url.URL{Scheme: env.WsScheme, Host: env.WsHost, Path: "/worker"}
	backoff := util.NewBackoff(0, time.Second*2, time.Second*60)
	state := router.NewWorkerState() // 재접속 루프 동안 procs·live conn 유지(재바인딩 기반)
	go handleSignals(state)
	for {
		log.Printf("접속 시도 worker -> supervisor")
		done, err := router.NewWorkerRouter(
			u,
			env.Name,
			env.ProcessRoot,
			store.GetStorePool(),
			state,
			func() { log.Printf("등록 완료, 접속 성공") }, // register 성공 시 호출
		)
		if err != nil {
			log.Printf("접속 실패(dial) %v", err)
			time.Sleep(backoff.Next())
			continue
		}

		log.Printf("WS 연결됨, 등록 시도 중...") // dial 성공(등록 성공은 위 콜백에서 로그)
		res := <-done                    // 세션 종료까지 대기(끊김/등록실패)
		if res.Reached {
			backoff.Reset() // 정상 가동했었으면 다음 재시도는 짧게
		}
		log.Printf("세션 종료(reached=%v) %v", res.Reached, res.Err)
		time.Sleep(backoff.Next())
	}
}

// shutdownGrace는 종료 신호 뒤 SIGTERM -> SIGKILL 유예(종료-1(유예 시간), systemd TimeoutStopSec보다 짧게).
const shutdownGrace = 5 * time.Second

// handleSignals는 종료 신호로 실행 중 process를 정리한 뒤 worker를 끝낸다(REF-worker-orphan).
// SIGHUP을 등록하는 것 자체가 안 나(SIGHUP 복원): nohup 등으로 무시 상태로 시작했어도 Go가 핸들러를
// 설치해 무시가 풀리고, exec되는 스크립트는 기본 처분으로 시작한다(무시는 exec로 상속되므로 이게 없으면
// worker가 죽어도 PTY hangup의 SIGHUP을 스크립트가 무시해 고아로 남음).
// hupIgnored는 Notify 전에 읽어야 한다(등록 뒤엔 무시가 풀려 항상 false).
func handleSignals(state *router.WorkerState) {
	hupIgnored := signal.Ignored(syscall.SIGHUP)
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	closing := false
	for sig := range sigs {
		switch {
		case sig == syscall.SIGHUP && hupIgnored: // nohup으로 띄운 의도 유지
			log.Printf("SIGHUP 받음, nohup 실행이라 무시하고 계속 실행")
		case !closing:
			closing = true
			log.Printf("%v 받음, 실행 중 process 정리 후 종료(다시 보내면 즉시 강제 종료)", sig)
			go func() {
				state.Shutdown(shutdownGrace)
				os.Exit(0)
			}()
		default: // 종료-3(두 번째 신호)
			log.Printf("%v 다시 받음, 즉시 강제 종료", sig)
			state.KillAll(true)
			os.Exit(1)
		}
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
