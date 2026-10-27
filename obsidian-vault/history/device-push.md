# history — 파트(장비 상태 push)
> → `REF-device-push.md` · → `CURRENT.md`

## 2026-10-27
- 계기: 사용자 질문 "장비 접속 현황 REST 폴링이었나" → 보드·내 장비·관리자 모두 10초 폴링 확인 → 느려서 불편, 소켓으로
- 조율: 알림 = 신호만(추천 나) · 받는 사람 = 주인+폴더 보는 사람+관리자 · 범위 = presence 빼고 전부 · 폴링 60초 안전망 · 라벨·등록/삭제·주인 변경 포함 · `SHARE:CHANGE`도 재조회 · 코드 = Claude 작성
- 구현: 브랜치 `feat/device-push`(master `4cf6855`에서). 확인 중 내 장비·관리자 페이지 직접 진입 시 소켓 미연결 발견 → 사용자 승인 받고 소켓 connect를 `App.vue`로 옮김
- 확인: 격리 받는 사람 10/10 · 보드 0.4초 · 관리 페이지 1.5초 안(→ REF "확인")
- 사용자 직접 확인(격리 3001, worker 끄고 켜기 요청 두 번) → "확인 끝" → 파트 마감·`feat/device-push` 커밋 → master `--no-ff` 머지 → `INDEX-device-push.md`
