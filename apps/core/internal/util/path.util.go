package util

import "os"

func CreateDir(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, 0755)
	}
}

func IsDir(p string) bool {
	if fi, err := os.Stat(p); err == nil {
		return fi.IsDir()
	}
	return false
}

func CreateFile(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		f, err := os.Create(p)
		if err != nil {
			panic(err)
		}
		f.Close()
	}
}
