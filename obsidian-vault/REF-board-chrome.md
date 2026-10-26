# REF — 보드 화면 헤더·푸터 개편

> 2026-10-20 신설. 이력 → `history/board-chrome.md` / 키보드 열림 확인·맞춤 계획 → `REF-board-keyboard.md` / 인디케이터 상세 → `REF-board-pager.md`(마감, `INDEX-board-pager.md`) / 키보드 제약 → `REF-mobile-viewport.md` 4 / 현재 진행 → `CURRENT.md`

**상태: 마감(2026-10-20, → `INDEX-board-chrome.md`) — 푸터-1(푸터 뼈대)~푸터-7(빈 곳 휠) 코드, 푸터-8(확인) 격리 헤드리스 40/40(2026-10-20). 남은 것 = 사용자 브라우저(특히 폰 실기기·키보드 열림) 확인 -> 파트 마감·커밋.**
- 시안 캔버스: https://claude.ai/artifact/8UVsShdtRn1hraKLcUh6DH — PC A(한 줄 바)·B(떠 있는 알약)·C(예약 띠 + 알약)·D(자동 숨김) / 모바일 A·B·C·E(접기형)

## 범위
**보드 페이지 평소 화면 중심.** 헤더 삭제는 앱 전체(로그인 포함). 전체보기 중 화면은 조작 자리(푸터 가운데)만 정했고, 내부 세부는 나중에

## 용어 (2026-10-20 확정)
| 위치 | 이름 | 컴포넌트 | 하는 일 |
|---|---|---|---|
| 헤더 | 로고 | `AppHead` | 누르면 `/`(기본 보드) |
| 헤더 | 보드 메뉴 | `BoardMenu` | 보드 전환·이름 변경·삭제·내가 공유한 것·계정 닉네임 |
| 헤더 | **인디케이터** | `TileNav` > `GridPager` (+ `TilePeek`) | 도트·드래그 칼럼 이동·당겨서 타일 선택. 옛 이름 "표시줄" |
| 헤더 | 알림 | `NoticeCenter` | 종 + 안 읽은 수, 오른쪽 드로어(헤더 높이만큼 내려 그림) |
| 헤더 | 전체 화면 | `AppHead` | |
| 헤더 | 로그아웃 | `AppHead` | |
| 푸터 | **Grid 정보** | `TileStrip` `#foot` | Grid마다 `Grid 2` 태그 + `row · 4타일`, 가로 스크롤 따라 움직임 |
| 푸터 | 전체보기 | `OverviewButton` | 푸터 줄 오른쪽 끝 고정, 스크롤바 두께만큼 비킴 |
| 푸터 | 가로 스크롤바 | `TileStrip` 스크롤바 | 푸터 줄 바로 아래 |

- "Grid 정보"를 고른 이유: 내용이 전부 Grid 하나 단위(번호·배치 모양·타일 수). "타일 정보"는 타일 하나로 읽힘
- 코드 주석에는 아직 "표시줄"이 남아 있음(`GridPager.vue` 등 14곳) — 코드는 사용자가 고침

## 새 배치 F(확정안) — 2026-10-20 사용자 정의, 시안 확인 대기
캔버스 맨 아래 줄 `PC-F` · `M-F`(눌러 볼 수 있음)

| 항목 | 결정 |
|---|---|
| 헤더 | **삭제**(앱 전체 — 페이지가 `index`·`Login`뿐, 로그인엔 불필요) |
| 푸터 | 지금 `TILE_FOOT` 20 + `TILE_GAP` 8보다 조금 높게(시안 PC 44 / 모바일 56), 화면 고정 한 줄 |
| 푸터 왼쪽 끝 | 로고 + 보드 메뉴(위로 열림). 모바일은 로고만, 보드 이름은 메뉴 안 |
| 푸터 가운데 | 인디케이터(항상 표시). **전체보기 버튼은 인디케이터 알약 안 오른쪽 끝**에 높이·오른쪽 끝까지 꽉 채움 + 왼쪽도 둥글게(버튼 자체가 알약), 색 유지(primary), PC = 아이콘 + "전체" / 모바일 = 아이콘만. 전체보기 중엔 같은 알약이 요약 + 축소 비율 `- 60% +` + 같은 자리 돌아가기 |
| 푸터 오른쪽 끝 | 알림 · 전체 화면 · 로그아웃. 좁으면(모바일) `...` 하나로 합치고 위로 펼침 + 안 읽은 알림 빨간 점 |
| Grid 정보 | Grid를 넘길 때마다 타일 왼쪽 위에 토스트, 일정 시간(시안 2.5초) 뒤 사라짐, 누르면 바로 사라짐. 타일 이름 줄을 잠깐 가리는 건 허용 |
| 가로 스크롤바 | **삭제**(인디케이터가 대신, 터치·트랙패드·Shift+휠은 그대로) |
| 휠 | 타일이 아닌 곳(타일 사이·푸터)에서 휠 = 페이지 이동. 타일 위 휠은 지금처럼 터미널 스크롤 |
| 타일 사이 클릭 메뉴 | **폐기** |

### 결정 (2026-10-20)
- 휠 1회 이동 단위 = **칼럼(반 칸)** — 인디케이터 드래그·정지 위치(`columnStops`)와 같은 단위. 트랙패드는 delta를 모아 기준 넘을 때 한 칸 + 간격 제한 필요
- 반 칸 위치 Grid 정보 토스트 = **`Grid 2-3`** 처럼 범위로. 인디케이터 숫자와 같은 계산(`GridPager.vue`의 `range`: 화면에 절반 이상 보이는 칼럼의 Grid 범위) 재사용 — 공용으로 빼야 함. 두 Grid면 정보(`row · 4타일`)도 둘 다
- 시안(`PC-F`)은 Grid 단위로만 움직임 — 반 칸은 구현에서

## 구현 계획 (2026-10-20 제시 · **사용자 허가**)
| 단계 | 하는 일 | 주요 파일 | 앞 단계 |
|---|---|---|---|
| 푸터-1(푸터 뼈대) | `BoardFooter.vue` 신설(왼·가운데·오른쪽 3구역), `TileWorkspace__body` 맨 아래 flex 자식. strip은 남은 높이만 씀 -> 영역 측정의 `- FOOT - GAP`에서 `FOOT` 제거, Grid 정보 `#foot` 삭제·`footHeight` 0 | `TileWorkspace` `TileStrip` `TileLayout`(foot 0) `tileGrids.store`(`TILE_FOOT`) | - |
| 푸터-2(요소 옮기기) | 왼쪽 = 로고 + `BoardMenu`(`location` 위로) / 가운데 = `TileNav`(Teleport 제거하고 직접) + 전체보기를 인디케이터 알약 안으로(`OverviewButton` 흡수, 전체보기 중 알약에도 돌아가기) / 오른쪽 = 알림 종(`NoticeCenter`에서 버튼만 분리) + 전체 화면·로그아웃(`AppHead`의 로직을 훅으로 이동), 폰은 `...` 하나로 합쳐 위로 | `BoardFooter` `BoardMenu` `TileNav` `GridPager` `OverviewButton` `NoticeCenter` `AppHead` | 푸터-1(푸터 뼈대) |
| 푸터-3(헤더 삭제) | `App.vue`에서 `AppHead` 제거. 헤더 높이 의존 정리: `AppBody`의 `--v-layout-top`, `TileWorkspace` 높이 계산 2곳, `NoticeCenter` 드로어 offset, `appWindown.store`가 읽는 `--v-layout-top`, 키보드 이동의 "헤더 아래" 기준 | `App.vue` `AppBody` `TileWorkspace` `NoticeCenter` `appWindown.store` `appHead.store` | 푸터-2(요소 옮기기) — Teleport 자리가 먼저 없어지면 안 됨 |
| 푸터-4(위로 열기) | 인디케이터 당김 방향 반대(`pagerDrag.hook`의 `dy > TILE_DOWN` / `dy < TILE_UP`), 안내·`TilePeek` 위치를 막대 위쪽 기준으로(`anchor.top = r.bottom + 8` -> 위), 안내 문구 "위로 당기면" | `pagerDrag.hook` `GridPager` `TilePeek` | 푸터-2(요소 옮기기) |
| 푸터-5(스크롤바 삭제) | strip 스크롤바 숨김(CSS), `scrollbar` 상태·`syncScrollbar`·전체보기 버튼의 두께 보정 제거 | `TileStrip` `tileGrids.store` | 푸터-2(요소 옮기기) |
| 푸터-6(Grid 정보 토스트) | `GridPager`의 `range`(절반 이상 보이는 칼럼의 Grid 범위)를 store로 올려 공용화. `GridToast.vue`: range가 바뀌고 스크롤이 멈추면 표시(드래그 중엔 대기), 2.5초 뒤 사라짐, 누르면 사라짐. `Grid 2-3` + 각 Grid `row · 4타일` | `tileGrids.store` `GridPager` `GridToast`(신설) | 푸터-1(푸터 뼈대) |
| 푸터-7(빈 곳 휠) | 타일 밖(`.TileFrame` 밖: 타일 사이·푸터) 휠 -> 반 칸 이동(store `goColumn(±1)`). delta 누적 + 기준 넘으면 1칸 + 간격 제한(트랙패드). 전체보기 중 기존 `onWheel`은 그대로 | `TileWorkspace` 또는 `TileStrip` | 푸터-2(요소 옮기기) |
| 푸터-8(확인) | 격리 헤드리스(PC·폰 크기) + 사용자 브라우저. 모바일 키보드 열림 때 푸터가 키보드 뒤로 가는지·키 줄과 안 겹치는지(`REF-mobile-viewport.md` 4) | - | 전부 |

### 구현 메모 (2026-10-20 푸터-1(푸터 뼈대)~푸터-7(빈 곳 휠))
| 파일 | 한 일 |
|---|---|
| `BoardFooter.vue`(신설) | grid `1fr auto 1fr`, 가운데 폭 = 전체 - 양쪽 중 넓은 쪽 x2(인디케이터가 화면 가운데 + 그 폭으로 도트 수). 높이·알약 높이 `SIZE` PC 44/34 · 폰 56/44(`--pill-h`). 키보드 열림 = `visibility: hidden`(높이 유지) |
| `FootPill.vue`(신설) | 알약 공통 모양(테두리·`elevation-2`·surface) |
| `TileNav.vue` | 인디케이터 `#end` 슬롯에 전체보기(primary, 폰 아이콘만) / 전체보기 중 = 요약·축소·돌아가기(tonal) 알약. `OverviewButton.vue` 삭제 |
| `GridPager.vue` | 바깥 알약 `.shell` + `end` 슬롯, 슬롯 폭을 도트 자리에서 뺌. 바 자체 배경 제거 |
| `BoardMenu.vue` | 활성자 = 로고 + 이름 + 위 화살표(`compact` = 로고만), `location="top start"`. 로고 눌러 `/` 이동은 없어짐(메뉴의 기본 보드로) |
| `appActions.hook.ts`(신설) | `AppHead`의 전체 화면·로그아웃 로직 |
| 알림 | 드로어 열림 = `tileNotice.store`의 `drawerOpen`, 종은 푸터(폰은 `...` 안 + 빨간 점). 토스트 `bottom` 56px 위로 |
| 헤더 삭제 | `AppHead.vue`·`appHead.store.ts` 삭제, `AppBody`의 `--v-layout-top` 덮어쓰기 제거, `TileWorkspace` 높이 `100dvh`·키보드 고정 `baseH`px |
| Grid 정보 | `TileStrip` `#foot`·`TILE_FOOT` 삭제, 영역 측정에서 `- FOOT - GAP` 제거 |
| 푸터-4(위로 열기) | `pagerDrag.hook` `TILE_PULL`/`TILE_BACK`(위로 36 = 타일, 16 아래로 = 칼럼). 안내·`TilePeek`은 `bottom` 기준(막대 위 8px), 등장 방향도 아래->위 |
| 푸터-5(스크롤바 삭제) | `.TileStrip:not(.overview)` 스크롤바 숨김(전체보기는 인디케이터가 없어 유지), store `scrollbar`·`syncScrollbar` 삭제 |
| 푸터-6(Grid 정보 토스트) | store `range` 공용화(`GridPager`도 이걸 씀). `GridToast.vue`: 범위 key가 바뀌고 스크롤이 150ms 멈추면 표시, 드래그(`scrubbing`) 중 대기, 2.5초, 누르면 닫힘(`pointerdown.prevent`로 포커스 유지), 전체보기 중 숨김·돌아오면 다시 표시. 위치 기준 = `TileWorkspace__stage`(strip 밖 래퍼) |
| 푸터-7(빈 곳 휠) | `gapWheel.hook.ts`(2026-10-26 `widget/hook`, 판 계약): 세로 우세 휠만, 대상이 `.Tile:not(.Tile--empty)` 밖이면 `preventDefault` + delta 누적 50px -> `goCol(±1)`, 간격 250ms, 200ms 멈추면 누적 버림, smooth 도중(600ms) 다음 칸은 목표 칼럼 기준. `TileWorkspace__body`의 `@wheel` |
| 버튼 크기 | 알약 안 `v-btn`은 `density="comfortable"`(기본 크기면 PC 40px·폰 48px 버튼이 34/44 알약을 넘침 — 헤드리스에서 발견) |

- 남은 빈틈: 보드 불러오기 실패·로딩 화면(`index.vue` state)엔 푸터가 없어 로그아웃 버튼이 없음 — "기본 보드로"로 나갈 수는 있음
- 작은 결정(플랜 허가로 추천안 채택): 폰 판정 = `cap.x === 1`(지금 `OverviewButton` 아이콘만 기준과 같게) / 푸터 높이 상수 PC 44 · 폰 56(시안 값)
- `TileLayout`의 foot는 전체보기 축소 배치에도 쓰이므로 0으로 바꿀 때 전체보기 "현재 배치 그대로" 확인

## 키보드 열림 확인 -> `REF-board-keyboard.md`
- Edge 전체 화면 키보드 가림 조사(측정 도구·판정) + 맞춤 계획(포커스 타일만 키보드 위에 맞춤)은 분리(2026-10-20, 10k 초과)

## 옮길 때 걸리는 것 (검토 메모)
- 인디케이터 "아래로 당기면 타일 선택" -> 하단이면 위로 당김, `TilePeek`도 위로 열림
- 모바일 키보드 열림(`resizes-content`) 중엔 하단 요소가 키보드 위 버튼 줄과 겹침 -> 숨김
- 버튼 누를 때 터미널 포커스가 빠지면 모바일 키보드가 닫힘 -> 포커스 안 뺏기, 닫을 때 active 터미널로 되돌림
- 알림 종이 메뉴 안으로 들어가면 메뉴 버튼에 빨간 점 배지 필요. `NoticeCenter` 드로어 위치(헤더 높이 기준)도 옮겨야 함
- 헤더 삭제 -> `TileWorkspace` 높이(`--v-layout-top` 빼기)·`AppBody`·`NoticeCenter` 드로어의 헤더 높이 의존 정리. 타일 높이 계산(`TileStrip`의 `- FOOT - GAP`)은 고정 푸터 기준으로
- 위로 열리게: 보드 메뉴(`location`), `TilePeek`·당김 안내
