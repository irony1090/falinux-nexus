# REF — 타일 이동 규칙 통일 · 커서 위치 페이지 표시

> 2026-10-21 신설(파트(페이지 이동)). 이력 → `history/tile-move.md` / 타일 위치 규칙의 원래 결정 → `REF-board-pager.md` "확정" / 확인 스크립트 → `REF-test-scripts-1020.md` "2026-10-21 페이지 이동" / 현재 진행 → `CURRENT.md`

**상태: 이동-1(타일 위치 규칙 통일)·이동-2(커서 위치 페이지 표시) 구현 + 격리 헤드리스 28/28 + 사용자 확인(2026-10-21). 이동-3(타일별 단축키 끔) 구현 + 17/17 + 이동-1·2 재확인 28/28 + 사용자 확인. 파트(페이지 이동) 마감·커밋 → `INDEX-tile-move.md`.**

## 번호-이름
| 번호(이름) | 내용 |
|---|---|
| 이동-1(타일 위치 규칙 통일) | 타일로 가는 모든 경로가 인디케이터(GridPager)과 같은 "타일 위치 규칙"을 따름 |
| 이동-2(커서 위치 페이지 표시) | Ctrl+Alt를 누르는 동안 마우스 커서 위치에 인디케이터 하나 더, 같은 조작 |
| 이동-3(타일별 단축키 끔) | 터미널 타일마다 옵션: 그 타일이 active인 동안 앱 단축키 전체를 끔 |

## 원칙 (사용자, 2026-10-21)
- **타일로 이동할 때는 기본적으로 페이지 표시의 규칙을 따른다.** 새 이동 경로를 만들 때도 `reveal`(= `revealCol`)을 쓴다. Grid 단위(`goGrid`)는 도트 클릭처럼 "Grid로 가기"가 목적일 때만

## 이동-1(타일 위치 규칙 통일)
- 원인: 이동 경로가 두 갈래였음 — 인디케이터 드래그 = `stepTo` -> `revealCol`(칼럼 단위, 막 다 보이는 만큼) / 나머지 = `activate` -> `reveal` -> `goGrid`(Grid 시작 = 페이지 단위)
- 고친 것: `tileGrids.store.ts` `reveal`이 `goGrid` 대신 `revealCol`로 계산. 기준 칼럼 = `stripView(geometry, el.scrollLeft, el.clientWidth)`(scrollX는 늦게 갱신되므로 요소에서 직접)
- 영향 받는 경로(전부 같은 규칙이 됨): Ctrl+Alt+숫자 · 폴더 열기(`nodeOps.open`) · 서버가 넣은 타일(exec) · RunRow · URL `?focus` · 크기 변경(`TileFrame`) · 새로고침 복원(view가 없을 때)
- 전체보기에서 타일 고름: `leaveOverview(tileId)` = 들어가기 전 화면으로 돌아간 뒤 그 기준으로 `reveal`(전엔 그 Grid 시작). 사용자가 따로 답하진 않았고 위 원칙으로 적용 — 확인 때 볼 것

## 이동-2(커서 위치 페이지 표시) — 결정
| 항목 | 결정 |
|---|---|
| 언제 뜨나 | **Ctrl+Alt를 누르자마자**(사용자 선택). Shift·Meta가 같이 있으면 안 뜸. 터미널 프로그램의 Ctrl+Alt+키·Windows AltGr에서도 잠깐 뜨는 건 감수 |
| 언제 닫히나 | Ctrl이나 Alt를 떼면. **드래그 중이면 손을 뗄 때까지 유지**. 창 blur에도 닫힘(Ctrl+Alt+화살표 창 전환에서 keyup을 못 받음) |
| 위치 | 알약 가운데 = 커서(가로·세로). 가장자리면 화면 안쪽 16px까지 밀어 넣음. 커서를 모르면(창에 안 들어옴·창 밖으로 나감) 화면 가운데. 뜬 뒤엔 커서를 따라가지 않음 |
| 타일 선택 | **위·아래 둘 다 당겨서**(사용자 선택). 타일 패널은 당긴 쪽에. 푸터 인디케이터은 그대로 위로만(아래는 화면 밖) |
| 안 띄움 | 폰(`pointer: fine` 아님) · 전체보기 중 |

## 이동-2 구조
- `PagerPopup.vue`(새 파일): body로 teleport, **늘 마운트하고 `visibility`만 바꿈** — 폭 측정(`useResizeCallback`)·드래그 상태를 유지하려고. `TileNav`의 평소 분기 안에 둠(= 전체보기·Grid 없음이면 마운트 안 됨)
- 포커스: 상자에 `@mousedown.prevent` -> 도트·드래그가 xterm 포커스를 가져가지 않음. 타일 이동으로 끝나면 `usePagerDrag.up`이 그 타일에 포커스(기존 동작)
- 클릭 통과: 상자(폭 560 고정)는 `pointer-events: none`, 알약(`.FootPill`)만 받음
- `usePagerDrag(pull)`: `'up'`(푸터) | `'both'`(팝업). `pullDir`로 들어간 쪽을 기억해 되돌림(`TILE_BACK`)도 그쪽 기준
- `GridPager` `pull` prop + `defineExpose({ mode })`(팝업이 드래그 중인지 봄) / `TilePeek` `below` prop(아래에 띄울 때 `transform-origin` 위)
- Ctrl+Alt+숫자(`tileShortcut.hook.ts`)와는 리스너가 따로 — 팝업이 떠 있는 채로 숫자를 누르면 바로 이동하고 팝업은 유지

## 남은 것 / 볼 것
- 사용자 실기기 확인(PC Chrome·Edge): Linux 창 관리자가 Alt+드래그를 창 이동으로 가로채는지(XFCE 등 Alt 기본인 환경) — 가로채면 Ctrl+Alt를 뗀 뒤 드래그(드래그 시작 전엔 닫히므로 지금은 안 됨) 방법을 다시 정해야 함
- 확인용 스크립트 = irony `test`(14) 아래 `ctrl-alt-test`(225, 2026-10-21 REST 등록): raw 모드로 받은 키 바이트(hex)와 해석을 찍음, `q`로 끝. 본문 원본 = DB `nodes.content`

## 이동-3(타일별 단축키 끔) — 계획·구현 (2026-10-21)
- 목적: 터미널 프로그램이 앱 단축키 조합(지금은 Ctrl+Alt+숫자·Ctrl+Alt)을 써야 할 때 그 타일에서만 앱이 가로채지 않게
- 원칙: **앱 단축키는 전부 검사 한 곳(`appShortcutsOff`)을 거친다** — 앞으로 단축키가 늘어도 자동으로 같이 꺼지게(MEMORY 불변 결정)

| 키 | 성격 | 끄면 |
|---|---|---|
| Ctrl+Alt+숫자 (`tileShortcut.hook.ts`) | 앱 단축키 | 꺼짐 -> 터미널로 감 |
| Ctrl+Alt 팝업 (`PagerPopup.vue`) | 앱 단축키 | 꺼짐(안 뜸) |
| Ctrl+C 복사 · Ctrl+V 붙여넣기 (`processTerm.store.ts` I4·I5) | 터미널 자체 동작 | 그대로 |
| Esc(전체보기·노드 선택) · 에디터 Ctrl+S · 인디케이터 ←→ | 터미널 active와 무관 / 그 요소 포커스 때만 | 영향 없음 |

| 단계(이름) | 내용 |
|---|---|
| 끔-1(서버 필드) | `router/tile.go` `tileItem`에 `ShortcutsOff bool json:"shortcutsOff,omitempty"` — 구조체가 필드를 정해 두므로 빠뜨리면 저장 때 사라짐 |
| 끔-2(트리 동작) | `tileTree.store.ts` `TerminalTile.shortcutsOff?: true` + `shortcutsOffOp`(`minimizeOp`과 같은 모양) + `setShortcutsOff` |
| 끔-3(검사 한 곳) | `tileGrids.store.ts` `appShortcutsOff` = active 타일이 터미널이고 `shortcutsOff`. `tileShortcut.hook`·`PagerPopup`의 keydown 맨 앞에서 확인 |
| 끔-4(메뉴·아이콘) | `TileFrame` ⋮ 메뉴(지금 폴더 타일만 보임)를 터미널 타일에도 -> "단축키 끄기/켜기" 토글. 꺼져 있으면 머리에 `mdi-keyboard-off` 아이콘 + title 설명. 볼 수 없는 타일(`unavailable`)은 안 보임 |
| 끔-5(확인) | 격리 헤드리스: 끈 타일 active -> Ctrl+Alt 팝업 안 뜸 · Ctrl+Alt+숫자가 xterm으로 감 / 다른 타일 active -> 그대로 동작 / 새로고침·다른 탭에서 유지(트리 저장) |

- 저장 단위 = 타일(타일 트리, 계정 안 탭·기기 동기화). 스크립트 노드 단위(DB 컬럼)·탭 메모리 단위는 안 함
- 꺼진 타일에서 나가기 = 마우스 클릭·푸터 인디케이터(키보드로는 못 나감, 아이콘으로 알림)

### 이동-3 구현 결과 (2026-10-21)
- 계획 그대로 끔-1~끔-4 구현. ⋮ 메뉴 버튼 title "보드" -> "메뉴"(터미널 타일에도 보이므로). 메뉴 항목 = "단축키 끄기/켜기" + 부제 "이 타일이 active일 때 Ctrl+Alt 단축키", 머리 아이콘 `.keys-off`(`mdi-keyboard-off-outline`)
- 끈 타일에서 Ctrl+Alt+3 -> 터미널에 `1b 33`(xterm.js가 Ctrl+Alt+숫자를 Alt+숫자로 보냄)
- 격리 헤드리스 `move/keys.mjs` 17/17: 메뉴·아이콘 / 서버 트리 저장 / 끈 타일 active = 팝업 안 뜸·숫자 이동 안 함·터미널로 감 / 다른 타일 active = 그대로 / 새로고침·다른 탭 유지 / 다시 켜기 실시간 반영 / 루트 폴더 메뉴 없음 그대로
- **주의: 개발 supervisor(8012, `go run`)는 새 필드를 모르는 옛 빌드면 저장 때 `shortcutsOff`를 버린다 -> 재시작 필요**
