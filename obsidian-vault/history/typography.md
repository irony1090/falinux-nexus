# history — 글자 크기 규칙

> → `REF-typography.md` / → `CURRENT.md`

## 2026-10-23 착수·조율 (코드 없음)
- 사용자 "터미널 글자 크기 관련해서는 전체 스타일 중 폰트들을 더 키워야 할 것 같아 … 맞춰가는 작업" -> "폰트별로 규칙(body-1·title-1·caption 등) + 중앙에서 한 번에" -> "vuetify에 맞추는게 좋을 것 같아"(결정)
- 사용자 "폰트 작업으로 돌아와서 진행하려는데" -> BACKLOG 항목을 파트(글자 크기 규칙)로 올림, 조사 → REF "조사"
- Q45(기본 크기) Vuetify 기본 · Q46(터미널 글자) 같은 배율 · Q47(폰·PC) 같게 · Q48(터치 스크롤 양) 나중에 -> 큰 계획 글자-1(중앙 값)~6(확인) 제안

## 2026-10-23 글자-1(중앙 값) 코드
- 브랜치 `feat/typography` 생성. `_variables.scss`: 안 쓰던 `$base-font`·`fontLevel()` 지움, `$font-scale`(1)·`@mixin font($name)`(Vuetify `$typography`의 size·line-height) 추가. `styles/global.scss`(html font-size = 100% × 배율)를 `main.ts`에서 import. `vite build` OK
- Vuetify 4.1.2는 단계 크기를 CSS 변수로 내지 않음(`--v-font-body`·`--v-font-heading`만) -> scss mixin으로 읽음

## 2026-10-23 글자-2~5
- 사용자 "진행" = 글자-2(대응표) 판단 지점 ①(13px)·②(배지 위계)·③(줄 높이) 제안대로 확정 → REF "글자-2(대응표) 확정"
- 글자-3(바꾸기) 73곳 일괄(python 치환) · 글자-4(터미널) `termFontSize()`
- 격리 dev 서버에서 sass 오류(`already loaded ... with`) -> 단계 값을 `_typography.scss`로 빼 Vuetify·mixin이 같이 읽게(한 곳 = 중앙 값이 진짜 한 곳이 됨)
- 글자-5(깨지는 곳 맞추기): master worktree(3002) vs 이 브랜치(3001) 스크린샷 비교 -> x-small `v-btn`이 scoped `button` 초기화에 여백·색을 잃던 것 발견, `button:where(:not(.v-btn))`(`:not`만은 우선순위 문제로 경로 줄 색·배지가 지워져 다시 고침)
- type-check·`vite build` 통과
- 사용자 "추천대로" = x-small 버튼 글자 (가)(label-medium 12로 덮기) -> `global.scss` 한 곳, 아이콘 버튼(⋮)은 제외 확인(10px 그대로)
