# REF — 집 서버 배포: updater 이후 (2026-11-03 신설)

> → 이력 `history/home-deploy-updater.md` · updater 도입·설치(11-03~11-04 묶음·updater 교체) `REF-home-deploy-updater-setup.md` · 그 전 수동 묶음(drag 등) `REF-home-deploy.md` · updater 구현 `REF-update-p4.md` · 진행 `CURRENT.md`
> 서버 PC 폴더 = `C:\Nexus`(supervisor `localhost:8013`, 작업 `Nexus`). 묶음 위치 = `build/home-update/`(git 제외)

## 묶음 규칙: 갈아치우기 (2026-11-08 사용자 "앞으로는 A로")
- **다음 묶음부터** `steps.json`의 `"uses"` 다음 줄에 `"replace": ["nexus-steps", "web"]`를 넣어 만든다(지난 묶음 것을 승계할 때도 이 줄이 있는지 확인)
- 근거: 집 서버 updater `UPDATER-FILES-261104`가 갈아치우기 지원 · 두 폴더 모두 updater 거절 조건에 안 걸림 · 단계 스크립트는 staging 사본으로 실행(`cmd/updater/job.go` `script`)이라 copy 중 `nexus-steps/`를 비워도 됨 · 지운 파일은 backup -> restore로 되돌아옴 · `web/`에 쌓인 옛 해시 assets 정리. 설계 → `REF-update-replace.md`
- 주의: 두 폴더에 묶음에 없는 파일을 손으로 넣었다면 지워짐(백업은 남음, `*.old-*`는 남김)

## 11-12 계정 묶음 (Claude가 만듦, 업로드·배포 = 사용자)
- `build/home-update/261112-account/`(파일 65개): `build/261112-account_win`의 `supervisor.exe`(`SUPERVISOR-ACCOUNT-261112`)·`worker.exe`(`WORKER-ACCOUNT-261112`, 코드 변경 없음·묶음 규칙상 포함)·`web/`(`index-Ca04n1MO.js` = test3용 `build/261112-account`와 같음) + 261111 묶음의 `steps.json`(`replace` 이미 있음)·`nexus-steps/` 그대로(diff 같음). updater 제외
- 내용 = 파트(계정)(`REF-account.md`) + 푸터 배지 잘림. **새 마이그레이션 `00020`(아이디 평문 이전)·`00021`(`user_sessions`)** — start 때 적용, **전원 한 번 재로그인**
- **배포 전 확인**: `build/home-update/261112-account-check.ps1`(사용자 요청으로 파일화, ASCII만 — `C:\Nexus\supervisor.env`로 psql 접속, 마지막 줄 `RESULT: SAFE/OK/STOP`, 이 PC엔 PowerShell이 없어 실행 시험 못 함) 또는 같은 내용 `.sql`(묶음 밖, 읽기만) -> 0행이면 안전. 관리자가 나오면 배포하지 말 것(아이디가 비어 아무도 못 고침)
- **rollback 주의**: `00020`이 아이디를 평문으로 바꿔서 **파일 복원(rollback)만으로는 옛 빌드에서 로그인이 안 됨**(옛 코드는 입력을 해시해서 비교) -> 되돌리려면 `01-dbdump` 덤프로 DB도 복원
- 배포 뒤 확인 = 공개 `index.html`이 `index-Ca04n1MO.js`를 가리키는지(nexus·direct) + `/api/users/sessions` = 401(새 API)
- **결과(11-12)**: 사용자가 배포(배포 전 확인 SAFE, 회원 1명). 공개 `index.html` = `index-Ca04n1MO.js`(nexus·direct 둘 다) · `/api/users/sessions` = 401 · 직전 `index-DYuchPlR.js` = `text/html`(갈아치우기 동작). Claude가 curl로 확인

## 11-11 칼럼 흐름 묶음 (Claude가 만듦, 업로드·배포 = 사용자)
- `build/home-update/261111-column-flow/`(파일 67개): `build/261111-column-flow_win`의 `supervisor.exe`(`SUPERVISOR-COLUMN-FLOW-261111`)·`worker.exe`(`WORKER-COLUMN-FLOW-261111`, 코드 변경 없음·묶음 규칙상 포함)·`web/`(`index-DYuchPlR.js` = test3용 빌드 `build/261111-column-flow`와 같음) + 261108 묶음의 `nexus-steps/` 그대로(diff 같음) + `steps.json`은 261108 것에 **`"replace": ["nexus-steps", "web"]`를 Claude가 넣음**(묶음 규칙: 갈아치우기 첫 적용). updater 제외
- 내용 = 칼럼 흐름(`REF-column-flow.md`, 프론트만). 11-08 이후 Go 코드 변경 없음·새 마이그레이션 없음(`00019` 그대로) -> rollback DB 복원 불필요
- 배포 뒤 확인 = 공개 `index.html`이 `index-DYuchPlR.js`를 가리키는지(nexus·direct, 옛 assets는 `content_type`으로 판정)
- **결과(11-11)**: 사용자가 배포. 공개 `index.html` = `index-DYuchPlR.js`(nexus·direct 둘 다) · 직전 빌드 `index-Ba5qQ-fM.js` = `text/html`(SPA 대체, 갈아치우기 동작). Claude가 curl로 확인

## 11-08 설정 두 겹 묶음 (Claude가 만듦, 업로드·배포 = 사용자)
- **이 묶음의 `replace`는 사용자가 업로드 전에 직접 넣음**(Claude는 안 고침, 11-08 사용자)
- `build/home-update/261108-settings-account/`(파일 67개): `build/261108-settings-account_win`의 `supervisor.exe`(`SUPERVISOR-SETTINGS-ACCOUNT-261108`)·`worker.exe`(`WORKER-SETTINGS-ACCOUNT-261108`, 코드 변경 없음·묶음 규칙상 포함)·`web/`(`index-Ba5qQ-fM.js` = test3와 같음) + 261107 묶음의 `steps.json`·`nexus-steps/` 그대로(diff 같음). updater 제외
- 내용 = 설정 두 겹(`REF-settings-account.md`). **새 마이그레이션 `00019`(`user_settings`)** — start 때 supervisor가 적용. 추가만 하는 테이블이라 rollback(파일 복원)만으로 옛 빌드가 돎, DB 복원 불필요(dbdump는 평소대로 남음)
- 배포 뒤 확인 = 공개 `index.html`이 `index-Ba5qQ-fM.js`를 가리키는지(nexus·direct)
- **결과(11-08)**: 사용자가 배포(replace 직접 넣음). 공개 `index.html` = `index-Ba5qQ-fM.js`(nexus·direct 둘 다) · `/api/settings` = 401(새 API 있음) · **`web/` 갈아치우기 확인**: 옛 빌드 `index-BIm258Bp.js`·`index-AJJLr_Z9.js`·`index-L-Np3jCk.js` = SPA 대체 HTML(1,731B `text/html`, 없는 파일과 같음), 새 것만 JS. 확인 함정: 없는 assets도 200으로 오므로 상태 코드 말고 `content_type`으로 판정

## 11-07 공용 타일 틀 묶음 (Claude가 만듦, 업로드·배포 = 사용자 11-07 끝)
- `build/home-update/261107-tile-shell/`(파일 67개): `build/261107-tile-shell_win`의 `supervisor.exe`(`SUPERVISOR-TILE-SHELL-261107`)·`worker.exe`(`WORKER-TILE-SHELL-261107`, 코드 변경 없음·묶음 규칙상 포함)·`web/`(`index-BIm258Bp.js` = test3와 같음) + 261106 묶음의 `steps.json`·`nexus-steps/` 그대로(diff 같음). updater 제외. Go 코드 변경 없음·새 마이그레이션 없음 -> rollback DB 복원 불필요
- 내용 = 공용 타일 틀·메뉴 줄 위치/숨김·설정 창·푸터 오른쪽 메뉴(`INDEX-tile-frame.md`). 배포 뒤 확인 = 공개 `index.html`이 `index-BIm258Bp.js`를 가리키는지
- **결과(11-07 밤)**: 사용자가 배포. 공개 `index.html` = `index-BIm258Bp.js`(nexus·direct 두 주소 모두, 묶음과 같음, Claude가 curl로 확인)

## 11-06 소켓 재연결 묶음 (Claude가 만듦, 업로드·배포 = 사용자 11-07 끝)
- `build/home-update/261106-socket-reconnect/`(파일 67개): `build/261106-socket-reconnect_win`의 `supervisor.exe`(`SUPERVISOR-SOCKET-RECONNECT-261106`)·`worker.exe`(`WORKER-SOCKET-RECONNECT-261106`, 코드 변경 없음·묶음 규칙상 포함)·`web/`(`index-AJJLr_Z9.js` = test3와 같음) + 261105 묶음의 `steps.json`·`nexus-steps/` 그대로. updater 제외. 새 마이그레이션 없음(00018) -> rollback DB 복원 불필요
- 내용 = 소켓 재연결·끊김 표시(`REF-socket-reconnect.md`). 배포 뒤 확인 = 공개 `index.html`이 `index-AJJLr_Z9.js`를 가리키는지
- **결과(11-07)**: 사용자가 배포. 공개 `index.html` = `index-AJJLr_Z9.js`(nexus·direct 두 주소 모두, 묶음과 같음, Claude가 curl로 확인)

## 11-05 보드 닫기 묶음 (Claude가 만듦, 업로드·배포 = 사용자 11-06 끝)
- `build/home-update/261105-board-close/`(파일 67개): `build/261105-board-close_win`의 `supervisor.exe`(`SUPERVISOR-BOARD-CLOSE-261105`)·`worker.exe`(`WORKER-BOARD-CLOSE-261105`)·`web/` + 261104-files의 `steps.json`·`nexus-steps/` 그대로. updater 제외. 새 마이그레이션 없음(00018) -> rollback DB 복원 불필요
- 사용자 요청 = Updater 화면에서 새 빌드 만들고 파일 업로드까지(배포는 안 함). 11-05 Claude in Chrome 확장 연결 안 됨(Chrome은 다시 켰음) -> 업로드 대기
- **결과(11-06)**: 사용자가 Updater 화면에서 직접 올리고 배포. 공개 `index.html` = `index-CqlA8_4K.js`(묶음과 같음, Claude가 curl로 확인). 집 서버 supervisor·worker·web = `BOARD-CLOSE-261105`, updater = `UPDATER-FILES-261104` 그대로
