# INDEX — 터미널 터치 스크롤 (마감)

> 2026-10-19 파트(터미널 터치 스크롤) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
모바일에서 터미널 스크롤. 안 가(스크롤바 터치) = 스크롤바에 `touch-action: none`(브라우저 팬 -> `pointercancel`로 드래그가 끊기던 것) / 안 나(쓸어서 스크롤) = `touchScroll.util.ts` `bindTouchScroll` — 10px 전엔 탭, 크기 고정 상자 먼저 -> 끝에 닿으면 터미널, 관성, 정상 화면 `term.scrollLines` / alt 화면·마우스 모드 합성 `WheelEvent`. 파일 = `ProcessTerminal.vue` + `touchScroll.util.ts`. 헤드리스 11/11 + 사용자 실기기 확인(둘 다)

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-term-touch.md`(터미널 터치 스크롤 — 스크롤바 `touch-action` + 쓸어서 스크롤) / `history/term-touch.md`
- 확인 스크립트 명세 → `REF-test-scripts-1014.md` "2026-10-19 터치 스크롤"

## HISTORY 색인 행 (원문)
### `history/term-touch.md` — 터미널 터치 스크롤 (2026-10-19 신설)
- 2026-10-19 원인 조사(`touch-action` 없음 -> `pointercancel`, 숨김 중 `pointer-events: none`) + 안 가(스크롤바 `touch-action: none`) 구현·헤드리스 확인
- 2026-10-19(2) 안 나(쓸어서 스크롤) 구현·헤드리스 11/11 -> 사용자 실기기 확인 -> 마감·커밋

## 이 파트에만 해당하는 결정
- 크기 고정 화면 = 상자 스크롤 먼저, 끝에 닿은 뒤 남은 양만 스크롤백 / 관성 넣음(마지막 100ms 속도, 16ms당 0.95)
- 전체보기 중엔 따로 끄지 않음(`TileFrame--overview`의 `pointer-events: none`으로 터치가 안 옴)
- 정상 화면에 합성 휠 금지(VS Code 스크롤 요소가 `wheelDeltaY`를 읽어 늘 0) -> `scrollLines`
- 남은 불편(필요하면): 스크롤바가 탭 전엔 숨김·slider 14px -> `@media (pointer: coarse)`에서 항상 보이게·넓게
