# history — 찰떡 규칙 (디자인 규칙)

> → `REF-design-chaltteok.md` / `CURRENT.md`

## 2026-10-25
- 사용자: 이번 반원형 스타일(부모 박스 엣지에 붙으면 부모와 같은 형태)을 컨셉으로 별도 정리하고 싶음. 흩어진 기록 위치 = `REF-tile-head.md`(조정-2 정정, C2a·C2b) · `REF-dir-grid.md`(판 D 상태 박스) · `styles/dir-grid.scss` 머리 주석
- 조율: Nexus 디자인 규칙으로, 다른 작업공간에서 직접 참조. 이름 후보(찰떡·말랑·껌딱지·젤리빈·코알라) 중 **찰떡** 확정 → 규칙 = 찰떡 규칙, 요소 = 찰떡 알약. 전역 `CLAUDE.md`에는 안 넣음
- "진행" → 아티팩트 "찰떡 규칙" 게시(https://claude.ai/artifact/JzUCKbGWTm4TL8A3GpqeRo — 규칙 끄기 비교 토글, 세 면, 모서리, 머리 줄·목록 칸 실제 예, CSS, 틀렸던 것, 확인 순서) + 이 REF 신설 + user vault `links.md` 한 줄 + 두 출처 REF에 포인터. 코드 변경 없음
- 사용자: 디렉토리 타일의 공유·장비 설정·터미널 탭 UI + 푸터 보드 선택·알림·전체 화면·로그아웃을 찰떡으로. 조율: 푸터 = 상자(본문 8px 간격, 아래 모서리 약간 둥글게), 터미널 = 탭 안 목록 행, 창 안 부모 벽 = 카드 + 목록 상자 → 시안 게시(https://claude.ai/artifact/1VjAAor4ZJFQU7soo75GKb, 찰떡-1(푸터) 가·나 + PC·폰, 찰떡-2~5 지금/찰떡 비교). 고르기 대기
- 사용자 고름: 푸터 = 가 + 상자 안 보이게(지금처럼), 창 3개 = 마음에 듦, 터미널 = 왼쪽 번호까지 벽. 시안 v2에 반영(같은 링크). 목록 상자 안 보이는 판 요청 → 별도 비교판 게시(https://claude.ai/artifact/FZza17LBRGDtL6XDiBnQjt: 상자 보임 / 다(상자 안 보임) / 라(창 벽까지)), 고르기 대기
- 사용자: 창 벽까지(라)는 아님, 상자 없이 보기·편집·조작·접속 n/m 칸까지 안쪽 반원이 보여야 → 비교판 v2에서 다를 다′로 고침(글자 칸에도 바탕, 칸 사이 흰 1px). 확인 대기
- 사용자: 다′로 확정 → 구현 계획 정리(코드 전)
- 구현 계획 조율: 칸 팝업(목록 칸 "▶ 숫자") 행도 번호를 팝업 왼쪽 벽에(들여쓰기 없앰) · 공유 창 단계 고르기 = 묶음 칸 + 메뉴(`v-select` 없앰) 사용자 OK
- "진행" -> 브랜치 `feat/chaltteok-ui`: 공통 `styles/chaltteok.scss`(`.ct-grp`·`.ct-seg`·`.ct-actions`·`.ct-card`) + `RunRow`·`ShareDialog`·`MySharesDialog`·`NodeDeviceDialog`·`FootPill`(`edge`)·`BoardFooter`·`BoardMenu`. 계획에선 공통 클래스를 `tile-buttons.scss`에 넣으려 했으나 규칙 이름으로 파일을 따로 둠. 공유 창 이어진 행 구분 = 행 점선 대신 아바타 점선(행 구분선이 없어져서). 장비 지정 창 `.roster` 버튼 초기화 규칙에서 `.ct-seg` 제외(안 하면 칸 여백이 지워짐)
- 확인: vue-tsc·빌드 통과, test1(8011 -> test3 API) 헤드리스 PC 1280·폰 390 스크린샷(푸터·터미널 탭 목록·칸 실행 팝업·장비 지정·내가 공유한 것·공유 창(빈 목록)). **실수**: 칸 팝업을 열려던 선택자(`button:has(.mdi-play)`)가 "실행" 버튼에 걸려 `move-tagtown`이 실행됨 + 좌표 클릭(공유 버튼이 늦게 그려져 자리가 밀림)이 "장비 디렉토리 열기"에 걸려 장비 타일이 열림. 교훈: 실제 계정 화면에서는 좌표 클릭 금지, 선택자는 클래스·title로 정확히(`button.dg-lbl:has(.mdi-play)`), 화면이 다 그려진 뒤(`.RootTabs` 대기 + 2초)
- 사용자가 잘못 생긴 타일 정리 + test3에 웹만 적용 요청 -> `VITE_API_URL= npx vite build`로 test3 `build/ctrl_panel/web` 교체(백업 안 둠, 새 번들 서빙 확인), test1 닫음, 파트(찰떡 UI) 커밋·master `--no-ff` 머지
