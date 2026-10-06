# CURRENT

## 현재 날짜
2026-10-06

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **4.6 EDIT(worker `vi` 편집) 완료(2026-10-06)**: E1(진입 위치)~E4(동시 편집) 제안대로 확정, 편집-a(프론트 진입) 커밋 `97fb5c4` + 편집-b(확인) 사용자 PC 전부 통과 → `INDEX-process-terminal.md`
- **vault 정리 완료(2026-10-06)**: process 터미널 기능(마감) → `INDEX-process-terminal.md` / node UI 연동(마감) → `INDEX-node-ui-link.md` 로 접음. CURRENT의 환경·테스트 함정 → `REF-test-env.md`, 타일 확정 규칙 요약 → `INDEX-node-ui-layout.md`
- **바로 다음 = 5 공유(P·R·T) 구조 합의** — presence(접속 상태)도 함께, 5의 첫 작업 단위로 → `REF-process-sync.md` "진행 순서"·"presence"·"필요할 필드". 구조 합의 전 코드 착수 금지
- 진행 순서(동기화 계열): 1 탭 id + S(구독 역할 분리) → 2 확인·커밋 → 2.5 터미널 입력 → 3 ⑪(타일 트리 서버 저장) → 4 O(계정 동기화) 화면 → 4.5 SNAPSHOT(화면복원) → 4.6 EDIT(worker `vi` 편집) **여기까지 완료** → 5 공유(P·R·T)
- 커밋 상태: 2026-10-02 커밋들 + `97fb5c4` 모두 push 안 함. 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치)
- 테스트·확인 준비(환경, 헤드리스, 함정, 남은 테스트 데이터) → `REF-test-env.md`
- 역할 분담(2026-09-29 개정): node 타일 UI·process 작업은 코드 대부분을 Claude가 작성하고, 로직 구조는 사용자와 먼저 맞춘 뒤 쓴다

---

## 🎯 node 카탈로그 타일 UI — 남은 것
레이아웃(→ `INDEX-node-ui-layout.md`)·연동(→ `INDEX-node-ui-link.md`) 마감. 컨셉 → `REF-node-ui.md`, 전체보기·가상 키보드 → `REF-node-ui-overview.md`.

- **미리보기**: `apps/frontend`에서 `npm run dev` → `/`. supervisor + 로그인 필요
- **색 규칙(2026-09-29)**: Vuetify 테마 색 이름 의존, hex 금지 → `REF-frontend.md` "색 규칙"
- **기타 유지 사항**: `position_x/y` 사용 중단(DB 컬럼 유지) / 트리 드래그 시 device_key 재상속 이슈(→ `REF-node-label.md`) / 외부 레이아웃 라이브러리 안 씀

**열린 질문 (번호 유지, 해소된 ②⑥~⑭는 REF/history에)**
1. 분할 UX — 버튼메뉴 vs 드래그드롭(VSCode류). 시안은 실행/새 타일 버튼
3. 라우팅 스킴 재검토(옛 `/nodes/:parentId?`가 이 모델에 맞는지)
4. `NODE:<parentId>` 동적 구독/해지 — node도 process처럼 REST로 갈지 미정. 4⑤(폴더 목록 동기화)가 여기 걸림
5. **네이밍**: 전체 개념 "그리드" vs "타일링", `Band`(가칭)
15. **모바일 가상 키보드**: A+B 확정·반영, **iPhone 실기기 확인만 남음** → `REF-node-ui-overview.md` "모바일 가상 키보드". **주의**: `appWindown.store.ts`의 `size.inner`(visualViewport)에 판정을 연결하면 안 됨

- ② `ProcessDialog` 다중 인스턴스 리팩터는 6 타일 임베드(J~N)에서 `ProcessDialog` 삭제로 해소(2026-10-06 정리 때 확인)

---

## 🎯 병행 작업: DraggableSession / useDragGhost 리팩터 (설계 확정, 구현 미착수)
구 test-jig 드래그 코드가 재활용하기 어려울 만큼 복잡해서 먼저 재설계함. 상세 → `REF-util-drag.md`.
- `DragListener` 유지 / `StartPolicy` 함수 2종(threshold-drag, long-press) / `DraggableSession`이 `EventInterface` 상속 — 이웃 재정렬 로직은 드롭
- `useDragGhost` 컴포저블 + `GhostArea.vue`·`ghost.store.ts` 포팅(`provideAppLayout.vue`에서 kick)
- 구 `common/listener/draggable.listener.ts`·`draggable/draggableLogic.ts`는 미사용 → 삭제 후 신규 작성

---

## 진행 중 / 잔여

### Node 모듈 — 남은 단계
> 설계 `REF-node-label.md` / 이력 `history/node-label.md`
4. **worker_instances roster**: `worker_instances`(PK main_key,sub_key + last_seen) → register에 DB upsert → `ListInstances(main_key)` → 활성/비활성 = roster ∩/− 레지스트리. ※subkey 위조검증 보류
6. **label 모듈**: labels 자기참조 + node_labels M:N → query → router
- 핸들러 책임 미적용: parentId owner 일치 검증 / 자기 자손으로 Move 사이클 방지
- ※ 마이그레이션 번호는 착수 때 실제 다음 번호로(`00005`는 `process_subscribers` 제거에 씀)

### process — 남은 것
> 배선 → `INDEX-process-wiring.md` / 터미널 기능 → `INDEX-process-terminal.md` / 동기화·공유 → `REF-process-sync.md`
- **5 공유(P·R·T) + presence(접속 상태)** — 다음 작업
- EXEC content→실행 세부정책(직접실행 vs `sh -c`) 미정
- **결정 필요**: 끊긴 창 kill 거절 vs 큐잉(입력은 I3(PENDING 중 입력) = 버림 확정) / 공유 kill 인가 / kill 에스컬레이션
- **정리 잔여(구)**: `register.go` 주석 SendBuffer 테스트. worker `baseDir` 필드·`instanceKey()` dead code화(정리 여부 판단)

### 프론트 user/login — WIP (→ `REF-frontend.md`)
- `feature/user`(`auth.store.ts` + `api/user.api.ts`), `pages/Login.vue` + `/login` 라우트, 전역 다이얼로그. 실서버 연동·가드 마무리 남음

---

## 미해결 이슈 (이월)
- **같은 터미널을 연 탭이 여러 개면 질의 시퀀스(`ESC[6n` 등)에 탭마다 xterm이 응답** → process가 응답을 탭 수만큼 받음. 방안 후보: 크기 소유 탭만 응답 전달 → `REF-process-snapshot-impl.md` "스냅샷-d"
- **input**: I3(PENDING 중 입력 버림) 확인만 남음(socat 재현 → `REF-process-reconnect.md` "재현 방법") → `REF-process-input.md`
- **Hub 막힘**: ping/pong(`e395433`)으로 해결, 남은 한계 = 막힘 최대 약 25초 → `REF-realtime.md` "발견"
- **파일 전송**: 구현 완료 / e2e 미검증. 잔여: e2e 스모크 / abort sentinel. `history/transfer.md`가 분할 기준 초과(17k자) — 다음에 고칠 때 분할
- **서브키 충돌/위조**: key↔subkey 결속 검증 미구현(node roster에서 닫을지 보류)
- **supervisor 영속성**: registry 메모리 → PG 미착수

## 잔여 (틈날 때)
- `api.util.ts` `throwCatch`가 HTTP 상태 코드를 버림 → 서버 미기동도 세션 만료처럼 `auth` null → 사용자 설정(`userPrefs`) 삭제. 공용 에러 처리 손볼 때 401만 삭제하도록(`auth.store.ts` TODO)
- `SESSION_KEY` 등 env화(현재 `"irony"` 하드코딩)
- checkSession createdAt=0(pgtype.Timestamptz gob 미직렬화) → sess.Data.ID로 DB 재조회
