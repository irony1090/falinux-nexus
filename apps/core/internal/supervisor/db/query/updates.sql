-- 업데이트 app 보관 메타 — REF-update-deploy.md "P1 착수 계획" 단위-1(DB)

-- ===== 장비 루트 =====

-- name: ListUpdateDevices :many
-- 목록 — 빌드 수·인스턴스 수·마지막 빌드 시각
SELECT d.main_key, d.memo, d.created_at,
       (SELECT COUNT(*) FROM update_bundles b WHERE b.kind = 'build' AND b.owner = d.main_key)::bigint AS build_count,
       (SELECT COUNT(*) FROM updaters u WHERE u.main_key = d.main_key)::bigint AS instance_count,
       (SELECT MAX(b.updated_at) FROM update_bundles b WHERE b.kind = 'build' AND b.owner = d.main_key)::timestamptz AS last_build_at,
       (d.api_key_hash <> '')::bool AS has_key
FROM update_devices d
ORDER BY d.main_key;

-- name: GetUpdateDevice :one
SELECT * FROM update_devices WHERE main_key = $1;

-- name: CreateUpdateDevice :one
INSERT INTO update_devices (main_key, memo, api_key_hash, api_key_enc, key_issued_at) VALUES ($1, $2, $3, $4, NOW()) RETURNING *;

-- name: UpdateUpdateDeviceKey :execrows
-- 발급·재발급 — 이미 붙은 연결은 그대로, 다음 접속부터 새 키
UPDATE update_devices SET api_key_hash = $2, api_key_enc = $3, key_issued_at = NOW() WHERE main_key = $1;

-- name: UpdateUpdateDeviceMemo :execrows
UPDATE update_devices SET memo = $2 WHERE main_key = $1;

-- name: DeleteUpdateDevice :execrows
DELETE FROM update_devices WHERE main_key = $1;

-- name: ListUpdatersByMainKey :many
SELECT * FROM updaters WHERE main_key = $1 ORDER BY created_at, sub_key;

-- name: GetUpdater :one
SELECT * FROM updaters WHERE main_key = $1 AND sub_key = $2;

-- name: UpsertUpdater :exec
-- updater 접속(HELLO) — 보고값 갱신. 별칭·메모는 그대로
INSERT INTO updaters (main_key, sub_key, os, arch, install_dir, installed_version, last_seen_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (main_key, sub_key)
DO UPDATE SET os = EXCLUDED.os, arch = EXCLUDED.arch, install_dir = EXCLUDED.install_dir,
              installed_version = EXCLUDED.installed_version, last_seen_at = NOW();

-- name: TouchUpdater :exec
UPDATE updaters SET last_seen_at = NOW() WHERE main_key = $1 AND sub_key = $2;

-- name: UpdateUpdaterLabel :execrows
UPDATE updaters SET label = $3, memo = $4 WHERE main_key = $1 AND sub_key = $2;

-- name: UpdateUpdaterInstallDir :execrows
UPDATE updaters SET install_dir = $3 WHERE main_key = $1 AND sub_key = $2;

-- name: UpdateUpdaterInstalled :exec
UPDATE updaters SET installed_version = $3 WHERE main_key = $1 AND sub_key = $2;

-- name: DeleteUpdater :execrows
DELETE FROM updaters WHERE main_key = $1 AND sub_key = $2;

-- ===== 번들(빌드·공용 묶음) =====

-- name: ListBundlesByOwner :many
-- locked = 잠금 행이 하나라도 있음, used_by = 이 공용을 쓰는 빌드 수
SELECT b.*,
       COALESCE(cu.nickname, '')::text AS created_by_name,
       EXISTS (SELECT 1 FROM update_locks l WHERE l.bundle_id = b.id)::bool AS locked,
       (SELECT COUNT(*) FROM update_uses u WHERE u.common_id = b.id)::bigint AS used_by,
       (SELECT COUNT(*) FROM update_files f WHERE f.bundle_id = b.id)::bigint AS file_count,
       (SELECT COALESCE(SUM(f.size), 0) FROM update_files f WHERE f.bundle_id = b.id)::bigint AS total_size,
       (SELECT COALESCE(MAX(d.id), 0) FROM update_deployments d WHERE d.bundle_id = b.id)::bigint AS last_deploy_id
FROM update_bundles b LEFT JOIN users cu ON cu.id = b.created_by
WHERE b.kind = $1 AND b.owner = $2
ORDER BY b.created_at DESC, b.id DESC;

-- name: ListCommonBundles :many
SELECT b.*,
       COALESCE(cu.nickname, '')::text AS created_by_name,
       EXISTS (SELECT 1 FROM update_locks l WHERE l.bundle_id = b.id)::bool AS locked,
       (SELECT COUNT(*) FROM update_uses u WHERE u.common_id = b.id)::bigint AS used_by,
       (SELECT COUNT(*) FROM update_files f WHERE f.bundle_id = b.id)::bigint AS file_count,
       (SELECT COALESCE(SUM(f.size), 0) FROM update_files f WHERE f.bundle_id = b.id)::bigint AS total_size,
       (SELECT COALESCE(MAX(d.id), 0) FROM update_deployments d WHERE d.bundle_id = b.id)::bigint AS last_deploy_id
FROM update_bundles b LEFT JOIN users cu ON cu.id = b.created_by
WHERE b.kind = 'common'
ORDER BY b.owner, b.created_at DESC, b.id DESC;

-- name: GetBundle :one
SELECT b.*,
       COALESCE(cu.nickname, '')::text AS created_by_name,
       EXISTS (SELECT 1 FROM update_locks l WHERE l.bundle_id = b.id)::bool AS locked,
       (SELECT COUNT(*) FROM update_uses u WHERE u.common_id = b.id)::bigint AS used_by,
       (SELECT COALESCE(MAX(d.id), 0) FROM update_deployments d WHERE d.bundle_id = b.id)::bigint AS last_deploy_id,
       update_lockable(b.id)::bool AS lockable
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
-- mode = 새 파일일 때만(덮어쓰기는 기존 권한 유지)
INSERT INTO update_files (bundle_id, scope, sub_key, path, size, sha, updated_by, mode)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (bundle_id, scope, sub_key, path)
DO UPDATE SET size = EXCLUDED.size, sha = EXCLUDED.sha, updated_by = EXCLUDED.updated_by, updated_at = NOW()
RETURNING *;

-- name: UpdateBundleFileMode :execrows
UPDATE update_files SET mode = $5, updated_by = $6, updated_at = NOW()
WHERE bundle_id = $1 AND scope = $2 AND sub_key = $3 AND path = $4;

-- name: DeleteBundleFile :execrows
DELETE FROM update_files WHERE bundle_id = $1 AND scope = $2 AND sub_key = $3 AND path = $4;

-- name: CopyBundleFiles :exec
-- 빌드 복사 — 메타만(내용은 보관소가 복사)
INSERT INTO update_files (bundle_id, scope, sub_key, path, size, sha, updated_by, mode)
SELECT sqlc.arg(dst_id)::bigint, f.scope, f.sub_key, f.path, f.size, f.sha, sqlc.arg(user_id)::bigint, f.mode
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

-- ===== updater 접속 =====

-- name: GetUpdateDeviceAuth :one
SELECT api_key_hash FROM update_devices WHERE main_key = $1;

-- name: GetBundleByID :one
SELECT * FROM update_bundles WHERE id = $1;

-- ===== 단계 실행 기록 =====

-- name: CreateStepRun :one
INSERT INTO update_step_runs (deployment_id, bundle_id, main_key, sub_key, version, list, step_id, step_kind, log_file, started_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: FinishStepRun :exec
UPDATE update_step_runs SET result = $2, exit_code = $3, log_tail = $4, ended_at = NOW() WHERE id = $1;

-- name: GetStepRun :one
SELECT * FROM update_step_runs WHERE id = $1;

-- name: ListStepRuns :many
SELECT * FROM update_step_runs WHERE main_key = $1 AND sub_key = $2 ORDER BY id DESC LIMIT $3;

-- name: ListTrialRuns :many
-- 시험 타일 — 배포 밖 단독 실행만(배포 기록이 섞이면 LIMIT 안에서 밀려난다)
SELECT * FROM update_step_runs WHERE main_key = $1 AND sub_key = $2 AND version = $3 AND deployment_id IS NULL ORDER BY id DESC LIMIT $4;

-- name: CancelLostStepRuns :many
-- updater 재접속 때 — 그 인스턴스의 끝나지 않은 기록 중 updater가 모르는 것(updater 재시작 등)
UPDATE update_step_runs SET result = 'cancel', log_tail = log_tail || E'\n[updater가 이 실행을 모름 — 결과를 받지 못함]', ended_at = NOW()
WHERE main_key = $1 AND sub_key = $2 AND result = 'running' AND NOT (id = ANY(sqlc.arg(keep)::bigint[]))
RETURNING id;

-- name: CreateWaitStepRun :one
-- 배포가 미리 만드는 단계 기록(아직 안 돎) — REF-update-p3.md 단위-1(DB)
INSERT INTO update_step_runs (deployment_id, bundle_id, main_key, sub_key, version, list, step_id, step_kind, started_by, result)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'wait')
RETURNING *;

-- name: StartStepRun :execrows
UPDATE update_step_runs SET result = 'running', started_at = NOW() WHERE id = $1 AND result = 'wait';

-- name: ListDeploymentRuns :many
SELECT * FROM update_step_runs WHERE deployment_id = $1 ORDER BY id;

-- name: SkipWaitRuns :exec
UPDATE update_step_runs SET result = 'skip', ended_at = NOW() WHERE deployment_id = $1 AND list = $2 AND result = 'wait';

-- name: DeleteWaitRuns :exec
DELETE FROM update_step_runs WHERE deployment_id = $1 AND list = $2 AND result = 'wait';

-- ===== 배포 =====

-- name: CreateDeployment :one
INSERT INTO update_deployments (bundle_id, main_key, sub_key, version, sha, uses_sig, auto_rollback, started_by)
VALUES ($1, $2, $3, $4, $5, update_uses_sig(sqlc.arg(sig_bundle)::bigint), $6, $7)
RETURNING *;

-- name: GetDeployment :one
SELECT d.*, COALESCE(u.nickname, '')::text AS started_by_name
FROM update_deployments d LEFT JOIN users u ON u.id = d.started_by
WHERE d.id = $1;

-- name: GetDeploymentForUpdate :one
SELECT * FROM update_deployments WHERE id = $1 FOR UPDATE;

-- name: FinishDeployment :execrows
UPDATE update_deployments SET result = $2, msg = $3, ended_at = NOW() WHERE id = $1 AND result = 'running';

-- name: SetRollbackResult :execrows
UPDATE update_deployments SET rollback_result = $2, msg = CASE WHEN sqlc.arg(msg)::text = '' THEN msg ELSE sqlc.arg(msg)::text END
WHERE id = $1;

-- name: LatestLiveDeploymentID :one
-- 되돌리기 차례 = 그 인스턴스에서 아직 되돌리지 않은 가장 나중 배포(0 = 없음) — REF-update-rules.md "되돌리기"
SELECT COALESCE(MAX(id), 0)::bigint FROM update_deployments WHERE main_key = $1 AND sub_key = $2 AND rollback_result <> 'ok';

-- name: CancelLostDeployments :many
-- updater 재접속 때 — 그 인스턴스의 끝나지 않은 배포·되돌리기 중 updater가 모르는 것(updater 재시작 = 중단, REF-update-p3.md "합의한 해석")
UPDATE update_deployments SET
    result = CASE WHEN result = 'running' THEN 'cancel' ELSE result END,
    rollback_result = CASE WHEN rollback_result = 'running' THEN 'cancel' ELSE rollback_result END,
    msg = 'updater가 이 배포를 모름 — updater 재시작 등으로 중단',
    ended_at = COALESCE(ended_at, NOW())
WHERE main_key = $1 AND sub_key = $2 AND (result = 'running' OR rollback_result = 'running') AND NOT (id = ANY(sqlc.arg(keep)::bigint[]))
RETURNING id;

-- name: SkipWaitRunsOf :exec
UPDATE update_step_runs SET result = 'skip', ended_at = NOW() WHERE deployment_id = ANY(sqlc.arg(ids)::bigint[]) AND result = 'wait';

-- name: IsLockable :one
SELECT update_lockable($1)::bool;

-- name: LockBuildWithUses :exec
-- 빌드 + 쓰는 공용 묶음마다 잠금 행(via = 이 빌드), sha = 지금 값
INSERT INTO update_locks (bundle_id, via_build_id, sha, locked_by)
SELECT b.id, sqlc.arg(build_id)::bigint, b.sha, sqlc.arg(locked_by)::bigint FROM update_bundles b
WHERE b.id = sqlc.arg(build_id)::bigint OR b.id IN (SELECT common_id FROM update_uses WHERE bundle_id = sqlc.arg(build_id)::bigint);

-- name: SetStepRunLogTail :exec
UPDATE update_step_runs SET log_tail = $2 WHERE id = $1;
