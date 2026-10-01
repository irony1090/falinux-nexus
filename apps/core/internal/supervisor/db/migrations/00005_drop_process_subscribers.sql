-- supervisor(PostgreSQL) — process_subscribers 제거.
-- 구독 역할 분리(S): 인가 = 계정(processes.owner_user_id) / 라우팅 = 소켓(Hub) / 크기 우선권 = 탭 id(메모리) → REF-process-sync.md

-- +goose Up
DROP TABLE process_subscribers;

-- +goose Down
CREATE TABLE process_subscribers (
    process_uid   TEXT        NOT NULL REFERENCES processes(uid) ON DELETE CASCADE,
    owner_user_id BIGINT      NOT NULL REFERENCES users(id),
    sid           TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (process_uid, sid)
);

CREATE INDEX idx_process_subscribers_sid ON process_subscribers(sid);
