package router

import (
	"context"
	"log"
	"time"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
)

// 종료 유예(서버 10초 규칙) — 종료 요청 뒤 이만큼 지나야 종료-4(모두 닫음)로 강제 종료한다.
// 브라우저가 기다리지 않아도(닫아도) 서버가 판정하도록 주기 검사로 둔다. 재시작해도 다음 검사가 이어받는다 — REF-board-close.md
// 남은 종료 요청이 없으면 killPending을 꺼 DB 조회를 쉰다(안 A(메모리 표시))
const (
	killGrace      = 10 * time.Second
	killSweepEvery = 2 * time.Second
)

func (r *supervisorRouter) runKillSweep() {
	r.killPending.Store(true) // 시작 = 재시작 전 요청이 남았을 수 있음
	for {
		time.Sleep(killSweepEvery)
		if !r.killPending.Load() {
			continue
		}
		// 순서 의존: 끄기 -> 판정 -> 남았으면 다시 켬. 그사이 들어온 요청은 markKillPending(커밋 뒤)이 다시 켜거나
		// 아래 조회에 잡힌다 — 끄기를 조회 뒤로 옮기면 그 틈의 요청을 놓친다
		r.killPending.Store(false)
		r.forceKills(true, []string{})
		left, err := store.GetStorePool().Queries().HasPendingKills(context.Background())
		if err != nil || left {
			if err != nil {
				log.Printf("[process] 남은 종료 요청 조회 실패: %v", err)
			}
			r.killPending.Store(true)
		}
	}
}

// markKillPending: 종료 요청이 커밋된 뒤 부른다(AfterCommit) — 주기 검사를 깨운다
func (r *supervisorRouter) markKillPending() { r.killPending.Store(true) }

// forceKills: all = 종료 요청 전부, 아니면 uids만. 강제로 바뀐 행만 돌아오므로 같은 process에 두 번 보내지 않는다
func (r *supervisorRouter) forceKills(all bool, uids []string) {
	rows, err := store.GetStorePool().Queries().ForceAbandonedKills(context.Background(), superdb.ForceAbandonedKillsParams{
		AllUids: all, Uids: uids, GraceSecs: int32(killGrace / time.Second), MinPerm: superdb.PermOperate,
	})
	if err != nil {
		log.Printf("[process] 모두 닫음 판정 실패 all=%v uids=%v: %v", all, uids, err)
		return
	}
	for _, row := range rows {
		r.sendKill(row)
	}
}
