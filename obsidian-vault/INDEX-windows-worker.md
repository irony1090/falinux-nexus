# INDEX — Windows worker (마감)

> 2026-10-19 파트(Windows worker) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
worker가 Windows에서 빌드조차 안 되던 문제(막히는 곳 = `internal/execute/pty` 하나). `pty`를 공통 + `_linux.go` + `_windows.go`로 나누고 Windows판을 ConPTY + Job Object로 구현. EXEC는 노드 이름 확장자로 실행기를 고름(`.ps1` = powershell, 그 외 = `cmd /c` 파일 연결, `#!` 안 씀), 실행 전 env `SCRIPT_FORMAT`으로 인코딩·줄 끝 손질(`.ps1` BOM, `.bat` CP949·CRLF 기본). 에디터 = `$VISUAL` > `EDITOR` > PATH 탐색(`edit` 등) + 인자 분리(Linux도 개선). worker OS를 등록·명부(`00011`)에 남기고 프론트에 L/W 표시. VM `nexus-win`(dockurr/windows)에서 17/17 확인 — 확인 중 WMI로 띄운 worker의 Ctrl+C 무시 속성이 스크립트로 상속되는 문제를 찾아 `SetConsoleCtrlHandler(NULL, FALSE)`로 수정. 재접속은 사용자 확인

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-windows-worker.md`(계획·결정 WW1(실행 규칙)~WW5(OS 표시)·작업 단위 윈도-0~윈도-6·VM 실행/준비·위험·마감 때 미확인 목록) / `history/windows-worker.md`
- `REF-windows-conpty.md`(윈도-0(빌드 분리) 결과 · ConPTY 시작/끝날 때 순서 · 윈도-6 ConPTY·종료 결과 · Ctrl+C 상속 수정) / `history/windows-conpty.md`
- `REF-windows-exec.md`(스크립트 형식 손질 · 윈도-4(에디터) 메모 · 윈도-5(OS 표시) · 윈도-6 실행 규칙 결과) / `history/windows-exec.md`
- 확인 스크립트 명세: `REF-test-scripts-1014.md` "2026-10-19 확인 스크립트 (Windows worker 윈도-6(확인), VM)" · 윈도-5 격리 `iso5/os_test.py` 줄
- EDIT 에디터 정정: `REF-process-exec-edit.md`

## HISTORY 색인 행 (원문)
### `history/windows-worker.md` — Windows worker 계획·결정 (2026-10-15 신설, 2026-10-19 마감)
- 2026-10-19 결정 WW1(실행 규칙)=개정(`.ps1` 표 + `cmd /c` 파일 연결) · WW2(종료 방식)=추천안 · WW3(에디터 기본값)=나 · WW5(OS 표시)=나 확정 + 윈도-0(빌드 분리)·윈도-3(실행 규칙)·윈도-4(에디터)·윈도-5(OS 표시) 구현 + VM 확인 + 사용자 직접 확인(8012 연결·재접속) → 마감
- 2026-10-15 Windows 빌드 실패 원인 = `pty` 패키지뿐 확인 + 계획 윈도-0~6·결정 WW1~WW5(코드 없음)

### `history/windows-conpty.md` — Windows ConPTY 구현 (2026-10-19 분할 신설)
- 2026-10-19 윈도-1(ConPTY 실행)·윈도-2(종료) 설계 → 구현 → 윈도-6 VM 15/18 → Ctrl+C 상속 수정 → 17/17

### `history/windows-exec.md` — Windows 실행 규칙·에디터 구현 (2026-10-19 분할 신설)
- 2026-10-19 분할 기록(윈도-3·윈도-4 구현 이력 본문은 `history/windows-worker.md` 2026-10-19)

## 모듈 현황표 행 (원문 일부)
- process 실행 행: "Windows worker(ConPTY·실행 규칙·에디터·OS 표시) 마감(2026-10-19) → `INDEX-windows-worker.md`"

## 이 파트에만 해당하는 결정
- WW1(실행 규칙): `.ps1` = `powershell -NoProfile -ExecutionPolicy Bypass -File` / 그 외 확장자 = `cmd /c`(Windows 파일 연결) / 확장자 없음 = 400(supervisor `processApi`) + worker 거절. 확장자 = `NodeExt`(영문·숫자 1~8자)
- WW2(종료 방식): 요청 = ConPTY 입력 `0x03` / 강제 = `TerminateJobObject` 137 / `KILL_ON_JOB_CLOSE` / Ctrl+C 종료 exit code는 그대로
- WW3(에디터 기본값): `$VISUAL` > `EDITOR` > Windows PATH `edit`->`nvim`->`vim`->`nano`(Linux = `vi`), 인자 분리는 공백만
- WW4(확인 환경): VM `dockurr/windows`(남겨 둠) / WW5(OS 표시): 등록 OS + `worker_instances.os`, 빈 값 = linux, 프론트 L/W + 툴팁
- `SCRIPT_FORMAT` = `.ext:인코딩,줄끝;...`(`utf8|utf8bom|cp949` × `keep|lf|crlf`), EDIT는 손질 안 함
- 마감 때 미확인(다시 볼 조건 = 실제로 쓸 때): `.py`(Python 설치 장비) · EDIT 저장 CRLF · TEMP 경로 공백
