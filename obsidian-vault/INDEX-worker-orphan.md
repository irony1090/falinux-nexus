# INDEX — 고아 스크립트 (마감)

> 2026-10-15 파트(고아 스크립트) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
worker를 꺼도 실행 중 스크립트가 남던 문제. 재현해 보니 원인은 `Setsid`가 아니라 worker를 SIGHUP 무시 상태(`nohup`)로 띄운 것이 스크립트에 상속된 것(기본 상태면 PTY master가 닫혀 SIGHUP으로 같이 죽음). 남은 스크립트는 다시 붙을 수 없는 누수라 정리 쪽으로 결정. 구현 = 안 나(SIGHUP 복원) + 안 다(정상 종료 정리): worker가 SIGHUP·SIGTERM·SIGINT를 받아 스크립트에 TERM -> 5초 뒤 KILL 후 종료, `nohup`으로 떴으면 SIGHUP만 무시. 격리 확인(수정 전 worker와 비교) + 사용자 재실행 확인

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-worker-orphan.md`(재현표 · 메커니즘 · 정리안 안 가~안 마 · 결정 · 구현 · 고아-3(확인) 결과 · systemd unit 예시) / `history/worker-orphan.md`
- 격리 함정: `REF-test-env.md` "끌 때 함정 — worker 자식"(원인 정정) · "`pgrep -f` 오판 함정"
- 확인 스크립트 명세: `REF-test-scripts.md` "2026-10-15 확인 스크립트 (고아 스크립트)"
- Windows 대응 메모(Job Object): `REF-windows-worker.md` "메모 — worker 종료 시 자식 정리"

## HISTORY 색인 행 (원문)
### `history/worker-orphan.md` — worker 종료 시 고아 스크립트 (2026-10-15 신설)
- 2026-10-15(2) 재현 진단: 원인 = `Setsid` 아님, worker의 SIGHUP 무시(`nohup`) 상속 + 스크립트 자체 회피. 정리안 안 가(현상 유지)~안 마(시작 시 청소) → 결정 = 안 나(SIGHUP 복원) + 안 다(정상 종료 정리)
- 2026-10-15(3) 구현(`handleSignals`·`shutdown.go`) + 고아-3(확인) 격리 6/6, 수정 전 worker와 비교
- 2026-10-15(4) SIGHUP 처리 개정(`nohup`이면 무시 / 아니면 정리 후 종료) + 재확인 → 사용자 재실행 확인 · 마감·커밋

## 이 파트에만 해당하는 결정
- 종료-1(유예 시간) 5초 상수(`shutdownGrace`, systemd `TimeoutStopSec`보다 짧게) / 종료-2(최종 상태 보고) 기존 pump 경로로 FAILED 143·137 / 종료-3(두 번째 신호) 즉시 강제 종료
- 안 라(systemd)는 unit 예시만(REF) — 배포 방식 미정, systemd 없는 장비도 있음 / 안 마(시작 시 청소)는 crash 뒤 고아가 실제로 보이면 그때
