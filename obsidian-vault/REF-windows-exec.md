# REF — Windows worker: 실행 규칙 · 에디터 · OS 표시 구현 (윈도-3 · 윈도-4 · 윈도-5)

> 2026-10-19 `REF-windows-worker.md`에서 분할(크기 기준). 결정 WW1(실행 규칙)·WW3(에디터 기본값) 본문 → `REF-windows-worker.md` / ConPTY → `REF-windows-conpty.md` / 이력 → `history/windows-exec.md`(2026-10-19 이전분은 `history/windows-worker.md`) / 현재 진행 → `CURRENT.md`

## 스크립트 형식 손질 (2026-10-19 확정·구현, 윈도-3)
- DB 본문 = UTF-8(BOM 없음)·LF → 해석기가 기대하는 형식으로 **EXEC 때 worker 쪽 파일만** 바꿈(DB·EDIT 무관). 문법이 아니라 저장 방식 차이라 "문법은 그 OS 그대로" 전제와 안 부딪힘
- **기본값(코드) + worker env `SCRIPT_FORMAT` 확장자별 덮어쓰기** — 장비마다(`pwsh`, 영문 Windows 등) 재배포 없이 조정
| 항목 | 값 |
|---|---|
| 형식 | `SCRIPT_FORMAT=.ps1:utf8bom,crlf;.bat:cp949,crlf` (줄 끝 생략 = `keep`, 잘못된 항목은 로그 후 무시) |
| 인코딩 | `utf8`(그대로) / `utf8bom` / `cp949`(`x/text` korean, CP949에 없는 문자는 `?` + 로그) |
| 줄 끝 | `keep` / `lf` / `crlf` |
| Windows 기본값 | `.ps1` = `utf8bom,crlf`(PS 5.1이 BOM 없으면 CP949로 읽어 한글 깨짐) · `.bat`·`.cmd` = `cp949,crlf`(한글·`goto` 오동작 해결) · 나머지 그대로 |
| Linux 기본값 | 없음(지금과 같음). env를 주면 그것만 적용 |
- 확인(2026-10-19): `go test -overlay` 임시 테스트 — 형식 파싱 5경우 · BOM+CRLF(두 번 적용해도 같음) · CP949 왕복 · 이모지 `?` · `utf8,keep` 무변경 · `NodeExt` 8경우 통과. Linux·Windows 빌드·vet 통과

## 윈도-4(에디터) 구현 메모 (2026-10-19)
- `router/process.go`: `resolveEditor` = `$VISUAL` > `EDITOR`(env) > `defaultEditor()` / `prepareEdit` = 공백으로 Cmd·Args 펼침 + 파일 경로 `filepath.Clean`, 에디터가 비면 Accept=false("worker env에 EDITOR를 설정") / exec 흐름 = `allLocalize` -> EDIT면 `prepareEdit`, 아니면 `prepareExec`(editPath는 펼치기 전 파일 경로)
- `defaultEditor`: `script_other.go` = `vi` / `script_windows.go` = PATH에서 `edit`·`nvim`·`vim`·`nano` 첫 번째
- `EDITOR` 실제 순서: `godotenv.Overload`라 **env 파일 `EDITOR`가 OS `EDITOR`를 덮음**(기존 동작, 그대로 둠)
- 주의: `$VISUAL`이 GUI 에디터(`code --wait` 등)면 PTY에 안 보임 — worker를 띄우는 셸의 `VISUAL` 확인
- EDIT는 `SCRIPT_FORMAT` 손질 안 함(read-back이 DB로 가므로). Windows 에디터가 CRLF로 저장하면 DB 줄 끝이 바뀜 → "위험" 절, 윈도-6에서
- 확인: `go test -overlay`(scratchpad `w4/`) — 기본 `vi`·`VISUAL` 우선·`"code  --wait"` 분리·경로 정리·빈 에디터 거절 + 윈도-3 테스트 재통과, Linux·Windows 빌드·vet 통과

## 윈도-5(OS 표시) 구현 (2026-10-19)
| 층 | 변경 |
|---|---|
| 프로토콜 | `RegisterRequest.OS`(`omitempty`) + `protocol.NormalizeOS`(빈 값 = linux, 옛 worker·이전 명부 행) |
| worker | `router/register.go`가 `runtime.GOOS`를 실음 |
| supervisor 메모리 | `workerOS` KeyValManager(InstanceKey -> OS), 등록 때 `workers`보다 먼저 넣고 끊길 때 뺌(`workers` 값 타입은 안 바꿈 — 사용처 11곳) |
| DB | `00011_worker_os.sql`(`os TEXT NOT NULL DEFAULT ''`) + `UpsertWorkerInstance`가 `os = EXCLUDED.os`로 갱신, `recordSeen(key, connected, os)` |
| API | `GET /workers`(메모리)·`GET /workers/instances`(DB) 응답에 `os`(정규화) |
| 실행 검사 | `processApi.go` 장비 일치 검사 다음: 대상 windows + EXEC + `process.NodeExt(이름) == ""` -> 400 "Windows에서는 노드 이름에 확장자(...)가 필요합니다". EDIT는 검사 안 함 |
| 프론트 | `feature/worker/util/workerOs.util.ts`(`osBadge` = L/W 한 글자, `osTitle` = 툴팁 "OS: Windows", `osName`) → `WorkerPickDialog`(subtitle 앞) · `NodeDeviceDialog`(인스턴스 줄 점 다음) · `FolderTileBody`(장비 칩 툴팁 줄에 `(Windows)`). **모든 인스턴스에 표시**(2026-10-19 사용자 결정: L/W + 툴팁). 세 곳 다 기본 화면에선 가려짐(칩 툴팁 · 장비 창 펼침 · 인스턴스 2개 이상일 때 고르기 창) — 사용자 브라우저 확인 후 "이대로"(장비 칩 자체 표시는 안 넣음) |

- 확인(2026-10-19, 격리 = 임시 DB `nexus_os_tmp` + 5051, 끝나고 프로세스 종료·DB 삭제): `00011` 적용 / 실제 Linux worker 등록 = 명부 `os=linux` / 가짜 worker(Python 소켓 `REGISTER`) `os=windows`·`os` 없음 -> API `windows`·`linux` / windows EXEC 확장자 없음·`v1.2 정리` = 400 / windows EDIT·`run.ps1` = 확장자 검사 통과(경로 `{WORKER_BASE}/28/<uid>.ps1`) / Linux EXEC 확장자 없는 경로 COMPLETED 0 / `EDITOR="cat -n"` EDIT COMPLETED 0(인자 분리 실동작) + read-back 내용 유지 → 14/14(끝난 process는 `GET /processes`에 안 나와 DB로 확인). `vue-tsc` 통과. 스크립트 명세 → `REF-test-scripts-1014.md` "2026-10-19"
- 적용 주의: 사용자가 띄운 supervisor는 재시작 때 `00011`이 자동 적용(추가 칸이라 무해). worker는 새 빌드여야 OS를 보냄(옛 빌드도 linux로 표시돼 문제없음)

## 윈도-6(확인) 결과 — 실행 규칙·에디터·OS 표시 (2026-10-19)
| 확인 | 결과 |
|---|---|
| 등록 `os=windows`(실제 worker.exe) | 통과 |
| `.ps1` -> powershell + BOM(한글) / `.bat` -> `cmd /c` + CP949·CRLF(한글, `goto` 정상) | 통과 |
| 확장자 없음 -> 400 | 통과 |
| `.py` -> `cmd /c`(파일 연결) | **멈춤** — 이 VM엔 Python이 없음(`python.exe` = Store 껍데기, `.py` 연결 없음) -> 연결 선택 창(GUI)이 세션 0에서 안 보이고 대기. 문서화한 "GUI 연결은 안 보임" 한계 그대로. 실제 Python 설치 시 동작은 미확인 |
| EDIT 기본 에디터 = `edit.exe`(Windows 11 25H2에 **기본 포함**) | 통과, 화면 정상 렌더, Ctrl+Q로 종료 COMPLETED, 저장 안 하면 DB 내용 그대로 |
| 참고 | Microsoft Edit 상태 줄에 `[CRLF]` — LF 파일을 열어도 저장하면 CRLF가 될 가능성(→ "위험" 절 EDIT 줄 끝, 미확인) |
