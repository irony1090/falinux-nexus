# INDEX — process 배선 (마감)

> 2026-10-02 색인 2단화로 MEMORY·HISTORY에서 접은 파트(사용자 확인). **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
process 실행 배선(2026-07-01~07-22): supervisor manager·entry·relay·router + worker procs/exec/pump/teardown, frontend REST 트리거(exec/kill) + 상태동기화 버그 3건, resize 배선 + `Layout` 에러 계약, 종료/재접속 모델(worker 끊김→PENDING→재바인딩), 세션→uid 원장(이후 2026-10-01 `process_subscribers` 제거 — `REF-process-sync.md`). 입력·SNAPSHOT·동기화는 이 파트 밖.

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- 설계/재사용 지식: `REF-process-wiring.md` `REF-process-trigger.md` `REF-process-resize.md` `REF-process-reconnect.md` `REF-process-subscription.md`
- 작업 이력: `history/process-wiring.md` `history/process-trigger.md` `history/process-resize.md` `history/process-reconnect.md` `history/process-subscription.md`

## 모듈 현황표 행 (MEMORY 원문)
| 모듈 | 상태 | 문서 |
|------|------|------|
| process 실행 (execute/pty/manager/bind) | supervisor+worker 양측 배선 완료(빌드/vet 통과, e2e 확인): manager·entry·path·relay·router + worker procs/exec/pump/teardown + worker 끊김→PENDING→재접속 재바인딩(`WorkerState`) + 세션→uid 원장(`process_subscribers`, REST 구독/해지+`browsers` registry) + **frontend 트리거(exec/kill REST, 자동구독)+종료 후 Hub 구독정리**(`startRelay`/`cleanupProcessTopic`) 전부 구현 완료(2026-07-16). **kill 실사용 테스트로 발견한 상태동기화 버그 3건 수정 완료**(PENDING 오삭제/entry.Record 박제/`Status()` 에러계약, 2026-07-22). **resize 배선 완료**(`Layout` 계약 `syscall.Errno`→`error` 정정 + 결과값 기반 DB/memory 동기화 + `MsgProcessUpdate` 발행, 2026-07-22) + **프론트 `ProcessDialog`(xterm) 착수**(DATA 출력 연동 완료). **input(키입력) 완료**(2026-10-02, 소켓 `PROCESS:INPUT` + xterm 입력·I4(Ctrl+C 복사)·I5(Ctrl+V 붙여넣기), 브라우저 확인 — `REF-process-input.md`). 남은 것=ring buffer SNAPSHOT(설계만, 미착수 — `REF-process-snapshot.md`) | `REF-process.md`(+`-wiring`/`-trigger`/`-reconnect`/`-subscription`/`-snapshot`/`-resize`) |

## 이 파트에만 해당하는 불변 결정
- 없음 — 관련 불변 결정(에러 처리 panic-style, 재연결 reconciliation 등)은 다른 파트에서도 지켜야 해서 MEMORY에 남김

## HISTORY 색인 (원문)
### `history/process-wiring.md` — process 실행 배선 (supervisor 아키텍처 + worker 실행부)
- 2026-07-13 worker 실행부 본체 구현 완료(procs/exec/pump/teardown/input·resize·kill) + Cwd 배선 + e2e 스모크(htop). PROC 토픽 무구독 미해결 발견
- 2026-07-03 folder-open 버그 수정 + worker EDITOR env + PTY 실행부 선행(ExecInteractive env·Pid 접근자)
- 2026-07-01(2) process 도메인 배선 완료 + 경로 조립(sup)/치환(worker) 책임 분리 `{WORKER_BASE}/nodeID/uid`
- 2026-07-01 supervisor ProcessManager·bind·router 골격 + 종료/재접속 모델(status 깔때기·worker끊김→PENDING) 확정

### `history/process-trigger.md` — frontend REST 트리거(exec/kill) + 상태동기화 버그 수정
- 2026-07-22 entry.Record memory 동기화(SetRecord) + kill exit code 유닉스 관례화 + `pty.Interactive.Status()` 에러 계약 버그 수정 — kill 실사용 검증 완료
- 2026-07-21(2) `processDto.go` 신설: `listSubscriptions` 응답 DTO 정정
- 2026-07-21(1) 토픽 접두사 `PROC:`→`PROCESS:` 정정
- 2026-07-16 frontend 트리거(exec/kill) REST 배선 완료(`router.Exec` 시그니처 변경 + `subscribeSid` 자동구독) + 종료 후 Hub 구독 정리(`startRelay`/`cleanupProcessTopic`) — PROC 토픽 무구독 백엔드 해소

### `history/process-resize.md` — process resize(rows/cols) 배선 + `Layout` 에러 계약 정정 + `ProcessDialog`(xterm)
- 2026-07-22(4) resize REST 결과값 기반 DB/memory 동기화 + `MsgProcessUpdate` 발행 배선 완료(build/vet 통과)
- 2026-07-22(3) `POST /processes/resize/:processId` 최초 핸들러(이후 (4)에서 결과값 무시 버그 발견·수정)
- 2026-07-22 프론트 `ProcessDialog.vue`+`processDialog.store.ts` 신설(xterm+FitAddon), `DATA` 이벤트 연동(uid 필터+base64 디코드), uid 전환 시 화면 리셋, `resizeProcess` 클라이언트 추가, `App.vue` 전역 마운트

### `history/process-reconnect.md` — 종료/재접속 모델 (worker 끊김→PENDING→재바인딩)
- 2026-07-22 PENDING 오삭제 버그(정상 실행 vs 끊김 합성 혼동) 수정
- 2026-07-14 (2) worker 끊김→PENDING→재접속 재바인딩 구현 완료 + e2e 검증(applyStatus 가드 완화, worker `WorkerState` 신설 포함)
- 2026-07-14 (1) 위 설계 확정(코드 변경 없음, 순수 설계)

### `history/process-subscription.md` — 세션→uid 원장 + REST 구독/해지 배선
- 2026-07-16 REST 구독/해지(GET/POST/DELETE /processes/*) + `browsers`(conn→sid) registry 배선 완료 — 세션→uid 원장 마지막 조각
- 2026-07-14 (3) 세션→uid 원장 설계 착수: sid 추출 배선(코드 완료, NameFunc에 req 추가) + `process_subscribers` 테이블 설계(미구현)
