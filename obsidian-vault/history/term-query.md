# history/term-query — 터미널 질의 응답 중복

> 요약·결정 → `REF-term-query.md` / 현재 진행 → `CURRENT.md`

## 2026-10-13 — 응답-1(구조안)~응답-3(확인), 격리 헤드리스 12/12
- 응답-1: 구조안 제시 → 사용자 "질의 응답 중복이 무엇에 대한건지 전혀 모르겠어" → 질의·응답 개념과 탭 수만큼 중복되는 경로를 도식으로 설명 → 이해 후 **질의를 처리하지 않는 방법**(parser 훅) 승인. 처음 확정한 (가)(`onData`에서 응답 모양 거름)보다 붙여 넣기 오탐이 없음
- 응답-2: `xtermQuery.util.ts` `muteQueries` + `createHandle`에서 호출. type-check 통과
- 응답-3(격리: 임시 DB `nexus_query_tmp` + supervisor 5051 + worker `NAME=query-dev` + 프론트 3001, scratchpad `pw/query.mjs`)
  - 시험 스크립트(bash): `stty -icanon -echo` → 한 글자 읽어 `q`=`ESC[6n` · `d`=`ESC[c` · `e`=`ESC[>c` · `o`=`OSC 11;?` · `m`=`ESC[?2004$p` · `s`=`DCS $qm` 출력 → 1초 뒤 0.5초 타임아웃으로 남은 입력을 모두 읽고 `ESC` 뒤 `[`·`]`·`P` 개수 = 응답 수를 `R_<글자>=n`으로 출력
  - 같은 계정 두 페이지 A·B, A의 탭 id로 `POST /processes/exec` → A 새로고침(API 실행이라 A는 응답 타일을 안 넣음) → A = 크기 소유, B = 아님(`.ProcessTerminal--fixed`)
  - 결과(수정 적용): 6종 질의 모두 응답 1 · B에서 입력해도 1 · A 닫은 직후(유예 중) 0 · 유예 뒤 B로 소유권 넘어오면 1 → **12/12**
  - 대조(수정 뺌, vite 변환 플러그인이 `BASELINE=1`이면 `muteQueries` 줄 제거): 6종 모두 응답 2 → 확인 스크립트가 중복을 잡아냄을 확인
  - 함정: 터미널 탭(`터미널 1`)은 경로 목록이지 터미널 자체가 아님 — 터미널 타일은 보드에 따로 있음
  - 확인 뒤 격리 프로세스 종료·DB 삭제

## 2026-10-13(2) — 사용자 Chrome(3000·5050, irony 계정)에서 실제 확인
- 사용자 요청 "내 크롬에 지금 켜놨거든? 네가 테스트해볼래?" → Claude in Chrome으로 탭 2개 열고 `test/bash` 실행(첫 탭 = 크기 소유)
- 셸에서 `stty -echo -icanon; printf '\033[6n\033[c\033[>c'; timeout 1 cat -v` → 응답 `^[[2;1R` · `^[[?1;2c` · `^[[>0;276;0c` 각 1개(탭 2개 열린 상태)
- `vim /tmp/nexus-query-check-1013.txt` → 두 탭 모두 화면에 이상 입력 없음 → `:q!`·`exit`(process exited: 0). 보드에 끝난 bash 터미널 타일이 남음, 탭 2개 닫음

## 2026-10-13(3) — 사용자 "마감하고 커밋해줘" → 파트(질의 응답 중복) 마감·커밋
