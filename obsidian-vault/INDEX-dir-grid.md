# INDEX — 목록 아이콘 보기 (마감)

> 2026-10-24 파트(목록 아이콘 보기) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
폴더 타일·장비 타일 목록을 표 줄에서 격자 아이콘 칸으로 **교체**(보기 전환 없음). 브랜치 `feat/dir-grid`(master `7d41f81`에서).
- 시안: 캔버스 https://claude.ai/artifact/MZ7QxvXoCJ3xYrx6YYiVsa — 판 A(모서리 표식)·B·C -> 판 D(상자 경계에 붙은 상태 박스) -> **판 D′(큰 아이콘·라벨 겹침·버튼 항상)** 확정
- 칸: 상자 ≈116×112(최소 100px), 아이콘 46px, 오른쪽 라벨(공유 수·실행 중 개수) · 왼쪽 라벨(장비·소유자·링크·끊김)이 아이콘 모서리를 덮음. 위쪽 버튼 항상: 왼쪽 위 = 새 타일(폴더) / 실행(스크립트·실행 파일), 오른쪽 위 = ⋮. hover 전용 기능 없음 — 정보 카드는 ⋮ 메뉴 맨 위
- 누르기 규칙(두 타일 공통): 한 번 = 고름(테두리) · 더블클릭·더블탭 = 열기·실행 · Ctrl+클릭·길게 누르기 = 체크(선택 모드에선 왼쪽 위가 체크박스, ⋮ 숨김)
- 추가: 칸에서 공유(⋮ "공유"·공유 수 라벨, 내 폴더만)·장비 지정(장비 라벨)
- 코드: `hook/itemPress.hook.ts` · `DirGridCell.vue` · `styles/dir-grid.scss` · `NodeRowMenu.vue`(제네릭 + `#head`) · `FolderTileBody.vue` · `DeviceTileBody.vue`
- 확인: 격리(PC 1280·폰 390 터치, 계정 2개 공유) + **사용자 test1(8011 개발 서버 -> test3 API) 확인 OK**(2026-10-24). 배포 = test3 `build/ctrl_panel/web`만 교체(백엔드 변경 없음)
- 커밋: 브랜치 `feat/dir-grid`에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-dir-grid.md`(출발 · 코드 조사 · 결정 아이콘-Q1~Q6 · 시안 · 판 D 조율 · 구현 계획 G0~G4 · 구현 · 확인 · 추가 요청) / `history/dir-grid.md`
- 확인 스크립트 명세 → `REF-test-scripts-1023.md` "아이콘-확인"

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: 목록 아이콘 보기 `REF-dir-grid.md`(2026-10-24 브랜치 `feat/dir-grid` 구현, 사용자 확인 대기)
- HISTORY: | `history/dir-grid.md` | 목록 아이콘 보기(2026-10-24, 진행 중) | 10-24 아이콘-Q1~Q6 답 · 시안 판 A~D′ · 교체·누르기 규칙 확정 · G0~G3 구현·격리 확인 |

## 이 파트에만 해당하는 결정 · 남은 것
- 아이콘-Q2(교체/전환) = 교체 — 장비 타일 크기·수정 시각은 ⋮ 카드로만(사용자 "크게 중요하지 않음")
- 실행 가능 ▶ 라벨 없음(실행 버튼과 중복, 실행 중 개수와 헷갈림) · 장비 라벨 숫자 = 접속 중 수(전체는 카드)
- 함정: body의 scoped 버튼 리셋에서 `.dg-btn, .dg-lbl`을 빼야 함 · 더블탭은 click 간격(300ms)으로 직접 잼 + 칸 `touch-action: manipulation`
