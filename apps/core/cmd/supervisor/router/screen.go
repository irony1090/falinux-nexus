package router

import (
	"log"
	"sync"
	"time"

	"nexus/cmd/supervisor/process"
	"nexus/internal/supervisor/bind"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 화면 복원 버퍼(SNAPSHOT) — REF-process-snapshot.md

const screenCap = 256 * 1024 // S5(ring 상한)

// screens는 ProcessEntry 밖에 둔다 — worker 재접속(Rebind) 때 entry·relay가 새로 만들어져도 출력과 offset이 이어져야 해서(S3(ring 위치·수명)).
// 지우는 곳 = cleanupProcessTopic(process 완전 종료). worker 끊김(Detach)은 유지. 메모리 전용이라 supervisor 재시작 후엔 비어 있다.
type screens struct {
	mu sync.Mutex
	m  map[string]*bind.Screen // uid -> 버퍼
}

func newScreens() *screens {
	return &screens{m: map[string]*bind.Screen{}}
}

// open은 uid의 버퍼를 반환한다. 없으면 만든다(exec·Rebind 때 relay에 넘김).
func (s *screens) open(uid string) *bind.Screen {
	s.mu.Lock()
	defer s.mu.Unlock()
	sc := s.m[uid]
	if sc == nil {
		sc = bind.NewScreen(screenCap)
		s.m[uid] = sc
	}
	return sc
}

func (s *screens) get(uid string) (*bind.Screen, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sc, ok := s.m[uid]
	return sc, ok
}

func (s *screens) drop(uid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, uid)
}

// snapshotResponse: 화면 복원 스냅샷. (GET /processes/snapshot/:processId)
type snapshotResponse struct {
	Data   []byte `json:"data"`   // base64
	Off    int64  `json:"off"`    // data 끝의 누적 바이트. off > len(data)면 앞이 밀려남
	Alt    bool   `json:"alt"`    // alt screen 중
	Redraw bool   `json:"redraw"` // 다시 그리기를 시킴 -> data 대신 off 뒤 DATA로 복원. alt인데 false(worker 끊김)면 data를 그대로 씀
	Cols   int16  `json:"cols"`
	Rows   int16  `json:"rows"`
}

// snapshotProcess는 실행 중(PENDING 포함) process의 화면 복원 버퍼를 반환한다. 버퍼가 없으면(끝남·supervisor 재시작) 404.
// 소유자 확인은 DB로 한다 — PENDING은 memory entry가 없다.
func (r *supervisorRouter) snapshotProcess(c echo.Context) error {
	sess := r.requireSession(c)
	uid := c.Param("processId")

	rows, err := TxQueries(c).ListProcessesByUids(c.Request().Context(), superdb.ListProcessesByUidsParams{
		OwnerUserID: sess.Data.ID,
		Uids:        []string{uid},
	})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	sc, ok := r.screens.get(uid)
	if len(rows) == 0 || !ok {
		panic(web.Err(404, "화면 복원 버퍼가 없습니다"))
	}
	data, off, alt := sc.Snapshot()
	entry, connected := r.processManager.Get(uid)
	// 크기를 모르면(아직 resize 전 = 0) 되돌릴 값이 없어 PTY를 0x0으로 만든다 — 2026-10-02 확인에서 실제로 발생
	redraw := alt && connected && entry.Inter != nil && entry.Record.Cols > 0 && entry.Record.Rows > 0
	if redraw {
		go r.redraw(entry, sc) // 스냅샷 뒤라 다시 그리기 출력은 off 뒤로 온다
	}
	return c.JSON(200, snapshotResponse{Data: data, Off: off, Alt: alt, Redraw: redraw, Cols: rows[0].Cols, Rows: rows[0].Rows})
}

const redrawWait = 500 * time.Millisecond

// redraw는 PTY 크기를 한 칸 바꿨다 되돌려 SIGWINCH로 TUI 전체 다시 그리기를 시킨다(같은 크기 설정은 커널이 신호를 안 보냄).
// DB·크기 우선권은 건드리지 않는다 — C(ring + alt screen만 다시 그리기)
//
// 순서 의존: 바꾼 크기로 출력이 나온 뒤에(최대 redrawWait) 되돌린다. 바로 되돌리면 TUI가 신호를 처리할 때 이미 원래 크기라
// "안 바뀜"으로 보고 다시 그리지 않는다(2026-10-02 확인: 간격 0/50/200ms에서 성공/실패/성공 — 고정 대기로는 못 막음).
// 대가로 화면이 두 번(cols-1, cols) 그려진다.
func (r *supervisorRouter) redraw(entry *process.ProcessEntry, sc *bind.Screen) {
	uid := entry.Record.Uid
	cols, rows := uint16(entry.Record.Cols), uint16(entry.Record.Rows)
	tmp := cols - 1
	if cols <= 1 {
		tmp = cols + 1
	}
	if err := entry.Inter.Layout(tmp, rows); err != nil {
		log.Printf("[snapshot] redraw uid=%s: %v", uid, err)
		return
	}
	base := sc.Off() // 바꾼 뒤 기준 — 스냅샷 기준이면 그 사이 평소 갱신 출력으로 기다림이 바로 끝난다
	for deadline := time.Now().Add(redrawWait); sc.Off() == base && time.Now().Before(deadline); {
		time.Sleep(20 * time.Millisecond)
	}
	if err := entry.Inter.Layout(cols, rows); err != nil {
		log.Printf("[snapshot] redraw 되돌리기 uid=%s: %v", uid, err)
	}
}
