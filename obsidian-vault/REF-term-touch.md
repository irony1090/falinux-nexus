# REF — 터미널 터치 스크롤

> 2026-10-19 신설. 이력 → `history/term-touch.md` / 모바일 키보드·뷰포트(반복 주제) → `REF-mobile-viewport.md` / PgUp·PgDn 버튼 → `INDEX-term-page-keys.md` / 현재 진행 → `CURRENT.md`

**상태: 마감(2026-10-19) → `INDEX-term-touch.md`. 안 가(스크롤바 터치)·안 나(쓸어서 스크롤) 둘 다 구현·헤드리스 확인·사용자 실기기 확인·커밋.**

## 증상 (2026-10-19 사용자)
- 터치에서 터미널 스크롤바 드래그가 제대로 안 됨(Windows PowerShell 타일에서 발견 — 스크롤백이 생기는 셸이라 처음 눈에 띔. 원인은 브라우저 쪽이라 Linux도 같음)
- 터미널 화면을 손가락으로 쓸어 스크롤하는 것도 원래 안 됨

## 원인 (xterm.js 6.0.0 소스 확인)
| 확인 | 내용 |
|---|---|
| 스크롤바 드래그 | VS Code `abstractScrollbar` — slider `pointerdown` -> `GlobalPointerMoveMonitor`(`setPointerCapture` + `pointermove`). pointer 이벤트라 **터치도 이미 받음** -> 터치 -> 마우스 변환 불필요 |
| 끊기는 이유 | 스크롤바에 `touch-action`이 없음(`auto`) -> 손가락이 움직이면 브라우저가 팬(페이지 스크롤)으로 가져가며 `pointercancel` -> 이후 `pointermove` 없음. monitor는 `pointercancel`을 안 들음 |
| 숨김 상태 | 스크롤바는 평소 `.invisible { opacity: 0; pointer-events: none }`, 마우스 over·스크롤 때 `.visible`. 터치엔 hover가 없어 **한 번 탭해야 나타남**(탭의 호환 mouseover로). 숨김 중 터치는 아래 화면으로 감 |
| 쓸어서 스크롤 | xterm 6 스크롤 영역은 overflow 숨김 + 자체 스크롤바라 브라우저 기본 터치 스크롤이 안 됨. VS Code `Gesture`(touch.ts)는 번들에 있지만 터미널 스크롤에 연결 안 됨 |

## 안 가(스크롤바 터치) — 구현
- `ProcessTerminal.vue` scoped style: `.ProcessTerminal :deep(.xterm-scrollable-element > .scrollbar) { touch-action: none; }`
- 헤드리스 확인(Chromium 터치 에뮬레이션, 독립 페이지 = xterm 6 + 300줄, CDP `Input.dispatchTouchEvent`로 slider 위 200px 위로 끌기, 탭으로 스크롤바 먼저 띄움):

| 조건 | 수정 전 | 수정 후 |
|---|---|---|
| 페이지 스크롤 없음 | viewportY 285 -> 267(몇 줄 뒤 끊김), `pointercancel` | 285 -> 105, `pointercancel` 없음 |
| 페이지가 김 | 285 -> 267 + **페이지가 202px 스크롤됨** | 285 -> 105, 페이지 안 움직임 |
| 탭 안 하고 바로 | 스크롤바 숨김(`pointer-events: none`) -> 화면이 터치를 받음, 둘 다 변화 없음 | 같음 |

- 남은 불편: 스크롤바가 탭 전엔 안 보임(탭 = xterm 포커스 -> 모바일 키보드도 열림) · slider 폭 14px로 손가락엔 좁음 -> 필요하면 `@media (pointer: coarse)`에서 항상 보이게·넓게 하는 안
- 확인 스크립트 → `REF-test-scripts-1014.md` "2026-10-19 터치 스크롤"
- 실기기 확인용 노드(irony 계정, REST `POST /nodes`로 등록 2026-10-19): Linux = 폴더 `test`(14) 아래 `scroll-test`(215, bash 300줄 후 `exec bash`) / Windows = 폴더 `win`(213, `win-dev`) 아래 `scroll-test.ps1`(216, 300줄 후 `powershell -NoLogo`). 둘 다 셸로 남아 `exit`로 종료. 필요 없어지면 노드 삭제

## 안 나(쓸어서 스크롤) — 검토 → 구현
- 충돌 검토(2026-10-19, 코드 조사만):

| 겹치는 것 | 위험 | 대응안 |
|---|---|---|
| 탭(포커스 -> 키보드 열림) | 낮음 | 움직임 약 10px 넘으면 쓸기로 판정, 아니면 탭 그대로 |
| 크기 고정 화면(크기 소유 아님, `ProcessTerminal--fixed` = `overflow: auto`) | 중간 — 지금은 쓸기로 타일보다 큰 터미널을 이동해 봄. 세로 쓸기를 가져가면 이동이 막힘 | 상자 스크롤 먼저 -> 끝에 닿으면 스크롤백(브라우저 중첩 스크롤과 같은 순서) |
| 전체보기(미리보기 터미널, `TileStrip` 가로 스크롤) | 중간 — 미리보기에서 쓸면 줄이 움직여야 함 | 전체보기 중엔 끔 |
| 전체 화면 앱(htop·vim, 스크롤백 없음) · 마우스 모드 앱 | 정할 것 | PC 마우스 휠과 같게(xterm이 alt 화면이면 화살표 키, 마우스 모드면 앱에 휠 보고) 또는 아무것도 안 함 |
| 모바일 키보드 열림(배치 고정) | 낮음 — 크기 안 바꿈. 터미널 위 `touch-action`이 문서 팬을 막는 정도 | iPhone 실기기 확인 때 같이 |
| 관성(손 떼면 미끄러짐) | 없으면 딱딱함 | 선택 |

- 구현 방식 후보: 가) 정상 화면에서만 `term.scrollLines(n)` / 나) 쓸기 -> 합성 `WheelEvent`를 xterm 요소에 보냄 = PC 휠과 같은 동작이 모든 모드에 그대로(추천)
- 화면 touchmove 세로 거리 -> `term.scrollLines(n)`
- 전체 화면 앱(alt 버퍼, htop·vim)은 스크롤백이 없으니 기본안 = 아무것도 안 함
- 탭(키보드 열기)·길게 눌러 선택은 그대로 / 모바일 키보드 규칙(열린 동안 배치 고정, `REF-mobile-viewport.md` 4)과 충돌 확인

### 안 나 구현 (2026-10-19, 사용자 "진행")
- 결정(추천대로): 크기 고정 화면 = **상자 먼저, 끝에 닿으면 터미널** / 관성 **넣음**. 전체보기는 이미 `TileFrame--overview`의 `pointer-events: none`으로 터치가 안 옴(따로 끌 필요 없음)
- `feature/process/util/touchScroll.util.ts` `bindTouchScroll(el, getTerm, scrollBox)`: 1손가락·`.scrollbar` 밖에서만 / 10px(`TAP_SLOP`) 전엔 탭으로 둠(포커스·키보드) / 가로가 더 크면 브라우저에 맡김(`touch-action: pan-x`) / 쓸기 중 `preventDefault`(뒤따르는 click 막음) / 관성 = 마지막 100ms 속도, 16ms당 0.95
- 넘기는 곳: 정상 화면 = `term.scrollLines`(px를 셀 높이로 나눠 쌓음) / alt 화면·마우스 모드 = 합성 `WheelEvent`(xterm 코어가 `deltaY`로 화살표 키·휠 보고)
- **함정(재발 방지)**: 정상 화면에 합성 휠 이벤트는 안 먹음 — VS Code 스크롤 요소(`StandardWheelEvent`)가 옛 `wheelDeltaY`를 읽고, 스크립트로 만든 이벤트는 그 값이 늘 0
- 크기 고정 상자: `scrollTop` 정수 반올림 찌꺼기가 스크롤백으로 새지 않게 상자가 끝에 닿았을 때만 남은 양을 넘김 / 크기 소유(`overflow: hidden`)일 땐 상자를 안 건드림(hidden이어도 `scrollTop`은 바뀌어 화면이 밀림)
- `ProcessTerminal.vue`: `onMounted` 연결·`onBeforeUnmount` 해제, `.ProcessTerminal { touch-action: pan-x }`
- 헤드리스 11/11(`touch/swipe.html`·`swipe.mjs`): 손가락 아래 = 스크롤백 위 / 위 = 아래 / 페이지 안 움직임 / 쓸기는 포커스 안 함 / 탭은 포커스 / 관성 / 가로 쓸기 무시 / 스크롤바 드래그 유지 / 크기 고정 = 작은 쓸기는 상자만, 상자 끝 뒤 스크롤백 / alt 화면 = 화살표 키
- 확인 함정: 테스트 상자보다 터미널이 크면 스크롤바 slider가 잘린 영역에 있어 터치가 안 닿음(실제 앱은 fit으로 맞춤)
