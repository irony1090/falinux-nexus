package util

import "sync"

type OneLine struct {
	data []byte
	mu   sync.Mutex
}

func NewOneLine(limit int) *OneLine {
	var data []byte
	if limit > 0 {

	} else {
		data = make([]byte, 0)
	}
	return &OneLine{
		data: data,
	}
}
