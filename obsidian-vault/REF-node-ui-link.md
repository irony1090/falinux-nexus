# REF — node 카탈로그 타일 UI 실제 연동 (더미 → node/process API)

> 2026-09-30 신설 — 타일 UI의 더미(`tileDummy.ts`)를 실제 API로 바꾸는 작업의 구조·결정. 이력 → `history/node-ui-link.md` / 데이터 층(유틸·스토어) → `REF-node-ui-impl.md` / 컴포넌트 → `REF-node-ui-render.md` / 백엔드 node·worker API → `REF-node-label.md` / 현재 진행 → `CURRENT.md`.

## 연동 1차(7-0) 항목
| 번호(이름) | 내용 | 상태 |
|---|---|---|
| 1(`GET /workers`) | 접속 인스턴스 목록 + `execProcess` 대상 장비 검증 | 완료(미커밋) → `REF-node-label.md` "Exec 인스턴스 선택" |
| 2(폴더 타일 목록) | 폴더 타일의 자식 목록·경로·이름을 node API로 | 완료(vue-tsc 통과 + 사용자 브라우저 확인 2026-09-30, 미커밋) |
| 3(node 관리 UI) | 새 폴더·스크립트 / 이름 변경·삭제 / `device_key` 설정 | 완료(2026-09-30, vue-tsc + 헤드리스 시나리오 통과 + 사용자 확인, 미커밋) |
| 4(스크립트 편집) | 브라우저 편집기 → `PATCH content` 먼저 | 완료(2026-09-30, vue-tsc + 헤드리스 시나리오 통과, 미커밋) |
| 5(실행) | `execProcess` → `addTerminal`, 인스턴스 선택 | 6(`ProcessDialog` 타일 임베드)과 함께 |

## 2(폴더 타일 목록) 구조 (2026-09-30)
새 스토어 없음 — 컴포넌트가 `node.api.ts`의 조회 훅(tanstack query)을 직접 부른다.

| 파일 | 더미 | 바뀐 것 |
|---|---|---|
| `FolderTileBody.vue` | `dummyChildren` | `useListChildren(nodeId ?? undefined)` (`GET /nodes?parentId=`, 없으면 루트) |
| `FolderTileBody.vue` | `dummyPath` | `useNodePath(nodeId)` — 결정 A(경로 조회 방식) |
| `TileFrame.vue` | `dummyNode().name` | `useGetNode` 2개(자기 노드 / 연 폴더 타일의 노드), 받기 전에는 `node {id}` |
| `TerminalTileBody.vue` | `dummyNode().lines` | 출력 줄 제거, 자리 표시 문구만 |
| `tileDummy.ts` | 가짜 노드 9개 + `seedTileTree` | 삭제. `dummyExec`/`dummyKill`/`dummyStatus`만 남음 |
| `pages/index.vue` | `seedTileTree(tileTree)` | 호출 제거 — 루트 폴더 타일 1개로 시작 |

### 결정
| 결정 | 내용 | 이유 |
|---|---|---|
| A(경로 조회 방식) = 프론트에서 부모 따라가기 (사용자 선택) | `fetchNodePath`: `getNode`를 `parentId` 따라 반복 호출, 깊이 상한 64. 서버 경로 API(`GET /nodes/:id/path`)는 만들지 않음 | 백엔드를 그대로 둠 |
| B(더미 시드 처리) = 시드 제거 (사용자 "지금 필요해??") | `seedTileTree` 삭제, 화면은 루트 폴더 타일 1개로 시작 | 시드의 가짜 id 1·2·3·4가 실제 DB 노드와 맞지 않음(id 2는 SCRIPT인데 폴더 타일로 열림). 실제 목록에서 직접 열 수 있어 자동 열기가 필요 없음 |
| C(실행 버튼 범위) = 더미 유지 (사용자 선택) | 실행·kill·상태는 `dummyExec` 등 그대로. 실제 `execProcess`는 5(실행)에서 | 출력 확인이 xterm 임베드에 묶임 |

### 구현 메모
- **경로의 조상은 DETAIL 캐시를 같이 쓴다**: `fetchNodePath`가 `queryClient.fetchQuery(Q_KEY.DETAIL(id))`로 부르므로 `TileFrame`의 `useGetNode`와 같은 캐시 항목. 깊이 N 폴더의 첫 조회만 N번, 이후는 캐시(staleTime 1분).
- **무효화**: `useNodeQueryClient().invalidateAll`에 `['NODE','PATH']` 전체 무효화 추가 — 조상의 이름·위치가 바뀌면 자손 경로가 달라지는데 어느 자손인지 키로 알 수 없어서.
- `useGetNode`는 id가 `NaN`이면 조회하지 않음(기존 규약) → 루트(`nodeId: null`)는 `?? NaN`으로 넘김.
- **목록 갱신**: NODE 실시간 구독 없음(열린 질문 4). 브레드크럼 줄 오른쪽 새로고침 버튼(`refetch`) + 3(node 관리 UI)에서 변경 뒤 무효화.
- **로딩 표시 = `VProgressLinear`(2px)**, `Skeleton` 위젯 미사용: `SkeletonGroup`의 clipPath id가 `CLIP_PATH` 고정이라 폴더 타일 여러 개가 동시에 로딩하면 id가 겹침(→ `REF-widget.md` 위젯 쪽 수정 필요).
- 오류 = 목록 자리에 서버 메시지 + "다시 시도". 기본 `retry: 2`라 401도 재시도 뒤에 표시됨.
- 미리보기는 이제 supervisor + 로그인 필요(세션 없으면 401).

## 3(node 관리 UI) 구조 (2026-09-30)
| 동작 | 화면 위치 | 호출 | 성공 후 |
|---|---|---|---|
| 새 폴더 / 새 스크립트 | 브레드크럼 줄 `+` 메뉴 → `NodeNameDialog` | `useNodeActions().create` | 목록·상세·경로 무효화 |
| 이름 변경 | 행 `⋮`(`NodeRowMenu`) → `NodeNameDialog` | `patch(id, {name})` | 같음 |
| 장비 지정 (폴더만) | 행 `⋮` → `NodeDeviceDialog` | `patch(id, {deviceKey})` | 같음 |
| 장비 지정 (지금 보는 폴더) | 브레드크럼 줄 오른쪽 장비 표시 버튼 → `NodeDeviceDialog` | `patch(id, {deviceKey})` | 같음 |
| 삭제 | 행 `⋮` → 확인 창(`AppDialog`) | `useNodeRemove().removeNode` | 타일 닫기 → 무효화 |

| 파일 | 책임 |
|---|---|
| `component/tile/NodeRowMenu.vue` | 행 `⋮` 메뉴, `rename`/`device`/`remove` emit만 |
| `component/dialog/NodeNameDialog.vue` | 이름 입력(생성·이름 변경 공용), `submit(name)` |
| `component/dialog/NodeDeviceDialog.vue` | 장비 키 직접 입력 + 접속 중 장비 칩에서 고르기 + 접속 상태 문구, `submit(key \| null)` |
| `hook/nodeRemove.hook.ts` | 삭제 + 타일 닫기(API와 타일 트리를 함께 다뤄서 컴포넌트·스토어 밖으로 뺌) |
| `node.api.ts` | `useNodeActions`(`create`/`patch` + 무효화), `useNodeQueryClient().fetchPath` |
| `worker.api.ts` | `useOnlineWorkers` — main_key별 접속 인스턴스 수, 10초 주기 조회 |
| `FolderTileBody.vue` | 위 조립. 창 상태(`nameOpen`/`nameTarget`, `deviceOpen`/`deviceTarget`)는 폴더 타일마다 |

### 결정
| 결정 | 내용 |
|---|---|
| D(조작 메뉴 형태) = 행마다 `⋮` 메뉴 하나 (추천대로) | 폰에서 타일 폭이 좁아 아이콘 3개를 나란히 둘 수 없음 |
| E(삭제된 노드를 보던 타일) = 전부 닫기 (사용자 "하위 폴더는 전부 닫아버려") | 삭제된 폴더나 그 하위를 보는 폴더 타일을 `close`. 루트 타일은 닫을 수 없어 기본 목록으로 `navigate`. 닫힌 타일이 연 타일들은 기존 닫기 규칙대로 부모 소속이 됨 |
| F(장비 후보 범위) = 직접 입력 + 접속 중 목록 (사용자: worker보다 먼저 지정하고 상태를 볼 수 있어야 함) | 폴더 행에 장비 키 + 점(접속 중 = `success` 색). 미접속 키도 저장 가능 |

### 구현 메모
- **재발 방지 — 삭제의 순서 의존**(`removeNode`): ① 닫을 타일은 **삭제 전에** 구한다(경로를 `getNode`로 따라가므로 삭제 후엔 404) ② 무효화는 타일 닫기 뒤 `nextTick` 다음(먼저 하면 닫힐 타일이 삭제된 노드를 다시 조회).
- 삭제 확인 창: 노드 이름은 `title`로만 넘김(`AppDialog`의 `content`는 `v-html`). 삭제 버튼에 Enter를 묶지 않음 — 묶지 않은 Enter·Space는 `AppDialog`가 닫기로 처리.
- 접속 상태는 실시간 알림이 없어 주기 조회(10초) + 새로고침 버튼. 같은 쿼리 키라 타일·창이 몇 개든 요청은 하나.
- **지금 보는 폴더의 장비 표시(사용자 추가 요청)**: 브레드크럼 줄 오른쪽에 `점 + 키`, 없으면 `장비 없음`. 실제 적용 장비 = 경로(`useNodePath`)를 뒤에서부터 훑어 처음 나오는 `deviceKey`(서버 `ResolveDeviceKey`와 같은 규칙, 추가 요청 없음). 상위 폴더 것이면 `(상속)` 표시 + 툴팁에 폴더 이름. 누르면 **지금 보는 폴더 자신의** 장비를 지정(상속 중이면 입력칸은 빈 값). 기본(루트) 목록에는 표시 없음.
- **확인(2026-09-30, 헤드리스 Chromium 1440×900, 실제 supervisor·worker)**: 임시 폴더 생성 → 들어가면 `장비 없음` → 미접속 키 지정(회색 점, "접속 안 됨") → 하위 폴더·스크립트 생성, 하위 폴더를 새 타일로 → `(상속)` 표시 → 상위를 접속 중 키로 바꾸면 하위 타일도 초록 점 → 하위 타일에서 자기 키 지정 → 목록에서 이름 변경하면 타일 제목·브레드크럼 반영 → 상위 삭제 시 하위 타일 닫힘, DB 원상 복구 확인. 폰 390×844 화면 배치도 확인.
- 범위 밖: 노드 이동·순서 변경, 삭제되는 스크립트의 실행 중 process 처리, 백엔드 부모 폴더 소유자 검증(→ `CURRENT.md` Node 모듈 잔여).

## 4(스크립트 편집) 구조 (2026-09-30)
스크립트 내용 = `nodes.content`(worker 파일 아님) → 브라우저에서 고쳐 `PATCH /nodes/:id`. 나중에 붙는 EDIT(worker `vi` PTY, `REF-process-exec-edit.md`)는 같은 내용을 고치는 **별도 경로**이고 6(타일 xterm 임베드) + input 배선이 선행.

| 단계 | 동작 |
|---|---|
| 열기 | 스크립트 행 `⋮` → "내용 편집", 또는 새 스크립트 생성 직후 자동(결정 I). 열 때마다 `getNode`로 새로 조회(목록 캐시는 최대 1분 묵음) |
| 편집 | `ScriptEditDialog.vue` — 제목(이름 + `수정됨`) + `VTextarea`(고정폭, 줄바꿈 없음, Tab = 탭 문자) |
| 저장 | 저장 버튼 = 저장 후 닫기 / Ctrl+S(⌘S) = 저장하고 계속 편집. 빈 내용은 `null` |
| 닫기 | Esc·바깥 클릭·닫기 버튼 공통, 고친 내용이 있으면 `AppDialog` 확인(계속 편집 / 저장하지 않고 닫기) |

| 결정 | 내용 |
|---|---|
| G(편집기 위치) = 창 (추천대로) | PC 폭 760·높이 640, 폰(`xs`)은 전체 화면. 편집기 타일은 타일 종류·⑪(타일 트리 서버 저장)에 영향이라 기각 |
| H(편집기 종류) = `VTextarea` (추천대로) | 새 라이브러리 없음. 본격 편집은 EDIT(`vi`)가 맡음 |
| I(새 스크립트 직후 동작) = 바로 편집 창 (추천대로) | 빈 스크립트는 실행해도 할 일이 없음. 새 폴더·이름 변경은 목록만 갱신 |

- **함정 — 창 높이는 `VDialog`의 `height`로 준다**: 카드는 `.v-overlay__content` 안의 flex 항목이라 카드에 준 `height`는 무시됨(실제 발생: 640 지정했는데 226px).
- `VTextarea`를 남는 높이에 채우기: `.v-input`의 `grid-template-rows: minmax(0, 1fr)` + `.v-input__control`/`.v-field`/`.v-field__field`/`textarea`에 `height: 100%`.
- **한계**: 저장 시 다른 곳의 변경을 검사하지 않음(나중에 저장한 쪽이 남음). 막으려면 백엔드 버전 검사 필요.
- **확인(2026-09-30, 헤드리스 PC 1440×900 / 폰 390×844)**: 새 스크립트 → 편집 창 자동 열림 → 입력(Tab 포함) → Ctrl+S 후 창 유지·`수정됨` 사라짐 → 더 고치고 Esc → 확인 창 → 계속 편집 → 저장 버튼 → 닫힘 → `⋮`로 다시 열면 저장된 내용 → 고치고 "저장하지 않고 닫기". DB `content`에 탭·줄바꿈 그대로 저장 확인.

