# REF: realtime push (socket)

> supervisor ↔ 웹 클라이언트 실시간 이벤트 평면. node/process CRUD·출력을 socket으로 push.
> 상세 이력 → `history/realtime.md` / 현재 진행 → `CURRENT.md`
> 전송 인프라(transport.Conn) → `REF-infra.md` / 프론트 hook 구현 → `REF-frontend.md` / node 백엔드 → `REF-node-label.md`

## 핵심 결정 (불변)

- **인가 ≠ 라우팅 분리** (이 모듈의 근간):
  - **인가 = DB, 구독 시점 1회** (느린 길). "이 유저가 이 node/process 볼 권한 있나(소유 OR 공유)" 검사 후 구독 등록.
  - **라우팅 = 메모리, 이벤트마다** (빠른 길). `subscribe.Hub` topic→구독자 맵 fan-out. **DB 무접촉**.
  - → process 고빈도 출력에서 DB 안 타고, 소유자 무관하게 "보고 있는 모두"에게 전달.
  - **DB ≠ 라우팅 authority** 불변 원칙을 worker→browser 평면으로 확장 (MEMORY의 worker 레지스트리 원칙과 동일 철학).
- **수신자 = presence(구독)이지 소유자가 아님**. 공유 기능·다중 유저 동일 화면이 전부 "구독자 집합" 한 개념으로 흡수됨.
- **토픽 granularity = 펼친 폴더 단위** `NODE:<parentId>`(루트=`NODE:0`). 유저 트리 전체 1토픽 안 씀.
  - 이유: 공유의 권한 단위(폴더)와 토픽이 1:1 → 누설 없음 + fan-out 정밀. 트리 전체 토픽은 공유 시 권한 밖 변경까지 누설되어 결국 쪼개야 함.
  - **이동/삭제 = old parent + new parent 2개 토픽**에 발행(양면).
- **비대칭 역할**: 브라우저 `call` → 서버 `Handle` / 서버 `Emit`(Publish) → 브라우저 `on`. **브라우저에 `Handle`(서버→브라우저 REQ 수신)은 두지 않음**.

## 프로토콜 (3모드, transport.Conn ↔ 프론트 hook 1:1)

| 모드 | Go Conn | 프론트 hook | 용도 |
|------|---------|-------------|------|
| REQ→RES | `Call`/`Handle` | `call` | 브라우저 요청→서버 응답 |
| EVENT ↑ | `Emit`/`On` | `emit` | 브라우저→서버 단방향 |
| EVENT ↓ | `Emit`/`On` | `on` | 서버→브라우저 push (push 본체) |

- 와이어 = Go `internal/protocol.Frame` `{k,id,t,e?,d?}` (Kind: REQ=0/RES=1/EVENT=2), gorilla **BinaryMessage**.
- **frame 형태 가변 전제**: 프론트는 `Protocol` 전략으로 와이어를 추상화 → frame 바뀌면 Protocol 구현 1개만 교체 (→ `REF-frontend.md`).

## supervisor 측 구현

- `internal/subscribe/hub.go`: `Hub[C comparable, T comparable]`. **`Publish(key, Kind T, payload)`** — Kind를 send로 전달(payload는 1회만 marshal). `send func(C, T, []byte)`.
- `cmd/supervisor/router/supervisorRouter.go`:
  - `subscribeHub *subscribe.Hub[*transport.Conn, protocol.MsgType]`
  - `subscribe.New(json.Marshal, func(c, k, b){ return c.Emit(k, json.RawMessage(b)) })`
    — ⚠️ **`json.RawMessage`로 감싸야** `Conn.Emit`의 재-marshal에서 이중 인코딩(base64) 방지.
  - 라우트 `GET /subscribe`.
- `cmd/supervisor/router/subscribe.go` `handleSubscribeWS`:
  - **`requireSession(c)`를 upgrade 전에** 호출(실패 시 panic 401 — 아직 hijack 전이라 정상 렌더). upgrade 후 panic은 hijack된 연결에 응답 못 씀.
  - `transport.New(ws)` → 구독 → `conn.Serve()` → `conn.Close(err)` → `subscribeHub.UnsubscribeAll(conn)`.
  - `nodeSubscribeTopic(parentId int64) = "NODE:%d"`(구 `nodeSubscribeKey`, `processSubscribeKey`와 나란히 있던 걸 2026-07-21 `processTopic()`으로 통합하며 리네임).
  - 현재 `NODE:0` 고정 구독만 남음. 검증용 `Handle("TEST")`/`On("TEST_ON")`/`Emit("TTTT")` 스모크는 **worktree에서 제거됨**(2026-07-13 확인, 커밋 아직 안 됨 — `apps/frontend/src/pages/index.vue` 쪽 대응 코드도 함께 제거된 상태).
- **node CRUD 발행처 배선 — 구현 완료 (2026-07-16)**: `cmd/supervisor/router/node.go`
  - `publishNode(kind, node)` = `node.ParentID`(NULL→0 정규화, `parentOrRoot`)로 토픽을 계산해 `publishNodeTopic`으로 위임. `publishNodeTopic(topic, kind, node)` = `subscribeHub.Publish(topic, kind, newNodeResponse(node))`(에러는 로그만).
  - **커밋 후 flush 보장**: tx 안에서 직접 Publish하면 롤백 시 누설되므로(위 "주의점 1"), `tx.go`에 신설한 **범용 `AfterCommit(c, fn)`** 훅으로 예약 — `txScope.hooks`에 쌓아 두고 `release()`가 **커밋 성공 시에만** 순서대로 실행(커밋 실패·롤백이면 실행 안 됨). 상세 메커니즘 → `REF-supervisor-web.md` "트랜잭션 미들웨어" 절.
  - `createNode`: 새 부모 토픽에 `NODE:CREATE` 1회.
  - `patchNode`: 요청에 `parentId`가 포함된 경우만 PATCH 전 old parent를 미리 조회(이동 여부 판단용) → 새 부모 토픽에 `NODE:UPDATE`, old≠new면 **old 부모 토픽에도 동일 payload로 한 번 더**(이동/이름변경 2토픽 원칙 그대로 적용).
  - `deleteNode`: 삭제 전에 대상 row를 조회해 "삭제 직전 스냅샷"을 확보(DeleteNode는 `:exec`라 반환행 없음) → 그 부모 토픽에 `NODE:DELETE`.
  - **자기 echo**는 아직 처리 안 함(프론트가 idempotent 재적용하기로 한 초기 방침 그대로, 백엔드 dedupe 없음).

## 주의점 (배선 시 반드시)

1. **commit 이후 Publish** — node CRUD는 txMiddleware 안. tx 안에서 Publish하면 롤백돼도 브로드캐스트 누설 → commit 성공 후 flush.
2. **자기 echo** — 요청자도 토픽 구독 중이면 자기 변경 되받음. 이벤트를 idempotent하게 만들어 재적용(초기), 정밀하게는 origin/req-id 태깅 dedupe.
3. **공유 철회 = 강제 unsubscribe** — presence가 권한보다 오래 살면 안 됨. browser 레지스트리(userID→conns)로 찾아 `Unsubscribe`.
4. **binaryType** — Go BinaryMessage라 프론트 `ws.binaryType='arraybuffer'` 필수.
5. **다중 탭/다중 유저는 공짜** — Hub가 연결(client) 단위 구독이므로 자연 처리.

## Kind(MsgType) 어휘 — node 도메인 확정 (2026-07-16)

- **`NODE:CREATE` / `NODE:UPDATE` / `NODE:DELETE` 3종으로 확정.** 세분화 kind(`node.created/moved/deleted/renamed` 등) 대신 **CRUD 3종 + payload=항상 전체 node 구조체**로 통일(초안이던 `node.change{op,node}` 단일봉투안은 기각 — 개별 kind 방식 채택).
- **이동(move)/이름변경(rename) = 별도 kind 없이 `NODE:UPDATE`에 흡수** — payload의 `parent_id`/`name` 필드 변화로 프론트가 판별. 기존 **"이동/삭제=old parent+new parent 2토픽 발행"** 원칙은 그대로 유지(라우팅은 topic이 담당, kind와 무관 — 서버가 mutation 직전 old `parent_id` 캡처해 두 토픽에 동일 payload 발행).
- **`NODE:DELETE`도 전체 구조체 payload**(id만 보내지 않음) — C/U/D 포맷 통일, 프론트 파싱 분기 단순화.
- 프론트 사용 형태: `socket.on<Node>('NODE:CREATE', handler)` 식으로 kind별 개별 등록(`websocket.hook.ts`의 `on(type, handler)` 그대로 재사용, 제네릭 payload=Node 구조체).
- **미정**: device presence(worker main#sub online/offline) kind 스타일 — 같은 개별-kind 패턴(`DEVICE:ONLINE`/`DEVICE:OFFLINE`)으로 갈지 별도 결정 필요.

## 현재 상태 / 다음

- ✅ **전송 토대 완성 + 3모드 e2e 검증**(2026-06-30, 커밋 3a8e92e + hook 견고성 e28252b): call(TEST→RES) / emit(TEST_ON) / on(TTTT). → `history/realtime.md`
- ✅ **node 도메인 Kind 어휘 확정**(2026-07-16, 위 절) — 구현은 미착수.
- ✅ **process 도메인 동적 구독/해지**(2026-07-16) — 소켓 메시지(`MsgSubscribe`/`MsgUnsubscribe`) 대신 **REST**로 확정·구현 완료. `browsers`(conn→sid) registry 포함. 상세 → `REF-process-subscription.md` "REST 구독/해지 배선" 절.
- ✅ **node CRUD 발행처 배선**(2026-07-16) — 위 절. `AfterCommit` 훅으로 커밋 후에만 Publish, 이동은 old+new 2토픽.
- ⬜ **NODE:<parentId> 동적 구독 어휘**: 위 process 쪽과 달리 아직 미정 — REST로 갈지(위 결정과 일관성) 소켓 메시지로 갈지부터 정해야 함. DB 인가 포함, `NODE:0` 고정 구독 대체. **발행은 이미 되지만** 프론트가 펼친 폴더 토픽을 구독할 방법이 없어 `NODE:0` 밖 변경은 아직 안 닿음.
- ⬜ **프론트 수신**: `on('NODE:CREATE'|'NODE:UPDATE'|'NODE:DELETE'|'process.output'|…)` 실제 핸들러 → 트리/캔버스 갱신. REST 클라이언트 함수(`GET/POST/DELETE /processes/subscribe*`)도 프론트에 아직 없음.
- ⬜ device presence kind 결정(`DEVICE:ONLINE`/`OFFLINE` 후보, 미확정).
- ✅ **process 도메인에도 CRUD류 kind 합류**(2026-07-22): 기존 `MsgData`/`MsgStatus`(저수준 스트림)에 더해 `MsgProcessUpdate`("PROCESS:UPDATE", 전체 process 구조체 payload — node의 CRUD kind와 동일 원칙)를 `PROCESS:<uid>` 토픽 위에 추가. 첫 사용처는 resize. → `REF-process-resize.md`.
