# CURRENT

## 현재 날짜
2026-10-26

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **(2026-10-26 오후) 파트(회원 관리) UI 개편 코드 적용 중** — UI-1(서버 목록)~UI-6(로그인·가입) + Ctrl+Alt 컨트롤 패널(보드와 공유 훅) + 알약 끌기(판 계약 `PagerBoard`, A안) 끝, 헤드리스 확인 끝 -> 사용자 브라우저 확인 OK("맘에 든다") -> 마감 전 회귀 17/17 + 옛 폴더 403 의도 확인 -> **파트 커밋 `1bd58e7`(feat/member)** -> 다음 = `feat/supervisor-move` 마감·master 머지 -> master를 `feat/member`에 머지 -> 배포(`REF-member.md` "배포 전 확인", 관리자 회원 묻기)(→ `REF-member-ui.md` "구현"). 그 뒤 = 마감 전 회귀(`REF-member-impl.md` "파트 마감까지 남은 것"). 격리 5051·3001·`jig-a`는 지난 세션 것이 그대로 떠 있음(scratchpad `e786c602…`)
- **(2026-10-26, 재부팅 직전) 파트(회원 관리) 진행 중 — 서버 전부 + 화면 첫 판 끝, 헤드리스 확인 OK, 커밋 전** → 결정 `REF-member.md` · 구현·다음 할 것 `REF-member-impl.md`
  - **다음 세션 첫 할 일(사용자 요청)**: ① `REF-member-impl.md` "격리 환경 다시 띄우기"로 5051 supervisor·3001 vite·`jig-a` worker를 띄운다(재부팅으로 scratchpad·프로세스 사라짐, DB `nexus_member_tmp`는 남음) ② 같은 파일 "다음 세션 첫 안내"의 **마지막 응답 원문을 그대로** 사용자에게 전한다 ③ 사용자가 화면 보며 지적 -> UI 다듬기
  - 작업 = **worktree `/home/jh-bae/irony/nexus-member`, 브랜치 `feat/member`**(머지 커밋 `6f87465` 위에 커밋 안 한 변경). 이 폴더(master)·`feat/supervisor-move`는 안 건드림. vault는 이 폴더에서 기록
  - 배포 전 필수: 관리자로 지정할 회원 묻기 · supervisor env `KEY_SECRET` · worker 실행 파일 + `API_KEY` 재배포(`REF-member.md` "배포 전 확인")
- **(2026-10-26 세션 끝) 다음 세션 = 파트(회원 관리) UI 개편 코드 적용** → `REF-member-ui.md` "코드 적용 계획"(UI-1(서버 목록)~UI-7(확인), 착수 전에 표를 보이고 맞춤) / 다시 띄우기 → `REF-member-impl.md` "다음 세션 시작". 시안 확정 https://claude.ai/artifact/XeSQZMThQjbs5CxGMage9r(사용자 "디자인 좋아 맘에 들어")
  - 순서(2026-10-26 사용자): 회원 관리 확인·UI 먼저 -> supervisor-move 머지(마감 커밋 + vault 충돌은 master 기준) -> 회원 관리 마감 때 master를 `feat/member`에 가져와 머지 -> 집 서버 재배포(master 정식 빌드, `REF-member.md` "배포 전 확인" — 가능함 확인, 모든 worker 실행 파일 + `API_KEY` 재배포 필요)
  - 머지 전 답 대기: 회원 관리 vault 파일(`REF-member*.md`·`history/member*.md`, 아직 git 밖)을 master vault 커밋에 같이 넣을지 · MEMORY 색인 2단화를 그때 같이 할지
  - 지금 떠 있는 것(재부팅하면 사라짐): 격리 5051·3001·`jig-a`(scratchpad) + 사용자가 띄운 `test-worker`(`go run ./cmd/worker -env test.env`)
- **(2026-10-26 회사) Linux B 빌드 `build/supervisor_move` 만듦**(태그 `SUPERVISORMOVE-261026`, embed env = test3 값) -> 사용자가 test3에 띄운 뒤 집 서버로 옮길 예정 → `REF-home-server.md` "Linux B 빌드"
- **(2026-10-25 세션 끝) 다음 = 회사에서 worker 전부 최신으로 + supervisor를 집 서버 주소로**(사용자 계획 2026-10-26) → `REF-home-server.md` "내일 할 것". 집 서버(운영) 구축 끝: Windows 서버 PC + Cloudflare Tunnel + 부팅 자동 시작, **B 빌드 = master + `feat/supervisor-move`(master 미머지 코드)로 운영 중**. 파트(supervisor 주소 실시간 변경)는 여전히 사용자 확인 -> master 머지 대기. 작업 트리 = master(파트(찰떡 UI) 머지 + vault), 떠 있는 개발 서버 없음. **MEMORY ~10k자** — 다음 vault 정리 때 색인 2단화(마감 파트 줄을 INDEX로)부터. 새 파트 후보 `BACKLOG.md`
- **2026-10-25 집 서버(운영) 구축 끝** → 별도 기록 `REF-home-server.md` / `history/home-server.md`(Nexus 기능 파트와 섞지 않음). 산출물(저장소 밖 `build/`, git 제외): `nexus-server.zip`(첫 설치) · `nexus-update.zip`(B 빌드 + `update.ps1`) · `nexus-autostart.ps1`(작업 스케줄러) · `nexus-create-user.sql`. 계정 `irony1090`(`worker_move`) 로그인 OK. Access는 BACKLOG(회사 worker가 터널로 `/worker`에 붙으면 Access가 막으니 경로 예외 필요)
- **2026-10-24 파트(supervisor 주소 실시간 변경) 구현·격리 확인 끝 — 사용자 확인 대기**(이 브랜치를 master에 머지할 때 vault 충돌(CURRENT·HISTORY·MEMORY)은 master 쪽 기준 — 2026-10-24 `feat/tile-head` 때 vault를 옮겨 와 커밋했기 때문)(→ `REF-supervisor-move.md` "구현"·"확인"): 브랜치 `feat/supervisor-move`(커밋 전). 주소-1(주소 목록)~주소-5(화면) 구현, 스크립트 이름 표시 추가(사용자 요청). 격리 확인 서버 29/29·화면 8/8, 격리 환경 정리함. **사용자 브라우저 확인 대기 — 2026-10-24 주말 외부라 다음 주에 확인**(사용자 요청으로 브랜치에 중간 커밋(`wip(supervisor-move)`), master 머지 안 함 — 확인 뒤 마감 커밋·머지). 돌아오면 확인할 것: ① 빌드 뒤 장비 타일 ⋮ 메뉴 "supervisor 주소 변경"·"미등록 process" ② 권한 SQL(`REF-supervisor-move.md` 00013 주석) 뒤 푸터 보드 메뉴 "전체 worker 주소 변경" ③ 빌드본 교체 주의 = `REF-build.md` 빌드본 적용 때 supervisor `00013`·worker `00003`·`00004` 마이그레이션이 붙음
- **지금 떠 있는 것**: (2026-10-26) 파트(회원 관리) 격리(5051·3001·worker `jig-a`)는 **재부팅으로 꺼짐 — DB `nexus_member_tmp`는 남김**(다시 띄우는 법 `REF-member-impl.md`). (2026-10-25) test1(8011) 꺼짐. test3(8013) = `build/ctrl_panel`, **`web/`만 파트(찰떡 UI) 빌드로 교체**(2026-10-25, 그 전 = 확대 막기 빌드, 백업 없음). 웹 백업(`web.bak*` 4개)은 사용자 요청으로 2026-10-25 전부 지움(되돌리려면 해당 커밋에서 다시 빌드). 위 파트 줄들의 `web.bak` 언급은 그때 기록 **8012는 지금 다른 프로젝트(`irony/community`) vite가 씀** / test3(8013) = `build/directory_path`(태그 `DIRECTORYPATH-261023`, 사용자가 빌드·실행, 경로 직접 입력 포함. `build/download`도 사용자 빌드). 보관 프로세스 19096·19684·61491 = `build/keep/worker -hold` -> **`build/keep` 지우지 말 것**. test1(8011)·개발(8012) 꺼짐. VM `nexus-win` 꺼짐(2026-10-23 확인 뒤 끔. 켜면 `start-test3.ps1`로 `win-dev`를 test3에 붙임, exe = `DIRECTORYPATH-261023` + `C:` 고침)(→ `REF-windows-vm.md` "VM 실행"). 이 세션 격리 환경(5051·3001·`nexus_xferext_tmp`·`nexus_xferdown_tmp`·`nexus_pathin_tmp`)은 정리함
- 빌드 = 태그(`apps/core/cmd/*/env` 첫 줄) 바꾸고 `scripts/build.sh <이름>`(태그 출력·같은 태그 경고) → `REF-build.md` "빌드 태그"
- 남은 흔적(정리 여부 사용자 미정): test1용 DB `nexus_keep`(시험 계정 `smoke<시각>` 1개)·`build/keep`의 env·sqlite(test1 = 재부팅 뒤 안 켬) · `nexus_build`에 `smoke1792672658797`·`smoke1792672663267`(폴더 233·스크립트 234). 보관 프로세스 일괄 종료 `pkill -f 'worker -hold'`는 **이 Claude 세션도 죽으니 금지**
- 두 빌드를 나란히 띄우는 법(env 분리·DB 복사) → `REF-build.md` "두 빌드 나란히"
- **나중에 할 것 = `BACKLOG.md`**(2026-10-22 신설, 세션 시작 시 안 읽음): 다음 파트 후보·알려진 버그·보류 설계·열린 질문·답 대기 / PC 재부팅은 다음 작업이 없을 때 확인
- **최근 마감 파트**(상세는 각 INDEX, 목록 = `HISTORY.md` "마감 파트"): 빈 곳 길게 누르기 `INDEX-press-pager`(10-25) · 목록 아이콘 보기 `INDEX-dir-grid`(10-24) · 타일 머리 줄 버튼 `INDEX-tile-head`(10-24) · Ctrl+Alt 팝업 버튼·터치 스크롤 양 `INDEX-pager-tools`(10-23) · 글자 크기 규칙 `INDEX-typography`(10-23) · 경로 직접 입력 `INDEX-path-input`(10-23) · 내려받기·worker 간 전송 `INDEX-transfer-down`(10-23) · 파일 전송 확장 `INDEX-transfer-ext`(10-23) · worker 폴더 타일 `INDEX-worker-dir`(10-23) · 타일 종류 정리 `INDEX-tile-kinds`(10-23) · process 살려 두기 `INDEX-process-keep`(10-22) · 새 보드·IUTF8 `INDEX-fix-board-iutf8`(10-22) · 빌드본 분리 `INDEX-build`(10-22) · 리팩터링 `INDEX-refactor`(10-21) · 페이지 이동 `INDEX-tile-move`(10-21) · 한글 입력 중복 `INDEX-term-ime` · 파일 전송 `INDEX-transfer-e2e` · 헤더·푸터 개편 `INDEX-board-chrome` · 페이지 표시 개편 `INDEX-board-pager`(10-20) · URL 바인딩 `INDEX-tile-url` · 터미널 터치 스크롤 `INDEX-term-touch` · Windows worker `INDEX-windows-worker` · PgUp·PgDn 버튼 `INDEX-term-page-keys` · 고아 스크립트 `INDEX-worker-orphan` · 보드 미리보기 `INDEX-board-preview`(10-15~19)
  - 남겨 둔 것: 확인용 노드 irony `test`/`ctrl-alt-test`(225)·`scroll-test`(215)·`scroll-test.ps1`(216) / VM `nexus-win`(`docker start/stop nexus-win`, 화면 `127.0.0.1:8006`) + VM worker `win-dev` → `REF-windows-vm.md` "VM 실행" / Windows 미확인 항목 → `BACKLOG.md`
- **시연 데이터(2026-10-16) 등록 → `REF-demo.md`**(시연 순서 점검은 `BACKLOG.md`): irony 계정 폴더 `시연-생산지그`(id 200, 안 C(생산 지그 시나리오)) + test1 편집·test2 조작·test3 보기 공유. 코드 변경 없음(REST로 데이터만, 커밋할 코드 없음)
- **외부 접속 설정 사용 중**: 프론트 https://test1.falinux.com(8011) · supervisor https://test2.falinux.com(8012) · worker -> localhost:8012 → `REF-external-access.md`. 로컬 전용(5050·3000)으로 되돌리는 법도 거기
- 남은 주의: 개발 DB에 테스트 계정 `active-<시각>` 3개(폴더 11개씩) / 개발용 로그인 기본값 = `apps/frontend/.env.development.local`(git 제외) / worker는 종료 정책 이후 빌드여야 force를 앎
- 커밋 상태(2026-10-26): **`feat/member`(worktree `/home/jh-bae/irony/nexus-member`) = 머지 커밋 `6f87465` + 파트 커밋 `1bd58e7`(2026-10-26), master 머지 전**. master 최신 = 파트(찰떡 UI) 머지(그 전 = 확대 막기 + vault 정리). 브랜치 `feat/dir-grid`·`feat/press-pager`·`feat/chaltteok-ui`도 남겨 둠. `feat/supervisor-move`는 중간 커밋(머지 안 함). 옛 브랜치 `refactor/socket`·`refactor/tile-kinds`·`feat/worker-dir`·`feat/transfer-ext`·`feat/transfer-down`·`feat/path-input`·`feat/typography`·`feat/pager-tools`·`feat/tile-head` 남겨 둠. **push는 한 번도 안 함**(분기 태그 `checkpoint/before-folder-tile-uid` = `d50f153`, 태그는 push 때 따로 `git push origin <tag>`). 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치). 커밋은 파트 단위(중간 커밋은 사용자가 요청할 때). user vault 전체 git 제외
- 테스트·확인 준비(환경, 헤드리스, 함정, **격리 실행 = 임시 DB + 5051**) → `REF-test-env.md`
- 역할 분담(2026-09-29 개정): node 타일 UI·process 작업은 코드 대부분을 Claude가 작성하고, 로직 구조는 사용자와 먼저 맞춘 뒤 쓴다


---

## 유지 사항 (항상)
- 색 규칙: Vuetify 테마 색 이름 의존, hex 금지 → `REF-frontend.md` "색 규칙" / `position_x/y` 사용 중단(DB 컬럼 유지) / 외부 레이아웃 라이브러리 안 씀 / 트리 드래그 device_key 재상속 이슈 → `REF-node-label.md`
- EXEC 실행 방식 = A(직접 실행, shebang 필수) 현상 유지 확정(2026-10-12) / kill 정책 → `REF-process-kill.md` "확정"
- 마이그레이션 번호는 착수 때 실제 다음 번호로(`00005`는 `process_subscribers` 제거에 씀)
- 모바일 키보드 주의: `appWindown.store.ts`의 `size.inner`(visualViewport)에 판정을 연결하면 안 됨 → `REF-node-ui-overview.md`
