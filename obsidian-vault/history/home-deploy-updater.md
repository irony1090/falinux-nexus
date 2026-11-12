# history — 집 서버 배포: updater 이후

→ `REF-home-deploy-updater.md` · 11-03~11-04 `history/home-deploy-updater-setup.md` · 그 전 `history/home-deploy.md` · `CURRENT.md`

## 2026-11-05 ~ 11-06 보드 닫기 묶음
- 11-05 Claude가 `scripts/build.sh 261105-board-close_win supervisor worker frontend --os windows`(태그 `SUPERVISOR-`·`WORKER-BOARD-CLOSE-261105`) -> 묶음 `build/home-update/261105-board-close/`(67개, steps·nexus-steps = 261104-files 그대로). Chrome 확장 연결 안 돼 업로드 못 함(회사 PC Chrome은 다시 켬)
- 11-06 사용자가 Updater 화면에서 직접 올리고 배포. 공개 `index-CqlA8_4K.js` = 묶음과 같음(curl 확인)

## 2026-11-06 (저녁)
- 사용자 요청으로 집 서버 묶음 `build/home-update/261106-socket-reconnect/` 만듦(supervisor·worker·web, Windows, 태그 `SOCKET-RECONNECT-261106`) -> 업로드·배포 = 사용자
- 2026-11-07 소켓 재연결 묶음(`261106-socket-reconnect`) 사용자 배포. 공개 `index-AJJLr_Z9.js` = 묶음과 같음(nexus·direct curl 확인)

## 2026-11-07 (밤)
- test3 supervisor 교체 `261106-socket-reconnect` -> `261107-tile-shell`(태그 `SUPERVISOR-TILE-SHELL-261107`, swap.sh, 21:48 켜짐, 공개 `index-BIm258Bp.js`, worker 다시 붙음) · 집 서버 묶음 `build/home-update/261107-tile-shell/` 만듦(업로드·배포 = 사용자) → `REF-home-deploy-updater.md` "11-07 공용 타일 틀 묶음"
- 사용자 집 서버 배포 -> 공개 `index-BIm258Bp.js` 일치(nexus·direct)
- 11-07 사용자: 집 서버에 `updater-new` 폴더 없음 -> 할 일에서 뺌(vault엔 이름이 "예"로만 적혀 있었음, 11-04 updater 교체용 임시 업로드 폴더)

## 2026-11-08
- test3 supervisor 교체 `261107-tile-shell` -> `261108-settings-account`(태그 `SUPERVISOR-SETTINGS-ACCOUNT-261108`, swap.sh, 11:06:30 켜짐 pid 749414, `nexus_build` 마이그레이션 `00019` 적용, 공개 `index-Ba5qQ-fM.js`, worker·updater 다시 붙음) · 집 서버 묶음 `build/home-update/261108-settings-account/` 만듦(업로드·배포 = 사용자) → `REF-home-deploy-updater.md` "11-08 설정 두 겹 묶음"
- 사용자 질문: 묶음 `nexus-steps/`·`web/`을 갈아치우기로 해도 되나 -> 확인(지원 버전·거절 조건·staging 실행·되돌리기) 후 가능. **앞으로 묶음은 steps.json에 replace를 미리 넣음(A)**, 이번 묶음은 사용자가 직접 고침 → `REF-home-deploy-updater.md` "묶음 규칙: 갈아치우기"
- 사용자 집 서버 배포(설정 두 겹, replace 사용자가 넣음) -> 공개 `index-Ba5qQ-fM.js` 일치(nexus·direct) · 옛 assets 사라짐 = `web/` 갈아치우기 동작(없는 파일도 200 + HTML이라 content_type으로 판정)

## 2026-11-11
- 파트(칼럼 흐름) 마감 뒤 사용자 요청 "test3용 빌드 + 집 서버 updater 묶음": 태그 `SUPERVISOR-/WORKER-COLUMN-FLOW-261111` -> `scripts/build.sh 261111-column-flow`(Linux) · `261111-column-flow_win supervisor worker frontend --os windows` -> 묶음 `home-update/261111-column-flow`(67개, `replace` 넣음) → `REF-home-deploy-updater.md` "11-11 칼럼 흐름 묶음"
- test3 교체용 `build/261111-column-flow/swap.sh`(되돌림 = `261108-settings-account`) 만들어 둠 — 실행 안 함(8013 재기동은 사용자 확인 뒤)
- 사용자가 test3·집 서버 둘 다 교체 → 확인: 8013 = `build/261111-column-flow/supervisor`(14:49:30 켜짐)·`index-DYuchPlR.js` / 집 서버 공개 nexus·direct = `index-DYuchPlR.js`, 옛 `index-Ba5qQ-fM.js` = HTML

## 2026-11-12 계정 묶음
- 사용자 "test3용 빌드 + 집 서버 updater 빌드 파일" -> 태그 `ACCOUNT-261112` · `build/261112-account`(+ `swap.sh`, 되돌림 = `261111-column-flow`) · `build/261112-account_win` · 묶음 `build/home-update/261112-account`(65개) + 배포 전 확인 SQL `261112-account-check.sql` → `REF-home-deploy-updater.md` "11-12 계정 묶음"
- test3 교체(사용자 swap.sh) 확인: 8013 = `261112-account`, `nexus_build` 마이그레이션 `00021`, 아이디 빈 계정 0/9, `/api/users/sessions` 401, rollback.log 없음
- 집 서버 배포 전 확인(사용자가 `C:\Nexus`에서 `261112-account-check.ps1` 실행): users=1, 아이디 빌 계정 0 -> `RESULT: SAFE`. 배포 대기
- 사용자 집 서버 배포 -> Claude 확인: nexus·direct 공개 `index-Ca04n1MO.js`, `/api/users/sessions` 401, 옛 js `text/html`
