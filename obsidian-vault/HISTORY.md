# HISTORY (인덱스)

> 상세 작업 기록은 **주제/모듈별로 `history/` 폴더에 분할**. 이 파일은 어느 파일에 뭐가 있는지 색인만.
> 요약·재사용 지식은 `REF-*.md`, 현재 진행은 `CURRENT.md`.

## 마감 파트 (INDEX — 세션 시작 시 안 읽음)
- 통신 인프라(마감) → `INDEX-infra.md` (transport·transfer·supervisor-web·project)
- process 배선(마감) → `INDEX-process-wiring.md` (process-wiring·trigger·resize·reconnect·subscription)
- node UI 레이아웃 설계(마감) → `INDEX-node-ui-layout.md` (node-ui-layout·projection·impl·render)
- node UI 연동(마감) → `INDEX-node-ui-link.md` (node-ui-link·terminal·save·save-impl·sync)
- process 터미널 기능(마감) → `INDEX-process-terminal.md` (process-input·snapshot·snapshot-impl·exec-edit)

## history/ 파일 색인 (진행 중 파트)

### `history/node-label.md` — Node/Label 모듈
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

### `history/process-sync.md` — process 동기화 범위 · 탭 id · 공유 (2026-10-01 신설, node-ui-terminal에서 분리)
- 2026-10-06 presence(접속 상태) 조회 시점 = 5 공유(P·R·T) 구조 합의 때 함께로 합의, 코드 없음
- 2026-10-01(4)(5) 1단계 코드·실행 확인 → `history/process-sync-impl.md`
- 2026-10-01(3) 1단계 작업 단위 1-a~1-g 제시(착수 승인 대기), 코드 없음
- 2026-10-01(2) Rebind 뒤 출력 끊김 실행 확인(socat 프록시로 worker 연결만 끊기) + 끊김 때 FAILED 502 발행 발견, 코드 수정 없음
- 2026-10-01 설계 개정 논의: O(계정 동기화)~V(탭 id 전달) 결정 + 진행 순서 1~5 + 1단계(탭 id + 구독 역할 분리) 구조 합의, 코드 없음

### `history/process-share.md` — process 공유 (5 공유(P·R·T) + presence, 2026-10-06 신설)
- 2026-10-06(2) SH2(받는 쪽 화면 위치) 확정(받는 쪽 `/` 루트) + 수락 없음 + 공유 사라짐 처리 + 받는 쪽 구분 표시·소유자 쪽 공유 목록 시안(제안), 코드 없음
- 2026-10-06 구조안 제시 + 확정 SH1(공유 단위)·SH3(받은 계정 권한)·SH8(받은 쪽이 만든 노드의 소유)·SH9(받은 쪽이 실행한 process의 소유)·SH10(노드 복사)+클립보드 표시, 코드 없음

### `history/workspace.md` — 작업 공간 여러 개 · 터미널 소유 표시 (2026-10-06 신설, 공유 논의에서 파생)
- 2026-10-06 작업 공간 규칙(`/`에서만 생성·`/w/{hash}`·폴더 단위 이동) + W1(분리 범위)=타일 트리 하위·W2(삭제 처리)=`/`로 되돌림·W3(이동 범위)=제한 없음 + 터미널 머리 줄 `[실행한 타일 번호] 경로` 확정, 코드 없음

### `history/workspace-impl.md` — 작업 공간 여러 개: 구현 구조·결정 (2026-10-07 신설)
- 2026-10-07(2) WS1~WS7 확정(코드 착수는 사용자 지시 뒤) + WS5 의미 보충
- 2026-10-07 구현 구조안 제시(WS1(트리 주소)~WS7(보는 작업 공간 전달)) + 받는 쪽 고아 안전망 문제 발견, 코드 없음


### `history/process-share-realtime.md` — 공유-d(실시간) (2026-10-07 신설)
- 2026-10-08(21) 실시간-7(확인: 조합 시나리오 17·회귀 52·화면 보강) 통과 — 공유-d(실시간) 마감, 코드 변경 없음
- 2026-10-07(9) 구조안 제시 SD1(출력 구독 범위)~SD9(경로 줄 멈춤) + 실시간-1~7, 코드 없음
- 2026-10-07(9) SD3 논의: 닫기 → 최소화 개정, 루트 탭 전환 + 스크립트 행 "실행 중 N", 목록 kill·모든 kill 확인창·종료하고 닫기, 머리 줄 A안 `RUN▾ [최소화|크기|×]` 확정 + 소유자 kill 조건, 시안 v1~v6 + 작업 순서 확정(실시간-0(터미널 UI) 맨 앞, 실시간-6 → 6a·6b 분할)

### `history/process-share-reconcile.md` — 공유-d 실시간-1~3(재계산·연결) (2026-10-07 신설)
- 2026-10-08(18) 실시간-5(노드 응답: 공유 루트 목록·`parentVisible`) 코드 + 격리 실행 12/12(커밋 전)
- 2026-10-08(17) 실시간-4(exec 타일) 코드 + 격리 실행 14/14(커밋 전)
- 2026-10-08(16) 실시간-3(노드 이동·삭제 연결) 코드 + 격리 실행 15/15(커밋 전)
- 2026-10-07(15) SH1(공유 단위) 개정 = 폴더만 공유(스크립트 PUT 400) + 재확인
- 2026-10-07(14) 실시간-2 코드(공유 PUT·DELETE 연결·받는 쪽 타일) + 격리 실행 35/35(커밋 전)
- 2026-10-07(13) 실시간-1 코드(`reconcileAccess`·쿼리 3개·`SHARE:CHANGE`) + 임시 DB 쿼리 확인(커밋 전)
- 2026-10-07(12) 실시간-1(재계산 함수) 구조안 + 정할 것 4개 제시, 코드 없음

### `history/process-share-terminal-place.md` — 터미널 타일 위치 · 폴더 타일 uid (2026-10-08, process-share-realtime에서 분할)
- 2026-10-12(29) 브랜치 커밋 + 분기 지점 태그 `checkpoint/before-folder-tile-uid` + `master` `--no-ff` 병합
- 2026-10-12(28) `openFolder` 제거 + 제거 전 23/23·후 21/21 격리 확인
- 2026-10-12(27) 사용자 직접 확인 통과 + 폴더 process 정리(`openFolder`는 안 쓰는 옛 경로), 코드 없음
- 2026-10-08(26) FU2 개정: 공유한 타일과 같은 uid + uid 우선 찾기(격리 11 통과, 커밋 전)
- 2026-10-08(25) 다른 보드로 가는 터미널 수정(placeTerminal 공용) + 폴더 타일 uid(FU1~FU4) 코드·격리 확인(브랜치 `feat/folder-tile-uid`)
- 2026-10-08(24) 다른 보드로 가는 터미널 조사(빈틈 3곳) + 폴더 타일 uid 지시(FU1~FU4 대기) + 체크포인트 커밋·분기, 코드 없음
- 2026-10-08(23) 공유 전에 실행한 터미널을 받는 쪽이 못 여는 빈틈 수정(공유 시 터미널 타일 + "내 보드에 열기")

### `history/process-share-realtime-ui.md` — 공유-d 실시간-0(터미널 UI) (2026-10-07 신설, process-share-realtime에서 분할)
- 2026-10-08(22) 루트 타일이 공유 폴더 안에서 막힘 수정(배너에 "기본으로 돌아가기") — 사용자 확인 통과
- 2026-10-08(20) 실시간-6b(공유 반영: 볼 수 없는 타일 배너·경로 멈춤·`SHARE:CHANGE` 재조회) 코드 + 헤드리스 확인(커밋 전)
- 2026-10-08(19) 실시간-6a(스크립트 행 "실행 중 N") 코드 + REST 11/11 + 헤드리스 화면 확인(커밋 전)
- 2026-10-07(11) 실시간-0 코드 + 격리 실행 헤드리스 24/24 통과(커밋 전)
- 2026-10-07(10) SD1~SD9 전부 확정 + 실시간-0 구현 계획·확인할 것 4개 제시, 코드 없음

### `history/process-share-display.md` — 공유-e(프론트 표시) (2026-10-08 신설, e-3·f는 분할)
- 2026-10-08(5) 공유-e-2(공유 걸기 창) 코드·격리 확인 11개 통과(커밋 전)
- 2026-10-08(4) 공유-e-2(공유 걸기 창) 구조안·시안 확정, 코드 없음
- 2026-10-08(3) 공유-e-1 격리 실행 확인 통과(+단계 변경 갱신 버그 수정)
- 2026-10-08(2) 공유-e-1(받는 쪽 표시) 코드·빌드 통과(커밋 전)
- 2026-10-08 공유-e-0(제안 항목 확정): ①~④·⑥ 채택·B 제외, 숨기기·O4·O5 분리, 코드 없음

### `history/process-share-owner.md` — 공유-e-3 소유자 쪽 목록(O1~O3·O5) (2026-10-08, display에서 분할)
- 2026-10-08(9) 공유-e-3c(내가 공유한 것) 코드·빌드 통과 → 사용자 확인 통과(커밋 전)
- 2026-10-08(8) 공유-e-3b(행 공유 개수) 코드·빌드 통과 → 3c와 함께 사용자 확인 통과(커밋 전)
- 2026-10-08(7) 공유-e-3a(공유 목록 창) 코드·격리 확인 11개 통과(커밋 전)
- 2026-10-08(6) 공유-e-3 구조안·시안 확정 + 공유 표시 색 청록(`share`) 적용

### `history/process-share-presence.md` — 공유-f 접속 상태(presence) (2026-10-08, display에서 분할)
- 2026-10-08(11) 공유-f-5(확인) 격리 API 25개 + 헤드리스 화면 통과
- 2026-10-08(10) 공유-f(presence) f-1~f-4 코드·빌드 통과, 확인 전(커밋 전)

### `history/workspace-front.md` — 보드(작업 공간 여러 개): 프론트 코드 작성 (2026-10-07 신설)
- 2026-10-07(8) 작업공간-e(확인): 헤드리스 34개 + 공유-b·c 회귀 59개 통과 + 발견(받는 쪽 실행 터미널이 소유자 기본 보드에 고아로 붙음, 공유-d 때 정함), 코드 수정 없음
- 2026-10-07(6) 화면 이름 "보드" 확정 + 작업공간-d(프론트) 코드·type-check 통과(커밋 전, 브라우저 확인은 작업공간-e)

### `history/workspace-server.md` — 작업 공간 여러 개: 서버 코드 작성 (2026-10-07, workspace-impl에서 분할)
- 2026-10-07(5) 작업공간-c(exec·안전망·탭) 코드(exec `workspace`·서버 고아 안전망·`TAB:VIEW`·`resyncSizeOwner`)·worker 포함 격리 실행 20개 통과(커밋 전)
- 2026-10-07(4) 작업공간-b(API) 코드(`/tiles/:hash`·WS5(터미널 타일 추가 주체) 검사·`workspace.go`)·격리 실행 41개 통과(커밋 전) + 과도기 위험(`adoptOrphans` PUT 400)
- 2026-10-07(3) 작업공간-a(DB) 코드(`00008_workspaces.sql`·쿼리)·확인 20개 + down/up 통과(커밋 전) + 실수로 supervisor 실행 → 개발 DB에 00007·00008 적용
### `history/process-share-node.md` — 공유: 권한 단계 · 노드 복사 · 클립보드 (2026-10-06, process-share REF 분할로 신설)
- 2026-10-06 신설(결정 이력 본문은 `history/process-share.md` 2026-10-06)

### `history/process-share-db.md` — 공유: 공유-a(DB) (2026-10-06 신설)
- 2026-10-07 공유-a 검토(커밋 막을 문제 없음) + `ListChildrenVisible` 소유자 40 고정 보완
- 2026-10-06(2) SA1(단계 저장 형식)~SA4(기존 행 채우기) 확정 + 공유-a 코드 작성, build/vet + 임시 DB 확인 통과(커밋 전)
- 2026-10-06 구조안 제시(`00007_shares.sql`·판정 함수·판정 쿼리, SA1~SA4) + sqlc 함수 반환형 인식 확인, 코드 없음

### `history/process-share-authz.md` — 공유: 공유-b(인가 교체) (2026-10-06 신설)
- 2026-10-07 인가-1(판정 헬퍼)~인가-8(확인) 코드 작성 + 격리 실행 53개 항목 통과(커밋 전)
- 2026-10-06(2) SB1(권한 부족 응답)~SB6(이동 사이클 검사) 확정, 코드 없음
- 2026-10-06 구조안 제시(엔드포인트별 요구 단계, SB1~SB6) + `NODE:0` 누설 발견, 코드 없음

### `history/process-share-api.md` — 공유: 공유-c(공유 API) (2026-10-07 신설)
- 2026-10-07(2) SC1~SC7 확정 + 공유API-1(판정 함수 정리)~공유API-7(확인) 코드 작성, 42개 + 회귀 50개 통과(커밋 전)
- 2026-10-07 구조안 제시(API 8개, SC1(출처 계산 위치)~SC7(복사 제외 개수)), 코드 없음

### `history/process-sync-impl.md` — process 동기화 1단계 구현 · 실행 확인 (2026-10-01, process-sync에서 분할)
- 2026-10-01(6) sid 충돌 수정(`signIn` nonce) + 커밋 전 사용자 검토로
- 2026-10-01(5) 2단계 실행 확인: 탭 id·계정 구독·크기 우선권·kill·X·W·Y·프론트(헤드리스) 전부 통과 + sid = 쿠키 원본값 충돌 발견(미해결)
- 2026-10-01(4) 1단계 코드 작성(1-a~1-g), build/vet/type-check 통과

### `history/frontend.md` — 프론트엔드 (apps/frontend)
- 2026-10-08 TileFrame 크기 팝업이 타일을 따라오게 수정(B안) — 사용자 화면 확인 완료
- 2026-10-07 앱 이름 "Nexus" + 로고(타일 2×2 + 허브) 헤더·favicon 적용
- 2026-09-29 색 규칙: Vuetify 테마 색 이름 의존 확정(prop → 유틸 클래스 → 테마 변수 → 커스텀 테마 색) + 위반 3곳 정리(커스텀 `terminal` 색 등록)
- 2026-08-07 앱 레이아웃 루트를 `ProvideAppLayout.vue`로 교체(공유 리사이즈 그룹 킥 겸임) + AppHead가 공유 리사이즈 그룹으로 이전
- 2026-06-30 apps/frontend 초기 스캐폴딩 (Vue3+TS+Vuetify, Router/Pinia X, Noto 폰트, .gitignore, 커밋 e511359)
- 2026-06-30 폰트 마무리 정리 (Roboto 제거, Noto 기본폰트 적용 확인)
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
- 2026-08-07(3) 포팅 설계 확정(코드 없음) — 구 test-jig 드래그 코드 문제진단(단일슬롯 콜백/synthetic dispatch/이웃정렬 혼입/ghost 손코딩 중복)+재설계 방향 확정

### `history/test-env.md` — 테스트 환경 · 확인용 클라이언트 (2026-10-06 신설)
- 2026-10-08(2) 공유-f·터미널 빈틈 확인 스크립트 요령
- 2026-10-08 공유-e 격리 확인 함정 3개(격리 프론트 설정·계정 동기화·헤드리스 계정)
- 2026-10-07 작업공간-e: 격리 프론트 3001(scratchpad vite 설정) + 함정 2개(PATCH `{valid, value}`·`pgrep -f` 자기 셸)
- 2026-10-07(2) 개발 DB 사고(실수로 supervisor 실행 → 00007·00008 적용) 함정 기록 + worker 포함 격리·소켓 클라이언트·`go build -overlay` 요령
- 2026-10-07 격리 실행 방법(임시 DB + 5051 + `TMPDIR`로 dev 판정 우회) 정리
- 2026-10-06 CURRENT의 환경·함정·DB 테스트 데이터 줄을 `REF-test-env.md`로 분리
