# CURRENT

## 현재 날짜
2026-11-06

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **(2026-11-06) 소켓 재연결 + 끊김 표시 — 마감, `fix/socket-reconnect` 커밋·master `--no-ff` 머지(push = 사용자)** · 사용자 모바일 확인 "잘 동작해"(test3 `261106-socket-reconnect`). 재연결-1(포기 없음)·재연결-2(화면 복귀·`online` 때 바로)·재연결-3(복귀 때 PING 3초)·화면-1(다시 붙으면 스냅샷으로 다시 그림)·끊김 알약. 못 본 것 = supervisor 재시작 뒤 화면-1. 집 서버 묶음 = `build/home-update/261106-socket-reconnect/`(11-06 만듦, 업로드·배포 = 사용자 대기 → `REF-home-deploy-updater.md` "11-06 소켓 재연결 묶음") → `REF-socket-reconnect.md` · `history/socket-reconnect.md`
- **(2026-11-05) 보드 닫기 + 종료 유예(S-b) — 마감, `feat/board-close` 커밋·master `--no-ff` 머지(push = 사용자)** · 사용자 화면 확인 "잘 동작해". 루트 타일 ×(기본 보드 제외)·보드 메뉴 "보드 삭제" = 같은 확인 창(취소 / 기본 보드로 보내기 / 모두 종료하고 닫기). 강제 종료 = 종료 요청 10초 뒤 서버 판정(2초 주기 검사, 남은 요청 없으면 쉼 = 안 A(메모리 표시)). 못 본 것 = 보기 권한 공유 터미널 → `REF-board-close.md` · `history/board-close.md`
- **(2026-11-06) 집 서버 = 보드 닫기 빌드(`BOARD-CLOSE-261105`)** · 사용자가 Updater로 직접 업로드·배포, 공개 화면 해시 일치 확인 → `REF-home-deploy-updater.md` "11-05 보드 닫기 묶음"
- **(2026-11-04) 집 서버 최신 배포 — 끝(14:33)** · supervisor·worker·web = updater 배포 #2 성공(빌드 `supervisor-url-261104` 잠김), updater = Nexus 디렉토리 타일에서 `setup.ps1` -> `UPDATER-FILES-261104` online. 남은 것 = `C:\Nexus\updater-new\` 삭제(사용자) → `REF-home-deploy-updater.md` "11-04 최신 배포 계획"
- 11-04 빌드 스크립트·로컬 updater 시험(주소 순위·되돌리기) 마감 줄 + 11-03 파트(업데이트 app) 진행 블록 원문 → `history/current-archive-7.md`(11-05 내려보냄). 구현 지도: `REF-update-p1.md`~`-p4.md` · 설계 `REF-update-deploy.md` · 규칙 `REF-update-rules.md` · `REF-build-scripts.md`
- **(2026-11-02) 10-30 작업 점검·수정 마감** — `fix/member-drag-review` master 머지. 코드 항목 수정 + 사용자 실기기 확인 끝 → `REF-review-1030.md` · `history/review-1030.md`. 확인 스크립트 = `scripts/check/run.sh`
  - **아직 안 한 것**: 새 빌드로 색인 금지 배포(코드-6(색인 금지 미배포)) = 사용자. **(11-03 사용자) 파트(업데이트 app)가 끝난 뒤 한 번에 배포**(updater 마이그레이션 `00016`~ 함께). `cmd/*/env` 태그를 바꿔 빌드, 집 서버 묶음은 worker 포함. 8013·집 서버는 지금 옛 빌드(10-30 이후 수정 없음)
- (2026-10-30) 파트(내 장비·관리자 타일 이동) 마감 → `INDEX-member-drag.md`. 남은 것 = `openParent`가 자식 묶음 중간에 낄 수 있음(옛 동작, 안 고침)
- (2026-10-29) 파트(타일 끌기) 마감 → `INDEX-tile-drag.md`
- (2026-10-29) 파트(영상 Range) 마감 → `INDEX-video-view.md`
- **(2026-10-29 저녁) 집 서버 drag 묶음 배포 완료**(`build/home-update/nexus-drag.zip`, 태그 `DRAG-HOMESERVER-261029`) → `REF-home-deploy.md` "drag 배포 묶음"
- **(2026-10-29) 보드 `fa24-PID`(192.168.40.228) worker = `VIEWER_IMAGE-261028`** → `REF-build.md` "보드 worker 교체"
- 옛 CURRENT 줄 원문 → `history/current-archive-6.md`(11-02 내려보냄: 업데이트 app 10-30 설계 상태·11-02 재검토 중 줄), 그 전 = `-5`(10-29 파트 마감·떠 있는 포트 표·build 흔적), 그 전 = `-4`(10-27 ~ 10-28)·`-3`·`-2`·`-1`·`current-archive.md`
- **그다음 파트 후보** = `BACKLOG.md`(사용자 예정 "타일 ESC로 닫기")
- 파일 편집기·보기 타일·영상·타일 끌기·내 장비 다섯 파트 = 11-03 INDEX로 접음(`INDEX-file-view`·`-view-tile`·`-video-view`·`-tile-drag`·`-member-drag`)

## 운영 상태 (2026-11-02 확인)
- **지금 떠 있는 것(11-06)**: 개발 test1(8011)·test2(8012)·worker `devtest-pc` = **11-06 사용자 요청으로 끔**(다시 띄우는 법 → `REF-test-env.md` "환경" 개발 화면 띄우기, 개발 DB에 계정 `devtest`·장비 `devtest-pc`·폴더 233 남음 — 비밀번호는 세션 스크래치라 사라짐, 다시 쓰려면 새 계정). 회사 PC Chrome = 11-05 Claude가 다시 켬. test3(8013) = supervisor **`build/261106-socket-reconnect`**(태그 `SUPERVISOR-SOCKET-RECONNECT-261106`, 11-06 20:58 사용자 요청으로 Claude가 교체 — 그 폴더 `swap.sh`로 옛 것 끄고 새 것 켬, 실패 시 `261105-board-close`로 되돌리게 함. `setsid`로 띄워 터미널 없음, 로그 = `build/261106-socket-reconnect/supervisor.log`, 다시 켜기 = `build`에서 `./261106-socket-reconnect/supervisor -env ../supervisor.env`. 직전 = `261105-board-close`(11-05 같은 방식)) · worker = `build/261104-files`(코드 변경 없음, 그대로)(DB `nexus_build` 마이그레이션 `00018`까지, 보관소 = `build/261104-files/update-store`). 보관 프로세스 하나(59860)는 실행 파일이 `build/261103_updater/worker` — 그 터미널을 닫기 전엔 지우지 말 것. 개발 DB `nexus` 마이그레이션 `00018`까지, 보관소 = `apps/core/cmd/supervisor/update-store`(git 제외). 5051·3001은 꺼져 있음. 옛 줄 원문 → `history/current-archive-7.md`
- **`build/`**: `261029_drag/`(지금 test3)·`261029_video_view/`·`video_view_tmp/`·`viewer_image/`·`supervisor.env`·`worker.env`·`worker-sqlite.db*`·`view-cache/`·`home-update/`
- **git**: master = `fix/socket-reconnect` 머지(11-06, 그 전 `feat/board-close` 11-05), **push = 사용자(아직)**. 남은 로컬 브랜치 = `feat/*`·`fix/*` 여러 개(머지됨) · `refactor/socket`·`refactor/tile-kinds`. 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치)·`cmd/worker/test.env`·user vault 전체. 커밋은 파트 단위
- 남은 판단: 임시 DB `nexus_fv_tmp`·`nexus_member_tmp`·`nexus_api_tmp`도 있음(11-02 목록에서 봄, 안 건드림)
- 빌드 = 태그(`apps/core/cmd/*/env` 첫 줄) 바꾸고 `scripts/build.sh <폴더명> [대상...] [--os] [--arch]`(11-04 대상별로 나눔, 태그 출력·같은 태그 경고) → `REF-build-scripts.md`
- 남은 흔적(정리 여부 사용자 미정): test1용 DB `nexus_keep` · `build/keep`의 env·sqlite · `nexus_build`의 `smoke…` 계정 2개(폴더 233·스크립트 234) · DB `nexus_td_tmp`·`nexus_vv_tmp`. 보관 프로세스 일괄 종료 `pkill -f 'worker -hold'`는 **이 Claude 세션도 죽으니 금지**
- 두 빌드를 나란히 띄우는 법(env 분리·DB 복사) → `REF-build.md` "두 빌드 나란히" / 테스트·확인 준비(헤드리스, 격리 = 임시 DB + 5051, 함정) → `REF-test-env.md`
- **최근 마감 파트**(전체 목록·순서 = `HISTORY.md` "마감 파트"): 젤리·관리 기록 `INDEX-jelly`(10-27) · 회원 관리 `INDEX-member`(10-26) · supervisor 주소 실시간 변경 `INDEX-supervisor-move`(10-26) · 빈 곳 길게 누르기 `INDEX-press-pager`(10-25) · 목록 아이콘 보기 `INDEX-dir-grid`(10-24) · 타일 머리 줄 버튼 `INDEX-tile-head`(10-24)
  - 남겨 둔 것: 확인용 노드 irony `test`/`ctrl-alt-test`(225)·`scroll-test`(215)·`scroll-test.ps1`(216) / VM `nexus-win`(`docker start/stop nexus-win`, 화면 `127.0.0.1:8006`) + VM worker `win-dev` → `REF-windows-vm.md` "VM 실행"
- **시연 데이터(2026-10-16) → `REF-demo.md`**: irony 계정 폴더 `시연-생산지그`(id 200) + test1 편집·test2 조작·test3 보기 공유. 코드 변경 없음
- **외부 접속 설정 사용 중**: 프론트 https://test1.falinux.com(8011) · supervisor https://test2.falinux.com(8012) · worker -> localhost:8012 → `REF-external-access.md`(로컬 전용으로 되돌리는 법도 거기)
- 남은 주의: 개발 DB에 테스트 계정 `active-<시각>` 3개(폴더 11개씩) / 개발용 로그인 기본값 = `apps/frontend/.env.development.local`(git 제외) / worker는 종료 정책 이후 빌드여야 force를 앎
- 역할 분담(2026-09-29 개정): node 타일 UI·process 작업은 코드 대부분을 Claude가 작성하고, 로직 구조는 사용자와 먼저 맞춘 뒤 쓴다

---

## 유지 사항 (항상)
- 색 규칙: Vuetify 테마 색 이름 의존, hex 금지 → `REF-frontend.md` "색 규칙" / `position_x/y` 사용 중단(DB 컬럼 유지) / 외부 레이아웃 라이브러리 안 씀 / 트리 드래그 device_key 재상속 이슈 → `REF-node-label.md`
- EXEC 실행 방식 = A(직접 실행, shebang 필수) 현상 유지 확정(2026-10-12) / kill 정책 → `REF-process-kill.md` "확정"
- 마이그레이션 번호는 착수 때 실제 다음 번호로(`00005`는 `process_subscribers` 제거에 씀)
- 모바일 키보드 주의: `appWindown.store.ts`의 `size.inner`(visualViewport)에 판정을 연결하면 안 됨 → `REF-node-ui-overview.md`
