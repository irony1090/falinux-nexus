package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	superdb "nexus/internal/supervisor/db/gen"
	"nexus/internal/supervisor/store"
)

// 옛 타일 문서(kids·adopted, order 없음) -> order 문서 — 변환-B(시작 때 한 번, REF-tile-drag.md). 옛 문서가 다 바뀐 뒤엔 지워도 됨

type legacyTile struct {
	Type    string   `json:"type"`
	Kids    []string `json:"kids"`
	Adopted bool     `json:"adopted"`
}

// legacyOrder: 옛 화면 순서 규칙 — 폴더 -> 직접 실행한 터미널 -> 넘겨받은 터미널 -> 연 하위 타일(재귀), 묶음 안은 kids 순
func legacyOrder(tree []byte) ([]string, error) {
	var old struct {
		RootID string                `json:"rootId"`
		Tiles  map[string]legacyTile `json:"tiles"`
	}
	if err := json.Unmarshal(tree, &old); err != nil {
		return nil, err
	}
	var out []string
	var walk func(id string)
	walk = func(id string) {
		t, ok := old.Tiles[id]
		if !ok {
			return
		}
		out = append(out, id)
		var procs, adopted, rest []string
		for _, k := range t.Kids {
			kt, ok := old.Tiles[k]
			switch {
			case !ok:
			case !tileKinds[kt.Type].holdsProc:
				rest = append(rest, k)
			case kt.Adopted:
				adopted = append(adopted, k)
			default:
				procs = append(procs, k)
			}
		}
		out = append(append(out, procs...), adopted...)
		for _, k := range rest {
			walk(k)
		}
	}
	walk(old.RootID)
	return out, nil
}

// MigrateTileOrder: order 없는 타일 문서를 옛 순서 그대로 바꿔 저장(version+1). 멱등 — supervisor 시작 때 마이그레이션 바로 뒤, 서빙 전
func MigrateTileOrder(ctx context.Context) error {
	tx := store.GetStorePool().Transaction()
	if err := tx.Begin(ctx); err != nil {
		return err
	}
	n, err := func() (int, error) {
		q, err := tx.Queries()
		if err != nil {
			return 0, err
		}
		recs, err := q.ListTileTreesWithoutOrder(ctx)
		if err != nil {
			return 0, err
		}
		for _, rec := range recs {
			order, err := legacyOrder(rec.Tree)
			if err != nil {
				return 0, fmt.Errorf("타일 문서 %s: %w", rec.Hash, err)
			}
			var doc tileDoc
			if err := json.Unmarshal(rec.Tree, &doc); err != nil {
				return 0, fmt.Errorf("타일 문서 %s: %w", rec.Hash, err)
			}
			doc.Order = order
			if err := doc.validate(); err != nil {
				return 0, fmt.Errorf("타일 문서 %s: %w", rec.Hash, err)
			}
			tree, err := doc.marshal()
			if err != nil {
				return 0, err
			}
			if _, err := q.UpdateTileTreeByHash(ctx, superdb.UpdateTileTreeByHashParams{UserID: rec.UserID, Hash: rec.Hash, Version: rec.Version, Tree: tree}); err != nil {
				return 0, fmt.Errorf("타일 문서 %s: %w", rec.Hash, err)
			}
		}
		return len(recs), nil
	}()
	if err != nil {
		_ = tx.Rollback(ctx, err)
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	if n > 0 {
		log.Printf("[tiles] 옛 타일 문서 %d개에 order를 넣음", n)
	}
	return nil
}
