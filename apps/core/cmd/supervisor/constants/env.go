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

var (
	separator  = string(filepath.Separator)
	currentDir = "." + separator
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
	// UploadDir string `iniName:"UPLOAD_DIR"`
	// AssetsDir string `iniName:"ASSETS_DIR"`
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

	if isNil {
		// 환경 변수 값으로 구조체 초기화
		env = &EnvVars{
			IsDev:       IsDev(),
			ProcessRoot: os.TempDir(),
		}
	}
	env.Port = port
	env.WorkerPath = workerPath
	// env.StaticDir = pathToAbsolutePath(projectDir, staticDir)
	env.ProjectDir = projectDir
	env.Ip = ip
	env.DBUser = dbUser
	env.DBPass = dbPass
	env.DBName = dbName
	env.DBHost = dbHost
	env.DBPort = dbPort

	log.Printf("root path %s", projectDir)
	log.Printf("Loaded %v", env)

	return env, nil
}

func pathToAbsolutePath(projectDir, p string) string {

	if strings.HasPrefix(p, currentDir) {
		p = p[len(currentDir):]
		return filepath.Join(projectDir, p)
	} else if filepath.IsAbs(p) {
		return p
	} else {
		return filepath.Join(projectDir, p)
	}
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
