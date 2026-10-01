# REF — process 터미널 입력 (키 입력 배선)

> 2026-10-01 신설. 진행 순서상 **⑪(타일 트리 서버 저장) 앞**에 넣음(`REF-process-sync.md` "진행 순서"). 이력 → `history/process-input.md` / 출력·상태 배선 → `REF-process-wiring.md` / 타일 터미널 → `REF-node-ui-terminal.md` / 모바일 가상 키보드(⑮, 이 작업 뒤) → `REF-node-ui-overview.md` / 현재 진행 → `CURRENT.md`.

**상태: 결정 확정(2026-10-01), 코드 미착수.** 착수 승인 전 코드 금지.

## 지금 있는 것 / 없는 것
| 구간 | 상태 |
|---|---|
| supervisor -> worker | **있음** — `AgentInteractive.Write` -> `onWrite`가 worker에 `MsgData`(EVENT) Emit (`process/manager.go` `newWorkerInteractive`) |
| worker -> PTY | **있음** — `worker/router/process.go` `input`이 `entry.inter.Write` |
| 브라우저 -> supervisor | **없음** |
| 프론트 xterm | 입력 꺼짐(`processTerm.store` `createHandle`의 `disableStdin: true`) |

## 결정 I1~I4 (2026-10-01 사용자 확정)
| 결정 | 내용 | 기각 |
|---|---|---|
| I1(전달 경로) | **소켓 EVENT** — 브라우저가 `PROCESS:INPUT` Emit, 응답 없음 | REST(키마다 요청 — 고빈도에 부적합) |
| I2(입력 권한) | **같은 계정이면 어느 탭이든** 입력 가능(크기 우선권과 무관) | 크기 소유 탭만 |
| I3(PENDING 중 입력) | **전부 버림** — worker가 끊긴 동안 온 입력은 쌓지 않음. 실행 직후 RUNNING 전 PENDING도 같음 | 쌓았다가 재접속 뒤 전달 |
| I4(Ctrl+C 복사) | **선택 영역이 있으면 복사, 없으면 `0x03` 전송**(VS Code 터미널·Windows Terminal 방식) — xterm `attachCustomKeyEventHandler`에서 `hasSelection()`이면 클립보드 복사 후 전송 막음 | 항상 `0x03`(복사는 Ctrl+Shift+C·우클릭) |

- 조합 키는 상태(PROCESS)만 맞으면 전부 바이트로 전달(Ctrl+C = `0x03` -> PTY가 SIGINT, Ctrl+X = `0x18` 등). 예외는 브라우저가 먼저 가져가는 Ctrl+W·Ctrl+T·Ctrl+N(페이지에 오지 않아 보낼 수 없음).
- I3로 CURRENT의 옛 결정 항목 "끊긴 창 입력 거절 vs 큐잉" 중 **입력 부분은 해소**(kill 쪽은 그대로 남음).

## 구조안
```
xterm.onData / onBinary
  -> socket.emit('PROCESS:INPUT', {uid, data: base64})
  -> handleSubscribeWS: conn.On(MsgProcessInput)   (userID captured at connect)
       entry = processManager.Get(uid)
       drop if: no entry (ended / PENDING-detached) | Inter nil (folder)
                | owner != userID | status != PROCESS
       entry.Inter.Write(data) -> worker MsgData -> PTY
```

| 작업 단위 | 내용 | 주요 파일 |
|---|---|---|
| 입력-a(서버) | `MsgProcessInput = "PROCESS:INPUT"`(browser->sup EVENT, payload = 기존 `DataEvent{uid, data}` 재사용) + 소켓 연결 핸들러에서 `conn.On` 등록(Serve 전) + 버림 조건 4가지 + `Inter.Write` | `protocol/messages.go`, `router/subscribe.go`(또는 새 `router/input.go`) |
| 입력-b(프론트) | `disableStdin` 해제 + `onData`·`onBinary` -> UTF-8 바이트 -> base64 -> emit. I4(Ctrl+C 복사) 키 핸들러. 상태가 PROCESS가 아니면 보내지 않음(서버도 버리지만 불필요한 전송 방지). 전체보기 모드에선 터미널에 포커스가 가지 않게(타일 본문 `inert` 확인) | `processTerm.store.ts`, `ProcessTerminal.vue`/`TileFrame.vue` |
| 입력-c(확인) | 대화형 스크립트(`read`/`cat`)·`bash`에서 입력·Ctrl+C(선택 없음 = 중단 / 선택 있음 = 복사)·Ctrl+X·방향키 / 다른 탭 입력 / PENDING 중 입력 버림(socat 프록시) / 다른 계정 uid 입력 버림 | Node 클라이언트 + 헤드리스 |

## 세부
- 버림은 조용히(응답·로그 없음 — 키마다 로그가 쌓이므로). 소유 계정 불일치만 로그 1줄 검토.
- 권한 확인은 소켓 연결 시점의 계정(`sess.Data.ID`) 기준 — 연결 뒤 로그아웃한 탭은 소켓이 끊길 때까지 입력 가능(1단계 "남은 한계(로그아웃)"와 같은 계열, `REF-process-sync-impl.md`).
- 이 배선이 끝나면 EDIT(worker `vi`, `REF-process-exec-edit.md`)를 타일에서 쓸 수 있는 조건이 갖춰진다.
