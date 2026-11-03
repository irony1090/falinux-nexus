# history/update-trial — 배포 전 단계별 시험 실행

→ REF-update-trial.md · → history/update-files-tile.md · → CURRENT

## 2026-11-03 밤 (조율·계획, 코드 전)
- **발단**: 사용자가 새 빌드에 `00-stop.ps1`을 올리고 Run · Log가 비활성인 걸 봄 -> "배포 전에 각 스크립트별 테스트 실행도 해봐야 돼"
- **조사**: 개별 실행 경로는 서버·updater에 이미 있고(`deployment_id NULL`), 화면 입구가 Run · Log 단계 행 `실행`뿐이라 배포 전에는 못 씀
- **조율**: 질문-1~5(입구·대상·범위·잠금·기록) -> 사용자: 입구 = 파일 행 버튼 or Run · Log를 배포 없이 열기 -> 파일 행은 steps.json 단계 id 기준이라 제외 / 같이 보낼 파일은 걱정 없음(합친 목록 전부 전달). 구분 걱정(시험 vs 배포 로그) -> `deployment_id` 로 구분됨·잠금 판정 제외. 질문-6~10 긍정, 질문-11(`시험` 버튼 따로)·질문-12(로그 상자 공용화) ok
- 결과: 결정-1~9 + 시험-1~5 계획(`REF-update-trial.md`). 다음 세션 = 시험-1(서버 조회)부터

## 2026-11-03 (새 세션) — 시험-1(서버 조회)
- 코드 대조 중 질문-1(기록 조회 쿼리 — 기존 쿼리는 배포 기록에 밀림)·질문-2(steps.json 오류 표시 `stepsIssue`)·질문-3(기록 정리 코드 없음 -> BACKLOG) 제기 -> 사용자 "제안대로"
- 사용자가 "빌드 전 Run/Log 활성화" 이야기를 확인 -> vault 기록(결정-1(별도 타일)·결정-7(입구))대로 맞음
- 구현: `updaterTrial.go`(`updGetTrial`·`trialSteps`) · `update.go` `b.GET("/trial")` · `updates.sql` `ListTrialRuns` + sqlc generate. go build·vet 통과
- 확인: 8012 개발 supervisor 다시 띄움(새 라우트), curl 경우별 결과 -> `REF-update-trial.md` "시험-1(서버 조회) 구현"

## 2026-11-03 (새 세션) — 시험-2(로그 상자 공용화)
- 계획 먼저 보임: 훅 경계 + 공용-1(행 표시 util로)·공용-2(스타일 updater.scss로) 제안 -> 사용자 "제안대로"
- 구현: 새 `hook/updLog.hook.ts`·`util/updRun.util.ts`, `UpdRunTile.vue`는 훅·util 사용으로 줄임, `styles/updater.scss`에 `.upd-ic`·`.upd-log`, `updater-deploy-ui.mjs` 선택자 `.upd-log`
- 확인: vue-tsc · `scripts/check/run.sh` 전체 통과 -> `REF-update-trial.md` "시험-2(로그 상자 공용화) 구현"

## 2026-11-03 (새 세션) — 시험-3(시험 타일)
- 계획 먼저 보임(화면 그림 + 맞춤-1(자동 선택)·맞춤-2(입구 버튼 시기)·맞춤-3(안내 문구)) -> 사용자 "그대로"
- 구현 중 사용자 배치 버그 신고로 화면 확인 전 멈춤 -> 배치 고친 뒤 이어 감(`history/member-ui.md` 11-03 배치)
- 화면 확인에서 label 없는 인스턴스가 `서브키 (서브키)`로 두 번 나와 고침. VM에서 `stop` 실제 실행 확인 -> `REF-update-trial.md` "시험-3(시험 타일) 구현"

## 2026-11-03 — 시험-4(입구·구분 표시)
- `UpdRunTile.vue` 절 제목 `개별 실행 N` -> `시험·단독 실행 N`(시험 타일 실행도 같은 조건 `deployment_id NULL`·같은 버전·배포 뒤라 여기 모임). 확인 스크립트는 이 글자를 안 봄

## 2026-11-03 — 시험-5(확인)
- 계획(검사-1~9 + 선택-1(두 인스턴스)·선택-2(잠긴 빌드 SQL)·선택-3(VM은 이미 끝)) -> 사용자 "제안대로"
- `updater-trial-ui.mjs` 추가 + run.sh 한 줄. run.sh 전체 통과 -> 시험 실행(시험-1~5) 끝 -> `REF-update-trial.md` "시험-5(확인) 구현"
- (11-03 저녁) 사용자가 test1에서 시험 타일을 직접 써 봄 -> "문제없어"(기록 17:05·17:06이 그때 것)
