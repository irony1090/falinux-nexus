-- supervisor(PostgreSQL) — 로그인 세션(로그인 1번 = 1행, id = 쿠키의 _nonce). 목록에 없는 쿠키는 무효 → REF-account.md 세션-1(세션 표)

-- +goose Up
CREATE TABLE user_sessions (
    id           TEXT        PRIMARY KEY,
    user_id      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_agent   TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_user_sessions_user ON user_sessions(user_id);

-- +goose Down
DROP TABLE user_sessions;
