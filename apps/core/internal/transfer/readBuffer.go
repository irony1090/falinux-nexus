package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"
)

// ReadBuffer는 ReadFile과 동일한 청크 송신 계약을 디스크가 아닌 메모리 바이트
// 슬라이스로 제공한다. os.File/os.FileInfo가 없으므로 이름·권한은 파일에서
// 캐내지 않고 생성 시점에 호출자가 직접 준다(EDIT seed처럼 "파일이 아닌 것"을
// 그대로 보내려는 용도). data는 불변으로 취급한다 — Read는 슬라이싱·복사만.
type ReadBuffer struct {
	data []byte

	read int64
	size int64
	mu   sync.Mutex // Read() 동시 호출 제어

	expire      time.Duration
	expireTimer *time.Timer
	done        chan struct{}

	onceClose sync.Once
	OnClose   func()

	onceHash sync.Once // Hash() 1회 계산 보장
	hashVal  string
}

// NewReadBuffer는 메모리 내용으로 reader를 만든다. ReadFile.NewReadFile과 달리
// 열 파일이 없어 실패 지점이 없으므로 error를 돌려주지 않는다.
func NewReadBuffer(data []byte, expire time.Duration) *ReadBuffer {
	r := &ReadBuffer{
		data:   data,
		size:   int64(len(data)),
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

	return r
}

func (r *ReadBuffer) Size() int64 { return r.size }

// SetOnClose는 Close() 시점에 1회 불릴 콜백을 건다(전송 맵에서 자기 제거 등).
func (r *ReadBuffer) SetOnClose(fn func()) { r.OnClose = fn }

// Hash는 내용 전체의 sha256(hex)을 반환한다. 1회 계산 후 캐시된다.
// 메모리 풀읽기라 IO 에러가 없어 error는 항상 nil이지만, ReadFile과 시그니처를
// 맞춰 드롭인 대체가 되도록 (string, error)를 유지한다.
func (r *ReadBuffer) Hash() (string, error) {
	r.onceHash.Do(func() {
		sum := sha256.Sum256(r.data)
		r.hashVal = hex.EncodeToString(sum[:])
	})
	return r.hashVal, nil
}

func (r *ReadBuffer) Current() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.read
}

// SeekTo는 이어받기를 위해 읽기 시작 위치를 offset으로 옮긴다.
// 파일 커서가 없으므로 진행 카운터(r.read)만 맞추면 된다. Read 루프 시작 전 1회만.
func (r *ReadBuffer) SeekTo(offset int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if offset < 0 || offset > r.size {
		return fmt.Errorf("offset %d 범위 밖 (size=%d)", offset, r.size)
	}
	r.read = offset
	return nil
}

func (r *ReadBuffer) Completed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.read >= r.size
}

func (r *ReadBuffer) Read(chunkSize int64) ([]byte, error) {
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
	// data는 불변이지만 호출자가 보관/변형할 수 있으니 ReadFile처럼 새 버퍼로 복사해 넘긴다.
	buf := make([]byte, chunkSize)
	n := copy(buf, r.data[r.read:r.read+chunkSize])
	r.read += int64(n)
	if r.expireTimer != nil {
		r.expireTimer.Reset(r.expire)
	}
	return buf[:n], nil
}

func (r *ReadBuffer) Close() {
	r.onceClose.Do(func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.expireTimer != nil {
			r.expireTimer.Stop()
			close(r.done)
		}
		if r.OnClose != nil {
			r.OnClose()
		}
	})
}
