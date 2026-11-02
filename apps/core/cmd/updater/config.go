package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// 설정 파일(updater.json) — 정-5(설치 경로): targetDir는 웹에서 바꾸면 updater가 이 파일을 고쳐 쓴다. 키가 들어 있어 0600
type config struct {
	Supervisor string `json:"supervisor"` // wss://host[:port] 또는 https://… (경로는 무시)
	MainKey    string `json:"mainKey"`
	APIKey     string `json:"apiKey"`
	TargetDir  string `json:"targetDir"`
}

type configFile struct {
	path string
	mu   sync.Mutex
	c    config
}

func loadConfig(path string) (*configFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if c.Supervisor == "" || c.MainKey == "" || c.APIKey == "" {
		return nil, fmt.Errorf("%s: supervisor · mainKey · apiKey는 필수입니다", path)
	}
	if _, _, err := endpoints(c.Supervisor); err != nil {
		return nil, err
	}
	return &configFile{path: path, c: c}, nil
}

func (f *configFile) get() config {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.c
}

// setTargetDir: 절대 경로만. 다른 항목은 파일에 적힌 그대로 두려고 원문 map을 고쳐 쓴다
func (f *configFile) setTargetDir(dir string) error {
	if dir != "" && !filepath.IsAbs(dir) {
		return errors.New("설치 경로는 절대 경로여야 합니다")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	raw := map[string]any{}
	if b, err := os.ReadFile(f.path); err == nil {
		_ = json.Unmarshal(b, &raw)
	}
	raw["targetDir"] = dir
	b, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFileAtomic(f.path, append(b, '\n'), 0o600); err != nil {
		return err
	}
	f.c.TargetDir = dir
	return nil
}

// endpoints: 설정 주소 -> WS 주소 · HTTP 기준 주소
func endpoints(base string) (ws string, http string, err error) {
	u, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil || u.Host == "" {
		return "", "", fmt.Errorf("supervisor 주소 %q 형식이 아닙니다(예: wss://test3.falinux.com)", base)
	}
	switch u.Scheme {
	case "ws", "http":
		return "ws://" + u.Host, "http://" + u.Host, nil
	case "wss", "https":
		return "wss://" + u.Host, "https://" + u.Host, nil
	}
	return "", "", fmt.Errorf("supervisor 주소는 ws · wss · http · https 중 하나로 시작합니다: %q", base)
}

// state.json — 서브키(supervisor 발급). 정-3(state.json 형식) = JSON
type state struct {
	SubKey    string `json:"subKey"`
	Installed string `json:"installed,omitempty"`
}

func loadState(path string) state {
	var s state
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &s)
	}
	return s
}

func saveState(path string, s state) error {
	b, _ := json.MarshalIndent(s, "", "  ")
	return writeFileAtomic(path, append(b, '\n'), 0o644)
}

// writeFileAtomic: 같은 폴더 임시 파일에 쓰고 이름 변경(중간에 꺼져도 반쪽 파일이 남지 않게)
func writeFileAtomic(path string, b []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), path)
}
