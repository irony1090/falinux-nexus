# history — 집 서버 배포: updater 이후

→ `REF-home-deploy-updater.md` · 그 전 `history/home-deploy.md` · `CURRENT.md`

## 2026-11-03 낮 — updater-p4 묶음
- 사용자: 관리자 = irony1090, 묶음 = `feat/updater-p4`로(집 서버 확인 뒤 P4 마감·머지)
- 빌드: 태그 바꿔 Windows exe 3개 + web -> `setup.ps1` 작성 -> VM 확인 6/6 -> `nexus-updater-p4.zip`(39MB). 폴더 이름 `updater/` -> `updater-p4/`로 바꿈(설치 자리와 겹침)
- (12:35) 사용자 집 서버 배포(회사에서 Nexus 터미널로): DB 백업 -> `update.ps1` -> `OK: new supervisor is up on port 8013`, 옛 파일 `*.old-20261103-123559`. 코드-6(색인 금지 미배포) 집 서버엔 들어감(로컬 test3 빌드본은 아직 옛것)
- (12:42) `setup.ps1` 실행 -> `C:\Nexus\updater` 생성·예약 작업 `NexusUpdater` 실행, 사용자 "잘 동작했어"
- (오후) 정리: `drag\`·`updater-p4\`·`nexus-updater-p4.zip`·`*.old-20261029-205139` 3개 지움. 오늘 12:33 백업이 drag 명령 그대로 `nexus-before-drag.dump`로 저장돼 옛 drag 백업을 덮어써서 -> `nexus-before-updater.dump`로 이름 바꿈(남은 DB 백업은 이것 하나). 남김 = `*.old-20261103-123559`(web·supervisor·worker, 새 빌드 확인 뒤 지움 — `worker.exe.old`는 배포 전 터미널이 닫혀야 지워짐) · dump(단위-7(교체 빌드)까지 끝난 뒤 지움)

## 2026-11-04 — 집 서버 최신 배포(261104) + updater 교체
- 사용자 계획: updater 뺀 나머지를 updater로 배포 -> 그 뒤 updater는 Nexus로 수동. 태그 그대로(`*-FILES-261104`, 로컬 test3와 같은 태그 -> 같은 태그 경고는 예상)
- 빌드(사용자): `scripts/build.sh 261104-files_win supervisor worker frontend updater --os windows`(처음 물은 `-os window`·`261104-files`(test3 실행 폴더)는 고쳐서). 폴더명은 밑줄
- 묶음(Claude): `build/home-update/261104-files/` = exe 2 + `web/`(60) + files-tile `steps.json`·`nexus-steps/` 그대로, `updater/` 뺌(보완-1(자기 폴더 보호)). 마이그레이션 변화 없음 확인
- (14:15~14:18) 사용자 업로드(빌드 `supervisor-url-261104`, 67개, 화면 캡처로 Claude 대조) -> 배포 #2 6단계 성공 -> 잠금. 사용자 질문 "배포 전에 잠그기 돼?" -> 안 됨(성공 배포 기록 필요, Claude가 순서를 잘못 말했던 것 정정)
- Claude가 Chrome 탭(사용자 로그인 세션)으로 Run·Log 읽음: dbdump 65,510 B · start pid 9024/9704 · prune `worker.exe.old-…182151` 지움. 공개 `index.html` 해시 = 묶음과 같음
- (14:33) updater 교체: `build/home-update/updater-261104/`(exe·install.ps1·setup.ps1) -> 사용자가 `C:\Nexus\updater-new\`에 올려 **Nexus 디렉토리 타일에서 `setup.ps1` 실행**(worker가 관리자 권한이라 통과). `LastTaskResult 267009` = 실행 중(정상). Updater 화면 online + `UPDATER-FILES-261104`
- 남은 것: `C:\Nexus\updater-new\` 삭제(사용자)

## 2026-11-05 ~ 11-06 보드 닫기 묶음
- 11-05 Claude가 `scripts/build.sh 261105-board-close_win supervisor worker frontend --os windows`(태그 `SUPERVISOR-`·`WORKER-BOARD-CLOSE-261105`) -> 묶음 `build/home-update/261105-board-close/`(67개, steps·nexus-steps = 261104-files 그대로). Chrome 확장 연결 안 돼 업로드 못 함(회사 PC Chrome은 다시 켬)
- 11-06 사용자가 Updater 화면에서 직접 올리고 배포. 공개 `index-CqlA8_4K.js` = 묶음과 같음(curl 확인)

## 2026-11-06 (저녁)
- 사용자 요청으로 집 서버 묶음 `build/home-update/261106-socket-reconnect/` 만듦(supervisor·worker·web, Windows, 태그 `SOCKET-RECONNECT-261106`) -> 업로드·배포 = 사용자
- 2026-11-07 소켓 재연결 묶음(`261106-socket-reconnect`) 사용자 배포. 공개 `index-AJJLr_Z9.js` = 묶음과 같음(nexus·direct curl 확인)

## 2026-11-07 (밤)
- test3 supervisor 교체 `261106-socket-reconnect` -> `261107-tile-shell`(태그 `SUPERVISOR-TILE-SHELL-261107`, swap.sh, 21:48 켜짐, 공개 `index-BIm258Bp.js`, worker 다시 붙음) · 집 서버 묶음 `build/home-update/261107-tile-shell/` 만듦(업로드·배포 = 사용자) → `REF-home-deploy-updater.md` "11-07 공용 타일 틀 묶음"
- 사용자 집 서버 배포 -> 공개 `index-BIm258Bp.js` 일치(nexus·direct)
- 11-07 사용자: 집 서버에 `updater-new` 폴더 없음 -> 할 일에서 뺌(vault엔 이름이 "예"로만 적혀 있었음, 11-04 updater 교체용 임시 업로드 폴더)
