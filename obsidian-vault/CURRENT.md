# CURRENT

## 현재 날짜
2026-10-24

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **2026-10-24 파트(타일 머리 줄 버튼) 마감·master 머지**(→ `INDEX-tile-head.md`): 머리 줄·경로 줄 버튼을 알약·끝 블록으로 통일 + 경로 줄 좁으면 두 줄. 사용자 test1 확인 OK. test3 `build/ctrl_panel/web`만 새 빌드로 교체(옛 것 = `web.bak`). 브랜치 `feat/tile-head` 남겨 둠, push 안 함. **vault는 `feat/supervisor-move` 것을 옮겨 와 같이 커밋됨**(supervisor-move REF·history 포함) — 그 브랜치를 master에 머지할 때 vault 충돌(CURRENT·HISTORY·MEMORY)은 master 쪽을 기준으로 정리
- **2026-10-24 파트(supervisor 주소 실시간 변경) 구현·격리 확인 끝 — 사용자 확인 대기**(→ `REF-supervisor-move.md` "구현"·"확인"): 브랜치 `feat/supervisor-move`(커밋 전). 주소-1(주소 목록)~주소-5(화면) 구현, 스크립트 이름 표시 추가(사용자 요청). 격리 확인 서버 29/29·화면 8/8, 격리 환경 정리함. **사용자 브라우저 확인 대기 — 2026-10-24 주말 외부라 다음 주에 확인**(사용자 요청으로 브랜치에 중간 커밋(`wip(supervisor-move)`), master 머지 안 함 — 확인 뒤 마감 커밋·머지). 돌아오면 확인할 것: ① 빌드 뒤 장비 타일 ⋮ 메뉴 "supervisor 주소 변경"·"미등록 process" ② 권한 SQL(`REF-supervisor-move.md` 00013 주석) 뒤 푸터 보드 메뉴 "전체 worker 주소 변경" ③ 빌드본 교체 주의 = `REF-build.md` 빌드본 적용 때 supervisor `00013`·worker `00003`·`00004` 마이그레이션이 붙음
- **2026-10-23 파트 2개 마감·master 머지**: 파트(타일 종류 정리) → `INDEX-tile-kinds.md` / 파트(worker 폴더 타일) → `INDEX-worker-dir.md`. 브랜치 `refactor/tile-kinds`·`feat/worker-dir` 남겨 둠, push 안 함
- **2026-10-23 파트(파일 전송 확장)·파트(내려받기·worker 간 전송) 마감·master 머지**(→ `INDEX-transfer-ext.md`·`INDEX-transfer-down.md`): 장비 타일 업로드·전송 상태 + 내려받기(파일·zip)·worker 간 복사(메뉴·끌어 놓기). 사용자 브라우저 확인 OK. 브랜치 `feat/transfer-ext`·`feat/transfer-down`·`feat/path-input` 남겨 둠, push 안 함
- **2026-10-23 파트(Ctrl+Alt 팝업 버튼·터치 스크롤 양) 마감·master 머지**(→ `INDEX-pager-tools.md`): Ctrl+Alt 팝업에 크기·전체보기 버튼 + 터치 길게 누르기 + 터미널 터치 스크롤 2배. 사용자가 이 브랜치로 빌드해 PC·폰 확인 OK(빌드 이름 미확인). 브랜치 `feat/pager-tools` 남겨 둠, push 안 함. 격리 환경(5051·3001·`nexus_pager_tmp`) 정리함. **다음 파트 = 미정**(후보 `BACKLOG.md`)
- **2026-10-23 파트(글자 크기 규칙) 마감·master 머지**(→ `INDEX-typography.md`): 단계 이름 mixin + 단계 값 한 곳 + 배율 손잡이, 터미널 14px. 브랜치 `feat/typography` 남겨 둠
- **2026-10-23 파트(경로 직접 입력) 마감·master 머지**(→ `INDEX-path-input.md`): 장비 타일·폴더 타일 경로 직접 입력 + 단계별 자동완성. 사용자 브라우저(Linux·Windows) 확인 OK. 브랜치 `feat/path-input` 남겨 둠, push 안 함
- **지금 떠 있는 것**: (2026-10-24) **test1(8011) = 임시 대비용으로 계속 띄워 둠(사용자 요청)** — 이 저장소 vite 개발 서버(작업 트리 = 지금 브랜치, `VITE_API_URL=https://test3.falinux.com`을 명령줄로 덮어 실행 — `.env`는 그대로 test2) -> API·데이터는 test3 빌드본 supervisor(8013, `nexus_build`). **8012는 지금 다른 프로젝트(`irony/community`) vite가 씀** / test3(8013) = `build/directory_path`(태그 `DIRECTORYPATH-261023`, 사용자가 빌드·실행, 경로 직접 입력 포함. `build/download`도 사용자 빌드). 보관 프로세스 19096·19684·61491 = `build/keep/worker -hold` -> **`build/keep` 지우지 말 것**. test1(8011)·개발(8012) 꺼짐. VM `nexus-win` 꺼짐(2026-10-23 확인 뒤 끔. 켜면 `start-test3.ps1`로 `win-dev`를 test3에 붙임, exe = `DIRECTORYPATH-261023` + `C:` 고침)(→ `REF-windows-vm.md` "VM 실행"). 이 세션 격리 환경(5051·3001·`nexus_xferext_tmp`·`nexus_xferdown_tmp`·`nexus_pathin_tmp`)은 정리함
- 빌드 = 태그(`apps/core/cmd/*/env` 첫 줄) 바꾸고 `scripts/build.sh <이름>`(태그 출력·같은 태그 경고) → `REF-build.md` "빌드 태그"
- **파트(process 살려 두기) 2026-10-22 마감·커밋**(→ `INDEX-process-keep.md`): 보관 프로세스(`worker -hold`)로 worker·supervisor 재시작에도 터미널·화면 유지. 미확인 K7(끝난 뒤 기다림) 10분 경과·K6(Windows)
- 남은 흔적(정리 여부 사용자 미정): test1용 DB `nexus_keep`(시험 계정 `smoke<시각>` 1개)·`build/keep`의 env·sqlite(test1 = 재부팅 뒤 안 켬) · `nexus_build`에 `smoke1792672658797`·`smoke1792672663267`(폴더 233·스크립트 234). 보관 프로세스 일괄 종료 `pkill -f 'worker -hold'`는 **이 Claude 세션도 죽으니 금지**
- 두 빌드를 나란히 띄우는 법(env 분리·DB 복사) → `REF-build.md` "두 빌드 나란히"
- **파트(새 보드·IUTF8) 2026-10-22 마감·커밋**(→ `INDEX-fix-board-iutf8.md`): 어느 보드에서든 새 보드로 분리 + worker PTY `IUTF8`. 빌드본 브라우저 확인 OK
- **나중에 할 것 = `BACKLOG.md`**(2026-10-22 신설, 세션 시작 시 안 읽음): 다음 파트 후보·알려진 버그·보류 설계·열린 질문·답 대기 / PC 재부팅은 다음 작업이 없을 때 확인
- 그 전 파트(리팩터링) 2026-10-21 마감(→ `INDEX-refactor.md`): 브랜치 `refactor/socket`의 커밋 3개(소켓 1차 `8cf32bc`·소켓 2차 `0dc83a7`·소켓 밖 정리 `38a24a7`)를 master에 `--no-ff` 머지, 브랜치는 남겨 둠
  - 리팩터링 뒤 옛 문서의 `tile.go` 함수 위치가 바뀜(→ `REF-refactor-cleanup.md` "결과"), 키보드 맞춤-1(타일 띄우기)은 `hook/tileFloat.hook.ts`
  - 확인 스크립트(scratchpad)는 세션 끝나면 사라짐 → 명세 `REF-test-scripts-1020.md` "2026-10-21 소켓 알림 리팩터"
- **최근 마감 파트**(상세는 각 INDEX, 목록 = `HISTORY.md` "마감 파트"): 타일 머리 줄 버튼 `INDEX-tile-head`(10-24) · Ctrl+Alt 팝업 버튼·터치 스크롤 양 `INDEX-pager-tools`(10-23) · 글자 크기 규칙 `INDEX-typography`(10-23) · 경로 직접 입력 `INDEX-path-input`(10-23) · 내려받기·worker 간 전송 `INDEX-transfer-down`(10-23) · 파일 전송 확장 `INDEX-transfer-ext`(10-23) · worker 폴더 타일 `INDEX-worker-dir`(10-23) · 타일 종류 정리 `INDEX-tile-kinds`(10-23) · process 살려 두기 `INDEX-process-keep`(10-22) · 새 보드·IUTF8 `INDEX-fix-board-iutf8`(10-22) · 빌드본 분리 `INDEX-build`(10-22) · 리팩터링 `INDEX-refactor`(10-21) · 페이지 이동 `INDEX-tile-move`(10-21) · 한글 입력 중복 `INDEX-term-ime` · 파일 전송 `INDEX-transfer-e2e` · 헤더·푸터 개편 `INDEX-board-chrome` · 페이지 표시 개편 `INDEX-board-pager`(10-20) · URL 바인딩 `INDEX-tile-url` · 터미널 터치 스크롤 `INDEX-term-touch` · Windows worker `INDEX-windows-worker` · PgUp·PgDn 버튼 `INDEX-term-page-keys` · 고아 스크립트 `INDEX-worker-orphan` · 보드 미리보기 `INDEX-board-preview`(10-15~19)
  - 남겨 둔 것: 확인용 노드 irony `test`/`ctrl-alt-test`(225)·`scroll-test`(215)·`scroll-test.ps1`(216) / VM `nexus-win`(`docker start/stop nexus-win`, 화면 `127.0.0.1:8006`) + VM worker `win-dev` → `REF-windows-vm.md` "VM 실행" / Windows 미확인 항목 → `BACKLOG.md`
- **시연 데이터(2026-10-16) 등록 → `REF-demo.md`**(시연 순서 점검은 `BACKLOG.md`): irony 계정 폴더 `시연-생산지그`(id 200, 안 C(생산 지그 시나리오)) + test1 편집·test2 조작·test3 보기 공유. 코드 변경 없음(REST로 데이터만, 커밋할 코드 없음)
- **외부 접속 설정 사용 중**: 프론트 https://test1.falinux.com(8011) · supervisor https://test2.falinux.com(8012) · worker -> localhost:8012 → `REF-external-access.md`. 로컬 전용(5050·3000)으로 되돌리는 법도 거기
- 남은 주의: 개발 DB에 테스트 계정 `active-<시각>` 3개(폴더 11개씩) / 개발용 로그인 기본값 = `apps/frontend/.env.development.local`(git 제외) / worker는 종료 정책 이후 빌드여야 force를 앎
- 커밋 상태(2026-10-24): master 최신 = 파트(타일 머리 줄 버튼) 머지. `feat/supervisor-move`는 중간 커밋(머지 안 함). 옛 브랜치 `refactor/socket`·`refactor/tile-kinds`·`feat/worker-dir`·`feat/transfer-ext`·`feat/transfer-down`·`feat/path-input`·`feat/typography`·`feat/pager-tools`·`feat/tile-head` 남겨 둠. **push는 한 번도 안 함**(분기 태그 `checkpoint/before-folder-tile-uid` = `d50f153`, 태그는 push 때 따로 `git push origin <tag>`). 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치). 커밋은 파트 단위(중간 커밋은 사용자가 요청할 때). user vault 전체 git 제외
- 이전 세션(2026-10-12~15)의 scratchpad 스크립트는 이미 사라짐 — 명세는 `REF-test-scripts.md`(~10-13)·`REF-test-scripts-1014.md`(10-14~)·`REF-test-scripts-1020.md`(10-20~)·`REF-test-scripts-1022.md`(10-22~) 날짜별 절. 그때 격리 프로세스·임시 DB는 전부 정리됨. PC 재부팅 뒤라면 `docker start postgres15` -> supervisor -> worker -> `npm run dev`(→ MEMORY "로컬 개발 환경")
- 테스트·확인 준비(환경, 헤드리스, 함정, **격리 실행 = 임시 DB + 5051**) → `REF-test-env.md`
- 역할 분담(2026-09-29 개정): node 타일 UI·process 작업은 코드 대부분을 Claude가 작성하고, 로직 구조는 사용자와 먼저 맞춘 뒤 쓴다


---

## 유지 사항 (항상)
- 색 규칙: Vuetify 테마 색 이름 의존, hex 금지 → `REF-frontend.md` "색 규칙" / `position_x/y` 사용 중단(DB 컬럼 유지) / 외부 레이아웃 라이브러리 안 씀 / 트리 드래그 device_key 재상속 이슈 → `REF-node-label.md`
- EXEC 실행 방식 = A(직접 실행, shebang 필수) 현상 유지 확정(2026-10-12) / kill 정책 → `REF-process-kill.md` "확정"
- 마이그레이션 번호는 착수 때 실제 다음 번호로(`00005`는 `process_subscribers` 제거에 씀)
- 모바일 키보드 주의: `appWindown.store.ts`의 `size.inner`(visualViewport)에 판정을 연결하면 안 됨 → `REF-node-ui-overview.md`
