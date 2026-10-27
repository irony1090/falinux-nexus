# REF — 파트(폴더에서 터미널): 장비 타일 "여기서 셸"
> → `history/dir-shell.md` · → `CURRENT.md` · 바탕 = `INDEX-worker-dir.md`(폴더-4(실행) = `REF-worker-dir-exec.md`)

## 결정 (2026-10-27 사용자)
| 항목 | 결정 |
|---|---|
| 기능 | 장비 디렉토리 타일에서 **지금 보는 폴더를 cwd로 대화형 셸**을 새 터미널 타일로 연다 |
| 누가 | **장비 주인만** = 기존 `requireDeviceDir`(Q5(누가 열 수 있나) = 폴더 소유자 + 장비 주인) 그대로. 셸은 `cd ..`로 폴더 밖까지 가므로 공유받은 사람에게 열지 않는다 |
| 버튼 | 경로 줄 새로고침(`mdi-refresh`) 바로 오른쪽 `mdi-console`. 켜짐 = `canRun`(저장 중 아님·탭 준비됨·접속 중) + `cur` 있음(Windows 드라이브 목록 제외) + 목록 에러 없음 |
| 셸 | Linux = worker의 `$SHELL`(절대 경로일 때) > PATH의 bash > `/bin/sh` / Windows = **`powershell.exe`**(Windows PowerShell, PATH에 없으면 `%SystemRoot%\System32\WindowsPowerShell\v1.0\`) |
| 셸 바꾸기(Linux) | 따로 만든 env 항목 없음. worker env 파일에 `SHELL="/bin/zsh"`를 넣으면 `godotenv.Overload`가 프로세스 환경 변수로 올려 `$SHELL` 단계에서 쓰임. systemd처럼 `$SHELL`이 없으면 bash |
| 업로드 실행 권한 | 따로 안 고침 — 이 셸에서 `chmod +x` 후 새로고침하면 "실행"이 붙음(BACKLOG "업로드 파일 실행 권한", 파트 마감 때 삭제) |

## 구현 (2026-10-27)
대화에서 말한 `shell: true` 필드 대신 **치환 토큰**으로 했다 — 기존 `{WORKER_EDITOR}`와 같은 방식이라 API·DB를 안 바꿈.

| 층 | 파일 | 내용 |
|---|---|---|
| protocol | `messages.go` | `PlaceholderWorkerShell = "{WORKER_SHELL}"` |
| worker | `process.go` `localize` | 토큰이 있을 때만 `shellPath()`로 치환(→ 절대 경로라 기존 `prepareDirect` 검사·`directRunner`를 그대로 지남: Linux 실행 비트 / Windows `.exe` 그대로) |
| worker | `dir_other.go` / `dir_windows.go` | `shellPath()` (OS 코드 분리 규칙) |
| supervisor | `workerExec.go` | 변경 없음(`cmd` = 토큰, 주석만). DB `processes.cmd`에 토큰 그대로 저장 → 재접속·`Entry.Spec()` 복원 때도 같은 토큰 |
| supervisor | `tileNotice.go` `fileBase` | 토큰이면 `"셸"` → 위치 알림 이름·실행 기록(`nodeRuns.go`) 이름 |
| 프론트 | `tileLabel.util.ts` | `SHELL_CMD`(서버 토큰과 같게) · `fileLabel` 토큰 = `"셸"` |
| 프론트 | `DeviceTileBody.vue` | `openShell` = `execFile(..., SHELL_CMD, cur)`, `.refresh, .shell` 같은 폭 |

- 터미널 타일 = 파일 실행과 같음(장비 타일의 자식, 경로 표시 = cwd 끝 두 단계). `exit` = COMPLETED(0)
- 가져오기(미등록 process)로 다시 붙이면 cmd = worker가 쥔 실제 경로(`/bin/bash`)라 이름 = `bash`

## 확인 (2026-10-27, 격리 5051 + `nexus_sh_tmp` + worker `sh-dev` + vite 3001, 헤드리스)
- 버튼 위치·켜짐 OK(스크린샷) · 경로 입력으로 픽스처 `fx` 이동 → 셸 → `pwd` = `fx`, `$0` = `/bin/bash`
- 644 `run.sh`: 셸에서 `chmod +x` → 새로고침 → "실행" 버튼 0 → 1
- 타일 이름 "셸"·경로 `…/sh/fx`·부모 = 장비 타일 / 실행 기록 `name: 셸`, `nodeId: null` / `exit` → COMPLETED 0
- Windows: `GOOS=windows go build ./cmd/worker/...` 통과
- **사용자 직접 확인(2026-10-27)**: 같은 격리(supervisor `0.0.0.0:5051`, 시험 계정 `sh1`) + VM `win-dev`(`shared/worker-shell.exe` + `shared/start-shell.ps1`, env `C:\nexus\env-shell`, 로그 `C:\nexus\worker-shell.log`) → "확인 끝" → **파트 마감**(2026-10-27, → `INDEX-dir-shell.md`)
- 정리 끝(두 번 다): 프로세스(exe·cwd로 고름)·DB DROP·`/tmp/nexus-hold-1000/sh-dev` / VM worker `taskkill /IM worker-shell.exe` + `docker stop nexus-win`. 5051·3001·8006 닫힘, 8013 그대로
