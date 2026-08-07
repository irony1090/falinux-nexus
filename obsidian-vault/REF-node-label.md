# REF — Node / Label 모듈 설계 (frontend 카탈로그)

> frontend에서 worker의 "앱/셸"을 나열·실행하는 **supervisor PG 영속 카탈로그**. process 모듈에 "무엇을 실행할지"를 먹이는 계층(**카탈로그=무엇 / process 모듈=어떻게**).
> 2026-06-26 설계 확정 / 2026-06-29 정련(device_key=main_key, ord 좌표분리, worker_instances roster) — **구현 착수**. DB+CRUD는 process 모듈과 독립 선행 가능, 실제 "실행"은 MsgExec 배선 필요.
> 논의 경로는 `history/node-label.md`. 현재 진행은 `CURRENT.md`. **frontend UI 컨셉(배치도/트리/모바일)** → `REF-node-ui.md`.
> ★용어 합의: 사용자가 "쉘/앱"이라 부른 것의 실체 = **디바이스(worker) 위 트리에 정리된 실행 정의**. 단일 엔티티가 아니라 folder+script 두 종류가 한 트리에 묶임.

## 통합 노드 트리 (`nodes` 단일 테이블)
- **두 종류(folder/script)를 한 트리에**: 폴더·스크립트를 frontend 한 목록에 섞어 나열+자유 재정렬해야 함 → 테이블 2개로 쪼개면 UNION+정렬좌표 분산으로 불편 → **단일 `nodes` + `kind` 구분**(파일시스템 inode 패턴). 한 폴더 자식 = `WHERE parent_id=$1 ORDER BY position_x,position_y` 한 줄
- **folder**=컨테이너(트리 노드, **worker 귀속 + 접속상태 창** 역할) / **script**=리프(셸 본문 텍스트 보관, 실행 시 worker로 인라인 전송 후 PTY)
- **worker 귀속 = 폴더 단위**(`device_key TEXT`, **FK 아님**). 스크립트 실행 대상 = 자기 폴더에서 위로 올라가며 **가장 가까운 device_key 상속**(재귀 CTE). device_key 없는(NULL) 폴더 = 순수 정리용
  - **`device_key` = main_key만**(2026-06-29) — `메인키#서브키`(InstanceKey) 아님. 폴더는 **장비 클래스(main_key)**에 묶이고, 한 main_key 아래 여러 인스턴스(main#sub)가 동시에 살 수 있음. **실행할 subkey는 Exec 시점에 선택**(아래 "Exec 인스턴스 선택")
  - main_key는 FK 무결성 없는 문자열. 결속 살아있나 = "**그 main_key로 살아있는 인스턴스가 ≥1개인가**"(런타임 레지스트리). DB는 문자열만 보관. ※구 `agents` 제거는 영속 금지 원칙 아니라 이름·역할 불명 탓(→ MEMORY 정정)
- **트리** = `parent_id` 자기참조(인접리스트), 루트=NULL
- **좌표 = 두 관심사 분리**(2026-06-29): 과거 x/y가 캔버스배치+그리드순서를 겸해 두 레이아웃이 강결합됐음 → 분리
  - **`position_x/y NUMERIC(5,2)` 0~100 = 캔버스(PC) 절대배치 전용**, NULL 가능(미배치 → 캔버스가 ord 순서로 자동 폴백). 부모별 로컬
  - **`ord BIGINT` = 그리드/리스트 순서 전용**(1D면 충분 — 행 줄바꿈은 레이아웃 엔진 몫). **정수**(2026-06-29 NUMERIC→BIGINT: pgtype.Numeric 번거로움 제거). **인접 정수 사이 삽입은 재번호**(형제 수 작아 무방, midpoint 포기). 동률은 name/id 타이브레이크
  - 그리드 드래그=`ord`만 / 캔버스 드래그=`x/y`만 → 서로 독립. 조회: 그리드 `ORDER BY ord, name, id`(인덱스 `(parent_id, ord)`) / 캔버스 `WHERE parent_id`(정렬 불필요)
- **전송=인라인**: 스크립트가 작은 텍스트라 청크 전송 모듈 불필요(실행 메시지에 `content` 실어 보냄). 큰 바이너리 생기면 그때 transfer 모듈 붙임(YAGNI). ※"전송 후 실행"은 실행 동작에 흡수
- **script 동작 두 가지 = process 모듈 실행 타입**(2026-06-26): **실행=`EXEC`**(content를 스크립트로 PTY 실행) / **편집=`EDIT`**(worker에서 실제 `vi` PTY로 띄워 편집→종료 시 read-back→`nodes.content` UPDATE). 둘 다 같은 PTY 엔진, type만 다름 → 상세 `REF-process-exec-edit.md`
- **CHECK 정합성**: `content`=script 전용(folder면 NULL), `device_key`=folder 전용(script면 NULL), x/y(있으면) 0~100. **ord엔 상한 CHECK 없음**

```
nodes( id, owner_user_id→users, parent_id→nodes NULL,
       kind 'FOLDER'|'SCRIPT', name,
       ord        BIGINT  NOT NULL,    -- 그리드/리스트 순서(부모별 로컬, 정수. 인접 삽입은 재번호)
       position_x NUMERIC(5,2) NULL,   -- 캔버스 절대배치 전용, 0~100 부모별 로컬(미배치=NULL)
       position_y NUMERIC(5,2) NULL,
       device_key TEXT NULL,     -- folder 전용, worker main_key 문자열(FK 아님)
       content    TEXT NULL,     -- script 전용
       created_at, updated_at,
       CHECK(kind='script' OR content    IS NULL),
       CHECK(kind='folder' OR device_key IS NULL),
       CHECK(position_x IS NULL OR position_x BETWEEN 0 AND 100),
       CHECK(position_y IS NULL OR position_y BETWEEN 0 AND 100) )
-- 인덱스: (parent_id, ord) 그리드/리스트 조회 / (device_key) WHERE device_key IS NOT NULL — worker(main_key)→폴더 역조회
```
- `users`=구현 완료(2026-06-26, 구 PortBridge 이식) / worker 라우팅=메모리 레지스트리(관측용 roster는 아래 worker_instances)

## Exec 인스턴스 선택 + worker_instances roster (2026-06-29)
- **device_key=main_key라 실행 대상 인스턴스(subkey)를 Exec 시점에 선택.** 흐름:
  1. 스크립트의 device_key(main_key) 상속 해석(재귀 CTE)
  2. `ListInstances(main_key)` — 레지스트리 `r.workers.FindAll(키의 main 파트 == main_key)`로 활성 인스턴스(subkey) 목록 산출
  3. 사용자가 subkey 선택(1개뿐이면 기본) → `authKey = main#sub` → `Exec(authKey, spec)` / `SendBuffer(authKey, …)`
- **활성/비활성 표시 = roster ∩/− 레지스트리.** 레지스트리는 살아있는 연결만 들고 있어 "비활성"(끊긴 subkey)을 보이려면 명부 필요 → DB roster:
  - `worker_instances(main_key, sub_key, first_seen, last_seen, PK(main_key, sub_key))`. register 시 **upsert**(`last_seen=now()`)
  - **online은 저장 안 하고 레지스트리로 파생**(크래시 stale 방지): `active = roster ∩ registry` / `inactive = roster − registry`
  - **roster ≠ authority** — 라우팅/접근결정은 레지스트리만. roster는 UI 관측용
  - register 경로에 **DB 쓰기 배선 필요**(현재는 메모리 레지스트리만 만지는 WS 핸드셰이크)
  - **미정/보류**: row 수명·프루닝(수동삭제 or last_seen 정리) / **subkey 위조검증 보류**(roster를 authority로 승격하는 건 나중)

## Label 모듈 (추후) — 직교하는 두 번째 트리
- **node 트리 = 구조/배치(부모 1개) vs label 트리 = 횡단 분류 + 공유(M:N)**. 서로 직교
- `labels(id, owner_user_id, parent_id→labels NULL, name)` 중첩 카테고리(**Gmail 라벨 패턴**) + `node_labels(node_id, label_id)` M:N
- **공유를 라벨이 떠안음** → 노드 단위 sharing 테이블 **지금 안 만듦**. "노드를 라벨에 담고 → 라벨 공유 → 그 노드 전부 접근". 라벨=순수 가산 join이라 나중에 무통증 추가(YAGNI). 소유는 노드별 `owner_user_id`, 공유는 라벨로

## 다음 — 실시간 push (미설계, 2026-06-30~)
- REST(CRUD)는 구현 완료. 다음은 **node 변경 발생 시 supervisor가 socket으로 프론트에 push**(실시간 트리/캔버스 반영) + `apps/web` 신규.
- 정할 것: 이벤트 단위(노드 delta vs 자식목록 갱신) / 구독 범위(유저별 트리, `internal/subscribe` Hub·EVENT 평면 재사용?) / REST↔socket 정합(낙관적 갱신 vs 자기요청도 socket 수신). → 설계 시 본 REF에 절 추가.
