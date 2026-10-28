# history/mobile-viewport — 모바일 가상 키보드·뷰포트 (반복 주제)

> 지식 → `REF-mobile-viewport.md` / Nexus 측정·구현 상세 경위 → `history/mobile-keyboard.md` / 색인 → `INDEX-mobile-keyboard.md`

## 2026-10-14 — 반복 주제로 분리 (Nexus 파트(모바일 가상 키보드) 마감 때)
- 사용자: "가상 키보드 관련 이슈는 별도 색인으로 처리하고 다른 프로젝트에서도 일어날 수 있으니 중요도가 높아" -> 프로젝트와 무관한 지식(뷰포트 메타 3가지, 열림 판정, 브라우저별 실측, 재배치 막기, 버튼 줄, 헤드리스 흉내)을 `REF-mobile-viewport.md`로 분리. 전역 CLAUDE.md "Vault 반복 주제 분리"에 지정 주제로 추가
- 근거 측정은 모두 Galaxy Fold 8(접음·펼침) + 크롬·Edge 사진(user vault `user/vrKeyboard/photo_2026-10-14_*`) — 경위는 `history/mobile-keyboard.md` 2026-10-14(3)~(6)

## 2026-10-20 — Edge 전체 화면 빈 띠 관찰 (조사 중)
- Nexus 헤더·푸터 개편 확인 중 사용자 사진: Edge 폰 전체 화면 + 키보드에서 키 줄과 키보드 사이 빈 띠, 기본 화면도 주소창 위에 빈 띠. 3절 "전체 화면 = 정확"과 어긋남 -> REF 3절에 의문 표시. 측정 도구·판정표 → Nexus `REF-board-keyboard.md` "전체 화면 키보드 가림"

## 2026-10-20 — Edge 폰 전체 화면 재측정: 약 40 위
- Nexus 측정 상자 사진: 전체 화면 키보드 열림 때 vv·inner·dvh/svh/lvh 전부 461, 실제 키보드 위 끝 약 501. 3절 표의 "정확"을 정정. 상세 → Nexus `REF-board-keyboard.md` "판정 결과"

## 2026-10-23 — 자동완성 목록을 눌러도 키보드 유지
- Nexus 파트(경로 직접 입력) 모바일 확인: 목록 상자 `mousedown.prevent`로 입력칸 포커스 유지 -> 키보드 안 내려감(사용자 실기기). REF 5-1절

## 2026-10-25 — 확대 막기(줌-1~3)
- 사용자: Nexus 웹 터치 줌 막기(아이폰 포함, 앱 전체, master 바로) -> 줌-1(viewport)·줌-2(전역 touch-action)·줌-3(iOS gesturestart) → REF 7절. 크롬 헤드리스로 막기 전후·방법별 비교(전부 1.00), test3 `web` 교체(옛 것 `web.bak`, 그 전 `web.bak.press-pager-전`)
- 사용자 확인 OK(아이폰은 미확인, "괜찮을 것 같다" → `BACKLOG.md`) -> test1 닫음
- 사용자 요청으로 test3 웹 백업 `web.bak*` 4개 삭제(13M씩)
