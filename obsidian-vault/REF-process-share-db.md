# REF — process 공유: 공유-a(DB) 구조안

> 2026-10-06 신설. 공유 결정·작업 순서 → `REF-process-share.md` / SH3(받은 계정 권한) 단계 정의 → `REF-process-share-node.md` / 이력 → `history/process-share-db.md` / 현재 진행 → `CURRENT.md`.

**상태(2026-10-07): 코드 작성 완료(커밋 전). 2026-10-07 검토 뒤 `ListChildrenVisible` 보완 + 공유-b에서 목록 쿼리 `sqlc.embed`화(→ `REF-process-share-authz.md` "구현 메모"). 2026-10-06: SA1(단계 저장 형식)~SA4(기존 행 채우기) 제안대로 확정. build/vet 통과 + 임시 DB에서 판정·목록 쿼리·Down 확인. 개발 DB `nexus`에는 supervisor 재기동 때 적용됨(작성 시점에 떠 있던 supervisor는 옛 바이너리).**

## 범위
| 넣음(공유-a) | 넣지 않음 |
|---|---|
| 마이그레이션 `00007_shares.sql`(표 2개 + 컬럼 2개 + 판정 함수 2개) | 공유 걸기·풀기·목록 쿼리 → 공유-c(공유 API) |
| `query/shares.sql` 판정 쿼리 | 기존 소유자 검사 호출부 교체 → 공유-b(인가 교체) |
| `CreateFolder`·`CreateScript`·`CreateProcess`에 `created_by`·`started_by` 인자 추가 + 호출부 컴파일 맞춤(값 = 지금은 요청한 계정) | 받는 쪽 `/` 루트 공유 목록(`ListShareRoots`) → 공유-d(실시간) |

## 결정 (SA = 공유-a, 2026-10-06 전부 제안대로 확정)
| 결정 | 묻는 것 | 제안 |
|---|---|---|
| SA1(단계 저장 형식) | 차단·보기·조작·편집을 어떻게 저장하나 | **SMALLINT, 간격을 둔 값** 0(차단)·10(보기)·20(조작)·30(편집), 표에는 CHECK `IN (0,10,20,30)`. **소유자 = 40**은 판정 함수의 반환값에만 존재. 중간 단계는 15 같은 값으로 끼워 넣음. 기각: PG ENUM(`ALTER TYPE ADD VALUE`는 같은 트랜잭션 안에서 새 값을 쓸 수 없음, sqlc가 문자열 타입으로 생성해 크기 비교가 Go에서 불편) |
| SA2(판정 위치) | "가까운 공유" 계산을 어디에 두나 | **DB 함수 `node_permission(user, node)` · `process_permission(user, uid)` 두 개**(`LANGUAGE sql STABLE`). 쿼리는 이 함수를 호출만 함 → SH5의 "가까운 공유 판정은 서버 한 곳에만"을 문자 그대로 만족. sqlc 1.31.1이 반환형을 `int16`으로 생성함(2026-10-06 스크래치 확인). 기각: 쿼리마다 `WITH RECURSIVE` 복사(규칙이 바뀌면 여러 곳을 고쳐야 함) / Go에서 계산(목록마다 N+1 조회) |
| SA3(숨기기 시점) | 숨기기 컬럼을 지금 넣나 | **나중(공유-e(프론트) 때)**. 숨기기는 화면·구독에서만 빼고 인가에는 영향이 없어서 판정 함수와 무관. 넣을 때는 `node_shares.hidden_at TIMESTAMPTZ NULL` 컬럼 하나 |
| SA4(기존 행 채우기) | `created_by`·`started_by`를 이미 있는 행에 무엇으로 채우나 | **소유자로 채운 뒤 NOT NULL**. 공유 전에는 만든 사람 = 소유자이므로 사실과 같음 |

## 마이그레이션 `00007_shares.sql`
```sql
-- +goose Up
CREATE TABLE node_shares (
    node_id    BIGINT      NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,  -- 노드 삭제 = 공유 자동 해제
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- 받는 계정
    permission SMALLINT    NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL,
    PRIMARY KEY (node_id, user_id),               -- 같은 노드·계정 = 1행(SH4 upsert)
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

ALTER TABLE nodes ADD COLUMN created_by BIGINT REFERENCES users(id);         -- SH8
UPDATE nodes SET created_by = owner_user_id;
ALTER TABLE nodes ALTER COLUMN created_by SET NOT NULL;

ALTER TABLE processes ADD COLUMN started_by BIGINT REFERENCES users(id);     -- SH9
UPDATE processes SET started_by = owner_user_id;
ALTER TABLE processes ALTER COLUMN started_by SET NOT NULL;

-- +goose StatementBegin
CREATE FUNCTION node_permission(p_user BIGINT, p_node BIGINT) RETURNS SMALLINT
LANGUAGE sql STABLE AS $$
  WITH RECURSIVE chain(id, parent_id, owner_user_id, depth) AS (
    SELECT id, parent_id, owner_user_id, 0 FROM nodes WHERE id = p_node
    UNION ALL
    SELECT n.id, n.parent_id, n.owner_user_id, c.depth + 1
    FROM nodes n JOIN chain c ON n.id = c.parent_id
    WHERE c.depth < 64
  )
  SELECT (CASE
    WHEN (SELECT owner_user_id FROM chain WHERE depth = 0) = p_user THEN 40
    ELSE COALESCE((SELECT s.permission FROM chain c
                   JOIN node_shares s ON s.node_id = c.id AND s.user_id = p_user
                   ORDER BY c.depth LIMIT 1), 0)
  END)::smallint
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION process_permission(p_user BIGINT, p_uid TEXT) RETURNS SMALLINT
LANGUAGE sql STABLE AS $$
  SELECT (CASE
    WHEN p.owner_user_id = p_user THEN 40
    WHEN ps.permission IS NOT NULL THEN ps.permission          -- process 단독 공유가 노드보다 가까움
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
ALTER TABLE processes DROP COLUMN started_by;
ALTER TABLE nodes DROP COLUMN created_by;
DROP TABLE process_shares;
DROP TABLE node_shares;
```

### 판정 함수 반환값
| 값 | 뜻 | 쓰는 곳(공유-b(인가 교체)에서) |
|---|---|---|
| 0 | 접근 없음(공유 없음 또는 가장 가까운 공유가 차단) | 404처럼 처리(존재 노출 안 함) |
| 10 | 보기 | 목록·내용·출력·스냅샷 |
| 20 | 조작 | + 실행·입력·kill·크기 우선권 후보 |
| 30 | 편집 | + 노드 추가·수정·삭제·이동 |
| 40 | 소유자 | + 공유 걸기·풀기(P(공유 권한)) |

- 비교는 항상 `>=`(예: 입력 허용 = `>= 20`). Go 쪽에 `PermBlock/PermView/PermOperate/PermEdit/PermOwner` 상수.
- 노드가 없으면 0(소유자 조회가 NULL → ELSE 쪽 → 공유 행 없음).
- 받는 계정이 소유자 자신인 공유 행은 표 제약으로 못 막음 → 공유-c(공유 API) 핸들러에서 거절(SH4 "자기 자신 거절").

## 판정 쿼리 `query/shares.sql`
| 쿼리 | 반환 | 쓰는 곳(공유-b(인가 교체)) |
|---|---|---|
| `NodePermission(user, node)` | int16 | 노드 단건 조회·수정·실행 검사 |
| `ProcessPermission(user, uid)` | int16 | kill·resize·snapshot·입력 검사(`requireOwnedEntry`, `input.go` 대체) |
| `ListChildrenVisible(user, parent, parent_perm)` | 자식 행 + `permission` | 폴더 목록. 자식 단계 = `COALESCE(자식 자신의 공유 행, parent_perm)`, 0이면 제외. 부모 단계는 호출자가 `NodePermission`으로 먼저 구함 → 재귀 없이 한 단계만 조인. 소유자도 같은 쿼리(parent_perm=40) |
| `ListVisibleLiveProcesses(user)` | process 행 + `permission` | 소켓 연결 때 구독(`subscribeLiveProcesses`) — 소유 + 공유분 |
| `ListVisibleProcessesByUids(user, uids)` | process 행 + `permission` | 끝난 process 조회(`ListProcessesByUids` 대체) |
| `ListProcessViewers(uid)` | `user_id`, `permission`(>0만) | exec·재접속 때 볼 수 있는 계정 전부 구독(`subscribeUser`), 크기 우선권 다음 후보(`handoffSizeOwner`) |

```sql
-- name: ListChildrenVisible :many
SELECT n.*, COALESCE(s.permission, sqlc.arg(parent_perm)::smallint) AS permission
FROM nodes n
LEFT JOIN node_shares s ON s.node_id = n.id AND s.user_id = sqlc.arg(user_id)
WHERE n.parent_id = sqlc.arg(parent_id)
  AND COALESCE(s.permission, sqlc.arg(parent_perm)::smallint) > 0
ORDER BY n.ord, n.name, n.id;

-- name: ListVisibleLiveProcesses :many
SELECT p.*, process_permission(sqlc.arg(user_id), p.uid) AS permission
FROM processes p
WHERE p.status IN ('PENDING','PROCESS')
  AND process_permission(sqlc.arg(user_id), p.uid) > 0
ORDER BY p.created_at ASC;

-- name: ListProcessViewers :many
WITH RECURSIVE chain(id, parent_id, depth) AS (
    SELECT n.id, n.parent_id, 0 FROM nodes n
    JOIN processes p ON p.node_id = n.id WHERE p.uid = sqlc.arg(uid)
  UNION ALL
    SELECT n.id, n.parent_id, c.depth + 1 FROM nodes n
    JOIN chain c ON n.id = c.parent_id WHERE c.depth < 64
), cand(user_id) AS (
    SELECT owner_user_id FROM processes WHERE uid = sqlc.arg(uid)
  UNION SELECT user_id FROM process_shares WHERE process_uid = sqlc.arg(uid)
  UNION SELECT s.user_id FROM chain c JOIN node_shares s ON s.node_id = c.id
)
SELECT user_id, process_permission(user_id, sqlc.arg(uid)) AS permission
FROM cand
WHERE process_permission(user_id, sqlc.arg(uid)) > 0;
```

- `ListVisibleLiveProcesses`는 살아 있는 process 전체에 함수를 부름. 살아 있는 process는 수십 개 수준이라 문제없음. 많아지면 받는 계정의 공유 행에서 아래로 펼치는 방식으로 교체(함수 시그니처는 그대로).
- `ListProcessViewers`의 후보 수집(`chain`)은 함수 밖에서 위로 올라가는 경로를 한 번 더 씀 — 후보를 모으는 용도이고 단계 판정은 여전히 함수 한 곳.

- **2026-10-07 판정 함수 재작성(SC1(출처 계산 위치))**: 가까운 공유 계산을 `nearest_node_share`(출처 노드 id 반환)로 옮기고 `node_permission`은 이것을 부름 — 위 SQL 블록은 옛 모양, 실제는 `00007_shares.sql` → `REF-process-share-api.md` "SC1"
- **2026-10-07 보완(`ListChildrenVisible`)**: `parent_perm = 40`이면 자식 공유 행을 무시하고 40 — 소유자 자신에게 건 공유 행이 있어도 `node_permission`과 같은 답. `parent_id`는 `::bigint`(루트 NULL 불가, 주석 명시). 판정 함수는 "하위 노드 소유자 = 상위 노드 소유자"에 기대며 DB로 강제하지 않음 — 공유-b `createNode`(소유 = 부모 소유자)·`checkMove`(같은 소유자)가 지킴

## 기존 쿼리 변경(공유-a에 포함)
| 쿼리 | 변경 | 호출부 |
|---|---|---|
| `CreateFolder` · `CreateScript` | `created_by` 인자 추가 | node 생성 핸들러: 지금은 `owner_user_id`와 같은 값(요청 계정). 받은 쪽 생성(SH8)은 공유-b에서 owner = 폴더 소유자로 바뀜 |
| `CreateProcess` | `started_by` 인자 추가 | exec 핸들러: 지금은 owner와 같은 값. SH9(owner = 노드 소유자)는 공유-b |
| `GetNode` 등 소유자 조건 쿼리 | **그대로** | 공유-b에서 판정 쿼리로 교체 |

## 구현 메모(2026-10-06)
- Go 상수 `superdb.PermBlock/PermView/PermOperate/PermEdit/PermOwner`(`gen/const.go`, 손으로 쓴 파일)
- `process_permission`은 process가 없으면 NULL → `ProcessPermission` 쿼리가 `COALESCE(..., 0)`. 목록 쿼리는 `> 0` 조건이라 NULL이 자연히 빠짐
- 함수 호출 결과 컬럼에는 `::smallint`를 붙여 sqlc가 `int16`으로 생성하게 함
- 호출부: `router/node.go` createNode(`CreatedBy: owner`), `process/manager.go` execScript(`StartedBy: owner.ID`)

## 확인 방법(공유-a 끝날 때)
- supervisor 기동 → goose가 `00007` 적용, `\df node_permission` 확인
- psql로 계정 2개 + 폴더 3단 + 공유 행(상위 보기, 하위 차단)을 넣고 `SELECT node_permission(...)`이 가까운 공유를 따르는지 확인 → 테스트 데이터는 `REF-test-env.md`에 기록
- `go build ./... && go vet ./...`
- **2026-10-06 실제 확인(임시 DB `nexus_share_tmp`, 확인 뒤 삭제)**: 00001~00007 Up 적용 → 계정 3(소유자 1, 받는 2·3) + 폴더 1>2>3>스크립트 4, 폴더 2>스크립트 5 + 공유(1→계정2 보기, 3→계정2 차단, 2→계정3 편집) + process `pa`(노드 4, 계정2에 단독 공유 조작)·`pb`(노드 5)·`pc`(노드 없음, 끝남)
  | 확인 | 결과 |
  |---|---|
  | 계정2: 노드 1·2·5 = 10, 노드 3·4 = 0(차단이 가까운 공유) | 맞음 |
  | 계정2: `pa` = 20(process 단독 공유가 노드 차단보다 우선), `pb` = 10 | 맞음 |
  | `ListProcessViewers(pa)` = 1(40)·2(20)·3(30) / `pc` = 1만 | 맞음 |
  | `ListChildrenVisible(계정2, 부모 2)` = 노드 5만(3은 차단으로 제외) | 맞음 |
  | 없는 노드 = 0, `ProcessPermission`(없는 uid) = 0 | 맞음 |
  | Down 적용 뒤 `created_by`·`started_by` 컬럼 0개 | 맞음 |
- 재현 방법: 마이그레이션 Up 구간만 잘라 `docker exec -i postgres15 psql`로 흘려 넣음(goose 없이). 개발 DB와 떠 있는 supervisor를 건드리지 않음
