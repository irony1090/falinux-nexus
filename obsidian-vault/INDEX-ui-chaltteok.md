# INDEX — 찰떡 마무리 (마감)

> 2026-10-27 파트(찰떡 마무리) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
브랜치 `feat/ui-chaltteok`(master `87ff6f2`에서). 프론트만.
- 계기: 사용자 "자잘한 UI 수정" — 경로 줄 괜한 여백 · 찰떡 안 된 창 · 관리자 필터 찰떡 + 이어서 관리 상세 상태 칸 정리
- 구현: 1(경로 줄 여백 = 2글자 이하 칸만 `.short`) · 2(관리자 필터 = 시안 A′) · 3(남은 창 5개 찰떡, `AppDialog` 버튼 종류 -> 칸 색) · 4(상세 상태 칸 = 안 2, 구분선 없음)
- 항목 5(select 목록 흰 바탕 흰 글씨) = 이 파트에서 빼고 **다음 파트(찰떡 입력칸)** 로 넘김
- 확인: 격리 5051 헤드리스 스크린샷 + 사용자 확인 OK(2026-10-27). 배포 안 함
- 커밋: 브랜치 `feat/ui-chaltteok`에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-ui-chaltteok.md`(항목 표 · 구현 · 확인 · 찰떡 입력칸 구상) / `history/ui-chaltteok.md`

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: **진행 중 파트**: 찰떡 마무리(2026-10-27, 브랜치 `feat/ui-chaltteok`) → `REF-ui-chaltteok.md` (직전 = 칸 메뉴 장비 디렉토리(마감) → `INDEX-row-open-dir.md`)
- HISTORY: | `history/ui-chaltteok.md` | 파트(찰떡 마무리, 진행 중) ↔ `REF-ui-chaltteok.md` | 10-27 경로 줄 여백·창 5개·필터 A′·상세 칸 안 2 구현(미커밋), 찰떡 입력칸 논의 |

## 이 파트에만 해당하는 결정 · 남은 것
- `<input readonly>`는 Chromium에서 키보드 커서가 안 움직임 -> 보기 전용 칸 = 일반 input + `@beforeinput.prevent`·`@drop.prevent` + `inputmode="none"`(→ REF "함정")
- 새 칸 색 클래스 `.ct-seg--warning`·`.ct-seg--danger`(`styles/chaltteok.scss`)
