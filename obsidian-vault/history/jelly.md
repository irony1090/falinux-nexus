# history — 젤리 효과 · 관리 페이지 기록 (파트(젤리·관리 기록))

→ REF `REF-jelly.md` · URL 규칙 `REF-member-ui.md` · 현재 `CURRENT.md`

## 2026-10-27
- 사용자 요청 3개: ①보드로(관리 페이지에서 원래 보던 보드로) ②관리 기록(진입만 push, 안은 replace) ③젤리 효과(hover·누름 말랑한 효과). 업로드 실행 권한 질문은 BACKLOG("업로드 파일 실행 권한")로 미룸
- 조율: "오른쪽 메뉴" = 사용자 착각, 왼쪽 로고 메뉴 / 보드 주소 = 활성 타일(`focus`)까지 이미 URL에 있음 / 내 장비↔관리자 = replace
- ③시안 아티팩트 https://claude.ai/artifact/Gvj3eoa77tsppmWHsXEG9n — v1 세기 4단계·알약 안쪽만/상자째·넓은 줄 토글·컨트롤 패널 / v2 아이콘 보기 칸 라벨 추가·패널 마우스 위치·통통 / v3 사용자가 패널이 오른쪽 아래 기준으로 커지는 버그 발견 -> `translate` 속성으로 수정, 세기 중간, 넓은 줄 제외
- 구현(`feat/jelly-nav`): `router/lastBoard.ts`(새) · `router/index.ts` · `MemberMenu.vue`(보드로 = `goLastBoard`, 내 장비·관리자 `replace`) · `memberUrl.hook.ts`(write 항상 replace, `pushPart` 삭제) · `memberBoard.store.ts`(`replaceNext` 삭제, `dropMissing = close`) · `styles/jelly.scss`(새) + `main.ts` import · `PagerPopup.vue`·`MemberPagerPopup.vue`(`jelly-pop`, `translate` 속성)
- 확인: `vue-tsc` 0 · `vite build` OK · test1 브라우저 — ①②는 REF "①②구현 메모" 확인 줄, ③ = `dg-btn` hover `jelly-hover`·기준점 왼쪽 위, 로고 메뉴 = 자식(로고·이름·화살표)만 `jelly-hover`, disabled `tile-btn` = 없음, 패널 열림·닫힘 동안 알약 가운데 고정(폭 129 -> 223 -> 211 -> 214, 넘쳤다 돌아옴) + 닫힌 뒤 hidden
- 개발 서버: supervisor 8012(`go run ./cmd/supervisor -env dev.env` — `-env`는 소스 폴더 기준 상대 경로, `cmd/supervisor/dev.env`로 주면 경로가 겹쳐 실패) · 프론트 8011. 로그인 기본값(`.env.development.local`)은 사용자가 주석 처리해 둠
- **파트 마감**: 사용자 "맘에 든다" -> vault 정리·커밋·마감. `INDEX-jelly.md`로 접음, 다른 파트에도 남는 규칙(새 버튼 클래스 -> `jelly.scss`, 팝업 = `translate` 속성 + `jelly-pop`)은 MEMORY 확정 결정 "젤리 효과" + `REF-decisions.md`로
