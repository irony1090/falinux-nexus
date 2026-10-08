# REF — process 공유: 공유-b(인가 교체) 구조안

> 2026-10-06 신설. 공유 결정·작업 순서 → `REF-process-share.md` / 단계 정의 → `REF-process-share-node.md` / 판정 함수·쿼리 → `REF-process-share-db.md` / 이력 → `history/process-share-authz.md` / 현재 진행 → `CURRENT.md`.

**상태(2026-10-07): 인가-1(판정 헬퍼)~인가-8(확인) 코드 작성·확인 완료(커밋 전). build/vet 통과 + 격리 supervisor로 53개 항목 통과(아래 "확인 결과").**

## 범위
| 넣음(공유-b) | 넣지 않음 |
|---|---|
| 소유자 검사 → 단계 판정(`superdb.Perm*`) 교체: process REST·소켓 입력·구독·크기 우선권·exec·node REST | 공유 걸기·풀기 API → 공유-c(공유 API) |
| SH8(받은 쪽이 만든 노드의 소유)·SH9(받은 쪽이 실행한 process의 소유) 적용 | 공유가 바뀔 때 구독 추가·해제, `SHARE:CHANGE` → 공유-d(실시간) |
| 응답에 `permission` 필드(프론트가 보기 전용 등을 판단) | 받는 쪽 타일 트리(`/` 루트 폴더 타일, exec 때 받는 쪽 트리) → 공유-d |
| `NODE:0` 구독 삭제(SB5) | 받는 쪽 루트 목록(`ListShareRoots`) → 공유-d |

공유 행을 만드는 API가 아직 없으므로 공유-b가 끝나도 실제 동작은 소유자 기준 그대로다. 받는 쪽 동작 확인은 psql로 공유 행을 직접 넣어서 한다.

## 결정 (SB = 공유-b, 2026-10-06 전부 제안대로 확정)
| 결정 | 묻는 것 | 제안 |
|---|---|---|
| SB1(권한 부족 응답) | 단계가 모자랄 때 상태 코드 | **단계 0(볼 수 없음) = 404**(존재를 드러내지 않음, 지금 `requireOwnedEntry`와 같음) / **보기 이상인데 요구 단계 미달 = 403**. 프론트가 "권한 없음"과 "없음"을 구분할 수 있음 |
| SB2(입력 권한 확인 방식) | 키 입력마다 단계를 어떻게 확인하나 | **소유자면 메모리(`Record.OwnerUserID`)로 끝, 아니면 입력마다 `ProcessPermission` DB 조회.** 사람 입력 속도(붙여넣기도 1프레임)라 부담 작고 공유 변경이 즉시 반영됨. 기각: 소켓별 단계 캐시(공유가 바뀔 때 무효화가 필요한데 그 신호는 공유-d에서 생김 — 공유-c~d 사이에 낡은 권한으로 입력이 통과) |
| SB3(크기 우선권 후보) | 크기 우선권을 넘길 다음 탭, 연결 때 빈 우선권을 누가 가져가나 | 후보 = **조작 이상 계정의 탭만**(보기 단계 탭은 후보 아님). 순서 = **process 소유 계정 탭 먼저 → 다른 계정 탭**, 각각 연결 순. `sizeOwnerOnConnect`도 조작 이상 uid만 넘김. 작업 공간 기준으로 더 좁히는 것은 작업공간-c |
| SB4(노드 쓰기 쿼리의 owner 인자) | `RenameNode`·`PatchNode`·`DeleteNode` 등 `owner_user_id = $x` 조건을 어떻게 하나 | **쿼리는 그대로 두고 인자를 "세션 계정" → "판정으로 확인한 대상 노드의 소유자"로 바꿈.** owner 조건이 "대상 노드의 실제 소유자"를 한 번 더 확인하는 안전망으로 남음. 기각: owner 조건 삭제(판정 누락 시 막을 것이 없음) |
| SB5(`NODE:0` 누설 수정) | 모든 소켓이 `NODE:0`을 구독해서 **다른 계정의 루트 노드 생성·수정·삭제 이벤트(내용 `content` 포함)가 전 계정에 전송됨**(2026-10-06 코드 읽고 발견, 프론트는 `NODE:*`를 아직 안 들어서 화면엔 안 보임) | **개정(2026-10-06, 사용자): `subscribe.go:43`의 `NODE:0` 고정 구독 한 줄 삭제.** 발행 코드는 그대로 — 구독자가 없으면 `Hub.Publish`가 바로 nil 반환(`hub.go:171`, 전송·로그 없음). `node.go:38` 주석의 "NODE:0 고정 구독" 언급도 같이 고침. 토픽 이름 설계는 프론트가 `NODE:*`를 받기 시작할 때 열린 질문 4(NODE 동적 구독)에서. 처음 안 `NODE:ROOT:<userId>`는 보류. 폴더 토픽 `NODE:<parentId>`는 아무도 구독하지 않아 지금은 누설 없음 — 동적 구독은 열린 질문 4(NODE 동적 구독)에서 판정과 함께. **사용처 검사(2026-10-06)**: 프론트 소켓 수신 등록은 `TAB:ID`·`TILES:UPDATE`·`DATA`·`STATUS`·`PROCESS:UPDATE`·`PROCESS:SIZE_OWNER` 6종뿐, `websocket.hook.ts` dispatch는 등록 안 된 type을 버림(전체 수신 핸들러 없음). 서버 쪽 `NODE:` 사용은 `subscribe.go:43` 구독 + `node.go` 발행뿐, 다른 클라이언트 없음 → 수신하는 곳 없음, 전송만 일어남 |
| SB6(이동 사이클 검사) | 기존 "핸들러 책임 미적용"인 자기 자손으로 이동 금지를 같이 넣나 | **같이 넣음** — 이동 검사를 어차피 새로 쓰므로 쿼리 하나(`IsDescendant`) 추가 |

## 엔드포인트별 요구 단계
| 자리 | 지금 | 공유-b 후 |
|---|---|---|
| `GET /processes`(살아 있는 목록) | `ListLiveProcessesByOwner` | `ListVisibleLiveProcesses` + 응답 `permission` |
| `GET /processes/lookup` | `ListProcessesByUids`(소유분) | `ListVisibleProcessesByUids` + 응답 `permission` |
| `GET /processes/snapshot/:id` | `ListProcessesByUids`(소유분) | 보기(10) |
| `POST /processes/kill/:id` | `requireOwnedEntry` | 조작(20) |
| `POST /processes/resize/:id` | `requireOwnedEntry` | 조작(20) + 크기 우선권 |
| 소켓 `PROCESS:INPUT` | `OwnerUserID != userID`면 버림 | 조작(20) 미만이면 버림(SB2(입력 권한 확인 방식)) |
| 소켓 연결 때 구독 | `ListLive(userID)` | `ListVisibleLiveProcesses` |
| exec·재접속 때 구독 | `subscribeUser(uid, owner)` | `subscribeViewers(uid)` = `ListProcessViewers` 계정 전부의 소켓 |
| `POST /processes/exec` | `GetNode(owner=세션)` | EXEC = 조작(20), EDIT(vi 편집) = 편집(30). process 소유 = 노드 소유자, `started_by` = 세션(SH9) |
| `GET /workers?nodeId=` | `resolveMainKey(owner=세션)` | 조작(20), device_key 상속 조회는 노드 소유자로 |
| `GET /nodes?parentId=` | `ListChildren(owner=세션)` | 부모 보기(10) → `ListChildrenVisible`(차단 자식 제외) + 응답 `permission`. **parentId 없음(루트) = 내 루트만**(공유받은 루트는 공유-d) |
| `GET /nodes/:id` | `GetNode(owner=세션)` | 보기(10) + 응답 `permission` |
| `POST /nodes` | owner = 세션 | 부모 있으면 부모 편집(30), owner = 부모 소유자, `created_by` = 세션(SH8). 루트 생성 = 내 것 |
| `PATCH /nodes/:id`(이름·내용·장비·순서) | owner = 세션 | 편집(30) |
| `PATCH /nodes/:id`(이동) | owner = 세션 | 대상 편집(30) + **대상 아래에 편집 미만 공유 행 없음** + 목적지 편집(30) + 목적지 소유자 = 대상 소유자 + 루트(NULL)로 이동은 소유자만 + 사이클 금지(SB6) |
| `DELETE /nodes/:id` | owner = 세션 | 대상 편집(30) + 대상 아래에 편집 미만 공유 행 없음(차단 하위가 `ON DELETE CASCADE`로 지워지는 것 방지, SH8) |
| `editResult`(vi 편집 회수) | `GetNode(owner=process 소유자)` | **그대로** — SH9로 process 소유자 = 노드 소유자라 성립 |

## 새 쿼리
| 쿼리 | 파일 | 용도 |
|---|---|---|
| `GetNodeByID(id)` | `nodes.sql` | owner 조건 없는 단건 조회 — **판정 뒤에만** 호출(주석에 명시) |
| `HasLowerShareBelow(user, node, min)` | `shares.sql` | 대상 아래(자기 제외) 그 계정의 `permission < min` 공유 행이 있나. 삭제·이동 검사 |
| `IsDescendant(ancestor, node)` | `nodes.sql` | SB6(이동 사이클 검사) |
| 제거: `ListLiveProcessesByOwner`·`ListProcessesByUids` | `processes.sql` | 교체 뒤 호출부 없음 |

## 작업 단위
| 단위 | 내용 | 파일 |
|---|---|---|
| 인가-1(판정 헬퍼) | `requireNode(c, id, min) (Node, perm)` · `requireProcess(uid, user, min) *ProcessEntry`(소유자면 메모리, 아니면 DB) — SB1 404/403 | `router/perm.go`(새 파일) |
| 인가-2(process REST) | kill·resize·snapshot·list·lookup, 응답 `permission` | `processApi.go`, `screen.go`, `processDto.go` |
| 인가-3(소켓) | 연결 때 구독, `subscribeViewers`, 입력(SB2) | `subscribe.go`, `process.go`, `input.go` |
| 인가-4(크기 우선권) | SB3 후보·연결 때 필터 | `sizeOwner.go` |
| 인가-5(exec) | 단계 검사 + `Exec`/manager 인자를 `owner superdb.User` → `ownerID, startedBy int64`로 | `processApi.go`, `process.go`, `workerApi.go`, `process/manager.go` |
| 인가-6(node REST) | 위 표 + 이동·삭제 경계 + SB6, 응답 `permission` | `node.go`, `nodeDto.go`, `query/nodes.sql`, `query/shares.sql` |
| 인가-7(루트 토픽) | SB5(`NODE:0` 누설 수정): `NODE:0` 구독 줄 삭제 + 주석 수정 | `subscribe.go`, `node.go`(주석만) |
| 인가-8(확인) | build/vet + 계정 2개, psql로 공유 행 삽입 → curl로 엔드포인트별 200/403/404, 입력·구독은 헤드리스 | `REF-test-env.md` |

## 알아 둘 한계(공유-d 전까지)
- 공유가 바뀌어도 이미 걸린 소켓 구독은 그대로(입력·REST는 매번 판정하므로 막힘, 출력 수신만 남음) — 공유-d의 SH5(공유 시작·취소 반영)에서 해제
- 소유자가 공유 폴더 아래에서 실행한 터미널은 받는 쪽 소켓으로 출력이 가지만 받는 쪽 타일 트리에는 아직 안 생김 — 공유-d

## 구현 메모(2026-10-07)
| 단위 | 실제 코드 |
|---|---|
| 인가-1(판정 헬퍼) | `router/perm.go`: `checkPerm(perm, min, 404문구)`(SB1) · `requireNodePerm`(단계만) · `requireNode`(단계 + `GetNodeByID`) · `processPerm`(소유자면 메모리, 아니면 `ProcessPermission`, 조회 실패 = 0) · `requireProcess`(판정 뒤 memory entry, 없으면 404). `requireOwnedEntry` 삭제 |
| 인가-2(process REST) | `processResponse.Permission`(`omitempty`, REST만) + `withPermission`. 목록 행은 `visibleProcess`로 통일(`liveRows`/`uidRows`). snapshot = `processPerm` 보기 + `GetProcess`로 크기. resize 응답에도 permission |
| 인가-3(소켓) | `subscribeLiveProcesses`가 조작(20) 이상 uid만 돌려줌(크기 우선권 후보). `subscribeUser` → `subscribeViewers(uid)`(`ListProcessViewers` 계정 전부의 소켓). 입력 = `processPerm < 조작`이면 버림 + 로그 |
| 인가-4(크기 우선권) | `processOwner` 삭제 → `sizeOwnerCandidates(uid)`: `ListProcessViewers`에서 40 = 소유 계정 탭(연결 순) 먼저, 20·30 계정 탭은 합쳐서 연결 순 |
| 인가-5(exec) | `Exec(startedBy, ...)` — 소유 = `node.OwnerUserID`. manager `Exec`/`ExecEdit`/`openFolder`/`execScript` 인자 `ownerID, startedBy int64`. EXEC = 조작, EDIT = 편집. `resolveMainKey`·`/workers?nodeId=` = 노드 소유자로 |
| 인가-6(node REST) | 위 표대로. `checkMove`: **같은 부모 안 순서 변경(`parentId` = 지금 부모)은 이동 검사 생략**(받는 쪽 순서 변경이 아래 공유 행 때문에 막히지 않게). 검사 순서 = 아래 편집 미만 공유 → 루트(소유자만) → 목적지 편집 → 같은 소유자(403) → 사이클(400). 이동 뒤 응답 permission은 새 위치로 다시 판정. `newOwnedNodeResponses`(루트, 40) / `newVisibleNodeResponses` |
| 인가-7(루트 토픽) | `subscribe.go`의 `NODE:0` 구독 줄 삭제, `node.go` `parentOrRoot` 주석 수정 |

- 쿼리: 목록 3개(`ListChildrenVisible`·`ListVisibleLiveProcesses`·`ListVisibleProcessesByUids`)를 `sqlc.embed`로 바꿈 → 행 = `{Node|Process, Permission}`. 새 쿼리 `GetNodeByID`·`IsDescendant`(nodes.sql)·`HasLowerShareBelow`(shares.sql). `ListLiveProcessesByOwner`·`ListProcessesByUids` 삭제
- **sqlc 함정(1.31.1)**: 재귀 CTE를 `EXISTS (SELECT ... FROM cte)` 안에서 쓰거나 `cte.col`로 한정하면 `table alias "..." does not exist` → `SELECT (COUNT(*) > 0)::boolean FROM cte WHERE col = sqlc.arg(x)::bigint`로 씀. 비교 대상이 NULL 허용 컬럼(`parent_id`)이면 인자가 `pgtype.Int8`로 생성되므로 `::bigint` 캐스트
- 상태 코드: 소유자 다른 폴더로 이동 = 403, 자기 아래로 이동 = 400, 루트로 이동(받는 쪽) = 403, 아래에 편집 미만 공유 = 403

## 확인 결과(2026-10-07)
격리 실행(→ `REF-test-env.md` "격리 실행")으로 계정 A(소유자)·B(받는 쪽)·C(무관) + psql 공유 행. 스크립트 = 세션 scratchpad `authz_test.py`(Python `websockets`, 재부팅 시 사라짐).

| 영역 | 확인 항목(전부 통과) |
|---|---|
| node REST | 목록: 차단 자식 제외·자식별 permission / 루트 = 내 것만 / 차단·차단 상속·무관 계정·없는 노드 = 404 / 조작으로 이름 변경·하위 생성 = 403 / 편집 폴더 아래 생성 = owner A·created_by B / 보기로 삭제 = 403 / 목적지 조작 = 403 / 자기 아래 = 400(받는 쪽·소유자 둘 다) / 루트로 = 403 / 같은 부모 순서 변경 = 200 / 소유자 다른 폴더 = 403 / 아래 차단 공유 있는 폴더 삭제 = 403 |
| workers | 조작 = 200 / 보기 = 403 / 무관 = 404 |
| exec | 조작 EXEC = 200, owner A·started_by B·permission 20 / 조작으로 EDIT = 403 / 보기 = 403 / 무관 = 404 |
| 입력·구독 | B 입력 → A·B 소켓 모두 출력 / 무관 C 입력 버림·구독 안 됨 / 새 소켓은 연결 때 구독 / process 단독 보기 공유 C: 목록 permission 10, snapshot 200, kill 403, 입력 버림, 출력 수신 |
| 크기 우선권 | 실행 탭 = 소유 / 소유자라도 다른 탭 resize = 403 / B 탭 모두 끊김 + 60초 → 소유 계정 A 탭으로 넘어감, 보기 C 탭은 후보 아님 |
| SB5 | A 루트 노드 생성 → 어떤 소켓에도 `NODE:*` 없음 |
| 공유-a 보완 | 소유자 자신에게 건 공유 행(단계 0)이 있어도 자식 목록 단계 40 그대로 |
