# history — CURRENT에서 내려보낸 줄 원문 (6, 2026-11-02 저녁)

> 업데이트 app 파트(10-30 설계 상태, 11-02 오후 재검토 중 줄)의 옛 상태. 현재 상태는 `CURRENT.md`, 결정은 `REF-update-redesign.md`. 앞 = `current-archive-5.md`

- **(2026-11-02 오후) 파트(업데이트 app) 설계 재검토 진행 중** — 설계-1(지우고 깔기)~설계-7(trap 신호)·설계-9(개별 실행과 설치본) 합의 끝, 릴리스 구조 개정(장비 루트 > 빌드, `common` = `uses` 의존)·`steps.json` `deploy`/`rollback` + `builtin`·웹 편집(P1 = 원문 편집 먼저). **결정 기록 = `REF-update-redesign.md`**. 남은 것 = 설계-8(supervisor 되돌리기)·설계-10(자기 교체)·설계-11·설계-12 닫음(시안: Claude 아티팩트 + DeepSeek v2, `user/updater-화면-*.html`, 사용자가 두 시안 확인 중) -> **다음 = DB 초안 개정** + 미확정 접속 토큰·공용 묶음 잠금·실행 권한. 다 정하면 `REF-update-deploy.md`·`REF-update-rules.md`·시안 일괄 개정
- **(2026-11-02) 다음 세션 첫 일 = 파트(업데이트 app) 설계 재검토·결정** — 코드 없음, 설계·시안만. `REF-review-1030.md` "설계" 표의 설계-1(지우고 깔기)~설계-6(버전 이름 체계)부터 하나씩 정하고, 설계-7(trap 신호)~설계-12(시안 표기)는 그다음. 정해지면 `REF-update-deploy.md`·`REF-update-rules.md`·시안에 반영
  - 볼 것 = `REF-review-1030.md` + `REF-update-deploy.md` + `REF-update-rules.md` + 시안 `user/배포-화면-찰떡.html`(4탭)
  - 배포 화면은 `MemberPage` 타일 틀을 쓰므로 열기 순서(연 곳 바로 뒤, 없으면 맨 앞)·끌기(`DragSession`)를 따른다 → 설계-11(배포 정보 타일 열기)
- **(2026-10-30) 파트(업데이트 app) — 설계만 끝, 코드 착수 전** → `REF-update-deploy.md`(설계·화면) · `REF-update-rules.md`(확정 규칙) · `history/update-deploy.md` · `history/update-rules.md`
  - 확정: 장비마다 `nexus-update` 상주(Job 밖) · 관리자 웹 화면 버튼 · 장비별+전체 배포 · 릴리스 이력+되돌리기 · **Linux 먼저**
  - 릴리스 = 폴더 트리(`common` + `<os>-<arch>`), 판정은 파일 유무 · 단계 = 스크립트 1개 + `steps.json` · **잠금 = arch 단위(`배포 완료` 확정)** · 되돌리기 = 이전 릴리스 재배포 · 버전 = 빌드에 박힌 env 첫 줄 · 내 장비 = 인스턴스별 `배포 정보` 타일(주인도)
  - **키 = `메인키#서브키`(업데이터 인스턴스)** — 같은 이름 두 대도 각각 타일·배포·이력·잠금
  - 화면 시안 = `user/배포-화면-찰떡.html`(4탭, 렌더 9장 = `user/배포-화면-찰떡/`) · 원판 `user/배포-화면.html` · 설계안 `user/배포-설계.html` → **사용자 확인 대기**
  - **다음 = P1(보관·화면)**. 착수 전 확정 3개 = 콘솔 인터랙션(가/나) · 스크립트를 `common`에 둘지 · 첫 시험 장비(`fa24-PID`/`imx8mp-f24`)
  - 참고: 이 세션 = `ANTHROPIC_API_KEY` 인증이라 아티팩트 게시 불가 → HTML을 `user/`에 둠

## 2026-11-02 밤 — P1 마감·P2 진행 중 CURRENT에서 내려보냄(세션 재시작 정리)
  - 확인: `scripts/check/run.sh` 전체 통과(회원 회귀 + `updates-api.py` 39 + `updater-ui.mjs` 27). 화면 = 보드·관리 메뉴 `Updater`(`/updater`)
  - (11-02 밤) 차이-1~5 사용자 확인 끝(다섯 개 모두 이대로) · P1 마감 커밋
  - (11-02 밤) 타일 아래 줄 넘침 수정(안내 글 `...` + 툴팁 `MbNote`, 버튼 줄바꿈) → `REF-design-chaltteok.md` "타일 아래 줄 넘침". 안내 글이 혼자 한 줄이면 줄바꿈(두 줄-1) · 벽에 붙은 입력칸 벽 쪽 선 없앰(벽-2) → `REF-design-chaltteok.md` "벽에 붙은 입력칸"
  - (11-02 밤) 샘플 데이터를 개발 DB `nexus`에 넣음(`sample-jig`·`sample-board`·공용 2개, 인스턴스·잠금 행 없음) → `REF-update-p1.md` "샘플 데이터"

## 2026-11-03 P3 머지 때 내려보냄 (원문)
- **(2026-11-02 밤) 파트(업데이트 app) — P1(보관·화면) · P2(updater Linux) 마감, master 머지 끝**(P1 `1f5eda9` · P2 `bd93fa4`, push = 사용자. P1 마감 때 세부 줄 → `history/current-archive-6.md`). 구현 지도·계획과 달라진 것(차이-1(장비 루트 테이블)~차이-5(폴더 업로드)) → `REF-update-p1.md` · 설계 `REF-update-deploy.md` · 규칙 `REF-update-rules.md`
  - P2(updater Linux) 구현 지도·동작 규칙 → `REF-update-p2.md`
  - **(11-02 밤 23:30~) `imx8mp-f24` 실기 확인 통과**(builtin 3개·script·interactive 입력·재접속·설치 경로 변경·화면) → `history/update-p2.md`. 보드 서비스 `nexus-updater`는 test2에 붙은 채 켜 둠 → `REF-update-p2.md` "남은 것"
  - **P2 마감, master 머지 끝**(`bd93fa4`, push = 사용자). 설계-8(supervisor 되돌리기)·설계-10(자기 교체) 11-02 밤 닫음 → `REF-update-deploy.md` "자기 교체"·"대상의 의존성". **P3(배포) 착수 계획 합의 끝** → `REF-update-p3.md`. **다음 = 사용자 "진행" 신호 뒤 브랜치 `feat/updater-p3`로 단위-1(DB)부터**(Linux 먼저, Windows 뒤)
  - **(11-03 아침) P3(배포) 구현·확인 끝** — 브랜치 `feat/updater-p3`(커밋 안 함), 로컬 `run.sh` 전체 통과 + `imx8mp-f24` 실기 통과 → `REF-update-p3.md` · `history/update-p3.md`. 8012 = P3 코드로 재시작(11-03 07:51, 개발 DB `00018` 적용), 보드 updater = P3 실행 파일(태그는 아직 P2). **11-03 마감 커밋 끝 → 다음 = master 머지 여부(사용자), 그 뒤 Windows(P4) 등은 `BACKLOG.md`**. updater 빌드 태그는 P2 이름 그대로 둠(사용자: 지금은 중요하지 않음)
  - **(11-03 오전) 화-10(타일 크기) 개정, 커밋 안 함(`feat/updater-p3` 작업 트리)** — 관리·Updater 페이지 상세 타일 처음 높이 = 1칸(Run·Log·Script만 2칸), URL 크기는 그대로 따름. `vue-tsc`·`run.sh` 통과, 사용자 화면 확인 대기 → `history/member-ui.md` 11-03. 이때 `REF-member-ui.md` 10k 넘어 `REF-member-ui-impl.md`(구현·판 계약) 분할
  - (11-02 밤) 개발 DB `nexus`에 `00017` 적용 · 샘플 `sample-jig`에 키 발급 · 로컬 시험 updater는 재시작 전에 끔(인스턴스 `asWTzHHK` 행은 offline으로 남음, 지워도 됨)
  - 개발 DB `nexus`에 마이그레이션 `00016`(11-02 17:50)·`00017`(23:13) 적용됨. 8013 빌드본(DB `nexus_build`)은 둘 다 아직 안 들어감

## 2026-11-03 밤 — Files 타일 개선 마감 때 내려보낸 CURRENT 줄(원문)
  - (11-03) steps.json 빈 틀 확장자를 설치본 os로 고름(windows만 = `.ps1`, 섞임 = `.sh` + 안내, 설치본 없음·공용 묶음 = `.sh`) — `UpdStepsTile.vue` + `useUpdDevice(mainKey, enabled)`. master 작업 트리에 미커밋, vue-tsc 통과, 화면 확인 = 사용자
  - **진행 중(11-03 밤) — 브랜치 `feat/update-files-tile`, 단위-가~라 구현 끝, 다음 = 파트 마감(커밋·머지) 준비** → `REF-update-files-tile.md` "단위-나"(나-1~4·끌기-1~3) · 시험 실행 기록 `REF-update-trial.md`. 규칙: 단위마다 코드 전에 이해 확인, 커밋은 파트 마감 때 한 번
    | 순서 | 작업 | 상태 |
    |---|---|---|
    | 0 | 단위-가(폴더 트리 + 선택 삭제) | 끝 · 사용자 화면 확인 끝(11-03) · 확인 `scripts/check/updater-files-ui.mjs` 28/28 |
    | 1 | 시험 실행(시험-1 서버 조회 -> 시험-2 로그 상자 공용화 -> 시험-3 시험 타일 -> 시험-4 입구·구분 표시 -> 시험-5 확인) | **끝(11-03)** — 시험-1~5 전부, run.sh에 `updater-trial-ui.mjs` 추가·통과. 시험 타일 사용자 화면 확인 끝(11-03, "문제없음") |
    | 2 | 단위-나(PC 탐색기 끌어 놓기 + 폴더 이름 유지) | **구현 끝(11-03 밤)** — `updater-files-ui.mjs` 43/43, run.sh 통과. 사용자 test1 폴더 끌어 놓기 확인 끝(11-03 "전부 확인") |
    | 3 | 단위-다(폴더별 합치기/갈아치우기, 부모 갈아치우기면 자식 토글은 켜진 채 흐림) | **구현 끝(11-03 밤)** — Go 시험·run.sh 통과. 사용자 화면 확인 끝(열린 편집 타일 맞추기 포함, 11-03). test2 supervisor 11-03 밤 단위-다 코드로 다시 띄움(`setsid go run`, 로그는 스크래치) |
    | 4 | 단위-라(`실행 켜기` -> `실행 권한` 이름) | **구현 끝(11-03 밤)** — 글자 칸, run.sh 통과. 사용자 화면 확인 대기 |
    | 5 | 파트 마감(커밋·머지, 집 서버 updater 새 빌드 교체 = 사용자) | |
  - **(11-03 끼어든 것) 타일 배치 버그 — 고침**: Grid 폭 = 쓴 열 수(화-7 개정), 점 = 3열 그대로. run.sh 통과, 사용자 화면 확인 끝(11-03) → `history/member-ui.md` 11-03 배치. 파일 = `memberGrid.util.ts`·`memberBoard.store.ts`·`MemberPage.vue`·`memberDrop.util.ts`·`memberDrag.store.ts`·`member-drop.check.ts`
  - **작업 트리 미커밋 목록**(브랜치 위): `UpdFilesTile.vue`·`updActions.hook.ts`·새 `updSelect.hook.ts`·`updTree.util.ts`(단위-가) / `UpdStepsTile.vue`·`updater.api.ts`(steps.json 틀 확장자) / `scripts/check/run.sh`·새 `updater-files-ui.mjs` / 새 `router/updaterTrial.go`·`update.go`·`updates.sql`·sqlc 생성본(시험-1) / `UpdRunTile.vue`·새 `updLog.hook.ts`·`updRun.util.ts`·`styles/updater.scss`·`updater-deploy-ui.mjs`(시험-2) / 새 `UpdTrialTile.vue`·`updater.api.ts`·`updGrid.util.ts`·`UpdTileView.vue`·`UpdBuildTile.vue`·`updater-ui.mjs`(시험-3) / 새 `updDrop.util.ts`·`UpdFilesTile.vue`·`updActions.hook.ts`·`UpdCommonTile.vue`·`updater-files-ui.mjs`·`updater-ui.mjs`(단위-나) / 새 `updstore/replace.go`·`replace_test.go`·`steps.go`·`updateFiles.go`·`update.go`·`updaterDeploy.go`·`updaterRun.go`·`updproto.go`·updater `builtin.go`·`builtin_test.go`·`deploy.go`·새 `updSteps.util.ts`·`UpdStepsTile.vue`·`updater-deploy.py`(단위-다) / 새 `updText.hook.ts`·`UpdTextTile.vue`(열린 편집 타일 맞추기) / 새 `updater-trial-ui.mjs`·`run.sh`(시험-5) / 배치 수정 파일(위 줄) / vault. `apps/frontend/web/`은 사용자 빌드 결과(커밋 대상 아님)
  - (11-03) 화-10(타일 크기) 개정: 관리·Updater 상세 타일 처음 높이 1칸(Run·Log·Script만 2칸) → `history/member-ui.md` 11-03. 이때 `REF-member-ui-impl.md` 분할
