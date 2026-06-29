-- supervisor(PostgreSQL) — nodes 테이블 (frontend 카탈로그 트리).
-- 폴더(컨테이너)와 스크립트(리프)를 한 트리에 섞는 inode 패턴(kind로 구분).
-- 트리=parent_id 자기참조(루트 NULL). 소유=owner_user_id. 상세 설계 → REF-node-label.md.
--
-- 좌표 분리: position_x/y=캔버스 절대배치 전용 / ord=그리드·리스트 순서 전용(서로 독립).
-- device_key=폴더가 묶인 worker 장비 클래스(main_key만, FK 아님). 실행할 subkey는 Exec 시점 선택.

-- +goose Up
CREATE TABLE nodes (
    id            BIGSERIAL    PRIMARY KEY,
    owner_user_id BIGINT       NOT NULL REFERENCES users(id),
    parent_id     BIGINT       NULL REFERENCES nodes(id) ON DELETE CASCADE,  -- 폴더 삭제 시 하위 subtree 동반 삭제(rm -r)
    kind          VARCHAR(16)  NOT NULL,                -- 'FOLDER' | 'SCRIPT'
    name          TEXT         NOT NULL,

    ord           BIGINT       NOT NULL,                -- 그리드/리스트 순서(부모별 로컬, 정수. 인접 삽입은 재번호)
    position_x    NUMERIC(5,2) NULL,                    -- 캔버스 절대배치 전용, 0~100 부모별 로컬(미배치=NULL)
    position_y    NUMERIC(5,2) NULL,

    device_key    TEXT         NULL,                    -- folder 전용: worker main_key 문자열(FK 아님). NULL=순수 정리용 폴더
    content       TEXT         NULL,                    -- script 전용: 셸 본문 텍스트

    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NULL,

    CHECK (kind IN ('FOLDER','SCRIPT')),
    CHECK (kind = 'SCRIPT' OR content    IS NULL),      -- content는 SCRIPT만
    CHECK (kind = 'FOLDER' OR device_key IS NULL),      -- device_key는 FOLDER만
    CHECK (position_x IS NULL OR position_x BETWEEN 0 AND 100),
    CHECK (position_y IS NULL OR position_y BETWEEN 0 AND 100)
    -- ord엔 상한 CHECK 없음(좌표 아닌 순서키)
);

-- 그리드/리스트 조회: WHERE parent_id=$1 ORDER BY ord (NULL parent=루트도 커버)
CREATE INDEX idx_nodes_children ON nodes(parent_id, ord);
CREATE INDEX idx_nodes_owner    ON nodes(owner_user_id);
-- worker(main_key)→귀속 폴더 역조회(folder 행만)
CREATE INDEX idx_nodes_device   ON nodes(device_key) WHERE device_key IS NOT NULL;

-- +goose Down
DROP TABLE nodes;
