package router

import (
	"crypto/sha256"
	"fmt"

	"nexus/internal/manager/session"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 구 PortBridge cmd/agent_v2 user 핸들러를 이식. echo 기본 핸들러(func(c) error)를 따르되
// 에러는 panic(web.Err(...))로 던져 PanicMiddleware가 JSON으로 응답하고,
// 트랜잭션은 txMiddleware가 깐 TxQueries(c)로 꺼낸다(롤백은 미들웨어 recover가 처리).

type userCreateRequest struct {
	Identification string `json:"identification" validate:"required"`
	Nickname       string `json:"nickname" validate:"required"`
	Password       string `json:"password" validate:"required"`
}

type userSignRequest struct {
	Identification string `json:"identification" validate:"required"`
	Password       string `json:"password" validate:"required"`
}

type userResponse struct {
	Identification string `json:"identification"`
	Nickname       string `json:"nickname"`
	CreatedAt      int64  `json:"createdAt"`
	UpdatedAt      *int64 `json:"updatedAt"`
}

func newUserResponse(u superdb.User) userResponse {
	r := userResponse{
		Identification: u.Identification,
		Nickname:       u.Nickname,
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

// toHash는 식별자/비밀번호를 sha256 hex로 변환한다(구 PortBridge와 동일 — 저장·조회 모두 해시).
func toHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

// requireSession은 로그인된 세션을 돌려주거나 401을 panic한다.
// 구 GetSessionOrPanic의 flag if/else 사슬을 단락 OR 한 줄로 축약했다.
func (router *supervisorRouter) requireSession(c echo.Context) *session.SessionElement[superdb.User] {
	sess := router.sessions.GetSession(c.Request(), nil)
	if sess == nil || sess.Session.IsNew || sess.Data.ID == 0 {
		panic(web.Err(401, "로그인 절차가 필요합니다"))
	}
	return sess
}

func (router *supervisorRouter) mountUsers(e *echo.Echo) {
	g := e.Group("/users")
	g.POST("", router.createUser)
	g.POST("/session", router.signIn)
	g.GET("/session", router.checkSession)
	g.DELETE("/session", router.signOut)
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

	q := TxQueries(c)
	ctx := c.Request().Context()
	idHash := toHash(body.Identification)

	if _, err := q.GetUserByIdentification(ctx, idHash); err == nil {
		panic(web.Err(400, "가입할 수 없는 ID입니다"))
	}

	user, err := q.CreateUser(ctx, superdb.CreateUserParams{
		Identification: idHash,
		Password:       toHash(body.Password),
		Nickname:       body.Nickname,
	})
	if err != nil {
		panic(web.Err(500, "%v", err)) // PanicMiddleware가 500 + txMiddleware가 롤백
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

	user, err := q.GetUserByIdentification(ctx, toHash(body.Identification))
	if err != nil || user.Password != toHash(body.Password) {
		// 계정 없음/비번 불일치를 같은 메시지로(존재 여부 노출 방지)
		panic(web.Err(400, "존재하지 않는 계정입니다"))
	}

	sess := router.sessions.GetSession(c.Request(), c.Response())
	sess.Data = &user
	if err := sess.Save(); err != nil {
		panic(web.Err(500, "%v", err))
	}

	return c.JSON(200, newUserResponse(user))
}

// checkSession은 현재 로그인 세션의 user를 반환한다(미로그인=401).
func (router *supervisorRouter) checkSession(c echo.Context) error {
	sess := router.requireSession(c)
	return c.JSON(200, newUserResponse(*sess.Data))
}

// signOut은 세션 쿠키를 만료시킨다.
func (router *supervisorRouter) signOut(c echo.Context) error {
	sess := router.sessions.GetSession(c.Request(), c.Response())
	sess.Session.Options.MaxAge = -1
	if err := sess.Save(); err != nil {
		panic(web.Err(500, "%v", err))
	}
	return c.JSON(200, map[string]any{})
}
