# INDEX — 닫은 뒤 active (마감)

> 2026-10-28 파트(닫은 뒤 active) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
브랜치 `feat/close-active`(master `bd0ea50`에서). 프론트만.
- 계기: 파트(패널 창 버튼) 재현 중 "active 타일을 닫으면 active 없음 -> 패널 버튼 없음" -> 사용자 "가까운 타일을 active로"
- 결정: 닫히면(패널·머리 줄 ×·터미널 종료 뒤) 위치 번호 이웃(다음, 마지막이면 앞)을 active + 화면 밖이면 `reveal`, 최소화·보드 바꾸기는 해제만, 보드·관리 페이지 둘 다
- 확인: 격리 헤드리스 6/6 + 사용자 확인 OK(2026-10-28), `npm run build` 통과. 격리 환경 끔(시험 셸 정리, `/tmp/claude-1000/pr-t` 지움)
- 배포: test3(8013) **웹만** 교체(`build/mochi/web`, 옛 것 `web.bak-*`) — Go 변경 없음
- 커밋: 브랜치에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-close-active.md`(확정 표 · 구현 · 확인) / `history/close-active.md`

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: **진행 중 파트**: 닫은 뒤 active(`feat/close-active`, 2026-10-28) → `REF-close-active.md` (직전 = 패널 창 버튼(마감) → `INDEX-pager-head-btns.md`)
- HISTORY: | `history/close-active.md` | 파트(닫은 뒤 active) — active 타일을 닫으면 이웃 active(2026-10-28 신설) | 10-28 구현·격리 확인 |

## 이 파트에만 해당하는 결정 · 남은 것
- 이웃 = 위치 번호(화면 거리 아님), 새 active로 포커스는 옮기지 않음(폰 키보드)
- 다른 파트에서도 지킬 것(active가 배치에서 빠질 때의 규칙)은 MEMORY·`REF-decisions.md` "닫은 뒤 active"에 남김
