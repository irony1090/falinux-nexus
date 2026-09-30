# history/node-ui-render — node 카탈로그 타일 UI 렌더 · 헤더 내비 (컴포넌트 층)

> `history/node-ui-impl.md`와 짝으로 2026-09-29 분할 — ⑤ 렌더 최소판·칼럼 스냅 이력((14)(15))은 분할 전이라 `history/node-ui-impl.md`에 있음. 요약·재사용 지식 → `REF-node-ui-render.md` / 현재 진행 → `CURRENT.md`

## 2026-09-30 — 타일 위치 따라가기 `reveal` 작성
- 사용자 요청: "사이즈 조절이 돼서 페이지가 옮겨졌을 경우 해당 위치로 이동되는 로직도 필요". 흐름안 제시 → 질문 3개 모두 추천안 선택: 화면 밖일 때만 이동 / 그 타일의 Grid 시작(`goGrid`) / 새 타일 생성(`onOpen`/`onExec`)에도 적용.
- **작성**: `tileGrids.store.ts` `reveal(tileId)`(`nextTick` 후 타일 x 구간을 화면과 비교) / `TileFrame.vue` `setSize` 끝에서 호출 / `FolderTileBody.vue` `onOpen`/`onExec`가 새 타일 id로 호출.
- **검증**: vue-tsc 통과 + 헤드리스 Chromium PC 1440×900 / 폰 390×844 시나리오 전부 기대대로(→ `REF-node-ui-render.md` "타일 위치 따라가기"). 미커밋.

## 2026-09-29(17) — 헤더 내비 작성 (Grid 탭 · 칼럼 위치 · 칼럼 이동)
- 구조 제안 → 사용자 승인("둘 다 추천대로"): **F(헤더 전달 방식) = Teleport**(provide를 `App.vue`로 올리지 않음, CURRENT의 "provide를 `AppHead`까지 올림" 계획 대체) / **G(화살표 = 칼럼 이동)** = VPagination `prev`/`next` 슬롯을 칼럼 단위 이동으로(열린 질문 ⑬ 해소).
- **작성**: `feature/node/component/{TileWorkspace.vue, TileNav.vue}` 신규. **개정**: `tileSplit.util.ts`(`stripGeometry`/`stripView`), `tileGrids.store.ts`(`area` 내부 생성·`TILE_GAP` export·스크롤 상태·`goGrid`/`goColumn`), `TileStrip.vue`(provide 제거, `stripEl`/`scrollX`/`viewW` 기록), `AppHead.vue`(`#app-head-nav` 자리 + 제목 flex), `pages/index.vue`(`<tile-workspace/>`).
- **발견해 고친 것**: ① 스크롤 끝에서 현재 탭이 G1에 남음 — Grid 구간 겹침에 칼럼 사이 gap 8px가 섞여 716 대 708 → 칼럼 겹침 합으로 변경. ② `text-caption`이 Vuetify 4에 없어 `칼럼` 표시가 16px로 나옴 → `text-body-small`.
- **검증**: `vue-tsc` 통과. 임시 dev 서버(3100) + 헤드리스 Chromium(playwright-core를 scratchpad에 설치, `~/.cache/ms-playwright/chromium-1234`)으로 PC 1440×900 / 폰 390×844 — 칼럼 이동·끝 비활성·탭 클릭 스크롤·직접 스크롤 시 탭 추종·Grid 10개 `…` 줄임 모두 기대대로. 스크린샷으로 헤더 한 줄 배치 확인.
- 이번에 vault 분할: `REF-node-ui-impl.md`가 헤더 내비 절 추가 시 10k 초과 → ⑤ 렌더 절 + 헤더 내비를 `REF-node-ui-render.md`로.
