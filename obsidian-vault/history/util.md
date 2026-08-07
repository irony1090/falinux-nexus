# HISTORY: 범용 유틸 (apps/frontend/src/common/util)

> 요약/재사용 지식 → `REF-util.md`. 현재 진행 → `CURRENT.md`.

## 2026-08-07 — `EventInterface` + `LifecycleRegistry` 신설, `Memoized` 확장
- `common/util/lifecycle/event.util.ts`: `EventInterface<EventMap>` 최소 pub/sub 베이스 클래스 신설.
- `common/util/lifecycle/lifecycle.util.ts`: `LifecycleRegistry<E>` 신설 — 조건부(lazy) init/release 생명주기 관리자. nexus 프로젝트에서 "범용 유틸 클래스" 반복 주제 첫 등장이라 전용 REF/history로 처음부터 분리.
- `common/util/index.util.ts`의 기존 `Memoized<K,V>`가 `EventInterface` 상속 + `has`/`keys` 추가 + `get`/`remove`에 `create`/`remove` 이벤트 발화 추가하도록 확장됨(`LifecycleRegistry` 내부 구현이 이 이벤트에 의존).
- 첫 소비처: `feature/widget/store/skeletonGroup.store.ts`(Skeleton/SkeletonGroup 위젯) — → `history/widget.md`.
