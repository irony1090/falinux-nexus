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
- 2026-08-07 컨셉 설계(코드 없음): 배치도(캔버스)=홈+트리=보조내비, PC 오버레이 토글/모바일 세그먼트 토글, 트리 재배치 드래그(device_key 상속변경·좌표NULL 이슈 발견). 모바일 트리 형태만 미정

### `history/frontend.md` — 프론트엔드 (apps/frontend)
- 2026-06-30 apps/frontend 초기 스캐폴딩 (Vue3+TS+Vuetify, Router/Pinia X, Noto 폰트, .gitignore, 커밋 e511359)
- 2026-06-30 폰트 마무리 정리 (Roboto 제거, Noto 기본폰트 적용 확인)
- 2026-06-30 user/login + 공용 API + 전역 다이얼로그 WIP (커밋 3a8e92e 동반)

### `history/realtime.md` — 실시간 push (socket)
- 2026-07-16 (2) node CRUD 발행처 배선 구현 완료(`AfterCommit` 훅 신설 + create/patch/delete 3핸들러 배선, 이동=2토픽)
- 2026-07-16 (1) node 도메인 Kind 어휘 확정(`NODE:CREATE/UPDATE/DELETE`, `node.change` 단일봉투안 기각) + process 동적구독은 REST로 결론(상세는 process-reconnect.md)
- 2026-06-30 supervisor↔웹 socket 전송 토대 완성 + 3모드(call/emit/on) e2e 검증 (Hub Kind 추가·subscribe.go 인증교정·프론트 hook 재설계)

### `history/widget.md` — 위젯 컴포넌트 (feature/widget)
- 2026-08-07 Skeleton/SkeletonGroup shimmer 로딩 위젯 신설(소비처 미배선)

### `history/util.md` — 범용 유틸 (common/util)
- 2026-08-07 `EventInterface`+`LifecycleRegistry` 신설, `Memoized` 확장(EventInterface 상속+create/remove 이벤트)

### `history/project.md` — 프로젝트 셋업 / worker 기반
- 2026-06-17 Nexus 프로젝트 분리 & vault 신규 구성
- 2026-06-17 apps/core 스캐폴딩 & 레포 초기 커밋
- 2026-06-18 worker 리팩토링(router 패턴) + SQLite 정체성 영속
