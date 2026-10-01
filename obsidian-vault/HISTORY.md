# HISTORY (인덱스)

> 상세 작업 기록은 **주제/모듈별로 `history/` 폴더에 분할**. 이 파일은 어느 파일에 뭐가 있는지 색인만.
> 요약·재사용 지식은 `REF-*.md`, 현재 진행은 `CURRENT.md`.

## history/ 파일 색인

### `history/transport.md` — 통신 인프라
- 2026-06-25 worker 재연결 루프 + backoff + subscribe 리네임(Manager→Hub)
- 2026-06-25 EVENT 평면(단방향 데이터 평면) 재구현 (`-race` 통과)
- 2026-06-17 재사용 구독 매니저 리팩터링 (subscribe.Manager)
- 2026-06-18 요청/응답 상관관계 프레임(call.Correlator) + 등록(REGISTER) 핸드셰이크

### `history/transfer.md` — 파일 전송 모듈
- 2026-06-29 reader 추상화(인터페이스) + 메모리 전송(SendBuffer) — EDIT seed 운반 수단
- 2026-06-24 파일 전송 본체 + abort 구현 (구현 완료, e2e 미검증)
- 2026-06-18 `internal/transfer` 검토 (readFile/saveFile, 미수정)
- 2026-06-19 `internal/transfer` 🔴 누수·레이스 수정 (done 채널 방식)
- 2026-06-19 전송 모듈 착수 준비 (conn 수명상태 / util·manager / supervisorRouter / 전송 프로토콜 설계)

### `history/supervisor-web.md` — supervisor web/HTTP 계층
- 2026-06-26 에러 처리 panic→return 전환 (이후 panic-style 번복)
- 2026-06-26 트랜잭션 미들웨어 + user 가입/로그인 핸들러 (e2e 검증)
- 2026-06-26 supervisor PG 스토어 이식 + DB 배선/마이그레이션 자동적용

### `history/node-label.md` — Node/Label 모듈
- 2026-09-30 (후속) 테스트 스크립트 id 2~4를 폴더 id 1로 이동 + `device_key` 지정(DB 직접 UPDATE)
- 2026-09-30 `GET /workers`(접속 인스턴스, `?nodeId=` 상속 해석) + `execProcess` 대상 장비 검증 — 연동 1차 1번
- 2026-06-29 Node 모듈 구현 (스키마·쿼리·핸들러·PatchNode·internal/patch, build/vet 통과)
- 2026-06-26 Node/Label 모듈 설계 (frontend 카탈로그)

### `history/process-wiring.md` — process 실행 배선 (supervisor 아키텍처 + worker 실행부)
- 2026-07-13 worker 실행부 본체 구현 완료(procs/exec/pump/teardown/input·resize·kill) + Cwd 배선 + e2e 스모크(htop). PROC 토픽 무구독 미해결 발견
- 2026-07-03 folder-open 버그 수정 + worker EDITOR env + PTY 실행부 선행(ExecInteractive env·Pid 접근자)
- 2026-07-01(2) process 도메인 배선 완료 + 경로 조립(sup)/치환(worker) 책임 분리 `{WORKER_BASE}/nodeID/uid`
- 2026-07-01 supervisor ProcessManager·bind·router 골격 + 종료/재접속 모델(status 깔때기·worker끊김→PENDING) 확정

### `history/process-trigger.md` — frontend REST 트리거(exec/kill) + 상태동기화 버그 수정
- 2026-07-22 entry.Record memory 동기화(SetRecord) + kill exit code 유닉스 관례화 + `pty.Interactive.Status()` 에러 계약 버그 수정 — kill 실사용 검증 완료
- 2026-07-21(2) `processDto.go` 신설: `listSubscriptions` 응답 DTO 정정
- 2026-07-21(1) 토픽 접두사 `PROC:`→`PROCESS:` 정정
- 2026-07-16 frontend 트리거(exec/kill) REST 배선 완료(`router.Exec` 시그니처 변경 + `subscribeSid` 자동구독) + 종료 후 Hub 구독 정리(`startRelay`/`cleanupProcessTopic`) — PROC 토픽 무구독 백엔드 해소

### `history/process-resize.md` — process resize(rows/cols) 배선 + `Layout` 에러 계약 정정 + `ProcessDialog`(xterm)
- 2026-07-22(4) resize REST 결과값 기반 DB/memory 동기화 + `MsgProcessUpdate` 발행 배선 완료(build/vet 통과)
- 2026-07-22(3) `POST /processes/resize/:processId` 최초 핸들러(이후 (4)에서 결과값 무시 버그 발견·수정)
- 2026-07-22 프론트 `ProcessDialog.vue`+`processDialog.store.ts` 신설(xterm+FitAddon), `DATA` 이벤트 연동(uid 필터+base64 디코드), uid 전환 시 화면 리셋, `resizeProcess` 클라이언트 추가, `App.vue` 전역 마운트

### `history/process-snapshot.md` — 화면복원 스냅샷 (ring buffer)
- 2026-07-16 ring buffer 설계 논의 착수(코드 없음, 순수 설계): supervisor-side 채택 + 스케일 검토 + worker-side 이전 시 필요한 protocol(RingBuffer/offset/MsgSnapshot) + snapshot↔live 이음매 race 발견(Hub 구조상 conn별 차등 라우팅 불가, `bind.CatchUp` 미완성)

### `history/process-reconnect.md` — 종료/재접속 모델 (worker 끊김→PENDING→재바인딩)
- 2026-07-22 PENDING 오삭제 버그(정상 실행 vs 끊김 합성 혼동) 수정
- 2026-07-14 (2) worker 끊김→PENDING→재접속 재바인딩 구현 완료 + e2e 검증(applyStatus 가드 완화, worker `WorkerState` 신설 포함)
- 2026-07-14 (1) 위 설계 확정(코드 변경 없음, 순수 설계)

### `history/process-subscription.md` — 세션→uid 원장 + REST 구독/해지 배선
- 2026-07-16 REST 구독/해지(GET/POST/DELETE /processes/*) + `browsers`(conn→sid) registry 배선 완료 — 세션→uid 원장 마지막 조각
- 2026-07-14 (3) 세션→uid 원장 설계 착수: sid 추출 배선(코드 완료, NameFunc에 req 추가) + `process_subscribers` 테이블 설계(미구현)

### `history/node-ui.md` — node 카탈로그 UI 컨셉
- 2026-08-10(4) 그리드 UI 구현 방식: 직접 제작 확정(조사+설계) — Splitpanes/Golden Layout/Dockview 조사 후 기각, 리사이즈만 DraggableSession 재사용 (2026-08-12 재확인, 결론 불변)
- 2026-08-10(3) `position_x/y` 사용 중단 확정(코드 없음) — 폴더 타일은 ord 기반 list/grid만, DB 컬럼은 유지·값만 미사용
- 2026-08-10 termspace 영감 그리드로 전면 피벗(코드 없음): 깊이2 고정 그리드(열/행 각 최대2, 비대칭 허용) 확정, 옛 캔버스+트리 Phase1-5 컨셉 폐기
- 2026-08-07(2) 구현 Phase 분할(1~5)+라우팅 스킴(`/nodes/:parentId`) 확정(코드 없음, 2026-08-10 폐기됨), Draggable 리팩터 설계로 이어짐
- 2026-08-07 컨셉 설계(코드 없음, 2026-08-10 상당부분 폐기): 배치도(캔버스)=홈+트리=보조내비, PC 오버레이 토글/모바일 세그먼트 토글, 트리 재배치 드래그(device_key 상속변경·좌표NULL 이슈 발견, 이 지식은 계승됨). 모바일 트리 형태만 미정

### `history/node-ui-layout.md` — node 카탈로그 레이아웃 (저장 모델 · 타일 순서 · Grid 배치 · 오버플로, node-ui에서 2026-08-12 분할)
- 2026-09-30 ⑧(같은 Grid 앞쪽 빈칸 채우기) 확정 = 커서 방식(앞으로만 채움), `A | C` / `B`는 허용 + `pack` 코드 반영(시나리오·vue-tsc 통과)
- 2026-09-29(18) 같은 Grid 안 순서 역전 발견(세로 1.0 변경 시 뒤 타일이 앞쪽 빈칸 채움) — 수정안(커서)만 제시, 결정 대기
- 2026-09-29(7) 빈 오른쪽 칼럼 줄이기 = 모든 Grid 적용으로 확정(비교 시안 2개 제작, 세로 빈칸은 유지)
- 2026-09-29(6) 현재 규칙 승인(⑨ 현행 유지) + 링크를 user vault에 기록, 다음은 코드 구현
- 2026-09-29(4) 시안을 두 기기(PC·폰) 동시 표시로 재작성, 교차 영향 확인(폰에서 연 1.0×0.5 타일이 PC에선 가로바가 되어 뒤를 밈)
- 2026-09-29(3) 타일별 크기 조절(가로·세로 0.5/1.0, 새 타일 = 화면 최소 단위) + 2×2칸 채우기 배치(안 들어가면 다음 Grid, 순서 유지), Band 방향은 배치 결과로 결정(코드 없음, 시안만)
- 2026-09-29(2) 타일 순서 규칙(폴더 → 실행한 프로세스 → 연 하위 폴더) + 전체 재배치(reflow) + axis 기기별 파생 확정, 저장 대상이 여는 관계 트리로 바뀜(08-19 axis 고정 저장 폐기), 투영을 별도 파일로 분할(코드 없음, 시안만)
- 2026-08-19(3) 오버플로 Grid 페이지 크기/페이징 아이디어 제안(확정 아님, 스냅은 09-29 칼럼 단위로 확정)
- 2026-08-19(2) 구조 확정 시점 & 오버플로 처리 확정(axis 고정 부분은 09-29 폐기)
- 2026-08-19(1) Phase1 구현 착수(TileLayout.vue)
- 2026-08-12 레이아웃 데이터 모델 axis 일반화(코드 없음) — 루트 칼럼우선 고정으로는 마스터-스택형(가로바+아래2분할) 표현 불가 발견 → `Grid.axis:'row'|'column'` 도입 + `Column`→`Band` 개명, 화면 크기별 투영도 primary/secondary 구조축 기준으로 재정리

### `history/node-ui-overview.md` — node 카탈로그 전체보기 · 화면 내비게이션 (node-ui-layout에서 2026-09-29 분할)
- 2026-09-30 ⑫(전체보기 세부) 확정 = 축소 비율 사용자 조절 + 거르지 않음
- 2026-09-29(10) 모바일 가상 키보드 위험 5가지 정리 + 시안 키보드 토글(v10)
- 2026-09-29(9) 구현 기준 시안 확정("타일링 시안 전체보기" v9) + Grid 탭은 Vuetify VPagination + vault 분할 정리
- 2026-09-29(8) 전체보기(Alt+Tab형) 모드 시안 — 긴 축 한 줄·짧은 축 가운데, 고정 버튼(정보 줄 높이), 스크롤바 auto + 두께 보정, Grid 탭 줄임·끝 페이지 잘림 버그 수정(v1~v9)

### `history/node-ui-impl.md` — node 카탈로그 타일 UI 코드 구현 (2026-09-29 신설)
- 2026-09-29(11) 구현 착수: 층 구조 ①~⑤ 합의(역할 분담 = Claude 코드·구조 합의) + ① ② 순수 유틸 작성(`tileOrder`/`tileSplit`, tsc·시나리오 테스트 통과) + layout/projection에서 코드 관련 절 분할
- 2026-09-29(12) ③ 트리 상태 스토어(`tileTree.store.ts`) 작성 + 닫기 규칙(⑩) 확정 + 실행 중 터미널 닫기 비활성(A)
- 2026-09-29(13) 넘겨받은 프로세스(`adopted`) 도입 — 폴더 닫기 후 부모가 직접 실행한 프로세스 뒤에 오도록 순서 규칙 확장
- 2026-09-29(14) ④ Grid 파생 + ⑤ 렌더 최소판(더미) 작성 — 브라우저에서 PC/폰 폭 reflow·실행·kill/닫기 동작 확인
- 2026-09-29(15) 칼럼 단위 스냅 추가(`TileLayout` overlay 슬롯 앵커) — Grid 경계 걸침 스냅 확인
- 2026-09-29(16) 사용자 직접 확인·승인("매우 맘에 들어") + 남은 작업·미확인 항목 정리

### `history/node-ui-render.md` — node 카탈로그 타일 UI 렌더 · 헤더 내비 (node-ui-impl에서 2026-09-29 분할)
- 2026-09-30(3) 전체보기 축소 비율 localStorage 저장(`userPrefs` 유틸 신규) + 로그인 풀리면(`auth` null) 삭제
- 2026-09-30(2) 전체보기 모드 + 고정 버튼 작성(`OverviewButton` 신규, 재마운트 금지 구조, 축소 비율 사용자 조절) — 헤드리스 PC·폰 확인
- 2026-09-30 타일 위치 따라가기 `reveal` — 크기 변경·새 타일로 화면 밖 Grid에 가면 그 Grid로 스크롤 (PC·폰 헤드리스 확인)
- 2026-09-29(17) 헤더 내비 작성 — Teleport로 AppHead에 VPagination Grid 탭 + 칼럼 이동 화살표 + `칼럼 a-b/N`, 현재 Grid 판정 버그(gap 섞임) 수정

### `history/node-ui-link.md` — node 카탈로그 타일 UI 실제 연동 (더미 → node/process API, 2026-09-30 신설)
- 2026-09-30(4) 연동 1차 4(스크립트 편집): `ScriptEditDialog`(브라우저 편집 → `PATCH content`, 새 스크립트 직후 자동 열기), 창 높이는 `VDialog` `height`로
- 2026-09-30(3) 폴더 안(브레드크럼 줄)에서 장비 지정·접속 상태 표시(상속 포함) + 헤드리스 시나리오 확인, 테스트 로그인은 `Login.vue` 기본값 사용 허용
- 2026-09-30(2) 연동 1차 3(node 관리 UI): 생성·이름 변경·장비 지정(직접 입력 + 접속 상태)·삭제(하위를 보던 타일 닫기), `hook/nodeRemove.hook.ts` 신규
- 2026-09-30 연동 1차 2(폴더 타일 목록): 더미 목록·경로·이름을 node API로(`useListChildren`/`useNodePath`/`useGetNode`), 경로 = `getNode` 반복 호출, 실행은 더미 유지

### `history/node-ui-projection.md` — node 카탈로그 화면 크기별 표시 (node-ui-layout에서 2026-09-29 분할)
- 2026-09-29(5) 페이지네이션 폐기(사용자 "원한 게 아님") → 좁은 축은 저장된 0.5를 1.0으로 보고 다시 채움, 채우기 방향 모든 기기 칼럼 우선으로 통일(코드 없음, 시안만)
- 2026-09-29(1) 타일링 시안(아티팩트) 제작 + 짧은축 규칙 폐기 → 축별 기준치 페이지네이션(W 728 / H 560, 상한 4장) + 판정 유틸 분리 결정 + 오버플로 Grid 칼럼 단위 스냅 확정(코드 없음)
- 2026-08-10(2) 화면 크기별 투영: 짧은축 페이지네이션 확정(코드 없음, 09-29 폐기) — 서버는 풀사이즈 그리드 하나만 저장, 뷰포트 종횡비로 짧은축 판정→페이지 분리(최대2장)

### `history/frontend.md` — 프론트엔드 (apps/frontend)
- 2026-09-29 색 규칙: Vuetify 테마 색 이름 의존 확정(prop → 유틸 클래스 → 테마 변수 → 커스텀 테마 색) + 위반 3곳 정리(커스텀 `terminal` 색 등록)
- 2026-08-07 앱 레이아웃 루트를 `ProvideAppLayout.vue`로 교체(공유 리사이즈 그룹 킥 겸임) + AppHead가 공유 리사이즈 그룹으로 이전
- 2026-06-30 apps/frontend 초기 스캐폴딩 (Vue3+TS+Vuetify, Router/Pinia X, Noto 폰트, .gitignore, 커밋 e511359)
- 2026-06-30 폰트 마무리 정리 (Roboto 제거, Noto 기본폰트 적용 확인)
- 2026-06-30 user/login + 공용 API + 전역 다이얼로그 WIP (커밋 3a8e92e 동반)

### `history/realtime.md` — 실시간 push (socket)
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

### `history/project.md` — 프로젝트 셋업 / worker 기반
- 2026-06-17 Nexus 프로젝트 분리 & vault 신규 구성
- 2026-06-17 apps/core 스캐폴딩 & 레포 초기 커밋
- 2026-06-18 worker 리팩토링(router 패턴) + SQLite 정체성 영속
