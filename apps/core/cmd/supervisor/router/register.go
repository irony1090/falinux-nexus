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
				if _, exist := r.connectors.Get(body.InstanceKey()); !exist {
					break
				}
			}
			key = body.InstanceKey()
		} else if _, exist := r.connectors.Get(key); exist {
			body.Key = ""
			body.SubKey = ""
			return "", fmt.Errorf("서브키 %q 는 이미 접속 중", key)
		}
		r.connectors.Append(key, conn)

		return protocol.RegisterResponse{SubKey: body.SubKey}, nil
	}
}
