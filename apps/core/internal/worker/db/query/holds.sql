-- name: InsertHold :exec
INSERT INTO holds (uid, kind, edit_path, sock, holder_pid, pid)
VALUES (?, ?, ?, ?, ?, ?);

-- name: DeleteHold :exec
DELETE FROM holds WHERE uid = ?;

-- name: ListHolds :many
SELECT * FROM holds ORDER BY created_at;
