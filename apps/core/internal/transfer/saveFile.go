package transfer

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type SaveFile struct {
	dest    *os.File
	written int64
	size    int64

	expire      time.Duration // 다음 chunk 대기 시간
	expireTimer *time.Timer
	mu          sync.Mutex // 파일 쓰기

	OnClose   func()
	onceClose sync.Once
}

var (
	expired = errors.New("만료된 객체입니다")
)

func NewSaveFile(path string, size int64, expire time.Duration, perm os.FileMode) (*SaveFile, error) {

	var written int64
	if s, err := os.Stat(path); err == nil {
		written = s.Size()
	} else if os.IsNotExist(err) {
		written = 0
	} else {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	dest, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		perm,
	)
	if err != nil {
		return nil, err
	}

	u := &SaveFile{
		dest:    dest,
		written: written,
		size:    size,
		expire:  expire,
	}

	if expire > 0 {
		u.expireTimer = time.NewTimer(expire)

		go func() {
			<-u.expireTimer.C
			u.Close()
		}()
	}

	return u, nil
}

func (s *SaveFile) Size() int64 {
	return s.size
}

func (s *SaveFile) Written() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.written
}

func (s *SaveFile) Validate() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.written == s.size {
		return true
	}
	return false
}

func (s *SaveFile) Completed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.written >= s.size {
		return true
	}
	return false
}

func (s *SaveFile) Write(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.expireTimer != nil && !s.expireTimer.Stop() {
		return expired
	}
	n, err := s.dest.Write(data)
	if err != nil {
		return err
	}
	s.written += int64(n)
	s.initChunkUnsafe()

	return nil
}

func (s *SaveFile) initChunkUnsafe() {
	if s.expireTimer == nil {
		return
	}

	s.expireTimer.Reset(s.expire)

}

func (s *SaveFile) Remove() error {
	name := s.dest.Name()
	s.Close()
	return os.Remove(name)
}
func (s *SaveFile) Close() {
	s.onceClose.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.dest.Close()
		if s.OnClose != nil {
			s.OnClose()
		}
	})
}

func IsExpired(err error) bool {
	return errors.Is(err, expired)
}
