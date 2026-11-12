package router

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"slices"
	"strings"

	"nexus/internal/manager/session"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/util"
	"nexus/internal/web"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const sessionNonceKey = "_nonce" // 로그인마다 새 값 → sid(쿠키 원본값)가 로그인마다 달라짐 + user_sessions 행 id(loginSessions.go)

// sessionLogin: 쿠키의 로그인 세션 id(옛 쿠키면 "")
func sessionLogin(sess *session.SessionElement[superdb.User]) string {
	id, _ := sess.Session.Values[sessionNonceKey].(string)
	return id
}

// 구 PortBridge cmd/agent_v2 user 핸들러를 이식. echo 기본 핸들러(func(c) error)를 따르되
// 에러는 panic(web.Err(...))로 던져 PanicMiddleware가 JSON으로 응답하고,
// 트랜잭션은 txMiddleware가 깐 TxQueries(c)로 꺼낸다(롤백은 미들웨어 recover가 처리).

// 아이디·닉네임 분리(아이디-2(가입 분리)), 모양 검사 = loginIDProblem·nicknameProblem(account.go)
type userCreateRequest struct {
	Identification string `json:"identification"`
	Nickname       string `json:"nickname"`
	Password       string `json:"password" validate:"required"`
}

type userSignRequest struct {
	Identification string `json:"identification" validate:"required"`
	Password       string `json:"password" validate:"required"`
}

type userResponse struct {
	ID             int64    `json:"id"`
	Identification string   `json:"identification"`
	Nickname       string   `json:"nickname"`
	CreatedAt      int64    `json:"createdAt"`
	UpdatedAt      *int64   `json:"updatedAt"`
	Perms          []string `json:"perms"` // 주소-3(권한 열) — 화면 메뉴 보이기용, 인가는 requirePerm이 DB에서 다시 읽음
	WorkerLimit    int32    `json:"workerLimit"`
}

func newUserResponse(u superdb.User) userResponse {
	r := userResponse{
		ID:             u.ID,
		Identification: u.Identification,
		Nickname:       u.Nickname,
		Perms:          u.Perms,
		WorkerLimit:    u.WorkerLimit,
	}
	if r.Perms == nil {
		r.Perms = []string{}
	}
	if u.CreatedAt.Valid {
		r.CreatedAt = u.CreatedAt.Time.Unix()
	}
	if u.UpdatedAt.Valid {
		t := u.UpdatedAt.Time.Unix()
		r.UpdatedAt = &t
	}
	return r
}

// toHash: sha256 hex — 옛 비밀번호 대조·장비 키(deviceKey.go)에만. 아이디는 평문(아이디-1(평문 이전))
func toHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

func hashPassword(pw string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		panic(web.Err(400, "%v", err)) // 72바이트 초과
	}
	return string(h)
}

// checkPassword: bcrypt면 bcrypt로, 아니면 옛 sha256 hex로 대조. legacy = 맞았는데 옛 방식(다시 저장 대상) — 회원-Q9(bcrypt 전환)
func checkPassword(stored, pw string) (ok, legacy bool) {
	if strings.HasPrefix(stored, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(pw)) == nil, false
	}
	ok = stored == toHash(pw)
	return ok, ok
}

// requireSession은 로그인된 세션을 돌려주거나 401을 panic한다.
// 구 GetSessionOrPanic의 flag if/else 사슬을 단락 OR 한 줄로 축약했다.
func (router *supervisorRouter) requireSession(c echo.Context) *session.SessionElement[superdb.User] {
	sess := router.sessions.GetSession(c.Request(), nil)
	if sess == nil || sess.Session.IsNew || sess.Data.ID == 0 {
		panic(web.Err(401, "로그인 절차가 필요합니다"))
	}
	if router.blocked.has(sess.Data.ID) { // 쿠키는 정지·삭제 전에 받은 것일 수 있다 — 회원-Q21
		panic(web.Err(401, "사용할 수 없는 계정입니다"))
	}
	if !router.logins.valid(sessionLogin(sess), sess.Data.ID) { // 세션 행이 지워짐(다른 기기에서 로그아웃 등)·옛 쿠키
		panic(web.Err(401, "로그인 절차가 필요합니다"))
	}
	return sess
}

// 계정 권한 이름(users.perms) — 주소-3(권한 열), 회원 관리 → REF-member.md
const (
	permWorkerMove     = "worker_move"     // 전체 worker 주소 변경
	permAdmin          = "admin"           // 관리자 페이지
	permWorkerRegister = "worker_register" // 장비 이름 등록(회원-Q3)
)

// requirePerm은 계정에 perm이 있는지 DB에서 다시 읽어 본다(쿠키의 sess.Data는 로그인 때 값이라 쓰지 않음). 없으면 403
func requirePerm(c echo.Context, userID int64, perm string) {
	user, err := TxQueries(c).GetUser(c.Request().Context(), userID)
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(401, "로그인 절차가 필요합니다"))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	if !slices.Contains(user.Perms, perm) {
		panic(web.Err(403, "권한이 없습니다"))
	}
}

func (router *supervisorRouter) mountUsers(e *echo.Group) {
	g := e.Group("/users")
	g.POST("", router.createUser)
	g.POST("/session", router.signIn)
	g.GET("/session", router.checkSession)
	g.DELETE("/session", router.signOut)
	g.GET("/lookup", router.lookupUser)
	router.mountAccount(g) // account.go
}

// userLookupResponse: 공유 대상 조회 결과 — 닉네임 + 내부 id만(SH4(공유 대상 찾기))
type userLookupResponse struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
}

// lookupUser는 아이디(?identification=, 대소문자 무시)가 일치하는 계정을 찾는다. 없으면 404, 자기 자신이면 400(SH4(공유 대상 찾기)).
func (router *supervisorRouter) lookupUser(c echo.Context) error {
	sess := router.requireSession(c)
	ident := c.QueryParam("identification")
	if ident == "" {
		panic(web.Err(400, "아이디를 입력하세요"))
	}
	user, err := TxQueries(c).GetUserByIdentification(c.Request().Context(), strings.TrimSpace(ident))
	if err != nil {
		panic(web.Err(404, "계정을 찾을 수 없습니다"))
	}
	if user.ID == sess.Data.ID {
		panic(web.Err(400, "자기 자신에게는 공유할 수 없습니다"))
	}
	return c.JSON(200, userLookupResponse{ID: user.ID, Nickname: user.Nickname})
}

// createUser는 신규 계정을 만든다. 같은 트랜잭션 안에서 중복 검사 후 INSERT.
func (router *supervisorRouter) createUser(c echo.Context) error {
	var body userCreateRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}

	ident := requireLoginID(body.Identification)
	nick := requireNickname(body.Nickname)
	q := TxQueries(c)
	ctx := c.Request().Context()

	// 순서 의존: 잠금 -> 회원 수 -> INSERT 를 한 트랜잭션에서. 잠금이 없으면 동시 첫 가입 둘 다 관리자가 된다 — 회원-Q5(관리자 판정)
	if err := q.LockUserCreate(ctx); err != nil {
		panic(web.Err(500, "%v", err))
	}
	requireFreeLoginID(c, ident)
	requireFreeNickname(c, nick)
	count, err := q.CountUsers(ctx)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	perms := []string{permWorkerRegister}
	if count == 0 {
		perms = append(perms, permAdmin)
	}

	user, err := q.CreateUser(ctx, superdb.CreateUserParams{
		Identification: ident,
		Password:       hashPassword(body.Password),
		Nickname:       nick,
		Perms:          perms,
	})
	if err != nil {
		panic(dbWriteErr(err)) // PanicMiddleware가 응답 + txMiddleware가 롤백
	}

	return c.JSON(200, newUserResponse(user))
}

// signIn은 식별자/비밀번호 해시를 대조하고 성공 시 세션에 user를 싣는다.
func (router *supervisorRouter) signIn(c echo.Context) error {
	var body userSignRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}

	q := TxQueries(c)
	ctx := c.Request().Context()

	user, err := q.GetUserByIdentification(ctx, strings.TrimSpace(body.Identification))
	ok, legacy := false, false
	if err == nil {
		ok, legacy = checkPassword(user.Password, body.Password)
	}
	if !ok {
		// 계정 없음/비번 불일치를 같은 메시지로(존재 여부 노출 방지)
		panic(web.Err(400, "존재하지 않는 계정입니다"))
	}
	if user.SuspendedAt.Valid {
		panic(web.Err(403, "정지된 계정입니다"))
	}
	if legacy {
		user.Password = hashPassword(body.Password)
		if err := q.UpdateUserPassword(ctx, superdb.UpdateUserPasswordParams{ID: user.ID, Password: user.Password}); err != nil {
			panic(web.Err(500, "%v", err))
		}
	}

	nonce, err := util.RandomKey(16, "", "")
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	sess := router.sessions.GetSession(c.Request(), c.Response())
	sess.Data = &user
	// 재발 방지: 쿠키 = date(초)|세션 값|mac 이라 같은 계정이 같은 초에 로그인하면 sid가 같아짐 — REF-process-sync-impl.md "발견"
	sess.Session.Values[sessionNonceKey] = nonce
	if err := q.CreateUserSession(ctx, superdb.CreateUserSessionParams{ID: nonce, UserID: user.ID, UserAgent: c.Request().UserAgent()}); err != nil {
		panic(web.Err(500, "%v", err))
	}
	// 순서 의존: 메모리 등록은 커밋을 기다리지 않는다. 응답(쿠키)은 커밋 전에 나가므로
	// AfterCommit으로 미루면 브라우저의 바로 다음 요청(GET /users/session)이 등록 전에 와서 401이 난다.
	// 롤백되면 메모리에만 남지만 그 쿠키는 재시작 때 무효가 된다.
	router.logins.add(nonce, user.ID)
	if err := sess.Save(); err != nil {
		panic(web.Err(500, "%v", err))
	}

	return c.JSON(200, newUserResponse(user))
}

// checkSession은 현재 로그인 세션의 user를 DB에서 다시 읽어 반환한다(미로그인·지워진 계정=401).
// 쿠키의 sess.Data는 gob이 pgtype.Timestamptz를 못 담아 createdAt이 0이라 쓰지 않는다.
func (router *supervisorRouter) checkSession(c echo.Context) error {
	sess := router.requireSession(c)
	user, err := TxQueries(c).GetUser(c.Request().Context(), sess.Data.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		panic(web.Err(401, "로그인 절차가 필요합니다"))
	}
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return c.JSON(200, newUserResponse(user))
}

// signOut은 세션 쿠키를 만료시키고 세션 행을 지운다. 로그인 상태였으면 그 세션 탭들의 크기 소유권을 먼저 넘긴다.
func (router *supervisorRouter) signOut(c echo.Context) error {
	sess := router.sessions.GetSession(c.Request(), c.Response())
	if sess.Data != nil && sess.Data.ID != 0 && !sess.Session.IsNew {
		router.sizeOwnerOnSignOut(sess.Name())
		if login := sessionLogin(sess); login != "" {
			if _, err := TxQueries(c).DeleteUserSession(c.Request().Context(), superdb.DeleteUserSessionParams{ID: login, UserID: sess.Data.ID}); err != nil {
				panic(web.Err(500, "%v", err))
			}
			AfterCommit(c, func() { router.endLogins(login) }) // 같은 브라우저의 다른 탭도 로그아웃
		}
	}
	sess.Session.Options.MaxAge = -1
	if err := sess.Save(); err != nil {
		panic(web.Err(500, "%v", err))
	}
	return c.JSON(200, map[string]any{})
}
