# REF — 모바일 가상 키보드 구현 (파트(모바일 가상 키보드), 2026-10-14 마감)

> 결정(A+B)·검토 근거 → `REF-node-ui-overview.md` "모바일 가상 키보드" / 이력 → `history/mobile-keyboard.md` / 입력 배선 → `REF-process-input.md` / 색인 → `INDEX-mobile-keyboard.md`
> **프로젝트와 무관한 지식(뷰포트 메타·열림 판정·브라우저별 실측·재배치 막기)은 반복 주제 `REF-mobile-viewport.md`에** — 이 파일은 Nexus 구현만

## 마감 상태 (2026-10-14)
- 최종 구조 = 키보드-0~4 + 고침-나(커서 기준 이동) + 라(resizes-content 전환). 확인 = 헤드리스 `kb.mjs` 22 · `kb2.mjs` 10 · `kb3.mjs` 12 + 사용자 Fold 8 크롬·Edge 사진 확인
- **Edge 제약**: 폰 크기 기본 화면(하단 막대)에서 키 줄 아래 빈틈 약 34(브라우저 값이 틀림, `REF-mobile-viewport.md` 3). 전체 화면이면 정확
- VK4(사용자 키 추가) = **안 함**(탭·누른 채·한글 자모 모두 동작해서 불필요)
- 측정용 `KbDebug.vue` 삭제. 폰 사진 업로드 도구는 `vite.config.mts`에서 주석 처리해 남김(→ `REF-test-env.md` "폰 사진 올리기")
- 남은 것: iPhone 실기기 확인(나중) / 별건 = 전체 화면을 켜면 같은 칼럼이 `single` -> `row`로 바뀜(CURRENT 미룬 항목)

## 시작 시점 코드 상태 (2026-10-14 확인)
- 결정만 있고 **코드 없음**: `index.html` viewport 메타에 `interactive-widget` 없음 / 특수 키 줄 없음 / visualViewport를 듣는 곳은 `appWindown.store.ts`(`size.inner`)뿐(`window` `resize`만 들음)
- 타일 영역 높이 = `TileWorkspace.vue` `calc(100dvh - var(--v-layout-top))` -> Android 기본(`resizes-content`)은 키보드가 뜨면 레이아웃 뷰포트가 줄어 `dvh`도 줄어듦 -> 타일 재배치(검토 기록 ①). iOS는 키보드가 겹쳐서 안 줄어듦
- 입력 경로: `processTerm.store.ts` `sendInput(uid, bytes)` -> `PROCESS:INPUT`. xterm `onData`·`onBinary`가 부름

## 구조안 (2026-10-14 제시, 코드 없음)
| 번호(이름) | 내용 | 파일 |
|---|---|---|
| 키보드-0(뷰포트 메타) | viewport 메타에 `interactive-widget=resizes-visual` -> Android도 키보드가 겹침 -> `100dvh` 그대로, 타일 재배치 없음(A) | `index.html` |
| 키보드-1(키보드 상태) | visualViewport `resize`·`scroll`을 듣는 모듈: 열림(레이아웃 높이 - 보이는 높이 > 기준), 보이는 영역 위·아래 끝, 포커스된 터미널 uid(xterm textarea `focusin`·`focusout`) | 새 `process/hook/virtualKeyboard.hook.ts`(컴포넌트 밖 reactive 모듈) |
| 키보드-2(가려짐 이동) | 포커스된 터미널 타일 아래 끝이 (보이는 영역 아래 끝 - 키 줄 높이)보다 아래면 타일 영역 전체를 그만큼 `translateY`. 배치·PTY 크기 불변(A, 검토 ③ = 유지) | `TileWorkspace.vue` |
| 키보드-3(특수 키 줄) | 키보드 바로 위 한 줄(B). 위치 = `position: fixed; top = offsetTop + height - 줄 높이`. 버튼 `pointerdown` 기본 동작 막음(포커스 유지). 키 -> `sendInput(포커스 uid, 바이트)`. 방향키는 xterm `modes.applicationCursorKeysMode`에 따라 `ESC O A` / `ESC [ A`. Ctrl·Alt = 토글, 다음 글자 하나에만 적용(스토어 `onData` 경로에서 변환: Ctrl = `& 0x1f`, Alt = 앞에 ESC) | 새 `process/component/TerminalKeyBar.vue` + `processTerm.store.ts` |
| 키보드-4(iOS 문서 스크롤) | 앱 셸 문서 스크롤 막기(`overscroll-behavior`, 필요 시 `overflow: hidden`) + `env(safe-area-inset-bottom)`. 키 줄 위치는 `offsetTop` 기준이라 iOS 밀어 올림에도 따라감. 실기기 확인은 나중(iPhone 없음) | 전역 스타일 |
| 키보드-5(확인) | 헤드리스: 메타·키 줄 전송 바이트(Esc 0x1b·Ctrl+C 0x03·방향키 두 모드)·포커스 유지. 가상 키보드 자체는 헤드리스 재현 불가 -> **사용자 Android 실기기**(test1.falinux.com): 재배치 없음·가려짐 이동·키 줄 위치·한글 입력 | — |

## 결정 (VK1~VK3 = 2026-10-14 사용자 "구현부터 시작"으로 권장안 채택, VK4 = 나중)
| 번호(이름) | 질문 | 권장 |
|---|---|---|
| VK1(키 목록) | 키 줄에 넣을 키 | `Esc` `Tab` `Ctrl` `Alt` `←` `↑` `↓` `→` `\|` `~` `/` `-` (Termux 기본 줄에서 한 줄로 줄임, 폰 세로 폭 약 380px에 12개) |
| VK2(표시 조건) | 키 줄을 언제 보이나 | 터미널 포커스 + 키보드 열림 + 조작 권한(보기 전용이면 숨김) + 실행 중(PENDING·종료 숨김) |
| VK3(Ctrl 동작) | 키 줄 Ctrl을 누른 채로 폰 가상 키보드 글자를 누를 수 있나(2026-10-14 사용자 질문) | 누른 채 = 가상 키보드는 별도 창(IME)이라 두 번째 손가락 터치가 그쪽으로 가는 동안 우리 버튼의 누름이 유지되는지 기기 따라 다름(`pointercancel` 가능) -> 확인 전 단정 못 함. **둘 다 지원**: 탭 = 다음 글자 하나(한 번), 누른 채 글자 입력 = 뗄 때 해제(휴대폰 Shift 방식). 판정 = 누르는 동안 글자가 들어왔는지. 실기기에서 누른 채 방식이 되는지 키보드-5에서 확인. 한글 자판 상태면 `c` 대신 `ㅊ`이 조합 입력으로 옴 -> Ctrl·Alt가 걸린 동안은 두벌식 자모를 같은 자리 영문으로 바꿔 보냄(ㅊ->c) |
| VK4(사용자 키 추가) | 사용자 제안(2026-10-14): 누른 채 방식이 안 되면 사용자가 키 조합 버튼을 직접 추가 | 탭 토글(VK3)만으로도 모든 조합이 가능(Ctrl 탭 -> c, 두 번 누름) -> 사용자 키는 **필수가 아니라 자주 쓰는 조합을 한 번에 누르는 편의**. 버튼 = 이름 + 보낼 것(조합 `Ctrl+C` 또는 문자열 `q` + Enter). 저장 = 기존 `userPrefs.util`(`nexus.user.*`, 브라우저별 — 폰마다 따로, 로그아웃 때 삭제). 폭이 넘치면 키 줄 가로 스크롤. **순서 = 키보드-5(확인)에서 실제로 불편한 조합이 보이면 키보드-6(사용자 키)으로 추가**(처음부터 넣지 않음) |

## 구현 (2026-10-14)
| 번호(이름) | 실제 구현 |
|---|---|
| 키보드-0(뷰포트 메타) | `index.html` viewport에 `interactive-widget=resizes-visual` -> **라-1(메타 변경)로 `resizes-content`로 바뀜 ↓** |
| 키보드-1(키보드 상태) | `common/hook/virtualKeyboard.hook.ts`: 모듈 상태 `{open, top, bottom, termUid}` + 사용하는 컴포넌트 수로 리스너 붙임·뗌. 열림 = `레이아웃 높이 - vv.height * vv.scale > 120`(핀치 확대 제외). 포커스 터미널 = `focusin`·`focusout`에서 `[data-term-uid]` 조상(`ProcessTerminal.vue`에 속성 추가) |
| 키보드-2(가려짐 이동) | `TileWorkspace.vue`: 열림·위치·포커스가 바뀌면 `nextTick` 뒤 그 터미널의 `.TileFrame` 위치를 재서 `translateY(-shift)`. shift = min(아래 끝 - (키보드 위 끝 - 키 줄 높이), 위 끝 - max(보이는 영역 위 끝, 헤더 아래)) — 타일 위 끝은 헤더 밑으로 안 들어감. 0.15초 전환. **고침-나(커서 기준 이동)로 바뀜 ↓** |
| 고침-나(커서 기준 이동) | 실기기에서 칼럼 모드 전체 높이 타일은 위 규칙으로 이동 0 -> 커서 줄이 키보드에 가려짐(이상-나(아래 가려짐)). `TileWorkspace.vue` `measure`: s = max(같은 터미널이면 이전 이동, 위 타일 규칙) -> 커서 줄 아래가 키 줄 위 끝보다 아래면 올림, 커서 줄 위가 헤더 아래보다 위면 내림 -> [0, 타일 아래 끝 - 키 줄 위 끝]으로 자름. **타일 머리가 헤더 밑으로 들어가도 됨**(헤더가 위에 그려짐). 커서 줄 = xterm `buffer.baseY + cursorY - viewportY`(화면 밖이면 무시) x `.xterm-screen` 높이 / rows. 이전 이동을 기준으로 삼아 커서가 보이는 띠 안에서 움직이면 그대로(htop처럼 커서가 뛰어도 덜 흔들림). 키보드가 열린 동안 그 터미널 `onCursorMove`·`onScroll`·`onResize`를 100ms 묶어 다시 잼. 전환 중 측정은 `getComputedStyle(root).transform`의 `m42`로 지금 이동량을 빼서 올리기 전 좌표로 계산 |
| 키보드-3(특수 키 줄) | `process/component/TerminalKeyBar.vue`(높이 40, `KEY_BAR_HEIGHT` 내보냄) — `TileWorkspace` 안이지만 올라가는 div **밖**에 둠(transform 안이면 fixed 기준이 바뀜). 버튼 = `div role=button` + `@pointerdown.prevent`(포커스 유지). 보이는 조건 VK2. Ctrl·Alt = `process/store/keyModifier.store.ts`(모듈 상태): 누름 `pressMod` / 뗌·`pointercancel` `releaseMod`, 글자 = `applyMods`(한 글자만, 두벌식 자모 -> 영문, Ctrl = `& 0x1f`·특수 `@[\\]^_/-?`, Alt = 앞에 ESC). 스토어 xterm `onData`가 `applyMods`를 거침. 방향키 = Ctrl·Alt 있으면 `ESC [1;<n>X`, 없으면 xterm `modes.applicationCursorKeysMode`에 따라 `ESC O X`/`ESC [ X`. 스토어에 `sendText`·`appCursor` 추가 |
| 키보드-4(iOS 문서 스크롤) | `App.vue` 전역 `html, body { overscroll-behavior: none }`만. 키 줄 위치가 `offsetTop` 기준이라 iOS 밀어 올림에 따라감. 그 이상(문서 스크롤 되돌리기)은 iOS 확인 때 |

## 키보드-5(확인) 헤드리스 결과 (2026-10-14, 격리, 22/22)
- 방법: 390x844 `isMobile`·`hasTouch` + `addInitScript`로 `window.visualViewport`를 가짜 EventTarget으로 바꿔 `__kb(높이)`로 키보드 열림 흉내(코드에 시험용 장치 없음). 시험 스크립트 = python `tty.setraw` 후 받은 바이트를 `[hh]`로 출력, `M`/`N`이면 커서 키 앱 모드 켬/끔. 바이트는 `GET /processes/snapshot/:uid`로 읽음
- 결과: 메타 · 키보드 닫힘이면 키 줄 없음 · 열림이면 키보드 바로 위(아래 끝 = 키보드 위 끝) · 아래쪽 타일 344px 올라가 아래 끝 = 키 줄 위 끝 · Esc·Tab · 키 줄 눌러도 포커스 유지 · Ctrl 탭 -> c = 03, 다음 c = 63 · 누른 채 c·d = 03 04, 뗀 뒤 63 · 누름 중 `pointercancel` -> 다음 c = 03 · 두 번 탭 = 해제 · Alt+x = 1b 78 · Ctrl+ㅊ(`insertText`) = 03 · 방향키 보통·앱 모드 · Ctrl+→ = `ESC[1;5C` · `| ~ / -` · Ctrl 걸림 표시 · 키보드 닫힘 = 키 줄 사라짐·타일 제자리 · 터미널 밖 포커스 = 사라짐 · 보기(10) 공유 계정 = 키 줄 없음
- 헤드리스로 못 보는 것: 실제 가상 키보드 높이·`interactive-widget` 효과, 누른 채 방식(실제 두 손가락), 한글 조합(IME는 조합 중엔 `onData`가 안 올 수 있음 — `insertText`는 조합 없이 바로 들어감), 이 PC 스크린샷에서 `↑`·`↓`·`|`가 색 있는 글꼴로 보임(글꼴 문제로 추정)

## 라(resizes-content 전환) (2026-10-14)
- 왜: Edge 폰 크기 기본 화면(하단 막대가 키보드 때 얇은 주소 줄로 줄어듦)에서 브라우저가 주는 모든 값(visualViewport·레이아웃·키보드 API·`env(keyboard-inset-*)`·svh/dvh/lvh)이 줄어든 막대를 반영 안 함. 그중 resizes-content의 오차가 가장 작음(81 -> 약 34) + 다른 웹 터미널(sessile·term-server)도 쓰는 방식. 측정·조사 경위 -> `history/mobile-keyboard.md` 2026-10-14(5)

| 번호(이름) | 구현 |
|---|---|
| 라-1(메타 변경) | `index.html` viewport `interactive-widget=resizes-content` — 키보드가 뜨면 레이아웃이 키보드 위까지 줄어듦. iOS는 이 값을 무시(보이는 영역만 줄어듦) -> 같은 훅 식으로 처리 |
| 라-2(높이 고정) | `TileWorkspace.vue`를 바깥(`height: calc(100dvh - 헤더)`, `overflow: clip`) + 안쪽 `__body`(transform·키보드 열린 동안 `height: calc(baseH px - 헤더)`)로 나눔. 열린 동안 Grid 수용량·타일 크기·xterm 줄 수·PTY 크기 불변, 넘치는 부분은 잘리고 가려짐 이동이 처리. `clip` = 포커스·스크립트로도 스크롤 안 됨(hidden이면 브라우저가 포커스 요소로 스크롤할 수 있음) |
| 라-3(열림 판정) | `virtualKeyboard.hook.ts`: 키보드 없을 때의 보이는 높이 `baseH`(= vv.height x scale, 닫힘일 때마다 갱신)보다 120 넘게 줄면 열림. 너비가 바뀌면(폴드 펴기·회전) 또는 `fullscreenchange` 뒤 1초 동안은 기준 다시 잡음(Edge 전체 화면 끝내기 = 110 줄어듦도 120 미만). `baseH`를 상태로 내보냄 |
| 라-4(키 줄 위치) | 바꾼 것 없음 — `bottom = vv.offsetTop + vv.height`가 resizes-content에선 줄어든 레이아웃 아래 끝, iOS에선 보이는 영역 아래 끝 |

- 앱 다른 곳: 다이얼로그는 줄어든 화면 가운데로 다시 정렬(작은 것 6개 높이 300 안팎), 스크립트 편집(폰 = 전체 화면)은 키보드 위까지 줄어 버튼이 보임. `appWindown.store`는 너비 등급·헤더 여백만 써서 영향 없음
- 헤드리스(격리) `pw/kb3.mjs` 12/12: 창 높이를 340 줄여 resizes-content 흉내 -> 키 줄 아래 끝 = 화면 아래 끝 · 타일 크기·Grid 정보(`Grid 1 row · 1타일`)·xterm 39줄·PTY 62x39 불변 · 커서 줄 보임 · 문서 스크롤 0 · 닫으면 제자리 · 56 줄어듦 = 키보드 아님 · 너비 바뀜 + 200 줄어듦 = 키보드 아님(기준 다시) · 새 기준에서 340 줄면 열림 · 이름 변경 다이얼로그 화면 안 · 내용 편집 전체 화면 460에 맞고 닫기·저장 보임. + 기존 `kb.mjs` 22/22(메타 검사만 resizes-content로) · `kb2.mjs` 10/10(가짜 visualViewport = iOS 경로)

## 고침-나(커서 기준 이동) 헤드리스 결과 (2026-10-14, 격리, 10/10 + 기존 22/22)
- `pw/kb2.mjs`: 412x800, 전체 높이(h 1) 타일을 칼럼 2로 이동(`button[title="다음 칼럼"]`) 뒤. 스크립트 = 소문자 `l` 10줄 출력 / `h` 화면 지우고 맨 위 / `b` 맨 아래 줄. 커서 위치 = `.xterm-cursor` 상자
- 결과: 키 줄 = 키보드 바로 위 · 커서 맨 아래에서 열기 -> 커서 줄 아래 = 키 줄 위 끝(타일 위 끝 -281, 헤더가 위에 그려짐) · 열린 채 `h` -> 커서가 띠 안(이동 줄어 타일 위 10) · 출력 40줄로 커서가 내려가도 따라 올라감 · 닫힘 = 제자리 · 커서 맨 위에서 열기 = 타일 머리 보임(이동 거의 0) · 열린 채 `b` -> 따라 올라감
- **헤드리스 함정**: playwright `keyboard.type('B')`(대문자, shiftKey 없는 keydown)를 터미널에 보내면 칼럼이 1로 바뀌었다(`press('X')`도 같음, `Shift+y`는 안 바뀜). 우리 코드에 이런 단축키는 없고 키보드가 닫힌 상태에서도 생김 -> 합성 이벤트 탓으로 보고 시험은 소문자로. 실기기에서 대문자 입력 후 칼럼이 바뀌면 다시 볼 것

## 사용자 Android 확인 목록 (https://test1.falinux.com, 새로고침 후) — 2026-10-14 결과: 3~7 모두 OK, 1·2는 Edge 폰 크기 기본 화면만 키 줄이 뜸(81 -> 라 뒤 34, → history 2026-10-14(5)·(6))
| 번호 | 확인 | 기대 |
|---|---|---|
| 1 | 아래쪽 터미널 타일을 눌러 키보드 띄우기 | 타일 배치·Grid 수 그대로, 그 타일이 키보드 위로 올라옴. 키보드 닫으면 제자리 |
| 2 | 키 줄 위치·모양 | 키보드 바로 위에 붙음, 키를 눌러도 키보드가 안 닫힘, 화살표 글자 모양 정상 |
| 3 | `cat`이나 htop에서 Ctrl 탭 -> c | 중단(Ctrl+C) |
| 4 | Ctrl을 누른 채 가상 키보드 c | 되면 VK3 누른 채 방식 OK / 안 되어도 다음 글자에 Ctrl이 걸림(=3과 같음) |
| 5 | 한글 자판 상태에서 Ctrl -> ㅊ | 바로 Ctrl+C면 OK. 조합 때문에 다음 키를 눌러야 나가면 알려 주기 |
| 6 | vim·htop에서 방향키·Esc | 이동·명령 모드 정상 |
| 7 | 터미널에 한글 그냥 입력 | 글자가 깨지거나 두 번 들어가지 않는지(기존 xterm 모바일 한글 문제 확인) |
