# CURRENT

## 현재 날짜
2026-06-29

---

## 🎯 현재 작업: process 도메인 배선 (EXEC 우선)

> EDIT(worker PTY vi 편집)·node "실행"의 전제. 설계=`REF-process.md`(실행타입 EXEC/EDIT + 배선 계획), 영속 모델=`REF-db.md`(processes 모델). EDIT는 EXEC 증명 후 얹음.

### 진행 상황
- **process sqlc 모델 확정**(2026-06-26): `00002_processes.sql` 스키마 + 쿼리 목록 = `REF-db.md` "processes 모델". 가변 레코드 + 출력 sink 나중, uid TEXT PK, type/status VARCHAR 대문자 CHECK, node_id nullable(FK 없음)

### 다음 할 것 (순서)
1. **process 영속**: `00002_processes.sql`(goose) + `query/processes.sql`(sqlc) 작성 → sqlc generate (`REF-db.md` 쿼리 목록대로)
2. **Phase A — worker** `cmd/worker/router/process.go`: exec/input/resize/kill 구현 + `pty.Interactive` 레지스트리(UID→핸들) + NewWorkerRouter 핸들러 등록 + conn 끊김 정리
3. **갭 처리(착수 전)**: ① 출력 **raw 패스스루**로(현 line-buffering=vi 멈춤) ② `ExecInteractive`에 ProcessSpec(Cwd/Env/초기 Rows·Cols) 배선 ③ ~~`ProcessSpec.Type` 필드~~ **완료**(EXEC/EDIT 프로토콜 어휘 구현 — ExecType/Seed/MsgEditResult/EditResult, `REF-process.md`)
4. **Phase B — supervisor** `cmd/supervisor/router/process.go`: Exec()/output/status + `AgentInteractive` 레지스트리 + handleAgentWS에 On(Data)/On(Status) 등록 + worker 끊김→Done(502)
5. **Phase C — 검증**: frontend 없이 임시 드라이버로 Exec("cat") 출력/입력 왕복 + Kill + 상태전이, `-race`

### 구도 핵심 (정정)
- **worker = `pty.Interactive`(실제 PTY)** / **supervisor = `AgentInteractive`(프록시 핸들)**. 양쪽 UID→핸들 레지스트리(KeyValManager)
- 범위 밖(이후): frontend fan-out(bind/subscribe/Hub) → apps/web xterm.js → **EDIT 타입**(seed→read-back)
  - ✅ EDIT seed **운반 수단 준비됨**(2026-06-29): transfer reader 인터페이스화 + `SendBuffer`(메모리 전송)로 디스크 안 거치고 seed 전송 가능 → `REF-transfer.md`. 배선(호출처)은 EDIT 타입 구현 시 연결

---

## ⏸ 보류: Node / Label 모듈 구현 (supervisor PG 카탈로그)

> process 배선 끝나면 재개. 설계 전부 `REF-node-label.md`. **DB+CRUD는 독립 선행 가능**, "실행"=EXEC / "편집"=EDIT 배선 필요.

### 핵심 설계 요약 (착수 기준)
- **단일 `nodes` 테이블** + `kind`('folder'|'script') 한 트리 (inode 패턴)
- 트리 = `parent_id` 자기참조(루트 NULL), 소유 = `owner_user_id→users`
- **worker 귀속 = 폴더 단위** `device_key TEXT NULL`(InstanceKey 문자열, **FK 아님**) → 스크립트는 위로 올라가며 가장 가까운 device_key 상속
- `position_x/y NUMERIC(5,2)` 0~100 실수, **부모별 로컬좌표**, 동률은 name/id 타이브레이크
- CHECK: `content`=script 전용 / `device_key`=folder 전용 / x,y 0~100
- 스크립트 실행=**인라인 전송**(청크 모듈 불필요)
- **Label = 직교 두 번째 트리**(M:N). node 트리는 구조/배치 / label 트리는 횡단분류+공유. 공유를 라벨이 떠안음 → 노드 단위 sharing 테이블 안 만듦

### 구현 순서 (계획 — 사용자가 직접 코딩, 큰 계획만)
1. **node 스키마**: goose `00002_nodes.sql` (nodes 테이블 + CHECK 3종 + 인덱스 `(parent_id, position_x, position_y)`)
2. **node sqlc**: `query/nodes.sql` — Create(folder/script), ListChildren(`WHERE parent_id ORDER BY x,y`), Get, Move/Reorder(parent+x/y 갱신), UpdateContent, Rename, Delete, **device_key 상속 조회**(재귀 CTE로 가장 가까운 조상 device_key)
3. **node 핸들러**: `cmd/supervisor/router/node.go` — echo 기본 핸들러(**panic-style**, `q,err:=TxQueries(c)`/`requireSession`), `mountNodes(e)`. owner_user_id = 세션 user
4. **검증/e2e**: build/vet + 실 서버 CRUD 스모크
5. **label 모듈**: `00003_labels.sql`(labels 자기참조 + node_labels M:N) → `query/labels.sql` → `router/label.go`

### 미정/주의
- device_key 결속 검증 = "그 키가 **런타임 메모리 레지스트리에 살아있나**"(끊긴 worker=키 없음). DB는 문자열만 보관 → 무결성 검증은 실행 시점
- "실행"(EXEC)·"편집"(EDIT) 동작은 process 모듈 MsgExec 배선 후 완성 (지금은 카탈로그 CRUD만). 실행타입 EXEC/EDIT 설계 확정 → `REF-process.md` "실행 타입 EXEC vs EDIT"
- 적용된 마이그레이션 수정은 반영 안 됨 → 항상 새 번호로 추가 (goose 멱등)
- 모든 supervisor echo 핸들러는 **panic-style 규약** 따를 것 (MEMORY "에러 처리 규약")

---

## ✅ 완료 (이번 세션, 상세는 HISTORY 2026-06-26)
- **user 모듈**: 세션 매니저 이식 + sqlc(users) + 가입/로그인/세션/로그아웃 핸들러 — **e2e 9시나리오 통과**
- **supervisor PG 스토어 계층**: 구 PortBridge pgx connectManager 이식 + goose 마이그레이션 자동적용(실 nexus DB에 00001_users 적용 확인)
- **트랜잭션 미들웨어** `tx.go`: 단일 `txScope`+`txMiddleware`+`Tx/TxQueries`, lazy Begin, error·panic 둘 다 롤백
- **에러 처리 = panic-style 회귀**(return-style 폐기, 사용자 번복). `web.Err`→`ClientError` panic→`PanicMiddleware` 렌더
- **agents 테이블 폐기**: worker 메모리 전용 확정 → node의 device 결속은 `device_key TEXT`(FK 아님)

## 잔여 (틈날 때)
- `SESSION_KEY` 등 env화 (현재 `"irony"` 하드코딩)
- checkSession의 createdAt=0 (pgtype.Timestamptz는 gob 직렬화 안 됨) → 필요시 sess.Data.ID로 DB 재조회

## 미해결 이슈 (이월)
- **process 실행 모듈**: 설계 확정(MEMORY "process 모듈 설계"), Step1(메시지 어휘) 완료 / 도메인 배선 미착수 — node/label 이후 후보
- **파일 전송 모듈**: 구현 완료 / e2e 미검증 (HISTORY 2026-06-24). 잔여: e2e 스모크 / register 내 임시 전송 분리 / abort sentinel
- **supervisorRouter register**: `Append` 반환 무시(원자적 점유 미확인) — cleanup은 conn 기준 FindAll로 전환됨
- **서브키 충돌/위조**: key↔subkey 결속 검증 미구현
- **supervisor 영속성**: registry 메모리 → PG 미착수
- (이후) 도메인 구독 계층, apps/web
