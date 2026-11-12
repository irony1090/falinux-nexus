# history — CURRENT에서 내려보낸 줄 원문 (2026-11-12~)

> → `CURRENT.md` · 앞 파일 `history/current-archive-7.md` · REF 없음(원문 보관용)

## 11-12 알림 찰떡 마감 때 내려보냄
- **(2026-11-12) 푸터 배지 잘림 — 수정·커밋·master `--no-ff` 머지**(사용자 "머지커밋") — `fix/footer-badge`. 원인 = 푸터 오른쪽 알약 `overflow: hidden`, `styles/global.scss` `.foot-badge`로 배지를 작게 해 아이콘에 겹침(알림 + 전송) → `history/board-chrome.md` "11-12"
- **(2026-11-11) 미니맵 진행도 — 마감**(`feat/peek-nudge` 커밋·master `--no-ff` 머지, 사용자 "맘에 들어") → `history/board-pager.md` · `REF-board-pager.md` "미니맵 진행도". test1·test2는 11-11 끔
- **(2026-11-11) 칼럼 흐름 빌드 — test3·집 서버 교체 끝**(사용자가 교체, Claude 확인): test3(8013) = `build/261111-column-flow`(14:49 켜짐) · 집 서버 = 묶음 `build/home-update/261111-column-flow` 배포, 공개 `index-DYuchPlR.js`(nexus·direct) → `REF-home-deploy-updater.md` "11-11 칼럼 흐름 묶음"
- **(2026-11-11) 파트(칼럼 흐름) — 마감**(사용자 "문제 없어 마감해도돼"), `feat/column-flow` 커밋·master `--no-ff` 머지 → `REF-column-flow.md` · `history/column-flow.md`. 보드(흐름-1(채우기)~흐름-5(인디케이터)) + URL focus(포커스-1(스크롤 위치)~포커스-3(경로 합치기)) + 관리 페이지(관리-1(채우기)~관리-5(인디케이터)·포커스-4(관리 페이지))
  - 확인용 dev supervisor 5051·vite 8011 = 마감 때 Claude가 끔(다시 띄우는 법 `REF-test-env.md` "8012가 다른 작업공간에 잡혀 있을 때"). 개발 DB에 계정 `mflow-*` 2개 남음
