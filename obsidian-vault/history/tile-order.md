# history — 타일 순서 구조 (파트(타일 끌기) 구조-1~3)

> 2026-10-29 `history/tile-drag.md`에서 분할 → `REF-tile-order.md` / `REF-tile-drag.md`

## 2026-10-29 구조-1~3 구현 (브랜치 `feat/tile-drag`)
- 사용자 "진행" -> 변환-B 확정 -> 서버(tileDoc·넣기 6곳·변환기)·프론트(store·배치) 작성, 상세 = REF "구조-1~3 구현"
- 격리 확인 전부 통과(REF 같은 절). 화면 확인 중 번호 `(1)`·`(2)`가 화면 순서를 따라 바뀌는 것 발견 -> 질문
- 남긴 것: 격리 환경(supervisor 5051·worker `td-dev`(TMPDIR `/run/user/1000/tdiso`)·vite 3001, DB `nexus_td_tmp`, 계정 td1·td2) = 끌기 단계 확인에 계속 씀, 스크립트 = 이 세션 scratchpad `iso/`
