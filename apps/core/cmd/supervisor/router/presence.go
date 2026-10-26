package router

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 접속 상태(presence) — SH6(presence 범위·전달)·SH7(브라우저 이름), REF-process-share.md. REST 조회만(push 없음).
// 탭 id·sid 원문은 다른 계정에 내보내지 않는다(탭 id = 크기 우선권 인가 열쇠, sid = 세션 쿠키 값).

func (r *supervisorRouter) mountPresence(e *echo.Group) {
	e.GET("/processes/:processId/presence", r.processPresence)
	e.GET("/users/presence", r.usersPresence)
}

type presenceTab struct {
	Connected bool  `json:"connected"` // false = 끊긴 뒤 60초 보관 중
	Viewing   bool  `json:"viewing"`   // 이 터미널이 있는 작업 공간을 띄운 탭
	SizeOwner bool  `json:"sizeOwner"` // PTY 크기 우선권을 가진 탭
	Since     int64 `json:"since"`     // 처음 연결된 시각(unix sec)
}

type presenceSession struct {
	Key       string        `json:"key"`       // sid의 해시 앞 8자 — 화면 구분용
	UserAgent string        `json:"userAgent"` // 원문, 브라우저·OS 추출은 프론트
	Tabs      []presenceTab `json:"tabs"`
}

type presenceAccount struct {
	UserID     int64             `json:"userId"`
	Nickname   string            `json:"nickname"`
	Permission int16             `json:"permission"`
	Online     bool              `json:"online"`
	LastSeen   *int64            `json:"lastSeen,omitempty"` // 접속 중이 아닐 때만, 기록 없으면 생략
	Sessions   []presenceSession `json:"sessions"`
}

func sidKey(sid string) string {
	sum := sha256.Sum256([]byte(sid))
	return hex.EncodeToString(sum[:])[:8]
}

// processPresence는 process를 볼 수 있는(보기 이상) 계정이면 그 process를 볼 수 있는 계정 전부의 접속 상태를 반환한다.
func (r *supervisorRouter) processPresence(c echo.Context) error {
	sess := r.requireSession(c)
	uid := c.Param("processId")
	checkPerm(r.processPerm(uid, sess.Data.ID), superdb.PermView, "존재하지 않는 process입니다")
	q, ctx := TxQueries(c), c.Request().Context()

	viewers, err := q.ListProcessViewers(ctx, uid)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	ids := make([]int64, len(viewers))
	for i, v := range viewers {
		ids[i] = v.UserID
	}
	nicks := r.nicknames(ctx, c, ids)
	trees, err := q.ListTreesWithTerminal(ctx, uid)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	ownerTab := r.sizeOwnerOf(uid)

	out := make([]presenceAccount, 0, len(viewers))
	for _, v := range viewers {
		acc := presenceAccount{UserID: v.UserID, Nickname: nicks[v.UserID], Permission: v.Permission, Sessions: []presenceSession{}}
		acc.Online, acc.LastSeen = r.accountPresence(v.UserID)
		idx := map[string]int{}
		for _, ts := range r.tabs.tabStatusesOfUser(v.UserID) {
			key := sidKey(ts.tab.sid)
			i, ok := idx[key]
			if !ok {
				i = len(acc.Sessions)
				idx[key] = i
				acc.Sessions = append(acc.Sessions, presenceSession{Key: key, UserAgent: ts.tab.userAgent, Tabs: []presenceTab{}})
			}
			acc.Sessions[i].Tabs = append(acc.Sessions[i].Tabs, presenceTab{
				Connected: ts.connected,
				Viewing:   ts.connected && tabViews(ts.tab, trees),
				SizeOwner: ts.tab.tabID == ownerTab,
				Since:     ts.tab.connectedAt.Unix(),
			})
		}
		out = append(out, acc)
	}
	// 접속 중인 계정 먼저, 소유자 먼저, 닉네임 순
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Online != out[j].Online {
			return out[i].Online
		}
		if (out[i].Permission == superdb.PermOwner) != (out[j].Permission == superdb.PermOwner) {
			return out[i].Permission == superdb.PermOwner
		}
		return out[i].Nickname < out[j].Nickname
	})
	return c.JSON(200, out)
}

func (r *supervisorRouter) accountPresence(userID int64) (bool, *int64) {
	online, seen := r.tabs.userPresence(userID)
	if online || seen.IsZero() {
		return online, nil
	}
	t := seen.Unix()
	return false, &t
}

func (r *supervisorRouter) nicknames(ctx context.Context, c echo.Context, ids []int64) map[int64]string {
	rows, err := TxQueries(c).ListUserNicknames(ctx, ids)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	m := make(map[int64]string, len(rows))
	for _, row := range rows {
		m[row.ID] = row.Nickname
	}
	return m
}

type userPresence struct {
	UserID   int64  `json:"userId"`
	Online   bool   `json:"online"`
	LastSeen *int64 `json:"lastSeen,omitempty"`
}

// usersPresence는 `?ids=1,2,3` 중 내가 공유를 건 계정의 지금 접속 중 / 마지막 접속만 반환한다 — O4(접속 표시). 그 밖의 id는 조용히 빠진다.
func (r *supervisorRouter) usersPresence(c echo.Context) error {
	sess := r.requireSession(c)
	targets, err := TxQueries(c).ListShareTargets(c.Request().Context(), sess.Data.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	allowed := make(map[int64]bool, len(targets))
	for _, id := range targets {
		allowed[id] = true
	}
	out := []userPresence{}
	for _, s := range strings.Split(c.QueryParam("ids"), ",") {
		id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil || !allowed[id] {
			continue
		}
		p := userPresence{UserID: id}
		p.Online, p.LastSeen = r.accountPresence(id)
		out = append(out, p)
		delete(allowed, id) // 중복 id 한 번만
	}
	return c.JSON(200, out)
}
