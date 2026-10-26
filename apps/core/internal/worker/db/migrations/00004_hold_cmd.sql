-- 보관 목록에 이름·실행 명령·폴더 — 미등록 process(붙은 supervisor DB에 행이 없음) 화면 표시용(주소-4(미등록 process), REF-supervisor-move.md)

-- +goose Up
ALTER TABLE holds ADD COLUMN label TEXT NOT NULL DEFAULT '';  -- 스크립트 노드 이름(장비 파일 실행은 빈 값)
ALTER TABLE holds ADD COLUMN cmd TEXT NOT NULL DEFAULT '';  -- 명령 + 인자 한 줄(표시용)
ALTER TABLE holds ADD COLUMN cwd TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE holds DROP COLUMN cwd;
ALTER TABLE holds DROP COLUMN cmd;
ALTER TABLE holds DROP COLUMN label;
