# history/process-share-terminal-place — 터미널 타일 위치 · 폴더 타일 uid 이력

> 2026-10-08 `history/process-share-realtime.md`에서 분할. 재사용 지식 → `REF-process-share-terminal-place.md` / 현재 진행 → `CURRENT.md`

## 2026-10-08(23) 공유 전에 실행한 터미널을 받는 쪽이 못 여는 빈틈 수정 (사용자 발견)
- 증상: 터미널을 켜 둔 채 폴더 공유를 걸면 받는 쪽엔 폴더 타일만 생기고 터미널 타일은 없음. 스크립트 행 "실행 중 N"은 보이는데 위치가 "위치 없음"이라 볼 방법이 없음(SD2는 exec 시점의 볼 수 있는 계정에만 넣어서 생긴 빈틈. 같은 날 앞서 "의도된 동작"이라 한 판단은 틀렸음)
- 수정 A: `putNodeShare`가 못 보다가 보게 될 때 `attachSharedTerminals`로 그 폴더 아래 살아 있는 process(볼 수 있는 것, 타일 없는 것)의 터미널 타일도 기본 보드 끝에 넣음(마지막 저장본에 폴더 타일까지 포함, 한 번만 push)
- 수정 B: `POST /processes/:uid/tile`(`openProcessTile`) 신설 — 보기(10) 이상·살아 있는 process, 내 어느 보드에도 없으면 기본 보드 끝에 넣음(있으면 `added=false`, 끝난 것 400, 볼 수 없음 404). `ScriptRunList`의 "위치 없음" 행에 "내 보드에 열기" 버튼
- 확인(격리 실행): API 12개(공유 시 폴더 타일 1 + 터미널 타일 2 · 단계 변경해도 중복 없음 · 이미 있으면 added=false · 타일 닫은 뒤 added=true · 연속 호출 false · 볼 수 없는 계정 404 · 보기 단계 계정도 2개) + presence 회귀 25개 + 헤드리스(위치 없음 행 -> 버튼 -> 타일 3→4, xterm 1). 커밋 전
- 안 본 것: 공유 중 폴더 이동·하위로 들어오는 경우(`reconcileAccess`)의 같은 빈틈

## 2026-10-08(24) 다른 보드로 가는 터미널 조사 (사용자 발견, 코드 없음)
- 사용자: 공유 폴더 타일을 다른 보드로 옮겨 둔 상태에서 (23)의 상황이면 터미널이 기본 보드에 열림 + 이런 경로가 더 있나 조사 요청
- 결과: A(`attachSharedTerminals`)·B(`openProcessTile`)·공유 폴더 하위로 노드 이동 3곳이 빈틈(SD2의 조상 폴더 타일 규칙을 안 씀). 표 → `REF-process-share-realtime.md` "다른 보드로 가는 터미널". 고칠 방향 제안, 사용자 결정 대기
- 이어서(같은 날): 사용자 "폴더 타일에 processUid를 넣으면 찾을 수 있지 않나" 논의 → 확인한 사실: 트리 JSON의 uid는 터미널 타일에만(키 `uid`), 폴더 타일은 계정마다 따로 생겨 uid가 계정 간 안 맞음 → 받는 쪽 위치 열쇠는 nodeId. 백엔드 `openFolder`(폴더 노드 exec → 메모리 전용 uid)는 미사용 옛 코드. 개발 DB 테스트1(65)의 `aAyf…` 터미널이 기본 보드, 폴더 14 타일은 `818d…` 보드 = 증상 실데이터 확인
- 사용자 지시 "폴더 타일에도 uid" → FU1(uid 발급원)~FU4(영속) 제시, 목적·결정 대기. 사용자 요청으로 vault 정리 + `master` 체크포인트 커밋 + 새 브랜치 `feat/folder-tile-uid`(위험 작업 분기)

## 2026-10-08(25) 다른 보드로 가는 터미널 수정 + 폴더 타일 uid (브랜치 `feat/folder-tile-uid`)
- 사용자 "같이 진행해줘" → FU1~FU4 추천안으로 확정, ③ 수정과 함께 코드 작성
- 서버: `insertTerminal`·`visibleAncestors`·`placeTerminal`·`folderTileOf`(tile.go) / `attachTerminal`·`attachSharedFolder`·`savedTrees`·`permIndex`·`attachNewlyVisible`·`publishSaved`(shareAccess.go) / share.go 세 핸들러 / node.go 이동. 폴더 uid = `marshal` 발급 + `keepFolderUIDs` + `validate` 허용. 프론트 = `FolderTile.uid?` 타입만
- 확인: 격리 실행 28 + 이벤트 2 + 회귀 37 통과, 헤드리스·사용자 확인 전. 커밋 전

## 2026-10-08(26) FU2 개정 — 공유한 타일과 같은 uid + uid 우선 찾기
- 사용자 "폴더 타일 uid는 따로 고르지 않았다니?" → FU1~FU4를 추천안으로 정했다고 설명(FU1은 ⓐ `openFolder` process가 아니라 ⓑ 트리 발급임을 짚음) → 사용자 "공유 때 소유자와 같아야 찾을 수 있는 거 아냐?" → 맞음(새 uid면 uid로 못 찾음), 같은 값으로 할 때 걸리는 점(소유자 타일 여러 개 → 공유 버튼 누른 타일, 다른 타일·새 타일·하위 폴더 = nodeId 필요) 제시 → 사용자 "그렇게 진행"
- 코드: 공유 PUT `tileId` + `sharedFromUID` + `attachSharedFolder(uid)`(받는 쪽 중복 시 새 발급) + `sourceFolderUID`·`linkedFolderTile`·`placeTerminal` ① 단계 + 프론트 `ShareDialog`·`FolderTileBody`·`share.api`. 격리 11 + 회귀 67 통과, 커밋 전

## 2026-10-12(27) 사용자 직접 확인 통과 + 폴더 process 정리 (코드 없음)
- 사용자가 postgres15·supervisor·worker 재시작(supervisor = `go run`, 브랜치 변경 포함 빌드) 후 확인 6개(B2로 옮긴 F 타일 아래 실행 / 내 보드에 열기 / 공유 해제 후 다시 공유 중복 없음 / 스크립트를 F 안으로 이동 / 같은 F를 두 보드에 열었을 때 uid 우선) → "되는 것 같다" 통과
- 논의: 사용자 "디렉토리 process pUid(메모리) 쓰는 작업이 끝났나" → 확인한 사실: 폴더를 process로 서버가 관리한 적 없음. 백엔드 `openFolder`는 2026-07 배선 때 만든 메모리 전용 경로(DB·worker·제거 경로 없음)이고 프론트는 폴더를 열 때 서버 exec를 안 부름. 2026-10-01 `REF-process-sync.md`의 "폴더 process 영속" 계획은 2026-10-06 공유 설계(node 공유 + 타일 트리)로 폐기. 이번 폴더 uid는 타일 트리 안 타일 식별자(FU1~FU4)
- 남은 것: 브랜치 커밋 → `master` 합칠지(사용자 결정) / `openFolder` 제거 여부(사용자 결정)

## 2026-10-12(28) openFolder 제거 + 제거 전·후 격리 확인 (브랜치 `feat/folder-tile-uid`, 커밋 전)
- 사용자 "앞으로도 쓸 일이 없으면 제거, 제거 전에 직접 테스트" → 프론트 exec 호출처가 SCRIPT 행뿐임을 확인 후 제거(manager `openFolder`·entry `HasProcess`·router 조기 반환 2곳, `execProcess` 폴더 400)
- `of_test.py`(scratchpad, 재부팅 시 사라짐): 제거 전 23/23 → 제거 후 21/21 + EDIT 확인. 첫 실행에서 R8 실패 1건은 테스트 기대값 오류(nodeId 규칙상 기본 보드가 맞음)로 정정

## 2026-10-12(29) 브랜치 커밋 + 분기 지점 태그 + `master` 병합
- 사용자 "중요 분기점이라 태그를 남기고 합치자" → 태그 위치 = 분기 지점(합치기 전 `master` `d50f153`) `checkpoint/before-folder-tile-uid`, 합치기 = `--no-ff`(merge 커밋 남김) — 둘 다 사용자 선택(추천안). 브랜치에 코드(`feat`)·vault(`docs`) 커밋 2개. push 안 함

