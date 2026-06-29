package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	mu   sync.Mutex // Read() 동시 호출 제어

	info os.FileInfo

	expire      time.Duration
	expireTimer *time.Timer
	done        chan struct{}

	onceClose sync.Once
	OnClose   func()

	onceHash sync.Once // Hash() 1회 계산 보장
	hashVal  string
	hashErr  error
}

func NewReadFile(path string, expire time.Duration) (*ReadFile, error) {
	var size int64
	var info os.FileInfo
	if s, err := os.Stat(path); err == nil {
		info = s
		size = s.Size()
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
		info:   info,
		expire: expire,
	}

	if expire > 0 {
		r.expireTimer = time.NewTimer(expire)
		r.done = make(chan struct{})

		go func() {
			select {
			case <-r.expireTimer.C:
				r.Close()
			case <-r.done:
			}
		}()
	}

	return r, nil
}

func (r *ReadFile) Info() os.FileInfo {
	return r.info
}

func (r *ReadFile) Size() int64 {
	return r.size
}

func (r *ReadFile) Perm() fs.FileMode {
	return r.info.Mode().Perm()
}

// SetOnClose는 Close() 시점에 1회 불릴 콜백을 건다(전송 맵에서 자기 제거 등).
// OnClose 필드는 인터페이스로 노출 못 하므로 transfer.Reader는 이 세터로 받는다.
func (r *ReadFile) SetOnClose(fn func()) { r.OnClose = fn }

// Hash는 파일 전체의 sha256(hex)을 반환한다. 결과는 1회 계산 후 캐시된다.
// 스트리밍에 쓰는 r.source 와 별개의 fd로 풀읽기하므로 읽기 커서·진행
// 카운터(r.read)·expire 타이머에 간섭하지 않는다 → 호출 순서·진행 상태와
// 무관하게 항상 같은 값을 돌려준다. (그래서 r.mu 도 불필요, sync.Once 가 캐시 동기화 전담)
func (r *ReadFile) Hash() (string, error) {
	r.onceHash.Do(func() {
		f, err := os.Open(r.source.Name())
		if err != nil {
			r.hashErr = err
			return
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			r.hashErr = err
			return
		}
		r.hashVal = hex.EncodeToString(h.Sum(nil))
	})
	return r.hashVal, r.hashErr
}
func (r *ReadFile) Current() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.read
}

// SeekTo는 이어받기를 위해 읽기 시작 위치를 offset으로 옮긴다.
// 파일 커서(*os.File)와 진행 카운터(r.read)를 함께 맞춘다 — 하나만 바꾸면
// 실제 읽는 위치와 라벨이 어긋난다. Read 루프 시작 전에 한 번만 호출할 것.
// (이름이 Seek면 io.Seeker 시그니처로 오인돼 go vet이 경고함)
func (r *ReadFile) SeekTo(offset int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if offset < 0 || offset > r.size {
		return fmt.Errorf("offset %d 범위 밖 (size=%d)", offset, r.size)
	}
	if _, err := r.source.Seek(offset, io.SeekStart); err != nil {
		return err
	}
	r.read = offset
	return nil
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
		if r.expireTimer != nil {
			r.expireTimer.Stop()
			close(r.done)
		}
		r.source.Close()
		if r.OnClose != nil {
			r.OnClose()
		}
	})
}
