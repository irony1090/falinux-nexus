-- supervisor(PostgreSQL) — process_subscribers 테이블.
-- 세션(브라우저)→구독 process uid 원장. 인가만 DB 담당, 실시간 라우팅은 subscribe.Hub(메모리).
-- sid=gorilla CookieStore 세션 쿠키 원본값(로그인마다 갱신, 같은 브라우저 내에선 고정). 해시 아님(단순 구현 우선, 2026-07-15 확정).
-- 상세 설계 → REF-process-reconnect.md "세션→uid 원장 구체화".

-- +goose Up
CREATE TABLE process_subscribers (
    process_uid   TEXT        NOT NULL REFERENCES processes(uid) ON DELETE CASCADE,
    owner_user_id BIGINT      NOT NULL REFERENCES users(id),
    sid           TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (process_uid, sid)
);

CREATE INDEX idx_process_subscribers_sid ON process_subscribers(sid);

-- +goose Down
DROP TABLE process_subscribers;
