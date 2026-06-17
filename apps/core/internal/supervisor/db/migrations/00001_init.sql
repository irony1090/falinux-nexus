-- supervisor(PostgreSQL) 초기 스키마.
-- 아래 agents 테이블은 시작용 예시 — 실제 필요한 컬럼으로 수정할 것.

-- +goose Up
CREATE TABLE agents (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    agent_id   TEXT NOT NULL UNIQUE,           -- worker 식별자
    name       TEXT NOT NULL DEFAULT '',
    last_seen  TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE agents;
