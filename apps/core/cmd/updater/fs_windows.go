package main

import (
	"errors"
	"io/fs"
	"os"
	"time"
)

// 단위-3(잠긴 파일): 실행 중인 exe는 덮어쓰기·삭제가 거절되지만 이름 변경은 된다 -> 기존 것을 <이름>.old-<시각>으로 물린 뒤 씀. *.old-*는 지우지 않음(REF-update-deploy.md "함정")

func oldName(p string) string { return p + ".old-" + time.Now().Format("20060102-150405") }

// replaceFile: tmp -> dst. 물린 파일 이름을 돌려줌(없으면 "")
func replaceFile(tmp, dst string) (string, error) {
	err := os.Rename(tmp, dst)
	if err == nil {
		return "", nil
	}
	if _, serr := os.Lstat(dst); serr != nil {
		return "", err
	}
	old := oldName(dst)
	if rerr := os.Rename(dst, old); rerr != nil {
		return "", err
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Rename(old, dst)
		return "", err
	}
	return old, nil
}

// removeFile: 지우기가 거절되면 이름만 물림
func removeFile(p string) (string, error) {
	err := os.Remove(p)
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	old := oldName(p)
	if rerr := os.Rename(p, old); rerr != nil {
		return "", err
	}
	return old, nil
}
