# REF — 공유-d 실시간-1~3: 구독·권한 재계산(`reconcileAccess`)과 연결

> 2026-10-07 신설. 공유-d 전체 결정(SD1~SD9)·작업 단위 → `REF-process-share-realtime.md` / 판정 함수 → `REF-process-share-db.md` / 인가 → `REF-process-share-authz.md` / 이력 → `history/process-share-reconcile.md` / 현재 진행 → `CURRENT.md`.

**상태(2026-10-08): 실시간-1(재계산 함수)·실시간-2(공유 API 연결)·실시간-3(노드 이동·삭제 연결)·실시간-4(exec 타일)·실시간-5(노드 응답) 코드·확인 완료(커밋 전). 다음 = 실시간-6a(스크립트 행 "실행 중 N").**

## 지금 코드 (2026-10-07 확인)
| 자리 | 지금 |
|---|---|
| 소켓 연결 | `subscribeLiveProcesses`: `ListVisibleLiveProcesses`(볼 수 있는 살아 있는 것) 전부 구독 → 조작(20) 이상 uid로 `sizeOwnerOnConnect` |
| exec·재접속 | `subscribeViewers(uid)`: `ListProcessViewers` 계정의 소켓 전부 구독(추가만) |
| 구독 해제 | 소켓 끊길 때 `UnsubscribeAll`뿐. 토픽 하나 해제(`Hub.Unsubscribe(key, conn)`)는 있으나 안 씀 |
| 크기 우선권 | `resyncSizeOwner(uid)`: 후보(`sizeOwnerCandidates` = 조작 이상 계정 + 그 터미널이 있는 보드를 보는 탭)에 지금 소유자가 없으면 첫 후보로 넘김. 끊긴 소유자(유예 타이머)는 건드리지 않음 → **공유를 풀면 이 함수만 부르면 소유권이 정리됨** |
| 입력 인가 | 입력마다 `processPerm`(DB) → 구독과 별개로 이미 막힘. 구독은 출력 전달만 |
| 노드 삭제 | `processes.node_id`는 FK 없음 → 노드를 지우면 받는 쪽 단계가 0이 됨(소유자는 40 유지) |

## 구조안
```
share PUT/DELETE, node move/delete (caller, after commit)
   |  users = affected accounts, uids = affected live processes
   v
reconcileAccess(users, uids)        -- accessMu held
   1. ListProcessPermissionMatrix(users x uids)   -- one query
   2. perm > 0 : Subscribe(PROCESS:<uid>, conns of user)
      perm = 0 : Unsubscribe(PROCESS:<uid>, conns of user)
   3. resyncSizeOwner(uid) for each uid
   4. publish SHARE:CHANGE on TILES:<user> for each user
```

| 파일 | 변경 |
|---|---|
| `query/shares.sql` | `ListProcessPermissionMatrix`(계정 배열 × uid 배열 → 단계, `process_permission`만 호출) / `ListLiveProcessesUnder(node_id)`(그 노드와 하위의 살아 있는 process) / `ListShareUsersAround(node_id)`(그 노드의 조상·자신·하위에 공유 행이 있는 계정 + 하위 살아 있는 process의 단독 공유 계정) |
| `router/shareAccess.go`(새) | `accessMu` + `reconcileAccess` + `publishShareChange` |
| `protocol/messages.go` | `MsgShareChange = "SHARE:CHANGE"` + `ShareChangeEvent{}`(내용 없음 — SD6(실효 단계 재조회): 받으면 다시 조회) |
| `subscribe.go` | `subscribeLiveProcesses`를 `accessMu` 안에서 |

### 정할 것 (2026-10-07 제안대로 확정)
| 질문 | 확정 | 이유 |
|---|---|---|
| `SHARE:CHANGE` 받을 계정 | **넘겨받은 `users` 전부**(단계가 실제로 바뀌었는지 안 따짐) | 노드 공유는 살아 있는 process가 없어도 폴더 목록이 바뀌어 알려야 함. 바뀌었는지 보려면 전·후 단계를 따로 구해야 하는데, 받은 쪽은 다시 조회만 하므로 얻는 것이 없음 |
| 경합 | **`accessMu` 하나로 재계산과 소켓 연결 구독을 한 줄로** | 재계산 A가 "볼 수 없음"을 읽은 뒤, 공유를 다시 건 재계산 B가 구독을 넣고, A가 늦게 해제하면 구독이 빠짐. 둘 다 잠금 안에서 커밋된 DB를 읽고 적용하면 마지막 커밋 기준으로 맞음. exec의 `subscribeViewers`는 새 uid에 추가만 해서 그대로 둠 |
| 영향 계산 위치 | 쿼리는 실시간-1에서 만들고, 호출(전·후 수집)은 실시간-2(공유 API 연결)·실시간-3(노드 이동·삭제 연결)에서 | 이동은 옮기기 **전** 조상과 **후** 조상 양쪽 공유가 걸리므로 호출부가 두 번 모아 합침. 삭제는 전만 |
| 확인 | 실시간-1은 `go build`·`vet` + 쿼리 결과만 확인(`-overlay`, `REF-test-env.md`). 동작 확인은 실시간-2에서 공유 걸기·풀기로 | 실시간-1만으로는 부르는 곳이 없음 |

## 실시간-1 구현 메모 (2026-10-07)
| 자리 | 한 것 |
|---|---|
| `shareAccess.go` | `reconcileAccess(users, uids)` = 중복 제거 → uids 있으면 `applyAccess`(잠금 안: 단계 행렬 조회 → `connsOfUser` 구독/해제) → `resyncSizeOwners`(잠금 밖, 크기 우선권 자체 잠금 사용) → users 전부 `publishShareChange`. 단계 조회 실패 = 로그만, SHARE:CHANGE는 그대로 보냄(받은 쪽 다시 조회). 제네릭 `uniq` |
| `supervisorRouter.go` | `accessMu sync.Mutex` 필드 |
| `subscribe.go` | `subscribeLiveProcesses` 전체를 `accessMu` 안에서 |
| `messages.go` | `MsgShareChange`·`ShareChangeEvent{}` |
| 아직 부르는 곳 없음 | `reconcileAccess`·`ListLiveProcessesUnder`·`ListShareUsersAround` → 실시간-2·3에서 |

## 실시간-1 확인 결과 (2026-10-07, 임시 DB `nexus_rt1_tmp`, 확인 뒤 삭제)
- 데이터: A 소유 F1 > F2 > S, F2 > F3 > S3, G > S2 / 공유 B→F1 보기(10), C→F2 차단(0), D→F3 조작(20) / process p1(S, PROCESS)·p2(S2, PENDING)·p3(S, COMPLETED)·p4(S3, PENDING), 단독 공유 C→p1 조작(20)·C→p2 보기(10)
- `ListProcessPermissionMatrix`: A 전부 40 · B p1·p3·p4 = 10, p2 = 0 · C p1 = 20(단독 공유가 차단보다 우선), p2 = 10, p3·p4 = 0 · D p4만 20 · 없는 uid = 0 — 전부 기대값
- `ListLiveProcessesUnder`: F1·F2 → p1·p4(COMPLETED p3 제외) / F3 → p4 / G → p2 / 스크립트 S → p1
- `ListShareUsersAround`: S → B·C / F3 → B·C·D / G → C(하위 PENDING process 단독 공유) / F1 → B·C·D
- 마이그레이션은 격리 supervisor를 임시 DB로 한 번 띄워 적용(`REF-test-env.md` "격리 실행"), 쿼리는 생성된 Go 상수에서 SQL을 꺼내 `PREPARE`/`EXECUTE`. 픽스처 함정: `nodes.created_by`·`processes.started_by` NOT NULL

## 실시간-2 구현 메모 (2026-10-07)
| 자리 | 한 것 |
|---|---|
| `share.go` PUT(노드·process) | upsert **전·후** 단계를 요청 트랜잭션에서 구함(`mustNodePerm`·`mustProcessPerm`, `perm.go`) → 0 → 0 초과면 `attachSharedTile` → `afterShareChange` |
| `share.go` DELETE(노드·process) | `afterShareChange`만(타일 그대로 — SD7(받는 쪽 타일 추가 시점) "풀 때") |
| `shareAccess.go` `attachSharedTile` | 받는 쪽 트리 없으면 만들고 → 받는 쪽 보드 전부 잠금(`lockWorkspaces`) → 어느 보드든 이미 있으면 안 넣음(`hasFolderOf`·`hasTerminalOf`) → 기본 보드 루트 끝, 0.5×0.5 |
| `afterShareChange` | 커밋 뒤 **재계산 먼저 → 그다음 `TILES:UPDATE`**(타일을 받은 탭이 복원 조회할 때 출력 구독이 이미 있어야) |
| 영향 uid | 노드 공유 = `ListLiveProcessesUnder`(요청 트랜잭션 안) / process 공유 = 그 uid |
| **구현 중 정한 것 → 개정** | 처음엔 스크립트 노드 공유 = 타일 없음으로 했다가, **SH1(공유 단위) 개정(2026-10-07 사용자): 폴더만 공유** → `putNodeShare`가 스크립트면 400, 타일 분기 삭제 |
| 잠금 순서 | 소유자 요청이 **받는 쪽 계정의 보드 행**을 잠금. 다른 경로는 자기 계정 것만 잠가서 지금은 교착 없음 — 실시간-4(exec 타일)가 여러 계정 트리를 잠글 때 계정 id 순으로 잠글 것 |
| 아직 남음 | 받는 쪽 폴더 타일의 경로 줄은 공유 루트 부모에서 404(SD9(경로 줄 멈춤) = 실시간-5) / 프론트는 `SHARE:CHANGE`를 아직 안 받음(실시간-6b) — `TILES:UPDATE`·출력은 지금도 동작 |

## 실시간-2 확인 결과 (2026-10-07, 격리 실행 임시 DB `nexus_rt2_tmp` + 5051 + worker, Python 소켓 클라이언트 31/31 + 크기 우선권 4/4)
- 공유 전 B 출력 못 받음 → 노드 공유(보기) PUT → B `SHARE:CHANGE` + `TILES:UPDATE`(폴더 타일 F) + 출력 받음
- 단계 변경(10 → 20) = `SHARE:CHANGE`만, 타일 추가 없음, 폴더 타일 1개 / DELETE → `SHARE:CHANGE` + 출력 끊김 + 폴더 타일 남음
- process 공유 PUT → 터미널 타일 + 출력 받음 / 풀기 → 출력 끊김 + 타일 남음 / 다시 노드 공유 = 이미 있는 폴더 타일이라 추가 없음 + 출력 받음
- 차단(0) 공유 = `SHARE:CHANGE`만, 타일 없음 / 스크립트 공유 = 처음엔 타일 없음 + 출력 받음 → 폴더만 개정 뒤 재실행: **PUT 400 + `SHARE:CHANGE` 없음 + 출력 못 받음**(31/31·4/4 다시 통과)
- 크기 우선권: A가 다른 보드를 봄(`TAB:VIEW`) → 먼저 연결한 D가 우선권 → D 공유 풀기 → 재계산으로 E에게 `PROCESS:SIZE_OWNER` / D resize 404
- 스크립트(scratchpad, 재부팅 시 사라짐): `rt2/rt2_test.py`·`rt2/rt2_owner.py`. 함정: 세션 응답(`GET /users/session`)에 id 없음 → `GET /users/lookup?identification=`로 id

## 실시간-3 구현 메모 (2026-10-08)
| 자리 | 한 것 |
|---|---|
| `node.go` `patchNode` | 부모가 실제로 바뀐 이동(`reparent` = `moving` && `!sameParent`)만: 옮기기 전 `collectAround`(`ListShareUsersAround` + `ListLiveProcessesUnder`) → PatchNode → 옮긴 뒤 `mergeUsers`(후 계정 합침) → `AfterCommit`에서 `reconcileAccess`. 순서만 바꾼 이동은 건너뜀 |
| `node.go` `deleteNode` | 삭제 전 `collectAround` → `AfterCommit`에서 `reconcileAccess` |
| 정한 것 | 소유자는 users에 따로 안 넣음(단계 불변) / 삭제된 폴더의 받는 쪽 타일 정리는 안 함(실시간-5) |

## 실시간-3 확인 결과 (2026-10-08, 격리 실행 임시 DB `nexus_rt3_tmp` + 5051 + worker, Python 소켓 클라이언트 15/15)
- 공유 폴더 밖 스크립트의 process: B 출력 못 받음 → 공유 폴더 안으로 이동 = B `SHARE:CHANGE` + 출력 받음 → 같은 부모 안 순서 변경 = `SHARE:CHANGE` 없음 → 밖으로 이동 = `SHARE:CHANGE` + 출력 끊김 → 다시 안으로 → 공유 폴더 삭제 = `SHARE:CHANGE` + 출력 끊김, 소유자 A는 계속 출력 받음
- 스크립트(scratchpad `rt3/rt3_test.py`, 재부팅 시 사라짐). 함정: `GET /users/lookup`은 자기 자신 400 → 두 계정이 서로 조회해서 id를 얻음

## 실시간-4(exec 타일) 구현 메모 (2026-10-08) — SD2(실행 때 터미널 타일을 넣을 계정)
| 자리 | 한 것 |
|---|---|
| `query/nodes.sql` | `ListNodeAncestors(node_id)` — 조상 id 가까운 순(자신 제외) |
| `tile.go` `attachViewerTerminal` | 계정 한 명의 모든 보드를 잠그고(별도 트랜잭션) 이미 uid가 있으면 nil / 가장 가까운 조상 폴더 타일 아래(여러 보드면 첫 발견) / 없으면 기본 보드 루트 끝, 0.5x0.5 |
| `processApi.go` `attachViewerTiles` | 실행자 타일·`publishTiles` 뒤에 `ListProcessViewers`(단계 > 0)에서 실행자를 뺀 계정을 id 순으로 순회, 계정마다 `attachViewerTerminal` → `publishTiles(user, tree, "")`. 실패 = 로그만(kill 안 함) |
| 정한 것 | 계정마다 별도 트랜잭션이라 교착 없음(여러 계정을 한 트랜잭션에서 잠그는 일이 없음) / 고아 안전망 `adoptOrphans`는 그대로 |

## 실시간-4 확인 결과 (2026-10-08, 격리 실행 임시 DB `nexus_rt4_tmp` + 5051 + worker, 소켓 클라이언트 14/14 + 실시간-3 회귀 15/15)
- A(소유자)가 `SH > sub > S` 실행: B(보기)는 SH 폴더 타일 아래에 터미널 타일 + `TILES:UPDATE` / D(보기, 폴더 타일 지움)는 기본 보드 루트에 / C(차단)는 타일·이벤트 없음 / A는 자기 타일
- B(조작 20)가 실행: A 트리에 B의 터미널 타일 + `TILES:UPDATE` / B 트리엔 그 uid 타일이 정확히 1개 / D도 받음
- 스크립트(scratchpad `rt3/rt4_test.py`, 재부팅 시 사라짐)

## 실시간-5(노드 응답) 구현 메모 (2026-10-08) — SD8(받는 쪽 루트 목록)·SD9(경로 줄 멈춤)
| 자리 | 한 것 |
|---|---|
| `query/shares.sql` `ListSharedRoots(user_id)` | 내 `node_shares` 행이 있는 남의 FOLDER 중 `node_permission > 0` + 부모가 없거나 부모 단계 0. 노드 + 단계 + 소유자 닉네임, 닉네임·이름·id 순 |
| `node.go` `listChildren` | `parentId` 없음 = 내 루트(40) 뒤에 공유 루트를 이어서 반환 |
| `node.go` `getNode` | `parentVisible` = 부모가 있고 (소유자 40 또는 부모 단계 > 0) |
| `nodeDto.go` | `ownerNickname`(공유 루트만)·`parentVisible`(getNode만) 포인터 `omitempty` |
| 정한 것 | `parentVisible`은 `getNode` 응답만(`fetchNodePath`·DETAIL 캐시가 `getNode`로만 채움) / 프론트 `fetchNodePath` 변경은 실시간-6b / 공유 루트 목록 갱신은 `SHARE:CHANGE` 재조회(6b) |

## 실시간-5 확인 결과 (2026-10-08, 격리 실행 임시 DB `nexus_rt5_tmp` + 5051, 소켓 클라이언트 없이 REST만 12/12)
- 공유 `F`(보기)·그 하위 `G`(편집 추가 행): B 루트 = F만(permission 10, ownerNickname) / C(F 조작 + 다른 폴더 H 차단) = F만 / D = 빈 목록 / A = 자기 루트 3개 40 + ownerNickname 없음
- B `GET F` parentVisible false / `GET G`·`GET S` true / A `GET F` false(부모 없음) · `GET G` true / B의 공유 안 된 폴더 = 404
- 스크립트(scratchpad `rt3/rt5_test.py`, 재부팅 시 사라짐)
