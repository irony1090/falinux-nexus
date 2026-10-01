# REF — node 타일 터미널 임베드 + 크기 우선권

> 2026-10-01 신설 — 타일 UI 6(`ProcessDialog` 타일 임베드) + 연동 1차 5(실행)의 구조·결정. 이력 → `history/node-ui-terminal.md` / 동기화 범위·탭 id·공유(O~V) → `REF-process-sync.md` / 연동 1차 1~4 → `REF-node-ui-link.md` / 렌더 층 → `REF-node-ui-render.md` / process 백엔드 → `REF-process-trigger.md`·`REF-process-resize.md` / 현재 진행 → `CURRENT.md`.

**상태: 구현 완료(2026-10-01, go build·vet + vue-tsc 통과), 실제 실행 확인 전·미커밋. 크기 우선권 부분은 개정 예정(→ `REF-process-sync.md`).**

## 결정 (번호는 `REF-node-ui-link.md`의 A~I에 이어서)
| 결정 | 내용 | 이유 |
|---|---|---|
| J(xterm 소유자) = 스토어 | uid마다 `Terminal` 객체를 스토어가 보유. 타일은 마운트 시 `terminal.element`를 자기 DOM에 붙이고 언마운트 시 뗀다. 객체 폐기 = 타일 닫을 때 | 아래 "재발 방지" 참조. 마운트 전에 온 `DATA`도 `write`해 둘 수 있어 exec 응답보다 먼저 온 출력을 잃지 않음 |
| K(상태 출처) = 같은 스토어 | `STATUS`·`PROCESS:UPDATE`를 **한 번만** 구독, uid별 상태 보관 | `dummyStatus`/`dummyKill` 대체. kill = `killProcess`, RUN 표시·닫기 비활성도 실제 상태 |
| L(실행 흐름) | `listWorkers(nodeId)` → 0개 = 오류 문구 / 1개 = 바로 / 여러 개 = 인스턴스 선택 창 → `execProcess` → 스토어 등록 → `addTerminal` → `reveal` | 연동 1차 5(실행) |
| M(크기 우선권) = 실행한 세션 | 아래 "크기 우선권" 절 (사용자 지정 2026-10-01) | 여러 세션이 같은 process를 볼 때 PTY 크기 충돌 방지 |
| N(기존 `ProcessDialog`) = 삭제 | `App.vue` 마운트·`processDialog.store`·`ProcessDialog.vue` 삭제 + `index.vue`의 미사용 테스트 함수(`onExec`/`onKill`/`onSubscribe` 등, 하드코딩 `authKey`) 정리 | `ProcessDialog` = 2026-07-22의 단일 창 xterm 시제품. 로직이 J의 스토어·`ProcessTerminal`로 옮겨짐 |

**재발 방지 — 타일은 Grid를 옮기면 재마운트된다**: `TileStrip.vue`는 Grid마다 `v-for` 목록이 따로라, 크기 변경·창 크기 변경으로 타일이 다른 Grid로 가면 컴포넌트가 새로 생성된다. xterm을 컴포넌트가 가지면 그때마다 화면(스크롤백)이 사라진다(화면복원 ring buffer 미구현). 전체보기 전환의 재마운트 금지(`REF-node-ui-render.md`)와 같은 문제의 다른 경로.

## 크기 우선권 (M)
| 상황 | 동작 |
|---|---|
| exec 직후 | 실행한 세션(sid) = 크기 소유자. supervisor 메모리(`ProcessEntry`)에 기록 |
| 소유자의 마지막 소켓이 끊김 | 타이머 **60초** 시작. 그 안에 재접속하면 취소 |
| 60초 초과 / 로그아웃 | 구독 순서(`entry.subscribers` = 구독한 순서)에서 다음 세션 중 **소켓이 연결된** 세션으로 넘김. 없으면 다음에 연결되는 구독자가 받음 |
| 원래 소유자가 나중에 돌아옴 | **되찾지 않음** — 일반 구독자 (사용자 선택, 크기가 갑자기 되돌아가지 않게) |
| 비소유자의 `resize` 요청 | 서버가 403 거절 |
| 비소유자 화면 | xterm을 소유자의 rows/cols로 고정, 타일보다 크면 타일 안 스크롤 (사용자 선택) |
| 소유자 변경 | 새 소유자 소켓에만 알림 → 새 소유자가 즉시 자기 타일 크기로 `resize` |
| 소유자 화면 | 타일 본문 `ResizeObserver` → `fit` → rows/cols가 바뀐 때만 `resizeProcess`(디바운스 ~200ms). 전체보기는 `transform` 축소라 영향 없음 |

- **보안 — sid를 다른 세션에 보내지 않는다**: `process_subscribers.sid` = 세션 쿠키 원본값. 소유자 여부는 모두에게 같은 `PROCESS:UPDATE`에 싣지 않고, REST 응답(exec·구독 목록)에 요청 세션 기준 `sizeOwner: bool`, 변경 시엔 해당 세션 소켓에만 개별 이벤트.
- **로그아웃 = 기존 `DELETE /users/session`(`signOut`) 확장**(새 API 아님 — 구현 중 이미 있음을 확인. 프론트 `auth.store` `logout`도 이걸 부름): 로그인 상태였으면 `leaveSession(sid)` = 그 sid 구독 행 전부 해지(추천대로, sid는 로그인마다 새로 생겨 다시 쓰이지 않음) + Hub 구독 해제 → 소유권 즉시 넘김.
- 소유자 기록은 메모리만. supervisor 재시작 후엔 비어 있고, 다음에 소켓이 붙거나 resize하는 구독자가 가져감.

## 개정 (2026-10-01) → `REF-process-sync.md`
M(크기 우선권)의 단위가 sid → **탭 id**로 바뀌고(Q), 구독은 계정 단위로(O·S) 개정됨. 아래 "구현"의 sid 단위 크기 우선권 코드는 1단계에서 탭 단위로 대체됨(2026-10-01, 터미널 임베드 J·K·L·N은 유지). 결정 O~V·진행 순서 → `REF-process-sync.md`, 1단계 구조·구현·확인 → `REF-process-sync-impl.md`.

## 구현 (2026-10-01)
| 층 | 파일 | 내용 |
|---|---|---|
| 백엔드 | `router/sizeOwner.go` (신규) | `sizeOwners`(uid→sid 맵 + 넘김 타이머) + `setSizeOwner`/`claimSizeOwner`/`releaseSizeOwner`/`sizeOwnerOnConnect`/`OnDisconnect`/`OnSignOut`/`handoffSizeOwner` |
| 백엔드 | `subscribe.go` | 소켓 연결 = 살아 있는 구독 uid로 `OnConnect`, 끊김 = `browsers` 제거 뒤 `OnDisconnect` |
| 백엔드 | `process.go` | `Exec` 성공 직후 `setSizeOwner`(알림 없음 — REST 응답으로 앎) / `applyStatus` 종료 분기에서 `releaseSizeOwner` |
| 백엔드 | `processApi.go`·`processDto.go`·`user.go` | resize 403 + `sizeOwner` 응답(`withSizeOwner`) + `leaveSession` + `signOut` |
| 백엔드 | `query/processSubscribers.sql` | `ListSubscriberSidsByProcess`(created_at 순) → `ProcessManager.SubscriberSids` |
| 백엔드 | `protocol/messages.go` | `MsgProcessSizeOwner = "PROCESS:SIZE_OWNER"` + `SizeOwnerEvent{uid}` |
| 프론트 | `process/store/processTerm.store.ts` (신규) | J·K: uid → 상태·소유 여부·xterm(`Map`, 비반응형). `exec`/`kill`/`syncSize`/`attach`/`detach`/`dispose` |
| 프론트 | `process/component/ProcessTerminal.vue` (신규) | attach/detach + `ResizeObserver`(150ms 묶음) → 소유자면 `proposeDimensions` → `syncSize`, 아니면 소유자 rows/cols로 `term.resize` + 스크롤 |
| 프론트 | `worker/component/WorkerPickDialog.vue` (신규) | 인스턴스 고르기 |
| 프론트 | `TerminalTileBody`·`TileFrame`·`FolderTileBody` | 터미널 임베드 / 상태 칩(WAIT·RUN·DONE·FAIL)·실제 kill·닫을 때 `dispose` / `onExec` = L |
| 프론트 | `pages/index.vue`·`App.vue` | `provideProcessTerm`을 타일 트리와 같은 자리(index)에. `ProcessDialog`·`processDialog.store`·`tileDummy.ts` 삭제 |

### 구현 메모
- **재발 방지 — 소유자 맵은 `ProcessEntry` 밖**: worker 재접속 시 `Rebind`가 entry를 새로 만들어 entry 안 필드(구독자 목록 포함)가 비워진다. 그래서 소유자는 router의 별도 맵, "다음 세션" 순서는 DB(`created_at`)에서 읽는다. 같은 이유로 `releaseSizeOwner`는 `cleanupProcessTopic`(PENDING 때도 불림)이 아니라 `applyStatus` 종료 분기에서.
- **재발 방지 — 넘김과 재접속의 경합**: connect는 `browsers` 등록 → `o.mu` 순서라, `handoffSizeOwner`가 잠금 안에서 떠나는 sid의 소켓을 보면 재접속이 이긴 것으로 보고 멈춘다(로그아웃 = `force`라 무시). 소유자가 이미 바뀌었으면 타이머를 건드리지 않는다(새 소유자 타이머일 수 있음).
- **exec 응답보다 먼저 오는 이벤트**: 요청 세션은 relay 기동 전에 구독되므로 `DATA`·`STATUS`가 응답보다 먼저 올 수 있다 → `execInFlight > 0`일 때만 모르는 uid의 출력(전부)·상태(최신 1개)를 모아 `register`에서 적용. 요청이 없을 때 모르는 uid는 버린다(새로고침 전부터 구독 중이던 process — ⑪ 전까지 타일 없음).
- 소켓 `STATUS.status`는 `execute.CommandStatus` **숫자**(0 PENDING / 1 PROCESS / 2 COMPLETED / 3 FAILED). `PROCESS:UPDATE`는 문자열 DTO.
- 소켓이 다시 붙으면(`CONNECTED`) `listSubscriptions`로 소유 여부를 서버 기준으로 다시 맞춤. resize 실패도 같은 동기화(`throwCatch`가 상태 코드를 버려 403 구분 불가).
- xterm `open()`은 한 번만 — 이후 attach는 `term.element`를 옮겨 붙임. 색은 테마 `terminal`/`on-terminal`(Vuetify 4 색 타입이 유니언이라 문자열일 때만).
- **기존 공백(이번 범위 밖)**: worker 재접속 뒤 `Rebind` entry의 메모리 구독자 목록이 비고, PENDING 때 `cleanupProcessTopic`이 Hub 구독도 지워서 브라우저가 소켓을 다시 붙이기 전까지 재바인딩된 process 출력을 못 받는 것으로 보임(코드 읽기로만 확인).

## 파일 배치 (계획 당시)
| 파일 | 책임 |
|---|---|
| `feature/process/store/processTerm.store.ts` (신규) | J + K: uid → `Terminal`·상태·`sizeOwner`, 소켓 이벤트 수신 |
| `feature/process/component/ProcessTerminal.vue` (신규) | DOM 붙이기·떼기 + 소유자면 fit/resize, 아니면 고정 크기 |
| `feature/worker/component/WorkerPickDialog.vue` (신규) | L 인스턴스 선택 |
| `TerminalTileBody` / `TileFrame` / `FolderTileBody` | `ProcessTerminal` 사용 / 실제 상태·kill / `onExec` 실제 실행 |
| 백엔드 `ProcessEntry`·`subscribe.go`·`processApi.go`·user 라우터 | 소유자 필드·끊김 타이머·403·`sizeOwner` 응답·logout (실제로는 `ProcessEntry` 대신 router 맵 — 위 구현 메모) |
| `tileDummy.ts` | 삭제 |

## 범위 밖
키입력(`disableStdin` 유지, input 배선 때) / 새로고침 뒤 이어 보기(⑪ 타일 트리 서버 저장 + ring buffer) / 공유 kill 인가.
