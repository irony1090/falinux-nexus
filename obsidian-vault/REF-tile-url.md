# REF — 보드 URL 바인딩 (새로고침 복원)

> 2026-10-19 신설. 이력 → `history/tile-url.md` / active 타일 → `REF-node-ui-active.md`(`INDEX-frontend-polish.md`) / `?focus=` 처음 생긴 곳 = 개편-8(알림 UI) → `REF-tile-binding-rework.md` / 현재 진행 → `CURRENT.md`

**상태: 마감(2026-10-19) → `INDEX-tile-url.md`. 구현 + 헤드리스 20/20 + 사용자 브라우저 확인 + 커밋.**

## 결정 (2026-10-19 사용자 "u1 나도 진행")
| 결정 | 확정 |
|---|---|
| U1(넣을 것) | 가 + 나: `?focus=<tileId>`(active 타일) + `?view=<tileId>`(보던 칼럼의 타일). Grid 번호는 화면 크기마다 달라 안 씀 — 타일 id는 어느 화면에서도 같은 타일 |
| U2(쓰는 방식) | `router.replace`(뒤로 가기 기록 안 쌓임), active·보던 칼럼이 바뀌고 300ms 뒤(스크롤이 멈춘 뒤). 전체보기 중엔 안 씀 |
| U3(복원 때 포커스) | PC(`pointer: fine`)에서만 포커스. 폰은 active 표시·스크롤만(새로고침마다 가상 키보드가 뜨지 않게) |
| U4(없는 타일) | 건너뛰고 지금 상태로 URL을 덮음 |
- 전체보기 상태는 URL에 안 넣음(새로고침하면 나옴). 보드는 원래 경로 `/w/:hash`

## 구조
| 자리 | 한 것 |
|---|---|
| `hook/tileUrl.hook.ts`(신규) `useTileUrl(ctx)` | 읽기: query 감시(`focus`·`view`·`tab` + Grid 측정 여부, `flush: 'post'`) — **view 있음 = 이 훅이 쓴 URL(새로고침) -> `restore`** / **view 없음 = 다른 화면이 보낸 이동(`focusPath`: 목록 "보드로 이동"·알림) -> 예전처럼 `activate`**(폰에서도 포커스). 쓰기: `[activeId, anchorId]` 감시 -> 300ms 뒤 `replace`, `tab`은 지움 |
| 같은 훅, 되먹임 막기 | `written`(마지막으로 쓴 값)과 query가 같으면 내 쓰기가 돌아온 것 -> 무시. `pending`(들어온 값을 아직 적용 안 함) 동안은 안 씀 — Grid 측정 전에 쓰면 복원할 값을 덮음 |
| `tileGrids.store.ts` | `columnsOf(tileId)`(reveal에서 뽑아냄) · `anchorId`(스냅 기준 왼쪽 칼럼 `view.col`을 차지하는 타일, 위쪽 먼저) · `restore(focus, view)`(view 칼럼으로 `instant` 스크롤, 없으면 active `reveal`) · `scrollTo`에 `behavior` 인자 |
| `TileWorkspace.vue` | 일회용 `?focus`·`?tab` watch(처리 뒤 query 지움)를 `useTileUrl(grids)`로 교체 |
- 보드 전환 = `TileWorkspace`가 `v-if="ready"`로 다시 마운트 -> 훅도 새로 시작(타이머는 `onBeforeUnmount`에서 지움)

## 헤드리스 확인 (2026-10-19, 격리 5051 + 프론트 3001, 20/20)
- PC 1400x900(가로 반·세로 전체 타일 8개 = Grid 4개): D1(처음 view) 첫 칼럼 타일 / D2(누르기) focus / D3(스크롤) view = 화면 왼쪽 칼럼 타일 / D8(기록 안 쌓임) / D4(새로고침) scrollLeft 4176 그대로·active·포커스·URL 그대로 / D5(focus만 이동) activate + 화면 안 + view 추가 / D6(없는 타일) URL 덮음·active 없음 / D9(`?tab`) 지워짐 / D7(전체보기) URL 안 바뀜
- 폰 400x800(`isMobile`·`hasTouch`): M1(폰 복원) active·왼쪽 칼럼 맞음, 포커스 안 함
- 봐 둘 것: 타일을 누르면 브라우저가 포커스된 타일을 화면 안으로 스크롤해서 view도 같이 바뀔 수 있음(정상) / D5는 smooth 스크롤 + 300ms 뒤에 URL이 바뀜
