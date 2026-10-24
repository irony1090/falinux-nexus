-- supervisor 주소 목록 — 순위(rank 오름차순)대로 접속 시도, 붙은 주소가 1순위. 최대 3개(주소-1(주소 목록), REF-supervisor-move.md)

-- +goose Up
CREATE TABLE supervisors (
    url  TEXT    PRIMARY KEY,  -- scheme://host (경로 /worker는 worker가 붙임)
    rank INTEGER NOT NULL
);

-- 작은 설정 값(env_seed = 마지막으로 본 env 주소 — env를 고쳤을 때만 그 주소를 1순위로)
CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- +goose Down
DROP TABLE settings;
DROP TABLE supervisors;
