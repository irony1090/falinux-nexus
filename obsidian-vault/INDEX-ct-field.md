# INDEX — 찰떡 입력칸 (마감)

> 2026-10-27 파트(찰떡 입력칸) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
브랜치 `feat/ct-field`(master `d5b4963`에서). 프론트만.
- 계기: 파트(찰떡 마무리) 항목 5(기본 `<select>` 목록 흰 바탕 흰 글씨) -> 사용자 제안 "Vuetify 입력칸을 찰떡 공용 컴포넌트로"
- 단계: F1(조각 `CtTextField`) -> F2(나머지 컴포넌트 `CtSelect`·`CtAutocomplete`·`CtTextarea`·`CtCheckbox`) -> F3(옮기기: 창 6개 -> 로그인·관리 페이지) -> F4(정리: `.mb-input` 등 지움)
- 진행 중 추가: 주소 변경 창 방식 묶음 오른쪽 정렬(2개) · 방식·호스트 칸 붙이기(묶음 메시지) · 관리 타일 포커스 선 잘림(`MemberPage .strip`)
- 확인: 격리 5051·3001 헤드리스 + 사용자 확인 OK(2026-10-27), `npm run build` 통과. 배포 안 함. 격리 환경 정리함(DB `nexus_dp_tmp` DROP)
- 커밋: 브랜치 `feat/ct-field`에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-ct-field.md`(범위 · 단계 · 추가 항목 · 함정 · F1~F4 구현) / `history/ct-field.md`
- 출발 구상 = `REF-ui-chaltteok.md` "찰떡 입력칸 구상"(→ `INDEX-ui-chaltteok`)

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: **진행 중 파트**: 찰떡 입력칸(2026-10-27, 브랜치 `feat/ct-field`) → `REF-ct-field.md` (직전 = 찰떡 마무리(마감) → `INDEX-ui-chaltteok.md`)
- HISTORY: | `history/ct-field.md` | 파트(찰떡 입력칸, 진행 중) ↔ `REF-ct-field.md` | 10-27 F1~F4 + 추가 항목 3개 구현, 빌드 통과, 마감 확인 대기 |

## 이 파트에만 해당하는 결정 · 남은 것
- 라벨 = 선 끊기(안 A, Vuetify `outlined`), 칸 높이 sm 32 / md 40, 끝 여백 = 반지름보다 안쪽(sm 20·md 24, 곧은 끝 14), 여러 줄 칸 = 16px 상자, 펼침 목록 사방 8px
- 붙인 묶음(`.ct-join`)의 메시지 = 묶음 아래 한 줄 `.ct-msg`(칸은 `error hide-details`)
- autocomplete = 관리자 장비 주인 고르기만(공유 대상 찾기는 글자 칸 — 회원 목록 노출 안 함)
- 다른 파트에서도 지킬 것(입력칸 = Ct* 컴포넌트, v-input 세로 늘어남 함정)은 MEMORY·`REF-decisions-ui.md` "입력칸"에 남김
