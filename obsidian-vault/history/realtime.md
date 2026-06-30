# history/realtime — 실시간 push (socket)

> 요약·재사용 지식 → `REF-realtime.md` / 현재 진행 → `CURRENT.md`
> 전송 인프라 → `REF-infra.md` / 프론트 hook → `REF-frontend.md`

## 2026-06-30 — supervisor↔웹 socket 전송 토대 완성 + 3모드 e2e 검증

### 설계 (이번 세션 확정)
- `transport.New`(transport.Conn) 재사용으로 브라우저 WS도 worker와 동일하게 다룸(MessageRW 의존, gorilla 만족). worker용 in-band REGISTER 대신 **HTTP 세션 인증**으로 대체.
- **인가(DB,구독1회) / 라우팅(메모리,Hub)** 분리 — 공유·다중 유저 동일 화면을 presence 구독으로 흡수. 토픽 = **펼친 폴더 단위** `NODE:<parentId>`. (상세 → `REF-realtime.md`)
- "유저 트리 전체 1토픽" vs "펼친 폴더 단위" 비교 후 후자 채택(공유 누설/over-fanout 회피).

### supervisor 구현
- `subscribe/hub.go`: `Hub`에 제네릭 `T` 추가 → `Publish(key, Kind T, payload)`, `send func(C,T,[]byte)`. Kind를 send로 전달(payload 1회 marshal).
- `supervisorRouter.go`: `subscribeHub` 필드(`Hub[*transport.Conn, protocol.MsgType]`), `subscribe.New(json.Marshal, c.Emit(k, json.RawMessage(b)))`, 라우트 `GET /subscribe`.
  - 🔴 교정: send에서 `json.RawMessage`로 감싸지 않으면 `Conn.Emit` 재-marshal이 []byte를 base64로 이중 인코딩.
- `subscribe.go` `handleSubscribeWS`: 🔴 인증 버그 2건 교정 — ① `sess != nil` 부등호 반전(로그인 유저가 막힘) + 기준 오류 → `requireSession` 사용 ② upgrade **후** panic은 hijack된 연결에 401 못 씀 → `requireSession`을 **upgrade 전**으로. 정리 = `Serve→Close→UnsubscribeAll`.

### 프론트 hook 재설계 (`apps/frontend/src/common/websocket/websocket.hook.ts`)
- 옛 `createWebsocketHook`(sendMessage/onMessage) 전면 폐기 → transport.Conn 대응 재작성. (상세 API → `REF-frontend.md`)
- `call`/`emit`/`on`/`status`/`connect`/`disconnect` + 주입형 `Protocol`(frame 가변). 기본 `jsonFrame`=Go Frame 미러.
- `index.vue` 테스트 페이지를 새 API로 마이그레이션(수동 프레임 → `call('TEST',…)`).
- `npm run type-check`(vue-tsc) 통과.

### e2e 검증 (실서버 5050, 로그인 후)
- `call('TEST','TEST_MESSAGE')` → 서버 TEST 핸들러 → `'RES'` resolve ✅ (REQ→RES)
- `emit('TEST_ON','TEST_ON_MESSAGE')` → 서버 `On("TEST_ON")` 수신 로그 ✅ (EVENT ↑)
- 서버 `Emit("TTTT","anyting~")` → 프론트 `on('TTTT')` 수신 `[TTTT] anyting~` ✅ (EVENT ↓)
- → Protocol·correlator·binaryType·watch 수명관리가 양방향 전부 정상.

### 남은 것
- 동적 `MsgSubscribe`/`MsgUnsubscribe` + DB 인가(현재 `NODE:0` 고정 구독·TEST 스모크 대체).
- node/process CRUD commit 후 `Publish` 배선. 프론트 `on` 실제 핸들러.
