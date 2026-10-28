# REF — 파트(찰떡 입력칸): 입력칸·select·체크박스를 찰떡 공용 컴포넌트로
> → `history/ct-field.md` · → `CURRENT.md` · 규칙 = `REF-design-chaltteok.md` · 출발 구상 = `REF-ui-chaltteok.md` "찰떡 입력칸 구상"(+ 범위 조율 줄)

브랜치 `feat/ct-field`(master `d5b4963`에서, 2026-10-27).

## 범위 (2026-10-27 사용자)
| 넣음 | 뺌 |
|---|---|
| Vuetify 입력칸 6개 · 직접 만든 `.mb-input` 16개(로그인·회원가입 포함) · 기본 `<select>` 3개 · API 키 보기 칸 · 체크박스(`DeviceMoveDialog`) | 경로 직접 입력칸(`PathInput.vue`) · 숨은 파일 선택(`DeviceTileBody`, 보이는 모양 없음) |

## 단계
| 단계 | 내용 | 상태 |
|---|---|---|
| F1(조각) | `CtTextField` + `styles/ct-field.scss`. 끝 모양 = 없음·`wall` `l`·`r`·`lr`·버튼 묶음 안 칸. 높이 sm 32 / md 40, 포커스 표시가 곧은 끝을 따름 | 구현·확인(메시지 줄 높이·묶음 메시지까지) |
| F2(나머지 컴포넌트) | `CtSelect`·`CtAutocomplete`·`CtTextarea`·`CtCheckbox` | 마감(사용자 OK, 여러 줄 = 16px 상자·체크 = 기본 네모 그대로) |
| F3(옮기기) | 전부 교체(창 묶음 -> 관리·로그인 페이지 묶음). autocomplete = 장비 주인 고르기만 | 구현·확인 |
| F4(정리) | `member.scss`에서 `.mb-input`·`.mb-select`(+`.chev`)·`.mb-hint` 지움(32줄), 임시 `CtLab.vue`·`/ct-lab` 라우트 지움 | 구현·빌드 통과 |

## 추가 항목 (2026-10-27 사용자, 진행 중 받음)
| 항목 | 내용 | 상태 |
|---|---|---|
| 방식 묶음 위치 | supervisor 주소 변경 창의 [저장만\|바로 옮기기]가 왼쪽 정렬로 보임 -> 사용자 "오른쪽" -> `.mode { align-self: flex-end }` | 구현·확인 |
| 방식·호스트 칸 붙이기 | 방식(ws/wss) 칸과 호스트:포트 칸이 닿는 끝을 곧게(한 덩어리). 자리 = `SupervisorMoveDialog` · `DeviceMoveDialog` 둘 다 같은 모양 | F3에서 |
| 관리 타일 포커스 선 잘림 | 장비·관리자 페이지 타일이 포커스됐을 때 파란 선이 화면 위아래(아마 좌우도)에서 잘림. 원인 = active 바깥 선(타일 밖 3px)인데 `MemberPage .strip`(`overflow-y: hidden`)이 위·아래·왼쪽 여백 0. 보드 `TileStrip`은 사방 padding 8 -> 사용자 "추천대로": `.strip { margin: -4px 0 0 -4px; padding: 4px 0 4px 4px; scroll-padding-left: 4px }`(위·왼쪽은 간격 안으로 넓혀 타일 자리 그대로, 아래만 4px 비움 = 타일 4px 낮아짐) + `viewW` = clientWidth − padding-left(칼럼 폭 계산 유지) + 폰은 margin·padding 0(Grid gap이 선 자리) | 구현·헤드리스 확인(PC 타일 432,8 / strip 428,4 / grid 아래 812·strip 816, 폰 strip 0) |

## 함정
- **(재발 방지) v-input은 `flex: 1 1 auto` + 칸 줄 `grid-template-rows: 1fr auto`** -> 세로 flex 상자에 남는 높이가 있으면 칸(`.v-field`)이 늘어나고 글자 칸(32)은 위에 붙어 "글자가 위로 쏠림". `.ct-field { grid-template-rows: auto auto; align-content: start }`로 칸은 내용 높이, 남는 높이는 칸 아래 빈 곳. 2026-10-27 실험 페이지(세로 flex·min-height 100vh)에서 발견. 디버거(Claude in Chrome)가 붙으면 화면 크기가 바뀌어 남는 높이가 없어져 정상으로 보였음 = 디버깅 중엔 재현 안 됨 -> **화면에 크기를 찍는 줄(probe)을 넣어 사용자 화면 값으로 잡았다**
- Vuetify `readonly` = native readonly -> Chromium 키보드 커서 안 움직임. API 키 보기 칸은 입력만 막는 방식 유지(→ `REF-ui-chaltteok.md` "함정")

## F1(조각) 구현·확인 (2026-10-27)
| 파일 | 내용 |
|---|---|
| `feature/widget/component/CtTextField.vue` | `v-text-field` + `solo-filled`·`flat`·`compact`·`hide-details="auto"` 기본값, 속성·슬롯 그대로 넘김(넘긴 속성이 기본값보다 우선). `size` sm/md · `wall` l/r/lr |
| `styles/ct-field.scss`(main.ts에서 import) | 둥글기 = `--ct-r-l`/`--ct-r-r`(wall이 0으로), 테두리 = `.v-field`에 inset box-shadow 1px(포커스 primary·에러 error) -> 곧은 끝도 테두리가 따라감. overlay 숨김. sm = control 32px·안쪽 padding 0. `.ct-join` = 옆 칸 붙이기(사이 1px, `.ct-grp` 높이 32) |
| `NodeNameDialog.vue` | 시험 적용(md + 라벨) |
| 임시 `pages/CtLab.vue` + 라우트 `/ct-lab` | **커밋 전 지움**. light/dark × md·sm·wall·join·에러·꺼짐 |
- 측정: md 40px · sm 32px, wall 둥글기 `0 999 999 0` 등 기대대로. 스크린샷 `scratchpad/ct/lab-focus.png`·`name-both.png`(세션 `20dde201`)
- 사용자 판단 거리: md 40px에서 뜬 라벨이 작고 빽빽함(48px 대안) · `hide-details="auto"`라 에러 글자가 생기면 창 높이가 늘어남(자리 미리 잡기 대안)
- 사용자 판단(2026-10-27): **md = 40px 유지** · **끝 여백 = 반원 경계(반지름)보다 안쪽** -> 반원 끝 sm 20px(반지름 16)·md 24px(반지름 20), 곧은 끝 14px(`--ct-pad-round`/`--ct-pad-flat`). 안쪽 아이콘은 끝 여백 −4px 자리, 글자는 아이콘 옆 6px. 에러 글자 들여쓰기도 같은 여백
- 사용자 "32 높이 글자가 위로 쏠림" -> 헤드리스 측정(한글 잉크): 위 9px · 아래 8px(1px 아래쪽) = 가운데. 어느 칸을 본 건지 사용자 확인 대기
- 2026-10-27 "32px 글자 위로 쏠림" 추적: 사용자 캡처(`user/ct-field/치우침.png`)에선 sm 칸이 38px·글자가 위쪽(위 ~9 / 아래 ~15). **Claude in Chrome이 붙으면(디버거) 32px로 정상** -> 사용자 "디버깅을 시작하면 크기가 잘 잡히고 안 잡으면 내가 말한 대로". 헤드리스 dpr 1·1.2·4, CSS zoom 1.2 전부 재현 안 됨. probe 결과(`user/ct-field/치우침2.png`): light 열 32 정상, dark 열만 커짐(probe가 light 열에만 있어 dark 열에 남는 높이) -> 원인 = 위 "함정" 첫 줄, 고침·다크 헤드리스 확인
- 메시지(2026-10-27 사용자 "message가 input 폭을 넘지 않게"): 원래 넘지 않음(details 폭 = 칸 폭, 접힘). 추천대로 진행 = ① 메시지 줄 높이 `body-small`(Vuetify 12px 고정은 2줄부터 붙음) ② 붙인 묶음은 칸마다 말고 **묶음 폭 아래 한 줄** `.ct-msg`(`--error`) — 칸은 `error hide-details`(테두리만 빨강), `.ct-join`은 `flex-wrap`으로 메시지만 다음 줄. F3에서 두 주소 변경 창에 적용
- 32px 쏠림 고침 사용자 확인 OK("됐다")
- 라벨 위치(2026-10-27 사용자 "라벨이 보더라인에 반쯤 걸치게") -> 안 A(선 끊기, Vuetify `outlined`) 확정 / 안 B(라벨 뒤 바탕 조각)는 안 씀. 구현: `variant="outlined"`(solo-filled·flat·box-shadow 테두리 지움), 선 두께 포커스 때도 1px·평소 opacity .18·hover .32, **포커스 색 primary**(Vuetify 기본은 color 속성 없으면 글자색 = 검정) · 뜬 라벨 12px(`--v-field-label-scale: .8571em`, 기본 .75em = 10.5px) · 선 끊김 시작 = 글자 시작(`.v-field__outline__start` flex-basis = 끝 여백 − 4px, 라벨 좌우 margin 4px) · 바탕 연회색은 `.v-field`에 그대로
- 남은 관찰: 뜬 라벨이 칸 위로 약 6px 나옴 -> 칸을 세로로 쌓을 때 위 칸 에러 글자와 가까워짐(옮길 때 간격 확인)

## F2(나머지 컴포넌트) 구현 (2026-10-27)
| 파일 | 내용 |
|---|---|
| `feature/widget/util/ctField.util.ts` | 공용 `ctFieldProps`(size·wall) + `ctFieldClass()` — CtTextField도 이걸로 바꿈 |
| `CtSelect.vue` · `CtAutocomplete.vue` | outlined·compact·hide-details auto, 펼침 목록 `menu-props.contentClass = 'ct-menu'`(목록은 앱 테마 = 항목 5 흰 글씨 해결) |
| `CtTextarea.vue` | `ct-field--area` = 반원 대신 16px 둥근 상자(여러 줄), wall만 받음(size 없음) |
| `CtCheckbox.vue` | `v-checkbox-btn` primary·compact, 네모 칸만(줄 글자는 부르는 쪽 label) |
| `styles/ct-field.scss` 추가 | 칸 안 라벨 = 글자 크기(Vuetify 16px 기본 -> inherit) · 펼침 화살표 왼쪽 여백 0(88px 방식 칸에 "wss" 들어감) · `.ct-check { flex: none }`(기본은 줄 폭을 차지해 옆 글자를 오른쪽 끝으로 밈) · `.ct-menu` = 16px 둥근 상자 + **사방 8px**(사용자 "균일하게, 8px 정도") 띄운 알약 줄 36px(Vuetify `rounded-0` !important를 덮음), 줄 글자 = md 칸 글자 세로줄(24px). 목록 상자 선·그림자는 그대로(사용자 철회) |
- 확인: 고르기 목록(light)·찾아 고르기 목록(dark, "o" 입력 -> o2) 앱 테마 색, 방식 88px + 호스트 붙임, sm 전부 32px. 스크린샷 `scratchpad/ct/f2*.png`(세션 `20dde201`)

## F3(옮기기) 1차 = 창 묶음 (2026-10-27)
| 파일 | 바꾼 것 |
|---|---|
| `ShareDialog` "아이디" 단계 | `v-text-field` -> `CtTextField`(md, 라벨·힌트 그대로) |
| `NodeDeviceDialog` 장비 키 | -> `CtTextField` |
| `ScriptEditDialog` | `v-textarea` -> `CtTextarea`(scoped `.editor` :deep 높이·monospace 그대로 먹음) |
| `SupervisorMoveDialog` | `.addr`(select + text 따로) -> `.ct-join`: `CtSelect wall=r`(88px) + `CtTextField wall=l` + 묶음 메시지 `.ct-msg--error`(칸은 `error hide-details`) |
| `DeviceMoveDialog` | 기본 `<select>`+`.mb-input` -> `.ct-join`(sm) + 묶음 메시지(아래 `mb-err` 지움) · 체크 = `CtCheckbox`(`.inst` label 안) |
- 확인: 헤드리스 스크린샷 `scratchpad/ct/f3/`(공유 아이디·장비 지정·스크립트·주소 변경 2개 에러 상태), 타입 검사. 포커스 테두리 = primary·opacity 1 측정(스크린샷이 흐려 보인 건 250ms 전환 중 캡처)
- 남은 질문: `DeviceMoveDialog` [저장만|바로 옮기기]도 오른쪽으로? / 공유 대상 찾기 autocomplete = 지금 "목록에서 고르지 않고 아이디를 정확히" 설계(회원 목록 노출 안 함) -> autocomplete면 회원 검색 API·노출 범위 결정 필요
- 사용자 답(2026-10-27): `DeviceMoveDialog` 방식 묶음 = 오른쪽(`.mode { align-self: flex-end }`) / 공유 대상 찾기 = **일반 글자 칸으로 충분**(autocomplete 안 함) / autocomplete = **관리자 페이지 장비 주인 변경**에만

## F3(옮기기) 2차 = 로그인·관리 페이지 (2026-10-27)
| 파일 | 바꾼 것 |
|---|---|
| `Login.vue` · `Signup.vue` | `.mb-input` -> `CtTextField size=sm`(id·name·autocomplete는 Vuetify가 input으로 넘김). 가입 힌트·비밀번호 불일치 = `hint`/`error-messages`(`.col` 감싸기 지움) |
| `MemberNewTile.vue` | 이름·API 키 = `CtTextField sm` + hint/error-messages, 주인(대신 등록) = `CtAutocomplete sm` |
| `MemberDeviceTile.vue` | 주인 = `CtAutocomplete sm wall=r`(`:model-value="d.ownerId"`에 묶어 확인 전엔 원래 값, 목록에 없는 지금 주인도 넣음) · 키 칸 보기 = `CtTextField wall=r` + `@beforeinput.prevent @drop.prevent inputmode=none`(헤드리스: End 뒤 입력해도 47자 그대로) · 가림 = `CtTextField readonly tabindex=-1` · 키 바꾸기 칸 · 인스턴스 라벨(`update:focused` false = 저장, Enter = blur) |
| `MemberList.vue` | 검색 = `CtTextField sm prepend-inner-icon=mdi-magnify`(직접 그린 돋보기·padding 지움) |
- 함정: 감싼 컴포넌트는 emit 타입이 없어 `@update:model-value="v => ..."`가 implicit any -> `(v: string) =>`처럼 타입을 적는다
- 확인: 헤드리스 `scratchpad/ct/f3/`(login·signup·admin-dev·admin-dev-key·admin-new·admin-list·owner-crop), 타입 검사. vue 파일에 `.mb-input`·`.mb-select` 남은 것 없음

## 마감 준비 (2026-10-27)
- `npm run build`(타입 검사 + vite 빌드) 통과. 관리 화면 다시 찍어 F4 뒤 깨짐 없음 확인(`scratchpad/ct/f3/f4.png`)
- 함정(재발 방지): scss 블록을 문자열로 지울 때 마지막 줄이 `{ ... }` 한 줄짜리면 그 줄 끝 `}`를 블록 끝으로 잘못 잡아 닫는 괄호가 남음 -> 빌드 `unmatched "}"`. 지운 뒤 반드시 빌드
- 커밋 범위: 프론트(위 표 전부 + `MemberPage.vue` 포커스 선) + vault. 커밋·머지는 사용자 확인 뒤
