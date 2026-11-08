# REF — 설정 저장소 · 설정 창 · 푸터 오른쪽 메뉴 통일 (2026-11-07 신설, 파트(공용 타일 틀) 단계-3·4)

> → 이력 `history/settings-footer.md` · 뒤 파트(설정 두 겹) `REF-settings-account.md` · 앞 단계(공용 틀·메뉴 줄) `REF-tile-frame.md` · 찰떡 규칙 `REF-design-chaltteok.md` · 푸터 배치 `REF-board-chrome.md` · 진행 `CURRENT.md`

## 설정 저장소
- 이미 `common/util/userPrefs.util.ts`(브라우저별 localStorage, 접두사 `nexus.user.`, **로그아웃 때 전부 삭제**)가 있다
- **Q14(로그아웃 때 설정) 확정(2026-11-07 사용자)**: 경로 줄 위치도 로그아웃하면 지워짐 -> 브라우저 범위 항목은 `userPrefs` 위에 얹는다

## 설정 두 겹
- 2026-11-08 파트(설정 두 겹)로 분리 → `REF-settings-account.md` (Q2(설정 범위) 개정: 계정 + 이 브라우저)

## 푸터 오른쪽 메뉴 (확정 2026-11-07, 사용자)
| 이름 | 결정 |
|---|---|
| Q15(공통 버튼) | PC 알약 = **[● 아이디 ▾] [알림] [전체 화면]** 순서. 모든 페이지 공통 |
| Q19(추가 메뉴 자리) | **아이디 버튼 안에 추가 메뉴**: 누르면 [설정] [로그아웃](공통) |
| Q16·Q20(폰 모양) | 폰 = 모든 페이지 공통으로 지금처럼 **[⋯] 하나**, 그 안에 아이디 · 알림 · 전체 화면 · 설정 · 로그아웃 |
| Q17(설정 창 구성) | 한 열: 묶음 제목("타일") + 항목("메뉴 줄 위치 [위 \| 아래]"). 항목이 늘면 왼쪽 분류 목록 |
| Q18(시안) | 코드 전에 시안 → https://claude.ai/artifact/KCroqig1AG5hHjq77h2YTY (Design 캔버스, PC 보드 ①~④·관리 페이지·폰 ①~③, `Main`·`Phone`은 눌러 볼 수 있음) |

- 지금(바꾸기 전): 보드 PC = [전송][위치 알림][전체 화면][로그아웃], 보드 폰 = [⋯](위로 펼친 알약) / 관리 PC = [액션] + [● 닉네임][로그아웃], 관리 폰 = [로그아웃]
- 시안의 가정 3개 -> 사용자 답(2026-11-07)으로 확정:

| 이름 | 결정 | 시안과 다른 점 |
|---|---|---|
| Q21(전송 자리) | **개정(2026-11-07 착수 때, 사용자 "전송도 알림과 같은 레벨로")**: [전송](`TransferButton`, 전송이 있을 때만 보임)은 [알림] 바로 앞, **알약 안**(PC) · ⋯ 안(폰, 전송 중이면 ⋯에 점). 바꾸기 전 보드 배치와 같음. (처음 답 = 알약 겉) | 시안에 전송 없음 -> 알림 앞에 추가 |
| Q22(관리 페이지 알림) | **개정(2026-11-07 착수 때, 사용자 "관리페이지에선 알림을 빼도록 하자")**: 관리 페이지엔 [알림] 없음. 이유 = 알림 받기(`useTileNoticeFeed`)가 보드 트리·보드 목록(`pages/index.vue` provide)에 기대서 관리 페이지에선 새 알림을 못 받음. (처음 답 = 관리 페이지에도 둠) | 시안과 같음(관리 페이지엔 없음) |
| Q23(설정 창 적용) | 고르면 바로 적용, 버튼은 [닫기]만 | 같음 |

- 시안은 위 차이를 고치지 않고 둔다(사용자: "아티팩트 수정하지 않아도 됨"). 시안 폰 화면에서 위/아래를 골라도 타일이 안 바뀌는 건 시안에서 연결하지 않았기 때문 — 실제 코드는 `TileScroll`이 기기 무관 같은 `tileBarAt`을 따르고 값은 브라우저별 저장이라 폰에도 적용됨

## 최종 모양 (구현 기준)
| 자리 | 보드 PC | 보드 폰 | 관리 PC | 관리 폰 |
|---|---|---|---|---|
| 오른쪽 알약 | [● 아이디 ▾] [전송(있을 때)] [알림] [전체 화면] | [⋯] | [● 아이디 ▾] [전체 화면] | [⋯] |
| 아이디 누름 / ⋯ 누름 | 추가 메뉴: [설정] [로그아웃] | 위로 펼친 알약: [● 아이디] [전송(있을 때)] [알림] [전체 화면] [설정] [로그아웃] | 추가 메뉴: [설정] [로그아웃] | 위로 펼친 알약: [● 아이디] [전체 화면] [설정] [로그아웃] |
| 관리 페이지 액션 | — | — | 알약 왼쪽 밖(지금처럼 `FootActions`) | 같음 |

- 공용 컴포넌트는 **보드 전용 버튼 자리(슬롯) 하나**(아이디 다음, 전체 화면 앞)와 **⋯ 점 켜기 값**만 받는다. 보드 푸터가 [전송][알림]을 넣고 점 조건(안 읽은 알림 또는 전송 중)을 넘김, 관리 페이지는 아무것도 안 넣음. 이유 = 전송(`useTransfers`)이 보드 페이지 provide라 공용 컴포넌트가 직접 부르면 관리 페이지에서 깨짐

## 구현 계획 (2026-11-07 정리 · 같은 날 구현)
| 순서 | 할 일 | 파일(안) |
|---|---|---|
| 1 | 설정 저장소: 항목표(키·기본값·범위) + 반응형 값. 브라우저 범위 = `userPrefs`(`readUserPref`/`writeUserPref`). **로그아웃(`clearUserPrefs`) 뒤 반응형 값도 기본값으로** 되돌리고, 다른 탭 변경은 `storage` 이벤트(`userPrefKey`)로 반영. 첫 항목 `tileBarAt`(기본 `'top'`) | 새 `feature/settings/store/settings.store.ts`(위치 안) · `widget/util/tileBar.util.ts`의 임시 `ref`를 설정 값으로 바꿈 |
| 2 | 설정 창: 앱에 하나(`App.vue`), `openSettings()`로 엶. 가운데 팝업 = `translate` 속성 + `jelly-pop`(젤리 규칙), 카드 `.ct-card` + 아래 끝 블록 `.ct-actions`([닫기]), 고르기 = `.ct-grp` > `.ct-seg`(창 오른쪽 벽에 붙음, 찰떡) | 새 `feature/settings/component/SettingsDialog.vue` |
| 3 | 공용 오른쪽 메뉴: 위 "최종 모양". 전체 화면 = `useAppActions`(`canFullscreen`·`toggleFullscreen`, 이미 공용) · 로그아웃 = `toggleAuth` · 보드 전용 버튼 = 슬롯(위 문단). 알약 모양 = `FootPill edge="right"` | 새 공용 컴포넌트(이름 안 `FootAccount.vue`) |
| 4 | `BoardFooter` 오른쪽(PC 알약·폰 ⋯) -> 공용 컴포넌트 | `feature/node/component/BoardFooter.vue` |
| 5 | `MemberFooter` 오른쪽 `.acct` 알약 -> 공용 컴포넌트(`FootActions`는 왼쪽에 그대로) | `feature/member/component/MemberFooter.vue` |
| 6 | 확인: type-check · `run.sh`(낡은 2건 외 통과) · 헤드리스(PC·폰 폭, 보드·관리 페이지(관리 페이지엔 알림 없음), 설정 바꾸면 타일 메뉴 줄 이동, 로그아웃하면 처음 값) · 사용자 화면 확인 | — |

- 시안 렌더 확인 = 캔버스 런타임 없이 상태 분기를 풀어 놓은 임시 사본을 headless Chrome으로 찍음(설정 아이콘이 해처럼 보여 톱니 모양으로 고침)

## 구현 결과 (2026-11-07, 사용자 화면 확인 "잘 동작해")
| 파일 | 바뀐 것 |
|---|---|
| 새 `feature/settings/store/settings.store.ts` | 항목표 `SPECS`(범위·기본값·모양 검사) + `settings`(읽기 전용) · `setSetting` · `openSettings`/`settingsDialog`. 저장 키 = `settings.<항목>`(userPrefs 접두사 붙어 `nexus.user.settings.tileBarAt`). 다른 탭 = `storage` 이벤트, 이 탭 로그아웃 = `authState` watch로 다시 읽음(+ 설정 창 닫음) — `nodeClipboard.store`와 같은 방식 |
| `widget/util/tileBar.util.ts` | 임시 `ref` -> `computed(settings.tileBarAt)` |
| 새 `feature/settings/component/SettingsDialog.vue` + `App.vue` | 앱에 하나. `v-dialog`(다른 창과 같은 방식) + `.ct-card` · 고르기 `.ct-grp--r`(창 오른쪽 벽) · `.ct-actions`[닫기] |
| 새 `feature/layout/component/FootAccount.vue` | 공용 오른쪽 메뉴. prop `phone`·`dot`, 기본 slot(slot prop `phone`) = 보드 전용 버튼 |
| `BoardFooter.vue` | 오른쪽 -> `FootAccount` + slot에 [전송][알림], 점 = 안 읽은 알림 \|\| 전송 중 |
| `MemberFooter.vue` | `.acct` 알약 -> `FootAccount`(slot 없음 = 알림 없음) |

- 계획과 다르게 한 것
  | 이름 | 계획 | 실제 | 이유 |
  |---|---|---|---|
  | 구현-1(설정 창 방식) | 가운데 팝업 `translate` + `jelly-pop` | `v-dialog` | 앱의 다른 창(주소 변경·공유 등)이 전부 `v-dialog`. `jelly-pop`은 컨트롤 패널용(REF-jelly.md) |
  | 구현-2(이름 표시) | [● 아이디] | **닉네임** | 서버가 `identification`을 해시로 줌(확인 중 title에 해시가 찍혀 발견). 옛 관리 푸터도 닉네임 |
  | 구현-3(계정 버튼 화살표) | 시안 ▾ | `mdi-chevron-up` | 왼쪽 보드 메뉴·관리 메뉴와 같게(셋 다 위로 열림) |
  | 구현-4(폰 펼침 모양) | 바꾸기 전 = 세로 알약 | **가로 알약**(시안대로) | 이름 칸이 들어가 가로가 맞음. 390px에서 폭 279px |
- 확인(격리 5051 + 임시 DB `nexus_sf_tmp`, 계정 `sfcheck`, worker 없음 — 끝나고 DB 지움): PC 보드 알약 순서·계정 메뉴[설정][로그아웃]·설정 창 처음 = 위·아래 고르면 창 그대로 + 저장값 + 메뉴 줄 아래·**다른 탭도 아래**·새로고침 유지 · 관리 페이지 알림 없음·설정 창 = 아래 · 관리 페이지에서 로그아웃 -> 저장값 지워짐 -> 다시 로그인 = 위 · 폰 보드 ⋯ 하나·펼침 [이름][알림][전체 화면][설정][로그아웃]·화면 안 · 폰 설정 창 · 폰 관리 펼침 알림 없음 · 콘솔 에러 없음 = **0 FAIL**(첫 판 2 FAIL = 해시 title, 구현-2로 고침). `run.sh` = 낡은 기대값 2건 외 통과. type-check 통과. 스크립트 = 세션 scratchpad `iso/setup.mjs`·`iso/check.mjs`(재부팅 시 사라짐)
- 못 본 것: ⋯ 빨간 점(안 읽은 알림·전송 중을 만들지 않음), 전송 버튼이 보일 때 알약 모양
