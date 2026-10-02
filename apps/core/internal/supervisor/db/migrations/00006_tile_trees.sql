-- supervisor(PostgreSQL) — 계정당 타일 트리 문서 1개(⑪(타일 트리 서버 저장)) → REF-node-ui-save.md
-- tree = 프론트 tileTree.store의 {rootId, tiles} 그대로(서버는 모양 검사만) / version = 저장 충돌 판정용

-- +goose Up
CREATE TABLE tile_trees (
    user_id    BIGINT      PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    tree       JSONB       NOT NULL,
    version    BIGINT      NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE tile_trees;
