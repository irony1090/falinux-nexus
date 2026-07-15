package router

import (
	"fmt"

	"nexus/internal/protocol"
	"nexus/internal/transport"
	"nexus/internal/util"
)

func (r *supervisorRouter) register(conn *transport.Conn, body *protocol.RegisterRequest) transport.Handler {

	return func(req protocol.Frame) (any, error) {
		if err := req.Bind(body); err != nil {
			return nil, err
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
		r.workers.Append(key, conn)

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
