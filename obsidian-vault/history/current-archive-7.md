# history — CURRENT에서 내려보낸 줄 원문 (2026-11-05~)

> → `CURRENT.md` · 앞 파일 `history/current-archive-6.md` · REF 없음(원문 보관용)

## 11-05 보드 닫기 마감 때 내려보냄

- **(2026-11-04) 빌드 스크립트 대상별로 — 마감, `feat/build-scripts` 커밋·master `--no-ff` 머지(push = 사용자)** · 사용자가 집 서버 빌드(`build/261104-files_win`)로 실사용 확인. `scripts/build.sh <폴더명> [대상...] [--os] [--arch]`(updater는 대상에 적어야 빌드) → `REF-build-scripts.md` · `history/build-scripts.md` 11-04
- **(2026-11-04) 로컬 updater 시험 + 주소 순위 눌러 입력 + 되돌리기 순서(안 A) — 마감, `feat/addr-rank-pick` 커밋·master `--no-ff` 머지(push = 사용자)** · 사용자 화면 확인 끝 → `history/update-rules.md` 11-04 · `history/supervisor-move.md` 11-04
  - 주소 순위 눌러 입력 = 순위 줄(밑줄 글자) 누르면 방식·호스트 입력칸에 → `REF-supervisor-move.md` "순위 눌러 입력"(Vuetify reset 함정 포함)
  - 되돌리기 = **최신만(역순)**, 이미 되돌린 배포 거절, 확인 창 = "배포 직전 상태로, 그 버전으로 가는 게 아님" → `REF-update-rules.md` "되돌리기"
  - **지금 켜져 있는 것**: 로컬 updater(11-04 11:5x 꺼짐) — 다시 = `apps/core`에서 `go run ./cmd/updater -c ~/upd-test/updater.json`(인스턴스 `dev-main#mzOwVeBF` -> test3 `ws://localhost:8013`, `targetDir` = `build`, 설치본 `test-1`). `updater.json`(키)·`state.json` = `~/upd-test/`로 옮김, `apps/core/.gitignore`에 `/cmd/updater/` 실행 흔적 5개 추가(11-04) / test3 = `build/261104-files`(11-04 11:39 새 supervisor, 보관소 = `build/update-store`(env `UPDATE_STORE_DIR="../update-store"`))
  - 시험 흔적(nexus_build): 장비 루트 `dev-main`, 빌드 `supervisor_url_click-261104`·`-01`·`test-1`·`test-2`, 배포 #1~#7. `build/261104-files/web/assets`에 직접 옮긴 옛 해시 파일 몇 개(안 씀)
- **(2026-11-03) 파트(업데이트 app) — P1(보관·화면) · P2(updater Linux) · P3(배포) · P4(Windows) 마감, master 머지 끝**(P1 `1f5eda9` · P2 `bd93fa4` · P3 = `feat/updater-p3` · P4 = `feat/updater-p4` 머지, push = 사용자). 사용자 P3 화면 확인 "나쁘지 않음"(11-03). 옛 줄 원문 → `history/current-archive-6.md` 11-03 절
  - 구현 지도: P1 `REF-update-p1.md` · P2 `REF-update-p2.md` · P3 `REF-update-p3.md` / 설계 `REF-update-deploy.md` · 규칙 `REF-update-rules.md`
  - P4(Windows) = 구현 · VM 확인 39/39 · 집 서버 설치(11-03 12:42 `C:\Nexus\updater` 동작) → `REF-update-p4.md` · 집 서버 → `REF-home-deploy-updater.md`(관리자 = irony1090)
  - **그다음 = 단위-7(교체 빌드)**(Files 타일 개선 끝, 정-1~3 답 대기 그대로): Updater 화면에서 집 서버 supervisor·worker를 updater로 교체(코드 없음, 보관소 빌드). **(11-03 밤) 착수 — updater 교체는 나중, 정-1(old 정리 단계)·정-2(DB 백업 자리) = 제안대로, 정-3(관찰 방식) = 회사에서 지켜봄. 묶음 `build/home-update/files-tile/` -> 11-03 밤 사용자가 집 서버 updater로 배포 완료, "잘 동작해"(첫 updater 배포 = 단위-7 끝). 집 서버 updater 교체 = **11-04 14:33 끝**(`UPDATER-FILES-261104`, 위 11-04 줄) · `nexus-before-updater.dump` 삭제(사용자)** → `REF-home-deploy-updater.md` "files-tile 묶음" → `REF-update-p4.md` "단위-7(교체 빌드) 후속". 답을 받으면 `steps.json`·스크립트 계획을 먼저 보여 줄 것
  - **(11-03 밤) Files 타일 개선 마감** — `feat/update-files-tile` 코드 커밋 `3caff7b` + vault 커밋 + master `--no-ff` 머지(push = 사용자). 단위-가(폴더 트리·선택 삭제)·시험 실행·단위-나(끌어 놓기)·단위-다(갈아치우기)·단위-라(실행 권한 이름) 전부 사용자 화면 확인 끝 → `REF-update-files-tile.md` · `REF-update-replace.md` · `REF-update-trial.md`. 집 서버 updater 새 빌드 교체 = 11-04 끝(이제 replace 먹음). 진행 블록 원문 → `history/current-archive-6.md` 11-03 밤 절
  - **지금 켜져 있는 것**: VM `nexus-win` + updater `test-win#BGRB5DdV`(test2, `C:\NexusU`, 예약 작업 아님) → `REF-windows-vm.md` 11-03 줄, 다 쓰면 `docker stop nexus-win` / 개발 서버 **8011(vite) = 11-03 밤 사용자 요청으로 끔**(다시 = `apps/frontend`에서 `npm run dev`) · 8012(supervisor `go run`, 11-03 밤 단위-다 코드로 다시 띄움 — 이전 메모: 11-03 16:39 이 세션이 `setsid`로 다시 띄움 — 시험-1 라우트 반영, 세션이 끝나도 남음, 로그는 스크래치라 사라짐) / 8013 빌드본. 시험 확인용 ssh 래퍼 `vssh`는 세션 스크래치라 사라짐(`REF-test-scripts-1014.md` "윈도-6"에서 다시 만듦)
  - **사용자 확인 대기**: DeepSeek API 키 재발급 + `ds.sh` 인자 노출 개선 여부 → `BACKLOG.md`
  - 집 서버 남은 정리(사용자가 지움): `*.old-20261103-123559` 3개(지워도 된다고 답함) · `nexus-before-updater.dump`(단위-7 `pg_dump` 뒤)
  - 로컬 VM `nexus-win`에 시험 흔적 `C:\NexusT`·`C:\NexusS` 남음(지워도 됨). 시험 스크립트(`upd_win*.py`)는 세션 스크래치라 사라짐 — 다시 필요하면 `REF-update-p4.md` "확인" 표로 다시 씀
  - 실기 장비: `imx8mp-f24` 서비스 `nexus-updater`(P3 실행 파일, 태그는 P2 이름 그대로 — 사용자: 중요하지 않음)가 test2에 붙어 있음. 개발 DB `nexus`·8013 빌드본 `nexus_build` 둘 다 마이그레이션 `00018`까지(11-04 확인)
  - 남은 흔적: 로컬 시험 updater 인스턴스 `sample-jig#asWTzHHK` 행 offline(지워도 됨)
  - 새 의존성: 프론트 `@codemirror/*`·`@lezer/highlight`(`npm i` 필요), 메인 보드 스크립트 편집도 CodeMirror(`CtCode`)로 바뀜. `tsconfig.app.json` `baseUrl` 제거(TS 6 경고 TS5101)
  - 시안(확정) = Claude 아티팩트 https://claude.ai/artifact/XGMJqjg7rHdAHUhRMXHbZa(v15). 시안·화면을 고칠 땐 먼저 `REF-design-chaltteok.md` "시안에서 정리한 규칙(2026-11-02)" 읽기(묶음 쪼개지 말 것). 시안 확인법 = headless Chrome 스크린샷

### 옛 "지금 떠 있는 것" 줄 · 화면 확인용 줄

- **지금 떠 있는 것(11-04 확인)**: test3(8013) = supervisor·worker **`build/261104-files`**(사용자 최신 빌드, DB `nexus_build` 마이그레이션 `00018`까지, 보관소 = `build/261104-files/update-store`). 보관 프로세스 하나(59860)는 실행 파일이 `build/261103_updater/worker` — 그 터미널을 닫기 전엔 지우지 말 것. **11-04 아침 test1(8011)·test2(8012) 둘 다 꺼져 있음.** 옛 메모: test1 = 11-03 밤 끔 · test2(8012, 개발 supervisor `go run`) Claude 세션이 다시 띄움 — 세션을 다시 시작하면 같이 꺼질 수 있음(`ss -ltnp | grep -E ':801[12]'`로 확인, 다시 띄우기 = `apps/core`에서 `go run ./cmd/supervisor -env dev.env` · `apps/frontend`에서 `npm run dev`)** — 개발 DB `nexus`에 마이그레이션 `00018`까지 적용됨, 보관소 = `apps/core/cmd/supervisor/update-store`(git 제외). 5051·3001은 꺼져 있음
  - **화면 확인용으로 띄운 것(11-05 09:07)**: 8012 supervisor `go run -env dev.env` · 8011 vite(`.env`를 안 고치고 실행 환경변수 `VITE_API_URL=https://test2.falinux.com`으로 띄움 — `.env`에선 그 줄이 주석이라 그냥 띄우면 화면이 API를 못 찾음) · worker `devtest-pc`(env·sqlite = 세션 스크래치 `dev/`). 테스트 계정 `devtest`(비밀번호 = 스크래치 `dev/creds.txt`, vault에 안 적음) + 기본 보드에 장비 폴더 타일. 접속 = https://test1.falinux.com

## 2026-11-07 밤 세션 끝에 내려보냄
- **(2026-11-07) 파트(공용 타일 틀) — 마감, `refactor/tile-shell` 커밋·master `--no-ff` 머지(push = 사용자)**
  - 끝: 단계-1(공용 틀 리팩터링)·단계-2(메뉴 줄 위치·숨김) 구현·확인 + 사용자 화면 확인("스크롤에 따른 숨김/보임 확인") → `REF-tile-frame.md`
  - 끝: 단계-3·4 조율 Q14(로그아웃 때 설정)~Q23(설정 창 적용) + 시안(https://claude.ai/artifact/KCroqig1AG5hHjq77h2YTY, 고치지 않음 — Q21·Q22로 달라진 점은 REF에)
  - 끝(11-07 밤): 단계-3(설정 저장소·dialog)·단계-4(오른쪽 메뉴 통일) 구현 + 헤드리스 0 FAIL · `run.sh` 낡은 2건 외 통과. 착수 때 Q22(관리 페이지 알림) = 없음 · Q21(전송 자리) = 알림과 같은 레벨로 개정. 계획과 다르게 한 것 구현-1(설정 창 방식)~구현-4(폰 펼침 모양) → `REF-settings-footer.md` "구현 결과"
  - 사용자 화면 확인 "잘 동작해"(11-07 밤) -> **파트 마감**(커밋·master `--no-ff` 머지, push = 사용자) → `INDEX-tile-frame.md`(커밋·머지, MEMORY에 불변 결정 한 줄: 푸터 오른쪽 = `FootAccount` 공용 + 보드 전용 slot / 새 설정 = `settings.store` `SPECS`)
  - 확인 스크립트 `run.sh`의 FAIL 2건 = 낡은 기대값(→ `BACKLOG.md`), 이번 작업과 무관
- **(2026-11-06) 소켓 재연결 + 끊김 표시 — 마감, `fix/socket-reconnect` 커밋·master `--no-ff` 머지(push = 사용자)** · 사용자 모바일 확인 "잘 동작해"(test3 `261106-socket-reconnect`). 재연결-1(포기 없음)·재연결-2(화면 복귀·`online` 때 바로)·재연결-3(복귀 때 PING 3초)·화면-1(다시 붙으면 스냅샷으로 다시 그림)·끊김 알약. 못 본 것 = supervisor 재시작 뒤 화면-1. 집 서버 묶음 = `build/home-update/261106-socket-reconnect/`(11-06 만듦, 업로드·배포 = 사용자 대기 → `REF-home-deploy-updater.md` "11-06 소켓 재연결 묶음") → `REF-socket-reconnect.md` · `history/socket-reconnect.md`
- **(2026-11-05) 보드 닫기 + 종료 유예(S-b) — 마감, `feat/board-close` 커밋·master `--no-ff` 머지(push = 사용자)** · 사용자 화면 확인 "잘 동작해". 루트 타일 ×(기본 보드 제외)·보드 메뉴 "보드 삭제" = 같은 확인 창(취소 / 기본 보드로 보내기 / 모두 종료하고 닫기). 강제 종료 = 종료 요청 10초 뒤 서버 판정(2초 주기 검사, 남은 요청 없으면 쉼 = 안 A(메모리 표시)). 못 본 것 = 보기 권한 공유 터미널 → `REF-board-close.md` · `history/board-close.md`
- **(2026-11-07) 집 서버 소켓 재연결 묶음 배포 끝** — 공개 `index-AJJLr_Z9.js` 일치(nexus·direct) → `REF-home-deploy-updater.md` "11-06 소켓 재연결 묶음"
- **(2026-11-06) 집 서버 = 보드 닫기 빌드(`BOARD-CLOSE-261105`)** · 사용자가 Updater로 직접 업로드·배포, 공개 화면 해시 일치 확인 → `REF-home-deploy-updater.md` "11-05 보드 닫기 묶음"
