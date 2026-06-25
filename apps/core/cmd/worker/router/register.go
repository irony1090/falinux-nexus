package router

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"nexus/internal/protocol"
	workerdb "nexus/internal/worker/db/gen"
)

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
		log.Printf("[WORKER-REGISTER] ERR %v", err)
		return err
	}

	var rr protocol.RegisterResponse
	if err := res.Bind(&rr); err != nil {
		return err
	}

	// 저장 경로에 쓸 서브키 확정(파일 수신 핸들러가 읽음).
	r.mu.Lock()
	r.subKey = rr.SubKey
	r.mu.Unlock()

	r.reached.Store(true) // supervisor가 받아줌 = 정상 가동 도달(backoff reset 기준)

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
