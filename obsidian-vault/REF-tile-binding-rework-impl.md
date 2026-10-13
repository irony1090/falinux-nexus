# REF — 바인딩 타일 규칙 개편: 구현 · 확인

> 2026-10-13 `REF-tile-binding-rework.md`에서 분리(결정·구조안은 거기). **사용자 확인 완료 → 파트(바인딩 타일) 마감·커밋(2026-10-13)**. 이력 → `history/tile-binding-rework-impl.md` / 현재 진행 → `CURRENT.md`

## 구현 메모 (2026-10-13~)
| 단계 | 구현 |
|---|---|
| 개편-1(고정 지도) | `protocol.TilePin{nodeId,hash,tileId}` + `TILES:PINS`(`TilesPinsEvent`). `pinsOf(hashes, docs)` = `folderTiles()`(루트 목록 제외 너비 우선)로 노드별 타일 수·`betterBinding`(bindingTile과 공용 비교) → 2개 이상만. `loadPins` = `ListTileTreeDocs`(**`ORDER BY id` 추가** — bindingTile의 `ListTileTreesForUpdate` 순서와 맞춰 동률 처리 일치). `GET /tiles` 응답 `pins`(`*[]`, 409 응답엔 없음) / `publishTiles` 끝에 `publishPins`(모든 호출이 커밋 뒤라 그 자리). 확인: build·vet + overlay 단위 테스트(먼저 연 타일·고정 우선·bindingTile 일치·타일 1개면 없음) |
| 개편-2(여기로 고정) | `tileBindRequest.Bound` 삭제 → 항상 그 타일 고정 + 같은 노드 다른 고정 풂 |
| 개편-3(목록 API) | `ListVisibleLiveProcessesInFolder` → `ListVisibleRunsInFolder` + 새 `ListVisibleRunsInSubtree`(`WITH RECURSIVE`, folder NULL = 볼 수 있는 전부) — 둘 다 `status 살아 있음 OR uid = ANY(tile_uids)` + `n.name`. `listNodeRuns`가 내 트리 터미널 uid를 먼저 모아 넘김, `?deep=1`. 응답 `name` 추가 |
| 개편-4(현재 타일에 바인딩) | `openProcessTile` → `bindProcessTile`(`share.go`, 같은 라우트) = `lockWorkspaces` → 어느 보드든 터미널 타일 `removeLeaf` → 지정 폴더 타일(없으면 그 보드 루트) kids 끝, 최소화 풂, 타일 id 유지 → 바뀐 보드 저장·push + BIND 알림. 타일 없음 + 끝난 process = 400. `boardPlace` 삭제(수동 열기가 이걸로 대체) |
| 개편-5(알림 이벤트) | `tileNotice.go`: `pinState`(계정별 마지막 지도 + 계정별 mutex로 읽기·비교·발행 한 줄) → `publishPins`가 PIN_MOVED(두 지도 모두에 있는 노드만, 서버 재시작 뒤 첫 계산은 기록만). `attachViewerTiles` → ATTACH / NO_PLACE. 공유 걸 때·노드 이동 때 붙는 터미널은 알림 없음(SHARE:CHANGE로 충분) |
| 개편-6(목록 UI) | `RunRow.vue`(행 공통: 위치·보기/열기/보드로 이동·현재 타일에 바인딩(이미 그 타일 아래면 숨김)·종료(조작 이상)/닫기(내 보드의 끝난 타일)) / `ScriptRunList` = RunRow 목록(`tileId` prop) / `RootTerminalList` = 루트 타일 nodeId로 `useScriptRuns(…, deep)`, 최소화 먼저. `TerminalListRow.vue` 삭제. 스크립트 행 문구 = 실행 중 N · 끝남 n · 최소화 n |
| 개편-7(고정 표시 UI) | `tileTree.store` `pins_`(GET·재접속 GET·`TILES:PINS`, GET 도중 push가 오면 GET 값 무시) + `pinByNode`. TileFrame 핀 = 그 노드 pin이 이 보드·이 타일 / 메뉴 "여기로 타일 고정" = pin 있고 이 타일이 아닐 때 |
| 개편-8(알림 UI) | `tileNotice.store.ts`(모듈 메모리: 내역 100개·안 읽은 수·토스트 5초·800ms 묶음·최대 3개) + `tileNotice.hook.ts`(`TILES:NOTICE` → 문구·토스트 여부·이동할 곳 / `useNoticeGo`) + `NoticeCenter.vue`(AppHead 새 자리 `#app-head-tools`의 종·배지 + 오른쪽 드로어 + `NoticeToasts.vue`). 이동 = `focusPath(board, tileId)` → `TileWorkspace`가 `?focus`(타일 생기면 최소화 풂·activate) / `?tab=terminals` 처리 후 query 지움. 로그인 계정 바뀌면 내역 비움 |
- 확인(코드): `go build`·`go vet`, `npm run type-check`, `vite build` 통과. ESLint는 설정 의존성(`eslint-config-vuetify`)이 설치돼 있지 않아 못 돌림
- 개편-9에서 고친 것: ① 드로어 윗부분을 헤더가 덮음(헤더 높이 auto라 Vuetify 레이아웃이 모름) → `AppBody`처럼 `useAppHead().size`로 `top`·`height` 직접 ② 터미널 탭 라벨이 옛 기준(이 보드 터미널 타일 수)이라 "터미널 0"인데 목록 2행 → TileFrame이 루트일 때만 같은 `deep` 조회(`useNodeRuns(…, enabled)`)로 셈

## 개편-9(확인) 결과 (2026-10-13, 격리 실행 — 임시 DB `nexus_rework_tmp` + supervisor 5051 + worker `NAME=rework-dev` + 프론트 3001, 끝난 뒤 끄고 DB 삭제)
- **API·소켓 34/34** (`rework_test.py`, A 소유·실행 / B 조작 공유, S > s1 · S > X > x1 · P > p1): 타일 1개면 pins 없음 / pins = 먼저 연 T1 + `TILES:PINS` / 기본 보드 보는 중 실행 → T1 아래 + ATTACH(to) / 보드2만 보면 NO_PLACE·타일 없음(x1도 S 기준) / 목록: S 직접 자식·S deep·X deep·루트 deep·name·위치 / 바인딩: 새로 만듦(BIND from 없음) · 다른 보드에서 이동(id 유지·최소화 풂·from) · 같은 자리 false · 없는 타일 = 보드 루트 · 고정 불변 / `bound:false`여도 고정 / PIN_MOVED T1→T3 · 고정 타일 닫음 → T1 + PIN_MOVED / 타일 1개 남으면 핀·알림 없음 / 사용자 시나리오 4단계(다시 S로 와도 보드2 유지) / 끝난 process: 내 타일 있음 = 남음 · 없음 = 빠짐 · 끝났고 타일 없음 바인딩 400 · 타일 있으면 이동 / 단독 공유 = 루트 deep에만
- **화면 23/23** (`pw/rework.mjs`, A 실행은 `aexec.py`): 핀(먼저 연 타일) · 메뉴 "여기로 타일 고정"(해제 없음) → 핀 이동 + 고정 이동 토스트 / 보는 보드에 붙음 + 토스트 / 노드의 상태 리스트 행(이미 이 타일 아래면 바인딩 숨김) / 위치 없음 토스트 → 누르면 터미널 탭 / 탭 라벨 = 목록 수 · 2행(보드2·위치 없음) / 바인딩 → 이 보드에 타일 + 토스트 + 행 "이 보드" / 보드로 이동 → 그 터미널 active·query 지움 / 종 배지 · 드로어(새것 먼저) · 열면 배지 사라짐 · 내역 누르면 그 보드·타일 active / 페이지 오류 없음
- 테스트 함정: `page.goto`는 새로고침이라 알림 내역이 비워짐(정상 — 내역은 탭 메모리)
- 재작성 명세(scratchpad, 재부팅 시 사라짐): 위 항목 그대로. 공유하면 받는 쪽 기본 보드에 S 타일이 생김(T1) / 보드2 = 타일 하나 더 열고 `POST /workspaces {folderTileId, name, version}` / A 실행은 소켓 붙은 세션의 `X-Tab-Id` 필요

## 사용자 확인 목록 (supervisor 재시작 + 프론트 새로고침, worker 무관)
| # | 확인 | 기대 |
|---|---|---|
| C1(핀 표시) | 같은 공유 폴더를 두 보드에 열기 | 먼저 연 타일에만 핀, 타일 1개일 땐 핀 없음 |
| C2(여기로 고정) | 다른 타일 ⋮ "여기로 타일 고정" | 핀 이동 + 하단 토스트, "해제" 메뉴 없음 |
| C3(자동으로 붙음) | 핀 타일 보드를 보는 중 공유자 실행 | 핀 타일 아래 터미널 + 토스트 |
| C4(위치 없음) | 다른 보드를 보는 중 공유자 실행 | 타일 안 생김 + "위치 없음" 토스트 → 누르면 터미널 탭 |
| C5(터미널 탭) | 루트 타일 터미널 탭 | 루트 경로 하위 process 전부(다른 보드·위치 없음 포함), 라벨 숫자 = 행 수, 루트 경로를 바꾸면 목록이 바뀜 |
| C6(현재 타일에 바인딩) | 탭·스크립트 행 펼침에서 버튼 | 다른 보드에 있던 타일이 이쪽으로 옮겨짐(원래 보드에서 사라짐), 이미 이 타일 아래면 버튼 없음 |
| C7(보드로 이동) | 다른 보드 행 "보드로 이동" | 그 보드로 가서 그 터미널이 active |
| C8(알림 내역) | 헤더 종 | 안 읽은 수 배지, 드로어 내역, 누르면 그 위치로. 새로고침하면 비워짐 |
| C9(끝난 터미널) | 내 타일에 있던 process 종료 | 목록에 남고(끝남) 닫기 버튼, 타일에 없던 process는 끝나면 목록에서 빠짐 |
