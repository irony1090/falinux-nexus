# REF — 터미널 타일 위치 · 폴더 타일 uid

> 2026-10-08 `REF-process-share-realtime.md`에서 분할(공유-d 마감 뒤 사용자 확인 중 나온 하위 주제). 공유-d 본문 → `REF-process-share-realtime.md` / 이력 → `history/process-share-terminal-place.md` / 현재 진행 → `CURRENT.md`

**상태(2026-10-12): 다른 보드로 가는 터미널 수정 + 폴더 타일 uid(FU1~FU4, FU2 개정 = 공유한 타일과 같은 uid + uid 우선 찾기) 코드·격리 확인·사용자 직접 확인 통과 + `openFolder` 제거, 2026-10-12 `master`에 병합(분기 지점 태그 `checkpoint/before-folder-tile-uid`).**

## 공유 전에 실행한 터미널 — 받는 쪽 빈틈 수정 (2026-10-08, 체크포인트 커밋 `3b80310`)
- 발견(사용자): 터미널을 켜 둔 채 폴더를 공유하면 받는 쪽엔 폴더 타일만 생김. 스크립트 행 "실행 중 N"은 보이지만 위치가 "위치 없음"이라 열 방법이 없음. 원인 = 터미널 타일을 넣는 곳이 exec(SD2(실행 때 터미널 타일을 넣을 계정): 그 시점에 볼 수 있는 계정)와 process 단독 공유뿐이라, 폴더 공유가 빈틈. 실시간-4(exec 타일) 확인 때 "공유 전에 실행한 터미널"은 안 봄
| 안 | 변경 | 파일 |
|---|---|---|
| A(공유 걸 때 터미널 타일도) | `putNodeShare`에서 못 보다가 보게 되는 순간(SD7(받는 쪽 타일 추가 시점)) 폴더 타일 + `attachSharedTerminals`: 폴더 아래 살아 있는 process(`liveUnder`) 중 받는 쪽이 볼 수 있고 타일이 없는 것의 터미널 타일을 기본 보드 루트 끝에. 마지막 저장본(폴더 타일 포함)을 한 번 push | `share.go`, `shareAccess.go` |
| B(내 보드에 열기) | `POST /processes/:uid/tile`(`openProcessTile`): 보기(10) 이상·살아 있는 process, 내 어느 보드에도 타일이 없으면 기본 보드 루트 끝에 넣음 → `{added}`(있으면 false, 끝난 것 400, 볼 수 없음 404). 커밋 뒤 `TILES:UPDATE` | `share.go`(라우트는 `mountShares`) |
| B 프론트 | `ScriptRunList`의 "위치 없음" 행에 "내 보드에 열기" 버튼 → `openProcessTile` → `invalidateNodeRuns`. 누르면 펼친 목록이 닫힘(행은 "보기"/"보드로 이동"으로 바뀜) | `share.api.ts`, `ScriptRunList.vue` |
- 확인(격리 실행): API 12개(공유 시 폴더 타일 1 + 터미널 타일 2 / 단계만 바꿔 다시 걸어도 중복 없음 / 이미 있으면 added=false / 받는 쪽이 타일을 닫은 뒤 added=true / 연속 호출 false / 볼 수 없는 계정 404 / 보기 단계 계정도 2개) + presence 회귀 25개 + 헤드리스(위치 없음 행 → 버튼 → 타일 3→4, xterm 1). 사용자 확인 중
- ~~안 본 것: 공유 중 폴더 이동·하위로 들어오는 경우~~ → 아래 "다른 보드로 가는 터미널"에서 확인(빈틈 있음)

## 다른 보드로 가는 터미널 (2026-10-08 사용자 발견)
- 증상: 받는 쪽이 공유 폴더 타일을 다른 보드(B2)로 옮겨 둔 상태에서 위 A·B로 터미널이 생기면 **기본 보드**에 생김
- 원인: A(`attachSharedTerminals`)·B(`openProcessTile`)가 `attachSharedTile`(= 무조건 기본 보드 루트 끝)을 씀. SD2(exec) 쪽 `attachViewerTerminal`만 "가장 가까운 조상 폴더 타일(모든 보드) → 없으면 기본 루트" 규칙
- 부수 문제: 처음 공유 때도 터미널이 폴더 타일 **아래가 아니라 루트 형제**로 붙음 → 나중에 폴더 타일을 B2로 옮겨도 W1(분리 범위 = 타일 하위)상 안 따라감

| 경로 | 지금 위치 | 판정 |
|---|---|---|
| exec 실행자(`appendTerminalTile`) | 요청 보드의 누른 타일 아래(보드 삭제 → 기본) | 정상 |
| exec 다른 계정(`attachViewerTerminal`, SD2) | 조상 폴더 타일 → 기본 루트 | 정상(기준 규칙) |
| 폴더 공유 시 실행 중 터미널(A) | 기본 루트 | **빈틈** |
| 내 보드에 열기(B) | 기본 루트 | **빈틈** |
| 공유 폴더 하위로 노드 이동(`moveNode` → `reconcileAccess`) | 타일 안 넣음 → "위치 없음" → B로 열면 기본 | **빈틈**(구독만 맞춤) |
| process 단독 공유(SD7) | 기본 루트 | 의도(받는 쪽은 폴더를 못 봄) |
| 고아 안전망(WS3)·보드 삭제(W2) | 기본 루트 | 의도 |
| 보드 간 폴더 이동(W3) | 폴더 타일 하위만 | 의도(단 A의 루트 형제 터미널은 안 따라감) |

## 수정 (2026-10-08, 브랜치 `feat/folder-tile-uid`)
| 자리 | 변경 | 파일 |
|---|---|---|
| 위치 고르기 공용화 | `insertTerminal`(계정 트리 전부 잠금 → 이미 있으면 nil → `placeTerminal` 위치에 넣음) + `visibleAncestors`(볼 수 있는 조상만, 공유가 풀려 배너만 남은 폴더 타일 아래로 안 넣게) + `placeTerminal`(가까운 조상 폴더 타일, 보드 id 순 / 없으면 기본 루트) + `folderTileOf`(루트에서 너비 우선, map 순서 비의존) | `tile.go` |
| exec(SD2) | `attachViewerTerminal` = 트랜잭션 + `insertTerminal` | `tile.go` |
| A·B·process 단독 공유 | `attachTerminal`(요청 트랜잭션 + `insertTerminal`). 폴더 공유는 폴더 타일(`attachSharedFolder`, 기본 루트 = SH2)을 **먼저** 넣고 터미널 → 폴더 타일 아래로 감 | `shareAccess.go`, `share.go` |
| 노드 이동 | 이동 전 단계(`permIndex`, 이동해 들어갈 쪽 공유 계정 포함) → 이동 뒤 "못 보다가 보게 됨"인 계정 × process에 `attachNewlyVisible`(계정 id 순 — 여러 계정 트리를 잠가 교착 방지) → 커밋 뒤 `reconcileAccess` + 보드별 push | `node.go`, `shareAccess.go` |
| 보드별 push | `savedTrees`(hash별 마지막 저장본) — 폴더 타일과 터미널이 다른 보드에 들어갈 수 있어 보드마다 `TILES:UPDATE` | `shareAccess.go` |

## 폴더 타일 uid (사용자 지시 2026-10-08)
- 논의 중 확인한 사실: 트리 JSON의 uid는 원래 터미널 타일에만(키 `uid`). 폴더 타일은 계정마다 따로 생겨 uid가 계정 간 안 맞음 → **받는 쪽 위치 찾기 열쇠는 nodeId**(위 수정이 이것). 폴더 uid는 위치 찾기에 쓰지 않음. 백엔드 `openFolder`(폴더 노드 exec → 메모리 전용 uid)는 프론트에서 안 쓰는 옛 코드(지울지 미정)
| 결정 | 확정(사용자 "같이 진행" = 추천안) |
|---|---|
| FU1(uid 발급원) | 서버 발급(`util.RandomKey(16)`, process uid와 같은 형식). `marshal`에서 uid 없는 폴더 타일에 발급 — 모든 저장이 지남 |
| FU2(공유 때 받는 쪽 uid) | **개정(사용자 "공유 때 소유자와 같아야 찾을 수 있지")**: 공유 버튼을 누른 폴더 타일의 uid를 받는 쪽 폴더 타일이 받음(공유 PUT 본문 `tileId` → 서버가 요청자 트리에서 그 타일·같은 nodeId 확인 → uid). 없거나 남의 타일 id면 새 uid. 받는 쪽 계정에 이미 같은 uid가 있으면 새 uid(계정 안 폴더 uid는 하나) |
| FU3(하위 폴더로 들어갈 때) | 유지(타일 id 기준) |
| FU4(영속) | 트리 JSON `uid`(터미널과 같은 키). `validate`가 폴더 uid 허용 |
- **찾는 순서(`placeTerminal`)**: ① 실행한 계정 트리에서 그 터미널이 붙은 폴더 타일 uid(`sourceFolderUID`, W1 = 실행 버튼을 누른 타일)와 같고 지금도 조상 폴더를 보여 주는 받는 쪽 폴더 타일 ② 가장 가까운 조상 nodeId의 폴더 타일 ③ 기본 루트. uid는 공유로 이어진 타일에만 같아서 ②는 빠질 수 없음(다른 타일에서 실행 / 받는 쪽이 새 타일로 다시 엶 / 이어진 타일이 하위로 들어감). ①이 의미 있는 경우 = 받는 쪽이 같은 폴더를 여러 타일로 열어 둘 때 "공유받은 그 타일"을 고름
- 프론트: `ShareDialog`에 `tileId` prop(`FolderTileBody`가 넘김) → `putNodeShare(..., tileId)`
- `PUT /tiles`: `keepFolderUIDs` — 프론트가 보낸 폴더 uid는 버리고 저장본 같은 id의 uid로, 새 타일은 비워서 발급(위조·복사 방지). 프론트 `FolderTile.uid?`(타입만, 쓰는 곳 없음)
- 기존 트리: 다음 저장 때 채워짐

## 확인 (격리 실행 임시 DB `nexus_fu_tmp` + 5051 + worker `fu-dev`, 삭제함)
- `fu_test.py` 28/28: FU 10(새 트리 루트 uid / 위조 무시 / 새 타일 발급 / 복사 uid 재발급 / 중복 없음 / 하위로 들어가도 유지 / 보드 분리 유지) + 다른 보드 18(B가 F 타일을 B2로 → 다시 공유 시 터미널 B2·폴더 중복 없음 / 닫고 "내 보드에 열기" → B2 / SD2 회귀 B2 / 스크립트 이동·폴더 이동 → B2, 오갔다 와도 중복 없음 / 볼 수 없는 남은 폴더 타일이 아니라 기본 루트)
- `ev_test.py` 2/2: 다시 공유·노드 이동 때 B에 B2 hash의 `TILES:UPDATE`
- 회귀: presence 25/25 · 공유 시 터미널 12/12
- FU2 개정 `fu2_test.py` 11/11: 받는 쪽 타일 uid = 공유한 타일 uid / tileId 없으면 새 uid / 남의 타일 id 무시 / **uid 우선**: 이어진 타일(B2)과 다른 F 타일(기본 보드, nodeId로는 이쪽이 먼저)이 있을 때 T에서 실행 → B2 / T2에서 실행 → 기본 보드(nodeId) / 이어진 타일이 하위로 들어가면 건너뜀 / 다시 공유 때 받는 쪽 uid 중복 없음. 회귀 다시: 28 + 2 + 25 + 12 통과
- 사용자 직접 확인(2026-10-12, 개발 서버 다시 빌드): 다른 보드(B2)로 옮긴 공유 폴더 타일 아래로 실행·"내 보드에 열기"·다시 공유·노드 이동·uid 우선 — "되는 것 같다" 통과. 헤드리스 화면 확인은 따로 안 함
- 폴더 uid ≠ 폴더 process: 백엔드 `openFolder`(`process/manager.go`, 폴더 노드 exec → 메모리 전용 레코드)는 프론트가 부른 적 없는 옛 경로. 폴더는 process로 관리하지 않고 공유 = node, 위치 = 타일 트리(uid는 타일 하나를 가리킴). **2026-10-12 제거**(아래 "openFolder 제거")

## openFolder 제거 (2026-10-12, 사용자 "앞으로도 쓸 일이 없으면 제거", 브랜치 `feat/folder-tile-uid` → `master` 병합)
- 판단: 프론트는 SCRIPT 행에서만 exec를 부름(폴더는 "새 타일" = 트리에 넣기). 폴더 공유 = node, 위치 = 타일 트리, 보는 중 = `TAB:VIEW` → 폴더 process가 필요한 곳 없음
| 자리 | 변경 |
|---|---|
| `process/manager.go` | `openFolder` 삭제. `Exec`은 SCRIPT 아니면 에러(`ExecEdit`과 같은 모양) |
| `process/entry.go` | `HasProcess` 삭제(쓰는 곳 = 폴더 조기 반환 2곳뿐), 주석의 folder-open 줄 정리 |
| `router/process.go` | `Exec`의 폴더 조기 반환 삭제 + 단계 번호 1~5로 |
| `router/processApi.go` | `execProcess`: `requireNode` 직후 SCRIPT 아니면 **400 "스크립트만 실행할 수 있습니다"** + 응답 직전 조기 반환 삭제 |
| 주석 | `perm.go` `requireProcess`·`processApi.go` `resizeProcess`의 "folder면 404" 문구 |
- `entry.Inter == nil` 검사들(input·kill·screen·status 등)은 방어 코드로 남김
- 확인(격리 임시 DB `nexus_of_tmp` + 5051 + worker `of-dev`, 삭제함, `of_test.py`): **제거 전 23/23**(폴더 exec 200·타일 없음·목록에 없음·kill 404 + 회귀 18) → **제거 후 21/21**(폴더 exec 400·목록 0 + 같은 회귀 18) + EDIT 스크립트 200·EDIT 폴더 400. 회귀 18 = FU1 uid 발급 / exec 타일 / 공유 시 같은 uid·실행 중 터미널 / 보드 B2 분리 / B2로 실행·`TILES:UPDATE` / uid 우선·nodeId 대체 / 내 보드에 열기 / 다시 공유 중복 없음 / 노드 이동
- 테스트 기대값 정정 1건: 노드 이동(R8)에서 받는 쪽이 기본 보드에도 F 타일을 열어 두면 ②(nodeId, 보드 id 순)로 기본 보드에 감 — 규칙대로(이어진 uid가 없는 G 타일에서 실행했으므로)

