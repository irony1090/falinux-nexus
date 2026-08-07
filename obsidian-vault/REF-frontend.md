# REF: frontend (apps/frontend)

> supervisor용 프론트엔드. 상세 이력 → `history/frontend.md`(+resize/다이얼로그 `history/process-resize.md`) / 현재 진행 → `CURRENT.md`
> node 카탈로그 관련 백엔드 설계 → `REF-node-label.md`. node 카탈로그 UI 컨셉(배치도/트리/모바일) → `REF-node-ui.md`. process resize 백엔드 배선 → `REF-process-resize.md`.

## 위치/정체성
- 경로: `apps/frontend` (package name `frontend-nexus-supervisor`)
- 커밋: `e511359`(초기 스캐폴딩) → `3a8e92e`(socket hook 재설계 + user/login·다이얼로그 WIP) → `e28252b`(hook 재연결 견고성) (모두 2026-06-30)

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

## user/login + 공용 API + 전역 다이얼로그 (WIP, 커밋 3a8e92e)
> 아직 마무리 전(실서버 연동·라우터 가드 남음). 구조만 자리잡음.
- **인증**: `feature/user/store/auth.store.ts`(auth 상태) + `feature/user/api/user.api.ts`. `pages/Login.vue` + 라우트 `/login`(router/index.ts, 수동 정의).
- **공용 API 계층**: `common/api/api.util.ts`(fetch 래퍼) + `query.util.ts`. socket hook과 별개의 REST 호출 통로.
- **전역 다이얼로그**: `feature/layout/component/AppDialog.vue` + `store/appDialog.store.ts`(reactive 모듈 패턴 — 컴포넌트 밖에서 다이얼로그 open). layout store 계열: `appHead`/`appNav`/`appWindown`/`appDialog`.

## REST API 클라이언트 (`feature/{node,process}/api`, 2026-07-21)
- `feature/node/api/node.api.ts`: `createNode`/`listChildren`/`getNode`/`patchNode`/`deleteNode` (node.go+nodeDto.go 미러) + vue-query `useListChildren`/`useGetNode`/`useNodeQueryClient`(invalidateAll/-Force). `Q_KEY`는 `LIST(parentId?)`/`DETAIL(id)`.
  - `PatchNodeRequest`는 백엔드 `patch.Field[T]`({valid,value})를 호출부에 노출하지 않는다 — 평범한 `T|null|undefined` 값만 받고 `patchNode()` 내부의 `toPatchField()`가 undefined→`{valid:false}` / null·값→`{valid:true,value}`로 변환(관리 부담 축소가 목적).
  - Response/Dto 패턴은 `user.api.ts`와 동일: camelCase `XResponse`(Date) + `Replace`로 파생한 `XResponseDto`(unix sec) + 변환 함수.
- `feature/process/api/process.api.ts`: `listSubscriptions`/`subscribeProcess`/`unsubscribeProcess`/`execProcess`/`killProcess`/`resizeProcess`(2026-07-22 추가, `POST /processes/resize/:processId`) (processApi.go 미러).
  - **선행 배선**: `listSubscriptions`가 원래 DTO 없이 `superdb.Process`(json 태그 없는 sqlc raw, PascalCase+RFC3339)를 그대로 내려보내던 걸 발견 → **백엔드에 `processDto.go` 신설**(`processResponse`+`newProcessResponse(s)`, nodeDto.go와 대칭)해 camelCase+unix sec로 정정한 뒤에 프론트 타입을 작성함(2026-07-21). 순서: 백엔드 응답 정합 먼저 → 프론트 타입은 그 위에.
  - `execProcess` 반환 타입은 사용자가 직접 `{uid}`→`ProcessResponse`로 바꿈(백엔드가 `newProcessResponse` 전체를 돌려주도록 바뀐 것과 짝) — 단 `toProcessResponse`(dto→response 날짜 변환) 경유는 아직 안 함, `resizeProcess`는 정상적으로 경유. 후속 정리 여지.
- node은 아직 UI 미연결. **process는 `ProcessDialog` 하나가 실제로 호출하는 첫 컴포넌트**(exec/kill/resize) — provide/inject 스토어(`processDialog.store.ts`)+xterm 연동 상세는 `REF-process-resize.md` "ProcessDialog" 절(백엔드 resize 배선과 한 문서에 묶임 — 이력도 `history/process-resize.md` 하나).

## 앱 레이아웃 루트 — `App.vue` → `ProvideAppLayout.vue`
- `App.vue`의 최상위가 `<v-app>` 직접 사용 → **`<provide-app-layout>`**(`feature/layout/component/provideAppLayout.vue`)로 감싸는 형태로 변경. 이 컴포넌트가 내부에서 `<v-app><slot/></v-app>`을 렌더하면서 동시에 `provideResizeGroupStore()` 호출 + `flag.value = true`로 **앱 전체 공유 리사이즈 관측 그룹을 1회 킥**(→ `REF-util.md` "공유 리사이즈 관측 그룹").
- `AppHead.vue`/`appHead.store.ts`도 이 공유 그룹으로 이전: 기존 `useResize(elRef)`(자체 옵저버) → `useResizeCallback(elRef, cb)` + `computeResizeSize()`로 `size`를 직접 갱신. `vElRef` 타입도 `any` → `InstanceType<typeof VAppBar>`로 정정.

## 미착수/다음
- user/login 마무리: 실서버 인증 왕복·라우터 가드(비로그인 → `/login`)·세션 복원.
- node 카탈로그 UI = 트리 + 캔버스(% 절대배치). **컨셉 설계 → `REF-node-ui.md`로 분리**(2026-08-07, 10k자 기준 분할). 구현은 미착수. node CRUD REST 클라이언트는 있으나 아직 호출하는 UI 없음.
- socket 수신 핸들러 실연동: node `on('NODE:CREATE'|…)` → 트리/캔버스 갱신 (→ `REF-realtime.md`). process는 `ProcessDialog`가 `DATA` 연동 완료, `PROCESS:UPDATE`/`STATUS`는 스텁만.
- process 입력(키스트로크) 배선 — 고빈도라 REST 부적합, 소켓 메시지 쪽이 유력하나 미정.
