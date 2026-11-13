# HISTORY (인덱스)

> 상세 작업 기록은 **주제/모듈별로 `history/` 폴더에 분할**. 이 파일은 어느 파일에 뭐가 있는지 색인만.
> 요약·재사용 지식은 `REF-*.md`, 현재 진행은 `CURRENT.md`.

## 상황 색인 (INDEX — 그 상황일 때만 연다)
- 빌드·배포(빌드 스크립트·test3 교체·집 서버 묶음·보드 worker·집 서버 접속) → `INDEX-deploy.md`

## 마감 파트 (INDEX — 세션 시작 시 안 읽음)
- 통신 인프라(마감) → `INDEX-infra.md` (transport·transfer·supervisor-web·project)
  - 2026-10-13 `history/supervisor-web.md`에 파트(서버 정리) 정리-2(세션 키 env)·정리-3(createdAt)·정리-4(확인) → 파트(서버 정리) 마감·커밋
- process 배선(마감) → `INDEX-process-wiring.md` (process-wiring·trigger·resize·reconnect·subscription)
- node UI 레이아웃 설계(마감) → `INDEX-node-ui-layout.md` (node-ui-layout·projection·impl·render)
- node UI 연동(마감) → `INDEX-node-ui-link.md` (node-ui-link·terminal·save·save-impl·sync)
- process 터미널 기능(마감) → `INDEX-process-terminal.md` (process-input·snapshot·snapshot-impl·exec-edit)
  - 2026-10-14 `history/process-input.md`에 실측-1(PENDING 입력 버림) 격리 확인 7/7 — I3 확인 끝
- process 동기화·공유·보드(마감) → `INDEX-process-share.md` (process-sync(+impl)·process-share 13개·workspace 4개)
  - 2026-10-14 `history/process-sync-impl.md`에 실측-2(탭 보관 60초) 격리 확인 9/9(presence 빠짐 60.1초)
- **★ 모바일 가상 키보드(마감, 중요 — 다른 프로젝트에서도 생김) → `INDEX-mobile-keyboard.md`** (mobile-keyboard·반복 주제 mobile-viewport)
- 터미널 동작 보강(마감) → `INDEX-terminal-ops.md` (process-kill·tile-binding 3개·term-query·size-take)
- 노드 조작(마감) → `INDEX-node-ops.md` (node-label·node-list-sync·node-clipboard·node-select)
- 프론트 마무리(마감) → `INDEX-frontend-polish.md` (node-ui-active·frontend-auth)
- 보드 미리보기(마감) → `INDEX-board-preview.md` (board-preview)
  - 2026-10-26 `history/board-preview.md`에 폰 전체보기 가로 한 줄(EV2 개정, `fix/phone-layout`)
- 고아 스크립트(마감) → `INDEX-worker-orphan.md` (worker-orphan)
- PgUp·PgDn 버튼(마감) → `INDEX-term-page-keys.md` (term-page-keys)
- Windows worker(마감) → `INDEX-windows-worker.md` (windows-worker·windows-conpty·windows-exec)
- 터미널 터치 스크롤(마감) → `INDEX-term-touch.md` (term-touch)
- URL 바인딩(마감) → `INDEX-tile-url.md` (tile-url)
- 페이지 표시 개편(마감) → `INDEX-board-pager.md` (board-pager)
  - 2026-11-11 `history/board-pager.md`에 미니맵 진행도(tile 모드 끄는 양만큼 active 칸이 이웃 쪽으로, `feat/peek-nudge`)
- 파일 전송(마감) → `INDEX-transfer-e2e.md` (transfer-e2e)
- 한글 입력 중복(마감) → `INDEX-term-ime.md` (term-ime)
- 헤더·푸터 개편(마감) → `INDEX-board-chrome.md` (board-chrome·board-keyboard)
  - 2026-11-12 `history/board-chrome.md`에 푸터 배지 잘림(`.foot-badge`, `fix/footer-badge`)
- 페이지 이동(마감) → `INDEX-tile-move.md` (tile-move)
- 리팩터링(마감) → `INDEX-refactor.md` (refactor-socket·refactor-socket-batches·refactor-cleanup)
- 빌드본 분리(마감) → `INDEX-build.md` (build)
  - 2026-11-12 `history/build.md`에 test3 worker 교체(옛 빌드 폴더 삭제로 새 process 실행 실패 -> 새 빌드 worker로)
- 새 보드·IUTF8(마감) → `INDEX-fix-board-iutf8.md` (fix-board-iutf8)
- process 살려 두기(마감) → `INDEX-process-keep.md` (process-keep)
- 타일 종류 정리(마감) → `INDEX-tile-kinds.md` (tile-kinds)
- worker 폴더 타일(마감) → `INDEX-worker-dir.md` (worker-dir-list·worker-dir-exec·worker-dir-board·전체 계획 worker-dir)
- 파일 전송 확장(마감, 업로드·전송 상태) → `INDEX-transfer-ext.md` (transfer-ext)
- 내려받기·worker 간 전송(마감) → `INDEX-transfer-down.md` (transfer-down)
- 경로 직접 입력(마감) → `INDEX-path-input.md` (path-input)
  - 2026-10-26 `history/path-input.md`에 장비 타일 자동완성 비침(`.crumbs` z-index, `fix/phone-layout`)
- 글자 크기 규칙(마감) → `INDEX-typography.md` (typography)
- Ctrl+Alt 팝업 버튼·터치 스크롤 양(마감) → `INDEX-pager-tools.md` (pager-tools)
- 타일 머리 줄 버튼(마감) → `INDEX-tile-head.md` (tile-head)
- 목록 아이콘 보기(마감) → `INDEX-dir-grid.md` (dir-grid)
- 빈 곳 길게 누르기(마감) → `INDEX-press-pager.md` (press-pager)
- supervisor 주소 실시간 변경(마감, 후속 가져오기는 회원 관리 브랜치) → `INDEX-supervisor-move.md` (supervisor-move)
  - 2026-11-04 `history/supervisor-move.md`에 주소 순위 눌러 입력(`feat/addr-rank-pick`)
- 회원 관리(마감, `/api` 정리·열린 장비 주소 일괄 변경·폰 수정 포함) → `INDEX-member.md` (member·member-impl·member-ui·member-ui-impl·api-prefix·device-move)
  - 2026-11-03 `history/member-ui.md`에 화-7(채우는 순서) 개정 — Grid 폭 = 쓴 열 수(배치 버그, `feat/update-files-tile`에서)
- 젤리·관리 기록(마감, 젤리 효과·보드로·관리 페이지 replace) → `INDEX-jelly.md` (jelly)
- 폴더에서 터미널(마감, 장비 타일 "여기서 셸") → `INDEX-dir-shell.md` (dir-shell)
- 장비 상태 push(마감, `DEVICE:CHANGE` 신호 + 폴링 60초) → `INDEX-device-push.md` (device-push)
- 칸 메뉴 장비 디렉토리(마감, 폴더 칸 ⋮ "장비 디렉토리 열기") → `INDEX-row-open-dir.md` (row-open-dir)
- 찰떡 마무리(마감, 경로 줄 여백·관리자 필터·남은 창 5개·상세 상태 칸) → `INDEX-ui-chaltteok.md` (ui-chaltteok)
- 찰떡 입력칸(마감, 입력칸·고르기·체크 = Ct* 컴포넌트 + 관리 타일 포커스 선) → `INDEX-ct-field.md` (ct-field)
- 경로 줄 정리(마감, 경로 줄 버튼 묶음·이동 버튼 경로 왼쪽·가로 스크롤) → `INDEX-path-row.md` (path-row)
- 패널 창 버튼(마감, 컨트롤 패널에 active 타일 [숨기기|크기|닫기]·위치 글자 빼기·도트 최소 폭, 미해결 1건) → `INDEX-pager-head-btns.md` (pager-head-btns)
- 닫은 뒤 active(마감, active 닫으면 위치 번호 이웃 active, 보드·관리) → `INDEX-close-active.md` (close-active)
- 파일 편집기(마감, 10-28) → `INDEX-file-view.md` (file-view)
- 보기 타일(마감, 10-28) → `INDEX-view-tile.md` (view-tile)
- 영상 Range(마감, 10-29) → `INDEX-video-view.md` (video-view)
- 타일 끌기(마감, 10-29) → `INDEX-tile-drag.md` (tile-drag·tile-drag-impl·tile-order)
- 내 장비·관리자 타일 이동(마감, 10-30) → `INDEX-member-drag.md` (member-drag)
- 업데이트 app(마감, 10-30~11-04, 설계·P1~P4·Files 타일·시험 실행·갈아치우기 + 10-30 점검) → `INDEX-update-app.md` (update-*·review-1030)
- 공용 타일 틀(마감, 11-07, 공용 틀·메뉴 줄·설정·푸터 오른쪽) → `INDEX-tile-frame.md` (tile-frame·settings-footer)

## 주제별 history 색인 (공용 주제 · 진행 중 파트)
> 날짜 엔트리는 각 파일에. 여기는 파일·주제·마지막 기록만.

| 파일 | 주제 | 마지막 기록 |
|---|---|---|
| `history/account.md` | 계정: 로그인 기기·비밀번호 변경·평문 아이디(11-12 신설) ↔ `REF-account.md` | 11-12 조율 · 구현 · 확인(API 44·화면 23) · 사용자 확인 · 마감 · 11-13 로그인 자동완성 끄기(`fix/login-autocomplete`) |
| `history/column-flow.md` | 칼럼 흐름: Grid 경계 걸침(11-11 신설) ↔ `REF-column-flow.md` | 11-11 조율·보드 구현·URL focus·관리 페이지·헤드리스 확인 · 사용자 확인 · 마감 |
| `history/settings-account.md` | 설정 두 겹: 계정 + 이 브라우저(11-08 신설) ↔ `REF-settings-account.md` | 11-08 조율·구현·확인 24/24 · 사용자 확인 · 마감 · test3·집 서버 배포 |
| `history/socket-reconnect.md` | 소켓 재연결 · 끊김 표시(11-06 신설) ↔ `REF-socket-reconnect.md` | 11-06 확인 14/14 · test3 교체 · 모바일 확인 · 마감 |
| `history/board-close.md` | 보드 닫기 · 종료 유예(2026-11-05 신설) ↔ `REF-board-close.md` | 11-05 확인 13/13·9/9 · 사용자 확인 · 마감 |
| `history/decisions.md` | 확정 결정 상세(2026-10-26 MEMORY 슬림화 때 신설) ↔ `REF-decisions.md` | 10-26 밤 MEMORY 9.9k -> 5.8k자 |
| `history/decisions-ui.md` | 확정 결정 상세: 화면(2026-10-29 분할) ↔ `REF-decisions-ui.md` | 10-29 터치 제스처·끌기(파트(타일 끌기) 마감) |
| `history/current-archive-8.md` | CURRENT에서 내려보낸 줄 원문(2026-11-12~) | 11-12: 11-11 미니맵·칼럼 흐름 마감·빌드 줄, 11-12 푸터 배지 줄 |
| `history/current-archive-7.md` | CURRENT에서 내려보낸 줄 원문(2026-11-05~) | 11-08: 11-07 파트(공용 타일 틀) 마감 줄·11-07 빌드 줄 |
| `history/current-archive-6.md` | CURRENT에서 내려보낸 줄 원문(2026-11-02~) | 11-02 업데이트 app 설계 줄 · 11-03 P1~P4 · 11-03 밤 Files 타일 개선 마감 때 진행 블록 |
| `history/current-archive.md` | CURRENT에서 내려보낸 줄 원문(REF 없음) | 10-29 영상 파트 마감 때 10-28 밤 ~ 10-26 밤 줄 |
| `history/current-archive-5.md` | 위 파일에서 분할(2026-10-30: 10-29 파트 마감 3건·떠 있는 포트 표·build 흔적) | 10-30 업데이트 app 세션 끝 |
| `history/current-archive-4.md` | 위 파일에서 분할(2026-10-29: 10-27 ~ 10-28 파트(경로 줄 정리)) | 10-28 세션 끝(집 서버 직통 전환 줄) |
| `history/current-archive-3.md` | 위 파일에서 분할(2026-10-28: 10-26 밤 fix/phone-layout ~ vault 색인 정리) | 10-26 밤 vault 색인 정리 |
| `history/current-archive-2.md` | 위 파일에서 분할(2026-10-26 밤 세션 끝) | 10-26 밤 세션 끝 |
| `history/current-archive-1.md` | 위 파일에서 분할(2026-10-26 낮까지: 재부팅 전·세션 끝 정리) | 10-26 세션 끝 |
| `history/windows-vm.md` | Windows VM `nexus-win`(2026-10-23 `windows-worker`에서 분할) | 10-23 경로 직접 입력 확인, win-dev를 test3에 다시 붙임 |
| `history/demo.md` | 시연 데이터(2026-10-16 신설) | 10-16 안 C(생산 지그 시나리오) 등록, 브라우저 경유 실행 미확인 |
| `history/design-chaltteok.md` | 찰떡 규칙(디자인 규칙, 2026-10-25 신설) | 11-12 알림 drawer·토스트·전송 찰떡(`feat/notice-chaltteok`), `.tile-btn-row-end` 빠뜨림 정정 |
| `history/worker-screen.md` | worker 화면 송출(아이디어) | 10-15 나(VNC 터널 + noVNC) 추천 · 10-25 VNC 서버 준비 방식(내려받기·설치) 재논의, 보류 |
| `history/node-ui.md` | node 카탈로그 UI 컨셉 | 08-10 그리드 직접 제작 확정, termspace 영감 피벗 |
| `history/node-ui-overview.md` | 전체보기 · 화면 내비게이션 | 10-06 active 타일 표시 = 나(약하게 흐림) + 위치 번호 단축키 |
| `history/frontend.md` | 프론트엔드 | 10-13 헤더 전체 화면 버튼 · vue-query queryKey 읽기 규칙 |
| `history/realtime.md` | 실시간 push | 10-01 Hub 막힘 -> ping/pong(`e395433`) |
| `history/widget.md` | 위젯 | 08-07 StickyBox · Skeleton |
| `history/util.md` | 범용 유틸(반복 주제) | 08-07 GroupedSet · 공유 리사이즈 관측 그룹 |
| `history/util-drag.md` | 드래그 인프라 | 10-12 보류 결정 |
| `history/test-env.md` | 테스트 환경 · 확인용 클라이언트 | 10-15 고아 스크립트 격리 정리 절차 · 11-11 8012 점유 시 5051 |
| `history/test-isolation.md` | 격리 실행(11-11 `test-env`에서 분할) ↔ `REF-test-isolation.md` | 10-08 공유-e 격리 확인 함정 |
| `history/photo-upload.md` | 폰 사진 올리기 도구(2026-11-02 `REF-test-env`에서 분할) ↔ `REF-photo-upload.md` | 11-02 분할 |
| `history/test-scripts.md` · `-1014.md` · `-1020.md` · `-1022.md` · `-1023.md` | 확인 스크립트 명세(날짜별) | 10-23 글자-5 전후 스크린샷 `ty/pw` |
| `history/external-access.md` | 외부 접속 · HTTPS 프록시 | 10-13 마감·커밋 |
