# history — 집 Windows 서버 PC (supervisor + Cloudflare Tunnel)

> → `REF-home-server.md` / 배포·업데이트 이력 `history/home-deploy.md` / `CURRENT.md`

## 2026-10-25
- 사용자 질문(작업 밖에서 시작): 고정 IP 없는 집(iptime)에서 도메인 + https로 남는 PC를 서버로 -> DDNS+포트포워딩 vs Cloudflare Tunnel 비교, 무료 범위, 기본 주소 형태(trycloudflare·pages.dev·workers.dev), supervisor 적합성(코드 확인), Windows 10 가능 여부(supervisor Windows 컴파일 통과)
- 사용자 Cloudflare 가입, 서버 PC에 PostgreSQL 15 설치(Server + CLT, 로케일 DEFAULT) -> pg_hba 내부 대역 + 방화벽(관리자 권한) -> 작업 PC DBeaver 접속 OK -> `irony` 계정 + `nexus` DB 생성 완료
- 사용자: master 최신·8013·worker 동반(`IRONY-SERVER-WORKER`) -> Windows 빌드 묶음 생성(REF "빌드 묶음"). 서버 PC에서 start.ps1 실행 -> 접속 OK(사용자, Windows supervisor 첫 실행 성공) -> 도메인 Cloudflare에서 구매 -> `winget install Cloudflare.cloudflared`(설치 직후 창은 PATH 미반영, 새 창) -> 대시보드 터널 `home-server` + `service install <토큰>`(관리자) 연결됨 -> "공개 애플리케이션 추가" `nexus` + `http://localhost:8013` -> 외부 접속 OK. `/worker` 무인증 확인 -> Access 권장 -> 재부팅 시 supervisor·worker 안 켜짐 -> `build/nexus-autostart.ps1`(작업 스케줄러, 부팅 30초 뒤, 내 계정) 등록 -> 재부팅 후 접속 OK(첫 번은 원격 데스크톱 로그인 뒤 확인 -> 다시 재부팅해 원격 접속 없이 폰으로 OK = 로그인 전 시작 확인)
- Access는 나중(BACKLOG). worker_move 필요 + 계정 `irony1090`(회원가입 화면 없음) -> B 선택: 임시 worktree에서 master + supervisor-move 합쳐 Windows 업데이트 묶음(`update.ps1`), 계정 INSERT·권한 SQL 안내
- 업데이트 적용(`update.ps1`, 옛 파일 지움) -> 계정 SQL 파일 `build/nexus-create-user.sql` -> `irony1090` 생성·로그인 OK. 주소 변경 사용법 안내(둘 다 새 빌드여야, 옛 빌드 worker는 env 교체). 사용자: 운영 서버 구축 마무리, 내일 회사에서 worker 전부 최신 + 집 주소로 -> REF "내일 할 것"

## 2026-10-28 — cloudflared 업데이트 메일
- 사용자: "인증서가 취약한 개인키 사용 중, 업데이트하라" 메일 -> Nexus로 원격 업데이트 방법 질문. 확인(Nexus 터미널): `cloudflared --version` = 2026.10.3, 서비스 Running·Automatic, 대시보드 커넥터 2026.10.3. 공지(eMudhra) 기준 2026.10.2 이상이라 조치 불필요 → REF-home-server "cloudflared 보안 공지". 최신 10.5는 집에 있을 때 선택
