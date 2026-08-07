# HISTORY: 범용 유틸 (apps/frontend/src/common/util)

> 요약/재사용 지식 → `REF-util.md`. 현재 진행 → `CURRENT.md`.

## 2026-08-07(2) — `GroupedSet` 신설 + 공유 리사이즈 관측 그룹(`feature/common`)
- `common/util/groupedSet.util.ts`: `GroupedSet<K,V>`(`Map<K,Set<V>>` 래퍼) 신설.
- `feature/common/store/resizeGroup.store.ts` + `component/ProvideResizeGroup.vue` 신설: 컴포넌트마다 각자 만들던 `ResizeObserver`/`MutationObserver`를 앱 전체 공유 옵저버 1쌍으로 통합. `GroupedSet`(엘리먼트당 콜백 다중 등록)+`LifecycleRegistry`(lazy 관찰 시작/해제, `flag`로 게이팅) 조합 — `LifecycleRegistry`의 두 번째 소비처.
- `common/hook/vue.hook.ts` 리팩터: `computeResizeSize()` 순수함수 추출, `useResize`에 `watchWindowResize` 옵션 추가(크기 그대로 위치만 바뀌는 케이스 보강), `useElementsChange()` 신설(복수 엘리먼트 scroll/resize/속성 diff 감시). 죽은 코드(`useStateDelay` 주석블록) 정리.
- 시작 트리거는 `App.vue`의 새 루트 `ProvideAppLayout.vue`(→ `history/frontend.md`)가 맡음.
- 소비처: `AppHead.vue`(→ `history/frontend.md`), `StickyBox.vue`(→ `history/widget.md`). 상세 → `REF-util.md` "공유 리사이즈 관측 그룹".

## 2026-08-07 — `EventInterface` + `LifecycleRegistry` 신설, `Memoized` 확장
- `common/util/lifecycle/event.util.ts`: `EventInterface<EventMap>` 최소 pub/sub 베이스 클래스 신설.
- `common/util/lifecycle/lifecycle.util.ts`: `LifecycleRegistry<E>` 신설 — 조건부(lazy) init/release 생명주기 관리자. nexus 프로젝트에서 "범용 유틸 클래스" 반복 주제 첫 등장이라 전용 REF/history로 처음부터 분리.
- `common/util/index.util.ts`의 기존 `Memoized<K,V>`가 `EventInterface` 상속 + `has`/`keys` 추가 + `get`/`remove`에 `create`/`remove` 이벤트 발화 추가하도록 확장됨(`LifecycleRegistry` 내부 구현이 이 이벤트에 의존).
- 첫 소비처: `feature/widget/store/skeletonGroup.store.ts`(Skeleton/SkeletonGroup 위젯) — → `history/widget.md`.
