package router

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"nexus/internal/protocol"

	"golang.org/x/sys/windows"
)

// topListing: 최상위 = 드라이브 목록(Q6(처음 여는 위치)). path·parent 빈 값
func topListing() (*dirListing, error) {
	mask, err := windows.GetLogicalDrives()
	if err != nil {
		return nil, err
	}
	var items []dirItem
	for i := range 26 {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		d := string(rune('A'+i)) + `:\`
		items = append(items, dirItem{entry: protocol.DirEntry{Name: d, Path: d, Kind: protocol.DirKindDir}})
	}
	return &dirListing{crumbs: []protocol.DirCrumb{{}}, items: items}, nil
}

// rootParent: 드라이브 루트(`C:\`)의 위 = 드라이브 목록
func rootParent() *string {
	s := ""
	return &s
}

// execOK: 실행기 확장자(PATHEXT + .ps1)가 있는 일반 파일 — Q24(실행 대상). Windows는 실행 비트가 없다
func execOK(name string, m fs.FileMode) bool {
	if !m.IsRegular() {
		return false
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return false
	}
	if ext == ".ps1" {
		return true
	}
	exts := os.Getenv("PATHEXT")
	if exts == "" {
		exts = ".COM;.EXE;.BAT;.CMD"
	}
	for _, e := range strings.Split(strings.ToLower(exts), ";") {
		if e == ext {
			return true
		}
	}
	return false
}

// directRunner: .ps1 = powershell / .exe·.com = 그대로 / 그 외 = cmd /c(파일 연결) — prepareExec(WW1(실행 규칙))와 같은 규칙, 형식 변환만 없음
func directRunner(body protocol.ProcessSpec) protocol.ProcessSpec {
	switch strings.ToLower(filepath.Ext(body.Cmd)) {
	case ".exe", ".com":
		return body
	}
	return scriptRunner(body, body.Cmd)
}
