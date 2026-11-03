# history — 집 서버 배포: updater 이후

→ `REF-home-deploy-updater.md` · 그 전 `history/home-deploy.md` · `CURRENT.md`

## 2026-11-03 낮 — updater-p4 묶음
- 사용자: 관리자 = irony1090, 묶음 = `feat/updater-p4`로(집 서버 확인 뒤 P4 마감·머지)
- 빌드: 태그 바꿔 Windows exe 3개 + web -> `setup.ps1` 작성 -> VM 확인 6/6 -> `nexus-updater-p4.zip`(39MB). 폴더 이름 `updater/` -> `updater-p4/`로 바꿈(설치 자리와 겹침)
- (12:35) 사용자 집 서버 배포(회사에서 Nexus 터미널로): DB 백업 -> `update.ps1` -> `OK: new supervisor is up on port 8013`, 옛 파일 `*.old-20261103-123559`. 코드-6(색인 금지 미배포) 집 서버엔 들어감(로컬 test3 빌드본은 아직 옛것)
- (12:42) `setup.ps1` 실행 -> `C:\Nexus\updater` 생성·예약 작업 `NexusUpdater` 실행, 사용자 "잘 동작했어"
- (오후) 정리: `drag\`·`updater-p4\`·`nexus-updater-p4.zip`·`*.old-20261029-205139` 3개 지움. 오늘 12:33 백업이 drag 명령 그대로 `nexus-before-drag.dump`로 저장돼 옛 drag 백업을 덮어써서 -> `nexus-before-updater.dump`로 이름 바꿈(남은 DB 백업은 이것 하나). 남김 = `*.old-20261103-123559`(web·supervisor·worker, 새 빌드 확인 뒤 지움 — `worker.exe.old`는 배포 전 터미널이 닫혀야 지워짐) · dump(단위-7(교체 빌드)까지 끝난 뒤 지움)
