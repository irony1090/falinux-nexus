# CURRENT

## 현재 날짜
2026-10-02

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시 (2026-10-01 세 번째 세션 종료 정리 — 사용자 재부팅)
- **1단계(탭 id + S(구독 역할 분리) + W·X·Y) 코드 + 2단계 실행 확인 완료(2026-10-01)** — 전 항목 통과, 커밋 `78bb8b2`. 결과 → `REF-process-sync-impl.md` "2단계 실행 확인", 이력 → `history/process-sync-impl.md`
- **sid 충돌 수정·확인함**(`signIn`에 `_nonce`, `78bb8b2`에 포함) — 재기동 후 연속 로그인 20쌍 쿠키 동일 0 → `REF-process-sync-impl.md` "발견"
- 전체 순서: 1 탭 id + S(구독 역할 분리)(완료) → 2 재기동·확인·커밋(완료: `78bb8b2`·`e5200ce`·`78ee99c`, push 안 함) → **2.5 터미널 입력(다음)** → 3 ⑪(타일 트리 서버 저장) → 4 O(계정 동기화) 화면 → **4.5 SNAPSHOT(화면복원 ring buffer)** → 5 공유(P·R·T)
- **2026-10-02 진행: 입력-a(서버) 코드 작성 완료**(`router/input.go`·`PROCESS:INPUT`, 서버 단독 확인 4항목 통과, 미커밋 → `REF-process-input.md` "입력-a 서버 단독 확인"). **입력-b(프론트) 코드 작성**(`processTerm.store.ts`) + I5(Ctrl+V 붙여넣기) = 붙여넣기 확정 + **입력-c 브라우저 7항목 통과**(미커밋). 남은 확인 = PENDING 중 입력 버림(socat). **다음 = 커밋 여부 확인 → 3 ⑪(타일 트리 서버 저장) 또는 5-1 ⑮(모바일 가상 키보드)**
- **2.5 터미널 입력**: 결정 I1(전달 경로)=소켓 / I2(입력 권한)=같은 계정 누구나 / I3(PENDING 중 입력)=버림 / I4(Ctrl+C 복사)=선택 있으면 복사·없으면 `0x03` / I5(Ctrl+V 붙여넣기)=붙여넣기 확정, 작업 단위 입력-a(서버)·입력-b(프론트)·입력-c(확인) → `REF-process-input.md`. (입력-a 착수 승인 받음, 2026-10-02)
- **3단계 ⑪(타일 트리 서버 저장) 구조·결정 확정(2026-10-01, 코드 미착수 — 사용자 "vault부터 정리, 코드 수정은 아직")**: ⑪-1(저장 형태)=계정당 JSON 문서 / ⑪-2(충돌 처리)=version 비교 + 재적용 / ⑪-3(새로고침 뒤 터미널 화면)=빈 화면 감수(SNAPSHOT 추후 필수) / ⑪-4(타일 없는 실행 중 process)=서버가 exec 때 터미널 타일 같이 넣기. 작업 단위 3-a(DB)·3-b(API)·3-b'(exec 연동)·3-c(스토어)·3-d(터미널 복원)·3-e(확인) → `REF-node-ui-save.md`. **착수 승인 대기**
- 커밋 제외 유지(미커밋): `apps/core/cmd/irony/`(사용자 스크래치), `provideAppLayout.vue`(빈 줄 하나)
- 재부팅으로 사라지는 것: 세션 scratchpad의 playwright-core·테스트 스크립트(`lib.mjs`·`part1*.mjs`·`part2.mjs`·`part3.mjs`·`ui.mjs`)·테스트 worker(`w2/`). 다시 필요하면 `REF-process-sync-impl.md` "2단계 실행 확인"·`REF-process-reconnect.md` "재현 방법" 보고 재작성
- 커밋 상태: 터미널 입력(입력-a·입력-b) + vault 커밋(2026-10-02, push 안 함)
- 타일 터미널(6) 실제 실행은 이번 세션에 동작 확인됨(사용자 탭 + 헤드리스 `tick` 출력). 사용자 탭에서 처음 한동안 출력이 안 보이다 뜬 일 1회 — 원인 미상, 재발 시 탭 콘솔 확인
- 연동 1차 1~4는 커밋됨(`ba5b857` feat / `8223f36` docs, push 안 함)
- 결정 위치: 6(타일 임베드) J~N → `REF-node-ui-terminal.md` / 동기화·탭 id·공유 O~V → `REF-process-sync.md` / 1단계 구조·W·X·Y·확인 → `REF-process-sync-impl.md` / ping/pong → `REF-realtime.md` "발견"
- 4(스크립트 편집) 저장 버튼 = 저장하고 계속 편집 — 코드 반영·커밋 `e5200ce`(2026-10-01) (→ `REF-node-ui-link.md` "4(스크립트 편집) 구조")
- 환경: supervisor·worker는 사용자가 띄움(`docker start postgres15` → supervisor → worker, dev 서버 3000). 테스트 로그인 = `pages/Login.vue` 기본값 계정. 헤드리스 = `~/.cache/ms-playwright/chromium-1234/chrome-linux64/chrome` + `playwright-core`(세션 scratchpad에 `npm i` — 재부팅하면 사라짐). worker 연결만 끊는 재현 = `REF-process-reconnect.md` "재현 방법", 멈춘 구독자 재현 = `REF-realtime.md` "발견"(테스트 스크립트들은 scratchpad에만 있어 재부팅 후 사라짐)
- 스크립트는 직접 실행이라 `#!/bin/sh` 등 shebang 필수(없으면 `exec format error`)
- DB 노드 현황: 폴더 id 1(`HTOP_TEST_SH_MODI`, `device_key=irony-MAC-ADDress1`) 안에 스크립트 2~4 / 폴더 `test`(14) > `ttt1`(15) > 스크립트 23(사용자 테스트용). 스크립트 2~4 `execProcess` 실제 실행 미확인
- `history/transfer.md`가 11.5k자로 분할 기준 초과(손대지 않아 그대로) — 다음에 그 파일을 고칠 때 분할

## 🎯 다음 작업: node 카탈로그 타일 UI — 구현 진행 중 (1~5단계 + 연동 1차(7-0) 1~5 + 6 타일 임베드 + 동기화 1단계 완료·커밋 `78bb8b2` → 다음 = 터미널 입력, 그 뒤 ⑪(타일 트리 서버 저장))

2026-08-10 termspace 영감 **2D 타일 그리드**로 전환 → 2026-09-29 아티팩트 시안으로 규칙을 다듬고, 같은 날 **"큰 틀은 잡혔다, 보여준 아티팩트 UI를 토대로 만들 것"**(사용자). **역할 분담(2026-09-29 개정)**: 이 작업은 코드 대부분을 Claude가 작성하고, 로직 구조는 사용자와 먼저 맞춘 뒤 쓴다(구조 합의 → 코드 순서, 합의 전 코드 착수 금지). 상세 → `REF-node-ui.md`(컨셉) / `REF-node-ui-layout.md`(저장·순서·채우기·오버플로) / `REF-node-ui-projection.md`(화면 크기별 표시) / `REF-node-ui-overview.md`(전체보기·버튼·스크롤바·Grid 탭).


**구현 기준 시안 = "타일링 시안 전체보기" (최신 v10, 가상 키보드 비교 포함)**: https://claude.ai/artifact/WPLqzcUKmTNwJp58Fi3zuM (`user/links.md`). 확정 규칙이 전부 들어 있는 유일한 시안(원본 Tz9D…·비교 시안 2개는 기록용). HTML 원본은 scratchpad에만 있었으므로 **Artifact `read`로 받아야 함** — 포팅 대상 = `TileOrder`/`TileSplit` 블록(탭 손코딩 로직은 제외, VPagination 사용).

**확정 규칙 요약 (2026-09-29)**
| 항목 | 규칙 |
|---|---|
| 저장 | 계정 단위 **여는 관계 트리**(tile: type·node·parent=연 폴더 타일·kids·size). Grid/axis는 저장 안 함 |
| 순서 | 폴더 → 직접 실행한 프로세스 → 넘겨받은 프로세스(`adopted`, 닫힌 폴더에서) → 연 하위 폴더(재귀) |
| 닫기 | 폴더 닫기 = 하위 폴더는 그 자리로, 터미널은 `adopted`로 부모 소속. 실행 중 터미널은 닫기 비활성(kill 먼저). 루트 불가 |
| 크기 | 타일별 가로·세로 0.5/1.0, 헤더 크기 버튼. **새 타일 = 누른 화면의 최소 단위**(PC 0.5×0.5 / 폰 세로 1.0×0.5) |
| 채우기 | 순서대로 2×2칸, 칼럼 우선(모든 기기). 안 들어가면 다음 Grid(이전 Grid로 안 돌아감). Band 방향 = 가로 1.0 타일 있으면 row, 아니면 column |
| 화면 크기 | Grid 영역 가로 ≥ 728 / 세로 ≥ 560이면 그 축 2칸. 1칸 축은 저장된 0.5를 1.0으로 보고 다시 채움. 페이지네이션(화면 분할) 없음 |
| 오버플로 | 옆 Grid 추가 + 가로 스크롤, 칼럼(Grid 가로 절반) 단위 스냅 |
| 빈 칼럼 | 오른쪽 칼럼이 통째로 빈 Grid는 **위치와 상관없이 1칼럼 폭**(`pack`이 `cols` 반환). 세로 빈칸은 유지 |
| 전체보기 | 버튼 → 현재 배치 그대로 축소, **긴 축 한 줄 + 짧은 축 가운데**, 자유 스크롤(스냅 끔), 타일 누르면 그 Grid로 이동, Esc = 원위치. 축소 비율 = 사용자 조절(15~60%, 기기별 localStorage, 로그인 풀리면 삭제), 거르지 않음 |
| 전체보기 버튼 | 화면 고정, Grid 하단 정보 줄 높이의 작은 강조색 버튼을 그 줄 오른쪽 끝에. 가로 2칸 = 아이콘+글자 / 1칸 = 아이콘만 |
| 스크롤바 | `overflow: auto` + 버튼이 현재 스크롤바 두께(`--sb-h/-w`)만큼 비킴. `capacity` 판정엔 스크롤바 반영 안 함(진동 방지) |
| Grid 탭 | Vuetify `VPagination`(폭 자동 맞춤·`…`). 현재 = 가장 많이 보이는 Grid(동점이면 뒤쪽). 칼럼 위치는 범위 표시(`칼럼 6-7/7`) |
| 폴더 열기 | 이름 클릭 = 타일 안 이동 / "새 타일" 버튼 = 새 폴더 타일 |

**구현 순서 (코드 구조·파일 매핑 → `REF-node-ui-impl.md` 데이터 층 / `REF-node-ui-render.md` 컴포넌트·헤더 내비)** — 층 구조 ①~⑤ 합의(2026-09-29), 첫 범위 = 1~3(타일 내용 더미)
1~5. **완료(2026-09-29~30)**: 유틸 포팅 / 트리 상태 / 렌더 최소판 + 칼럼 스냅 / 헤더 내비 / 4-1 타일 따라가기(`reveal`) + ⑧ 커서 채우기(`5b87578`) / 5 전체보기 + 고정 버튼 + 축소 비율 저장(`0855041`). 상세 → `REF-node-ui-impl.md`·`REF-node-ui-render.md`, 이력 → `history/node-ui-impl.md`·`history/node-ui-render.md`. ⑭(전체보기 버튼 최종 위치)는 현재안으로 확정(2026-10-01)
5-1. ⑮(모바일 가상 키보드) — **방향 확정 A+B(2026-10-01)**: A(타일 UI 유지, 가려지면 타일 영역 이동) + B(키보드 위 특수 키 줄). iOS 실기기 확인은 나중(iPhone 없음). 순서 = 2.5 터미널 입력 뒤 → `REF-node-ui-overview.md` "모바일 가상 키보드"
6. **(완료·커밋 `78bb8b2`)** `ProcessDialog` 타일 임베드 → `REF-node-ui-terminal.md`. 전체보기 전환은 타일을 재마운트하지 않으므로 xterm을 그대로 얹을 수 있음
6-1. ⑪(타일 트리 서버 저장) — **구조·결정 확정(2026-10-01), 코드 미착수** → `REF-node-ui-save.md`
7-0. **연동 1차 (2026-09-30)** — 항목·구조·결정 → `REF-node-ui-link.md`, 이력 → `history/node-ui-link.md`
   - **완료(커밋 `ba5b857`)**: 1(`GET /workers` + exec 대상 장비 검증) / 2(폴더 타일 목록·경로·이름) / 3(node 관리 UI: 생성·이름 변경·장비 지정·접속 상태·삭제) / 4(스크립트 편집 창)
   - 5(실행)도 6과 함께 구현·커밋(`78bb8b2`) — `tileDummy.ts` 삭제, 실행·kill·상태 = 실제 API·소켓 (→ `REF-node-ui-terminal.md` L)
7. **더미 → 실제 연동 나머지**: 실행(위 5) / kill·상태 = process API·소켓(`PROCESS:UPDATE`/`STATUS`) / 타일 트리 = 서버 저장본(⑪) 로드. EDIT(worker `vi`)는 6 + input 배선 뒤

**미리보기 방법**: `apps/frontend`에서 `npm run dev` → `/`. 루트 폴더 타일 1개로 시작, **supervisor + 로그인 필요**(목록이 실제 node API).

**색 규칙(2026-09-29 확정)**: 데모부터 이후 레이아웃 전부 Vuetify 테마 색 이름 의존, hex 금지 → `REF-frontend.md` "색 규칙". 기존 위반 3곳 정리 완료(터미널 = 커스텀 `terminal` 테마 색).

**기타 유지 사항**: `position_x/y` 사용 중단(2026-08-10, DB 컬럼 유지) / 트리 드래그 시 device_key 재상속 이슈만 유효(→ `REF-node-label.md`) / 외부 레이아웃 라이브러리 안 씀(Vuetify 기본 컴포넌트는 사용).

**열린 질문 (번호 유지, 해소된 ⑥⑦⑧⑨⑩⑪⑫⑬⑭는 REF/history에 — ⑪은 2026-10-01 `REF-node-ui-save.md`로 확정)**
1. 분할 UX — 버튼메뉴 vs 드래그드롭(VSCode류). 시안은 실행/새 타일 버튼
2. `ProcessDialog` 다중 인스턴스 리팩터 구체 범위
3. 라우팅 스킴 재검토(옛 `/nodes/:parentId?`가 이 모델에 맞는지)
4. `NODE:<parentId>` 동적 구독/해지 — process는 REST로 확정했으나 node도 같은 길로 갈지 미정(이월)
5. **네이밍**: 전체 개념 "그리드" vs "타일링", `Band`(가칭)
15. **모바일 가상 키보드**: 방향 확정 A+B(2026-10-01) — ①(판정) 레이아웃 뷰포트 / ②(입력 모드) 기각 / ③(입력 중 터미널 크기) 유지 / ④(보조 키 줄) Esc·Tab·Ctrl·Alt·방향키 등 / ⑤(iOS 문서 스크롤) 대응은 반영, **iPhone 실기기 확인만 남음** → `REF-node-ui-overview.md` "모바일 가상 키보드". **주의**: `appWindown.store.ts`의 `size.inner`(visualViewport)에 판정을 연결하면 안 됨

**커밋 (2026-09-30)**: `5b87578` feat(node) ⑧ 커서 + reveal / `e55fb98` docs(vault) / `0855041` feat(node) 전체보기 + 고정 버튼 + 축소 비율 저장 / docs(vault) 정리. 제외(미커밋 유지): `apps/core/cmd/irony/`(사용자 스크래치), `provideAppLayout.vue`(빈 줄 하나).

---

## 🎯 병행 작업: DraggableSession / useDragGhost 리팩터 설계 (구현 미착수)

Node UI Phase 2/3(드래그)에 앞서, 구 test-jig의 드래그 코드(`DraggableListener`+`DraggableLogic`+ghost 손코딩)가 "만든 본인도 재활용 못 할 정도로 복잡"하다는 판단으로 먼저 재설계했다. **설계만 확정, 코드는 다음 세션 작성 예정.** 상세 → `REF-util-drag.md`.

- `DragListener`(순수 포인터 추적)는 유지 / `StartPolicy` 함수 2종(threshold-drag, long-press)이 구 `DraggableLogic` 클래스계층+synthetic event 재발사를 대체 / `DraggableSession`이 `EventInterface` 상속(start/move/end/progress) — 이웃 재정렬 로직은 드롭
- `useDragGhost` 컴포저블 — Icon/Window에 중복돼있던 ghost 추종 watch 블록 흡수
- `GhostArea.vue`+`ghost.store.ts` 포팅 확정(test-jig에서 그대로) — `feature/layout`에 `AppDialog`와 동일한 자리, `provideAppLayout.vue`에서 같이 kick 예정
- 구 `common/listener/draggable.listener.ts`+`draggable/draggableLogic.ts`는 nexus 어디서도 미사용 확인됨 → 교체(삭제 후 신규 작성) 예정

---

## 진행 중 / 잔여

### Node 모듈 — 남은 단계 (DB+CRUD+핸들러+PatchNode 커밋 `9c9d22e`, HTTP e2e 미검증)
> 설계 `REF-node-label.md` / 이력 `history/node-label.md`
4. **worker_instances roster**: `00004_worker_instances.sql`(PK main_key,sub_key + last_seen) → register에 DB upsert → `ListInstances(main_key)` → 활성/비활성 = roster ∩/− 레지스트리. ※subkey 위조검증 보류
5. **HTTP e2e**: 가입→세션→node CRUD 왕복 스모크(서버 띄우면 00003 실DB 적용)
6. **label 모듈**: `00005_labels.sql`(labels 자기참조 + node_labels M:N) → query → router
- 핸들러 책임 미적용: parentId owner 일치 검증 / 자기 자손으로 Move 사이클 방지

### process 도메인 배선 — 남은 건 input뿐
> 설계·이력 → `REF-process-wiring.md`/`-trigger.md`/`-reconnect.md`/`-subscription.md`(history 동일 접두)

완료: supervisor+worker 실행부 전체(exec/kill/resize/재접속/구독) 배선·e2e 검증 끝(2026-07-01~07-22, kill 실사용 테스트로 발견한 상태동기화 버그 3건 포함 → `history/process-trigger.md`).

**남은 것**
1. **input(키입력)**: 입력-a(서버)·입력-b(프론트) 작성 + 입력-c 확인 통과(2026-10-02, 미커밋). PENDING 중 입력 버림 확인만 남음 → `REF-process-input.md`
2. **화면복원**: supervisor-side ring buffer(SNAPSHOT) — 설계만 확정, 코드 미착수(→ `REF-process-snapshot.md`). **추후 반드시 구현(사용자 2026-10-01)** — 새로고침·다른 탭 터미널 빈 화면의 유일한 해결책. ⑪(타일 트리 서버 저장)에선 빼고 감수(⑪-3).
3. EXEC content→실행 세부정책(직접실행 vs `sh -c`) 미정.

**결정 필요**: 끊긴 창 kill 거절 vs 큐잉(입력은 I3(PENDING 중 입력)=버림으로 확정) / 공유 kill 인가 / kill 에스컬레이션.
**정리 잔여(구)**: `register.go` 주석 SendBuffer 테스트. worker `baseDir` 필드·`instanceKey()`가 resolveDest 재설계로 dead code화(정리 여부 판단).

### 프론트 user/login — WIP (커밋 `3a8e92e` 동반, → `REF-frontend.md`)
- `feature/user`(`auth.store.ts` + `api/user.api.ts`), `pages/Login.vue` + `/login` 라우트, `common/api`(api/query util), `feature/layout` 전역 다이얼로그(`AppDialog.vue` + `appDialog.store`). 실서버 연동·가드 마무리 남음.

---

## 미해결 이슈 (이월)
- **Hub 막힘** → ping/pong으로 **해결·확인·커밋 `e395433`**(2026-10-01, push 안 함: `internal/transport/keepalive.go` 신규 + `subscribe.go`·`supervisorRouter.go`·`workerRouter.go` 한 줄씩). 남은 한계 = 막힘이 최대 약 25초. 사용자 supervisor·worker는 옛 빌드라 재기동해야 반영 → `REF-realtime.md` "발견"
- **`ProcessDialog` `PROCESS:UPDATE`/`STATUS` 리스너**: 아직 `console.log` 스텁(`patchStatus` 진입점은 이미 있어 연결만 하면 됨).
- **파일 전송**: 구현 완료 / e2e 미검증. 잔여: e2e 스모크 / abort sentinel
- **서브키 충돌/위조**: key↔subkey 결속 검증 미구현(node roster에서 닫을지 보류)
- **supervisor 영속성**: registry 메모리 → PG 미착수

## 잔여 (틈날 때)
- `api.util.ts` `throwCatch`가 HTTP 상태 코드를 버림 → 서버 미기동도 세션 만료처럼 `auth` null → 사용자 설정(`userPrefs`) 삭제. 공용 에러 처리 손볼 때 401만 삭제하도록(`auth.store.ts` TODO)
- `SESSION_KEY` 등 env화(현재 `"irony"` 하드코딩)
- checkSession createdAt=0(pgtype.Timestamptz gob 미직렬화) → sess.Data.ID로 DB 재조회
