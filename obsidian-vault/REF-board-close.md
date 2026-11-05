# REF — 보드 닫기 · 종료 유예(서버 10초 규칙) (2026-11-05 신설)

> → 이력 `history/board-close.md` · 종료 정책 원본 `REF-process-kill.md` · 보드(작업 공간) `REF-workspace.md` · 진행 `CURRENT.md`

## 요청 (2026-11-05, 사용자)
- 보드에서 **루트 타일을 닫을 수 있게**, 닫기 직전에 **보드의 모든 타일 종료 / 기본 보드로 보내기**를 묻는다
- 조율 답: Q1(대상 보드) = 기본 보드(`/`, 가입 직후 보이는 보드)는 제외 · Q2(루트 보드 뜻) = 기본 보드 · Q3(모두 종료의 뜻) = 실행 중 process까지 종료(가) · Q4(진입점) = 보드 메뉴 "보드 삭제"도 같은 확인 창
- 10초 대기를 어디서: 사용자 **S-b(모든 종료 요청을 서버가 10초 뒤 판정)** 선택 — "종료 요청 직후 브라우저를 닫아도 문제없게"
- D2(종료 못 하는 터미널) = 확인 창에 한 줄 · D3(실행 중 터미널 없음) = 두 버튼 그대로 · D4(종료하고 닫기) = 요청 직후 바로 닫기 · D5(판정 방식) = 주기 검사 — 전부 추천대로(사용자 "추천대로")

## 종료 유예 (S-b) — 바뀐 규칙
| 항목 | 전(10-12 결정) | 지금 |
|---|---|---|
| 강제 종료 조건 | 종료 요청 + 조작 이상 계정 보드에 타일 없음 | **+ 종료 요청 뒤 10초(`killGrace`) 지남** |
| 판정 시점 | 타일 닫기·공유 변경 때만 | 그때(유예 지난 것만) + **2초마다 주기 검사**(`runKillSweep`, 유예 지난 종료 요청 전부) |
| 요청 때 판정 | 안 함 | 10초 뒤엔 함 — 스크립트 행 "위치 없음"에서 종료한 것도 TERM을 무시하면 10초 뒤 SIGKILL |
| 타일이 열려 있음 | SIGTERM만 | 그대로(강제 = 두 번째 종료) |
| "종료하고 닫기" | 프론트가 최대 10초 기다린 뒤 닫음 | **요청 직후 바로 닫음**(`CLOSE_WAIT`·`waitEnded` 삭제). 브라우저를 닫아도 서버가 판정 |
| supervisor 재시작 | — | 주기 검사가 다음 회차에 이어받음(타이머 없음) |
- 실제 강제 시점 = 요청 뒤 10~12초(검사 간격 2초). worker가 끊겨 있으면 `kill_forced`만 기록되고 재접속 때 전달(전과 같음)
- **쉬기(안 A(메모리 표시), 11-05 사용자)**: `killPending`(atomic) 꺼져 있으면 주기 검사가 DB를 안 봄. 켜는 곳 = supervisor 시작 · 종료 요청 커밋 뒤(`killProcess`·`killWorkspace`의 `markKillPending`). 검사 순서 = 끄기 -> 판정 -> `HasPendingKills`(실행 중·요청됨·아직 강제 아님)가 참이면 다시 켬(순서를 바꾸면 그 틈의 요청을 놓침). 타일을 열어 둔 "종료 중" process가 있으면 끝날 때까지 2초마다 돎
- 같은 행을 두 번 보내지 않음: UPDATE가 `NOT kill_forced`인 행만 바꿔 돌려줌(타일 닫기 판정과 주기 검사가 겹쳐도 한 번)

## 보드 닫기
| 진입 | 확인 창 `'<보드 이름>' 닫기` |
|---|---|
| 루트 타일 ×(기본 보드 아닐 때만, 컨트롤 패널 닫기 포함) · 보드 메뉴 "보드 삭제" | `취소`(Esc) / `기본 보드로 보내기` / `모두 종료하고 닫기` + 안내 3줄(보내기 · 모두 종료 · 보기 권한 공유 터미널은 종료 안 되고 이 보드에서만 빠짐) |
- 기본 보드로 보내기 = 전과 같은 `DELETE /workspaces/:hash`(W2(삭제 처리))
- 모두 종료하고 닫기 = `DELETE /workspaces/:hash?mode=kill`: 트리 안 터미널 uid 중 **조작 이상**(`processPerm >= PermOperate`)이고 실행 중인 것에 `RequestProcessKill`(SIGTERM) → 보드 행 삭제(옮기지 않음) → 커밋 뒤 `sendKill` · `resyncSizeOwners` · `forceAbandoned`(전에 요청돼 유예가 지난 것은 바로). 응답 = 기본 보드 트리(그대로). 끝나면 프론트가 `/`로
- 종료 요청된 process는 고아 안전망(`adoptOrphans`)이 기본 보드에 다시 붙이지 않음(KP3, 확인함). 끝나기 전까지 기본 보드 루트의 터미널 탭(경로 기준 조회)에 "종료 중"으로 보임

## 코드 지도
| 파일 | 내용 |
|---|---|
| `internal/supervisor/db/query/processes.sql` `ForceAbandonedKills` | 인자 `all_uids`·`grace_secs` 추가(`kill_requested_at <= NOW() - make_interval(secs => grace)`) |
| `cmd/supervisor/router/killGrace.go` | `killGrace`·`killSweepEvery` · `runKillSweep` · `forceKills(all, uids)` · `markKillPending` |
| `processes.sql` `HasPendingKills` · `supervisorRouter.go` `killPending` | 쉬기 판단 |
| `processApi.go` `forceAbandoned` | `forceKills(false, uids)` 호출만 |
| `supervisorRouter.go` | `go router.runKillSweep()` |
| `workspace.go` `deleteWorkspace`·`killWorkspace` | `?mode=kill` 분기(그 외 값 400) |
| 프론트 `api/board.api.ts`·`store/boards.store.ts` | `BoardCloseMode = 'move' \| 'kill'` |
| `hook/boardClose.hook.ts` | `useBoardClose` = `closable`(지금 라우트 보드가 기본이 아님) · `closing`(모듈 상태) · `confirmClose` |
| `component/tile/TileFrame.vue` | 루트 × = `closeShown`·`onClose`(루트면 보드 닫기, 아니면 `emit('close')`), `tileHeadActs`도 같게 |
| `component/BoardMenu.vue` | "보드 삭제" = `confirmClose` |
| `hook/terminalKill.hook.ts` | 종료하고 닫기 = 요청 뒤 바로 `closeTile`, 안내 문구 |
| `FolderTile.vue`·`DeviceTile.vue` | 루트 "볼 수 없음" 안내 문구(×로 보드 닫기) |

## 확인 (2026-11-05, 격리 DB `nexus_bc_tmp` + 5051 + 실제 worker, 끝나고 지움)
- (쉬기 추가 뒤 다시) API 13/13: 아래 11개 + 남은 요청 없을 때 10초 동안 검사 쿼리 0번 · 종료 요청 뒤 6초에 3번. **측정 함정**: `pg_stat_user_tables` 스캔 수는 늦게 반영돼 쉬는데도 +3~5로 보임 -> 임시 DB에 `ALTER DATABASE … SET log_statement = 'all'` 후 `docker logs postgres15 --since`에서 `kill_requested_at <= NOW()` 줄 수로 셈
- API 11/11(`bc_test.py`, 스크래치 — 다시 쓸 땐 이 목록대로): mode=kill 보드 삭제 · TERM 스크립트 0.1초 종료 · TERM 무시(`trap '' TERM` 루프) 6초엔 살아 있음 · 기본 보드로 안 옮겨짐 · 11.5초 강제(`FAILED,true`) / 종료+타일 바로 닫기 4초엔 살아 있음 · 11.3초 강제 / 타일 열어 둠 14초 뒤에도 살아 있음 · 두 번째 종료로 끝 / 유예 안에 supervisor 끔 → 12초 뒤 다시 켬 → 1.8초 만에 강제 / mode 생략 = 기본 보드로 옮겨지고 계속 실행
- 화면 9/9(`bc_ui.mjs`, 헤드리스): 기본 보드 루트 × 없음 · 보드 루트 × 있음 · 확인 창 버튼 3개·제목 · 취소 = 그대로 · 모두 종료하고 닫기 = `/`로, 보드 삭제, 폴더 타일 안 옮김 · 보드 메뉴 "보드 삭제" = 같은 창 · 보내기 = 폴더 타일 기본 보드로 · 페이지 오류 없음
- 사용자 화면 확인(11-05, 개발 8011·8012 + 계정 `devtest` + worker `devtest-pc`): "잘 동작해" -> 마감, master 머지
- 못 본 것: 보기 권한만 있는 공유 터미널(두 번째 계정·공유 필요) — 코드는 권한 검사로 건너뜀
- 확인 클라이언트 함정: 실행(`/workers/.../exec`)은 `X-Tab-Id` 필요 → `/api/subscribe` 소켓을 열어 `TAB:ID`를 받아 헤더에 넣음. supervisor를 다시 켜면 소켓도 다시
