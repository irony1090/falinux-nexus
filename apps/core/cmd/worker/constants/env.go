package constants

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	Name        string `iniName:"NAME"`
	WsHost      string `iniName:"WS_HOST"`
	WsScheme    string `iniName:"WS_SCHEME"`
	Sqlite      string `iniName:"SQLITE"`
	Editor      string `iniName:"EDITOR"`
	ProcessRoot string
	ProjectDir  string
	Ip          string
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

	name := os.Getenv(("NAME"))
	if name == "" {
		if isNil {
			return nil, fmt.Errorf("환경변수 NAME 값 필수")
		} else {
			name = env.Name
		}
	}

	wsHost := os.Getenv(("WS_HOST"))
	if wsHost == "" {
		if isNil {
			return nil, fmt.Errorf("환경변수 WS_HOST 값 필수")
		} else {
			wsHost = env.WsHost
		}
	}

	wsScheme := os.Getenv(("WS_SCHEME"))
	if wsScheme == "" {
		if isNil {
			return nil, fmt.Errorf("환경변수 WS_SCHEME 값 필수")
		} else {
			wsScheme = env.WsScheme
		}
	}

	sqlite := os.Getenv(("SQLITE"))
	if wsScheme == "" {
		if isNil {
			return nil, fmt.Errorf("환경변수 SQLITE 값 필수")
		} else {
			wsScheme = env.WsScheme
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

	// EDITOR는 선택값 — 비면 worker의 resolveEditor가 $VISUAL>$EDITOR>vi로 폴백한다.
	editor := os.Getenv(("EDITOR"))
	if editor == "" && !isNil {
		editor = env.Editor
	}

	if isNil {
		// 환경 변수 값으로 구조체 초기화
		env = &EnvVars{
			IsDev:       IsDev(),
			ProcessRoot: os.TempDir(),
		}
	}
	env.Name = name
	env.WsHost = wsHost
	env.WsScheme = wsScheme
	env.Sqlite = pathToAbsolutePath(projectDir, sqlite)
	// env.StaticDir = pathToAbsolutePath(projectDir, staticDir)
	env.ProjectDir = projectDir
	env.Ip = ip
	env.Editor = editor

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
