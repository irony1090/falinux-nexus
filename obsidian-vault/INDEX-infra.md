# INDEX — 통신 인프라 (마감)

> 2026-10-02 색인 2단화로 MEMORY·HISTORY에서 접은 파트(사용자 확인). **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
worker<->supervisor 통신 토대(transport·subscribe Hub·call 상관관계·EVENT 평면·재연결 backoff) + 파일 전송(transfer, e2e 미검증·abort sentinel 잔여) + supervisor web 계층(tx 미들웨어·panic-style 에러·user/session) + 프로젝트 셋업(2026-06-17~06-29). 이후 변경은 각 기능 파트에서 다룸(예: ping/pong은 `REF-realtime.md`).

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- 설계/재사용 지식: `REF-infra.md` `REF-transfer.md` `REF-supervisor-web.md`
- 작업 이력: `history/transport.md` `history/transfer.md` `history/supervisor-web.md` `history/project.md`

## 모듈 현황표 행 (MEMORY 원문)
| 모듈 | 상태 | 문서 |
|------|------|------|
| 통신 인프라 (transport/subscribe/call/EVENT/util) | 구현 완료 | `REF-infra.md` |
| 파일 전송 (transfer) | 구현 완료, e2e 미검증 | `REF-transfer.md` |
| supervisor web (tx·error·user·session) | 구현+e2e 완료 | `REF-supervisor-web.md` |

## 이 파트에만 해당하는 불변 결정
- 없음 — 관련 불변 결정(에러 처리 panic-style, 재연결 reconciliation 등)은 다른 파트에서도 지켜야 해서 MEMORY에 남김

## HISTORY 색인 (원문)
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

### `history/project.md` — 프로젝트 셋업 / worker 기반
- 2026-06-17 Nexus 프로젝트 분리 & vault 신규 구성
- 2026-06-17 apps/core 스캐폴딩 & 레포 초기 커밋
- 2026-06-18 worker 리팩토링(router 패턴) + SQLite 정체성 영속
