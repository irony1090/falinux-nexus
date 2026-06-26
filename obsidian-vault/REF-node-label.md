# REF — Node / Label 모듈 설계 (frontend 카탈로그)

> frontend에서 worker의 "앱/셸"을 나열·실행하는 **supervisor PG 영속 카탈로그**. process 모듈에 "무엇을 실행할지"를 먹이는 계층(**카탈로그=무엇 / process 모듈=어떻게**).
> 2026-06-26 설계 확정 — **구현 착수**. DB+CRUD는 process 모듈과 독립 선행 가능, 실제 "실행"은 MsgExec 배선 필요.
> 논의 경로는 `history/node-label.md`. 현재 진행은 `CURRENT.md`.
> ★용어 합의: 사용자가 "쉘/앱"이라 부른 것의 실체 = **디바이스(worker) 위 트리에 정리된 실행 정의**. 단일 엔티티가 아니라 folder+script 두 종류가 한 트리에 묶임.

## 통합 노드 트리 (`nodes` 단일 테이블)
- **두 종류(folder/script)를 한 트리에**: 폴더·스크립트를 frontend 한 목록에 섞어 나열+자유 재정렬해야 함 → 테이블 2개로 쪼개면 UNION+정렬좌표 분산으로 불편 → **단일 `nodes` + `kind` 구분**(파일시스템 inode 패턴). 한 폴더 자식 = `WHERE parent_id=$1 ORDER BY position_x,position_y` 한 줄
- **folder**=컨테이너(트리 노드, **worker 귀속 + 접속상태 창** 역할) / **script**=리프(셸 본문 텍스트 보관, 실행 시 worker로 인라인 전송 후 PTY)
- **worker 귀속 = 폴더 단위**(`device_key TEXT`, **FK 아님**). 스크립트 실행 대상 worker = 자기 폴더에서 위로 올라가며 **가장 가까운 device_key 상속**. device_key 없는(NULL) 폴더 = 순수 정리용
  - ★**worker는 메모리 전용**(DB `agents` 테이블 폐기, 2026-06-26) → `device_key`는 worker의 `InstanceKey`(`메인키#서브키`) **문자열**일 뿐 FK 무결성 없음. 폴더↔worker 결속 검증 = "그 키가 **런타임 메모리 레지스트리에 살아있나**"로 함(끊긴 worker = 키가 메모리에 없음). DB는 키 문자열만 보관
- **트리** = `parent_id` 자기참조(인접리스트), 루트=NULL
- **position_x/y = `NUMERIC(5,2)` 실수 0~100**(정수면 한 부모에 101칸뿐 + 두 노드 사이 끼워넣기 불가라 실수 채택). PC=캔버스 **% 절대배치** / 그외=`ORDER BY x,y` 리스트·그리드. **부모별 로컬좌표**, x·y 동률은 name/id 타이브레이크
- **전송=인라인**: 스크립트가 작은 텍스트라 청크 전송 모듈 불필요(실행 메시지에 `content` 실어 보냄). 큰 바이너리 생기면 그때 transfer 모듈 붙임(YAGNI). ※"전송 후 실행"은 실행 동작에 흡수
- **script 동작 두 가지 = process 모듈 실행 타입**(2026-06-26): **실행=`EXEC`**(content를 스크립트로 PTY 실행) / **편집=`EDIT`**(worker에서 실제 `vi` PTY로 띄워 편집→종료 시 read-back→`nodes.content` UPDATE). 둘 다 같은 PTY 엔진, type만 다름 → 상세 `REF-process.md` "실행 타입 EXEC vs EDIT"
- **CHECK 정합성**: `content`=script 전용(folder면 NULL), `device_key`=folder 전용(script면 NULL), x/y 0~100

```
nodes( id, owner_user_id→users, parent_id→nodes NULL,
       kind 'folder'|'script', name,
       position_x NUMERIC(5,2), position_y NUMERIC(5,2),  -- 0~100 부모별 로컬
       device_key TEXT NULL,     -- folder 전용, worker InstanceKey 문자열(FK 아님)
       content    TEXT NULL,     -- script 전용
       created_at, updated_at,
       CHECK(kind='script' OR content    IS NULL),
       CHECK(kind='folder' OR device_key IS NULL),
       CHECK(position_x BETWEEN 0 AND 100 AND position_y BETWEEN 0 AND 100) )
```
- `users`=구현 완료(2026-06-26, 구 PortBridge 이식) / worker=메모리 전용(DB 테이블 없음)

## Label 모듈 (추후) — 직교하는 두 번째 트리
- **node 트리 = 구조/배치(부모 1개) vs label 트리 = 횡단 분류 + 공유(M:N)**. 서로 직교
- `labels(id, owner_user_id, parent_id→labels NULL, name)` 중첩 카테고리(**Gmail 라벨 패턴**) + `node_labels(node_id, label_id)` M:N
- **공유를 라벨이 떠안음** → 노드 단위 sharing 테이블 **지금 안 만듦**. "노드를 라벨에 담고 → 라벨 공유 → 그 노드 전부 접근". 라벨=순수 가산 join이라 나중에 무통증 추가(YAGNI). 소유는 노드별 `owner_user_id`, 공유는 라벨로
