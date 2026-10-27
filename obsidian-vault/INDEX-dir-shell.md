# INDEX — 폴더에서 터미널 (마감)

> 2026-10-27 파트(폴더에서 터미널) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
브랜치 `feat/dir-shell`(master `9a56ab9`에서).
- 계기: 브라우저 업로드 파일의 실행 권한이 사라짐(브라우저가 권한 비트를 못 줌) → 권한 기능 대신 **장비 타일에서 지금 폴더로 셸을 바로 열기**로 방향 전환, 권한은 셸에서 `chmod`
- 구현: 장비 타일 경로 줄 새로고침 오른쪽 `mdi-console` → 기존 폴더-4(실행) API에 cmd = 치환 토큰 `{WORKER_SHELL}` → worker가 Linux `$SHELL`(>bash>sh) / Windows `powershell.exe`로 바꿈. 이름 "셸"
- 확인: 격리 헤드리스 Linux + 사용자 직접 확인(격리 3001 + VM `win-dev`) OK(2026-10-27). 배포 안 함
- 커밋: 브랜치 `feat/dir-shell`에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-dir-shell.md`(결정 · 구현 표 · 확인) / `history/dir-shell.md`
- 바탕: 접힌 파트(worker 폴더 타일)의 `REF-worker-dir-exec.md`(폴더-4(실행))

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: **진행 중 파트**: 폴더에서 터미널(2026-10-27, 장비 타일 "여기서 셸", 브랜치 `feat/dir-shell`) → `REF-dir-shell.md`
- HISTORY: | `history/dir-shell.md` | 파트(폴더에서 터미널) ↔ `REF-dir-shell.md` | 10-27 구현·격리 확인 + 사용자 확인(VM 포함) |

## 이 파트에만 해당하는 결정 · 남은 것
- 셸은 장비 주인만(기존 `requireDeviceDir`) · Windows = Windows PowerShell 고정 · 셸 바꾸기(Linux) = worker env `SHELL=`
- 서버 토큰 `protocol.PlaceholderWorkerShell`과 프론트 `SHELL_CMD`는 같은 값이어야 함(이름 "셸" 표시가 둘에 나뉨)
- 업로드 실행 권한(BACKLOG)은 이 파트로 대체돼 삭제
