-- supervisor(PostgreSQL) — processes 테이블.
-- in-memory 레지스트리의 write-through 영속층(재시작 복구·frontend 목록).
-- process 1개 = 행 1개(가변 레코드). 상태전이는 UPDATE로 덮어씀.
-- 출력 스크롤백 영속(DB sink)은 별 테이블·후순위(현재 범위 아님).

-- +goose Up
CREATE TABLE processes (
    uid           TEXT        PRIMARY KEY,          -- 와이어 라우팅 키(RandomKey), 그대로 PK
    type          VARCHAR(16) NOT NULL,             -- 'EXEC' | 'EDIT'
    owner_user_id BIGINT      NOT NULL REFERENCES users(id),
    node_id       BIGINT      NULL,                 -- 스크립트 기원(있으면). FK 없음(nodes 미생성)
    device_key    TEXT        NOT NULL,             -- worker InstanceKey(메인키#서브키), FK 아님

    cmd           TEXT        NOT NULL,
    args          TEXT[]      NOT NULL DEFAULT '{}',
    env           TEXT[]      NOT NULL DEFAULT '{}',
    cwd           TEXT        NOT NULL DEFAULT '',
    rows          SMALLINT    NOT NULL DEFAULT 0,   -- 마지막 알려진 PTY 크기
    cols          SMALLINT    NOT NULL DEFAULT 0,

    status        VARCHAR(16) NOT NULL DEFAULT 'PENDING',  -- PENDING|PROCESS|COMPLETED|FAILED
    pid           INT         NULL,                 -- PROCESS 시점 worker 로컬 PID
    exit_code     INT         NULL,                 -- 종료 시점만 유효

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),  -- exec 요청 시각
    started_at    TIMESTAMPTZ NULL,                    -- RUNNING(PROCESS) 수신
    finished_at   TIMESTAMPTZ NULL,                    -- 종료(COMPLETED|FAILED)
    updated_at    TIMESTAMPTZ NULL,

    CHECK (type   IN ('EXEC','EDIT')),
    CHECK (status IN ('PENDING','PROCESS','COMPLETED','FAILED'))
);

CREATE INDEX idx_processes_owner  ON processes(owner_user_id);
CREATE INDEX idx_processes_device ON processes(device_key);
-- 재시작 복구·미종료 조회용 부분 인덱스(종료건 제외)
CREATE INDEX idx_processes_active ON processes(device_key)
    WHERE status IN ('PENDING','PROCESS');

-- +goose Down
DROP TABLE processes;
