-- 로그인 세션(user_sessions) — REF-account.md 세션-1(세션 표)

-- name: CreateUserSession :exec
INSERT INTO user_sessions (id, user_id, user_agent) VALUES ($1, $2, $3);

-- name: ListAllUserSessions :many
-- 시작 때 메모리로
SELECT id, user_id, created_at FROM user_sessions;

-- name: ListUserSessions :many
SELECT * FROM user_sessions WHERE user_id = $1 ORDER BY created_at DESC;

-- name: DeleteUserSession :execrows
DELETE FROM user_sessions WHERE id = $1 AND user_id = $2;

-- name: DeleteUserSessionsOfUser :many
-- keep_id = 남길 세션(없으면 '')
DELETE FROM user_sessions WHERE user_id = $1 AND id <> @keep_id::text RETURNING id;

-- name: TouchUserSessions :exec
UPDATE user_sessions SET last_seen_at = @seen::timestamptz WHERE id = ANY(@ids::text[]);

-- name: DeleteExpiredUserSessions :many
DELETE FROM user_sessions WHERE created_at < @before::timestamptz RETURNING id;
