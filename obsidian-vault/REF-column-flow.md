# REF — 칼럼 흐름 (타일이 Grid 경계에 걸쳐 놓임)

> 2026-11-11 신설. 이력 → `history/column-flow.md` / 옛 규칙 → `REF-node-ui-projection.md` 규칙 3 / 미니맵 진행도 → `REF-board-pager.md` / 현재 진행 → `CURRENT.md`

## 계기
- 사용자 질문(2026-11-11): A[.5,.5] B[1,.5] C[1,.5]에서 C가 A 옆으로 안 옴 → 원인 = `pack`의 "Grid = 2×2 상자, 경계를 넘지 않음" 전제(금지 결정 기록 없음, 처음 설계부터의 전제). 사용자 "1.0을 유지한 채 B와 절반 겹치고 0.5만큼 더 나아가면 되잖아" → 칼럼 흐름으로 바꿈
- 화면상으로 가능한 근거: strip padding = Grid 간격 = 타일 간격이라 경계에 걸쳐도 1.0 폭이 맞음, 스냅도 이미 칼럼 단위

## 결정 (2026-11-11 사용자 "진행")
| 번호(이름) | 결정 |
|---|---|
| 칼럼-1(범위) | 보드 먼저, **바로 이어서 관리 페이지**(내 장비·관리자·updater, 별도 N×M `pack`) |
| 칼럼-2(Grid 개념 폐지) | 보드 배치 = 2행짜리 칼럼이 이어진 판 하나. 인디케이터 점 하나 = 페이지(화면 폭 = 칼럼 `cap.x`개), 미니맵 = 이어진 한 줄, GridToast(Grid 정보 안내) 삭제 |
| 칼럼-3(채우는 순서) | 위→아래, 다음 칼럼, 앞으로만(⑧(같은 Grid 앞쪽 빈칸 채우기 안 함) 그대로). 바뀐 건 "안 들어가면 다음 Grid" → "경계를 넘어 이어서" |
| 칼럼-4(빈 칸에 놓기) | 빈 칸 = 채우는 순서상 **그 칸 앞 마지막 타일 뒤**(관리 페이지 끌기와 같은 방식). 판 오른쪽 너머 = 맨 뒤 |
| 칼럼-5(전체보기) | 보드 하나 = 이어진 한 띠, `G1 · n타일` 캡션 없음 |

- 폰 세로(가로 1칸) = 칼럼 하나 = 화면 폭이라 모든 타일 1칼럼(지금과 같음). 폰 가로(2×1)는 PC처럼 걸침
- 저장값(크기·순서)·백엔드 변경 없음 — 배치는 프론트에서 매번 계산

## 구현 지도 (보드, `feat/column-flow`)
| 번호(이름) | 파일 | 내용 |
|---|---|---|
| 흐름-1(채우기) | `widget/util/tileOrder.util.ts` | `pack` → `FlowPlan`(items 전역 칼럼 `c`, `empty`, `cols`). 칸 번호 = `c*2 + r`(채우는 순서), cursor 뒤에서 첫 맞는 칸. `spanOf(size, cap)` = 가로 1칸 화면은 늘 1칼럼, 세로 1칸이면 2행. 빈 칸 = 칼럼이 통째로 비면 슬롯 하나. `lastBefore` = 칼럼-4용 |
| 흐름-2(기하) | `widget/util/tileSplit.util.ts` · `node/store/tileGrids.store.ts` | `colWidth`·`flowWidth`, `stripGeometry` = 칼럼 목록만, `stripPos` = 칼럼 단위. 스토어 `plan`(옛 `grids`)·`pages`·`pos`(페이지 단위)·`range`(페이지)·`goPage`(옛 `goGrid`)·`colRange` = 전역 칼럼. `enterOverview` = 보던 칼럼 위치 × 축소 |
| 흐름-3(그리기) | `node/component/TileStrip.vue` | `TileLayout` 하나(2행 × `cols`), `.cell` 폭 = `flowWidth` px, 스냅 앵커 칼럼마다. 다른 보드 줄도 같은 모양. 타일이 경계를 넘어도 부모가 같아 다시 마운트되지 않음 |
| 흐름-4(끌기) | `widget/util/tileDrop.util.ts` · `node/store/tileDrag.store.ts` | `DropFrame.origin(flowW)` 하나(옛 `boxes(cols)`), 전체보기 가운데 정렬은 판 폭으로. 빈 칸 = `lastBefore`(묶음은 묶음 빼고) |
| 흐름-5(인디케이터) | `GridPager.vue` · `TilePeek.vue` · `TileNav.vue` · `pagerBoard.store.ts` | 점 = 페이지, 숫자 `1-2 / 2` = 페이지. `PeekGrid.half` → `size`(화면 몇 개, 보드 = `cols / cap.x`, 관리 페이지 = 1). 전체보기 요약 `Grid N` → `타일 N`. `GridToast.vue` 삭제 |

## 확인 (2026-11-11, 헤드리스, `/api/tiles` GET을 가로채 크기만 바꿈·PUT 막음)
- `pack` 단독: 사용자 예 = `A@0,0 B@0,1(2칼럼) C@1,0(2칼럼)`, 빈 칸 (2,1) / 0.5×0.5 6개 = 3칼럼 꽉 / 섞인 크기 = 앞으로만이라 (2,0) 빈칸 남음
- PC 1400×800: C(3번 타일)가 x 704부터 1384폭 = A 옆 걸침, 끝 스크롤 696(칼럼 하나), 인디케이터 `1-2 / 2`, 전체보기 한 띠
- 폰 390×844: 전부 1칼럼, `2 / 2`
- 미니맵 = 화면 1.5개 폭 한 줄, 진행도(`.puck`) 그대로
- 끌기: B(1.0)를 빈 칸 (2,1)에 → "C 뒤"로 가서 B = (1,0)~(2,0), 빈 칸 그대로 → 다음 판정에서 앞 타일 = B 자신이라 그대로(흔들림 없음). 넓은 타일은 반 칸 빈칸에 못 들어감(규칙대로)
- 주의: 8012를 다른 작업공간(community 프론트)이 쓰고 있어 dev supervisor = 5051 + vite `VITE_API_URL=http://localhost:5051`로 확인. 소켓은 연결 안 됨(`WebSocket closed without opened`, 배치 확인과 무관)

## URL focus (2026-11-11 사용자 "진행")
| 번호(이름) | 결정 |
|---|---|
| 포커스-1(스크롤 위치) | 새로고침·접속 = `?focus` 타일로 타일 위치 규칙(`reveal`) — 왼쪽 끝에서 막 다 보이는 만큼. 첫 적용은 `instant`, 그 뒤(다른 화면이 보낸 이동) `smooth` |
| 포커스-2(focus 없음) | 맨 왼쪽에서 시작 |
| 포커스-3(경로 합치기) | 옛 "view 있음 = 새로고침 복원(`restore`) / 없음 = `activate`"를 `restore(focusId, behavior)` 하나로. 포커스는 PC에서만(U3) → 폰에서 실행 목록·알림·전체보기 다른 보드 이동으로 터미널에 가도 키보드 안 뜸(사용자 "항상 뜨면 난감"). URL을 안 거치는 `activate`(단축키·이 탭에서 연 타일)는 그대로 |
| 포커스-4(관리 페이지) | 보드와 같게 — active 타일을 `?focus`로, `?view`(주소-4(보던 위치)) 삭제. 관리 페이지 칼럼 흐름 작업 때 |

- 구현: `tileUrl.hook.ts`(view 쓰기·읽기 삭제, `first` 플래그) · `tileGrids.store.ts`(`anchorId` 삭제, `restore` 개정, `reveal(tileId, behavior)`)
- 확인(헤드리스, 응답에 폴더 타일 사본 하나 끼움): 쿼리 없음 = 스크롤 0 / PC `?focus=3번` = 스크롤 696(칼럼 1)·틀 포커스 / 옛 `?focus&view` URL = view 지워짐 / 폰 `?focus=3번` = 스크롤 382·포커스 없음

## 구현 지도 (관리 페이지 — 내 장비·관리자·updater, 2026-11-11 사용자 "진행")
| 번호(이름) | 파일 | 내용 |
|---|---|---|
| 관리-1(채우기) | `member/util/memberGrid.util.ts` | `pack` → `MemberPlan`(items 전역 칼럼 `c`, `cols`). 칸 번호 `cellIndex(c, r, rows)` = `c*rows + r`(행 1~2, 보드 `pack`은 2행 고정이라 따로 둠). 타일 폭 = `min(w, 화면 열 수)` 그대로 |
| 관리-2(그리기) | `MemberPage.vue` | `.mb-grid` 하나(CSS grid `cols`열). PC 폭 = `100% * cols / 화면 열 수`(칼럼 폭 옛 Grid와 같음), 폰 = `100cqw * cols` + 칼럼 사이 `2gap`(양쪽 여백 gap), 스냅 = 칼럼마다(폰은 `scroll-margin-left: gap`) |
| 관리-3(스토어) | `memberBoard.store.ts` | `grids` → `plan`(측정 전 null), `colRange` = `lead + c`, 미니맵 = 판 하나(`size = cols / span`) |
| 관리-4(끌기) | `memberDrop.util.ts` · `memberDrag.store.ts` | 판 상자 하나. `MemberFrame.gapX/gapY`(폰 칼럼 사이 2gap을 `columnGap`으로 잼 — 옛 코드는 폰에서 Grid마다 gap만큼 어긋났음). 빈 칸 = 그 칸 앞 마지막 타일 뒤, 판 오른쪽 너머 = 맨 뒤 |
| 관리-5(인디케이터) | `MemberDots.vue` | 라벨 `Grid N` → `페이지 N`(점은 이미 화면 폭 단위) |
| 포커스-4(관리 페이지) | `memberUrl.hook.ts` · 스토어 | `anchorKey`·`restoreView`·`?view` 삭제, `?focus` = active(있는 타일만). 들어올 때 `restore(key)` = active + `reveal(instant)` + 포커스는 PC에서만(`pointer: fine`, 타일 틀 `section`이라 폰에서도 키보드는 안 뜨지만 보드와 규칙 하나로). 관리 페이지는 다른 화면이 보내는 이동이 없어 늘 instant. 판 측정 전이면 `waitPlan`으로 미룸 |

## 확인 (관리 페이지, 2026-11-11, 헤드리스 `/devices`, 테스트 계정 `mflow-*` + 장비 3개 만들고 끝에 장비 삭제)
- 사용자 예와 같은 모양 A 1×1·B 2×1·C 2×1, PC 1400×800(상세 2열×2행): A (0,0), B (0,1)~(1,1), C (1,0)~(2,0) = 페이지 경계 걸침(x 936, 폭 920)
- `?focus=C` = 스크롤 464(칼럼 1, C 오른쪽 끝이 막 보임)·C active·포커스 C / focus 없음 = 스크롤 0·포커스 없음 / 옛 `?view` URL = view 지워짐
- 폰 390×844: 전부 1칼럼(A·B 한 칼럼 위아래, C 다음 칼럼), `?focus=C` = 스크롤 780(목록 + 1칼럼)·포커스 없음, 점 = `목록|페이지 1|페이지 2`
- 끌기: B를 빈 칸 (2,1)에 → C 뒤로 가서 C (0,1)~(1,1)·B (1,0)~(2,0), 빈 칸 그대로라 다시 판정해도 흔들림 없음 → 놓은 뒤 URL `t` 순서·`focus=B`
- updater 페이지(`/updater`)는 같은 틀이지만 관리자 계정이 없어 화면 확인 안 함

## 마감 (2026-11-11)
- 사용자 "문제 없어 마감해도돼" → `feat/column-flow` 커밋·master `--no-ff` 머지. 같은 날 test3용 빌드·집 서버 묶음 → `REF-home-deploy-updater.md` "11-11 칼럼 흐름 묶음"
