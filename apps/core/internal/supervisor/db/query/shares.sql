-- 공유 판정 쿼리. 설계 → REF-process-share-db.md
-- 규약: 단계 계산은 node_permission/process_permission 함수만 호출한다(가까운 공유 로직을 여기에 복사하지 말 것).
-- 단계 값: 0 접근 없음 / 10 보기 / 20 조작 / 30 편집 / 40 소유자 → 비교는 항상 >=.

-- name: NodePermission :one
SELECT node_permission(sqlc.arg(user_id), sqlc.arg(node_id))::smallint AS permission;

-- name: ProcessPermission :one
-- process 없음 = NULL → 0으로.
SELECT COALESCE(process_permission(sqlc.arg(user_id), sqlc.arg(uid)), 0)::smallint AS permission;

-- name: ListChildrenVisible :many
-- 한 폴더의 자식 + 각 자식의 단계. parent_perm = 호출자가 NodePermission으로 먼저 구한 부모 단계(소유자면 40).
-- 자식 단계 = 자식 자신의 공유 행 > 부모 단계(가까운 공유를 한 단계만 적용). 0이면 제외.
-- 부모 소유자(40)는 공유 행을 무시 — 자기 자신에게 건 공유 행이 있어도 node_permission과 같은 답(40).
-- parent_id NULL(루트)은 맞는 행이 없음 → 루트 목록에 쓰지 말 것.
SELECT sqlc.embed(n), (CASE WHEN sqlc.arg(parent_perm)::smallint = 40 THEN 40
                  ELSE COALESCE(s.permission, sqlc.arg(parent_perm)::smallint) END)::smallint AS permission
FROM nodes n
LEFT JOIN node_shares s ON s.node_id = n.id AND s.user_id = sqlc.arg(user_id)
WHERE n.parent_id = sqlc.arg(parent_id)::bigint
  AND (sqlc.arg(parent_perm)::smallint = 40 OR COALESCE(s.permission, sqlc.arg(parent_perm)::smallint) > 0)
ORDER BY n.ord, n.name, n.id;

-- name: ListVisibleLiveProcesses :many
-- 소켓 연결 때 구독 대상 = 소유 + 공유받은 살아 있는 process.
SELECT sqlc.embed(p), process_permission(sqlc.arg(user_id), p.uid)::smallint AS permission
FROM processes p
WHERE p.status IN ('PENDING','PROCESS')
  AND process_permission(sqlc.arg(user_id), p.uid) > 0
ORDER BY p.created_at ASC;

-- name: ListVisibleProcessesByUids :many
-- 요청한 uid만, 상태 무관. 볼 수 없는 것·없는 uid는 빠짐.
SELECT sqlc.embed(p), process_permission(sqlc.arg(user_id), p.uid)::smallint AS permission
FROM processes p
WHERE p.uid = ANY(sqlc.arg(uids)::text[])
  AND process_permission(sqlc.arg(user_id), p.uid) > 0
ORDER BY p.created_at ASC;

-- name: ListProcessViewers :many
-- 볼 수 있는 계정 전부(소유자 포함). chain은 후보 수집용, 단계 판정은 process_permission.
WITH RECURSIVE chain(id, parent_id, depth) AS (
    SELECT n.id, n.parent_id, 0
    FROM nodes n
    JOIN processes p ON p.node_id = n.id
    WHERE p.uid = sqlc.arg(uid)
  UNION ALL
    SELECT n.id, n.parent_id, c.depth + 1
    FROM nodes n
    JOIN chain c ON n.id = c.parent_id
    WHERE c.depth < 64
), cand(user_id) AS (
    SELECT owner_user_id FROM processes WHERE uid = sqlc.arg(uid)
  UNION
    SELECT ps.user_id FROM process_shares ps WHERE ps.process_uid = sqlc.arg(uid)
  UNION
    SELECT s.user_id FROM chain c JOIN node_shares s ON s.node_id = c.id
)
SELECT cand.user_id::bigint AS user_id,
       process_permission(cand.user_id, sqlc.arg(uid))::smallint AS permission
FROM cand
WHERE process_permission(cand.user_id, sqlc.arg(uid)) > 0;

-- ===== 재계산(reconcileAccess, 공유-d 실시간-1) — REF-process-share-reconcile.md =====

-- name: ListProcessPermissionMatrix :many
-- 계정 배열 × uid 배열의 단계. 없는 uid = 0.
SELECT u.user_id::bigint AS user_id, p.uid::text AS uid,
       COALESCE(process_permission(u.user_id, p.uid), 0)::smallint AS permission
FROM unnest(sqlc.arg(user_ids)::bigint[]) AS u(user_id)
CROSS JOIN unnest(sqlc.arg(uids)::text[]) AS p(uid);

-- name: ListNodePermissionMatrix :many
-- 계정 배열 × 노드 배열의 단계(원격 active 배지 — REF-node-ui-active.md). 없는 노드 = 0.
SELECT u.user_id::bigint AS user_id, n.node_id::bigint AS node_id,
       COALESCE(node_permission(u.user_id, n.node_id), 0)::smallint AS permission
FROM unnest(sqlc.arg(user_ids)::bigint[]) AS u(user_id)
CROSS JOIN unnest(sqlc.arg(node_ids)::bigint[]) AS n(node_id);

-- name: ListLiveProcessesUnder :many
-- 노드 자신과 하위 스크립트에서 실행된 살아 있는(PENDING/PROCESS) process.
WITH RECURSIVE sub(id, depth) AS (
    SELECT sqlc.arg(node_id)::bigint, 0
  UNION ALL
    SELECT n.id, s.depth + 1
    FROM nodes n
    JOIN sub s ON n.parent_id = s.id
    WHERE s.depth < 64
)
SELECT p.uid
FROM processes p
JOIN sub ON p.node_id = sub.id
WHERE p.status IN ('PENDING','PROCESS')
ORDER BY p.created_at ASC;

-- name: ListShareUsersAround :many
-- 영향 계정 후보(판정 아님): 노드의 조상·자신·하위에 공유 행이 있는 계정 + 하위 살아 있는 process의 단독 공유 계정.
-- 이동은 옮기기 전·후 두 번 불러 합친다(조상이 바뀜), 삭제는 전에만.
WITH RECURSIVE up(id, parent_id, depth) AS (
    SELECT n.id, n.parent_id, 0
    FROM nodes n
    WHERE n.id = sqlc.arg(node_id)::bigint
  UNION ALL
    SELECT n.id, n.parent_id, u.depth + 1
    FROM nodes n
    JOIN up u ON n.id = u.parent_id
    WHERE u.depth < 64
), down(id, depth) AS (
    SELECT sqlc.arg(node_id)::bigint, 0
  UNION ALL
    SELECT n.id, d.depth + 1
    FROM nodes n
    JOIN down d ON n.parent_id = d.id
    WHERE d.depth < 64
)
SELECT s.user_id
FROM node_shares s
WHERE s.node_id IN (SELECT id FROM up UNION SELECT id FROM down)
UNION
SELECT ps.user_id
FROM process_shares ps
JOIN processes p ON p.uid = ps.process_uid
JOIN down ON p.node_id = down.id
WHERE p.status IN ('PENDING','PROCESS');

-- name: HasLowerShareBelow :one
-- 노드 아래(자기 제외)에 그 계정의 min_perm 미만 공유 행이 있나. 받은 쪽 삭제·이동 검사(차단된 하위를 같이 지우거나 옮기는 것 방지).
WITH RECURSIVE sub(id, depth) AS (
    SELECT n.id, 1
    FROM nodes n
    WHERE n.parent_id = sqlc.arg(node_id)::bigint
  UNION ALL
    SELECT n.id, s.depth + 1
    FROM nodes n
    JOIN sub s ON n.parent_id = s.id
    WHERE s.depth < 64
)
SELECT (COUNT(*) > 0)::boolean AS found
FROM sub
JOIN node_shares ns ON ns.node_id = sub.id
WHERE ns.user_id = sqlc.arg(user_id)
  AND ns.permission < sqlc.arg(min_perm)::smallint;

-- ===== 공유 걸기·풀기·목록(공유-c(공유 API)) — 소유자 검사는 핸들러(router/share.go) =====

-- name: GetNodeShare :one
SELECT * FROM node_shares
WHERE node_id = $1 AND user_id = $2;

-- name: UpsertNodeShare :one
-- 같은 노드·계정 = 단계만 바꿈(SH4(공유 대상 찾기)).
INSERT INTO node_shares (node_id, user_id, permission)
VALUES ($1, $2, $3)
ON CONFLICT (node_id, user_id) DO UPDATE
SET permission = EXCLUDED.permission,
    updated_at = NOW()
RETURNING *;

-- name: DeleteNodeShare :execrows
DELETE FROM node_shares
WHERE node_id = $1 AND user_id = $2;

-- name: CountDirectShares :many
-- 노드별 직접 건 공유 수(차단(0) 제외) — 폴더 목록 행의 shareCount(O3). 공유가 없는 노드는 행이 없음.
SELECT node_id, COUNT(*)::bigint AS cnt
FROM node_shares
WHERE node_id = ANY(sqlc.arg(node_ids)::bigint[])
  AND permission > 0
GROUP BY node_id;

-- name: ListMyNodeShares :many
-- 내가 소유한 노드에 직접 건 공유 전부(차단 포함) — 내가 공유한 것(O5). 해제는 기존 DELETE /nodes/:id/shares/:userId.
SELECT n.id AS node_id, n.name AS node_name, s.user_id, u.nickname, s.permission::smallint AS permission
FROM node_shares s
JOIN nodes n ON n.id = s.node_id
JOIN users u ON u.id = s.user_id
WHERE n.owner_user_id = sqlc.arg(owner_user_id)
ORDER BY n.name, n.id, u.nickname, u.id;

-- name: ListMyProcessShares :many
-- 내가 소유한 살아 있는 process에 직접 건 공유(O5). 이름 = 연결된 스크립트 이름, 없으면 명령어.
SELECT p.uid AS process_uid, COALESCE(n.name, p.cmd)::text AS process_name, ps.user_id, u.nickname, ps.permission::smallint AS permission
FROM process_shares ps
JOIN processes p ON p.uid = ps.process_uid
LEFT JOIN nodes n ON n.id = p.node_id
JOIN users u ON u.id = ps.user_id
WHERE p.owner_user_id = sqlc.arg(owner_user_id)
  AND p.status IN ('PENDING','PROCESS')
ORDER BY p.created_at, p.uid, u.nickname, u.id;

-- name: ListShareTargets :many
-- 내가 직접 공유를 건 계정(내 노드 + 내 살아 있는 process) — 접속 조회(O4)를 허용하는 범위.
SELECT s.user_id::bigint AS user_id FROM node_shares s JOIN nodes n ON n.id = s.node_id WHERE n.owner_user_id = sqlc.arg(owner_user_id)
UNION
SELECT ps.user_id FROM process_shares ps JOIN processes p ON p.uid = ps.process_uid WHERE p.owner_user_id = sqlc.arg(owner_user_id);

-- name: ListNodeShareEntries :many
-- 이 노드에 실제로 적용되는 공유 전부(직접 + 상위에서 이어진 것, 차단 포함, 소유자 제외) + 출처.
-- chain은 후보 수집용, 가까운 공유는 nearest_node_share(SC1(출처 계산 위치)).
WITH RECURSIVE chain(id, parent_id, depth) AS (
    SELECT n.id, n.parent_id, 0
    FROM nodes n
    WHERE n.id = sqlc.arg(node_id)::bigint
  UNION ALL
    SELECT n.id, n.parent_id, c.depth + 1
    FROM nodes n
    JOIN chain c ON n.id = c.parent_id
    WHERE c.depth < 64
), cand(user_id) AS (
    SELECT DISTINCT s.user_id
    FROM chain c
    JOIN node_shares s ON s.node_id = c.id
)
SELECT u.id AS user_id, u.nickname, src.id AS source_node_id, src.name AS source_name, s.permission
FROM cand
JOIN users u ON u.id = cand.user_id
JOIN node_shares s ON s.user_id = cand.user_id
                  AND s.node_id = nearest_node_share(cand.user_id, sqlc.arg(node_id)::bigint)
JOIN nodes src ON src.id = s.node_id
WHERE u.id <> (SELECT owner_user_id FROM nodes WHERE id = sqlc.arg(node_id)::bigint)
ORDER BY u.nickname, u.id;

-- name: GetProcessShare :one
SELECT * FROM process_shares
WHERE process_uid = $1 AND user_id = $2;

-- name: UpsertProcessShare :one
INSERT INTO process_shares (process_uid, user_id, permission)
VALUES ($1, $2, $3)
ON CONFLICT (process_uid, user_id) DO UPDATE
SET permission = EXCLUDED.permission,
    updated_at = NOW()
RETURNING *;

-- name: DeleteProcessShare :execrows
DELETE FROM process_shares
WHERE process_uid = $1 AND user_id = $2;

-- name: ListProcessShareEntries :many
-- process에 적용되는 공유 전부(process 단독 + 노드에서 이어진 것, 차단 포함, 소유자 제외). direct = process 단독 공유 행.
WITH RECURSIVE chain(id, parent_id, depth) AS (
    SELECT n.id, n.parent_id, 0
    FROM nodes n
    JOIN processes p ON p.node_id = n.id
    WHERE p.uid = sqlc.arg(uid)::text
  UNION ALL
    SELECT n.id, n.parent_id, c.depth + 1
    FROM nodes n
    JOIN chain c ON n.id = c.parent_id
    WHERE c.depth < 64
), cand(user_id) AS (
    SELECT ps.user_id FROM process_shares ps WHERE ps.process_uid = sqlc.arg(uid)::text
  UNION
    SELECT s.user_id FROM chain c JOIN node_shares s ON s.node_id = c.id
)
SELECT u.id AS user_id, u.nickname,
       process_permission(u.id, sqlc.arg(uid)::text)::smallint AS permission,
       (ps.user_id IS NOT NULL)::boolean AS direct
FROM cand
JOIN users u ON u.id = cand.user_id
LEFT JOIN process_shares ps ON ps.process_uid = sqlc.arg(uid)::text AND ps.user_id = u.id
WHERE u.id <> (SELECT owner_user_id FROM processes WHERE uid = sqlc.arg(uid)::text)
ORDER BY u.nickname, u.id;

-- ===== 노드 응답(실시간-5) — SD8(받는 쪽 루트 목록) =====

-- name: ListSharedRoots :many
-- 공유받은 최상위 폴더: 내 공유 행이 있는 남의 폴더 중 단계 > 0이고 부모가 없거나 부모를 볼 수 없는 것. 단계 판정은 node_permission만.
SELECT sqlc.embed(n), node_permission(sqlc.arg(user_id), n.id)::smallint AS permission, u.nickname AS owner_nickname
FROM node_shares s
JOIN nodes n ON n.id = s.node_id
JOIN users u ON u.id = n.owner_user_id
WHERE s.user_id = sqlc.arg(user_id)
  AND n.kind = 'FOLDER'
  AND n.owner_user_id <> sqlc.arg(user_id)
  AND node_permission(sqlc.arg(user_id), n.id) > 0
  AND (n.parent_id IS NULL OR node_permission(sqlc.arg(user_id), n.parent_id) = 0)
ORDER BY u.nickname, n.name, n.id;

-- ===== 스크립트 행 "실행 중 N"(실시간-6a) =====

-- name: ListVisibleRunsInFolder :many
-- 폴더의 직접 자식 스크립트에서 실행된 process 중 내가 볼 수 있는 것 = 살아 있는 것 + 내 보드에 터미널 타일이 있는 끝난 것(R2-a(끝난 process)).
-- 단계 판정은 process_permission만.
SELECT p.uid, p.node_id, p.status, process_permission(sqlc.arg(user_id), p.uid)::smallint AS permission,
       p.started_by, COALESCE(u.nickname, '')::text AS started_by_nickname,
       (p.kill_requested_at IS NOT NULL)::boolean AS terminating, p.kill_forced, n.name, p.cmd, p.cwd, p.args
FROM processes p
JOIN nodes n ON n.id = p.node_id
LEFT JOIN users u ON u.id = p.started_by
WHERE n.parent_id IS NOT DISTINCT FROM sqlc.narg(folder_id)::bigint -- NULL = 루트
  AND n.kind = 'SCRIPT'
  AND (p.status IN ('PENDING','PROCESS') OR p.uid = ANY(sqlc.arg(tile_uids)::text[]))
  AND process_permission(sqlc.arg(user_id), p.uid) > 0
ORDER BY p.created_at ASC;

-- name: ListVisibleRunsInSubtree :many
-- 터미널 탭(개편-3(목록 API)): 폴더와 그 하위 전체 스크립트의 process — 위와 같은 조건. folder_id NULL = 기본 루트 = 볼 수 있는 process 전부
-- (단독 공유 process 포함, R3(경로 없는 process)). 장비 실행(node_id NULL)은 dir_node_id가 하위면 — Q10(루트가 아닐 때), 이름은 router가 cmd로
WITH RECURSIVE sub AS (
    SELECT id FROM nodes WHERE id = sqlc.narg(folder_id)::bigint
    UNION ALL
    SELECT c.id FROM nodes c JOIN sub ON c.parent_id = sub.id
)
SELECT p.uid, p.node_id, p.status, process_permission(sqlc.arg(user_id), p.uid)::smallint AS permission,
       p.started_by, COALESCE(u.nickname, '')::text AS started_by_nickname,
       (p.kill_requested_at IS NOT NULL)::boolean AS terminating, p.kill_forced, COALESCE(n.name, '')::text AS name, p.cmd, p.cwd, p.args
FROM processes p
LEFT JOIN nodes n ON n.id = p.node_id
LEFT JOIN users u ON u.id = p.started_by
WHERE (sqlc.narg(folder_id)::bigint IS NULL OR p.node_id IN (SELECT id FROM sub) OR p.dir_node_id IN (SELECT id FROM sub))
  AND (n.kind = 'SCRIPT' OR p.node_id IS NULL)
  AND (p.status IN ('PENDING','PROCESS') OR p.uid = ANY(sqlc.arg(tile_uids)::text[]))
  AND process_permission(sqlc.arg(user_id), p.uid) > 0
ORDER BY p.created_at ASC;

-- name: ListVisibleRunsOnDevice :many
-- 루트 장비 타일 터미널 탭(이동-3(장비 목록 API), Q25(루트 장비 탭 범위)): 그 장비(main_key, 인스턴스 전부)의 process — 위와 같은 조건.
-- 어느 타일·폴더에서 띄웠든, 단독 공유 포함. 서브키는 마지막 '#' 뒤(splitInstanceKey와 같게)
SELECT p.uid, p.node_id, p.status, process_permission(sqlc.arg(user_id), p.uid)::smallint AS permission,
       p.started_by, COALESCE(u.nickname, '')::text AS started_by_nickname,
       (p.kill_requested_at IS NOT NULL)::boolean AS terminating, p.kill_forced, COALESCE(n.name, '')::text AS name, p.cmd, p.cwd, p.args
FROM processes p
LEFT JOIN nodes n ON n.id = p.node_id
LEFT JOIN users u ON u.id = p.started_by
WHERE regexp_replace(p.device_key, '#[^#]*$', '') = sqlc.arg(main_key)::text
  AND (n.kind = 'SCRIPT' OR p.node_id IS NULL)
  AND (p.status IN ('PENDING','PROCESS') OR p.uid = ANY(sqlc.arg(tile_uids)::text[]))
  AND process_permission(sqlc.arg(user_id), p.uid) > 0
ORDER BY p.created_at ASC;
