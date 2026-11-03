//go:build !windows

package main

import "os"

// Linux: 실행 중 파일도 이름 변경·삭제가 된다(옛 프로세스는 옛 inode) — 물리는 일 없음
func replaceFile(tmp, dst string) (string, error) { return "", os.Rename(tmp, dst) }

func removeFile(p string) (string, error) { return "", os.Remove(p) }
