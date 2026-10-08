-- supervisor(PostgreSQL) — 작업 공간 여러 개: tile_trees를 계정당 1행 → 여러 행으로. 설계 → REF-workspace-impl.md
-- hash = URL `/w/{hash}`·API 주소(WS1(트리 주소)), 소문자 hex 12자. 새 행의 hash는 Go(newTreeHash)가 만든다

-- +goose Up
ALTER TABLE tile_trees DROP CONSTRAINT tile_trees_pkey;
ALTER TABLE tile_trees ADD COLUMN id BIGSERIAL PRIMARY KEY;
ALTER TABLE tile_trees ADD COLUMN hash TEXT;
ALTER TABLE tile_trees ADD COLUMN name TEXT NOT NULL DEFAULT '';
ALTER TABLE tile_trees ADD COLUMN ord INT NOT NULL DEFAULT 0;
ALTER TABLE tile_trees ADD COLUMN is_default BOOLEAN NOT NULL DEFAULT false;

-- 기존 행 = 각 계정의 기본 작업 공간
UPDATE tile_trees
SET hash = substr(replace(gen_random_uuid()::text, '-', ''), 1, 12), is_default = true;

ALTER TABLE tile_trees ALTER COLUMN hash SET NOT NULL;
ALTER TABLE tile_trees ADD CONSTRAINT uq_tile_trees_hash UNIQUE (hash);

-- 계정당 기본 하나. CreateTileTree의 ON CONFLICT가 이 인덱스를 쓴다
CREATE UNIQUE INDEX uq_tile_trees_default ON tile_trees(user_id) WHERE is_default;
CREATE INDEX idx_tile_trees_user ON tile_trees(user_id);

-- +goose Down
DELETE FROM tile_trees WHERE NOT is_default;
DROP INDEX idx_tile_trees_user;
DROP INDEX uq_tile_trees_default;
ALTER TABLE tile_trees DROP COLUMN is_default;
ALTER TABLE tile_trees DROP COLUMN ord;
ALTER TABLE tile_trees DROP COLUMN name;
ALTER TABLE tile_trees DROP COLUMN hash;
ALTER TABLE tile_trees DROP COLUMN id;
ALTER TABLE tile_trees ADD PRIMARY KEY (user_id);
