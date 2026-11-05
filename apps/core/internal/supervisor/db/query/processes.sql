-- name: CreateProcess :one
INSERT INTO processes (
    uid,
    type,
    owner_user_id,
    node_id,
    device_key,
    cmd,
    args,
    env,
    cwd,
    rows,
    cols,
    started_by,
    dir_node_id
) VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13 )
RETURNING *;

-- name: GetProcess :one
SELECT * FROM processes WHERE uid = $1;

-- name: MarkProcessRunning :one
UPDATE processes
SET status     = 'PROCESS',
    pid        = $2,
    started_at = NOW(),
    updated_at = NOW()
WHERE uid = $1
RETURNING *;

-- name: MarkProcessPending :one
UPDATE processes
SET status     = 'PENDING',
    updated_at = NOW()
WHERE uid = $1
RETURNING *;

-- name: MarkProcessDone :one
UPDATE processes
SET status      = $2,
    exit_code   = $3,
    finished_at = NOW(),
    updated_at  = NOW()
WHERE uid = $1
RETURNING *;

-- name: UpdateProcessLayout :one
UPDATE processes
SET rows       = $2,
    cols       = $3,
    updated_at = NOW()
WHERE uid = $1
RETURNING *;

-- name: ListProcessesByOwner :many
SELECT * FROM processes
WHERE owner_user_id = $1
ORDER BY created_at DESC;

-- name: ListActiveByDevice :many
SELECT * FROM processes
WHERE device_key = $1
  AND status IN ('PENDING','PROCESS');

-- name: ListOwnLiveProcesses :many
-- 고아 안전망 대상 = 내가 실행한 살아 있는 process — WS3(고아 안전망 위치). 내 소유라도 남이 실행한 것은 받는 타일 규칙을 따른다 — BD8(고아 안전망).
-- 종료 요청된 것은 빼야 닫기가 유지된다 — KP3(자동으로 다시 붙이기). node_id NULL(장비 실행, 폴더-4(실행))도 터미널 타일이 받는다
SELECT uid, node_id
FROM processes
WHERE started_by = $1
  AND status IN ('PENDING','PROCESS')
  AND kill_requested_at IS NULL
ORDER BY created_at ASC;

-- name: RequestProcessKill :one
-- 종료 요청 기록. 두 번째 요청 시각은 덮지 않고, force는 한 번 켜지면 유지. 끝난 process면 행 없음
UPDATE processes
SET kill_requested_at = COALESCE(kill_requested_at, NOW()),
    kill_forced       = kill_forced OR sqlc.arg(force)::boolean,
    updated_at        = NOW()
WHERE uid = sqlc.arg(uid)
  AND status IN ('PENDING','PROCESS')
RETURNING *;

-- name: ForceAbandonedKills :many
-- 종료-4(모두 닫음): 종료 요청 뒤 grace_secs가 지났고 min_perm(조작) 이상 계정의 어느 보드에도 터미널 타일이 없으면 강제로 바꾼다.
-- all_uids = 주기 검사(uids 무시), 아니면 uids만(타일 닫기·공유 변경). 유예 = 서버 10초 규칙(REF-board-close.md 종료 유예)
-- 보기 계정 타일은 세지 않는다(2026-10-12(2)). 같은 순간 다시 연 타일은 못 볼 수 있음 — KP4(동시성 한계)
UPDATE processes p
SET kill_forced = true,
    updated_at  = NOW()
WHERE (sqlc.arg(all_uids)::boolean OR p.uid = ANY(sqlc.arg(uids)::text[]))
  AND p.status IN ('PENDING','PROCESS')
  AND p.kill_requested_at IS NOT NULL
  AND p.kill_requested_at <= NOW() - make_interval(secs => sqlc.arg(grace_secs)::int)
  AND NOT p.kill_forced
  AND NOT EXISTS (
    SELECT 1 FROM tile_trees t
    WHERE jsonb_path_exists(t.tree, '$.tiles.* ? (@.type == "terminal" && @.uid == $uid)', jsonb_build_object('uid', p.uid))
      AND process_permission(t.user_id, p.uid) >= sqlc.arg(min_perm)::smallint
  )
RETURNING p.*;

-- name: HasPendingKills :one
-- 판정을 기다리는 종료 요청이 남았는가(실행 중 · 요청됨 · 아직 강제 아님) — 주기 검사를 쉴지 정함(REF-board-close.md)
SELECT EXISTS (
  SELECT 1 FROM processes
  WHERE status IN ('PENDING','PROCESS') AND kill_requested_at IS NOT NULL AND NOT kill_forced
);

-- name: ListExpiredProcesses :many
-- 끝난 지 오래된 process — K2(supervisor 정리, REF-process-keep.md)
SELECT uid FROM processes
WHERE status IN ('COMPLETED', 'FAILED')
  AND COALESCE(finished_at, updated_at, created_at) < sqlc.arg(before)::timestamptz;

-- name: DeleteProcesses :exec
-- process_shares는 ON DELETE CASCADE
DELETE FROM processes WHERE uid = ANY(sqlc.arg(uids)::text[]);
