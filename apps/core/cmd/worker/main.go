// Command worker는 supervisor 서버에 접속하는 에이전트다.

package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
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
	"nexus/internal/execute/hold"
	"nexus/internal/util"
	"nexus/internal/worker/db/migrations"
	"nexus/internal/worker/store"
)

//go:embed env
var envContent string

// 빌드본은 실행 파일 옆 env를 읽는다(-env로 바꿈). 개발은 -env dev.env 필수
var envFile = flag.String("env", "", "env 파일 (개발 = 필수·상대 경로는 소스 폴더 / 빌드본 = 기본 env·상대 경로는 실행 파일 폴더)")

func init() {
	if holdSock() != "" {
		return // 보관 프로세스는 env·sqlite를 쓰지 않는다
	}
	initEnv()
	env := constants.GetEnv()

	mountStore(env.Sqlite)

}

// holdSock은 `worker -hold <소켓>`(보관 프로세스 모드)이면 소켓 경로 — worker가 process마다 스스로 띄운다(REF-process-keep.md)
func holdSock() string {
	if len(os.Args) == 3 && os.Args[1] == "-hold" {
		return os.Args[2]
	}
	return ""
}

func main() {
	if sock := holdSock(); sock != "" {
		hold.Main(sock)
		return
	}
	env := constants.GetEnv()

	u := url.URL{Scheme: env.WsScheme, Host: env.WsHost, Path: "/worker"}
	backoff := util.NewBackoff(0, time.Second*2, time.Second*60)
	holdDir, err := router.HoldDir(env.Name)
	if err != nil {
		log.Fatalf("worker: 보관 프로세스 폴더 %v", err)
	}
	state := router.NewWorkerState(store.GetStorePool(), holdDir) // 재접속 루프 동안 procs·live conn 유지(재바인딩 기반)
	state.Recover()
	go handleSignals()
	for {
		log.Printf("접속 시도 worker -> supervisor")
		done, err := router.NewWorkerRouter(
			u,
			env.Name,
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

// handleSignals는 종료 신호에 worker만 끝낸다 — process는 보관 프로세스에 남는다(K4(기본 종료 동작), REF-process-keep.md).
// SIGHUP을 등록하는 것 자체가 안 나(SIGHUP 복원): nohup 등으로 무시 상태로 시작했어도 Go가 핸들러를
// 설치해 무시가 풀리고, exec되는 보관 프로세스는 기본 처분으로 시작한다(무시는 exec로 상속됨).
// hupIgnored는 Notify 전에 읽어야 한다(등록 뒤엔 무시가 풀려 항상 false).
func handleSignals() {
	hupIgnored := signal.Ignored(syscall.SIGHUP)
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	for sig := range sigs {
		if sig == syscall.SIGHUP && hupIgnored { // nohup으로 띄운 의도 유지
			log.Printf("SIGHUP 받음, nohup 실행이라 무시하고 계속 실행")
			continue
		}
		log.Printf("%v 받음, worker 종료(process는 보관 프로세스에서 계속)", sig)
		os.Exit(0)
	}
}

// mountStore는 worker SQLite를 열고(PRAGMA 옵션) goose 마이그레이션을 적용한다.
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

// embed env = 배포 값 -> 빌드본에서 파일이 없을 때 내뱉는 용도로만 쓴다 (REF-build.md "env 내뱉기 계획")
func initEnv() {
	projectDir := getProjectRoot()
	flag.Parse()
	envPath := *envFile
	if constants.IsDev() {
		// 개발은 -env 필수: 빼먹으면 embed(배포 값)로 떠서 빌드본 supervisor(8013)에 같은 메인키로 붙는다
		if envPath == "" {
			panic("개발 실행은 -env 필수 (예: go run ./cmd/worker -env dev.env)")
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
		log.Printf("worker: env 파일이 없어 배포 기본값으로 만듦 -> %s (값 확인 후 다시 실행)", envPath)
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
