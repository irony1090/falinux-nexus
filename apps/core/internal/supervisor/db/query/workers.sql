-- worker_instances 명부 쿼리. 설계 → REF-node-label.md "roster 계획"

-- name: UpsertWorkerInstance :exec
-- 등록 때 — 처음이면 추가, 있으면 last_seen만 갱신(first_seen 유지)
INSERT INTO worker_instances (main_key, sub_key)
VALUES ($1, $2)
ON CONFLICT (main_key, sub_key) DO UPDATE SET last_seen = NOW();

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
