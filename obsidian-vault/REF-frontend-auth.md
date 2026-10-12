# REF: frontend — 로그인 · 공용 API · 전역 다이얼로그

> `REF-frontend.md`에서 분리(2026-10-12). 이력 → `history/frontend-auth.md` / 프론트 전반 → `REF-frontend.md` / 현재 진행 → `CURRENT.md`

## user/login + 공용 API + 전역 다이얼로그 (마무리 2026-10-12 — 파트(user/login 마무리))
> 작업 단위: 로그인-1(상태를 모듈로) · 로그인-2(라우터 가드) · 로그인-3(401만 로그아웃) · 로그인-4(화면 정리) · 로그인-5(확인)

| 위치 | 내용 |
|---|---|
| `feature/user/store/auth.store.ts` | auth 상태 = **모듈 단일 값**(`undefined` 확인 전 / `null` 비로그인 / 값). `ensureSession()` = 처음 한 번만 `GET /users/session`(진행 중 Promise 공유). `provideAuthStore`(App.vue) / `useAuthStore`는 같은 ctx 반환 — 로그인-1(상태를 모듈로) |
| `router/authGuard.ts` | `beforeEach`: `ensureSession` 기다린 뒤 비로그인 → `/login?redirect=원래주소`(`/`면 query 없음), 로그인 상태로 `meta.public`(`/login`) → redirect(같은 사이트 경로만, `//host` 거부). **보는 중 상태 변화**(로그아웃·만료·로그인 성공)는 같은 파일의 `watch(authState)`가 이동 — 첫 이동 전(`matched` 없음)은 가드가 맡음. 페이지에서 직접 이동하지 않는다 — 로그인-2(라우터 가드) |
| `common/api/api.util.ts` `onUnauthorized` | 응답 401이면 등록된 콜백 호출(`/users/session` 제외 — auth가 직접 처리). common → feature 역의존을 피한 등록 방식 — 로그인-3(401만 로그아웃) |
| 사용자 설정 삭제 | `expire()`(401·로그아웃)에서만 `clearUserPrefs`. 서버 연결 실패는 `offline` 표시만(처음 확인이면 auth = null로 로그인 화면, 설정 유지) |
| `pages/Login.vue` | 폼 기본값 = `VITE_DEV_LOGIN_ID/PW`(`.env.development.local`, git 제외, `env.d.ts`에 타입). `offline`이면 "서버에 연결할 수 없습니다". 회원가입 버튼 **제거**(회원가입 파트에서 다시) — 로그인-4(화면 정리) |
| `AppHead.vue` | 로그아웃 = AppDialog 확인 뒤(Enter = 로그아웃, Esc = 취소) |

- 로그아웃 때 `boards.reset()`(목록 비움)은 보드 삭제가 아님 — `index.vue` "보던 보드가 사라짐" watch는 `boards.loaded`가 true일 때만 판정(2026-10-12 사용자 확인 중 발견)
- 로그인 실패는 서버가 **400**(401 아님) → 401 알림과 겹치지 않음
- 공용 API 계층: `common/api/api.util.ts`(axios 래퍼, `throwThen`은 200만 통과 → 삭제 API도 200 `{}`) + `query.util.ts`
- **전역 다이얼로그**: `feature/layout/component/AppDialog.vue` + `store/appDialog.store.ts`(reactive 모듈 패턴 — 컴포넌트 밖에서 다이얼로그 open). layout store 계열: `appHead`/`appNav`/`appWindown`/`appDialog`.
- **`AppDialog` 주의(2026-09-30)**: `content`는 `v-html`로 그려진다 → 사용자 입력(노드 이름 등)은 `content`가 아니라 `title`(텍스트 보간)로 넘긴다. 버튼에 `bindButton`을 묶지 않은 Enter·Space·Esc는 창 닫기로 처리되므로, 삭제 같은 되돌릴 수 없는 버튼에는 Enter를 묶지 않는다(사용 예: `FolderTileBody.vue` 삭제 확인, `ScriptEditDialog.vue` 닫기 확인).

## 회원가입 (보류 — 2026-10-12 사용자 결정 "나중에")
- 가입 정책(누구나 / 관리자만 / env로 차단) + 장비 실행 권한(지금은 아무 계정이나 아무 `deviceKey`로 실행 가능) + 비밀번호 bcrypt 전환(지금 솔트 없는 sha256, `user.go` `toHash`). 서버 `POST /users`는 로그인 없이 열려 있음
