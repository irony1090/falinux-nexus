# history/util-drag — 드래그 인프라 (DraggableSession / useDragGhost / GhostArea)

> `history/util.md`에서 10k자 기준 분할(2026-08-07). 요약·재사용 지식 → `REF-util-drag.md` / 현재 진행 → `CURRENT.md`.

## 2026-08-07(3) — DraggableSession/useDragGhost/GhostArea 포팅 설계 확정 (코드 없음)
- Node 카탈로그 UI Phase2/3(드래그) 논의 중 구 test-jig 드래그 코드(`DraggableListener`+`DraggableLogic`+ghost 손코딩) 재사용 불가 수준 진단 → 재설계 착수.
- 문제: 단일슬롯 콜백 7개 / 판정→추적을 synthetic event 재발사로 봉합 / 상태기계에 이웃재정렬 도메인로직 혼입 / ghost 추종 로직 소비처마다 중복.
- 방향: `DragListener` 유지 + `StartPolicy` 함수 2종(threshold/long-press)이 클래스계층+synthetic dispatch 대체 + `DraggableSession`(`EventInterface` 상속) + `useDragGhost` 컴포저블 + `GhostArea`/`ghost.store` 포팅(`feature/layout`, `AppDialog`와 동일한 자리).
- 구현은 다음 세션. 상세 → `REF-util-drag.md`.
