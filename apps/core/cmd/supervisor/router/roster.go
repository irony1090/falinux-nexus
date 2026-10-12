package router

import (
	"context"
	"log"
	"time"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
)

// recordSeen은 명부(worker_instances)에 접속(connected=true)·끊김을 기록한다 — RS1(기록 시점), REF-node-label.md "roster 계획".
// 명부는 라우팅 authority가 아니므로 실패해도 로그만 남긴다. goroutine으로 돌리면 끊김 기록이 등록보다 먼저 실행될 수 있어 동기로 쓴다
func recordSeen(key string, connected bool) {
	main, sub := splitInstanceKey(key)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	q := store.GetStorePool().Queries()
	var err error
	if connected {
		err = q.UpsertWorkerInstance(ctx, superdb.UpsertWorkerInstanceParams{MainKey: main, SubKey: sub})
	} else {
		err = q.TouchWorkerInstance(ctx, superdb.TouchWorkerInstanceParams{MainKey: main, SubKey: sub})
	}
	if err != nil {
		log.Printf("[roster] 기록 실패 key=%s connected=%v: %v", key, connected, err)
	}
}
