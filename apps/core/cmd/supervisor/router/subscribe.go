package router

import (
	"fmt"
	"log"
	"nexus/internal/transport"

	"github.com/labstack/echo/v4"
)

func nodeSubscribeKey(parentId int64) string {
	return fmt.Sprintf("NODE:%d", parentId)
}

func processSubscribeKey(processId string) string {
	return fmt.Sprintf("PROCESS:%s", processId)
}

func (r *supervisorRouter) handleSubscribeWS(c echo.Context) error {
	req, res := c.Request(), c.Response()
	r.requireSession(c)
	ws, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Printf("supervisor: upgrade 실패 %v", err)
		return nil
	}

	conn := transport.New(ws)

	//node:{parentId} - 0이면 nil인 node들을 구독한다
	r.subscribeHub.Subscribe(nodeSubscribeKey(0), conn)

	err = conn.Serve()
	conn.Close(err)
	r.subscribeHub.UnsubscribeAll(conn)

	return nil
}
