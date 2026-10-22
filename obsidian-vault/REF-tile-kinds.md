# REF — 타일 종류 정리 (새 타일 종류를 분기 추가 없이 넣을 수 있게)

> 2026-10-22 신설. 이력 → `history/tile-kinds.md` / 현재 진행 → `CURRENT.md` / 다음 파트 → `REF-worker-dir.md`(worker 폴더 타일, 장비 타일 규칙 장비-1~3·Q9·Q10)

**상태: 정리-1~7 구현·확인 끝, 브랜치 `refactor/tile-kinds`에 커밋(2026-10-22, master 미머지). 남은 것 = 사용자 브라우저 확인 -> master `--no-ff` 머지·마감. 역할 = Claude 작성.**

## 왜 (2026-10-22 사용자)
터미널 타일만 process를 쓰다가 장비 타일이 추가됨 -> 앞으로도 타일 종류가 늘 수 있음(이미지·텍스트 보기 Q4, 장치파일 타일 Q7). 그 전에 구조를 정리. **동작 변경 없음**이 이 파트의 기준.

## 파트 3개 (2026-10-22 사용자 "네가 편한 대로" -> 제안 승인)
| 순서 | 파트 | 내용 |
|---|---|---|
| 1 | 파트(타일 종류 정리) | 이 문서 |
| 2 | 파트(worker 폴더 타일) | 폴더-1(목록 요청)·2(장비 타일)·3(여는 곳)·4(실행)·6(보드 이동) → `REF-worker-dir.md` |
| 3 | 파트(파일 전송 확장) | 폴더-8(전송 상태)·7(내려받기)·5(업로드)·9(worker 간 전송, 같은 장비 두 타일 사이 포함) → `REF-worker-dir.md` |

## 조사 (2026-10-22)
| 위치 | 종류 분기 수 | 메모 |
|---|---|---|
| 프론트 `TileFrame.vue`(580줄) | 35 | 아이콘·색·메뉴·종료·공유 표시가 한 파일에 종류별로 섞임 |
| 프론트 `tileTree.store.ts` | 9 | `folderIn`(자식 가능 검사)·kids 정렬(folder/terminal 나눔) |
| 프론트 `tileGrids.store.ts` | 6 | 터미널 목록·포커스·단축키 끔·URL 대상 |
| 프론트 나머지 16파일 | 1~5씩 | 대부분 `t.type === 'terminal' ? t.uid` (process uid 꺼내기) |
| 서버 `tileDoc.go` | 13 | 모양 검사(루트·parent = folder만), uid 발급, 찾기 함수 |
| 서버 나머지 7파일 | 1~4씩 | 타일 만들 때(`tilePlace`·`share`·`shareAccess`), 부모 검사, 터미널 uid 모으기 |
- `type === 'folder'` 검사의 실제 뜻은 대부분 **"루트 가능 / 자식 가능 / 터미널 탭 있음"**, `type === 'terminal'`은 **"process를 담음"**. 장비 타일은 앞의 세 능력을 다 가짐
- process 쪽: `processes.node_id` = 스크립트 노드(권한 `process_permission`·실행 목록·EDIT 결과·터미널 타일 `nodeId`가 모두 이 뜻에 기댐). 장비 타일 실행은 스크립트 노드가 없음 -> 컬럼 변경은 파트(worker 폴더 타일) 폴더-4(실행)에서, 이 파트는 "nodeId = 스크립트"에 기대는 곳을 한 함수로 모으기만
- 테스트: 타일 관련 Go·프론트 단위 테스트 없음

## 능력 목록 (제안, 확인 전)
| 이름 | 뜻 | folder | terminal | device(다음 파트) |
|---|---|---|---|---|
| 능력-1(루트 가능) | 보드의 루트 타일이 될 수 있음 | O | X | O |
| 능력-2(자식 가능) | 다른 타일을 열어 kids로 가짐 | O | X | O |
| 능력-3(process 담음) | `uid` = process uid, 화면·종료·최소화·단축키 끔 | X | O | X |
| 능력-4(노드 연결) | 계정 폴더 노드에 묶임(공유·선택·삭제 정리 대상) | O(null = 루트 목록) | O(스크립트) | O(장비 디렉토리) |
- 터미널 탭 = 능력-1(루트 가능) 타일이 루트일 때. 별도 능력으로 두지 않음

## 큰 계획 (제안, 확인 전)
| 단계 | 내용 | 손댈 곳 |
|---|---|---|
| 정리-1(종류 명세) | 서버 `tileKinds` 표(능력 + 종류별 필드 검사), 프론트 `tileKind` 표(능력 + 아이콘·색). 아직 호출처는 안 바꿈 | 새 파일 서버 1·프론트 1 |
| 정리-2(서버 분기 교체) | `tileDoc` 검사·찾기 함수와 7파일의 종류 비교를 능력 검사 함수로. 타일 만드는 곳은 생성 함수로 모음 | `tileDoc.go`·`tilePlace.go`·`share*.go`·`tile.go`·`workspace.go`·`nodeRuns.go`·`processPurge.go` |
| 정리-3(서버 확인) | `tileDoc` 단위 테스트 새로(정리-2 전 동작 기준으로 먼저 써서 전후 같은지) | `tileDoc_test.go` |
| 정리-4(프론트 store·util 교체) | `tileTree`·`tileGrids`·`tileOrder`·`boardTrees`·hook들의 비교를 능력 함수로(`canParent(t)`·`procUid(t)` 등) | store 4·util 1·hook 5·컴포넌트 6 |
| 정리-5(TileFrame 나누기) | 공통 틀(제목줄·크기·닫기 버튼 자리)만 남기고 종류별 제목줄 내용·메뉴를 `FolderTileHead`·`TerminalTileHead`로 | `TileFrame.vue` + 새 컴포넌트 2 |
| 정리-6(스크립트 노드 한 곳으로) | "터미널 `nodeId` = 스크립트 노드"에 기대는 곳(경로 표시·EDIT 결과)을 함수 하나 뒤로 | 프론트 2~3, 서버 1~2 |
| 정리-7(확인) | `go build`·`go test`·`vue-tsc` + 격리 실행(5051) 시험 스크립트 + 사용자 브라우저 확인 | — |

## 구현 (2026-10-22, Claude 작성)
| 단계 | 결과 |
|---|---|
| 정리-1(종류 명세) | 서버 `router/tileKind.go`(`tileKinds` 표 + `canRoot`·`canParent`·`holdsProc`·`procUID` + `newTerminalTile`·`newFolderTile`) / 프론트 `widget/util/tileKind.util.ts`(`TILE_KINDS` = 능력 + 아이콘·색) + `tileTree.store`의 `ParentTile`·`ProcTile`·`isParentTile`·`isProcTile`·`procUid` |
| 정리-2(서버 분기 교체) | `tileDoc.validate` = 종류 표로 검사(루트 = 능력-1, parent = 능력-2, 터미널 필드 = 능력-3, kids = 능력-2, 종류만의 `check`). 터미널 찾기·모으기 = 능력-3. 타일 만들기 5곳 = 생성 함수. `detachFolder` = 능력-1 검사. **폴더 전용(uid 발급 FU1·shownAt·bound·공유·고정)은 `tileFolder` 비교 유지** |
| 정리-3(서버 확인) | `tileDoc_test.go` 9개(검사 거부 17가지 포함) — 바꾸기 전 코드에서 먼저 통과시킨 뒤 바꿈 |
| 정리-4(프론트 교체) | store 4·util 1·hook 3·컴포넌트 4의 공통 로직을 능력 함수로. 닫기 규칙(closeOp)·타일 순서(flatten) = "process 타일 / 나머지"로 일반화 |
| 정리-5(TileFrame 나누기) | `TileFrame.vue` = 공통 틀(props + slot `title`·`controls`·`menu`·`tabs`·`unavailable`·`ov-extra`·`extra`, emit `close`·`minimize`). 종류 컴포넌트 `FolderTile.vue`·`TerminalTile.vue`가 채움. 고르기 = `tileComponents.ts` `TILE_COMPONENTS`(TileStrip). 보드 분리·이동 = `hook/tileBoard.hook.ts`(장비 타일 재사용), 이름 = `util/tileLabel.util.ts` |
| 정리-6(스크립트 노드) | **축소**: 터미널 `nodeId` 사용처가 `TerminalTile`·`OverviewTile`(이름)·`instanceOf`(같은 스크립트 번호) 3곳으로 모임. 서버 process `node_id` 일반화는 폴더-4(실행) 때(실제 요구가 그때 정해짐) |
| 정리-7(확인) | 아래 |
- 종류 비교 남은 곳: 프론트 13곳(전부 폴더 노드 전용·터미널 전용 컴포넌트) / 서버 폴더 전용 10곳 / SQL `type == "terminal"` 3곳(`processes.sql`·`tileTrees.sql`, `tileKind.go` 주석에 포인터)
- **함정(scoped CSS + slot)**: slot 내용은 종류 컴포넌트 scope라 틀의 `.TileFrame__head > *`가 안 닿음 -> 틀은 `> :deep(*) { flex: none }`, 줄어들 것(이름·경로·opener·실행자)은 종류 컴포넌트에서 `.TileFrame__head > .name`처럼 한 단계 높여 덮음(파일 간 CSS 순서에 기대지 않게)

## 새 종류를 더할 때 (다음 파트 장비 타일 기준)
1. `tileKind.go` `tileKinds` 줄 + 그 종류만의 `check` / 프론트 `TILE_KINDS` 줄 + `tile.type.ts` `TileType` + `tileTree.store` 타입(능력에 맞게 `ParentTile`·`ProcTile` 합집합에 넣기)
2. 종류 컴포넌트(`TileFrame` 채우기) + `TILE_COMPONENTS` 줄
3. 폴더 전용으로 남긴 것 중 그 종류에도 필요한 것(루트 탭·opener·보드 메뉴 hook)을 골라 씀

## 확인 (2026-10-22, 정리-7)
- `go build`·`go vet`·`go test`(router 9개) · `vue-tsc` · `vite build` 통과
- 격리(supervisor 5051 + 임시 DB `nexus_tk_tmp` + worker `tk-dev`) 위에 **옛 프론트(HEAD, `git archive`) 3002 / 새 프론트 3001**을 띄워 같은 헤드리스 시나리오 -> 단계마다 타일 머리 줄 글자·버튼 title/disabled·클래스·자식 계산 CSS(flex·글자 크기·opacity·굵기) 비교: **11단계 + 보드 분리·이동 6항목 모두 같음**(명세 → `REF-test-scripts-1020.md` "2026-10-22 타일 종류 정리")
- 옛 쪽에만 403 3건 = 옛 트리 `node_modules` symlink라 Vite가 아이콘 글꼴을 막은 것(환경), 앱 차이 아님
- 정리 끝: 격리 프로세스·보관 프로세스(`tk-dev`만)·소켓 폴더·임시 DB 삭제, test3 영향 없음 확인
