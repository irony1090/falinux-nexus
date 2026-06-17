-- worker(SQLite) 초기 스키마. 설정/메타데이터만 영속한다(process는 메모리 관리).
-- 아래 settings 테이블은 시작용 예시 — 실제 필요한 컬럼으로 수정할 것.

-- +goose Up
CREATE TABLE settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- +goose Down
DROP TABLE settings;
