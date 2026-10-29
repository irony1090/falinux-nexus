# REF — 파트(타일 끌기) 끌기-1·끌기-2·제스처-1 구현

> 2026-10-29 `REF-tile-drag.md`에서 분할. 계획·결정·설계 → `REF-tile-drag.md` / 이력 → `history/tile-drag-impl.md` / 끌기 기반(범용) → `REF-util-drag.md` / 현재 진행 → `CURRENT.md`

## 구현 (2026-10-29)
| 곳 | 내용 |
|---|---|
| `common/listener/dragSession.ts` | 끌기-1(끌기 기반) 범용 `DragSession` → 반복 주제 `REF-util-drag.md` |
| `feature/widget/util/tileDrop.util.ts` `dropOrder` | 놓을 자리 = 흔들림 방지 ①②③. 배치 위치는 `pack`+`stripGeometry`로 직접 계산(평소 모드 CSS 가정). 빈 칸 = 그 Grid 마지막 타일 뒤, 모든 Grid 뒤 = 맨 뒤(둘 다 ③ 없이 — 옮긴 뒤 끄는 타일이 그 끝이라 다시 안 옮김) |
| `feature/node/store/tileDrag.store.ts` | 보드 연결(`TileWorkspace`에서 provide). 시작 자리 `dragFrom` = 마우스 머리 줄 빈 곳 / 터치 + 몸통 `data-pager-press` 칸 밖. 끄는 동안 `grids.preview`(미리보기 순서) + `scrubbing`(스냅 끔) + 화면 끝 48px 자동 스크롤. 놓으면 `place(X, 미리보기에서 X 앞 타일)` + X active. 취소 = Esc·키보드 열림·전체보기·끄는 타일 사라짐 |
| `tileTree.store` `placeOp` | X를 Y 바로 뒤로(Y 없으면 맨 뒤, 루트 불가). 최소화 타일은 문서 order에 그대로 → 상대 위치 유지 |
| `tileGrids.store` `preview` | 있으면 `layoutOf`가 문서 order 대신 씀(위치 번호도 미리보기 기준) |
| `TileFrame` | `--slot`(빈 자리: 내용 visibility hidden, 그대로 마운트, **배경 불투명**(surface + 파랑 8% 겹침) — 반투명이면 GPU 잔상, 아래 함정) · `--pressing`(터치 누르는 중) |
| `TileDragChip.vue` | 안 1 작은 표시(아이콘·이름), 끄는 동안 `html.tile-dragging`(커서·글자 선택) |
| 제스처-1 `pagerOpen.hook` | 터치 = 두 손가락 펼치기(touch 이벤트로 잼, 40px 벌어지면, 두 손가락 가운데). 길게 누르기 코드 삭제. 옵션 `pressable` → `pinchable`. 이미지 보기 몸통 = `data-pager-pinch-off`. **관리 페이지(`MemberPagerPopup`)도 펼치기로 같이 바뀜**(길게 누르기 없음) |
- 남은 것: 전체보기 끌기(보드-1과 같이 — 전체보기는 축소·캡션이라 배치 위치 계산이 다름) · 폰 실기기(Android "길게 누른 뒤 움직이기") · 미리보기 중 다른 Grid로 옮겨 가는 타일은 다시 마운트됨(터미널은 attach/detach라 화면 유지, 기존 순서 변경과 같음)
- 함정(10-29 사용자 발견): 빈 자리 배경이 반투명이면 실제 Chrome(GPU)에서 옆 칸(루트) 글자 조각이 비침 — 루트가 보이는 화면에서 3번 자리로 끌 때 잦음. 헤드리스(GPU 없음)는 재현 안 됨, DOM상 보이는 요소 0. 추정 = 다시 쓰는 GPU 조각을 안 지우고 반투명만 칠함 -> 불투명 바탕으로 고침 -> 사용자 PC Chrome에서 재현 안 됨 확인(10-29). 스크린샷 = user vault `user/스크린샷 2026-10-29 15-48-27.png`
- 함정(10-29 사용자 폰): 끌어서 다른 페이지(Grid)로 넘어가면 끊김 = 미리보기에서 끄는 타일이 다른 Grid로 가며 다시 마운트 -> 손가락 댄 머리 줄이 문서에서 빠져 touchmove가 strip에 안 옴 -> 스크롤 -> pointercancel. `DragSession`이 시작 요소에도 막기 리스너를 걸어 고침(→ `REF-util-drag.md` 함정). 재현 스크립트 = scratchpad `ui/phone.mjs`(폭 400 터치, 고친 뒤 끝까지 끌림). 사용자 폰(Android) 확인 OK
- 헤드리스 확인 21/21(격리 5051·3001, td1): 마우스 임계·끌기·저장·active·Esc·루트 제외·자동 스크롤 / 터치 짧게 움직임 = 스크롤·길게 누른 뒤 끌기·패널 안 뜸 / 펼치기 = 패널(가운데)·바깥 탭 닫힘. 스크립트 = 이 세션 scratchpad `ui/drag.mjs`·`setup.py`(재부팅 때 사라짐)

## 보드-1(보드 간 이동 API) 구현 (2026-10-29)
| 곳 | 내용 |
|---|---|
| 서버 `POST /api/workspaces/:hash/tiles/move` `{tileId, to, after, version?}` (`workspace.go moveWorkspaceTile`) | 옛 `/:hash/move`(폴더만) 대체. 같은 보드 = `placeAfter`로 그 타일 하나 / 다른 보드 = `takeOpened`(연 곳 사슬 묶음, 출발 order 순) -> `putOpened`(after 뒤, 연 곳이 묶음 밖인 타일 = 도착 루트). after `""` = 루트 바로 뒤, 없는 타일 = 맨 뒤. version은 보낼 때만 검사(지금 보는 보드), 도착은 검사 없음. `deleteWorkspace`도 `putOpened`로 |
| 프론트 `moveBoardTile` · `tileTree.moveTile(from, id, to, after)` | 출발 = 이 보드면 `serverChange`(version), 아니면 `flush` 뒤 검사 없이. 메뉴 "다른 보드로" = after `""` |
| `tileGrids.preview` | `{ orders: {hash: order}, tiles }` — 보드별 미리보기 + 보드를 옮겨 가는 묶음 타일. 지금 보드·다른 보드 줄 `layoutOf`가 같이 씀 |
| `tileDrop.util` | `DropFrame` = Grid 상자 위치 함수 + 축소 비율(평소·전체보기 공용). `dropBundle` = 다른 보드 줄: 묶음 위 = 그대로, 타일 위 = 앞/뒤 중 **묶음 어느 타일이든** 커서를 덮는 쪽(③ 묶음 단위), 빈 칸 = 그 Grid 마지막(묶음 아닌) 타일 뒤, 줄의 빈 곳 = 맨 뒤 |
| `tileDrag.store` | 전체보기 = `downOv`(TileStrip 타일 아무 데서나, 모든 줄), 줄 판정 = `elementFromPoint` -> `.board[data-hash]`. 전체보기 위치 = 첫 상자를 재고 상자 폭 + 14px로 이음(보드 하나면 가운데 정렬 보정). 놓으면 같은 보드·지금 보드 = `place`, 그 밖 = `moveTile`(응답까지 미리보기 유지). 상하좌우 자동 스크롤 |
| `TileStrip` | `.board`에 `data-hash` · 빈 자리 = `<tile>`의 `drag-slot`(불투명 바탕, TileFrame `--slot` 대체) · 다른 보드에서 들어온 타일 = `OverviewTile` · 전체보기 캡션 폭 = 상자 폭(`.cap` width 0 + min-width 100%, 줄임표) |
- 묶음 표시(보드Q6): `tileDrag.bundleIds`(다른 보드 줄 위일 때만, 끄는 타일 제외) -> `<tile>` `drag-bundle`(::after 실선 + 바탕, 두께 `calc(3px / var(--s))`) · `bundleChips`(시작 때 `.Tile[data-tile-id] .ov-name .name` 글자, 화면에 없는 최소화 터미널 = "최소화한 터미널") -> `TileDragChip` 아래 줄. 전체보기 빈 자리 테두리도 축소 보정. 헤드리스 `ui/bundle.mjs` 4/4
- 함정(10-29): 묶음 첫 타일 ≠ 맨 위 타일 — order는 연 곳과 무관해 연 터미널이 폴더보다 앞일 수 있음. 서버 `putOpened` 처음 구현이 첫 타일 연 곳을 바꿔 500(검사 실패). 프론트 ③도 같은 이유로 묶음 단위
- 헤드리스: 전체보기 15/15(`ui/ov.mjs`: 같은 보드·다른 보드로 묶음·다른 보드 줄 안·다른 보드에서 지금 보드 맨 뒤·Esc·클릭 = 이동) + 평소 21/21. 서버 테스트 `TestMoveOpenedBetweenBoards`(도착 루트 id 다르게, 연 터미널이 앞). 시험 데이터 = `ui/setup.py`(기본 보드)·`ui/b3save.py save|restore`(b3)

## 칸-1(node 순서 API)·칸-2(node 칸 끌기) 구현 (2026-10-29)
| 곳 | 내용 |
|---|---|
| 서버 `POST /api/nodes/order` `{parentId, ids, after}` (`nodeOrder.go`) | 칸Q2(저장 방식). 부모 편집 이상(`requireFolder`), parentId null = 내 루트만(칸Q3). `ListSiblingOrds`(차단 형제 포함, `FOR UPDATE`) -> `placeSiblings`(옮길 것 = 형제 순서 유지, after 없음 = 맨 앞, 형제에 없음 = 맨 뒤) -> 바뀐 것만 `SetNodeOrds`(1부터). `NODE:UPDATE` 한 번 |
| `SetNodeOrds` 쿼리 | 두 배열 `unnest`는 sqlc 분석기가 못 읽음(`function unnest(unknown, unknown) does not exist`) -> `ords[array_position(ids, id)]` |
| `patchNode` 이동 | 부모가 바뀌고 ord를 안 보내면 새 폴더 끝(`MaxChildOrd + 1`). 전에는 옛 ord를 들고 가 새 폴더 아무 데나 끼어듦 |
| 프론트 `hook/cellDrag.hook.ts` `useCellDrag` | 폴더 타일 몸통마다 하나 + 모듈 상태(끄는 중 1개, 몸통 등록 = 다른 타일 놓기). `DragSession`(끌기-1) 재사용 |
| 판정 | 칸 찾기 = `offsetLeft/Top`(비켜 가는 transform 무시, 흔들림 방지 ①). 끄는 칸(빈 자리) 위 = 그대로(②). 넣을 수 있는 폴더 칸 = 가장자리 ¼ 순서·가운데 0.6초 머묾 = 넣기 / 그 밖 칸 = 반씩 앞·뒤. 격자 빈 곳 = 순서 칸 맨 뒤. 넣기 되면 빈 자리 원위치(폴더 칸이 커서 밑에서 비켜 나갈 때만 그대로). 다른 폴더 타일 = 칸 위면 그 폴더, 아니면 그 타일 폴더(SD7), 순서는 원래대로. 넣기 가능 = `pasteBlock` |
| 터치(칸끌기-A) | 500ms = 들림(`lifted`, 체크 안 함) -> 8px 움직이면 끌기 / 안 움직이고 떼면 체크. `itemPress` 옵션 `hold: false`로 옛 길게 누르기 체크 끔 |
| 저장 | 순서 = 낙관 표시(`pending`, 새 목록이 오면 지움) + `orderNodes`, 실패 = 알림 + 무효화. 넣기 = `doPaste`(붙여넣기 길), Ctrl = 복사(칸Q4) |
| 모양(칸Q5) | `TransitionGroup`(`.dg-move` 180ms) · 빈 자리 = 불투명 + 점선 · 머묾 = `@property --dg-dwell` conic 테두리 600ms · 넣기 = 바탕·2px 테두리·1.05배 · 넣은 칸 = `dg-land`(줄며 사라짐, 목록이 다시 올 때까지·최대 3초) · 받은 칸·타일 = `.jelly-land`(`jelly.scss`, jelly-hover 재사용) · 작은 표시 = `TileDragChip`이 `cellChip`도 그림(외 N개, "→ 폴더 이름(에 복사)") |
| 삭제 | 폴더 칸 HTML5 끌기(`draggable`·dragover/drop)와 `nodePaste.hook`의 `dragging` 모듈 |
- 헤드리스 34/34(격리 5051·3001, td1 폴더 `칸시험` 243): 마우스 순서·앞으로·Esc·빈 곳 맨 뒤 / 머묾 테두리·넣기·표시 글자·빈 자리 원위치·사라짐·젤리·옮겨짐 / 머물기 전 놓기 = 안 넣음 / 가장자리 지나 가운데 머묾 / Ctrl 복사 / 다른 타일 넣기 / 터치 스와이프·들림·떼면 체크·움직이면 끌기. API 13건(형제 아님 400·공유 없음 404·보기 공유 403·편집 공유 가능·차단 형제 상대 위치 유지·넣기 끝 붙음)
- 스크립트(이 세션 scratchpad, 재부팅 때 사라짐): `/tmp/claude-1000/-home-jh-bae-irony-nexus/1dbb5e34-df83-47a1-b7aa-5d313a551b3b/scratchpad/cell/` `cells.mjs`(헤드리스) · `api.py`(API) · `board.py`(td1 보드 = 루트 + X(A) + C(칸시험) + 칸 되돌리기) · `seed.py`(칸 만들기, 이미 함)
- **함정(10-29 사용자 발견)**: f3 안 s5를 끌었다 놓은 뒤 칸시험으로 가면 s5가 맨 앞에 잠깐 보였다 사라짐(서버는 정상). 원인 = `TransitionGroup`이 빠지는 칸을 CSS 전환 시간(칸의 `scale` 150ms)만큼 남겨 둠 -> 폴더를 바꿀 때마다 옛 칸이 남음. 고침 = `:css="false"`(들어오고 빠지기는 바로, 비켜 가기 move만). 그 결과 넣기 응답이 빠르면 줄어드는 칸이 애니메이션 전에 빠져서 -> 넣는 동안 놓을 때 목록을 붙잡음(`held`, 요청 + 240ms 둘 다 끝나고 새 목록에서 빠질 때까지, 최대 3초). 재현 = scratchpad `cell/back.mjs`
- 시험 중 함정: 시험 배치에서 타일이 화면 밖·다음 페이지면 좌표 끌기가 빗나감(코드 문제 아님) -> 순서 `[루트, X, C]`, C = 0.5×1
- **함정(10-29 사용자 폰, 원래 있던 문제)**: 선택 모드 켜기·끄기(길게 누르기·⋮ "선택") 때 화면이 엉뚱한 타일로 감. 원인 = 선택 막대(45px)가 strip 높이를 줄여 `TILE_THRESHOLD` 높이 560을 넘나들면 `capacity` y가 2<->1로 바뀜 -> 배치·페이지 수가 바뀌는데 scrollLeft는 그대로. 헤드리스 400 x 640~670에서 재현(3 -> 4페이지, C 대신 X). 고침 = 안 1(막대 높이 빼고 판정): `tileGrids.barH`를 칸 수 판정 높이에 더함, 값은 `TileStrip` 크기 측정 때 같이 읽음(따로 재면 한 프레임 다른 배치). 재현 = scratchpad `cell/phone.mjs`(`H=` 화면 높이)
- 추가(10-29 사용자 폰): 칸 끌기에 **옆 페이지 넘기기**. 처음엔 흐르는 가로 스크롤(48px·바로 -> 화면 끝 칸에서 바로 넘어가 32px·0.4초)이었다가 아래 **끝 띠**로 바꿈. 옵션 `pager`(FolderTileBody가 `useTileGrids`의 `edge*`를 넘김). 폰 한 페이지 = 타일 하나라 이게 없으면 다른 타일 넣기를 못 함. 확인 = scratchpad `cell/phone3.mjs`(고른 2칸 -> 왼쪽 끝 머묾 -> 1페이지 X에 넣기)

## 끝 띠(페이지 넘기기) — 칸·타일 끌기 공용 (2026-10-29)
- 사용자 결정: 끝 띠 + "이전/다음 페이지" 글자, 페이드 인·아웃 / 띠가 끝 칸을 일부 가려도 됨 / 넘기기 = **칼럼 하나(0.5페이지)씩 한 번에** / 칸·타일 끌기 **전부 맞춤**. 전체보기는 칼럼이 없어 흐르는 스크롤 그대로
| 곳 | 내용 |
|---|---|
| `tileGrids` `edge`·`edgeStart`·`edgeTrack`·`edgeEnd` | `EDGE_BAND` 48px 안에 `EDGE_HOLD_MS` 400ms 머물면 `goCol(칼럼 ±1)`(부드러운 스크롤, 스냅 그대로), 계속 머물면 다시 차올라 한 칼럼 더. 넘길 쪽이 없으면 그 띠 없음(`prev`·`next`). 넘긴 칼럼은 따로 셈(스크롤 중 `view.col`은 옛 값) |
| `PageEdgeBands.vue`(TileWorkspace stage) | 끄는 동안만 양쪽 띠(그라데이션, 화살표 + 세로 글자) 페이드 인, 들어간 쪽은 진해지고 `fill`이 아래에서 위로 400ms 차오름(키 = 들어간 때). `pointer-events: none`(끌기 판정 `elementFromPoint`를 안 가리게) |
| `tileDrag.store` 평소 모드 | 옛 흐르는 스크롤·`scrubbing` 대신 `edgeTrack`, 스크롤이 움직이면 매 프레임 `judge` |
| `cellDrag.hook` | 같은 `edgeTrack`, 스크롤 움직이면 `judge` |
- 확인: 칸 폰 `cell/phone3.mjs`(띠 2개·왼쪽 차오름 -> 1페이지 X에 2칸 넣기) · 칸 회귀 34/34 · 타일 평소 `ui/drag_edge.mjs`(drag.mjs 복사, M6을 0.65초 머묾 = 한 칼럼 + 띠 표시·사라짐 E1·E2) 23/23 · 전체보기 `ui/ov.mjs` 15/15. 옛 `ui/drag.mjs` M6(1.8초 머묾)은 칼럼 4개를 넘겨 순서 기대가 달라짐 -> drag_edge.mjs 쓸 것
