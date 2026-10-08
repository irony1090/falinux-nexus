# REF — process 공유: 공유-c(공유 API) 구조안

> 2026-10-07 신설. 공유 결정·작업 순서 → `REF-process-share.md` / SH3(받은 계정 권한) 단계·SH10(노드 복사) → `REF-process-share-node.md` / 판정 함수 → `REF-process-share-db.md` / 판정 헬퍼(`perm.go`) → `REF-process-share-authz.md` / 이력 → `history/process-share-api.md` / 현재 진행 → `CURRENT.md`.

**상태(2026-10-07): SC1(출처 계산 위치)~SC7(복사 제외 개수) 제안대로 확정(사용자 "진행해줘") → 공유API-1(판정 함수 정리)~공유API-7(확인) 코드 작성·확인 완료(커밋 전, 파트 마감 때 커밋). 42개 항목 + 공유-b 회귀 50개 통과.**

## 범위
| 넣음(공유-c) | 넣지 않음 |
|---|---|
| SH4(공유 대상 찾기): 아이디로 계정 조회 | 공유가 바뀔 때 구독 추가·해제, `SHARE:CHANGE` → 공유-d(실시간) |
| 노드 공유 걸기(upsert)·풀기·목록(O2(적용 공유 목록)에 필요한 출처 포함) | 받는 쪽 `/` 루트 목록·폴더 타일 → 공유-d |
| process 단독 공유 걸기·풀기·목록 | 숨기기 → 공유-e(프론트) |
| SH10(노드 복사) API — SC4(노드 복사 시점) | O3(행 공유 개수)·O5(내가 공유한 것 전체) → SC5(소유자 쪽 집계 API) |

**알아 둘 한계(공유-d 전까지)**: 공유를 걸어도 받는 쪽에 **이미 연결된 소켓**은 구독이 늘지 않는다(REST 조회·새로 연결한 소켓은 바로 반영). 풀어도 이미 걸린 구독은 남는다(입력·REST는 매번 판정하므로 막힘, 출력 수신만 남음). 공유-b의 한계와 같은 것.

## API
| 메서드 · 경로 | 권한 | 요청 | 응답 |
|---|---|---|---|
| `GET /users/lookup?identification=` | 로그인 | — | `{id, nickname}` / 없는 아이디 404 / 자기 자신 400 |
| `GET /nodes/:id/shares` | 노드 소유자(40). 보기 이상인데 소유자 아님 403 / 볼 수 없음 404 | — | 이 노드에 실제로 적용되는 공유 전부(직접 + 상위에서 이어진 것, 차단 포함, 소유자 제외): `[{userId, nickname, permission, sourceNodeId, sourceName, direct}]` |
| `PUT /nodes/:id/shares/:userId` | 노드 소유자, **폴더만**(스크립트 = 400, 2026-10-07 SH1 개정) | `{permission: 0·10·20·30}` | 그 계정의 위 행 1개 + `previous`(이 노드에 있던 직접 단계, 없으면 null) |
| `DELETE /nodes/:id/shares/:userId` | 노드 소유자 | — | 지운 뒤 그 계정의 행(상위에서 이어지면 그 값) 또는 `null` |
| `GET /processes/:uid/shares` | process 소유자 | — | `[{userId, nickname, permission, direct}]`(`direct` = process 단독 공유 행인가). `ListProcessViewers` + 차단 행 |
| `PUT /processes/:uid/shares/:userId` | process 소유자 | `{permission}` | 위 행 1개 + `previous` |
| `DELETE /processes/:uid/shares/:userId` | process 소유자 | — | 지운 뒤 행 또는 `null` |
| `POST /nodes/:id/copy` | 원본 보기(10) + 목적지 편집(30), 목적지 없음 = 내 루트 | `{parentId: number \| null}` | `{nodes: [만든 노드(최상위만, permission 포함)]}` |

공유 걸기(PUT) 서버 검사 — SH4(공유 대상 찾기) 그대로:
| 검사 | 응답 |
|---|---|
| 노드(또는 process) 소유자가 아님 = 재공유 불가(P(공유 권한)) | 볼 수 없음 404 / 보기 이상 403 |
| `userId` 계정 없음 | 404 |
| `userId` = 소유자 자신 | 400 |
| `permission`이 0·10·20·30 밖 | 400(DB CHECK 전에 거름) |
| 같은 노드·계정에 행이 이미 있음 | 단계만 바꿈(`ON CONFLICT DO UPDATE`, `updated_at = NOW()`) |

## 결정 (SC = 공유-c, 2026-10-07 전부 제안대로 확정)
| 결정 | 묻는 것 | 제안 |
|---|---|---|
| SC1(출처 계산 위치) | 공유 목록의 출처("여기 / 상위 폴더 ○○")를 어디서 구하나 | **판정 함수를 `nearest_node_share(user, node) RETURNS TABLE(node_id, permission)` 하나로 모으고 `node_permission`은 이것을 부르게 고침.** 가까운 공유 계산이 계속 한 곳. `00007_shares.sql`은 개발 DB `nexus`에 아직 적용 전(임시 DB에만 적용했다 지움)이라 새 마이그레이션 없이 고쳐도 됨. 기각: 목록 쿼리에서 `DISTINCT ON`으로 다시 계산(SA2(판정 위치) 규약 위반) / 출처를 안 보여 줌(O2의 "여기서 바꾸기"를 못 함) |
| SC2(차단 단독 공유) | 이어받는 공유가 없는 노드에 차단(0)을 걸면 | **허용.** 결과가 공유 없음과 같아 무해하고, 상위 공유를 풀면 이런 행은 어차피 생기므로 서버에서 막아도 불변식이 되지 않음. 프론트가 선택지를 끔(공유 걸기 창 제안 그대로) |
| SC3(URL 모양) | 공유 API 경로 | **대상 아래 중첩**(`/nodes/:id/shares/:userId`, PUT = upsert). 대상 종류별 소유자 검사가 경로로 나뉨. 기각: `POST /shares {nodeId \| processUid, userId, permission}`(핸들러 안에서 대상 종류 분기) |
| SC4(노드 복사 시점) | SH10(노드 복사)을 공유-c에 넣나 | **넣음(마지막 단위).** 판정 헬퍼를 그대로 쓰고 타일 트리와 무관. 기각: 공유-e(프론트) 때(그때 API·확인을 따로 한 번 더 함) |
| SC5(소유자 쪽 집계 API) | O3(행 공유 개수)·O5(내가 공유한 것 전체) API | **공유-e 때.** O3·O5가 아직 제안 단계(시안 보며 확정)라 API 모양도 그때 정함 |
| SC6(process 단독 공유 대상) | 끝난 process에도 공유를 걸 수 있나 | **살아 있는 것(PENDING·PROCESS)만, 아니면 400.** 끝난 process는 화면 복원 버퍼가 없어 공유해도 볼 것이 없음. 이미 걸린 행은 process가 끝나도 그대로(지울 일 없음) |
| SC7(복사 제외 개수) | 복사 응답에 "볼 수 없는 하위 N개 제외"를 넣나(SH10에서 "있으면 좋음") | **넣지 않음.** 차단은 존재를 숨기는 단계라 개수도 드러내지 않음(SB1(권한 부족 응답)의 404와 같은 이유) |

## SC1(출처 계산 위치) — 함수 모양(구현됨)
```sql
CREATE FUNCTION nearest_node_share(p_user BIGINT, p_node BIGINT) RETURNS BIGINT
LANGUAGE sql STABLE AS $$
  WITH RECURSIVE chain(id, parent_id, depth) AS ( ... WHERE c.depth < 64 )
  SELECT c.id FROM chain c
  JOIN node_shares s ON s.node_id = c.id AND s.user_id = p_user
  ORDER BY c.depth LIMIT 1
$$;
```
- 반환 = 가까운 공유의 **출처 노드 id**(없으면 NULL). 단계는 `(node_id, user_id)`로 `node_shares`에서 읽음
- `node_permission` = 소유자면 40, 아니면 `COALESCE((SELECT s.permission FROM node_shares s WHERE s.node_id = nearest_node_share(u, n) AND s.user_id = u), 0)`
- 처음 안(`RETURNS TABLE(node_id, permission)`)은 **sqlc 1.31.1이 그 컬럼을 못 읽음**(`column "node_id" does not exist`) → 스칼라로 바꿈
- 목록 쿼리(`ListNodeShareEntries`) = 노드와 조상에 공유 행이 있는 계정(후보) → `node_shares`를 `node_id = nearest_node_share(후보, 노드)`로 조인 → 출처 이름. `direct`(출처 = 이 노드)는 Go에서 붙임. 소유자 제외

## SH10(노드 복사) 구현 모양
| 항목 | 내용 |
|---|---|
| 하위 수집 | Go에서 폴더마다 `ListChildrenVisible`을 불러 아래로 내려감 → 차단 자식(과 그 아래 전체)은 자연히 빠짐, 판정은 기존 쿼리 그대로 |
| 순서 | **원본 하위를 전부 읽은 뒤 넣음** → 폴더를 자기 아래로 복사해도 무한 반복 없음 |
| 소유 | 복사본 소유자 = 목적지 소유자(목적지 없음 = 나), `created_by` = 복사한 사람(SH8과 같은 규칙) |
| 장비 지정 | 노드 자신의 `device_key`만 복사(상속값은 안 따라감) |
| 위치 | 최상위 복사본은 목적지 형제 끝(ord = max+1), 하위는 원본 ord 유지 |
| 한 번에 복사할 수 있는 수 | 상한 둠(예: 1000개, 넘으면 400) — 실수로 큰 트리를 복사하는 것 방지 |
| 실시간 | 최상위 복사본만 `NODE:CREATE` 발행(지금 구독자는 없음, SB5) |

## 작업 단위
| 단위 | 내용 | 파일 |
|---|---|---|
| 공유API-1(판정 함수 정리) | SC1: `nearest_node_share` + `node_permission` 재작성 | `migrations/00007_shares.sql` |
| 공유API-2(쿼리) | `UpsertNodeShare`·`DeleteNodeShare`·`ListNodeShareEntries` · process 쪽 같은 3개 | `query/shares.sql` |
| 공유API-3(대상 조회) | `GET /users/lookup` | `router/user.go` |
| 공유API-4(노드 공유) | `GET/PUT/DELETE /nodes/:id/shares...` | `router/share.go`(새 파일), `shareDto` |
| 공유API-5(process 공유) | `GET/PUT/DELETE /processes/:uid/shares...` + SC6 | `router/share.go` |
| 공유API-6(노드 복사) | `POST /nodes/:id/copy` | `router/node.go` 또는 `nodeCopy.go` |
| 공유API-7(확인) | build/vet + 격리 실행(`REF-test-env.md` "격리 실행"): 계정 3개로 조회·걸기·upsert·풀기·목록 출처·재공유 403·자기 자신 400·복사(차단 하위 제외·자기 아래 복사·소유자·장비) | — |

## 구현 메모(2026-10-07)
| 단위 | 실제 코드 |
|---|---|
| 공유API-1(판정 함수 정리) | `00007_shares.sql`: `nearest_node_share` 추가 + `node_permission` 재작성 + Down에 DROP. 개발 DB `nexus`는 그때까지 버전 6(00007 미적용) 확인 후 고침 |
| 공유API-2(쿼리) | `shares.sql`: `Get/Upsert/DeleteNodeShare`·`ListNodeShareEntries` + process 쪽 같은 4개. Delete는 `:execrows`(0행 = 404) |
| 공유API-3(대상 조회) | `user.go` `GET /users/lookup` — 빈 아이디도 400 |
| 공유API-4·5(노드·process 공유) | `router/share.go`(새 파일, `mountShares`): 소유자 검사 = `requireNode(..., PermOwner)` / `processPerm` + `checkPerm(PermOwner)`. 검사 순서 = 소유자 → 받을 계정 있음(404) → 자기 자신(400) → 단계 값(400). PUT·DELETE 응답의 `entry`는 바뀐 뒤 목록을 다시 읽어 그 계정 행을 찾음(없으면 null) |
| 공유API-6(노드 복사) | `router/nodeCopy.go`(`POST /nodes/:id/copy`): 목적지가 스크립트면 400(구조표에 없던 검사 — 새 코드라 넣음). `copyMax` = 1000 |

## 확인 결과(2026-10-07)
격리 실행으로 계정 A(소유자)·B·C. 스크립트 = 세션 scratchpad `share_test.py`(`authz_test.py`의 클라이언트를 import).

| 영역 | 확인 항목(전부 통과) |
|---|---|
| 대상 조회 | 있는 아이디 200 `{id, nickname}` / 없음 404 / 자기 자신·빈 값 400 |
| 노드 공유 | 걸기(direct·출처) / 다시 걸기 = upsert(`previous`, 행 1개) / 하위 목록에 상위 출처로 보임 / 차단 걸기 → 받는 쪽 404 / SC2 차단 단독 허용 / 조작·편집 단계 받은 쪽이 다시 공유 = 403, 목록 = 403 / 무관 = 404 / 자기 자신 400 / 없는 계정 404 / 단계 15·40·없음 400 / 풀면 상위 단계로 돌아감 / 다시 풀기 404 |
| process 공유 | 단독 공유(direct) / 목록 = 노드에서 이어진 것 + 단독 / 받은 쪽 목록·공유 = 403 / 자기 자신 400 / 없는 process 404 / 풀기 → null / 끝난 process 공유 400(SC6), 목록은 200 |
| 노드 복사 | 보기 이상 → 내 루트(40) / 차단 하위 빠짐 / 소유·`created_by` = 복사한 사람 / 폴더 자신의 `device_key`·스크립트 내용 복사 / 상속받던 스크립트 복사 = 장비 미지정 / 차단 원본 404 / 목적지 조작 403 / 목적지 스크립트 400 / 남의 편집 폴더로 = owner 목적지 소유자·permission 30 / 자기 아래로 복사 = 원본 수만큼만 늘어남 |
| 회귀 | 함수 재작성 뒤 공유-b `authz_test.py` 50개 다시 통과 |
