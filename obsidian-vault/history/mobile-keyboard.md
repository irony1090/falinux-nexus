# history/mobile-keyboard — 모바일 가상 키보드 구현

> 요약·결정 → `REF-mobile-keyboard.md` / 결정 근거 → `REF-node-ui-overview.md` "모바일 가상 키보드" / 현재 진행 → `CURRENT.md`

## 2026-10-14 — 파트 착수: 코드 상태 확인 + 구조안(키보드-0~5) 제시 (코드 없음)
- vault에 "A+B 확정·반영, iPhone 확인만 남음"으로 적혀 있었으나 **반영 = 결정 문서뿐, 코드 없음** 확인 -> CURRENT·MEMORY 문구 정정
- 구조안 키보드-0(뷰포트 메타)~키보드-5(확인) + 결정 대기 VK1(키 목록)·VK2(표시 조건) 제시
- 사용자 질문 "Ctrl을 누른 채로 안드로이드 가상 키보드에서 c" -> 기기 의존이라 단정 못 함, 탭·누른 채 둘 다 지원 + 한글 자모 -> 영문 변환을 VK3(Ctrl 동작)로 추가
- 사용자 제안 "안 된다면 이용자가 키조합을 추가" -> VK4(사용자 키 추가): 탭 토글로 이미 가능하므로 편의 기능, 키보드-5 확인 뒤 필요하면 키보드-6으로 권장

## 2026-10-14(2) — 키보드-0~4 구현 + 키보드-5 헤드리스 22/22 (커밋 전, 사용자 Android 확인 대기)
- 사용자 "구현부터 시작, 내가 확인이 필요한 부분까지 쭉" -> VK1~VK3 권장안으로 구현. 새 파일 `virtualKeyboard.hook.ts`·`keyModifier.store.ts`·`TerminalKeyBar.vue`, 수정 `index.html`·`ProcessTerminal.vue`·`processTerm.store.ts`·`TileWorkspace.vue`·`App.vue`
- 격리(임시 DB `nexus_kb_tmp` + 5051 + worker `kb-dev` + 3001)에서 가짜 visualViewport로 22/22. 스크립트 실수 2번(화면 밖 머리 줄 탭, 공유 대상 id를 세션 응답에서 찾음 -> `/users/lookup`) 고친 뒤 통과. 정리·DB 삭제
- 스크린샷의 색 있는 화살표 -> `font-variant-emoji: text` 넣어 봤으나 그대로라 되돌림(이 PC 글꼴 문제로 추정, 실기기 확인 목록 2번)

## 2026-10-14(3) — 첫 실기기 화면(Galaxy Fold 8 커버 화면, 칼럼 모드 htop)
- 사용자 스크린샷 `user/vrKeyboard/photo_2026-10-14_14-01-37.jpg`
- 정상: 키 줄 표시·모양(화살표도 흰 글자로 정상 -> 색 화살표는 PC 글꼴 문제 확정)
- 이상-가(키 줄 뜸): 키 줄이 키보드 위 끝보다 약 70 CSS px 위에 있고 그 사이로 htop 줄이 보임. 키보드 위에 `test1.falinux.com` 줄(브라우저 하단 주소창으로 추정)이 있음. vv.height가 실제 보이는 영역보다 작게 나오는 것으로 추정, 원인 미확정
- 이상-나(아래 가려짐): 칼럼 모드 타일이 화면 전체 높이라 "타일 위 끝은 헤더 아래" 규칙 때문에 이동 0 -> 타일 아래 줄이 키보드에 가려짐(셸 프롬프트가 맨 아래면 입력 줄이 안 보임). 설계 한계
- 제안: 고침-가(측정 표시: 임시로 visualViewport 값 화면 표시 -> 원인 확정 후 제거), 고침-나(커서 기준 이동: 커서 줄이 키 줄 위에 오게, 헤더 넘어가도 허용). 사용자 답 대기
- 사진 두 장 추가(`user/vrKeyboard/photo_2026-10-14_14-17-14.jpg` 기본, `..._14-17-15.jpg` 전체 화면 버튼 켬)
  - 전체 화면: 키보드 위 `test1.falinux.com` 줄이 사라지고 **키 줄이 키보드 바로 위에 정확히 붙음** -> 이상-가(키 줄 뜸)는 브라우저 하단 줄(주소창 추정)이 있을 때만 생김
  - 기본: 키 줄 아래 빈틈 약 85 CSS px(사진 2.5px = 1 CSS px 기준), 하단 줄 높이 약 30 CSS px. 85 = 주소창 원래 높이 56 + 줄 30으로 맞아떨어짐(추정)
  - 이상-나(아래 가려짐)는 두 모드 모두 그대로
- 사진 두 장 더(`user/vrKeyboard/photo_2026-10-14_14-21-37.jpg` 전체 화면·키보드 전, `..._14-21-38.jpg` 기본·키보드 전)
  - 브라우저 = **Microsoft Edge Android**(주소창 Copilot 아이콘 + 하단 주소창 + 하단 탐색 줄 back/forward/+/탭 수/메뉴)
  - 키보드 전 하단 브라우저 UI = 주소창 + 탐색 줄 약 116 CSS px, 키보드가 뜨면 약 30 CSS px 줄로 줄어듦
  - 빈틈 85 ≈ 116 - 30 -> **Edge가 하단 UI가 줄어든 것을 visualViewport에 반영하지 않는다**는 가설이 더 잘 맞음(앞의 56 + 30 추정 대체)
  - 별건(미룸): 전체 화면을 켜면 같은 칼럼이 `single` -> `row`(빈 칸 하나 추가)로 바뀜. 화면 높이에 따른 배치 판정으로 추정, 키보드 파트와 무관 -> 2026-10-15 버그 아님(2026-10-15 확인): 전체 화면으로 Grid 영역 높이가 `TILE_THRESHOLD.h` 560 이상 -> `capacity` y 1 -> 2 -> `effectiveSize`가 저장값 h 0.5를 그대로 써서 `single` -> `row`(1칸 축의 0.5를 1로 보는 규칙, `tileSplit.util.ts`)
- 크롬(Android, 상단 주소창) 사진 4장 `user/vrKeyboard/photo_2026-10-14_14-29-32.jpg`(전체 화면·키보드 전) `..._14-29-33.jpg`(기본·키보드 전) `..._14-29-33_2.jpg`(전체 화면·키보드) `..._14-29-34.jpg`(기본·키보드)
  - **두 모드 모두 키 줄이 키보드 바로 위에 정확히 붙음** -> 이상-가(키 줄 뜸)는 **Edge(하단 주소창) 한정**. 크롬 하단 주소창 설정에서도 생기는지는 미확인
  - 이상-나(아래 가려짐) 그대로: 타일은 위 여유(약 8 CSS px)만큼만 올라가고 아래 줄은 가려짐

## 2026-10-14(4) — 고침-나(커서 기준 이동) 구현 + 헤드리스 10/10·22/22 (고침-가는 Edge 확인 때)
- 사용자 "그대로 진행" -> 고침-나 먼저, 고침-가(측정 표시)는 그다음 Edge에서만
- `TileWorkspace.vue` measure 변경 + 커서 이동 구독. 상세 -> `REF-mobile-keyboard.md` 구현표 "고침-나"·결과 절
- 시험 실수: 첫 실행에서 터미널이 화면 밖 칼럼에 있었음(`[data-term-uid]` 상자는 화면 안으로 나옴 -> `.xterm-screen` 상자로 판정 + "다음 칼럼" 버튼). 대문자 키 입력으로 칼럼이 바뀌는 헤드리스 현상 발견(REF "헤드리스 함정")
- 격리 다시 띄움(`nexus_kb_tmp`) -> 끝나고 정리·DB 삭제

## 2026-10-14(5) — 사용자 폰 확인(사진 14장) + 고침-가(측정 표시) 넣음
- 사진 `user/vrKeyboard/photo_2026-10-14_14-58-08.jpg` ~ `..._14-58-18.jpg`(폴드 접음, Edge·크롬, 기본·전체 화면, bash·htop) + `..._15-02-20.jpg`·`..._15-02-20_2.jpg`(폴드 펼침 = 태블릿 크기, Edge)
- 고침-나(커서 기준 이동): 모든 사진에서 커서 줄(bash 프롬프트·htop 아래 줄)이 키 줄 바로 위에 보임 -> 동작 확인
- 이상-가(키 줄 뜸): **Edge + 폰 크기(접음)에서만**. 사용자 분석: 폰 크기 Edge는 주소창·탐색 줄이 하단에 있고 스크롤하면 숨겨지는 방식, 태블릿 크기(펼침)는 상단 고정이라 숨겨지지 않음 -> 태블릿 크기 Edge는 정상. 키 줄 아래 빈틈엔 올라간 타일 영역 밑의 빈 공간이 보임
- 고침-가: 임시 `process/component/KbDebug.vue`(TileWorkspace에 마운트, 커밋 전 제거) — 보이는 영역 위에 노란 3줄: `ih oh ch sh ah dpr` / `vvH vvT vvS pageT scrY` / `fixedBottom`(bottom:0 고정 요소 아래 끝) `vk`(VirtualKeyboard API) `kb`. 0.3초마다 갱신
- 측정 사진 8장 `user/vrKeyboard/photo_2026-10-14_15-45-38.jpg`~`..._15-45-44_2.jpg`(폰 크기). 수치(CSS px, dpr 2.6, sh 752):

| 경우 | ih(닫힘) | vvH(열림) | ih - vvH | 키 줄 |
|---|---|---|---|---|
| 크롬 전체 화면 | 711 | 461 | 250 | 정상 |
| 크롬 기본(상단 주소창) | 655 | 405 | 250 | 정상 |
| Edge 전체 화면 | 751 | 500.6 | 250.6 | 정상 |
| Edge 기본(하단 막대) | 641 | 390.5 | 250.6 | 약 81 뜸 |

  - 모든 경우 키보드 높이 ≈ 250으로 같고, `ih`·`fixedBottom`은 키보드가 열려도 그대로(resizes-visual 동작). `vvT`·`pageT`·`scrY` 모두 0, `vk`(VirtualKeyboard API) 있음, overlaysContent 꺼져 있어 0+0
  - **결론**: Edge는 vvH = ih - 키보드 높이로 계산. 그러나 키보드가 뜨면 하단 막대(sh - ih = 111)가 얇은 줄(약 30)로 줄어 실제 보이는 아래 끝은 그만큼(약 81) 더 아래. 크롬은 하단에 막대가 없으니 같은 식이 맞음
- 사용자 "가 실험" -> `KbDebug.vue`에 안 가(가상 키보드 API 실험) 추가: 노란 표시 끝 `VK on/off` 버튼(localStorage `kbdebug.vk`, 누르면 새로고침) -> 켜면 `navigator.virtualKeyboard.overlaysContent = true`. 4째 줄 `VK ON/off overlays envTop envH`(CSS `env(keyboard-inset-*)` 측정 요소), **빨간 선 = boundingRect 위 끝, 초록 선 = visualViewport 아래 끝**. VK 켜면 화면 영역이 안 줄어 우리 훅은 키보드를 못 알아챔(키 줄 안 나옴, 실험 중엔 정상)
- test1 헤드리스(로그인 폼 + `.env.development.local` 주석의 계정)로 버튼 켬·끔 확인 후 끔 상태로 되돌림
- 안 가(가상 키보드 API 실험) 결과: 사진 `user/vrKeyboard/photo_2026-10-14_15-57-02.jpg`·`..._15-57-14.jpg`(둘 다 Edge 폰 크기, VK ON, 키보드 열림)
  - `overlaysContent true`, `boundingRect` = 391 + 250, CSS `env(keyboard-inset-top)` = 391 -> **visualViewport와 같은 값**(641 - 250). 빨간 선이 실제 키보드 위 끝(주소 줄 위 약 472)보다 약 81 위 -> API도 같은 오류, **안 가 실패**
  - 결론: Edge 폰 크기 기본 화면에서는 브라우저가 주는 어떤 값도 줄어든 하단 막대를 반영하지 않음 -> 안 나(Edge 제약으로 기록, 전체 화면 버튼 안내) 권장
- 사용자 Android 확인 목록 결과: 3(Ctrl 탭 -> c) OK · 4(Ctrl 누른 채 c) OK · 5(한글 자판 Ctrl -> ㅊ) 바로 종료 OK · 6(vim·htop 방향키·Esc) OK · 7(한글 입력) 한 번만 들어감 OK. 1·2는 앞 사진들로 확인(Edge 폰 크기 기본 화면 제외)
- 사용자 질문 "키보드가 나온다는 걸 알면 fixed 등으로 특수 키 줄을 하단에 붙이면?" -> resizes-visual에선 `bottom: 0`이 레이아웃 아래 끝(641, 측정값 fixedBottom이 키보드 열려도 641)이라 키보드 뒤에 가려짐. 붙이려면 resizes-content(레이아웃이 키보드만큼 줄어듦)여야 하고 Edge가 그때 맞는 높이로 줄이는지는 미확인 -> 안 라(resizes-content 실험) 제안
- 사용자 "진행" -> 안 라(resizes-content 실험): `KbDebug.vue`에 `RC on/off` 버튼(localStorage `kbdebug.rc`, 켜면 viewport 메타 `resizes-visual` -> `resizes-content`로 바꿔 새로고침) + **파란 선 = `position: fixed; bottom: 0`**(기존 probe를 4px 선으로). 표시 줄 `RC ON/off meta interactive-widget=...` 추가. test1 헤드리스로 켬·끔 확인 후 끔으로 되돌림
- 안 라(resizes-content 실험) 결과: 사진 `user/vrKeyboard/photo_2026-10-14_16-08-17.jpg`(Edge 폰 크기, RC ON, 키보드 열림)
  - 레이아웃이 줄어듦: `ih` 641 -> 442.7, 파란 선(`bottom: 0`)·초록 선 = 442.7. 실제 주소 줄 위 끝 약 478 -> **여전히 약 34 뜸**(81보다는 줄었음). 타일 재배치(Grid 정보 줄이 키보드 위로 올라옴) 발생
  - 결론: resizes-content도 Edge에서 딱 맞지 않음 + 재배치 비용 -> **안 라 실패**, 안 나(Edge 제약 기록) 권장
- 사용자 질문 "fixed bottom 0이면 걸쳐지지 않나" / "안 보이는 100% 태그 + resize 이벤트로 위치를 알 수 없나" -> 둘 다 레이아웃 뷰포트 기준이라 이미 잰 값(파란 선 641 / 442.7)과 같음. 다만 CSS 단위 `svh·dvh·lvh`는 안 재 봤음 -> 사용자 "재 보자" -> `KbDebug.vue`에 `100svh/100dvh/100lvh` 높이 줄 추가(fixed top 0, visibility hidden). test1 헤드리스에서 표시 확인
- svh·dvh·lvh 결과: 사진 `user/vrKeyboard/photo_2026-10-14_16-22-46.jpg`·`..._16-22-47.jpg`(Edge 폰 크기, VK·RC 끔, 키보드 닫힘·열림)
  - 닫힘·열림 모두 `svh 641.1 dvh 641.1 lvh 751.2` -> **키보드가 떠도 세 단위 모두 그대로**, 단위 방식 실패
  - 참고: 실제 키보드 위 끝(주소 줄 아래) ≈ lvh - 키보드 높이(751.2 - 250.6 = 500.6), 주소 줄 약 28 -> 보이는 아래 끝 ≈ 472. 계산식으로 맞출 수는 있으나 주소 줄 높이를 짐작해야 하고 크롬(상단 주소창)에는 안 맞는 Edge 전용 식 = 안 다(Edge 전용 보정)와 같은 문제
- 사용자 "비슷한 사례 해결 자료?" -> 웹 조사 결과
  - Edge 하단 막대 + 키보드 오차를 직접 다룬 글은 못 찾음. 가장 가까운 것 = Samsung 개발자 포럼 "Bottom address bar causing wrong initial visualviewport height?"(삼성 인터넷, 주소창 아래일 때만 vv 높이가 틀림, 확대·스크롤 뒤 되돌리는 꼼수만, 스크롤 안 되는 페이지엔 못 씀)
  - 다른 웹 터미널 키 줄: sessile PR #114·term-server PR #194 = resizes-content + "가장 컸던 높이"를 기준값으로 150px 넘게 줄면 키보드 열림(너비 바뀌면 기준값 다시), codeg PR #647·knowledge-medium PR #1128 = visualViewport 기준(우리 지금 방식). 하단 막대 브라우저를 따로 다룬 곳 없음
  - Chromium 493412612(키보드 API 높이가 하단 시스템 inset만큼 큼, M152 병합)는 비슷한 결이나 같은 버그인지는 불명

## 2026-10-14(6) — 라(resizes-content 전환) 구현 + 헤드리스 12/12·22/22·10/10 (폰 확인 대기)
- 사용자 질문 "라로 바꾸면 타일 배치가 바뀌나" -> 라-2(높이 고정) 없으면 바뀜(Grid 수용량 = 영역 크기). "Dialog 입력 시 UI가 많이 깨지나" -> 작은 다이얼로그는 다시 정렬, 스크립트 편집은 버튼이 보이게 돼 오히려 나아짐 -> 사용자 "라로 진행"
- 구현·결과 -> `REF-mobile-keyboard.md` "라(resizes-content 전환)". `KbDebug.vue`는 폰 확인용 두 줄(ih·vvH·kb·baseH)로 줄임(VK·RC 버튼 제거), 커밋 전 제거
- 시험 실수: 스크립트가 폴더 `k` 안이라 행이 안 보였음 -> `button.nm`(폴더 이름) 눌러 들어감, 메뉴 = `button[title="메뉴"]`
- 폰 확인(사진 `user/vrKeyboard/photo_2026-10-14_16-48-37.jpg`~`..._16-48-39.jpg` Edge, `..._16-49-42.jpg`~`..._16-49-44.jpg` 크롬, 접음)
  - 크롬: 키 줄이 키보드 바로 위 · 이름 변경 다이얼로그 보임 · 내용 편집 닫기·저장이 키보드 바로 위 · 닫으면 제자리
  - Edge: 키 줄 아래 빈틈 약 34(예상대로, 81에서 줄어듦) · 이름 변경 보임(Edge 자동 완성 줄이 추가로 뜨면 ih 394로 더 줄어도 맞춤) · 내용 편집 버튼 보임 · 닫으면 `Grid 2 single · 1타일` 그대로
  - 측정 표시: Edge `kb OPEN 0..442.7 baseH 641.1`, 닫힘 `closed 0..641.1` -> 라-3(열림 판정) 정상

## 2026-10-14(7) — 파트(모바일 가상 키보드) 마감·커밋
- 사용자 "사진 업로드 도구는 주석 처리, 마감, vault 정리와 커밋까지" + "가상 키보드 이슈는 별도 색인, 다른 프로젝트에서도 생길 수 있어 중요도 높음"
- `KbDebug.vue` 삭제 / 업로드 도구 = `vite.config.mts` import·plugins 두 줄 주석(파일 `dev-upload.plugin.mts`는 남김, 토큰은 `.env.development.local`) — 끈 상태에서 `/__upload` = 앱 화면, POST 404 확인
- 반복 주제 분리: `REF-mobile-viewport.md`·`history/mobile-viewport.md` 신설 + 전역 CLAUDE.md 지정 주제 추가 / 파트 색인 `INDEX-mobile-keyboard.md`(MEMORY·HISTORY에 중요 표시 한 줄)
- 사진(user vault `user/vrKeyboard/photo_2026-10-14_*` 약 35장)은 용량 때문에 커밋 안 함(저장소 폴더라 재부팅해도 남음)
- 마감 뒤: 사용자가 사진 44장을 `user/vrKeyboard/`로 옮김 -> vault 경로 갱신. 사용자 "user vault 자체를 깃 이그노어" -> 루트 `.gitignore`에 `obsidian-vault/user/` + 이미 추적되던 `user/links.md`·`user/kill-test-scripts.md`는 추적만 해제(파일은 그대로)
