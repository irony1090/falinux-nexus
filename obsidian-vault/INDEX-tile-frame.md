# INDEX — 공용 타일 틀 (마감)

> 2026-11-07 파트 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
11-07 하루에 시작·마감. 브랜치 `refactor/tile-shell` -> master `--no-ff` 머지. 사용자 화면 확인 "스크롤에 따른 숨김/보임 확인"(단계-2) · "잘 동작해"(단계-3·4)
| 단계 | 내용 |
|---|---|
| 단계-1(공용 틀 리팩터링) | `TileShell`(바탕·머리 줄 slot) · `TileScroll`(본문 스크롤 + `#bar`) — 보드·관리·updater 타일 전부 |
| 단계-2(메뉴 줄 위치·숨김) | 메뉴 줄 = 스크롤 안 sticky, 위/아래 · 내리면 숨김·올리면 보임 |
| 단계-3(설정 저장소·dialog) | `settings.store`(항목표 `SPECS`, 브라우저 범위 = userPrefs, 로그아웃하면 처음 값) · `SettingsDialog` |
| 단계-4(오른쪽 메뉴 통일) | `FootAccount` 공용: PC [● 닉네임 ^][전송][알림][전체 화면] / 폰 [⋯], 관리 페이지엔 전송·알림 없음 |
- 다른 파트에서도 지킬 결정(타일 틀·푸터 오른쪽·설정·git 머지)은 MEMORY "확정 결정"과 `REF-decisions-ui.md`에 남김
- 남은 것: ⋯ 빨간 점·전송 버튼이 보일 때 알약 모양은 헤드리스로 못 봄 / `run.sh` 낡은 기대값 2건 = `BACKLOG.md`

## REF / history 목록
- `REF-tile-frame.md` / `history/tile-frame.md` — 공용 틀·메뉴 줄(단계-1·2)
- `REF-settings-footer.md` / `history/settings-footer.md` — 설정·푸터(단계-3·4)

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: - **공용 타일 틀·메뉴 줄(11-07)**: `REF-tile-frame.md`(TileShell·TileScroll·메뉴 줄 위/아래 숨김) · 설정·푸터 `REF-settings-footer.md`(설정 저장소·설정 창·오른쪽 메뉴 통일)
- HISTORY: | `history/settings-footer.md` | 설정 저장소 · 설정 창 · 푸터 오른쪽 메뉴(11-07 신설) ↔ `REF-settings-footer.md` | 11-07 조율 Q14~Q23 · 시안 · 구현 계획 |
- HISTORY: | `history/tile-frame.md` | 공용 타일 틀 · 메뉴 줄 위치/숨김 · 설정(11-07 신설) ↔ `REF-tile-frame.md` | 11-07 조율 Q1~Q13 · 계획 · 단계-1·2 구현·확인 |
