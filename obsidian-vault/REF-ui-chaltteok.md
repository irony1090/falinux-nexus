# REF — 파트(찰떡 마무리): 경로 줄 여백 · 남은 창 찰떡 · 관리자 필터
> → `history/ui-chaltteok.md` · → `CURRENT.md` · 규칙 = `REF-design-chaltteok.md`

## 항목 (2026-10-27 사용자)
| 항목 | 결정 | 상태 |
|---|---|---|
| 1 경로 줄 여백 | `.path > * { min-width: 2.2em }`(앞쪽 칸 "글자 하나 + …"까지만 줄게)은 남기고, **2글자 이하 이름 칸만 `.short`(`flex: none; min-width: 0`)**. 지우면 긴 경로에서 앞 칸이 폭 0까지 줄어 "/ / / 지금"만 남음 | 구현·확인 |
| 2 관리자 필터 | 필터 칩·"한 페이지" [10\|20\|50]에 찰떡. "전체" = 왼쪽 벽, "50" = 오른쪽 벽. 칩을 한 묶음으로 할지 떨어진 알약으로 할지 = **아티팩트로 사용자가 보며 정함** https://claude.ai/artifact/6G4KvdUKYG3CheUwwTkWz7 (지금 / A 한 묶음(추천) / B 떨어진 알약) → 사용자 "전체 왼쪽·50 오른쪽에 4px, 검색칸도 맞춰" → 4px 시안 https://claude.ai/artifact/PYwZDUAKF8Vwwq7F4NMVLv (A′·B′, 줄 좌우 여백 4px, 곧은 끝 유지, "1–20 / 45" 글자는 12px 그대로) → 사용자 "여백 = 검색칸 위 여백과 같게" → 같은 링크 v2: 줄 좌우 10px(`.tools` padding-top과 같음), 끝 모양 곧게/반원 전환 → **사용자 확정: 시안 A′(칩 한 묶음, 줄 좌우 여백 10px = 검색칸 위 여백, 끝 곧게)** | 구현·확인 |
| 5 select 목록 안 보임 | 기본 `<select>` 목록 창 = OS 기본 흰 바탕 + 앱 테마 글자색(흰) → 흰 바탕 흰 글씨. 자리 = 장비 상세·대신 등록 "주인", 열린 장비 주소 변경 "방식". 사용자 제안(2026-10-27): **Vuetify 입력칸을 찰떡 모양 공용 컴포넌트로 감싸기**(autocomplete 필요, 애니메이션·에러 메시지 등 이점) → "진행 말고 고민" → autocomplete 쓸 곳 = **주인 고르기 · 공유 대상 찾기** | **다음 파트(찰떡 입력칸)로 넘김**(2026-10-27, 이 파트는 5 빼고 마감) |
| 4 관리 상세 상태 박스 | 회원 상세(상태·가입·장비 상한·장비 등록·관리자)·장비 상세(상태·주인·등록·API 키) 칸이 난잡 → 정리. 사용자가 고른 원인 = 가(높이·모양 제각각)·나(폭 제각각)·라(줄 간격)·마(찰떡 안 됨), 다(API 키 두 줄)는 안 고름. 시안 https://claude.ai/artifact/TM1f143ViK7TFKBqCdSkha (지금 / 안 1 한 줄 44px + 오른쪽 벽 / 안 2 값 칸 폭 통일 / 안 3 라벨 위·값 아래) → **사용자 "안 2"** + "API 키 보기 때 영역 넘침 신경" → v2(키 칸 줄바꿈) → 사용자 "높이가 두꺼워짐, 보기 때 나오던 NAME·API_KEY env 두 줄이면 충분" → v3: **키 칸은 한 줄 고정(넘치면 …), 보기 = 아래 env 두 줄 상자(지금과 같음, 가로 스크롤)** → 사용자 "보기 때 가린 칸을 input처럼 커서 이동되게" → v4: 보기 = `<input readonly>`(한 줄, 커서·선택으로 끝까지), 숨김 = • 가림 글자. env 상자 만족 → 사용자 "정보 줄 구분선 없는 쪽도 보고 싶음" → 비교 시안 https://claude.ai/artifact/84xRmpVsesJnDF8yoHJavV (구분선 있음/없음) → **사용자 "구분선 없는 쪽으로 진행"** | 구현·확인 |
| 3 남은 창 | 찰떡 안 된 창 전부: `NodeNameDialog`(보드 이름 변경·새 보드·노드 새로 만들기·이름 변경 공용) · `PresenceDialog` · `ScriptEditDialog` · `SupervisorMoveDialog` · `AppDialog`(공용 확인 창) | 구현·확인 |

## 구현 (2026-10-27, 프론트만)
| 파일 | 내용 |
|---|---|
| `FolderTileBody.vue` · `DeviceTileBody.vue` | 경로 줄 버튼에 `.short`(2글자 이하, `[...name].length`) |
| 창 5개 | 카드 `ct-card` + `v-card-actions` → `.ct-actions > .ct-grp--r > .ct-seg`(확인·저장 = `--primary`) |
| `ScriptEditDialog.vue` | "Ctrl+S 저장" 안내는 `.ct-actions` 안 왼쪽(`margin-right: auto; align-self: center`) |
| `SupervisorMoveDialog.vue` | `v-btn-toggle`(저장만/바로 옮기기) → `.ct-grp--sm` 묶음, 고른 칸 = `--primary`(`.mode { align-self: flex-start }`) |
| `AppDialog.vue` | 버튼 종류 → 칸 색: info = 기본 · warning = `ct-seg--warning`(새) · error = `ct-seg--danger`(새, 빨강 바탕) · 종류 없음 = 창 종류. 아래 구분선 지움(위 구분선은 그대로) |
| `styles/chaltteok.scss` | `.ct-seg--warning`(옅은 경고색) · `.ct-seg--danger`(빨강 바탕, 옆 칸 구분선 없음) |

| `MemberList.vue` | 필터 = `.mb-grp--sm.chips`(한 묶음, 왼쪽 끝 곧게) · `.tools` padding 10px(좌우 = 위) · 한 페이지 묶음 오른쪽 끝 곧게 · 범위 글자 `padding-left: 2px`(벽에서 12px) |
| `styles/member.scss` `.mb-fields--wall` | 상세 상태 칸 안 2: 줄 `minmax(44px, auto)`·구분선 없음·오른쪽 여백 0, 값 칸 묶음/선택 = 폭 전부 + 오른쪽 끝 곧게, 상태 태그 = 글자(+점), `.full` = 오른쪽 12px |
| `MemberUserTile.vue` · `MemberDeviceTile.vue` | `mb-fields--wall` · 상한 [− 1/3 +](−·+ 44px) · 장비 상태 점 · API 키 칸 + [보기\|복사\|바꾸기] 한 덩어리 |
- 사용자 확인 뒤 조정(2026-10-27): env 상자가 오른쪽 벽까지 붙음(`.full`에 padding을 줘서 상자 안쪽만 늘었음) → `margin: 2px 16px 6px 0`(좌우 16px 대칭) · API 줄이 "인스턴스" 띠에 붙어 보임 → `.mb-fields--wall` 아래 padding 4px → 10px(마지막 조작 칸 아래 ≈ 위쪽 첫 줄 글자까지 ≈ 16px)
- **함정(재발 방지)**: `<input readonly>`는 Chromium에서 키보드(End·방향키)로 커서가 안 움직임(드래그만 됨). 그래서 보기 때 키 칸 = 일반 `input` + `@beforeinput.prevent`·`@drop.prevent`(고치기만 막음) + `inputmode="none"`(폰 키보드 안 뜸). 헤드리스로 End = 47자 끝·입력 무시 확인

## 확인 (2026-10-27, 격리 5051 + vite 3001, 헤드리스 스크린샷 `scratchpad/dp/ui/`)
- 장비 타일 `/home/jh-bae`: `/` 뒤 빈 여백 없어짐 · 폴더 타일 `기본 / 라인` 그대로
- 이름 변경·삭제 확인(삭제 = 빨강 바탕)·스크립트 편집·접속 상태·supervisor 주소 변경 창: 오른쪽 아래 끝 블록, 주소 변경 방식 묶음 OK
- 관리자 필터·상세 상태 칸(회원 o2·장비 dp-dev, 보기 + env 상자) 스크린샷 `ui/admin2*.png`
- 타입 검사 통과

## 찰떡 입력칸 구상 (2026-10-27, 논의만 — 코드 없음)
지금 입력칸 = Vuetify 6개(창: `v-text-field` 4 · `v-select` 1(주소 변경 방식) · `v-textarea` 1(스크립트 편집)) + 직접 만든 `.mb-input` 16개(관리·내 장비 6개 파일). Vuetify 쪽 이점 = 에러·힌트·라벨 움직임·지우기·**테마 따르는 펼침 목록**·키보드·autocomplete

| 고민 | Claude 생각(사용자 미확정) |
|---|---|
| 변형 | `outlined`는 라벨이 뜰 때 테두리에 홈 → 알약에 어색. **`solo-filled` + `flat`**(연회색 바탕 알약 = 지금 `.mb-input`과 가까움, 한쪽 끝 곧게도 쉬움) |
| 높이 | Vuetify `compact` = 40px, 줄 안 칸 = 32px. `--v-input-control-height`·`--v-field-input-padding-*`(Vuetify 4 `VField.css`에 있음)로 줄일 수 있는지 **조각 하나 먼저 만들어 확인**(안쪽 아이콘·지우기 정렬) |
| 라벨 | sm(32px) = 라벨 칸 밖(`mb-lbl`)·안은 placeholder / md(40~48px) = 창, 라벨이 칸 안에서 뜸 |
| 모양 | `CtTextField`·`CtSelect`·`CtAutocomplete`(필요하면 `CtTextarea`) — Vuetify 속성 그대로 넘김 + `wall`(`l`/`r`, 벽 쪽 끝 곧게) + `size`(`sm`/`md`). 스타일 한 곳 `styles/ct-field.scss`(젤리처럼) |
| 순서 | 조각 하나 → 실제 앱에서 높이·끝 확인 → 옮기기. 키 표시 칸 같은 특수 칸은 안 옮길 수도 |
| 첫 적용 | 항목 5(select 흰 글씨) = `CtSelect`로 / autocomplete = 장비 "주인" 고르기 · 공유 창 대상 찾기 |
- **사용자 답(2026-10-27)**: 지금 파트는 항목 5 빼고 마감 -> 새 파트(찰떡 입력칸) / 범위 = 페이지의 input·select **전부**(Vuetify + 직접 만든 `.mb-input`, 로그인·회원가입 포함) / 순서 = 조각 하나 먼저 만들고 확인
- 범위 조율(2026-10-27): 경로 직접 입력칸(`PathInput.vue`) = **뺌** / API 키 보기 칸 = **넣음**(Vuetify `readonly`도 native readonly라 Chromium 키보드 커서 함정 그대로 -> 입력만 막는 방식 유지) / 끝 모양 = 없음·`l`·`r`·`lr`·버튼 묶음 안 칸 / 체크박스(`DeviceMoveDialog`)·숨은 파일 선택(`DeviceTileBody`) = 설명 뒤 사용자 답 대기
