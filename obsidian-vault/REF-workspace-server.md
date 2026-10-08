# REF — 작업 공간 여러 개: 서버 구현 메모 · 확인 결과 (작업공간-a~c)

> 2026-10-07 `REF-workspace-impl.md`에서 분할. 구조·결정(WS1(트리 주소)~WS7(보는 작업 공간 전달))·API·작업 단위 → `REF-workspace-impl.md` / 이력 → `history/workspace-server.md` / 현재 진행 → `CURRENT.md`.

## 작업공간-a(DB) 구현 메모 (2026-10-07)
| 자리 | 내용 |
|---|---|
| `00008_workspaces.sql` | PK `user_id` → `id BIGSERIAL`(FK `user_id`는 유지) + `hash`·`name`(기본 `''`)·`ord`(기본 0)·`is_default`. 기존 행 = 기본, hash는 `gen_random_uuid()`에서 hex 12자. `uq_tile_trees_hash` UNIQUE + `uq_tile_trees_default`(`user_id`, `WHERE is_default`) + `idx_tile_trees_user`. Down = 기본 아닌 행 삭제 뒤 원래 모양 |
| hash 형식 | 소문자 hex 12자. 기존 행은 SQL, 새 행은 Go `newTreeHash()`(`tile.go`, `crypto/rand` 6바이트) |
| 기본 대상 쿼리(이름 유지) | `GetTileTree`·`CreateTileTree`(`hash` 인자 추가, `ON CONFLICT (user_id) WHERE is_default`)·`UpdateTileTree`·`GetTileTreeForUpdate` — 모두 `AND is_default`. `/tiles` 동작은 그대로 |
| 추가 쿼리 | `GetTileTreeByHash`/`…ForUpdate`(user_id + hash, 다른 계정 hash = 행 없음) · `UpdateTileTreeByHash`(version 규칙 동일) · `ListTileTrees`(tree 제외, `ord, id` 순) · `ListTileTreesForUpdate` · `CreateWorkspaceTree`(ord = 최대 + 1) · `RenameTileTree`(version 안 올림) · `DeleteTileTree`(`:execrows`, 기본은 0행) |
| **잠금 순서** | 여러 행을 잠그는 곳(WS2(변경 방식) 이동·WS3(고아 안전망 위치))은 `ListTileTreesForUpdate`(계정 전체, id 순)만 쓴다 — 순서가 섞이면 두 트랜잭션이 서로의 행을 기다리는 데드락. 한 행만 잠그는 exec·PUT은 해당 없음 |
| Go 변경 | `tile.go`: `CreateTileTree`에 `Hash` 전달 + `newTreeHash()`만. 나머지는 작업공간-b |

### 확인 결과 (2026-10-07, 임시 DB `nexus_ws_tmp`, 확인 뒤 삭제)
- sqlc generate + `go build ./...` + vet 통과
- 00007까지 적용 → 계정 2개에 트리 행 → 00008: 기존 행 기본·hash 12자·version 보존
- 쿼리 20개 항목 통과: 기본 중복 생성 = 행 없음 / 기본 없는 계정 생성 / ord 1·2 / hash 중복 거절 / 계정당 기본 2개 거절 / ByHash 조회·다른 계정 = 행 없음 / 기본 갱신이 다른 작업 공간 안 건드림 / ByHash version 409 / 목록 순서 / 이름 변경 version 불변 / 전체 잠금 id 순 / 기본·다른 계정 삭제 0행·일반 1행 / 삭제 뒤 ord = 최대 + 1
- down → 원래 모양(PK `user_id`, 기본 아닌 행 삭제) → 다시 up 통과
- 확인 프로그램은 저장소에 파일을 두지 않고 `go build -overlay`로 `cmd/wstmp/main.go`를 가상으로 끼워 빌드(내부 패키지 import 가능)

## 작업공간-b(API) 구현 메모 (2026-10-07)
| 자리 | 내용 |
|---|---|
| `tile.go` | `/tiles/:hash` GET·PUT 추가(`requestTileTree`: hash 없음 = 기본(없으면 만듦), 다른 계정 hash = 404). PUT = 저장본을 읽어 version 비교(409) → **WS5 검사 `checkTerminalUIDs`**(저장본에 없던 uid 추가 400 + 한 트리 안 uid 중복 400) → `UpdateTileTreeByHash`(version 조건, 그 사이 끼면 409) |
| `workspace.go` (새 파일) | `GET /workspaces`(기본 없으면 만듦) · `POST /workspaces`(만들기) · `POST /workspaces/:hash/move` · `PATCH /workspaces/:hash` · `DELETE /workspaces/:hash`. 트리 조작 = `tileDoc.detachFolder`(폴더 타일 하위째 떼어 새 문서의 루트로) + `attachUnderRoot`(다른 문서 루트 끝에 붙임). 바꾼 트리는 `validate()` 뒤 저장 |
| 잠금 | 만들기 = 기본 한 행(`GetTileTreeForUpdate`) / 이동·삭제 = `lockWorkspaces`(= `ListTileTreesForUpdate`, id 순) |
| 응답 | `workspaceChangeResponse{workspace?, tree, version}` — tree = 요청한 탭이 보던 트리(만들기·이동 = 출발, 삭제 = 기본), 이름 변경은 workspace만. 409 = 출발 트리 `{tree, version}`(PUT과 같은 모양) |
| 이벤트 | `TILES:UPDATE`에 `hash` 추가(WS6(TILES 이벤트 구분)), `publishTiles(userID, rec, tabID)`로 시그니처 변경. 서버 동작 API는 **tabId를 비워** 요청한 탭도 받게 함(그 탭에도 없는 변경). 목록 변경(만들기·이름·삭제) = `WORKSPACES:UPDATE {workspaces}`(같은 `TILES:<userId>` 토픽, 커밋 전에 목록을 읽고 커밋 뒤 발행) |
| 오류 코드 | 없는 타일·작업 공간·다른 계정 = 404 / 폴더 아님·루트 폴더 타일·같은 곳으로 이동·기본 삭제·이름 64자 초과 = 400 / 출발 version 다름 = 409 |
| 이름 | 앞뒤 공백 제거, 빈 이름 허용(화면 표시 명칭은 작업공간-d 전에 정함) |

**과도기 위험(작업공간-d 전까지, 작업공간-c 뒤 대부분 해소)**: 프론트 `adoptOrphans`가 터미널 타일을 PUT으로 넣으면 이제 400(WS5). 작업공간-c부터는 `GET /tiles`가 서버 안전망을 먼저 돌려 프론트가 찾을 고아가 남지 않음(그 GET 뒤에 생긴 고아만 해당). 프론트 save는 400을 재시도 몇 번 뒤 멈추고 그 사이 다른 저장 대기분도 함께 막힘(트리 전체 PUT) → 새로고침 필요. 고아는 exec 타일 저장 실패 + kill 실패 때만 생겨 드묾. 작업공간-c WS3(고아 안전망 위치) + 작업공간-d `adoptOrphans` 제거로 해소. 또 지금 프론트는 `hash`를 안 보므로 다른 작업 공간의 `TILES:UPDATE`도 적용함 — 작업 공간을 만드는 UI가 아직 없어 실제로는 안 생김

### 확인 결과 (2026-10-07, 격리 실행: 임시 DB `nexus_ws_tmp` + 5051, 확인 뒤 삭제)
41개 항목 통과 (`ws_test.py`, 세션 scratchpad — 재부팅 시 사라짐. 터미널 타일은 exec 대신 DB에 직접 넣어 만듦):
| 묶음 | 항목 |
|---|---|
| `/tiles` | 기본 생성·`/tiles/:기본hash` 같은 트리 / 폴더 추가 PUT / 새 터미널 uid 400 / 기존 터미널 크기 변경 200 / 같은 uid 두 번 400 / 옛 version 409 + 서버 트리 / 다른 계정 hash GET·PUT 404 / 다른 트리의 uid를 PUT 400 / `PUT /tiles/:hash` |
| 만들기 | 옛 version 409 / 루트·터미널 400·없는 타일 404 / 이름 공백 제거·ord 1 / 출발 트리에서 하위째 빠짐 / 새 트리 루트 = 그 폴더(parent null) / `TILES:UPDATE` 2번(hash 구분, tabId 빈 값) + `WORKSPACES:UPDATE` 1번 / 두 번째 ord 2 / 삭제 뒤 ord = 최대 + 1 |
| 이동 | 루트 폴더·같은 곳 400 / 없는 도착·다른 계정 404 / 옛 version 409 / 출발에서 빠지고 기본 루트 끝에 하위째 / `TILES:UPDATE` 2번·목록 이벤트 없음 |
| 이름·삭제 | 이름 변경·다른 계정 404·65자 400·version 불변 / 기본 삭제 400·다른 계정 404 / 루트 폴더째 기본 루트 끝 / 이벤트 TILES 1 + 목록 / 삭제된 hash 404 / 목록 순서 |
| 전체 | 계정 전체 트리에서 터미널 uid 중복 없음 |

## 작업공간-c(exec·안전망·탭) 구현 메모 (2026-10-07)
| 자리 | 내용 |
|---|---|
| exec | `execRequest.workspace`(hash, 빈 값 = 기본). 실행 **전에** 작업 공간 확인(없으면 404, process 안 만듦). `appendTerminalTile`은 계정 트리 **전부를 잠그고**(`ListTileTreesForUpdate`) uid가 이미 어느 트리에 있으면 그 타일을 돌려줌(고아 안전망과의 경합에서 uid 중복 방지). 대상 트리가 그 사이 삭제됐으면 기본 |
| 고아 안전망(WS3(고아 안전망 위치)) | `tile.go` `adoptOrphans` — 대상 `ListOwnLiveProcesses`(소유 또는 `started_by` = 나, PENDING·PROCESS, node 있음). 먼저 잠그지 않고 `ListTileTreeDocs`로 보고, 고아가 있을 때만 전체 잠금 뒤 다시 계산 → 기본 루트 끝에 `adopted: true`, 0.5×0.5. 커밋 뒤 `TILES:UPDATE`(tabId 빈 값) + 크기 우선권 다시 맞춤 |
| **REF와 다른 점** | 안전망을 `GET /tiles`(기본)만이 아니라 **`GET /tiles/:hash`에서도** 돈다 — `/w/{hash}`로만 접속하는 탭도 있어서. 붙이는 곳은 언제나 기본 |
| 탭 등록부 | `browserTab.viewing`(hash, `""` = 기본). `TAB:VIEW {hash}`(브라우저 → 서버 EVENT, `conn.On`)로 갱신. **TAB:VIEW를 보내기 전 = 기본을 봄** → 지금 프론트(TAB:VIEW 없음, `/`만)는 동작이 전과 같음. 같은 id로 다시 붙은 탭은 이전 값 유지 |
| 크기 우선권 후보(WS7(보는 작업 공간 전달)) | `sizeOwnerCandidates`에 필터 추가: `ListTreesWithTerminal(uid)`(jsonb_path로 그 uid 터미널이 있는 트리, 계정 무관)에서 **탭의 계정 트리 중 탭이 보는 것**이 있어야 후보. 끊김 넘김·로그아웃 넘김에도 그대로 적용 |
| `resyncSizeOwner(uid)` (새로) | 비어 있으면 첫 후보가 가져감 / **연결된 소유자가 후보가 아니면**(다른 작업 공간을 봄·터미널이 옮겨짐) 첫 후보로 넘김, 후보가 없으면 비움 / 끊긴 소유자(tabGrace 대기)는 기존 규칙 / 끝난 process(소유 기록 없음 + memory에 없음)는 건드리지 않음 |
| resync 부르는 곳 | 연결 직후(`sizeOwnerOnConnect` — 예전 "빈 것 바로 가져가기"를 이것으로 대체) · `TAB:VIEW` 받을 때(그 계정이 조작 이상인 살아 있는 process 전부) · 만들기·이동·삭제 커밋 뒤(옮겨진 터미널 uid) · 고아 안전망 뒤 |
| 비용 | resync 1번 = DB 쿼리 2번(`ListProcessViewers` + `ListTreesWithTerminal`, 후자는 `tile_trees` 전체 스캔). 연결·TAB:VIEW 때 uid 수만큼 — 지금 규모에선 문제없음, 커지면 계정 트리를 한 번 읽어 uid 위치 맵으로 바꿀 것 |

### 확인 결과 (2026-10-07, 격리 실행: 임시 DB + 5051 + worker `ws-dev`, 확인 뒤 삭제)
20개 항목 통과 (`wsc_test.py`, 세션 scratchpad. 스크립트 `#!/bin/sh\ncat`):
| 묶음 | 항목 |
|---|---|
| exec | 기본 = 기본 루트 / 없는 작업 공간 = 404 + process 행 안 늘어남 / 작업 공간 + 부모 폴더 / 부모 없음 = 그 작업 공간 루트 |
| 크기 우선권 | 실행한 탭 = 소유자 / 소유 탭이 다른 작업 공간으로 가면 비움(보는 탭 없음) / 기본을 보는 탭이 생기면 가져감(`PROCESS:SIZE_OWNER` 받음) / 새 탭 연결 때 다른 작업 공간의 것은 안 가져감 / 보던 작업 공간을 떠난 소유자 → 그 작업 공간을 보는 탭으로 |
| 작업 공간 동작 | W 안 PUT으로 터미널 위치 변경 / 이동 뒤 옮겨진 것만 기본을 보는 탭으로, 남은 것은 유지 / 삭제 뒤 기본을 보는 탭으로 |
| 고아 안전망 | 트리에서 지운 살아 있는 process → `GET /tiles/:hash`(기본 아님)로도 기본 루트 끝에 adopted / `TILES:UPDATE`(기본) / 다른 작업 공간에 있는 것은 안 붙임·다시 GET해도 version 그대로 / 계정 전체 uid 중복 없음 / 공유받은 쪽(조작 20) GET = 안 붙임 |
| 끊김 넘김(62초 대기) | 소유 탭 끊김 → 기본에 있는 것은 기본을 보는 탭(t3)으로, 먼저 연결됐지만 다른 작업 공간을 보는 탭(t1)은 건너뜀 |
- 확인 못 한 것: exec와 고아 안전망이 실제로 겹치는 경합(uid가 이미 있으면 그 타일 반환) — 재현이 어려워 코드 경로만. 공유-b·c 확인 스크립트는 재부팅으로 사라져 회귀 실행 안 함(크기 우선권 후보 함수가 바뀌었으니 공유-g(확인) 때 공유 계정 탭 포함해 다시 볼 것)

## 작업공간-d 때 서버 변경 (2026-10-07)
- `tileTreeResponse`에 `hash`(GET·PUT 409·분리·이동·삭제 응답 모두) — 프론트가 기본 보드의 실제 hash를 알아 `TILES:UPDATE`를 거름
- 사용자에게 보이는 오류 문구 "작업 공간" → "보드"(화면 이름 확정). 로그·주석은 그대로

## 작업공간-e(확인) 때 서버 회귀 · 발견 (2026-10-07)
**공유-b·공유-c 회귀 59개 통과** (`reg/share_reg.py`, 세션 scratchpad. 계정 A(소유자)·B·C, 공유 행은 psql 대신 공유-c API로 걸음): 대상 조회 4 · 노드 공유 걸기·upsert·출처·재공유 403·자기 자신·단계 값 11 · node REST(목록·404·403·이동·삭제 경계) 15 · workers·exec·소유(`started_by`) 6 · 입력·구독·보기 단계 6 · 크기 우선권 2 · process 공유 4 · SB5 1 · 풀기·복사 6 · 넘김 3. 크기 우선권 넘김만 WS7 때문에 기대값이 바뀜 → 아래 "발견"

**발견 — 받는 쪽이 실행한 터미널이 소유자 기본 보드에 고아로 붙음(공유-d 착수 때 정할 것)**
| 단계 | 일어나는 일 |
|---|---|
| B(조작 공유)가 A의 공유 스크립트 실행 | SH9로 process 소유 = A, `started_by` = B. 터미널 타일은 B 트리에만(`appendTerminalTile`은 실행한 계정 트리) |
| A가 `GET /tiles`(페이지 열기·새로고침·보드 전환) | WS3 고아 안전망 대상 = "소유 또는 실행" → A 트리 어디에도 그 uid가 없으므로 **A 기본 보드 루트 끝에 `adopted: true`로 붙음** + `TILES:UPDATE` |
| 크기 우선권 | 붙기 전: B 탭이 모두 끊겨도 A 탭은 후보 아님(트리에 없음, WS7) → 소유자 비어 있음. 붙은 뒤: B 끊김 + 60초 → A 탭으로 넘어감 |
- 문제: **A가 그 사이 `GET /tiles`를 했는지에 따라** A 화면에 터미널이 생기는지, 크기 우선권이 넘어가는지가 달라짐(이미 열려 있던 A 탭엔 안 생기고 새로고침해야 생김). 안전망이 "저장 실패 복구"가 아니라 "받는 쪽 실행을 소유자에게 보여 주기" 역할을 우연히 하고 있음
- 정할 것(공유-d, SH2 경계 ①의 반대 방향): 받는 쪽이 실행한 터미널을 소유자 트리에 넣나 — 넣는다면 exec 때 소유자 트리에도(공유 폴더 타일이 있는 보드, 없으면 기본 루트) 넣어 시점을 맞추고, 안 넣는다면 WS3 대상을 "소유 + 실행 = 나" 대신 "실행 = 나, 또는 소유 = 나이고 `started_by` 없음/나"로 좁힘
