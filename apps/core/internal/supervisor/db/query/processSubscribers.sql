-- name: CreateProcessSubscriber :exec
INSERT INTO process_subscribers (process_uid, owner_user_id, sid)
VALUES ($1, $2, $3)
ON CONFLICT (process_uid, sid) DO NOTHING;

-- name: DeleteProcessSubscriber :exec
DELETE FROM process_subscribers WHERE process_uid = $1 AND sid = $2;

-- name: ListProcessesBySid :many
SELECT p.* FROM processes p
JOIN process_subscribers ps ON ps.process_uid = p.uid
WHERE ps.sid = $1
ORDER BY ps.created_at DESC;
