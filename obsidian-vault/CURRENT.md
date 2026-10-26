# CURRENT

## 현재 날짜
2026-10-26

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **(2026-10-26 오후) `feat/member` master 머지 끝(`4ae97a7`, push = 사용자) · 집 서버 회원 관리 배포 진행 중 = Nexus 터미널만으로** → `REF-home-server.md` "회원 관리 배포 — Nexus만으로" **집 서버 배포 완료(16:59 `state=DONE`)**. 파트(회원 관리) 마감 정리는 아직(사용자 확인 대기)
- **(2026-10-26 저녁 세션 끝, 사용자 퇴근 -> 집에선 집 서버 ssh·원격 가능)** 브랜치 `refactor/api-prefix`(작업 폴더 = 원래 폴더, **커밋 전**) = ① 브라우저 API 전부 `/api`(`REF-api-prefix.md`, 격리 14/14) ② 열린 장비 타일 supervisor 주소 일괄 변경 + 푸터 액션 자리(`REF-device-move.md`, 격리 23/24 — 실패 1 = 주소 3개 상한 함정, BACKLOG)
  - 커밋함(2026-10-26 저녁, 사용자 "vault 정리 및 커밋") — master 머지는 아직
  - 이 PC 재빌드(사용자 진행 예정): 빌드명 추천 `api_prefix`(태그 `SUPERVISOR-/WORKER-API_PREFIX-261026`, `cmd/*/env`는 git 밖이라 태그는 손으로 되돌림). **supervisor + web만 교체, worker는 재시작 불필요**(이번 변경에 `cmd/worker`·`internal` 차이 없음, `/worker` 그대로, 끊기면 스스로 재접속) -> worker는 `take_process/worker`로 계속 돌므로 **`build/take_process/` 지우지 말 것**
  - 다음: 사용자 화면 확인 -> 파트 마감(master 머지) -> 배포 = **supervisor + 웹만**(둘을 같이, worker 그대로): test3 = 새 빌드 / 집 서버 = 집에서 원격으로 exe·web 교체(남겨 둔 `build/home-update/` 스크립트 재사용 가능, env 키는 이미 있음)
  - 파트(회원 관리) 마감 정리(MEMORY 불변 결정·`INDEX-member` 접기 — MEMORY 10.3k자라 이때 줄임)도 같이
  - `build/` 정리함(사용자 요청): 남은 것 = `take_process/`·`supervisor.env`·`worker.env`·`worker-sqlite.db*` + `home-update/`의 ps1 스크립트만. 옛 빌드본 4개·`supervisor_move`·집 서버 옛 zip 2개·`nexus-autostart.ps1`·`nexus-create-user.sql` 삭제(내용은 `REF-home-server.md`에 설명 남음). `supervisor_move/worker -hold`(pid 771160, 12:12 터미널)는 실행 파일만 지워졌고 계속 돎(take_process worker가 붙어 있음)
  - 격리 흔적: DB `nexus_api_tmp`(시험 계정들) 남김 — 필요 없으면 DROP
- **(2026-10-26 세션 끝, 사용자 "여기까지 오케이" -> 세션 재시작)** 다음 세션 첫 할 일 = 사용자가 test3에서 써 본 결과 듣기 -> 문제없으면 파트(회원 관리) 마감
- **(2026-10-26 세션 끝) 파트(회원 관리) = 코드·확인·커밋 끝, 사용자가 test3.falinux.com에 직접 새 배포함** — test3(8013) = **`build/take_process`**(사용자 빌드, `nexus_build` goose v14 = `00014` 적용 확인). 사용자 확인 결과 듣고 이어서
  - 브랜치: `feat/member`(worktree `/home/jh-bae/irony/nexus-member`) = `1bd58e7`(회원 관리 전부: 서버·UI 개편·컨트롤 패널·알약 끌기·빈 곳 휠·판 계약) -> `fb2fa56`(master 머지 = supervisor-move 포함) -> `fa50c64`(미등록 process 가져오기). master = `efbe403`(supervisor-move 머지) + vault. **push = 사용자가 함**(분기 태그는 따로 `git push origin <tag>`)
  - (해결된 듯 — 사용자가 `build/take_process`로 배포) 배포 중 막힘: 사용자가 원래 폴더에서 `git checkout feat/member` -> "already checked out at nexus-member". 제안 = ① worktree 폴더에서 바로 빌드(권장, env·`KEY_SECRET` 이미 있음) / ② `git worktree remove --force` 뒤 원래 폴더에서 체크아웃(`test.env`·`node_modules` 링크·env 복사본 지워짐 — 옮긴 뒤)
  - **관리자 지정 끝**: 사용자 "test3는 irony 닉네임 계정" -> `nexus_build`에 닉네임 `irony` 없음, 로그인 아이디 `irony` = 닉네임 `아이로니`(id 3, sha256 일치) -> 사용자 "관리자 지정해줘" -> Claude가 `UPDATE ... WHERE id = 3` 실행, perms = `{worker_register,admin}`
  - **보드 worker 교체 끝(2026-10-26, 사용자 요청)**: `fa24-PID`(transitBoard 보드 `root@192.168.40.203`, `/home/root/nexus-worker`) · `imx8mp-f24`(mirror-info 보드 `falinux@192.168.80.206`, `/opt/nexus-worker`, sudo) — `fa50c64` arm64 정적 + env `API_KEY` + `systemctl restart` -> 둘 다 test3 등록 성공·보관 process 다시 붙음. 기록 = 각 작업공간 vault(transitBoard `REF-board-nexus.md` · mirror-info `REF-board.md` "Nexus worker")
  - 배포 전 확인 나머지 → `REF-member.md` "배포 전 확인"(`KEY_SECRET`·worker 실행 파일 + `API_KEY`·기존 폴더 403). 새 supervisor DB면 장비를 **같은 키로** 등록해야 worker가 붙음(→ `REF-supervisor-move.md` "가져오기 구현" 함정)
  - 배포 뒤 = 파트(회원 관리) 마감: MEMORY 불변 결정 정리(판 계약·회원 권한 등) + `INDEX-member` 접기 + `feat/member` master 머지
  - 이번 세션 경위 → `history/member-ui.md`·`history/member-impl.md`·`history/supervisor-move.md` 2026-10-26 오후. 옛 줄(재부팅 직전·UI 계획·집 서버·supervisor-move 확인 대기)은 `history/current-archive.md` "2026-10-26 세션 끝"
- **지금 떠 있는 것**(2026-10-26 세션 끝): 파트(회원 관리) 격리 **전부 닫음**(5051~5053·3001~3003·worker `jig-a`) — DB `nexus_member_tmp`는 남김(다시 띄우는 법 `REF-member-impl.md`), 임시 `nexus_adopt_b`·`_c` 지움. **test3(8013) = `build/take_process`**(DB `nexus_build`, goose v14, 사용자가 2026-10-26 배포). 8012는 다른 프로젝트(`irony/community`) vite. test1(8011)·개발(8012) 꺼짐. 보관 프로세스 19096·19684·61491 = `build/keep/worker -hold` -> **`build/keep` 지우지 말 것**. VM `nexus-win` 꺼짐(→ `REF-windows-vm.md` "VM 실행"). 웹 백업(`web.bak*`)은 2026-10-25 전부 지움
- 빌드 = 태그(`apps/core/cmd/*/env` 첫 줄) 바꾸고 `scripts/build.sh <이름>`(태그 출력·같은 태그 경고) → `REF-build.md` "빌드 태그"
- 남은 흔적(정리 여부 사용자 미정): test1용 DB `nexus_keep`(시험 계정 `smoke<시각>` 1개)·`build/keep`의 env·sqlite(test1 = 재부팅 뒤 안 켬) · `nexus_build`에 `smoke1792672658797`·`smoke1792672663267`(폴더 233·스크립트 234). 보관 프로세스 일괄 종료 `pkill -f 'worker -hold'`는 **이 Claude 세션도 죽으니 금지**
- 두 빌드를 나란히 띄우는 법(env 분리·DB 복사) → `REF-build.md` "두 빌드 나란히"
- **나중에 할 것 = `BACKLOG.md`**(2026-10-22 신설, 세션 시작 시 안 읽음): 다음 파트 후보·알려진 버그·보류 설계·열린 질문·답 대기 / PC 재부팅은 다음 작업이 없을 때 확인
- **최근 마감 파트**(상세는 각 INDEX, 목록 = `HISTORY.md` "마감 파트"): supervisor 주소 실시간 변경 `INDEX-supervisor-move`(10-26) · 빈 곳 길게 누르기 `INDEX-press-pager`(10-25) · 목록 아이콘 보기 `INDEX-dir-grid`(10-24) · 타일 머리 줄 버튼 `INDEX-tile-head`(10-24) · Ctrl+Alt 팝업 버튼·터치 스크롤 양 `INDEX-pager-tools`(10-23) · 글자 크기 규칙 `INDEX-typography`(10-23) · 경로 직접 입력 `INDEX-path-input`(10-23) · 내려받기·worker 간 전송 `INDEX-transfer-down`(10-23) · 파일 전송 확장 `INDEX-transfer-ext`(10-23) · worker 폴더 타일 `INDEX-worker-dir`(10-23) · 타일 종류 정리 `INDEX-tile-kinds`(10-23) · process 살려 두기 `INDEX-process-keep`(10-22) · 새 보드·IUTF8 `INDEX-fix-board-iutf8`(10-22) · 빌드본 분리 `INDEX-build`(10-22) · 리팩터링 `INDEX-refactor`(10-21) · 페이지 이동 `INDEX-tile-move`(10-21) · 한글 입력 중복 `INDEX-term-ime` · 파일 전송 `INDEX-transfer-e2e` · 헤더·푸터 개편 `INDEX-board-chrome` · 페이지 표시 개편 `INDEX-board-pager`(10-20) · URL 바인딩 `INDEX-tile-url` · 터미널 터치 스크롤 `INDEX-term-touch` · Windows worker `INDEX-windows-worker` · PgUp·PgDn 버튼 `INDEX-term-page-keys` · 고아 스크립트 `INDEX-worker-orphan` · 보드 미리보기 `INDEX-board-preview`(10-15~19)
  - 남겨 둔 것: 확인용 노드 irony `test`/`ctrl-alt-test`(225)·`scroll-test`(215)·`scroll-test.ps1`(216) / VM `nexus-win`(`docker start/stop nexus-win`, 화면 `127.0.0.1:8006`) + VM worker `win-dev` → `REF-windows-vm.md` "VM 실행" / Windows 미확인 항목 → `BACKLOG.md`
- **시연 데이터(2026-10-16) 등록 → `REF-demo.md`**(시연 순서 점검은 `BACKLOG.md`): irony 계정 폴더 `시연-생산지그`(id 200, 안 C(생산 지그 시나리오)) + test1 편집·test2 조작·test3 보기 공유. 코드 변경 없음(REST로 데이터만, 커밋할 코드 없음)
- **외부 접속 설정 사용 중**: 프론트 https://test1.falinux.com(8011) · supervisor https://test2.falinux.com(8012) · worker -> localhost:8012 → `REF-external-access.md`. 로컬 전용(5050·3000)으로 되돌리는 법도 거기
- 남은 주의: 개발 DB에 테스트 계정 `active-<시각>` 3개(폴더 11개씩) / 개발용 로그인 기본값 = `apps/frontend/.env.development.local`(git 제외) / worker는 종료 정책 이후 빌드여야 force를 앎
- 커밋 상태(2026-10-26 세션 끝): 위 첫 줄(브랜치). 남겨 둔 브랜치 `feat/dir-grid`·`feat/press-pager`·`feat/chaltteok-ui`·`feat/supervisor-move`(머지됨)·`refactor/socket`·`refactor/tile-kinds`·`feat/worker-dir`·`feat/transfer-ext`·`feat/transfer-down`·`feat/path-input`·`feat/typography`·`feat/pager-tools`·`feat/tile-head`. 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치)·worktree `cmd/worker/test.env`. 커밋은 파트 단위(중간 커밋은 사용자가 요청할 때). user vault 전체 git 제외
- 테스트·확인 준비(환경, 헤드리스, 함정, **격리 실행 = 임시 DB + 5051**) → `REF-test-env.md`
- 역할 분담(2026-09-29 개정): node 타일 UI·process 작업은 코드 대부분을 Claude가 작성하고, 로직 구조는 사용자와 먼저 맞춘 뒤 쓴다


---

## 유지 사항 (항상)
- 색 규칙: Vuetify 테마 색 이름 의존, hex 금지 → `REF-frontend.md` "색 규칙" / `position_x/y` 사용 중단(DB 컬럼 유지) / 외부 레이아웃 라이브러리 안 씀 / 트리 드래그 device_key 재상속 이슈 → `REF-node-label.md`
- EXEC 실행 방식 = A(직접 실행, shebang 필수) 현상 유지 확정(2026-10-12) / kill 정책 → `REF-process-kill.md` "확정"
- 마이그레이션 번호는 착수 때 실제 다음 번호로(`00005`는 `process_subscribers` 제거에 씀)
- 모바일 키보드 주의: `appWindown.store.ts`의 `size.inner`(visualViewport)에 판정을 연결하면 안 됨 → `REF-node-ui-overview.md`
