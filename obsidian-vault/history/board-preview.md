# history — 보드 미리보기

> → `REF-board-preview.md` / 보드 개념 → `REF-workspace.md` / 현재 진행 → `CURRENT.md`

## 2026-10-15
- 사용자 확인 목록 1~7 "확인했어" → 파트 마감: 코드 커밋 `52ae8b9`, `INDEX-board-preview.md`로 접음
- vault 파악 → 미결정(전체보기를 닫을 때 다른 보드 xterm) 중 사용자 "중간안으로 가자" → 계획 중간-1(등록 출처 기억)·중간-2(놓아주기)·중간-3(확인) 제시, 등록은 `terminalRestore`에 둠(`boardTrees`로 안 옮김) → 사용자 "진행해줘"
- 구현 `terminalRestore.hook.ts`만(`viaOthers`·`othersOnly`·`release`, `next()`에서 호출). 경계 = `ready`일 때만 놓아줌(누르고 이동 시 빈 트리로 판정 방지) · 조회 중 닫기는 조회 끝에서 정리. type-check 통과
- 중간-3(확인) 격리 헤드리스 `pw/mid.mjs` 15/15(판정 = 스냅샷 요청 횟수). 1차 실패 2건은 스크립트 문제(`tick-B 0`이 이미 위로 지나감 → 줄 번호 연속성으로 판정 / `unroute`가 붙잡은 요청을 먼저 처리 → 놓은 뒤 기다렸다 해제)

## 2026-10-14
- ③ 파트(보드 미리보기) 착수. 정할 것 PV1(표시 위치)·PV2(카드 내용)·PV3(출력 구독 범위) 제시
- 아이디어용 시안 A~E 캔버스(https://claude.ai/artifact/KSBX7CyuzQP3soj9tWnQd9) → 사용자가 E(전체보기 보드 줄) 선택 → PV1~PV3 대신 EV1(다른 보드 출력)~EV3(다른 보드 크기) 제시, 사용자 "네 추천대로" → EV1 나 · EV2 PC 가/폰 나 · EV3 가 + 기본값 EV4~EV6
- 구조안 준비 중 발견: 서버가 이미 모든 살아 있는 process를 탭마다 구독(`subscribeLiveProcesses`) → EV1의 "열린 동안만 구독"은 전제가 틀림, "열 때 xterm 등록"으로 고쳐 적음. 구조안 미리보기-1(다른 보드 트리)~미리보기-6(확인) 작성(코드 없음)
- 사용자 "내 확인이 필요할때까지 진행해줘" → 미리보기-1~5 구현(새 파일 `boardTrees.store.ts`·`OverviewTile.vue`, 수정 `tileGrids.store.ts`·`TileStrip.vue`·`terminalRestore.hook.ts`·`pages/index.vue`), type-check 통과 → 미리보기-6(확인) 격리 헤드리스 31/31 + 다시 열기 지연 확인(커밋 전, 사용자 확인 대기)
- 격리 정리 중 `pgrep -af vite` 함정 재발(자기 셸 종료, exit 144) — 포트로 남은 것 확인 후 DB 삭제
- 사용자 질문: "알아두실 동작"의 "그 탭"·"닫은 뒤"가 모호 → 브라우저 탭·전체보기를 닫은 뒤로 풀어 답하고 REF 문장 고침
- 사용자 질문: 스냅샷만 읽을 수 없나 / 지금 모두 소켓 연결인가 / 실시간 비용 → 소켓 단계·xterm 단계 구분, 비용 표, 선택지(지금·중간안·스냅샷) 정리 → **결정 미룸**(세션 끝)

## 2026-10-26 밤 — 폰 전체보기도 가로 한 줄 (브랜치 `fix/phone-layout`)
- 사용자: 메인 보드 푸터 "전체보기"를 폰에서도 PC처럼 가로로. 보드 여럿일 때 폰은 이름을 줄 위로(사용자 선택, 140px 이름 칸이 폰 폭을 먹음)
- 수정: `tileGrids.store` `ovDir`(가로·세로 긴 축 판정) 삭제 -> 들어갈 때 scrollLeft만 맞춤 / `TileStrip` `.row`·`.col` 갈래 삭제, 가로 규칙을 기본으로, `.multi.phone`(폰 = `cap.x === 1`) = 보드 세로 쌓기 + 이름 칸 가로·왼쪽 정렬. EV2 개정(`REF-board-preview.md`)
- 확인: 폰 Grid 4개 같은 높이 가로·가로 스크롤 / PC 그대로 / 보드 2개(타일 트리 PUT으로 폴더 타일 + `POST /api/workspaces`) 폰 이름이 줄 위 — 첫 확인 때 PC용 가운데 정렬이 남아 이름이 오른쪽으로 치우쳐 `justify-content: flex-start` 추가
