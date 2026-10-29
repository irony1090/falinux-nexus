# REF — 집 서버 직통 접속 (iptime 포트포워딩 + Caddy, 2026-10-28 `REF-home-server.md`에서 분할)

> → 이력 `history/home-direct.md` · → 서버 PC 구성·옛 터널 `REF-home-server.md` · → 진행 `CURRENT.md`
> 사용자 체크리스트 = user vault `user/home-server/직통-전환.md`(1~7단계) · `user/home-server/주소-전환-명령.md`(8(주소 전환))

## 지금 구성 (2026-10-28 밤~)
| 자리 | 값 |
|---|---|
| 경로 | 브라우저 -> `125.143.1.21`(iptime, DDNS `irony1090.iptime.org`) :80·:443 -> 서버 PC `192.168.0.2` Caddy -> `localhost:8013` |
| DNS(Cloudflare) | `nexus`·`direct` 둘 다 CNAME `irony1090.iptime.org`, **DNS 전용(회색)** |
| Caddy | `C:\Caddy\`(v2.11.7) · 서비스 `caddy` 자동 시작 · 인증서 `C:\Caddy\data` · 로그 `C:\Caddy\caddy.log` · 사이트 `nexus.irony1090.app, direct.irony1090.app` |
| iptime | 서버 PC `192.168.0.2` DHCP 고정 할당(10-28) · 포트포워딩 `80`·`https`(TCP 80·443 -> `.2`), 기존 `front` 5173~5174는 비활성화 |
| cloudflared | Stopped·Disabled(삭제 전). 되돌리기 = `주소-전환-명령.md` "되돌리기" |
| 인증서 | Let's Encrypt, `direct` = tls-alpn-01 · `nexus` = http-01로 발급 -> 80·443 둘 다 밖에서 열림 |

## 인증서 갱신 (자동)
- Caddy가 알아서 갱신(사람 할 일 없음). 2026-10-28 발급분의 Let's Encrypt 권장 갱신 구간(로그 `got renewal info`) = **2026-12-26 ~ 12-28** — 만료(약 90일)보다 한 달쯤 앞이라 실패해도 재시도할 여유가 있음
- 조건: 서비스 `caddy` 실행 중(자동 시작) · 포트포워딩 80·443 유지 · DNS가 집 IP를 가리킴(IP가 바뀌면 iptime DDNS가 따라감)
- 확인: `caddy.log`에서 `certificate renewed`·`renew` 실패 줄, 또는 브라우저 자물쇠 -> 인증서 유효 기간

- PowerShell `sc` = `Set-Content` 별칭 -> 서비스 명령은 `sc.exe`(아니면 `C:\query` 같은 파일만 생김)
- `caddy validate`도 Caddyfile의 `log` 설정대로 `caddy.log`에 줄을 남김 -> 로그 볼 때 시각(ts) 확인
- 전환 직후 쓰던 브라우저는 옛 연결·DNS 기억으로 한동안 Cloudflare 경유(trace 열림). `nslookup`은 OS 캐시를 안 거침 -> 다른 브라우저 또는 `ipconfig /flushdns` + 브라우저 소켓 Flush
- `nexus` 레코드는 구름만 회색으로 바꾸면 안 됨(대상 `cfargotunnel.com`은 공개 IP 없음) — 대상도 iptime으로
- DNS가 바뀐 뒤에 Caddyfile에 이름 추가(먼저 하면 발급 실패 반복 -> Let's Encrypt 제한)

## 진행 기록 (2026-10-28 계획 -> 같은 날 밤 완료)
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
- **2(포트포워딩) 입력(10-28 밤, 집)**: 규칙 `80` TCP 80 -> `192.168.0.2:80` · `https` TCP 443 -> `192.168.0.2:443`(사용자 스크린샷 `user/스크린샷_28-10-2026_212828_192.168.0.1.jpeg`). 같은 화면의 다른 규칙 `localhost5000`·`vRising`(-> .15)·`front` 5173~5174(-> .2)는 사용자 기존 것. `.2` = 서버 PC 확인(사용자), `front` 규칙은 비활성화(사용자). 1(내부 IP 고정) 등록 여부는 미확인 -> 다음 = 6(시험 실행)
- **6(시험 실행) 인증서 발급 OK(10-28 21:35, 집)**: `sc.exe start caddy` -> Let's Encrypt 계정 생성 -> **tls-alpn-01(443)로 검증** -> `certificate obtained successfully`. 첫 실행의 `default.json: cannot find the path` = 계정 파일이 아직 없어서 나오는 정상 info. 접속·`cdn-cgi/trace`·타자 반응 확인 -> `start= auto`는 남음
  - 함정: **PowerShell의 `sc` = `Set-Content` 별칭** -> `sc query caddy`가 `C:\query` 파일을 만들고 아무 출력 없음. 서비스 명령은 항상 `sc.exe`
- **6(시험 실행) 완료(10-28 밤)**: `direct.irony1090.app` 접속 OK · `cdn-cgi/trace` 안 열림(Cloudflare 안 거침) · `start= auto` 전환. 타자 반응 체감은 사용자가 시험 중 -> 결과 보고 8(주소 전환) 여부 결정
- **속도 확인(10-28 밤, 사용자)**: 같은 터미널 비교에서 `direct`(직통) 타자 반응이 터널(`nexus`)보다 훨씬 좋음 -> LAX 엣지 원인 판단 맞음. 다음 = 8(주소 전환) 또는 `direct`만 쓰기 결정
- **8(주소 전환) 진행(10-28 밤, 집)**: 1(public hostname `nexus` 삭제)·2(DNS `nexus` CNAME -> `irony1090.iptime.org` 회색) 완료, `nslookup nexus.irony1090.app 1.1.1.1` = iptime DDNS -> 집 공인 IP 확인. `direct` 레코드는 되돌아갈 길로 남김(이름 바꾸기 비추천). 다음 = 3(Caddyfile에 `nexus` 추가). 8 전용 명령 모음·상태 = user vault `user/home-server/주소-전환-명령.md`(cloudflared는 정지 + `Disabled`까지 — 서비스가 Automatic이라 정지만 하면 재부팅 때 다시 켜짐)
- 3(Caddyfile 수정) 완료(10-28 21:52): `nexus` 인증서 **http-01(80)로 발급** -> 80·443 둘 다 밖에서 열림 확인(`direct`는 tls-alpn-01이었음, Caddy가 방식을 번갈아 고름). 다음 = 4(접속 확인) -> 5(cloudflared 정지)
- 4(접속 확인)(10-28 22시): `nexus` 로그인 OK, trace = 다른 브라우저에선 안 열림(쓰던 브라우저는 옛 연결·DNS 기억으로 한동안 Cloudflare 경유 — `nslookup`은 OS 캐시를 안 거치므로 맞게 나와도 브라우저는 옛 답일 수 있음). 보드 online 확인 남음. 집 공인 IP = `125.143.1.21`(`history/home-server.md` 10-28 오후 외부 시험과 같은 값, 앞의 `.2`는 복사 잘림)
- **8(주소 전환) 완료(10-28 22시)**: 5(cloudflared 정지) = `Stopped`·`Disabled`(정지 대기 경고 여러 줄 = 연결 정리 대기, 정상). 남은 것: 보드 `imx8mp-f24` online 확인 · 며칠 뒤 `cloudflared service uninstall` + Zero Trust 터널 삭제
