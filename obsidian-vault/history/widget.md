# HISTORY: 위젯 컴포넌트 (apps/frontend/src/feature/widget)

> 요약/재사용 지식 → `REF-widget.md`. 현재 진행 → `CURRENT.md`.

## 2026-08-07(2) — StickyBox 컴포넌트 구현 + 스토어 리팩터
- `component/StickyBox.vue` 신설(스토어는 기존) — 중첩 sticky header/footer가 서로의 차지 영역만큼 밀려 쌓이는 위젯 실체 구현.
- `store/stickyBox.store.ts` 리팩터: `thisClient`(ref 공유) 방식 → `reportSelf(v)`(콜백 보고) 방식으로 읽기/쓰기 역할 분리. **`viewportClient` 신규**(다이얼로그 등 조상이 CSS containing block을 만들 때 `position:fixed` 기준을 보정하는 값, `rootClient`/`maxClient`와는 역할이 다름).
- 관찰(리사이즈)은 개별 `useResize` 대신 공유 그룹(`useResizeCallback`) 소비 — → `history/util.md`.
- 상세 → `REF-widget.md` "StickyBox".

## 2026-08-07 — Skeleton / SkeletonGroup shimmer 로딩 위젯 신설
- `component/Skeleton.vue` + `component/SkeletonGroup.vue` + `store/skeletonGroup.store.ts` 추가.
- 그룹 단위 shimmer 애니메이션(SVG clipPath 합성) + 개별 컴포넌트 wrapper 조합 구조. 관찰 리소스(ResizeObserver/MutationObserver)는 로딩 중일 때만 활성화(`LifecycleRegistry`, → `history/util.md`).
- 실제 소비처(어느 화면이 이 위젯을 쓰는지)는 아직 미배선.
