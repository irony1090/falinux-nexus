# CURRENT

## 현재 날짜
2026-11-02

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **(2026-11-02) 다음 세션 첫 일 = 파트(업데이트 app) 설계 재검토·결정** — 코드 없음, 설계·시안만. `REF-review-1030.md` "설계" 표의 설계-1(지우고 깔기)~설계-6(버전 이름 체계)부터 하나씩 정하고, 설계-7(trap 신호)~설계-12(시안 표기)는 그다음. 정해지면 `REF-update-deploy.md`·`REF-update-rules.md`·시안에 반영
  - 볼 것 = `REF-review-1030.md` + `REF-update-deploy.md` + `REF-update-rules.md` + 시안 `user/배포-화면-찰떡.html`(4탭)
  - 배포 화면은 `MemberPage` 타일 틀을 쓰므로 열기 순서(연 곳 바로 뒤, 없으면 맨 앞)·끌기(`DragSession`)를 따른다 → 설계-11(배포 정보 타일 열기)
- **(2026-11-02) 10-30 작업 점검·수정 마감** — `fix/member-drag-review` master 머지. 코드 항목 수정 + 사용자 실기기 확인 끝 → `REF-review-1030.md` · `history/review-1030.md`. 확인 스크립트 = `scripts/check/run.sh`
  - **아직 안 한 것**: 새 빌드로 색인 금지 배포(코드-6(색인 금지 미배포)) = 사용자. `cmd/*/env` 태그를 바꿔 빌드, 집 서버 묶음은 worker 포함. 8013·집 서버는 지금 옛 빌드(10-30 이후 수정 없음)
- **(2026-10-30) 파트(업데이트 app) — 설계만 끝, 코드 착수 전** → `REF-update-deploy.md`(설계·화면) · `REF-update-rules.md`(확정 규칙) · `history/update-deploy.md` · `history/update-rules.md`
  - 확정: 장비마다 `nexus-update` 상주(Job 밖) · 관리자 웹 화면 버튼 · 장비별+전체 배포 · 릴리스 이력+되돌리기 · **Linux 먼저**
  - 릴리스 = 폴더 트리(`common` + `<os>-<arch>`), 판정은 파일 유무 · 단계 = 스크립트 1개 + `steps.json` · **잠금 = arch 단위(`배포 완료` 확정)** · 되돌리기 = 이전 릴리스 재배포 · 버전 = 빌드에 박힌 env 첫 줄 · 내 장비 = 인스턴스별 `배포 정보` 타일(주인도)
  - **키 = `메인키#서브키`(업데이터 인스턴스)** — 같은 이름 두 대도 각각 타일·배포·이력·잠금
  - 화면 시안 = `user/배포-화면-찰떡.html`(4탭, 렌더 9장 = `user/배포-화면-찰떡/`) · 원판 `user/배포-화면.html` · 설계안 `user/배포-설계.html` → **사용자 확인 대기**
  - **다음 = P1(보관·화면)**. 착수 전 확정 3개 = 콘솔 인터랙션(가/나) · 스크립트를 `common`에 둘지 · 첫 시험 장비(`fa24-PID`/`imx8mp-f24`)
  - 참고: 이 세션 = `ANTHROPIC_API_KEY` 인증이라 아티팩트 게시 불가 → HTML을 `user/`에 둠
- **(2026-10-30) 파트(내 장비·관리자 타일 이동) 마감** — `feat/member-drag` master 머지 → `REF-member-drag.md` · `history/member-drag.md`. 미확인 = 회원→장비 자식 붙기(dev DB에 장비 없음) · `openParent`가 자식 묶음 중간에 낄 수 있음(옛 동작, 안 고침)
- **(2026-10-29) 파트(타일 끌기) 마감** — `feat/tile-drag` 머지 → `REF-tile-drag.md` · `REF-tile-drag-impl.md` · `REF-tile-order.md`
- **(2026-10-29) 파트(영상 Range) 마감** — `feat/video-view` 머지(마이그레이션 `00015`) → `REF-video-view.md`. 집 서버엔 아직 안 보냄 → `BACKLOG.md`
- **(2026-10-29 저녁) 집 서버 drag 묶음 배포 완료**(`build/home-update/nexus-drag.zip`, 태그 `DRAG-HOMESERVER-261029`) → `REF-home-deploy.md` "drag 배포 묶음"
- **(2026-10-29) 보드 `fa24-PID`(192.168.40.228) worker = `VIEWER_IMAGE-261028`** → `REF-build.md` "보드 worker 교체"
- 옛 CURRENT 줄 원문 → `history/current-archive-5.md`(10-30 내려보냄: 10-29 파트 마감·떠 있는 포트 표·build 흔적), 그 전 = `-4`(10-27 ~ 10-28)·`-3`·`-2`·`-1`·`current-archive.md`
- **그다음 파트 후보** = `BACKLOG.md`(사용자 예정 "타일 ESC로 닫기")
- 파일 편집기·보기 타일·영상·타일 끌기·내 장비 다섯 파트 REF는 아직 INDEX로 안 접음(MEMORY·HISTORY가 10k에 닿으면 접기, 파트 묶음은 사용자 확인)

## 운영 상태 (2026-11-02 확인)
- **지금 떠 있는 것**: test3(8013) = supervisor·worker **`build/261029_drag`**(11-02 09:59 사용자 실행). 보관 프로세스 2개(48509·66320)도 실행 파일이 `build/261029_drag/worker` — 그 터미널을 닫기 전엔 지우지 말 것. test1(8011)·test2(8012)·5051·3001은 꺼져 있음(11-02 실기기 확인 뒤 끔)
- **`build/`**: `261029_drag/`(지금 test3)·`261029_video_view/`·`video_view_tmp/`·`viewer_image/`·`supervisor.env`·`worker.env`·`worker-sqlite.db*`·`view-cache/`·`home-update/`
- **git**: master = `fix/member-drag-review` 머지(11-02), push = 사용자. 남은 로컬 브랜치 = `feat/*`·`fix/*` 여러 개(머지됨) · `refactor/socket`·`refactor/tile-kinds`. 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치)·`cmd/worker/test.env`·user vault 전체. 커밋은 파트 단위
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
