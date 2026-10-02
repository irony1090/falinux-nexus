# REF — node 타일 트리 서버 저장 (⑪)

> 2026-10-01 신설. 진행 순서 3단계(`REF-process-sync.md` "진행 순서"). 서버 구현·확인(3-a(DB)·3-b(API)·3-b'(exec 연동)) → `REF-node-ui-save-impl.md`. 저장 대상 모델(여는 관계 트리)·순서·닫기 규칙 → `REF-node-ui-layout.md` / 프론트 스토어 → `REF-node-ui-impl.md` / 이력 → `history/node-ui-save.md` / 현재 진행 → `CURRENT.md`.

**상태: 서버 쪽 3-a(DB)·3-b(API)·3-b'(exec 연동) 완료·커밋 `2aa0920`(2026-10-02) → `REF-node-ui-save-impl.md`. 3-c(스토어) 작성·확인 통과(2026-10-02, 커밋 `2f0170f` — 구조안과 달라진 점은 `REF-node-ui-save-impl.md` "3-c(스토어) 구현"). 3-d(터미널 복원)도 작성·확인 통과(같은 커밋). 다음 = 3-e(확인) — 대부분 3-c·3-d 확인에서 이미 봄.** 작업 단위마다 착수 승인 받고 진행.

## 목표와 범위
| 이 단계(3) | 다음 단계(4 O(계정 동기화) 화면) |
|---|---|
| 같은 탭 새로고침·다른 기기 로그인 시 **저장된 타일 트리를 불러옴** | 트리 변경을 다른 탭에 **실시간 push**(`TILES:<userId>` 토픽 안) |
| exec와 터미널 타일 추가를 서버에서 한 번에 | 다른 탭이 push로 받은 터미널 타일을 바로 표시 |

## 결정 ⑪-1~⑪-4 (2026-10-01 사용자 확정)
| 결정 | 내용 | 기각 |
|---|---|---|
| ⑪-1(저장 형태) | **계정당 JSON 문서 1개** — `tileTree.store`의 `{rootId, tiles}` 그대로. 닫기 규칙(⑩)을 Go로 다시 짜지 않음 | 타일마다 DB 행(연산마다 서버 로직 중복) |
| ⑪-2(충돌 처리) | **version 비교 저장**: PUT에 읽은 version을 실어 보내고 다르면 409 + 서버 트리 → 프론트가 서버 트리 위에 **아직 확정 안 된 자기 동작을 다시 적용**해 재전송 | 나중 저장이 이김(다른 탭 변경이 사라짐) |
| ⑪-3(새로고침 뒤 터미널 화면) | **빈 화면 감수** — 새 출력부터 보임. 해결은 SNAPSHOT(ring buffer) — **추후 반드시 구현**(사용자, `REF-process-snapshot.md`) | 이번에 SNAPSHOT까지 |
| ⑪-4(타일 없는 실행 중 process) | **서버가 exec 때 트리에 같이 넣기** — exec 요청에 부모 타일 id·크기를 싣고, 실행 성공 직후 서버가 그 폴더 타일 `kids` 끝에 터미널 타일을 추가하고 version을 올림. 실행과 타일 추가 사이 빈틈이 없어짐 | 불러올 때 루트에 `adopted`로 붙이기(원래 자리 잃음) / 4단계로 미루기 |

- ⑪-4가 막는 상황: 지금 흐름은 exec 성공 → 프론트가 타일 추가 → 저장의 **요청 두 번**이라, 그 사이 탭 닫기·서버 재시작·네트워크 끊김·409 재시도 실패면 process는 도는데 서버 트리에 타일이 없어 새로고침 후 보이지도 kill되지도 않는다.

## 흐름
```
load : login -> GET /tiles -> tileTree.store (version kept)
         none in DB -> server creates root-only tree, returns it (root id same for all tabs)
         terminal tiles -> processTerm: live = GET /processes, ended = lookup by uids
edit : open/navigate/close/resize -> local apply -> PUT /tiles {tree, version} (coalesced)
         409 {tree, version} -> take server tree -> replay pending ops -> PUT again
exec : POST /processes/exec {nodeId, authKey, parentTileId, size}
         -> worker exec ok -> server appends terminal tile (tx, version+1)
         -> response {...process fields, tile, tileVersion} -> store applies same tile
```

## 작업 단위 3-a~3-e
| 작업 단위 | 내용 | 주요 파일 |
|---|---|---|
| 3-a(DB) | `tile_trees(user_id PK FK users, tree JSONB NOT NULL, version BIGINT NOT NULL, updated_at)` + 조회 + **version 비교 갱신**(`UPDATE ... WHERE user_id=$1 AND version=$2 RETURNING`) + 없을 때 생성 | `migrations/00006_tile_trees.sql`, `query/tileTrees.sql` |
| 3-b(API) | `GET /tiles`(없으면 루트만 있는 트리 생성 후 반환) / `PUT /tiles {tree, version}` → 200 `{version}` / 409 `{tree, version}`. 트리 검사 | 새 `router/tile.go`(+DTO) |
| 3-b'(exec 연동, ⑪-4) | exec 요청에 `parentTileId`·`size` 추가 → 실행 성공 후 트리 행 잠금(`SELECT ... FOR UPDATE`) → 터미널 타일 추가(id = 서버 발급 UUID) → version+1. 응답에 `tile`·`version` | `processApi.go`, `router/tile.go` |
| 3-c(스토어) | 로그인 시 불러오기(불러오기 전 타일 UI는 로딩) · 변경마다 저장(짧은 간격 묶음) · 409 재적용 · exec는 PUT 없이 응답의 타일 반영 | `tileTree.store.ts`, 새 `node/api/tile.api.ts`, `FolderTileBody.vue`(exec 인자) |
| 3-d(터미널 복원) | 불러온 터미널 타일 uid를 `processTerm`에 등록(xterm 생성): 살아 있는 것 = `GET /processes`(sizeOwner 포함) / 끝난 것 = uid 목록 조회(소유 검사, 신규) + **안전망: 살아 있는데 트리에 타일이 없는 process는 루트에 `adopted` 터미널로 붙이고 저장**(3-b' 결정 ③) | `processApi.go`, `process.api.ts`, `processTerm.store.ts` |
| 3-e(확인) | 새로고침 후 타일 유지 · 탭 2개 동시 변경 409 재적용 · exec 타일이 서버 트리에 있음 · 부모 타일이 닫힌 뒤 exec | 헤드리스 + Node 클라이언트 |

## 세부 규칙 (구조 단계에서 정한 것)
| 항목 | 규칙 |
|---|---|
| 문서 모양 | `{rootId, tiles: {[id]: Tile}}` — `Tile` = 지금 스토어 타입(`type`·`parent`·`kids`·`size`·`nodeId`·`uid`·`adopted`). version은 문서 밖(컬럼) |
| 서버 검사 | 루트 1개(parent null, folder) · 모든 parent/kids 쌍 일치 · 고아·순환 없음 · type folder/terminal · size 각 0.5/1.0 · terminal은 uid·nodeId 필수. 본문 크기 상한(예: 256KB). uid 소유는 검사 안 함(타일은 참조일 뿐, process API가 소유 검사) |
| exec 부모 없음 | `parentTileId`가 트리에 없으면(다른 탭이 닫음) **루트 kids 끝**에 붙임 |
| 재적용 대상 | 프론트는 마지막 확정 version 이후의 자기 동작(함수)을 큐로 보관 → 409 때 서버 트리에 순서대로 다시 실행. 대상 타일이 없어진 동작은 버림. 재시도 상한(3회) 넘으면 서버 트리로 다시 불러오고 안내 |
| 저장 묶음 | 연속 변경(크기 조절 등)은 짧게 모아 PUT 1번. 요청 중 생긴 변경은 응답 뒤 다음 PUT |
| exec와 대기 중 저장 | exec 응답의 version이 로컬보다 앞서면, 대기 중 PUT은 옛 version이라 409 → 재적용으로 자연히 합쳐짐 |

## 이 단계에서 안 하는 것
- 다른 탭 실시간 반영(4단계) · SNAPSHOT(추후 필수) · 공유 트리(5단계).
- 끝난 터미널 타일 자동 정리(사용자가 닫음 — 기존 규칙 유지).

## 3-c(스토어) 구조안 (2026-10-02 확정·구현 — `base`는 구현에서 뺌)
### 지금 스토어와 바뀌는 점
| 지금 | 3-c 뒤 |
|---|---|
| `provideTileTree()`가 루트를 직접 만들고 `rootId`는 상수 | 로그인 뒤 `GET /tiles`로 받음. 받기 전엔 `ready=false`(타일 UI 대신 로딩). `rootId`는 반응형으로 바뀜 -> 소비처 `tileGrids.store`·`nodeRemove.hook` 수정 |
| 동작(`openFolder`·`navigate`·`close`·`resize`)이 `tiles_`를 바로 바꿈 | 동작을 **트리를 바꾸는 함수(op)** 로 만들어 `tiles_`에 적용하고 `pending`에 쌓음 -> 저장 예약 |
| exec 성공 -> `addTerminal`(프론트가 타일 생성) | exec 요청에 `parentTileId`·`size`를 싣고, 응답의 `tile`·`tileVersion`을 반영(`addTerminal` 대신 `applyServerTile`) |

### 상태와 저장 흐름
```
state : base (last tree the server confirmed) + version
        pending: Op[]   (my ops not yet confirmed)
        tiles_ = base + pending applied   (what the screen shows)

op    : apply to tiles_ -> push to pending -> schedule save (300ms, one PUT in flight)
save  : PUT {tree: tiles_, version}; remember sent = pending.length
        200 -> base = sent tree, version = res.version, pending.splice(0, sent)
        409 -> base = server tree, version = server version
               tiles_ = clone(base), replay pending (op throws on missing tile -> drop it)
               retry PUT (max 3, then reload server tree + notice)
        other error -> keep pending, retry later (3s)
exec  : res.tileVersion == version + 1 and nothing pending/in flight
          -> fast path: add tile to base and tiles_, version = tileVersion (no PUT)
        else -> push idempotent op "upsert this tile" -> next PUT gets 409 -> replay skips it (already in server tree)
```

| 항목 | 규칙 |
|---|---|
| op 안의 id | 새 타일 id(`crypto.randomUUID`)는 **op를 만들 때** 정한다(적용할 때 X) — 재적용해도 같은 id라 화면·`reveal`이 안 흔들림 |
| 재적용 실패 | 대상 타일이 서버 트리에 없으면(다른 탭이 닫음) 그 op만 버림 |
| 닫기 직전 저장 (2026-10-02 확정) | `visibilitychange`로 `hidden`이 되면 300ms 묶음을 기다리지 않고 **바로 저장**, 이때만 `fetch(..., {keepalive: true})` 직접 사용(axios 미지원, 본문 64KB 이하일 때만). 응답 처리(200/409)는 평소 경로 그대로 — 페이지가 살아 돌아오면 정상 처리, 사라졌으면 409분은 유실 감수. `pagehide`는 모바일에서 OS가 탭을 죽이면 안 오므로 기각 |
| 로그아웃 | `auth` null -> `ready=false`, base·pending 비움 |
| 터미널 타일 | 불러온 터미널 타일의 xterm 등록은 3-d(터미널 복원) 몫. 3-c에선 등록 안 된 uid 타일이 빈 화면이어도 깨지지 않게만 |

### 파일 매핑
| 파일 | 변경 |
|---|---|
| `feature/node/api/tile.api.ts`(신규) | `getTiles()` / `putTiles(tree, version)`(409를 에러가 아닌 결과로 구분해 돌려줌 — `api.util`의 `throwCatch`가 상태 코드를 버리므로 직접 처리) |
| `feature/node/store/tileTree.store.ts` | base·version·pending·ready + op 큐 + 저장 묶음 + 409 재적용 + `applyServerTile` + `load`/`reset` |
| `pages/index.vue` | `auth` 감시로 `load`/`reset`, `ready` 전엔 로딩 표시 |
| `feature/node/store/tileGrids.store.ts`, `hook/nodeRemove.hook.ts` | `rootId` 반응형 대응 |
| `feature/process/api/process.api.ts`, `store/processTerm.store.ts` | exec 요청에 `parentTileId`·`size`, 응답 타입에 `tile`·`tileVersion`, `exec`가 이것도 반환 |
| `feature/node/component/tile/FolderTileBody.vue` | `start`: `addTerminal` 대신 `applyServerTile` 후 `reveal(tile.id)` |

## 3-d(터미널 복원) 구조안 (2026-10-02 확정·구현 — 3-d①~④ 전부, 구현 기록 `REF-node-ui-save-impl.md`)
### 흐름
```
trigger : tileTree.ready && tabReady  (tab id first -> sizeOwner correct; reload within 60s keeps same tab id)
order   : tree loaded FIRST, then processes  (exec in another tab in between -> orphan op -> 409 -> replay skips by uid)
fetch   : GET /processes (live, as is) + GET /processes/lookup?uids=a&uids=b (listed only, any status, owner checked)
apply   : processTerm.restore(list)  -> register xterm; ended -> write "[process exited: N]" line
orphan  : live but no tile with that uid -> tileTree.adoptOrphans -> root kids end, adopted, 0.5x0.5 (PUT)
```

### 결정 후보
| 결정 | 제안 | 대안 |
|---|---|---|
| 3-d①(끝난 process 조회) **확정 2026-10-02** | **나누기**: 살아 있는 것 = `GET /processes` 그대로 / 타일 uid = 새 `GET /processes/lookup?uids=a&uids=b`(요청한 uid만, 상태 무관, 소유 검사). 쿼리 파라미터 반복 형식(쉼표 구분 X) — Echo `QueryParams()["uids"]`, axios는 `paramsSerializer: { indexes: null }`(기본은 `uids[]=`) | 기각: `GET /processes?uids=`로 묶기(파라미터가 결과를 넓히는 "더하기"라 의미 혼동, 요청 사이 변화는 상태 이벤트·uid 검사로 이미 흡수) / 끝난 건 조회 안 함 |
| 3-d②(복원 전 닫기) **확정** | 복원이 끝나기 전엔 터미널 타일 닫기 비활성 — 지금은 상태를 모르면 "실행 중 아님"으로 보여 실행 중 타일이 닫힘 | 막지 않음 |
| 3-d③(복원 안내 줄) **확정** | 살아 있는 터미널에 흐린 한 줄 `[reconnected — earlier output not shown]` — 빈 화면이 고장처럼 보이지 않게(⑪-3) | 안내 없음 |
| 3-d④(계정 바뀜) **확정** | 로그아웃·계정 전환 때 `processTerm`도 비움(xterm dispose) — 지금은 이전 계정 xterm이 메모리에 남음 | 이번 범위 밖 |

### 세부 규칙
| 항목 | 규칙 |
|---|---|
| 안전망 op | `adoptOrphans`는 **uid로** 있는지 검사(타일 id 아님) — 재적용 때 다른 탭이 넣은 같은 uid 타일이 있으면 건너뜀. nodeId 없는 process는 건너뜀(터미널 타일은 nodeId 필수) |
| 크기 | 안전망 타일 = 0.5×0.5 (어느 화면에서든 최소 단위, 서버 exec 기본값과 같음). `newSize`는 TileWorkspace 안에서만 있어 못 씀 |
| 크기 우선권 이벤트 | `PROCESS:SIZE_OWNER`가 등록 전에 오면 지금은 버림 -> 모르는 uid도 `owners_`에 기록 |
| 다시 복원 | 같은 페이지에서 소켓 재연결은 기존 `syncOwners`로 충분, 복원은 트리 load마다 1번 |
| 위치 | 조율은 새 `feature/node/hook/terminalRestore.hook.ts`(index.vue에서 호출) — process 스토어가 node 스토어를 모르게 유지 |

### 파일 매핑
| 파일 | 변경 |
|---|---|
| `query/processes.sql` + sqlc | `ListProcessesByUids` (`owner_user_id=$1 AND uid = ANY($2::text[])`) |
| `processApi.go` | 새 `lookupProcesses`(`GET /processes/lookup`, uids 반복 파라미터, 개수 상한) — `listProcesses`는 그대로 |
| `process.api.ts` | `lookupProcesses(uids)` |
| `processTerm.store.ts` | `restore(list)` + `reset()` + SIZE_OWNER 미등록 uid 기록 |
| `tileTree.store.ts` | `adoptOrphans(procs)` op |
| `terminalRestore.hook.ts`(신규) · `pages/index.vue` | 트리거·순서·`restoring` 상태 |
| `TileFrame.vue` | `restoring` 중 터미널 닫기 비활성 |
