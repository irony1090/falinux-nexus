package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"

	"github.com/jackc/pgx/v5/pgtype"
)

// 끝난 process 주기 삭제 — K2(supervisor 정리)·K3(타일 같이 정리), REF-process-keep.md 살리기-7(supervisor 주기 삭제)

const (
	purgeAge   = 24 * time.Hour
	purgeEvery = time.Hour
)

func (r *supervisorRouter) runPurge() {
	for {
		if err := r.purgeProcesses(time.Now().Add(-purgeAge)); err != nil {
			log.Printf("[purge] %v", err)
		}
		time.Sleep(purgeEvery)
	}
}

// purgeProcesses는 before 전에 끝난 process를 지운다. 타일을 먼저 빼고, 그래도 타일이 남은 uid(트리 저장 실패)는 이번엔 남긴다
func (r *supervisorRouter) purgeProcesses(before time.Time) error {
	ctx := context.Background()
	q := store.GetStorePool().Queries()
	uids, err := q.ListExpiredProcesses(ctx, pgtype.Timestamptz{Time: before, Valid: true})
	if err != nil || len(uids) == 0 {
		return err
	}
	refs, err := q.ListTerminalRefs(ctx, uids)
	if err != nil {
		return err
	}
	byUser := map[int64]map[string]bool{}
	for _, ref := range refs {
		if byUser[ref.UserID] == nil {
			byUser[ref.UserID] = map[string]bool{}
		}
		byUser[ref.UserID][ref.Uid] = true
	}
	for userID, set := range byUser {
		if err := r.purgeTiles(ctx, userID, set); err != nil {
			log.Printf("[purge] 타일 정리 실패 user=%d: %v", userID, err)
		}
	}

	left, err := q.ListTerminalRefs(ctx, uids)
	if err != nil {
		return err
	}
	keep := map[string]bool{}
	for _, ref := range left {
		keep[ref.Uid] = true
	}
	del := make([]string, 0, len(uids))
	for _, uid := range uids {
		if !keep[uid] {
			del = append(del, uid)
		}
	}
	if err := q.DeleteProcesses(ctx, del); err != nil {
		return err
	}
	log.Printf("[purge] 끝난 process %d개 삭제(타일 남아 미룸 %d개)", len(del), len(uids)-len(del))
	return nil
}

// purgeTiles는 계정의 모든 보드에서 uids의 터미널 타일을 빼고 바뀐 보드를 push한다(계정 트리 전부 잠금 — lockWorkspaces와 같은 순서)
func (r *supervisorRouter) purgeTiles(ctx context.Context, userID int64, uids map[string]bool) (err error) {
	tx := store.GetStorePool().Transaction()
	if err := tx.Begin(ctx); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx, err)
		}
	}()
	q, err := tx.Queries()
	if err != nil {
		return err
	}
	recs, err := q.ListTileTreesForUpdate(ctx, userID)
	if err != nil {
		return err
	}
	var saved []superdb.TileTree
	for _, rec := range recs {
		var doc tileDoc
		if err := json.Unmarshal(rec.Tree, &doc); err != nil {
			return fmt.Errorf("트리 %s: %w", rec.Hash, err)
		}
		removed := false
		for id, t := range doc.Tiles {
			if t.holdsProc() && uids[t.UID] {
				doc.remove(id)
				removed = true
			}
		}
		if !removed {
			continue
		}
		if err := doc.validate(); err != nil {
			return fmt.Errorf("트리 %s: %w", rec.Hash, err)
		}
		tree, err := doc.marshal()
		if err != nil {
			return err
		}
		upd, err := q.UpdateTileTreeByHash(ctx, superdb.UpdateTileTreeByHashParams{UserID: userID, Hash: rec.Hash, Version: rec.Version, Tree: tree})
		if err != nil {
			return fmt.Errorf("트리 %s: %w", rec.Hash, err)
		}
		saved = append(saved, upd)
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	for _, rec := range saved {
		r.publishTiles(userID, rec, "")
	}
	return nil
}
