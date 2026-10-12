# REF — process 종료 정책 (종료 요청 · 끊긴 장비 kill · 강제 종료)

> 2026-10-12 신설. 이력 → `history/process-kill.md` / 재접속 동기화 → `INDEX-process-wiring.md`(REF-process-reconnect) / 타일 저장·고아 안전망 → `INDEX-node-ui-link.md`(REF-node-ui-save) / 현재 진행 → `CURRENT.md`

## 확정 (2026-10-12 사용자)
| 항목 | 결정 |
|---|---|
| EXEC 실행 방식 | A: 직접 실행, shebang 필수(현상 유지) |
| kill 첫 요청 | SIGTERM(정리 기회) → 상태 "종료 요청됨" |
| 끊긴 장비 kill | DB에 요청 기록 → 재접속 동기화 때 worker에 전달, 결과 수신 |
| 강제 종료 조건 ① | 종료 요청 뒤 **조작(20) 이상 계정의 보드에 그 터미널 타일이 없음**(최소화는 열린 것으로 셈) → SIGKILL. 보기(10) 계정 타일은 세지 않음(그 타일엔 종료 결과만 보임) |
| 종료 요청 권한 | 조작(20) 이상(기존 kill 인가 그대로, 보기 = 403 — SH3(받은 계정 권한)) |
| 강제 종료 조건 ② | "종료 요청됨"에서 다시 종료 = 강제 종료(칩 `종료 중▾` → 강제 종료… / X 창 = 취소·강제 종료·닫기). 매번 확인 창 |
| 타일 닫기 | 종료 요청 뒤엔 실행 중이어도 닫기 허용(지금은 PENDING·PROCESS면 막힘) |
| 다시 열기 | 모두 닫기 전이면 스크립트 행 "실행 중 N" → "위치 없음" → 내 보드에 열기(`openProcessTile`, 기존) |

## 조사 (2026-10-12)
- kill 경로: `processApi.go killProcess` → `requireProcess`(메모리 entry 필수 → **PENDING이면 404 "존재하지 않는 process"**) → `AgentInteractive.Kill` → `MsgKill` → worker `kill` → `pty.Interactive.Kill` = 프로세스 그룹에 SIGKILL
- 재접속: `reconcileReconnect`(교집합 Rebind + applyStatus / DB만 = FAILED / worker만 = 무시)
- **고아 안전망 `adoptOrphans`(tile.go)**: 살아 있는데 내 트리에 타일이 없는 process를 기본 보드 루트에 다시 붙임 → 종료 요청된 process는 빼야 "닫기"가 유지됨
- 타일 트리 쓰기: `putTiles`(사용자 저장 = 닫기가 여기) / `placeTerminal`·`adoptOrphans`·작업 공간 이동(추가·이동) / 공유 해제 정리
- 폴더 타일에도 `uid`가 있음 → "모두 닫음" 판정은 `type = terminal`만

## 구조안 (2026-10-12 제시 · 2026-10-12(2) 종료-4(모두 닫음) 개정 · **2026-10-12(3) 전부 권장안대로 승인**)
| 단계 | 할 일 | 주요 파일 |
|---|---|---|
| 종료-1(신호 구분) | `KillRequest.Force` + `Interactive.Kill(force)` → 프로세스 그룹에 SIGTERM/SIGKILL | `protocol/messages.go`, `execute/*`(AgentInteractive·pty·tmux·fifo), worker `process.go` |
| 종료-2(요청 기록) | `00010`: `processes.kill_requested_at`·`kill_forced`. kill API `{force}` = DB 기록 + 접속 중이면 전달(PENDING도 404 아님). `PROCESS:UPDATE`로 종료 요청됨 전파 | `processApi.go`, `perm.go`, `processDto.go` |
| 종료-3(재접속 전달) | `reconcileReconnect` Rebind 뒤 요청 있으면 전달(강제 = `kill_forced` 또는 모두 닫음) | `process.go` |
| 종료-4(모두 닫음) | `putTiles` 커밋 뒤 빠진 터미널 uid 중 종료 요청된 것 + **공유 해제·권한 변경 커밋 뒤** 그 공유에 걸린 종료 요청 process → `process_permission >= 20` 계정 트리에 `type=terminal` 없음이면 `kill_forced` + 전달. `adoptOrphans`·공유 자동 배치는 종료 요청된 것 건너뜀 | `tile.go`, 공유 해제·권한 변경, 쿼리 1개 |
| 종료-5(프론트) | 칩 `종료 중▾`(강제 종료…) / X 창(취소·강제 종료·닫기) / 종료 요청 뒤 닫기 허용 / "종료하고 닫기" = 요청 후 바로 닫기 / 목록 "종료 중" | `terminalKill.hook.ts`, 머리 줄, `TerminalListRow`, `ScriptRunList` |
| 종료-6(확인) | 격리 + 실제 worker. `trap '' TERM` 스크립트·TERM에 끝나는 스크립트 / 끊긴 장비 kill → 재접속 전달 / 모두 닫음 → SIGKILL / 다 닫기 전 다시 열기 / 두 번째 = 강제 | — |

| 결정 | 권장안 | 다른 안 |
|---|---|---|
| KP1(기록 위치) | `processes` 컬럼(supervisor 재시작에도 남음) | 메모리만 |
| KP2(상태 표현) | 상태 값 유지(PENDING/PROCESS) + 응답 `terminating`·`forced` — 상태 전이 단일 깔때기 불변 | 새 상태 `TERMINATING`(CHECK·깔때기 변경) |
| KP3(자동으로 다시 붙이기) | 고아 안전망·공유 자동 배치는 건너뜀, "내 보드에 열기"(수동)는 허용 | 모두 막음 |
| KP4(동시성 한계) | 커밋 뒤 한 번 판정. 같은 순간 다른 계정이 다시 열면 SIGKILL이 갈 수 있음 → 다시 연 터미널에 종료 결과가 보이는 정도로 허용 | 전 계정 트리 잠금(무거움) |

## 추가 결정 (2026-10-12(3), 구현 중 사용자 확인)
| 항목 | 결정 |
|---|---|
| "종료하고 닫기" | **기다렸다 닫기**: SIGTERM → 최대 10초 종료 대기 → 타일 닫기. 그때도 살아 있으면 닫기로 종료-4(모두 닫음) 판정 → SIGKILL. (바로 닫으면 혼자 보던 터미널은 SIGTERM 직후 SIGKILL이라 정리 틈이 없음) |
| 요청 때 판정 | **안 함**: 요청 시점에 조작 계정 보드에 타일이 이미 없어도(스크립트 행 "위치 없음"에서 종료) SIGTERM만. 판정은 타일 닫기·공유 변경 때만 → 안 끝나면 두 번째 종료(강제 종료) |

## 구현 메모 (2026-10-12(3), 종료-1(신호 구분)~종료-5(프론트) 코드 · 종료-6(확인) 통과 · 2026-10-12(6) 사용자 확인·마감)
| 단계 | 구현 |
|---|---|
| 종료-1(신호 구분) | `KillRequest.Force` / `IInteractive.Kill(force)` / `pty.Interactive.Kill` = 프로세스 그룹에 SIGTERM·SIGKILL. exec 보상(타일 저장 실패)·worker uid 중복은 `Kill(true)`. tmux·fifo는 IInteractive 구현체가 아니라 안 건드림 |
| 종료-2(요청 기록) | `00010_process_kill.sql`(`kill_requested_at`·`kill_forced`). `RequestProcessKill`(시각은 첫 요청 유지, force는 OR, 끝난 process = 행 없음 → 404). `killProcess`는 `{force}` 받고 `requireProcess`(memory 필수) 대신 `processPerm` + DB → PENDING도 기록. 커밋 뒤 `sendKill` = `PROCESS:UPDATE` 발행 + 접속 중이면 전달. 응답 = processResponse(`terminating`·`forced`) |
| 종료-3(재접속 전달) | `reconcileReconnect` 교집합에서 Rebind·relay·applyStatus 뒤 `kill_requested_at`이면 `Kill(kill_forced)` — EVENT 배달 고루틴이라 `go`로 띄움 |
| 종료-4(모두 닫음) | `ForceAbandonedKills`(UPDATE 한 문장, `process_permission >= min_perm` 계정 트리에 terminal 없음 → `kill_forced`). 부르는 곳 = `putTiles` 커밋 뒤(`closedTerminals` = 저장본에서 빠진 uid) + `reconcileAccess`(공유 걸기·풀기·노드 이동·삭제 커밋 뒤 공통 지점) |
| KP3(자동으로 다시 붙이기) | `ListOwnLiveProcesses`에 `kill_requested_at IS NULL`(고아 안전망) / `attachSharedTerminals`·`putProcessShare` 자동 배치 건너뜀 / `openProcessTile`(내 보드에 열기)은 허용 |
| 종료-5(프론트) | `pillOf` = 실행 중 + terminating이면 칩 "종료 중"(warning). 칩 메뉴: 처음 = 종료…·종료하고 닫기… / 종료 중 = 강제 종료…. X·목록 행 = `confirmKill` 기본 선택지가 상태별(처음 = 종료·종료하고 닫기 / 종료 중 = 강제 종료·닫기). 스크립트 행 = 종료 / 종료 중이면 강제 종료(`runs` 응답에 `terminating`·`forced`, `PROCESS:UPDATE`에 목록 갱신) |
| 프론트 부수 | kill 응답은 terminating·forced만 반영(상태는 STATUS가 먼저 올 수 있음) / `PROCESS:UPDATE`가 끝난 process를 실행 중으로 되돌리지 않게 무시 / `prune`이 트리에서 빠진 종료 요청 터미널 xterm도 정리 |

- 한계: kill 커밋과 worker 재접속이 겹치면 SIGTERM이 두 번 갈 수 있음(`sendKill` + 재접속 전달). 첫 TERM을 trap하는 스크립트만 영향 → KP4(동시성 한계)와 같은 수준으로 둠
- 보기(10) 계정 X(2026-10-12(3) 사용자 결정 "그냥 닫기"): 실행 중이어도 확인창 없이 타일만 닫음 + 칩 종료 메뉴 숨김(TileFrame·TerminalListRow). 전엔 종료 창 → 403
- 배포 주의: supervisor 재시작 때 `00010` 자동 적용 / worker도 다시 빌드해야 함(옛 worker는 force를 몰라 늘 SIGKILL — 옛 `Kill()`)

## 종료-6(확인) 결과 (2026-10-12(3), 격리 실행)
- 환경: 임시 DB `nexus_kill_tmp` + supervisor 5051 + worker `NAME=kill-dev`를 **socat 5152 → 5051 경유**(끊김 재현) + 격리 프론트 3001. 끝난 뒤 전부 끄고 DB 삭제
- 스크립트: trap = `trap 'echo got-term' TERM` + `while true; do sleep 0.2; done`(SIGTERM에 안 끝남) / plain = 같은 루프만(SIGTERM에 143)
- **API 22/22** (`kill_test.py`, 계정 A 소유 · B 보기 · C 조작): SIGTERM → 143 / 끝난 것 404 / trap은 종료 중으로 남고 두 번째(force) → 137 / 종료 요청 뒤 혼자 닫음 → 137 / 보기 kill 403 / C(조작)가 열어 두면 A가 닫아도 실행 중 / 고아 안전망이 다시 안 붙임 / runs `terminating` / 내 보드에 열기 허용 / C를 보기로 내림 → 137(B·C 타일 남음) / 요청 때 타일 없어도 SIGTERM만 / 끊김 → PENDING kill 200 → 재접속 143 / 끊긴 동안 모두 닫음 → `forced` 기록 → 재접속 137 / C 소켓에 `PROCESS:UPDATE`(terminating)
- **화면 16/16** (`pw/kill.mjs`, A·B 브라우저): 칩 RUN → 종료 → "종료 중"(B 화면도) / 확인창 처음 = 취소·종료·종료하고 닫기, 종료 중 = 취소·강제 종료·닫기 + 제목 "종료 중" / 칩 메뉴 = 강제 종료… / 보기 계정 X = 확인창 없이 닫힘, process 유지 / 강제 종료 → FAIL 137 / 종료하고 닫기: plain = 곧 닫힘 143, trap = 5초엔 남고 10초 뒤 닫힘 → 137(B 보기 타일 있어도)
- 격리에서 확인 안 한 것(→ 2026-10-12(6) 사용자가 확인): 스크립트 행(ScriptRunList) "종료 중"·강제 종료 버튼 화면 / supervisor 재시작 뒤 기록 유지(KP1(기록 위치))
- 재작성 명세(scratchpad라 재부팅 시 사라짐): 위 항목 그대로. 끊김 = `pkill -f 'TCP-LISTEN:515[2]'` → 다시 `socat TCP-LISTEN:5152,reuseaddr,fork TCP:localhost:5051`. 화면 exec = `page.request.post`(헤더 `X-Tab-Id` = sessionStorage의 tab 키) 후 새로고침. 다른 함정은 `REF-test-env.md`
- 사용자 확인용 스크립트(trap·plain + 종료 코드 표) → `user/kill-test-scripts.md`

## 실행 실패 정리 (2026-10-12(5), 사용자 테스트에서 발견)
| 문제 | 원인 | 수정 |
|---|---|---|
| shebang 없는 스크립트 실행 실패가 "실행 중"(PENDING)으로 남음 → 종료해도 대상 없음 | `Exec` 실패 경로 4곳(선배치·MsgExec 전송·응답 해석·worker 거부)이 메모리만 지우고 DB 행은 PENDING 그대로(기존 버그) | `failExec` = `applyStatus(FAILED, 126)` + (EDIT면) `Remove`. DB가 먼저 닫히고 relay가 FAILED 발행 → 목록 재조회에서 바로 빠짐. 126 = 실행 불가 관례 |
| worker 재접속으로 유령 행이 FAILED가 돼도 프론트는 새로고침해야 반영 | 실패한 실행은 relay 종료 때 토픽 구독이 정리돼 `publishLost`의 FAILED를 받을 소켓이 없음 | `publishLost` 앞에 `subscribeViewers(uid)` |
- 뺀 것: 실패 알림에 "shebang 필요" 안내 — 거부 사유가 shebang만은 아니라서(사용자 결정)
- 격리 확인 6/6(`fail_test.py`): 실행 실패 500 · DB FAILED 126 · 소켓 STATUS FAILED(126) · runs 빈 목록 / 소켓 연결 뒤 넣은 유령 PENDING 행 → worker 재시작 → DB 502 + 새로고침 없이 소켓에 STATUS FAILED
- 사용자 테스트 중 생긴 개발 DB 유령 4건(node 181)은 worker 재시작으로 FAILED 502 정리됨(사용자 확인)
