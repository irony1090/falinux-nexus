-- 계정 권한 = 기능별 이름 목록 — 주소-3(권한 열), REF-supervisor-move.md. 주는 법 = DB 직접 UPDATE(주소-Q23)
-- worker_move = 전체 worker 주소 변경. identification은 sha256 hex라 닉네임이나 id로 고른다
--   UPDATE users SET perms = array_append(perms, 'worker_move') WHERE nickname = 'irony' AND NOT 'worker_move' = ANY(perms);

-- +goose Up
ALTER TABLE users ADD COLUMN perms TEXT[] NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE users DROP COLUMN perms;
