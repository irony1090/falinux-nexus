# REF — Windows worker: ConPTY 구현 (윈도-0 · 윈도-1 · 윈도-2)

> 2026-10-19 `REF-windows-worker.md`에서 분할(크기 기준). 파트 계획·결정 WW1(실행 규칙)~WW5(OS 표시) → `REF-windows-worker.md` / 이력 → `history/windows-conpty.md` (윈도-0 이력은 `history/windows-worker.md` 2026-10-19) / 현재 진행 → `CURRENT.md`

## 윈도-0(빌드 분리) 결과 (2026-10-19 완료, 커밋 전)
| 파일 | 내용 |
|---|---|
| `pty/execInteractive.go`(공통) | `Interactive`(+ `sys sysProc` 필드) · `newInteractive` · 상태 큐 메서드 · `streamOutputLogic(io.Reader)`/`streamInputLogic(io.Writer)` · `publishLine`·`newLineBytes` |
| `pty/execInteractive_linux.go` | `openPty` · `sysProc{master, cmd}` · `Kill`·`Layout`·`Pid`·`exitCodeOf`·`Winsize`·`ExecInteractive` (옛 코드 그대로 이동) |
| `pty/execInteractive_windows.go` | `sysProc{}` 빈 구조체 + 같은 이름 함수가 `errNotImplemented` 반환 → 윈도-1에서 채움 |
| `pty/execFIFO_linux.go` | 이름만 변경(`git mv`) |
- 확인: Linux `go build ./...`·`go vet` / `GOOS=windows go build ./cmd/worker`·`./cmd/supervisor`·`go vet ./internal/execute/pty` 통과 / Linux PTY 동작 = `go test -overlay`로 임시 테스트를 패키지에 끼워 실행(저장소에 파일 안 남김): 한글 출력·`Layout(100,30)` -> `stty size` = `30 100`·입력 왕복·exit 3 / `Kill(false)` = 143 · `Kill(true)` = 137 전부 통과

## 윈도-1(ConPTY 실행) · 윈도-2(종료) — 설계대로 구현(2026-10-19, 커밋 전, 실행 미확인)
API는 전부 `golang.org/x/sys/windows` v0.46.0에 있음(확인함): `CreatePseudoConsole`·`ResizePseudoConsole`·`ClosePseudoConsole`·`NewProcThreadAttributeList`·`PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE`·`CreateJobObject`·`SetInformationJobObject`·`AssignProcessToJobObject`·`TerminateJobObject`·`StartupInfoEx`·`CREATE_SUSPENDED`·`ResumeThread`·`GetExitCodeProcess`·`ComposeCommandLine`

### 시작 순서 (`ExecInteractive`)
| 순서 | 하는 일 | 틀리면 |
|---|---|---|
| 1 | `CreatePipe` 2쌍: 입력(ConPTY 읽기 끝 / worker 쓰기 끝), 출력(ConPTY 쓰기 끝 / worker 읽기 끝) | — |
| 2 | `CreatePseudoConsole(크기, 입력 읽기 끝, 출력 쓰기 끝, flags=0)` | flags에 `INHERIT_CURSOR`를 주면 ConPTY가 `ESC[6n`을 보내고 응답을 기다림 -> 응답할 탭이 없으면 멈춤 |
| 3 | worker 쪽에서 **ConPTY에 넘긴 두 끝(입력 읽기·출력 쓰기)을 닫음** | 안 닫으면 출력 읽기가 EOF를 영원히 못 받음(worker가 쓰기 끝을 들고 있으므로) |
| 4 | 속성 목록 `PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE = hpc` | — |
| 5 | `CreateJobObject` + `KILL_ON_JOB_CLOSE` 설정 | — |
| 6 | `exec.LookPath(command)`로 실행 파일 경로(PATHEXT 해석) + `ComposeCommandLine` + 환경 블록(UTF-16, `env == nil`이면 상속) + 작업 폴더 | `CreateProcess`는 PATHEXT를 안 봄 |
| 7 | `CreateProcess(..., inheritHandles=false, EXTENDED_STARTUPINFO_PRESENT \| CREATE_SUSPENDED \| CREATE_UNICODE_ENVIRONMENT)` | 상속을 켜면 worker 핸들(소켓 등)이 스크립트로 새어 나감 |
| 8 | `AssignProcessToJobObject` -> `ResumeThread` -> 스레드 핸들 닫기 | **멈춘 채 job에 넣은 뒤 재개**해야 함. 먼저 재개하면 job에 넣기 전에 자식을 띄워 그 자식이 job 밖으로 빠짐 |
| 9 | 입력 쓰기 끝·출력 읽기 끝을 `os.NewFile`로 감싸 공통 `streamInputLogic`·`streamOutputLogic`에 넘김 | — |

### 끝날 때 순서 (대기 고루틴)
| 순서 | 하는 일 | 틀리면 |
|---|---|---|
| 1 | `WaitForSingleObject(프로세스)` -> `GetExitCodeProcess` | — |
| 2 | `closed = true`(뮤텍스 안) — 이후 `Layout`·`Kill`은 아무것도 안 함 | 닫힌 핸들 값을 OS가 다른 객체에 재사용 -> 엉뚱한 것을 리사이즈·종료 |
| 3 | `ClosePseudoConsole` | 출력 읽기 고루틴이 돌고 있어야 함 — 옛 Windows 10은 출력 파이프가 비워질 때까지 이 호출이 막힘 |
| 4 | 출력 고루틴 EOF를 기다림(안전 상한 3초) | 먼저 `setStatusCompleted`(큐 닫기)하면 마지막 출력 유실 |
| 5 | 입력·출력 파일, 프로세스, job 핸들 닫기 | job 닫기 = `KILL_ON_JOB_CLOSE`로 남은 자손 정리(Linux에서 세션 리더 종료 뒤 SIGHUP과 같은 효과) |
| 6 | `setStatusCompleted(exitCode, err)` | — |

### 메서드
| 메서드 | Windows 구현 |
|---|---|
| `Kill(false)` | 입력 큐에 `0x03` -> ConPTY가 CTRL_C_EVENT로 바꿈. WW2(종료 방식) |
| `Kill(true)` | `TerminateJobObject(job, 137)` -> exit code 137 |
| `Layout` | `ResizePseudoConsole(hpc, Coord{cols, rows})` |
| `Pid` | `CreateProcess`가 준 ProcessId |
| `sysProc` | `hpc`·`job`·`process` 핸들 + pid + `mu`·`closed` |
- worker 종료(`router/shutdown.go`)와 kill(`router/process.go`)은 둘 다 `Interactive.Kill`을 부르므로 이것만 구현하면 TERM -> 5초 -> KILL 흐름이 그대로 동작
- `ctx`: 지금 호출자는 `context.Background()`뿐. `ctx.Done()`이 오면 `TerminateJobObject`(Linux `CommandContext`와 같은 의미)
- Linux 쪽 "출력 EOF 기다린 뒤 Completed"는 손대지 않음(현재 동작 확인됨) — Windows만

### 구현 메모 (2026-10-19)
- 파일 = `pty/execInteractive_windows.go` 하나. `sysProc{c *conpty}` — 뮤텍스가 있어 포인터(값 복사 시 vet copylocks)
- `go.mod`: `golang.org/x/sys v0.46.0`이 indirect -> direct(`go mod tidy`, 버전 그대로)
- 설계에 없던 것: `si.Flags = STARTF_USESTDHANDLES`(핸들 0) — worker 표준 입출력이 리다이렉트돼 있으면(nohup 로그 등) 자식이 그걸 물려받는 문제 회피. 오픈소스 Go ConPTY 구현들이 쓰는 방식으로 기억 — **윈도-6에서 확인 1순위**
- 속성 값: `Update(PSEUDOCONSOLE, *(*unsafe.Pointer)(unsafe.Pointer(&hpc)), ...)` — HPCON 값 자체를 넘김(vet unsafeptr 회피 변환)
- exit code = `GetExitCodeProcess`의 uint32를 `int`로(Ctrl+C면 3221225786 = `0xC000013A`). 0이 아니면 `exit status N` 에러
- 환경 블록: 같은 키는 뒤의 값(대소문자 무시) — worker가 `os.Environ()` 뒤에 `TERM`을 덧붙이므로 필요
- 출력 EOF 3초 넘기면 `outFile.Close()`를 고루틴으로(막힌 `ReadFile`이 끝나야 Close가 반환)

### 확인 상태
- 이 장비: `GOOS=windows go build`(worker·supervisor)·`GOOS=windows go vet ./...`·Linux `go vet ./...` 통과. wine 없음 → 실제 실행은 윈도-6(확인)에서 VM으로
- 윈도-6 확인 목록(이 구현분): 표준 입출력 리다이렉트한 worker에서 출력 보임 / 한글 / 크기 바꾸기 / Ctrl+C 종료·강제 137 / 손자 프로세스까지 강제 종료 / worker 강제 종료 시 스크립트도 꺼짐(`KILL_ON_JOB_CLOSE`) / 마지막 출력 유실 없음 / 끝난 뒤 `Layout`·`Kill` 무해

## 윈도-6(확인) 결과 — ConPTY·종료 (2026-10-19, VM `nexus-win`)
환경: Windows 11 25H2(26200) 한국어, 코드 페이지 949, PowerShell 5.1, worker = WMI(`Win32_Process.Create`)로 **세션 0(Services)** + 출력 파일 리다이렉트(`cmd /c worker.exe > worker.log`) / 격리 supervisor `172.17.0.1:5051`(`nexus_win_tmp`) / 스크립트 `iso6/win_test.py`
| 확인 | 결과 |
|---|---|
| 출력 리다이렉트한 worker에서 스크립트 출력 보임(`STARTF_USESTDHANDLES`) | **통과** |
| 세션 0(서비스 세션)에서 ConPTY | **통과** |
| 한글 출력(ps1 BOM / bat CP949, ConPTY는 UTF-8 VT로 내보냄) | **통과** |
| 크기 바꾸기(`ResizePseudoConsole`) -> `$Host.UI.RawUI.WindowSize` 100x30 | **통과** |
| exit code(ps1 `exit 3` / bat `exit /b 4`) | **통과** |
| 강제 종료 137 + 손자(`Start-Process ping`)까지 job으로 종료 | **통과** |
| worker 강제 종료(`taskkill /F`) -> 실행 중 스크립트 전부 종료(`KILL_ON_JOB_CLOSE`) | **통과** |
| ConPTY 시작 시퀀스 | `ESC[?9001h ESC[?1004h ESC[?25l ESC[2J ...`(win32-input-mode·포커스 요청, xterm.js는 모르는 모드 무시) |
| **종료 요청(Ctrl+C)** | **실패 -> 원인 확인**: WMI로 띄운 worker에 "Ctrl+C 무시" 속성이 걸려 있고 **자식이 상속**(Linux 고아 스크립트의 SIGHUP 무시 상속과 같은 구조, → `REF-worker-orphan.md`). overlay 실험: worker 시작 때 `SetConsoleCtrlHandler(NULL, FALSE)` 한 줄 -> Ctrl+C로 루프 종료(COMPLETED 0) 확인. **저장소 반영(사용자 승인) -> 재확인 통과** |
- `SetConsoleCtrlHandler`는 `x/sys/windows`에 없음 -> `NewLazySystemDLL("kernel32.dll").NewProc(...)`
- 반영 위치: `execInteractive_windows.go` `allowCtrlCInChildren()` — `ExecInteractive` 맨 앞에서 `sync.Once`로 한 번(worker 자신의 속성을 되돌려 이후 자식이 정상 상속). 띄우는 방식(WMI·서비스·작업 스케줄러)과 무관하게 동작
- 반영 후 저장소 빌드로 재실행: `win_test.py ps bat noext ctrlc force edit` **17/17 통과**(Ctrl+C -> COMPLETED 0). `.py`는 제외(Python 없음)
