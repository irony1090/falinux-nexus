-- supervisor(PostgreSQL) — 장비 디렉토리에서 실행한 process(폴더-4(실행)) → REF-worker-dir-exec.md
-- 장비 실행 = node_id NULL + dir_node_id = 장비 타일을 연 폴더(Q15(장비 디렉토리 id)). 폴더 터미널 탭 목록 근거(Q10(루트가 아닐 때))

-- +goose Up
ALTER TABLE processes ADD COLUMN dir_node_id BIGINT NULL REFERENCES nodes(id) ON DELETE SET NULL;
CREATE INDEX idx_processes_dir_node ON processes(dir_node_id) WHERE dir_node_id IS NOT NULL;

-- +goose Down
DROP INDEX idx_processes_dir_node;
ALTER TABLE processes DROP COLUMN dir_node_id;
