-- supervisor(PostgreSQL) — worker 인스턴스 OS(등록 때 보고, 빈 값 = 옛 worker = linux). 표시용 → REF-windows-worker.md WW5(OS 표시)

-- +goose Up
ALTER TABLE worker_instances ADD COLUMN os TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE worker_instances DROP COLUMN os;
