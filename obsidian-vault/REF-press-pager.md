# REF — 빈 곳 길게 누르기 · 고른 칸 라벨 강조

> → `history/press-pager.md` / `CURRENT.md` / 앞 파트 `INDEX-dir-grid.md`(목록 아이콘 보기) · `INDEX-pager-tools.md`(팝업-6(길게 누르기로 띄우기))

## 출발 (2026-10-25)
- 사용자: 파트(목록 아이콘 보기) 뒤 터치 길게 누르기 = 타일 컨트롤 패널(`PagerPopup`, Ctrl+Alt 팝업)이 경로 줄·목록 빈 곳에서 안 뜸. 머리 줄 빈 곳은 뜸(확인)
- 원인: `PagerPopup.vue` `pressable` = `.TileFrame__head` 또는 `.TileStrip` 안 `.TileFrame` 밖뿐. 경로 줄(`.crumbs`)·목록은 몸통 안이라 대상 아님
- 같이 묶음(사용자 동의): 고른 칸의 누를 수 있는 상태 라벨 강조. 앞서 낸 "라벨에 꺾쇠·테두리 상시 표시" 안 A~C는 사용자가 "애매"로 철회 -> 대신 한 번 누르기(고름) 때만 강조

## 확정
| 자리(터치 길게) | 결과 |
|---|---|
| 아이템 칸 안(아이콘·이름 글자·바탕) | 체크(그대로) |
| 칸의 버튼·라벨 / 경로 줄 `기본` 등 이동 단계 버튼 | 각자 동작(그대로) |
| 경로 줄 빈 곳 + 현재 폴더 이름 글자(편집 중 아닐 때) | **패널**. 짧게 = 경로 입력칸(그대로), 길게 = 패널만(기존 `eatClick`이 손 뗄 때 click을 먹음) |
| 목록 칸 밖 빈 곳 · "비어 있음" | **패널** |
| 터미널 타일 화면 | 제외(탭 = 포커스, 그대로) |
- 대상 타일 = 폴더 타일 + 장비 타일
- 함정(팝업-6 때 확인): 칸 사이 6px은 Chrome 터치 대상 보정으로 옆 칸이 잡힘 -> 실감되는 빈 곳 = 목록 아래 남는 영역·"비어 있음"
- 라벨 강조: 고른(`focused`) 칸의 `button.dg-lbl`(공유 수·실행 중 ▶ 숫자·장비, 권한 될 때만 button)만 테두리가 상자색 -> 라벨 색으로 **짧은 transition(≈0.2초)으로 부드럽게**(깜빡임 아님). 누를 수 없는 `span` 라벨·선택 모드는 그대로

## 계획 (브랜치 `feat/press-pager`, master에서)
| 단계 | 내용 | 파일 |
|---|---|---|
| 길게-1(자리 표시) | 경로 줄(편집 중 아닐 때)·목록 몸통 루트에 표시 속성(`data-pager-press` 류) | `FolderTileBody.vue`, `DeviceTileBody.vue` |
| 길게-2(판정 확장) | `pressable`에 "표시 안 + 버튼·입력·`.DirGridCell` 밖" 추가 | `PagerPopup.vue` |
| 길게-3(짧은 누르기 충돌) | 짧게 = 입력칸 / 길게 = 패널만 — 기존 `eatClick`으로 되는지 확인 | — |
| 길게-4(글자 선택 끔) | `pointer: coarse`에서 표시 자리 `user-select: none`·`touch-callout: none` | 위 두 body |
| 라벨-1(고른 칸 라벨 강조) | `.DirGridCell.focused button.dg-lbl` 테두리 색 transition | `styles/dir-grid.scss` 또는 `DirGridCell.vue` |
| 길게-5(확인) | 격리, 폰 390 터치 + PC: 위 표 전부 + 라벨 강조(고름·선택 모드·span 라벨) | scratchpad |

## 구현 (2026-10-25, 브랜치 `feat/press-pager`)
- 길게-1(자리 표시): `FolderTileBody`·`DeviceTileBody` 루트에 `data-pager-press`, 경로 줄은 편집 중 `data-pager-press-off`(입력칸·자동완성 자리)
- 길게-2(판정 확장): `PagerPopup.vue` `pressable` — 머리 줄 다음에 `[data-pager-press]` 안이면 `.DirGridCell`·`[data-pager-press-off]` 밖일 때 true(버튼·입력 제외는 기존 첫 줄)
- 길게-3(짧은 누르기 충돌): 코드 변경 없음 — 기존 `eatClick`으로 길게 = 패널만
- 길게-4(글자 선택 끔): 두 body 루트 `@media (pointer: coarse)` user-select·touch-callout 끔, `input`은 되살림(iOS는 조상이 막으면 입력이 안 됨)
- 라벨-1(고른 칸 라벨 강조): `dir-grid.scss` `.dg-lbl`에 `--ring`(보통 = `--c`, `--solid` = `--on-c`) + `box-shadow: 상자색 바깥 1.5px, inset 1.5px rgba(--ring, 0)` + `transition: box-shadow 0.2s`. `.DirGridCell.focused:not(.checking) button.dg-lbl` = inset 알파 0.9. `DirGridCell` 루트에 `checking` 클래스 추가

## 확인 (2026-10-25)
- test1(8011 vite, `VITE_API_URL=https://test3.falinux.com npm run dev`) + 헤드리스 폰 390 CDP 터치(`Input.dispatchTouchEvent`), 로그인 = `.env.development.local` 주석 줄 계정으로 test3 `/users/session`. 데이터 바꾸는 동작(경로 이동·실행)은 안 누름
- 13/13: 경로 줄 빈 곳 길게 = 패널·입력칸 안 열림 · 바깥 짧게 = 닫힘 · 경로 줄 빈 곳 짧게 = 입력칸 · Esc 닫힘 · 현재 폴더 글자 길게 = 패널·글자 선택 없음 · 도구 줄(둘째 줄) 빈 곳 길게 = 패널 · 목록 아래 빈 곳 길게 = 패널 · 칸 길게 = 체크(패널 아님) · 칸 짧게 = 체크 풀림 · 머리 줄 길게 = 패널(회귀) · 고른 칸 button 라벨 = 테두리 0.9 · 안 고른 칸 = 0
- 함정: 칸 길게 = **보드 전체 선택 모드**(위 "N개 선택" 막대) — 체크를 다 풀어도 ✓ 전까지 유지(기존 동작). 막대가 타일을 아래로 밀어 이후 좌표가 어긋남 -> ✓(390폭 기준 362,22)를 눌러 끝낸 뒤 진행
- 장비 타일은 보드에 없어 헤드리스로 안 봄(같은 표시·스타일) — 사용자 test1 확인 때 같이
