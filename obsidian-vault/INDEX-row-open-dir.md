# INDEX — 칸 메뉴 장비 디렉토리 (마감)

> 2026-10-27 파트(칸 메뉴 장비 디렉토리) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
브랜치 `feat/row-open-dir`(master `e16a519`에서).
- 계기: 폴더 칸 ⋮ 메뉴에서도 장비 디렉토리를 열고 싶음(경로 줄 버튼은 지금 연 폴더만)
- 구현: `FolderTileBody.vue` `menuItems`에 "장비 디렉토리 열기" — 내 폴더 + 장비(상속 포함)일 때만, 오프라인 = 꺼짐, 동작 = 경로 줄 버튼과 같은 `ops.openDir`
- 확인: 격리 헤드리스(o2 주인·v3 공유받음) + 사용자 직접 확인 OK(2026-10-27). 배포 안 함
- 커밋: 브랜치 `feat/row-open-dir`에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-row-open-dir.md`(결정 · 구현 · 확인) / `history/row-open-dir.md`

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: **진행 중 파트**: 칸 메뉴 장비 디렉토리(2026-10-27, 브랜치 `feat/row-open-dir`) → `REF-row-open-dir.md`
- HISTORY: | `history/row-open-dir.md` | 파트(칸 메뉴 장비 디렉토리, 진행 중) ↔ `REF-row-open-dir.md` | 10-27 구현·격리 헤드리스 확인, 사용자 확인 대기 |

## 이 파트에만 해당하는 결정 · 남은 것
- 장비 없는 내 폴더 = 숨김(꺼짐 아님) · 상속 포함 · 스크립트 칸 제외 · 공유받은 폴더 = 안 보임
- 칸의 상속 장비 = `n.deviceKey ?? 지금 경로의 장비`(서버 `ResolveDeviceKey`와 같은 결과여야 함)
