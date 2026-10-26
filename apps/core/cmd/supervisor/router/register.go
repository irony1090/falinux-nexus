package router

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"time"

	"nexus/internal/protocol"
	"nexus/internal/supervisor/store"
	"nexus/internal/transport"
	"nexus/internal/util"
)

func (r *supervisorRouter) register(conn *transport.Conn, body *protocol.RegisterRequest) transport.Handler {

	return func(req protocol.Frame) (any, error) {
		if err := req.Bind(body); err != nil {
			return nil, err
		}
		if err := checkWorkerKey(body.Key, body.APIKey); err != nil {
			body.Key = "" // 끊길 때 정리 대상 아님(등록 안 됨)
			body.SubKey = ""
			return "", err
		}
		key := body.InstanceKey()
		if body.SubKey == "" {
			for {
				subKey, _ := util.RandomKey(8, "", "")
				body.SubKey = subKey
				if _, exist := r.workers.Get(body.InstanceKey()); !exist {
					break
				}
			}
			key = body.InstanceKey()
		} else if _, exist := r.workers.Get(key); exist {
			body.Key = ""
			body.SubKey = ""
			return "", fmt.Errorf("서브키 %q 는 이미 접속 중", key)
		}
		os := protocol.NormalizeOS(body.OS)
		r.workerOS.Remove(key)
		r.workerOS.Append(key, os) // workers보다 먼저 — 보이는 worker는 OS도 있게
		r.workers.Append(key, conn)
		recordSeen(key, true, os)

		// q := store.GetStorePool().Queries()
		// ctx := context.Background()
		// n, err := q.GetNode(ctx, superdb.GetNodeParams{ID: 3, OwnerUserID: 3})
		// u, err := q.GetUser(ctx, n.OwnerUserID)
		// if err == nil {

		// 	err := r.Exec(u, key, protocol.ExecTypeExec, n)
		// 	if err != nil {
		// 		log.Printf("[PROCESS ERR] %v", err)
		// 	} else {
		// 		log.Printf("[PROCESS SUC]")
		// 	}
		// } else {
		// 	log.Printf("[ERR] %v", err)
		// }

		return protocol.RegisterResponse{SubKey: body.SubKey}, nil
	}
}

// checkWorkerKey: 장비 이름이 등록돼 있고 API 키가 맞아야 등록 — 회원-4(worker 접속 인증), REF-member.md
func checkWorkerKey(mainKey, apiKey string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	d, err := store.GetStorePool().Queries().GetDeviceAuth(ctx, mainKey)
	if err != nil {
		log.Printf("[WORKER-REGISTER] 거부 %q: 등록되지 않은 장비 이름(%v)", mainKey, err)
		return errors.New("등록되지 않은 장비 이름이거나 API 키가 다릅니다")
	}
	if d.OwnerSuspended {
		log.Printf("[WORKER-REGISTER] 거부 %q: 주인 계정 정지", mainKey)
		return errors.New("장비 주인 계정이 정지되었습니다")
	}
	if subtle.ConstantTimeCompare([]byte(d.ApiKeyHash), []byte(apiKeyHash(apiKey))) != 1 {
		log.Printf("[WORKER-REGISTER] 거부 %q: API 키 불일치", mainKey)
		return errors.New("등록되지 않은 장비 이름이거나 API 키가 다릅니다")
	}
	return nil
}
