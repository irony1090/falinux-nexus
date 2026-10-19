# history — 터미널 터치 스크롤

> → `REF-term-touch.md` / 현재 진행 → `CURRENT.md`

## 2026-10-19
- 사용자 질문 "터치에서 Windows PowerShell 스크롤바 드래그가 안 됨, 터치 -> 마우스 변환해야 하나?" → xterm 6 소스 조사: 이미 pointer 이벤트, 원인 추정 = `touch-action` 없음 -> `pointercancel` / 쓸어서 스크롤은 xterm 6에 없음 → 안 가(스크롤바 `touch-action`)·안 나(쓸어서 스크롤) 제시
- 사용자: Linux는 스크롤바 생기는 앱이 없어 미확인, 쓸어 스크롤은 원래 안 됐음 → "안 가 먼저" → `ProcessTerminal.vue`에 CSS 한 줄
- 헤드리스 확인: 처음엔 수정 전후 둘 다 실패 → 원인 = 스크롤바가 숨김 중 `pointer-events: none`(터치엔 hover 없음) → 탭으로 띄운 뒤 끌기: 수정 전 285->267 + `pointercancel` + 페이지 스크롤 / 수정 후 285->105, `pointercancel` 없음. `vue-tsc` 통과. 사용자 실기기 확인 대기
- 사용자 요청 → irony 계정에 확인용 노드 2개 REST 등록: `scroll-test`(215, 폴더 test 14) · `scroll-test.ps1`(216, 폴더 win 213)
- 사용자 실기기 확인: 막대 터치 드래그 동작 → 안 가 확인 끝
- 곁가지: 셸 노드 종료가 10초 걸림 → 설계대로(대화형 셸이 종료 요청 무시) 그대로 둠 → `REF-process-kill.md` 메모
- 안 나(쓸어서 스크롤) 충돌 검토 → 구조(쓸기 -> 휠, 상자 먼저, 관성) → 사용자 "진행" → `touchScroll.util.ts` 구현. 헤드리스 첫 실행 6/11: 정상 화면 합성 휠 무시(`wheelDeltaY` 0) → 정상 화면은 `scrollLines`로 바꿈 / 스크롤바 실패는 테스트 상자 잘림 → 11/11. 상자 반올림 찌꺼기 보정 추가. 사용자 실기기 확인 대기

## 2026-10-19(2)
- 사용자 실기기 확인: 안 나(쓸어서 스크롤) "잘돼" -> 파트(터미널 터치 스크롤) 마감, `INDEX-term-touch.md`로 접음, 코드(`ProcessTerminal.vue`·`touchScroll.util.ts`)+vault 한 커밋. 확인용 노드 215·216은 남겨 둠
