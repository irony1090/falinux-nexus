# REF — 업데이트 app P2(updater Linux)

> 2026-11-02 밤 신설. 설계 본문 → `REF-update-deploy.md` · 규칙 → `REF-update-rules.md` · P1 구현 → `REF-update-p1.md` · 이력 → `history/update-p2.md` · 현재 진행 → `CURRENT.md`

**상태(2026-11-02 밤): 단위-1(DB)~단위-8(확인) 끝 — 로컬(`run.sh` 전체 통과) + `imx8mp-f24` 실기 확인 통과. 남은 것 = 마감 커밋. 브랜치 `feat/updater-p2`.**

## P2 전 확정 (2026-11-02 밤)
| 항목 | 결정 |
|---|---|
| 정-1(접속 인증) | **메인키별 API 키**(안 B, 다시 보기 방식 사용자 확인). 장비 루트 타일에서 발급·다시 보기·재발급. worker 장비 키와 같은 방식(접속 검증 = 해시, 다시 보기 = `KEY_SECRET` AES-GCM 봉인), worker 키와는 별개. **첫 접속 때 장비 루트 자동 생성은 폐기** — 관리자가 장비 루트를 먼저 만들고 키를 updater 설정에 넣는다 |
| 정-2(실행 권한) | 파일마다 권한(mode)을 DB에 둔다. 기본 `.sh` = 755, 나머지 644. Files 타일에서 실행 켜기/끄기 + 8진수 입력 둘 다. updater가 `copy` 때 적용(Windows는 무시). 폴더 sha에 mode 포함(권한만 바꿔도 배포 내용이 바뀐 것) |
| 정-3(`state.json` 형식) | JSON |
| 정-4(첫 시험 장비) | `imx8mp-f24`(`falinux@`, sudo, 지금 test3에 붙음). updater = root systemd 서비스 |
| 정-5(설치 경로) | updater 설정 파일에 `TARGET_DIR`. 웹에서 바꾸면 updater가 설정 파일을 고쳐 씀(가). 배포 중엔 거절, 그 밖엔 바로 적용. `manifest.json`에 그때 `TARGET_DIR`을 남겨 되돌리기는 옛 경로로 |
| 정-6(콘솔) | **대화형 지원**. 단계 스크립트는 PTY에서 실행, 웹 xterm 입력을 중계. 단계에 `"interactive": true`면 그 단계는 timeout 끔(입력 대기 감지는 불가능해서). 비대화형 단계는 stdin 닫음. 입력 창 화면은 P3(Run·Log), P2는 updater PTY + 입력 중계까지 |

## 착수 계획 (2026-11-02 밤 확정)
| 단위 | 내용 |
|---|---|
| 단위-1(DB) | `00017`: 장비 루트 키(해시·봉인) · `update_files.mode` · `steps.json` `interactive` 검증 |
| 단위-2(키·권한 화면) | 장비 루트 타일 키 발급·다시 보기·재발급 / Files 타일 권한 열(켜기·끄기 + 8진수) |
| 단위-3(접속) | supervisor updater 전용 WS(`/api/updater/ws`, 키 인증) · 서브키 발급 · `updaters` 행 갱신 · online = 메모리. transport 패키지 재사용, 메시지 정의는 updater 전용 |
| 단위-4(updater 본체) | `cmd/updater`: 설정 파일(주소·메인키·키·`TARGET_DIR`) · `state.json` · 재접속 · systemd 유닛 |
| 단위-5(받기·builtin) | 합친 파일을 staging에 받기(sha 확인) · `backup`/`copy`/`restore` + 권한 적용. `manifest.json`에 그때 `TARGET_DIR` 기록(경로가 바뀌어도 복원은 옛 경로) |
| 단위-6(script 실행) | PTY 실행 · 환경변수 · timeout · 프로세스 그룹 종료 · 출력·입력 중계 · 로그(DB 요약 + 파일) |
| 단위-7(인스턴스 화면) | 장비 타일 인스턴스 칸(online·별칭·메모·OS/CPU·`TARGET_DIR` 변경 → updater에 전달, 배포 중이면 거절) |
| 단위-8(확인) | 로컬 updater + 임시 DB 5051 자동 확인 → `imx8mp-f24` 실기 |
- P2에서 단계 실행 확인 = 확인용 API·스크립트만(임시 버튼 없음, Run·Log 화면은 P3)

## 구현 지도 (2026-11-02 밤)
| 단위 | 파일 |
|---|---|
| 단위-1(DB) | `00017_updater_conn.sql`(`update_devices` 키 3열 · `update_files.mode`, 기존 `.sh` = 755) · `query/updates.sql`(키·updaters·단계 기록 쿼리) |
| 보관소·검증 | `updstore/store.go`(`DefaultMode` · `FolderSha`에 mode — **기본 권한이면 P1 형식 그대로**라 P2 전 잠금 sha가 안 바뀜) · `steps.go`(`interactive` · Step에 Timeout·Manual·Interactive, interactive+timeout = 경고) |
| 메시지 | `internal/updproto`(worker `protocol`과 분리, 봉투·전송은 재사용) |
| supervisor | `updaterKey.go`(키 `nxu_`, worker `nxk_`와 접두어로 구분, 해시·봉인 함수는 같음) · `updaterConn.go`(연결 목록·WS·HELLO·인스턴스 고치기/지우기·파일 받기) · `updaterRun.go`(단계 실행·로그·입력·취소, 합친 파일 목록) · `update.go`(장비 만들 때 키 발급·`nokey` 필터·장비 지우면 연결 끊음) · `updateFiles.go`(`PATCH …/mode`) |
| updater | `cmd/updater/`: `main.go` · `config.go`(updater.json·state.json·원자적 쓰기) · `agent.go`(접속·재접속·결과 재전송) · `job.go`(staging·script) · `builtin.go` · `exec*.go`(PTY·파이프, Linux 프로세스 그룹) · `env`(태그 `UPDATER-P2-261102`) |
| 설치 | `scripts/updater/nexus-updater.service`(`KillMode=process`) · `install.sh`(`/opt/nexus-updater`, 기존 updater.json은 안 덮음) |
| 화면 | 장비 타일: 접속 키(보기·복사·재발급, 키 없으면 `키 발급`) + updater.json 상자 · 인스턴스 행(설정 = 별칭·메모·설치 경로 / offline만 지우기, 10초마다 다시 조회) / Files·공용 묶음 타일: 실행 켜기·끄기 + 8진수 버튼(`useUpdMode`) / 목록 필터 `키 없음` |
| 확인 | `scripts/check/updater-run.py`(34) + `updates-api.py` 권한 5개 추가 → `run.sh`에 포함 |

## 동작 규칙 (구현에서 정한 것)
| 항목 | 규칙 |
|---|---|
| 경로 | WS `/updater/ws` · 파일 `/updater/file?b=&scope=&sub=&path=`(머리 `X-Nexus-Main`·`X-Nexus-Key`) — `/api` 밖(worker 경로와 같은 이유). 빌드는 자기 메인키 것만, 공용은 모두 |
| HELLO | 키 검사 → 서브키 발급(빈 값일 때) → 같은 서브키가 이미 붙어 있으면 거절 → `updaters` upsert(지운 인스턴스가 다시 붙으면 새 행) → updater가 모르는 `running` 기록 = cancel |
| 실행 | 인스턴스당 하나. supervisor가 steps.json에서 단계를 찾아 합친 파일 목록(공용 uses 순서 → main → 그 서브키) + stage sha를 보냄. 기록은 요청 트랜잭션 밖에서 바로 커밋(출력·끝이 커밋 전에 올 수 있어서) |
| staging | `staging/<버전>-<stage 16자>/`, sha 같은 파일은 다시 안 받음, 받은 뒤 sha·크기 확인 + 권한, 최근 3개 |
| script | 작업 폴더 = staging. 환경변수 `TARGET_DIR`·`VERSION`·`STEP_ID`·`STAGE_DIR`·`DEPLOY_ID`(+백업 있으면 `BACKUP_DIR`·`NEW_FILES`). 비대화형 = stdin 없음·기본 10분·LF→CRLF / interactive = PTY·시간 제한 없음 / 종료 = 그룹 SIGTERM → 5초 뒤 SIGKILL |
| builtin | backup id = 배포 id, 개별 실행이면 `r<기록 id>` / restore·`BACKUP_DIR` = 배포 id의 백업, 개별 실행이면 가장 최근 / copy = 같은 폴더 임시 이름 → 권한 → 이름 변경(Linux는 실행 중 파일도 됨) |
| 끊김 | 출력은 updater `logs/<id>.log`에만 남고, 결과(End)는 받았다는 답이 올 때까지 재접속 뒤 다시 보냄. supervisor 재시작이면 HELLO Running으로 기록을 다시 엶 |
| 로그 | supervisor `<UPDATE_STORE_DIR>/logs/<id>.log` + 메모리 끝 64KB(끝난 뒤 10분), DB `log_tail` 4KB |
| 설치 경로 변경 | 접속 중 + 실행 중 아님 → updater가 updater.json의 targetDir만 고쳐 씀(다른 항목 보존, 0600). 상대 경로 거절 |

## 남은 것 / 알려진 한계
- 실기 시험 장비(켜 둠, P3(배포) 시험에 다시 씀): `imx8mp-f24`(192.168.80.119) 서비스 `nexus-updater` = `/opt/nexus-updater`(태그 `UPDATER-P2-261102`) -> **개발 supervisor test2**, 서브키 `pVMx1WwE`(별칭 `f24 실기`), 배포 대상 = `/opt/nexus-updater-test`(지금 복원 상태 `app.conf` ver=0). 개발 DB 장비 루트 `imx8mp-f24` 빌드 `v1`(hello·ask·check 스크립트). test3로 옮기려면 test3 빌드본에 P2가 들어간 뒤 updater.json 주소·키를 바꾼다(키는 DB마다 다름)
- 화면 실시간 알림 없음(장비 타일 10초 조회) — Run·Log와 같이 P3
- 끊긴 동안 출력은 supervisor 로그에 빠짐(updater 로그엔 있음)
- PTY 끝 출력: `pty` 패키지가 Wait 뒤 출력 큐를 닫아 마지막 몇 바이트가 빠질 수 있음(worker와 같은 옛 동작)
