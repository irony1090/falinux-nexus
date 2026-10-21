# history: 리팩터링 — 소켓 밖 영역 정리 후보

> → REF `REF-refactor-cleanup.md` / → 현재 `CURRENT.md`

## 2026-10-21 R1(목록 조사) 2회차 — worker·internal·프론트·REST
- 사용자 요청: 소켓 2차 뒤 "조금 더 리팩터링할 부분 파악"(process 상태 전이 흩어짐 조사는 불필요 — 소켓-6(process 구독 지점 흩어짐) 정리로 충분)
- 방법·결과 → `REF-refactor-cleanup.md`: 후보 10개 정돈-1(PortBridge 실행 잔재)·정돈-2(옛 WS 클라이언트)·정돈-3(worker 저장소 트랜잭션)·정돈-4(작은 잔재)·정돈-5(실행 인터페이스 안 쓰는 메서드)·정돈-6(tile.go 세 주제)·정돈-7(주석 "표시줄")·정돈-8(TileFrame.vue 섞인 관심사)·정돈-9(프론트 공용 유틸 미사용)·정돈-10(파일 전송 송신부). 코드 변경 없음, R2(범위 결정) 대기
