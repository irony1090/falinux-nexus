# REF-update-files-tile — Files 타일 개선 (폴더 트리 · 선택 삭제 · 끌어 놓기 · 갈아치우기)

→ history/update-files-tile.md · → REF-update-trial.md(시험 실행) · → REF-update-replace.md(단위-다 갈아치우기) · → CURRENT · 파트(업데이트 app) 후속. 구현 지도·계획은 코드 착수 때 이 문서에 이어 쓴다.
관련: `REF-update-deploy.md`(설계) · `REF-update-p1.md`(Files 타일 P1) · `REF-update-p4.md`(updater `copy`·`backup`) · `REF-design-chaltteok.md`(묶음 규칙)

## 한눈에 (2026-11-03 밤 마감: 단위-가~라·시험 실행 전부 사용자 확인 끝, 커밋 `3caff7b` + master 머지)
| 단위 | 내용 | 영역 |
|---|---|---|
| 단위-가(폴더 트리 + 선택 삭제) | 폴더 접기/펴기, 길게 눌러 여러 개 선택해 한꺼번에 삭제 | 프론트 |
| 단위-나(끌어 놓기) | PC 탐색기에서 파일·폴더를 Files 타일로 끌어 올리기(폴더 이름 유지) | 프론트 |
| 단위-다(갈아치우기) | 배포 때 장비의 같은 이름 폴더를 합칠지·비우고 채울지 폴더별 선택 | 서버 검증 + updater + 프론트 토글 |
| 단위-라(실행 권한 이름) | `실행 켜기`/`실행 끄기` -> `실행 권한` (버튼이 실행 비트만 바꾸는데 바로 실행되는 줄 오해됨, 11-03 사용자 지시) | 프론트(Files 타일·공용 묶음 타일, `useUpdMode`) |

순서 = (시험 실행) -> 나 -> 다 -> 라. 단위-가는 끝. 브랜치 `feat/update-files-tile` 1개, 커밋은 파트 단위(마감 때 머지).

## 결정표
| 결정 | 내용 |
|---|---|
| 결정-1(선택 방식) | 행을 길게 누르면 선택 모드(마우스·터치). 이후 짧게 눌러 토글. 폴더 행을 체크하면 하위 전체 체크. 푸터 = `N개 선택 \| 전체 선택 · 취소 · 삭제 N`. 기존 길게 누르기 훅(`feature/node/hook/itemPress.hook.ts`) 재사용 가능한지 먼저 본다 |
| 결정-2(삭제 확인) | 확인창 한 번, 개수 + 처음 몇 개 경로. `steps.json` 포함이면 기존 문구 "uses도 비워집니다". 순서대로 지우다 하나가 실패하면 거기서 멈추고 지운 개수·실패 경로를 보인다(원자적 아님) |
| 결정-3(폴더 트리) | 서버에는 폴더가 없고 경로 문자열뿐 -> `/`로 쪼개 트리로 그림. 폴더 행 = 접힘 화살표 · 파일 개수 · 삭제 버튼(하위 전체 삭제, 확인창에 개수). 처음엔 모두 펼침, 접힘 상태는 타일을 닫으면 잊음. 빈 폴더는 없음 |
| 결정-4(끌어 놓기 출처) | **PC 탐색기만.** Nexus 안의 다른 타일(장비 파일 목록)에서 끌어 오는 것은 범위 밖(`transfers.store`와 얽힘) |
| 결정-5(끌어 놓기 규칙) | 폴더 행에 놓으면 그 폴더 안, 파일 행에 놓으면 그 파일이 든 폴더, 빈 곳 = 루트. 메인키 탭에서만 받고 다른 탭은 무시. 폴더째 놓으면 **폴더 이름을 유지**(`web` -> `web/…`). 겹치면 기존처럼 덮어쓰기 확인 1회. 터치에는 없는 기능 |
| 결정-6(폴더 업로드 버튼) | 끌어 놓기와 같은 규칙으로 **최상위 폴더 이름 유지**. 지금 `hook/updActions.hook.ts` `relPathOf`가 `.slice(1)`로 이름을 떼고 올림 -> 고쳐야 함(새로 올리는 것부터만 적용) |
| 결정-7(잠긴 빌드) | 선택 삭제 · 끌어 놓기 · 토글 모두 막힘, 접기/펴기만 |
| 결정-8(갈아치우기의 뜻) | **배포 때** updater가 장비 대상 폴더에서 같은 이름 폴더를 만났을 때 처리. 합치기(기본) = 겹치는 파일만 교체, 나머지는 둠 / 갈아치우기 = 폴더 안 기존 파일을 먼저 전부 지우고 새로 채움. (11-03 처음에 "올린 직후 서버 정리"로 잘못 이해했다가 정정 — 올릴 때 묻는 방식 아님) |
| 결정-9(저장 위치) | 빌드의 `steps.json`에 `"replace": ["web"]`(`uses`와 같은 최상위). 폴더 행 토글이 이 항목을 자동으로 고쳐 씀. `steps.json`이 없으면 틀과 함께 생성. 폴더가 사라져 목록에 없는 경로가 남으면 검증 경고 |
| 결정-10(비우는 시점) | `copy` 시작 때 대상 폴더를 먼저 비우고 그다음 공용 -> 메인키 -> 서브키 순으로 채움(공용이 같은 폴더에 넣는 파일도 비워지지 않고 채워짐) |
| 결정-11(적용 범위) | 메인키 탭 폴더만. 공용 묶음 폴더·서브키 탭은 이번 범위 밖 |
| 결정-12(되돌리기) | 지워질 파일도 `backup` 단계에 들어가 `rollback`의 `restore`가 복원. 백업 용량이 커지는 것은 감수 |
| 결정-13(안전장치) | 갈아치우기 대상은 빌드가 가리키는 장비 대상 폴더 안으로 제한(`..`·대상 폴더 밖 거부) |
| 결정-14(옛 updater) | 막지 않음. 이번 버전부터 첫 배포로 본다(집 서버 updater도 새 빌드로 교체 필요) |
| 결정-15(토글 모양) | 폴더 행 오른쪽 묶음 안, 삭제 버튼 왼쪽에 **`합치기 \| 갈아치우기` 두 칸** 한 묶음(`.ct-grp`), 오른쪽 벽에 붙음. 켜진 칸만 강조(갈아치우기 = 주황 경고색), 켤 때 확인 안내 한 번, 잠기면 두 칸 모두 흐림 |

## 시안 (Claude 아티팩트)
- **확정 = https://claude.ai/artifact/N5USsPArC598DeZy1ApqPK** (v3, 7장면: ① 처음 화면 ② 길게 눌러 선택 ③ 탐색기 끌어 놓기 ④ 갈아치우기 켜기(확인 안내) ⑤ 켠 뒤 ⑥ steps.json 반영·배포 로그 ⑦ 잠긴 빌드). 데스크톱·폰(400px) 렌더 확인함. 시안 장면은 구현 때 `Files 타일` 화면 확인 기준
- 참고 = https://claude.ai/artifact/PvMXf4H2yL66rYETnLrx7G (DeepSeek flash). 두 칸 토글·폴더 `일부 선택` 표시 아이디어만 채택. **steps.json 구조는 틀림**(`build`/`steps`/`copy` 같은 없는 구조, 없는 `nexus deploy` 명령, `저장` 버튼) — 요청에 실제 스키마(`uses`/`deploy`/`rollback`)를 안 넣어서 지어냄. 다음에 DeepSeek에 시킬 땐 실제 스키마 발췌를 넣고 모델은 기본 `deepseek-flash`(내가 `v4-pro`로 바꿔 보냈다가 사용자가 지적)

## 단위-라(실행 권한 이름) — 구현 끝(11-03)
- 사용자 "추천대로" = **글자 칸**: 파일 행 묶음 `[실행 권한 | 755 | 받기 | 삭제]`, 켜짐 = `ct-seg--accent`, 꺼짐 = `ct-seg--dim`. ▶ 아이콘은 "지금 실행"으로 읽혀서 뺌. 툴팁 = "실행 권한 켜기/끄기(장비에 깔릴 때 실행 비트 …, 지금 실행하지 않음)"
- 곳 = `UpdFilesTile.vue`·`UpdCommonTile.vue`(동작 `useUpdMode` 그대로). 확인 = `updater-files-ui.mjs` 2개(.sh 켜짐 -> 끄면 644)

## 구현 때 볼 곳 (착수 전 조사 목록)
| 단위 | 파일 |
|---|---|
| 단위-가·나 | `apps/frontend/src/feature/updater/component/UpdFilesTile.vue`(행·푸터·`remove`) · `hook/updActions.hook.ts`(`useUpdUpload.upload`·`relPathOf`) · `api/updater.api.ts`(`deleteFile` = 파일 1개씩) · `feature/node/hook/itemPress.hook.ts`(길게 누르기) · `feature/node/component/tile/DeviceTileBody.vue` 306행 근처(드롭에서 `webkitGetAsEntry().isDirectory` 처리 선례) |
| 단위-다 | 서버 `apps/core/internal/supervisor/updstore/steps.go`(검증 스키마에 `replace` 추가) · updater `apps/core/cmd/updater/builtin.go`(`backup`·`copyFiles`·`restore`) · 프론트 폴더 행 토글 + `steps.json` 자동 고쳐 쓰기 · `scripts/updater/build.sh`로 updater 새 빌드 -> 집 서버 updater 교체 |

## 주의
- 서버 파일 삭제 API는 파일 1개(`DELETE …/file?path=`). 선택 삭제는 프론트가 차례로 호출(서버 변경 없음)
- 단위-다는 updater·서버·프론트 세 곳이 바뀜 -> 파급이 가-나보다 큼. `replace`를 모르는 옛 updater는 항목을 무시하고 합치기로 배포함(결정-14(옛 updater)로 수용)
- 시안 소스(HTML)는 scratchpad에만 있어 세션이 끝나면 사라짐. 아티팩트에서 `read`로 다시 받을 수 있음

## 구현 계획 (2026-11-03 조사, 브랜치 `feat/update-files-tile`) — 코드는 진행 신호 뒤
> 조사한 파일 = 위 "구현 때 볼 곳" 전부 + `updproto.go`(`Run`·`Deploy`) · `updaterDeploy.go`(228행 배포 메시지) · `updaterRun.go`(396행 개별 실행) · `updater/deploy.go`(101행 Deploy -> Run) · `update.go`(빌드 상세 응답)

### 단위-가(폴더 트리 + 선택 삭제) — 프론트만 · **구현 끝(11-03), 확인 28/28**
| 파일 | 변경 |
|---|---|
| 새 `updater/util/updTree.util.ts` | 경로 목록 -> 트리(폴더 = 하위 파일 개수) -> 접힘 집합을 반영한 화면 행 배열(깊이 포함) |
| 새 `updater/hook/updSelect.hook.ts` | 길게 누르기(마우스·터치 500ms, 10px 넘게 움직이면 취소) -> 선택 모드, 선택 = 파일 경로 집합, 폴더 상태 = 없음·일부·전체 |
| `UpdFilesTile.vue` | 메인키 탭을 트리 행으로. 폴더 행 = 접힘 화살표·이름·개수·[삭제], 파일 행 = 들여쓰기 + 파일 이름. 선택 모드 푸터 = `N개 선택 \| 전체 선택 · 취소 · 삭제 N` |
| `updActions.hook.ts` | `removeMany(paths)` = 확인창 1회(개수 + 처음 5개, steps.json 포함이면 uses 문구) -> 차례로 `deleteFile` -> 실패하면 멈추고 "N개 지움, 실패 경로" |
- 화면 순서 = 폴더마다 파일 먼저, 그다음 하위 폴더(시안 ① 루트 순서). 트리 행에 `data-pager-press-off`(터치 길게 누르기가 타일 끌기로 가지 않게)
- 확인 = `scripts/check/updater-files-ui.mjs`(run.sh에 넣음): 트리·접기·Script 타일 열기·길게 누르기 선택·폴더 일부 선택·전체 선택·선택 삭제·폴더 삭제·한 개 삭제 문구·잠김
- `itemPress.hook.ts`를 그대로 안 쓰는 이유: 마우스 길게 누르기를 건너뛰고(271행), 고름·더블클릭 동작이 섞여 있음. 상수(500ms·10px)만 맞춘다

### 단위-나(끌어 놓기 + 폴더 이름 유지) — 프론트만 · **구현 끝(11-03), 확인 43/43 · run.sh 통과**
| 파일 | 변경 |
|---|---|
| 새 `updater/util/updDrop.util.ts` | `DataTransfer.items` -> `webkitGetAsEntry()` -> 폴더면 `readEntries`를 빈 배열이 나올 때까지 반복(한 번에 100개씩 옴) -> `{file, path}` 목록(폴더 이름 유지) |
| `updActions.hook.ts` | `upload`가 `{file, path}[]`를 받게. `relPathOf`의 `.slice(1)` 제거(결정-6(폴더 업로드 버튼)) |
| `UpdFilesTile.vue` | `dragover`·`drop`. 행의 `data-dir`로 놓을 폴더 결정(결정-5(끌어 놓기 규칙)), 놓일 폴더 행 강조. 메인키 탭 + 안 잠김일 때만 |
- 주의: `items`의 entry는 `drop` 처리 안에서 동기로 먼저 꺼내 둬야 한다(await 뒤엔 비어 있음)
- 11-03 저녁 코드 대조로 확정한 단계(사용자 "제안대로"):
  | 단계 | 내용 |
  |---|---|
  | 나-1(폴더 펼치기) | `updDrop.util.ts`: 놓인 항목 -> `{file, path}[]`, 폴더 이름 유지 |
  | 나-2(올리기 공용화) | `upload(r, items: {file, path}[], have)`. 버튼 쪽은 `relPathOf`(`.slice(1)` 제거)로 만들어 넘김. 겹침 확인 1회 그대로 |
  | 나-3(놓기 화면) | 타일 몸통 `dragover`·`dragleave`·`drop`. 받는 조건 = 메인키 탭 · 안 잠김 · 선택 모드 아님 · 올리는 중 아님. 끄는 동안 = 타일 점선 + 놓일 폴더 행 강조 + 행 끝 "여기에 놓기"(시안 ③) |
  | 나-4(확인) | `updater-files-ui.mjs`: 파일을 폴더 행·파일 행·빈 곳에 놓은 경로, 잠김·다른 탭 무시, 폴더 업로드 버튼 = 폴더 이름 유지 |
  | 끌기-1(안내 문구) | 브라우저는 놓기 전엔 끄는 항목 이름을 안 알려 줌 -> 시안 ③ 말풍선을 "놓으면 `conf/` 안으로 올라갑니다(폴더 이름 유지)"로(이름 없이), 위치는 시안처럼 타일 아래쪽 말풍선 |
  | 끌기-2(빈 폴더) | 서버에 폴더가 없어 빈 폴더는 안 올라감. 놓은 것이 빈 폴더뿐이면 "올릴 파일이 없습니다" |
  | 끌기-3(자동 확인 범위) | headless에선 폴더 entry를 못 만듦 -> 자동 확인 = 파일 끌기 + 폴더 업로드 버튼, 실제 폴더 끌어 놓기는 사용자가 test1에서 직접 |
- **구현 결과(11-03)**: `updDrop.util.ts`(`takeDropped` 동기 꺼내기 · `readDropped(taken, dir)` · `pickedItems`) / `upload(r, items, have)` — 빈 목록이면 `message` "올릴 파일이 없습니다"(끌기-2(빈 폴더)), `UpdCommonTile.vue`도 `pickedItems`로 / `UpdFilesTile.vue` 행 `data-drop-dir`(폴더 행 = 자기 경로, 파일 행 = `parentDir`), 행 밖 = 루트
- 받지 않는 때(다른 탭·잠김·선택 모드·올리는 중)도 `dragover`에서 `preventDefault` + `dropEffect = 'none'` — 안 막으면 브라우저가 파일을 열어 페이지를 떠남
- 안내 말풍선 = 몸통 끝 `position: sticky; bottom`. 목록이 짧으면 시안 ③(푸터 바로 위)과 달리 목록 바로 아래에 뜸
- `updater-ui.mjs`의 폴더 업로드 기대값을 `release/…`(이름 유지)로 고침 — 예전 동작(이름 뗌)을 확인하던 줄

### 단위-다(갈아치우기) → `REF-update-replace.md`
- 계획 표 · 착수 전 사용자 확인(확인-1~4) · 구현 결과(다-1~6)는 분리한 파일에

## 시험 실행 → 별도 REF
- 배포 전 단계별 시험 실행(시험 타일)은 하위 주제가 달라 `REF-update-trial.md`로 분리(11-03). 이 파트의 순서: 시험 실행 -> 단위-나 -> 단위-다 -> 단위-라(실행 권한 이름) -> 마감
