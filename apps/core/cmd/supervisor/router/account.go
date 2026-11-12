package router

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
)

// 내 계정 — 중복 확인·비밀번호·닉네임·로그인한 기기 → REF-account.md 세션-2(기기 API)·세션-4(비밀번호 API)·아이디-2(가입 분리)

func (router *supervisorRouter) mountAccount(g *echo.Group) {
	g.GET("/available", router.checkAvailable)
	g.PUT("/password", router.changePassword)
	g.PUT("/nickname", router.changeNickname)
	g.GET("/sessions", router.listLogins)
	g.DELETE("/sessions", router.endAllLogins)
	g.DELETE("/sessions/:key", router.endLogin)
}

// ===== 아이디·닉네임 모양·중복 =====

const nameMax = 64 // nickname VARCHAR(64), 아이디도 같게

// loginIDProblem: 아이디 = 공백 없는 1~64자. 문제없으면 ""
func loginIDProblem(s string) string {
	switch {
	case s == "":
		return "아이디를 입력하세요"
	case utf8.RuneCountInString(s) > nameMax:
		return "아이디는 64자까지입니다"
	case strings.ContainsFunc(s, unicode.IsSpace):
		return "아이디에 공백을 넣을 수 없습니다"
	}
	return ""
}

// nicknameProblem: 닉네임 = 앞뒤 공백을 뺀 1~64자
func nicknameProblem(s string) string {
	switch {
	case s == "":
		return "닉네임을 입력하세요"
	case utf8.RuneCountInString(s) > nameMax:
		return "닉네임은 64자까지입니다"
	}
	return ""
}

func requireLoginID(s string) string {
	s = strings.TrimSpace(s)
	if p := loginIDProblem(s); p != "" {
		panic(web.Err(400, "%s", p))
	}
	return s
}

func requireNickname(s string) string {
	s = strings.TrimSpace(s)
	if p := nicknameProblem(s); p != "" {
		panic(web.Err(400, "%s", p))
	}
	return s
}

func loginIDTaken(c echo.Context, s string) bool {
	taken, err := TxQueries(c).IdentificationTaken(c.Request().Context(), s)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return taken
}

func nicknameTaken(c echo.Context, s string) bool {
	taken, err := TxQueries(c).NicknameTaken(c.Request().Context(), s)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return taken
}

func requireFreeLoginID(c echo.Context, s string) {
	if loginIDTaken(c, s) {
		panic(web.Err(400, "이미 쓰는 아이디입니다"))
	}
}

func requireFreeNickname(c echo.Context, s string) {
	if nicknameTaken(c, s) {
		panic(web.Err(400, "이미 쓰는 닉네임입니다"))
	}
}

// dbWriteErr: 중복 확인과 저장 사이에 다른 요청이 같은 값을 먼저 쓴 경우(UNIQUE 위반)는 400, 나머지는 500
func dbWriteErr(err error) web.ClientError {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		if strings.Contains(pg.ConstraintName, "nickname") {
			return web.Err(400, "이미 쓰는 닉네임입니다")
		}
		return web.Err(400, "이미 쓰는 아이디입니다")
	}
	return web.Err(500, "%v", err)
}

type availableResponse struct {
	Available bool   `json:"available"`
	Message   string `json:"message"` // 못 쓰는 이유(쓸 수 있으면 "")
}

// checkAvailable: ?identification= 또는 ?nickname= 하나 — 가입 화면 입력 중 확인(로그인 없이). 저장 때 서버가 다시 확인한다
func (router *supervisorRouter) checkAvailable(c echo.Context) error {
	if v, ok := c.QueryParams()["identification"]; ok {
		s := strings.TrimSpace(strings.Join(v, ""))
		if p := loginIDProblem(s); p != "" {
			return c.JSON(200, availableResponse{Message: p})
		}
		if loginIDTaken(c, s) {
			return c.JSON(200, availableResponse{Message: "이미 쓰는 아이디입니다"})
		}
		return c.JSON(200, availableResponse{Available: true})
	}
	s := strings.TrimSpace(c.QueryParam("nickname"))
	if p := nicknameProblem(s); p != "" {
		return c.JSON(200, availableResponse{Message: p})
	}
	if nicknameTaken(c, s) {
		return c.JSON(200, availableResponse{Message: "이미 쓰는 닉네임입니다"})
	}
	return c.JSON(200, availableResponse{Available: true})
}

// ===== 비밀번호·닉네임 =====

type passwordChangeRequest struct {
	Current string `json:"current" validate:"required"`
	Next    string `json:"next" validate:"required"`
}

type nicknameChangeRequest struct {
	Nickname string `json:"nickname"`
}

// changePassword: 현재 비밀번호 확인 -> 저장 -> 지금 기기만 남기고 나머지 세션 종료
func (router *supervisorRouter) changePassword(c echo.Context) error {
	sess := router.requireSession(c)
	var body passwordChangeRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "비밀번호를 입력하세요"))
	}
	u := requireUser(c, sess.Data.ID)
	if ok, _ := checkPassword(u.Password, body.Current); !ok {
		panic(web.Err(400, "현재 비밀번호가 다릅니다"))
	}
	router.setPassword(c, u.ID, body.Next, sessionLogin(sess))
	return c.JSON(200, map[string]any{})
}

// setPassword: 저장 + keep 말고 그 계정 세션 전부 종료(커밋 뒤). 관리자 재설정도 같이 쓴다
func (router *supervisorRouter) setPassword(c echo.Context, userID int64, pw, keep string) {
	q, ctx := TxQueries(c), c.Request().Context()
	if err := q.UpdateUserPassword(ctx, superdb.UpdateUserPasswordParams{ID: userID, Password: hashPassword(pw)}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	router.dropLogins(c, userID, keep)
}

// dropLogins: 그 계정의 세션 행을 keep 말고 지우고, 커밋 뒤 소켓을 끊는다
func (router *supervisorRouter) dropLogins(c echo.Context, userID int64, keep string) {
	ids, err := TxQueries(c).DeleteUserSessionsOfUser(c.Request().Context(), superdb.DeleteUserSessionsOfUserParams{UserID: userID, KeepID: keep})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	AfterCommit(c, func() { router.endLogins(ids...) })
}

func (router *supervisorRouter) changeNickname(c echo.Context) error {
	sess := router.requireSession(c)
	var body nicknameChangeRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	router.setNickname(c, sess.Data.ID, body.Nickname)
	return c.JSON(200, map[string]any{})
}

// setNickname: 모양·중복 확인 후 저장(자기 지금 값이면 그대로). 관리자 수정도 같이 쓴다
func (router *supervisorRouter) setNickname(c echo.Context, userID int64, raw string) {
	nick := requireNickname(raw)
	u := requireUser(c, userID)
	if u.Nickname == nick {
		return
	}
	if !strings.EqualFold(u.Nickname, nick) { // 대소문자만 바꾸는 건 자기 값이라 통과
		requireFreeNickname(c, nick)
	}
	if err := TxQueries(c).UpdateUserNickname(c.Request().Context(), superdb.UpdateUserNicknameParams{ID: userID, Nickname: nick}); err != nil {
		panic(dbWriteErr(err))
	}
}

// setLoginID: setNickname과 같은 방식(관리자 수정)
func (router *supervisorRouter) setLoginID(c echo.Context, userID int64, raw string) {
	ident := requireLoginID(raw)
	u := requireUser(c, userID)
	if u.Identification == ident {
		return
	}
	if !strings.EqualFold(u.Identification, ident) {
		requireFreeLoginID(c, ident)
	}
	if err := TxQueries(c).UpdateUserIdentification(c.Request().Context(), superdb.UpdateUserIdentificationParams{ID: userID, Identification: ident}); err != nil {
		panic(dbWriteErr(err))
	}
}

// ===== 로그인한 기기 =====

type loginResponse struct {
	Key        string `json:"key"`
	UserAgent  string `json:"userAgent"` // 원문, 브라우저·OS 풀이는 프론트(presence와 같은 함수)
	CreatedAt  int64  `json:"createdAt"`
	LastSeenAt int64  `json:"lastSeenAt"` // 5분 단위로 기록(loginFlushEvery)
	Current    bool   `json:"current"`    // 이 요청을 보낸 기기
	Online     bool   `json:"online"`     // 연결된 탭이 있음
}

func (router *supervisorRouter) listLogins(c echo.Context) error {
	sess := router.requireSession(c)
	rows, err := TxQueries(c).ListUserSessions(c.Request().Context(), sess.Data.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	me, online := sessionLogin(sess), router.tabs.onlineLogins(sess.Data.ID)
	out := make([]loginResponse, len(rows))
	for i, s := range rows {
		out[i] = loginResponse{
			Key: loginKey(s.ID), UserAgent: s.UserAgent,
			CreatedAt: s.CreatedAt.Time.Unix(), LastSeenAt: s.LastSeenAt.Time.Unix(),
			Current: s.ID == me, Online: online[s.ID],
		}
	}
	return c.JSON(200, map[string]any{"items": out})
}

// expireCookie: 이 브라우저 쿠키 만료(+ 이 sid 탭들의 크기 우선권 넘김)
func (router *supervisorRouter) expireCookie(c echo.Context) {
	sess := router.sessions.GetSession(c.Request(), c.Response())
	router.sizeOwnerOnSignOut(sess.Name())
	sess.Session.Options.MaxAge = -1
	if err := sess.Save(); err != nil {
		panic(web.Err(500, "%v", err))
	}
}

// endAllLogins: 모든 기기에서 로그아웃(이 기기 포함)
func (router *supervisorRouter) endAllLogins(c echo.Context) error {
	sess := router.requireSession(c)
	router.dropLogins(c, sess.Data.ID, "")
	router.expireCookie(c)
	return c.JSON(200, map[string]any{})
}

// endLogin: 기기 하나 로그아웃. 이 기기면 쿠키도 만료 — current로 알려 브라우저가 로그아웃 처리
func (router *supervisorRouter) endLogin(c echo.Context) error {
	sess := router.requireSession(c)
	key := c.Param("key")
	q, ctx := TxQueries(c), c.Request().Context()
	rows, err := q.ListUserSessions(ctx, sess.Data.ID)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	for _, s := range rows {
		if loginKey(s.ID) != key {
			continue
		}
		if _, err := q.DeleteUserSession(ctx, superdb.DeleteUserSessionParams{ID: s.ID, UserID: sess.Data.ID}); err != nil {
			panic(web.Err(500, "%v", err))
		}
		id := s.ID
		AfterCommit(c, func() { router.endLogins(id) })
		current := id == sessionLogin(sess)
		if current {
			router.expireCookie(c)
		}
		return c.JSON(200, map[string]any{"current": current})
	}
	panic(web.Err(404, "이미 로그아웃된 기기입니다"))
}
