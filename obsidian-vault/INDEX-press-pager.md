# INDEX — 빈 곳 길게 누르기 (마감)

> 2026-10-25 파트(빈 곳 길게 누르기) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
파트(목록 아이콘 보기) 뒤 터치 길게 누르기 = 타일 컨트롤 패널(`PagerPopup`)이 경로 줄·목록 빈 곳에서 안 뜨던 것 + 고른 칸의 누를 수 있는 라벨 강조. 브랜치 `feat/press-pager`(master `e6ff22e`에서).
- 자리: 경로 줄 빈 곳·현재 폴더 글자·도구 줄 빈 곳·목록 칸 밖 = 패널(짧게 = 경로 입력칸 그대로) / 칸 안 = 체크 / 버튼·라벨 = 각자 / 터미널 화면 제외. 폴더·장비 타일
- 라벨: 고른 칸의 `button.dg-lbl`만 안쪽 1.5px 테두리 0.2초 전환(장비 = 라벨 색, 공유 수(solid) = `--on-c`), 선택 모드 제외
- 코드: `PagerPopup.vue` `pressable` · 두 body 루트 `data-pager-press`(+ 경로 편집 중 `data-pager-press-off`) · `dir-grid.scss` `--ring` · `DirGridCell` `checking` 클래스
- 후속(마감 뒤): 루트 탭 줄(`RootTabs`)도 대상 → REF "후속"
- 확인: 헤드리스 폰 390 CDP 터치 13/13(test1 -> test3 API) + **사용자 test1 확인 OK**(폴더·장비 타일, 2026-10-25). 배포 = test3 `build/ctrl_panel/web`만 교체(백엔드 변경 없음)
- 커밋: 브랜치 `feat/press-pager`에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-press-pager.md`(출발 · 확정 자리 표 · 계획 길게-1~5·라벨-1 · 구현 · 확인) / `history/press-pager.md`

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: 빈 곳 길게 누르기 `REF-press-pager.md`(2026-10-25 구현, 사용자 확인 대기)
- HISTORY: | `history/press-pager.md` | 빈 곳 길게 누르기 · 고른 칸 라벨 강조(2026-10-25, 진행 중) | 10-25 조율·구현·헤드리스 13/13, 사용자 확인 대기 |

## 이 파트에만 해당하는 결정 · 남은 것
- 라벨 상시 표시(꺾쇠·눌림 반응·테두리 안 A~C)는 사용자 철회 — 고름 때만 강조
- 다른 파트에도 남는 규칙(새 몸통 = `data-pager-press`)은 `REF-tile-kinds.md` "새 종류를 더할 때" 4번으로 옮김
- 함정: 칸 길게 = 보드 전체 선택 모드, 체크를 다 풀어도 ✓ 전까지 유지(기존 동작) — 헤드리스 좌표가 밀림
