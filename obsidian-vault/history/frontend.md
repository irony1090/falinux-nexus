# history/frontend — apps/frontend

> 요약·재사용 지식 → `REF-frontend.md` / 현재 진행 → `CURRENT.md`

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
