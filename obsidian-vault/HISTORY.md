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

## history/ 파일 색인 (진행 중 파트)

### `history/node-label.md` — Node/Label 모듈
- 2026-10-13(2) 정리-1(부모 종류 검증) `requireFolder` 구현·격리 확인 → 파트(서버 정리) 마감·커밋
- 2026-10-13 핸들러 책임 재확인 — owner·사이클 검증은 이미 있음, 빈틈 = 부모 종류 검증(다음 파트(서버 정리))
- 2026-10-12(4) 사용자 확인 → 파트(worker 명부(roster)) 마감·커밋
- 2026-10-12(3) 명부-2(조회·삭제 API)~명부-4(확인) 구현, 격리 API 10/10 + 헤드리스 11/11(커밋 전, 사용자 확인 대기)
- 2026-10-12(2) 명부-1(기록) 구현·격리 확인(커밋 전)
- 2026-10-12 roster 계획(RS1(기록 시점)~RS6(위조 검증))·UI 스토리 확정, 다음 = 명부-1
- 2026-10-12 Label 모듈 보류 결정(공유는 5 공유로 해결·모아 보기 = 보드·검색이 먼저)
- 2026-09-30 (후속) 테스트 스크립트 id 2~4를 폴더 id 1로 이동 + `device_key` 지정(DB 직접 UPDATE)
- 2026-09-30 `GET /workers`(접속 인스턴스, `?nodeId=` 상속 해석) + `execProcess` 대상 장비 검증 — 연동 1차 1번
- 2026-06-29 Node 모듈 구현 (스키마·쿼리·핸들러·PatchNode·internal/patch, build/vet 통과)
- 2026-06-26 Node/Label 모듈 설계 (frontend 카탈로그)

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

### `history/node-ui-active.md` — active 타일 표시 · 위치 번호 단축키 (2026-10-12 신설)
- 2026-10-12(6) 사용자 확인 → 파트 마감·커밋
- 2026-10-12(5) 원격 active 배지(터미널 + 폴더) 구조 확정·구현, 격리 헤드리스 11/11(커밋 전)
- 2026-10-12(4) 로컬 active 사용자 확인 완료
- 2026-10-12(3) AT4(단축키 조합) = `Ctrl+Alt+숫자`로 변경, 헤드리스 14/14
- 2026-10-12(2) AT 확정 + 활성-1(상태·누르기)~활성-5(확인) 구현, 헤드리스 13/13(커밋 전)
- 2026-10-12(1) 구조안 제시 AT1(상태 위치)~AT6(이동할 때 같이 할 일) + 활성-1~5 + 다른 계정 active 배지 RA1(맞추는 기준)~RA6(순서), 코드 없음

### `history/frontend.md` — 프론트엔드 (apps/frontend)
- 2026-10-13(2) 헤더 전체 화면 버튼(모바일 주소창 숨기기)
- 2026-10-13 vue-query queryFn이 ref를 읽던 버그 수정(queryKey 읽기 규칙)
- 2026-10-08 TileFrame 크기 팝업이 타일을 따라오게 수정(B안) — 사용자 화면 확인 완료
- 2026-10-07 앱 이름 "Nexus" + 로고(타일 2×2 + 허브) 헤더·favicon 적용
- 2026-09-29 색 규칙: Vuetify 테마 색 이름 의존 확정(prop → 유틸 클래스 → 테마 변수 → 커스텀 테마 색) + 위반 3곳 정리(커스텀 `terminal` 색 등록)
- 2026-08-07 앱 레이아웃 루트를 `ProvideAppLayout.vue`로 교체(공유 리사이즈 그룹 킥 겸임) + AppHead가 공유 리사이즈 그룹으로 이전
- 2026-06-30 apps/frontend 초기 스캐폴딩 (Vue3+TS+Vuetify, Router/Pinia X, Noto 폰트, .gitignore, 커밋 e511359)
- 2026-06-30 폰트 마무리 정리 (Roboto 제거, Noto 기본폰트 적용 확인)

### `history/frontend-auth.md` — 로그인 · 공용 API · 전역 다이얼로그 (2026-10-12 frontend에서 분리)
- 2026-10-13 헤더 로그아웃 버튼 = 햄버거 → `mdi-logout` 아이콘 버튼
- 2026-10-12(3) 사용자 확인 → 파트(user/login 마무리) 마감·커밋
- 2026-10-12(2) 보드에서 로그아웃 시 "보드 삭제" 오인 알림 수정(`boards.loaded` 조건)
- 2026-10-12 파트(user/login 마무리) 로그인-1(상태를 모듈로)~로그인-5(확인) 구현, 헤드리스 12/12
- 2026-06-30 user/login + 공용 API + 전역 다이얼로그 WIP (커밋 3a8e92e 동반)

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

### `history/size-take.md` — 크기 우선권 즉시 가져오기 (2026-10-13 신설)
- 2026-10-14(4) 사용자 확인 → 파트(확인·실측) 마감·커밋
- 2026-10-14(3) 밀림-가(먼저 발행)·밀림-나(즉시 반영) 구현 + 격리 재확인 3/3
- 2026-10-14(2) 실측-3(htop 줄 밀림) 발견: 잃은 탭이 키우는 방향에서 밀린 채 남음(서버 UPDATE 늦음 + 프론트 150ms 묶음), 고칠 방향 밀림-가·밀림-나 제시
- 2026-10-14 사용자 확인 → 파트(크기 가져오기) 마감·커밋 (htop 줄 밀림 관찰은 미해결로 이월)
- 2026-10-13(5) 사용자 환경 헤드리스 확인: 통제 실행 정상(넘김 15초·칩), 사용자 Chrome 열린 상태의 넘김 지연 = 탭 연쇄 끊김 대기 누적으로 추정·종결
- 2026-10-13(4) K5(빼앗긴 탭 안내) = (ㄴ) 확정 → K4·K5 구현, 격리 헤드리스 19/19(커밋 전)
- 2026-10-13(3) K4(소유 계정 우선) 확정 · K5(빼앗긴 탭 안내) 미정(코드 없음)
- 2026-10-13(2) 크기-0(여유 시간)~크기-3(버튼) 구현 + 크기-4(확인) 격리 헤드리스 12/12(커밋 전, 사용자 확인 대기)
- 2026-10-13 구조안(크기-0~4) + K1(다른 계정 소유)·K2(버튼 모양)·K3(여유 시간 = 크기 넘김만 15초) 확정(코드 없음)

### `history/test-scripts.md` — 확인 스크립트 재작성 명세 (2026-10-14, test-env에서 분할)
- 2026-10-14 분할 + 파트(확인·실측) 스크립트(`chk/client.py`·`i3.py`·`grace.py`·`order.py`, `pw/htop*.mjs`)

### `history/external-access.md` — 외부 접속(모바일 확인) · HTTPS 프록시 (2026-10-13 신설)
- 2026-10-13(3) 마감·커밋(노드 클립보드와 나눠서)
- 2026-10-13(2) 30초마다 새로고침 = nginx 유휴 끊김 30초 + vite 재연결 새로고침 → HMR keepalive 플러그인
- 2026-10-13 test1~3.falinux.com(8011~8013) 설정 + 소켓 wss·HMR 443·0.0.0.0 바인드, 모바일 헤드리스 확인(커밋 전)

### `history/node-select.md` — 노드 선택 모드·우클릭 메뉴·드래그 (2026-10-13 node-clipboard에서 분리)
- 2026-10-13(5) 사용자 확인 → 파트(노드 클립보드) 마감·커밋
- 2026-10-13(4) 선택-1(동작 분리)~선택-6(드래그) 구현 + queryFn 버그 수정, 격리 헤드리스 44/44(커밋 전, 사용자 확인 대기)

### `history/node-clipboard.md` — 노드 복사·잘라내기·붙여넣기 UI (2026-10-13 신설)
- 2026-10-13(3) 선택 모드 논의 → 시안 4장 → 승인(SM1~SM6·드래그 확정, 코드 없음)
- 2026-10-13(2) CP5(여러 개 담기)·CP6(표시 위치) 확정 + 클립보드-1~5 구현·격리 26/26(커밋 전)
- 2026-10-13 논의 시작, CP1(범위)~CP4(붙여넣은 뒤) + 부분 실패·계정 전환 확정(코드 없음)

### `history/term-query.md` — 터미널 질의 응답 중복 (2026-10-13 신설)
- 2026-10-13(3) 사용자 확인 → 파트(질의 응답 중복) 마감·커밋
- 2026-10-13(2) 사용자 Chrome에서 탭 2개 실제 확인(질의 3종 응답 각 1개 · vim 이상 입력 없음)
- 2026-10-13 응답-1(구조안)~응답-3(확인) — parser 훅으로 크기 소유 탭만 응답, 격리 헤드리스 12/12 + 대조(수정 뺌) 응답 2 확인

### `history/process-kill.md` — process 종료 정책 (2026-10-12 신설)
- 2026-10-12 결정 필요 항목 정리 → 종료 정책 확정(SIGTERM 먼저 · 끊긴 장비 kill 기록 · 모두 닫으면/두 번째 종료면 SIGKILL)
- 2026-10-12(2) 보기(10) 계정은 "모두 닫음" 판정에서 제외(조작 이상만 셈) + 공유 해제·권한 변경 뒤에도 판정
- 2026-10-12(3) 구조안 승인 → 종료-1(신호 구분)~종료-5(프론트) 구현(빌드·타입 체크 통과)
- 2026-10-12(4) 종료-6(확인) 격리 API 22/22 · 화면 16/16 + 보기 계정 X = 그냥 닫기(커밋 전, 사용자 확인 대기)
- 2026-10-12(5) 사용자 테스트에서 발견: 실행 실패가 PENDING으로 남음 + 재접속 정리가 프론트에 안 감 → `failExec`·`publishLost` 재구독, 격리 6/6
- 2026-10-12(6) 사용자 확인 완료(스크립트 행 강제 종료 · 재시작 뒤 유지) → 파트 마감·커밋

### `history/node-list-sync.md` — 폴더 목록 동기화 (2026-10-12 신설)
- 2026-10-12 문제 제기(공유 폴더 목록이 바로 안 갱신됨) → NB(계정 토픽으로 보내기) 선택 → 구조안 제시
- 2026-10-12(2) 구조안 승인 → 동기화-1(서버 발행 대상)·동기화-2(프론트 수신) 구현(커밋 전)
- 2026-10-12(3) 동기화-3(확인) 격리 API·소켓 26/26 + 화면 11/11 → 사용자 확인 → 파트 마감·커밋

### `history/tile-binding.md` — 바인딩 타일 (2026-10-12 신설)
- 2026-10-12 문제 발견(공유자 실행 터미널이 기본 보드로만) → 바인딩 설계 논의·확정(코드 없음)
- 2026-10-12(2) 구조안 제시(BD1(우선순위)~BD10(기존 타일) + 바인딩-1(타일 메타)~바인딩-6(확인)) → 고아 안전망 그림 추가 후 전부 승인
- 2026-10-12(3) 바인딩-1(타일 메타)~바인딩-5(프론트) 구현 + 바인딩-6(확인) 격리 API 21/21 · 화면 8/8(커밋 전, 사용자 확인 대기)

### `history/tile-binding-rework.md` — 바인딩 타일 규칙 개편 (2026-10-13 tile-binding에서 분리)
- 2026-10-13 사용자 확인 전 규칙 개편 논의(명칭 3개 확정 · R1(이미 그런 것)~R8(구분 대상 불명) 검토 · 고정핀 시나리오 확인 · 목록 행 명령, 코드 없음)

### `history/tile-binding-rework-impl.md` — 바인딩 규칙 개편 구현·확인 (2026-10-13 rework에서 분리)
- 2026-10-13(2) 구조안 개편-1(고정 지도)~개편-9(확인) 승인 → 구현 + 격리 API 34/34 · 화면 23/23 → 사용자 확인 → 파트(바인딩 타일) 마감·커밋

### `history/test-env.md` — 테스트 환경 · 확인용 클라이언트 (2026-10-06 신설)
- 2026-10-14 함정 3줄(socat fork 자식·pkill 자기 셸·`/users/presence` 범위) + 날짜별 확인 스크립트 절을 `REF-test-scripts.md`로 분할
- 2026-10-08(2) 공유-f·터미널 빈틈 확인 스크립트 요령
- 2026-10-08 공유-e 격리 확인 함정 3개(격리 프론트 설정·계정 동기화·헤드리스 계정)
- 2026-10-07 작업공간-e: 격리 프론트 3001(scratchpad vite 설정) + 함정 2개(PATCH `{valid, value}`·`pgrep -f` 자기 셸)
- 2026-10-07(2) 개발 DB 사고(실수로 supervisor 실행 → 00007·00008 적용) 함정 기록 + worker 포함 격리·소켓 클라이언트·`go build -overlay` 요령
- 2026-10-07 격리 실행 방법(임시 DB + 5051 + `TMPDIR`로 dev 판정 우회) 정리
- 2026-10-06 CURRENT의 환경·함정·DB 테스트 데이터 줄을 `REF-test-env.md`로 분리
