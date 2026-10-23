# REF — worker 폴더 타일: 폴더-4(실행)

> → `history/worker-dir-exec.md` / 앞 단계(폴더-1·2·3·뒤로-1) → `REF-worker-dir-list.md` / 상위 계획 → `REF-worker-dir.md` / → `CURRENT.md`
> 2026-10-23 `REF-worker-dir-list.md`에서 분할(10k자 근접).

## 폴더-4(실행) 세부 설계 (2026-10-22 밤 제시, 2026-10-23 확정)
| 항목 | 설계안 |
|---|---|
| DB | 마이그레이션 `00012`: `processes.dir_node_id BIGINT NULL REFERENCES nodes ON DELETE SET NULL`. 장비 실행 = `node_id` NULL + `dir_node_id` = 연 폴더(Q15). 권한 = 기존 `process_permission` 그대로(node 없음 -> 소유자·process 공유만 = Q3 공유 없음과 맞음) |
| REST | `POST /workers/:mainKey/:subKey/exec` `{folderId, cmd(파일 전체 경로), cwd(지금 보는 폴더), workspace, parentTileId, size}` -> 인가 = 목록과 같음 -> 응답 = 기존 실행 응답(터미널 타일 포함). cwd를 프론트가 주므로 서버는 경로 구분자를 안 다룸 |
| 실행 | 파일 직접 실행(Linux 실행 비트·shebang, Windows 확장자 실행기 = 기존 `prepareExec`). 본문 선배치 없음 -> manager에 파일 실행 등록 함수 하나 |
| 터미널 타일 | 장비 타일의 자식. 서버 터미널 check의 nodeId 필수 -> 선택값, 프론트 `TerminalTile.nodeId` = `number | null`. 이름 = cmd 끝 이름, 경로 표시 = cwd |
| 목록(Q10) | `ListVisibleRunsInSubtree`: `JOIN nodes` -> 스크립트 또는 `dir_node_id ∈ 하위`, 이름 = 노드 이름 또는 cmd 끝 이름 |
| Q9(루트 장비 탭) | 폴더-6(보드 이동) 때 |

사용자 답(2026-10-23): **Q23(누르는 방식) = 행의 "실행" 버튼** · **Q24(실행 대상) = 실행 비트(Linux)·실행기 확장자(Windows) 있는 파일만** · 인자 입력은 나중. 착수 전에 Windows VM에서 폴더-1·2 확인 먼저(사용자 요청)

## 구현 (2026-10-23)
| 층 | 파일 | 내용 |
|---|---|---|
| DB | `migrations/00012_process_dir.sql` | `processes.dir_node_id`(FK nodes, ON DELETE SET NULL) + 부분 인덱스 |
| 쿼리 | `processes.sql` | `CreateProcess`에 `dir_node_id`($13) · `ListOwnLiveProcesses`에서 `node_id IS NOT NULL` 뺌(장비 실행도 고아 안전망이 받음) |
| 쿼리 | `shares.sql` | `ListVisibleRunsInSubtree` = `LEFT JOIN nodes` + `dir_node_id ∈ 하위` + `(n.kind = 'SCRIPT' OR p.node_id IS NULL)`(지워진 스크립트 노드의 process는 예전처럼 빠짐). 두 runs 쿼리에 `p.cmd, p.cwd`(열 맞춤 — 하위 행을 폴더 행 타입으로 변환함) |
| protocol | `ProcessSpec.Direct` | 장비 파일 그대로 실행. supervisor `Entry.Spec()`이 "노드 없는 EXEC"면 켬 |
| worker | `dir_exec.go` `prepareDirect` | 절대 경로·`os.Stat`(링크 = 원본)·`execOK`·cwd 폴더 확인 -> `directRunner`. **형식 변환(`applyScriptFormat`) 안 함** — 사용자 파일을 고쳐 쓰면 안 되므로(기존 스크립트 경로는 노드 본문을 선배치한 사본이라 변환해도 됨) |
| worker | `dir_other.go` / `dir_windows.go` | `execOK` = Linux 실행 비트 / Windows 확장자가 `PATHEXT`(없으면 .COM;.EXE;.BAT;.CMD) 또는 .ps1. `directRunner` = Linux 그대로 / Windows .exe·.com 그대로, 그 외 `scriptRunner`(기존 `prepareExec`에서 뽑음: .ps1 = powershell -File, 그 외 cmd /c) |
| worker | `dir.go` `fill()` | `Exec` = `execOK`, 파일 링크는 원본 기준 |
| supervisor | `process.go` | `Exec`의 3~5단계를 `launch`로 뽑고 `ExecFile` 추가. worker 거부 = `errExecRejected`(문구 그대로) |
| supervisor | `manager.go` | `ExecFile` + 공통 `register`(memory 등록 -> pool 영속, 실패 시 롤백) |
| supervisor | `workerExec.go` | `POST /workers/:mainKey/:subKey/exec {folderId, cmd, cwd, workspace, parentTileId, size}`. 인가 = `requireDeviceDir`(`workerDir.go`로 뽑음, 목록과 같음) -> `EndTx` -> `ExecFile`. 거부 400 / 그 외 500 |
| supervisor | `processApi.go` | `execTileSize`·`execTile`로 뽑아 스크립트 실행과 같이 씀. 장비 실행은 `attachViewerTiles` 안 부름(볼 계정 = 소유자뿐) |
| supervisor | 타일 | 터미널 check의 nodeId 필수 해제 · `newTerminalTile(nodeID *int64)` · `bindProcessTile`이 노드 없는 process도 받음 · 알림 이름 = `procName`(노드 이름 또는 `fileBase(cmd)`, `\`도 구분자) · runs 응답 `nodeId` null + `cwd`, 이름 = 파일 이름 |
| 프론트 | `process.api.ts` `execWorkerFile` / `processTerm.store` `execFile` / `nodeOps.store` `execFile` | **실행 전에 `flush()`**(타일 트리 저장 대기분) — 아래 "함정" |
| 프론트 | `DeviceTileBody.vue` | 행 끝 "실행" 버튼(`e.exec`, 접속 중·탭 준비됨일 때), cwd = 그 파일이 든 폴더(`parentPathOf`) |
| 프론트 | `TerminalTile.vue`·`RunRow.vue` | nodeId null = 이름 `fileLabel(cmd)`, 경로 = `shortPath(cwd)`(끝 두 단계, 전체는 title) |
| 프론트 | `tileGrids.store` `instanceOf`·`scriptRuns.hook` | nodeId null은 번호·노드별 묶음에서 뺌 |

### 함정
- **막 열거나 옮긴 장비 타일에서 바로 "실행"하면 터미널이 루트에 붙음**(헤드리스에서 발견): 프론트 타일 트리 저장이 지연돼 서버 트리에 장비 타일이 아직 없음 -> `appendTerminalTile`이 부모를 못 찾아 루트로. `execFile` 앞에 `flush()`로 해결. **스크립트 실행(`nodeOps.start`)도 같은 구조**(새로 연 폴더 타일에서 바로 실행) — 사용자 요청으로 이 파트에서 같이 고침(아래 "확인")
- 화면 경로: 장비 경로가 길면 이름이 가려짐 -> `shortPath`

## 확인 (2026-10-23)
- 격리 REST 31/31(scratchpad `iso/exec_test.py`, 세션 끝나면 사라짐): 목록 실행 표시 · 실행 200 · process 필드(nodeId null·cmd·cwd·단계 40) · 장비 타일 아래 터미널 · 끝남 exit 0 · 출력 pwd·인자 0 · **파일 안 바뀜(md5)** · runs deep(root·child·전체 = 포함, 다른 폴더·얕은 목록 = 제외) · 공백 이름 · 실행 파일 링크 · 오래 도는 것 · 거부(실행 비트 없음·없는 파일·폴더·상대 경로·없는 cwd = 400, 탭 없음 400, 다른 계정 404, 장비 없는 폴더 403, 끊긴 인스턴스 409) · 거부된 것이 실행 중으로 안 남음 · 타일 지운 뒤 고아 안전망이 루트에 붙임(nodeId null·adopted) · 장비 타일로 바인딩 · 입력 되돌아옴 · 종료
  - 한 번(9회 중 1회) "공백 이름" 출력이 시험 대기(끝난 뒤 0.3초) 안에 안 잡힘, 종료 코드는 0. 재현 안 됨
- 격리 헤드리스 13/13(scratchpad `pw/exec.mjs`): 실행 버튼 있음/없음 · 실행 -> 터미널 타일(이름 hello.sh, 경로 `…/iso/fx` + title 전체) · 출력·pwd · 장비 타일 그대로 · 서버 트리에서 장비 타일의 자식 · 오래 도는 것에 입력 · 루트 터미널 탭에 장비 실행 2개 · 새로고침 뒤 복원
- Windows VM REST 11/11(scratchpad `win/exec_win.py`, `start-dir.ps1`에 픽스처 추가): 실행 표시(.ps1·.bat·공백 .bat = 예, .txt = 아니오) · .ps1이 cwd에서 · **.ps1·.bat 해시 그대로(BOM·CP949 변환 안 됨)** · .bat `cd` = fx · 공백 이름 .bat · `hostname.exe` 직접 · runs 이름·cwd · .txt·junction 폴더·상대 경로 400
- `go test`·`go vet`(Linux·GOOS=windows)·`vue-tsc`·`vite build`(scratchpad) 통과. VM은 확인 뒤 끔
- 스크립트 실행 `flush()`(헤드리스 scratchpad `pw/script_flush.mjs`): 폴더 타일 "새 타일"로 열고 기다리지 않고 그 안 스크립트 "실행" -> 터미널 부모 = 새 폴더 타일. 고친 뒤 8/8, `flush` 빼고 같은 시험 0/8(전부 루트) = 실제 있던 버그. 함정: REST로 노드 내용을 넣을 땐 `{"content": {"valid": true, "value": ...}}`(PATCH 필드 감싸기, 그냥 문자열이면 400)
- 미확인: 장비 끊김 중 실행 버튼 비활성 표시, 504(worker 무응답), 모바일, 사용자 브라우저
