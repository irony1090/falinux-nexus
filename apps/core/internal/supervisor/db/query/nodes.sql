-- nodes 카탈로그 쿼리. 설계 → REF-node-label.md.
-- 규약:
--  · 모든 조회/변경은 owner_user_id로 스코핑(per-user 카탈로그 — 남의 노드 접근 차단).
--  · 루트 자식은 parent_id IS NULL → `IS NOT DISTINCT FROM`로 NULL=루트도 한 쿼리에서 커버.
--  · ord(그리드 순서)는 호출자가 midpoint/append로 계산해 넘긴다(MaxChildOrd 참고).
--  · 주의(핸들러 책임): ① parent_id는 같은 owner 노드여야 함 ② 자기 자신의 자손으로 Move 금지(사이클).

-- name: CreateFolder :one
INSERT INTO nodes (
    owner_user_id, parent_id, kind, name, ord, position_x, position_y, device_key
) VALUES ( $1, $2, 'FOLDER', $3, $4, $5, $6, $7 )
RETURNING *;

-- name: CreateScript :one
INSERT INTO nodes (
    owner_user_id, parent_id, kind, name, ord, position_x, position_y, content
) VALUES ( $1, $2, 'SCRIPT', $3, $4, $5, $6, $7 )
RETURNING *;

-- name: GetNode :one
SELECT * FROM nodes
WHERE id = $1 AND owner_user_id = $2;

-- name: ListChildren :many
-- 한 폴더의 자식(또는 parent NULL이면 루트) 그리드/리스트 순서로.
SELECT * FROM nodes
WHERE owner_user_id = sqlc.arg(owner_user_id)
  AND parent_id IS NOT DISTINCT FROM sqlc.narg(parent_id)
ORDER BY ord, name, id;

-- name: MaxChildOrd :one
-- append 위치 계산용(없으면 0). midpoint 삽입은 호출자가 이웃 ord로 별도 계산.
SELECT COALESCE(MAX(ord), 0)::BIGINT AS max_ord
FROM nodes
WHERE owner_user_id = sqlc.arg(owner_user_id)
  AND parent_id IS NOT DISTINCT FROM sqlc.narg(parent_id);

-- name: MoveNode :one
-- 재부모화 + 그리드 순서 갱신(같은 부모로 넘기면 단순 reorder). 캔버스 좌표는 PlaceNode.
UPDATE nodes
SET parent_id  = $2,
    ord        = $3,
    updated_at = NOW()
WHERE id = $1 AND owner_user_id = $4
RETURNING *;

-- name: PlaceNode :one
-- 캔버스 절대배치 좌표만 갱신(그리드 순서 ord와 독립).
UPDATE nodes
SET position_x = $2,
    position_y = $3,
    updated_at = NOW()
WHERE id = $1 AND owner_user_id = $4
RETURNING *;

-- name: UpdateNodeContent :one
-- 스크립트 본문 갱신(EDIT read-back 등). SCRIPT 행만 대상.
UPDATE nodes
SET content    = $2,
    updated_at = NOW()
WHERE id = $1 AND owner_user_id = $3 AND kind = 'SCRIPT'
RETURNING *;

-- name: RenameNode :one
UPDATE nodes
SET name       = $2,
    updated_at = NOW()
WHERE id = $1 AND owner_user_id = $3
RETURNING *;

-- name: DeleteNode :exec
-- 폴더면 하위 subtree가 FK ON DELETE CASCADE로 동반 삭제된다.
DELETE FROM nodes
WHERE id = $1 AND owner_user_id = $2;

-- name: ResolveDeviceKey :one
-- 스크립트 실행 대상 = 자기 폴더에서 위로 올라가며 가장 가까운(depth 최소) device_key(main_key).
-- 없으면 0행 → 핸들러는 "귀속 worker 없음"으로 처리. depth 가드로 사이클 폭주 방지.
WITH RECURSIVE chain(id, parent_id, device_key, depth) AS (
    SELECT n.id, n.parent_id, n.device_key, 0 AS depth
    FROM nodes n
    WHERE n.id = $1 AND n.owner_user_id = $2
  UNION ALL
    SELECT n.id, n.parent_id, n.device_key, c.depth + 1
    FROM nodes n
    JOIN chain c ON n.id = c.parent_id
    WHERE c.depth < 64
)
SELECT chain.device_key FROM chain
WHERE chain.device_key IS NOT NULL
ORDER BY chain.depth
LIMIT 1;

-- name: PatchNode :one
-- 부분 업데이트(PATCH): 보낸 필드만 갱신. 프론트 {valid,value} → 핸들러가 아래로 매핑.
--  · NOT NULL 필드(name, ord) = COALESCE(narg, 컬럼): narg가 NULL이면 skip(값은 항상 non-null만 전달).
--  · NULL이 유의미한 필드(parent_id, position_x/y, device_key, content) = set_* 플래그로
--    "skip vs 갱신(NULL 포함)" 구분. set_*=true면 narg 적용(narg가 NULL이면 컬럼을 NULL로).
--    ※ set_*(=프론트 valid, "패치할지") ≠ pgtype.Valid("non-null?") — 핸들러에서 분리할 것.
--  · position_x/y는 한 좌표쌍이라 set_position 하나로 함께 제어.
--  · device_key=FOLDER만 / content=SCRIPT만 — 잘못된 kind면 CHECK가 거부(핸들러 선검증 권장).
--  · parent_id를 자기 자손으로 바꾸면 사이클 — 핸들러 책임.
UPDATE nodes
SET name       = COALESCE(sqlc.narg('name'), name),
    ord        = COALESCE(sqlc.narg('ord'),  ord),
    parent_id  = CASE WHEN sqlc.arg('set_parent')::boolean   THEN sqlc.narg('parent_id')  ELSE parent_id  END,
    position_x = CASE WHEN sqlc.arg('set_position')::boolean THEN sqlc.narg('position_x') ELSE position_x END,
    position_y = CASE WHEN sqlc.arg('set_position')::boolean THEN sqlc.narg('position_y') ELSE position_y END,
    device_key = CASE WHEN sqlc.arg('set_device')::boolean   THEN sqlc.narg('device_key') ELSE device_key END,
    content    = CASE WHEN sqlc.arg('set_content')::boolean  THEN sqlc.narg('content')    ELSE content    END,
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND owner_user_id = sqlc.arg('owner_user_id')
RETURNING *;
