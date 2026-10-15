# INDEX — 프론트 마무리 (마감)

> 2026-10-14 색인 2단화로 MEMORY·HISTORY에서 접음(사용자 "이대로 진행"). **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
2026-06-30~10-13. 프론트의 로그인 마무리와 active 타일 표시를 묶음.
| 파트 | 마감 | 한 줄 | REF |
|---|---|---|---|
| user/login 마무리 | 2026-10-12 | 로그인 상태를 모듈로 · 라우터 가드 · 401만 로그아웃 · 전역 다이얼로그(AppDialog). 회원가입은 별도 파트로 보류 | `REF-frontend-auth.md` |
| active 타일 · 원격 배지 | 2026-10-12 | 내 active 타일(약하게 흐림) + 위치 번호 단축키 `Ctrl+Alt+숫자` + 다른 계정 active 배지 | `REF-node-ui-active.md` |

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-frontend-auth.md`(로그인 상태·라우터 가드·401·AppDialog) / `history/frontend-auth.md`
- `REF-node-ui-active.md`(active 타일·단축키 `Ctrl+Alt+숫자`·다른 계정 active 배지) / `history/node-ui-active.md`

## MEMORY 모듈 현황표 행 (접기 전 원문)
| 모듈 | 상태 | 문서 |
|------|------|------|
| 프론트엔드 (apps/frontend) | 스캐폴딩·socket hook·user/login(라우터 가드·401 처리)·전역 다이얼로그 + node 타일 UI(레이아웃 → `INDEX-node-ui-layout.md`, 연동 → `INDEX-node-ui-link.md`, 둘 다 마감) + 터미널 기능(→ `INDEX-process-terminal.md`). ⑮(모바일 가상 키보드) = **마감(2026-10-14) → `INDEX-mobile-keyboard.md`** / user/login 마무리 = **마감(2026-10-12, `REF-frontend-auth.md`)**, 회원가입은 별도 파트로 보류 / active 타일 표시·위치 번호 단축키·다른 계정 active 배지 = **마감(2026-10-12, `REF-node-ui-active.md`)** | `REF-frontend.md`(+`REF-node-ui.md` 컨셉 / `REF-node-ui-overview.md` 전체보기·가상 키보드) |

## HISTORY 색인 행 (원문)
### `history/node-ui-active.md` — active 타일 표시 · 위치 번호 단축키 (2026-10-12 신설)
- 2026-10-12(6) 사용자 확인 → 파트 마감·커밋
- 2026-10-12(5) 원격 active 배지(터미널 + 폴더) 구조 확정·구현, 격리 헤드리스 11/11(커밋 전)
- 2026-10-12(4) 로컬 active 사용자 확인 완료
- 2026-10-12(3) AT4(단축키 조합) = `Ctrl+Alt+숫자`로 변경, 헤드리스 14/14
- 2026-10-12(2) AT 확정 + 활성-1(상태·누르기)~활성-5(확인) 구현, 헤드리스 13/13(커밋 전)
- 2026-10-12(1) 구조안 제시 AT1(상태 위치)~AT6(이동할 때 같이 할 일) + 활성-1~5 + 다른 계정 active 배지 RA1(맞추는 기준)~RA6(순서), 코드 없음
### `history/frontend-auth.md` — 로그인 · 공용 API · 전역 다이얼로그 (2026-10-12 frontend에서 분리)
- 2026-10-13 헤더 로그아웃 버튼 = 햄버거 → `mdi-logout` 아이콘 버튼
- 2026-10-12(3) 사용자 확인 → 파트(user/login 마무리) 마감·커밋
- 2026-10-12(2) 보드에서 로그아웃 시 "보드 삭제" 오인 알림 수정(`boards.loaded` 조건)
- 2026-10-12 파트(user/login 마무리) 로그인-1(상태를 모듈로)~로그인-5(확인) 구현, 헤드리스 12/12
- 2026-06-30 user/login + 공용 API + 전역 다이얼로그 WIP (커밋 3a8e92e 동반)

## 이 파트에만 해당하는 결정
- 로그아웃은 401일 때만(로그인-3(401만 로그아웃)) → `REF-frontend-auth.md`
- active 단축키 = `Ctrl+Alt+숫자`(AT4(단축키 조합)) → `REF-node-ui-active.md`
