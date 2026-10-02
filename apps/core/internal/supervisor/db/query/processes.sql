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
    cols
) VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 )
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

-- name: ListLiveProcessesByOwner :many
SELECT * FROM processes
WHERE owner_user_id = $1
  AND status IN ('PENDING','PROCESS')
ORDER BY created_at ASC;

-- 요청한 uid만, 상태 무관(끝난 것 포함) — 3-d①(끝난 process 조회). 남의 것·없는 uid는 빠짐
-- name: ListProcessesByUids :many
SELECT * FROM processes
WHERE owner_user_id = sqlc.arg(owner_user_id)
  AND uid = ANY(sqlc.arg(uids)::text[])
ORDER BY created_at ASC;
