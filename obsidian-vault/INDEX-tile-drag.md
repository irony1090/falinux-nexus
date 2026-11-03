# INDEX — 타일 끌기 (마감)

> 2026-11-03 파트(업데이트 app) Files 타일 개선 마감 때 HISTORY가 10k자에 닿아 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
10-29 마감. 브랜치 `feat/tile-drag` -> master 머지(`29ae134`). 구조-1~3(순서 배열 `order`·연 곳 `parent`) · 끌기-1·2 · 제스처-1 · 보드-1 · 칸-1·2 · 끝 띠. 사용자 PC·폰 확인 끝. 남은 것(회원 관리 판 끌기)은 다음 파트(내 장비·관리자 타일 이동)에서 처리
- 다른 파트에서도 지킬 결정(타일 순서·끌기·터치 제스처·타일 추가 주체 등)은 MEMORY "확정 결정"과 `REF-decisions-ui.md`에 그대로 남김

## REF / history 목록
- `REF-tile-drag.md`(계획·결정) · `REF-tile-drag-impl.md`(끌기 구현) · `REF-tile-order.md`(순서 구조) / `history/tile-drag.md` · `history/tile-drag-impl.md` · `history/tile-order.md`. 끌기 기반은 반복 주제 `REF-util-drag.md`(접지 않음)

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY 진행 중 파트 줄: 그 전 = 타일 끌기(마감, 10-29, `feat/tile-drag` 머지) → `REF-tile-drag.md` · 순서 구조 `REF-tile-order.md` · 끌기 구현 `REF-tile-drag-impl.md`
- HISTORY: | `history/tile-drag.md` | 파트(타일 끌기)(2026-10-29 신설) ↔ `REF-tile-drag.md` | 10-29 끌기 착수 전까지(방향·계획·끄는 모양) |
- HISTORY: | `history/tile-drag-impl.md` | 파트(타일 끌기) 끌기-1·2·제스처-1·보드-1·칸-1·2 구현(2026-10-29 분할) ↔ `REF-tile-drag-impl.md` | 10-29 파트 마감(칸-1·2·끝 띠, 커밋 + master 머지) |
- HISTORY: | `history/tile-order.md` | 파트(타일 끌기) 구조-1~3 순서 구조(2026-10-29 분할) ↔ `REF-tile-order.md` | 10-29 구현 + 격리 확인 |
