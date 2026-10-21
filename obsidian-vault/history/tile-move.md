# history/tile-move — 타일 이동 규칙 통일 · 커서 위치 페이지 표시

> 재사용 지식 → `REF-tile-move.md` / 확인 스크립트 → `REF-test-scripts-1020.md` "2026-10-21 페이지 이동" / 현재 진행 → `CURRENT.md`

## 2026-10-21
- 사용자: Ctrl+Alt+숫자가 무조건 페이지 단위로 이동 -> 인디케이터의 타일 이동 규칙과 같게. "타일 이동 시에는 기본적으로 페이지네이션 규칙을 따른다"
- 원인 확인: `reveal` -> `goGrid`(Grid 시작) vs 인디케이터 `stepTo` -> `revealCol`. 이동 규칙이 두 군데로 갈라져 있었음
- 사용자 추가 요청: Ctrl+Alt를 누르면 커서 위치(가로·세로 가운데)에 인디케이터, 같은 조작
- 결정(사용자): 뜨는 때 = 누르자마자 / 닫힘 = 떼면(드래그 중 유지) / 커서 모르면 화면 가운데 / 타일 선택 = 위·아래 둘 다 / 폰·전체보기 안 띄움
- 구현: 이동-1(타일 위치 규칙 통일) `tileGrids.store.ts` `reveal`·`leaveOverview(tileId)`·`activate` / 이동-2(커서 위치 페이지 표시) `PagerPopup.vue` 신설 + `TileNav`에 마운트, `pagerDrag.hook.ts` `pull`·`pullDir`, `GridPager` `pull`·expose, `TilePeek` `below`
- `vue-tsc` 통과 + 격리 헤드리스 `move/move.mjs` 28/28(이동-1 7개 · 이동-2 21개) + 스크린샷(팝업·아래 패널) 확인
- 전체보기에서 고른 타일도 같은 규칙으로 바꿈(사용자 따로 답 없음, 원칙으로 적용) — 확인 때 물을 것
- 2026-10-21(2) 확인용 스크립트 `ctrl-alt-test`(225)를 irony `test`(14)에 등록(로컬 PTY 확인 후 REST)
- 2026-10-21(3) 사용자 확인: `ctrl-alt-test`로 이동-1·이동-2 정상. 추가 요청 -> 이동-3(타일별 단축키 끔) 계획(끔-1~끔-5, 저장 = 타일 트리, 끄는 것 = 앱 단축키 전체, Ctrl+C·V 복사·붙여넣기는 그대로) 승인
- 2026-10-21(4) 이동-3 구현: `tile.go` `ShortcutsOff`(+폴더 타일 금지 검증) / `tileTree.store` `shortcutsOffOp`·`setShortcutsOff` / `tileGrids.store` `appShortcutsOff` / `tileShortcut.hook`·`PagerPopup` 맨 앞 검사 / `TileFrame` ⋮ 메뉴 토글·아이콘. `go build`·`vet`·`vue-tsc` 통과, 격리(supervisor+worker+3001) `keys.mjs` 17/17 + `move.mjs` 28/28
- 실수: `go build ./cmd/supervisor`를 `apps/core`에서 돌려 `apps/core/supervisor`(gitignore)를 덮어쓰고 지움 — 원래 파일이 있었는지 모름. 이후 빌드는 `-o <scratchpad>`로
