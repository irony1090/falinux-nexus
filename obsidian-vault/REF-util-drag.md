# REF: 드래그 인프라 (`DragSession`, 옛 설계 DraggableSession / useDragGhost / GhostArea)

> `REF-util.md`(범용 유틸)에서 10k자 기준 분할(2026-08-07) — 드래그/고스트만 별도 하위주제로 분리. 나머지 범용 유틸(`EventInterface`/`Memoized`/`LifecycleRegistry`/`GroupedSet`/공유 리사이즈 관측 그룹)은 `REF-util.md`. 이력 → `history/util-drag.md`. 소비처(Node UI 배선) → `REF-node-ui.md`.

## `DragSession` (구현 2026-10-29, 파트(타일 끌기) 끌기-1(끌기 기반))
> 아래 옛 설계(DraggableSession/useDragGhost/GhostArea)의 "다시 볼 조건"(타일 순서 드래그)이 와서 소비처에 맞춰 다시 만들었다. 옛 코드 4개(`drag`·`longPress`·`draggable` listener, `draggableLogic`)는 삭제
- 위치: `apps/frontend/src/common/listener/dragSession.ts` / 첫 소비처 = `feature/node/store/tileDrag.store.ts`(→ `REF-tile-drag.md`)
- 하는 일 = **시작 판정 + 포인터 추적만.** 놓을 자리·고스트 그리기는 소비처(옛 설계 문제 3·4 피함)

| 항목 | 내용 |
|---|---|
| 시작 판정 | 마우스 = `threshold`(6px) 움직이면 / 터치·펜 = `holdMs`(500) 누르고 있으면, 그 전에 `holdSlop`(10px) 움직이면 취소(= 스크롤) |
| 이벤트(`EventInterface`) | `hold`(터치 누르기 시작, down 안에서 바로 나옴 — 소비처 상태를 down 전에 채울 것) · `start` · `move` · `drop` · `cancel`(시작 전 취소 포함) |
| 쓰는 법 | 소비처 pointerdown에서 대상이면 `down(e)` → 진행 중이 아니면 true. `state` = idle/pending/dragging |
| 취소 | Esc(dragging만) · pointercancel(브라우저가 스크롤로 가져감) · 다른 포인터 down(두 번째 손가락) · 창 blur · `cancel()` |
| 터치 스크롤 막기 | `blockTouch(root)` — 비수동 `touchmove`를 **미리** 걸어 둠(dragging일 때만 preventDefault). 터치 중간에 거는 리스너는 Chrome compositor가 모름. **+ 터치 down 때 손가락 댄 요소 자체에도** 같은 리스너(끝나면 뗌) — 아래 함정 |
| 먹기 | 누르는 동안 contextmenu(길게 누르기 기본 메뉴) · 끈 직후 400ms click·contextmenu(손 뗀 자리 클릭 방지) |
| 정리 | `dispose()`(먹는 리스너는 첫 down에서 걸고 여기까지 둠) |
- **함정(10-29 폰 실측)**: 끄는 중 손가락 댄 요소가 다시 마운트돼 문서에서 빠지면 `touchmove`는 **빠진 그 요소로만** 가고 root·window까지 안 올라온다 -> 막히지 않아 브라우저가 스크롤로 가져감 -> `pointercancel`로 끌기 끊김. 소비처가 끄는 동안 DOM을 옮기면(타일이 다른 Grid로) 반드시 생김. 해법 = 시작 요소에 직접 리스너
- 함정: `window.addEventListener`를 변수에 담아 부르면 this가 빠져 Illegal invocation → 분기로 직접 호출

## (옛 설계) DraggableSession / useDragGhost / GhostArea (설계 확정 2026-08-07 → 보류 2026-10-12 → **10-29 위 `DragSession`으로 대체**)
> **보류 근거**: 첫 소비처였던 노드 캔버스 좌표 드래그·트리 드래그 이동이 2026-08-10 타일 레이아웃 전환으로 폐기됨(`position_x/y` 사용 중단). 타일 크기도 메뉴(0.5/1)라 리사이즈 드래그 없음 → 지금 쓸 곳이 없고, 소비처 없이 만들면 설계를 검증할 수 없음
> **다시 볼 조건**: 드래그 UX를 넣기로 할 때 — 열린 질문 1(분할 UX, VSCode식 드래그 앤 드롭) / 타일 순서 드래그 변경. 그때 아래 설계를 그 소비처 요구에 맞춰 재검토(위임 처리·롱프레스가 실제로 필요한지 등)
> 구 `draggable.listener.ts`·`draggableLogic.ts`는 nexus에서 여전히 미사용 상태로 남아 있음
> Node 카탈로그 UI Phase2(캔버스 좌표 드래그)/Phase3(트리 드래그 이동) 착수 전 선행 리팩터. 논의 계기 → node UI 계획 중 "구 드래그 코드는 만든 사람도 재활용 못 할 만큼 복잡하다"는 자체 진단.

### 구코드(test-jig) 진단
- `common/listener/draggable.listener.ts`(`DraggableListener`)+`draggable/draggableLogic.ts`(`DraggableLogic`+2 서브클래스)는 nexus에 손 안 댄 채 이식만 돼있고(diff 0) 소비처 全무. `feature/layout/{component/GhostArea.vue,store/ghost.store.ts}`는 아직 이식 전.
- 문제 4가지:
  1. **단일슬롯 콜백 7개**(`onHold`/`onType`/`onPointer`.../`onAround`) — 재할당 시 이전 리스너 조용히 증발, 멀티 컨슈머 불가.
  2. **판정→추적 로직을 synthetic event 재발사로 봉합** — `DraggableLogic.onDown()`이 가짜 `MouseEvent`/`TouchEvent`를 `hold`에 dispatch해서 새로 만든 `DragListener`의 down 핸들러를 트리거.
  3. **범용 드래그 상태기계에 "이웃 재정렬" 도메인 로직 내장** — `Position`/`itemPositions`/`setupAround`(TOP/RIGHT/BOTTOM/LEFT 이웃거리 계산)까지 `DraggableListener` 안에 있어, 자유배치만 필요한 소비처도 그 비용을 짊어짐.
  4. **Ghost 추종 로직이 소비처마다 손코딩 중복** — `DesktopIcon.vue`/`DesktopWindow.vue` 둘 다 `watch(ghost)`/`watch(pointer)` → `style.left/top` 직접 대입 블록을 거의 그대로 복붙.
  - 부수: 앵커 매칭이 `event.target` 정확일치(`closest()` 아님 — 타이틀바 자식 클릭 시 드래그 미시작 가능) / `desktop.store.ts`에 관성·플릭 계산 죽은 코드 40+줄.

### 재설계 방향
```
DragListener (그대로 유지 — 순수 포인터 추적(down/move/up→좌표)은 문제없음)
        │  down 처리부를 외부에서 직접 호출 가능한 public 메서드로 분리
        │  → StartPolicy가 synthetic dispatch 없이 바로 호출
        ▼
StartPolicy (구 DraggableLogic 클래스 계층 → 정책 함수 2종으로 대체)
  ├─ createThresholdPolicy(maxDiff)  — 이동거리 임계값 넘으면 확정
  └─ createLongPressPolicy(delay)    — 롱프레스로 확정(progress 이벤트 포함)
        ▼
DraggableSession extends EventInterface<{start, move, end, progress}>
  - 컨테이너+selector 이벤트 위임(closest() 매칭)은 유지 — Node 캔버스처럼
    아이콘 여러 개를 한 컨테이너가 위임 처리해야 하는 요구는 실재
  - Position/neighbor(이웃 재정렬) 계산은 드롭 — Node UI엔 불필요한 도메인 로직
  - 콜백 슬롯 7개 → `EventInterface` on/emit(`LifecycleRegistry`와 동일 패턴, → `REF-util.md`)
```
- **useDragGhost 컴포저블**: `(session, elRef) => { isDragging, selfStyle, ghostStyle }`. `session`의 `start`에서 `elRef` clone → `useGhostArea().ghostRef`에 꽂고 offset 보정 계산 / `move`에서 좌표 갱신 / `end`에서 정리 — Icon/Window에 중복돼있던 watch 블록 두 벌을 이 훅 하나로 흡수.
- **GhostArea 포팅**: `feature/layout/component/GhostArea.vue`+`store/ghost.store.ts` — test-jig 것을 거의 그대로(body레벨 teleport 타겟 `#Ghost`, z-index 9999, `v-show="!!ghostRef"`). 배치는 `AppDialog`(전역 오버레이+provide/inject)와 동일한 자리(`feature/layout`)로 결정, **`provideAppLayout.vue`에서 `provideResizeGroupStore()`와 나란히 1회 kick** 예정(→ `REF-frontend.md` "앱 레이아웃 루트"). 대안(teleport 없이 로컬 fixed)도 검토했으나 overflow:hidden 컨테이너(Node 캔버스 등) 안에서 잘릴 위험 때문에 포팅 쪽으로 확정.

### 파일 계획 (다음 세션 작성 예정, 코드 없음)
- `common/listener/draggable/draggableSession.ts` — **신규, 구 `draggable.listener.ts`+`draggableLogic.ts` 대체**(둘 다 nexus 미사용 확인됨 → 교체, 존치 안 함)
- `common/hook/dragGhost.hook.ts` — `useDragGhost`
- `feature/layout/component/GhostArea.vue` + `feature/layout/store/ghost.store.ts` — 포팅
