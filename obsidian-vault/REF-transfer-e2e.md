# REF — 파일 전송 e2e 확인

> 2026-10-19 신설(파트(파일 전송) 안 가(검증만)). 이력 → `history/transfer-e2e.md` / 모듈 설계 → `REF-transfer.md`(파트(통신 인프라), `INDEX-infra.md`) / 확인 스크립트 명세 → `REF-test-scripts-1014.md` "2026-10-19 파일 전송" / 현재 진행 → `CURRENT.md`

**상태: 시나리오 9개 확인 끝(34/34), 코드 수정 없음. F1(재시도 창)·F2(청크 왕복) 고칠지 사용자 결정 대기.**

## 확인 방법
- 격리 supervisor(5051, 임시 DB `nexus_xfer_tmp`) + 지연·끊김 TCP 프록시(5152, Python) + 격리 worker(`NAME=xfer-dev`, `TMPDIR` = 수신 루트)
- `SendFile`·`AbortFile`은 호출처가 없어서 **`go build -overlay`로 확인용 REST를 끼움**(저장소 파일 없음): `cmd/supervisor/main.go` 대체본이 `NewSupervisorRouter`가 돌려준 router에 `/xfer/file`·`/xfer/buf`·`/xfer/abort`·`/xfer/jobs`·`/xfer/workers`를 붙임 + `router` 패키지에 `XferWorkers`·`XferJobs`(진행 중 transferId) 파일 추가
- 끊김 = 프록시에 SIGUSR1(지금 연결 전부 abort) / 지연 = 프록시 `DELAY`(편도 초)

## 결과 (2026-10-19)
| 시나리오 | 내용 | 결과 |
|---|---|---|
| T1(큰 파일) | 50MB `SendFile` | 통과, sha 일치, 57MB/s(로컬) |
| T2(버퍼 여러 청크) | 5MB `SendBuffer` | 통과 |
| T3(끊김 이어받기) | 300MB 중 40%에서 연결 끊기 -> 같은 파일 다시 보내기 | 첫 send는 실패(F1(재시도 창)), `.part` 123MB 남음 -> 다시 보내면 123MB부터 이어받아 sha 일치 |
| T4(abort) | 300MB 중 30%에서 `AbortFile` | send 즉시 끝남(`전송 중단됨: context canceled`), `.part` 삭제, job 정리 |
| T5(해시 불일치) | 끊긴 뒤 원본 앞 1MB를 바꾸고 다시 보내기 | 1차 = 이어받은 결과 hash 불일치 -> `.part` 삭제 -> 2차에서 처음부터 받아 성공(바뀐 원본과 sha 일치) |
| T6(경로 이탈) | `../x`·`{WORKER_BASE}/../x`·`/etc/x`·`a/../../x` | 4개 다 즉시 거절(재시도 없음, 0ms), 밖에 파일 없음. `sub/dir/ok.bin`은 허용 |
| T7(동시 전송) | 30MB 3개 동시 | 셋 다 성공, sha 일치 |
| T8(지연) | 2MB `SendBuffer`, 왕복 10ms / 50ms | 2.7MB/s / **600KB/s**(F2(청크 왕복)) |
| T9(worker 재시작) | 300MB 중 40%에서 worker SIGTERM -> 재실행 -> 다시 보내기 | 같은 서브키로 재접속, 121MB부터 이어받아 성공 |
- 끝난 뒤 supervisor·worker에 원본·`.part` fd 남지 않음, 진행 중 job 0

## 찾은 것
- **F1(재시도 창)** 실측: 끊기고 1.06초 만에 `3회 시도 실패: 대상 연결 없음`, worker 재접속은 2.07초 -> **같은 send 안에서는 이어받기가 일어나지 않는다.** 호출자가 다시 보내야 `.part`에서 이어받음(transferId가 달라도 이어받기는 수신 경로 `.part` 기준이라 됨)
- **F2(청크 왕복)** 실측: 32KB마다 응답을 기다려 속도 ≈ 32KB / 왕복 시간. 외부 접속(왕복 수십 ms)에서 MB급 파일은 느림
- **abort sentinel(구 잔여 3번) 불필요**: abort 에러가 `%w`로 `context.Canceled`를 감싸므로 호출자가 `errors.Is(err, context.Canceled)`로 구분 가능
- 해시 불일치는 끝까지 받은 뒤에야 판정 -> 전체를 한 번 더 보냄(설계대로, 이어받는 사이 원본이 바뀐 경우만 생김)
- 코드로만 본 것(확인 안 함): worker가 끊긴 상태에서 `AbortFile`하면 `MsgFileAbort`를 못 보내 `.part`가 남음(best-effort) / 끊긴 전송의 worker `SaveFile`은 `recvExpire` 60초 동안 fd를 들고 있음(새 전송과 같은 `.part`를 열지만 옛 것은 쓰지 않음)

## F1·F2 고치는 안 (결정 전)
| 항목 | 안 | 비고 |
|---|---|---|
| F1(재시도 창) | 가) "대상 연결 없음"이면 worker 재접속을 일정 시간(예: 10초) 기다린 뒤 재시도 / 나) 그대로 — 호출자가 다시 보냄 | 지금 호출처 = EXEC·EDIT 선배치(작은 파일)뿐이라 영향 없음 |
| F2(청크 왕복) | 가) 청크 키움(32KB -> 256KB 등, 와이어는 base64 JSON) / 나) 응답을 기다리지 않고 여러 청크를 먼저 보내는 창 방식 / 다) 그대로 | 같음 — 큰 파일을 보내는 기능(안 나(업로드 기능))을 만들 때 의미가 생김 |
