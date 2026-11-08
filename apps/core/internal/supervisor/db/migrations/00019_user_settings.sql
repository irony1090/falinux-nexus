-- supervisor(PostgreSQL) — 계정 설정(설정 두 겹 중 계정 겹): 키 = 프론트 SPECS 항목 이름, 값 모양 검사는 프론트 → REF-settings-footer.md "설정 두 겹"

-- +goose Up
CREATE TABLE user_settings (
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key        TEXT        NOT NULL,
    value      JSONB       NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, key)
);

-- +goose Down
DROP TABLE user_settings;
