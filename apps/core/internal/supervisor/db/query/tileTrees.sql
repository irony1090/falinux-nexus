-- tile_trees 쿼리. 설계 → REF-node-ui-save.md "작업 단위 3-a~3-e", 작업 공간 여러 개 → REF-workspace-impl.md
-- 이름에 ByHash가 없는 쿼리 = 기본 작업 공간(is_default) 대상

-- name: GetTileTree :one
SELECT * FROM tile_trees
WHERE user_id = $1 AND is_default;

-- name: CreateTileTree :one
-- 이미 있으면(다른 탭이 먼저 만듦) 행 없음(pgx.ErrNoRows) → 호출자가 GetTileTree로 다시 읽는다
INSERT INTO tile_trees (user_id, hash, tree, is_default)
VALUES ($1, $2, $3, true)
ON CONFLICT (user_id) WHERE is_default DO NOTHING
RETURNING *;

-- name: UpdateTileTree :one
-- 읽은 version일 때만 갱신 — 행 없음(pgx.ErrNoRows) = 그 사이 다른 저장이 있었음(409) — ⑪-2(충돌 처리)
UPDATE tile_trees
SET tree = sqlc.arg(tree), version = version + 1, updated_at = NOW()
WHERE user_id = sqlc.arg(user_id) AND is_default AND version = sqlc.arg(version)
RETURNING *;

-- name: GetTileTreeForUpdate :one
-- exec 때 터미널 타일 추가 전 행 잠금(트랜잭션 안에서만) — ⑪-4(타일 없는 실행 중 process)
SELECT * FROM tile_trees
WHERE user_id = $1 AND is_default
FOR UPDATE;

-- name: GetTileTreeByHash :one
-- 다른 계정의 hash면 행 없음(404)
SELECT * FROM tile_trees
WHERE user_id = $1 AND hash = $2;

-- name: GetTileTreeByHashForUpdate :one
SELECT * FROM tile_trees
WHERE user_id = $1 AND hash = $2
FOR UPDATE;

-- name: UpdateTileTreeByHash :one
-- UpdateTileTree와 같은 version 규칙 — PUT /tiles/:hash, 서버 동작 API(WS2(변경 방식))
UPDATE tile_trees
SET tree = sqlc.arg(tree), version = version + 1, updated_at = NOW()
WHERE user_id = sqlc.arg(user_id) AND hash = sqlc.arg(hash) AND version = sqlc.arg(version)
RETURNING *;

-- name: ListTileTrees :many
-- 작업 공간 목록용이라 tree는 뺀다
SELECT id, hash, name, ord, is_default, version, updated_at
FROM tile_trees
WHERE user_id = $1
ORDER BY ord, id;

-- name: ListTileTreesForUpdate :many
-- 계정의 트리 전부 잠금 — WS2(변경 방식)의 두 트리 이동·WS3(고아 안전망 위치)의 전체 검사.
-- 여러 행을 잠그는 곳은 반드시 이 쿼리만 쓴다: 잠금 순서를 id 순으로 고정해야 두 트랜잭션이
-- 서로 상대가 잡은 행을 기다리는 데드락이 생기지 않는다(한 행만 잠그는 쿼리는 해당 없음)
SELECT * FROM tile_trees
WHERE user_id = $1
ORDER BY id
FOR UPDATE;

-- name: CreateWorkspaceTree :one
-- 기본이 아닌 작업 공간 만들기(분리) — 목록 끝(ord 최대 + 1)
INSERT INTO tile_trees (user_id, hash, name, tree, ord)
VALUES (
    sqlc.arg(user_id), sqlc.arg(hash), sqlc.arg(name), sqlc.arg(tree),
    (SELECT COALESCE(MAX(ord), 0) + 1 FROM tile_trees WHERE user_id = sqlc.arg(user_id))
)
RETURNING *;

-- name: RenameTileTree :one
-- 이름은 트리 내용이 아니라 version을 올리지 않는다
UPDATE tile_trees
SET name = sqlc.arg(name), updated_at = NOW()
WHERE user_id = sqlc.arg(user_id) AND hash = sqlc.arg(hash)
RETURNING *;

-- name: DeleteTileTree :execrows
-- 기본은 삭제 불가(W2(삭제 처리)) — 0행 = 없음 또는 기본
DELETE FROM tile_trees
WHERE user_id = $1 AND hash = $2 AND NOT is_default;

-- name: ListTileTreeDocs :many
-- 잠그지 않는 전체 읽기 — 고아 안전망(WS3(고아 안전망 위치))의 먼저 보기, 고아가 있을 때만 ListTileTreesForUpdate로 잠근다
SELECT hash, is_default, tree
FROM tile_trees
WHERE user_id = $1
ORDER BY id; -- ListTileTreesForUpdate와 같은 순서 — 고정 지도가 bindingTile과 같은 동률 처리를 하게

-- name: ListTreesWithTerminal :many
-- 그 uid의 터미널 타일이 있는 작업 공간(계정 무관) — 크기 우선권 후보 좁힘(WS7)
SELECT user_id, hash, is_default
FROM tile_trees
WHERE jsonb_path_exists(tree, '$.tiles.* ? (@.type == "terminal" && @.uid == $uid)', jsonb_build_object('uid', sqlc.arg(uid)::text));

-- name: ListTerminalRefs :many
-- uids 중 아직 터미널 타일이 남은 것(계정·uid 쌍) — K3(타일 같이 정리)
SELECT DISTINCT tt.user_id, (t.value->>'uid')::text AS uid
FROM tile_trees tt, jsonb_each(tt.tree->'tiles') t
WHERE t.value->>'type' = 'terminal'
  AND t.value->>'uid' = ANY(sqlc.arg(uids)::text[]);
