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
## files-tile 묶음 (2026-11-03 밤) — 첫 updater 배포(단위-7(교체 빌드))
- `build/home-update/files-tile/`(이 폴더째 Updater 화면 "폴더 업로드", 장비 루트 `home-server` 빌드 `261103_files`): master `8123223` Windows `supervisor.exe`·`worker.exe`(태그 `SUPERVISOR-FILES-HOMESERVER-261103`·`WORKER-FILES-HOMESERVER-261103`, 빌드 뒤 env 태그 `…-UPDATER-261103`으로 되돌림) · `web/`(`VITE_API_URL=`, noindex) · `steps.json` · `nexus-steps/*.ps1`(ASCII, LF — updater가 BOM·CRLF 사본으로 실행)
- deploy = `stop`(00-stop: `-hold` 아닌 worker + supervisor 종료) -> `dbdump`(01-dbdump: `supervisor.env` DB_* + `PGPASSWORD`, `C:\Program Files\PostgreSQL\*\bin\pg_dump.exe` -> `updater\db-backups\<db>-<시각>-<버전>.dump`, 최근 3개) -> `backup` -> `copy` -> `start`(03-start: `Start-ScheduledTask Nexus` + 새 supervisor `/api/nope` 404 60초, 주 worker는 경고만) -> `prune`(90-prune-old: `C:\Nexus` 바로 아래 + `web\`의 `*.old-*`, 사용 중이면 넘어감). rollback = stop -> restore -> start, DB 복원 없음(새 마이그레이션 없음)
- 확인(VM `nexus-win`, PS 5.1.26100): 네 스크립트 문법 OK · prune 범위(위 두 곳만 지움, `update-store`·`updater` 남김) OK · dbdump env 읽기 OK(VM엔 PostgreSQL 없어 pg_dump 없음 오류로 끝남 = 예상). 실제 pg_dump·start는 집 서버에서 처음
- 정-3(관찰 방식) = 사용자가 회사에서 Updater 화면으로 지켜봄(11-03 밤). 되돌리기 시험은 따로 안 함
- **결과(11-03 밤, 사용자 "잘 동작해")**: 빌드 이름 = `file-tile-261103`(staging `file-tile-261103-d718f93c1759af83`, 파일 67개). 모든 단계 성공. prune = 지움 0 · 넘어감 1(`worker.exe.old-20261103-182151` 액세스 거부 = 보관 터미널이 쥔 옛 worker, 예상대로). 11-03 낮 `*.old-20261103-123559` 3개는 이미 없었음(지운 것으로 보임). `worker.exe.old-…182151`은 터미널이 닫힌 뒤 다음 배포의 prune이 지움
- 남은 것: 집 서버 updater 새 빌드 교체(나중, `install.ps1` 수동 — 옛 P4 updater는 replace 무시) · `C:\Nexus\nexus-before-updater.dump` 삭제(사용자, 이제 `updater\db-backups\`에 새 dump 있음)
- (옛 줄) 다음: 단위-7(교체 빌드) — Updater 화면에서 집 서버 supervisor·worker 교체 빌드(`steps.json`)를 만들어 배포 → `REF-update-p4.md` "단위-7(교체 빌드) 후속"(정-1~정-3 사용자 답 대기)

## updater 교체 — 사용자 직접 빌드 (2026-11-04 예정, 11-03 밤 사용자 "내일은 내가 직접")
| 순서 | 할 일 |
|---|---|
| 1. 태그 | `apps/core/cmd/updater/env` 첫 줄을 새 태그로(지금 `# UPDATER-P4-261103`, 예 `# UPDATER-FILES-261104`). 같은 태그면 `build.sh`가 "이미 있음"으로 멈춤 |
| 2. 빌드 | 저장소 루트에서 `scripts/updater/build.sh` -> `build/updater/<태그>/windows-amd64/{updater.exe, install.ps1}`(linux-arm64도 같이 나옴) |
| 3. 묶기 | 위 두 파일 + `setup.ps1`을 한 폴더에. **`setup.ps1`은 git에 없음** — `build/home-update/updater-p4/updater/setup.ps1`에서 복사(`$PSScriptRoot`의 `install.ps1`·`updater.exe`를 씀) |
| 4. 확인 | Updater 화면에서 `home-server`에 진행 중 배포가 없는지(설치가 `NexusUpdater` 작업을 멈춤) |
| 5. 설치 | 집 서버 관리자 PowerShell: `powershell -ExecutionPolicy Bypass -File .\setup.ps1` -> "이미 설치됨 — 설정은 그대로 두고 exe만" -> 작업 재시작. 키 다시 안 넣음 |
| 6. 확인 | Updater 화면에서 인스턴스 online + 버전 = 새 태그. 이후 배포부터 `replace`(갈아치우기)가 먹음 |
- 태그를 빌드 뒤 되돌릴지는 사용자 판단(P4 때 Claude는 supervisor·worker env만 되돌렸고 updater env는 그대로 둠)

