# history: 리팩터링 — supervisor 소켓 알림 규약 정리

> → REF `REF-refactor-socket.md` / → 현재 `CURRENT.md`

## 2026-10-21 R1(목록 조사) — 소켓 알림 범위
- 사용자 요청: supervisor 소켓 알림의 계정별·탭(기기)별 규약이 여러 번 고쳐져 다듬을 필요가 있는지
- 읽은 곳: `internal/subscribe/hub.go`, router `subscribe.go`·`tabs.go`·`tile.go`·`tileNotice.go`·`node.go`·`active.go`·`workspace.go`·`shareAccess.go`·`process.go`·`processApi.go`·`sizeOwner.go`·`presence.go`·`roster.go`, `protocol/messages.go`, 프론트 `on('...')` 구독처
- 결과: 후보 9개 → `REF-refactor-socket.md` — 소켓-1(계정 토픽 이름)·소켓-2(여러 계정에 보내기 중복)·소켓-3(볼 수 있는 계정 판정 중복)·소켓-4(탭 하나로 보내기)·소켓-5(커밋 뒤 발행 규약 섞임)·소켓-6(process 구독 지점 흩어짐)·소켓-7(빈 패키지·낡은 주석)·소켓-8(kind 이름 접두어 제각각)·소켓-9(worker·presence push 없음). 코드 변경 없음. R2(범위 결정) 대기

## 2026-10-21(2) R2(범위 결정)·R3(묶음별 정리) 1차
- 사용자 결정: 목적 = 사람이 읽기 쉽게(성능은 떨어지지 않는 선까지), 주석은 손대는 파일만 한 줄 기준(가리키는 다른 파일은 연관 주석만), 1차 범위 = 소켓-1(계정 토픽 이름)·소켓-5(커밋 뒤 발행 규약 섞임)·소켓-6(process 구독 지점 흩어짐), 코드는 Claude가, 이름만 바뀌는 파일도 주석 정리
- 소켓-4(탭 하나로 보내기)는 실제 대상이 `emitSizeOwner` 1곳이라 헬퍼 세트에 합치기로
- 한 것·결과 → `REF-refactor-socket.md` "1차 범위 결과". 소켓-5는 코드가 이미 이름 규칙대로였고 규칙이 한 곳에 안 적힌 게 문제였음(예외 `afterShareChange` 하나만 이름 바꿈)
- 확인 중 발견(버그 아님): 서버는 `TAB:ID`를 보낸 뒤 process 토픽을 구독한다 -> 확인 스크립트가 `TAB:ID` 직후 입력을 보내면 출력을 못 받음(→ `REF-test-env.md` 함정). 이미 소유자인 탭이 resize하면 `PROCESS:SIZE_OWNER`가 안 옴(정상)

## 2026-10-21(3) PC 꺼짐 뒤 재개 · 브랜치 · 1차 커밋
- 재시작 뒤 확인: 작업 트리 그대로, build·vet 통과, 임시 DB·worktree 정리돼 있음. scratchpad의 코드만 patch는 사라짐
- 사용자 질문 "기능 차이 없나" -> 주석 뺀 diff 재검토: 이름 2개 + 함수 3개 그대로 이동뿐
- 사용자 결정: 파트(리팩터링)는 브랜치 `refactor/socket`, 마감 때 머지. 관계없는 `websocket.hook.ts` WireData 타입 수정은 master `03bc9a6`에 먼저
- 1차를 브랜치에 한 커밋(코드 + 주석 + vault)
