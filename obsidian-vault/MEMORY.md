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
  - **관측용 roster는 DB 허용**: `worker_instances(main_key, sub_key, last_seen)` = UI에 main_key별 인스턴스 활성/비활성 표시용. online은 저장 말고 레지스트리로 **파생**(roster ≠ authority). 착수 예정 (→ `REF-node-label.md`)
  - node의 device 결속 = FK 아닌 `device_key TEXT`(**= main_key**, 폴더는 장비 클래스에 묶임 / subkey는 실행 시 런타임 선택) (→ `REF-node-label.md`)
- **agent 식별자: 사전 지정 메인키 + supervisor가 접속 시 부여하는 서브키.** 저장/조회는 `메인키#서브키`(`InstanceKey()`). MAC 폐기
- **process 영속성**: worker 휘발(메모리) / supervisor 영속(PG, 재시작 복구)
- **재연결 reconciliation (2026-07-01 개정 — 구 `Done(502)` 비관 폐기)**: worker 끊김→관련 process **PENDING**(죽이지 않음) / 재접속→worker가 자기 live 상태 보고로 재동기화(낙관적). **모든 상태전이=process 라우터 `status` 단일 깔때기**(worker 보고 or 끊김 시 supervisor 합성). **frontend 끊김≠종료**(명시적 kill 전까진 실행 유지, 브라우저 종료 무시), **같은 세션 재접속=보던 화면 복원**(ring+세션원장) (→ `REF-process-reconnect.md` "종료/재접속 모델")
- **에러 처리 = panic-style**(2026-06-26 재확정): 핸들러 `panic(web.Err(...))` → `PanicMiddleware` 렌더 (→ `REF-supervisor-web.md`)
- **실시간 push = 인가/라우팅 분리**(2026-06-30): 인가=DB(구독 시점 1회) / 라우팅=메모리(`subscribe.Hub` topic Publish, DB 무접촉). 수신자=presence(구독)≠소유자 → 공유·다중 유저 동일 화면 흡수. 토픽 단위=**펼친 폴더** `NODE:<parentId>`. "DB≠라우팅 authority"를 browser 평면으로 확장. 비대칭: 브라우저 call→서버 Handle / 서버 Emit→브라우저 on (→ `REF-realtime.md`)

## 기술 스택 (확정)
- **레포**: `nexus` / 모노레포 `apps/core`(Go) + `apps/frontend`(프론트, 스캐폴딩 완료 2026-06-30)
- **프론트**: Vue3 + TS + Vuetify4 (scratch preset, Router=standard / **Pinia 미사용**, CSS framework 없음). 상태공유=provide/inject(+컴포넌트 밖 필요 시 reactive 모듈 패턴). 기본폰트=Noto Sans Korean (→ `REF-frontend.md`)
- **Go**: 단일 모듈 `nexus`, 멀티 cmd(`cmd/supervisor`, `cmd/worker`), Go 1.26
- **DB**: worker → SQLite(`modernc.org/sqlite`, CGO 없음) / supervisor → PostgreSQL(`pgx/v5`)
- **sqlc**: 두 엔진 한 `sqlc.yaml` → `workerdb`/`superdb` 분리 생성
- **마이그레이션**: goose (스키마 단일 진실 = goose 파일, sqlc가 schema로 읽음)
- **터미널 출력**: xterm.js (프론트 착수 시)

## 로컬 개발 환경 (DB)
- supervisor PG = docker 컨테이너 **`postgres15`**(이미지 `postgres:15`). 시작: `docker start postgres15` (※ `docker start postgres`는 그런 컨테이너 없음 — 보여지는 건 이미지명)
- 컨테이너 env = supervisor env 일치: user `irony` / pass `!Fa1289` / 포트 `5432`. **단 기본 DB는 `workspace`(구 PortBridge)** → nexus용 `nexus` DB는 **수동 생성함**(`CREATE DATABASE nexus OWNER irony;`, 2026-06-26 1회)
- supervisor 띄우면 `init()`의 mountStore가 goose 마이그레이션 자동 적용(멱등)
- **함정**: 잔류 supervisor가 5050 점유 시 `bind: address already in use`(마이그레이션은 init서 정상, 바인드만 실패). `ss -ltnp | grep :5050`로 pid 찾아 kill

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
| 통신 인프라 (transport/subscribe/call/EVENT/util) | 구현 완료 | `REF-infra.md` |
| process 실행 (execute/pty/manager/bind) | supervisor+worker 양측 배선 완료(빌드/vet 통과, e2e 확인): manager·entry·path·relay·router + worker procs/exec/pump/teardown + worker 끊김→PENDING→재접속 재바인딩(`WorkerState`) + 세션→uid 원장(`process_subscribers`, REST 구독/해지+`browsers` registry) + **frontend 트리거(exec/kill REST, 자동구독)+종료 후 Hub 구독정리**(`startRelay`/`cleanupProcessTopic`) 전부 구현 완료(2026-07-16). **kill 실사용 테스트로 발견한 상태동기화 버그 3건 수정 완료**(PENDING 오삭제/entry.Record 박제/`Status()` 에러계약, 2026-07-22). **resize 배선 완료**(`Layout` 계약 `syscall.Errno`→`error` 정정 + 결과값 기반 DB/memory 동기화 + `MsgProcessUpdate` 발행, 2026-07-22) + **프론트 `ProcessDialog`(xterm) 착수**(DATA 출력 연동 완료). 남은 것=input(키입력)+ring buffer SNAPSHOT(설계 논의만 진행, 미착수 — `REF-process-snapshot.md`) | `REF-process.md`(+`-wiring`/`-trigger`/`-reconnect`/`-subscription`/`-snapshot`/`-resize`) |
| 파일 전송 (transfer) | 구현 완료, e2e 미검증 | `REF-transfer.md` |
| DB/스토어 (sqlc·goose·store) | 구현 완료 | `REF-db.md` |
| supervisor web (tx·error·user·session) | 구현+e2e 완료 | `REF-supervisor-web.md` |
| Node 카탈로그 | DB+CRUD+핸들러+PatchNode 커밋(9c9d22e, e2e 미검증), roster/label 남음 | `REF-node-label.md`(+`REF-node-ui.md` UI 컨셉) |
| 공용 PATCH 래퍼 (`internal/patch`) | `patch.Field[T]` 3-state(`{valid,value}`), worker도 재사용 예정 | `REF-node-label.md` |
| 프론트엔드 (apps/frontend) | 스캐폴딩(e511359) + socket hook 3모드 e2e(3a8e92e/e28252b) + user·login·전역다이얼로그 WIP + node/process REST 클라이언트(2026-07-21, node는 UI 미연결) + websocket hook provide/inject 전환 + **`ProcessDialog`(xterm) 착수**(2026-07-22, exec/kill/resize+DATA 출력 연동, `PROCESS:UPDATE`/`STATUS`는 스텁). node 카탈로그 UI = 컨셉 설계 완료(2026-08-07, 배치도+트리+모바일 대응), 구현 미착수 | `REF-frontend.md`(+`REF-node-ui.md` UI 컨셉 / `REF-process-resize.md` "ProcessDialog" 절) |
| 실시간 push (socket) | 전송 토대·3모드 e2e 커밋(3a8e92e). node Kind 어휘 확정 + process 동적구독 REST 완료 + **node CRUD 발행처 배선 완료**(`AfterCommit` 훅, 이동=2토픽, 2026-07-16, 빌드/vet 통과·e2e 미검증). 남은 것=NODE 동적구독 어휘·프론트 수신 | `REF-realtime.md` |
| 위젯 (Skeleton/SkeletonGroup/StickyBox) | Skeleton/SkeletonGroup shimmer 로딩 위젯 신설(2026-08-07). 소비처 미배선 | `REF-widget.md` |
| 범용 유틸 (EventInterface/Memoized/LifecycleRegistry) | LifecycleRegistry 신설 + Memoized를 EventInterface 상속으로 확장(2026-08-07). 반복 주제(전역 규칙)라 전용 REF 분리 | `REF-util.md` |

## reference 인덱스
- 설계/재사용 지식: `REF-infra.md` `REF-process.md` `REF-process-exec-edit.md` `REF-process-wiring.md` `REF-process-trigger.md` `REF-process-resize.md` `REF-process-reconnect.md` `REF-process-subscription.md` `REF-process-snapshot.md` `REF-transfer.md` `REF-db.md` `REF-supervisor-web.md` `REF-node-label.md` `REF-node-ui.md` `REF-frontend.md` `REF-realtime.md` `REF-widget.md` `REF-util.md`
- 작업 이력(주제별): `history/transport.md` `history/transfer.md` `history/supervisor-web.md` `history/node-label.md` `history/node-ui.md` `history/process-wiring.md` `history/process-trigger.md` `history/process-resize.md` `history/process-reconnect.md` `history/process-subscription.md` `history/process-snapshot.md` `history/frontend.md` `history/realtime.md` `history/project.md` `history/widget.md` `history/util.md`
- 통신/PTY 상세 PLAN: `PLAN-agent-comm.md` / 구독 모델: `PLAN-subscription.md`
- 현재 진행: `CURRENT.md`
