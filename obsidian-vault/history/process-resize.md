# history/process-resize — process resize(rows/cols) 배선 + `Layout` 에러 계약 정정

> 설계·재사용 지식 → `REF-process-resize.md` / exec·kill 트리거 이력 → `history/process-trigger.md` / 계약·원칙 → `REF-process.md` / 프론트 실제 배선 이력 → `history/frontend.md` / 현재 진행 → `CURRENT.md`

## 2026-07-22(4) — resize REST 배선 + DB/memory 동기화 + `MsgProcessUpdate` 발행 (build/vet 통과)

사용자가 `POST /processes/resize/:processId` 초안(entry.Inter.Layout 호출만 하고 응답 200 고정) 이후 "worker와 결과값 주고받는 로직과 entry의 process 객체 동기화는 아직 없지?"라고 직접 지적하며 시작. 대화로 두 가지를 순서대로 정리:

**① `AgentInteractive.Layout`이 `worker.Call`의 결과를 버리고 있던 것 발견** — `onLayout(cols, rows)`만 호출하고 무조건 `syscall.Errno(0)`을 반환하던 코드를 짚음. 사용자가 "worker의 resize 핸들러에서 응답도 해줘야 하는거지?"라고 물어, worker `resize` 핸들러(`cmd/worker/router/process.go`)는 이미 정상적으로 RES(에러 유무)를 돌려주고 있었고 손볼 곳은 supervisor `AgentInteractive.Layout`/`onLayout` 클로저 하나뿐임을 확인.

**② DB/memory 동기화 시점** — "REST 핸들러에서 곧바로 DB/memory를 갱신할지, worker 응답을 기다렸다 갱신할지" 사용자 고민에 기존 패턴(`MsgExec`/`MsgKill`=접수 ack만 동기, 실제 상태는 `MsgStatus` 비동기 이벤트에서만 확정)을 근거로 두 옵션을 제시. `MsgResize`는 뒤이어 오는 별도 확인 이벤트가 없는 순수 REQ/RES라 "결과값 기반"도 같은 핸들러 안에서 동기로 끝낼 수 있다는 점을 짚어 B안(결과값 기반) 권장 → 사용자 채택, 전제조건으로 `Layout` 에러 계약 정정을 먼저 요청("error를 error로 바꾸고 onLayout·Layout까지 고쳐줘").

**구현**:
- `IInteractive.Layout(cols, rows uint16) syscall.Errno` → `error`로 인터페이스 변경. `pty.Interactive.Layout`은 `errNo != 0`이면 그 값을(자체 `Error()` 있음) 반환, 아니면 `nil`. `AgentInteractive`의 `onLayout` 필드/파라미터 타입도 `func(cols, rows uint16) error`로 맞추고 `Layout`은 `return a.onLayout(cols, rows)`로 단순화. `newWorkerInteractive`(`manager.go`)의 `onLayout` 클로저가 `worker.Call` 에러를 그대로 반환하도록 수정. 콜사이트 두 곳(`errno != 0` → `err != nil`) 정정: worker `resize` 핸들러, `resizeProcess`. `go build`/`go vet` 클린.
- `internal/supervisor/db/query/processes.sql`의 `UpdateProcessLayout`을 `:exec` → `:one RETURNING *`로 바꾸고 `sqlc generate` 재생성.
- `resizeProcess`(`processApi.go`): `entry.Inter.Layout(...)` 동기 대기(실패 시 500, DB/memory 미변경) → 성공 시 `TxQueries(c).UpdateProcessLayout(...)` → `entry.SetRecord(&row)` → `AfterCommit(c, func(){ r.publishProcess(row) })`(node CRUD의 "커밋 후에만 발행" 안전장치 재사용) → 응답도 `newProcessResponse(row)`.
- `protocol.MsgProcessUpdate = "PROCESS:UPDATE"` 신설(`messages.go`) + `publishProcess(rec)` 헬퍼(`process.go`, `publishNode`와 대칭).

상세 → `REF-process-resize.md`.

## 2026-07-22(3) — `POST /processes/resize/:processId` 최초 핸들러 (build/vet 통과)

사용자가 `execProcess`의 응답을 `{uid}`에서 `processResponse` 전체로 바꾼 뒤(사용자 직접 수정, `toProcessResponse` 변환은 아직 안 거침 — `REF-frontend.md` 참조), 이어서 "우선 frontend에서의 resize를 처리할 API의 핸들러가 필요해"라고 요청. `mountProcesses`에 라우트 추가 + `resizeRequest{rows, cols}` 바인딩 + `entry.Inter.Layout(cols, rows)` 호출만 있는 최소 구현으로 시작(이 시점엔 `Layout`이 아직 결과를 안 버렸는지 확인 안 된 상태 — 위 2026-07-22(4)에서 버그로 드러남). `entry`가 없거나 `Inter == nil`(folder-open)이면 404.

## 프론트: `ProcessDialog.vue` + `processDialog.store.ts` — xterm 연동 (사용자 주도 구현, 2026-07-22)

`feature/process/store/processDialog.store.ts`(스켈레톤 요청 후 assistant가 `appDialog.store.ts`와 동일 provide/inject 패턴으로 작성 — 현재 다이얼로그가 보여주는 process 1개만 관리, `openProcessDialog`/`closeProcessDialog`/`patchStatus`)와 `feature/process/component/ProcessDialog.vue`(v-dialog+title bar(uid/status)+Close 버튼(kill 안 함, 다이얼로그만 닫음)+xterm 마운트 지점)를 assistant가 우선 생성. `open` 전이 시 `nextTick` 후 `Terminal`+`FitAddon` 생성/`fit()` → 확정된 rows/cols를 로컬 ref에 보관.

**소켓 이벤트 필터링(DATA/STATUS)은 사용자가 직접 구현하겠다고 선언** — assistant는 그 자리를 완전히 비워둠(스텁도 안 남김). 사용자가 이후 직접: `PROCESS:UPDATE`/`DATA` 리스너 등록(스텁 `console.log`), `rows`/`cols`/`process` 변경 watch → `resizeProcess` API 호출(2026-07-22(3) 핸들러 대상)까지 붙임.

이어서 사용자가 "DATA 이벤트 xterm과 연동시켜줄 수 있어?"라고 요청 → assistant가 `on<DataEventPayload>('DATA', ...)`에서 `ev.uid === process.value?.uid`로 걸러 base64(`protocol.DataEvent.Data`가 `[]byte`라 JSON에선 base64 문자열)를 `Uint8Array`로 디코드해 `terminal.write()`. 부가로 "다이얼로그가 열린 채로 대상 process(uid)만 바뀌는 경우" 기존 `watch(open,...)`가 `true→true` 전이라 재발화하지 않아 이전 화면이 남는 문제를 발견해 `watch(() => process.value?.uid, ...)`로 `terminal.reset()` 추가.

`process.api.ts`에 `resizeProcess(processId, {rows, cols})` 클라이언트 함수 추가(다른 함수들과 동일하게 `ProcessResponseDto`→`toProcessResponse` 변환 경유). `App.vue`에 `provideProcessDialog()` + `<process-dialog/>` 전역 마운트(`AppDialog`와 동일 위치 패턴). `index.vue`(임시 테스트 페이지) EXEC 버튼이 `execProcess` 성공 시 `openProcessDialog(res)` 호출하도록 변경.

상세 → `REF-process-resize.md` "ProcessDialog" 절(REST 클라이언트 표는 `REF-frontend.md`).
