# history — 업데이트 app P2(updater Linux)

> ↔ `REF-update-p2.md` · 설계 `REF-update-deploy.md` · 현재 `CURRENT.md`

## 2026-11-02 밤 — P1 머지 · P2 전 확정
- `feat/updater-p1` master 머지(`1f5eda9`, push = 사용자)
- P2 전 확정 6개: 제안표(공유 토큰·`+x`·JSON·`fa24-PID`·설정 파일·보기 전용) → 사용자 답으로 정-1(메인키별 API 키, 자동 생성 폐기)·정-2(파일 mode, 켜기·끄기 + 8진수)·정-4(`imx8mp-f24`, 지금 test3에 붙음)·정-6(대화형) 바뀜, 정-3(JSON)·정-5(설정 파일, 웹 변경 = 가) 그대로
- 정정: 처음에 "키 원문은 발급 때 한 번만"이라 했으나 worker 장비 키는 봉인해서 다시 보기가 됨(회원-Q16) → 같은 방식
- 사용자 질문 "공유 토큰 = 같은 장비명 공유?" → 아님(메인키는 updater마다 설정, 토큰은 공통 비밀번호)이라 답함
- 정-1(접속 인증) 키 저장 방식 = 다시 보기 가능(해시 + `KEY_SECRET` 암호화 원문, worker 키와 같음) 사용자 확인. 한 번만 표시 안은 안 씀
- 정-5 실시간 변경 규칙(배포 중 거절·manifest에 경로) · P2 확인 = API·스크립트만 → 사용자 "진행해줘", 브랜치 `feat/updater-p2`

## 2026-11-02 밤 — P2 구현(로컬 확인까지)
- 단위-1(DB) `00017` · 단위-2(키·권한 화면) · 단위-3(접속) · 단위-4(updater 본체) · 단위-5(받기·builtin) · 단위-6(script 실행) · 단위-7(인스턴스 화면) · 단위-8(확인) 작성 → 구현 지도 `REF-update-p2.md`
- 결정(구현 중): updater 키 접두어 `nxu_`(worker `nxk_`와 구분) · FolderSha는 기본 권한이면 P1 형식 그대로(잠금 sha 호환) · 장비 지우면 붙은 updater 연결 끊음 · 비대화형 출력 LF→CRLF(웹 터미널 표시) · 시작 때 running 일괄 취소 대신 HELLO Running 비교
- 확인: `run.sh` 전체 통과(회원 회귀 + updates-api 44 + updater-ui + updater-run 34), `go test ./...`·vet 통과, 헤드리스 화면(키·updater.json·인스턴스 설정·권한 창) 확인
- 개발 supervisor(8012) 다시 띄움 → 개발 DB `nexus`에 `00017` 적용(23:13). 샘플 `sample-jig` 키 발급 + 로컬 updater 하나 붙임(scratchpad `upd-sample/`, 설치 경로도 scratchpad)

## 2026-11-02 밤 (23:30~) — `imx8mp-f24` 실기 확인 통과
- 사용자 확인: "설치 경로 `/opt/nexus-updater-test`" = 배포 대상 폴더(`TARGET_DIR`), updater 자체는 `/opt/nexus-updater`. 장비 루트 만들기·화면 확인은 Claude가 먼저(API + 헤드리스)
- arm64 빌드(sha 앞 `e340f898`) -> scp -> `install.sh` -> `접속됨 imx8mp-f24#pVMx1WwE`(test2). 장비 루트·키는 API로 발급(메모 "P2 실기 확인(11-02)"), 대상 폴더에 옛 `app.conf`(ver=0) 미리 둠
- 빌드 `v1` = app.conf(ver=1)·hello.sh·ask.sh(interactive)·check.sh·steps.json. 권한 자동(.sh 755, 나머지 644)
- 단계 실행(API) 6개 전부 ok: backup(있던 1·새 4, `backups/r1`) -> copy(권한 그대로) -> hello(uname aarch64, `TARGET_DIR`·`VERSION`·`STEP_ID` 확인) -> ask(입력 전 `name? ` 출력, 입력 `f24-real` -> `hi f24-real`) -> check(`BACKUP_DIR=…/r1`, ver=1) -> restore(되돌림 1·새 파일 지움 4, ver=0 확인)
- 서비스 재시작 = 같은 서브키(state.json)로 다시 붙음 / 설치 경로 웹 변경: 절대 경로 -> updater.json 갱신(600 유지), 상대 경로 -> 409 "절대 경로여야" / 원래 경로로 되돌림
- 화면(test1 헤드리스): 장비 타일에 키(가림·보기·복사·재발급) · 인스턴스 online·`linux/arm64`·경로·태그 · 빌드 v1. 콘솔 에러 없음. 기록 보기 화면은 P3(Run·Log)라 API로만 확인
- 서비스는 켜 둠(사용자 (가))
- 11-02 밤 master 머지(`bd93fa4`, push = 사용자). 이 줄과 CURRENT·색인의 머지 표시는 다음 파트 커밋에 같이 들어감
