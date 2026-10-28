# REF — 집 서버 배포·업데이트 기록 (2026-10-28 `REF-home-server.md`에서 분할)

> → 이력 `history/home-deploy.md` · → 구성·터널·재부팅 `REF-home-server.md` · → 진행 `CURRENT.md`
> 서버 PC 폴더 = `C:\Nexus`. 묶음 위치 = `build/home-update/`(git 제외)

## 업데이트 빌드 B (2026-10-25, worker_move 포함)
- 사용자 결정 B(확인 전 코드라도 집 서버에 먼저): 임시 worktree(scratchpad)에서 master `19cc8bd` + `feat/supervisor-move` `40e0942`를 `merge --no-commit`(코드 충돌 없음, vault 4개만 충돌) -> Windows 빌드 -> worktree 지움. **master·브랜치는 그대로**
- 묶음 `build/nexus-update.zip`(첫 묶음 = `build/nexus-server.zip`, git 제외): exe 2개 + `web/` + `update.ps1`(supervisor·주 worker 끄고, 실행 중 exe는 덮어쓸 수 없어 `*.old-<시각>`으로 이름 바꾼 뒤 복사, `start.ps1`) + README. env·sqlite는 그대로. 켜질 때 supervisor `00013`(users.perms)·worker `00003`·`00004` 자동
- 계정: 회원가입 화면 없음 -> SQL로 생성. `identification`·`password` 둘 다 `sha256 hex`(`toHash`) = `encode(sha256(convert_to('<값>','UTF8')),'hex')`. `createUser`는 INSERT 말고 다른 일 없음. 권한 = `perms = '{worker_move}'` -> SQL 파일 `build/nexus-create-user.sql`(생성·확인·권한만·비밀번호 변경)

## 내일 할 것 (2026-10-26 회사, 사용자 계획)
- 회사에 설치한 worker 전부 최신(B 빌드 = supervisor-move 포함)으로 바꾸고, 붙을 supervisor를 집 서버(`wss://nexus.<도메인>`)로
- 필요한 것: **Linux worker B 빌드는 아직 없음**(이번엔 Windows exe만) -> 회사 가기 전이나 가서 같은 방법(임시 worktree에서 master + supervisor-move)으로 Linux·Windows worker 빌드. 또는 supervisor-move를 확인·master 머지한 뒤 정식 빌드(그러면 B 빌드와 같은 코드)
- 옮기는 법: 옛 빌드 worker는 주소 변경 명령을 모름 -> **worker 실행 파일 교체 + env `WS_SCHEME="wss"`·`WS_HOST="nexus.<도메인>"`로 재시작**이 가장 간단(env 주소가 지난번과 다르면 1순위). 이후엔 집 서버 화면(장비 ⋮ "supervisor 주소 변경" / 보드 메뉴 "전체 worker 주소 변경")으로 옮길 수 있음
- 주의: 집 서버 DB는 test3와 따로(폴더·스크립트·공유 없음) -> 옮긴 worker의 실행 중 터미널은 집 서버에서 "미등록 process"(장비 ⋮에서 보기·끄기). worker sqlite 서브키는 그대로 이어짐. worker 쪽 마이그레이션 `00003`·`00004`는 켜질 때 자동
- 주의: 나중에 Access를 호스트 전체에 걸면 회사 worker의 `/worker` 접속도 막힘 -> `/worker` 경로 예외(Bypass) 또는 서비스 토큰

## Linux B 빌드 (2026-10-26 회사)
- `build/supervisor_move/{supervisor, worker, web/}`(Linux amd64), 태그 `SUPERVISORMOVE-261026`. 코드 = 집 서버 B 빌드와 같음(master `2cbd306`의 `apps/` = `19cc8bd` + `feat/supervisor-move` `40e0942`). 만든 법 = scratchpad 임시 worktree에서 `merge --no-commit`(vault 4개만 충돌) + `cmd/*/env` 태그만 바꿔 복사 + `node_modules` 링크 -> `scripts/build.sh` -> 결과를 `build/`로 옮기고 worktree 지움
- embed env = test3 값 그대로(`localhost:8013`·`nexus_build`). 사용자 계획: 먼저 test3(8013) 자리에 띄움 -> 그 뒤 화면 "supervisor 주소 변경"으로 집 서버로 옮김
- 띄울 때 붙는 것: supervisor `00013`(nexus_build) · worker `00003`·`00004`(공용 `build/worker-sqlite.db`). worker 교체 = test3 위 process 종료(REF-build)

## 회원 관리 배포 — Nexus만으로 (2026-10-26 회사, 완료)
- 조건: 회사에서 집 서버로 ssh·원격 불가 -> **집 서버 화면에서 `IRONY-SERVER-WORKER` 터미널로만** 업데이트. 집 서버 = supervisor-move B 빌드, 다른 worker는 전부 test3. 관리자 = `irony1090`
- 함정-A(Job): Windows 터미널 process는 Job(`KILL_ON_JOB_CLOSE`, `execInteractive_windows.go`) 안 -> 터미널에서 `Start-Process`로 띄운 supervisor는 스크립트가 끝나면 같이 죽음 -> 터미널은 `Start-ScheduledTask Nexus`만 부르고 교체·재시작은 스케줄러 쪽에서
- 함정-B(키): 새 supervisor는 키 없는 worker 거부 -> 키(`nxk_`)를 미리 `worker.env`에 넣고, 업데이트 뒤 "내 장비"에 **키 붙여넣기**로 등록(worker 재시도 최대 60초)
- 업데이트-1(사전 점검) `build/home-update/check.ps1`(읽기만: 권한·Job·부모 체인·env 항목 이름·`Nexus` 작업·psql·DB 회원/goose/worker) -> 업데이트-2(묶음 업로드) -> 업데이트-3(준비: 압축 해제·`KEY_SECRET`·`API_KEY`·관리자 SQL·`Start-ScheduledTask`) -> 업데이트-4(교체: 15분 안에 서버 PC worker가 새 supervisor에 안 붙으면 옛 파일로 자동 되돌림, 판정 = `worker_instances.last_seen` > 시작 시각) -> 업데이트-5(화면: 키 붙여넣기)
- 업데이트-1 결과(2026-10-26): elevated True · Job 안 True(함정-A 실재) · 부모 = `worker.exe -hold`(**Windows 보관 프로세스 동작 = K6 일부 확인**) · `Nexus` 작업 Ready(strik, Highest, Password, `start.ps1`) · goose 13 · `irony1090`(id 1) `{worker_move}` · 업로드 + 폴더 타일 직접 실행 OK
- 롤백 안전: 옛 B 빌드 supervisor가 v14 DB 복사본(`nexus_member_tmp` 템플릿)에서 정상 기동(GET 200, 버전 14 유지) -> DB 복원 없이 파일만 되돌리면 됨
- 묶음 `build/home-update/`(git 제외): `nexus-member.zip`(master `4ae97a7`, 태그 `MEMBER-261026`, Windows exe 2개 + `web/` + `apply.ps1`·`common.ps1`·`start.ps1`(PENDING 훅)·`status.ps1`·`VERSION.txt`) + 따로 올리는 `prepare.ps1`(압축 해제 -> env `KEY_SECRET`·`API_KEY`(`nxk_`) 추가 -> `pg_dump` -> 관리자 SQL -> start.ps1 교체 -> 키 출력 + `GO` 입력 -> `PENDING` + `Start-ScheduledTask`). `apply.ps1` = 끄기 -> `now()` -> `*.old-<stamp>` -> 켜기 -> HTTP 60초 -> 15분 `last_seen > t0` 기다림, 실패 = `Restore`(새 로그 `update\new-*.log`)
- 함정: `cmd/*/env`는 git 추적 밖 -> 태그 되돌리기는 `git checkout` 말고 손으로
- **결과(2026-10-26 16:56~16:59)**: `prepare.ps1` -> GO -> apply 16:56:52 끄기 -> 16:57:01 supervisor up -> 키 붙여넣기 등록 -> 16:59:03 DONE. 남은 것(서버 PC): `*.old-20261026-165652`(exe 2개·web)·`start.ps1.old-*`·`update\`(dump 포함) — 지워도 됨. `start.ps1`의 PENDING 훅은 그대로 둬도 무해, **다음 업데이트도 같은 방법**(zip + `prepare.ps1`, 이미 있는 env 키는 재사용)

## 정리 (2026-10-26 저녁)
- `build/`에서 `nexus-server.zip`·`nexus-update.zip`·`nexus-autostart.ps1`·`nexus-create-user.sql` 삭제(사용자: 새 배포에 필요 없으면 지움 — 작업 스케줄러는 이미 등록, 가입 화면이 생겨 계정 SQL 불필요). 다시 필요하면 `REF-home-server.md` "재부팅"·이 파일 "업데이트 빌드 B" 절대로 다시 작성
- `build/home-update/`는 ps1만 남김(zip·exe·web·`ct/` 삭제). 다음 배포(`/api` 정리 + 주소 일괄 변경)는 supervisor.exe + `web/`만 바꾸면 됨. 집에선 ssh·원격 가능
- **집 서버 다음 배포 묶음(2026-10-26 저녁)** `build/home-update/nexus-api-prefix.zip`(폴더 `api_prefix/`): master `c5f9749` Windows `supervisor.exe`(태그 `SUPERVISOR-API_PREFIX-HOMESERVER-261026`, 빌드 뒤 `API_PREFIX-261026`로 되돌림) + `web/`(`VITE_API_URL=`, `/api` 확인) + `update.ps1` + README. worker·env 그대로
  - **2026-10-26 밤 다시 묶음**: 웹만 master `8de3e05`(`fix/phone-layout` 머지 — 폰 수정 포함, test3 확인본과 같은 빌드)으로 교체, `supervisor.exe`는 그대로(c5f9749 이후 Go 변경 없음). README 첫 줄 갱신. 옛 zip은 세션 scratchpad에만
  - **배포 결과(2026-10-26 22:19)**: 사용자가 `C:\Nexus\api_prefix`에 풀고 관리자 PowerShell `update.ps1` -> `OK`(8013). 서버 PC에 `*.old-20261026-221903`(supervisor.exe·web)·`C:\Nexus\api_prefix` 남음 — 화면 확인 뒤 지워도 됨. 하위 폴더에 풀어도 됨(README "C:\Nexus 말고" = 실행 중 파일 위에 직접 풀지 말라는 뜻)
  - **서버 PC 정리 기준(2026-10-26 밤 사용자 질문)**: 남길 것 = `supervisor.exe`·`worker.exe`·`web\`·`*.env`·worker sqlite·`start.ps1`(작업 스케줄러 `Nexus`가 부팅·`update.ps1` 재시작 때 실행)·`stop.ps1`(수동 끄기). 지워도 됨 = `*.old-*`·`update\`(PENDING 없으면 안 씀, 단 `nexus-before.dump` = 유일한 DB 백업)·`api_prefix\`·zip·`prepare.ps1`·`check.ps1`·`nexus-autostart.ps1`(작업 이미 등록, 다시 필요하면 `REF-home-server.md` "재부팅" 절로 재작성)·첫 묶음 README. `worker.exe.old-*`는 그 전에 연 터미널의 `-hold`가 쓰는 중이면 안 지워짐
  - `update.ps1`: supervisor + 주 worker 끄기 -> `supervisor.exe`·`web` -> `*.old-<stamp>` -> 새 것 복사 -> **`Start-ScheduledTask Nexus`로 켬**(ssh 세션·Nexus 터미널에서 띄운 프로세스는 세션이 끝나면 같이 죽으므로 어디서 실행해도 되게) -> `/api/nope` = 404면 새 빌드 확인(옛 빌드는 index.html 200) -> 60초 안에 안 되면 자동 되돌림. 관리자 PowerShell 권장
  - 사용자가 집에서 직접 배포(원격·ssh 가능)

## mochi 배포 묶음 (2026-10-27 밤)
- `build/home-update/nexus-mochi.zip`(폴더 `mochi/`): master `28053b7`(오늘 6파트) Windows `supervisor.exe`·**`worker.exe`**(여기서 셸 `{WORKER_SHELL}` = powershell) + `web/` + `update.ps1` + README. 태그 `MOCHI-HOMESERVER-261027`(빌드 뒤 `MOCHI-261027`로 되돌림). 마이그레이션 없음, 관리자 `irony1090` 그대로
- `update.ps1` 바뀐 점: 교체 대상에 `worker.exe` 추가 / 새 빌드 판정 = supervisor 프로세스 시작 시각 > 교체 시각 + `/api/nope` 404(옛 빌드도 404라 예전 판정으로는 구분 안 됨) / 주 worker 안 뜨면 WARN만(되돌리지 않음)
- **배포 결과(2026-10-27 밤)**: 사용자 `update.ps1` -> `OK`. 남은 것 정리 기준 = 위 "서버 PC 정리 기준"(+ 이번 `*.old-<stamp>` 3개·`mochi\`·zip)

## 웹만 교체 (2026-10-28)
- 사용자가 test3 웹 빌드(`build/mochi/web`, master `02dc54c` = 경로 줄 정리·패널 창 버튼·닫은 뒤 active)를 압축해 Nexus 전송으로 서버 PC에 올림 -> PowerShell `Expand-Archive` -> `web.new` -> `web` -> `web.bak-<시각>`, `web.new` -> `web`(재시작 없음, `WEB_DIR="./web"`는 요청 때 디스크에서 읽음) -> **완료(사용자)**. exe는 mochi 그대로(Go 변경 없음). 되돌리기 = 폴더 이름만 바꾸기
