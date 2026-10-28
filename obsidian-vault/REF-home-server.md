# REF — 집 Windows 서버 PC에 supervisor 올리기 (Cloudflare Tunnel)

> → 이력 `history/home-server.md` · → 배포·업데이트 `REF-home-deploy.md` · → 진행 `CURRENT.md`
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

## 재부팅 (2026-10-25)
- PostgreSQL·cloudflared = 서비스라 자동 / supervisor·worker = `start.ps1` 수동이라 안 켜짐 -> 작업 스케줄러 등록 스크립트 `build/nexus-autostart.ps1`(저장소 밖, git 제외): 부팅 30초 뒤(DB 먼저), 내 계정(SYSTEM 아님 = 터미널 권한·홈), 로그온 여부 무관, `-Force`로 다시 실행 가능. 해제 = `Unregister-ScheduledTask -TaskName Nexus`
- 확인(2026-10-25 사용자): 등록(State Ready) -> 재부팅 -> 접속 OK. 다시 재부팅해 원격 데스크톱 없이 폰으로 접속 OK = **로그인 전 자동 시작 확인**. `Start-Process` 자식은 작업이 끝나도 살아 있음
- .ps1은 영어만(PowerShell 5.1은 BOM 없는 UTF-8을 CP949로 읽어 한글이 깨짐)

## 겪은 것
- `Restart-Service`·`New-NetFirewallRule`은 관리자 PowerShell 필요
- DBeaver 편집기가 다른(MySQL) 연결에 붙어 `CREATE USER` 문법 오류 -> 편집기 활성 연결 확인(`Ctrl+9`)

## cloudflared 보안 공지 (2026-10-28)
- 메일 = eMudhra 중간 인증서 8개 개인키 유출(10-19) -> cloudflared가 그 인증서를 믿지 않게 고친 버전으로 올리라는 공지. 고친 버전 = 2026.10.2 이상(Cloudflare 블로그 `blog.cloudflare.com/emudhra-certificate-compromise`)
- 서버 PC = `2026.10.3`(10-25 첫 설치가 공지 뒤라 이미 해당), 서비스 Running·Automatic, 대시보드 커넥터 버전도 `2026.10.3` -> **할 일 없음**. origin = `http://localhost:8013`라 origin 인증서 교체도 해당 없음
- 원격 업데이트가 필요해지면: 브라우저 화면 자체가 터널 경유라 업데이트 중 끊기고, 실패하면 회사에서 복구 수단 없음 -> 집에서 `winget upgrade Cloudflare.cloudflared`(관리자) 권장. 회사에서 해야 하면 작업 스케줄러 1회 작업(`REF-home-deploy.md` 함정-A(Job) 회피) + 서비스 Running 확인·옛 exe 되돌림 스크립트(미작성)
