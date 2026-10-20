# INDEX — 페이지 표시 개편 (마감)

> 2026-10-20 파트(페이지 표시 개편) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
보드 헤더의 `VPagination`(Grid 탭) + 칼럼 화살표를 인디케이터(`GridPager`)로 교체: Grid마다 도트, 반 칸 = 두 도트 사이 작은 도트, 클릭 = Grid 이동, x축 드래그 = 칼럼 이동, 드래그한 채 아래로 = 타일 패널(`TilePeek`, Grid별 미니맵) + x축 = active 타일 이동(막 다 보이는 만큼 스크롤). Grid가 많으면 창 방식. 디자인 캔버스 프로토타입(https://claude.ai/artifact/XLvPvzUkZQrfD1FN6Vaqo5) -> 구현 N1(위치 계산)~N6(TileNav 교체) -> 격리 헤드리스 31/31 + 사용자 확인

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-board-pager.md`(보드 페이지 표시 개편 — 도트·드래그·타일 표시, 0.5 Grid 단위) / `history/board-pager.md`
- 확인 스크립트 명세 → `REF-test-scripts-1020.md` "2026-10-20 페이지 표시"

## HISTORY 색인 행 (원문)
### `history/board-pager.md` — 보드 페이지 표시 개편 (2026-10-20 신설)
- 2026-10-20 아티팩트 프로토타입 1·2차, 이동 단위·2.5 표시·타일 위치 규칙 확정
- 2026-10-20(2) 구현 N1~N6 + 헤드리스 31/31
- 2026-10-20(3) 사용자 확인 -> 마감·커밋

## 이 파트에만 해당하는 결정
- 위치 단위 = 칼럼(0.5 Grid, 폰 1 Grid). 지금 앱의 칼럼 정지가 이미 그 단위(`Grid.cols` = 오른쪽 칼럼이 비었나)라 `columnStops` 변경 없음
- 2.5 표시 = 가운데 작은 도트(막대가 줄어듦) / 타일 이동 = 직전 화면 기준 막 다 보이는 만큼, 1.0 타일은 그 Grid 전체 / 좁은 공간 = 창 방식 / 인디케이터 헤더 가운데
- 보드 이동 애니메이션 = 브라우저 기본 smooth / 타일 모드에서 놓을 때 포커스는 `pointer: fine`만 / ←·→ 키 = 칼럼 ±1
