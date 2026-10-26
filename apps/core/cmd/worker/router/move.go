package router

import (
	"errors"
	"log"
	"time"

	"nexus/internal/protocol"
)

// ===== supervisor 주소 옮기기 — 주소-2(옮기기 명령), REF-supervisor-move.md =====

var errMoved = errors.New("supervisor 주소 옮기기")

// move: MsgMove(REQ). 새 주소를 1순위로 저장하고, Now면 응답이 나간 뒤 지금 연결을 끊는다(main 루프가 1순위부터 다시 시도).
// 새 주소에 못 붙으면 다음 순위(옛 주소)로 돌아간다
func (r *workerRouter) move(req protocol.Frame) (any, error) {
	var body protocol.MoveRequest
	if err := req.Bind(&body); err != nil {
		return nil, err
	}
	addr, err := protocol.NormalizeSupervisorAddr(body.URL)
	if err != nil {
		return nil, err
	}
	list, err := r.state.addrs.Promote(addr)
	if err != nil {
		return nil, err
	}
	log.Printf("[move] 주소 저장 %s now=%v 목록=%v", addr, body.Now, list)
	if body.Now {
		go func() { // 응답은 핸들러가 끝난 뒤 dispatchReq가 쓴다 — 그보다 먼저 끊지 않게 잠깐 기다림
			time.Sleep(300 * time.Millisecond)
			r.conn.Close(errMoved)
		}()
	}
	return protocol.AddrsResponse{Addrs: list}, nil
}

// addrList: MsgAddrs(REQ). 저장된 주소 목록(1순위부터)
func (r *workerRouter) addrList(protocol.Frame) (any, error) {
	return protocol.AddrsResponse{Addrs: r.state.addrs.List()}, nil
}
