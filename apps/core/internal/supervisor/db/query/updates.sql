-- 업데이트 app 보관 메타 — REF-update-deploy.md "P1 착수 계획" 단위-1(DB)

-- ===== 장비 루트 =====

-- name: ListUpdateDevices :many
-- 목록 — 빌드 수·인스턴스 수·마지막 빌드 시각
SELECT d.main_key, d.memo, d.created_at,
       (SELECT COUNT(*) FROM update_bundles b WHERE b.kind = 'build' AND b.owner = d.main_key)::bigint AS build_count,
       (SELECT COUNT(*) FROM updaters u WHERE u.main_key = d.main_key)::bigint AS instance_count,
       (SELECT MAX(b.updated_at) FROM update_bundles b WHERE b.kind = 'build' AND b.owner = d.main_key)::timestamptz AS last_build_at
FROM update_devices d
ORDER BY d.main_key;

-- name: GetUpdateDevice :one
SELECT * FROM update_devices WHERE main_key = $1;

-- name: CreateUpdateDevice :one
INSERT INTO update_devices (main_key, memo) VALUES ($1, $2) RETURNING *;

-- name: UpdateUpdateDeviceMemo :execrows
UPDATE update_devices SET memo = $2 WHERE main_key = $1;

-- name: DeleteUpdateDevice :execrows
DELETE FROM update_devices WHERE main_key = $1;

-- name: ListUpdatersByMainKey :many
SELECT * FROM updaters WHERE main_key = $1 ORDER BY sub_key;

-- ===== 번들(빌드·공용 묶음) =====

-- name: ListBundlesByOwner :many
-- locked = 잠금 행이 하나라도 있음, used_by = 이 공용을 쓰는 빌드 수
SELECT b.*,
       COALESCE(cu.nickname, '')::text AS created_by_name,
       EXISTS (SELECT 1 FROM update_locks l WHERE l.bundle_id = b.id)::bool AS locked,
       (SELECT COUNT(*) FROM update_uses u WHERE u.common_id = b.id)::bigint AS used_by,
       (SELECT COUNT(*) FROM update_files f WHERE f.bundle_id = b.id)::bigint AS file_count,
       (SELECT COALESCE(SUM(f.size), 0) FROM update_files f WHERE f.bundle_id = b.id)::bigint AS total_size
FROM update_bundles b LEFT JOIN users cu ON cu.id = b.created_by
WHERE b.kind = $1 AND b.owner = $2
ORDER BY b.created_at DESC, b.id DESC;

-- name: ListCommonBundles :many
SELECT b.*,
       COALESCE(cu.nickname, '')::text AS created_by_name,
       EXISTS (SELECT 1 FROM update_locks l WHERE l.bundle_id = b.id)::bool AS locked,
       (SELECT COUNT(*) FROM update_uses u WHERE u.common_id = b.id)::bigint AS used_by,
       (SELECT COUNT(*) FROM update_files f WHERE f.bundle_id = b.id)::bigint AS file_count,
       (SELECT COALESCE(SUM(f.size), 0) FROM update_files f WHERE f.bundle_id = b.id)::bigint AS total_size
FROM update_bundles b LEFT JOIN users cu ON cu.id = b.created_by
WHERE b.kind = 'common'
ORDER BY b.owner, b.created_at DESC, b.id DESC;

-- name: GetBundle :one
SELECT b.*,
       COALESCE(cu.nickname, '')::text AS created_by_name,
       EXISTS (SELECT 1 FROM update_locks l WHERE l.bundle_id = b.id)::bool AS locked,
       (SELECT COUNT(*) FROM update_uses u WHERE u.common_id = b.id)::bigint AS used_by
FROM update_bundles b LEFT JOIN users cu ON cu.id = b.created_by
WHERE b.kind = $1 AND b.owner = $2 AND b.version = $3;

-- name: GetBundleForUpdate :one
-- 쓰기 전 잠금 판정 — 같은 번들 동시 쓰기를 줄 세운다(폴더 sha 재계산 경합)
SELECT b.id, b.kind, b.owner, b.version,
       EXISTS (SELECT 1 FROM update_locks l WHERE l.bundle_id = b.id)::bool AS locked
FROM update_bundles b
WHERE b.kind = $1 AND b.owner = $2 AND b.version = $3
FOR UPDATE OF b;

-- name: CreateBundle :one
INSERT INTO update_bundles (kind, owner, version, note, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateBundleNote :execrows
UPDATE update_bundles SET note = $2 WHERE id = $1;

-- name: UpdateBundleSha :exec
UPDATE update_bundles SET sha = $2, updated_at = NOW() WHERE id = $1;

-- name: DeleteBundle :execrows
DELETE FROM update_bundles WHERE id = $1;

-- ===== 파일 =====

-- name: ListBundleFiles :many
SELECT * FROM update_files WHERE bundle_id = $1 ORDER BY scope, sub_key, path;

-- name: GetBundleFile :one
SELECT * FROM update_files WHERE bundle_id = $1 AND scope = $2 AND sub_key = $3 AND path = $4;

-- name: UpsertBundleFile :one
INSERT INTO update_files (bundle_id, scope, sub_key, path, size, sha, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (bundle_id, scope, sub_key, path)
DO UPDATE SET size = EXCLUDED.size, sha = EXCLUDED.sha, updated_by = EXCLUDED.updated_by, updated_at = NOW()
RETURNING *;

-- name: DeleteBundleFile :execrows
DELETE FROM update_files WHERE bundle_id = $1 AND scope = $2 AND sub_key = $3 AND path = $4;

-- name: CopyBundleFiles :exec
-- 빌드 복사 — 메타만(내용은 보관소가 복사)
INSERT INTO update_files (bundle_id, scope, sub_key, path, size, sha, updated_by)
SELECT sqlc.arg(dst_id)::bigint, f.scope, f.sub_key, f.path, f.size, f.sha, sqlc.arg(user_id)::bigint
FROM update_files f WHERE f.bundle_id = sqlc.arg(src_id)::bigint;

-- ===== 쓰는 공용 묶음 =====

-- name: DeleteBundleUses :exec
DELETE FROM update_uses WHERE bundle_id = $1;

-- name: InsertBundleUse :exec
INSERT INTO update_uses (bundle_id, common_id, ord) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;

-- name: CopyBundleUses :exec
INSERT INTO update_uses (bundle_id, common_id, ord)
SELECT sqlc.arg(dst_id)::bigint, u.common_id, u.ord FROM update_uses u WHERE u.bundle_id = sqlc.arg(src_id)::bigint;

-- name: ListBundleUses :many
-- 이 빌드가 쓰는 공용 묶음(깔 순서)
SELECT b.id, b.owner, b.version, b.sha,
       EXISTS (SELECT 1 FROM update_locks l WHERE l.bundle_id = b.id)::bool AS locked
FROM update_uses u JOIN update_bundles b ON b.id = u.common_id
WHERE u.bundle_id = $1
ORDER BY u.ord;

-- name: ListBundleUsers :many
-- 이 공용 묶음을 쓰는 빌드
SELECT b.id, b.owner, b.version,
       EXISTS (SELECT 1 FROM update_locks l WHERE l.bundle_id = b.id)::bool AS locked
FROM update_uses u JOIN update_bundles b ON b.id = u.bundle_id
WHERE u.common_id = $1
ORDER BY b.owner, b.version;
