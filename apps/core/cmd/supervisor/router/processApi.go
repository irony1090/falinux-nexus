package router

import (
	"log"

	"nexus/cmd/supervisor/process"
	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

func (r *supervisorRouter) mountProcesses(e *echo.Echo) {
	g := e.Group("/processes")
	g.GET("/subscriptions", r.listSubscriptions)
	g.POST("/subscribe/:processId", r.subscribeProcess)
	g.DELETE("/subscribe/:processId", r.unSubscribeProcess)
	g.POST("/exec", r.execProcess)
	g.POST("/kill/:processId", r.killProcess)
}

func (r *supervisorRouter) listSubscriptions(c echo.Context) error {
	sess := r.requireSession(c)
	sid := sess.Name()

	processes, err := r.processManager.ListSubscriptions(sid)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}

	return c.JSON(200, newProcessResponses(processes))
}

// subscribeSid는 sid를 uid의 구독자로 등록(memory+DB write-through, ProcessManager.
// SubscribeProcess)하고, 이미 열려 있는 sid의 소켓 연결에도 즉시 hub 구독을 반영한다.
// 수동 구독(subscribeProcess)과 자동 구독(execProcess/killProcess) 둘 다 이 지점 하나를
// 공유한다 — REST로 직접 구독하든, 실행/종료에 딸려오든 결과가 완전히 같다.
func (r *supervisorRouter) subscribeSid(uid string, sub process.Subscriber) error {
	if err := r.processManager.SubscribeProcess(uid, sub); err != nil {
		return err
	}
	for _, conn := range r.browsersForSid(sub.Sid) {
		r.subscribeHub.Subscribe(processTopic(uid), conn)
	}
	return nil
}

// subscribeProcess는 현재 세션(sid)을 uid의 구독자로 등록한다. entry가 없으면(=이미
// 종료됐거나 존재한 적 없는 uid) 404.
func (r *supervisorRouter) subscribeProcess(c echo.Context) error {
	sess := r.requireSession(c)
	uid := c.Param("processId")

	sub := process.Subscriber{Sid: sess.Name(), OwnerUserID: sess.Data.ID}
	if err := r.subscribeSid(uid, sub); err != nil {
		panic(web.Err(404, "존재하지 않는 process입니다"))
	}
	return c.JSON(200, map[string]any{})
}

// unSubscribeProcess는 현재 세션(sid)을 uid의 구독자에서 뺀다(memory+DB, entry 유무와
// 무관하게 DB delete는 항상 시도 — ProcessManager.UnsubscribeProcess 참조).
func (r *supervisorRouter) unSubscribeProcess(c echo.Context) error {
	sess := r.requireSession(c)
	sid := sess.Name()
	uid := c.Param("processId")

	if err := r.processManager.UnsubscribeProcess(uid, sess.Name()); err != nil {
		panic(web.Err(500, "%v", err))
	}
	for _, conn := range r.browsersForSid(sid) {
		r.subscribeHub.Unsubscribe(processTopic(uid), conn)
	}
	return c.JSON(200, map[string]any{})
}

// execRequest: 노드 실행 요청. (POST /processes/exec)
type execRequest struct {
	NodeID  int64  `json:"nodeId" validate:"required"`
	AuthKey string `json:"authKey" validate:"required"` // worker 인스턴스 키(main#sub). 인스턴스
	// 선택 UI는 별도 미착수(REF-node-label.md "Exec 인스턴스 선택") — 지금은 프론트가 직접 지정한다.
	Type string `json:"type" validate:"omitempty,oneof=EXEC EDIT"` // 비우면 EXEC
}

// execProcess는 frontend의 "이 노드 실행" 요청을 받아 worker에 명령한다(router.Exec 위임).
// 요청 세션(sid)은 실행 성공 시 별도 구독 요청 없이 자동으로 이 process의 구독자가 된다
// — router.Exec 내부에서 subscribeSid로 처리하며, relay 기동 전이라 RUNNING 등 초기 이벤트도
// 놓치지 않는다.
func (r *supervisorRouter) execProcess(c echo.Context) error {
	sess := r.requireSession(c)
	var body execRequest
	if err := c.Bind(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}
	if err := c.Validate(&body); err != nil {
		panic(web.Err(400, "%v", err))
	}

	node, err := TxQueries(c).GetNode(c.Request().Context(), superdb.GetNodeParams{
		ID:          body.NodeID,
		OwnerUserID: sess.Data.ID,
	})
	if err != nil {
		panic(web.Err(404, "노드를 찾을 수 없습니다"))
	}

	sub := process.Subscriber{Sid: sess.Name(), OwnerUserID: sess.Data.ID}
	uid, err := r.Exec(*sess.Data, body.AuthKey, protocol.ExecType(body.Type), node, sub)
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return c.JSON(200, map[string]any{"uid": uid})
}

// killProcess는 실행 중인 process를 종료한다(worker에 MsgKill 전달, entry.Inter.Kill()).
// 종료 요청자도 자동으로 구독자가 된다 — subscribeSid를 그대로 재사용해, 뒤이어 오는 종료
// STATUS(Completed/Failed) 이벤트를 별도 구독 요청 없이 받아본다. relay는 실행 시점부터 이미
// 돌고 있으므로(exec 쪽과 달리) 순서 걱정 없이 아무 때나 구독해도 된다.
func (r *supervisorRouter) killProcess(c echo.Context) error {
	sess := r.requireSession(c)
	uid := c.Param("processId")
	// log.Printf("[PROCESS_API] KILL_PROCESS")
	// r.processManager.Test(func(s string, pe *process.ProcessEntry) bool {
	// 	log.Printf("[PROCESS_API] memory.items: %s", s)
	// 	return false
	// })
	entry, ok := r.processManager.Get(uid)
	log.Printf("[%s] %v", uid, entry)
	if !ok || entry.Inter == nil {
		panic(web.Err(404, "존재하지 않는 process입니다"))
	}
	if err := entry.Inter.Kill(); err != nil {
		panic(web.Err(500, "%v", err))
	}

	sub := process.Subscriber{Sid: sess.Name(), OwnerUserID: sess.Data.ID}
	if err := r.subscribeSid(uid, sub); err != nil {
		log.Printf("[process] kill 시 자동구독 실패 uid=%s: %v", uid, err)
	}
	return c.JSON(200, map[string]any{})
}
