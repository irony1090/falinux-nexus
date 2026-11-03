# REF — 업데이트 app P3(배포)

> 2026-11-02 밤 신설. 설계 본문 → `REF-update-deploy.md`("자기 교체"·"대상의 의존성" 포함) · 규칙(잠금·되돌리기·개별 실행) → `REF-update-rules.md` · P2 구현 → `REF-update-p2.md` · 이력 → `history/update-p3.md` · 현재 진행 → `CURRENT.md`

**상태(2026-11-03 아침): 단위-1(DB)~단위-7(확인) 끝 — 로컬 `run.sh` 전체 통과 + `imx8mp-f24` 실기 통과(→ `history/update-p3.md`). 11-03 마감, 브랜치 `feat/updater-p3` 커밋(master 머지 여부 = 사용자). updater 빌드 태그는 `UPDATER-P2-261102` 그대로(사용자: 지금은 중요하지 않음).**

## 착수 계획 (2026-11-02 밤 합의)
| 단위 | 내용 |
|---|---|
| 단위-1(DB) | `00018`: `update_deployments`에 목록 종류(deploy/rollback) 열 · 잠그기 가능 판정 DB 함수 한 곳(현재 sha와 같은 성공 배포가 있는가) |
| 단위-2(배포 메시지) | `updproto`: 배포 시작(배포 id·버전·단계 목록·합친 파일 목록) · 단계별 시작·끝 보고 · 배포 끝(재접속 뒤 재전송) · 로그 전체 재전송(설계-10(자기 교체) 10-1(로그 끊김)) |
| 단위-3(updater 배포 실행) | updater가 `deploy` 목록을 차례로 실행(10-2(순서 주체), `manual` 건너뜀) · 진행 위치 `state.json` · 실패 = 멈춤, `autoRollback`이면 `rollback` 목록 · 끝까지 성공 = 설치본 이름(`Installed`) 기록 |
| 단위-4(supervisor API) | 배포 시작(인스턴스마다 배포 행, 동시 진행, 오프라인 거절) · 되돌리기 · 취소 · 잠그기 · 배포 목록·상세 |
| 단위-5(실시간 알림) | 소켓 토픽 `UPDATE:<메인키>` — 단계 상태·출력 push, 장비 타일 10초 조회 대체 |
| 단위-6(화면) | 빌드 타일 `배포` -> 배포 대상 **가운데 팝업** -> Run·Log 타일(단계 행 `실행` · 로그 xterm · interactive 입력 · 아래 줄 `되돌리기`·`잠그기`) + 잠그기 확인 창 |
| 단위-7(확인) | 로컬 자동 확인 -> `imx8mp-f24` 실기(일부러 실패 -> 되돌리기, 성공 -> 잠그기). Windows는 P3 뒤 |

## 합의한 해석 (2026-11-02 밤)
| 항목 | 결정 |
|---|---|
| 배포 대상 고르기 | 가운데 팝업(설계 문서 "배포 팝업"대로, 시안의 타일처럼 그린 것은 따르지 않음) |
| Run·Log 타일 범위 | **배포 하나 × 인스턴스 하나**(시안 `dep-0142 · a41c`). 두 대에 배포하면 타일 두 개. 장비 타일 빌드 행의 `Run · Log` = 그 빌드의 **마지막 배포** |
| 배포 중 updater 재시작 | 실행 중이던 단계 = "중단"으로 기록, 배포 멈춤. 재부팅 단계(`"reboot": true` 같은 표시로 이어 가기)는 `BACKLOG.md` |
| 용어 | 배포 = `deploy` 목록 전체 한 번 / 단계 실행 = 단계 하나(개별 실행, P2 방식) |

## 계획과 달라진 것 (구현 중)
| 차이 | 내용 · 이유 |
|---|---|
| 차이-1(되돌리기 기록) | 배포 기록에 "목록 종류" 열 대신 **같은 배포 안 `rollback_result`** — 시안 Run·Log가 한 타일에 두 목록을 보여서. 되돌리기 단계 기록 = 같은 `deployment_id`, `list = rollback` |
| 차이-2(진행 위치) | 계획 단위-3의 "진행 위치 `state.json` 저장"은 안 씀 — 재시작 = 중단(합의한 해석)이라. supervisor가 HELLO `Deploying`에 없는 배포를 중단으로 기록, 남은 `wait`는 `skip` |
| 차이-3(처리 순서) | transport에 `HandleOrdered` 추가 — REQ(단계 끝)가 고루틴이라 앞선 EVENT(시작 알림·출력)보다 먼저 처리되는 경합(화면 확인에서 발견: 인스턴스가 "실행 중"으로 남고 짧은 단계 로그가 비어 있음). 단계 끝·배포 끝·로그 재전송을 EVENT 디스패처 순서로 |
| 차이-4(로그 이음매) | `UPD:OUTPUT`에 누적 바이트 `off`, 실행 조회에 `logEnd` — 조회와 실시간 조각이 겹치거나 빠지지 않게(터미널 `writeLive`와 같은 방식) |
| 차이-5(자동 되돌리기 기록) | `autoRollback`이면 배포 시작 때 rollback `wait` 행도 미리 만들고, 되돌리기를 안 하면 배포 끝에서 지움 |

## 구현 지도 (2026-11-03)
| 단위 | 파일 |
|---|---|
| 단위-1(DB) | `00018_update_deploy.sql`(배포 열 4개 · 결과 `wait`/`skip` · 함수 `update_uses_sig`·`update_lockable`) · `query/updates.sql`(배포·wait 기록·잠그기·설치본) |
| 단위-2(배포 메시지) | `updproto`: `UPD_DEPLOY`·`UPD_STEP_START`·`UPD_DEPLOY_END`·`UPD_LOG` · HELLO `Deploying` · `Cancel.DeployID` · `DeployTag`(`d<id>`) |
| 단위-3(updater) | `cmd/updater/deploy.go`(목록 실행·자동 되돌리기·설치본·결과 재전송·`state.json` 잠금) · `agent.go`(출력 `out` + 재접속 로그 재전송, 잠금 순서 outMu -> mu) · `builtin.go`(manifest `prevInstalled`) |
| 단위-4(supervisor) | `updaterDeploy.go`(배포·되돌리기·취소·상세·잠그기 API, 시작 알림·배포 끝 처리) · `updaterConn.go`(`deploying`·`busy`·HELLO 정리·`HandleOrdered`) · `updaterRun.go`(wait 기록 허용·`replaceLog`·`emit`+off) · `update.go`(`lastDeployId`·`lockable`·라우트) · `updstore/steps.go`(`AutoRollback`) · `transport/conn.go`(`HandleOrdered`) |
| 단위-5(실시간) | `protocol/messages.go`(`UPD:WATCH`·`UPD:CHANGE`·`UPD:OUTPUT`) · `subscribe.go`(`onUpdWatch`) · 프론트 `hook/updWatch.hook.ts`(페이지 한 번, 바뀜 = 다시 조회, 출력 = runId별) · 장비 조회 10초 -> 60초 안전망 |
| 단위-6(화면) | `UpdDeployDialog.vue`(가운데 팝업) · `UpdRunTile.vue`(키 `r<id>`) · `UpdBuildTile`·`UpdDeviceTile`(Run · Log = 마지막 배포, 배포 버튼) · `UpdTileView`(경로) · `updGrid.util.ts` |
| 단위-7(확인) | `scripts/check/updater-deploy.py`(38) · `updater-deploy-ui.mjs`(18) → `run.sh` |

## 동작 규칙 (구현에서 정한 것)
| 항목 | 규칙 |
|---|---|
| 배포 | 인스턴스마다 배포 하나, 동시에. 오프라인·바쁨 = 그 인스턴스만 오류(응답 `items`). `manual` 단계는 기록도 안 만듦 |
| 단계 상태 | `wait` -> `running`(시작 알림 또는 HELLO Running) -> `ok`/`fail`/`timeout`/`cancel`, 멈춘 뒤 남은 것 `skip` |
| 바쁨 | 인스턴스당 단계 하나 또는 배포 하나(`running`·`deploying`). 배포 중 개별 실행·다른 배포·되돌리기·설치 경로 변경 = 409 |
| 취소 | 배포 취소 = 지금 단계 cancel + 다음으로 안 넘어감. 취소면 자동 되돌리기 안 함 |
| 되돌리기 | 배포가 끝난 뒤(성공이어도). 백업 = 그 배포 것(`d<id>`), 3회 보존 밖이면 restore 실패 + 이유("백업이 없습니다"). 빌드가 지워지면 409 |
| 설치본 이름 | 배포 끝까지 성공 = 그 버전 / 되돌리기 성공 = 백업 때 이름(`prevInstalled`) / 개별 실행은 안 바꿈 |
| 잠그기 | 판정 = DB 함수(빌드 sha + 공용 서명이 같은 성공 배포). 빌드 + 쓰는 공용 묶음 잠금 행. Run·Log 확인 창에 같이 잠기는 공용 묶음 |
| Run·Log | 단계 행 = 지금 `steps.json` 정의 + 그 배포의 마지막 기록(빌드가 없으면 기록만). 행 `실행` = 개별 실행(`d<id>` 문맥, 아래 "개별 실행"에 따로). 로그 = 고르지 않으면 진행 중·마지막 단계를 따라감, interactive 진행 중이면 터미널 입력 전달 |

## P3에 안 넣음 (→ `BACKLOG.md`)
- 백업 타일(시안 ③의 백업 목록) · `steps.json` 시각 편집 · 서브키 파일 계층 · 재부팅 단계 · updater 웹 자동 교체
