-- 계정 설정 — REF-settings-footer.md "설정 두 겹"

-- name: ListUserSettings :many
SELECT key, value, updated_at FROM user_settings
WHERE user_id = $1
ORDER BY key;

-- name: SaveUserSetting :exec
INSERT INTO user_settings (user_id, key, value)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();
