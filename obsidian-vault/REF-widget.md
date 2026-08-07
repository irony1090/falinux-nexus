# REF: 위젯 컴포넌트 (apps/frontend/src/feature/widget)

> 프로젝트 UI 위젯(재사용 가능하지만 프로젝트 로직과는 결합된 컴포넌트) 전용 문서. 이력 → `history/widget.md`. 기반 유틸(`LifecycleRegistry`) → `REF-util.md`. 프론트 전반 구조 → `REF-frontend.md`.

## Skeleton / SkeletonGroup — shimmer 로딩 위젯 (2026-08-07 추가)

### 설계 배경 (왜 그룹으로 묶었나)
- **이전 방식의 문제**: 스켈레톤 하나당 `@keyframes` 애니메이션을 하나씩 따로 돌리는 구조 — 화면에 스켈레톤 사용 빈도(개수)가 늘어나면 동시에 도는 애니메이션 인스턴스 수도 같이 늘어 성능 이슈가 발생.
- **현재 해법**: `@keyframes loading` 애니메이션을 **그룹당 1개**만 돌리고(`SkeletonGroup`의 `.shimmer-mask::after`), 실제로 "어디에 보여줄지"는 각 `Skeleton`의 사각형을 모은 **SVG `<clipPath>`로 마스킹**해서 해결 — 개별 엘리먼트마다 애니메이션을 켜는 대신 하나의 공유 shimmer surface를 자르는 방식이라 스켈레톤 개수가 늘어도 애니메이션 인스턴스 수는 1로 고정됨.

### 구성
- `component/Skeleton.vue`: 임의 엘리먼트(`:is="component"`, 기본 `div`)를 감싸 `loading` 상태일 때 반투명 오버레이(`opacity` + `::after` 배경색)를 씌우는 wrapper. `loading` prop이 꺼져 있어도 부모 `SkeletonGroup`의 `isLoading`이 켜져 있으면 로딩으로 표시(`computed(() => pLoading.value || isLoading.value)`), 즉 그룹 전체 on/off와 개별 on/off를 OR로 합성.
- `component/SkeletonGroup.vue`: 하위 `Skeleton`들의 공용 로딩 상태·좌표를 provide/inject로 공급하는 컨테이너. `v-model`(`isLoadingModel`)로 그룹 전체를 켜고 끔.
- `store/skeletonGroup.store.ts`: `provideSkeletonGroupStore()` / `useSkeletonGroupStore()` — provide/inject 컨텍스트 팩토리. 그룹 밖에서 `Skeleton` 단독 사용 시 `useSkeletonGroupStore`는 no-op 기본값(`isLoading=false`, `registryElement`/`releaseElement`=noop)을 반환해 안전.

### 동작 흐름
1. 각 `Skeleton`이 mount 시 자기 root element를 `skeletonRef`로 잡아 그룹 store에 `registryElement(el)`(`LifecycleRegistry.put`), unmount/교체 시 `releaseElement`(`LifecycleRegistry.remove`) — → `REF-util.md`.
2. 그룹의 `isLoading`이 켜지면 `LifecycleRegistry.initAll()`이 등록된 모든 엘리먼트에 `ResizeObserver`+`MutationObserver` 관찰을 건다(꺼지면 `releaseAll()`로 일괄 해제) — 로딩 중이 아닐 땐 관찰 비용 자체를 안 씀.
3. 관찰 콜백(`onRefresh`)이 그룹 기준 좌표(`getBoundingClientRect()` 오프셋)로 각 Skeleton의 사각형(`offsetRects`, radius=min(w,h)/2로 pill 형태)을 재계산.
4. `SkeletonGroup`이 이 `offsetRects`로 `<svg><clipPath>` 안에 `<rect>`들을 그리고, 그룹 배경 위 shimmer gradient(`.shimmer-mask::after`, `linear-gradient` sweep 애니메이션)에 그 clip-path를 적용 — 여러 조각난 Skeleton이 마치 하나의 큰 shimmer surface를 공유하는 것처럼 보이는 효과를 냄.
- 색상: `SkeletonGroup`의 `colors`(팔레트 맵) + `color`/`bgColor`/`highlightColor` prop 조합, CSS 커스텀 프로퍼티 `--skeleton-color`로 개별 `Skeleton`에도 전파(`v-bind(color)` scss).

### 참고
- 같은 폴더의 `store/stickyBox.store.ts`(기존, provide/inject 기반 sticky 영역 좌표 공유)는 이번 추가와 무관한 별개 위젯.
- 아직 실제 사용처(어느 페이지/컴포넌트가 이 위젯을 소비하는지) 미배선 — 컴포넌트·스토어만 추가된 상태.

## StickyBox — 중첩 가능한 sticky header/footer 위젯
> 스토어(`store/stickyBox.store.ts`)는 기존, 이번에 `component/StickyBox.vue`(실제 컴포넌트) 추가 + 스토어 리팩터. 관찰 리소스는 공유 그룹(`REF-util.md` "공유 리사이즈 관측 그룹") 소비.

### 하는 일
- 슬롯 `header`/`footer`를 `position: fixed`로 화면에 붙이되, **중첩된 StickyBox끼리 서로의 header/footer 높이만큼 밀어내며 쌓임**(예: 바깥 StickyBox의 header 아래에 안쪽 StickyBox의 header가 붙는 식) — 각 레벨이 자기 차지 영역을 부모에게 보고하고, 부모는 그 최댓값을 다음 레벨의 기준으로 물려줌.

### 스토어 리팩터 — `thisClient`(ref 공유) → `reportSelf`(콜백 보고)
- 기존엔 `provideStickyBox()`가 자기 rect를 담는 `thisClient` ref를 만들고 부모가 그걸 `watch`하는 구조라, "내가 읽는 것"과 "내가 보고하는 것"이 섞여 있었음.
- 지금은 **자식이 `reportSelf(v: Vec4)`를 호출해서만 부모에게 알림** / 부모는 `rootClient`(자기가 물려받은 기준)·`maxClient`(자식들 보고의 누적 최댓값)만 내려줌 — 역할이 읽기/쓰기로 명확히 분리됨.

### `viewportClient` — CSS containing block 보정 (신규 개념)
- `position: fixed`는 원래 브라우저 뷰포트 기준이지만, **조상에 `transform`/`contain: layout` 등이 걸리면 그 조상이 새로운 containing block이 되어버려 `fixed` 자식이 뷰포트가 아니라 그 조상 기준으로 붙는다** — 다이얼로그 카드 안에 StickyBox를 쓰는 경우가 대표적.
- `viewportClient`(rootClient와 동일 규약: `[top,right,bottom,left]` inset, 기본 `null`=진짜 뷰포트 그대로)를 **그 조상 쪽 코드가 직접 갱신**해줘야 함. `rootClient`/`maxClient`(내부에서 자동 누적)와는 역할이 달라 섞어 쓰면 안 됨 — 하나는 "내부에서 자동 계산되는 형제 영역", 하나는 "바깥에서 강제로 알려주는 좌표계 보정".

### `StickyBox.vue` 컴포넌트
- `relation` prop(단일/배열, `useElementsChange`로 감시) — 이 StickyBox 바깥에서 일어나는, 재계산이 필요한 scroll/resize/속성변화의 원천(다이얼로그 카드, `document.documentElement` 등)을 명시적으로 지정.
- 자기 자신·head·foot 세 엘리먼트 각각 공유 리사이즈 그룹(`useResizeCallback`)으로 크기 관찰(`REF-util.md`).
- `refresh()`가 `rootClient`+`viewportClient`+실측 rect로 clamp된 `thisRectClient`(자기 차지 영역)를 계산 → `reportToParent`로 부모에 보고 + head/foot의 `position:fixed` 좌표(`top/right/bottom/left`)로도 사용.
- 재계산 트리거 총 4가지: 자기 크기 변화(`stickyResize`) / `rootClient`·`viewportClient` 변화 / head·foot 실측 높이 확정 시점(`nextTick`) / `relation` 대상들의 변화(`useElementsChange`) — 그리고 창 자체 리사이즈.
