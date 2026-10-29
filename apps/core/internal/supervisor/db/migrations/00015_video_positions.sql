-- supervisor(PostgreSQL) — 영상 이어보기 S1(이어보기): 계정·파일마다 마지막 재생 위치 → REF-video-view.md
-- 키 = 파일(instance+path), 타일 아님 — 영상 보기 타일은 같은 파일이 계정 전체 하나. 타일을 닫으면 지움(P1(위치 지우는 시점))

-- +goose Up
CREATE TABLE video_positions (
    user_id    BIGINT           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    instance   TEXT             NOT NULL,
    path       TEXT             NOT NULL,
    pos        DOUBLE PRECISION NOT NULL,
    updated_at TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, instance, path)
);

-- +goose Down
DROP TABLE video_positions;
