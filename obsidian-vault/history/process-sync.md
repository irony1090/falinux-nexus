# HISTORY — process 동기화 범위 · 탭 id · 공유

> 설계·결정 → `REF-process-sync.md` / 1단계 구현·확인 이력 → `history/process-sync-impl.md` / 출발점(터미널 임베드·sid 단위 크기 우선권) → `history/node-ui-terminal.md` / 현재 진행 → `CURRENT.md`.

---

## 2026-10-06 — presence(접속 상태) 조회 시점 논의 (코드 없음)
- 사용자 질문: process별(폴더 포함?)로 보는 계정·세션의 접속 상태를 볼 수 있나, 공유에 필요할 텐데 언제 할지
- 코드 확인: `tabRegistry`(탭 id·sid·계정·연결 시각)·`released`·`sizeOwners`는 메모리에 있으나 조회 API 없음, User-Agent 없음, process별 보는 탭은 S(구독 역할 분리)로 계정 탭 전부
- 합의: 5 공유(P·R·T) 구조 합의 때 함께, 5의 첫 작업 단위로 → `REF-process-sync.md` "presence"
- 같이 답한 것: A 브라우저 종료 → 60초(`tabGrace`) 뒤 B로 크기 소유권 넘김 → B가 `watch(owner)`로 fit·resize. 비정상 끊김은 ping/pong 감지까지 최대 약 25초 추가

## 2026-10-01(3) — 1단계 작업 단위 제시 (코드 없음)
- 사이에 Hub 막힘 → ping/pong 해결·커밋 `e395433`(→ `history/realtime.md`). 브라우저 끊김(기기 종료) 시 P는 계속 실행·출력은 구독자 없으면 버려짐·타일은 ⑪ 전까지 안 보임 설명.
- 사용자 "다음 순서 플랜 보여줘" → 1-a(DB 정리)~1-g(프론트) 작업 단위 제시(`REF-process-sync-impl.md` "1단계 작업 단위"), 착수 승인 대기 중 재부팅.

## 2026-10-01(2) — Rebind 뒤 출력 끊김 실행 확인 (코드 수정 없음)
- 사용자 "실행해서 확인까지해" → supervisor·worker를 scratchpad 빌드로 직접 띄우고 worker를 socat 프록시 경유로 붙여 프록시만 끊는 방식으로 재현(방법 → `REF-process-reconnect.md` "재현 방법").
- 확인: 기존 소켓은 재접속 후 DATA 0건, 같은 세션 새 소켓은 즉시 수신 → 원인 = `cleanupProcessTopic`이 끊김 때 Hub 구독을 지우고 Rebind가 다시 붙이지 않음(추정이 맞았음).
- 추가 발견: 끊김 순간 브라우저에 `STATUS FAILED 502`가 감(process는 살아 있음). 둘 다 `REF-process-sync-impl.md` "발견".
- 사용자 "추천대로" → 두 문제 모두 1단계에 포함 확정. 구조 W(재접속 재구독)·X(끊김 = PENDING 발행 + `Detach`)·Y(재접속 소실 = FAILED 직접 발행) 정리, 코드 미착수.
- 정리: 테스트 node 20~22 삭제(API), 띄운 supervisor·worker·socat 종료. `processes`에 FAILED 행 3개(node 20~22, 이미 삭제된 node 참조)는 이력으로 남음.

## 2026-10-01 — 설계 개정 논의 (코드 없음)
- 발단: 사용자 질문 "같은 계정 크롬 탭 2개에서 한쪽에서 process를 켜면 다른 쪽도 켜지나?" → 서버는 같은 sid의 모든 소켓에 보내지만 화면은 실행한 탭만(타일 트리가 탭 메모리). 다른 브라우저·기기는 sid가 달라 아예 안 받음.
- **실수 인정**: 크기 우선권(M)의 "세션"을 sid로 해석해 같은 브라우저 탭 사이 우선권이 없었음. 사용자 "아까 우선권 문제에서 물어본 내용 아니었어? 잘 처리했을 줄 알았는데". 해석이 갈리는 지점은 확인했어야 함.
- 탭 식별: 처음엔 프론트 생성 + `BroadcastChannel` 복제 감지를 제안 → 사용자 "백엔드에서 생성해서 넘겨줄 수 있나?" → 서버 발급 + `sessionStorage` 보관으로 변경(탭 복제·위조를 서버가 해결).
- 사용자 가정 기능(공유: 특정 process·폴더를 공유하면 그 아래 자식 실행도 동기화) 검토 → 결정 O(계정 동기화)·P(공유 권한)·Q(크기 우선권 단위 = 탭 id)·R(공유 대상 단위 = 계정) 사용자 지정.
- 사용자 "여러 항목을 같이 처리하려니까 힘들다" → 전역 규칙 신설(`~/.claude/CLAUDE.md` "여러 질문을 한꺼번에 받았을 때"): 연관 없는 질문은 하나씩, 미루는 것은 밝히고 vault 기록.
- S(구독 역할 분리) 확정: 인가 계정 / 라우팅 탭 / 크기 순서 탭 id → `process_subscribers` 제거. 사용자 처음 제안은 "구독 범위를 탭 id까지"였고, DB엔 계정만 두고 탭은 메모리로 나누는 안에 동의.
- T(공유 자식 상속) = 볼 때 확인(추천대로).
- 진행 순서 1~5 확정(탭 id + S → 확인·커밋 → ⑪ → O 화면 → 공유).
- 사용자 질문 "sessionStorage가 1분 유지 후 넘기는 거지?" → 아님: `sessionStorage`는 id 보관만(만료 없음), 60초는 서버가 셈.
- 1단계 구조 제시 → U(목록 API) = `GET /processes`, V(탭 id 전달) = `TAB:ID` 이벤트(둘 다 추천대로). 사용자 재부팅 전 vault 정리 요청으로 세션 마감, 코드 미착수.
