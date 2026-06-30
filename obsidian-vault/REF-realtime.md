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
  - `nodeSubscribeKey(parentId int64) = "NODE:%d"`.
  - 현재 `NODE:0` 고정 구독 + 검증용 `Handle("TEST")`/`On("TEST_ON")`/`Emit("TTTT")` (스모크, 추후 제거).

## 주의점 (배선 시 반드시)

1. **commit 이후 Publish** — node CRUD는 txMiddleware 안. tx 안에서 Publish하면 롤백돼도 브로드캐스트 누설 → commit 성공 후 flush.
2. **자기 echo** — 요청자도 토픽 구독 중이면 자기 변경 되받음. 이벤트를 idempotent하게 만들어 재적용(초기), 정밀하게는 origin/req-id 태깅 dedupe.
3. **공유 철회 = 강제 unsubscribe** — presence가 권한보다 오래 살면 안 됨. browser 레지스트리(userID→conns)로 찾아 `Unsubscribe`.
4. **binaryType** — Go BinaryMessage라 프론트 `ws.binaryType='arraybuffer'` 필수.
5. **다중 탭/다중 유저는 공짜** — Hub가 연결(client) 단위 구독이므로 자연 처리.

## 현재 상태 / 다음

- ✅ **전송 토대 완성 + 3모드 e2e 검증**(2026-06-30): call(TEST→RES) / emit(TEST_ON) / on(TTTT). → `history/realtime.md`
- ⬜ **동적 구독 어휘**: `MsgSubscribe`/`MsgUnsubscribe` 핸들러 + 그 안에서 **DB 인가** → `NODE:0` 고정 구독 대체.
- ⬜ **발행처 배선**: node/process CRUD 핸들러에서 commit 후 `subscribeHub.Publish(topic, kind, payload)`.
- ⬜ **프론트 수신**: `on('node.created'|'process.output'|…)` 실제 핸들러 → 트리/터미널 갱신.
- ⬜ Kind(MsgType) 어휘 확정: `node.created/moved/deleted/renamed`, `process.output/status` 등.
