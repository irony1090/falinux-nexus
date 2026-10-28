package constants

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// EnvVars는 자주 사용하는 환경 변수 값들을 저장하는 구조체
type EnvVars struct {
	IsDev       bool
	Port        int
	WorkerPath  string
	ProcessRoot string
	ProjectDir  string
	Ip          string
	DBUser      string
	DBPass      string
	DBName      string
	DBHost      string
	DBPort      int
	SessionKey  string // 세션 쿠키 서명 키
	KeySecret   string // 장비 API 키 암호화 비밀값 — 잃으면 저장된 키를 다시 볼 수 없음(REF-member.md 회원-Q16)
	WebDir      string // 빌드된 웹 파일 폴더(비면 서빙 안 함). 상대 경로 = 실행 파일 폴더 기준
	ViewCache   string // 이미지 사본 폴더(빈 값 = 사본 안 만듦). 키가 없으면 ./view-cache, 상대 경로 = WEB_DIR과 같음(REF-view-tile.md)
}

var env *EnvVars

func LoadEnv(envPath string, projectDir string) (*EnvVars, error) {
	isNil := env == nil
	// .env 파일 로드 (Overload: 기존 환경변수도 덮어씀)
	if err := godotenv.Overload(envPath); err != nil && isNil {
		return nil, fmt.Errorf("[%s] 파일 로드 실패: %w", envPath, err)
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		if isNil {
			port = 5050
		} else {
			port = env.Port
		}
	}

	workerPath := os.Getenv("WORKER_PATH")
	if workerPath == "" {
		if isNil {
			return nil, fmt.Errorf("환경변수 WORKER_PATH 값 필수")
		} else {
			workerPath = env.WorkerPath
		}
	}

	ip := os.Getenv(("IP"))
	if ip == "" {
		if isNil {
			// return nil, fmt.Errorf("환경변수 IP 값 필수")
		} else {
			ip = env.Ip
		}
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		if isNil {
			return nil, fmt.Errorf("환경변수 DB_USER 값 필수")
		} else {
			dbUser = env.DBUser
		}
	}

	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		if isNil {
			return nil, fmt.Errorf("환경변수 DB_PASS 값 필수")
		} else {
			dbPass = env.DBPass
		}
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		if isNil {
			return nil, fmt.Errorf("환경변수 DB_NAME 값 필수")
		} else {
			dbName = env.DBName
		}
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		if isNil {
			return nil, fmt.Errorf("환경변수 DB_HOST 값 필수")
		} else {
			dbHost = env.DBHost
		}
	}

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		if isNil {
			dbPort = 5432
		} else {
			dbPort = env.DBPort
		}
	}

	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		if isNil {
			return nil, fmt.Errorf("환경변수 SESSION_KEY 값 필수")
		} else {
			sessionKey = env.SessionKey
		}
	}

	keySecret := os.Getenv("KEY_SECRET")
	if keySecret == "" {
		if isNil {
			return nil, fmt.Errorf("환경변수 KEY_SECRET 값 필수")
		} else {
			keySecret = env.KeySecret
		}
	}

	webDir := os.Getenv("WEB_DIR")
	if webDir == "" && !isNil {
		webDir = env.WebDir
	} else if webDir != "" && !filepath.IsAbs(webDir) {
		webDir = filepath.Join(projectDir, webDir)
	}

	// 빈 값(VIEW_CACHE_DIR="") = 끔이라 키가 있는지로 가린다
	viewCache, ok := os.LookupEnv("VIEW_CACHE_DIR")
	if !ok {
		if isNil {
			viewCache = "view-cache"
		} else {
			viewCache = env.ViewCache
		}
	}
	if viewCache != "" && !filepath.IsAbs(viewCache) {
		viewCache = filepath.Join(projectDir, viewCache)
	}

	if isNil {
		// 환경 변수 값으로 구조체 초기화
		env = &EnvVars{
			IsDev:       IsDev(),
			ProcessRoot: os.TempDir(),
		}
	}
	env.Port = port
	env.WorkerPath = workerPath
	env.ProjectDir = projectDir
	env.Ip = ip
	env.DBUser = dbUser
	env.DBPass = dbPass
	env.DBName = dbName
	env.DBHost = dbHost
	env.DBPort = dbPort
	env.SessionKey = sessionKey
	env.KeySecret = keySecret
	env.WebDir = webDir
	env.ViewCache = viewCache

	log.Printf("root path %s", projectDir)
	log.Printf("Loaded %v", env)

	return env, nil
}

func LoadEnvFromString(envContent, projectDir string) (*EnvVars, error) {
	tmpFile, err := os.CreateTemp("", "env-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(envContent); err != nil {
		return nil, err
	}

	tmpFile.Close()

	return LoadEnv(tmpFile.Name(), projectDir)
}

// func getTag(s interface{}, name string) (string, bool) {

// }

func GetEnv() EnvVars {
	if env == nil {
		panic("환경 변수가 로드되지 않았습니다. LoadEnv를 먼저 호출하세요")
	}
	return *env
}

func IsDev() bool {
	ex, _ := os.Executable()
	return strings.Contains(ex, os.TempDir()) || strings.Contains(ex, "go-build")
}
