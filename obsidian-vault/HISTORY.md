# HISTORY (인덱스)

> 상세 작업 기록은 **주제/모듈별로 `history/` 폴더에 분할**. 이 파일은 어느 파일에 뭐가 있는지 색인만.
> 요약·재사용 지식은 `REF-*.md`, 현재 진행은 `CURRENT.md`.

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
- 고아 스크립트(마감) → `INDEX-worker-orphan.md` (worker-orphan)
- PgUp·PgDn 버튼(마감) → `INDEX-term-page-keys.md` (term-page-keys)
- Windows worker(마감) → `INDEX-windows-worker.md` (windows-worker·windows-conpty·windows-exec)
- 터미널 터치 스크롤(마감) → `INDEX-term-touch.md` (term-touch)
- URL 바인딩(마감) → `INDEX-tile-url.md` (tile-url)
- 페이지 표시 개편(마감) → `INDEX-board-pager.md` (board-pager)
- 파일 전송(마감) → `INDEX-transfer-e2e.md` (transfer-e2e)
- 한글 입력 중복(마감) → `INDEX-term-ime.md` (term-ime)

## history/ 파일 색인 (진행 중 파트)

### `history/demo.md` — 시연 데이터 (2026-10-16 신설)
- 2026-10-16 플랜 안 A·B·C 제시 → 안 C(생산 지그 시나리오) 선택, 스크립트 6개 + test1~3 공유 등록(REST), 브라우저 경유 실행 미확인

### `history/worker-screen.md` — worker 화면 송출 (2026-10-15 신설, 아이디어)
- 2026-10-15 방식 가·나·다 비교, 조작 필요 → 나(VNC 터널 + noVNC) 추천, 기록만

### `history/node-ui.md` — node 카탈로그 UI 컨셉
- 2026-08-10(4) 그리드 UI 구현 방식: 직접 제작 확정(조사+설계) — Splitpanes/Golden Layout/Dockview 조사 후 기각, 리사이즈만 DraggableSession 재사용 (2026-08-12 재확인, 결론 불변)
- 2026-08-10(3) `position_x/y` 사용 중단 확정(코드 없음) — 폴더 타일은 ord 기반 list/grid만, DB 컬럼은 유지·값만 미사용
- 2026-08-10 termspace 영감 그리드로 전면 피벗(코드 없음): 깊이2 고정 그리드(열/행 각 최대2, 비대칭 허용) 확정, 옛 캔버스+트리 Phase1-5 컨셉 폐기
- 2026-08-07(2) 구현 Phase 분할(1~5)+라우팅 스킴(`/nodes/:parentId`) 확정(코드 없음, 2026-08-10 폐기됨), Draggable 리팩터 설계로 이어짐
- 2026-08-07 컨셉 설계(코드 없음, 2026-08-10 상당부분 폐기): 배치도(캔버스)=홈+트리=보조내비, PC 오버레이 토글/모바일 세그먼트 토글, 트리 재배치 드래그(device_key 상속변경·좌표NULL 이슈 발견, 이 지식은 계승됨). 모바일 트리 형태만 미정

### `history/node-ui-overview.md` — node 카탈로그 전체보기 · 화면 내비게이션 (node-ui-layout에서 2026-09-29 분할)
- 2026-10-06 active 타일 표시 확정 = 나(약하게 흐림) + 위치 번호 단축키, 코드 없음
- 2026-10-01(2) ⑮(모바일 가상 키보드) 방향 = A+B(타일 UI 유지 + 특수 키 줄) 확정, iOS 실기기 확인은 나중
- 2026-10-01 ⑭(전체보기 버튼 최종 위치) 확정 = 현재안 그대로
- 2026-09-30 ⑫(전체보기 세부) 확정 = 축소 비율 사용자 조절 + 거르지 않음
- 2026-09-29(10) 모바일 가상 키보드 위험 5가지 정리 + 시안 키보드 토글(v10)
- 2026-09-29(9) 구현 기준 시안 확정("타일링 시안 전체보기" v9) + Grid 탭은 Vuetify VPagination + vault 분할 정리
- 2026-09-29(8) 전체보기(Alt+Tab형) 모드 시안 — 긴 축 한 줄·짧은 축 가운데, 고정 버튼(정보 줄 높이), 스크롤바 auto + 두께 보정, Grid 탭 줄임·끝 페이지 잘림 버그 수정(v1~v9)

### `history/frontend.md` — 프론트엔드 (apps/frontend)
- 2026-10-13(2) 헤더 전체 화면 버튼(모바일 주소창 숨기기)
- 2026-10-13 vue-query queryFn이 ref를 읽던 버그 수정(queryKey 읽기 규칙)
- 2026-10-08 TileFrame 크기 팝업이 타일을 따라오게 수정(B안) — 사용자 화면 확인 완료
- 2026-10-07 앱 이름 "Nexus" + 로고(타일 2×2 + 허브) 헤더·favicon 적용
- 2026-09-29 색 규칙: Vuetify 테마 색 이름 의존 확정(prop → 유틸 클래스 → 테마 변수 → 커스텀 테마 색) + 위반 3곳 정리(커스텀 `terminal` 색 등록)
- 2026-08-07 앱 레이아웃 루트를 `ProvideAppLayout.vue`로 교체(공유 리사이즈 그룹 킥 겸임) + AppHead가 공유 리사이즈 그룹으로 이전
- 2026-06-30 apps/frontend 초기 스캐폴딩 (Vue3+TS+Vuetify, Router/Pinia X, Noto 폰트, .gitignore, 커밋 e511359)
- 2026-06-30 폰트 마무리 정리 (Roboto 제거, Noto 기본폰트 적용 확인)

### `history/realtime.md` — 실시간 push (socket)
- 2026-10-01 Hub 막힘 발견(동기 순차 Publish + write deadline 없음 + 출력 큐 상한 없음) + 사용자 실행 확인 + ping/pong(`KeepAlive`) 해결·확인(`e395433`)
- 2026-07-16 (2) node CRUD 발행처 배선 구현 완료(`AfterCommit` 훅 신설 + create/patch/delete 3핸들러 배선, 이동=2토픽)
- 2026-07-16 (1) node 도메인 Kind 어휘 확정(`NODE:CREATE/UPDATE/DELETE`, `node.change` 단일봉투안 기각) + process 동적구독은 REST로 결론(상세는 process-reconnect.md)
- 2026-06-30 supervisor↔웹 socket 전송 토대 완성 + 3모드(call/emit/on) e2e 검증 (Hub Kind 추가·subscribe.go 인증교정·프론트 hook 재설계)

### `history/widget.md` — 위젯 컴포넌트 (feature/widget)
- 2026-08-07(2) StickyBox 컴포넌트 구현 + 스토어 리팩터(reportSelf 패턴, viewportClient 신규)
- 2026-08-07 Skeleton/SkeletonGroup shimmer 로딩 위젯 신설(소비처 미배선)

### `history/util.md` — 범용 유틸 (common/util)
- 2026-08-07(2) `GroupedSet` 신설 + 공유 리사이즈 관측 그룹(`feature/common`, LifecycleRegistry 두 번째 소비처) + `vue.hook.ts` 리팩터
- 2026-08-07 `EventInterface`+`LifecycleRegistry` 신설, `Memoized` 확장(EventInterface 상속+create/remove 이벤트)

### `history/util-drag.md` — 드래그 인프라 (DraggableSession/useDragGhost/GhostArea)
- 2026-10-12 보류 결정(소비처였던 캔버스·트리 드래그 폐기, 드래그 UX 넣을 때 재검토)
- 2026-08-07(3) 포팅 설계 확정(코드 없음) — 구 test-jig 드래그 코드 문제진단(단일슬롯 콜백/synthetic dispatch/이웃정렬 혼입/ghost 손코딩 중복)+재설계 방향 확정

### `history/test-scripts-1014.md` — 확인 스크립트 명세 2026-10-14~ (2026-10-19 분할)
- 2026-10-19 분할 + Windows worker(윈도-5·윈도-6)·터치 스크롤 스크립트

### `history/test-scripts.md` — 확인 스크립트 재작성 명세 (2026-10-14, test-env에서 분할)
- 2026-10-14 분할 + 파트(확인·실측) 스크립트(`chk/client.py`·`i3.py`·`grace.py`·`order.py`, `pw/htop*.mjs`)

### `history/external-access.md` — 외부 접속(모바일 확인) · HTTPS 프록시 (2026-10-13 신설)
- 2026-10-13(3) 마감·커밋(노드 클립보드와 나눠서)
- 2026-10-13(2) 30초마다 새로고침 = nginx 유휴 끊김 30초 + vite 재연결 새로고침 → HMR keepalive 플러그인
- 2026-10-13 test1~3.falinux.com(8011~8013) 설정 + 소켓 wss·HMR 443·0.0.0.0 바인드, 모바일 헤드리스 확인(커밋 전)

### `history/test-env.md` — 테스트 환경 · 확인용 클라이언트 (2026-10-06 신설)
- 2026-10-15 worker를 꺼도 스크립트가 고아로 남음 — 격리 정리 절차 추가
- 2026-10-14 함정 3줄(socat fork 자식·pkill 자기 셸·`/users/presence` 범위) + 날짜별 확인 스크립트 절을 `REF-test-scripts.md`로 분할
- 2026-10-08(2) 공유-f·터미널 빈틈 확인 스크립트 요령
- 2026-10-08 공유-e 격리 확인 함정 3개(격리 프론트 설정·계정 동기화·헤드리스 계정)
- 2026-10-07 작업공간-e: 격리 프론트 3001(scratchpad vite 설정) + 함정 2개(PATCH `{valid, value}`·`pgrep -f` 자기 셸)
- 2026-10-07(2) 개발 DB 사고(실수로 supervisor 실행 → 00007·00008 적용) 함정 기록 + worker 포함 격리·소켓 클라이언트·`go build -overlay` 요령
- 2026-10-07 격리 실행 방법(임시 DB + 5051 + `TMPDIR`로 dev 판정 우회) 정리
- 2026-10-06 CURRENT의 환경·함정·DB 테스트 데이터 줄을 `REF-test-env.md`로 분리
