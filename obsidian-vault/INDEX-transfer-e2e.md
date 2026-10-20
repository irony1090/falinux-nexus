# INDEX — 파일 전송 (마감)

> 2026-10-20 파트(파일 전송) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). 모듈 설계(2026-06) → `INDEX-infra.md`의 `REF-transfer.md`. → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
호출처가 EXEC·EDIT 선배치(작은 `SendBuffer`)뿐이던 파일 전송을 overlay 확인용 REST(`/xfer/*`) + 지연·끊김 프록시 + 격리 worker로 e2e 확인(안 가(검증만), T1(큰 파일)~T9(worker 재시작) 34/34). 거기서 찾은 F1(재시도 창)·F2(청크 왕복)를 고침: F1-a(재접속 대기 10초)·F1-b(끊길 때 reader 강제 Close 제거)·F2(청크 32KB -> 256KB) -> 재확인 21/21. 끊김·worker 재시작이 같은 send 안에서 이어받기로 끝나고, 왕복 50ms에서 615KB/s -> 4.5MB/s

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-transfer-e2e.md`(파일 전송 e2e 확인 — T1~T10, F1(재시도 창)·F2(청크 왕복) 고친 것) / `history/transfer-e2e.md`
- 확인 스크립트 명세 → `REF-test-scripts-1014.md` "2026-10-19 파일 전송"(2026-10-20 보강 줄 포함)

## HISTORY 색인 행 (원문)
### `history/transfer-e2e.md` — 파일 전송 e2e 확인 (2026-10-19 신설)
- 2026-10-19 안 가(검증만): overlay 확인용 REST + 프록시로 T1~T9 34/34, F1(재시도 창)·F2(청크 왕복) 실측, 코드 수정 없음
- 2026-10-20 F1-a(재접속 대기)·F1-b(강제 Close 제거)·F2(청크 256KB) 구현 + 재확인 21/21 -> 마감·커밋

## 이 파트에만 해당하는 결정
- 재시도 = 3회, 연결이 없으면 시도마다 최대 10초(200ms 간격) 재접속을 기다림, abort면 즉시 빠짐. 끝내 안 되면 실패까지 ~21초(설계대로)
- 미완 전송의 reader 수명 = `send`의 `defer Close`만 정함(연결 종료 처리에서 닫지 않음)
- F2 나)(창 방식) 안 함 — worker가 REQ를 고루틴으로 처리해 청크 순서가 바뀌고 `SaveFile.written`이 최대 끝 위치라 빈 구멍 뒤부터 이어받게 됨
- abort sentinel 안 만듦 — `errors.Is(err, context.Canceled)`로 충분
- 안 나(업로드 기능)는 안 함 — 큰 파일을 보내는 기능을 만들 때 그때 결정
