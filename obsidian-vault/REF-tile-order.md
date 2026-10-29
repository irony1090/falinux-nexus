# REF — 타일 순서 구조 (파트(타일 끌기) 구조-1~3)

> 2026-10-29 `REF-tile-drag.md`에서 분할. 파트 전체 계획·결정 → `REF-tile-drag.md` / 이력 → `history/tile-order.md` / `CURRENT.md`

## 구조-1(순서 배열) 상세 (확정·구현 2026-10-29)
- 문서: `{rootId, order: [id...], tiles: {id: tile}}`. 타일에서 `kids`·`adopted` 삭제, `parent` = 연 곳(루트만 null). Go `tileDoc.Order` / 프론트 `TileTree.order`
- 검사(`validate`): ① 루트 = parent null·능력-1(루트 가능)·`order[0]` ② `order` = 모든 타일 정확히 한 번(모르는 id·빠짐·중복 없음) ③ 루트 말고는 parent 있음 + 그 타일 존재 + 능력-2(자식 가능) ④ parent 사슬이 순환 없이 루트에 닿음 ⑤ 기존 종류·크기·필드 검사 그대로
- 옛 문서 변환 = 변환-B(시작 때 한 번, 추천): supervisor 시작 때 마이그레이션 직후 `order` 없는 `tile_trees`를 옛 `flatten` 규칙(Go로 옮김)으로 변환·저장·version+1. 멱등. 응답·push가 저장본 그대로(`rec.Tree`)라 변환기를 한 곳에만 두려면 이게 맞음. goose Go 마이그레이션은 store 패키지가 router의 tileDoc을 못 불러서(순환 import) 안 씀. 기각 = 변환-A(읽을 때마다 변환: 프론트에도 변환기 필요)
- 구조-1~3은 한 묶음으로 넣고 확인(검사가 바뀌면 프론트 store도 같이 바뀌어야 저장됨). 배포 = 프론트·supervisor 같이, 열린 옛 탭은 새로고침

## 구조-1~3 구현 (2026-10-29)
| 곳 | 바뀐 것 |
|---|---|
| `tileDoc.go` | `Order` 추가, `Kids`·`Adopted` 삭제, validate 개정(①~⑤), `openTile`(연 곳 바로 뒤, 없으면 루트 바로 뒤)·`remove`(연 곳 위로)·`subtree`, `folderTiles` = order 순 |
| `tileKind.go` | 새 타일 생성 함수에서 parent·kids 인자 빼기(자리는 openTile) |
| 서버 넣기 | `tilePlace`(exec·insertTerminal·고아)·`share`(바인딩)·`shareAccess`(공유 받음)·`videoTile` = `openTile` / `processPurge` = `remove` / `viewerList` = order 순 / `workspace` 분리·붙이기 = subtree·루트 바로 뒤 |
| 변환-B | `tileOrderMigrate.go` `MigrateTileOrder`(main `init`에서 마이그레이션 바로 뒤) + 쿼리 `ListTileTreesWithoutOrder` |
| 프론트 | `TileTree.order` · `attachOp`(연 곳 바로 뒤)·`closeOp`(연 곳 위로) · `ordered()`가 `flatten` 대체 · `tileGrids` 배치 = order · `TileNode`에서 `kids`·`adopted` 삭제 · 보기 타일 중복 = `parent === 그 타일` |
- 격리 확인(복사 DB `nexus_td_tmp` + 5051 + worker `td-dev` + 프론트 3001): 변환 test3 복사 13문서/53타일 65/65(옛 flatten과 순서 일치·version+1·kids 없음·parent 그대로) · 개발 DB 복사 10문서 50/50 · 재시작 멱등 · PUT 검사 9/9 · 화면(실행·재실행·영상·닫기·새로고침·터미널 탭) 14/14 · 서버 경로(바인딩·공유·분리·보드 삭제·폴더 이동·고아) 19/19 · Go 테스트 통과 · 타입 검사 통과
- 실데이터 두 DB에 `adopted` 타일 없음 -> adopted 변환은 단위 테스트(`TestLegacyOrder`)로만 확인
- **번호 = 번호-A(화면 순서) 확정(사용자 10-29)** — 그대로 둠(기각 = 번호-B(실행 순서)). 발견 경위: 같은 노드 여러 번 실행 번호 `(1)`·`(2)`(`tileGrids instanceOf`)가 화면 순서 기준이라, 최신이 앞에 오면서 **기존 터미널 번호가 바뀜**(먼저 실행한 것이 (1) -> (2))

