# REF — node 타일 트리 서버 저장 (⑪)

> 2026-10-01 신설. 진행 순서 3단계(`REF-process-sync.md` "진행 순서"). 저장 대상 모델(여는 관계 트리)·순서·닫기 규칙 → `REF-node-ui-layout.md` / 프론트 스토어 → `REF-node-ui-impl.md` / 이력 → `history/node-ui-save.md` / 현재 진행 → `CURRENT.md`.

**상태: 구조·결정 확정(2026-10-01), 코드 미착수.** 사용자 지시 "vault부터 정리, 코드 수정은 아직" — 착수 승인 전 코드 금지.

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
         -> response {process, tile, version} -> store applies same tile
```

## 작업 단위 3-a~3-e
| 작업 단위 | 내용 | 주요 파일 |
|---|---|---|
| 3-a(DB) | `tile_trees(user_id PK FK users, tree JSONB NOT NULL, version BIGINT NOT NULL, updated_at)` + 조회 + **version 비교 갱신**(`UPDATE ... WHERE user_id=$1 AND version=$2 RETURNING`) + 없을 때 생성 | `migrations/00006_tile_trees.sql`, `query/tileTrees.sql` |
| 3-b(API) | `GET /tiles`(없으면 루트만 있는 트리 생성 후 반환) / `PUT /tiles {tree, version}` → 200 `{version}` / 409 `{tree, version}`. 트리 검사 | 새 `router/tile.go`(+DTO) |
| 3-b'(exec 연동, ⑪-4) | exec 요청에 `parentTileId`·`size` 추가 → 실행 성공 후 트리 행 잠금(`SELECT ... FOR UPDATE`) → 터미널 타일 추가(id = 서버 발급 UUID) → version+1. 응답에 `tile`·`version` | `processApi.go`, `router/tile.go` |
| 3-c(스토어) | 로그인 시 불러오기(불러오기 전 타일 UI는 로딩) · 변경마다 저장(짧은 간격 묶음) · 409 재적용 · exec는 PUT 없이 응답의 타일 반영 | `tileTree.store.ts`, 새 `node/api/tile.api.ts`, `FolderTileBody.vue`(exec 인자) |
| 3-d(터미널 복원) | 불러온 터미널 타일 uid를 `processTerm`에 등록(xterm 생성): 살아 있는 것 = `GET /processes`(sizeOwner 포함) / 끝난 것 = uid 목록 조회(소유 검사, 신규) | `processApi.go`, `process.api.ts`, `processTerm.store.ts` |
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
