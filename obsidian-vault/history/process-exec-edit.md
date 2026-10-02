# history/process-exec-edit — 실행 타입 EXEC vs EDIT (worker vi 편집)

> 설계·재사용 지식 → `REF-process-exec-edit.md` / 현재 진행 → `CURRENT.md`
> 2026-06-26 설계·프로토콜 어휘 이력은 이 파일 신설 전이라 `REF-process-exec-edit.md` 본문에만 있음.

## 2026-10-02 — 4.6 EDIT 프론트 연결 구조안 제시 (코드 없음)
- 4.5 SNAPSHOT 커밋 직후 사용자 "다음 플랜" → 코드 확인(`editResult`, worker `teardown` 순서, `ScriptEditDialog`가 열 때 `getNode`) 후 구조안 제시
- 계획 정정: 노드 내용 캐시 갱신 불필요
- 결정 E1(진입 위치)~E4(동시 편집) 제안 + 작업 단위 편집-a(프론트 진입)·편집-b(확인). **사용자 답 전 세션 종료** — E2(끝난 뒤 타일)·E3(저장 결과 표시) 확인부터 이어 갈 것
