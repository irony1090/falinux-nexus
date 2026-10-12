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
  - **`position_x/y NUMERIC(5,2)` 0~100 = 캔버스(PC) 절대배치 전용**, NULL 가능(미배치 → 캔버스가 ord 순서로 자동 폴백). 부모별 로컬. **2026-08-10부터 UI에서 미사용 확정**(캔버스 홈 폐기 → termspace 그리드로 전환, → `REF-node-ui.md`) — **컬럼·마이그레이션은 그대로 유지, 값만 안 씀**(YAGNI, 스키마 정리는 필요해지면 나중에)
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
> 프론트 UI(2026-09-30): 폴더 `device_key` 지정(직접 입력 + 접속 중 목록)·접속 상태 표시는 구현됨 → `REF-node-ui-link.md` "3(node 관리 UI)". 실행 시 인스턴스 선택 UI는 5(실행)에서.
- **device_key=main_key라 실행 대상 인스턴스(subkey)를 Exec 시점에 선택.** 흐름:
  1. 스크립트의 device_key(main_key) 상속 해석(재귀 CTE)
  2. `ListInstances(main_key)` — 레지스트리 `r.workers.FindAll(키의 main 파트 == main_key)`로 활성 인스턴스(subkey) 목록 산출
  3. 사용자가 subkey 선택(1개뿐이면 기본) → `authKey = main#sub` → `Exec(authKey, spec)` / `SendBuffer(authKey, …)`
  - **구현(2026-09-30)**: `cmd/supervisor/router/workerApi.go` — `GET /workers`(전체, 폴더 `device_key` 후보) / `GET /workers?nodeId=`(1+2를 서버에서: `ResolveDeviceKey` 재귀 CTE → 그 main_key 인스턴스만, 귀속 폴더 없으면 404). 응답 `[{mainKey, subKey, instanceKey}]` 정렬, 키 분리 = 마지막 `#`. 프론트 `feature/worker/api/worker.api.ts` `listWorkers(nodeId?)`. **1개면 선택 없이 바로 실행**(사용자 확정, UI는 연동 1차 5번에서)
  - **`execProcess` 검증 추가(2026-09-30)**: `authKey`의 main_key ≠ 스크립트 상속 `device_key` → 400(다른 장비로 잘못 실행 방지). 서브키 위조 검증은 여전히 별도 이슈
- **활성/비활성 표시 = roster ∩/− 레지스트리.** 레지스트리는 살아있는 연결만 들고 있어 "비활성"(끊긴 subkey)을 보이려면 명부 필요 → DB roster:
  - `worker_instances(main_key, sub_key, first_seen, last_seen, PK(main_key, sub_key))`. register 시 **upsert**(`last_seen=now()`)
  - **online은 저장 안 하고 레지스트리로 파생**(크래시 stale 방지): `active = roster ∩ registry` / `inactive = roster − registry`
  - **roster ≠ authority** — 라우팅/접근결정은 레지스트리만. roster는 UI 관측용
  - register 경로에 **DB 쓰기 배선 필요**(현재는 메모리 레지스트리만 만지는 WS 핸드셰이크)
  - **미정/보류**: row 수명·프루닝(수동삭제 or last_seen 정리) / **subkey 위조검증 보류**(roster를 authority로 승격하는 건 나중)

### roster 계획 (2026-10-12 **확정** — RS1(기록 시점)~RS6(위조 검증) 제안대로 + UI 스토리 승인, 코드 없음)
조사: worker는 받은 subkey를 SQLite `identity`에 저장해 재시작해도 같은 subkey로 접속(→ 명부가 쓰레기로 안 쌓임) / 등록 = `router/register.go`(`r.workers.Append`) / 끊김 = `supervisorRouter.go` `workers.Remove` / `GET /workers` = 레지스트리만, 프론트 10초 폴링(`useOnlineWorkers`) — 폴더 장비 칩·실행 인스턴스 고르기·장비 지정 다이얼로그가 씀 / 다음 마이그레이션 = `00009`

| 결정 | 제안 |
|---|---|
| RS1(기록 시점) | 등록 때 upsert(`first_seen` 유지, `last_seen = now`) + 끊길 때 `last_seen = now`. DB 실패해도 등록은 진행(로그만 — roster ≠ authority) |
| RS2(API 모양) | 기존 `GET /workers`(접속 중만, 실행 대상 고르기) 그대로 + 신규 `GET /workers/instances[?mainKey=]` = 명부 전체 `{mainKey, subKey, instanceKey, online, firstSeen, lastSeen}`(online = 레지스트리로 계산) |
| RS3(보이는 범위) | 로그인한 계정 전부(지금 `GET /workers`와 같음 — 장비는 계정 소유가 아님) |
| RS4(정리) | 끊긴 인스턴스만 수동 삭제 `DELETE /workers/instances/:mainKey/:subKey`(접속 중이면 409). 자동 정리 없음 |
| RS5(표시 위치) | ① 실행 인스턴스 고르기: 끊긴 것도 흐리게 + "마지막 접속 n분 전"(누를 수 없음) ② 장비 지정 다이얼로그: main_key별 접속/전체 수 ③ 폴더 장비 칩 title: "접속 1 · 끊김 2" |
| RS6(위조 검증) | 계속 보류(명부를 인가에 안 씀) |

- UI 스토리 시안(2026-10-12): https://claude.ai/artifact/3M5jJDyHoFPGMPxdUnKsc2 — 1 폴더 장비 칩 `1/3` + hover 인스턴스 목록 / 2 인스턴스 고르기(접속 2개 이상일 때만, 끊긴 것 흐리게·못 누름) / 3 장비 지정 다이얼로그에 명부(main_key별 접속/전체, 펼치면 인스턴스, 끊긴 것만 삭제) / 4 삭제 확인 → 목록·칩 갱신

작업 단위: 명부-1(기록) → 명부-2(API 조회·삭제) → 명부-3(프론트 표시 RS5) → 명부-4(확인) — **전부 구현·격리 확인·사용자 확인 → 마감 2026-10-12**

구현 메모(2026-10-12):
| 위치 | 내용 |
|---|---|
| `router/roster.go` `recordSeen` | 등록(`Append` 뒤)·끊김(`Remove` 뒤)에서 동기 + 2초 timeout. goroutine이면 끊김 기록이 등록보다 먼저 갈 수 있음 |
| `workerApi.go` | `GET /workers/instances` / `DELETE .../:mainKey/:subKey`(409·404·200 `{}`). 삭제도 200 — 프론트 `throwThen`이 200 아니면 던짐. `pathParam` = RawPath일 때만 디코딩(이중 디코딩 방지) |
| `worker.api.ts` `useWorkerRoster` | 명부 10초 폴링(키 `WORKER/ROSTER`), `useOnlineWorkers` 대체. 칩·다이얼로그의 접속 여부도 명부의 `online`(= 레지스트리) |
| 실행 대상 | 여전히 `GET /workers?nodeId=`(레지스트리) — 명부는 표시만. 고르기 창의 끊긴 목록 = 명부(최대 10초 전) − 방금 받은 접속 목록 |

- 한계: 등록 때 DB 기록이 실패한 인스턴스는 접속 중이어도 칩·다이얼로그에 안 보임(실행은 됨). 다음 등록 때 다시 기록됨

## Label 모듈 — **보류(2026-10-12 사용자 결정)** · 직교하는 두 번째 트리
> **보류 근거**: ① 원래 맡기려던 공유는 파트(5 공유)의 폴더·process 단위 공유로 해결(아래 "공유를 라벨이 떠안음"은 **폐기된 계획**) ② 위치가 다른 노드 모아 보기 = 보드가 이미 함 ③ 찾기가 어려워지면 검색이 먼저 ④ 나중에 더해도 테이블 2개 추가뿐(기존 스키마·코드 무변경) — 지금 만들면 붙이기·거르기 UI·공유 범위 결정만 늘어남
> **다시 볼 조건**: 같은 기준으로 여러 장비 폴더에 흩어진 스크립트를 반복해서 찾아야 하는데 검색·보드로 안 될 때

- **node 트리 = 구조/배치(부모 1개) vs label 트리 = 횡단 분류 + 공유(M:N)**. 서로 직교
- `labels(id, owner_user_id, parent_id→labels NULL, name)` 중첩 카테고리(**Gmail 라벨 패턴**) + `node_labels(node_id, label_id)` M:N
- **공유를 라벨이 떠안음** → 노드 단위 sharing 테이블 **지금 안 만듦**. "노드를 라벨에 담고 → 라벨 공유 → 그 노드 전부 접근". 라벨=순수 가산 join이라 나중에 무통증 추가(YAGNI). 소유는 노드별 `owner_user_id`, 공유는 라벨로

## 다음 — 실시간 push (미설계, 2026-06-30~)
- REST(CRUD)는 구현 완료. 다음은 **node 변경 발생 시 supervisor가 socket으로 프론트에 push**(실시간 트리/캔버스 반영) + `apps/web` 신규.
- 정할 것: 이벤트 단위(노드 delta vs 자식목록 갱신) / 구독 범위(유저별 트리, `internal/subscribe` Hub·EVENT 평면 재사용?) / REST↔socket 정합(낙관적 갱신 vs 자기요청도 socket 수신). → 설계 시 본 REF에 절 추가.
