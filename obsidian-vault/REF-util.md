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

## `groupedSet.util.ts` — `GroupedSet<K, V>`
- 위치: `apps/frontend/src/common/util/groupedSet.util.ts`
- `Map<K, Set<V>>` 래퍼 — 키 하나에 값 여러 개(중복 없이) 묶어서 관리. `put`/`remove`(set 비면 키 자체 삭제)/`has`/`forEach(key, fn)`.
- 프로젝트 로직과 무관한 범용 자료구조. 첫 소비처는 아래 "공유 리사이즈 관측 그룹"의 `callbacks`(엘리먼트 하나에 콜백 여러 개 붙는 경우).

## `lifecycle.util.ts` — `LifecycleRegistry<E>`
- 위치: `apps/frontend/src/common/util/lifecycle/lifecycle.util.ts`
- `EventInterface<{registered, unregistered, inited, released}>` 상속.
- 내부에 `Memoized<E, {init, release, isInit}>` 하나를 들고, 키(`E`, 보통 `HTMLElement`)마다 `init`/`release` 콜백 쌍을 lazy 생성.
- API:
  - `put(el)` / `remove(el)` — 엘리먼트 등록/해제(`remove`는 등록돼 있으면 먼저 `release()` 호출 후 캐시에서 제거).
  - `initAll()` / `releaseAll()` — 내부 `flag`를 켜고/끄고, 아직 안 된 엔트리들만 일괄 `init()`/`release()`(이미 상태가 같으면 스킵 — 중복 호출 안전).
  - `registered` 이벤트 리스너에서 "이미 `isLoading`이면 방금 등록된 것도 바로 init" 패턴으로 사용(아래 소비처 참고).
- **용도**: 비용이 드는 감시 작업(예: `ResizeObserver`/`MutationObserver` 구독)을 "로딩 중" 같은 조건이 켜져 있을 때만 걸고, 꺼지면 한꺼번에 해제하기 위한 범용 lazy 생명주기 관리자. 프로젝트 로직(스켈레톤·shimmer)과 무관 — 어떤 "조건부로 관찰/해제해야 하는 엘리먼트 집합"에도 재사용 가능.
- 첫 소비처: `feature/widget/store/skeletonGroup.store.ts` (→ `REF-widget.md`). 두 번째 소비처는 아래 "공유 리사이즈 관측 그룹".

## 공유 리사이즈 관측 그룹 — `feature/common/store/resizeGroup.store.ts` + `component/ProvideResizeGroup.vue`
> `common/util`이 아니라 `feature/common`에 있지만, 프로젝트 로직과 무관한 범용 성능 패턴이라 여기 기록. 소비처는 `REF-frontend.md`(AppHead)·`REF-widget.md`(StickyBox).
- **문제**: 컴포넌트마다 각자 `ResizeObserver`+`MutationObserver`를 새로 만들던 기존 `useResize` 방식은, 리사이즈 감시가 필요한 컴포넌트 수가 늘면 옵저버 인스턴스 수도 그만큼 늘어남.
- **해법**: 앱 전체에 **공유 옵저버 1쌍**(`resizeObserver`/`mutationObserver`)만 두고, 각 컴포넌트는 `useResizeCallback(elRef, callback)`으로 자기 엘리먼트+콜백만 등록. 내부 구조:
  - `GroupedSet<HTMLElement, () => void>`(`callbacks`) — 엘리먼트 하나에 콜백 여러 개(여러 컴포넌트가 같은 엘리먼트를 볼 수도 있음) 매핑.
  - `LifecycleRegistry<HTMLElement>`(`cycle`) — 실제 `observe`/`unobserve` 호출(비용 있는 부분)을 엘리먼트 단위로 lazy 관리. `flag`(그룹 활성 여부)가 켜져 있어야 `registered` 이벤트에서 바로 `initAll()`.
  - `mutationManager`: 네이티브 `MutationObserver`엔 대상별 `unobserve`가 없어서, 하나 뺄 때 **전체 `disconnect()` 후 남은 대상 전부 재관찰**하는 방식으로 우회.
- **시작 트리거**: `ProvideAppLayout.vue`(App.vue 루트, `v-app`을 감싸는 래퍼)가 마운트 시 `provideResizeGroupStore()` 호출 + `flag.value = true`로 앱 전체 관찰을 1회 킥. 그 전에 등록된 엘리먼트는 대기하다 이 시점에 한꺼번에 `initAll()`됨.
- **`vue.hook.ts` 쪽 변화**(공유 그룹 도입에 맞춘 리팩터):
  - `computeResizeSize(el)`: 기존 `useResize` 내부 계산 로직을 순수 함수로 추출 — 공유 콜백 안에서 각 컴포넌트가 직접 호출해 자기 `size` state를 갱신.
  - `useResize`에 `watchWindowResize` 옵션 추가(기본 꺼짐) — `ResizeObserver`/`MutationObserver`는 **크기는 그대로인데 위치(rect)만 바뀌는 경우**(예: `max-width`에 걸려 폭 고정, 중앙정렬 offset만 이동)를 못 잡음. rect 정확도가 중요한 호출부만 켜서 `window resize` 리스너를 추가 비용으로 문다.
  - `useElementsChange(targets, onChange)` 신설 — 배열(ref/computed)로 주어진 여러 엘리먼트의 scroll/resize/속성변화를 한 콜백으로 감시, `targets` 변경 시 필요한 것만 attach/detach(diff). `StickyBox`의 `relation` prop(다이얼로그 카드 등 바깥 재계산 트리거)이 이걸 씀.
