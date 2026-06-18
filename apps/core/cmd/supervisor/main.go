// Command supervisor는 worker agent의 연결을 받는 총괄 서버다.
//
// 등록(REGISTER) 핸드셰이크:
//   - 최초 접속: worker가 고유키만 보냄 → supervisor가 새 서브키를 부여
//   - 재접속:   worker가 고유키+서브키를 보냄 → 인식 (단, 그 서브키가 이미
//     접속 중이면 차단)
package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"nexus/internal/protocol"
	"nexus/internal/transport"

	"github.com/gorilla/websocket"
)

const addr = "localhost:8080"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// registry는 현재 접속 중인 서브키 장부다(최소 구현: 메모리만).
type registry struct {
	mu     sync.Mutex
	active map[string]bool // 접속 중인 서브키
	seq    uint64          // 서브키 생성용 일련번호
}

func newRegistry() *registry {
	return &registry{active: make(map[string]bool)}
}

// register는 등록을 처리하고 확정된 서브키를 반환한다.
// owned에 이 연결이 점유한 서브키를 기록해(같은 락 안에서) 종료 시 해제에 쓴다.
func (r *registry) register(req protocol.RegisterRequest, owned *string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sub := req.SubKey
	switch {
	case sub == "":
		// 최초 접속: 고유키 + 일련번호로 서브키 부여 (같은 고유키도 서로 구분됨)
		r.seq++
		req.SubKey = fmt.Sprintf("%d", r.seq)
		// sub = fmt.Sprintf("%s#%d", req.Key, r.seq)
	case r.active[req.InstanceKey()]:
		// 재접속인데 그 서브키가 이미 접속 중 → 차단
		return "", fmt.Errorf("서브키 %q 는 이미 접속 중", sub)
	}
	*owned = req.InstanceKey()
	r.active[*owned] = true

	return req.SubKey, nil
}

// release는 연결 종료 시 점유 서브키를 장부에서 빼고, 해제한 서브키를 반환한다.
func (r *registry) release(owned *string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	sub := *owned
	if sub != "" {
		delete(r.active, sub)
	}
	return sub
}

type server struct {
	reg *registry
}

func main() {
	s := &server{reg: newRegistry()}
	http.HandleFunc("/agent", s.handleAgent)
	log.Printf("supervisor: %s 에서 worker 연결 대기", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

// handleAgent는 worker가 접속할 때마다 호출된다.
func (s *server) handleAgent(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("supervisor: upgrade 실패 %v", err)
		return
	}

	conn := transport.New(ws)

	// 이 연결이 점유한 서브키. register가 같은 락 안에서 기록 → 종료 시 release.
	var owned string

	conn.Handle(protocol.MsgRegister, func(req protocol.Frame) (any, error) {
		var rr protocol.RegisterRequest
		if err := req.Bind(&rr); err != nil {
			return nil, err
		}
		sub, err := s.reg.register(rr, &owned)
		if err != nil {
			log.Printf("supervisor: 등록 거부 key=%s sub=%s (%v)", rr.Key, rr.SubKey, err)
			return nil, err
		}
		if rr.SubKey == "" {
			log.Printf("supervisor: 최초 등록 key=%s → 서브키 %s 부여", rr.Key, sub)
		} else {
			log.Printf("supervisor: 재접속 등록 key=%s 서브키 %s", rr.Key, sub)
		}
		return protocol.RegisterResponse{SubKey: sub}, nil
	})

	err = conn.Serve() // 블록: 요청 수신 루프
	if sub := s.reg.release(&owned); sub != "" {
		log.Printf("supervisor: 연결 종료, 서브키 해제 %s", sub)
	}
	conn.Close(err)
}
