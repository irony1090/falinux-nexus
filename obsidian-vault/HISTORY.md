# HISTORY (인덱스)

> 상세 작업 기록은 **주제/모듈별로 `history/` 폴더에 분할**. 이 파일은 어느 파일에 뭐가 있는지 색인만.
> 요약·재사용 지식은 `REF-*.md`, 현재 진행은 `CURRENT.md`.

## 마감 파트 (INDEX — 세션 시작 시 안 읽음)
- 통신 인프라(마감) → `INDEX-infra.md` (transport·transfer·supervisor-web·project)
- process 배선(마감) → `INDEX-process-wiring.md` (process-wiring·trigger·resize·reconnect·subscription)
- node UI 레이아웃 설계(마감) → `INDEX-node-ui-layout.md` (node-ui-layout·projection·impl·render)

## history/ 파일 색인 (진행 중 파트)

### `history/node-label.md` — Node/Label 모듈
- 2026-09-30 (후속) 테스트 스크립트 id 2~4를 폴더 id 1로 이동 + `device_key` 지정(DB 직접 UPDATE)
- 2026-09-30 `GET /workers`(접속 인스턴스, `?nodeId=` 상속 해석) + `execProcess` 대상 장비 검증 — 연동 1차 1번
- 2026-06-29 Node 모듈 구현 (스키마·쿼리·핸들러·PatchNode·internal/patch, build/vet 통과)
- 2026-06-26 Node/Label 모듈 설계 (frontend 카탈로그)

### `history/process-snapshot.md` — 화면복원 스냅샷 (ring buffer)
- 2026-07-16 ring buffer 설계 논의 착수(코드 없음, 순수 설계): supervisor-side 채택 + 스케일 검토 + worker-side 이전 시 필요한 protocol(RingBuffer/offset/MsgSnapshot) + snapshot↔live 이음매 race 발견(Hub 구조상 conn별 차등 라우팅 불가, `bind.CatchUp` 미완성)

### `history/node-ui.md` — node 카탈로그 UI 컨셉
- 2026-08-10(4) 그리드 UI 구현 방식: 직접 제작 확정(조사+설계) — Splitpanes/Golden Layout/Dockview 조사 후 기각, 리사이즈만 DraggableSession 재사용 (2026-08-12 재확인, 결론 불변)
- 2026-08-10(3) `position_x/y` 사용 중단 확정(코드 없음) — 폴더 타일은 ord 기반 list/grid만, DB 컬럼은 유지·값만 미사용
- 2026-08-10 termspace 영감 그리드로 전면 피벗(코드 없음): 깊이2 고정 그리드(열/행 각 최대2, 비대칭 허용) 확정, 옛 캔버스+트리 Phase1-5 컨셉 폐기
- 2026-08-07(2) 구현 Phase 분할(1~5)+라우팅 스킴(`/nodes/:parentId`) 확정(코드 없음, 2026-08-10 폐기됨), Draggable 리팩터 설계로 이어짐
- 2026-08-07 컨셉 설계(코드 없음, 2026-08-10 상당부분 폐기): 배치도(캔버스)=홈+트리=보조내비, PC 오버레이 토글/모바일 세그먼트 토글, 트리 재배치 드래그(device_key 상속변경·좌표NULL 이슈 발견, 이 지식은 계승됨). 모바일 트리 형태만 미정

### `history/node-ui-overview.md` — node 카탈로그 전체보기 · 화면 내비게이션 (node-ui-layout에서 2026-09-29 분할)
- 2026-10-01(2) ⑮(모바일 가상 키보드) 방향 = A+B(타일 UI 유지 + 특수 키 줄) 확정, iOS 실기기 확인은 나중
- 2026-10-01 ⑭(전체보기 버튼 최종 위치) 확정 = 현재안 그대로
- 2026-09-30 ⑫(전체보기 세부) 확정 = 축소 비율 사용자 조절 + 거르지 않음
- 2026-09-29(10) 모바일 가상 키보드 위험 5가지 정리 + 시안 키보드 토글(v10)
- 2026-09-29(9) 구현 기준 시안 확정("타일링 시안 전체보기" v9) + Grid 탭은 Vuetify VPagination + vault 분할 정리
- 2026-09-29(8) 전체보기(Alt+Tab형) 모드 시안 — 긴 축 한 줄·짧은 축 가운데, 고정 버튼(정보 줄 높이), 스크롤바 auto + 두께 보정, Grid 탭 줄임·끝 페이지 잘림 버그 수정(v1~v9)

### `history/process-sync.md` — process 동기화 범위 · 탭 id · 공유 (2026-10-01 신설, node-ui-terminal에서 분리)
- 2026-10-01(4)(5) 1단계 코드·실행 확인 → `history/process-sync-impl.md`
- 2026-10-01(3) 1단계 작업 단위 1-a~1-g 제시(착수 승인 대기), 코드 없음
- 2026-10-01(2) Rebind 뒤 출력 끊김 실행 확인(socat 프록시로 worker 연결만 끊기) + 끊김 때 FAILED 502 발행 발견, 코드 수정 없음
- 2026-10-01 설계 개정 논의: O(계정 동기화)~V(탭 id 전달) 결정 + 진행 순서 1~5 + 1단계(탭 id + 구독 역할 분리) 구조 합의, 코드 없음

### `history/process-input.md` — process 터미널 입력 (2026-10-01 신설)
- 2026-10-02(3) I5(Ctrl+V 붙여넣기) = 붙여넣기 확정·반영 + 입력-c 브라우저 확인 7항목 통과
- 2026-10-02(2) 입력-b(프론트) 코드 작성(`processTerm.store.ts`, I4(Ctrl+C 복사) 포함), type-check 통과·브라우저 확인 전 + I5(Ctrl+V 붙여넣기) 발견
- 2026-10-02 입력-a(서버) 코드 작성(`router/input.go`, `PROCESS:INPUT`) + 서버 단독 확인 4항목 통과
- 2026-10-01 결정 I1~I4 확정(소켓 / 같은 계정 누구나 / PENDING 중 버림 / Ctrl+C = 선택 있으면 복사) + 순서를 ⑪(타일 트리 서버 저장) 앞으로, 코드 없음

### `history/node-ui-save.md` — node 타일 트리 서버 저장 ⑪ 설계·3-c 구조안 (2026-10-01 신설)
- 2026-10-02(6) 3-d 구조안 + 3-d①(끝난 process 조회 = 나누기, 반복 파라미터)·②~④ 확정
- 2026-10-02(5) 3-c 열린 항목(닫기 직전 저장) 확정 = `visibilitychange` hidden 즉시 keepalive 저장
- 2026-10-02(4) 서버 쪽 커밋(`2aa0920`) + 3-c(스토어) 구조안 제시(사용자 확인 전) + REF/history `-impl` 분할
- 2026-10-01 구조안 제시 + 결정 ⑪-1(저장 형태)·⑪-2(충돌 처리)·⑪-3(새로고침 뒤 터미널 화면)·⑪-4(타일 없는 실행 중 process) 확정, 코드 없음

### `history/node-ui-sync.md` — 4 O(계정 동기화) 화면: 타일 트리·터미널 실시간 반영 (2026-10-02 신설)
- 2026-10-02(5) 사용자 PC 확인 + 커밋 `4b6bbcd`, 세션 종료 정리
- 2026-10-02(4) 4-c(터미널 등록 일원화) 작성(보관함·restore 일원화·prune) + 탭 2개 확인 통과
- 2026-10-02(3) 4-b(트리 수신) 작성(push 반영·op 멱등·재연결 재동기화) + 탭 2개 확인 통과
- 2026-10-02(2) 4-a(서버 발행) 작성 + 소켓 확인 7항목 통과
- 2026-10-02 구조안 제시 + 4①~4⑤ 제안대로 확정(작업 단위 4-a~4-d), SNAPSHOT과 따로 진행 합의, 코드 없음

### `history/node-ui-save-impl.md` — node 타일 트리 서버 저장 구현 3-a~3-d (2026-10-02, node-ui-save에서 분할)
- 2026-10-02(5) 3-d(터미널 복원) 작성(lookup API·restore·안전망) + 함정 2개(provide 자신 inject 불가·복원 타일 attach 누락) 수정, 확인 통과, 커밋 `2f0170f`
- 2026-10-02(4) 3-c(스토어) 작성(op 큐·409 재적용·keepalive·exec 빠른 경로) + 헤드리스 8항목 통과(커밋 `2f0170f`)
- 2026-10-02(3) 3-b'(exec 연동) 실패 대처 결정(별도 트랜잭션·kill 보상·3-d 안전망) + 작성 + HTTP 6항목 통과
- 2026-10-02(2) 3-b(API) 작성(`router/tile.go` GET/PUT + 트리 검사) + HTTP 확인·첫 GET 경합 통과
- 2026-10-02 3-a(DB) 작성(`tile_trees` + 쿼리 4개), PG 롤백 실행 확인
### `history/process-sync-impl.md` — process 동기화 1단계 구현 · 실행 확인 (2026-10-01, process-sync에서 분할)
- 2026-10-01(6) sid 충돌 수정(`signIn` nonce) + 커밋 전 사용자 검토로
- 2026-10-01(5) 2단계 실행 확인: 탭 id·계정 구독·크기 우선권·kill·X·W·Y·프론트(헤드리스) 전부 통과 + sid = 쿠키 원본값 충돌 발견(미해결)
- 2026-10-01(4) 1단계 코드 작성(1-a~1-g), build/vet/type-check 통과

### `history/node-ui-terminal.md` — node 타일 터미널 임베드 + 크기 우선권 (2026-10-01 신설)
- 2026-10-01(3) 설계 개정 논의 → `history/process-sync.md`
- 2026-10-01(2) 구현: 크기 우선권 백엔드(`sizeOwner.go`, signOut 확장) + `processTerm.store`·`ProcessTerminal`·`WorkerPickDialog` + 타일 연결 + `ProcessDialog` 삭제 (미커밋, 실행 확인 전)
- 2026-10-01 구조 합의 J~N(xterm 스토어 소유·상태 단일 구독·실행 흐름·크기 우선권 = 실행한 세션·`ProcessDialog` 삭제), 코드 없음

### `history/node-ui-link.md` — node 카탈로그 타일 UI 실제 연동 (더미 → node/process API, 2026-09-30 신설)
- 2026-10-01 스크립트 편집 창 저장 동작 통일 결정(저장 버튼도 저장하고 계속 편집, 코드 미반영)
- 2026-09-30(4) 연동 1차 4(스크립트 편집): `ScriptEditDialog`(브라우저 편집 → `PATCH content`, 새 스크립트 직후 자동 열기), 창 높이는 `VDialog` `height`로
- 2026-09-30(3) 폴더 안(브레드크럼 줄)에서 장비 지정·접속 상태 표시(상속 포함) + 헤드리스 시나리오 확인, 테스트 로그인은 `Login.vue` 기본값 사용 허용
- 2026-09-30(2) 연동 1차 3(node 관리 UI): 생성·이름 변경·장비 지정(직접 입력 + 접속 상태)·삭제(하위를 보던 타일 닫기), `hook/nodeRemove.hook.ts` 신규
- 2026-09-30 연동 1차 2(폴더 타일 목록): 더미 목록·경로·이름을 node API로(`useListChildren`/`useNodePath`/`useGetNode`), 경로 = `getNode` 반복 호출, 실행은 더미 유지

### `history/frontend.md` — 프론트엔드 (apps/frontend)
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

