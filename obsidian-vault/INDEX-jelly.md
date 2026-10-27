# INDEX — 젤리·관리 기록 (마감)

> 2026-10-27 파트(젤리·관리 기록) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
사용자 요청 3개를 한 파트로. 브랜치 `feat/jelly-nav`(master `a563234`에서).
- ①보드로: 관리 페이지 로고 메뉴 '보드로' = 직전 기록이 보드면 `router.back()`, 아니면 마지막 보드 주소(`focus`·`view` 포함)로 push — `router/lastBoard.ts`
- ②관리 기록: 관리 페이지는 들어올 때만 push, 안에서 바뀌는 URL·내 장비↔관리자 이동은 replace(`REF-member-ui.md` 주소-1(진입만 push) 개정·주소-7(보드로) 추가)
- ③젤리 효과: 누를 수 있는 것 hover(커졌다 돌아옴)·누름(눌렸다 튀어 돌아옴) + 컨트롤 패널 등장·퇴장, 세기 = 중간 — `styles/jelly.scss`. 시안 아티팩트 https://claude.ai/artifact/Gvj3eoa77tsppmWHsXEG9n
- 확인: `vue-tsc`·`vite build` OK, test1 브라우저(①② 기록 수·복귀 주소, ③ 계산된 애니메이션·패널 가운데 고정) + **사용자 확인 OK**(2026-10-27). 배포 안 함(test3 웹 교체는 사용자 판단)
- 커밋: 브랜치 `feat/jelly-nav`에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-jelly.md`(요청 3개 · 확정 값 · 적용 범위 · 함정 · ①②구현 메모) / `history/jelly.md`
- URL 규칙 본문은 접힌 파트(회원 관리)의 `REF-member-ui.md` "URL 저장"

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: **진행 중 파트**: 젤리·관리 기록(2026-10-27, `feat/jelly-nav`) → `REF-jelly.md`(젤리 효과·보드로·관리 페이지 replace, 함정 = `transform` translate + `scale` 기준점)
- HISTORY: | `history/jelly.md` | 파트(젤리·관리 기록) ↔ `REF-jelly.md` | 10-27 시안 v1~v3·구현·test1 확인 |

## 이 파트에만 해당하는 결정 · 남은 것
- 넓은 줄(`v-list-item`·목록 줄) 제외, 폰 = 누름만, 묶음(`ct-grp`·`mb-grp`·`tile-btn-end`) 안 칸 = 상자째(바깥 끝은 잘림), 페이지 도트 = hover만
- 다른 파트에도 남는 규칙 2개(새 버튼 클래스 -> `$box`, 팝업 = `translate` 속성 + `jelly-pop`)는 MEMORY 확정 결정 "젤리 효과" + `REF-decisions.md`로
- 관리자 페이지·관리자 컨트롤 패널은 Claude가 직접 못 봄(시험 계정 irony = 관리자 아님), 같은 코드라 빌드로만 확인
