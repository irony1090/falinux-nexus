# INDEX — 패널 창 버튼 (마감)

> 2026-10-28 파트(패널 창 버튼) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
브랜치 `feat/pager-head-btns`(master `8fa5bdd`에서). 프론트만.
- 계기: 사용자 요청 — Ctrl+Alt 컨트롤 패널에 active 타일 머리 줄 창 버튼 `[숨기기 │ 크기 │ 닫기]`를 그대로
- 단계: 패널-1(창 버튼 등록 `tileHeadActs`) -> 패널-2(묶음, 보드·관리 패널) -> 패널-3(패널 위치 글자 빼기) -> 패널-5(도트 자리 최소 3개 폭)
- 확인: 격리 5051·3001 헤드리스(타일 종류별 버튼·동작·관리 페이지·최소 폭) OK, `npm run build` 통과. 배포 안 함. 격리 환경 끔(시험 셸 정리, `/tmp/claude-1000/pr-t` 지움)
- **미해결**: 사용자 브라우저에서 2번 타일 패널이 빔(새로고침 뒤에도) — 재현 못 하고 사용자 결정으로 마감 → `REF-pager-head-btns.md` "미해결", `BACKLOG.md`
- 커밋: 브랜치에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-pager-head-btns.md`(확정 표 · 타일별 버튼 · 패널-1~5 · 함정 · 확인 · 미해결) / `history/pager-head-btns.md`

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: **진행 중 파트**: 패널 창 버튼(`feat/pager-head-btns`, 2026-10-28) → `REF-pager-head-btns.md` (직전 = 경로 줄 정리(마감) → `INDEX-path-row.md`)
- HISTORY: | `history/pager-head-btns.md` | 파트(패널 창 버튼) — 컨트롤 패널에 active 타일 창 버튼(2026-10-28 신설) | 10-28 패널-1~3·격리 확인 |

## 이 파트에만 해당하는 결정 · 남은 것
- 패널 버튼 모양 = 투명 칸 + 1px 선, 위치 = 왼쪽, 없는 버튼은 뺌(비활성 X), ⋮ 메뉴는 넣지 않음(사용자 "3가지")
- 패널에서 닫은 뒤 이웃 타일 자동 active는 제안만(사용자 답 없이 마감)
- 다른 파트에서도 지킬 것(머리 줄 창 버튼을 바꾸면 패널 등록도 같이)은 MEMORY·`REF-decisions.md` "패널 창 버튼"에 남김
