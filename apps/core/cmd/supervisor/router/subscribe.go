package router

import (
	"fmt"
	"log"
	"nexus/internal/transport"

	"github.com/labstack/echo/v4"
)

func nodeSubscribeTopic(parentId int64) string {
	return fmt.Sprintf("NODE:%d", parentId)
}

func (r *supervisorRouter) handleSubscribeWS(c echo.Context) error {
	req, res := c.Request(), c.Response()
	sess := r.requireSession(c)
	sid := sess.Name() // sid 겠지?
	// log.Printf("\tSID - %s", sid)
	ws, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Printf("supervisor: upgrade 실패 %v", err)
		return nil
	}

	conn := transport.New(ws)
	r.browsers.Append(conn, sid)

	// 구독중이던 process
	processes, err := r.processManager.ListSubscriptions(sid)
	for _, proc := range processes {
		if _, ok := r.processManager.Get(proc.Uid); ok {
			r.subscribeHub.Subscribe(processTopic(proc.Uid), conn)
		}
	}
	//node:{parentId} - 0이면 nil인 node들을 구독한다
	r.subscribeHub.Subscribe(nodeSubscribeTopic(0), conn)

	err = conn.Serve()
	conn.Close(err)
	// 모든 구독 해제
	r.subscribeHub.UnsubscribeAll(conn)
	r.browsers.Remove(conn)

	return nil
}
