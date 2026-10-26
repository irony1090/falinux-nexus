# INDEX — supervisor 주소 실시간 변경 (마감)

> 2026-10-26 파트(supervisor 주소 실시간 변경) master 머지(`efbe403`) 뒤 vault 정리 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
supervisor를 옮기거나 업데이트할 때 worker 연결을 끊지 않고 주소를 바꾼다. 브랜치 `feat/supervisor-move`(중간 커밋 `40e0942` = 구현 전부).
- 주소-1(주소 목록): worker sqlite `supervisors`(최대 3개, 순위) · 주소-2(옮기기 명령): 바로 옮기기/저장만, 새 쪽 실패 시 옛 주소로 · 주소-3(권한 열): users `perms`(`00013`), `worker_move` = 전체 worker 옮기기
- 주소-4(미등록 process): 붙은 supervisor DB에 행이 없는 process — 죽이지 않고 장비 타일 메뉴 목록에서 보고 끔. worker holds에 label·cmd·cwd(`00004`)
- 주소-5(화면): 장비 타일 메뉴 "supervisor 주소 변경"·"미등록 process" + 보드 메뉴 "전체 worker 주소 변경"
- 확인: 격리 서버 29/29·화면 8/8 + **사용자 확인 = 회사 Linux B 빌드로 test3 -> 집 서버로 옮기기 OK**(2026-10-26)
- 후속(`feat/member`에서, 2026-10-26): **미등록 process 가져오기**(가져오기-1(화면 다시 요청)~가져오기-5(확인)) — 새 supervisor의 장비 주인이 등록·터미널 타일 → REF "가져오기"·"가져오기 구현", 커밋 `fa50c64`

## REF / history 목록
- `REF-supervisor-move.md`(결정 주소-Q · 계획 주소-1~6 · 구현 · 확인 · 가져오기 조율·구현) / `history/supervisor-move.md`

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: supervisor 주소 실시간 변경 `REF-supervisor-move.md`(브랜치 중간 커밋, 사용자 확인 대기)
- HISTORY: | `history/supervisor-move.md` | supervisor 주소 실시간 변경(2026-10-24 착수, 진행 중) | 10-26 사용자 확인 OK(집 서버로 옮기기), 머지 대기 |

## 이 파트에만 해당하는 결정 · 남은 것
- 이름 = "미등록 process"("고아" 안 씀) · 미등록은 자동으로 끄지 않음 · 목록 권한 = 그 장비 폴더 소유자(회원 관리 뒤 = 장비 주인)
- 주인 = 붙어 있는 supervisor: 가져오기는 가져온 쪽 장비 주인 것, worker가 옛 쪽으로 돌아가면 옛 행으로 이어짐(추가 기록 없음, 사용자 ①)
- 가져오기는 EXEC만(EDIT는 결과 저장 노드를 확정 못 해 끄기만)
- 키는 supervisor DB마다 따로 — 새 DB로 옮기면 장비를 같은 키로 등록해야 worker가 붙음
