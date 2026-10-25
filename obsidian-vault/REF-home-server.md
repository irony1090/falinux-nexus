# REF — 집 Windows 서버 PC에 supervisor 올리기 (Cloudflare Tunnel)

> → 이력 `history/home-server.md` · → 진행 `CURRENT.md`
> Nexus 기능 파트와 별개인 **운영 환경** 기록(사용자 요청 2026-10-25: 집 서버 이야기는 따로). 서버 PC 쪽 폴더 = `C:\Nexus`

## 구성 (2026-10-25 조율)
| 자리 | 내용 |
|---|---|
| 서버 PC | 집의 옛 게이밍 PC, Windows 10, 유선(Private 프로필), iptime 공유기 아래. 작업 PC와 물리적으로 다르고 같은 망 |
| 외부 노출 | Cloudflare Tunnel(무료, 포트포워딩 없음, https·wss 자동). 고정 주소는 도메인 필요(Cloudflare Registrar에서 원가 구매 가능, `.kr` 등 일부 안 팜). 도메인 없이 시험만 = Quick Tunnel `*.trycloudflare.com`(매번 주소 바뀜) |
| DB | PostgreSQL 15(winget EDB, 구성 요소 = Server + Command Line Tools), 로케일 DEFAULT, 계정 `irony` + DB `nexus`(OWNER irony, UTF8, template0). 5432는 같은 망에만: `pg_hba.conf` `host all all <내부대역>/24 scram-sha-256` + 방화벽 규칙 "PostgreSQL 5432"(`-RemoteAddress LocalSubnet`). **iptime 포트포워딩 금지** |
| supervisor | 마이그레이션은 시작 때 자동(goose embed, master = 00001~00012). `env` 없으면 embed 값으로 내뱉고 종료(REF-build E3) -> 고쳐서 재실행. Windows 빌드 = `GOOS=windows GOARCH=amd64 go build ./cmd/supervisor` + `web/`(컴파일만 확인, 실행 미확인) |

## 빌드 묶음 (2026-10-25, master `19cc8bd`)
- 결정: master 최신 · PORT 8013 · 서버 PC에 worker도 같이(`NAME="IRONY-SERVER-WORKER"`, `ws://localhost:8013`)
- 만든 것(scratchpad `nexus-server/` + `.zip`, 세션 끝나면 사라짐): `supervisor.exe`·`worker.exe`(`GOOS=windows GOARCH=amd64`) + `web/`(`VITE_API_URL=` 빌드) + `supervisor.env`(DB `irony`/`nexus`, `DB_PASS=CHANGE_ME`, `SESSION_KEY` 무작위 48자, `IP=0.0.0.0`) + `worker.env` + `start.ps1`/`stop.ps1`(-hold는 남김) + `README.txt`. 두 exe가 한 폴더라 기본 이름 `env`가 겹침 -> 각자 `-env supervisor.env`/`-env worker.env`
- 태그: 빌드 동안만 `cmd/*/env` 첫 줄을 `*-HOMESERVER-261025`로 바꿨다가 원래(`CTRLPANEL-261023`)로 되돌림. exe 안 embed env는 test3 배포 값(내뱉기용, `env` 파일이 있으면 안 씀)
- env 줄 끝 CRLF 괜찮음(godotenv)

## 코드 쪽 확인 결과 (Cloudflare 거칠 때)
- 웹·API 같은 주소(supervisor가 `web/` 서빙) -> 터널 하나 `-> http://localhost:<PORT>`
- 브라우저 소켓 https면 `wss` 자동, 앱 ping 10초라 Cloudflare 유휴 100초 끊김 안 걸림
- 업로드 1MB 조각(`uploadChunkMax`) -> 무료 요금제 요청 100MB 제한 무관
- 쿠키 `SameSite=Lax`·`Secure=false` 같은 주소라 동작
- 집 밖 worker = `WS_SCHEME=wss`·`WS_HOST=<도메인>`
- **`/worker`는 인증 없음**(2026-10-25 확인, `handleWorkerWS` = 업그레이드 후 `MsgRegister`만): 터널 주소로 누구나 아무 메인키로 worker 등록 가능 -> 그 장비로 실행되는 스크립트 내용·출력이 넘어갈 수 있음. 서버 PC worker는 `localhost`로 붙으므로 **호스트 전체에 Cloudflare Access**(또는 WAF 규칙 `/worker` 차단)로 막는다
- 보안: 화면 앞에 Cloudflare Access(무료 50명) 권장, worker 경로는 Access 밖(별도 주소 또는 서비스 토큰). `SESSION_KEY`는 긴 무작위 값으로

## 업데이트 빌드 B (2026-10-25, worker_move 포함)
- 사용자 결정 B(확인 전 코드라도 집 서버에 먼저): 임시 worktree(scratchpad)에서 master `19cc8bd` + `feat/supervisor-move` `40e0942`를 `merge --no-commit`(코드 충돌 없음, vault 4개만 충돌) -> Windows 빌드 -> worktree 지움. **master·브랜치는 그대로**
- 묶음 `build/nexus-update.zip`(첫 묶음 = `build/nexus-server.zip`, git 제외): exe 2개 + `web/` + `update.ps1`(supervisor·주 worker 끄고, 실행 중 exe는 덮어쓸 수 없어 `*.old-<시각>`으로 이름 바꾼 뒤 복사, `start.ps1`) + README. env·sqlite는 그대로. 켜질 때 supervisor `00013`(users.perms)·worker `00003`·`00004` 자동
- 계정: 회원가입 화면 없음 -> SQL로 생성. `identification`·`password` 둘 다 `sha256 hex`(`toHash`) = `encode(sha256(convert_to('<값>','UTF8')),'hex')`. `createUser`는 INSERT 말고 다른 일 없음. 권한 = `perms = '{worker_move}'` -> SQL 파일 `build/nexus-create-user.sql`(생성·확인·권한만·비밀번호 변경)

## 재부팅 (2026-10-25)
- PostgreSQL·cloudflared = 서비스라 자동 / supervisor·worker = `start.ps1` 수동이라 안 켜짐 -> 작업 스케줄러 등록 스크립트 `build/nexus-autostart.ps1`(저장소 밖, git 제외): 부팅 30초 뒤(DB 먼저), 내 계정(SYSTEM 아님 = 터미널 권한·홈), 로그온 여부 무관, `-Force`로 다시 실행 가능. 해제 = `Unregister-ScheduledTask -TaskName Nexus`
- 확인(2026-10-25 사용자): 등록(State Ready) -> 재부팅 -> 접속 OK. 다시 재부팅해 원격 데스크톱 없이 폰으로 접속 OK = **로그인 전 자동 시작 확인**. `Start-Process` 자식은 작업이 끝나도 살아 있음
- .ps1은 영어만(PowerShell 5.1은 BOM 없는 UTF-8을 CP949로 읽어 한글이 깨짐)

## 내일 할 것 (2026-10-26 회사, 사용자 계획)
- 회사에 설치한 worker 전부 최신(B 빌드 = supervisor-move 포함)으로 바꾸고, 붙을 supervisor를 집 서버(`wss://nexus.<도메인>`)로
- 필요한 것: **Linux worker B 빌드는 아직 없음**(이번엔 Windows exe만) -> 회사 가기 전이나 가서 같은 방법(임시 worktree에서 master + supervisor-move)으로 Linux·Windows worker 빌드. 또는 supervisor-move를 확인·master 머지한 뒤 정식 빌드(그러면 B 빌드와 같은 코드)
- 옮기는 법: 옛 빌드 worker는 주소 변경 명령을 모름 -> **worker 실행 파일 교체 + env `WS_SCHEME="wss"`·`WS_HOST="nexus.<도메인>"`로 재시작**이 가장 간단(env 주소가 지난번과 다르면 1순위). 이후엔 집 서버 화면(장비 ⋮ "supervisor 주소 변경" / 보드 메뉴 "전체 worker 주소 변경")으로 옮길 수 있음
- 주의: 집 서버 DB는 test3와 따로(폴더·스크립트·공유 없음) -> 옮긴 worker의 실행 중 터미널은 집 서버에서 "미등록 process"(장비 ⋮에서 보기·끄기). worker sqlite 서브키는 그대로 이어짐. worker 쪽 마이그레이션 `00003`·`00004`는 켜질 때 자동
- 주의: 나중에 Access를 호스트 전체에 걸면 회사 worker의 `/worker` 접속도 막힘 -> `/worker` 경로 예외(Bypass) 또는 서비스 토큰

## 겪은 것
- `Restart-Service`·`New-NetFirewallRule`은 관리자 PowerShell 필요
- DBeaver 편집기가 다른(MySQL) 연결에 붙어 `CREATE USER` 문법 오류 -> 편집기 활성 연결 확인(`Ctrl+9`)
