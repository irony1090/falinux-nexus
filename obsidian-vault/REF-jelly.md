# REF — 젤리 효과 · 관리 페이지 기록 (파트(젤리·관리 기록))

→ 이력 `history/jelly.md` · URL 규칙 본문 `REF-member-ui.md` "URL 저장"(주소-1(진입만 push)·주소-7(보드로)) · 현재 `CURRENT.md`

**상태: 마감(2026-10-27) — 사용자 확인 OK, `feat/jelly-nav` 커밋 -> master `--no-ff` 머지. 색인 = `INDEX-jelly.md`**

## 요청 3개 (2026-10-27)
| 이름 | 결정 | 코드 |
|---|---|---|
| ①보드로 | 관리 페이지 로고 메뉴 '보드로' = 원래 보던 보드 주소 전체(`focus` 활성 타일·`view` 보던 칼럼 포함) | `router/lastBoard.ts` · `MemberMenu.vue` |
| ②관리 기록 | 관리 페이지(내 장비·관리자)는 들어올 때만 push, 안에서 바뀌는 URL·두 페이지 사이 이동은 replace | `memberUrl.hook.ts` · `MemberMenu.vue` |
| ③젤리 효과 | 누를 수 있는 것 hover·누름 + 컨트롤 패널 나타남·사라짐에 젤리 효과 | `styles/jelly.scss` · `PagerPopup.vue` · `MemberPagerPopup.vue` |

## ③젤리 효과 — 확정 값 (세기 = 중간)
- 시안 아티팩트: https://claude.ai/artifact/Gvj3eoa77tsppmWHsXEG9n (세기 지금·약하게·중간·통통 비교, v3). 사용자: 통통 고른 뒤 "중간이 낫다"로 바꿈

| 동작 | 값 |
|---|---|
| hover | `scale` 1 -> 1.08 -> 0.98 -> 1, 380ms ease-out (커진 채 머물지 않음) |
| 누름 | `transform: scale(0.92)` 70ms |
| 뗌 | 380ms `cubic-bezier(.3, 2, .5, 1)` — 1을 넘쳤다 돌아와서 튀는 느낌 |
| 패널 나타남 | opacity 0·0.6 -> 1.08 -> 0.97 -> 1, 380ms |
| 패널 사라짐 | 1 -> 1.05 -> opacity 0·0.6, 270ms. 그동안 `visibility`를 늦춰 보이게 |

## ③적용 범위
| 대상 | 방식 |
|---|---|
| `.v-btn` · `.tile-btn` · `.mb-btn` · `.ct-seg` · `.dg-btn` · `.dg-lbl` (button·a만, disabled 제외) | 상자째 |
| 목록 아이콘 보기 칸의 `dg-btn`·누를 수 있는 `dg-lbl` | 상자째, **붙은 벽 쪽이 기준점**(`transform-origin`) — 칸이 자르지 않게 |
| 푸터 알약(`.FootPill`) 안 버튼 | **안쪽만**(자식 요소·`.v-btn__content`) — 알약 끝이 버튼을 자르므로. 페이지 도트 `.dot`은 위치에 따라 크기(transform)를 직접 바꿔서 누름 제외, hover만 |
| 컨트롤 패널 `PagerPopup`·`MemberPagerPopup` | 클래스 `jelly-pop`, 열림 = `.shown`. 가운데 = 마우스(화면 밖이면 안쪽 최대점), 커지는 기준점 = 패널 가운데 |
| 넓은 줄(`v-list-item`·목록 줄 `mb-row`) | **제외**(사용자 확정) — 배경색만 |
| 폰(`hover: none`) | 누름만 |
| `prefers-reduced-motion: reduce` | 전부 끔 |

- 묶음(`ct-grp`·`mb-grp`·`tile-btn-end`) 안 칸은 상자째: 묶음이 `overflow: hidden`이라 바깥 끝은 잘리고 안쪽으로만 커진다. 칸 글자는 대부분 텍스트 노드라 "안쪽만"이 안 됨(템플릿 110곳을 감싸야 함) -> 상자째로 둠
- 새 버튼 클래스를 만들면 `jelly.scss` `$box` 목록에 넣는다

## 함정 (재발 방지)
- **hover는 `scale` 속성, 누름은 `transform`** — 한 속성에 두면 hover 애니메이션이 도는 동안 누름 transform이 묻힌다(애니메이션이 transition보다 우선)
- **`transform: translate(-50%, -50%)` + `scale`을 같이 쓰면 scale이 옮기기 전 상자 가운데 기준 = 화면상 오른쪽 아래에서 커진다.** 개별 속성 순서가 translate -> rotate -> scale -> transform이라 transform(옮기기)이 먼저 적용된 뒤 scale이 원래 상자 기준으로 걸림. -> `translate: -50% -50%` 속성으로 옮긴다(시안 v2에서 사용자가 발견)
- 인라인 `transform`이 있는 요소(페이지 도트)에 transform 누름을 걸면 인라인이 이겨서 안 먹고, transition만 붙어 크기 변화가 튀게 된다 -> 제외

## ①②구현 메모
- 마지막 보드 = `router.afterEach`에서 `meta.board`인 이동마다 `nexus.user.board.last`에 `fullPath`(보드 URL은 보는 중에도 `focus`·`view`가 replace로 바뀜). 로그아웃 때 `clearUserPrefs`로 같이 지워짐
- 직전 기록 판단 = `router.options.history.state.back`(vue-router가 history.state에 둠). replace는 `back`을 유지하므로 관리 페이지 안에서 몇 번 바뀌어도 보드를 가리킴
- 확인(2026-10-27, test1): 보드 -> 내 장비 = 기록 +1, 필터·줄 수 변경 = 그대로, 보드로 = back으로 `view` 그대로 복귀(기록 +0) / 주소창으로 `/devices` 직접 -> 보드로 = 마지막 보드 push
