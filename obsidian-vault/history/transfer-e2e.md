# history — 파일 전송 e2e 확인

> → `REF-transfer-e2e.md` / 모듈 이력(2026-06) → `history/transfer.md` / 현재 진행 → `CURRENT.md`

## 2026-10-19
- 사용자가 다음 파트로 파일 전송 선택 -> 코드 조사: 실사용 = EXEC·EDIT 선배치 `SendBuffer`(작은 파일, 청크 1개)뿐, `SendFile`·`AbortFile` 호출처 없음. 코드로 F1(재시도 창)·F2(청크 왕복) 발견 -> 안 가(검증만)·안 나(업로드 기능 + 검증) 제시 -> 사용자 "가로 진행"
- 격리 supervisor + overlay 확인용 REST(`/xfer/*`) + 지연·끊김 프록시 + 격리 worker로 T1~T9 실행. 함정 2개: 키의 `#`가 URL을 자름(쿼리 인코딩) / `pgrep -x worker`가 사용자 worker를 먼저 잡음(cwd 검사로 막힘 -> cwd로 고름)
- 결과 34/34, 코드 수정 없음. F1(재시도 창) 실측 1.06초 실패 vs 재접속 2.07초 / F2(청크 왕복) 실측 왕복 50ms에 600KB/s / abort sentinel은 `errors.Is(context.Canceled)`로 충분
