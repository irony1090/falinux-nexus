# REF: frontend (apps/frontend)

> supervisor용 프론트엔드. 상세 이력 → `history/frontend.md` / 현재 진행 → `CURRENT.md`
> node 카탈로그 관련 백엔드 설계 → `REF-node-label.md`

## 위치/정체성
- 경로: `apps/frontend` (package name `frontend-nexus-supervisor`)
- 커밋: `e511359` (2026-06-30, 초기 스캐폴딩)

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
- unplugin-fonts(Roboto fontsource)는 주석 처리/제거 방향. `@fontsource/roboto`는 죽은 의존성(삭제 가능). `import 'unfonts.css'`도 Fonts 플러그인 제거 시 함께 제거.
- weight 200/600 `@font-face` 없음 → 사용 시 브라우저 합성(faux). 필요 시 추가.

## scss 구조
- `_variables.scss`: 프로젝트 공용 토큰(spacing/weight/alpha 함수). vite `additionalData`로 전 scss에 주입.
- `settings.scss`: Vuetify 내부 변수 오버라이드 전용(`configFile`로 연결). 폰트 등은 여기서.
- 둘 역할 분리 — 혼용 금지.

## 미착수/다음
- node 카탈로그 UI = 트리 + 캔버스(% 절대배치). 터미널(xterm.js)은 EXEC/EDIT용 별개.
- supervisor → socket으로 node 변경 push 연동(채널/이벤트 형식 미설계).
