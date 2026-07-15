# CURRENT

## 현재 날짜
2026-07-14

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## 🎯 다음 작업: node 실시간 연동 + 카탈로그 UI

전송 토대·hook·node CRUD는 모두 커밋 완료. 이제 **node 변경 → supervisor commit 후 socket push → 프론트 트리/캔버스 반영**을 잇고 UI를 올린다.

**남은 배선** (설계 → `REF-realtime.md` / node 백엔드 → `REF-node-label.md`)
- 동적 `MsgSubscribe`/`MsgUnsubscribe` 어휘 + 핸들러 안에서 **DB 인가** → 현재 `subscribe.go`의 `NODE:0` 고정구독 + TEST/TTTT 스모크 대체.
- Kind(MsgType) 어휘 확정: `node.created/moved/deleted/renamed`, `process.output/status`.
- node/process CRUD 핸들러: **commit 후** `subscribeHub.Publish(topic, kind, payload)` (롤백 누설 방지). 자기 echo는 idempotent 재적용(초기).
- 프론트 `on('node.created'|…)` 실제 핸들러 → 트리/터미널 갱신. 컴포넌트 밖 반영 필요 시 `reactive` 모듈 패턴(→ `REF-frontend.md`).
- **UI**: node 카탈로그 = 트리 + 캔버스(% 절대배치). 터미널(xterm.js)은 EXEC/EDIT용 별개(미착수).

---

## 진행 중 / 잔여

### Node 모듈 — 남은 단계 (DB+CRUD+핸들러+PatchNode 커밋 `9c9d22e`, HTTP e2e 미검증)
> 설계 `REF-node-label.md` / 이력 `history/node-label.md`
4. **worker_instances roster**: `00004_worker_instances.sql`(PK main_key,sub_key + last_seen) → register에 DB upsert → `ListInstances(main_key)` → 활성/비활성 = roster ∩/− 레지스트리. ※subkey 위조검증 보류
5. **HTTP e2e**: 가입→세션→node CRUD 왕복 스모크(서버 띄우면 00003 실DB 적용)
6. **label 모듈**: `00005_labels.sql`(labels 자기참조 + node_labels M:N) → query → router
- 핸들러 책임 미적용: parentId owner 일치 검증 / 자기 자손으로 Move 사이클 방지

### process 도메인 배선 (supervisor 측 완료 → worker 실행부·router 제어 남음)
> 설계·결정 상세 `REF-process-wiring.md` / 이력 `history/process-wiring.md`. 재접속·세션 복원은 `REF-process-reconnect.md` / `history/process-reconnect.md`.

**완료 상태**: supervisor 배선(manager/entry/bind/router, 경로 계약 `{WORKER_BASE}/<node.ID>/<proc.Uid>`) + worker 실행부 본체(procs/exec/pump/teardown/input·resize·kill, Cwd 배선) + worker 끊김→PENDING→재접속 재바인딩(`WorkerState`) — 전부 구현·build/vet 통과·e2e 검증 완료(2026-07-01~07-14). 상세 → `history/process-wiring.md`, `history/process-reconnect.md`.

**다음 배선 (우선순위)**
1. **frontend 트리거·제어 + PROC 동적 구독**: 소켓 `Handle(실행요청)`→`router.Exec` / `input(MsgData)`→`Inter.Write` / `kill`→`Inter.Kill`(Ctrl+C=input, 버튼=kill 구분). **`PROC:<uid>` 토픽 동적 구독 배선도 여기 포함**(안 하면 출력이 `subscribeHub.Publish`에서 계속 조용히 폐기됨) — `REF-realtime.md` "동적 구독 어휘"와 합류 지점.
2. **화면복원**: bind에 ring buffer(SNAPSHOT) 상시 적재 + 세션→uid 원장 + 재접속 SNAPSHOT 전송.
   - **세션→uid 원장 구체 설계 완료(2026-07-14), 구현은 미착수**: sid 추출 배선(코드 완료) + `process_subscribers`(process_uid,owner_user_id,sid,created_at) 테이블/쿼리 3종 설계 → `REF-process-reconnect.md` "세션→uid 원장 구체화" 절. 남은 것=마이그레이션·쿼리 파일 생성+`sqlc generate`, CREATE/DELETE 실제 호출 지점(동적 구독·UNSUBSCRIBE 배선과 합류, 위 1번과 동시 진행 필요).
3. EXEC content→실행 세부정책(직접실행 vs `sh -c`).

**결정 필요**: 끊긴 창 입력/kill 거절 vs 큐잉 / 공유 kill 인가 / kill 에스컬레이션.
**정리 잔여(구)**: `register.go` 주석 SendBuffer 테스트. worker `baseDir` 필드·`instanceKey()`가 resolveDest 재설계로 dead code화(정리 여부 판단).

### 프론트 user/login — WIP (커밋 `3a8e92e` 동반, → `REF-frontend.md`)
- `feature/user`(`auth.store.ts` + `api/user.api.ts`), `pages/Login.vue` + `/login` 라우트, `common/api`(api/query util), `feature/layout` 전역 다이얼로그(`AppDialog.vue` + `appDialog.store`). 실서버 연동·가드 마무리 남음.

---

## 미해결 이슈 (이월)
- **PROC 토픽 무구독**(2026-07-13 발견): worker→sup 출력이 `subscribeHub.Publish("PROC:<uid>",...)`에서 구독자 0명이면 조용히 폐기(에러·로그 없음). 프론트 동적 구독 배선 전까진 process 출력이 브라우저에 절대 안 닿음. → 위 "다음 배선" 1번.
- **파일 전송**: 구현 완료 / e2e 미검증. 잔여: e2e 스모크 / abort sentinel (register 임시전송은 주석처리됨)
- **서브키 충돌/위조**: key↔subkey 결속 검증 미구현(node roster에서 닫을지 보류)
- **supervisor 영속성**: registry 메모리 → PG 미착수

## 잔여 (틈날 때)
- `SESSION_KEY` 등 env화(현재 `"irony"` 하드코딩)
- checkSession createdAt=0(pgtype.Timestamptz gob 미직렬화) → sess.Data.ID로 DB 재조회
- `supervisorRouter.go` `getSessionKey`: `req.Cookie(key)` 에러 미가드 → 쿠키 없을 때 nil pointer panic(2026-07-14 발견, → `REF-process-reconnect.md` "세션→uid 원장")
