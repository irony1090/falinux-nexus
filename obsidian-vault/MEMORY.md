# MEMORY

> 매 세션 시작 시 항상 로드. **불변 결정 + 항상 필요한 것만** 간결하게.
> 상세 설계·재사용 지식은 `REF-*.md`, 작업 이력은 `history/*.md`, 현재 진행은 `CURRENT.md`.

## 경로 참조 (중요)
- **현재 프로젝트 (Nexus)**: `/home/jh-bae/irony/nexus`
  - Go 코어: `/home/jh-bae/irony/nexus/apps/core`
- **이전 프로젝트 (PortBridge)**: `/home/jh-bae/irony/test-jig`
  - Go: `.../apps/test-jig-api` / 프론트: `.../apps/test-jig-web` / vault: `.../obsidian-vault`

> 사용자가 **"test-jig" / "예전·구·이전 프로젝트" / "PortBridge"** 로 칭하면 `/home/jh-bae/irony/test-jig`. 해당 코드 참조할 것.

## 프로젝트 개요
- **Nexus** = 2-tier 에이전트 관리 플랫폼
  - **Worker Agent**: 자기 process 관리(실행/모니터링/종료)
  - **Supervisor Agent**: 다수 worker 관리(연결/상태/명령 분배)
- 출발점: PortBridge의 agent↔server 통신 아키텍처 계승. 방향이 "테스트 장비 데스크탑 UI"→"에이전트 계층 관리"로 전환되어 신규 분리. (PortBridge 전용 UI·아이콘·그룹 로직 폐기)

## 확정 결정 (불변 — 장기 유지)
- **계층: 2단계 고정** (supervisor → worker, 중간 노드 없음. 식별자/라우팅/구독 1홉 평면)
- **worker 라우팅/신원 = 메모리 레지스트리**(`InstanceKey`=`메인키#서브키`) — 살아있는 연결로만 라우팅, **DB가 라우팅 authority 아님**. ※구 `agents` 테이블 제거(2026-06-26)는 "영속 금지 원칙"이 아니라 **이름·역할 불명** 때문이었음 — 영속 자체는 가능(2026-06-29 정정)
  - **관측용 roster는 DB 허용**: `worker_instances(main_key, sub_key, last_seen)` = UI에 main_key별 인스턴스 활성/비활성 표시용(+`first_seen`). online은 저장 말고 레지스트리로 **파생**(roster ≠ authority). 구현 마감 2026-10-12 (→ `REF-node-label.md` "roster 계획")
  - node의 device 결속 = FK 아닌 `device_key TEXT`(**= main_key**, 폴더는 장비 클래스에 묶임 / subkey는 실행 시 런타임 선택) (→ `REF-node-label.md`)
- **agent 식별자: 사전 지정 메인키 + supervisor가 접속 시 부여하는 서브키.** 저장/조회는 `메인키#서브키`(`InstanceKey()`). MAC 폐기
- **process 영속성**: worker 휘발(메모리) / supervisor 영속(PG, 재시작 복구)
- **재연결 reconciliation (2026-07-01 개정 — 구 `Done(502)` 비관 폐기)**: worker 끊김→관련 process **PENDING**(죽이지 않음) / 재접속→worker가 자기 live 상태 보고로 재동기화(낙관적). **모든 상태전이=process 라우터 `status` 단일 깔때기**(worker 보고 or 끊김 시 supervisor 합성). **frontend 끊김≠종료**(명시적 kill 전까진 실행 유지, 브라우저 종료 무시), **같은 세션 재접속=보던 화면 복원**(ring+세션원장) (→ "종료/재접속 모델")
- **에러 처리 = panic-style**(2026-06-26 재확정): 핸들러 `panic(web.Err(...))` → `PanicMiddleware` 렌더 (→)
- **process 동기화 범위 = 계정**(2026-10-01, 1단계 코드·실행 확인·커밋 완료): 같은 계정이면 탭·브라우저·기기 무관 동기화. 인가 = 계정(DB) / 출력 라우팅 = 탭 소켓(메모리) / PTY 크기 우선권 = 서버 발급 탭 id(메모리) → `process_subscribers` 제거(`00005`). 공유는 계정 단위·소유자만 신청 (→ `REF-process-sync.md`, 구현·확인 → `REF-process-sync-impl.md`)
- **공유 판정 = DB 함수 한 곳**(2026-10-06, 공유-a): 가까운 공유 단계 계산은 `node_permission`/`process_permission`(`00007_shares.sql`)에만 두고 쿼리·Go는 호출만. 단계 = 0 차단·10 보기·20 조작·30 편집·40 소유자(함수 반환만), 비교는 `>=`. 볼 수 없음 = 404 / 단계 미달 = 403. **공유 단위 = 폴더 + process 단독**(스크립트 노드 공유 금지, 2026-10-07 SH1 개정) (→ `REF-process-share-db.md`, `REF-process-share-authz.md`, `REF-process-share.md`)
- **실시간 push = 인가/라우팅 분리**(2026-06-30): 인가=DB(구독 시점 1회) / 라우팅=메모리(`subscribe.Hub` topic Publish, DB 무접촉). 수신자=presence(구독)≠소유자 → 공유·다중 유저 동일 화면 흡수. 토픽 단위: process 출력 = `PROCESS:<uid>` / **node 변경 = 볼 수 있는 접속 계정의 `TILES:<userId>`, 발행 시점 DB 판정**(2026-10-12 개정 — 옛 "펼친 폴더 `NODE:<parentId>`"는 구독 코드가 없어 폐기, → `REF-node-list-sync.md`). "DB≠라우팅 authority"를 browser 평면으로 확장. 비대칭: 브라우저 call→서버 Handle / 서버 Emit→브라우저 on (→ `REF-realtime.md`)

- **모바일 키보드 = `resizes-content` + 열린 동안 배치 영역은 키보드 전 높이(px) 고정**(2026-10-14): 타일 재배치·PTY 크기 변경 금지. `dvh`로 배치하는 새 영역도 지킬 것 (→ `REF-mobile-viewport.md` 4)

## 기술 스택 (확정)
- **레포**: `nexus` / 모노레포 `apps/core`(Go) + `apps/frontend`(프론트, 스캐폴딩 완료 2026-06-30)
- **프론트**: Vue3 + TS + Vuetify4 (scratch preset, Router=standard / **Pinia 미사용**, CSS framework 없음). 상태공유=provide/inject(+컴포넌트 밖 필요 시 reactive 모듈 패턴). 기본폰트=Noto Sans Korean. **색 = Vuetify 테마 색 이름 의존, hex 금지**(2026-09-29), 공유 표시 = 테마 `share`(청록) (→ `REF-frontend.md`)
- **Go**: 단일 모듈 `nexus`, 멀티 cmd(`cmd/supervisor`, `cmd/worker`), Go 1.26
- **DB**: worker → SQLite(`modernc.org/sqlite`, CGO 없음) / supervisor → PostgreSQL(`pgx/v5`)
- **sqlc**: 두 엔진 한 `sqlc.yaml` → `workerdb`/`superdb` 분리 생성
- **마이그레이션**: goose (스키마 단일 진실 = goose 파일, sqlc가 schema로 읽음)
- **터미널 출력**: xterm.js (프론트 착수 시)

## 로컬 개발 환경 (DB)
- supervisor PG = docker 컨테이너 **`postgres15`**(이미지 `postgres:15`). 시작: `docker start postgres15` (※ `docker start postgres`는 그런 컨테이너 없음 — 보여지는 건 이미지명)
- 컨테이너 env = supervisor env 일치: user `irony` / pass `!Fa1289` / 포트 `5432`. **단 기본 DB는 `workspace`(구 PortBridge)** → nexus용 `nexus` DB는 **수동 생성함**(`CREATE DATABASE nexus OWNER irony;`, 2026-06-26 1회)
- supervisor 띄우면 `init()`의 mountStore가 goose 마이그레이션 자동 적용(멱등)
- **포트(2026-10-13~)**: 외부 접속 설정 중 = supervisor 8012(0.0.0.0) · 프론트 8011 · worker -> localhost:8012. 옛 5050·3000으로 되돌리는 법 → `REF-external-access.md`
- **함정**: 잔류 supervisor가 5050(지금은 8012) 점유 시 `bind: address already in use`(마이그레이션은 init서 정상, 바인드만 실패). `ss -ltnp | grep :5050`로 pid 찾아 kill

## apps/core 디렉토리
```
cmd/{supervisor,worker}/{main.go, router/, constants/}
internal/
  protocol/ transport/ call/ execute/{,pty}  공유
  subscribe/ manager/ transfer/ util/ web/   재사용 코어
  worker/{store, db/{migrations,query,gen=workerdb}}
  supervisor/{registry,bind,subscribe, store, db/{...,gen=superdb}}
sqlc.yaml  README.md  .gitignore
```

## 모듈 현황 (상세는 각 REF)
| 모듈 | 상태 | 문서 |
|------|------|------|
| 통신 인프라·파일 전송·supervisor web (마감) | 구현 완료(파일 전송 e2e 미검증) | → `INDEX-infra.md` |
| process 실행 (execute/pty/manager/bind) | 배선 마감 → `INDEX-process-wiring.md` / 터미널 기능(입력·SNAPSHOT·EDIT) 마감 → `INDEX-process-terminal.md` / **동기화·공유(5 공유(P·R·T))·보드·presence 마감(2026-10-12)** → `INDEX-process-share.md`. EXEC 실행 방식 = 현상 유지 확정 / **종료 정책(SIGTERM 먼저·끊긴 장비 kill 기록·모두 닫음/두 번째 = SIGKILL) 마감(2026-10-12)**(→ `REF-process-kill.md`) / **바인딩 타일(다른 계정 실행 터미널 위치 · 타일 고정 · 터미널 탭 = 루트 경로 하위 · 현재 타일에 바인딩 · 위치 알림) 마감(2026-10-13)**(→ `REF-tile-binding.md`, `REF-tile-binding-rework.md`, `REF-tile-binding-rework-impl.md`) / **질의 응답 중복(크기 소유 탭만 xterm 질의에 응답) 마감(2026-10-13)**(→ `REF-term-query.md`) / **크기 가져오기(같은 계정·스크립트 소유 계정 우선, 넘김 15초) 마감(2026-10-14)**(→ `REF-size-take.md`) | `REF-process.md` + 마감 INDEX 3개 |
| DB/스토어 (sqlc·goose·store) | 구현 완료 | `REF-db.md` |
| Node 카탈로그 | DB+CRUD+핸들러+PatchNode(`9c9d22e`) + `GET /workers`·exec 대상 장비 검증(`ba5b857`). UI 연동(연동 1차·타일 터미널·⑪(타일 트리 서버 저장)·4 O(계정 동기화) 화면) 마감 → `INDEX-node-ui-link.md`. **roster(명부) 마감**(2026-10-12, `GET/DELETE /workers/instances` + 칩·고르기·장비 지정 다이얼로그, `REF-node-label.md` "roster 계획") · label 보류 / **복사·잘라내기·붙여넣기 + 선택 모드·우클릭·드래그 마감(2026-10-13)** → `REF-node-clipboard.md`·`REF-node-select.md` | `REF-node-label.md`(+`REF-node-ui.md` UI 컨셉 / `REF-node-ui-overview.md` 전체보기·가상 키보드) |
| 공용 PATCH 래퍼 (`internal/patch`) | `patch.Field[T]` 3-state(`{valid,value}`), worker도 재사용 예정 | `REF-node-label.md` |
| 프론트엔드 (apps/frontend) | 스캐폴딩·socket hook·user/login(라우터 가드·401 처리)·전역 다이얼로그 + node 타일 UI(레이아웃 → `INDEX-node-ui-layout.md`, 연동 → `INDEX-node-ui-link.md`, 둘 다 마감) + 터미널 기능(→ `INDEX-process-terminal.md`). ⑮(모바일 가상 키보드) = **마감(2026-10-14) → `INDEX-mobile-keyboard.md`** / user/login 마무리 = **마감(2026-10-12, `REF-frontend-auth.md`)**, 회원가입은 별도 파트로 보류 / active 타일 표시·위치 번호 단축키·다른 계정 active 배지 = **마감(2026-10-12, `REF-node-ui-active.md`)** | `REF-frontend.md`(+`REF-node-ui.md` 컨셉 / `REF-node-ui-overview.md` 전체보기·가상 키보드) |
| 실시간 push (socket) | 전송 토대·3모드 e2e 커밋(3a8e92e). **ping/pong(`transport.KeepAlive`, 10초/25초)로 응답 없는 소켓의 Hub 막힘 해소**(`e395433`, 2026-10-01). node Kind 어휘 확정 + process 동적구독 REST 완료 + **node CRUD 발행처 배선 완료**(`AfterCommit` 훅, 이동=2토픽, 2026-07-16, 빌드/vet 통과·e2e 미검증). **node 변경 전달 = 계정 토픽 + 발행 시점 판정, 마감**(2026-10-12, → `REF-node-list-sync.md`) | `REF-realtime.md` |
| 위젯 (Skeleton/SkeletonGroup/StickyBox/TileLayout) | Skeleton/SkeletonGroup shimmer 로딩 위젯 신설 + StickyBox 컴포넌트 구현·스토어 리팩터(reportSelf/viewportClient, 2026-08-07) + TileLayout(CSS grid 타일 위젯, node UI Phase1 뼈대, 2026-08-19). 셋 다 실사용 미배선(TileLayout은 index.vue 데모만) | `REF-widget.md` |
| 범용 유틸 (EventInterface/Memoized/LifecycleRegistry/GroupedSet) | 4종 + 공유 리사이즈 관측 그룹(`feature/common`, AppHead·StickyBox가 씀, 2026-08-07). 반복 주제(전역 규칙)라 전용 REF | `REF-util.md`(+`REF-util-drag.md` DraggableSession/useDragGhost/GhostArea, 설계 확정·**보류**(2026-10-12)) |

## reference 인덱스
- **★ 반복 주제(프로젝트 무관, 다른 프로젝트에서도 생김)**: `REF-mobile-viewport.md`(모바일 키보드·뷰포트, Edge 하단 막대 오류 실측) / `REF-util.md`(범용 유틸)
- 설계/재사용 지식: `REF-process.md` `REF-db.md` `REF-node-label.md` `REF-node-ui.md` `REF-node-ui-overview.md` `REF-node-ui-active.md`(active 타일·단축키 `Ctrl+Alt+숫자`·다른 계정 active 배지) `REF-frontend.md` `REF-frontend-auth.md`(로그인 상태·라우터 가드·401·AppDialog) `REF-realtime.md` `REF-widget.md` `REF-util.md` `REF-util-drag.md` `REF-test-env.md`(테스트 환경·함정) `REF-test-scripts.md`(날짜별 확인 스크립트 재작성 명세) `REF-process-kill.md`(종료 요청·끊긴 장비 kill·강제 종료·실행 실패 정리) `REF-tile-binding.md`(다른 계정 실행 터미널을 받는 폴더 타일 — 처음 연 타일·고정·보는 보드, 처음 설계) `REF-node-list-sync.md`(폴더 목록 동기화 — node 변경을 볼 수 있는 접속 계정 토픽으로, 마감) `REF-tile-binding-rework.md`(바인딩 규칙 개편 — 결정·구조안) `REF-tile-binding-rework-impl.md`(개편 구현·격리 확인·사용자 확인 목록, 마감) `REF-term-query.md`(질의 응답 중복 — 크기 소유 탭만 xterm 질의에 응답) `REF-node-clipboard.md`(노드 복사·잘라내기·붙여넣기 UI) `REF-node-select.md`(노드 선택 모드·우클릭 메뉴·드래그) `REF-external-access.md`(외부 접속 test1~3.falinux.com = 8011~8013, HTTPS 프록시) `REF-size-take.md`(크기 우선권 즉시 가져오기)
- 작업 이력(주제별): `history/node-label.md` `history/node-ui.md` `history/node-ui-overview.md` `history/node-ui-active.md` `history/frontend.md` `history/frontend-auth.md` `history/realtime.md` `history/widget.md` `history/util.md` `history/util-drag.md` `history/test-env.md` `history/test-scripts.md` `history/process-kill.md` `history/tile-binding.md` `history/tile-binding-rework.md` `history/tile-binding-rework-impl.md` `history/node-list-sync.md` `history/term-query.md` `history/node-clipboard.md` `history/node-select.md` `history/external-access.md` `history/size-take.md`
- **마감 파트(세션 시작 시 안 읽음, REF가 필요하면 INDEX부터)**: ★ 모바일 가상 키보드(마감, 중요) → `INDEX-mobile-keyboard.md` / 통신 인프라(마감) → `INDEX-infra.md` / process 배선(마감) → `INDEX-process-wiring.md` / node UI 레이아웃 설계(마감) → `INDEX-node-ui-layout.md` / node UI 연동(마감) → `INDEX-node-ui-link.md` / process 터미널 기능(마감) → `INDEX-process-terminal.md` / process 동기화·공유·보드(마감) → `INDEX-process-share.md`
- 통신/PTY 상세 PLAN: `PLAN-agent-comm.md` / 구독 모델: `PLAN-subscription.md`
- 현재 진행: `CURRENT.md`
