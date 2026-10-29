# history — 집 서버 배포·업데이트 (2026-10-28 `history/home-server.md`에서 분할)

> → `REF-home-deploy.md` / 구성 이력 `history/home-server.md` / `CURRENT.md`

## 2026-10-26 (회사)
- 사용자: 집 서버 배포 버전으로 Linux supervisor·worker 하나씩, 우선 test3(`localhost:8013`)에 붙이고 그 뒤 집 서버로 -> `build/supervisor_move` 빌드(태그 `SUPERVISORMOVE-261026`, REF "Linux B 빌드"). 실행은 사용자
- 사용자: test3에 띄운 뒤 화면으로 집 서버로 옮기기 **잘 됨** -> 파트(supervisor 주소 실시간 변경) 머지해도 됨(순서는 회원 관리 확인 뒤, `CURRENT.md`). 회원 관리 뒤 master 정식 빌드로 집 서버 재배포 가능(모든 worker 실행 파일 + `API_KEY`, `REF-member.md` "배포 전 확인")
- (오후) master에 `feat/member` 머지(`4ae97a7`). 사용자: 회사라 집 서버 원격 불가, Nexus 시스템만으로 배포 -> 함정-A(Job)·함정-B(키) 확인, 업데이트-1~5 흐름 합의(관리자 `irony1090`, 다른 worker는 전부 test3, 서버 PC 터미널 실행은 처음). 업데이트-1(사전 점검) 스크립트 `build/home-update/check.ps1` 작성(ASCII만, 로컬 문법 검사 도구 없음)
- 업데이트-1 결과 받음(전부 기대대로, REF). 롤백 안전 확인(옛 supervisor + v14 DB). master `4ae97a7` Windows 빌드 + 업데이트-3·4 스크립트 -> `build/home-update/nexus-member.zip`·`prepare.ps1`. 로컬 PowerShell 없어 문법 미검사
- 사용자 실행: 업로드 -> `prepare.ps1` -> GO -> 키 붙여넣기 -> `status.ps1` `state=DONE`(16:59:03). 집 서버 = master `4ae97a7`(회원 관리) 배포 완료, Nexus 터미널만으로 원격 업데이트 성공

## 2026-10-26 밤 — `/api` + 폰 수정 배포
- 묶음 `nexus-api-prefix.zip` 다시 만듦: 웹 = master `8de3e05`(`fix/phone-layout` 머지, test3 확인본과 같은 빌드) / `supervisor.exe` = `c5f9749` 빌드 그대로(Go 변경 없음)
- 사용자: `C:\Nexus\api_prefix`에 압축 해제 -> 관리자 PowerShell `update.ps1` -> 22:19 `OK: new supervisor is up on port 8013`. 남은 것 `*.old-20261026-221903`

## 2026-10-27 밤 — mochi 묶음
- 사용자 요청(집에서 직접 올림): master `28053b7` Windows 빌드 -> `build/home-update/nexus-mochi.zip`. worker 포함(사용자: 집 서버 = 회사 업데이트 전 버전과 같아 worker도 필요). 배포 `OK`(사용자, 10-27 밤) → REF-home-deploy "mochi 배포 묶음"

## 2026-10-28 — 웹만 교체
- 사용자: 웹 빌드 zip을 Nexus로 집 서버에 올림 -> 압축 풀기·폴더 바꿔치기 명령 안내 -> 교체 완료(사용자) → REF-home-deploy "웹만 교체"

## 2026-10-28 밤 — viewer-image 묶음
- 사용자 요청: 오늘 뷰어 기능 포함 최신으로, 이름 viewer-image, worker·supervisor 둘 다 -> master `709095c` Windows 빌드 -> `build/home-update/nexus-viewer-image.zip`. web은 새로 빌드(test2 주소 안 들어감 확인). 배포 완료(사용자, 10-28 밤) → REF-home-deploy "viewer-image 배포 묶음"
- 같은 밤: 집 서버 "편집" 거절 메시지(EDITOR 안내) 확인 -> `winget install Microsoft.Edit` + worker 재시작(`Start-ScheduledTask Nexus`) -> 정상(사용자). 명령 = user vault `user/home-server/편집기-설치.md`
