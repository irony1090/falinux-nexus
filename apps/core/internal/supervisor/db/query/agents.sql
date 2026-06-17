-- name: GetAgentByAgentID :one
SELECT * FROM agents WHERE agent_id = $1;

-- name: ListAgents :many
SELECT * FROM agents ORDER BY created_at DESC;

-- name: UpsertAgent :one
INSERT INTO agents (agent_id, name, last_seen)
VALUES ($1, $2, now())
ON CONFLICT (agent_id) DO UPDATE SET name = excluded.name, last_seen = now()
RETURNING *;
