package router

import (
	"log"
	"net/http"

	"nexus/internal/manager"
	"nexus/internal/protocol"
	"nexus/internal/transport"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type supervisorRouter struct {
	connectors *manager.KeyValManager[string, *transport.Conn]
	readers    *manager.KeyValManager[string, *fileSend]
}

func NewSupervisorRouter(addr, path string) (*http.Server, *supervisorRouter) {

	log.Printf("[supervisor] SERVER 실행")
	router := &supervisorRouter{
		connectors: manager.NewKeyValManager[string, *transport.Conn](),
		readers:    manager.NewKeyValManager[string, *fileSend](),
	}

	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("supervisor: upgrade 실패 %v", err)
			return
		}

		conn := transport.New(ws)

		var auth protocol.RegisterRequest
		conn.Handle(protocol.MsgRegister, router.register(conn, &auth))

		err = conn.Serve()

		// 연결 종료: 등록 해제 + 이 연결 소속 미완 전송(reader) 정리.
		key := auth.InstanceKey()
		if key != "" {
			router.connectors.Remove(key)
			for _, fr := range router.readers.FindAll(func(_ string, v *fileSend) bool {
				return v.authKey == key
			}) {
				fr.Val.reader.Close() // OnClose가 readers에서 제거
			}
		}
		conn.Close(err)

	})

	srv := &http.Server{Addr: addr, Handler: mux}

	return srv, router
}
