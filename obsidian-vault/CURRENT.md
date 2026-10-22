# CURRENT

## 현재 날짜
2026-10-22

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **2026-10-23 출근 후 사용자와 같이 정리**(2026-10-22 밤 결정): ① 사용자가 test3를 자기 터미널에서 다시 켬(`CLAUDE*` 변수 상속 제거) ② test1 내릴지 결정 — 내리면 `kill 4124134 4112207`(PID 지정, `pkill` 금지), `build/keep` 폴더·바이너리는 test3가 써서 남김(env·sqlite·`nexus_keep`만) → `REF-build.md` "두 빌드 나란히" ③ `build/iutf` 삭제·시험 계정 정리 여부 (재배포-6(마감) 커밋은 2026-10-22 밤에 끝냄). 그때까지 `build/iutf`·시험 계정·test1 전부 그대로 둠
- **2026-10-22 밤 세션 = test3(8013)에서 접속 확인**(조상 worker 4178326 `-env ../worker.env`). 이 세션은 `CLAUDE*` 변수 상속으로 **transcript 저장 안 됨**(`/resume` 불가) -> 사용자가 변수 지우고 Claude를 다시 띄울 수 있음. 근본 해결 = test3 worker를 사용자 터미널에서 다시 띄우기 → `REF-build.md` "Claude가 띄운 빌드본". **사용자가 2026-10-23에 직접 다시 켜기로 함**(Claude는 띄우지 않음)
- 지금 떠 있는 것(2026-10-22 21:36~): test1 = supervisor 4124134·worker 4112207(`build/keep`, `-env ./*.env`, DB `nexus_keep`, sqlite `build/keep/worker-sqlite.db`) / test3 = supervisor 4178138·worker 4178326(`build/keep`, `-env ../*.env`, DB `nexus_build`, sqlite `build/worker-sqlite.db`, 로그 `build/keep/log/*-8013.log`). 옛 `build/iutf` 프로세스는 없음
- **지금: 파트(worker 폴더 타일) 구상·조율 단계(2026-10-22 착수, 코드 없음)** → `REF-worker-dir.md`. Q1~Q11 답 받음(별도 장비 타일·장비-1~3 규칙), 개정 큰 계획 폴더-1(목록 요청)·2(장비 타일)·3(여는 곳)·4(실행)·6(보드 이동)·5(업로드) 확인 대기 + **역할 분담(Claude가 코드 작성 vs 사용자 직접) 답 대기**. 안 나(업로드 기능)는 이 파트의 폴더-5(업로드)로 흡수
- **파트(process 살려 두기) 2026-10-22 마감·커밋**(→ `INDEX-process-keep.md`): 보관 프로세스(`worker -hold`)로 worker·supervisor 재시작에도 터미널·화면 유지. 미확인 K7(끝난 뒤 기다림) 10분 경과·K6(Windows)
- 남은 흔적: `nexus_keep`에 시험 계정 `smoke<시각>` 1개 · `nexus_build`에 `smoke1792672658797`·`smoke1792672663267`(폴더 233·스크립트 234) / test1에서 띄운 터미널은 보관 프로세스(`worker -hold`)로 남음(정리 = 해당 타일에서 종료, 일괄은 `pkill -f 'worker -hold'`인데 **test1 세션도 같이 죽으니 금지**)
- 두 빌드를 나란히 띄우는 법(env 분리·DB 복사) → `REF-build.md` "두 빌드 나란히"
- **파트(새 보드·IUTF8) 2026-10-22 마감·커밋**(→ `INDEX-fix-board-iutf8.md`): 어느 보드에서든 새 보드로 분리 + worker PTY `IUTF8`. 빌드본 브라우저 확인 OK
- **나중에 할 것 = `BACKLOG.md`**(2026-10-22 신설, 세션 시작 시 안 읽음): 다음 파트 후보·알려진 버그·보류 설계·열린 질문·답 대기 / PC 재부팅은 다음 작업이 없을 때 확인
- 그 전 파트(리팩터링) 2026-10-21 마감(→ `INDEX-refactor.md`): 브랜치 `refactor/socket`의 커밋 3개(소켓 1차 `8cf32bc`·소켓 2차 `0dc83a7`·소켓 밖 정리 `38a24a7`)를 master에 `--no-ff` 머지, 브랜치는 남겨 둠
  - 리팩터링 뒤 옛 문서의 `tile.go` 함수 위치가 바뀜(→ `REF-refactor-cleanup.md` "결과"), 키보드 맞춤-1(타일 띄우기)은 `hook/tileFloat.hook.ts`
  - 확인 스크립트(scratchpad)는 세션 끝나면 사라짐 → 명세 `REF-test-scripts-1020.md` "2026-10-21 소켓 알림 리팩터"
- **최근 마감 파트**(상세는 각 INDEX, 목록 = `HISTORY.md` "마감 파트"): process 살려 두기 `INDEX-process-keep`(10-22) · 새 보드·IUTF8 `INDEX-fix-board-iutf8`(10-22) · 빌드본 분리 `INDEX-build`(10-22) · 리팩터링 `INDEX-refactor`(10-21) · 페이지 이동 `INDEX-tile-move`(10-21) · 한글 입력 중복 `INDEX-term-ime` · 파일 전송 `INDEX-transfer-e2e` · 헤더·푸터 개편 `INDEX-board-chrome` · 페이지 표시 개편 `INDEX-board-pager`(10-20) · URL 바인딩 `INDEX-tile-url` · 터미널 터치 스크롤 `INDEX-term-touch` · Windows worker `INDEX-windows-worker` · PgUp·PgDn 버튼 `INDEX-term-page-keys` · 고아 스크립트 `INDEX-worker-orphan` · 보드 미리보기 `INDEX-board-preview`(10-15~19)
  - 남겨 둔 것: 확인용 노드 irony `test`/`ctrl-alt-test`(225)·`scroll-test`(215)·`scroll-test.ps1`(216) / VM `nexus-win`(`docker start/stop nexus-win`, 화면 `127.0.0.1:8006`) + VM worker `win-dev` → `REF-windows-worker.md` "VM 실행" / Windows 미확인 항목 → `BACKLOG.md`
- **시연 데이터(2026-10-16) 등록 → `REF-demo.md`**(시연 순서 점검은 `BACKLOG.md`): irony 계정 폴더 `시연-생산지그`(id 200, 안 C(생산 지그 시나리오)) + test1 편집·test2 조작·test3 보기 공유. 코드 변경 없음(REST로 데이터만, 커밋할 코드 없음)
- **외부 접속 설정 사용 중**: 프론트 https://test1.falinux.com(8011) · supervisor https://test2.falinux.com(8012) · worker -> localhost:8012 → `REF-external-access.md`. 로컬 전용(5050·3000)으로 되돌리는 법도 거기
- 남은 주의: 개발 DB에 테스트 계정 `active-<시각>` 3개(폴더 11개씩) / 개발용 로그인 기본값 = `apps/frontend/.env.development.local`(git 제외) / worker는 종료 정책 이후 빌드여야 force를 앎
- 커밋 상태(2026-10-21): 마감 파트는 전부 master에 커밋·머지(파트별 코드 + vault). master 최신 = 파트(빌드본 분리) 커밋(2026-10-22, master에 바로), 그 앞이 파트(리팩터링) `--no-ff` 머지(브랜치 `refactor/socket` 남김). **push는 한 번도 안 함**(분기 태그 `checkpoint/before-folder-tile-uid` = `d50f153`, 태그는 push 때 따로 `git push origin <tag>`). 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치). 커밋은 파트 단위(작업 단위마다 안 함, 사용자가 중간 커밋을 요청하면 그때). user vault 전체 git 제외
- 이전 세션(2026-10-12~15)의 scratchpad 스크립트는 이미 사라짐 — 명세는 `REF-test-scripts.md`(~10-13)·`REF-test-scripts-1014.md`(10-14~)·`REF-test-scripts-1020.md`(10-20~) 날짜별 절. 그때 격리 프로세스·임시 DB는 전부 정리됨. PC 재부팅 뒤라면 `docker start postgres15` -> supervisor -> worker -> `npm run dev`(→ MEMORY "로컬 개발 환경")
- 테스트·확인 준비(환경, 헤드리스, 함정, **격리 실행 = 임시 DB + 5051**) → `REF-test-env.md`
- 역할 분담(2026-09-29 개정): node 타일 UI·process 작업은 코드 대부분을 Claude가 작성하고, 로직 구조는 사용자와 먼저 맞춘 뒤 쓴다


---

## 유지 사항 (항상)
- 색 규칙: Vuetify 테마 색 이름 의존, hex 금지 → `REF-frontend.md` "색 규칙" / `position_x/y` 사용 중단(DB 컬럼 유지) / 외부 레이아웃 라이브러리 안 씀 / 트리 드래그 device_key 재상속 이슈 → `REF-node-label.md`
- EXEC 실행 방식 = A(직접 실행, shebang 필수) 현상 유지 확정(2026-10-12) / kill 정책 → `REF-process-kill.md` "확정"
- 마이그레이션 번호는 착수 때 실제 다음 번호로(`00005`는 `process_subscribers` 제거에 씀)
- 모바일 키보드 주의: `appWindown.store.ts`의 `size.inner`(visualViewport)에 판정을 연결하면 안 됨 → `REF-node-ui-overview.md`
