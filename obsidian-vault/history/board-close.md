# history — 보드 닫기 · 종료 유예 (2026-11-05 신설)

> → `REF-board-close.md` · 종료 정책 `history/process-kill.md` 쪽 결정의 개정 · `CURRENT.md`

## 2026-11-05 — 루트 타일 닫기 + 서버 10초 규칙
- 사용자: 보드 루트 타일 닫기 + 닫기 전 "모두 종료 / 루트 보드로 보내기" 묻기. 조율 Q1(대상 보드)~Q4(진입점) 답 → REF "요청"
- 사용자 질문으로 짚은 것: 종료 요청만으로 10초 뒤 자동 강제 종료가 되는 게 아님(10초는 프론트 대기, 강제는 타일 닫힘 판정) / 타일이 아무 데도 없다고 서버가 종료하지 않음(종료 요청이 있어야 함) / 루트 터미널 탭은 경로 기준 조회라 타일을 닫아도 끝날 때까지 보임 / 종료 푸시는 이미 있음 — D5는 "언제 SIGKILL"을 서버가 정하는 문제
- 프론트 방식 대신 서버 방식 비교 → S-a(보드 닫기만)·S-b(모든 종료 요청) 중 사용자 **S-b**. 타일 닫기 판정에도 유예 조건을 넣어야 함(안 그러면 바로 닫는 순간 SIGKILL)
- 구현 브랜치 `feat/board-close`: SQL·`killGrace.go`·`killWorkspace` / `boardClose.hook.ts`·TileFrame 루트 ×·BoardMenu·종료하고 닫기 즉시
- 사용자 질문 "종료 요청이 없어도 계속 검사?" -> 예(2초마다 UPDATE 1번). 안 A(메모리 표시)·안 B(부분 인덱스) 중 사용자 **안 A** -> `killPending` + `HasPendingKills`. 확인 13/13(쉬는 동안 0번), 스캔 수 측정이 늦게 반영되는 함정 -> 쿼리 로그로 셈
- 확인: API 11/11 · 화면 9/9(첫 실행 1건 FAIL = 확인 스크립트가 URL 끝을 비교했는데 `?focus=…&view=…`가 붙음, pathname 비교로 고쳐 통과)
- 11-05 화면 확인용 개발 서버(8011·8012) + 테스트 계정 `devtest` + worker `devtest-pc` 띄움 -> 사용자 화면 확인 "잘 동작해"
- 11-05 15:10 test3(8013) 적용: 사용자가 외부(Nexus 세션)라 Claude가 직접 — 태그 `SUPERVISOR-BOARD-CLOSE-261105` · `scripts/build.sh 261105-board-close supervisor frontend` · `swap.sh`(옛 556117 끔 -> 새 것 켬 -> `/` 200 확인, 실패 시 `261104-files`로 되돌림) -> 1초 만에 켜짐, 브라우저·worker 다시 붙음(PROCESS 2 · PENDING 0), 마이그레이션 변화 없음(00018). worker는 그대로
