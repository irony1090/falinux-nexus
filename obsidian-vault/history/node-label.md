# HISTORY — Node / Label 모듈

> frontend 카탈로그(nodes 트리 + labels 분류) 설계·구현 상세 기록.
> 확정 설계는 `REF-node-label.md`, 현재 상황은 `CURRENT.md`.

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
