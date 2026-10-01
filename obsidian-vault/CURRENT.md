# CURRENT

## 현재 날짜
2026-09-30

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시 (2026-09-30 세션 종료 정리)
- **미커밋 작업 있음 — 커밋할지 먼저 물을 것**(사용자 "커밋은 나중에"). 연동 1차(7-0) 1~4 전체:
  - 백엔드: `apps/core/cmd/supervisor/router/{workerApi.go(신규), processApi.go, supervisorRouter.go}`
  - 프론트: `feature/worker/`(신규) / `feature/node/{api/node.api.ts, hook/(신규), component/dialog/(신규), component/tile/{NodeRowMenu.vue(신규), FolderTileBody, TileFrame, TerminalTileBody}, dev/tileDummy.ts}` / `pages/index.vue`
  - vault 변경(`REF-node-ui-link.md`·`history/node-ui-link.md` 신규 포함)
  - 제외 유지: `apps/core/cmd/irony/`(사용자 스크래치), `provideAppLayout.vue`(빈 줄 하나)
- **바로 다음 = 6(`ProcessDialog` 타일 임베드) 구조 합의** — 연동 1차의 남은 항목 5(실행: `execProcess` → `addTerminal`, 인스턴스 선택)가 여기에 묶임. 구조 합의 → 코드 순서
- **사용자 확인 대기**: 4(스크립트 편집) 편집 창(헤드리스는 통과). 세부 하나 — 저장 버튼 = 저장 후 닫기 / Ctrl+S = 저장하고 계속 편집(한쪽으로 맞출지)
- 환경: supervisor·worker는 사용자가 띄움(`docker start postgres15` → supervisor → worker, dev 서버 3000). 테스트 로그인 = `pages/Login.vue` 기본값 계정을 그대로 써도 됨(헤드리스는 `/login`에서 제출 버튼만) → 프론트는 직접 로그인해 확인하고 보고. 헤드리스 = `~/.cache/ms-playwright/chromium-1234` + `playwright-core`(세션 scratchpad에 설치)
- DB 노드 현황: 폴더 id 1(`HTOP_TEST_SH_MODI`, `device_key=irony-MAC-ADDress1`) 안에 스크립트 2~4 / 사용자가 만든 폴더 `test`(14) > `ttt1`(15). 스크립트 2~4는 폴더로 옮긴 뒤 **`execProcess` 실제 실행 미확인**
- `history/transfer.md`가 11.5k자로 분할 기준 초과(이번 세션에 손대지 않아 그대로) — 다음에 그 파일을 고칠 때 분할

## 🎯 다음 작업: node 카탈로그 타일 UI — 구현 진행 중 (1~5단계 + 연동 1차(7-0) 1~4 완료, 다음 = 6 `ProcessDialog` 타일 임베드)

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
1~5. **완료(2026-09-29~30)**: 유틸 포팅 / 트리 상태 / 렌더 최소판 + 칼럼 스냅 / 헤더 내비 / 4-1 타일 따라가기(`reveal`) + ⑧ 커서 채우기(`5b87578`) / 5 전체보기 + 고정 버튼 + 축소 비율 저장(`0855041`). 상세 → `REF-node-ui-impl.md`·`REF-node-ui-render.md`, 이력 → `history/node-ui-impl.md`·`history/node-ui-render.md`. 남은 것 = ⑭ 전체보기 버튼 위치 사용자 검토
5-1. 모바일 가상 키보드 대응: viewport 메타 `interactive-widget=resizes-visual` + 판정은 레이아웃 뷰포트만 + 입력 모드·보조 키 줄(input 배선과 함께) — ⑮. **순서상 7 뒤로**(process input 배선이 선행, 2026-09-30 계획)
6. **(다음)** `ProcessDialog` 타일 임베드(다중 인스턴스, ②) — 착수 시 ② 범위부터 구조 합의. 전체보기 전환은 타일을 재마운트하지 않으므로 xterm을 그대로 얹을 수 있음
6-1. ⑪(타일 트리 서버 저장) 테이블/API 설계 — 7 전에 해두면 막히지 않음
7-0. **연동 1차 (2026-09-30)** — 항목·구조·결정 → `REF-node-ui-link.md`, 이력 → `history/node-ui-link.md`
   - **완료(미커밋)**: 1(`GET /workers` + exec 대상 장비 검증) / 2(폴더 타일 목록·경로·이름) / 3(node 관리 UI: 생성·이름 변경·장비 지정·접속 상태·삭제) / 4(스크립트 편집 창)
   - **남음**: 5(실행 = `execProcess` → `addTerminal`, 인스턴스 1개면 바로 실행·여러 개면 선택) — 6번과 함께. 지금 실행·kill·상태는 `tileDummy.ts` 더미
7. **더미 → 실제 연동 나머지**: 실행(위 5) / kill·상태 = process API·소켓(`PROCESS:UPDATE`/`STATUS`) / 타일 트리 = 서버 저장본(⑪) 로드. EDIT(worker `vi`)는 6 + input 배선 뒤

**미리보기 방법**: `apps/frontend`에서 `npm run dev` → `/`. 루트 폴더 타일 1개로 시작, **supervisor + 로그인 필요**(목록이 실제 node API).

**색 규칙(2026-09-29 확정)**: 데모부터 이후 레이아웃 전부 Vuetify 테마 색 이름 의존, hex 금지 → `REF-frontend.md` "색 규칙". 기존 위반 3곳 정리 완료(터미널 = 커스텀 `terminal` 테마 색).

**기타 유지 사항**: `position_x/y` 사용 중단(2026-08-10, DB 컬럼 유지) / 트리 드래그 시 device_key 재상속 이슈만 유효(→ `REF-node-label.md`) / 외부 레이아웃 라이브러리 안 씀(Vuetify 기본 컴포넌트는 사용).

**열린 질문 (번호 유지, 해소된 ⑥⑦⑧⑨⑩⑫⑬은 REF/history에)**
1. 분할 UX — 버튼메뉴 vs 드래그드롭(VSCode류). 시안은 실행/새 타일 버튼
2. `ProcessDialog` 다중 인스턴스 리팩터 구체 범위
3. 라우팅 스킴 재검토(옛 `/nodes/:parentId?`가 이 모델에 맞는지)
4. `NODE:<parentId>` 동적 구독/해지 — process는 REST로 확정했으나 node도 같은 길로 갈지 미정(이월)
5. **네이밍**: 전체 개념 "그리드" vs "타일링", `Band`(가칭)
11. **타일 트리 서버 저장**: 테이블/API 미설계(계정 단위, node·process uid 참조, 순서 = kids 순서)
14. **전체보기 버튼 최종 위치**: 현재안(정보 줄 오른쪽 끝)은 사용자가 더 검토하겠다고 함. 폰에서 22px 터치 크기도 함께
15. **모바일 가상 키보드**: ① 판정은 레이아웃 뷰포트로만 + `interactive-widget=resizes-visual` / ② 입력 모드 채택 여부 / ③ 입력 모드 터미널 크기 유지+스크롤 vs 맞춤(resize) / ④ 보조 키 줄 구성 / ⑤ iOS 문서 스크롤 — `REF-node-ui-overview.md` "모바일 가상 키보드", 시안 v10에서 비교 가능. **주의**: `appWindown.store.ts`의 `size.inner`(visualViewport)에 판정을 연결하면 안 됨

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
1. **input(키입력)**: `MsgData` 역방향, `Inter.Write` 배선. 고빈도라 REST 부적합 — 소켓 메시지 쪽 유력하나 미정.
2. **화면복원**: supervisor-side ring buffer(SNAPSHOT) — 설계만 확정, 코드 미착수(→ `REF-process-snapshot.md`). 세션→uid 원장은 구현 완료, 프론트가 엔드포인트를 부르는 UI만 없음.
3. EXEC content→실행 세부정책(직접실행 vs `sh -c`) 미정.

**결정 필요**: 끊긴 창 입력/kill 거절 vs 큐잉 / 공유 kill 인가 / kill 에스컬레이션.
**정리 잔여(구)**: `register.go` 주석 SendBuffer 테스트. worker `baseDir` 필드·`instanceKey()`가 resolveDest 재설계로 dead code화(정리 여부 판단).

### 프론트 user/login — WIP (커밋 `3a8e92e` 동반, → `REF-frontend.md`)
- `feature/user`(`auth.store.ts` + `api/user.api.ts`), `pages/Login.vue` + `/login` 라우트, `common/api`(api/query util), `feature/layout` 전역 다이얼로그(`AppDialog.vue` + `appDialog.store`). 실서버 연동·가드 마무리 남음.

---

## 미해결 이슈 (이월)
- **`ProcessDialog` `PROCESS:UPDATE`/`STATUS` 리스너**: 아직 `console.log` 스텁(`patchStatus` 진입점은 이미 있어 연결만 하면 됨).
- **파일 전송**: 구현 완료 / e2e 미검증. 잔여: e2e 스모크 / abort sentinel
- **서브키 충돌/위조**: key↔subkey 결속 검증 미구현(node roster에서 닫을지 보류)
- **supervisor 영속성**: registry 메모리 → PG 미착수

## 잔여 (틈날 때)
- `api.util.ts` `throwCatch`가 HTTP 상태 코드를 버림 → 서버 미기동도 세션 만료처럼 `auth` null → 사용자 설정(`userPrefs`) 삭제. 공용 에러 처리 손볼 때 401만 삭제하도록(`auth.store.ts` TODO)
- `SESSION_KEY` 등 env화(현재 `"irony"` 하드코딩)
- checkSession createdAt=0(pgtype.Timestamptz gob 미직렬화) → sess.Data.ID로 DB 재조회
