# INDEX — 새 보드·IUTF8 (마감)

> 2026-10-22 파트(새 보드·IUTF8) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
BACKLOG "같이 진행"의 작은 수정 두 개. master에 바로 한 커밋.
- 수정-1(새 보드): W1(분리 범위) "`/`에서만"을 "어느 보드에서든"으로 개정. 서버 `workspaceCreateRequest.From`(출발 hash) + 프론트 `createBoard(from)`·메뉴 조건 삭제
- 수정-2(IUTF8): worker PTY slave에 `IUTF8` 켜기(`setUTF8`) -> canonical 모드에서 한글 백스페이스가 한 글자를 지움
- 마감 확인 = 빌드본 `build/iutf`(8013) 브라우저 실사용

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-fix-board-iutf8.md`(새 보드·IUTF8) / `history/fix-board-iutf8.md`

## HISTORY 색인 행 (원문)
| `history/fix-board-iutf8.md` | 새 보드·IUTF8(2026-10-22 신설) | 10-22 수정-1(새 보드)·수정-2(IUTF8) 작성·확인 / 10-22(2) 빌드본 브라우저 확인 OK, 마감·커밋 |

## 이 파트에만 해당하는 결정
- 확인 함정(IUTF8): 조합 중인 글자에서 누른 백스페이스는 브라우저 IME가 처리해 PTY로 안 감 -> 확인은 조합을 확정(한/영 키)한 뒤 백스페이스
- 반영 범위: 수정-1은 supervisor + 웹을 같이 바꿔야 함(웹만이면 옛 서버가 `from` 무시 -> 404) / 수정-2는 worker 재실행 뒤 새 process부터
