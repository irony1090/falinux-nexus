-- worker_instances 명부 쿼리. 설계 → REF-node-label.md "roster 계획"

-- name: UpsertWorkerInstance :exec
-- 등록 때 — 처음이면 추가, 있으면 last_seen·os만 갱신(first_seen 유지)
INSERT INTO worker_instances (main_key, sub_key, os)
VALUES ($1, $2, $3)
ON CONFLICT (main_key, sub_key) DO UPDATE SET last_seen = NOW(), os = EXCLUDED.os;

-- name: TouchWorkerInstance :exec
-- 끊길 때
UPDATE worker_instances SET last_seen = NOW()
WHERE main_key = $1 AND sub_key = $2;

-- name: ListWorkerInstances :many
-- main_key = '' 이면 전체
SELECT * FROM worker_instances
WHERE sqlc.arg(main_key)::text = '' OR main_key = sqlc.arg(main_key)::text
ORDER BY main_key, sub_key;

-- name: DeleteWorkerInstance :execrows
DELETE FROM worker_instances
WHERE main_key = $1 AND sub_key = $2;

-- name: UpdateWorkerInstanceLabel :execrows
-- 인스턴스 라벨(회원-Q11(인스턴스 구분))
UPDATE worker_instances SET label = $3
WHERE main_key = $1 AND sub_key = $2;

-- name: ListOwnedWorkerInstances :many
-- 내 장비 페이지 — 내 장비 이름의 인스턴스만
SELECT wi.* FROM worker_instances wi
JOIN devices d ON d.main_key = wi.main_key
WHERE d.owner_user_id = $1
ORDER BY wi.main_key, wi.sub_key;

-- name: ListVisibleWorkerInstances :many
-- 명부 보이는 범위(회원-6(실행 권한)) = 내 장비 이름 + 내가 볼 수 있는 폴더에 지정된 장비. main_key = '' 이면 그 안 전부
SELECT wi.* FROM worker_instances wi
WHERE (sqlc.arg(main_key)::text = '' OR wi.main_key = sqlc.arg(main_key)::text)
  AND (wi.main_key IN (SELECT d.main_key FROM devices d WHERE d.owner_user_id = sqlc.arg(user_id))
    OR wi.main_key IN (SELECT n.device_key FROM nodes n
                       WHERE n.device_key IS NOT NULL AND node_permission(sqlc.arg(user_id), n.id) >= 10))
ORDER BY wi.main_key, wi.sub_key;

-- name: FilterDeviceWatchers :many
-- 장비 상태 알림(DEVICE:CHANGE) 받을 계정 = user_ids 중 관리자 + 주인 + 그 장비가 지정된 폴더를 볼 수 있는 계정(ListVisibleWorkerInstances와 같은 범위)
SELECT u.id FROM users u
WHERE u.id = ANY(sqlc.arg(user_ids)::bigint[])
  AND ('admin' = ANY(u.perms)
    OR EXISTS (SELECT 1 FROM devices d WHERE d.main_key = sqlc.arg(main_key)::text AND d.owner_user_id = u.id)
    OR EXISTS (SELECT 1 FROM nodes n WHERE n.device_key = sqlc.arg(main_key)::text AND node_permission(u.id, n.id) >= 10));

-- name: ListOwnedMainKeys :many
SELECT main_key FROM devices WHERE owner_user_id = $1;
