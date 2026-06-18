-- worker(SQLite) 초기 스키마. 자기 자신의 정체성(메인키/서브키)을 영속한다.
-- process는 메모리 관리, 여기엔 재부팅 후 재접속에 필요한 식별자만 저장.

-- +goose Up
CREATE TABLE identity (
    main_key   TEXT PRIMARY KEY,
    sub_key    TEXT NOT NULL,
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    UNIQUE(main_key, sub_key)
);

-- +goose Down
DROP TABLE identity;
