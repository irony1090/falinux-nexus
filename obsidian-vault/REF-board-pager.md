# REF — 보드 페이지 표시 개편 (TileNav)

> 이름: GridPager = **인디케이터**(2026-10-20 "표시줄"에서 변경, → `REF-board-chrome.md`)

> 2026-10-20 신설. 이력 → `history/board-pager.md` / 지금 화면 이동 구조 → `REF-node-ui-overview.md` / URL 복원 → `INDEX-tile-url.md` / 현재 진행 → `CURRENT.md`

**상태: 마감(2026-10-20) — 헤드리스 31/31 + 사용자 확인("매우 맘에 든다"). 좁은 Grid 표시(막대 정수 + 숫자 `3-4`)·폰 도트 3개는 현재 동작 그대로 둠.** 마감 색인 → `INDEX-board-pager.md`
- 프로토타입: https://claude.ai/artifact/XLvPvzUkZQrfD1FN6Vaqo5 (데스크톱 2x2 Grid 5개 / 폰 세로 1x2 Grid 8개, Tweaks = 속도·드래그 감도·강조색)

## 바꾸는 것
지금 `TileNav.vue` = `VPagination`(Grid 탭) + prev/next 화살표(칼럼 이동) + "칼럼 n/m" 텍스트 → 아래 표의 인디케이터 하나로 바꾼다.

| 조작 | 동작 |
|---|---|
| 도트 클릭 | 그 Grid 시작으로 이동 |
| 인디케이터 잡고 x축 드래그 | 정지 위치 단위로 이동(보드가 손을 바로 따라오고, 놓으면 가까운 정지 위치로) |
| 드래그한 채 아래로 | 타일 표시(Grid별 미니맵 + 보이는 범위 틀)로 바뀜 |
| 그 상태로 x축 드래그 | 타일 순서대로 active 이동 + 아래 "타일 위치 규칙"대로 스크롤 |
| 위로 되돌림 / 손 뗌 | Grid 이동으로 복귀 / 접힘(위치 유지) |

## 확정 (2026-10-20 사용자)
| 항목 | 결정 |
|---|---|
| 이동 단위 | **가장 좁은 타일 가로 폭** = 데스크톱 0.5 Grid(칼럼), 폰(가로 1칸) 1 Grid |
| 반 칸 정지 범위 | **가) 모든 Grid에서 0.5 단위로 정지** — 1.0 타일만 있는 Grid도 반 칸에 멈출 수 있음(1.0 타일이 잘려 보여도 허용). **※정정(2026-10-20 코드 확인): 지금 앱이 이미 이렇게 동작** — `Grid.cols`는 타일 폭이 아니라 "오른쪽 칼럼이 통째로 비었나"(`tileOrder.util.ts`)라 1.0 타일 Grid도 2칼럼 = 반 칸 정지 있음. 오른쪽이 빈 Grid는 폭 자체가 반(`narrow`). `columnStops` 변경 불필요 |
| 2.5 표시 | **두 도트 사이 작은 도트** — 현재 위치 막대가 반 칸으로 갈수록 작은 도트로 줄어듦. 숫자 표시 = `2-3 / 5` |
| 타일 위치 규칙 | 대상이 화면 밖일 때만, **막 다 보이는 만큼만** 이동: 오른쪽 밖 = 화면 오른쪽 끝에, 왼쪽 밖 = 왼쪽 끝에. 이미 보이면 그대로. **1.0 폭 타일이면 그 Grid 전체가 보이는 위치** |
| 기준 화면 | 드래그 중 **타일이 바뀔 때마다 직전 화면 기준**으로 계산(중간에 1.0 타일을 지나면 그 Grid 전체로 갔다가 다음 타일에서 다시 반 칸) — 사용자가 원한 동작 |
| 인디케이터 위치 | **헤더 가운데 정렬** — 폰(좌우 영역 숨김)에서도 가운데 |
| Grid가 많을 때 | **창 방식** — 도트·미니맵 각각 최대 9개만 보이고 현재 위치 주변으로 미끄러짐. 더 있는 쪽 끝 도트는 작게, 미니맵 끝은 흐리게. 드래그는 px 기준이라 영향 없음. 실제 구현은 9 대신 폭을 재서 정할 것(지금 `VPagination`이 하던 일) |
| 타일 표시로 바뀔 때 | active가 화면에 없으면 화면에 보이는 첫 타일을 active로 |

- 2026-10-21 이 규칙을 단축키·폴더 열기·URL 등 모든 타일 이동에 적용 -> `REF-tile-move.md` 이동-1(타일 위치 규칙 통일)

예: 1페이지 -> 3페이지 `serial-log`(왼쪽 칼럼) = 2페이지 오른쪽 절반 + serial-log가 화면 오른쪽 / 5페이지 -> serial-log = 3페이지 전체(serial-log가 화면 왼쪽)

## 프로토타입 계산 (구현 때 옮길 것)
- 위치 = 화면 왼쪽 끝 칸 번호 `v`(칸 = 1/cols Grid). 범위 `0 ~ (Grid 수 - 1) * cols`
- 타일 칸 범위 `[l, r)` = `Grid 번호 * cols + x * cols` ~ `+ w * cols`
- `reveal`: `r - l >= cols` -> `v = l` / `l < v` -> `v = l` / `r > v + cols` -> `v = r - cols`
- 드래그 6px 전엔 포인터 캡처 안 함 -> 도트 클릭(button)이 그대로 동작. 인디케이터 `touch-action: none`
- 아래로 36px = 타일 표시, 16px 위로 = 복귀. 페이지 드래그는 칸당 36px, 타일 드래그는 타일당 약 29px

## 구현 (2026-10-20, N1~N6)
| 단계 | 파일 | 내용 |
|---|---|---|
| N1(위치 계산) | `widget/util/tileSplit.util.ts` | `StripColumn.pos`(오른쪽 칼럼 = +0.5) · `stripPos`(scrollX -> Grid 단위 실수) · `colScroll`(칼럼 실수 -> scrollLeft) · `revealCol` |
| N2(스토어 동작) | `node/store/tileGrids.store.ts` | `pos` · `maxCol`(스크롤 끝에 막히는 칼럼 제외) · `scrubbing` · `goCol` · `scrubTo` · `colRange` · `stepTo` · `focusTile` 노출 |
| N3(스냅 끄기) | `TileStrip.vue` | `.scrubbing` = `scroll-snap-type: none` |
| N4(인디케이터) | `GridPager.vue` + `hook/pagerDrag.hook.ts` | 도트·막대(반 칸 = 작은 도트)·창 방식(폭 재서 개수, 도트 5개도 안 들어가면 숫자 숨김)·←/→ 키. 드래그 상수 START 6 · TILE_DOWN 36 · TILE_UP 16 · COL_PX 36 · TILE_PX 29 |
| N5(타일 미니맵) | `TilePeek.vue` | Grid 칸 비율 미니맵 + 보이는 범위 틀(strip 좌표 -> 미니맵 선형) + 창 방식 + `타일 i / N` |
| N6(TileNav 교체) | `TileNav.vue` | `VPagination`·칼럼 화살표·"칼럼 n/m" 제거 -> `GridPager`. 전체보기 분기 그대로 |
- 안내·패널은 앱 바(`v-app-bar`, overflow 잘림 + transform) 안에서 못 나오므로 **body로 teleport + fixed**, 위치는 드래그 시작 때 인디케이터 rect로
- 놓으면 `settle`: 가까운 칼럼 `goCol` -> 그 위치에 도착(`scrollend` + 위치 확인, 최대 700ms)한 뒤 스냅 복원 — smooth 도중 스냅을 켜면 다른 칼럼으로 튈 수 있음
- 타일 모드에서 놓을 때 포커스는 `pointer: fine`일 때만(U3(복원 때 포커스)와 같은 규칙)
- **좁은 Grid(오른쪽 칼럼이 빈 Grid)**: 그 칼럼이 왼쪽 끝이면 막대는 그 Grid 도트(정수), 숫자는 `3-4 / 6`(다음 Grid 왼쪽 절반도 보이므로) — 현재 동작 유지(마감 때 사용자 확인)
- 폰 헤더(390px)는 내비 자리가 약 104px -> 도트 3개 + 숫자 숨김

## 남은 것 / 구현 때 본 것
- 터치 드래그는 헤드리스로 못 봄(마우스로만) -> 사용자 실기기 확인
- 보드 이동 애니메이션 = 브라우저 기본 smooth(결정 1), 전체보기 TileNav는 그대로
- 확인 스크립트 명세 → `REF-test-scripts-1020.md` "2026-10-20 페이지 표시"
