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

## 터널 속도 (2026-10-28 질문)
- 인터넷에 도는 "속도 올리기" = 서비스 `ImagePath`를 `tunnel run --protocol quic --token-file <파일>`로 바꾸기. **판단: 속도 효과는 거의 없음**
  - `--protocol` 기본값 `auto`가 이미 QUIC 우선(안 되면 http2로 내려감). 강제 quic은 지금 http2로 떨어져 있을 때만 의미, 대신 UDP 7844가 막히면 내려가지 못하고 끊김
  - `--token-file`은 그 파일을 먼저 만들어야 함. `service install <토큰>`으로 깔았으면 ImagePath에 `--token eyJ...`가 들어 있음 -> 파일 없이 바꾸면 재시작 때 서비스 시작 실패 = 터널 끊김
  - `Set-ItemProperty`만으론 적용 안 됨(`Restart-Service cloudflared`, 관리자). 재시작 = 터널 끊김이라 회사에서 원격으로 하지 말 것(위 "보안 공지"의 원격 업데이트 주의와 같음)
- 체감 속도는 보통 **어느 Cloudflare 엣지에 붙었나**(한국 무료 요금제는 해외 엣지로 잡히는 경우가 많음)와 집 업로드 대역이 좌우 -> 대시보드 커넥터의 엣지 위치·로그 `Registered tunnel connection ... protocol=` 먼저 확인
- 증상(사용자 2026-10-28): 첫 로딩 느림 + 터미널 타자 반응 늦음. 코드 확인 결과
  - 터미널 입력 = `term.onData` -> 바로 전송(지연·묶음 없음), worker 출력도 묶음 타이머 없음 -> 타자 지연 = 거의 전부 왕복 시간(브라우저 -> 엣지 -> 터널 -> 집 -> worker). 회사 worker 터미널이면 회사 -> 집 -> 회사로 터널을 두 번 지남
  - 첫 로딩: supervisor 정적 서빙(echo `Static`)에 `Cache-Control` 없음 · `noto-sans-korean.scss`의 `src` 순서가 eot -> **otf** -> woff -> woff2라 브라우저가 otf(~400KB/굵기)를 받음(woff2를 앞에 둬야 함) · js·css 합계 ~1.5MB + 아이콘 woff2 400KB
  - 확인법: 브라우저로 `https://<도메인>/cdn-cgi/trace`의 `colo=`(ICN = 서울, NRT·KIX = 일본, 그 밖 = 더 멂) · 개발자도구 Network의 각 요청 Waiting(TTFB) · 응답 헤더 `cf-cache-status`
- **측정(2026-10-28 사용자)**: `cdn-cgi/trace` = `colo=LAX`·`loc=KR`·`http=http/3` -> 한국에서 접속해도 Cloudflare 엣지가 로스앤젤레스로 잡힘(무료 요금제). 같은 집 worker를 회사 supervisor(Cloudflare 안 거침)에 붙이면 빠름 -> **원인 = Cloudflare 경로(LAX 왕복)로 확정**, 프로토콜·앱 코드 아님. 대안(엣지를 안 거치는 경로) 검토 중

## 직통 전환 계획(안, 2026-10-28 — 미착수, 사용자 결정 대기)
- 방향: Cloudflare = DNS만(회색 구름), 트래픽 중계(cloudflared 터널) 뺌. iptime 포트포워딩 80·443 -> 서버 PC Caddy(HTTPS 자동) -> `localhost:8013`
- 유동 IP = iptime DDNS(**이미 발급받은 `*.iptime.org` 있음**, 사용자) + Cloudflare DNS CNAME(회색)
- 순서(안): WAN IP 공인 확인 -> 서버 PC 내부 IP 고정(iptime DHCP 고정 할당) -> Caddy 설치·서비스 등록 + 방화벽 80·443 -> 포트포워딩 -> 시험 주소 `direct` CNAME으로 확인 -> 되면 `nexus`를 터널에서 CNAME으로 전환 -> cloudflared 정지(삭제는 안정 확인 뒤). 전환 단계는 집에서
- 진행: 1단계(공인 IP 확인) = DDNS가 공인 대역 가리킴(10-28, iptime 화면 WAN과 일치는 사용자 확인). 인증서 발급 실패가 반복되지 않게 **DNS·포트포워딩을 먼저, Caddy 시작은 마지막**으로 순서 조정. Caddy = `C:\Caddy\`(exe + Caddyfile), 서비스 = `sc.exe create caddy start= auto binPath= "...caddy.exe run --config ...Caddyfile"`
- 회사에서 할 수 있는 범위(10-28): DNS `direct` 추가(대시보드) · Caddy 파일 준비(Nexus 터미널 `IRONY-SERVER-WORKER`) · 방화벽·`sc.exe`는 그 터미널이 관리자일 때만 / iptime 1·2단계(내부 IP 고정·포트포워딩)와 시험 실행은 **집에서**. 임시 터널 경로로 iptime 관리 화면 열기 = 공유기 관리 화면을 인터넷에 여는 것이라 비추천
- DNS 추가 화면(사용자 스크린샷 `user/home-server/dns_추가.png`): 형식 CNAME · 이름 `direct` · 대상 `irony1090.iptime.org` · **프록시 상태 끔(회색 = DNS 전용)** · TTL 자동
- **3단계(DNS) 완료(10-28)**: `direct` CNAME -> `irony1090.iptime.org` DNS 전용, 메모 "iptime ddns를 통한 직통"(스크린샷 `user/home-server/dns_추가완료.png`). `nslookup direct.irony1090.app 1.1.1.1` = 집 공인 IP로 풀림 확인
- Caddy 설치 = PowerShell로 GitHub zip 직접 받기(10-28 기준 `v2.11.7`, `caddy_2.11.7_windows_amd64.zip`). winget(`CaddyServer.Caddy`)도 있지만 Nexus 터미널(작업 스케줄러로 뜬 worker 아래)에선 winget이 안 될 수 있어 직접 받기 권장. PS 5.1은 `Tls12` 지정 + `$ProgressPreference='SilentlyContinue'`(진행 표시 때문에 느림)
- **사용자용 체크리스트 = user vault `user/home-server/직통-전환.md`**(10-28, 단계·명령·상태 칸). 단계가 진행되면 그 파일 상태 칸도 같이 갱신
- **5단계(Caddy 준비) 완료(10-28)**: `C:\Caddy\caddy.exe` v2.11.7 + Caddyfile, `validate` = Valid(들여쓰기 형식 경고만 = Caddyfile은 탭 들여쓰기 기대, 동작 무관, `caddy fmt --overwrite`로 정리 가능). 남은 것: 0(권한 확인) -> 4(방화벽)·7(서비스) / 집에서 1·2·6
- 0단계 = Nexus 터미널 관리자 `True`(10-28) -> 4(방화벽)·7(서비스 등록) 회사에서 가능. Caddyfile에 전역 `storage file_system { root C:/Caddy/data }`(서비스 SYSTEM 계정과 인증서 위치 공유) + `log { output file C:/Caddy/caddy.log }`(서비스는 화면 출력 없음 -> 원격 확인용) 추가. 서비스는 `start= demand`로 등록만, 집에서 포트포워딩 뒤 `sc.exe start` -> 로그 확인 -> `start= auto`. 체크리스트 순서도 7 -> 6으로 바꿈
- **회사에서 할 것 끝(10-28)**: 4(방화벽 규칙 "Caddy 80-443") · Caddyfile 전역 설정 판 Valid · 7(서비스 `caddy` 등록, 수동·Stopped). 집에서 남은 것 = 1(내부 IP 고정) -> 2(포트포워딩) -> 6(시험 실행 -> `start= auto`) -> (선택) 8(주소 전환)
- 8(주소 전환) 주의: `nexus` 레코드 **구름만 회색 금지** — 대상 `<터널ID>.cfargotunnel.com`은 공개 IP가 없어 접속 끊김. public hostname 삭제 -> CNAME `irony1090.iptime.org`(회색) -> Caddyfile에 `nexus` 추가 순. `direct` 직통 확인 뒤에만(안 되면 회사에서 들어갈 경로가 없어짐)
