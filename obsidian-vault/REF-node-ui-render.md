# REF — node 카탈로그 타일 UI 렌더 · 헤더 내비 (컴포넌트 층)

> `REF-node-ui-impl.md`에서 10k자 기준 분할(2026-09-29) — **데이터 층(① ~ ④ 유틸·스토어)은 impl에, 그걸 그리고 조작하는 컴포넌트(⑤ 렌더·헤더 내비)는 여기에.** 이력 → `history/node-ui-render.md`(그 전 (14)(15)는 `history/node-ui-impl.md`) / 데이터 층 → `REF-node-ui-impl.md` / 내비 규칙(무엇을 보여줄지) → `REF-node-ui-overview.md` / 현재 진행 → `CURRENT.md`.

## ⑤ 렌더 최소판 (2026-09-29 작성, 더미 미리보기)
| 컴포넌트 | 위치 | 책임 |
|---|---|---|
| `pages/index.vue` | - | `provideTileTree()` + `seedTileTree`(더미) + `<tile-workspace/>`. **D(index.vue 처리)**: 스크립트(소켓·테스트 함수)는 그대로, 템플릿만 교체 |
| `TileWorkspace` | `feature/node/component/` | `provideTileGrids()` + `<tile-strip/>` + `TileNav`를 헤더로 Teleport (헤더 내비 단계에서 신설) |
| `TileStrip` | `feature/node/component/` | 측정(border-box - padding - 정보 줄 20 - 간격 8, 스크롤바 안 뺌) → 스토어 `area`에 기록 + `viewW`/`scrollX` 기록 → Grid 나열·가로 스크롤. `cols=1`이면 `.narrow`(`(100% - gap)/2`). **E(⑥ 높이)** = `calc(100dvh - var(--v-layout-top))` |
| `TileLayout` / `Tile` | `feature/widget/component/` | 범용 위젯(→ `REF-widget.md`) |
| `TileFrame` | `feature/node/component/tile/` | 헤더: 순번·아이콘·이름(+`(n)`)·`← 연 폴더`·RUN/DONE+kill·크기 `VMenu`(1칸 축 0.5 비활성, `=1×.5` 표시)·닫기(루트/실행 중 비활성) |
| `FolderTileBody` / `TerminalTileBody` | 같은 곳 | 브레드크럼·노드 목록(이름 = `navigate`, 새 타일 = `openFolder`, 실행 = `addTerminal`) / 더미 출력 |
| `tileDummy.ts` | `feature/node/dev/` | 가짜 노드 9개(시안 NODES), 더미 status(reactive Map, `dummyExec`/`dummyKill`), `seedTileTree`(시안 "순서 예시") |

- **칼럼 스냅(2026-09-29 추가)**: strip `scroll-snap-type: x mandatory` + `scroll-padding: var(--gap)`. 앵커 = `TileLayout`의 `overlay` 슬롯에 `columnStops(area.w, cap, GAP, g.cols)` 위치마다 1px 절대배치 `.snap`(`scroll-snap-align: start`). 전체 폭 Grid 폭 = 측정한 `area.w`라 추가 측정 없음. 확인: 앵커 `0, 923 | 1845, 2768 | 3690(1칼럼 Grid)`, 923 = G1 오른쪽 칼럼 + G2 왼쪽 칼럼으로 경계 걸침 스냅 동작, 스크롤 끝은 max로 멈춤.
- **GAP 하나로 통일**: strip padding = Grid 간격 = 타일 간격(8px) = `TILE_GAP`(`tileGrids.store.ts` export, 헤더 내비 단계에서 `TileStrip` 지역 상수에서 옮김). 1칼럼 Grid 폭 `(100% - gap)/2`와 `stripGeometry`가 이 가정에 의존.
- **색**: Vuetify 테마 색 이름만 사용(→ `REF-frontend.md` "색 규칙"). 터미널 배경 = 커스텀 테마 색 `bg-terminal`.
- **함정**: Vuetify 4에서 네이티브 `<button>`에 UA 기본 배경(회색)이 그대로 나옴 → 컴포넌트에서 `background: none; border: 0; color: inherit; font: inherit` 리셋 필요.
- **브라우저 확인(2026-09-29, 임시 dev 서버)**: PC 폭 = G1 `기본 \| sup(2)` / `sup(1) \| Rack A`, G2 1칼럼 폭(`htop / 2F`) / Rack A에서 fw-flash 실행 → htop 뒤 6번으로 들어가 G2 = `htop, fw-flash \| 2F` / strip 390px(폰 폭) → Grid당 2타일·크기 버튼 `.5×.5=1×.5` / 실행 중 닫기 비활성 → kill 후 활성 → 닫으면 `(n)` 표기 사라지고 reflow. 창 크기 조절은 창 관리자가 막아 strip 폭을 직접 줄여 확인.

## 헤더 내비 (2026-09-29 작성) — 구현 순서 4
규칙(무엇을 보여줄지)은 `REF-node-ui-overview.md` "헤더 내비게이션", 여기는 코드 구조.

| 파일 | 책임 |
|---|---|
| `AppHead.vue` | 제목과 메뉴 아이콘 사이에 빈 자리 `<div id="app-head-nav">`(flex:1, min-width:0)만 둠. 노드 기능은 모름. 제목은 `flex: 0 1 auto`(Vuetify 기본 `flex: 1 1`이면 자리와 폭을 나눔) |
| `TileWorkspace.vue` | `provideTileGrids()` → `<tile-strip/>` + `<teleport defer to="#app-head-nav"><tile-nav/></teleport>` |
| `TileNav.vue` | `VPagination`(`:model-value="view.grid + 1"`, `@update:model-value` → `goGrid`) + `prev`/`next` 슬롯 = 칼럼 이동 버튼 + `칼럼 a-b/N` |
| `tileSplit.util.ts` | `stripGeometry(gridCols, areaW, cap, gap)` → Grid·칼럼 x/w / `stripView(geo, scrollX, viewW, gap)` → `{grid, col, from, to}` (순수 함수) |
| `tileGrids.store.ts` | `area`를 스토어 안에서 생성(인자 제거) + `stripEl`/`scrollX`/`viewW` + `geometry`/`view` computed + `goGrid(i)`/`goColumn(±1)`(`stripEl.scrollTo`, smooth) |

- **F(헤더 전달 방식) = Teleport**(사용자 승인): `AppHead`는 `router-view`와 형제라 페이지 아래 provide를 inject할 수 없음. provide를 `App.vue`로 올리는 대신 **내비를 페이지 트리 안에 두고 DOM만 헤더로** 보냄 → 노드 상태는 페이지 범위 유지, 다른 페이지(`/login`)에선 자리가 빔. `defer` 필수(Vue 3.5+): 첫 마운트 때 `AppHead`의 DOM은 아직 문서에 붙기 전이라 즉시 Teleport는 대상을 못 찾음.
- **G(화살표 = 칼럼 이동)**(열린 질문 ⑬ 해소): VPagination `prev`/`next` 슬롯을 직접 그린 `VBtn`으로 교체, Grid 단위 이동은 탭 클릭으로. prev 비활성 = `col === 0`, next 비활성 = 마지막 칼럼이 이미 보임(`to === N-1`, 스크롤 끝).
- **칼럼 = 스냅 한 칸**: 가로 2칸 화면의 2칼럼 Grid = 2칼럼, 1칼럼 Grid·가로 1칸 화면(폰)의 Grid = 1칼럼. `from-to` = 절반 이상 보이는 칼럼 범위.
- **재발 방지 — 현재 Grid는 칼럼 겹침 합으로 잰다**: Grid 구간 `[x, x+w]`로 재면 칼럼 사이 간격(8px)까지 세어서, 스크롤 끝(G1 오른쪽 칼럼 + G2 1칼럼)에서 716 대 708로 G1이 이겨 "동점이면 뒤쪽" 규칙이 깨졌음(실제 발생·수정). 동점 허용 오차 0.5px.
- **되먹임 방지**: 탭 표시는 스크롤에서만 파생(`:model-value`), 클릭은 스크롤 요청만. VPagination은 `modelValue`가 넘어오면 내부 값을 바꾸지 않고 emit만 함(`useProxiedModel`) — smooth 스크롤 중엔 탭 강조가 지나가는 Grid를 따라 움직임.
- **탭 표시 = 숫자만**(VPagination 기본 버튼). 시안의 `G1` 표기는 `item` 슬롯에서 `VBtn`을 직접 그려야 하는데, 그러면 `VPaginationBtn` 기본값(크기·색·variant) 상속이 끊겨 기각. Grid 이름은 Grid 하단 정보 줄의 `Grid n`으로 충분.
- **Vuetify 4 타이포 클래스**: `text-caption` 없음(MD3 이름으로 바뀜: `text-body-small` 12px / `text-label-small` 11px 등) — 없는 클래스를 쓰면 조용히 무시됨.
- **확인(2026-09-29, 헤드리스 Chromium 1440×900 / 390×844)**: PC 시드 = Grid 2·칼럼 3, `칼럼 1-2/3` → next → `[2]`·`칼럼 2-3/3`·next 비활성 / 폰 = Grid 3, 칼럼 이동마다 탭 `1→2→3` / 직접 `scrollTo(0)` → 탭 `[1]` 복귀 / 실행 14번 → PC Grid 5 전부 표시, 폰 Grid 10 → `[1] 2 3 … 10`, 마지막 탭 → `1 … 8 9 [10]`·`칼럼 10/10`. 콘솔 에러는 백엔드 미기동 API 실패뿐.
- 미확인: 실제 기기 터치 스크롤, 헤더 높이 변화 시(`height="auto"`) 레이아웃, 탭 가운데 정렬 유지 여부(VPagination 기본 `justify-content: center`).

## 타일 위치 따라가기 `reveal` (2026-09-30 작성)
크기 변경이나 새 타일 생성으로 타일이 화면 밖 Grid로 가면 그 Grid로 스크롤한다.

| 항목 | 결정 (사용자 선택) |
|---|---|
| 이동 조건 | **화면 밖일 때만**. 타일이 다 보이면 스크롤하지 않음(같은 Grid 안에서 자리만 바뀌면 화면 고정) |
| 이동 위치 | **그 타일의 Grid 시작** = `goGrid(gi)`. Grid 폭 = 화면 폭이라 타일이 전부 보이고 Grid 탭과 일치 |
| 적용 범위 | 크기 변경(`TileFrame.setSize`) + 새 타일(`FolderTileBody`의 `onOpen`/`onExec`, `openFolder`/`addTerminal`이 돌려준 id) |

- `tileGrids.store.ts` `reveal(tileId)`: `nextTick` 대기 → `grids`에서 타일의 Grid·칸 찾기 → `geometry.columns`(해당 Grid)로 타일 x 구간 → `stripEl`의 `scrollLeft`/`clientWidth`와 비교(1px 오차 허용) → 밖이면 `goGrid`.
- **재발 방지 — 순서 의존**: `nextTick` 전에 `scrollTo`를 부르면 늘어난 Grid가 아직 DOM에 없어 옛 스크롤 폭에서 잘림. 스크롤 위치도 `scrollX`(스크롤 이벤트로 늦게 갱신)가 아니라 요소에서 직접 읽음 — 폭이 줄며 잘린 직후일 수 있음.
- 따라가는 대상 = 조작한 타일 하나. 그 여파로 밀려난 다른 타일(예: sup(2))은 따라가지 않음.
- **확인(2026-09-30, 헤드리스 Chromium)**: PC 1440×900 — sup(1) 세로 1.0(G1 유지) → 스크롤 0 유지 / 이어서 가로 1.0(G2로) → G2로 이동 / 되돌리면 G1로 복귀 / 실행 3번(G2·G3에 생김) → 매번 새 타일이 보임. 폰 390×844 — sup(1) 세로 1.0(G2로) → 이동, 되돌리면 복귀, 실행 3번 → G4·G5로 이동.
