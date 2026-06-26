package router

import (
	"log"
	"net/http"

	"nexus/internal/manager"
	"nexus/internal/manager/session"
	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/transport"
	"nexus/internal/web"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type supervisorRouter struct {
	workers  *manager.KeyValManager[string, *transport.Conn]
	readers  *manager.KeyValManager[string, *fileSend]
	sessions *session.SessionManager[superdb.User]
}

// NewSupervisorRouter는 echo 서버를 세우고 worker WS 연결 라우트를 단다.
// REST(user/session 등)도 같은 echo 인스턴스에 붙여 WS와 단일 서버에서 공존시킨다.
// 서버 기동(addr 바인딩)은 호출자가 e.Start(addr)로 한다.
func NewSupervisorRouter(workerPath string) (*echo.Echo, *supervisorRouter) {

	log.Printf("[supervisor] SERVER 실행")
	router := &supervisorRouter{
		workers:  manager.NewKeyValManager[string, *transport.Conn](),
		readers:  manager.NewKeyValManager[string, *fileSend](),
		sessions: session.NewSessionManager[superdb.User]("irony", "sid", nil),
	}

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(web.PanicMiddleware) // 핸들러가 panic(web.Err(...))로 던진 ClientError를 JSON으로 렌더(+예기치 못한 panic도)
	e.Use(web.LogMiddelware)
	e.Use(txMiddleware(store.GetStorePool())) // PanicMiddleware 안쪽: 재-panic이 위로 전파되도록
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http*://*", "ws*://*"},
		AllowCredentials: true,
		// AllowOriginFunc: func(origin string) (bool, error) {
		// 	log.Printf("[FUNC] %s\n", origin)
		// 	return true, nil
		// },
		AllowMethods: []string{"GET", "POST", "PATCH", "DELETE", "PUT", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization", "MAC"},
	}))

	e.GET(workerPath, router.handleAgentWS)
	router.mountUsers(e)

	return e, router
}

// handleAgentWS는 worker의 WS 업그레이드를 받아 transport.Conn으로 다룬다.
// 업그레이드 성공 시 연결이 hijack되므로 echo는 비켜선다 → 본문은 기존 net/http
// 핸들러와 동일하고, 끝에서 반드시 nil을 반환해야 한다(에러 반환 시 echo가 hijack된
// 연결에 응답을 쓰려다 실패). upgrade 실패도 gorilla가 이미 HTTP 에러를 썼으므로 nil.
func (router *supervisorRouter) handleAgentWS(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("supervisor: upgrade 실패 %v", err)
		return nil
	}

	conn := transport.New(ws)

	var auth protocol.RegisterRequest
	conn.Handle(protocol.MsgRegister, router.register(conn, &auth))

	err = conn.Serve()

	// 연결 종료: 등록 해제 + 이 연결 소속 미완 전송(reader) 정리.
	key := auth.InstanceKey()
	if key != "" {
		router.workers.Remove(key)
		for _, fr := range router.readers.FindAll(func(_ string, v *fileSend) bool {
			return v.authKey == key
		}) {
			fr.Val.reader.Close() // OnClose가 readers에서 제거
		}
	}
	conn.Close(err)

	return nil
}
