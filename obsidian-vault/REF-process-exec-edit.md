# REF — 실행 타입: EXEC vs EDIT (2026-06-26 확정)

> 계약/설계 원칙 본체는 `REF-process.md`(여기는 그 하위 절이 커져서 분리됨, 2026-07-22). script 편집 = worker PTY vi 왕복.
> node script 편집 = frontend→supervisor→worker로 worker의 실제 `vi`($EDITOR)를 **PTY로 띄워** 편집, 종료 시 내용 회수. **PTY 엔진의 특수 사례** — 새 메커니즘 아님. 카탈로그(`REF-node-label.md`) "무엇"에 "어떻게(편집)"를 먹이는 동작.

- **단일 `MsgExec{ type, spec }` + 단일 결과채널**에 `type` 디스크리미네이터. 제어/스트림(Data·Resize·Kill·Status) 공유라 메시지 안 가르고 type만 추가(separate MsgEditExec보다 깔끔)
- **두 타입(닫힌 집합)**:
  | type | 출력 스트림 | DB sink | 산출물 | PTY |
  |------|------------|---------|--------|-----|
  | `EXEC` | 라이브 fan-out | **ON**(스크롤백 영속) | 없음(exit code) | O |
  | `EDIT` | 라이브 fan-out | **OFF**(vi 화면=버림) | **파일 read-back 내용** | O |
- **공유 엔진 / type 무지**: `IInteractive`/PTY/MsgData·Resize·Status 100% 재사용, 런타임 핸들은 type 모름. type 소비자 = ① **worker bracket 핸들러**(seed 깔까/read-back 할까) ② **supervisor bind 배선**(sink on·off / 결과 라우팅). 엔진은 한 줄도 안 갈라짐
- **EDIT 흐름**: supervisor가 노드 content + 대상 worker(device_key 상속) 해석 → `MsgExec{EDIT, content 인라인}` → worker가 tmp에 content 기록 → `vi <tmp>` PTY 실행 → 종료(STOPPED) 시 tmp 재읽기 → `MsgEditResult{UID, content}`(인라인 REQ) → tmp 정리 → supervisor `nodes.content` UPDATE
- **저장판별 = read-back & diff**: vi는 `:wq`·`:q!` 모두 exit 0 → 종료코드로 저장여부 못 가림. **종료 후 tmp 무조건 재읽기→원본과 비교→바뀌면 UPDATE/같으면 no-op**(`:w` 안 했으면 파일 불변→자연히 "저장 안 함"). `:cq`(non-zero)=명시 취소는 선택적
- **구조체**: base `ProcessSpec`(UID/Cmd/Env/…) 공유 + **type별 addendum**(EDIT만 seedContent/tmp 정책). 3개 평행 구조체 금지
- **editor 선택 = worker 책임**($EDITOR/OS 기본). 1차 **Linux worker 한정**(Windows PTY=ConPTY 별도 이슈)
- **이름**: `EXEC`/`EDIT` 둘 다 **동사로 일관**. `EDITOR`(도구명—레이어 샘) 기각, `RETURN`(모호—exit code도 return) 기각. RETURN은 EDIT의 후보명이었을 뿐 동의어
- **의존성**: process 도메인 배선 선행 필수(현재 Step1만). frontend xterm.js 필요(apps/web 미착수) — e2e는 테스트 클라 conn으로 PTY 왕복 검증 가능
- **YAGNI**: "인터랙티브 세션이 아티팩트 반환" 거창한 프레임워크 금지. **EDIT 한 동작만**. 비인터랙티브 출력 캡처(CAPTURE류)는 실수요 나올 때 별 type으로
- **프로토콜 어휘 구현됨(2026-06-26)** `protocol/messages.go`: `ExecType`(EXEC|EDIT) + `ProcessSpec.Type`(빈값=EXEC, `Kind()` 정규화) + `MsgEditResult`(worker→sup REQ) + `EditResult{UID,Content}`. 와이어 = 공통 ProcessSpec + Type 구분. 저장판별=supervisor diff. ※구 `ProcessSpec.Seed []byte`/"인라인 content" 서술 폐기 → `REF-process-wiring.md` "경로 조립" 참조.
