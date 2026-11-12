-- supervisor(PostgreSQL) — 로그인 아이디 평문 저장(sha256 hex였음) + 아이디·닉네임 대소문자 무시 중복 금지 → REF-account.md 아이디-1(평문 이전)

-- +goose Up
-- 해시가 맞는 첫 후보로 평문 지정(닉네임 -> 알려진 아이디). 아무것도 안 맞으면 '' = 아이디 없음(로그인 불가, 관리자가 지정)
UPDATE users u SET identification = COALESCE((
    SELECT c.cand FROM (VALUES (1, u.nickname), (2, 'irony'), (3, 'test1'), (4, 'test2'), (5, 'test3')) AS c(ord, cand)
    WHERE u.identification = encode(sha256(convert_to(c.cand, 'UTF8')), 'hex')
    ORDER BY c.ord LIMIT 1
), '');

ALTER TABLE users DROP CONSTRAINT users_identification_key;
DROP INDEX idx_users_identification;
DROP INDEX idx_users_nickname;
CREATE UNIQUE INDEX uq_users_identification ON users (lower(identification)) WHERE identification <> '';
CREATE UNIQUE INDEX uq_users_nickname ON users (lower(nickname));

-- +goose Down
DROP INDEX uq_users_nickname;
DROP INDEX uq_users_identification;
UPDATE users SET identification = CASE WHEN identification = '' THEN 'unset-' || id
    ELSE encode(sha256(convert_to(identification, 'UTF8')), 'hex') END;
ALTER TABLE users ADD CONSTRAINT users_identification_key UNIQUE (identification);
CREATE INDEX idx_users_identification ON users(identification);
CREATE INDEX idx_users_nickname ON users(nickname);
