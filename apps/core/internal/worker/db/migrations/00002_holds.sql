-- 보관 프로세스 목록 — 살아 있는 process만(끝나면 바로 삭제). worker가 다시 켜지면 sock으로 다시 붙는다(REF-process-keep.md K1(worker 목록))

-- +goose Up
CREATE TABLE holds (
    uid        TEXT    PRIMARY KEY,
    kind       TEXT    NOT NULL,             -- EXEC | EDIT
    edit_path  TEXT    NOT NULL DEFAULT '',  -- EDIT 종료 때 읽어 보낼 파일
    sock       TEXT    NOT NULL,
    holder_pid INTEGER NOT NULL,             -- 소켓이 망가졌을 때 정리용(cmdline으로 확인 후)
    pid        INTEGER NOT NULL,             -- 스크립트 pid
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- +goose Down
DROP TABLE holds;
