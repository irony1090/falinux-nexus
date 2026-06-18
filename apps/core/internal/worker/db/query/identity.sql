-- name: GetIdentity :one
SELECT * FROM identity WHERE main_key = ?;

-- name: UpsertIdentity :exec
INSERT INTO identity (main_key, sub_key, updated_at)
VALUES (?, ?, unixepoch())
ON CONFLICT(main_key) DO UPDATE SET sub_key = excluded.sub_key, updated_at = unixepoch();