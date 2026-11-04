# CURRENT

## 현재 날짜
2026-11-04

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **(2026-11-04) 로컬 updater 시험 + 주소 순위 눌러 입력 + 되돌리기 순서(안 A) — 마감, `feat/addr-rank-pick` 커밋·master `--no-ff` 머지(push = 사용자)** · 사용자 화면 확인 끝 → `history/update-rules.md` 11-04 · `history/supervisor-move.md` 11-04
  - 주소 순위 눌러 입력 = 순위 줄(밑줄 글자) 누르면 방식·호스트 입력칸에 → `REF-supervisor-move.md` "순위 눌러 입력"(Vuetify reset 함정 포함)
  - 되돌리기 = **최신만(역순)**, 이미 되돌린 배포 거절, 확인 창 = "배포 직전 상태로, 그 버전으로 가는 게 아님" → `REF-update-rules.md` "되돌리기"
  - **지금 켜져 있는 것**: 로컬 updater `go run ./cmd/updater -c apps/core/cmd/updater/updater.json`(인스턴스 `dev-main#mzOwVeBF` -> test3 `ws://localhost:8013`, `targetDir` = `build`, 설치본 `test-1`). **`updater.json`(키 들어 있음)·`state.json`이 저장소 안·git 제외 아님 — 커밋 때 빼야 함** → `BACKLOG.md` / test3 = `build/261104-files`(11-04 11:39 새 supervisor, 보관소 = `build/update-store`(env `UPDATE_STORE_DIR="../update-store"`))
  - 시험 흔적(nexus_build): 장비 루트 `dev-main`, 빌드 `supervisor_url_click-261104`·`-01`·`test-1`·`test-2`, 배포 #1~#7. `build/261104-files/web/assets`에 직접 옮긴 옛 해시 파일 몇 개(안 씀)
- **(2026-11-03) 파트(업데이트 app) — P1(보관·화면) · P2(updater Linux) · P3(배포) · P4(Windows) 마감, master 머지 끝**(P1 `1f5eda9` · P2 `bd93fa4` · P3 = `feat/updater-p3` · P4 = `feat/updater-p4` 머지, push = 사용자). 사용자 P3 화면 확인 "나쁘지 않음"(11-03). 옛 줄 원문 → `history/current-archive-6.md` 11-03 절
  - 구현 지도: P1 `REF-update-p1.md` · P2 `REF-update-p2.md` · P3 `REF-update-p3.md` / 설계 `REF-update-deploy.md` · 규칙 `REF-update-rules.md`
  - P4(Windows) = 구현 · VM 확인 39/39 · 집 서버 설치(11-03 12:42 `C:\Nexus\updater` 동작) → `REF-update-p4.md` · 집 서버 → `REF-home-deploy-updater.md`(관리자 = irony1090)
  - **그다음 = 단위-7(교체 빌드)**(Files 타일 개선 끝, 정-1~3 답 대기 그대로): Updater 화면에서 집 서버 supervisor·worker를 updater로 교체(코드 없음, 보관소 빌드). **(11-03 밤) 착수 — updater 교체는 나중, 정-1(old 정리 단계)·정-2(DB 백업 자리) = 제안대로, 정-3(관찰 방식) = 회사에서 지켜봄. 묶음 `build/home-update/files-tile/` -> 11-03 밤 사용자가 집 서버 updater로 배포 완료, "잘 동작해"(첫 updater 배포 = 단위-7 끝). 남은 것 = 집 서버 updater 교체(**사용자가 직접 빌드, 11-04엔 로컬 시험부터 함**, 순서 = `REF-home-deploy-updater.md` "updater 교체 — 사용자 직접 빌드") · `nexus-before-updater.dump` 삭제(사용자)** → `REF-home-deploy-updater.md` "files-tile 묶음" → `REF-update-p4.md` "단위-7(교체 빌드) 후속". 답을 받으면 `steps.json`·스크립트 계획을 먼저 보여 줄 것
  - **(11-03 밤) Files 타일 개선 마감** — `feat/update-files-tile` 코드 커밋 `3caff7b` + vault 커밋 + master `--no-ff` 머지(push = 사용자). 단위-가(폴더 트리·선택 삭제)·시험 실행·단위-나(끌어 놓기)·단위-다(갈아치우기)·단위-라(실행 권한 이름) 전부 사용자 화면 확인 끝 → `REF-update-files-tile.md` · `REF-update-replace.md` · `REF-update-trial.md`. 남은 것 = **집 서버 updater 새 빌드 교체(사용자)** — 옛 updater는 replace를 무시하고 합치기로 배포. 진행 블록 원문 → `history/current-archive-6.md` 11-03 밤 절
  - **지금 켜져 있는 것**: VM `nexus-win` + updater `test-win#BGRB5DdV`(test2, `C:\NexusU`, 예약 작업 아님) → `REF-windows-vm.md` 11-03 줄, 다 쓰면 `docker stop nexus-win` / 개발 서버 **8011(vite) = 11-03 밤 사용자 요청으로 끔**(다시 = `apps/frontend`에서 `npm run dev`) · 8012(supervisor `go run`, 11-03 밤 단위-다 코드로 다시 띄움 — 이전 메모: 11-03 16:39 이 세션이 `setsid`로 다시 띄움 — 시험-1 라우트 반영, 세션이 끝나도 남음, 로그는 스크래치라 사라짐) / 8013 빌드본. 시험 확인용 ssh 래퍼 `vssh`는 세션 스크래치라 사라짐(`REF-test-scripts-1014.md` "윈도-6"에서 다시 만듦)
  - **사용자 확인 대기**: DeepSeek API 키 재발급 + `ds.sh` 인자 노출 개선 여부 → `BACKLOG.md`
  - 집 서버 남은 정리(사용자가 지움): `*.old-20261103-123559` 3개(지워도 된다고 답함) · `nexus-before-updater.dump`(단위-7 `pg_dump` 뒤)
  - 로컬 VM `nexus-win`에 시험 흔적 `C:\NexusT`·`C:\NexusS` 남음(지워도 됨). 시험 스크립트(`upd_win*.py`)는 세션 스크래치라 사라짐 — 다시 필요하면 `REF-update-p4.md` "확인" 표로 다시 씀
  - 실기 장비: `imx8mp-f24` 서비스 `nexus-updater`(P3 실행 파일, 태그는 P2 이름 그대로 — 사용자: 중요하지 않음)가 test2에 붙어 있음. 개발 DB `nexus`·8013 빌드본 `nexus_build` 둘 다 마이그레이션 `00018`까지(11-04 확인)
  - 남은 흔적: 로컬 시험 updater 인스턴스 `sample-jig#asWTzHHK` 행 offline(지워도 됨)
  - 새 의존성: 프론트 `@codemirror/*`·`@lezer/highlight`(`npm i` 필요), 메인 보드 스크립트 편집도 CodeMirror(`CtCode`)로 바뀜. `tsconfig.app.json` `baseUrl` 제거(TS 6 경고 TS5101)
  - 시안(확정) = Claude 아티팩트 https://claude.ai/artifact/XGMJqjg7rHdAHUhRMXHbZa(v15). 시안·화면을 고칠 땐 먼저 `REF-design-chaltteok.md` "시안에서 정리한 규칙(2026-11-02)" 읽기(묶음 쪼개지 말 것). 시안 확인법 = headless Chrome 스크린샷
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
- **지금 떠 있는 것(11-04 확인)**: test3(8013) = supervisor·worker **`build/261104-files`**(사용자 최신 빌드, DB `nexus_build` 마이그레이션 `00018`까지, 보관소 = `build/261104-files/update-store`). 보관 프로세스 하나(59860)는 실행 파일이 `build/261103_updater/worker` — 그 터미널을 닫기 전엔 지우지 말 것. **11-04 아침 test1(8011)·test2(8012) 둘 다 꺼져 있음.** 옛 메모: test1 = 11-03 밤 끔 · test2(8012, 개발 supervisor `go run`) Claude 세션이 다시 띄움 — 세션을 다시 시작하면 같이 꺼질 수 있음(`ss -ltnp | grep -E ':801[12]'`로 확인, 다시 띄우기 = `apps/core`에서 `go run ./cmd/supervisor -env dev.env` · `apps/frontend`에서 `npm run dev`)** — 개발 DB `nexus`에 마이그레이션 `00018`까지 적용됨, 보관소 = `apps/core/cmd/supervisor/update-store`(git 제외). 5051·3001은 꺼져 있음
- **`build/`**: `261029_drag/`(지금 test3)·`261029_video_view/`·`video_view_tmp/`·`viewer_image/`·`supervisor.env`·`worker.env`·`worker-sqlite.db*`·`view-cache/`·`home-update/`
- **git**: master = `feat/update-files-tile` 머지 `8123223`(11-03 밤), **push = 사용자(아직)**. 마감 뒤 vault 정리분은 master 작업 트리에 미커밋(다음 파트 브랜치에서 같이 커밋). 남은 로컬 브랜치 = `feat/*`·`fix/*` 여러 개(머지됨) · `refactor/socket`·`refactor/tile-kinds`. 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치)·`cmd/worker/test.env`·user vault 전체. 커밋은 파트 단위
- 남은 판단: 임시 DB `nexus_fv_tmp`·`nexus_member_tmp`·`nexus_api_tmp`도 있음(11-02 목록에서 봄, 안 건드림)
- 빌드 = 태그(`apps/core/cmd/*/env` 첫 줄) 바꾸고 `scripts/build.sh <이름>`(태그 출력·같은 태그 경고) → `REF-build.md` "빌드 태그"
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
