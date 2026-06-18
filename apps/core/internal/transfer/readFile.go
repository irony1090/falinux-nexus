package transfer

import (
	"io"
	"io/fs"
	"os"
	"sync"
	"time"
)

type ReadFile struct {
	source *os.File

	read int64
	size int64
	perm fs.FileMode
	mu   sync.Mutex // Read() 동시 호출 제어

	expire      time.Duration
	expireTimer *time.Timer

	onceClose sync.Once
	OnClose   func()
}

func NewReadFile(path string, expire time.Duration) (*ReadFile, error) {
	var size int64
	var perm fs.FileMode
	if s, err := os.Stat(path); err == nil {
		size = s.Size()
		perm = s.Mode().Perm()
	} else {
		return nil, err
	}

	source, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	r := &ReadFile{
		source: source,
		size:   size,
		perm:   perm,
		expire: expire,
	}

	if expire > 0 {
		r.expireTimer = time.NewTimer(expire)

		go func() {
			<-r.expireTimer.C
			r.Close()
		}()
	}

	return r, nil
}

func (r *ReadFile) Size() int64 {
	return r.size
}

func (r *ReadFile) Perm() fs.FileMode {
	return r.perm
}
func (r *ReadFile) Current() int64 {
	return r.read
}

func (r *ReadFile) Completed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.read >= r.size
}

func (r *ReadFile) Read(chunkSize int64) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.expireTimer != nil && !r.expireTimer.Stop() {
		return nil, expired
	}
	if r.read >= r.size {
		return nil, io.EOF
	}
	if remain := r.size - r.read; remain < chunkSize {
		chunkSize = remain
	}
	buf := make([]byte, chunkSize)
	n, err := r.source.Read(buf)
	if err != nil {
		return nil, err
	}
	r.read += int64(n)
	if r.expireTimer != nil {
		r.expireTimer.Reset(r.expire)
	}
	return buf[:n], nil
}

func (r *ReadFile) Remove() error {

	// log.Printf("[ReadFile] Remove %s", r.source.Name())
	name := r.source.Name()
	r.Close()
	return os.Remove(name)
}

func (r *ReadFile) Close() {
	r.onceClose.Do(func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.source.Close()
		if r.OnClose != nil {
			r.OnClose()
		}
	})
}
