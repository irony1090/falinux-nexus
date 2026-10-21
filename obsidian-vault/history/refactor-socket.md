# history: 리팩터링 — supervisor 소켓 알림 규약 정리

> → REF `REF-refactor-socket.md` / → 묶음 이력 `history/refactor-socket-batches.md` / → 현재 `CURRENT.md`

## 2026-10-21 R1(목록 조사) — 소켓 알림 범위
- 사용자 요청: supervisor 소켓 알림의 계정별·탭(기기)별 규약이 여러 번 고쳐져 다듬을 필요가 있는지
- 읽은 곳: `internal/subscribe/hub.go`, router `subscribe.go`·`tabs.go`·`tile.go`·`tileNotice.go`·`node.go`·`active.go`·`workspace.go`·`shareAccess.go`·`process.go`·`processApi.go`·`sizeOwner.go`·`presence.go`·`roster.go`, `protocol/messages.go`, 프론트 `on('...')` 구독처
- 결과: 후보 9개 → `REF-refactor-socket.md` — 소켓-1(계정 토픽 이름)·소켓-2(여러 계정에 보내기 중복)·소켓-3(볼 수 있는 계정 판정 중복)·소켓-4(탭 하나로 보내기)·소켓-5(커밋 뒤 발행 규약 섞임)·소켓-6(process 구독 지점 흩어짐)·소켓-7(빈 패키지·낡은 주석)·소켓-8(kind 이름 접두어 제각각)·소켓-9(worker·presence push 없음). 코드 변경 없음. R2(범위 결정) 대기
