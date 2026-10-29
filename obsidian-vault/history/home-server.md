# history — 집 Windows 서버 PC (supervisor + Cloudflare Tunnel)

> → `REF-home-server.md` / 배포·업데이트 이력 `history/home-deploy.md` / 직통 전환 이력 `history/home-direct.md` / `CURRENT.md`

## 2026-10-25
- 사용자 질문(작업 밖에서 시작): 고정 IP 없는 집(iptime)에서 도메인 + https로 남는 PC를 서버로 -> DDNS+포트포워딩 vs Cloudflare Tunnel 비교, 무료 범위, 기본 주소 형태(trycloudflare·pages.dev·workers.dev), supervisor 적합성(코드 확인), Windows 10 가능 여부(supervisor Windows 컴파일 통과)
- 사용자 Cloudflare 가입, 서버 PC에 PostgreSQL 15 설치(Server + CLT, 로케일 DEFAULT) -> pg_hba 내부 대역 + 방화벽(관리자 권한) -> 작업 PC DBeaver 접속 OK -> `irony` 계정 + `nexus` DB 생성 완료
- 사용자: master 최신·8013·worker 동반(`IRONY-SERVER-WORKER`) -> Windows 빌드 묶음 생성(REF "빌드 묶음"). 서버 PC에서 start.ps1 실행 -> 접속 OK(사용자, Windows supervisor 첫 실행 성공) -> 도메인 Cloudflare에서 구매 -> `winget install Cloudflare.cloudflared`(설치 직후 창은 PATH 미반영, 새 창) -> 대시보드 터널 `home-server` + `service install <토큰>`(관리자) 연결됨 -> "공개 애플리케이션 추가" `nexus` + `http://localhost:8013` -> 외부 접속 OK. `/worker` 무인증 확인 -> Access 권장 -> 재부팅 시 supervisor·worker 안 켜짐 -> `build/nexus-autostart.ps1`(작업 스케줄러, 부팅 30초 뒤, 내 계정) 등록 -> 재부팅 후 접속 OK(첫 번은 원격 데스크톱 로그인 뒤 확인 -> 다시 재부팅해 원격 접속 없이 폰으로 OK = 로그인 전 시작 확인)
- Access는 나중(BACKLOG). worker_move 필요 + 계정 `irony1090`(회원가입 화면 없음) -> B 선택: 임시 worktree에서 master + supervisor-move 합쳐 Windows 업데이트 묶음(`update.ps1`), 계정 INSERT·권한 SQL 안내
- 업데이트 적용(`update.ps1`, 옛 파일 지움) -> 계정 SQL 파일 `build/nexus-create-user.sql` -> `irony1090` 생성·로그인 OK. 주소 변경 사용법 안내(둘 다 새 빌드여야, 옛 빌드 worker는 env 교체). 사용자: 운영 서버 구축 마무리, 내일 회사에서 worker 전부 최신 + 집 주소로 -> REF "내일 할 것"

## 2026-10-28 — cloudflared 업데이트 메일
- 사용자: "인증서가 취약한 개인키 사용 중, 업데이트하라" 메일 -> Nexus로 원격 업데이트 방법 질문. 확인(Nexus 터미널): `cloudflared --version` = 2026.10.3, 서비스 Running·Automatic, 대시보드 커넥터 2026.10.3. 공지(eMudhra) 기준 2026.10.2 이상이라 조치 불필요 → REF-home-server "cloudflared 보안 공지". 최신 10.5는 집에 있을 때 선택

## 2026-10-28 터널 속도 질문(`--protocol quic` + `--token-file`)
- 사용자가 찾은 ImagePath 변경 명령 검토 -> 속도 효과 거의 없음·token 파일 없으면 서비스 시작 실패 위험. 판단 근거 → `REF-home-server.md` "터널 속도". 적용 여부 미정(사용자)
- 같은 날: 증상 = 첫 로딩·타자 지연. 코드상 입력 지연 없음 확인, `colo=LAX` 측정 + 회사 supervisor에 붙이면 빠름 -> 원인 = Cloudflare 엣지 배정(LAX). 대안 비교 중 → REF "터널 속도"
- 같은 날: 대안 비교(안 A Tailscale / 안 B 포트포워딩 / 안 C 유지) -> 사용자가 안 B 쪽 질문 진행(HTTPS = Caddy, 유동 IP = 이미 있는 iptime DDNS). 계획(안) → `REF-home-direct.md` "진행 기록"(2026-10-28 분할)
- 같은 날: DNS `direct` 추가 완료(3단계). Caddy 설치 = PowerShell zip 직접 받기. 사용자 요청으로 user vault에 체크리스트 `user/home-server/직통-전환.md` 작성 + user vault 이미지 주제별 폴더 정리(`board-keyboard`·`phone-bug-1026`·`ct-field`·`term-ime`·`home-server`, `vrKeyboad` -> `vrKeyboard` 철자 수정), Claude vault 참조 11개 파일 경로 갱신
- 같은 날 오후: 회사에서 4(방화벽)·Caddyfile 전역 설정 판(storage·log)·7(서비스 `caddy` 수동 등록, Stopped) 완료. 외부 접속 시험 = `125.143.1.21:443` 시간 초과(포트포워딩 전, 예상대로). 설명: Caddy 역할(리버스 프록시 + 인증서), CNAME 조회 순서, `nexus` 구름만 회색 금지. **사용자: 오늘 밤 집에서 1·2·6 진행 예정**
