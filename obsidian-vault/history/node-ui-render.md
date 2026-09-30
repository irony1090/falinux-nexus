# history/node-ui-render — node 카탈로그 타일 UI 렌더 · 헤더 내비 (컴포넌트 층)

> `history/node-ui-impl.md`와 짝으로 2026-09-29 분할 — ⑤ 렌더 최소판·칼럼 스냅 이력((14)(15))은 분할 전이라 `history/node-ui-impl.md`에 있음. 요약·재사용 지식 → `REF-node-ui-render.md` / 현재 진행 → `CURRENT.md`

## 2026-09-30(3) — 전체보기 축소 비율 localStorage 저장 + 로그인 풀리면 삭제
- 사용자 제안: "축소 비율은 localStorage에 저장해서 로그아웃 전까지 유지". 구조안 → 사용자: 세션 만료에도 지워야 함 → 소유자 기록으로 더미 미리보기를 보호하는 안 제시 → 사용자: "더미 미리보기는 실 환경에선 필요 없는 경우" → **`auth`가 null이면 무조건 삭제**로 단순화.
- **작성**: `common/util/userPrefs.util.ts` 신규 / `tileGrids.store.ts`(`readOvPercent` 검증 + 저장 watch) / `auth.store.ts`(`watch(auth)` null → `clearUserPrefs`, 서버 미기동 한계 TODO).
- **검증**: vue-tsc + 헤드리스(세션 응답 모킹) 저장·복원·401 삭제·단위 밖 값 무시. 사용자가 스크롤바 두께 보정 실기 확인. 커밋됨.

## 2026-09-30(2) — 전체보기 모드 + 고정 버튼 작성 (구현 순서 5)
- 구조안 제시(상태·나열·축소·타일·버튼·헤더 층 + 재마운트 금지) → ⑫(전체보기 세부) 질문: 축소 비율 = **사용자 조절**, 거르기 = **안 거름**. ⑭(전체보기 버튼 최종 위치)는 현재안으로 만들고 화면 보며 조정.
- **작성**: `OverviewButton.vue` 신규. **개정**: `tileGrids.store.ts`(전체보기 상태·진입/복귀·배율·스크롤바), `TileStrip.vue`(`.cell/.box` 구조, 축소·휠·Esc·타일 이동), `TileFrame.vue`(`ov-name`·inert·폴더 흐림), `TileWorkspace.vue`(래퍼 + 버튼), `TileNav.vue`(전체보기 요약 + `- n% +`).
- **발견해 고친 것**: 돌아가기 버튼 `color="on-surface"`가 배경으로 안 먹어 밝게 나옴 → `surface-variant`.
- **검증**: vue-tsc 통과 + 헤드리스 PC/폰 시나리오(→ `REF-node-ui-render.md` "전체보기 모드"). 커밋됨.

## 2026-09-30 — 타일 위치 따라가기 `reveal` 작성
- 사용자 요청: "사이즈 조절이 돼서 페이지가 옮겨졌을 경우 해당 위치로 이동되는 로직도 필요". 흐름안 제시 → 질문 3개 모두 추천안 선택: 화면 밖일 때만 이동 / 그 타일의 Grid 시작(`goGrid`) / 새 타일 생성(`onOpen`/`onExec`)에도 적용.
- **작성**: `tileGrids.store.ts` `reveal(tileId)`(`nextTick` 후 타일 x 구간을 화면과 비교) / `TileFrame.vue` `setSize` 끝에서 호출 / `FolderTileBody.vue` `onOpen`/`onExec`가 새 타일 id로 호출.
- **검증**: vue-tsc 통과 + 헤드리스 Chromium PC 1440×900 / 폰 390×844 시나리오 전부 기대대로(→ `REF-node-ui-render.md` "타일 위치 따라가기"). 커밋됨.

## 2026-09-29(17) — 헤더 내비 작성 (Grid 탭 · 칼럼 위치 · 칼럼 이동)
- 구조 제안 → 사용자 승인("둘 다 추천대로"): **F(헤더 전달 방식) = Teleport**(provide를 `App.vue`로 올리지 않음, CURRENT의 "provide를 `AppHead`까지 올림" 계획 대체) / **G(화살표 = 칼럼 이동)** = VPagination `prev`/`next` 슬롯을 칼럼 단위 이동으로(열린 질문 ⑬ 해소).
- **작성**: `feature/node/component/{TileWorkspace.vue, TileNav.vue}` 신규. **개정**: `tileSplit.util.ts`(`stripGeometry`/`stripView`), `tileGrids.store.ts`(`area` 내부 생성·`TILE_GAP` export·스크롤 상태·`goGrid`/`goColumn`), `TileStrip.vue`(provide 제거, `stripEl`/`scrollX`/`viewW` 기록), `AppHead.vue`(`#app-head-nav` 자리 + 제목 flex), `pages/index.vue`(`<tile-workspace/>`).
- **발견해 고친 것**: ① 스크롤 끝에서 현재 탭이 G1에 남음 — Grid 구간 겹침에 칼럼 사이 gap 8px가 섞여 716 대 708 → 칼럼 겹침 합으로 변경. ② `text-caption`이 Vuetify 4에 없어 `칼럼` 표시가 16px로 나옴 → `text-body-small`.
- **검증**: `vue-tsc` 통과. 임시 dev 서버(3100) + 헤드리스 Chromium(playwright-core를 scratchpad에 설치, `~/.cache/ms-playwright/chromium-1234`)으로 PC 1440×900 / 폰 390×844 — 칼럼 이동·끝 비활성·탭 클릭 스크롤·직접 스크롤 시 탭 추종·Grid 10개 `…` 줄임 모두 기대대로. 스크린샷으로 헤더 한 줄 배치 확인.
- 이번에 vault 분할: `REF-node-ui-impl.md`가 헤더 내비 절 추가 시 10k 초과 → ⑤ 렌더 절 + 헤더 내비를 `REF-node-ui-render.md`로.
