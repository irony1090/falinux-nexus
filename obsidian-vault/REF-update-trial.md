# REF-update-trial — 배포 전 단계별 시험 실행 (시험 타일)

→ history/update-trial.md · → CURRENT · REF-update-files-tile.md와 같은 브랜치 `feat/update-files-tile`(파트 업데이트 app 후속). 관련: `REF-update-rules.md`(개별 실행·잠금) · `REF-update-p3.md`(Run · Log 타일·개별 실행 합의) · `REF-update-p4.md`(Windows 단계 실행)

## 요구와 현재 상태 (2026-11-03)
- 요구: 배포하기 전에 steps.json 단계(스크립트)를 하나씩 장비에서 시험 실행하고 로그를 본다
- 이미 있는 것: 서버 `updStartRun`(`cmd/supervisor/router/updaterRun.go` 330행~, 배포 id 없이 가능, 잠긴 빌드도 허용) · 기록 `update_step_runs.deployment_id NULL` · updater 개별 실행
- 막힌 곳: 화면 입구가 Run · Log 단계 행 `실행`뿐인데 Run · Log(`r<배포 id>`)는 배포가 있어야 열림. 배포하면 steps.json 전체가 도니 "배포 전 시험"이 안 됨
- 파일 행에 시험 버튼을 두는 안은 제외: 개별 실행은 파일 경로가 아니라 steps.json 단계 id(`stepId`)로 돌아 steps.json에 없는 스크립트는 못 돌림. 같이 보낼 파일 걱정은 없음(서버가 공용 -> 메인키 -> 서브키 합친 목록 전부를 보내고 updater가 staging에 받아 거기서 실행, 배포와 같음)

## 결정 (11-03 사용자 확정)
| 결정 | 내용 |
|---|---|
| 결정-1(별도 타일) | 시험 = 새 타일(키 `x<메인키>~<버전>`, 종류 `trial`). Run · Log는 배포 기록 전용으로 그대로 |
| 결정-2(구분) | 시험 기록 = `deployment_id NULL`, 배포 기록 = 값 있음. 잠금 판정은 배포 기록만 봄 -> 시험이 잠그기 조건을 못 채움. Run · Log의 "개별 실행" 절 이름은 "시험·단독 실행"으로 |
| 결정-3(범위) | script 단계만 실행. `copy`·`backup`·`restore` 행은 흐리게 "배포에서만" |
| 결정-4(잠금) | 잠긴 빌드에서도 시험 허용(서버가 이미 허용) |
| 결정-5(기록) | 남김(끝난 행의 결과·`log_tail` 4KB + 전체 `logs/<id>.log`) |
| 결정-6(인스턴스) | 타일 위에서 고름, 하나뿐이면 자동 선택, 오프라인·바쁨(배포 중·단계 실행 중)은 선택 불가 |
| 결정-7(입구) | 빌드 타일 열기 줄에 `시험` 버튼을 새로 둠. `Run · Log`는 그대로(한 버튼이 두 뜻이 되면 섞임) — 질문-11 답 |
| 결정-8(코드 공용화) | `UpdRunTile`의 로그 상자(xterm·조회와 실시간 이음매·선택 따라가기)를 훅으로 빼서 시험 타일과 같이 씀 — 질문-12 답 |
| 결정-9(순서) | 시험 실행을 단위-나(끌어 놓기)보다 먼저 |

## 구현 계획
| 단계 | 내용 | 파일 |
|---|---|---|
| 시험-1(서버 조회) **끝(11-03)** | `GET /admin/updates/bundles/build/:owner/:version/trial?sub=` = deploy·rollback 단계 정의 + `stepsIssue` + 그 인스턴스·버전의 단독 실행 기록(최근 50). 실행은 기존 `updStartRun` 그대로 | `router/updaterTrial.go` + `update.go` 라우트 + sqlc `ListTrialRuns` |
| 시험-2(로그 상자 공용화) **끝(11-03)** | 로그 상자를 `hook/updLog.hook.ts`로 + 공용-1(행 표시)·공용-2(스타일) | `UpdRunTile.vue` 리팩터(동작 그대로) |
| 시험-3(시험 타일) **끝(11-03, 사용자 화면 확인 끝)** | 위 = 인스턴스 고르기(`Ct*` 컴포넌트) / deploy·rollback 전환 / 단계 행 `실행`(script만) / 로그 / 시험 기록 / 푸터 취소 | 새 `UpdTrialTile.vue` · `updGrid.util.ts`(KIND·키·`updTargetOf`·크기 2칸) · `UpdTileView.vue`(71행 근처) |
| 시험-4(입구·구분 표시) **끝(11-03, 절 이름만 — 입구·안내는 시험-3에서)** | 빌드 타일 열기 줄 `시험` 버튼 + 시험 타일 안내 "배포가 아닙니다 · 잠금 판정에 안 들어갑니다" + Run · Log "개별 실행" 절 이름 | `UpdBuildTile.vue`(98행 근처) · `UpdRunTile.vue` |
| 시험-5(확인) **끝(11-03)** | `scripts/check`에 시험 타일 검사(배포 없이 열림·자동 선택·script 실행·로그·내장 행 꺼짐·잠겨도 허용·잠그기 조건 안 채움) + `updater-ui.mjs` 66행 열기 줄 기대값(`Files\|steps.json\|Run · Log`)에 `시험` 반영 + VM `test-win`에서 `00-stop.ps1` 실제 실행 | `scripts/check/*` |

```
[빌드 타일] --시험--> [시험 타일] --실행--> 서버 updStartRun --> updater
                          |                        |
                          +-- 시험 기록 조회 <-- update_step_runs (deployment_id NULL)
```

## 시험-1(서버 조회) 구현 (11-03, 사용자 질문-1~3 제안대로)
| 항목 | 내용 |
|---|---|
| 질문-1(기록 조회 쿼리) | 새 sqlc `ListTrialRuns`(`deployment_id IS NULL AND version = $3`). 기존 `ListStepRuns` + Go 거르기는 배포 단계 기록이 LIMIT 안을 채워 시험 기록이 밀려남 |
| 질문-2(steps.json 오류 표시) | 응답 `stepsIssue` = `""`(정상) · `steps.json이 없습니다` · `읽을 수 없습니다` · `오류가 있습니다`. 단계 목록은 정상일 때만 채움 |
| 질문-3(기록 쌓임) | 정리 코드 없음 확인(DB 행·`logs/<id>.log` 계속 쌓임) -> 이번엔 안 고침, `BACKLOG.md` "시험 기록 정리" |
| 응답 | `{deploy, rollback: updStepDef[], stepsIssue, runs: updRunItem[]}`. `sub` 없으면 `runs = []`, 잘못된 sub 400, common 종류 400, 없는 빌드 404 |
| 확인 | 개발 8012에서 curl: imx8mp-f24/v2(시험 기록 15 · 배포 기록 제외) · test-win/INIT(`stop` 1건) · sample-board(steps.json 없음) · 오류 경우 400/404/401. "steps.json 오류" 경우는 데이터가 없어 못 봄(코드 경로는 `updGetDeployment`와 같음) |

## 미확인
- 시험 타일의 인스턴스 목록 출처: `useUpdDevice(mainKey)`(online·busy 포함) 재사용 예정, 시험-3에서 확인
## 시험-2(로그 상자 공용화) 구현 (11-03)
| 항목 | 내용 |
|---|---|
| `hook/updLog.hook.ts` | `useUpdLog({ runs, interactive })` -> `box`(template ref)·`selected`·`shown`·`pick(r)`·`follow(runId)`(방금 시작한 실행 고정)·`unpin()`(다시 따라가기)·`size()`. xterm·ResizeObserver·입력·조회/실시간 이음매·`updLogReload` 다시 읽기 포함. `useUpdNow()`(1초 시계)도 여기 |
| 공용-1(행 표시) | `util/updRun.util.ts`: `RUN_ICON`·`RUN_TEXT`·`runTone`·`runText(now, r, def)` |
| 공용-2(스타일) | `.log`·`.ic`(scoped) -> `styles/updater.scss` `.upd-log`·`.upd-ic`. `updater-deploy-ui.mjs` 선택자 3곳 같이 고침 |
| 타일에 남는 것 | 단계 정의 찾기(`defOf`·`shownDef`)는 타일마다 다름 -> 훅엔 `interactive` 함수만 넘김 |
| 확인 | `vue-tsc` 통과 · `scripts/check/run.sh` 전체 통과(Run · Log 프롬프트·입력·따라가기 포함). eslint는 프로젝트에 미설치라 못 돌림 |

## 시험-3(시험 타일) 구현 (11-03, 맞춤-1~3 사용자 확정)
| 항목 | 내용 |
|---|---|
| 맞춤-1(자동 선택) | 고를 수 있는(온라인·대기) 인스턴스가 하나뿐이면 자동. 고른 값은 타일 안에만(새로고침하면 다시 자동 선택) |
| 맞춤-2(입구 버튼 시기) | 빌드 타일 열기 줄 `Files \| steps.json \| 시험 \| Run · Log`를 시험-3에 넣음 -> 시험-4엔 Run · Log 절 이름만 남음. `updater-ui.mjs` 66행 기대값도 고침 |
| 맞춤-3(안내 문구) | 기본 "배포가 아닙니다 · 잠금 판정에 안 들어갑니다", 막힌 이유(요청 오류·steps.json 문제·updater 없음·인스턴스 안 고름·오프라인·배포 중·다른 단계 실행 중)가 있으면 그것 |
| 파일 | 새 `UpdTrialTile.vue` · `updater.api.ts`(`Trial`·`getTrial`·`useTrial`, 조회 키 `UPDATES/TRIAL` = `UPD:CHANGE`로 다시 조회, 인스턴스 바뀔 때 옛 기록이 남지 않게 `keepPreviousData` 안 씀) · `updGrid.util.ts`(키 `x`, 종류 `trial`, 처음 높이 2) · `UpdTileView.vue`(경로 `메인키 > 버전 > 시험`, 아이콘 `mdi-flask-outline`) · `UpdBuildTile.vue` · 훅에 `reset()`(인스턴스 바꿀 때 로그 비움) · `.upd-lists`도 updater.scss로(Run · Log와 공용) |
| 확인 | vue-tsc · run.sh 통과 · 개발 8011 headless: 자동 선택 · 내장 행 흐림 · script 행만 `실행` · VM `test-win`에서 `stop`(00-stop.ps1) 실제 실행 -> 로그(staging 경로·TARGET_DIR) · 기록 행 추가 · 안내 문구 · rollback 전환 · 브라우저 오류 없음 |

## 시험-5(확인) 구현 (11-03, 선택-1~3 사용자 확정)
- 새 `scripts/check/updater-trial-ui.mjs`(run.sh에서 `updater-deploy-ui.mjs` 다음) — 격리 supervisor + Linux updater 2개(폴더마다 state.json). 24개 항목 처음부터 통과
- 검사-1(조회 API: sub 없음·steps.json 없음) · 검사-2(입구: 열기 줄·배포 전 Run · Log 꺼짐·자식 2칸) · 검사-3(자동 선택·기본 안내) · 검사-4(내장 행 실행 없음) · 검사-5(실행 -> 로그·완료·기록) · 검사-6(배포 뒤에도 시험 기록만) · 검사-7(시험 성공해도 lockable 아님) · 검사-8(잠긴 빌드 시험) · 검사-9(브라우저 오류 없음)
- 선택-1(두 인스턴스): 둘 다 온라인 = 자동 선택 안 함·실행 꺼짐, B 끈 뒤 = B 항목 고를 수 없음 / 선택-2(잠긴 빌드): `update_locks` 행을 SQL로 직접(정식 잠그기는 `updater-deploy-ui.mjs`) / 선택-3(VM): 시험-3 때 `test-win` 실제 실행으로 끝, 자동 검사엔 안 넣음
- 못 넣은 것: "steps.json에 오류가 있습니다" 경우 — 텍스트 저장 API가 오류 있는 steps.json을 422로 거절해서 만들 길이 없음(코드 경로는 Run · Log 조회와 같음)

## 시험 환경
- VM `nexus-win` + updater `test-win#BGRB5DdV`(test2, `C:\NexusU`, 11-03 저녁 실행) → `REF-windows-vm.md`. Windows 단계 스크립트 규칙 = `REF-update-p4.md`(`$ErrorActionPreference='Stop'`, SYSTEM 아님 주의: 이 VM updater는 일반 사용자 계정으로 띄움)
- 단계 스크립트 환경변수·작업 폴더: 작업 폴더 = staging(`cmd.Dir`), `TARGET_DIR`·`VERSION`·`STEP_ID`·`STAGE_DIR`·`DEPLOY_ID`·(백업 있으면)`BACKUP_DIR`·`NEW_FILES`
