# REF — node 타일 트리 서버 저장 구현 (⑪ 3-a~3-d)

> 2026-10-02 `REF-node-ui-save.md`에서 분할 — 구현·확인 기록(서버 3-a(DB)·3-b(API)·3-b'(exec 연동) + 프론트 3-c(스토어)·3-d(터미널 복원)). 설계·결정·3-c 구조안 → `REF-node-ui-save.md` / 이력 → `history/node-ui-save-impl.md` / 현재 진행 → `CURRENT.md`.

**상태: 3-a·3-b·3-b' 작성 + 확인 통과, 커밋 `2aa0920`(2026-10-02, push 안 함). 3-c(스토어)·3-d(터미널 복원) 작성 + 브라우저 확인 통과 + 사용자 PC 확인, 커밋 `2f0170f`(2026-10-02, push 안 함, 3-d는 supervisor 재시작 필요했음).**

## 3-a(DB) 구현 (2026-10-02)
| 파일 | 내용 |
|---|---|
| `migrations/00006_tile_trees.sql` | `tile_trees(user_id PK FK users ON DELETE CASCADE, tree JSONB, version BIGINT DEFAULT 1, updated_at)` |
| `query/tileTrees.sql` | `GetTileTree` / `CreateTileTree`(`ON CONFLICT DO NOTHING` — 이미 있으면 `pgx.ErrNoRows` → 다시 Get) / `UpdateTileTree`(`WHERE version=$` + `version+1`, 행 없음 = 409) / `GetTileTreeForUpdate`(3-b'(exec 연동)용 행 잠금) |
| `gen` | `superdb.TileTree{UserID, Tree []byte, Version, UpdatedAt}` — JSONB는 `[]byte`로 통과 |

- 확인: `docker exec postgres15 psql` 트랜잭션 안에서 생성·중복 생성(0행)·version 일치 갱신(version 2)·옛 version 갱신(0행)·FOR UPDATE 실행 → ROLLBACK(DB에 안 남음). 실제 적용은 supervisor 재기동 시 goose
- 주의: JSONB는 키 순서를 바꿔 저장한다(`{"rootId","tiles"}` → `{"tiles","rootId"}`). 3-b에서 Go map으로 다시 만들 때도 키가 정렬된다. **확인(2026-10-02): 영향 없음** — 화면 순서는 `kids` 배열(JSONB·Go slice 모두 순서 유지)로만 정해짐(`tileOrder.util.ts`). `tiles` 객체를 순회하는 곳은 `nodeRemove.hook.ts` `tilesUnder` 하나(삭제된 폴더를 보던 타일 목록)이고, 닫는 순서가 바뀌어도 결과가 같다(중첩 A>B를 A 먼저/B 먼저 닫아도 부모에 붙는 터미널 순서 동일 — 닫기 규칙이 직접 실행 -> adopted 순으로 다시 정렬하므로)

## 3-b(API) 구현 (2026-10-02)
새 파일 `router/tile.go` + `supervisorRouter.go`에 `mountTiles` 한 줄. `google/uuid`가 indirect -> direct(`go mod tidy`, go.mod 한 줄 이동만).

| 부분 | 내용 |
|---|---|
| `GET /tiles` | `ensureTileTree` — 없으면 `newRootDoc`(루트 id = UUID, 0.5×0.5) 생성, 생성이 행 없음이면(다른 탭이 먼저) 다시 Get. 응답 `{tree, version}` |
| `PUT /tiles` | 본문 256KB(`middleware.BodyLimit`) -> `tileDoc`로 해석 -> `validate` -> `marshal`(모르는 필드 버림, kids null -> []) -> `UpdateTileTree`. 행 없음 = 409 `{tree, version}`(`c.JSON` 직접 — 본문에 서버 트리를 실어야 해서 panic-style 아님) |
| `validate` | 루트 존재·parent null·folder / 키 = id / size 0.5·1 / folder엔 uid·adopted 없음 / terminal은 uid·nodeId 필수, kids 비어야 함 / parent는 존재하는 folder이고 그 kids에 정확히 한 번 / 루트에서 kids로 BFS해 전부 닿아야 함(고아·순환·kid->parent 불일치·중복) |

- 검사 확인: 임시 `_test.go`로 12케이스(정상 2 + 위반 10: 루트 없음·크기·키 불일치·uid 없음·kid 역참조·중복 kid·순환·터미널 부모·루트 2개·type) 통과 후 파일 삭제
- HTTP 확인 스크립트 = 세션 scratchpad `input/tiles.mjs`(계정 `input-test-b` 사용 — 실제 계정 트리 안 건드림): GET 생성·재GET 동일 / PUT 200 version+1 / 옛 version 409 / 잘못된 트리 400 / 원복
- **HTTP 확인 통과(2026-10-02, supervisor 재기동 후)**: `tiles.mjs` 5항목(GET 생성·재GET 같은 루트 / PUT 200 version 1->2 / 옛 version 409 + 서버 트리 / 크기 0.3 -> 400 / 원복 200) + `tiles-race.mjs`(새 계정으로 첫 GET 8개 동시 -> 루트 id 1개). DB: goose version 6, `input-test-b` 행 version 3. 경합 테스트 계정 `tile-race-<시각>` 1개가 DB에 남음

## 3-b'(exec 연동) 구현 (2026-10-02)
### 실패 대처 결정 (2026-10-02 사용자 확정)
타일 추가를 요청 트랜잭션에 넣으면 **커밋이 응답 뒤(tx 미들웨어)에 일어나** 핸들러가 커밋 실패를 모른다 → 프론트는 200 + 타일을 받았는데 DB엔 없음. 그래서:

| 단계 | 내용 |
|---|---|
| ① 별도 트랜잭션 | `appendTerminalTile`이 pool에서 직접 트랜잭션을 열고(`loadOrCreateTileTree` -> `GetTileTreeForUpdate` -> 추가 -> `UpdateTileTree`) 핸들러 안에서 커밋까지 |
| ② 보상 | 실패하면 `entry.Inter.Kill()` + 500("터미널 타일 저장에 실패해 실행을 중단했습니다"). kill 실패는 로그 |
| ③ 안전망(3-d) | kill도 실패하거나 실행~커밋 사이 supervisor가 죽은 경우 → 불러올 때 타일 없는 live process를 루트 `adopted`로. ⑪-4(타일 없는 실행 중 process)에서 주 수단으론 기각했던 안을 보험으로만 사용 |

- 잠금 시점 = **실행 후**(worker 실행 동안 다른 탭 PUT이 막히지 않게). 남는 빈틈은 ②③이 받음
- 컨텍스트 = `context.WithTimeout(Background, 5s)` — 요청 컨텍스트면 브라우저가 응답 전에 끊을 때 저장이 취소돼 멀쩡한 process가 kill됨
- 동시 PUT: PUT의 `UPDATE ... WHERE version=$`이 행 잠금을 기다렸다가 커밋 뒤 다시 평가 → 0행 → 409(재적용으로 합쳐짐)

### 코드
| 파일 | 내용 |
|---|---|
| `processApi.go` | `execRequest`에 `parentTileId?`·`size?`(선택 — 3-c 전 프론트 호환, 없으면 루트·0.5×0.5, 크기 검사는 실행 전 400). 응답 `execResponse` = process 응답 필드 그대로(임베드) + `tile`·`tileVersion`(process 없는 실행이면 생략) — REF 초안의 `{process, tile, version}`에서 바꿈(지금 프론트가 안 깨지게) |
| `tile.go` | `loadOrCreateTileTree`(ctx·q 받는 형태로 분리, `ensureTileTree`는 감싸기만) + `appendTerminalTile`(부모가 없거나 폴더가 아니면 루트, kids 끝에 추가, id = UUID) + `marshal`이 error 반환(`mustMarshal`은 핸들러용) |

- 확인 스크립트 = scratchpad `input/exec-tile.mjs`(irony 계정 트리 사용 후 원복): 부모 없음 -> 루트 / 부모 f1 + 크기 / 없는 부모 -> 루트 / exec 뒤 옛 version PUT 409 / 잘못된 크기 400·실행 안 됨 / 트리를 `[]`로 망가뜨려 저장 실패 -> 500 + kill
- **3-b' HTTP 확인 통과(2026-10-02, 재기동 후)**: 6항목 전부(보상 경로 포함 — 트리를 `[]`로 바꾸자 500 "터미널 타일 저장에 실패해 실행을 중단했습니다: json: cannot unmarshal array..." + 새 live process 0개). 끝난 뒤 live process 없음, irony 트리는 루트만(version 12)
- 테스트 함정: `users.identification`은 **해시로 저장**됨 — psql에서 `identification='irony'`로 찾으면 0행. 행은 `tree->>'rootId'` 등으로 찾을 것

## 3-c(스토어) 구현 (2026-10-02)
### 구조안에서 바뀐 점
| 구조안 | 구현 |
|---|---|
| `base`(서버 확정 트리)를 따로 보관 | **보관 안 함** — 화면 트리 = 확정분 + pending이 항상 성립하고, 409 땐 서버 트리를 새 기준으로 쓰므로 `base`를 읽는 곳이 없음. `version`만 보관 |
| exec 빠른 경로 조건 = version+1 · 대기 없음 | + `tileVersion <= version`이면 무시(이미 받은 서버 트리에 반영돼 있음) |
| 다른 오류 = 3초 뒤 재시도 | 3초 간격 최대 3번, 그 뒤엔 다음 변경 때 다시 보냄(400 같은 영구 오류에서 무한 재시도 방지 — `throwCatch`가 상태 코드를 버려 구분 불가) |
| `pagehide` | `visibilitychange`(hidden) 즉시 저장 — 저장 중 요청이 끝난 뒤에도 hidden이면 keepalive로 이어서 보냄 |

### 코드
| 파일 | 내용 |
|---|---|
| `feature/node/api/tile.api.ts`(신규) | `getTiles` / `putTiles(tree, version, keepalive)` — 409를 `{ok:false, tree, version}` 결과로 돌려줌(axios `validateStatus`). keepalive면 `fetch` 직접(본문 64KB 넘으면 axios로) |
| `tileTree.store.ts` | 동작 = `Op`(트리를 바꾸는 함수: `attachOp`·`navigateOp`·`closeOp`·`resizeOp`·`upsertOp`). **검사(throw) 먼저, 변경은 뒤** — 재적용 중 throw하면 그 op만 버림(재적용은 op마다 복제본에서). 상태 = `version`·`pending`·`inFlight`·`gen`(load/reset마다 증가, 이전 응답 버림). `addTerminal` 삭제 -> `applyServerTile`. `rootId`·`ready`·`loadError` 반응형 |
| `pages/index.vue` | `auth.identification` 감시로 `load`/`reset`(같은 계정 세션 재확인으로 pending을 버리지 않게). ready 전 = 로딩, 실패 = 메시지 + 다시 시도 |
| `process.api.ts` | exec 요청 `parentTileId`·`size`, 응답 = `{proc, tile?, tileVersion?}`(`ExecTile` 와이어 타입 — process가 node 스토어 타입을 import하지 않게) |
| `processTerm.store.ts` | `exec(nodeId, authKey, place)` -> 응답 그대로 반환 |
| `FolderTileBody.vue` | `start`: `applyServerTile` 후 `reveal` |
| `tileGrids.store.ts`·`nodeRemove.hook.ts` | `rootId.value` |

### 확인 (2026-10-02, 헤드리스, `type-check` 통과)
- `tiles/store.mjs`(새 계정 `tile-store-<시각>`): ① 첫 로드 version 1 ② 새 타일 -> PUT 200, 서버 2개 ③ 새로고침 후 유지 ④ 탭 2개: A PUT 200 -> B PUT 409 -> 재적용 PUT 200, 서버 4개(B 화면 4개 / A 화면 3개 = 실시간 반영은 4단계 몫) ⑤ 변경 30ms 뒤 탭 닫기 -> 서버에 반영(keepalive 경로 — 300ms 묶음 전이라 이 경로뿐)
- `tiles/exec.mjs`(새 계정 `tile-exec-<시각>`, worker `irony-MAC-ADDress1` 접속 상태): ① 대기 없음 -> exec 뒤 PUT 없음(빠른 경로) ② 이동 2번(대기 중) 직후 exec -> PUT 409 -> 재적용 PUT 200, 서버에 터미널 타일 2개 ③ 새로고침 후 터미널 타일 2개 유지(xterm 등록은 3-d)
- 테스트 계정 2개 DB에 남음. 스크립트는 scratchpad `tiles/`(재부팅 시 사라짐)

## 3-d(터미널 복원) 구현 (2026-10-02)
### 코드
| 파일 | 내용 |
|---|---|
| `query/processes.sql` | `ListProcessesByUids`(`owner_user_id = $ AND uid = ANY($::text[])`) |
| `processApi.go` | `GET /processes/lookup` -> `lookupProcesses`: `c.QueryParams()["uids"]`(반복 파라미터), 0개 = `[]`, 256개 초과 = 400, sizeOwner는 요청 탭 기준 |
| `process.api.ts` | `lookupProcesses(uids)` — `paramsSerializer: { indexes: null }` |
| `processTerm.store.ts` | `restore(fetch)`: 요청 중 출력·상태를 모으는 카운터를 exec와 공유(`fetchInFlight`) / 등록 안 된 uid만 xterm 생성 -> 안내 줄(살아 있음 = `reconnected — earlier output not shown`, 끝남 = `process exited: N`) -> `register`(모아 둔 출력은 안내 줄 뒤) / `restored` 플래그 / `reset()` / 종료 줄은 실행 중 -> 끝남 전이 때만(복원 때 쓴 줄과 중복 방지) / SIZE_OWNER는 미등록 uid도 기록하고 `register`는 `||`로 유지 |
| `tileTree.store.ts` | `adoptOrphans(procs)` -> `adoptOp`: uid로 검사, 루트 kids 끝, `adopted`, 0.5×0.5, 타일 id는 op 만들 때 |
| `hook/terminalRestore.hook.ts`(신규) | `ready && tabReady`일 때 트리 load마다 1번: lookup(타일 uid) + list(live) 병렬 -> uid 중복 제거 -> `restore` -> 살아 있는 것으로 `adoptOrphans`. 실패하면 다음 tabReady 때 재시도 |
| `pages/index.vue` | 스토어 두 개를 hook에 **인자로** 넘김 / 계정 바뀌면 `resetTerms()` |
| `TileFrame.vue` | 터미널 타일 + 상태 모름 + 복원 전 = 닫기 비활성 |
| `ProcessTerminal.vue` | `procs[uid]`가 생기면 그때 `attach` |

### 확인 중 발견한 함정 2개
| 증상 | 원인 | 처리 |
|---|---|---|
| 첫 화면 안 뜸 "TileTreeStore is not provided" | `provide`한 컴포넌트 **자신**은 그 값을 `inject` 못 함(자식만) — hook이 `index.vue`에서 `useTileTree()` 호출 | hook이 스토어를 인자로 받음 |
| 복원한 터미널 2개 빈 화면(안전망 타일만 정상) | `ProcessTerminal`은 마운트 때 1번만 attach — 복원 타일은 xterm 등록 전에 마운트됨(안전망 타일은 등록 뒤 생겨서 정상) | 등록 감시 후 attach |
| (도구) Vite가 새 파일을 빈 내용으로 캐시 | heredoc으로 만들 때 빈 순간을 읽음 -> "does not provide an export" | `touch`로 다시 읽힘 |

### 확인 (2026-10-02, supervisor 재시작 후, 헤드리스 1600×1200)
- `tiles/restore.mjs`(새 계정 `tile-restore-<시각>`·`tile-other-<시각>`): lookup 반복 파라미터·없는 uid 빠짐 / 257개 400 / 남의 계정 빈 목록 / 새로고침 뒤 터미널 3개(실행 중·끝남·안전망) / 실행 중 = 안내 줄 + 새 출력 / 끝남 = `process exited: 3`(FAIL) / 안전망 타일 서버 저장(`adopted`, 부모 루트) / 복원 뒤 kill 200 + 종료 줄 타일당 1번 — 전부 통과
- 3-c 회귀(`store.mjs`·`exec.mjs`) 같은 결과
- 화면 밖 Grid의 xterm은 `innerText`에 안 잡힘 -> `.xterm-rows` `textContent`로 읽을 것
