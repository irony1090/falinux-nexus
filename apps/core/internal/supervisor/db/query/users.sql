-- name: CreateUser :one
INSERT INTO users (
    identification,
    password,
    nickname,
    perms
) VALUES ( $1, $2, $3, @perms::text[] )
RETURNING *;

-- name: LockUserCreate :exec
-- 가입 직렬화(트랜잭션 끝에 풀림) — 첫 회원 = 관리자 판정이 동시 가입에 두 번 걸리지 않게(회원-Q5(관리자 판정))
SELECT pg_advisory_xact_lock(hashtext('users_create'));

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: UpdateUserPassword :exec
-- 옛 sha256 -> bcrypt 다시 저장(회원-Q9(bcrypt 전환))
UPDATE users SET password = $2 WHERE id = $1;

-- name: ListUsersWithDevices :many
-- 관리자 페이지 — 회원과 등록한 장비 이름 수
SELECT u.id, u.nickname, u.perms, u.worker_limit, u.created_at, u.suspended_at,
       (SELECT COUNT(*) FROM devices d WHERE d.owner_user_id = u.id)::int AS device_count
FROM users u
ORDER BY u.id;

-- name: UpdateUserMember :one
-- 관리자 페이지 — 상한·권한
UPDATE users SET worker_limit = $2, perms = @perms::text[], updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByIdentification :one
SELECT * FROM users WHERE identification = $1;

-- name: ListUserNicknames :many
SELECT id, nickname FROM users WHERE id = ANY(@ids::bigint[]);

-- name: ListSuspendedUserIDs :many
SELECT id FROM users WHERE suspended_at IS NOT NULL;

-- name: SetUserSuspended :one
UPDATE users SET suspended_at = CASE WHEN @suspend::bool THEN NOW() ELSE NULL END, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountActiveProcessesOfUser :one
-- 회원 삭제 전 — 소유했거나 실행한 process 중 살아 있는 것
SELECT COUNT(*) FROM processes
WHERE (owner_user_id = $1 OR started_by = $1) AND status IN ('PENDING', 'PROCESS');

-- name: DeleteProcessesOfUser :exec
DELETE FROM processes WHERE owner_user_id = $1 OR started_by = $1;

-- name: ReassignNodeCreatedBy :exec
-- 남의 폴더에 이 회원이 만든 노드는 남기고 만든 사람을 노드 주인으로(created_by NOT NULL)
UPDATE nodes SET created_by = owner_user_id WHERE created_by = $1 AND owner_user_id <> $1;

-- name: DeleteNodesOfUser :exec
DELETE FROM nodes WHERE owner_user_id = $1;

-- name: DeleteUser :execrows
-- devices·공유·타일 트리는 CASCADE
DELETE FROM users WHERE id = $1;
