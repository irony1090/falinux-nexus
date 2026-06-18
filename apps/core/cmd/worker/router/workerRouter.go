package router

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/url"
	"nexus/internal/protocol"
	"nexus/internal/transport"
	workerdb "nexus/internal/worker/db/gen"
	"nexus/internal/worker/store"
	"time"

	"github.com/gorilla/websocket"
)

type workerRouter struct {
	conn      *transport.Conn
	store     *store.StorePool
	uniqueKey string
}

func NewWorkerRouter(supervisor url.URL, uniqueKey string, store *store.StorePool) (*transport.Conn, chan error) {

	ws, _, err := websocket.DefaultDialer.Dial(supervisor.String(), nil)
	if err != nil {
		log.Fatalf("worker: 연결 실패 %v", err)
	}
	conn := transport.New(ws)
	serveErr := make(chan error, 1)
	go func() { serveErr <- conn.Serve() }() // 응답 수신용

	router := workerRouter{conn, store, uniqueKey}

	go func() {
		err = router.register()
		if err != nil {
			serveErr <- err
		}
	}()

	return conn, serveErr
}

func (r *workerRouter) register() error {

	q := r.store.Queries()
	ctx, selectCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer selectCancel()
	identity, err := q.GetIdentity(ctx, r.uniqueKey)
	var subKey string
	switch {
	case err == nil:
		subKey = identity.SubKey
	case errors.Is(err, sql.ErrNoRows):
		// 최초 접속: 저장된 서브키 없음 (정상)
	default:
		log.Printf("[WORKER-REGISTER] GetIdentity 실패 %v", err)
	}

	req := protocol.NewRegisterRequest(r.uniqueKey, subKey)

	// REGISTER 요청 → 응답이 이 한 줄에 묶여 돌아온다.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := r.conn.Call(ctx, protocol.MsgRegister, req)
	if err != nil {
		// 서브키 중복 등으로 supervisor가 차단하면 여기로 온다.
		// log.Fatalf("worker: 등록 실패 - %v", err)
		log.Printf("[WORKER-REGISTER] ERR %v", err)
		return err
	}

	var rr protocol.RegisterResponse
	if err := res.Bind(&rr); err != nil {
		return err
	}
	if rr.SubKey != subKey {
		ctx, updateCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer updateCancel()
		if err := q.UpsertIdentity(ctx, workerdb.UpsertIdentityParams{
			MainKey: r.uniqueKey,
			SubKey:  rr.SubKey,
		}); err != nil {
			log.Printf("[WORKER-REGISTER] UpsertIdentity 실패 %v", err)
		}
	}

	return nil
}
