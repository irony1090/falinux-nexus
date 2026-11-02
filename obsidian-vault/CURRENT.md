# CURRENT

## 현재 날짜
2026-11-02

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **(2026-11-02 밤) 파트(업데이트 app) — P1(보관·화면) 마감 커밋** — 브랜치 `feat/updater-p1`(커밋됨, **master 머지는 사용자 확인 대기**). 구현 지도·계획과 달라진 것(차이-1(장비 루트 테이블)~차이-5(폴더 업로드)) → `REF-update-p1.md` · 설계 `REF-update-deploy.md` · 규칙 `REF-update-rules.md`
  - 확인: `scripts/check/run.sh` 전체 통과(회원 회귀 + `updates-api.py` 39 + `updater-ui.mjs` 27). 화면 = 보드·관리 메뉴 `Updater`(`/updater`)
  - (11-02 밤) 차이-1~5 사용자 확인 끝(다섯 개 모두 이대로) · P1 마감 커밋
  - **다음 = ① master 머지 여부 사용자 확인 ② P2(updater Linux) 착수 — 먼저 `BACKLOG.md` "P2(updater) 전 확정 6개"를 사용자와 정함**(updater 프로그램은 아직 없음)(코드 + vault, master 머지는 사용자 확인 후) → 그다음 **P2(updater Linux)** 착수 = 먼저 `BACKLOG.md` "P2(updater) 전 확정 6개"를 사용자와 정함(updater 프로그램은 아직 없음 — P1은 페이지·보관소까지)
  - (11-02 밤) 타일 아래 줄 넘침 수정(안내 글 `...` + 툴팁 `MbNote`, 버튼 줄바꿈) → `REF-design-chaltteok.md` "타일 아래 줄 넘침". 안내 글이 혼자 한 줄이면 줄바꿈(두 줄-1) · 벽에 붙은 입력칸 벽 쪽 선 없앰(벽-2) → `REF-design-chaltteok.md` "벽에 붙은 입력칸"
  - (11-02 밤) 샘플 데이터를 개발 DB `nexus`에 넣음(`sample-jig`·`sample-board`·공용 2개, 인스턴스·잠금 행 없음) → `REF-update-p1.md` "샘플 데이터"
  - 개발 DB `nexus`에 마이그레이션 `00016` 적용됨(11-02 17:50). 8013 빌드본(DB `nexus_build`)은 아직 안 들어감
  - 새 의존성: 프론트 `@codemirror/*`·`@lezer/highlight`(`npm i` 필요), 메인 보드 스크립트 편집도 CodeMirror(`CtCode`)로 바뀜. `tsconfig.app.json` `baseUrl` 제거(TS 6 경고 TS5101)
  - 남은 설계: 설계-8(supervisor 되돌리기)·설계-10(자기 교체) = P3 전 / P2 전 = 접속 토큰·실행 권한(`+x`)·`state.json` 형식·첫 시험 장비·설치 경로·콘솔 인터랙션 → `REF-update-deploy.md` "미정" · `BACKLOG.md`
  - 시안(확정) = Claude 아티팩트 https://claude.ai/artifact/XGMJqjg7rHdAHUhRMXHbZa(v15). 시안·화면을 고칠 땐 먼저 `REF-design-chaltteok.md` "시안에서 정리한 규칙(2026-11-02)" 읽기(묶음 쪼개지 말 것). 시안 확인법 = headless Chrome 스크린샷
- **(2026-11-02) 10-30 작업 점검·수정 마감** — `fix/member-drag-review` master 머지. 코드 항목 수정 + 사용자 실기기 확인 끝 → `REF-review-1030.md` · `history/review-1030.md`. 확인 스크립트 = `scripts/check/run.sh`
  - **아직 안 한 것**: 새 빌드로 색인 금지 배포(코드-6(색인 금지 미배포)) = 사용자. `cmd/*/env` 태그를 바꿔 빌드, 집 서버 묶음은 worker 포함. 8013·집 서버는 지금 옛 빌드(10-30 이후 수정 없음)
- **(2026-10-30) 파트(내 장비·관리자 타일 이동) 마감** — `feat/member-drag` master 머지 → `REF-member-drag.md` · `history/member-drag.md`. 회원→장비 자식 붙기는 11-02 확인 끝. 남은 것 = `openParent`가 자식 묶음 중간에 낄 수 있음(옛 동작, 안 고침)
- **(2026-10-29) 파트(타일 끌기) 마감** — `feat/tile-drag` 머지 → `REF-tile-drag.md` · `REF-tile-drag-impl.md` · `REF-tile-order.md`
- **(2026-10-29) 파트(영상 Range) 마감** — `feat/video-view` 머지(마이그레이션 `00015`) → `REF-video-view.md`. 집 서버엔 아직 안 보냄 → `BACKLOG.md`
- **(2026-10-29 저녁) 집 서버 drag 묶음 배포 완료**(`build/home-update/nexus-drag.zip`, 태그 `DRAG-HOMESERVER-261029`) → `REF-home-deploy.md` "drag 배포 묶음"
- **(2026-10-29) 보드 `fa24-PID`(192.168.40.228) worker = `VIEWER_IMAGE-261028`** → `REF-build.md` "보드 worker 교체"
- 옛 CURRENT 줄 원문 → `history/current-archive-6.md`(11-02 내려보냄: 업데이트 app 10-30 설계 상태·11-02 재검토 중 줄), 그 전 = `-5`(10-29 파트 마감·떠 있는 포트 표·build 흔적), 그 전 = `-4`(10-27 ~ 10-28)·`-3`·`-2`·`-1`·`current-archive.md`
- **그다음 파트 후보** = `BACKLOG.md`(사용자 예정 "타일 ESC로 닫기")
- 파일 편집기·보기 타일·영상·타일 끌기·내 장비 다섯 파트 REF는 아직 INDEX로 안 접음(MEMORY·HISTORY가 10k에 닿으면 접기, 파트 묶음은 사용자 확인)

## 운영 상태 (2026-11-02 확인)
- **지금 떠 있는 것**: test3(8013) = supervisor·worker **`build/261029_drag`**(11-02 09:59 사용자 실행). 보관 프로세스 2개(48509·66320)도 실행 파일이 `build/261029_drag/worker` — 그 터미널을 닫기 전엔 지우지 말 것. **test1(8011, vite)·test2(8012, 개발 supervisor `go run`, 브랜치 `feat/updater-p1`) 11-02 밤 Claude 세션이 다시 띄움 — 세션을 다시 시작하면 같이 꺼질 수 있음(`ss -ltnp | grep -E ':801[12]'`로 확인, 다시 띄우기 = `apps/core`에서 `go run ./cmd/supervisor -env dev.env` · `apps/frontend`에서 `npm run dev`)** — 개발 DB `nexus`에 마이그레이션 `00016` 적용됨, 보관소 = `apps/core/cmd/supervisor/update-store`(git 제외). 5051·3001은 꺼져 있음
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
