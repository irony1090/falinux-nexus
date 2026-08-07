# HISTORY: 위젯 컴포넌트 (apps/frontend/src/feature/widget)

> 요약/재사용 지식 → `REF-widget.md`. 현재 진행 → `CURRENT.md`.

## 2026-08-07 — Skeleton / SkeletonGroup shimmer 로딩 위젯 신설
- `component/Skeleton.vue` + `component/SkeletonGroup.vue` + `store/skeletonGroup.store.ts` 추가.
- 그룹 단위 shimmer 애니메이션(SVG clipPath 합성) + 개별 컴포넌트 wrapper 조합 구조. 관찰 리소스(ResizeObserver/MutationObserver)는 로딩 중일 때만 활성화(`LifecycleRegistry`, → `history/util.md`).
- 실제 소비처(어느 화면이 이 위젯을 쓰는지)는 아직 미배선.
