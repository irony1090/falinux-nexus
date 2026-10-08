-- supervisor(PostgreSQL) — 공유(5 공유(P·R·T)) 테이블 + 판정 함수. 설계 → REF-process-share-db.md
-- permission = 0(차단)·10(보기)·20(조작)·30(편집), 소유자 40은 판정 함수 반환값에만 있음(SA1(단계 저장 형식))
-- 가까운 공유 판정은 node_permission/process_permission 한 곳에만 둔다(SA2(판정 위치)) — 쿼리는 호출만 할 것

-- +goose Up
CREATE TABLE node_shares (
    node_id    BIGINT      NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,  -- 노드 삭제 = 공유 자동 해제
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- 받는 계정
    permission SMALLINT    NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL,

    PRIMARY KEY (node_id, user_id),
    CHECK (permission IN (0, 10, 20, 30))
);

CREATE INDEX idx_node_shares_user ON node_shares(user_id);

CREATE TABLE process_shares (
    process_uid TEXT        NOT NULL REFERENCES processes(uid) ON DELETE CASCADE,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission  SMALLINT    NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NULL,

    PRIMARY KEY (process_uid, user_id),
    CHECK (permission IN (0, 10, 20, 30))
);

CREATE INDEX idx_process_shares_user ON process_shares(user_id);

-- SH8(받은 쪽이 만든 노드의 소유) / SH9(받은 쪽이 실행한 process의 소유). 기존 행은 소유자로 채움(SA4(기존 행 채우기))
ALTER TABLE nodes ADD COLUMN created_by BIGINT REFERENCES users(id);
UPDATE nodes SET created_by = owner_user_id;
ALTER TABLE nodes ALTER COLUMN created_by SET NOT NULL;

ALTER TABLE processes ADD COLUMN started_by BIGINT REFERENCES users(id);
UPDATE processes SET started_by = owner_user_id;
ALTER TABLE processes ALTER COLUMN started_by SET NOT NULL;

-- 가까운 공유 = 노드에서 루트 쪽으로 올라가며 처음 만나는 그 계정의 공유 행. 그 행의 node_id(출처)를 반환, 없으면 NULL
-- 가까운 공유 계산은 여기 한 곳뿐 — node_permission·공유 목록(출처 표시)이 모두 이 함수를 쓴다(SC1(출처 계산 위치)).
-- RETURNS TABLE은 sqlc 1.31.1이 컬럼을 못 읽어 출처 id만 돌려주고 단계는 (node_id, user_id)로 node_shares에서 읽는다
-- +goose StatementBegin
CREATE FUNCTION nearest_node_share(p_user BIGINT, p_node BIGINT) RETURNS BIGINT
LANGUAGE sql STABLE AS $$
    WITH RECURSIVE chain(id, parent_id, depth) AS (
        SELECT n.id, n.parent_id, 0
        FROM nodes n
        WHERE n.id = p_node
      UNION ALL
        SELECT n.id, n.parent_id, c.depth + 1
        FROM nodes n
        JOIN chain c ON n.id = c.parent_id
        WHERE c.depth < 64
    )
    SELECT c.id
    FROM chain c
    JOIN node_shares s ON s.node_id = c.id AND s.user_id = p_user
    ORDER BY c.depth
    LIMIT 1
$$;
-- +goose StatementEnd

-- 노드에 대한 계정의 단계. 소유자 40 / 그 외 가까운 공유의 단계 / 공유 없음·노드 없음 0
-- +goose StatementBegin
CREATE FUNCTION node_permission(p_user BIGINT, p_node BIGINT) RETURNS SMALLINT
LANGUAGE sql STABLE AS $$
    SELECT (CASE
        WHEN (SELECT n.owner_user_id FROM nodes n WHERE n.id = p_node) = p_user THEN 40
        ELSE COALESCE((
            SELECT s.permission
            FROM node_shares s
            WHERE s.node_id = nearest_node_share(p_user, p_node) AND s.user_id = p_user
        ), 0)
    END)::smallint
$$;
-- +goose StatementEnd

-- process 단독 공유가 노드 공유보다 가깝다 → 있으면 우선. process 없음 = NULL
-- +goose StatementBegin
CREATE FUNCTION process_permission(p_user BIGINT, p_uid TEXT) RETURNS SMALLINT
LANGUAGE sql STABLE AS $$
    SELECT (CASE
        WHEN p.owner_user_id = p_user THEN 40
        WHEN ps.permission IS NOT NULL THEN ps.permission
        WHEN p.node_id IS NOT NULL THEN node_permission(p_user, p.node_id)
        ELSE 0
    END)::smallint
    FROM processes p
    LEFT JOIN process_shares ps ON ps.process_uid = p.uid AND ps.user_id = p_user
    WHERE p.uid = p_uid
$$;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION process_permission;
DROP FUNCTION node_permission;
DROP FUNCTION nearest_node_share;
ALTER TABLE processes DROP COLUMN started_by;
ALTER TABLE nodes DROP COLUMN created_by;
DROP TABLE process_shares;
DROP TABLE node_shares;
