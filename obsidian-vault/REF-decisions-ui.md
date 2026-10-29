# REF — 확정 결정 상세: 화면(UI) (불변)

> 2026-10-29 `REF-decisions.md`에서 분할(10k 기준, 젤리 효과 ~ 타일 순서 + 이후 화면 결정). → 이력 `history/decisions-ui.md` · 한 줄 요약 = `MEMORY.md` "확정 결정"(같은 이름) · 나머지 결정 = `REF-decisions.md` · → `CURRENT.md`
> 결정을 바꾸면 **이 파일과 MEMORY 한 줄을 같이** 고친다.

## 젤리 효과
- 2026-10-27 파트(젤리·관리 기록). 누를 수 있는 것 hover·누름 + 컨트롤 패널 등장·퇴장 = `styles/jelly.scss` 한 곳(세기 중간). **새 버튼 클래스를 만들면 `$box` 목록에 넣는다.** 알약(`FootPill`) 안은 안쪽만, 넓은 줄은 제외
- 가운데 정렬 팝업에 젤리(scale)를 걸 땐 `transform: translate(-50%,-50%)` 금지 -> `translate: -50% -50%` 속성 + 클래스 `jelly-pop`(열림 = `.shown`). 안 지키면 화면상 오른쪽 아래 기준으로 커짐 → `REF-jelly.md` "함정"

## 입력칸
- 2026-10-27 파트(찰떡 입력칸). 화면의 input·select·textarea·체크 = `CtTextField`·`CtSelect`·`CtAutocomplete`·`CtTextarea`·`CtCheckbox`(Vuetify 감싸기, 속성·슬롯 그대로 넘김) + 공용 `size`(sm 32 / md 40)·`wall`(l/r/lr = 곧은 끝). 모양은 `styles/ct-field.scss` 한 곳. **기본 `<select>` 금지**(목록이 OS 흰 바탕 + 앱 글자색 = 흰 바탕 흰 글씨)
- 함정: v-input은 `flex: 1 1 auto` + 칸 줄 `1fr` -> 세로 flex 상자에서 칸이 늘어나 글자가 위로 쏠림. `.ct-field`에서 막아 둠 → `REF-ct-field.md` "함정"
- 예외: 경로 직접 입력칸(`PathInput.vue`, 자체 자동완성) · 숨은 파일 선택

## 닫은 뒤 active
- 2026-10-28 파트(닫은 뒤 active). active 타일이 닫히면(패널·머리 줄 ×·터미널 종료 뒤·다른 보드로 이동) 위치 번호 이웃 = 같은 번호(다음 타일), 마지막이면 앞 번호를 active + 화면 밖이면 `reveal`(포커스는 안 옮김). 최소화·보드 바꾸기 = 해제만. 보드 = `tileGrids.store` `watch([posOf, hash])`, 관리 = `memberBoard.store` `close` → `REF-close-active.md`

## 패널 창 버튼
- 2026-10-28 파트(패널 창 버튼). Ctrl+Alt 컨트롤 패널(보드 `PagerPopup`·관리 `MemberPagerPopup`) 왼쪽 = active 타일 머리 줄 창 버튼 `[숨기기 | 크기 | 닫기]` 중 그 타일에 있는 것만(비활성으로 자리 차지 X), 동작 = 머리 줄과 같음. 보드는 `TileFrame`이 `node/store/tileHeadActs.store.ts`에 등록 -> 머리 줄 창 버튼 조건·동작을 바꾸면 등록도 같이. 패널엔 위치 글자 없음·도트 자리 최소 3개 폭
- 미해결 1건(사용자 브라우저에서 2번 타일 패널이 빔, 재현 못 함) → `REF-pager-head-btns.md` "미해결"

## 경로 줄 도구
- 2026-10-28 파트(경로 줄 정리). 폴더 타일·장비 타일 경로 줄의 버튼 = 찰떡 묶음 `.ct-grp` > `.ct-seg`(비슷한 기능끼리 한 알약, 묶음 사이 4px, 구분선 없음, 기능 색 = 디렉토리 초록·공유 청록·붙여넣기 파랑). `[뒤로 | 앞으로 | 새로고침]`(`TileNavButtons`)은 경로 왼쪽·타일 왼쪽 벽, 머리 줄엔 두지 않음. 가운데 묶음이 넘치면 `HScroll`(가로 스크롤), 파란 끝 블록은 고정
- 함정: 두 몸통의 `button:where(:not(...))` 초기화 규칙이 몸통 안에 직접 쓴 버튼 클래스의 바탕을 지움 -> 새 버튼 클래스는 제외 목록에 넣는다 → `REF-path-row.md` "함정"

## 타일 추가 주체
- 2026-10-29 파트(영상 Range) 조율. **하나만 있어야 하는 타일(외부 자원에 묶임) = 서버가 트리에 넣음**: 터미널(process uid, 기존 WS5(터미널 타일 추가 주체)·`checkTerminalUIDs`) · 영상 보기(같은 파일 = 계정 전체 하나). 서버 = 계정 트리 전부 잠금(`lockWorkspaces`) 뒤 같은 것 찾기/넣기, PUT은 새로 추가를 거절. 프론트 = `flush` -> API -> `applyServerTile`(adopt와 같은 모양)
- 나머지 타일(이미지 보기·장비·폴더 = 중복 허용)과 배치(끌기·크기·닫기·최소화) = 프론트 `Op` + `PUT /tiles` 그대로. "모든 타일을 서버가 관리" 안은 사용자 질문으로 검토 -> 안 함(낙관 적용·rebase가 그대로 남고 동작이 TS·Go 두 벌이 됨) → `REF-video-view.md` "결정"

## 타일 순서
- 2026-10-29 파트(타일 끌기). 화면 순서 = 문서 `order`(연 곳 `parent`와 분리, 끌어서 자유 이동). 새 타일 = 연 곳 바로 뒤(최신이 앞), 연 곳 없음 = 루트 바로 뒤. 닫기 = 연 곳 한 단계 위로. **parent null = 루트뿐**(터미널 탭 판정) → `REF-tile-order.md`

## 터치 제스처
- 2026-10-29 파트(타일 끌기) 제스처-1(펼치기 패널)·칸끌기-A. 두 손가락 펼치기 = 컨트롤 패널(두 손가락 가운데, 보드·관리 페이지 둘 다) / 빈 곳 길게 누르기(머리 줄 + 경로 줄·목록 빈 곳) = 타일 끌기 / 목록 칸 길게 누르기 = 들림 -> 안 움직이고 떼면 체크, 움직이면 칸 끌기. 이미지 보기 몸통 안 펼치기 = 확대(그 타일은 머리 줄에서만 패널). 마우스 = 머리 줄 빈 곳·칸을 누른 채 6px 움직이면 바로 끌기 → `REF-tile-drag.md` "확정"

## 끌기
- 2026-10-29 파트(타일 끌기). 같은 보드 안 = 그 타일 하나만, 다른 보드로(전체보기) = 연 타일까지 묶음째(가(같은 보드 안끼리 통일)). 끌기 시작 판정 = `DragSession` 한 곳(`common/listener`, → `REF-util-drag.md`), 끄는 동안 = 작은 표시 + 빈 자리(안 1)
- node 칸 순서 = `POST /nodes/order`(X(들)를 Y 뒤로, 서버가 안 보이는 형제까지 1부터 다시 매김, 부모 편집 권한, 루트 = 내 루트끼리만). 칸을 폴더로 옮기면 그 폴더 끝. 폴더 칸 가운데 0.6초 머묾 = 넣기, Ctrl = 복사(넣기에만)
- 화면 끝 띠 = 칸·타일 끌기 공용(`tileGrids.edge*` + `PageEdgeBands`), 48px 안 0.4초 머물면 칼럼 하나씩. 전체보기는 흐르는 스크롤
- 선택 막대처럼 strip 위에 잠깐 끼어드는 막대는 칸 수 판정 높이에서 뺀다(`tileGrids.barH`) — 안 그러면 켜고 끌 때 배치·페이지가 바뀜 → `REF-tile-drag-impl.md`
