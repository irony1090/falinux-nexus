# REF — 브라우저 소켓 재연결 · 끊김 표시 (2026-11-06 신설)

> → 이력 `history/socket-reconnect.md` · 실시간 push 구조 `REF-realtime.md` · 터미널 복원·이음매 `REF-process-keep.md` · 진행 `CURRENT.md`

## 증상 (2026-11-06, 사용자)
- 모바일에서 브라우저를 껐다 켜면(백그라운드 -> 복귀) 터미널 출력이 멈춤. TUI 터미널은 화면이 안 바뀜
- 기대 = 화면으로 돌아오면 저절로 다시 붙고, 그동안 놓친 화면도 보임(사용자 "맞아")

## 원인
| 이름 | 내용 |
|---|---|
| 가(포기) | `provideTestSocket`이 `maxRetries: 2` — 2초 간격 2번 실패하면 `PENDING`으로 멈추고 다시 시도하는 곳이 없음. 모바일 복귀 직후엔 네트워크가 덜 깨어 2번이 쉽게 실패 |
| 나(화면 빠짐) | 다시 붙으면 서버가 실행 중 process 토픽을 다시 구독(`subscribeLiveProcesses`)해 새 출력은 오지만 끊긴 사이 출력은 버려짐. `restore`는 이미 연 터미널을 건너뜀 -> TUI(바뀐 곳만 다시 그림)는 옛 화면이 남음. `writeLive`가 빈 곳에 `earlier output not shown` 줄만 씀 |
| 다(+버그) | 옛 `onclose`는 `status !== 'CONNECTING'`일 때만 재연결 예약 -> 연결 중에 오류 없이 닫히면 영영 재시도 안 함 |

## 고친 것 (사용자 확정: 재연결-3·끊김 표시 포함)
| 이름 | 내용 | 곳 |
|---|---|---|
| 재연결-1(포기 없음) | 횟수 제한 없음, 간격 = `retryInterval * 2^실패수`(2·4·8·16·30초, `retryMax` 30초). 예약 타이머 하나(`timer`)만 — 겹쳐 예약 안 함 | `websocket.hook.ts` `onclose` |
| 재연결-2(돌아오면 바로) | `visibilitychange`(visible)·`pageshow`·`online` -> `socket.wake()`: 끊겼으면 간격 무시하고 바로 연결 + 실패수 0. 연결 중인 채 5초(`STALE_MS`) 넘은 시도는 버리고 새로 | `socketWake.util.ts`(App.vue에서 한 번) · hook `wake` |
| 재연결-3(죽은 연결 확인) | `wake` 때 연결됨이면 `PING` 요청(3초, `PROBE_MS`). **시간 초과만** 죽은 연결 -> 버리고 새로 연결. 닫혀서 실패 = onclose가 이미 예약, 오류 응답 = 살아 있음(옛 서버엔 PING 핸들러가 없어 "핸들러 없음" 오류 응답 -> 살아 있다고 봄) | hook `probe` · 서버 `subscribe.go` `conn.Handle(MsgPing)` · `protocol.MsgPing = "PING"` |
| 화면-1(다시 받기) | 한 번 붙었다 끊긴 뒤 다시 붙으면(`dropped`) 실행 중 터미널마다 `lookupProcesses` -> 아직 실행 중 = 스냅샷 받아 `term.reset()` + `writeSnapshot` + 스냅샷 off 뒤 출력만, 끊긴 사이 끝남 = 모은 출력 쓰고 종료 상태. 처음 붙을 때는 안 함(복원 `restore`가 그림) | `processTerm.store.ts` `resync`·`redraw`·`resyncing` |
| 끊김 표시 | 로그인 중 소켓이 1.5초 넘게 연결 안 됨이면 위 가운데 알약 "연결 끊김 · 다시 연결 중"(테마 `warning`, 점 깜빡임, `pointer-events: none`). 붙으면 바로 사라짐 | `common/websocket/SocketLostPill.vue`(App.vue) |

- **순서 의존**: `resync`는 `resyncing` 표시를 **동기로 먼저** 한다. 다시 붙은 직후 오는 DATA가 스냅샷보다 먼저 xterm에 쓰이면 안 됨(DATA 핸들러가 `resyncing`에 있으면 모아 둠). `watch(socketStatus)`는 onopen 뒤 microtask에서 돌아 첫 onmessage보다 먼저
- `drop()` = 핸들러를 떼고 닫음 + 매달린 call 거절 — 옛 소켓 onclose가 재연결을 또 예약하지 않게. `disconnect`(로그아웃)는 `active = false` -> wake·onclose 모두 아무것도 안 함
- 알약 위치 = 위 가운데(첫 타일 머리 줄 오른쪽 버튼을 일부 가림, 누르기는 통과). 다른 자리를 원하면 사용자와 다시 정할 것

## 확인 (2026-11-06, 격리 DB `nexus_sr_tmp` + 5051 + 실제 worker, 끝나고 지움) — 14/14
- 방법: 브라우저 -> **node 프록시 5052** -> 5051. 프록시 모드 = `down`(연결 전부 끊고 새 연결 거절 = 네트워크 끊김) / `freeze`(기존 소켓 연결만 전달 멈춤 = 열린 채 죽은 연결) / `normal`. 스크립트 `tui.sh` = 1행 TICK 숫자 계속, `mark` 파일이 생기면 3행에 `MARK <값>` 한 번만(커서 위치로만 그림 = TUI 흉내). 화면 = `.xterm-rows` 글자
- A 12초 끊김(끊긴 사이 MARK A1): 알약 보임 -> 네트워크 복구 2.0초 뒤 저절로 붙음 -> MARK A1 보임, TICK 이어짐
- B 35초 끊김(간격 30초까지 늘어남): 복구 + `visibilitychange` -> 0.2초 만에 붙음 · MARK B1 보임
- C 얼림(MARK C1): 화면 TICK 멈춤 -> `visibilitychange` -> PING 3초 시간 초과 -> 새 연결 -> 3.2초에 MARK C1, TICK 다시 감, 알약 없음
- D 연결된 채 `visibilitychange` = PING만, 화면 그대로 · 페이지 오류 0
- 확인 함정: ① 실행 API를 **그 페이지 탭 id**(`X-Tab-Id`)로 보내면 타일 push를 그 탭이 "내 변경"으로 무시 -> 페이지 새로고침으로 복원해야 터미널이 보임 ② worker `TMPDIR`이 길면 보관 프로세스 소켓 경로 초과(119 > 85) -> `TMPDIR=/tmp/srt`처럼 짧게(실행 파일은 스크래치에 둬도 dev 판정 안 됨)
- **실제 모바일 = 사용자 확인 "모바일에서 잘 동작해"(11-06, test3 `261106-socket-reconnect`)**
- 못 본 것: supervisor 재시작 뒤 화면-1(스냅샷 출처가 바뀜, 이번 시험은 브라우저 쪽 끊김만)
