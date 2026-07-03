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

### `history/process.md` — process 실행 모듈 (supervisor 측)
- 2026-07-03 folder-open 버그 수정 + worker EDITOR env + PTY 실행부 선행(ExecInteractive env·Pid 접근자)
- 2026-07-01(2) process 도메인 배선 완료 + 경로 조립(sup)/치환(worker) 책임 분리 `{WORKER_BASE}/nodeID/uid`
- 2026-07-01 supervisor ProcessManager·bind·router 골격 + 종료/재접속 모델(status 깔때기·worker끊김→PENDING) 확정

### `history/frontend.md` — 프론트엔드 (apps/frontend)
- 2026-06-30 apps/frontend 초기 스캐폴딩 (Vue3+TS+Vuetify, Router/Pinia X, Noto 폰트, .gitignore, 커밋 e511359)
- 2026-06-30 폰트 마무리 정리 (Roboto 제거, Noto 기본폰트 적용 확인)
- 2026-06-30 user/login + 공용 API + 전역 다이얼로그 WIP (커밋 3a8e92e 동반)

### `history/realtime.md` — 실시간 push (socket)
- 2026-06-30 supervisor↔웹 socket 전송 토대 완성 + 3모드(call/emit/on) e2e 검증 (Hub Kind 추가·subscribe.go 인증교정·프론트 hook 재설계)

### `history/project.md` — 프로젝트 셋업 / worker 기반
- 2026-06-17 Nexus 프로젝트 분리 & vault 신규 구성
- 2026-06-17 apps/core 스캐폴딩 & 레포 초기 커밋
- 2026-06-18 worker 리팩토링(router 패턴) + SQLite 정체성 영속
