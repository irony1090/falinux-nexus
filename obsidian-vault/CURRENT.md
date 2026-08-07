# CURRENT

## 현재 날짜
2026-07-22

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## 🎯 다음 작업: node 실시간 연동 + 카탈로그 UI

전송 토대·hook·node CRUD는 모두 커밋 완료. process 도메인의 구독 배선도 이번 세션에 끝남(아래). 이제 **node 변경 → supervisor commit 후 socket push → 프론트 트리/캔버스 반영**을 잇고 UI를 올린다.

**남은 배선** (설계 → `REF-realtime.md` / node 백엔드 → `REF-node-label.md`)
- ~~Kind(MsgType) 어휘 확정~~ → **node 도메인은 확정 완료**(2026-07-16): `NODE:CREATE`/`NODE:UPDATE`/`NODE:DELETE` 3종, move/rename은 UPDATE 흡수, payload=항상 전체 구조체. process는 기존 `MsgData`/`MsgStatus` 유지. → `REF-realtime.md` "Kind 어휘" 절.
- ~~node CRUD 발행처 배선~~ → **완료(2026-07-16)**: `createNode`/`patchNode`/`deleteNode`가 커밋 성공 후에만 `subscribeHub.Publish`하도록 `tx.go`에 범용 `AfterCommit(c, fn)` 훅 신설. 이동(parentId 변경)은 old+new 부모 토픽 양쪽에 `NODE:UPDATE` 발행. → `REF-realtime.md` "supervisor 측 구현" 절 / `REF-supervisor-web.md` "트랜잭션 미들웨어" 절.
- `NODE:<parentId>` 동적 구독/해지 — **아직 미정**(process는 REST로 확정했지만 NODE도 같은 길로 갈지 별도 결정 필요). 현재도 `subscribe.go`는 `NODE:0` 고정구독뿐 → 발행은 되지만 펼친 다른 폴더는 프론트가 구독할 방법이 아직 없음.
- 프론트 `on('NODE:CREATE'|…)` 실제 핸들러 → 트리/터미널 갱신. 컴포넌트 밖 반영 필요 시 `reactive` 모듈 패턴(→ `REF-frontend.md`).
- ~~REST 클라이언트 함수 없음~~ → **node/process 둘 다 완료(2026-07-21, process는 2026-07-22 resize 추가)**: `feature/node/api/node.api.ts`(createNode/listChildren/getNode/patchNode/deleteNode + vue-query `useListChildren`/`useGetNode`/`useNodeQueryClient`), `feature/process/api/process.api.ts`(listSubscriptions/subscribeProcess/unsubscribeProcess/execProcess/killProcess/resizeProcess). node는 아직 호출하는 UI 없음. **process는 `ProcessDialog`가 exec/kill/resize를 실제로 호출하는 첫 UI**(→ `REF-process-resize.md` "ProcessDialog" 절).
- **UI**: node 카탈로그 = 트리 + 캔버스(% 절대배치) — **컨셉 설계 완료**(2026-08-07, 구현은 미착수): 배치도(캔버스)=홈+트리=보조내비(breadcrumb), PC=캔버스+모서리 리스트/트리 오버레이 토글, 모바일=`[지도|리스트]` 세그먼트 토글, 트리 재배치 드래그 지원(device_key 상속변경 확인 UI 필요 + 이동 노드 좌표 NULL 처리) → `REF-node-ui.md`. **미정**: 모바일 트리 화면 형태. 터미널(xterm.js)은 **`ProcessDialog`로 착수**: DATA 출력 연동 완료, 화면복원(스크롤백)·키 입력은 아직.

---

## 진행 중 / 잔여

### Node 모듈 — 남은 단계 (DB+CRUD+핸들러+PatchNode 커밋 `9c9d22e`, HTTP e2e 미검증)
> 설계 `REF-node-label.md` / 이력 `history/node-label.md`
4. **worker_instances roster**: `00004_worker_instances.sql`(PK main_key,sub_key + last_seen) → register에 DB upsert → `ListInstances(main_key)` → 활성/비활성 = roster ∩/− 레지스트리. ※subkey 위조검증 보류
5. **HTTP e2e**: 가입→세션→node CRUD 왕복 스모크(서버 띄우면 00003 실DB 적용)
6. **label 모듈**: `00005_labels.sql`(labels 자기참조 + node_labels M:N) → query → router
- 핸들러 책임 미적용: parentId owner 일치 검증 / 자기 자손으로 Move 사이클 방지

### process 도메인 배선 (supervisor 측 완료 → worker 실행부·router 제어 남음)
> 설계·결정 상세 `REF-process-wiring.md` / 이력 `history/process-wiring.md`. frontend 트리거·상태동기화 버그수정은 `REF-process-trigger.md` / `history/process-trigger.md`. 재접속 모델은 `REF-process-reconnect.md` / `history/process-reconnect.md`. 세션→uid 원장·REST 구독 배선은 `REF-process-subscription.md` / `history/process-subscription.md`.

**완료 상태**: supervisor 배선(manager/entry/bind/router, 경로 계약 `{WORKER_BASE}/<node.ID>/<proc.Uid>`) + worker 실행부 본체(procs/exec/pump/teardown/input·resize·kill, Cwd 배선) + worker 끊김→PENDING→재접속 재바인딩(`WorkerState`) — 전부 구현·build/vet 통과·e2e 검증 완료(2026-07-01~07-14). 상세 → `history/process-wiring.md`, `history/process-reconnect.md`.

**process 상태 관리 버그 3건 수정 완료(2026-07-22, kill 실사용 테스트로 검증)**: 사용자가 frontend에서 exec→kill을 직접 테스트하며 연쇄로 발견.
1. 정상 실행 직후 PENDING 오보고로 process가 즉시 삭제되던 버그 — worker가 exec 시작 시 항상 보고하는 정상 PENDING과 끊김 시 supervisor 합성 PENDING이 `applyStatus`의 같은 분기를 타던 것. "현재 상태와 같으면 무시" 가드로 해결(→ `REF-process-reconnect.md` "PENDING 오삭제 버그").
2. memory `entry.Record`가 생성 시점에 박제(Status/Pid/ExitCode 등 DB 갱신과 안 맞음)돼 있던 문제 — `Mark*` 쿼리의 `RETURNING` row로 통째 교체(`ProcessEntry.SetRecord`)로 해결. 위 가드가 정확히 동작하기 위한 전제조건이기도 했음(→ `REF-process-trigger.md` "entry.Record memory 동기화").
3. kill/비정상 종료 시 `pty.Interactive.Status()`가 큐-종료 신호 대신 프로세스 종료 에러를 반환해 마지막 상태 이벤트가 supervisor `applyStatus`에 아예 안 가던 근본 버그 — 계약대로 큐 에러를 반환하도록 수정(부수로 kill exit code도 유닉스 관례 128+시그널로 정정). `REF-process.md`에 `Status()` 계약 명문화(→ `REF-process-trigger.md` "kill 종료 이벤트 유실 버그 수정").

**다음 배선 (우선순위)**
1. ~~PROC 동적 구독~~ → **완료(2026-07-16)**: `GET /processes/subscriptions` + `POST/DELETE /processes/subscribe/:processId`(REST, 소켓 메시지 아님) + `browsers`(conn→sid) registry로 이미 열려있는 소켓도 즉시 라우팅 반영. → `REF-process-subscription.md` "REST 구독/해지 배선" 절.
   ~~frontend 트리거(실행/종료)~~ → **완료(2026-07-16, 2)**: `POST /processes/exec` / `POST /processes/kill/:processId`(REST, 소켓 `Handle` 아님 — 위 구독 결정과 일관). 실행·종료 둘 다 별도 구독 요청 없이 요청 세션이 자동 구독됨(`subscribeSid` 공유 헬퍼 — 수동 구독과 동일 경로). `router.Exec` 시그니처가 `(uid string, error)`로 바뀜. **종료 후 Hub 구독 정리**(`startRelay`/`cleanupProcessTopic`, relay가 마지막 이벤트까지 다 흘려보낸 뒤에만 해제 — race 없음)도 이번에 같이 닫음. → `REF-process-trigger.md` "frontend 트리거(exec/kill)" 절.
   ~~resize~~ → **완료(2026-07-22)**: `POST /processes/resize/:processId` — `entry.Inter.Layout`이 worker 응답을 실제로 기다려(`syscall.Errno`→`error` 계약 정정) 성공했을 때만 DB(`UpdateProcessLayout` RETURNING)+memory(`entry.SetRecord`) 동기화 후 `PROCESS:UPDATE`(`MsgProcessUpdate` 신설) 발행. `ProcessDialog`가 xterm `fit()` 결과로 자동 호출. → `REF-process-resize.md`.
   **남은 건 input뿐**: `input(MsgData)`→`Inter.Write`(Ctrl+C 등 키입력, `MsgData` 역방향). 고빈도라 REST 왕복은 부적합 — 소켓 메시지 쪽이 유력하나 미정.
2. **화면복원**: bind에 ring buffer(SNAPSHOT) 상시 적재 + 재접속 SNAPSHOT 전송 — **미착수(설계 논의만 완료)**. supervisor-side ring buffer로 방향 확정(스케일·htop 케이스 검토 끝), worker-side 이전 옵션은 snapshot↔live 이음매 race(유실/중복) 미해결로 보류. → `REF-process-snapshot.md`.
   - ~~세션→uid 원장~~ → **구현 완료(2026-07-16)**: 마이그레이션·쿼리·CREATE/DELETE 호출 지점(위 1번) 전부 끝남. 남은 건 ring buffer SNAPSHOT 자체와 **프론트가 이 엔드포인트들을 실제로 부르는 것**(REST 클라이언트 함수·UI 미착수).
3. EXEC content→실행 세부정책(직접실행 vs `sh -c`).

**결정 필요**: 끊긴 창 입력/kill 거절 vs 큐잉 / 공유 kill 인가 / kill 에스컬레이션.
**정리 잔여(구)**: `register.go` 주석 SendBuffer 테스트. worker `baseDir` 필드·`instanceKey()`가 resolveDest 재설계로 dead code화(정리 여부 판단).

### 프론트 user/login — WIP (커밋 `3a8e92e` 동반, → `REF-frontend.md`)
- `feature/user`(`auth.store.ts` + `api/user.api.ts`), `pages/Login.vue` + `/login` 라우트, `common/api`(api/query util), `feature/layout` 전역 다이얼로그(`AppDialog.vue` + `appDialog.store`). 실서버 연동·가드 마무리 남음.

---

## 미해결 이슈 (이월)
- ~~**PROC 토픽 무구독**~~ → **완전히 해소(2026-07-22)**: 백엔드(2026-07-16)·REST 클라이언트(2026-07-21)에 이어 `ProcessDialog`가 exec 성공 시 자동 구독된 process의 `DATA`를 실제로 xterm에 그려 보여준다 — "백엔드 무배선"→"프론트 함수 없음"→"UI 미착수"로 이어지던 원인 이동이 끝남. 남은 건 `PROCESS:UPDATE`/`STATUS` 리스너가 아직 `console.log` 스텁이라는 것뿐(기능적으로는 `patchStatus` 진입점이 이미 있어 연결만 하면 됨).
- **파일 전송**: 구현 완료 / e2e 미검증. 잔여: e2e 스모크 / abort sentinel (register 임시전송은 주석처리됨)
- **서브키 충돌/위조**: key↔subkey 결속 검증 미구현(node roster에서 닫을지 보류)
- **supervisor 영속성**: registry 메모리 → PG 미착수

## 잔여 (틈날 때)
- `SESSION_KEY` 등 env화(현재 `"irony"` 하드코딩)
- checkSession createdAt=0(pgtype.Timestamptz gob 미직렬화) → sess.Data.ID로 DB 재조회
- ~~`getSessionKey` nil pointer panic~~ → 해소 확인(작업트리에 가드 적용돼 있음, `REF-process-subscription.md` "세션→uid 원장" 참조)
