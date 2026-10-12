-- supervisor(PostgreSQL) — process 종료 요청 기록. 끊긴 장비도 기록해 두고 재접속 때 전달 → REF-process-kill.md
-- 상태 값은 그대로(PENDING/PROCESS) — 종료 요청은 상태가 아니라 별도 표시(KP2(상태 표현))

-- +goose Up
ALTER TABLE processes ADD COLUMN kill_requested_at TIMESTAMPTZ NULL;
ALTER TABLE processes ADD COLUMN kill_forced BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE processes DROP COLUMN kill_forced;
ALTER TABLE processes DROP COLUMN kill_requested_at;
