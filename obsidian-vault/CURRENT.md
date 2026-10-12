# CURRENT

## 현재 날짜
2026-10-12

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **2026-10-12 — 브랜치 `feat/folder-tile-uid` 커밋 → `master`에 `--no-ff` 병합 완료, 분기 지점(`d50f153`)에 태그 `checkpoint/before-folder-tile-uid`(되돌릴 땐 이 태그로)**: ③(다른 보드로 가는 터미널) 수정 + 폴더 타일 uid(FU1~FU4, FU2 개정) **사용자 직접 확인 통과**(2026-10-12) + 백엔드 `openFolder` 제거(폴더 exec = 400, 격리 확인 제거 전 23/23·후 21/21). 변경 = 서버 `tile.go`·`share.go`·`shareAccess.go`·`node.go`·`process.go`·`processApi.go`·`perm.go`·`process/{manager,entry}.go` + 프론트 `tileTree.store.ts`·`ShareDialog.vue`·`FolderTileBody.vue`·`share.api.ts` + vault. **다음 = 공유-g(확인) → 파트(5 공유) 마감**
- **2026-10-08 세션 끝 상태(정리 3) — 분기 지점**: 아래 공유-a~f + ①② 수정까지 전부 `master`에 **체크포인트 커밋 `3b80310`(코드)·`d50f153`(vault)**(사용자 요청, 파트 마감 아님 — 공유-g(확인) 남음). 이어서 할 작업(③ 수정 + 폴더 타일 uid)은 위험해 보여 **새 브랜치 `feat/folder-tile-uid`** 에서 진행(사용자 "현 지점을 분기로"). 되돌릴 땐 `master`로
- (정리 2 요약) 공유-d(실시간) 마감 + 공유-e(프론트 표시)·공유-f(presence) 완료(사용자 확인 통과), 지금은 체크포인트 커밋됨 → `REF-process-share-display.md` / `REF-process-share-owner.md` / `REF-process-share-presence.md`
- **사용자 확인 중 발견·수정(2026-10-08)**: ① 루트 타일을 공유 폴더 안으로 열었다가 공유가 풀리면 빠져나올 수 없음 → 배너에 "기본으로 돌아가기"(프론트만, **사용자 확인 통과**) → `REF-process-share-realtime-ui.md` "루트 타일 막힘" ② 터미널을 켜 둔 채 폴더를 공유하면 받는 쪽이 터미널을 열 방법이 없음 → 공유 걸 때 터미널 타일도 넣음 + `POST /processes/:uid/tile` + 스크립트 행 "내 보드에 열기"(격리 확인 통과, **사용자 확인 통과 2026-10-12**) → `REF-process-share-terminal-place.md` "공유 전에 실행한 터미널"
- **③ 다른 보드로 가는 터미널 + 폴더 타일 uid — 코드·격리 확인·사용자 확인 통과(2026-10-12, `master`에 병합)**: 위치 찾기 `placeTerminal`(uid 우선 → nodeId → 기본 루트) → `REF-process-share-terminal-place.md`
- **다음 = 공유-g(확인) → 파트(5 공유) 마감 커밋**. 공유-g에 넣을 것: 공유-b·c 회귀(합본 스크립트 재작성) / 3b(행 공유 개수)·3c(내가 공유한 것)·공유-f 시나리오 / 루트 타일을 공유 폴더 안으로 열기 → 해제 → 돌아가기 / 실행 중 터미널이 있는 폴더 공유 → 받는 쪽 터미널 타일·"내 보드에 열기" / 공유 중 폴더 이동·하위로 들어오기 — ③에서 수정·격리 확인. 시안 링크 → `user/links.md`
- **공유 표시 색 = 청록 `share`**(2026-10-08 사용자: 공유 관련만). 새 공유 표시도 `share` 색 사용. **사소한 미결**: ⑥(`실행: 이름`)이 내가 실행한 터미널에도 내 닉네임으로 나옴(시안은 "나") — 사용자 판단 / 숨기기(복원 UI "숨긴 공유 N" + `hidden_at`)는 나중 / 단축키·잘라내기(SH10) 나중
- 알아 둘 동작: 폴더 **자신**에 직접 공유 행이 있으면 그 폴더를 다른 폴더 아래로 옮겨도 접근 유지(가까운 공유 행 기준) — "범위 밖 이동"은 공유 폴더 **안의 하위 노드**를 밖으로 옮길 때만. 루트 스크립트는 장비가 없어 실행 불가. `throwCatch`가 이제 `status`(HTTP 상태)를 던짐(401만 골라 처리하는 수정은 아직 — 아래 잔여)
- 시안: "터미널 최소화 시안"(채택 기준, 메인 보드 + 머리 줄 비교 보드) https://claude.ai/artifact/Q2X7sYsDAHPwmVZmE1icgP · 이전 "터미널 다시 열기 시안"(대체됨) → user vault `user/links.md`. 고칠 땐 Artifact `read`로 `project/*.dc.html`을 받아서
- 커밋은 파트(5 공유) 마감 때 한 번(2026-10-07 사용자, 작업 단위마다 안 함). 예외: 2026-10-08 사용자가 위험 작업 전 체크포인트 커밋을 직접 요청
- 재부팅으로 사라지는 것(세션 scratchpad `/tmp/claude-1000/.../scratchpad/`): (2026-10-12 세션 `12185a69…`) `of_test.py`(openFolder 제거 전·후 + 다른 보드·폴더 uid 회귀 18, Client 클래스 포함 단독 실행) · `iso/{sup,wrk}`(DB `nexus_of_tmp`, 삭제함) / (2026-10-08 정리 4 세션, 재부팅으로 이미 사라짐) `fu_test.py`(다른 보드·폴더 uid 28개, 앞 세션 `presence_test.py`의 `Client`를 `exec`) · `fu2_test.py`(FU2 개정 11개) · `ev_test.py`(보드별 `TILES:UPDATE` 2개) · `reg/`(presence·attach를 worker `fu-dev`용으로 바꾼 복사본) · `iso/{sup,wrk}`(DB `nexus_fu_tmp`, 삭제함) / 이전 세션 것(`312eca0a…/scratchpad/pres/`의 presence·attach·ui 스크립트, 격리 프론트 `vite.iso.mts`). **다시 쓸 땐 `REF-test-env.md` "격리 실행"·"2026-10-08 확인 스크립트 요령" 보고 재작성**. 개발 환경은 영향 없음(임시 DB는 전부 삭제함, 5050·3000은 사용자 것)
- 어디에 무엇이:
  | 파일 | 내용 |
  |---|---|
  | `REF-process-share.md` | SH1(공유 단위)~SH10(노드 복사) 결정표(전부 확정), 받는 쪽·소유자 쪽 표시(확정·제안·미정), **작업 단위·순서** |
  | `REF-process-share-node.md` | SH3(받은 계정 권한) 단계·SH10(노드 복사)·클립보드 |
  | `REF-process-share-db.md` (신설) | 공유-a(DB): `00007_shares.sql`·판정 함수·판정 쿼리, SA1(단계 저장 형식)~SA4(기존 행 채우기) 확정·구현·확인 결과. 판정 쿼리 표의 "쓰는 곳" 열 = 공유-b(인가 교체) 작업 목록 |
  | `REF-process-share-authz.md` | 공유-b(인가 교체): 엔드포인트별 요구 단계, SB1(권한 부족 응답)~SB6(이동 사이클 검사), 작업 단위, **구현 메모·확인 결과(2026-10-07)** |
  | `REF-process-share-realtime.md` (신설) | 공유-d(실시간): 범위·지금 코드·SD1~SD9(전부 확정)·지킬 조건·작업 단위 실시간-0~7 |
  | `REF-process-share-reconcile.md` (신설) | 실시간-1~3: `reconcileAccess` 구조안·쿼리·경합·정할 것 |
  | `REF-process-share-realtime-ui.md` (신설) | 실시간-0(터미널 UI): SD3 확정 전문·지금 코드·구현 계획·확인할 것 |
  | `REF-process-share-api.md` (신설) | 공유-c(공유 API): API 표, SC1(출처 계산 위치)~SC7(복사 제외 개수) 확정, **구현 메모·확인 결과** |
  | `REF-workspace-impl.md` (신설) | 작업공간-a~e 구현 구조: 지금 구조·발견(받는 쪽 고아 안전망)·WS1~WS7(확정)·API·작업 단위 |
  | `REF-workspace-front.md` (신설) | 작업공간-d(프론트): 파일별 변경·구현 중 정한 것·확인할 것 |
  | `REF-workspace-server.md` (신설) | 작업공간-a(DB)·작업공간-b(API)·작업공간-c(exec·안전망·탭) 구현 메모·확인 결과·과도기 위험 |
  | `REF-workspace.md` (신설) | 작업 공간 여러 개(`/`·`/w/{hash}`, `/`에서만 생성, W1(분리 범위)~W3(이동 범위)), 터미널 머리 줄 `[실행한 타일 번호] 경로 / 이름`, 작업공간-a~e |
  | `REF-node-ui-overview.md` "active 타일 표시" | 나(약하게 흐림) + 위치 번호 단축키 확정(구현 시점 미정, 공유와 별개) |
  | user vault `user/links.md` | 오늘 시안 4개 링크·읽는 법·시안별 확정/제안 |
- **작업 순서 = C(끼워 넣기)**: 공유-a(DB) → 공유-b(인가 교체) → 공유-c(공유 API) → 작업공간-a~e → 공유-d(실시간) → 공유-e(프론트) → 공유-f(presence) → 공유-g(확인)
- **뒤에 정해도 되는 것**: 숨기기(SA3(숨기기 시점) = 공유-e(프론트) 때, `node_shares.hidden_at` 컬럼 하나) / 잘라내기(SH10 구현 때) / **보드 미리보기**(`/` 루트 타일 안 다른 보드 카드, REF-workspace "희망안") = 나중에(2026-10-07 사용자) — 덧붙이는 기능이라 막는 작업 없음. 단 "보는 보드만 출력 구독"으로 바꿀 거면 구독 로직을 건드리는 **공유-d(실시간) 착수 때 한 번 확인** / 제안 단계 표시 ①~④·⑥·B·O1~O5·단계 낮출 때 열린 터미널(공유-e 때 시안 보며) / 단축키 키 조합
- **시안 읽는 법**: 각 시안 맨 위 "결론 반영" 상자가 기준, 본문의 "제안"은 미확정. 시안과 REF가 다르면 REF 우선
- 진행 순서(동기화 계열): 1 탭 id + S(구독 역할 분리) → 2 확인·커밋 → 2.5 터미널 입력 → 3 ⑪(타일 트리 서버 저장) → 4 O(계정 동기화) 화면 → 4.5 SNAPSHOT(화면복원) → 4.6 EDIT(worker `vi` 편집) **여기까지 완료** → 5 공유(P·R·T)
- 커밋 상태: **2026-10-08 공유-a~f·작업공간-a~e·앱 이름·로고·①② 수정 코드 + vault 전부 `master`에 체크포인트 커밋함(분기 지점)**. 이후 작업은 `feat/folder-tile-uid` → 2026-10-12 `master`에 병합(태그 `checkpoint/before-folder-tile-uid` = 분기 지점). 태그·병합도 push 안 함. 2026-10-02 커밋들 + `97fb5c4` 모두 push 안 함. vault 변경(2026-10-06 공유 구조 합의분)은 커밋함(push 안 함). 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치)
- 테스트·확인 준비(환경, 헤드리스, 함정, 남은 테스트 데이터, **격리 실행 = 임시 DB + 5051**) → `REF-test-env.md`
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
5. **네이밍**: 전체 개념 "그리드" vs "타일링", `Band`(가칭) — 작업 공간 화면 이름은 **"보드"로 확정**(2026-10-07)
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
- **5 공유(P·R·T) + presence(접속 상태)** — 공유-a(DB)·공유-b(인가 교체)·공유-c(공유 API)·작업공간-a~e·실시간-0~7 완료(공유-d 마감) + 공유-e(프론트 표시)·공유-f(presence) 완료(전부 커밋 전), 다음 공유-g(확인) (`REF-process-share.md`, `REF-workspace.md`)
- EXEC content→실행 세부정책(직접실행 vs `sh -c`) 미정
- **결정 필요**: 끊긴 창 kill 거절 vs 큐잉(입력은 I3(PENDING 중 입력) = 버림 확정) / ~~공유 kill 인가~~(SH3(받은 계정 권한)으로 해소: 조작 이상) / kill 에스컬레이션
- **정리 잔여(구)**: `register.go` 주석 SendBuffer 테스트. worker `baseDir` 필드·`instanceKey()` dead code화(정리 여부 판단)

### 프론트 user/login — WIP (→ `REF-frontend.md`)
- `feature/user`(`auth.store.ts` + `api/user.api.ts`), `pages/Login.vue` + `/login` 라우트, 전역 다이얼로그. 실서버 연동·가드 마무리 남음

---

## 미해결 이슈 (이월)
- **받는 쪽 실행 터미널이 소유자 기본 보드에 고아로 붙음(2026-10-07 작업공간-e 발견)** → SD2(실행 때 터미널 타일을 넣을 계정) 확정(볼 수 있는 계정 전부), **실시간-4(exec 타일)에서 구현** → `REF-workspace-server.md` "발견"
- ~~받는 쪽 고아 안전망(2026-10-07 발견)~~ → **해소(2026-10-07)**: 작업공간-c WS3(고아 안전망 위치)로 서버 안전망(대상 = 내가 소유·실행한 것, 공유받은 쪽 GET = 안 붙임 확인) + 작업공간-d에서 프론트 `adoptOrphans` 삭제. 공유받은 터미널의 받는 쪽 배치는 공유-d가 SH2(받는 쪽 화면 위치)로 → `REF-workspace-server.md`
- **`NODE:0` 누설(2026-10-06 발견)**: 2026-10-07 인가-7(루트 토픽)에서 구독 줄 삭제·확인(커밋 전). 루트 토픽 설계는 열린 질문 4(NODE 동적 구독) → `REF-process-share-authz.md`
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
