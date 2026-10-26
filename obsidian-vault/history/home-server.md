# history — 집 Windows 서버 PC (supervisor + Cloudflare Tunnel)

> → `REF-home-server.md` / `CURRENT.md`

## 2026-10-25
- 사용자 질문(작업 밖에서 시작): 고정 IP 없는 집(iptime)에서 도메인 + https로 남는 PC를 서버로 -> DDNS+포트포워딩 vs Cloudflare Tunnel 비교, 무료 범위, 기본 주소 형태(trycloudflare·pages.dev·workers.dev), supervisor 적합성(코드 확인), Windows 10 가능 여부(supervisor Windows 컴파일 통과)
- 사용자 Cloudflare 가입, 서버 PC에 PostgreSQL 15 설치(Server + CLT, 로케일 DEFAULT) -> pg_hba 내부 대역 + 방화벽(관리자 권한) -> 작업 PC DBeaver 접속 OK -> `irony` 계정 + `nexus` DB 생성 완료
- 사용자: master 최신·8013·worker 동반(`IRONY-SERVER-WORKER`) -> Windows 빌드 묶음 생성(REF "빌드 묶음"). 서버 PC에서 start.ps1 실행 -> 접속 OK(사용자, Windows supervisor 첫 실행 성공) -> 도메인 Cloudflare에서 구매 -> `winget install Cloudflare.cloudflared`(설치 직후 창은 PATH 미반영, 새 창) -> 대시보드 터널 `home-server` + `service install <토큰>`(관리자) 연결됨 -> "공개 애플리케이션 추가" `nexus` + `http://localhost:8013` -> 외부 접속 OK. `/worker` 무인증 확인 -> Access 권장 -> 재부팅 시 supervisor·worker 안 켜짐 -> `build/nexus-autostart.ps1`(작업 스케줄러, 부팅 30초 뒤, 내 계정) 등록 -> 재부팅 후 접속 OK(첫 번은 원격 데스크톱 로그인 뒤 확인 -> 다시 재부팅해 원격 접속 없이 폰으로 OK = 로그인 전 시작 확인)
- Access는 나중(BACKLOG). worker_move 필요 + 계정 `irony1090`(회원가입 화면 없음) -> B 선택: 임시 worktree에서 master + supervisor-move 합쳐 Windows 업데이트 묶음(`update.ps1`), 계정 INSERT·권한 SQL 안내
- 업데이트 적용(`update.ps1`, 옛 파일 지움) -> 계정 SQL 파일 `build/nexus-create-user.sql` -> `irony1090` 생성·로그인 OK. 주소 변경 사용법 안내(둘 다 새 빌드여야, 옛 빌드 worker는 env 교체). 사용자: 운영 서버 구축 마무리, 내일 회사에서 worker 전부 최신 + 집 주소로 -> REF "내일 할 것"

## 2026-10-26 (회사)
- 사용자: 집 서버 배포 버전으로 Linux supervisor·worker 하나씩, 우선 test3(`localhost:8013`)에 붙이고 그 뒤 집 서버로 -> `build/supervisor_move` 빌드(태그 `SUPERVISORMOVE-261026`, REF "Linux B 빌드"). 실행은 사용자
- 사용자: test3에 띄운 뒤 화면으로 집 서버로 옮기기 **잘 됨** -> 파트(supervisor 주소 실시간 변경) 머지해도 됨(순서는 회원 관리 확인 뒤, `CURRENT.md`). 회원 관리 뒤 master 정식 빌드로 집 서버 재배포 가능(모든 worker 실행 파일 + `API_KEY`, `REF-member.md` "배포 전 확인")
- (오후) master에 `feat/member` 머지(`4ae97a7`). 사용자: 회사라 집 서버 원격 불가, Nexus 시스템만으로 배포 -> 함정-A(Job)·함정-B(키) 확인, 업데이트-1~5 흐름 합의(관리자 `irony1090`, 다른 worker는 전부 test3, 서버 PC 터미널 실행은 처음). 업데이트-1(사전 점검) 스크립트 `build/home-update/check.ps1` 작성(ASCII만, 로컬 문법 검사 도구 없음)
- 업데이트-1 결과 받음(전부 기대대로, REF). 롤백 안전 확인(옛 supervisor + v14 DB). master `4ae97a7` Windows 빌드 + 업데이트-3·4 스크립트 -> `build/home-update/nexus-member.zip`·`prepare.ps1`. 로컬 PowerShell 없어 문법 미검사
- 사용자 실행: 업로드 -> `prepare.ps1` -> GO -> 키 붙여넣기 -> `status.ps1` `state=DONE`(16:59:03). 집 서버 = master `4ae97a7`(회원 관리) 배포 완료, Nexus 터미널만으로 원격 업데이트 성공

## 2026-10-26 밤 — `/api` + 폰 수정 배포
- 묶음 `nexus-api-prefix.zip` 다시 만듦: 웹 = master `8de3e05`(`fix/phone-layout` 머지, test3 확인본과 같은 빌드) / `supervisor.exe` = `c5f9749` 빌드 그대로(Go 변경 없음)
- 사용자: `C:\Nexus\api_prefix`에 압축 해제 -> 관리자 PowerShell `update.ps1` -> 22:19 `OK: new supervisor is up on port 8013`. 남은 것 `*.old-20261026-221903`
