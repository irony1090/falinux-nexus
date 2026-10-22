# history — 새 보드·IUTF8

> → `REF-fix-board-iutf8.md` / → `CURRENT.md`

## 2026-10-22
- BACKLOG "같이 진행"에서 착수(사용자: worker 폴더 타일보다 먼저)
- 원인 조사: 새 보드 = W1(분리 범위) "`/`에서만" 결정(버그 아님) -> 사용자가 "어느 보드에서든"으로 개정 / IUTF8 = PTY termios 미설정
- Claude가 작성: 서버 `from`·프론트 메뉴·`setUTF8`. 확인 IUTF8 overlay 테스트 + 새 보드 격리 16/16
- 반영 범위 정리: 수정-1 = supervisor + 웹(웹만 바꾸면 옛 서버가 `from`을 무시해 404), 수정-2 = worker 재실행 뒤 새 process부터. 빌드본 worker 교체는 사용자 세션을 끊으므로 세션 끝에. 사용자가 supervisor + 웹 재빌드 시작, vault 정리

## 2026-10-22(2)
- 사용자가 새 빌드 `build/iutf`로 worker·supervisor 교체 후 재접속. 세션 PTY `stty`에 `iutf8` 확인
- 브라우저 확인: 수정-1(새 보드) OK / 수정-2(IUTF8) 1차 `가ㄴ` = 조합 중 백스페이스를 IME가 처리(이 수정과 무관), 2차(한/영으로 확정 후) `가` OK
- 파트 마감·커밋(파트(worker 폴더 타일) 구상 기록·오늘 vault 정리 포함). 옛 `build/const`·`build/_check`는 사용자가 삭제
