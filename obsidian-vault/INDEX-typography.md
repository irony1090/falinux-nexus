# INDEX — 글자 크기 규칙 (마감)

> 2026-10-23 파트(글자 크기 규칙) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
전체 화면 글자를 키우고, 글자 크기를 **Vuetify(MD3) 단계 이름**으로 바꿔 **중앙에서 한 번에** 조절. 브랜치 `feat/typography`(master `83678c8`에서).
- 글자-1(중앙 값): 단계 값 한 곳 `styles/_typography.scss`(Vuetify `settings.scss`와 우리 mixin이 같이 읽음) · 배율 `$font-scale` -> `styles/global.scss` html font-size · mixin `vars.font(이름)`
- 글자-2(대응표): 9~10.5px = label-small(11) · 11·11.5 = body-small(12) · 12·12.5 = body-medium(14) · 13 = 쓰임별 14
- 글자-3(바꾸기): 27개 파일 73곳, px 글자 크기 0곳
- 글자-4(터미널): `termFontSize()` = html font-size × 0.875(12px -> 14px)
- 글자-5(깨지는 곳 맞추기): scoped `button` 초기화가 `v-btn`에 걸리던 것 `button:where(:not(.v-btn))` · x-small 글자 버튼 = label-medium(12)(`global.scss`)
- 글자-6(확인): 격리 환경 전후 스크린샷 + **사용자 브라우저 확인 OK**(2026-10-23)
- 커밋: 브랜치 `feat/typography`에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-typography.md`(조사 · 착수 결정 · 큰 계획 글자-1~6 · 결과 · 함정) / `history/typography.md`
- 확인 스크립트 명세 → `REF-test-scripts-1023.md` "글자-5 전후 스크린샷"

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: `REF-typography.md`(파트(글자 크기 규칙), 진행 중)
- HISTORY: | `history/typography.md` | 글자 크기 규칙(진행 중, 2026-10-23 신설) | 10-23 글자-1~5 완료, 글자-6(확인) 대기 |

## 이 파트에만 해당하는 결정 · 남은 것
- Q45(기본 크기) Vuetify 기본값 · Q46(터미널 글자) 같은 배율 · Q47(폰·PC) 같게 · Q48(터치 스크롤 양) BACKLOG
- 파트 공통으로 남긴 결정(MEMORY 기술 스택): 글자 크기 = 단계 이름 mixin, px 금지
- 함정(dev 서버에서만 나는 sass `with` 오류 · mixin 줄 높이 순서 · scoped `button` 초기화) → `REF-typography.md` "함정"
