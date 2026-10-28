# REF — 경로 줄 정리 (파트(경로 줄 정리), 2026-10-28)

> → 이력 `history/path-row.md` · → 진행 `CURRENT.md` · 모양 규칙 `REF-design-chaltteok.md`
> 브랜치 `feat/path-row`. 시안 https://claude.ai/artifact/MoF8UXStkXxLBDg2yUVgej (v4 = 지금 / 안 A(따로 알약) 확정)

## 확정 (사용자 2026-10-28)
| 항목 | 결정 |
|---|---|
| 범위 | 폴더 타일·장비 타일 **경로 줄만**(머리 줄·관리 타일 제외) |
| 모양 | 안 A(따로 알약): 비슷한 기능끼리 옅은 회색 알약 하나(`.ct-grp`), 칸 사이 1px, 묶음 사이 4px, 구분선(`tile-btn-sep`) 없음. 안 B(한 줄 묶음)는 탈락 |
| 색 | 기능 색: 장비 디렉토리 초록(접속 중일 때) · 공유 청록 · 붙여넣기 파랑. 나머지 아이콘 = 중간 강조 회색 |
| 폴더 타일 | `[● 장비 │ 장비 디렉토리 열기]` · `[공유]` · `[붙여넣기 │ 비우기]` · 파란 "＋새로" |
| 장비 타일 | `[N개 │ 내려받기 │ 보내기 │ 해제]` · `[● 상태 │ 터미널]`(점 = 표시만, 옛 점 자리 = "올리기" 앞) · 파란 "올리기" |
| 이동 버튼 | `[뒤로 │ 앞으로 │ 새로고침]` = 경로 **왼쪽**, 타일 왼쪽 벽에 붙음(왼쪽 끝 곧게). 머리 줄(`TileFrame` `#controls`)에서 **뺌**. 칸 폭 42px(양 끝 44px) |
| 장비 이름 | 최대 폭(150px) + `…` (지금처럼) |
| 경로 | 지금처럼 앞쪽 칸부터 `…`로 줄어듦 — **스크롤 안 함** |
| 넘칠 때 | 가운데 묶음만 가로 스크롤(`HScroll`): 막대 숨김 · 넘친 쪽 가장자리 흐림(mask 18px) · 세로 휠 = 가로. 파란 끝 블록은 오른쪽 벽에 고정 |
| 줄바꿈 | 이동 묶음 + 경로 = `flex-basis 270px`가 안 되면 도구가 둘째 줄로(전엔 160px). 둘째 줄 도구는 왼쪽 4px 띄움 |

## 구현 (2026-10-28, Claude 작성)
| 이름 | 파일 | 내용 |
|---|---|---|
| 구현-1(묶음 스타일) | `styles/chaltteok.scss` | `.ct-seg--dim`(글자만 중간 강조) · `.ct-seg--ok`(초록 글자) · `.ct-grp--keep`(꺼진 칸도 바탕 유지, 글자만 흐림) 추가. 나머지는 기존 `.ct-grp`·`.ct-seg` 재사용 |
| 구현-2(가로 스크롤) | `feature/widget/component/HScroll.vue`(새 파일) | 바깥·안쪽 `ResizeObserver`로 흐림 갱신, `@wheel` 세로 -> 가로(넘칠 때만 `preventDefault`) |
| 구현-3(이동 버튼) | `TileNavButtons.vue` · `FolderTile.vue` · `DeviceTile.vue` | `.ct-grp--l` 묶음 + `busy` prop + `refresh` 이벤트. 머리 줄 `#controls`에서 제거 |
| 구현-4(폴더 타일 도구) | `FolderTileBody.vue` | 위 표대로, `.crumbs` 왼쪽 여백 0(이동 묶음이 벽에 붙도록) |
| 구현-5(장비 타일 도구) | `DeviceTileBody.vue` | 위 표대로 |

## 함정 (재발 방지)
- 두 타일 몸통의 버튼 초기화 규칙 `button:where(:not(.v-btn, .tile-btn, ...))`이 **`.ct-seg` 바탕을 지움**(scoped라 자식 컴포넌트 루트만 걸리고, 몸통 안에 직접 쓴 `.ct-seg`엔 걸림) -> 제외 목록에 `.ct-seg` 추가. 몸통 안에 새 버튼 클래스를 쓰면 이 목록부터 볼 것
- 축소 스크린샷에서 1px 칸 경계가 묶음 사이 틈처럼 보임 -> 판정은 2배(`deviceScaleFactor: 2`) 잘라 보기로

## 확인 (격리, 2026-10-28)
- 5051 supervisor(`nexus_member_tmp`) + worker `jig-a`(TMPDIR 짧게: `/tmp/claude-1000/pr-t`, 길면 보관 소켓 경로 초과) + vite 3001, alice/pw. 헤드리스(`pw/rows.mjs`·`pw/busy.mjs`): 폭 1400·700·400·320 / 클립보드·선택 있음 / 휠 스크롤(폴더 sl 0->22, 장비 0->4, 흐림 쪽 바뀜) OK. `vue-tsc` 통과
- 미확인: 실제 폰 터치로 가운데 묶음 끌기(보드 좌우 넘김과 겹치는지) · 어두운 테마
