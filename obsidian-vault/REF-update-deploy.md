# REF — 업데이트·배포 체계 (파트(업데이트 app) 설계)

> 2026-10-30 신설. 이력 → `history/update-deploy.md` / 지금 배포 방식 → `REF-home-deploy.md`(update.ps1)·`REF-build.md` "보드 worker 교체" / 화면 시안 → `user/배포-화면.html` · 설계안 → `user/배포-설계.html` / 현재 진행 → `CURRENT.md`

**상태(2026-10-30): 설계만 끝, 코드 착수 전. 화면 시안 사용자 확인 필요.**

## 왜 (지금 손으로 하는 이유)
| 막히는 곳 | 내용 |
|---|---|
| supervisor 자가 교체 | 실행 중 exe는 덮어쓸 수 없고 자기를 재시작할 수도 없음 → 지금은 `Start-ScheduledTask`로 우회 |
| worker는 Job | 터미널에서 띄운 프로세스는 터미널이 끝나면 같이 죽음(JS-1) |
| 버전이 안 보임 | 어느 장비가 어느 빌드인지 DB·화면에 없음. 태그는 `cmd/*/env` 첫 줄 주석뿐 |
| 장비가 늘어남 | 집 서버 + 보드 2대 + 회사 PC 여러 대를 한 번에 관리할 수단이 없음 |

## 확정 (사용자, 2026-10-30)
| 항목 | 결정 |
|---|---|
| 교체 주체 | **장비마다 `nexus-update` 상주**(Job 밖, worker와 비슷한 에이전트). supervisor·worker 둘 다 담당 |
| 실행 방식 | **관리자 웹 화면(배포 페이지)** 에서 버튼 |
| 배포 단위 | 장비별 + **전체 일괄** |
| 버전 관리 | **릴리스 이력 + 되돌리기** |
| 순서 | **Linux 먼저**(보드에서 시험) → Windows |
| 릴리스 형태 | 폴더 트리. `common`(모든 장비 공통) + `<os>-<arch>`(자동 매칭) |
| 폴더 추가 | **가안** — 이름 직접 입력(`+ 폴더`). `os-arch` 형식이면 자동 배포, 아니면 보관만 |
| 빈 폴더 | 판정은 **파일 유무**. 빈 폴더는 흐리게, 배포 대상 아님(기본 칸을 만들어 둬도 안 쓰면 그만) |
| 단계 | **스크립트 1개 = 한 단계** + `steps.json`(순서·`required`·`manual`·`timeout`). 각 단계 개별 실행 |
| 되돌리기 | 제안 = **역순 실행**(스크립트가 스스로 판단). 단계마다 rollback 짝지정은 기각 쪽 |

## 흐름
| 단계 | 내용 |
|---|---|
| ① 준비 | 배포 페이지에서 버전 폴더를 만들고 OS별 폴더에 파일·단계 스크립트 채움 |
| ② 등록 | supervisor가 보관(파일 + DB 기록: 버전·sha·올린 사람·시각) |
| ③ 배포 | 장비 선택(또는 전체) → supervisor가 대상 updater로 전송 |
| ④ 실행 | updater가 설치 위치에 펼치고 `steps.json` 순서대로 실행, 로그를 실시간 중계 |
| ⑤ 보고 | 장비별 현재 버전·단계 상태·결과. 실패 = 되돌리기 제안 |

## 릴리스 폴더
```
261030_drag/linux-arm64/
+-- worker
+-- steps.json
+-- 00-stop.sh
+-- 01-backup.sh
+-- 02-copy.sh
+-- 03-start.sh
+-- 04-verify.sh
+-- 90-cleanup.sh      (manual: 배포엔 안 끼고 손으로만 실행)
```

```json
[{ "id": "stop", "name": "끄기", "script": "00-stop.sh", "required": true },
 { "id": "verify", "name": "확인", "script": "04-verify.sh", "timeout": "15m" },
 { "id": "cleanup", "name": "정리", "script": "90-cleanup.sh", "manual": true }]
```
- 단계에 넘기는 환경변수 = `TARGET_DIR`(설치 경로) · `VERSION` · `STAGE`(deploy/rollback) · `STEP_ID`
- `steps.json`이 없으면 옛 방식(`update.sh`/`update.ps1` 한 번)으로 떨어뜨리는 게 좋음(기존 묶음 호환)

## 기록할 것 (DB 초안)
| 테이블 | 컬럼 | 쓰는 곳 |
|---|---|---|
| `releases` | 버전 · 폴더(common/os-arch) · sha256 · 크기 · 올린 사람 · 시각 · 보관 경로 | 릴리스 목록·되돌리기 |
| `deployments` | 장비(메인키) · 릴리스 · 시작/끝 · 결과 · 단계별 상태·로그 요약 | 장비별 현재 버전·이력 |
| `updaters` | 메인키 · OS/arch · 설치 경로 · 마지막 접속 | 배포 가능 장비 목록·online 판정 |

## 화면 (시안 = `user/배포-화면.html`)
| 탭 | 내용 |
|---|---|
| ① 기본 | 왼쪽 목록(릴리스/장비 탭) + 오른쪽 타일(릴리스·장비). 타일 틀 = `MemberPage`와 같은 머리 줄 36·몸통·아래 줄 40 |
| ② 새 릴리스 | 폴더 칩(`common`·`windows-amd64`·`linux-arm64`·`+ 폴더`). 빈 폴더 = 점선·흐림, 규칙 밖 이름 = 경고 |
| ③ 배포 진행 | 단계 목록(이름·스크립트·상태·소요) + 고른 단계의 로그 + 수동 단계 `실행` 버튼 |

## 단계 계획
| 단계 | 내용 | 확인 |
|---|---|---|
| P1(보관·화면) | 배포 페이지, 릴리스 업로드, DB 기록 | 올린 폴더가 그대로 보관되는가 |
| P2(updater Linux) | systemd 상주, 접속·파일 수신·steps.json 실행·로그 | 보드 한 대 worker 교체 |
| P3(배포) | 장비별·전체, 진행 표시, 이력, 되돌리기 | 일부러 실패시켜 되돌림 확인 |
| P4(Windows) | 스케줄 작업, `update.ps1` 규약, supervisor 교체 | 집 서버에서 supervisor·worker 교체 |

## 미정 (착수 때 확정)
- **콘솔 인터랙션**: 가(대화형 + 입력 대기 중 되돌림 시계 멈춤) / 나(보기 전용 + "대화형 실행" 버튼) — 자동 되돌림과 상충
- **스크립트 위치**: OS 폴더 안(현재 시안) / `common`에도 허용(`update.sh`·`update.ps1` 둘 다)
- **첫 시험 장비**: `fa24-PID`(root, `/home/root/nexus-worker`) / `imx8mp-f24`(falinux+sudo, `/opt/nexus-worker`)
- **updater 인증**: worker API 키와 별도 토큰(worker 교체 중에도 살아 있어야 함)
- **설치 경로**: updater `env`의 `TARGET_DIR`

## 함정 (지난 배포에서 실제로 물린 것)
- 실행 중 exe → `*.old-<시각>`으로 물러나게 한 뒤 복사(Windows·Linux 공통)
- `-hold`가 옛 exe를 쥠 → 남은 보관 프로세스를 세어 결과에 표시
- Job(JS-1) → updater는 서비스/스케줄 작업으로 떠 있어야 함
- 자동 되돌림 vs 입력 대기 → 대화형이면 시계를 멈춤
- 서버·DB 버전 짝 → 업데이트 스크립트가 배포 전 `pg_dump`를 남김(옛 타일 순서 변환 때 겪음)

## 참고
- 아티팩트 게시 불가(이 세션 = `ANTHROPIC_API_KEY` 인증) → 시안 HTML을 `user/`에 둠. claude.ai 로그인하면 아티팩트로 공유 가능
