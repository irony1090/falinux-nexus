# REF — process resize(rows/cols) 배선 + `Layout` 에러 계약 정정

> exec/kill 트리거는 `REF-process-trigger.md`. 계약/설계 원칙은 `REF-process.md`(`Layout` 계약 최신화 완료). 발행 패턴(node 도메인)은 `REF-realtime.md`. frontend 전반 규약(상태관리 방침·socket hook 등)은 `REF-frontend.md` — 이 파일에 backend resize 배선 + frontend `ProcessDialog`(다이얼로그·xterm) 실제 배선을 함께 묶는다. 작업 이력은 `history/process-resize.md`.

## 배경 — `AgentInteractive.Layout`이 worker 응답을 버리고 있었음 (발견 2026-07-22)

`entry.Inter.Layout(cols, rows)` → `AgentInteractive.Layout` → `onLayout` 클로저(`process/manager.go`의 `newWorkerInteractive`)가 `worker.Call(MsgResize, ...)`을 실제로 보내긴 했지만:
```go
func (a *AgentInteractive) Layout(cols, rows uint16) syscall.Errno {
    a.onLayout(cols, rows)   // worker.Call의 (Frame, error) 리턴값을 버림
    return 0                 // 항상 성공 고정
}
```
worker `resize` 핸들러(`cmd/worker/router/process.go`)는 원래도 정상적으로 RES(에러 유무)를 돌려주고 있었다 — 문제는 오직 supervisor 쪽에서 그 결과를 버리는 지점 하나였다.

## 결정 — DB/memory 동기화는 "worker가 확인해준 결과값" 기반 (B안)

이 코드베이스는 지금까지 "명령 접수(REQ/RES 동기 ack)"와 "실제 상태 확정(worker의 비동기 `MsgStatus` 보고 시점)"을 분리해왔다 — `MsgExec`/`MsgKill` 둘 다 접수 확인만 동기로 받고, DB `Mark*`+`entry.SetRecord`는 항상 뒤이어 오는 `MsgStatus` 핸들러(`applyStatus`)에서만 수행한다(`REF-process-trigger.md` "entry.Record memory 동기화" 절과 동일 원칙).

`MsgResize`는 다르다 — **뒤따르는 별도 확인 이벤트가 없는 순수 REQ/RES**라, worker 응답을 그 자리에서 기다리기만 하면 같은 REST 핸들러 안에서 동기로 "결과값 기반" 동기화를 끝낼 수 있다. 그래서:
- 초안(낙관적 — REST 성공 시 무조건 DB/memory 갱신)은 기각.
- **B안 채택**: `entry.Inter.Layout(...)`이 실제로 성공을 반환했을 때만 DB/memory를 갱신. 전제조건은 `Layout`이 진짜 결과를 반환하도록 계약부터 고치는 것.

## 구현 1 — `IInteractive.Layout`: `syscall.Errno` → `error`

- `internal/execute/iInteractive.go`: `Layout(cols, rows uint16) syscall.Errno` → `error`.
- `internal/execute/pty/execInteractive.go`(`Interactive.Layout`): ioctl `errNo != 0`이면 그 `syscall.Errno`를 그대로 반환(자체 `Error()` 구현이 있어 `error`로 대입 가능), 0이면 `nil`.
- `internal/execute/agentInteractive.go`: `onLayout` 필드/생성자 파라미터 타입을 `func(cols, rows uint16) error`로 변경, `AgentInteractive.Layout`은 `return a.onLayout(cols, rows)`(위임만, 자체 로직 없음).
- `cmd/supervisor/process/manager.go`의 `newWorkerInteractive`: `onLayout` 클로저가 `worker.Call(MsgResize, ...)`의 에러를 버리지 않고 반환.
- 콜사이트 두 곳(`errno != 0` → `err != nil`) 정정: `cmd/worker/router/process.go`의 `resize` 핸들러, `cmd/supervisor/router/processApi.go`의 `resizeProcess`.
- `IInteractive`/`AgentInteractive`/`pty.Interactive` 세 구현 전부 "그 자리에서 확인된 진짜 결과를 돌려준다"는 계약은 동일 — 로컬 PTY는 ioctl errno, 원격은 worker RES 에러.

## 구현 2 — `POST /processes/resize/:processId` (`processApi.go`)

```
resizeProcess(c)
  ├─ entry.Inter.Layout(cols, rows)          worker.Call(MsgResize) 왕복 대기(동기)
  │    실패 → 500, DB/memory/소켓 전부 미변경
  │    성공
  ├─ TxQueries(c).UpdateProcessLayout(uid, rows, cols)   DB 갱신 (:one RETURNING *)
  ├─ entry.SetRecord(&row)                                memory 갱신
  ├─ AfterCommit(c, func(){ r.publishProcess(row) })      commit 성공 후에만 소켓 발행
  └─ return newProcessResponse(row)                       REST 응답도 최신 객체
```
- `entry`가 없거나(`Inter == nil`, folder-open) 이미 종료된 uid는 404 — kill과 동일.
- **DB 쿼리 변경**: `UpdateProcessLayout`을 `:exec` → `:one ... RETURNING *`로 바꿔 `sqlc generate` 재생성(다른 `Mark*` 쿼리와 동일 패턴 — 갱신된 row를 그대로 `entry.SetRecord`에 쓰기 위함).
- **`AfterCommit` 재사용**: node CRUD(`node.go`)가 이미 쓰던 "트랜잭션 커밋 성공 후에만 훅 실행" 장치(`tx.go`)를 그대로 가져옴 — 롤백되면 소켓 발행도 안 나간다.

## 구현 3 — `protocol.MsgProcessUpdate` 신설 (`messages.go`)

```go
const MsgProcessUpdate MsgType = "PROCESS:UPDATE" // sup→browser EVENT: 전체 process 구조체
```
`PROCESS:<uid>` 토픽 위에서 기존 `MsgData`(출력 바이트)/`MsgStatus`(worker 원본 상태전이 중계)와 별개 채널 — "process 레코드 자체가 바뀌었다"를 알릴 때만 쓰며, payload는 항상 전체 `processResponse`(node 도메인 `NODE:CREATE/UPDATE/DELETE`와 동일 원칙, `REF-realtime.md` 참조). `process.go`에 `publishProcess(rec)` 헬퍼 신설(`node.go`의 `publishNode`와 대칭). 첫 사용처는 resize지만, 이후 다른 process 필드 변경도 같은 kind로 통일해서 얹을 수 있다.

## `ProcessDialog` — 실행 중 process 표시 다이얼로그 (2026-07-22, xterm 연동)
> `feature/process/store/processDialog.store.ts` + `feature/process/component/ProcessDialog.vue`. frontend 전반 규약은 `REF-frontend.md`(상태관리 방침·socket hook 등).

- **스토어 패턴**: `appDialog.store.ts`와 동일한 provide/inject(`provideProcessDialog`/`useProcessDialog`). 현재 다이얼로그가 보여주는 process 1개(`ProcessResponse|null`)만 관리 — 동시에 여러 개 띄우는 건 범위 밖. `open = computed(!!process)`, `openProcessDialog`/`closeProcessDialog`, `patchStatus(uid, status)`(소켓 STATUS 반영용 진입점만 제공 — 실제 구독은 컴포넌트 쪽 책임).
- **xterm 마운트**: `open`이 true로 바뀌면 `nextTick` 후 `Terminal`+`@xterm/addon-fit`로 생성/`open()`/`fit()` → `fit()` 직후 확정되는 `rows`/`cols`를 로컬 ref에 보관, `watch([process, cols, rows], ...)`가 `resizeProcess(uid, {rows, cols})`를 호출(위 "구현 2"가 worker 확인 후 DB/memory 동기화).
- **`DATA` 이벤트 연동**: `protocol.DataEvent`가 와이어에선 `{uid, data}`(`data`는 `[]byte`라 base64 문자열)로 옴 → `ev.uid === process.value?.uid`로 필터링 후 base64를 `Uint8Array`로 디코드해 `terminal.write()`(PTY 출력이 유효 UTF-8 보장이 없어 문자열이 아닌 바이트로 전달).
- **소켓 이벤트 필터링(DATA/STATUS)은 사용자가 직접 구현**(assistant는 스텁도 안 남기고 자리만 비움) — `PROCESS:UPDATE` 리스너도 현재 `console.log` 스텁.
- **uid 전환 시 화면 리셋**: 다이얼로그가 열린 채로 대상 process(uid)만 바뀌면 `watch(open,...)`가 `true→true`라 재발화하지 않아 이전 화면이 남는다 — 별도 `watch(() => process.value?.uid, ...)`로 `terminal.reset()`.
- **Close 버튼**: kill 안 함, 다이얼로그만 닫음(구독은 유지 — 다시 열면 이어서 받음).
- **알려진 한계**: 다이얼로그를 열 때 지금까지의 출력(스크롤백)은 못 불러옴(`REF-process-snapshot.md` ring buffer 미착수). 키 입력(`MsgData` 역방향)도 미배선 — 지금은 표시 전용.
- `App.vue`에 `provideProcessDialog()` + `<process-dialog/>` 전역 마운트(`AppDialog`와 동일 위치).

## 다음

- **입력(키스트로크) 미배선**: `MsgData` 역방향(browser→worker)이 아직 없음 — 지금은 표시 전용(위 "ProcessDialog" 절). 고빈도라 REST 부적합, 소켓 메시지 쪽이 유력하나 미정(`CURRENT.md`와 동일 이월 항목).
- 화면복원(ring buffer SNAPSHOT)은 여전히 별개 미착수 항목 — `REF-process-snapshot.md`.
