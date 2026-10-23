package router

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 전송 목록 — 폴더-8(전송 상태), REF-transfer-ext.md. 메모리만(Q29(전송 목록 보관)), 진행은 계정 토픽으로 push
const (
	transferKeep  = 5 * time.Minute        // 끝난 항목을 목록에 남겨 두는 시간
	uploadIdle    = 10 * time.Minute       // 올리는 중 조각이 이만큼 안 오면 중단(브라우저 닫힘)
	transferSweep = 30 * time.Second       // 멈춘 올리기 검사 간격
	progressEvery = 250 * time.Millisecond // 진행률 push 간격 하한
)

type transferItem struct {
	mu        sync.Mutex
	info      protocol.TransferInfo
	userID    int64
	overwrite bool               // Q23(같은 이름 파일): 확인 창에서 덮어쓰기를 골랐을 때만
	cancel    context.CancelFunc // 보내는 중 단계의 send를 깨움(시작 전 중단도 먹게 미리 만듦)
	ctx       context.Context
	tmp       *os.File // 올리는 중 단계의 임시 파일(supervisor)
	tmpPath   string
	src       *workerReader // 내려받기: 시작 전에 연 worker 읽기
	started   bool          // 내려받기: GET이 한 번 왔음(한 번만)
	touched   time.Time     // 마지막 조각 시각(uploadIdle)
	lastPush  time.Time
}

func (it *transferItem) snapshot() protocol.TransferInfo {
	it.mu.Lock()
	defer it.mu.Unlock()
	return it.info
}

func (r *supervisorRouter) mountTransfers(e *echo.Echo) {
	g := e.Group("/transfers")
	g.GET("", r.listTransfers)
	g.DELETE("/:id", r.abortTransfer)
	g.PUT("/:id/data", r.uploadData)      // upload.go
	g.POST("/:id/finish", r.finishUpload) // upload.go
	g.GET("/:id/data", r.downloadData)    // download.go
	g.POST("/copies", r.startCopy)        // copy.go
}

// pushTransfer: 진행률은 progressEvery마다, force(상태·단계 변화)는 바로
func (r *supervisorRouter) pushTransfer(it *transferItem, force bool) {
	it.mu.Lock()
	if !force && time.Since(it.lastPush) < progressEvery {
		it.mu.Unlock()
		return
	}
	it.lastPush = time.Now()
	info, userID := it.info, it.userID
	it.mu.Unlock()
	r.publishUser(userID, protocol.MsgTransferUpdate, info)
}

// endTransfer: 한 번만 끝낸다(먼저 끝낸 쪽이 이김). 임시 파일 정리 + transferKeep 뒤 목록에서 뺌
func (r *supervisorRouter) endTransfer(it *transferItem, state, code, reason string) bool {
	it.mu.Lock()
	if it.info.State != protocol.TransferRunning {
		it.mu.Unlock()
		return false
	}
	it.info.State, it.info.Code, it.info.Reason = state, code, reason
	if it.tmp != nil {
		it.tmp.Close()
		it.tmp = nil
	}
	tmpPath := it.tmpPath
	pending := it.src != nil && !it.started // 시작한 내려받기는 GET 핸들러가 닫음
	it.mu.Unlock()

	if pending {
		it.src.Close()
	}

	it.cancel()
	if tmpPath != "" {
		os.Remove(tmpPath)
	}
	r.pushTransfer(it, true)
	time.AfterFunc(transferKeep, func() {
		r.transfers.Remove(it.info.ID)
		r.publishUser(it.userID, protocol.MsgTransferGone, protocol.TransferGoneEvent{ID: it.info.ID})
	})
	return true
}

// requireTransfer: 내 항목만(남의 것 = 없는 것과 같게 404)
func (r *supervisorRouter) requireTransfer(c echo.Context, userID int64) *transferItem {
	it, ok := r.transfers.Get(c.Param("id"))
	if !ok || it.userID != userID {
		panic(web.Err(404, "전송을 찾을 수 없습니다"))
	}
	return it
}

// GET /transfers: 내 전송 목록(시작 순)
func (r *supervisorRouter) listTransfers(c echo.Context) error {
	sess := r.requireSession(c)
	list := []protocol.TransferInfo{}
	for _, f := range r.transfers.FindAll(func(_ string, it *transferItem) bool { return it.userID == sess.Data.ID }) {
		list = append(list, f.Val.snapshot())
	}
	sort.Slice(list, func(i, j int) bool { return list[i].StartedAt < list[j].StartedAt })
	return c.JSON(200, list)
}

// DELETE /transfers/:id: 중단. 올리는 중 = 임시 파일 버림 / 보내는 중 = send 취소 + worker .part 정리(best-effort)
func (r *supervisorRouter) abortTransfer(c echo.Context) error {
	sess := r.requireSession(c)
	it := r.requireTransfer(c, sess.Data.ID)
	phase := it.snapshot().Phase
	if r.endTransfer(it, protocol.TransferAborted, "", "사용자가 중단함") && phase == protocol.TransferPhaseSend {
		if err := r.AbortFile(it.info.ID, "사용자 중단"); err != nil {
			log.Printf("[transfer] %s 중단 통보 생략: %v", it.info.ID, err)
		}
	}
	return c.NoContent(204)
}

// runTransferSweep: 조각이 끊긴 올리기를 중단으로 정리
func (r *supervisorRouter) runTransferSweep() {
	for range time.Tick(transferSweep) {
		for _, f := range r.transfers.FindAll(func(_ string, it *transferItem) bool {
			it.mu.Lock()
			defer it.mu.Unlock()
			return it.info.State == protocol.TransferRunning && it.info.Phase == protocol.TransferPhaseUpload && time.Since(it.touched) > uploadIdle
		}) {
			r.endTransfer(f.Val, protocol.TransferAborted, "", "올리기가 멈춤(브라우저 연결 끊김)")
		}
	}
}

// uploadDir: 실행마다 새 임시 폴더(두 빌드를 같은 장비에 띄워도 서로 안 지움). 하루 지난 옛 폴더는 시작 때 지움
func newUploadDir() string {
	base := os.TempDir()
	if olds, err := filepath.Glob(filepath.Join(base, "nexus-upload-*")); err == nil {
		for _, d := range olds {
			if st, err := os.Stat(d); err == nil && st.IsDir() && time.Since(st.ModTime()) > 24*time.Hour && strings.HasPrefix(filepath.Base(d), "nexus-upload-") {
				os.RemoveAll(d)
			}
		}
	}
	dir, err := os.MkdirTemp(base, "nexus-upload-")
	if err != nil {
		log.Fatalf("[transfer] 임시 폴더 생성 실패: %v", err)
	}
	return dir
}
