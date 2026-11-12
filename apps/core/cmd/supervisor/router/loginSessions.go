package router

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"sync"
	"time"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"

	"github.com/jackc/pgx/v5/pgtype"
)

// 로그인 세션 — 로그인 1번 = user_sessions 1행(id = 쿠키의 _nonce), 목록에 없는 쿠키는 무효 → REF-account.md 세션-1(세션 표)

const (
	loginTTL        = 30 * 24 * time.Hour // 쿠키 MaxAge와 같게(sessionManager.go)
	loginFlushEvery = 5 * time.Minute     // last_seen_at 기록·만료 정리 주기
)

type loginSession struct {
	userID    int64
	createdAt time.Time
}

// loginSessions: requireSession이 매 요청 DB를 읽지 않게 세션 행을 메모리에 든다. 시작 때 DB에서 채운다
type loginSessions struct {
	mu   sync.Mutex
	byID map[string]loginSession
	seen map[string]bool // 마지막 기록 뒤 요청이 온 세션 — 다음 정리 때 last_seen_at
}

func loadLoginSessions() *loginSessions {
	l := &loginSessions{byID: map[string]loginSession{}, seen: map[string]bool{}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := store.GetStorePool().Queries().ListAllUserSessions(ctx)
	if err != nil {
		log.Printf("[login] 세션 목록 조회 실패: %v", err)
	}
	for _, s := range rows {
		l.byID[s.ID] = loginSession{userID: s.UserID, createdAt: s.CreatedAt.Time}
	}
	return l
}

// valid: 그 계정의 살아 있는 세션이면 true(+활동 기록)
func (l *loginSessions) valid(id string, userID int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	s, ok := l.byID[id]
	if !ok || s.userID != userID || time.Since(s.createdAt) > loginTTL {
		return false
	}
	l.seen[id] = true
	return true
}

func (l *loginSessions) add(id string, userID int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.byID[id] = loginSession{userID: userID, createdAt: time.Now()}
}

func (l *loginSessions) remove(ids ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, id := range ids {
		delete(l.byID, id)
		delete(l.seen, id)
	}
}

func (l *loginSessions) takeSeen() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	ids := make([]string, 0, len(l.seen))
	for id := range l.seen {
		ids = append(ids, id)
	}
	l.seen = map[string]bool{}
	return ids
}

// loginKey: 화면·API에 내보내는 세션 구분값(nonce 원문은 안 내보냄)
func loginKey(id string) string {
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:])[:16]
}

func (r *supervisorRouter) runLoginSweep() {
	for {
		time.Sleep(loginFlushEvery)
		r.sweepLogins()
	}
}

// sweepLogins: 활동 시각 기록 + 30일 지난 세션 삭제
func (r *supervisorRouter) sweepLogins() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	q := store.GetStorePool().Queries()
	now := time.Now()
	if ids := r.logins.takeSeen(); len(ids) > 0 {
		if err := q.TouchUserSessions(ctx, superdb.TouchUserSessionsParams{Seen: pgtype.Timestamptz{Time: now, Valid: true}, Ids: ids}); err != nil {
			log.Printf("[login] 활동 기록 실패: %v", err)
		}
	}
	expired, err := q.DeleteExpiredUserSessions(ctx, pgtype.Timestamptz{Time: now.Add(-loginTTL), Valid: true})
	if err != nil {
		log.Printf("[login] 만료 정리 실패: %v", err)
		return
	}
	r.endLogins(expired...)
}

// endLogins: 세션을 메모리에서 빼고 그 세션 소켓에 SESSION:END를 보낸 뒤 끊는다(커밋 뒤에 부른다)
func (r *supervisorRouter) endLogins(ids ...string) {
	if len(ids) == 0 {
		return
	}
	r.logins.remove(ids...)
	for _, conn := range r.tabs.connsOfLogins(ids) {
		_ = conn.Emit(protocol.MsgSessionEnd, struct{}{})
		conn.Close(errors.New("로그인 세션 종료"))
	}
}
