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
- **process 영속성**(2026-10-22 개정): process는 **worker 밖 보관 프로세스(`worker -hold`, process당 1개)** 가 PTY·화면 ring을 쥐고, worker는 sqlite 목록(살아 있는 것만)으로 다시 붙는다 → worker·supervisor 재시작에도 터미널·화면 유지 / supervisor 영속(PG, 끝난 지 24시간 지나면 주기 삭제) (→ `REF-process-keep.md` K1~K7). 옛 "worker 휘발(메모리)" 폐기
- **재연결 reconciliation (2026-07-01 개정 — 구 `Done(502)` 비관 폐기)**: worker 끊김→관련 process **PENDING**(죽이지 않음) / 재접속→worker가 자기 live 상태 보고로 재동기화(낙관적). **모든 상태전이=process 라우터 `status` 단일 깔때기**(worker 보고 or 끊김 시 supervisor 합성). **frontend 끊김≠종료**(명시적 kill 전까진 실행 유지, 브라우저 종료 무시), **같은 세션 재접속=보던 화면 복원**(ring+세션원장) (→ "종료/재접속 모델")
  - **worker 자체 종료 = process 살려 둠**(2026-10-22 개정, K4(기본 종료 동작)): 종료 신호를 받아도 worker만 끝나고 보관 프로세스·스크립트는 남는다. 옛 "종료 신호 = 스크립트 정리(TERM -> 5초 -> KILL)"(2026-10-15, `INDEX-worker-orphan.md`) 폐기. 전부 종료 수단은 BACKLOG
- **worker는 Linux·Windows 둘 다**(2026-10-19): OS에 묶인 코드는 `internal/execute/pty`의 `_linux.go`/`_windows.go`(+ worker `script_windows.go`/`script_other.go`·`dir_windows.go`/`dir_other.go`)에만 둔다. **스크립트 노드 문법 = 그 OS에서 그대로 실행되게 작성**(Windows는 노드 이름 확장자로 실행기, `#!` 안 씀). worker를 띄우는 방식이 남긴 신호 무시 속성(Linux SIGHUP·Windows Ctrl+C)은 자식에 상속되므로 worker가 되돌린다 (→ `INDEX-windows-worker.md`)
- **에러 처리 = panic-style**(2026-06-26 재확정): 핸들러 `panic(web.Err(...))` → `PanicMiddleware` 렌더 (→)
- **process 동기화 범위 = 계정**(2026-10-01, 1단계 코드·실행 확인·커밋 완료): 같은 계정이면 탭·브라우저·기기 무관 동기화. 인가 = 계정(DB) / 출력 라우팅 = 탭 소켓(메모리) / PTY 크기 우선권 = 서버 발급 탭 id(메모리) → `process_subscribers` 제거(`00005`). 공유는 계정 단위·소유자만 신청 (→ `REF-process-sync.md`, 구현·확인 → `REF-process-sync-impl.md`)
- **공유 판정 = DB 함수 한 곳**(2026-10-06, 공유-a): 가까운 공유 단계 계산은 `node_permission`/`process_permission`(`00007_shares.sql`)에만 두고 쿼리·Go는 호출만. 단계 = 0 차단·10 보기·20 조작·30 편집·40 소유자(함수 반환만), 비교는 `>=`. 볼 수 없음 = 404 / 단계 미달 = 403. **공유 단위 = 폴더 + process 단독**(스크립트 노드 공유 금지, 2026-10-07 SH1 개정) (→ `REF-process-share-db.md`, `REF-process-share-authz.md`, `REF-process-share.md`)
- **실시간 push = 인가/라우팅 분리**(2026-06-30): 인가=DB(구독 시점 1회) / 라우팅=메모리(`subscribe.Hub` topic Publish, DB 무접촉). 수신자=presence(구독)≠소유자 → 공유·다중 유저 동일 화면 흡수. 토픽 단위: process 출력 = `PROCESS:<uid>` / **node 변경 = 볼 수 있는 접속 계정의 `USER:<userId>`(2026-10-21 `TILES:`에서 이름 바꿈 — 옛 문서의 `TILES:<userId>` = 같은 것), 발행 시점 DB 판정**(2026-10-12 개정 — 옛 "펼친 폴더 `NODE:<parentId>`"는 구독 코드가 없어 폐기, → `REF-node-list-sync.md`). "DB≠라우팅 authority"를 browser 평면으로 확장. 비대칭: 브라우저 call→서버 Handle / 서버 Emit→브라우저 on (→ `REF-realtime.md`)

- **모바일 키보드 = `resizes-content` + 열린 동안 배치 영역은 키보드 전 높이(px) 고정**(2026-10-14): 타일 재배치 금지. `dvh`로 배치하는 새 영역도 지킬 것 (→ `REF-mobile-viewport.md` 4). **2026-10-20 개정**: 포커스 터미널 타일만 키보드 위 영역에 맞춤(그 PTY rows만 바뀜) (→ `REF-board-keyboard.md` "맞춤 계획")

- **타일 이동 = 페이지 표시의 "타일 위치 규칙"**(2026-10-21): 타일로 가는 경로는 전부 `reveal`(= `revealCol`, 화면 밖일 때만 막 다 보이는 만큼). Grid 단위 이동은 "Grid로 가기"가 목적일 때만 (→ `REF-tile-move.md`)

- **앱 단축키 = 검사 한 곳 `appShortcutsOff`(`tileGrids.store`)를 맨 앞에서 거친다**(2026-10-21): 터미널 타일마다 "단축키 끄기"(타일 트리 `shortcutsOff`) — 그 타일이 active면 앱 단축키 전부 꺼짐. **새 단축키도 반드시 이 검사를 넣는다**. 터미널 자체 동작(Ctrl+C 복사·Ctrl+V 붙여넣기)은 대상 아님 (→ `REF-tile-move.md` 이동-3)
- **타일 종류 = 능력으로 묻는다**(2026-10-22, 파트(타일 종류 정리)): 여러 종류가 같이 쓰는 로직은 `type` 비교 대신 능력-1(루트 가능)·2(자식 가능)·3(process 담음) — 서버 `router/tileKind.go` `tileKinds` / 프론트 `widget/util/tileKind.util.ts` `TILE_KINDS`(두 표를 같게). 화면 = 공통 틀 `TileFrame` + 종류 컴포넌트(`tileComponents.ts`). 새 종류 추가 절차 → `REF-tile-kinds.md` "새 종류를 더할 때"(첫 사례 = 장비 타일 `device`, `INDEX-worker-dir`)

- **전송 = supervisor 전송 목록(메모리) + 계정 토픽 `TRANSFER:*` push**(2026-10-23, 파트(파일 전송 확장)): 업로드·내려받기·worker 간 전송이 같은 목록을 쓴다(Kind만 추가). 브라우저 -> supervisor 본문은 **HTTP 조각 512KB**(앞단 nginx `client_max_body_size` 기본 1MB). worker -> supervisor 읽기는 끌어오기(`FILE_OPEN`·`READ`·`CLOSE`). 장비 디렉토리 인가 = `requireDeviceDir` (→ `REF-transfer-ext.md`·`REF-transfer-down.md`)

## 기술 스택 (확정)
- **레포**: `nexus` / 모노레포 `apps/core`(Go) + `apps/frontend`(프론트, 스캐폴딩 완료 2026-06-30)
- **프론트**: Vue3 + TS + Vuetify4 (scratch preset, Router=standard / **Pinia 미사용**, CSS framework 없음). 상태공유=provide/inject(+컴포넌트 밖 필요 시 reactive 모듈 패턴). 기본폰트=Noto Sans Korean. **색 = Vuetify 테마 색 이름 의존, hex 금지**(2026-09-29), 공유 표시 = 테마 `share`(청록). **글자 크기 = Vuetify(MD3) 단계 이름 `@include vars.font('body-small')`, px 금지**(2026-10-23, 단계 값 한 곳 `styles/_typography.scss`·전체 배율 `$font-scale`, → `INDEX-typography`) (→ `REF-frontend.md`)
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
- **빌드본 = 사용용(2026-10-22~)**: `build/<이름>`(`scripts/build.sh`) + 공통 `build/{supervisor,worker}.env`·`worker-sqlite.db`, **8013(test3) · DB `nexus_build`**. 개발(8011·8012, `nexus`)과 분리, 사용자 Claude 세션이 그 위에서 돎. **Claude는 8013을 끄지 않는다**. 재기동은 사용자 몫, Claude가 띄우면 `CLAUDE*` 변수를 지우고(상속되면 그 위 Claude가 transcript 저장 안 함, → `REF-build.md`). 빌드본 worker 교체 = 그 위 process 전부 종료(사용자 세션 포함, supervisor만 교체는 PENDING 뒤 이어짐). embed `cmd/*/env` = 배포 값 / **개발 실행 = `go run ./cmd/supervisor -env dev.env` 필수** (→ `INDEX-build.md` -> `REF-build.md`)
- **Windows VM `nexus-win`(RAM 6G) = 시험할 때만 켠다(2026-10-23 사용자)**: 작업(코드 작성) 중엔 `docker stop nexus-win`, Windows 확인 단계에서 `docker start` -> 확인 끝나면 바로 끈다 (→ `REF-windows-vm.md`)
- **함정**: 잔류 supervisor가 5050(지금은 8012) 점유 시 `bind: address already in use`(마이그레이션은 init서 정상, 바인드만 실패). `ss -ltnp | grep :5050`로 pid 찾아 kill

## apps/core 디렉토리
```
cmd/{supervisor,worker}/{main.go, router/, constants/}
internal/
  protocol/ transport/ call/ execute/{,pty}  공유
  subscribe/ manager/ transfer/ util/ web/   재사용 코어
  worker/{store, db/{migrations,query,gen=workerdb}}
  supervisor/{bind, store, db/{...,gen=superdb}}
sqlc.yaml  README.md  .gitignore
```

## 모듈 현황 (상세는 각 REF)
| 모듈 | 상태 | 문서 |
|------|------|------|
| 통신 인프라·파일 전송·supervisor web | 마감 | `INDEX-infra.md` · `INDEX-transfer-e2e.md` |
| process 실행 (execute/pty/manager/bind) | 배선·터미널 기능·동기화/공유/보드·종료 정책·고아 스크립트·Windows worker 전부 마감. EXEC 실행 방식 = 현상 유지 확정 | `REF-process.md` + 마감 INDEX(`HISTORY.md` "마감 파트") |
| DB/스토어 (sqlc·goose·store) | 구현 완료 | `REF-db.md` |
| Node 카탈로그 | Node 모듈·roster(명부)·폴더 목록 동기화·노드 클립보드·UI 연동 마감, label 보류 | `REF-node-label.md` + `INDEX-node-ops.md` |
| 공용 PATCH 래퍼 (`internal/patch`) | `patch.Field[T]` 3-state(`{valid,value}`), worker도 재사용 예정 | `REF-node-label.md` |
| 프론트엔드 (apps/frontend) | 타일 UI·터미널·⑮(모바일 가상 키보드)·로그인·active 타일·보드 미리보기 전부 마감, 회원가입 보류 | `REF-frontend.md` · `REF-node-ui.md`(컨셉) · `REF-node-ui-overview.md`(전체보기·가상 키보드) |
| 실시간 push (socket) | 전송 토대·3모드·ping/pong·node CRUD 발행(`AfterCommit`)·계정 토픽, 마감 | `REF-realtime.md` |
| 위젯 (Skeleton/SkeletonGroup/StickyBox/TileLayout) | 구현만, 실사용 미배선(TileLayout = index.vue 데모만) | `REF-widget.md` |
| 범용 유틸 (EventInterface/Memoized/LifecycleRegistry/GroupedSet) | 4종 + 공유 리사이즈 관측 그룹(`feature/common`), 반복 주제 | `REF-util.md` · `REF-util-drag.md`(드래그, 보류) |

## reference 인덱스
- **★ 반복 주제(프로젝트 무관, 다른 프로젝트에서도 생김)**: `REF-mobile-viewport.md`(모바일 키보드·뷰포트, Edge 하단 막대 오류 실측) / `REF-util.md`(범용 유틸)
- 설계/재사용 지식(위 모듈 현황표에 있는 것 말고): `REF-test-env.md`(테스트 환경·함정) · `REF-test-scripts*.md`(확인 스크립트 재작성 명세, 날짜별 5개) · `REF-external-access.md`(외부 접속 test1~3 = 8011~8013) · `REF-demo.md`(시연 데이터) · `REF-worker-screen.md`(worker 화면 송출, 아이디어) · `REF-worker-dir.md`(worker 폴더 타일 전체 계획, 폴더-7(내려받기)·9(worker 간 전송) 원래 표)
- 작업 이력: `HISTORY.md`가 색인(`history/{주제}.md`, REF와 1:1)
- **마감 파트(세션 시작 시 안 읽음, REF가 필요하면 INDEX부터)**: 목록·최근 순서 = `HISTORY.md` "마감 파트"(최근 = 글자 크기 규칙 `INDEX-typography`)
- 통신/PTY 상세 PLAN: `PLAN-agent-comm.md` / 구독 모델: `PLAN-subscription.md`
- 현재 진행: `CURRENT.md` / 나중에 할 것: `BACKLOG.md`(세션 시작 시 안 읽음)
