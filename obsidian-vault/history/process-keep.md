# history — process 살려 두기

> → `REF-process-keep.md` / → `CURRENT.md`

## 2026-10-22
- 사용자 요청(worker 폴더 타일보다 먼저): worker 재시작에도 process 유지 + sqlite 매칭
- 원인 정리: supervisor 재시작 = process 삶·ring(화면) 사라짐 / worker 재시작 = 정리 또는 SIGHUP으로 죽음
- 안 가(process마다 보관 프로세스)~라(tmux) 제시, 1차 답: 화면 복원도 대상·비정상 종료까지·끌 때 전부 종료 수단·Linux 먼저
- 2차 답: 기본 = 살려 둠 / 통신 형식 최대한 유지(버전 번호) / nohup·& 문제없음 확인 / 끝난 process 주기 삭제 요청 -> 3차 질문
- 3차: 사용자 'worker 테이블 불필요, supervisor 주기 삭제' -> 소켓 폴더 훑기로 sqlite 대체 가능, supervisor 삭제는 R2-a(타일 있는 끝난 process)와 충돌 지점 확인 중
- 4차: K1(worker 목록 = sqlite, 끝나면 바로 삭제)·K2(supervisor 24시간 주기 삭제)·K3(타일 같이 정리)·K4(기본 = 살려 둠)·K5(형식 유지·버전 번호)·K6(Linux 먼저) 확정 -> 큰 계획 제시
- 실행 방식 질문(모드로 띄움·두 번 실행?·worker 종료 시 보관 프로세스?) 답함, K7(끝난 뒤 기다림) = 10분 확정

## 2026-10-22(2) 구현·확인
- 진행 신호("네가 코드를 짜는거야. 진행해줘") -> MEMORY 불변 결정 2개 개정(process 영속성·worker 종료 = 살려 둠)
- 살리기-1~7 작성: `internal/execute/hold` 새 패키지, worker `router/hold.go`·`00002_holds.sql`, Screen을 `internal/ring`으로 옮기고 `WriteAt`·`Seed`, AgentInteractive `Chunk`, MsgScreen, 프론트 `writeLive`(lastOff), supervisor `processPurge.go`
- 설계 중 판단: offset을 보관 프로세스 값으로 통일 -> supervisor 재시작 뒤 열린 브라우저도 겹침 없이 이어짐(프론트가 lastOff로 거름). 스냅샷도 출력과 같은 줄(AgentInteractive 큐)로 relay해 발행 순서 유지
- 격리 확인 `keep_test.py` 40/40. 처음 전체 실행 뒤 lost 시나리오의 소켓·로그가 남음 -> `hold.Cleanup` 추가 후 재확인. Go 단위 테스트·`vue-tsc`·Windows 빌드 통과

## 2026-10-22(3) 시험 배포(test1)·사용자 확인
- 사용자(외부 접속 중): 8013 빌드본은 두고 8011(test1)에 따로 배포 요청, 시작 전 확인 -> DB 안 가(`nexus_build` 복사본 `nexus_keep`) 선택. 같은 DB 위험 설명 중 "서로 502로 닫음"은 틀린 설명이었음(인스턴스키가 sqlite마다 다름) -> 정정
- `scripts/build.sh keep`, env `build/keep/*.env`(8011·`nexus_keep`·sqlite 따로·메인키 같음), setsid 실행. 복사본 purge 173개
- Claude 확인: test1 경유 로그인·소켓·실행. 사용자 터미널로 worker 재시작·supervisor 재시작 -> 사용자 "화면 이어지고 입력도 돼"
- 사용자 결정: 새 Claude 세션을 test1에서 열고 8013(test3)을 새 빌드로 재배포 -> CURRENT 재배포-1~6(test3 재배포)

## 2026-10-22(4) test3 재배포
- 새 세션(test1 위)에서 CURRENT 재배포-1~6 진행. 재배포-2 전 사용자 확인(test3에 막 열린 터미널 1개 끝남) -> "진행해줘"
- 8013을 `build/keep` 바이너리 + `../*.env`로 기동, purge 173개, worker sqlite 00002 첫 적용. Node 시험 스크립트로 출력 10줄·exit 0 확인(첫 실행은 바이너리 프레임 파싱 실패 -> 시험 계정 1개 더 생김)
- 사용자 "동작 확인했어, 그쪽에서 세션 다시 접속" -> vault 정리. 남은 것 = 재배포-6(마감)

## 2026-10-22(5) 마감
- test3 위 새 세션에서 `CLAUDE*` 변수 상속 발견 -> `REF-build.md` "Claude가 띄운 빌드본", BACKLOG "PTY 환경에서 `CLAUDE*` 빼기"
- test1을 끌 때 test3 영향 검토: PID 지정 SIGTERM이면 없음, `pkill` 패턴·`build/keep` 폴더 삭제는 test3에 영향 -> `REF-build.md` "두 빌드 나란히"
- 사용자 "커밋하고 worker 폴더 타일 시작" -> `go build`·`go vet`·`go test`·`vue-tsc` 통과, 재배포-6(마감): INDEX-process-keep으로 접고 코드 + vault 한 커밋(`apps/core/cmd/irony/` 제외)
- 정리 3개(test1 내릴지·`build/iutf`·시험 계정)는 2026-10-23 사용자와 같이 -> CURRENT
