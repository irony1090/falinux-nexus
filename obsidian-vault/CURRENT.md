# CURRENT

## 현재 날짜
2026-06-30

---

## ✅ 프론트엔드 스캐폴딩 완료 (2026-06-30, 커밋 e511359)

> `apps/frontend` 생성 — Vue3 + TS + Vuetify4(scratch, Router=standard / **Pinia X** / CSS framework 없음). 상세 → `REF-frontend.md`, 이력 → `history/frontend.md`.
- Noto Sans Korean 폰트 배치(`public/fonts/` + `noto-sans-korean.scss` `@font-face`), `main.ts` import.
- `.gitignore` 작성(`.env`/node_modules/dist 무시), 81개 파일 커밋.
- 상태관리 = provide/inject (컴포넌트 밖 접근 필요 시 reactive 모듈 패턴 검토).

### 프론트 마무리 잔여 (선택)
- 기본폰트 적용 = Vuetify `$body-font-family` 오버라이드(`settings.scss`) 확인.
- Roboto 정리: `@fontsource/roboto` 삭제 + `import 'unfonts.css'` 제거(Fonts 플러그인 주석화 상태).

---

## 🎯 다음 작업: node 실시간 연동 + UI

> node CRUD(REST)는 구현 완료, 프론트 스캐폴딩 완료. 이제 **supervisor가 node 변화를 socket으로 프론트에 push**(실시간 반영) + node 카탈로그 UI(트리+캔버스) 구현.

### 방향 (설계 필요)
- **흐름**: node 변경(생성/이동/수정/삭제) 발생 → supervisor가 socket으로 결과 push → 프론트 트리/캔버스 갱신
- **미설계(내일 정할 것)**:
  - socket 채널/이벤트 형식(어떤 단위로 push: 노드 1건 delta vs 부모 자식목록 갱신)
  - 구독 범위(유저별 node 트리) — 기존 `internal/subscribe`(Hub)·EVENT 평면 재사용 여부
  - REST(요청) ↔ socket(브로드캐스트) 정합(자기 요청 결과도 socket으로 받을지/낙관적 갱신)
- **프론트**: `apps/frontend` 스캐폴딩 완료. node 카탈로그 UI = 트리 + 캔버스(% 절대배치) / 터미널(xterm.js)은 EXEC·EDIT용 별개 (UI 미착수)

---

## ✅ Node 모듈 — 이번 세션 구현 (DB+CRUD+핸들러 / HTTP e2e 미검증 / **uncommitted**)

> 설계 전부 `REF-node-label.md`, 구현 상세 `history/node-label.md`(2026-06-29). build/vet/gofmt 통과.

### 완료 (steps 1~3)
1. **`00003_nodes.sql`**(00002는 processes 선점) — nodes + CHECK 5종 + 인덱스. 실DB 롤백검증
2. **`query/nodes.sql`**(sqlc) — Create(Folder/Script)/Get/ListChildren/MaxChildOrd/Move/Place/UpdateContent/Rename/Delete/ResolveDeviceKey + **PatchNode**(tri-state). owner 스코핑
3. **`node.go` 핸들러**(REST) + **`internal/patch`**(공용 `patch.Field[T]`) + **`nodeDto.go`**(DTO·변환·bind·toParams 분리)

### 남은 단계
4. **worker_instances roster**: `00004_worker_instances.sql`(PK main_key,sub_key + last_seen) → register에 DB upsert 배선 → `ListInstances(main_key)`(레지스트리 FindAll) → 활성/비활성 = roster ∩/− 레지스트리. ※subkey 위조검증 **보류**
5. **HTTP e2e**: 가입→세션→node CRUD 왕복 스모크(서버 띄우면 00003 실DB 적용됨)
6. **label 모듈**: `00005_labels.sql`(labels 자기참조 + node_labels M:N) → query → router

### 핵심 설계 요약 (상세 `REF-node-label.md`)
- 단일 `nodes` + `kind`('FOLDER'|'SCRIPT') inode 트리. parent 자기참조(루트 NULL, CASCADE), 소유=owner_user_id
- **device_key = main_key만**(폴더=장비 클래스). 스크립트는 최근접 조상 device_key 상속, subkey는 Exec 시점 선택
- **좌표 분리**: `position_x/y NUMERIC(5,2)`=캔버스 / `ord BIGINT` 정수=그리드 순서(인접삽입=재번호)
- **worker_instances roster**(DB 관측용≠authority): online은 레지스트리 파생
- **PatchNode**: 프론트 `{valid,value}`→`patch.Field[T]`→NOT NULL은 COALESCE/nullable은 set_* 플래그

### 미정/주의
- device_key(main_key) 결속 검증 = "그 main_key로 ≥1 인스턴스 살아있나"(런타임). 검증은 실행 시점
- 핸들러 책임 미적용: parentId의 owner 일치 / 자기 자손으로 Move 사이클 방지
- 적용된 마이그레이션 수정 금지 → 항상 새 번호(goose 멱등) / supervisor echo 핸들러는 panic-style 규약

---

## ⏸ 보류: process 도메인 배선 (EXEC 우선) — **uncommitted WIP 있음**

> EDIT·node "실행"의 전제. 설계 `REF-process.md`. 작업트리에 미커밋 진행분이 남아 있음.

### 미커밋 WIP (정리 후 커밋 필요)
- `messages.go`(EDIT 어휘 + **Seed 인라인 폐기** → 초기내용은 SendBuffer 선전송), worker/supervisor `process.go`(배선 일부 + **디버그 로그**), `register.go`(**임시 SendBuffer 테스트** + 주석 Exec), `env`(DB_* 추가)
- → 디버그 로그·register 임시테스트 정리 후 `feat(process)` 커밋

### 다음 할 것 (순서)
1. **갭 처리**: ① 출력 raw 패스스루(line-buffering=vi 멈춤) ② `ExecInteractive`에 ProcessSpec(Cwd/Env/Rows·Cols) 배선
2. **Phase A — worker** `process.go`: exec/input/resize/kill + `pty.Interactive` 레지스트리(UID→핸들) + conn 끊김 정리
3. **Phase B — supervisor** `process.go`: Exec()/output/status + `AgentInteractive` 레지스트리 + worker 끊김→Done(502)
4. **Phase C — 검증**: 임시 드라이버로 Exec("cat") 왕복 + Kill + 상태전이, `-race`
- 구도: worker=`pty.Interactive`(실제 PTY) / supervisor=`AgentInteractive`(프록시). 양쪽 UID→핸들 레지스트리
- EDIT seed 운반은 `SendBuffer`(메모리 전송) 준비됨 → EDIT 타입 구현 시 연결

---

## ✅ 완료 (이전, 상세 HISTORY)
- **transfer reader 인터페이스화 + SendBuffer**(2026-06-29, 커밋 713c50f/4284bb7) — `REF-transfer.md`
- **user 모듈**(가입/로그인/세션, e2e 9시나리오) / **supervisor PG 스토어 + tx 미들웨어** / **에러 = panic-style**(2026-06-26)

## 잔여 (틈날 때)
- `SESSION_KEY` 등 env화(현재 `"irony"` 하드코딩)
- checkSession createdAt=0(pgtype.Timestamptz gob 직렬화 안 됨) → sess.Data.ID로 DB 재조회

## 미해결 이슈 (이월)
- **파일 전송**: 구현 완료 / e2e 미검증. 잔여: e2e 스모크 / register 임시전송 분리 / abort sentinel
- **서브키 충돌/위조**: key↔subkey 결속 검증 미구현(node roster에서 닫을지 보류)
- **supervisor 영속성**: registry 메모리 → PG 미착수
