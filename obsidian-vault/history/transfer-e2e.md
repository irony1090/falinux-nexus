# history — 파일 전송 e2e 확인

> → `REF-transfer-e2e.md` / 모듈 이력(2026-06) → `history/transfer.md` / 현재 진행 → `CURRENT.md`

## 2026-10-19
- 사용자가 다음 파트로 파일 전송 선택 -> 코드 조사: 실사용 = EXEC·EDIT 선배치 `SendBuffer`(작은 파일, 청크 1개)뿐, `SendFile`·`AbortFile` 호출처 없음. 코드로 F1(재시도 창)·F2(청크 왕복) 발견 -> 안 가(검증만)·안 나(업로드 기능 + 검증) 제시 -> 사용자 "가로 진행"
- 격리 supervisor + overlay 확인용 REST(`/xfer/*`) + 지연·끊김 프록시 + 격리 worker로 T1~T9 실행. 함정 2개: 키의 `#`가 URL을 자름(쿼리 인코딩) / `pgrep -x worker`가 사용자 worker를 먼저 잡음(cwd 검사로 막힘 -> cwd로 고름)
- 결과 34/34, 코드 수정 없음. F1(재시도 창) 실측 1.06초 실패 vs 재접속 2.07초 / F2(청크 왕복) 실측 왕복 50ms에 600KB/s / abort sentinel은 `errors.Is(context.Canceled)`로 충분

## 2026-10-20
- 사용자 "F1, F2 수정 진행" -> 코드 재확인 중 발견: 끊길 때 연결 종료 처리가 reader를 `Close()` -> 재접속을 기다려도 다음 시도 `SeekTo` 실패 + `AbortFile`이 job을 못 찾음(T3에선 마지막 에러 "대상 연결 없음"에 가려짐) -> F1-a(재접속 대기)·F1-b(강제 Close 제거)·F2(청크 256KB) 제시 -> 사용자 "진행"
- 구현 후 격리 재확인 21/21(T1·T3·T4·T4b(대기 중 abort)·T10(재접속 안 됨)·T9·T8). T8은 HEAD 빌드(32KB)로 같은 프록시에서 기준을 다시 재서 비교 — 50ms 왕복 615KB/s -> 4.5MB/s
- 사용자 "마감" -> `INDEX-transfer-e2e.md`로 접음 + 커밋(코드 + vault)
- 스크립트 함정: `SendFile` 수신 이름 = 원본 파일 이름(`Info().Name()`) -> 시나리오마다 원본을 하드링크로 다른 이름으로 / 옛 프록시의 "읽을 때마다 sleep" 지연은 큰 프레임에서 대역폭까지 깎음 -> 받은 시각 + 지연 큐로 바꿈
