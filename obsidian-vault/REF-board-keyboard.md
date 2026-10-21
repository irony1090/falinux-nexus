# REF — 보드 화면 키보드 열림 (파트(헤더·푸터 개편)의 키확인)

> 2026-10-20 `REF-board-chrome.md`에서 분리. 이력 → `history/board-keyboard.md` / 헤더·푸터 본편 → `REF-board-chrome.md` / 일반 지식 → `REF-mobile-viewport.md` / Nexus 키보드 구현(2026-10-14) → `REF-mobile-keyboard.md` / 현재 진행 → `CURRENT.md`

## 전체 화면 키보드 가림 (2026-10-20 조사 중, Edge 폰)
- 증상(사진 `user/photo_2026-10-20_18-08-40.jpg` 전체 화면 · `18-09-15` 기본 화면): 두 화면 모두 키 줄과 키보드(또는 주소창) 사이에 빈 띠(약 23~30px). 기본 화면은 타일이 다 보이고, 전체 화면은 커서 줄 아래(상태 줄)가 키 줄에 가려짐
- 후보: ① `visualViewport` 아래 끝이 실제 키보드 위 끝보다 위(`REF-mobile-viewport.md` 3은 "Edge 전체 화면 = 정확"이었음) ② 전체 화면은 타일이 길어(헤더 삭제로 약 28px 더) 키 줄 위에 다 안 들어감 -> 커서 기준 이동 규칙대로 아래가 가려짐
- 판정표(키확인-2(원인 판정)) — 상자의 숫자(레이아웃 CSS px)로:

| 보는 값 | 이러면 |
|---|---|
| `kb bot`(= vv top + h) = `bar` 아래 끝인데 사진에선 그 아래 빈 띠 | ① 브라우저 vv 값이 키보드 위 끝보다 위. 빈 띠 px ≈ 사진 띠 높이 / dpr 배율. `vk`(가상 키보드 API)·`inner`·`client`·`lvh - 키보드`가 실제와 맞는지 비교해 대체 값 후보 찾기 |
| `tile` 아래 끝 > `bar` 위 끝 | ② 타일이 키 줄 위 공간보다 김 -> 커서 기준 이동 규칙(`TileWorkspace` `measure`)대로 아래가 가려짐. 기본 화면과 `base`(baseH) 차이 = 막대 높이 |
| 둘 다 | ①을 고치면 ②의 가려지는 양이 그만큼 줄어듦 |

### 판정 결과 (2026-10-20 키확인-2(원인 판정)) = **둘 다**
사진 `user/photo_2026-10-20_21-27-0{5,5_2,6,6_2}.jpg`(기본 닫힘·기본 열림·전체 열림·전체 닫힘). Galaxy Fold 8 접음, dpr 2.625, screen 476x752 -> 사진 px ≈ CSS px x 2.62

| 화면(키보드 열림) | 페이지 아래 끝(vv·inner·dvh·lvh) | 실제 위 끝 | 빈 띠 | 타일 아래 vs 키 줄 위 | 이동량 |
|---|---|---|---|---|---|
| 기본 | 442.7 | Edge 얇은 주소 줄 약 472 (자판 약 501) | 약 29 | 402.7 = 402.7(다 보임) | 174.5 = 상한(`bottom - limit`) |
| 전체 | 461.0 | 키보드 약 501 | **약 40** | 456.8 > 421.0(**35.8 가려짐**) | 230.4 = 커서 기준 |

- ①(vv 값 오차) 확인: 전체 화면에서 Edge가 레이아웃을 291(752 - 461) 줄이는데 키보드는 약 251 -> 40을 더 줄임. 빈 띠는 **페이지 밖(브라우저 영역)** 이라 페이지가 그릴 수 없고, 맞는 값을 주는 API도 없음(`inner`·`client`·`dvh/svh/lvh` 전부 461, `vk` 0 = `overlaysContent` 꺼짐이라 값 없음). 페이지 쪽에서 고칠 길은 `overlays-content` 전환뿐(`REF-mobile-viewport.md` 4의 resizes-content 결정과 충돌, 실험 필요)
- ②(타일이 김) 확인: 원래 타일 8~687.2(높이 679.2), `limit` = 461 - 40(키 줄) = 421. 아래 맞춤이면 이동 266.2인데 `measure`는 커서 줄만 키 줄 위로 보장(`c.bottom - limit` = 230.1) -> 커서 아래 35.8(Claude Code 구분선·상태 줄)이 키 줄 뒤. 기본 화면은 이동량이 상한에 닿아 다 보였음
- 수정 방향 후보(키확인-3(수정 방향)에서 사용자와 결정): ② = 타일이 보이는 공간보다 길면 **아래 맞춤 우선**(`s = bottom - limit`), 커서 줄 위 끝이 보이는 영역 위로 나갈 때만 덜 올림(`min(…, c.top - minTop)`) / ① = 받아들이고 `REF-mobile-viewport.md` 3 표 정정

- 측정 도구(임시, 끝나면 삭제): `DevProbe.vue`(`TileWorkspace`에 `import.meta.env.DEV`일 때) — 📎 버튼 = 사진을 user vault로 바로 올림(토큰은 첫 사용 때 입력, 폰 localStorage) / `?kb=1` = vv·inner·dvh/svh/lvh·kb·가상 키보드 API·타일·키 줄 좌표 상자. `vite.config.mts` devUpload 두 줄 주석 풀어 둠(켠 동안 `vue-tsc`가 vite.config import에서 TS5097·TS6307 — 끌 때 같이 원복)

## 맞춤 계획 — 포커스 타일만 키보드 위에 맞춤 (2026-10-20 사용자 결정·허가, **구현·헤드리스 28/28·실기기 확인·마감**)
- 계기: 키확인-3(수정 방향)에서 "아래 맞춤 우선"(위로 올리는 양만 바꿈)을 제안 -> 사용자: 올리면 윗줄이 안 보임 -> **사용자 규칙**: 키보드가 뜨면 전체 타일은 키보드가 겹친 것처럼 그대로, **포커스된 터미널 타일만 원래 가로 위치·폭 유지 + 높이 = 키보드 없는 영역**(위 끝 = 보이는 영역 위 끝, 아래 끝 = 키 줄 위 끝). 터미널 크기는 바뀌어도 됨
- 옛 결정과의 관계: 2026-10-01 ③(입력 중 터미널 크기) = 유지 / ②(입력 모드) = 기각(`REF-node-ui-overview.md` "모바일 가상 키보드")을 **뒤집음**. ①(배치 재계산)은 그대로 막음(나머지 타일·Grid 수용량 불변). 가로 폭 그대로라 cols 불변 -> xterm 줄바꿈 재계산 없음, rows만 바뀜
- **KB1(소유자 아닐 때) = (가) 확정**: 크기 우선권 없는 탭은 타일만 맞추고 터미널은 소유자 크기 고정 + 스크롤(`ProcessTerminal--fixed` 그대로). 사용자: "우선권으로 pty 크기가 안 맞는 건 문제없음, 키보드·특수 키 영역·터미널 영역이 제자리를 찾으면 됨"
- ①(vv 값 오차, Edge 빈 띠 40)은 이 방식에서도 남음(페이지 밖)

| 단계 | 하는 일 | 파일 |
|---|---|---|
| 맞춤-1(타일 띄우기) | `TileFrame`이 `useVirtualKeyboard`의 `open`·`termUid`가 자기 터미널이면, 열리기 직전 자기 rect의 left·width를 잡아 `position: fixed` + `top = kb.top` · `height = kb.bottom - KEY_BAR_HEIGHT - kb.top` · 다른 타일 위 z-index(키 줄 아래). 타일 자리는 `Tile`의 grid 좌표라 빠져도 다른 타일 안 움직임(normal 모드 strip·layout에 transform 없음 확인) | `hook/tileFloat.hook.ts`(`useTileFloat`, 2026-10-21 정돈-8(TileFrame.vue 섞인 관심사)로 `TileFrame.vue`에서 옮김) |
| 맞춤-2(올리기 제거) | `TileWorkspace`의 `shift`·`measure`·`cursorRect`·`translateY`·커서 이동 감시 삭제. 키보드 열린 동안 `baseH` px 높이 고정(라-2(높이 고정))은 유지 | `TileWorkspace.vue` |
| 맞춤-3(크기 반영) | 코드 추가 없음 예상: `ProcessTerminal`의 ResizeObserver(150ms 묶음) -> 소유 탭이면 PTY rows 전송, 아니면 고정 + 스크롤. 닫으면 원래 높이 -> 원래 rows | `ProcessTerminal.vue`(확인만) |
| 맞춤-4(확인) | 격리 헤드리스(창 높이 줄여 resizes-content 흉내, 기존 `kb3` 방식): 다른 타일 rect 불변 · 포커스 타일 아래 끝 = 키 줄 위 끝 · cols 불변·rows 감소 · 닫으면 복원 · 포커스 다른 터미널로 옮김 · 비소유 탭 스크롤. 이후 사용자 Edge 실기기(기본·전체 화면) | - |

- 정한 기본값: 열린 동안 strip이 가로로 움직이면 **위치 따라가기**(2026-10-20 사용자 질문 -> 추천, 원래 grid 칸 `Tile`이 빈 채로 남으므로 `scroll` 때 그 rect의 left를 다시 읽어 `left`만 갱신. 복잡하면 키보드 내리기가 사용자 대안) — 화면 밖으로 밀려도 키보드는 열린 채(지금과 같음) / 다른 터미널로 포커스가 옮겨 가면 옛 타일은 제자리, 새 타일을 띄움

### 구현 메모 (2026-10-20 맞춤-1(타일 띄우기)~맞춤-3(크기 반영))
| 파일 | 한 일 |
|---|---|
| `tile/TileFrame.vue` | `floating` = `kb.open` · 전체보기 아님 · 터미널 · `kb.termUid` = 자기 uid. 켜질 때 부모 `.Tile`(빈 grid 칸)의 left·width를 읽고 `position: fixed` · `top = kb.top` · `height = kb.bottom - KEY_BAR_HEIGHT - kb.top` · `z-index 1000`(키 줄 2000 아래). 켜진 동안 `window` scroll(capture)에서 대상이 이 타일을 품을 때만 rAF로 칸 위치 다시 읽음(터미널 안 스크롤은 무시) |
| `TileWorkspace.vue` | `shift`·`measure`·`cursorRect`·`translateY`·transition·커서 감시·`root` ref 삭제. 남은 것 = 열린 동안 `baseH` px 높이(라-2(높이 고정)) |
| `ProcessTerminal.vue` | 바꾼 것 없음 — ResizeObserver 150ms 묶음으로 소유 탭은 PTY rows 전송, 비소유는 `--fixed` |
| `TerminalKeyBar.vue` | `KEY_BAR_HEIGHT` 주석만 |

### 맞춤-4(확인) 헤드리스 결과 (2026-10-20, 28/28)
- 폰 390x844, 창 높이 340 줄여 열림 흉내. 처음 터미널 59x20(타일 382px) -> 열림 59x25(띄운 타일 0~464, 키 줄 464~504) -> 닫힘 59x20
- 통과: fixed · 가로 자리·폭 그대로 · 위 0 · 아래 = 키 줄 위 · 루트 타일 rect 불변 · 다른 터미널 안 띄움 · cols 그대로·rows만(서버 = xterm) · 화면 아래 끝 <= 키 줄 · 문서 스크롤 0 · 포커스 유지 · strip 한 Grid(382) 이동 -> 칸 따라감 · 되돌림 · 다른 터미널로 포커스 -> 옛 타일 제자리·rows 복원, 새 타일 띄움 · 닫으면 풀림·rows 복원 · **비소유 탭**: 타일만 띄우고 서버 PTY 그대로·xterm rows = 소유자 rows
- 이 경우 타일이 오히려 커짐(382 -> 464, rows 20 -> 25): 규칙이 "키보드 없는 영역 채우기"라 원래보다 길어질 수 있음
- 남은 확인 = 키확인-4(나머지 실기기): Edge 기본·전체 화면 실제 키보드, 한/영·이모지 패널로 키보드 높이 바뀔 때, 같은 터미널을 PC에서도 열어 둔 상태

### 실기기 확인 (2026-10-20, Edge Fold 8 접음) — 사용자 "잘 표현되는 것 같아"
- 사진 `user/photo_2026-10-20_22-29-1{8,8_2,9,9_2}.jpg`(전체 열림·전체 닫힘·기본 열림·기본 닫힘). 측정 상자 shift 0
- 전체 화면 열림: tile 0~421 · bar 421~461 / 기본 화면 열림: tile 0~402.7 · bar 402.7~442.7 -> 둘 다 **타일 아래 끝 = 키 줄 위 끝**, Claude Code 상태 줄까지 보임. 닫힘은 원래 배치
- Edge 빈 띠(①(vv 값 오차))는 키 줄 아래에 그대로(예상대로)
- 남은 실기기: 한/영·이모지로 키보드 높이 바뀔 때 / PC에서 같은 터미널 열어 둔 상태 / 키확인-4(나머지 실기기)의 푸터·당김·PC 휠
- (2026-10-20) 사용자 "전부 확인" -> 키확인-4(나머지 실기기) 끝. 키확인-5(정리·마감): `DevProbe.vue` 삭제·`TileWorkspace`의 측정 줄 제거·`vite.config.mts` devUpload 다시 주석, `npm run type-check` 통과
