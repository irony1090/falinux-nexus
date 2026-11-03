# history — 업데이트 app P3(배포)

> ↔ `REF-update-p3.md` · 설계 `REF-update-deploy.md` · 현재 `CURRENT.md`

## 2026-11-02 밤 — 착수 계획 합의
- P2 master 머지(`bd93fa4`) · 설계-8(supervisor 되돌리기, updater 범위 밖)·설계-10(자기 교체) 닫은 뒤 계획 초안 단위-1(DB)~단위-7(확인)
- 확정 시안 ③(배포 진행·잠그기) 대조: Run·Log = 배포 id·인스턴스 단위, 배포 대상은 시안에 타일처럼 그려져 있었음
- 사용자 답: 배포 대상 = 팝업 / Run·Log 해석(배포 × 인스턴스, 빌드 행 = 마지막 배포) 동의 / 재부팅 단계는 다음번 -> 이번엔 중단 기록 + 멈춤
- BACKLOG에서 P3로 올림: updater 화면 실시간 알림 · 끊긴 동안 출력(-> 단위-5(실시간 알림)·단위-2(배포 메시지))

## 2026-11-03 새벽 — P3 구현(로컬 확인까지)
- 브랜치 `feat/updater-p3`. 단위-1(DB)~단위-6(화면) 작성 → 구현 지도·차이-1(되돌리기 기록)~차이-5(자동 되돌리기 기록) `REF-update-p3.md`
- 화면 확인(격리 5051 + 로컬 updater, 헤드리스)에서 발견: 짧은 단계면 인스턴스가 "실행 중"으로 남고 로그가 빔 → 원인 = supervisor가 단계 끝(REQ, 고루틴)을 시작 알림·출력(EVENT, 디스패처)보다 먼저 처리 → transport `HandleOrdered`(차이-3) + 반복 배포 5회 확인 항목. 로그 조회·실시간 이음매 off(차이-4). 터미널은 상자가 생길 때 만듦(배포 조회 뒤에 그려짐)
- 확인: `run.sh` 전체 통과(새 항목 updater-deploy 38 · updater-deploy-ui 18), `go test ./...`(transport `TestHandleOrderedAfterEvents`·steps `TestAutoRollback`)
- 실기 전 막힘: 개발 supervisor(8012) 재시작 명령이 권한 확인에서 거절됨 → 사용자에게 재시작 요청

## 2026-11-03 아침 — `imx8mp-f24` 실기 확인 통과
- 사용자 요청으로 Claude가 8012 재시작(07:51) → 개발 DB `nexus`에 `00018` 적용. 보드 updater를 `install.sh`로 P3 실행 파일로 교체(updater.json 유지, `pVMx1WwE` 재접속). 태그는 아직 `UPDATER-P2-261102`(`cmd/updater/env` 안 바꿈)
- 확인 스크립트(scratchpad `f24/p3.py`, API) 17개 통과: v1 배포 5단계(ask 입력 `hi f24`) → 설치본 v1 / v2(`bad.sh` exit 3 + `autoRollback`) → 배포 실패·자동 되돌리기 성공·check 건너뜀·app.conf ver=1 복귀·설치본 v1 유지 / v1 잠그기(v2 배포는 잠그기 불가) / v1 배포 수동 되돌리기 → app.conf ver=0
- 1개 FAIL은 기대값 오류: 두 번째 v1 배포를 되돌리면 설치본 = 그 배포 직전 값 v1(보드 백업 `d2/manifest.json` `prevInstalled: v1`로 확인)이 맞음
- test1 화면(헤드리스): v2 빌드 행 Run·Log → 배포 #3 타일 = 배포 실패·되돌리기 성공 태그, `rollback · 자동`, 단계 표시·rollback 로그, 브라우저 오류 없음
- 배포 id 3개(#1~#3)·빌드 v2·v1 잠김은 개발 DB에 남겨 둠, 보드 서비스는 test2에 붙은 채 켜 둠
- 빌드 태그를 P3로 바꿀지 물었더니 사용자: 지금은 중요하지 않음 → 그대로. P3 마감 → 브랜치에 코드 커밋 + vault 커밋

## 2026-11-03 오전 — 사용자 화면 확인 · 질문 정리 · master 머지
- Claude가 test1(8011)·test2(8012) 다시 띄움 → 보드 `pVMx1WwE` 재접속. 사용자 확인 "P3까지 나쁘진 않은 것 같아"
- 질문 답(코드 변경 없음): builtin은 `backup`·`copy`·`restore` 셋뿐(나머지는 script) / `bash`를 넣은 단계는 `"interactive": true`가 있어야 입력 가능(없으면 stdin `/dev/null` -> 바로 EOF로 끝남) / 보관소 구조(`devices/`·`common/`·`logs/`·`.nexus-db`) / 실행 로그 보관 기한 → `REF-update-p3.md` "로그 보관"
- 로그 보관 기한은 정하지 않음(사용자: 잊어도 됨, 필요하면 손으로 지움)
- 같이 커밋: 화-10(타일 크기) 개정(→ `history/member-ui.md` 11-03). `feat/updater-p3` master `--no-ff` 머지

