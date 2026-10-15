# history — worker 종료 시 고아 스크립트

> → `REF-worker-orphan.md` / 현재 진행 → `CURRENT.md`

## 2026-10-15(2) — 진단: 원인은 `Setsid`가 아니라 SIGHUP 무시 상속
- 사용자가 다음 파트로 고아 스크립트를 고름(Windows worker는 뒤로 미룸)
- scratchpad `orphan/`에서 worker의 PTY 시작 방식을 복제해 재현 → 기본 상태에서는 worker가 죽으면 PTY master가 닫히고, 커널이 SIGHUP을 보내 스크립트도 죽음 / `nohup`으로 띄우면 SIGHUP 무시가 상속돼 살아남음 / 스크립트가 `trap '' HUP`·`setsid`·`nohup`을 쓰면 남음
- 살아남은 스크립트는 새 worker가 다시 붙을 수 없음(master는 worker 메모리에만 있음) → "작업 유지"는 성립 안 함, 누수
- 정리안 안 가(현상 유지)~안 라(SIGKILL에도 정리) 정리, 사용자 결정 대기
- `REF-test-env.md` "끌 때 함정 — worker 자식" 원인 정정 + `pgrep -f` 오판 함정 추가(재현 중 첫 결과를 오판함)
- 결정: 배포 방식 systemd 고려 → "systemd 없는 장비도 있음"(사용자) → 이번 파트 = 안 나(SIGHUP 복원) + 안 다(정상 종료 정리), 안 라(systemd)는 unit 예시만, 안 마(시작 시 청소) 신설 후 보류. Windows Job Object 메모를 `REF-windows-worker.md`에

## 2026-10-15(3) — 안 나(SIGHUP 복원) + 안 다(정상 종료 정리) 구현 + 고아-3(확인)
- 세부 확정(사용자 "네 제안대로"): 종료-1(유예 시간) 5초 · 종료-2(최종 상태 보고) 기존 pump 경로로 보고 · 종료-3(두 번째 신호) 즉시 강제 종료
- 코드: `cmd/worker/main.go`(`handleSignals`·`shutdownGrace`) · 신규 `cmd/worker/router/shutdown.go` · `workerRouter.go`(`WorkerState.mu`·`closing`) · `process.go`(`admit`). `go vet` 통과
- 확인: 격리 supervisor 5051 + 임시 DB `nexus_orphan_tmp` + worker `orphan-dev`, 수정 전 worker(HEAD `git archive`)와 비교. 6개 경우 전부 예상대로 → `REF-worker-orphan.md` "고아-3(확인) 결과". 격리 종료 · 임시 DB 삭제함
- systemd unit 예시를 REF에(저장소엔 안 둠)

## 2026-10-15(4) — SIGHUP 처리 개정
- 사용자가 `go run` 터미널을 닫았는데 worker가 남음(SIGHUP 늘 무시 탓) → 남은 worker(pid 2870433) SIGTERM으로 끔
- 개정: 시작 때 SIGHUP이 무시 상태(`nohup`)면 무시, 아니면 SIGTERM과 같이 정리 후 종료(`signal.Ignored`를 `Notify` 전에). 격리 재확인 H1·H2·A 통과, 임시 DB 삭제
- 사용자 재실행 확인 → 파트(고아 스크립트) 마감, INDEX 접기(`INDEX-worker-orphan.md`) + MEMORY 불변 결정 "worker 자체 종료 ≠ 끊김" 추가 · 커밋
