-- supervisor(PostgreSQL) — 업데이트 app P2(updater Linux): 장비 루트 API 키 · 파일 권한 → REF-update-p2.md "P2 전 확정"

-- +goose Up
-- 정-1(접속 인증): 메인키별 키. hash = 접속 검증(sha256 hex) / enc = 다시 보기용 암호문(AES-GCM, env KEY_SECRET) — worker 장비 키(00014)와 같은 방식
ALTER TABLE update_devices
    ADD COLUMN api_key_hash  TEXT        NOT NULL DEFAULT '',
    ADD COLUMN api_key_enc   BYTEA       NULL,
    ADD COLUMN key_issued_at TIMESTAMPTZ NULL;

-- 정-2(실행 권한): 파일 권한 비트(0o644 = 420). .sh 기본 0o755(493) — updater가 copy 때 적용
ALTER TABLE update_files ADD COLUMN mode INT NOT NULL DEFAULT 420;
UPDATE update_files SET mode = 493 WHERE lower(path) LIKE '%.sh';

-- +goose Down
ALTER TABLE update_files DROP COLUMN mode;
ALTER TABLE update_devices DROP COLUMN key_issued_at, DROP COLUMN api_key_enc, DROP COLUMN api_key_hash;
