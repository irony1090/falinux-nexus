# REF — 모바일 가상 키보드·뷰포트 (반복 주제, 프로젝트 무관)

> **다른 프로젝트에서도 같은 문제가 생기므로 프로젝트 고유 내용과 섞지 않는다**(전역 CLAUDE.md "Vault 반복 주제 분리"). 새 프로젝트에선 이 파일 이름 그대로 새로 만들고 아래 지식을 옮겨 쓴다.
> 처음 정리 = Nexus 2026-10-14 파트(모바일 가상 키보드). Nexus 구현 → `REF-mobile-keyboard.md` / 이력 → `history/mobile-viewport.md`(이 주제) · `history/mobile-keyboard.md`(Nexus 측정 경위) / 색인 → `INDEX-mobile-keyboard.md`

## 1. 뷰포트 메타 `interactive-widget` (Android 크롬 계열, iOS는 무시)
| 값 | 키보드가 뜨면 | `position: fixed; bottom: 0` | 쓸 때 |
|---|---|---|---|
| `resizes-visual`(크롬 108+ 기본) | 레이아웃 그대로, visualViewport만 줄어듦 | **키보드 뒤에 가려짐** | 레이아웃을 절대 안 바꾸고 싶을 때. 위치는 `vv.offsetTop + vv.height`로 직접 |
| `resizes-content` | 레이아웃이 키보드 위까지 줄어듦(`innerHeight`·`dvh`·`100%` 모두) | 키보드 바로 위 | **권장**(아래 3의 브라우저 오차가 가장 작음, 웹 터미널들(sessile·term-server)도 이 방식). 레이아웃이 줄어 생기는 재배치는 4로 막음 |
| `overlays-content` | 아무것도 안 줄어듦 | 키보드 뒤 | 가상 키보드 API(`navigator.virtualKeyboard.overlaysContent = true`)와 같은 효과 |

- iOS Safari는 이 메타를 무시: 레이아웃은 그대로, visualViewport만 줄고 **문서를 밀어 올림**(`vv.offsetTop` > 0) -> `bottom: 0`은 키보드 뒤. 위치는 항상 `vv.offsetTop + vv.height`로 잡으면 두 경우 모두 맞음

## 2. 키보드 열림 판정
- "레이아웃 높이 - 보이는 높이"는 resizes-content에선 0이라 못 씀
- **기준 높이 방식**: 키보드가 없을 때의 보이는 높이(`vv.height * vv.scale`, 핀치 확대 제외)를 기억 -> 그보다 120 넘게 줄면 열림, 닫힘일 때마다 기준 갱신
  - 너비가 바뀌면(폴드 펴기·회전) 기준 다시
  - `fullscreenchange` 뒤 약 1초는 기준을 따라감(Edge 전체 화면 끝내기 = 110 줄어듦, 주소창 접힘 = 약 56 — 둘 다 120 미만이지만 여유 확보)
- 실측 키보드 높이(Galaxy Fold 8 접음, dpr 2.6): 약 250 CSS px, 화면 752의 1/3. 120 문턱은 넉넉함

## 3. 브라우저별 실측 (Galaxy Fold 8, 2026-10-14)
| 브라우저 | 키보드 위 끝 계산 | 실제와 차이 |
|---|---|---|
| 크롬(주소창 위, 기본·전체 화면) | `innerHeight - 키보드` | **정확** |
| Edge 태블릿 크기(펼침, 막대 위 고정) | 같음 | 정확 |
| Edge 폰 크기 전체 화면 | 같음 | **약 40 위**(2026-10-20 재측정, 10-14 "정확" 정정) — 레이아웃을 291 줄이는데 키보드는 약 251. 모든 값(vv·inner·dvh/svh/lvh)이 같은 461이라 대체 값 없음 |
| **Edge 폰 크기 기본 화면**(주소창·탐색 줄이 **아래**, 스크롤하면 숨는 막대 111) | resizes-visual = 641 - 250 = 391 | **약 81 위** — 키보드가 뜨면 막대가 얇은 주소 줄(약 28)로 줄어드는데 반영 안 함 |
| 〃 | resizes-content = 레이아웃 442.7 | **약 34 위**(가장 작음) |

- Edge 폰 크기 기본 화면에서 **브라우저가 주는 모든 값이 틀림**: visualViewport, 레이아웃(`innerHeight`·고정 요소), 가상 키보드 API `boundingRect`(391+250), CSS `env(keyboard-inset-top)`(391), `svh/dvh/lvh`(641/641/751, 키보드가 떠도 그대로). 안 보이는 100% 요소 + resize 이벤트도 같은 값을 물려받음
- `lvh - 키보드 - 주소 줄(28)`로 맞출 수는 있으나 주소 줄 높이를 짐작해야 하고 크롬(위 주소창)에는 안 맞음 -> 쓰지 않음
- 같은 증상 제보: 삼성 인터넷 "Bottom address bar causing wrong initial visualviewport height?"(Samsung 개발자 포럼, 주소창 아래일 때만, 공식 해결 없음). 크롬 계열 **하단 주소창** 브라우저 공통 문제로 봄
- 사용자 안내: 그런 브라우저에선 전체 화면(Fullscreen API)이면 정확
- **(2026-10-20 정정)** Edge 폰 전체 화면도 약 40 위 -> 위 표 고침. 빈 띠는 페이지 밖이라 페이지가 그릴 수 없음. "전체 화면이면 정확" 안내는 크롬에만 해당. 측정 → Nexus `REF-board-keyboard.md` "판정 결과"

## 4. resizes-content에서 레이아웃 재배치 막기
- 화면 크기로 배치를 정하는 영역(타일·그리드)은 키보드가 열린 동안 **키보드 전 높이(px)로 고정**하고, 바깥 상자에 `overflow: clip`
  - `hidden`이 아니라 `clip`: hidden은 브라우저가 포커스 요소를 보이게 하려고 스크롤할 수 있음. clip은 스크립트·포커스로도 스크롤 안 됨
  - 높이를 `dvh` 그대로 두면 한 프레임이라도 줄어 ResizeObserver -> 터미널 fit -> PTY 크기 변경이 퍼질 수 있음. 상태(px)로 주면 레이아웃이 먼저 줄지 않음
- 가려지는 부분 처리 두 방식:
  - (옛 Nexus) 영역 전체를 `translateY`로 올림(커서 줄 기준). 크기 불변이지만 화면보다 긴 영역은 윗부분이나 커서 아래가 안 보임. 전환 중 측정은 `getComputedStyle(el).transform`의 `m42`로 지금 이동량을 빼서 계산
  - (2026-10-20 Nexus 현행) **포커스된 입력 영역만 `position: fixed`로 띄워** 가로 자리·폭은 두고 높이 = 보이는 영역 위 끝~키 줄 위 끝. 나머지 배치는 그대로(빠진 grid 칸은 빈 채 남음), 터미널은 rows만 바뀜(cols 그대로라 줄바꿈 재계산 없음). 상세 → Nexus `REF-board-keyboard.md` "맞춤 계획"
- 키보드 위 고정 요소(특수 키 줄)는 transform 걸린 요소 **밖**에 둠(transform 안이면 fixed 기준이 그 요소로 바뀜)
- 다른 화면 영향: 다이얼로그는 줄어든 화면 가운데로 다시 정렬, 전체 화면 다이얼로그는 키보드 위까지 줄어 아래 버튼이 보임(resizes-visual보다 나음)

## 5. 키보드 위 버튼 줄
- 버튼은 `pointerdown`에서 `preventDefault` -> 입력 포커스 유지(빠지면 키보드가 닫힘). `<button>` 대신 `div role=button`
- 수정 키(Ctrl·Alt): 탭 = 다음 글자 하나, 누른 채 = 뗄 때 해제 둘 다 지원. 누르는 중 가상 키보드를 치면 `pointercancel`이 올 수 있어 걸린 채 남겨 다음 글자에 적용 — Android(Fold 8, 삼성 키보드)에서 누른 채 c 동작 확인
- 한글 자판 상태 Ctrl + 자모: 두벌식 자리 -> 영문으로 바꿈(ㅊ -> c). 삼성 키보드에서 바로 나감(조합 지연 없음) 확인

## 5-1. 입력 중 목록을 눌러도 키보드 유지 (2026-10-23 Nexus 실기기 확인)
| 문제 | 입력칸 아래 자동완성 목록을 누르면 입력칸이 blur -> 모바일 키보드가 내려감(다시 올라오며 화면이 출렁임) |
|---|---|
| 방법 | 목록 상자에 `@mousedown.prevent`(= `mousedown`에서 `preventDefault`). 터치 뒤 호환 `mousedown`의 기본 동작(포커스 이동)이 막혀 입력칸이 포커스 유지, `click`은 그대로 옴 -> 고른 뒤 바로 다음 글자 입력 |
| 확인 | 사용자 폰 "자동완성을 눌러도 키보드 안 내려가서 좋더라"(Nexus 파트(경로 직접 입력), `PathInput.vue`) |
| 주의 | 목록의 버튼은 `tabindex="-1"`(Tab이 목록으로 안 감). blur로 닫는 입력칸이면 이게 없으면 목록을 누르기도 전에 닫힘 |

## 6. 확인 방법 (헤드리스)
- resizes-visual·iOS 흉내: `addInitScript`로 `window.visualViewport`를 가짜 EventTarget(높이 = 실제 - 키보드)로 바꾸고 `resize` 이벤트 발행
- resizes-content 흉내: `page.setViewportSize`로 창 높이를 키보드만큼 줄임(레이아웃·보이는 영역이 같이 줄어듦)
- 함정: playwright `keyboard.type('B')`(대문자, shiftKey 없는 합성 keydown)는 실기기와 다르게 동작할 수 있음 -> 명령 키는 소문자로
- 헤드리스로 못 보는 것: 실제 키보드 높이, 브라우저 막대 변화(3의 Edge 문제), IME 조합 -> 실기기 사진으로 확인(Nexus는 개발 서버 사진 업로드 도구 사용 → Nexus `REF-test-env.md` "폰 사진 올리기")
