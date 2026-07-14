# REF: frontend (apps/frontend)

> supervisor용 프론트엔드. 상세 이력 → `history/frontend.md` / 현재 진행 → `CURRENT.md`
> node 카탈로그 관련 백엔드 설계 → `REF-node-label.md`

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
- `createWebsocketHook(url, opts)` → composable `useX(): SocketContext`. **url당 단일 ctx 공유**(소켓·correlator·핸들러).
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

## 미착수/다음
- user/login 마무리: 실서버 인증 왕복·라우터 가드(비로그인 → `/login`)·세션 복원.
- node 카탈로그 UI = 트리 + 캔버스(% 절대배치). 터미널(xterm.js)은 EXEC/EDIT용 별개.
- socket 수신 핸들러 실연동: `on('node.created'|'process.output'|…)` → 트리/터미널 갱신 (→ `REF-realtime.md`).
