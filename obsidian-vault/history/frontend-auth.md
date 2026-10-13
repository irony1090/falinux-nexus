# history/frontend-auth — 로그인 · 공용 API · 전역 다이얼로그

> `history/frontend.md`에서 분리(2026-10-12). 요약 → `REF-frontend-auth.md` / 현재 진행 → `CURRENT.md`

## 2026-10-13 — 로그아웃 버튼 아이콘
- 사용자 요청: 헤더 햄버거(`VAppBarNavIcon`)를 로그아웃 아이콘 버튼으로 → `AppHead.vue` `v-btn` `mdi-logout`(title "로그아웃", 비로그인 `mdi-login`), 동작(확인 창)은 그대로. type-check 통과 → 사용자 화면 확인, 파트(서버 정리)와 같이 커밋

## 2026-10-12(3) — 사용자 확인 → 파트(user/login 마무리) 마감·커밋

## 2026-10-12(2) — 사용자 확인 중 발견: 보드(`/w/{hash}`)에서 로그아웃하면 "보던 보드가 삭제되어" 알림 + 기본 보드로 이동
- 원인: 로그아웃 → `index.vue`가 `boards.reset()`(목록 비움) → "보던 보드가 목록에서 사라짐" watch가 삭제로 오인
- 수정: watch 소스에 `boards.loaded` 추가, `loaded = false`(reset)면 삭제로 보지 않음
- 확인(격리 헤드리스): B2 보드에서 로그아웃 → `/login?redirect=/w/..` + 알림 없음 → 다시 로그인 → 같은 보드 4/4 / 다른 세션이 실제로 삭제 → 알림 + `/` 유지 확인
- 시험 데이터 함정: 기본 보드 트리를 그대로 복사해 보드를 만들면 타일 id가 겹쳐 `DELETE /workspaces/:hash`가 500(기본 보드로 타일을 옮기다 충돌) → 복사할 땐 `id`·`uid`를 새로

## 2026-10-12 — 파트(user/login 마무리) 로그인-1~5 구현·격리 확인
- 사용자 "프론트 user/login 내용 알려줘" → 코드 기준 현황표(vault 기록이 2026-06-30에 멈춰 있었음) → "회원가입까지 같이?" → 권장: 따로(가입이 열려 있으면 누구나 모든 worker에서 실행 가능 + sha256 무솔트) → 사용자 "회원가입은 나중에, 구조안" → 로그인-1(상태를 모듈로)~로그인-5(확인) 제시, 정할 것 2개 권장안대로 "승인"
- 구현: `auth.store.ts` 모듈 상태 + `ensureSession` / `router/authGuard.ts`(가드 + 상태 변화 watch) / `api.util.ts` `onUnauthorized` + 응답 인터셉터 / `Login.vue` 기본값 env화·offline 안내·회원가입 버튼 제거·이동 코드 제거 / `AppHead.vue` 로그아웃 확인 / `env.d.ts` 타입. `.env.development.local`(git 제외)에 기존 기본값을 옮겨 만들어 둠
- 확인: vue-tsc 통과 + 격리(임시 DB + 5051 + 프론트 3001) 헤드리스 12/12 — 비로그인 `/`·`/w/..` → login(redirect) / 로그인 뒤 원래 주소 / 새로고침 유지 / 로그인 상태 `/login` → `/` / 바깥 redirect 거부 / 로그아웃 확인 창 → 설정 삭제 / 쿠키 삭제 뒤 다른 API 401 → login + 설정 삭제 / 서버 꺼짐 → 안내 + 설정 유지

## 2026-06-30 — user/login + 공용 API + 전역 다이얼로그 WIP (커밋 3a8e92e에 동반)
- `feature/user`: `store/auth.store.ts`(auth 상태) + `api/user.api.ts`. `pages/Login.vue` + 라우트 `/login`(수동 정의).
- `common/api`: `api.util.ts`(fetch 래퍼) + `query.util.ts` — socket hook과 별개 REST 통로.
- `feature/layout` 전역 다이얼로그: `AppDialog.vue` + `appDialog.store.ts`(reactive 모듈 패턴, 컴포넌트 밖 open). layout store: appHead/appNav/appWindown/appDialog.
- 상태: 구조만 자리잡음, 실서버 연동·라우터 가드 미완(→ REF-frontend 미착수/다음).
