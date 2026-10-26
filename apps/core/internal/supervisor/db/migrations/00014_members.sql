-- supervisor(PostgreSQL) — 회원 관리: 장비 이름(메인키) 소유·API 키·등록 상한·인스턴스 라벨 → REF-member.md 회원-1(DB)
-- 권한 이름(users.perms): admin = 관리자 페이지 / worker_register = 장비 이름 등록(회원-Q3(이름 등록 가능 여부))
-- 회원이 이미 있는 DB는 첫 회원 = 관리자 규칙이 안 걸린다 → 배포 전 사용자에게 관리자로 지정할 회원을 묻고(REF-member.md "배포 전 확인"):
--   UPDATE users SET perms = array_append(perms, 'admin') WHERE nickname = '<닉네임>' AND NOT 'admin' = ANY(perms);

-- +goose Up
ALTER TABLE users ADD COLUMN worker_limit INT NOT NULL DEFAULT 3;
-- 정지(회원-Q21) — 로그인·세션·그 회원 장비의 worker 접속을 막고 데이터는 남긴다. NULL = 정상
ALTER TABLE users ADD COLUMN suspended_at TIMESTAMPTZ NULL;

-- 기존 회원도 새 가입자와 같게 등록 권한을 켠다(회원-Q6(새 가입자 기본값))
UPDATE users SET perms = array_append(perms, 'worker_register') WHERE NOT 'worker_register' = ANY(perms);

-- api_key_hash = 접속 검증(sha256 hex) / api_key_enc = 다시 보기용 암호문(AES-GCM, env KEY_SECRET) — 회원-Q16(키 다시 보기)
CREATE TABLE devices (
    main_key      TEXT        PRIMARY KEY,
    owner_user_id BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_hash  TEXT        NOT NULL,
    api_key_enc   BYTEA       NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    key_issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_devices_owner ON devices(owner_user_id);

-- 인스턴스(메인키#서브키) 표시 이름 — 회원-Q11(인스턴스 구분). 빈 값 = 라벨 없음
ALTER TABLE worker_instances ADD COLUMN label TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE worker_instances DROP COLUMN label;
DROP TABLE devices;
UPDATE users SET perms = array_remove(perms, 'worker_register');
ALTER TABLE users DROP COLUMN suspended_at;
ALTER TABLE users DROP COLUMN worker_limit;
