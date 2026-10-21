# INDEX — 리팩터링 (마감)

> 2026-10-21 파트(리팩터링) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
목적 = 사람이 Claude가 짠 코드를 읽기 쉽게, 기준 = 성능이 떨어지지 않는 선까지, 주석 = 손대는 파일만 한 줄 기준. 브랜치 `refactor/socket`(master `03bc9a6`에서 분기)에서 작업하고 마감 때 master에 `--no-ff` 머지.
- 소켓 알림 1차(`8cf32bc`): 소켓-1(계정 토픽 이름, `TILES:` -> `USER:`)·소켓-5(커밋 뒤 발행 규약 섞임)·소켓-6(process 구독 지점 흩어짐) + 주석 13파일, 격리 32/32
- 소켓 알림 2차(`0dc83a7`): 소켓-2(여러 계정에 보내기 중복)·소켓-3(볼 수 있는 계정 판정 중복)·소켓-4(탭 하나로 보내기)·소켓-7(빈 패키지·낡은 주석), 격리 35/35
- 소켓 밖 정리(`38a24a7`): 묶음 가(안 쓰는 코드 삭제, 약 1,200줄)·나(tile.go 나누기, `tile.go`·`tileDoc.go`·`tilePlace.go`)·다(작은 것, `IInteractive` 2메서드·"표시줄" -> "인디케이터") + 정돈-8(TileFrame.vue 섞인 관심사, 맞춤-1(타일 띄우기) -> `hook/tileFloat.hook.ts`). 격리 35/35 + 폰 실기기 5/5

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-refactor-socket.md`(파트(리팩터링) — 목적·기준 + supervisor 소켓 알림 정리 후보) / `history/refactor-socket.md`
- `REF-refactor-socket-batches.md`(같은 파트 묶음별 계획·결과) / `history/refactor-socket-batches.md`
- `REF-refactor-cleanup.md`(같은 파트 — 소켓 밖 정리 후보·결과, 옛 `tile.go` 함수의 새 위치) / `history/refactor-cleanup.md`
- 확인 스크립트 명세 → `REF-test-scripts-1020.md` "2026-10-21 소켓 알림 리팩터"

## HISTORY 색인 행 (원문)
### `history/refactor-socket.md` — 리팩터링: supervisor 소켓 알림 후보 조사 (2026-10-21 신설)
- 2026-10-21 R1(목록 조사) 소켓 알림 범위 — 후보 9개, 코드 변경 없음

### `history/refactor-cleanup.md` — 리팩터링: 소켓 밖 정리 후보 (2026-10-21 신설)
- 2026-10-21 R1(목록 조사) 2회차 — 후보 10개(정돈-1(PortBridge 실행 잔재) ~ 정돈-10(파일 전송 송신부)), 코드 변경 없음
- 2026-10-21(2) 묶음 가(안 쓰는 코드 삭제)·나(tile.go 나누기)·다(작은 것), 격리 35/35
- 2026-10-21(3) 정돈-8(TileFrame.vue 섞인 관심사) 맞춤-1(타일 띄우기) -> `useTileFloat`, 폰 확인 5/5
- 2026-10-21(4) 파트(리팩터링) 마감 — 커밋 `38a24a7` -> INDEX 접기 -> master `--no-ff` 머지

### `history/refactor-socket-batches.md` — 리팩터링: 소켓 알림 정리 묶음 (2026-10-21 분할)
- 2026-10-21(2) 1차 정리: 소켓-1(계정 토픽 이름)·소켓-5(커밋 뒤 발행 규약 섞임)·소켓-6(process 구독 지점 흩어짐) + 주석 13파일, 격리 32/32
- 2026-10-21(3) PC 꺼짐 뒤 재개, 브랜치 `refactor/socket` 분기, 1차 한 커밋
- 2026-10-21(4) 2차: 소켓-2(여러 계정에 보내기 중복)·소켓-3(볼 수 있는 계정 판정 중복)·소켓-4(탭 하나로 보내기)·소켓-7(빈 패키지·낡은 주석), 격리 35/35

## 이 파트에만 해당하는 결정
- 남긴 것: 소켓-8(kind 이름 접두어 제각각) 보류(프론트 동시 수정 비용) / 소켓-9(worker·presence push 없음) 범위 밖 = 기능 변경 후보 / 정돈-9(프론트 공용 유틸 미사용) 유지(범용 유틸, `REF-util.md`) / 정돈-10(파일 전송 송신부) 유지(안 나(업로드 기능)에서 씀)
- 다른 파트에도 지킬 것(계정 토픽 이름 `USER:<userId>`)은 MEMORY 불변 결정 "실시간 push"에 이미 있음
