# REF: frontend (apps/frontend)

> supervisor용 프론트엔드. 상세 이력 → `history/frontend.md` / 로그인·API·다이얼로그 → `REF-frontend-auth.md`(+resize/다이얼로그 `history/process-resize.md`) / 현재 진행 → `CURRENT.md`
> node 카탈로그 관련 백엔드 설계 → `REF-node-label.md`. node 카탈로그 UI 컨셉(배치도/트리/모바일) → `REF-node-ui.md`. process resize 백엔드 배선 → `REF-process-resize.md`.

## 위치/정체성
- 경로: `apps/frontend` (package name `frontend-nexus-supervisor`)
- 커밋: `e511359`(초기 스캐폴딩) → `3a8e92e`(socket hook 재설계 + user/login·다이얼로그 WIP) → `e28252b`(hook 재연결 견고성) (모두 2026-06-30)

## 앱 이름·로고 (2026-10-07)
- 이름 = **Nexus**(`index.html` `<title>`). 로고 = `src/assets/logo.svg`(헤더 `AppHead.vue`) = `public/favicon.svg` + `public/favicon.ico`(16·32·48, ImageMagick으로 생성)
- 모양: 어두운 바탕(`terminal` 색 #10161e) 위 타일 2×2 + 가운데 허브 원. 켜진 칸(하늘색, `>_`) = 실행 중 터미널, 허브 = supervisor. 로고는 정적 파일이라 hex 사용(색 규칙의 예외 — 컴포넌트 색이 아님). 기본 Vuetify `logo.png`는 `HelloWorld.vue`(미사용)만 씀

## 스택 확정 (불변)
- **Vue3 + TypeScript + Vuetify 4** (`npm create vuetify`, **scratch** preset)
- **Vue Router (standard, 코드 정의 방식)** — file-based 아님
- **Pinia 미사용** — 의도적으로 뺌
- **CSS 프레임워크 없음** (None) — Vuetify 컴포넌트/유틸리티만
- 빌드: Vite, dev 포트 `3000`
- alias `@` → `src`

## 상태관리 방침
- 1차: **provide/inject** (Vue 내장 DI). 트리 한정 공유엔 충분.
- 한계: **컴포넌트 밖(소켓 핸들러·router 가드·.ts 모듈)에서 inject 불가**.
  → 소켓 실시간 node delta 반영처럼 컴포넌트 밖 갱신이 필요해지면
  **`reactive` 모듈 패턴**(별도 모듈에서 `reactive` 객체 export) 채택 검토.
  의존성 0 + 컴포넌트 밖 접근 동시 확보(단 devtools/HMR/SSR 미지원). Pinia는 규모·디버깅 중시 시 대안.

## 폰트
- 기본 폰트 = **Noto Sans Korean**.
- 로드: `public/fonts/` 정적 에셋 + `src/styles/noto-sans-korean.scss`(`@font-face`), `main.ts`에서 import.
- 적용: Vuetify `$body-font-family` 오버라이드(`src/styles/settings.scss`의 `@use 'vuetify/settings' with (...)`).
- ✅ Roboto 정리 완료(2026-06-30): `@fontsource/roboto` 삭제, `main.ts` `import 'unfonts.css'` 주석화, `vite.config.mts` unplugin-fonts `Fonts()` 주석화. 기본폰트는 `settings.scss` `$body-font-family` 오버라이드로 적용 확인.
- weight 200/600 `@font-face` 없음 → 사용 시 브라우저 합성(faux). 필요 시 추가.

## scss 구조
- `_variables.scss`: 프로젝트 공용 토큰(spacing/weight/alpha 함수). vite `additionalData`로 전 scss에 주입.
- `settings.scss`: Vuetify 내부 변수 오버라이드 전용(`configFile`로 연결). 폰트 등은 여기서.
- 둘 역할 분리 — 혼용 금지.

## 색 규칙 — Vuetify 컬러 네이밍 의존 (2026-09-29 사용자 지시, 불변)
- **방향(확정)**: node 타일 데모 레이아웃부터 이후 모든 레이아웃은 **Vuetify 테마 색 이름(`primary`/`surface`/`on-surface`/`warning`…)에 최대한 의존**. hex·`rgba(0,0,0,…)` 직접 지정 금지.
- **쓰는 순서(2026-09-29 확정)**
  | 순위 | 수단 | 예 |
  |---|---|---|
  | 1 | 컴포넌트 `color` prop | `<v-chip color="success" variant="tonal">`, `<v-btn color="primary">` |
  | 2 | 유틸 클래스 | `text-warning` `bg-surface` `text-medium-emphasis` `border-b` |
  | 3 | scoped CSS의 테마 변수 | `rgb(var(--v-theme-primary))`, `rgba(var(--v-border-color), var(--v-border-opacity))` — hover 등 1·2로 안 될 때만 |
  | 4 | 테마에 없는 색 | `plugins/vuetify.ts` 테마에 커스텀 색 등록(light/dark 양쪽) 후 1~3으로 사용 |
- **등록된 커스텀 색**: `terminal`/`on-terminal`(light·dark 동일한 어두운 배경, `plugins/vuetify.ts`). 등록하면 `bg-terminal`/`text-terminal` 유틸 클래스가 자동 생성됨. xterm 테마도 나중에 이 값을 읽을 후보.
- **주의**: Vuetify는 `on-*` 색에 `bg-*` 클래스를 만들지 않음(`.on-surface` 글자색 클래스만 있음) → 반전 배지(`TileFrame` `.pos`)는 3순위 CSS 유지. `text-medium-emphasis`는 `on-background` 기준이므로 `bg-terminal` 같은 다른 배경 위에서는 `rgba(var(--v-theme-on-X), var(--v-medium-emphasis-opacity))`로 씀.
- **타이포 클래스(Vuetify 4)**: `text-caption` 등 MD2 이름 없음 → MD3 이름(`text-body-small` 12px, `text-label-small` 11px …). 없는 클래스는 조용히 무시되니 주의.
- **2026-09-29 정리 완료(위반 3곳)**: `TerminalTileBody.vue` hex → `bg-terminal` + `on-terminal` 변수 / `settings.scss` `.wk-scollbar` 트랙 → `on-surface` 0.1 / `TileFrame.vue` 상태 배지 → `v-chip color="success" variant="tonal"`, 종류 아이콘 → `v-icon :color`, 여는 폴더 이름·실제 표시 크기 → `text-warning`. vue-tsc·vite build 통과. `FolderTileBody.vue`의 `.folder`/`.script` 등은 3순위(테마 변수)로 남아 있음(규칙 위반 아님).

## WebSocket hook (socket 전송계층) — `src/common/websocket/websocket.hook.ts`
> 실시간 push 아키텍처(인가/라우팅·토픽·서버측) → `REF-realtime.md`. 여기는 프론트 hook 자체.

- **Go `transport.Conn` 1:1 대응**: `call`↔Call(REQ→RES) / `emit`↔Emit(EVENT↑) / `on`↔On(EVENT↓) / `status`=watch 기반(Serve 루프 대용).
- `createWebsocketHook(url, opts)` → **url당 단일 ctx 공유**(소켓·correlator·핸들러).
- **설계 결정(2026-07-21, 구현 완료)**: 반환 형태를 `useX(): SocketContext` 단일 함수 → **`[provideSocket, useSocket]` 튜플**로 변경. 기존엔 모듈 스코프 `let ctx`가 첫 호출 시점에 암묵적으로 build()돼 생명주기가 불투명 → `appDialog.store.ts`와 동일한 provide/inject 패턴(`provide(KEY, ctx)` + `inject<T>(KEY)!` 후 `if(!context) throw`)으로 초기화 지점을 명시화(root에서 `provideSocket()` 1회 호출 → 하위 트리 `useSocket()`).
  - **주입 키 = `Symbol()`(문자열 아님)** — `appDialog`는 앱에 인스턴스가 하나뿐이라 문자열 키로 충분하지만, **소켓은 여러 인스턴스(복수 url)를 만들 예정이 확정**돼 있어 팩토리 호출마다 새로 생성되는 `Symbol()`로 인스턴스 간 키 충돌을 원천 차단.
  - 기존 `useSocket` 내부의 `onUnmounted` 로컬 구독 해제 로직은 그대로 유지.
- `SocketContext` = `{ status, event, connect, disconnect, call, emit, on }`.
- **`Protocol` 주입형(frame 가변 전제)**: 코어는 프레임 필드를 모르고 "의도(request/event) → 와이어" + "와이어 → 분류(Inbound: response/event/request/unknown)"만 위임. frame 바뀌면 **Protocol 구현 1개 교체**. 기본 `jsonFrame` = Go Frame `{k,id,t,e,d}`(REQ0/RES1/EVENT2) 미러.
- 구현 결정(확정):
  - `binaryType='arraybuffer'`(Go BinaryMessage 수신) — 전송계층 설정.
  - **correlator**: `++seq` number id, `pending Map<id>`. **끊김(onclose/disconnect) 시 pending 일괄 reject**(Go corr.Close 대응).
  - `call`: 미연결 시 즉시 reject. `callTimeout`(기본 10000, 0=무제한) + per-call `timeout`/`AbortSignal`.
  - `emit`: 미연결 시 조용히 drop.
  - `on`: type당 **다중 핸들러(Set)**, 해제 함수 반환 + `onUnmounted` 자동 해제.
  - **브라우저에 `Handle` 없음**(서버→브라우저 REQ 미수신, 비대칭).
- 검증: 3모드 e2e 통과(2026-06-30) — `REF-realtime.md`/`history/realtime.md`.
- **`WireData` = `string | Blob | BufferSource`(2026-10-21)**: VSCode 내장 TS 6.0.3은 `send`에서 `SharedArrayBuffer`를 빼서 옛 `ArrayBufferLike` 포함 타입이 에디터에서만 에러(프로젝트 TS 5.9.3은 통과). 지금 타입은 두 버전 다 통과

## user/login · 공용 API · 전역 다이얼로그 → `REF-frontend-auth.md`
- 로그인 상태(모듈 단일 값)·라우터 가드·401 처리·AppDialog 주의(content = v-html)·회원가입 보류 메모는 분리 파일에

## REST API 클라이언트 (`feature/{node,process}/api`, 2026-07-21)
- `feature/node/api/node.api.ts`: `createNode`/`listChildren`/`getNode`/`patchNode`/`deleteNode` (node.go+nodeDto.go 미러) + vue-query `useListChildren`/`useGetNode`/`useNodeQueryClient`(invalidateAll/-Force). `Q_KEY`는 `LIST(parentId?)`/`DETAIL(id)`.
  - `PatchNodeRequest`는 백엔드 `patch.Field[T]`({valid,value})를 호출부에 노출하지 않는다 — 평범한 `T|null|undefined` 값만 받고 `patchNode()` 내부의 `toPatchField()`가 undefined→`{valid:false}` / null·값→`{valid:true,value}`로 변환(관리 부담 축소가 목적).
  - Response/Dto 패턴은 `user.api.ts`와 동일: camelCase `XResponse`(Date) + `Replace`로 파생한 `XResponseDto`(unix sec) + 변환 함수.
- `feature/process/api/process.api.ts`: `listSubscriptions`/`subscribeProcess`/`unsubscribeProcess`/`execProcess`/`killProcess`/`resizeProcess`(2026-07-22 추가, `POST /processes/resize/:processId`) (processApi.go 미러).
  - **선행 배선**: `listSubscriptions`가 원래 DTO 없이 `superdb.Process`(json 태그 없는 sqlc raw, PascalCase+RFC3339)를 그대로 내려보내던 걸 발견 → **백엔드에 `processDto.go` 신설**(`processResponse`+`newProcessResponse(s)`, nodeDto.go와 대칭)해 camelCase+unix sec로 정정한 뒤에 프론트 타입을 작성함(2026-07-21). 순서: 백엔드 응답 정합 먼저 → 프론트 타입은 그 위에.
  - `execProcess` 반환 타입은 사용자가 직접 `{uid}`→`ProcessResponse`로 바꿈(백엔드가 `newProcessResponse` 전체를 돌려주도록 바뀐 것과 짝) — 단 `toProcessResponse`(dto→response 날짜 변환) 경유는 아직 안 함, `resizeProcess`는 정상적으로 경유. 후속 정리 여지.
- node은 아직 UI 미연결. **process는 `ProcessDialog` 하나가 실제로 호출하는 첫 컴포넌트**(exec/kill/resize) — provide/inject 스토어(`processDialog.store.ts`)+xterm 연동 상세는 `REF-process-resize.md` "ProcessDialog" 절(백엔드 resize 배선과 한 문서에 묶임 — 이력도 `history/process-resize.md` 하나).

## vue-query — queryFn은 queryKey에서 값을 읽는다 (2026-10-13, 불변)
- `queryFn: ({ queryKey }) => api(queryKey[2])` 형태로. **ref(`id.value`)를 읽지 말 것** — 같은 키의 쿼리는 마지막으로 조회한 observer의 `queryFn`을 기억하므로, 그 observer(타일)가 다른 폴더로 옮겨 간 뒤 무효화 재조회가 엉뚱한 값을 그 키에 넣는다(루트 타일에 다른 폴더 목록이 뜸). 경위 → `REF-node-select.md` "함께 고친 것"

## 앱 레이아웃 루트 — `App.vue` → `ProvideAppLayout.vue`
- `App.vue`의 최상위가 `<v-app>` 직접 사용 → **`<provide-app-layout>`**(`feature/layout/component/provideAppLayout.vue`)로 감싸는 형태로 변경. 이 컴포넌트가 내부에서 `<v-app><slot/></v-app>`을 렌더하면서 동시에 `provideResizeGroupStore()` 호출 + `flag.value = true`로 **앱 전체 공유 리사이즈 관측 그룹을 1회 킥**(→ `REF-util.md` "공유 리사이즈 관측 그룹").
- `AppHead.vue`/`appHead.store.ts`도 이 공유 그룹으로 이전: 기존 `useResize(elRef)`(자체 옵저버) → `useResizeCallback(elRef, cb)` + `computeResizeSize()`로 `size`를 직접 갱신. `vElRef` 타입도 `any` → `InstanceType<typeof VAppBar>`로 정정.

## 전체 화면 버튼 (2026-10-13)
- `AppHead.vue` 로그아웃 왼쪽. 목적 = 모바일 주소창 숨기기(문서 스크롤이 없어 브라우저가 스스로 안 숨김). `document.fullscreenEnabled`가 false(iPhone 등)면 숨김, 상태는 `fullscreenchange`로 읽음(뒤로가기·Esc 해제 반영). 검토한 대안: PWA 설치(주소창 없음, 나중 후보) / 문서 스크롤 만들기(높이 변동으로 타일·터미널 재계산, 기각)

## 미착수/다음
- 회원가입 파트(보류) → `REF-frontend-auth.md` "회원가입"
- node 카탈로그 UI = termspace 영감 2D 타일 그리드(터미널/폴더 타일, 깊이2 고정)로 전면 재설계 중(2026-08-10, 옛 캔버스+트리 컨셉 폐기). **설계 → `REF-node-ui.md`**(열린 질문 다수) / **레이아웃(순서·reflow) → `REF-node-ui-layout.md` / 화면 크기별 표시 → `REF-node-ui-projection.md`**. Phase1 첫 코드(`TileLayout.vue`, → `REF-widget.md`) 착수(2026-08-19)했으나 아직 axis/Band 모델과 미연동. node CRUD REST 클라이언트는 있으나 아직 호출하는 UI 없음.
- socket 수신 핸들러 실연동: node `on('NODE:CREATE'|…)` → 트리/캔버스 갱신 (→ `REF-realtime.md`). process는 `ProcessDialog`가 `DATA` 연동 완료, `PROCESS:UPDATE`/`STATUS`는 스텁만.
- process 입력(키스트로크) 배선 — 고빈도라 REST 부적합, 소켓 메시지 쪽이 유력하나 미정.
