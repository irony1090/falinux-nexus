# REF-update-replace — Files 타일 단위-다(갈아치우기)

→ history/update-replace.md · → REF-update-files-tile.md(결정표 결정-8~15 · 시안 ④~⑦) · → CURRENT
배포 때 updater가 장비 대상 폴더에서 같은 이름 폴더를 **비우고 채울지(갈아치우기) / 겹치는 파일만 바꿀지(합치기)** 폴더별로 고르는 기능. 2026-11-03 `REF-update-files-tile.md`가 10k자를 넘어 분리.

## 계획 · **구현 끝(11-03), Go 시험·run.sh 통과**
| 곳 | 변경 |
|---|---|
| `updstore/steps.go` | `Steps.Replace []string`. 검사: 문자열 배열, 상대 경로(`..`·절대·빈 값 거절), 중복 거절. `CheckRefs`: 그 폴더 아래 메인키 파일이 없으면 경고 |
| `updproto.go` | `Run`·`Deploy`에 `Replace []string json:"replace,omitempty"`. updater `deploy.go` 101행에서 Run으로 넘김 |
| 서버 `updaterDeploy.go` 228행 · `updaterRun.go` 396행 | steps의 `Replace`를 메시지에 채움(롤백 메시지 316행은 불필요 — restore는 백업만 씀) |
| 서버 `update.go` 빌드 상세 | `replace` 필드 추가(토글 표시용, steps.json 오류면 빈 배열) |
| updater `builtin.go` `backup` | replace 폴더 안 기존 파일 중 `run.Files`에 없는 것도 `files/`에 저장(보통 파일 아니면 오류) |
| updater `builtin.go` `copyFiles` | 쓰기 전에 replace 폴더 검사(대상 폴더 안·`guardSelf`, 하나라도 걸리면 아무것도 안 함) -> 폴더 안 파일 지움(`removeFile`, Windows 잠긴 파일 = 이름 변경) -> 기존대로 채움. 빈 하위 폴더는 남겨도 됨 |
| updater `restore` | 변경 없음(지운 파일은 `files/`에 있으므로 되돌아옴) |
| 프론트 `UpdFilesTile.vue` | 폴더 행 `합치기 \| 갈아치우기`(결정-15(토글 모양)), 켤 때 확인 1회 -> steps.json 고쳐 쓰기 |
| 시험 | `steps_test.go`(replace 검사) · `builtin_test.go`(replace backup -> copy -> restore 왕복) |
| 배포 | updater 새 빌드(11-04~ `scripts/build.sh <폴더명> updater`), 집 서버 교체는 사용자 |
- **구현 결과(11-03)** — 단계 이름: 다-1(steps 검사) · 다-2(updater 처리) · 다-3(배포 메시지) · 다-4(steps.json 고쳐 쓰기) · 다-5(폴더 행 토글) · 다-6(확인)
  | 단계 | 실제 코드 |
  |---|---|
  | 다-1(steps 검사) | 새 `updstore/replace.go`: `checkReplace`(배열·`CleanRel`·중복 = error, 부모·자식 = warn) · `EffectiveReplace`(메인키 파일 없는 폴더·부모 있는 자식 뺌) · `SetReplace`(값 자리만 바꿈 / 없으면 `uses` 다음, `uses` 없으면 맨 앞 / 빈 목록 = 항목째 지움, 줄바꿈·들여쓰기·CRLF 따름). 파서 `jnode.end` 추가. `CheckRefs`에 `mainPaths` 인자(폴더에 파일 없음 경고) |
  | 다-2(updater 처리) | `builtin.go`: `replaceDirs`(대상 밖 · updater 폴더 안 · updater 파일을 품은 폴더 거절) · `replaceVictims`(폴더 안 기존 파일 중 이번에 쓰지 않는 것, `*.old-*` 남김, 보통 파일 아니면 오류). `backup` = victims도 `files/`에 저장, `copy` = 검사 통과 뒤 victims 지우고 채움. `guardSelf`의 목록을 `ownPaths`로 뺌 |
  | 다-3(배포 메시지) | `updproto` `Run.Replace`·`Deploy.Replace` · updater `runList`가 넘김 · 서버 `replaceFor`(빌드 메인키 파일 기준) -> `startDeploy`·`updStartRun` |
  | 다-4(steps.json 고쳐 쓰기) | `PUT …/bundles/build/:owner/:version/replace {replace}` = `updSetReplace`(steps.json 없음 404, 문법 오류 422, 결과 검증 오류 422 + 줄). 저장 공통부 `storeText`로 뺌(`updSaveText`와 공용). 빌드 상세 `replace`(steps.json 그대로, 오류면 []) |
  | 다-5(폴더 행 토글) | `UpdFilesTile.vue` 폴더 행 `[합치기 \| 갈아치우기 \| 삭제]`, 켤 때 확인, 끌 때 확인 없음, 부모 켜면 자식 항목 지움, 자식 = 켜진 채 흐림(disabled), 행 주황 강조 + 이유 한 줄, 푸터 `갈아치우기 N곳`. steps.json 없으면 틀(설치본 os 확장자)부터 — 틀은 새 `util/updSteps.util.ts`로 옮겨 steps.json 타일과 공용 |
  | 다-6(확인) | `replace_test.go`(검사·EffectiveReplace·SetReplace 9가지) · `builtin_test.go` 왕복(backup -> copy -> restore)·거절 · `updater-deploy.py` 7절(실제 updater 배포·되돌리기) · `updater-files-ui.mjs` 토글 11개 |
- **화면 확인(11-03 사용자)**: 토글 1~3 확인. 4(steps.json 타일 반영)는 새로고침해야 보임 -> 아래 "열린 편집 타일 맞추기"로 고침
- **열린 편집 타일 맞추기(11-03, 사용자 "제안대로, Script 타일도")**: 새 `hook/updText.hook.ts` `useUpdLiveText` — 빌드 상세의 그 파일 sha가 바뀌면 고치던 것이 없을 때 바로 다시 읽고, 있으면 덮지 않고 푸터 "다른 곳에서 … 바뀌었습니다" + `다시 불러오기`(버릴 때 확인). steps.json 타일·Script 타일 공용. 저장 응답 sha를 맞춰 자기 저장은 다시 읽지 않음. steps.json 없을 때 틀은 고친 것으로 안 침(닫을 때 확인 안 뜸). 같은 탭 변경만 바로 — 다른 곳은 다음 조회 때. 확인 = `updater-files-ui.mjs` 5개
- **폴더를 지워도 replace는 그대로(11-03 사용자 "그냥 냅두자")**: 지운 뒤 같은 이름 폴더를 다시 올리면 갈아치우기가 그대로 살아나야 해서(새 웹 빌드로 바꾸는 흐름). 남은 항목은 검증 경고 + 배포 메시지에서 빠짐(확인-3(사라진 폴더))
- 남은 것: 집 서버 updater 새 빌드 교체(결정-14(옛 updater) — 옛 updater는 replace를 무시하고 합치기로 배포) = 파트 마감 때 사용자

## 착수 전 사용자 확인 (2026-11-03 답 받음)
| 확인 | 답 |
|---|---|
| 확인-1(토글 깊이) | **모든 깊이 폴더**. 부모가 갈아치우기면 자식 토글 = 켜진 채 흐림(부모를 따름, 따로 못 끔). 검사: 부모·자식이 둘 다 replace에 있으면 경고 + 배포 메시지엔 부모만. 토글로 부모를 켜면 자식 항목은 steps.json에서 자동으로 지움(11-03 사용자 "제안대로") |
| 확인-2(steps.json 고쳐 쓰기) | **서버가 `replace` 값 자리만 바꿈**(새 API, 파서 위치 사용). 없으면 `uses` 다음 줄에 새로 넣음, 빈 목록이 되면 항목째 지움 |
| 확인-3(사라진 폴더) | **배포 메시지에서 뺌** + 검증 경고 |
| 확인-4(미커밋 변경) | **이 브랜치에 같이**, 파트 마감 때 커밋 |

