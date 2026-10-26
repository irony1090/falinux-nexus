-- devices = 장비 이름(메인키) 소유·API 키 — REF-member.md 회원-1(DB)

-- name: CreateDevice :one
INSERT INTO devices (main_key, owner_user_id, api_key_hash, api_key_enc)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetDevice :one
SELECT * FROM devices WHERE main_key = $1;

-- name: CountDevicesByOwner :one
SELECT COUNT(*) FROM devices WHERE owner_user_id = $1;

-- name: ListDevicesByOwner :many
SELECT * FROM devices WHERE owner_user_id = $1 ORDER BY main_key;

-- name: ListAllDevices :many
-- 관리자 페이지 — 주인 닉네임 포함
SELECT d.*, u.nickname AS owner_nickname
FROM devices d JOIN users u ON u.id = d.owner_user_id
ORDER BY d.main_key;

-- name: UpdateDeviceKey :one
UPDATE devices SET api_key_hash = $2, api_key_enc = $3, key_issued_at = NOW()
WHERE main_key = $1
RETURNING *;

-- name: UpdateDeviceOwner :one
-- 주인 바꾸기(회원-Q18)
UPDATE devices SET owner_user_id = $2
WHERE main_key = $1
RETURNING *;

-- name: DeleteDevice :execrows
DELETE FROM devices WHERE main_key = $1;

-- name: GetDeviceAuth :one
-- worker 등록 검사 — 키 해시 + 주인 정지 여부(회원-4)
SELECT d.api_key_hash, (u.suspended_at IS NOT NULL)::bool AS owner_suspended
FROM devices d JOIN users u ON u.id = d.owner_user_id
WHERE d.main_key = $1;
