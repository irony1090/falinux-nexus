# REF — node 타일 트리 서버 저장 구현 (⑪ 3-a~3-b')

> 2026-10-02 `REF-node-ui-save.md`에서 분할 — 서버 쪽 구현·확인 기록(3-a(DB)·3-b(API)·3-b'(exec 연동)). 설계·결정·3-c 구조안 → `REF-node-ui-save.md` / 이력 → `history/node-ui-save-impl.md` / 현재 진행 → `CURRENT.md`.

**상태: 3-a·3-b·3-b' 작성 + 확인 통과, 커밋 `2aa0920`(2026-10-02, push 안 함).**

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
