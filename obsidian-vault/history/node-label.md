# HISTORY — Node / Label 모듈

> frontend 카탈로그(nodes 트리 + labels 분류) 설계·구현 상세 기록.
> 확정 설계는 `REF-node-label.md`, 현재 상황은 `CURRENT.md`.

---

### 2026-06-29 - Node 모듈 구현 (스키마·쿼리·핸들러, build/vet 통과 / HTTP e2e 미검증)

> 설계 정련 후 nodes 카탈로그를 DB→쿼리→REST 핸들러까지 구현. 설계 확정은 `REF-node-label.md`.

#### 설계 정련 (이번 세션 확정)
- **device_key/content 분리 유지**: kind로 상호배타지만 합치지 않음(NULL은 PG에서 공짜, 의미·인덱스·향후 분기 위해 분리 = STI 정석). device_key는 부분인덱스(역조회)·content는 비인덱스 대용량
- **device_key = main_key만**(서브키 제거): 폴더는 장비 클래스(main_key)에 묶이고 한 main_key 아래 다중 인스턴스 가능 → **실행 subkey는 Exec 시점 선택**. 결속 검증 = "그 main_key로 ≥1 인스턴스 살아있나"
- **좌표 분리**: x/y(과거 캔버스+그리드 겸용 → 강결합)를 분리. `position_x/y`=캔버스 절대배치 / `ord`=그리드 순서. 처음 NUMERIC(midpoint)였다가 **BIGINT 정수로 변경**(pgtype.Numeric 번거로움 제거, 인접삽입=재번호 수용)
- **kind enum 대문자** `'FOLDER'|'SCRIPT'`(processes의 EXEC/EDIT 컨벤션 일치)
- **worker_instances roster 결정**(DB 허용): "agents 제거는 영속 금지 원칙이 아니라 이름·역할 불명 탓"이라는 사용자 해명 → 관측용 roster는 OK(online은 레지스트리 파생, roster≠authority). **subkey 위조검증은 보류**
- `parent_id ON DELETE CASCADE`(폴더 삭제=subtree 동반)

#### 구현물
- **`00003_nodes.sql`**(00002는 processes 선점): nodes(BIGSERIAL id, parent 자기참조 CASCADE, kind, name, ord BIGINT, x/y NUMERIC(5,2) NULL, device_key/content TEXT NULL) + CHECK 5종 + 인덱스 `(parent_id,ord)`/`(device_key) WHERE NOT NULL`. 실DB 롤백검증(제약 4종 위반 거부 확인)
- **`query/nodes.sql`**(sqlc): CreateFolder/CreateScript, GetNode, ListChildren(`IS NOT DISTINCT FROM`로 루트커버), MaxChildOrd, MoveNode, PlaceNode, UpdateNodeContent, RenameNode, DeleteNode, ResolveDeviceKey(재귀 CTE 최근접 조상 — CTE 컬럼목록 명시로 sqlc 모호성 해소), PatchNode. 전부 owner 스코핑
- **`PatchNode` 부분수정**: NOT NULL(name/ord)=COALESCE(narg) / nullable(parent/x/y/device/content)=`set_* 플래그 + 값` CASE. tri-state 실DB 검증(skip/null/값)
- **`internal/patch`**(신규 공용 패키지): `patch.Field[T]` = 프론트 `{valid,value}` 3-state 래퍼(순수 JSON, DB 무관 → worker도 재사용 예정). pgtype 매핑은 supervisor 전용
- **`cmd/supervisor/router/node.go`**: REST(POST/GET목록/GET:id/PATCH/DELETE) panic-style + TxQueries/requireSession + mountNodes 배선
- **`nodeDto.go`**: DTO + pgtype변환 + `bindCreateRequest`/`bindPatchRequest`(Bind·Validate 흡수) + `nodePatchRequest.toParams()`(매핑 흡수) → 핸들러 슬림화

#### 미검증/남음
- **HTTP e2e 미검증**(build/vet/gofmt만). 가입→세션→CRUD 왕복 안 돌림(서버 띄우면 00003 실DB 적용됨)
- 핸들러 책임 미적용: parentId의 owner 일치 검증 / 자기 자손으로 Move 사이클 방지
- 남은 단계: worker_instances(00004) → label(00005) → HTTP e2e

---

### 2026-06-26 - Node/Label 모듈 설계 (frontend 카탈로그, 미구현)

> 코드 변경 없음. process 모듈 착수 전, frontend가 worker의 "앱/셸"을 나열·실행하기 위한 supervisor PG 카탈로그를 설계만 확정. 상세 결론은 `REF-node-label.md`.

#### 논의 경로 (요약)
- 사용자 요구를 좁혀가는 대화: 처음엔 "frontend에서 process 실행용 파일 모듈" → AI가 파일전송(브라우저→sup→worker 2-hop)으로 오해 → 사용자 정정: "DB에 저장된 쉘/앱을 골라 worker에서 실행" → 다시 "디바이스의 앱들" + 유저별 소유/가끔 공유 → **폴더(worker 귀속, 접속상태)+실행파일(셸 본문) 트리 구조**로 수렴
- **이름**: 단일 엔티티 작명이 안 잡힌 이유 = 실은 folder+script 두 종류 → 통합 `Node`(kind 구분)로 귀결
- **단일 테이블 결정 계기**: 사용자가 "둘 분리하면 frontend 나열 순서 불편하지 않냐" 지적 → 파일시스템 inode 패턴(단일 `nodes`+kind)이 정답. 한 폴더 자식 한 줄 조회·단일 정렬좌표·자유 재정렬
- **position**: 단일 정수 → `position_x/y` 실수(NUMERIC(5,2)) 0~100. PC 캔버스 % 배치 + 끼워넣기 위해 정수 아닌 실수
- **공유**: 노드 단위 sharing 테이블을 만들려다, 사용자가 "정리·공유용 Label 모듈(트리)"을 추후 추가 의향 → 공유를 라벨로 이전, 현 nodes 스키마에선 sharing 제외(YAGNI)

#### 결정 스냅샷
- `nodes`(단일 테이블, folder/script) / worker 귀속=폴더 device_key 상속 / 스크립트=인라인 전송 / Label(추후, 직교 M:N 트리)=분류+공유. 스키마·CHECK 제약은 `REF-node-label.md` 참조
- ※ 이후 worker 메모리 전용 확정으로 `device_id→agents` FK는 `device_key TEXT`(InstanceKey 문자열, FK 아님)로 변경
