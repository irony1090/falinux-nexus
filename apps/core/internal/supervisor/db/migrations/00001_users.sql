-- supervisor(PostgreSQL) — users 테이블 (구 PortBridge 이식).

-- +goose Up
CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    identification  VARCHAR(256) NOT NULL UNIQUE,
    password        VARCHAR(256) NOT NULL,
    nickname        VARCHAR(64) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NULL
);

-- identification은 UNIQUE 제약이 이미 인덱스를 만든다(아래 별도 인덱스는 중복 — 필요 없으면 제거).
CREATE INDEX idx_users_identification ON users(identification);
CREATE INDEX idx_users_nickname ON users(nickname);

-- +goose Down
DROP TABLE users;
