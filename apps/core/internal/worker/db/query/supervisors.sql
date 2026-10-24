-- name: ListSupervisors :many
SELECT url FROM supervisors ORDER BY rank;

-- name: DeleteSupervisors :exec
DELETE FROM supervisors;

-- name: InsertSupervisor :exec
INSERT INTO supervisors (url, rank) VALUES (?, ?);

-- name: GetSetting :one
SELECT value FROM settings WHERE key = ?;

-- name: SetSetting :exec
INSERT INTO settings (key, value) VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value;
