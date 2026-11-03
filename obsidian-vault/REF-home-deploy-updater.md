# REF — 집 서버 배포: updater 이후 (2026-11-03 신설)

> → 이력 `history/home-deploy-updater.md` · 그 전 수동 묶음(drag 등) `REF-home-deploy.md` · updater 구현 `REF-update-p4.md` · 진행 `CURRENT.md`
> 서버 PC 폴더 = `C:\Nexus`(supervisor `localhost:8013`, 작업 `Nexus`). 묶음 위치 = `build/home-update/`(git 제외)

## updater-p4 묶음 (2026-11-03) — 마지막 수동 배포
- `build/home-update/nexus-updater-p4.zip`(폴더 `updater-p4/` — `updater/`로 하면 `C:\Nexus\updater`(updater 설치 자리)에 풀림): `feat/updater-p4`(base master `cb75903` = P1~P3 + 코드-6(색인 금지 미배포)) Windows `supervisor.exe`·`worker.exe`·`web/`(`VITE_API_URL=`, JS에 test2 주소 없음, index.html에 색인 금지 meta) + `update.ps1`(drag 것 그대로) + `updater/{updater.exe, install.ps1, setup.ps1}` + README
- 태그: `SUPERVISOR-UPDATER-HOMESERVER-261103` · `WORKER-UPDATER-HOMESERVER-261103` · `UPDATER-P4-261103`(빌드 뒤 supervisor·worker는 `…-UPDATER-261103`으로 되돌림)
- DB: 마이그레이션 `00016`~`00018` = `update_*` 새 테이블만(기존 데이터 안 건드림). 보관소 = `C:\Nexus\update-store`(env 없음 = 기본값). 새 env 필수 항목 없음
- 관리자 = `irony1090`(사용자 11-03, MEMORY "배포 전 필수")
- `setup.ps1`: 키는 `Read-Host -AsSecureString`(명령 기록에 안 남게, `-Key`는 시험용) · `supervisor.env` PORT로 `ws://localhost:<PORT>` · `targetDir` = `C:\Nexus` · `updater.json`은 BOM 없이(Go JSON) · 이미 있으면 그대로 두고 exe만. VM 확인 6/6(역방향 ssh 터널로 VM `localhost:5051` -> 격리 supervisor)
- 순서: DB 백업 -> `update.ps1` -> 브라우저 Updater 화면에서 장비 루트 `home-server` 만들고 키 복사 -> `setup.ps1` -> 화면에서 인스턴스 online·설치 경로 `C:\Nexus` 확인
- **결과(11-03)**: 12:35 `update.ps1` OK · 12:42 `setup.ps1` OK -> `C:\Nexus\updater`·작업 `NexusUpdater` 동작
- 정리 뒤 `C:\Nexus`(11-03 오후): 쓰는 것 = `update-store`·`updater`·`view-cache`·`web`·exe 2개·`*.env`·`worker-sqlite.db*`·`start.ps1`·`stop.ps1`·로그. 남은 정리 = `*.old-20261103-123559` 3개(사용자 "바로 지울까?" -> 지워도 됨, git `29ae134`로 다시 만들 수 있음, `worker.exe.old`는 배포 전 터미널이 닫혀야) · `nexus-before-updater.dump`(하나뿐인 DB 백업, 단위-7의 `pg_dump`가 한 번 돈 뒤 지우기 권함)
- 다음: 단위-7(교체 빌드) — Updater 화면에서 집 서버 supervisor·worker 교체 빌드(`steps.json`)를 만들어 배포 → `REF-update-p4.md` "단위-7(교체 빌드) 후속"(정-1~정-3 사용자 답 대기)
