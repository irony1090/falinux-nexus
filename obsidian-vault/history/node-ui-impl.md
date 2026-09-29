# history/node-ui-impl — node 카탈로그 타일 UI 구현 (코드)

> 2026-09-29 신설 — 규칙 설계 이력은 `history/node-ui-layout.md`/`-projection.md`/`-overview.md`, 여기는 **코드 구현 이력**. 요약·재사용 지식 → `REF-node-ui-impl.md` / 현재 진행 → `CURRENT.md`

## 2026-09-29(11) — 구현 착수: 층 구조 합의 + ① ② 유틸 작성
- **역할 분담 변경**: 사용자 "코드는 대부분 네가 짤 건데 로직 구조는 같이 맞춰나갈 거야" → 이 작업은 Claude가 코드 작성, 구조는 먼저 합의.
- **층 구조 합의**: ① 순서·채우기 / ② 화면 판정 / ③ 트리 상태(provide/inject) / ④ `useTileGrids` 파생 / ⑤ `TileStrip` > `TileLayout` > `Tile` 렌더.
- 사용자에게 던진 질문 Q1~Q5를 추천안대로 확정: Q1(역할 분담) 위와 같이 / Q2(첫 범위) 유틸+트리+렌더러 / Q3(렌더 방식) CSS grid 좌표 배치, Band 버림 / Q4(트리 보관) 메모리만 / Q5(유틸 위치, ⑦) `feature/widget/util/`.
- 합의 중 발견: 측정을 `TileLayout`(Grid 하나)에서 하면 1칼럼 폭 Grid가 절반 폭을 재서 판정 진동 → **`TileStrip`에서 한 번만 측정**으로 결정.
- **작성**: `apps/frontend/src/feature/widget/util/{tile.type.ts, tileOrder.util.ts, tileSplit.util.ts}`. 시안 `TileOrder.toGrids` → `pack`(좌표 출력 + `empty` 슬롯, `trim` 인자 제거), `TileSplit`에서 `primPhys`/`flexDir` 제외, `minSize` 추가.
- **검증**: `vue-tsc --noEmit` 통과. scratchpad 시나리오 테스트(node strip-types)로 REF 예시 대조 — 순서 표(기본/sup1/sup2/A/htop/X/2F), 크기 섞기 예(G1 column: 기본 | sup,sup / G2 row: A 가로바 + htop | 2F), 중간·마지막 빈 칼럼(G1·G3 `cols=1`), 폰 세로(전부 가로 1.0, 2타일씩) 모두 일치.
- vault: `REF-node-ui-layout.md`에서 하위 주제(코드 구현)가 갈라지므로 "순서 유틸"·"프로토타입"·"구현 착수" 절을 이 쌍(`REF-node-ui-impl.md`)으로 분할, projection의 "판정 유틸" 절도 이동.

## 2026-09-29(12) — ③ 트리 상태 스토어 작성 + 닫기 규칙 확정
- 구조 제안 → 사용자 승인("둘 다 추천대로"): A(닫기와 kill) = 실행 중 터미널 닫기 비활성·kill 후만 닫기(UI가 막음) / ⑩(타일 닫기 규칙) = 시안 가정 확정.
- **작성**: `apps/frontend/src/feature/node/store/tileTree.store.ts` — `provideTileTree`/`useTileTree`, `FolderTile`/`TerminalTile`, `openFolder`/`navigate`/`addTerminal`/`close`/`resize`. 합의안의 `exec` → `addTerminal`로 개명(API는 밖), `navigate` 추가. `tile.type.ts`의 `kids`를 `ReadonlyArray`로(readonly `tiles`를 `flatten`에 직결).
- **검증**: `vue-tsc` 통과 + readonly `tiles` → `flatten`/`pack` 타입 호환 probe 통과. 런타임 시나리오: 순서 예시 트리 구성, 폴더 A 닫기 → htop·X가 A 자리로 올라가 root 소속, 루트 닫기 무시, navigate/resize, 터미널을 부모로 쓰면 throw — 모두 기대대로.
- 발견: 폴더를 닫으면 그 폴더가 실행한 프로세스가 부모 자신의 프로세스보다 앞으로 옴(규칙의 결과, REF에 기록).

## 2026-09-29(13) — 넘겨받은 프로세스(`adopted`) 순서 도입
- 사용자: (12)의 부작용(폴더 닫으면 그 폴더의 프로세스가 부모 자신의 프로세스보다 앞) → "뒤에 오는 게 맞다", 나아가 "부모가 나중에 sup3을 실행해도 `root / sup1 / sup2 / sup3 / htop` 순"을 원함.
- `kids` 순서만으로는 직접 실행/넘겨받음 구분 불가 → `TileNode.adopted?: true` 추가. 순서 규칙 = 폴더 → 직접 실행 → 넘겨받음 → 하위 폴더(`flatten`만 수정). `close` = 하위 폴더는 닫힌 자리, 터미널은 `adopted` + `kids` 끝(닫힌 폴더 안 순서 유지). 플래그는 해제 안 함.
- 검증: tsc·probe 통과, 시나리오(X 닫기 → A 닫기 → sup3 실행 → `root / sup1 / sup2 / sup3 / htop* / serial* / 2F`, root가 B·A 연 경우 `root / htop* / B`) 기대대로.

## 2026-09-29(14) — ④ Grid 파생 + ⑤ 렌더 최소판 작성, 더미 미리보기 동작 확인
- 결정: B(전달 방식) = provide/inject, C(측정 전 상태) = 그리지 않음, ④ 위치 `feature/node/store/tileGrids.store.ts`, D(index.vue 처리) = 스크립트 유지·템플릿만 교체, E(⑥ 높이) = `calc(100dvh - var(--v-layout-top))`.
- **작성**: `feature/node/store/tileGrids.store.ts`, `feature/node/component/{TileStrip.vue, tile/TileFrame.vue, tile/FolderTileBody.vue, tile/TerminalTileBody.vue}`, `feature/node/dev/tileDummy.ts`. **개정**: `feature/widget/component/{TileLayout.vue(cols/rows/gap/footHeight + foot 슬롯), Tile.vue(cell 좌표 + empty)}`, `pages/index.vue`(템플릿 데모·resize 로그 → `<tile-strip/>`, `#Index` 스타일 제거).
- **검증**: `vue-tsc` + `vite build` 통과. 임시 dev 서버(5199) + Chrome으로 확인 — PC 배치·1칼럼 폭 Grid·실행 후 reflow·폰 폭(strip 390px) 2타일씩·kill 전 닫기 비활성 모두 규칙대로. 발견해 고친 것: 네이티브 `<button>` 회색 UA 배경(리셋 추가).

## 2026-09-29(15) — 칼럼 단위 스냅 추가
- `TileLayout`에 `overlay` 슬롯(+ 루트 `position: relative`), `TileStrip`에 `scroll-snap-type: x mandatory` + `columnStops` 앵커. `TileLayout`의 기본 슬롯에 앵커를 넣으면 grid 칸을 차지하므로 별도 슬롯으로(추천안 채택).
- 검증: tsc 통과, Chrome에서 `scrollTo` 시도값 → 가장 가까운 앵커로 스냅(`500→923`, `1900→1845`, `2500→2768`), 923 위치 스크린샷으로 Grid 경계 걸침(G1 오른쪽 + G2 왼쪽) 확인.

## 2026-09-29(16) — 사용자 직접 확인·승인 + vault 정리
- 사용자가 더미 미리보기(①~⑤ + 칼럼 스냅)를 직접 확인 → **"매우 맘에 들어"**. 구조 합의 → Claude 코드 → 브라우저 검증 흐름이 이 작업에서 잘 맞음.
- 남은 것(CURRENT 구현 순서 4~7): 헤더 내비(VPagination, provide 위치를 `AppHead`까지 올림) / 전체보기 / 모바일 키보드 / ProcessDialog 임베드 / 더미 → 실제 연동(node API·execProcess+인스턴스 선택·서버 저장본).
- 미확인: 휠·트랙패드 스냅 체감, 실제 폰·DevTools 기기 모드(창 관리자 때문에 strip 폭 축소로만 확인), 폰 세로 Grid 단위 스냅.
