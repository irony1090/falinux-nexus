package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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
	done        chan struct{}
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

	// O_APPEND 금지: WriteAt(offset)은 append 모드 파일에서 에러난다.
	// 위치 지정은 written/offset으로 직접 한다.
	dest, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE,
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
		u.done = make(chan struct{})

		go func() {
			select {
			case <-u.expireTimer.C:
				u.Close()
			case <-u.done:
			}
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

// Hash는 현재까지 디스크에 쓰인 파일 전체의 sha256(hex)을 반환한다.
// 송신측 FileInit.Hash 와 비교해 무결성을 검증하는 용도.
//
// ReadFile.Hash 와 달리 캐시하지 않는다 — 수신 파일은 전송 동안 계속 변하므로
// "완료(Completed) 이후 한 번" 호출해야 의미가 있다. WriteAt 의 비순차/재전송
// 쓰기·이어받기에도 결과 파일을 그대로 풀읽기하므로 정확하다(running hash 로는
// 이들을 다룰 수 없다). s.mu 를 잡아 in-flight 쓰기와의 일관성을 보장한다.
func (s *SaveFile) Hash() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Open(s.dest.Name())
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Sync는 지금까지 쓴 데이터를 디스크에 강제 반영한다(fsync).
// 완료 선언(rename) 직전에 호출해 크래시 시 버퍼 유실을 막는다.
func (s *SaveFile) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dest.Sync()
}

func (s *SaveFile) Completed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.written >= s.size {
		return true
	}
	return false
}

// Write는 현재 진행 끝(written)에 순차로 이어 쓴다. O_APPEND를 쓰지 않으므로
// OS 커서가 아니라 written을 위치로 삼는다(이어받기 시 기존 크기부터 이어짐).
func (s *SaveFile) Write(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.expireTimer != nil && !s.expireTimer.Stop() {
		return expired
	}
	n, err := s.dest.WriteAt(data, s.written)
	if err != nil {
		return err
	}
	s.written += int64(n)
	s.initChunkUnsafe()

	return nil
}

// WriteAt은 data를 offset 위치에 직접 쓴다(이어받기·멱등 쓰기용).
// 같은 구간을 재전송으로 다시 써도 안전하다. written은 최대 도달 지점
// (high-water mark)으로 갱신한다 — 순차/이어받기에선 정확하나, 구멍을 남기는
// 비순차 쓰기에선 Completed/Validate가 실제 채움보다 앞설 수 있으니 최종 hash로 검증할 것.
func (s *SaveFile) WriteAt(offset int64, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.expireTimer != nil && !s.expireTimer.Stop() {
		return expired
	}
	if offset < 0 {
		return fmt.Errorf("offset %d 음수", offset)
	}
	n, err := s.dest.WriteAt(data, offset)
	if err != nil {
		return err
	}
	if end := offset + int64(n); end > s.written {
		s.written = end
	}
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
		if s.expireTimer != nil {
			s.expireTimer.Stop()
			close(s.done)
		}
		s.dest.Close()
		if s.OnClose != nil {
			s.OnClose()
		}
	})
}

func IsExpired(err error) bool {
	return errors.Is(err, expired)
}
