# REF — 내 장비·관리자 판 타일 끌기 (파트(내 장비·관리자 타일 이동))

> 2026-10-30 신설. 이력 → `history/member-drag.md` / 열기 규칙(화-9) → `REF-member-ui.md` / 보드 끌기 원본 → `REF-tile-drag.md`·`REF-tile-drag-impl.md` / 끌기 기반 → `REF-util-drag.md` / 현재 진행 → `CURRENT.md`

**상태(2026-10-30): 1단계 열기 규칙 + 2단계 끌기 구현 끝, 헤드리스 통과. 사용자 PC·폰 확인 대기.**

## 번호-이름
| 번호(이름) | 내용 |
|---|---|
| 규칙-1(열기 규칙) | 화-9(자식 자리)를 보드 `openTile`과 같게 — 연 곳 바로 뒤(최신이 앞), 연 곳 없으면 **맨 앞** |
| 멤버-1(배치 계산) | `pack` 결과 슬롯 + 끌기 시작 때 잰 Grid 상자로 rect 계산. 타일 위 = **위 절반이면 앞·아래 절반이면 뒤**(2026-10-30 개정, 아래 "발견") |
| 멤버-2(끌기 연결) | `DragSession` + 미리보기 순서(`preview`) + 놓으면 배열 재정렬 |
| 멤버-3(모양) | 끄는 타일 자리 = 빈 자리, 손가락 옆 `TileDragChip` |
| 멤버-4(페이지 넘기기) | 끝 띠(칼럼 하나씩) — `PageEdgeBands` 공용화 + `makeEdgeBands` |
| 멤버-5(폰·예외) | 폰 같은 규칙(목록 칸 제외), 키보드 열림 = 막기, Esc 취소 |
| 멤버-6(확인) | 헤드리스 + 사용자 PC·폰 |

## 확정 (사용자, 2026-10-30)
| 항목 | 결정 |
|---|---|
| 같은 판 안 끌기 | **그 타일만**(보드와 같음) — 회원을 끌어도 그 아래 장비 타일은 제자리 |
| 자식 떼기 | 허용, `parent` 유지 — 떨어져 놓아도 경로·옅은 파랑 테두리 남음 |
| 시작 자리 | 머리 줄 + 몸통 빈 곳(터치) — 보드 `dragFrom`과 같음. 입력칸·버튼·장비 줄 제외 |
| 순서 저장 | **URL `t`에만**(서버 저장 없음, 사용자) — 새로고침·뒤로 가기 유지, 다른 기기엔 안 감 |
| 페이지 넘기기 | 끝 띠 재사용(칼럼 하나씩, 400ms 머묾) — 보드·칸 끌기와 같은 감각 |
| 범위 | `/devices`(내 장비)·`/admin`(관리자) 둘 다(같은 store) · 폰 목록 칸은 드래그 대상 아님 |

## 구현 (2026-10-30)
| 곳 | 내용 |
|---|---|
| `util/memberGrid.util.ts` | `insertTile` = 연 곳 바로 뒤 / 연 곳 없으면 맨 앞. `isDescendant` 삭제 |
| `util/memberDrop.util.ts`(새) | `memberDropOrder` — 판정은 `pack` 배치 계산 위치. 타일 위 = 위 절반이면 그 타일 앞 / 아래 절반이면 뒤 · 커서가 끄는 타일(빈 자리) 위면 그대로 · **빈 칸 = 그 칸 자리**(채우는 순서에서 그 칸보다 앞인 마지막 타일 뒤, 없으면 그 Grid 첫 타일 앞) · Grid 오른쪽 너머 = 맨 뒤. Grid가 늘어나면 잰 간격으로 이어 붙임 |
| `store/memberBoard.store.ts` | `preview`(미리보기 순서)·`shown` 추가 — `grids`·`posOf`가 미리보기를 씀(위치 번호도 따라 바뀜) |
| `store/memberDrag.store.ts`(새) | `DragSession`(holdMs 500) + `memberDragFrom`(시작 자리) + `memberChip`(아이콘·이름). 끌기 시작 때 Grid 상자 측정, 끝 띠 스크롤마다 다시 판정, 놓으면 `tiles = preview` + `activate` |
| `component/MemberTileView.vue` | `@pointerdown` 끌기 시작, `.slot`(빈 자리 — 내용 숨김, 불투명 바탕)·`.pressing`(터치 대기). **끌기 시작하면 크기 팝업을 닫는다**(열린 채 끌면 팝업이 버튼을 따라가 화면이 흔들림) |
| `component/MemberPagerPopup.vue` | 끌기 시작하면 컨트롤 패널·크기 카드도 닫는다 |
| `component/MemberPage.vue` | `provideMemberDrag`, `PageEdgeBands`(푸터 위 영역), `TileDragChip`, `.grid`에 `mb-grid` |
| `MemberUserTile`·`MemberDeviceTile`·`MemberNewTile` | 몸통에 `data-pager-press`(터치 시작 자리) |
| `PageEdgeBands.vue` | 보드 저장소 대신 **`edge` prop** — 보드·관리 페이지가 같이 씀 |
| `widget/util/edgeBand.util.ts`(새) | `makeEdgeBands`(띠 판정·차오름·칼럼 넘김) — `tileGrids`도 이걸로 교체(동작 그대로) |
| `TileDragChip.vue` | 보드 저장소 대신 **prop**(`chip`·`bundleIds`·`bundleChips`·`cellChip`) — `TileWorkspace`가 넘긴다 |
| `TileWorkspace.vue` | `:edge="grids.edge"`, 칩 prop 연결 |

## 확인 (2026-10-30 헤드리스, test1·test2·DB `nexus`, 계정 `irony`)
| 스크립트 | 결과 |
|---|---|
| `mdrag/check.mjs` | 7/7 — 열기 규칙(연 타일이 맨 앞, 3번), 끌기 미리보기·놓기, URL `t` 순서 일치, 새로고침 유지, 콘솔 오류 없음 |
| `mdrag/edge.mjs` | 6/6 — 타일 7개 = Grid 2개, 오른쪽 끝 띠 표시, 머물면 다음 칼럼(scrollLeft 440), 2페이지로 놓기, 띠 사라짐 |
| `mdrag/touch.mjs` | 7/7 — 폰 400×800 터치, 몸통 길게 누르기 → 시작 대기 → 끌기 → 순서 저장 |
| `mdrag/board3.mjs` | 6/6 — 보드 회귀(칩 prop·끝 띠 prop·`makeEdgeBands` 교체 뒤 끌기 그대로) |
- 스크립트 = 세션 scratchpad `mdrag/`(재부팅 때 사라짐). `playwright` = `~/.npm/_npx/e41f203b7505f1fb/node_modules` 심볼릭 링크로 씀
- 함정(확인 중): 폰은 스크롤 스냅 때문에 **보이는 타일끼리** 좌표를 다시 재야 한다(스와이프 뒤 옛 좌표로 끌면 빗나감 — 앱 문제 아님)
- ESLint는 이 저장소에 `eslint-config-vuetify`가 없어 못 돌림(vue-tsc는 통과)

## 발견·고침 (2026-10-30, 사용자 실기기)
- 증상(사용자): 장비 타일(세로 2칸) **하단쪽에 놓으면 배치가 안 되고, 원래 자리로 돌아오기도 힘들다**
- 원인: 보드에서 가져온 ③(옮긴 결과에서 끄는 타일이 커서를 덮을 때만 적용). 크기가 섞이면 앞/뒤 어느 쪽으로 옮겨도 끄는 타일이 커서 밑에 오지 않아 **아무 일도 안 일어남**. 게다가 앞/뒤 판정이 "지금 순서에서 앞이냐 뒤냐"라서 큰 타일 아래쪽 = 앞 후보만 남아 막힘
- 고침: 판정을 **커서가 그 타일의 위 절반이냐 아래 절반이냐**로. 옮긴 결과와 무관하게 바로 적용(관리 판은 애니메이션이 없어 흔들림이 없음). 빈 칸은 그 Grid 끝으로 가되 **끄는 타일이 이미 그 Grid 안이면 그대로**(옮긴 뒤 생긴 빈 칸에서 다시 옮기는 흔들림 방지)
- 확인 = `mdrag/halves.mjs`(1×1 / 1×2 / 1×2 배치): 아래 절반 → 뒤로, 위 절반 → 앞으로(되돌아옴). 기존 `check`·`edge`·`touch`·`board3`도 다시 통과

## 발견·고침 2 (2026-10-30, 사용자 실기기 2차)
- 증상(사용자): 세로 0.5 타일(h=1, 2행 중 한 칸)·1.0 타일(h=2)이 섞였을 때 **반 칸 타일 아래 빈 칸**으로 끌어가도 반응이 없음
- 원인: 빈 칸 규칙이 "그 Grid **끝**으로"였고, 끄는 타일이 이미 그 Grid 안이면 그대로 두는 안전장치가 있어 빈 칸이 죽은 자리가 됨. 게다가 `pack`은 앞으로만 채워서 그 빈 칸은 다음 타일이 채우지 못하고 계속 빈다
- 고침: 빈 칸 = **그 칸 자리**에 넣는다. 채우는 순서(`c*rows + r`)에서 그 칸보다 앞인 마지막 타일 뒤로 `moveAfter`(없으면 그 Grid 첫 타일 앞). 반 칸 타일 아래 빈 칸 → 그 타일 바로 뒤 → `pack`이 그 칸에 넣는다
- 확인 = 단위 테스트(scaffold `mdrag/unit.ts`, tsc로 컴파일해 실행) 5/5 — 빈 칸 자리·1×2 위/아래 절반·자기 자리·Grid 너머 + 브라우저 `halves` 4/4 · `edge` 6/6 · `touch` 7/7 · `board3` 6/6 · `check` 7/7

## 발견·고침 3 (2026-10-30, 사용자)
- 증상(사용자): **크기 조절 패널이 열린 채로 타일을 끌면 페이지가 움직인다** -> 제안 "끌면 다른 패널들이 모두 닫히면 문제 없지 않을까"
- 고침: 끌기 시작(`dragKey`)에 `MemberTileView`의 크기 팝업(`v-model` → `sizeOpen`)과 `MemberPagerPopup`의 패널·크기 카드를 닫는다. 헤드리스로 화면 이동 자체는 재현 안 됨(열린 채 끌기·놓기 모두 스크롤 0) — 닫는 쪽으로 정리
- 확인 = `mdrag/menu2.mjs` 4/4(팝업 열림 -> 끌기 시작 -> 닫힘 -> 놓아도 안 열림), `check` 8/8 · `halves` 4/4 · `touch` 7/7 재통과
- **보드도 같은 규칙으로 맞춤(사용자 10-30)**: `TileFrame` 크기 팝업·⋮ 메뉴에 `v-model`(`sizeOpen`·`menuOpen`), `drag.dragId`가 잡히면 닫는다. 보드는 원래 화면 이동이 없었지만 팝업이 남는 건 같았음 -> `mdrag/boardmenu2.mjs` 7/7 · `board3` 6/6 재통과 → `REF-tile-drag-impl.md`

## 남은 것
- 사용자 PC·폰 실기기 확인(폰 Android 길게 누르기 — 보드 때와 같은 위험, `REF-util-drag.md` 함정)
- 회원 타일 아래 장비 타일이 붙는지(관리자 → 회원 타일 → 장비 줄)는 dev DB에 장비가 없어 미확인
- 경로 재열림(`openParent` → `insertBefore` "그 자식 바로 앞")은 옛 동작 그대로 — 자식이 여럿이면 부모가 자식 묶음 중간에 낄 수 있음(뜻이 안 맞는지 다음에 볼 것)
