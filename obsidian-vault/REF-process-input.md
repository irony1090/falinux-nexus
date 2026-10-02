# REF — process 터미널 입력 (키 입력 배선)

> 2026-10-01 신설. 진행 순서상 **⑪(타일 트리 서버 저장) 앞**에 넣음(`REF-process-sync.md` "진행 순서"). 이력 → `history/process-input.md` / 출력·상태 배선 → `REF-process-wiring.md` / 타일 터미널 → `REF-node-ui-terminal.md` / 모바일 가상 키보드(⑮, 이 작업 뒤) → `REF-node-ui-overview.md` / 현재 진행 → `CURRENT.md`.

**상태: 결정 확정(2026-10-01). 입력-a(서버)·입력-b(프론트) 코드 + 입력-c(확인) 서버 4항목·브라우저 7항목 통과(2026-10-02, 미커밋).** 남은 확인 = PENDING 중 입력 버림(socat 프록시)뿐.

## 지금 있는 것 / 없는 것
| 구간 | 상태 |
|---|---|
| supervisor -> worker | **있음** — `AgentInteractive.Write` -> `onWrite`가 worker에 `MsgData`(EVENT) Emit (`process/manager.go` `newWorkerInteractive`) |
| worker -> PTY | **있음** — `worker/router/process.go` `input`이 `entry.inter.Write` |
| 브라우저 -> supervisor | **입력-a로 작성**(2026-10-02) — `router/input.go` `onProcessInput`/`input`, `subscribe.go`에서 Serve 전 등록 |
| 프론트 xterm | **입력-b로 켬**(2026-10-02) — `processTerm.store` `createHandle(uid)`에서 `onData`/`onBinary` -> `sendInput` -> `emit('PROCESS:INPUT')` |

## 결정 I1~I5 (I1~I4 2026-10-01 / I5 2026-10-02 사용자 확정)
| 결정 | 내용 | 기각 |
|---|---|---|
| I1(전달 경로) | **소켓 EVENT** — 브라우저가 `PROCESS:INPUT` Emit, 응답 없음 | REST(키마다 요청 — 고빈도에 부적합) |
| I2(입력 권한) | **같은 계정이면 어느 탭이든** 입력 가능(크기 우선권과 무관) | 크기 소유 탭만 |
| I3(PENDING 중 입력) | **전부 버림** — worker가 끊긴 동안 온 입력은 쌓지 않음. 실행 직후 RUNNING 전 PENDING도 같음 | 쌓았다가 재접속 뒤 전달 |
| I4(Ctrl+C 복사) | **선택 영역이 있으면 복사, 없으면 `0x03` 전송**(VS Code 터미널·Windows Terminal 방식) — xterm `attachCustomKeyEventHandler`에서 `hasSelection()`이면 클립보드 복사 후 전송 막음 | 항상 `0x03`(복사는 Ctrl+Shift+C·우클릭) |
| I5(Ctrl+V 붙여넣기) | **Ctrl+V = 붙여넣기**(2026-10-02 사용자 확정, I4와 같은 Windows Terminal 방식) — 핸들러가 `false`를 반환해 xterm의 ^V 변환을 건너뛰고 브라우저 paste 이벤트 -> xterm `onData` | 그대로(^V 전송, 붙여넣기는 Ctrl+Shift+V) |

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

## 입력-a 서버 단독 확인 (2026-10-02, 통과)
Node 22 내장 `WebSocket` 클라이언트로 브라우저 역할 대신. 스크립트 노드(`echo READY; cat`)를 폴더 1 아래 만들어 실행 -> 끝나고 삭제.

| 항목 | 결과 |
|---|---|
| A 계정 `hello\n` 입력 | 통과 — 출력 `READY\r\nhello\r\nhello\r\n`(PTY 에코 + cat) |
| B 계정(`input-test-b`, 이번에 가입)이 같은 uid로 입력 | 통과 — A 출력에 안 나타남 + supervisor 로그 `[process] 다른 계정의 입력 버림 uid=... user=4` 1줄(사용자 확인) |
| A 계정 `0x03` | 통과 — STATUS FAILED, exitCode 130(SIGINT) |
| 종료된 uid로 입력 | 통과 — 출력 없음, 소켓 유지 |

- 남은 확인(입력-c): PENDING 중 입력 버림(socat 프록시), 다른 탭 입력, 프론트 I4(Ctrl+C 복사)
- 테스트 클라이언트 작성 시 함정: 서버 프레임은 **binary**(`binaryType='arraybuffer'` 후 JSON) / exec REST는 **`X-Tab-Id` 헤더 필수**(소켓의 `TAB:ID` 값) / STATUS의 `status`는 **숫자**(1=PROCESS, 2=COMPLETED, 3=FAILED) / 입력 프레임 = `{k:2, t:'PROCESS:INPUT', d:{uid, data:base64}}`
- 스크립트는 세션 scratchpad `input/test.mjs`에만 있음(재부팅하면 사라짐)

## 입력-b 구현 (2026-10-02)
수정 파일은 `feature/process/store/processTerm.store.ts` 하나(xterm을 스토어가 가지므로 — `REF-node-ui-terminal.md` J).

| 부분 | 내용 |
|---|---|
| `createHandle(uid)` | `disableStdin` 제거, uid를 받아 입력 핸들러를 건다 |
| `sendInput` | `procs[uid].status === 'PROCESS'`일 때만 `emit('PROCESS:INPUT', {uid, data: base64})`. 소켓이 끊겨 있으면 `emit`이 버림(hook 동작) |
| `onData` | 문자열 -> `TextEncoder`(UTF-8) -> base64 |
| `onBinary` | 문자 하나 = 바이트 하나(마우스 보고 등) -> `charCodeAt & 0xff`. UTF-8 재인코딩 금지 |
| `clipboardKeys` | I5(Ctrl+V 붙여넣기) = Ctrl+V면 `false`(xterm이 ^V로 바꾸지 않고 브라우저 paste로). I4(Ctrl+C 복사) — `attachCustomKeyEventHandler`. `ev.code === 'KeyC'`(한글 입력 상태에서도 동작) + 선택 있음 -> keydown에서 `navigator.clipboard.writeText` + `clearSelection`, `false` 반환으로 `0x03` 막음. 클립보드 API가 없는 비보안 http에서도 막기만 함(복사하려던 Ctrl+C로 프로세스가 끊기지 않게) |
| 전체보기 | `TileFrame.vue`가 이미 `terminal-tile-body`에 `:inert="overview"` — 코드 변경 없음, 포커스 빠지는지는 입력-c에서 확인 |

- 발견 — **Ctrl+V는 xterm 6 기본 동작으로 `0x16`(^V)을 보내고 브라우저 붙여넣기를 막는다**(Ctrl+문자 -> keyCode-64, cancel). 붙여넣기는 Ctrl+Shift+V·우클릭 메뉴로는 됨(리눅스 터미널 관례와 같음). → I5(Ctrl+V 붙여넣기) = 붙여넣기로 확정·반영

## 입력-c 브라우저 확인 (2026-10-02, 7항목 통과)
playwright-core(세션 scratchpad) + `~/.cache/ms-playwright/chromium-1234`. 로그인(`Login.vue` 기본값) -> 폴더 `HTOP_TEST_SH_MODI` -> 테스트 스크립트(`echo READY; cat`) 실행 버튼 -> 타일 터미널.

| 항목 | 결과 |
|---|---|
| 1 타이핑 `hello` + Enter | 통과(에코 + cat 출력) |
| 2 방향키 | 통과(`^[[A` 전달) |
| 3 Ctrl+V | 통과 — 클립보드 내용 붙여넣기, `^V` 없음 |
| 4 같은 계정 다른 탭(별도 소켓)에서 입력 | 통과 — 브라우저 터미널에 표시 |
| 5 전체보기 중 키 입력 | 통과 — 포커스가 xterm 밖(`inert`), 입력 안 감 |
| 6 선택 있음 + Ctrl+C | 통과 — 클립보드 = 선택 글자, process 유지 |
| 7 선택 없음 + Ctrl+C | 통과 — `^C`, `[process exited: 130]` |

- 함정: 브라우저 쿠키 도메인은 `localhost` — Node에서 쿠키를 꺼내 소켓을 열 땐 `ctx.cookies('http://localhost:5050')` + `ws://localhost:5050`(127.0.0.1로 하면 쿠키 없음 -> 업그레이드 실패)
- 참고: Ctrl+C로 끝난 process는 상태 칩이 FAIL(exit 130 = 시그널 종료, 기존 동작)
- 스크립트 = 세션 scratchpad `input/ui.mjs`(재부팅하면 사라짐)
- 사용자 PC 직접 사용 확인(2026-10-02): "PC 기준으론 딱히 이상 없음" — 한글(IME) 입력을 따로 확인했다는 언급은 없음
