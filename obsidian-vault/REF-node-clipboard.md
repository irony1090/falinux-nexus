# REF — 노드 복사·잘라내기·붙여넣기 UI (파트(노드 클립보드), 2026-10-13~)

> 이력 → `history/node-clipboard.md` / 현재 진행 → `CURRENT.md` / 원래 설계(SH10(노드 복사)·클립보드 표시) → `REF-process-share-node.md` / 서버 API → `REF-process-share-api.md`

## 출발점 (2026-10-13 확인)
- 서버: 복사 `POST /nodes/:id/copy {parentId}`(SH10) · 이동 `PATCH /nodes/:id {parentId}`(`checkMove`: 목적지 편집(30) 이상·같은 소유자·자기 아래 금지·폴더만, 루트로는 소유자만) — 둘 다 있음
- 프론트: **복사·이동 UI 모두 없음**. 행 메뉴(`NodeRowMenu.vue`) = 내용 편집·vi로 편집·이름 변경·장비 지정·삭제뿐. 클립보드 표시는 설계만 있었음
- 그래서 "잘라내기 추가"가 아니라 복사·잘라내기·붙여넣기 UI 전체가 이 파트

## 결정표
| 번호(이름) | 내용 | 상태 |
|---|---|---|
| CP1(범위) | 복사 + 잘라내기 같이. 잘라내기 = UI에서 노드를 옮기는 유일한 길 | 확정(2026-10-13) |
| CP2(표시 구분) | 잘라낸 행만 흐리게, 복사는 "복사됨 N개" 표시만(윈도우 탐색기와 같음) — 원래 설계 "복사도 흐리게"를 바꿈 | 확정(2026-10-13) |
| CP3(붙여넣기 제약) | 안 되는 위치는 붙여넣기 버튼 비활성 + 이유 툴팁. 최종 판정은 서버 | 확정(2026-10-13) |
| CP4(붙여넣은 뒤) | 복사 = 클립보드 유지 / 잘라내기 = 비움 | 확정(2026-10-13) |
| CP-부분 실패 | 노드마다 요청(일괄 API 없음) → 성공한 건 그대로, "N개 중 M개 실패" 알림 | 확정(2026-10-13) |
| CP-계정 전환 | 클립보드가 다른 계정에 보이면 안 됨 → 기존 `userPrefs.util`(`nexus.user.*`, 로그아웃 때 전부 삭제)에 저장해 해결(키를 계정별로 나누지 않음) | 확정(2026-10-13) |
| CP5(여러 개 담기) | ~~같은 모드면 덧붙임~~ → SM3(클립보드 담기 = 고른 것으로 바꾸기)로 대체(2026-10-13, `REF-node-select.md`) | 대체됨 |
| CP6(표시 위치) | 타일마다 "복사됨" 줄 대신 머리 줄에 "붙여넣기 N" 버튼 + 비우기(x), 담긴 이름은 버튼 툴팁 | 확정(2026-10-13) |

## 데이터 흐름
- 복사·잘라내기 = 프론트만: `localStorage`에 `{mode: copy|cut, nodes: [{id, name, kind, ...}]}`. `name`·`kind`는 표시용
- 붙여넣기 = `id`마다 서버 요청(복사 `POST .../copy`, 잘라내기 `PATCH parentId`). 서버가 그 시점 DB로 권한·소유자·사이클·폴더 여부를 다시 판정 → 결과는 항상 서버 현재 상태 기준

## 구조안 (2026-10-13 확정, 사용자 "권장대로 진행")
| 단계 | 파일 | 내용 |
|---|---|---|
| 클립보드-1(소유자 필드) | `nodeDto.go`·`node.api.ts` | 노드 응답에 `ownerUserId` — 같은 소유자 판정용 |
| 클립보드-2(상태) | `feature/node/store/nodeClipboard.store.ts`(신규) | `userPrefs` 저장 + `storage` 이벤트로 탭 간 맞춤 |
| 클립보드-3(메뉴·흐림) | `NodeRowMenu.vue`·`FolderTileBody.vue` | 행 메뉴 "복사"·"잘라내기", 잘라낸 행 흐리게 |
| 클립보드-4(붙여넣기) | `FolderTileBody.vue` | 머리 줄 버튼, CP3 사전 판정, 노드마다 차례로 요청, 부분 실패 알림 |
| 클립보드-5(확인) | — | 격리 헤드리스 |

CP3 사전 판정: 목적지 `permission` < 30 = 막음(공통) / 잘라내기: 목적지 경로에 잘라낸 id가 있음 = 막음, `ownerUserId` 다름 = 막음, 루트 목적지면 잘라낸 노드가 `permission` 40이어야 함

## 구현 (2026-10-13, 사용자 확인·마감)
| 단계 | 파일 | 내용 |
|---|---|---|
| 클립보드-1(소유자 필드) | `nodeDto.go`·`node.api.ts` | `nodeResponse.OwnerID`(`ownerUserId`) — `newNodeResponse`에서 채우므로 목록·상세·실시간 발행 모두 포함. `copyNode` API 함수 추가 |
| 클립보드-2(상태) | `nodeClipboard.store.ts`(신규)·`userPrefs.util.ts` | 모듈 상태 `{mode, nodes[{id, name, kind, ownerUserId, permission}]}`. 저장 키 `nexus.user.nodeClipboard`. 다른 탭 = `storage` 이벤트(`userPrefKey` 추가), 이 탭 로그아웃 = `authState` watch로 다시 읽음 |
| 클립보드-3(메뉴·흐림) | `NodeRowMenu.vue`·`FolderTileBody.vue` | 메뉴 "복사"·"잘라내기"(편집 미만이면 잘라내기 비활성), 잘라낸 행 `.row.cut` opacity 0.45 |
| 클립보드-4(붙여넣기) | `FolderTileBody.vue` | 머리 줄 `.paste`(아이콘 + 개수 + 비우기 x), 툴팁 = 담긴 이름 + 막힌 이유. `pasteBlock`이 CP3 판정. 노드마다 `await` 차례로, 끝나면 `invalidateAllForce`. 잘라내기는 성공한 id만 클립보드에서 뺌(실패는 남겨 재시도). 실패 알림 내용은 이름·메시지를 escape(AppDialog content = v-html) |

## 클립보드-5(확인) 결과 → `history/node-clipboard.md` 2026-10-13(2)

## 선택 모드 → `REF-node-select.md` (2026-10-13 분리)
- 클립보드에 영향: SM3(클립보드 담기 = 고른 것으로 바꾸기)이 CP5(덧붙임)를 대체 — `nodeClipboard.store.ts` `put`이 바꾸기로, `copy`·`cut`이 배열을 받음
- 붙여넣기 판정·요청은 `hook/nodePaste.hook.ts`로 옮김(드롭과 공용) + 판정에 "이미 이 폴더에 있습니다"(잘라내기, `ClipNode.parentId`) 추가
