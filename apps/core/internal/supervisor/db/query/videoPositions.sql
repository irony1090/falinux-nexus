-- 영상 이어보기 S1(이어보기) — REF-video-view.md

-- name: GetVideoPosition :one
SELECT pos FROM video_positions
WHERE user_id = $1 AND instance = $2 AND path = $3;

-- name: SaveVideoPosition :execrows
-- 그 파일의 보기 타일이 내 트리에 있을 때만 — 닫힌 뒤 늦게 온 저장이 행을 되살리지 않게(P1(위치 지우는 시점))
INSERT INTO video_positions (user_id, instance, path, pos)
SELECT sqlc.arg(user_id), sqlc.arg(instance)::text, sqlc.arg(path)::text, sqlc.arg(pos)
WHERE EXISTS (
    SELECT 1 FROM tile_trees t, jsonb_each(t.tree->'tiles') AS e(id, tile)
    WHERE t.user_id = sqlc.arg(user_id) AND e.tile->>'type' = 'viewer'
      AND e.tile->>'instance' = sqlc.arg(instance)::text AND e.tile->>'path' = sqlc.arg(path)::text
)
ON CONFLICT (user_id, instance, path) DO UPDATE SET pos = EXCLUDED.pos, updated_at = NOW();

-- name: DeleteVideoPosition :exec
DELETE FROM video_positions
WHERE user_id = $1 AND instance = $2 AND path = $3;

-- name: DeleteOrphanVideoPositions :execrows
-- 주기 정리 — 저장과 닫기가 엇갈려 남은 행(보기 타일이 어느 트리에도 없는 것)
DELETE FROM video_positions v
WHERE NOT EXISTS (
    SELECT 1 FROM tile_trees t, jsonb_each(t.tree->'tiles') AS e(id, tile)
    WHERE t.user_id = v.user_id AND e.tile->>'type' = 'viewer'
      AND e.tile->>'instance' = v.instance AND e.tile->>'path' = v.path
);
