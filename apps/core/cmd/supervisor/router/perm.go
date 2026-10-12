package router

import (
	"context"
	"log"

	"nexus/cmd/supervisor/process"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
	"nexus/internal/web"

	"github.com/labstack/echo/v4"
)

// 공유 단계 판정 헬퍼 — 단계 계산은 DB 함수(node_permission/process_permission)에만 있다. REF-process-share-authz.md

// checkPerm은 단계가 모자라면 panic한다. 0(볼 수 없음) = 404(존재를 드러내지 않음) / 보기 이상인데 미달 = 403 — SB1(권한 부족 응답)
func checkPerm(perm, min int16, notFound string) {
	if perm <= superdb.PermBlock {
		panic(web.Err(404, "%s", notFound))
	}
	if perm < min {
		panic(web.Err(403, "권한이 없습니다"))
	}
}

// requireNodePerm은 노드에 대한 요청 계정의 단계를 반환한다(노드 행은 읽지 않음).
func requireNodePerm(c echo.Context, userID, nodeID int64, min int16) int16 {
	perm := mustNodePerm(c, userID, nodeID)
	checkPerm(perm, min, "노드를 찾을 수 없습니다")
	return perm
}

// mustNodePerm·mustProcessPerm: 검사 없이 단계만(요청 트랜잭션 — 같은 요청의 변경이 보임)
func mustNodePerm(c echo.Context, userID, nodeID int64) int16 {
	perm, err := TxQueries(c).NodePermission(c.Request().Context(), superdb.NodePermissionParams{UserID: userID, NodeID: nodeID})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return perm
}

func mustProcessPerm(c echo.Context, userID int64, uid string) int16 {
	perm, err := TxQueries(c).ProcessPermission(c.Request().Context(), superdb.ProcessPermissionParams{UserID: userID, Uid: uid})
	if err != nil {
		panic(web.Err(500, "%v", err))
	}
	return perm
}

// requireNode는 판정을 통과한 노드와 단계를 반환한다.
func requireNode(c echo.Context, userID, nodeID int64, min int16) (superdb.Node, int16) {
	perm := requireNodePerm(c, userID, nodeID, min)
	node, err := TxQueries(c).GetNodeByID(c.Request().Context(), nodeID)
	if err != nil {
		panic(web.Err(404, "노드를 찾을 수 없습니다"))
	}
	return node, perm
}

// processPerm은 process에 대한 계정의 단계를 반환한다. 소유자면 메모리로 끝, 아니면 DB — SB2(입력 권한 확인 방식)
func (r *supervisorRouter) processPerm(uid string, userID int64) int16 {
	if entry, ok := r.processManager.Get(uid); ok && entry.Record != nil && entry.Record.OwnerUserID == userID {
		return superdb.PermOwner
	}
	perm, err := store.GetStorePool().Queries().ProcessPermission(context.Background(), superdb.ProcessPermissionParams{UserID: userID, Uid: uid})
	if err != nil {
		log.Printf("[perm] process 단계 조회 실패 uid=%s user=%d: %v", uid, userID, err)
		return superdb.PermBlock
	}
	return perm
}

// requireProcess는 판정을 통과한 실행 중 entry를 반환한다. 끝났거나 worker 끊김(memory에 없음)이면 404.
func (r *supervisorRouter) requireProcess(uid string, userID int64, min int16) *process.ProcessEntry {
	checkPerm(r.processPerm(uid, userID), min, "존재하지 않는 process입니다")
	entry, ok := r.processManager.Get(uid)
	if !ok || entry.Inter == nil || entry.Record == nil {
		panic(web.Err(404, "존재하지 않는 process입니다"))
	}
	return entry
}
