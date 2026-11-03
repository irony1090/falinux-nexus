# REF — 회원 관리: 화면 개편 구현 · 판 계약

→ 결정(화-1~16·URL 저장) `REF-member-ui.md` · 이력 `history/member-ui-impl.md` · 현재 `CURRENT.md`
> 2026-11-03 `REF-member-ui.md`에서 분할(10k자 넘음). 옮긴 절 = "코드 적용 계획"·"구현"·"판 계약"

## 코드 적용 계획 (다음 세션, 확인 전 초안)
| 단계 | 내용 | 손댈 곳(예상) |
|---|---|---|
| UI-1(서버 목록) | 화-4: `/admin/users`·`/admin/devices`·`/devices`에 `q`·`f`·`page`·`size` -> `{ items, total }`. 필터 개수(알약 옆 숫자)도 같이 내려줌 | `router/admin.go`·`device.go`, `query/users.sql`·`devices.sql`·`workers.sql` + sqlc |
| UI-2(틀) | 목록 열 + 상세 strip + 아래 푸터. 화-6 계산 함수, 화-7 N×M `pack`은 **관리 페이지 전용 새 util**(보드 `tileOrder.util.ts`는 안 건드림). 이동 = 보드 `revealCol` 재사용 | `feature/member/util/memberGrid.util.ts`(새), `MemberPage.vue` 교체, `FootPill`·`BoardMenu` 로고 재사용 |
| UI-3(목록 타일) | 화-3: 검색·필터·줄 수·페이지 번호, 회원 행 / 장비 행, 열린 항목 바탕 | `feature/member/component/` 새 컴포넌트 |
| UI-4(상세 타일) | 화-8~11·15: 회원 타일 / 장비 타일(키·env·인스턴스 라벨) / 등록 타일, 경로·크기·닫기, 자식 끼워 넣기 | 위 + `DeviceCard.vue` 대체 |
| UI-5(URL) | 주소-1~6 | `feature/member/hook/` 새 훅(보드 `tileUrl.hook.ts` 참고) |
| UI-6(로그인·가입) | 화-16 | `pages/Login.vue`·`Signup.vue` |
| UI-7(확인) | 격리(5051·3001) 헤드리스 PC 1280·1920·폰 390 + 사용자 브라우저 | scratchpad |
- 순서 제안: UI-1 -> UI-2 -> UI-3 -> UI-4 -> UI-5 -> UI-6 -> UI-7. 착수 전에 이 표를 사용자에게 보이고 맞춘다(전역 규칙)

## 구현 (2026-10-26 오후, 커밋 전)
| 단계 | 결과 |
|---|---|
| UI-1(서버 목록) | 새 `router/memberList.go`(`parseListQuery`·`pageList` 제네릭·필터 표). 접속 여부가 메모리라 **SQL·sqlc는 안 바꾸고 Go에서 거름**. `/admin/users`·`/admin/devices`(+`owner=`)·`/devices` -> `{items,total,page,counts}`, `size=0` = 전부. 타일용 하나 조회 `GET /admin/users/:id`·`/admin/devices/:mainKey`·`/devices/:mainKey` |
| UI-2(틀) | `util/memberGrid.util.ts`(layoutOf·pack·insertTile·encode/decodeTiles) · `store/memberBoard.store.ts`(provide/inject, 칼럼 스크롤 + 보드 `revealCol`) · `MemberPage`(PC = strip만 스크롤·칼럼마다 스냅 표시 / 폰 = 목록 + Grid 한 줄, `100cqw`) · `MemberFooter`·`MemberMenu`·`MemberDots`. 공용 모양 `styles/member.scss`(`mb-*`) |
| UI-3(목록 타일) | `MemberList.vue` — 검색 0.5초 지연·필터 알약 개수·줄 수·페이지 번호(끝 블록) |
| UI-4(상세 타일) | `MemberTileView`(경로·크기·닫기·active 테두리) + `MemberUserTile`·`MemberDeviceTile`(라벨 칸 바로 고침, blur·Enter 저장)·`MemberNewTile`(등록 -> 같은 자리 장비 타일, 키 펼침). `DeviceCard.vue` 지움, `pages/Admin.vue`·`Devices.vue`는 provide + `MemberPage`만 |
| UI-5(URL) | `hook/memberUrl.hook.ts` — push(`t`·`tab`·`f`) / replace(`p`·`n`·`q`·`view`), 마지막 구성 = `readUserPref`(`member.last.<page>`, 로그아웃 때 지워짐), 없는 대상 = 타일 쿼리 404면 `dropMissing`(replace로 덮음 — push면 뒤로 가기가 그 자리에서 맴돎). 함정: route 감시 원천을 배열로 주면 apply가 매번 새 타일 배열을 만들어 무한 재실행 -> 글자열로. 첫 측정 전엔 폭 0 = 폰 판정이라 PC·폰이 바뀔 때 스크롤 요소를 다시 고른다(`watch(phone, measure)`) |
| UI-6(로그인·가입) | `component/AuthTile.vue`(가운데 타일·머리 줄·끝 블록) + `pages/Login.vue`(오류 = 알약, 서버 연결 불가도 알약, 아이디 저장 = 토글 알약, 옛 오류 창 없앰)·`Signup.vue`("‹ 로그인" 왼쪽 벽, 확인 불일치 힌트) |
| UI-7(확인) | 헤드리스: 화면 `shot.mjs`(1280·1920·390·내 장비) · 패널 `panel.mjs` 17/18 · URL `url.mjs` 15/15 · 로그인 `auth.mjs` 5/5(가입 시험 계정 `zz60597` 생김). 사용자 브라우저 확인 남음. ESLint는 이 worktree에 `eslint-config-vuetify`가 없어 못 돌림(vue-tsc·go vet 통과) |
- **화-10 개정(2026-10-26 사용자 B안)**: 머리 줄 ↔·↕ 대신 **보드식 [크기 | 닫기] 끝 블록** — 크기 = 팝업 `MemberSizeCard`(폭 1~열 수·높이 1~행 수, 보드 `TileSizeCard` 모양), 끝 블록은 active일 때만 진한 파랑(보드와 같음, 시안의 "닫기 늘 파랑" 대신). 머리 줄 왼쪽에 위치 번호(보드 `.pos`)
- **컨트롤 패널(2026-10-26 사용자 요청)**: `MemberPagerPopup` = 보드 `PagerPopup`과 같은 여는 길(Ctrl+Alt 누르는 동안 / 길게 누르기 — 자리 = `.mb-head` + 타일 밖 빈 곳) + 알약 [크기 | 도트](전체보기 없음) + Ctrl+Alt+숫자 = 위치 번호로 active. **공유 방식(사용자 선택)**: 여는 로직 `widget/hook/pagerOpen.hook.ts` `usePagerOpen`, 숫자 단축키 `widget/hook/digitShortcut.hook.ts` `useDigitShortcut`로 빼고 보드 `PagerPopup`·`useTileShortcut`도 이것을 씀(보드 동작 그대로 — 헤드리스 회귀 확인)
- **알약 끌기 = 보드와 같음(2026-10-26 사용자 "둘 다"·A안)**: 푸터 알약(위로 당김)·컨트롤 패널 알약(위·아래) 모두 — 가로 끌기 = 칼럼 이동, 당긴 채 끌기 = 위치 번호 순으로 active 한 칸씩 + 미니맵. 폰 목록 칸에서 당기면 그 뒤 첫 타일부터. PC만 끝나고 포커스(폰은 가상 키보드 때문에 안 줌, 보드 U3과 같음)
- 확인(헤드리스 `drag.mjs`) 15/15 — 회원 PC·폰 + 보드 푸터 회귀
- **빈 곳 휠(2026-10-26 사용자)**: 보드 푸터-7(빈 곳 휠)과 같음 — 타일 밖(타일 사이·빈 칸·푸터) 세로 휠 = 칼럼 한 칸, 목록·상세 타일(`.mb-tile`) 위는 그 안 스크롤. `useGapWheel`을 `widget/hook`으로 옮겨 판 계약으로 씀(판을 provide한 같은 컴포넌트에서 부르므로 inject 대신 `ctx.pager`를 넘김). 헤드리스 `wheel.mjs` 6/6(보드 쪽은 시험 보드에 타일 1개라 스크롤 없음 확인만)
- **끌기 간격 1.5배(2026-10-26 사용자 "간격이 짧아 세세하게 맞추기 힘듦")**: `pagerDrag.hook` `COL_PX` 36 -> 54 · `TILE_PX` 29 -> 44. 공용이라 보드도 같이 바뀜(사용자 "보드도 같이 바껴도 돼"). 도트 알약 ←/→ 키(보드에 있음)는 안 넣음(사용자 "필요없을 듯")

### 판 계약 (PagerBoard, 2026-10-26 A안)
타일 판을 또 만들 때 알약 끌기·미니맵·컨트롤 패널을 그대로 붙이려고, 보드 저장소에 묶여 있던 것을 계약 하나로 뗐다(사용자: "별도로 떼서 쓸 수 있게"). 판 전체(배치·스크롤·URL) 공통화 = B안은 안 함 — 보드 고유 동작(전체보기·Grid별 칼럼 수·미리보기)이 많아서. 계약이 B안의 출발점.

| 파일(`feature/widget/`) | 역할 |
|---|---|
| `store/pagerBoard.store.ts` | 계약 `PagerBoard` + `providePagerBoard`/`usePagerBoard`. 칼럼(스크롤 멈춤 한 칸)·`span`·`maxCol`·`viewCol`·타일 순서·active·`scrubbing`·`colLeft`/`goCol`/`scrubTo`/`colRange`/`setActive`/`focusTile` + 미니맵 `peekGrids`(스크롤 좌표 x·w + cols·rows·half)·`peekView`·`aspect` |
| `hook/pagerDrag.hook.ts` | 끌기(옛 `node/hook`). `stepTo`(타일 한 칸 = revealCol)는 보드 저장소에서 여기로 옮김 |
| `component/TilePeek.vue` | 미니맵(옛 `node/component`). 칸 높이 = `r / rows`(보드는 rows 2) |
| `component/PagerFloat.vue` | 끌기 중 안내 + 미니맵 teleport(옛 `GridPager` 안). 안내 class `.GridPager__hint` -> **`.PagerFloat__hint`** |
| `hook/pagerOpen.hook.ts` · `hook/digitShortcut.hook.ts` | 컨트롤 패널 여는 길 · Ctrl+Alt+숫자(옛 `node/hook`) |
| `hook/gapWheel.hook.ts` | 빈 곳 휠(옛 `node/hook`). `useGapWheel(판, isTile)` — 휠을 그대로 둘 타일 판정은 판마다 |

- 채우는 쪽: 보드 `tileGrids.store.ts` 끝(strip 기하 그대로, 스냅 = 칼럼 x - gap) / 관리 `memberBoard.store.ts`(칼럼 = 화면 폭 / 열 수, 폰은 0 = 목록. Grid 안쪽 여백 PC 오른쪽만·폰 양쪽). 판은 `scrubbing`이면 스냅을 꺼야 한다(`MemberPage` `.scrubbing`, 보드 `TileStrip`)
- 도트 모양은 판마다 따로: 보드 `GridPager`(창 방식·숫자) / 관리 `MemberDots`(폰 목록 도트 진하게)
- 확인(헤드리스 `panel.mjs`) 17/18 — 남은 1개(보드 폰 첫 길게 누르기)는 옛 코드도 똑같음(로딩 직후 그 자리에 다른 요소)
- 푸터 높이 = 보드 푸터와 같게(PC 44·폰 56, 시안 50 아님)
- 격리 DB `nexus_member_tmp`에 시험 데이터 추가: 가입 회원 24명(minjun~parkj, 비번 `pw`) · 대신 등록 장비 22개 · `doyun`(12) 등록 막힘, `seoyeon`(9) 관리자
- 헤드리스(`shot.mjs`, scratchpad) PC 1280·1920·폰 390·내 장비 1280 화면 확인
