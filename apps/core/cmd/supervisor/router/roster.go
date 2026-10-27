package router

import (
	"context"
	"log"
	"slices"
	"time"

	"nexus/internal/protocol"
	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
)

// recordSeen은 명부(worker_instances)에 접속(connected=true, os 갱신)·끊김을 기록한다 — RS1(기록 시점), REF-node-label.md "roster 계획".
// 명부는 라우팅 authority가 아니므로 실패해도 로그만 남긴다. goroutine으로 돌리면 끊김 기록이 등록보다 먼저 실행될 수 있어 동기로 쓴다
func recordSeen(key string, connected bool, os string) {
	main, sub := splitInstanceKey(key)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	q := store.GetStorePool().Queries()
	var err error
	if connected {
		err = q.UpsertWorkerInstance(ctx, superdb.UpsertWorkerInstanceParams{MainKey: main, SubKey: sub, Os: os})
	} else {
		err = q.TouchWorkerInstance(ctx, superdb.TouchWorkerInstanceParams{MainKey: main, SubKey: sub})
	}
	if err != nil {
		log.Printf("[roster] 기록 실패 key=%s connected=%v: %v", key, connected, err)
	}
}

// publishDevice: 장비 상태가 바뀌었다는 신호(DEVICE:CHANGE)를 볼 수 있는 접속 계정에 — 받은 쪽이 다시 조회(REF-device-push.md).
// 상태를 DB·레지스트리에 반영한 뒤 부른다. also = 바뀌기 전에만 보던 계정(삭제·주인 변경 전 주인). 조회 실패 = also에만
func (r *supervisorRouter) publishDevice(mainKey string, also ...int64) {
	online := r.tabs.onlineUsers()
	if len(online) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	users, err := store.GetStorePool().Queries().FilterDeviceWatchers(ctx, superdb.FilterDeviceWatchersParams{UserIds: online, MainKey: mainKey})
	if err != nil {
		log.Printf("[roster] 알림 받을 계정 조회 실패 main=%s: %v", mainKey, err)
	}
	for _, u := range also {
		if slices.Contains(online, u) && !slices.Contains(users, u) {
			users = append(users, u)
		}
	}
	r.publishUsers(users, protocol.MsgDeviceChange, protocol.DeviceChangeEvent{MainKey: mainKey})
}
