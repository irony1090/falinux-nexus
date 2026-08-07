# REF: 범용 유틸 (apps/frontend/src/common/util)

> 프로젝트 로직과 무관한 재사용 가능한 유틸리티 전용 문서(전역 규칙 — 반복 주제는 프로젝트 고유 REF와 분리). 이력 → `history/util.md`. 첫 소비처(위젯) → `REF-widget.md`.

## `event.util.ts` — `EventInterface<EventMap>`
- 위치: `apps/frontend/src/common/util/lifecycle/event.util.ts`
- 최소 pub/sub 추상 베이스 클래스. `EventMap`은 `{ [type]: 인자튜플 }` 형태.
- `on(type, func)`: 핸들러 등록, **타입당 슬롯 1개(덮어씀, Set 아님)** — 마지막 `on()` 호출만 유효. 해제 함수(자기 자신 참조 시에만 초기화) 반환.
- `emit(type, ...args)` (protected): 서브클래스 내부에서만 발화.
- `Memoized`(`index.util.ts`)와 `LifecycleRegistry`(아래)가 상속해서 사용.

## `index.util.ts` — `Memoized<K, V>` (기존 클래스, 이번에 확장)
- 최초 도입은 초기 스캐폴딩(`e511359`)의 단순 캐시(`get`/`remove`)였음.
- **오늘 확장(2026-08-07)**: `EventInterface<{create, remove}>` 상속 + `has(k)`/`keys()` 추가, `get`/`remove`가 캐시 미스 채움/삭제 시 `create`/`remove` 이벤트 발화.
- 목적: `LifecycleRegistry`가 "새 엘리먼트가 처음 등록됐다"를 이벤트로 감지해 초기화 훅을 걸 수 있게 함.

## `lifecycle.util.ts` — `LifecycleRegistry<E>`
- 위치: `apps/frontend/src/common/util/lifecycle/lifecycle.util.ts`
- `EventInterface<{registered, unregistered, inited, released}>` 상속.
- 내부에 `Memoized<E, {init, release, isInit}>` 하나를 들고, 키(`E`, 보통 `HTMLElement`)마다 `init`/`release` 콜백 쌍을 lazy 생성.
- API:
  - `put(el)` / `remove(el)` — 엘리먼트 등록/해제(`remove`는 등록돼 있으면 먼저 `release()` 호출 후 캐시에서 제거).
  - `initAll()` / `releaseAll()` — 내부 `flag`를 켜고/끄고, 아직 안 된 엔트리들만 일괄 `init()`/`release()`(이미 상태가 같으면 스킵 — 중복 호출 안전).
  - `registered` 이벤트 리스너에서 "이미 `isLoading`이면 방금 등록된 것도 바로 init" 패턴으로 사용(아래 소비처 참고).
- **용도**: 비용이 드는 감시 작업(예: `ResizeObserver`/`MutationObserver` 구독)을 "로딩 중" 같은 조건이 켜져 있을 때만 걸고, 꺼지면 한꺼번에 해제하기 위한 범용 lazy 생명주기 관리자. 프로젝트 로직(스켈레톤·shimmer)과 무관 — 어떤 "조건부로 관찰/해제해야 하는 엘리먼트 집합"에도 재사용 가능.
- 첫 소비처: `feature/widget/store/skeletonGroup.store.ts` (→ `REF-widget.md`).
