# REF — 글자 크기 규칙 (전체 화면 글자 키우기)

> 2026-10-23 신설(조율 중, 코드 없음). 이력 → `history/typography.md` / 터치 스크롤 → `INDEX-term-touch.md` / → `CURRENT.md`

## 출발 (BACKLOG에서 올림)
- 원래 항목 "터미널 글자 크기·터치 스크롤 양"(2026-10-23) -> 사용자: 터미널만이 아니라 **전체 화면 글자를 키우고, 커지며 깨지는 UI를 맞춰 가기** + **글자 규칙(단계 이름)을 정하고 중앙에서 한 번에 바꾸기** + **이름은 Vuetify(MD3)에 맞춤**(결정)

## 조사 (2026-10-23)
| 대상 | 지금 |
|---|---|
| 우리 컴포넌트 | px 고정 27개 파일: 11px 20 · 12px 15 · 10.5px 10 · 11.5px 9 · 10px 7 · 13px 6 · 12.5px 5 · 9px 1 |
| 터미널 | `processTerm.store.ts` `fontSize: 12`(xterm, px 숫자) |
| Vuetify 4.1.2 | MD3 단계가 **rem**: title-large 22 · title-medium 16 · title-small 14 · body-large 16 · body-medium 14 · body-small 12 · label-large 14 · label-medium 12 · label-small 11(px, 루트 16 기준). 버튼 = `label-large` 등 컴포넌트가 `$typography`를 참조. 바꾸는 곳 = `styles/settings.scss`의 `@use 'vuetify/settings' with (...)` |
| 우리 scss | `styles/_variables.scss`(전 파일 `vars.`로 주입): spacing·weight 단계 + 안 쓰는 `$base-font`·`fontLevel()` |

- 결론: 우리 화면(10~13px)이 Vuetify 기본(12~16px)보다 한 단계쯤 작음. Vuetify가 rem이라 **루트 글자 크기 하나로 Vuetify 전체가 같이 움직임** -> 우리 px도 단계(rem)로 바꾸면 중앙 하나로 전부

## 착수 결정 (2026-10-23 사용자 답)
| 번호(이름) | 결정 |
|---|---|
| (이름 체계) | **Vuetify(MD3) 단계 이름**(`body-small/medium/large`·`label-*`·`title-*`) |
| Q45(기본 크기) | **Vuetify 기본값 그대로** — 본문 body-medium 14px · 작은 정보 body-small 12px · label-small 11px |
| Q46(터미널 글자) | **같은 배율을 따름** — 터미널도 단계 값을 읽음 |
| Q47(폰·PC) | **같게** — 배율 하나 |
| Q48(터치 스크롤 양) | **나중에 따로**(BACKLOG) — 글자 파트 뒤 줄 높이가 바뀐 상태에서 |

## 큰 계획 (2026-10-23 제안)
| 단계 | 내용 | 손댈 곳 |
|---|---|---|
| 글자-1(중앙 값) | 크기는 Vuetify `$typography`(rem) 그대로 + 루트 배율 손잡이 하나(지금 1). 우리 scss용 `vars.font(이름)` mixin = Vuetify 단계의 크기·줄 높이를 그대로 씀. 안 쓰는 `$base-font`·`fontLevel()` 정리 | `styles/settings.scss`·`_variables.scss` |
| 글자-2(대응표) | 지금 px -> 단계: 9·10·10.5px = label-small(11) · 11·11.5px = body-small(12) · 12·12.5px = body-medium(14) · 13px = 쓰임 보고 title-small(14) 또는 body-large(16). 위계(작은 것 < 본문 < 제목)는 유지 | 표로 먼저 |
| 글자-3(바꾸기) | 27개 파일의 `font-size: Npx` -> `@include vars.font(...)` | `feature/**`·`common/**` |
| 글자-4(터미널) | `fontSize: 12` -> 루트 글자 크기 × body-medium 비율(14px) | `processTerm.store.ts` |
| 글자-5(깨지는 곳 맞추기) | 헤드리스 스크린샷 전후 비교(보드·폴더 타일·장비 타일·터미널·헤더·푸터·다이얼로그·페이지 표시·폰 폭) -> 넘침·잘림·줄바꿈 고치기, 글자 옆 아이콘 크기도 같이 | 해당 컴포넌트 |
| 글자-6(확인) | 스크린샷 + 사용자 브라우저 확인 | scratchpad |

## 글자-1(중앙 값) 결과 (2026-10-23)
| 무엇 | 어디 |
|---|---|
| **단계 값 = 한 곳** `$typography`(size·line-height, Vuetify 4.1.2 기본값) | `styles/_typography.scss` — `settings.scss`가 `vuetify/settings`에 `$typography`로 넘기고, `_variables.scss`의 mixin도 같은 map을 읽음 -> 단계 크기를 여기서 바꾸면 Vuetify 컴포넌트와 우리 화면이 같이 바뀜 |
| 배율 손잡이 `$font-scale`(지금 1) | `styles/_variables.scss` -> `styles/global.scss`가 `html { font-size: 100% × 배율 }`(`main.ts` import). 전체를 더 키우려면 이 값 하나 |
| 단계 mixin `@include vars.font('body-small')` | `styles/_variables.scss` — size·line-height만(weight는 기존 `vars.$weight-*`) |
| 터미널 | `processTerm.store.ts` `termFontSize()` = html font-size × 0.875(body-medium) — 배율을 따름 |

### 함정 (재발 방지)
- **`_variables.scss`(·`_typography.scss`)에서 `vuetify/settings`를 `@use`하면 안 됨**: `additionalData`로 Vuetify 컴포넌트 sass에도 주입되어 `settings.scss`의 `with`보다 먼저 로드 -> `This module was already loaded, so it can't be configured using "with"`. **`vite build`는 통과하고 dev 서버에서만** 남(2026-10-23 실제로 겪음)
- mixin이 line-height도 넣으므로, 따로 쓴 `line-height`는 include **뒤**에 있어야 이김(지금 2곳: `ActiveBadges` 1 · `ScriptEditDialog` textarea 1.5)
- scoped `button { padding:0; font:inherit; background:none }` 초기화는 **자식 `v-btn` 루트에도 걸린다** -> x-small 버튼 여백·tonal 색이 사라져 있었음(글자가 커지며 드러남). `button:where(:not(.v-btn))`로 좁힘(`:not()`만 쓰면 우선순위가 올라 `.crumbs button` 색·`button.runs` 배지가 지워짐) — `FolderTileBody`·`DeviceTileBody`·`NodeDeviceDialog`

## 글자-2(대응표) 확정 (2026-10-23 사용자 "진행")
| 지금 | 단계 |
|---|---|
| 9 · 10 · 10.5px | label-small(11) — `ActiveBadges` 10px 배지·9px `.more`가 같은 크기가 됨(둘 다 배지라 허용) |
| 11 · 11.5px | body-small(12) |
| 12 · 12.5px | body-medium(14) |
| 13px | 다이얼로그 본문(`MySharesDialog`·`PresenceDialog`)·스크립트 편집 textarea = body-medium(14) / `TerminalKeyBar .key` = label-large(14) / `TileStrip .board-cap b` = title-small(14, 주변과 같은 크기·굵기로 구분) |

## 글자-3~5 결과 (2026-10-23)
- 글자-3(바꾸기): 27개 파일 73곳 `font-size: Npx` -> `@include vars.font(...)`, px 글자 크기 0곳
- 글자-4(터미널): 12px -> 14px(배율 1)
- 글자-5(깨지는 곳 맞추기): 격리 환경 전후 스크린샷(PC 1440 · 폰 390) — 보드·폴더 타일·장비 타일·터미널 헤더·푸터·공유·스크립트 편집·장비 지정 다이얼로그·전체 보기. 깨진 곳 = 위 `v-btn` 초기화 1건만 고침. 폰 폭에서 긴 폴더 이름이 말줄임으로 바뀜(정상 동작)
- x-small 버튼 글자(2026-10-23 사용자 "추천대로"): Vuetify 기본 10px(0.625rem)이 본문보다 작아짐 -> `global.scss`에서 `.v-btn--size-x-small`(아이콘·stacked 제외)만 label-medium(12), 높이 20px 그대로. Vuetify 컴포넌트 CSS는 `@layer vuetify-components` 안이라 layer 밖 전역 규칙이 우선순위와 무관하게 이김
- 안 본 곳: 알림 센터·토스트·내 공유·접속자 다이얼로그·폰 특수 키 줄(키보드 열린 상태)·전송 목록 -> 글자-6(확인)에서 사용자 브라우저로
