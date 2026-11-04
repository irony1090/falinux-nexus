# REF — 업데이트 app P4(Windows)

> 2026-11-03 신설. 설계 본문 → `REF-update-deploy.md`("단계 계획"·"함정"·"대상의 의존성") · 규칙 → `REF-update-rules.md` · P3 구현 → `REF-update-p3.md` · Windows worker 실행 규칙 → `REF-windows-worker.md`·`REF-windows-exec.md` · 집 서버 배포 방식 → `REF-home-deploy.md` · 이력 → `history/update-p4.md` · 현재 진행 → `CURRENT.md`

**상태(2026-11-03 오후): P4(Windows) 마감 · master 머지(`feat/updater-p4`).** 단위-1(스크립트 실행)~단위-5(빌드) 구현 + VM `nexus-win` 확인 39/39 + 집 서버 설치(`C:\Nexus\updater`, 사용자 "잘 동작했어"). **단위-7(교체 빌드)은 코드가 아니라 보관소 빌드라 P4에서 떼어 후속으로** → 아래 "단위-7(교체 빌드) 후속"

## 지금 코드에서 Windows가 비어 있는 곳 (2026-11-03 조사)
| 곳 | 지금 | 문제 |
|---|---|---|
| `cmd/updater/exec_other.go` | `setGroup` 빈 함수, `killGroup` = 부모 프로세스만 Kill | 시간 제한·취소 때 스크립트가 띄운 자식이 남음 |
| `cmd/updater/exec.go` `runPipe` | `exec.Command(path)` 로 파일을 바로 실행 | Windows에서 `.ps1`은 바로 실행이 안 됨(파일 연결 = 메모장). 해석기를 붙여야 함 |
| 스크립트 인코딩 | 보관소 파일 그대로(UTF-8, BOM 없음, LF) | PowerShell 5.1은 BOM 없는 파일을 CP949로 읽어 한글이 깨짐(`REF-windows-exec.md`) |
| `builtin.go` `copyOne` | 임시 파일 -> `os.Rename(tmp, dst)` | 실행 중인 exe는 Windows에서 덮어쓰기(이름 변경)가 거절됨. 설계 문서의 "잠긴 파일은 `*.old-<시각>`으로 물린 뒤 씀"이 **코드에 아직 없음** |
| interactive 단계 | `pty.ExecInteractive`(Windows = ConPTY, 있음) | ConPTY 쪽 Job이 `KILL_ON_JOB_CLOSE`라 interactive 단계에서 띄운 프로세스는 단계가 끝나면 같이 종료됨(집 서버 함정-A와 같은 원인) |
| 설치 | `scripts/updater/install.sh` + systemd 단위 | Windows 설치 스크립트·예약 작업 없음 |
| 빌드 | updater는 `go build`로만, `scripts/build.sh`에 없음 | Windows exe 빌드 절차 없음 |

## 착수 계획 초안 (2026-11-03, 조율 전)
| 단위 | 내용 |
|---|---|
| 단위-1(스크립트 실행) | 확장자별 해석기: `.ps1` = `powershell -NoProfile -ExecutionPolicy Bypass -File` / `.bat`·`.cmd` = `cmd /c` / `.exe` = 그대로. 실행 전에 staging 파일만 형식 변환(`.ps1` = UTF-8 BOM + CRLF, `.bat`·`.cmd` = CP949 + CRLF) — worker `cmd/worker/router/script.go`의 변환을 공용 패키지로 옮겨 같이 씀. Linux는 지금 그대로 |
| 단위-2(자식까지 종료) | Windows 비대화형 단계 = Job에 넣고 시간 제한·취소 때 `TerminateJobObject`. **`KILL_ON_JOB_CLOSE`는 안 씀** — 단계가 정상으로 끝났을 때 그 단계가 띄운 프로세스까지 죽이면 안 되므로 |
| 단위-3(잠긴 파일) | `copy`·`restore`: 이름 변경이 거절되면 기존 파일을 `<이름>.old-<시각>`으로 바꾼 뒤 다시. `*.old-*`는 지우지 않음(설계 그대로). Linux는 거절이 안 나니 동작 같음 |
| 단위-4(설치) | `scripts/updater/install.ps1`: 설치 폴더에 exe·`updater.json`(관리자만 읽게 ACL) + 예약 작업(부팅 때 시작·멈추면 다시 시작). 다시 실행 = exe 교체 + 재시작, 설정 유지(Linux `install.sh`와 같은 계약) |
| 단위-5(빌드) | updater Windows/amd64 빌드 절차(태그 = `cmd/updater/env` 첫 줄) |
| 단위-6(확인) | 로컬 자동 확인(`run.sh`에 Windows 항목은 못 넣음 — 변환·이름 변경 단위 테스트만) -> VM `nexus-win` -> 집 서버 |
| 단위-7(교체 빌드) | 집 서버용 빌드 예시 `steps.json`: supervisor·주 worker 끄기 -> `pg_dump` -> `backup` -> `copy` -> `Start-ScheduledTask Nexus` -> 응답 확인. `rollback` = `restore` -> 켜기, DB 복원은 `manual: true` |

## 조율 결과 (2026-11-03 사용자)
| 이름 | 결정 |
|---|---|
| 조-1(붙을 supervisor) | **가(집 서버 자신)**. updater 기능이 없는 지금 집 서버엔 첫 배포를 지난번처럼 Claude가 만든 `update.ps1` 묶음으로 수동 |
| 조-2(시험 순서) | VM `nexus-win` 먼저 -> 집 서버 |
| 조-3(설치 자리·계정) | 사용자 제안 `C:\Nexus\updater`(집 서버 Nexus 폴더 `C:\Nexus` 안) — 겹침·권한 보완 2개 확인 중(아래 조-3 보완) · 계정 SYSTEM |

### 조-3 보완 (2026-11-03 사용자 확인 "그 순서 맞아, 진행해줘")
- 보완-1(자기 폴더 보호): `TARGET_DIR` = `C:\Nexus`면 updater 폴더가 배포 대상 안 -> 빌드에 `updater/...` 경로가 있으면 `updater.json`·`state.json`·백업을 덮어씀. `copy`·`restore`가 updater 자기 폴더 안을 쓰려 하면 그 단계 실패
- 보완-2(폴더 권한): updater = SYSTEM이라 일반 사용자가 `updater.exe`를 바꿀 수 있으면 SYSTEM 권한으로 실행됨 -> `install.ps1`이 `C:\Nexus\updater`의 상속을 끊고 Administrators·SYSTEM만 쓰기

## 조율할 것 (2026-11-03 처음 질문)
| 이름 | 질문 | 제안 |
|---|---|---|
| 조-1(붙을 supervisor) | 집 서버 updater가 어느 supervisor에 접속하나 | 아래 "조-1 선택지" |
| 조-2(시험 순서) | VM `nexus-win` 먼저인가 | VM 먼저(단위-1~4를 집 서버 전에 확인) |
| 조-3(설치 자리·계정) | 설치 폴더·예약 작업 실행 계정 | `C:\nexus-updater` · SYSTEM. SYSTEM은 사용자 PATH·프로필이 없어서 단계 스크립트는 절대 경로를 써야 함 |

### 조-1 선택지
| 안 | 내용 | 장점 | 단점 |
|---|---|---|---|
| 가(집 서버 자신) | 집 서버 updater -> 집 서버 supervisor | 최종 모습 그대로. 외부 의존 없음 | 집 서버 supervisor에 updater 기능(마이그레이션 `00016`~`00018`)이 먼저 있어야 함 -> 첫 배포는 지금처럼 `update.ps1`로 수동. supervisor를 교체하는 동안 updater 접속이 끊김(P3 재전송으로 결과는 이어 받음) |
| 나(회사 개발 supervisor) | 집 서버 updater -> test2(8012) | 집 서버 supervisor와 무관해서 교체 중에도 접속 유지, 지금 바로 시험 가능 | 최종 모습이 아님. 집 서버 PC 관리가 회사 개발 서버에 묶임 |

## 구현 지도 (2026-11-03)
| 단위 | 파일 |
|---|---|
| 단위-1(스크립트 실행) | 새 `internal/scriptfmt`(worker `router/script.go`에서 옮김, `scriptfmt.Windows` 기본표) · `cmd/updater/exec_windows.go` `scriptCmd`(`.ps1` = powershell `-File`, `.bat`·`.cmd`·그 밖 = `cmd /c`, `.exe` = 그대로, 변환 사본 `.nexus-run-<id>-<이름>`을 같은 폴더에 만들고 끝나면 지움 — 원본은 `copy` 대상이라 그대로) · `outText` = 줄 단위로 UTF-8이 아니면 CP949로 읽음(PowerShell·옛 프로그램 파이프 출력) |
| 단위-2(자식까지 종료) | `exec_windows.go` `procGroup`: Job(`KILL_ON_JOB_CLOSE` 없음) + `CREATE_SUSPENDED` 시작 -> Job 등록 -> `NtResumeProcess`. 종료 = `TerminateJobObject`(SIGTERM 없음, force 무관). Linux는 `exec_linux.go` 같은 이름(프로세스 그룹) |
| 단위-3(잠긴 파일) | `fs_windows.go` `replaceFile`·`removeFile`(거절되면 `<이름>.old-<YYYYMMDD-HHMMSS>`로 물림) / `fs_other.go` = 이름 변경 그대로. `copyOne`이 물린 이름을 돌려주고 `copy`·`restore` 로그에 "실행 중이던 옛 파일 -> …" |
| 보완-1(자기 폴더 보호) | `builtin.go` `guardSelf`: updater가 쓰는 것만(설정 파일·`state.json`·`staging`·`backups`·`logs`·`updater.log*`·실행 파일) — `copy`는 전부 먼저 검사(하나라도 걸리면 아무것도 안 씀), `restore`도 같은 검사 |
| 단위-4(설치) | `scripts/updater/install.ps1`(UTF-8 BOM·CRLF): `-Dir` 기본 `C:\Nexus\updater` · icacls 상속 끊고 SID `S-1-5-32-544`·`S-1-5-18`만 · 예약 작업 `NexusUpdater`(SYSTEM, 부팅 때, `ExecutionTimeLimit` 0, 1분 간격 재시작) · `updater.exe -log updater.log` |
| 단위-4 로그 | `main.go` `-log`(상대 경로 = 설정 폴더, 켤 때 5MB 넘으면 `.1`로 넘김) — 예약 작업엔 journal이 없어서 |
| 단위-5(빌드) | `scripts/updater/build.sh` -> `build/updater/<태그>/{linux-arm64, windows-amd64}` + 설치 스크립트. 태그 `UPDATER-P4-261103` |

## 계획과 달라진 것 (구현 중)
| 차이 | 내용 · 이유 |
|---|---|
| 차이-1(보호 범위) | 보완-1을 "updater 폴더 전체"로 막았더니 설치 경로가 updater 폴더 **안**인 확인 스크립트(`work/target`)에서 모든 copy가 막힘 -> updater가 쓰는 파일만 막도록 좁힘 |
| 차이-2(Start-Process 자식) | 처음엔 "updater 작업을 재시작하면 단계가 `Start-Process`로 띄운 프로세스도 같이 꺼진다"고 봤으나 **VM에서 남음** 확인 -> `install.ps1` 주석 고침. 부팅 뒤에도 켜져야 하는 것은 여전히 자기 예약 작업으로 |
| 차이-3(출력 인코딩) | 계획엔 없던 것: 비대화형(파이프)에선 PowerShell이 CP949로 내보냄 -> 줄 단위 CP949 읽기 추가. interactive는 ConPTY라 UTF-8 그대로 |

## 확인 (2026-11-03, VM `nexus-win` + 격리 supervisor `172.17.0.1:5051`·임시 DB)
- scratchpad `updwin/upd_win.py` 33/33 + `upd_win2.py` 6/6(세션 끝나면 사라짐): 설치·접속(os=windows)·폴더 권한·로그 파일·예약 작업 설정 / `.ps1` 한글·`.bat` CP949·종료 코드 3·stderr / 정상 끝 단계의 자식은 남음·시간 제한은 자식까지(Job) / interactive `Read-Host` 입력 / 실행 중 `app.exe` 덮어쓰기·복원 때 물림·실행 중 새 exe 지우기 = 이름만 / `*.old-*` 남김 / 보완-1 / 다시 설치 = 설정 유지·같은 서브키 / **배포 중 supervisor를 내렸다 올려도 updater가 혼자 이어 가고 결과·끊긴 동안 로그를 다시 보냄**(조-1(가) 자기 교체의 전제)
- 로컬: `go test ./...` · `GOOS=windows go vet` · `scripts/check/run.sh` 193/193
- VM에 남김: `C:\NexusT\`(updater 설치·`*.old-*`), 작업 `NexusUpdater`는 멈춰 둠. VM은 `docker stop`
- 단계 스크립트 작성 규칙(Windows): `$ErrorActionPreference='Stop'`이 없으면 PowerShell 오류가 종료 코드 0으로 끝날 수 있음 / SYSTEM이라 사용자 PATH 없음 -> 절대 경로

## 단위-7(교체 빌드) 후속 (2026-11-03 오후, P4 머지 뒤)
- 무엇: Updater 화면에서 장비 루트 `home-server`에 빌드를 만들어 집 서버 supervisor·worker를 updater로 교체. 코드 변경 없음(빌드 = 보관소 `steps.json` + `.ps1`). 문제가 나오면 새 브랜치로 고침
- 흐름 초안: deploy = supervisor·주 worker 끄기 -> `pg_dump` -> `backup` -> `copy` -> `Start-ScheduledTask Nexus` -> 8013 응답 확인 / rollback = `restore` -> 켜기, DB 복원 = `manual: true`. 올릴 것 = 지금과 같은 코드, 태그만 바꾼 exe
- builtin `backup` = 파일만(`updater\backups\<배포 id>`, 최근 3회 보존). DB는 빌드 스크립트가 맡음
- **사용자 답(11-03 밤)**: updater 교체 = **나중**(이번 배포는 옛 P4 updater -> replace 무시, `web/`도 합치기) · 정-1(old 정리 단계) = **넣음**(배포 끝 `*.old-*` 삭제, 사용 중이면 넘어감) · 정-2(DB 백업 자리) = **`updater\db-backups\` + 최근 3개** · 정-3(관찰 방식) = 아직 답 없음
- 알아 둘 것: 단계 스크립트·`steps.json`도 빌드 파일이라 `copy`가 `TARGET_DIR`(`C:\Nexus`)에 같이 깖(`mergedFiles`가 빌드 파일 전부) -> 스크립트는 하위 폴더 하나에 모음
- **처음 질문(11-03)**:
  - 정-1(old 정리 단계): 배포 중 보관 프로세스가 `worker.exe`를 쥐고 있으면 `worker.exe.old-<시각>`이 생김 -> 배포 끝에 `*.old-*` 삭제 시도(사용 중이면 넘어감) 단계를 넣을지
  - 정-2(DB 백업 자리): `pg_dump` 결과를 `C:\Nexus`가 아니라 updater 쪽(예: `updater\db-backups\`)에 두고 최근 3개만 남길지
  - 정-3(관찰 방식): 배포 중 supervisor가 꺼져 Updater 화면·Nexus 터미널이 잠깐 끊김 -> 회사에서 이대로 지켜봐도 되는지 / 실패 되돌리기도 집 서버에서 한 번 볼지
