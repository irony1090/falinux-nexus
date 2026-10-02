-- tile_trees 쿼리. 설계 → REF-node-ui-save.md "작업 단위 3-a~3-e".

-- name: GetTileTree :one
SELECT * FROM tile_trees
WHERE user_id = $1;

-- name: CreateTileTree :one
-- 이미 있으면(다른 탭이 먼저 만듦) 행 없음(pgx.ErrNoRows) → 호출자가 GetTileTree로 다시 읽는다
INSERT INTO tile_trees (user_id, tree)
VALUES ($1, $2)
ON CONFLICT (user_id) DO NOTHING
RETURNING *;

-- name: UpdateTileTree :one
-- 읽은 version일 때만 갱신 — 행 없음(pgx.ErrNoRows) = 그 사이 다른 저장이 있었음(409) — ⑪-2(충돌 처리)
UPDATE tile_trees
SET tree = sqlc.arg(tree), version = version + 1, updated_at = NOW()
WHERE user_id = sqlc.arg(user_id) AND version = sqlc.arg(version)
RETURNING *;

-- name: GetTileTreeForUpdate :one
-- exec 때 터미널 타일 추가 전 행 잠금(트랜잭션 안에서만) — ⑪-4(타일 없는 실행 중 process)
SELECT * FROM tile_trees
WHERE user_id = $1
FOR UPDATE;
