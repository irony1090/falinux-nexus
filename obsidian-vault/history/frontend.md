# history/frontend — apps/frontend

> 요약·재사용 지식 → `REF-frontend.md` / 현재 진행 → `CURRENT.md`

## 2026-07-21 — websocket hook provide/inject 전환 + node/process REST 클라이언트

- **`websocket.hook.ts`**: `createWebsocketHook`의 반환을 `useX(): SocketContext` 단일 함수 → `[provideSocket, useSocket]` 튜플로 변경. 팩토리 호출마다 고유한 `Symbol()` 주입 키로 `provide`/`inject`(`appDialog.store.ts`와 동일한 `inject<T>(key)!` + `if(!context) throw` 패턴). 계기: 기존 모듈 스코프 `let ctx` lazy-singleton이 첫 호출 시점에 암묵적으로 초기화돼 생명주기가 불투명했던 점. 여러 소켓 인스턴스를 만들 계획이 확정되어 있어 `appDialog`처럼 고정 문자열 키가 아니라 `Symbol` 채택. `useTestSocket`도 `[provideTestSocket, useTestSocket]`로 갱신.
- **`feature/node/api/node.api.ts`** 신설: `createNode`/`listChildren`/`getNode`/`patchNode`/`deleteNode` (node.go+nodeDto.go 미러, `user.api.ts`의 Response/Dto+Replace 변환 패턴 재사용) + vue-query(`useListChildren`/`useGetNode`/`useNodeQueryClient`, bbs 참조 예시 패턴 이식). `PatchNodeRequest`는 백엔드 `patch.Field[T]` 3-state를 호출부에 노출하지 않고 평범한 값만 받아 내부에서 변환(`toPatchField`).
- **`feature/process/api/process.api.ts`** 신설: `listSubscriptions`/`subscribeProcess`/`unsubscribeProcess`/`execProcess`/`killProcess` (processApi.go 미러). 작성 중 `listSubscriptions`가 DTO 없이 raw `superdb.Process`(PascalCase+RFC3339)를 그대로 내보내던 걸 발견 → **백엔드 먼저 고침**: `processDto.go` 신설(`processResponse`+`newProcessResponse(s)`, nodeDto.go 대칭) 후 `listSubscriptions`가 이걸 쓰도록 교체(build/vet 통과). 이후 프론트 타입을 정합된 응답 위에 작성.
- 둘 다 REST 함수만 존재, 호출하는 UI 컴포넌트는 아직 없음.

## 2026-06-30 — apps/frontend 초기 스캐폴딩 (커밋 e511359)
- `npm create vuetify` (scratch preset) 로 `apps/frontend` 생성 (Vue3 + TS + Vuetify4).
- 선택: Router = Vue Router(standard) / Pinia 제외 / CSS framework None / preset 저장 안 함.
- **폰트**: Noto Sans Korean 적용.
  - `public/fonts/`에 NotoSans 9종(eot/otf/woff/woff2) 배치.
  - `src/styles/noto-sans-korean.scss`에 `@font-face`(weight 100~900) 작성, `main.ts`에서 import.
  - 기본 적용은 Vuetify `$body-font-family`(settings.scss) 오버라이드로 처리(진행).
  - 기존 Roboto(unplugin-fonts/fontsource)는 주석 처리. `@fontsource/roboto`는 삭제 대상.
- **`.gitignore`** 작성: node_modules/dist/.env/.vite/coverage 등. `.env` 무시, `package-lock.json`은 추적.
- 커밋: 81개 파일 `feat(frontend): apps/frontend 워크스페이스 초기 구성`.
- 상태관리: Pinia 대신 provide/inject 사용 중. 컴포넌트 밖 접근 필요 시 reactive 모듈 패턴 검토(→ REF-frontend).

## 2026-06-30 — 폰트 마무리 정리
- 기본폰트 적용 확인: `settings.scss` `$body-font-family: ('Noto Sans Korean', sans-serif)`.
- Roboto 제거: `@fontsource/roboto` 의존성 삭제, `main.ts` `import 'unfonts.css'` 주석화, `vite.config.mts` unplugin-fonts `Fonts()` 주석화.

## 2026-06-30 — user/login + 공용 API + 전역 다이얼로그 WIP (커밋 3a8e92e에 동반)
- `feature/user`: `store/auth.store.ts`(auth 상태) + `api/user.api.ts`. `pages/Login.vue` + 라우트 `/login`(수동 정의).
- `common/api`: `api.util.ts`(fetch 래퍼) + `query.util.ts` — socket hook과 별개 REST 통로.
- `feature/layout` 전역 다이얼로그: `AppDialog.vue` + `appDialog.store.ts`(reactive 모듈 패턴, 컴포넌트 밖 open). layout store: appHead/appNav/appWindown/appDialog.
- 상태: 구조만 자리잡음, 실서버 연동·라우터 가드 미완(→ REF-frontend 미착수/다음).
