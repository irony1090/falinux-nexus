# history: 리팩터링 — 소켓 밖 영역 정리 후보

> → REF `REF-refactor-cleanup.md` / → 현재 `CURRENT.md`

## 2026-10-21 R1(목록 조사) 2회차 — worker·internal·프론트·REST
- 사용자 요청: 소켓 2차 뒤 "조금 더 리팩터링할 부분 파악"(process 상태 전이 흩어짐 조사는 불필요 — 소켓-6(process 구독 지점 흩어짐) 정리로 충분)
- 방법·결과 → `REF-refactor-cleanup.md`: 후보 10개 정돈-1(PortBridge 실행 잔재)·정돈-2(옛 WS 클라이언트)·정돈-3(worker 저장소 트랜잭션)·정돈-4(작은 잔재)·정돈-5(실행 인터페이스 안 쓰는 메서드)·정돈-6(tile.go 세 주제)·정돈-7(주석 "표시줄")·정돈-8(TileFrame.vue 섞인 관심사)·정돈-9(프론트 공용 유틸 미사용)·정돈-10(파일 전송 송신부). 코드 변경 없음, R2(범위 결정) 대기

## 2026-10-21(2) R2(범위 결정)·R3(묶음별 정리) — 가·나·다
- 사용자 결정: 가(안 쓰는 코드 삭제) -> 나(tile.go 나누기) -> 다(작은 것), 그 전에 소켓 2차 커밋(`0dc83a7`)
- 한 것·확인 → `REF-refactor-cleanup.md` "결과". 정돈-4 중 `ConnState`는 getter만이 아니라 쓰기만 하던 필드까지 같이 지움. 정돈-1의 Go 범용 유틸(path·byte·LimitArray)은 REF-util(프론트 범용 유틸)과 달리 정돈-1 목록에 넣어 승인받은 것이라 삭제
- vault 고침: `REF-infra.md`(ConnState)·`REF-process.md`(IInteractive·execute 보유)·`REF-db.md`(worker Transaction)·CURRENT 정리 잔여 줄

## 2026-10-21(3) 정돈-8(TileFrame.vue 섞인 관심사) — 맞춤-1(타일 띄우기) composable로
- 조사 때 보류 추천이었으나 사용자가 진행 결정. 위험 질문에 "로직 그대로 옮김, 실수 여지 = `.value` 넘김(반응성 끊김, type-check로 안 잡힘)·setup 밖 호출" 답 -> 인자를 `Ref` 타입으로 받아 막기로
- 한 것·확인 → `REF-refactor-cleanup.md` "결과". 사용자 폰 확인 5/5 통과
- vault 고침: `REF-board-keyboard.md` 맞춤-1 파일 칸

## 2026-10-21(4) 파트(리팩터링) 마감
- 사용자 "지금 마감하자" -> 남은 후보(정돈-9(프론트 공용 유틸 미사용)·정돈-10(파일 전송 송신부) 유지, 소켓-8(kind 이름 접두어 제각각) 보류)는 하지 않고 마감
- 브랜치에 소켓 밖 정리 한 커밋(`38a24a7`) -> vault 접기(`INDEX-refactor.md`, MEMORY·HISTORY에서 줄 뺌) -> master에 `--no-ff` 머지, 브랜치는 남김
