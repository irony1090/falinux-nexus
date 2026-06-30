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

		// d := transfer.NewReadBuffer([]byte("IRONY TEST\nHello World"), time.Second*2)
		// k, err := r.SendBuffer(key, d, "ironyMemory.txt", 0777)
		// if err != nil {
		// 	log.Printf("[SendBuffer] Err: %v", err)
		// } else {
		// 	log.Printf("[SendBuffer] Suc: %s", k)
		// }
		// uid, _ := util.RandomKey(8, "", "")
		// r.Exec(key, protocol.ProcessSpec{
		// 	UID:  uid,
		// 	Cmd:  fmt.Sprintf("vi %s", protocol.PlaceholderWorkerBase),
		// 	Args: []string{fmt.Sprintf("%s/TEST", protocol.PlaceholderWorkerBase)},
		// })

		return protocol.RegisterResponse{SubKey: body.SubKey}, nil
	}
}
