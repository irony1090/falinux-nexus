# INDEX — Ctrl+Alt 팝업 버튼 · 터치 스크롤 양 (마감)

> 2026-10-23 파트(Ctrl+Alt 팝업 버튼·터치 스크롤 양) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
PC Ctrl+Alt 팝업(`PagerPopup`)에 크기·전체보기 버튼 + 터치 길게 누르기로도 팝업 + 터미널 터치 스크롤 2배. 브랜치 `feat/pager-tools`(master `0b104e9`에서).
- 팝업-1(크기 카드 공용화): `TileSizeCard.vue` + `hook/tileSize.hook.ts`, 따라다니기(`relocateSizeMenu`)는 머리 줄에만
- 팝업-2(왼쪽 슬롯): `GridPager` `#start`
- 팝업-3(팝업 버튼): 크기(맨 왼쪽) -> 버튼 가운데에 카드 + 알약 숨김, 자리 고정 / 전체보기(오른쪽 끝, 아이콘만)
- 팝업-6(길게 누르기로 띄우기): 머리 줄·타일 밖 빈 곳 500ms, 폰 포함, 바깥 탭·동작으로 닫힘
- 팝업-4(터치 스크롤 배율): `TERM_GAIN = 2`(터미널로 넘기는 양에만)
- 팝업-5(확인): 헤드리스 PC 21/21 · 터치 17/17 + **사용자 빌드로 PC·폰 확인 OK**(2026-10-23)
- 커밋: 브랜치 `feat/pager-tools`에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-pager-tools.md`(결정 Q48·Q49·Q50·Q53~Q57 · 큰 계획 팝업-1~6 · 구현 · 확인 · 함정) / `history/pager-tools.md`
- 확인 스크립트 명세 → `REF-test-scripts-1023.md` "팝업-5"

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: `REF-pager-tools.md`(파트(Ctrl+Alt 팝업 버튼·터치 스크롤 양), 진행 중)
- HISTORY: | `history/pager-tools.md` | Ctrl+Alt 팝업 버튼·터치 스크롤 양(진행 중, 2026-10-23 신설) | 10-23 착수 Q48·Q53·Q54, 큰 계획 팝업-1~5 |

## 이 파트에만 해당하는 결정 · 남은 것
- Q53(크기 팝업 유지) PC = 키 떼면 닫힘 · Q55(카드 뜨면 팝업) 알약 숨김 · Q56(카드 위치) 버튼 가운데 · Q57-1~4(터치: 자리·닫는 법·폰 포함·500ms) · Q48(터치 스크롤 양) 2배
- 함정: Chrome 터치 대상 보정 — 좁은 빈 곳 길게 누르기는 옆 요소로 대상이 옮겨질 수 있음(머리 줄이 확실) → `REF-pager-tools.md`
