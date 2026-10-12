-- supervisor(PostgreSQL) — worker 인스턴스 명부(roster). 관측용, 라우팅·인가에 안 씀 → REF-node-label.md "roster 계획"
-- online은 저장하지 않는다(레지스트리로 계산). ?mainKey= 조회는 PK 앞부분이 맡는다

-- +goose Up
CREATE TABLE worker_instances (
    main_key   TEXT        NOT NULL,
    sub_key    TEXT        NOT NULL,
    first_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (main_key, sub_key)
);

-- +goose Down
DROP TABLE worker_instances;
