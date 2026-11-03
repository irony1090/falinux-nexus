# history/update-files-tile — Files 타일 개선

→ REF-update-files-tile.md · → history/update-trial.md(시험 실행) · → history/update-replace.md(단위-다) · → CURRENT

## 2026-11-03 (P4 마감 뒤, 사용자 실사용 중 나온 요구)
- **발단**: 집 서버 `home-server` 장비를 만들어 새 빌드를 열었더니 `steps.json` 빈 틀이 `.sh`로 나옴(Windows인데) -> `UpdStepsTile.vue`의 틀 상수가 `.sh` 고정이었음.
- **코드 변경(미커밋, master 작업 트리)**: `UpdStepsTile.vue` 틀을 `template(ext)`로, 같은 메인키 설치본 `os`로 확장자 선택(windows만 = `.ps1`, 섞임·없음·공용 = `.sh`, 섞임이면 안내 줄) + `api/updater.api.ts` `useUpdDevice(mainKey, enabled)`. `vue-tsc` 통과, 화면 확인 = 사용자. 웹만 교체하면 반영(supervisor 재시작 불필요).
- **Files 타일 요구**: 여러 파일 선택 삭제 -> 폴더 접기/펴기 -> PC 탐색기 끌어 놓기 -> 폴더별 갈아치우기. 조율 질문-1~21(번호는 대화 중 임시)을 거쳐 결정-1~15로 정리(`REF-update-files-tile.md`).
- **오해 정정**: "폴더 안 겹치지 않는 파일 정리"를 처음엔 **올린 직후 서버 정리**로 읽어 시안 ④⑤를 그렸다가, 사용자가 **배포 때 updater의 폴더 처리**라고 바로잡음 -> 시안 ④⑤ 폐기·재작성.
- **시안**: Claude 아티팩트(확정 v3) + DeepSeek 별도 시안(참고). 폴더 행 토글은 사용자 의견으로 `합치기 | 갈아치우기` 두 칸 묶음으로 확정.
- **교훈**: ① DeepSeek 요청에 실제 `steps.json` 스키마를 안 넣으면 구조를 지어낸다. ② 모델 기본은 `deepseek-flash`, 임의로 `v4-pro`로 바꾸지 말 것. ③ 전역 TOOLS.md "읽을 조건"을 세션 시작 때 확인해야 함(DeepSeek 도구가 있는데 없다고 답했다가 정정). ④ `ds.sh` 실행 중 `pgrep -af`로 명령줄을 찍으면 API 키가 노출됨 -> 사용자에게 알리고 재발급 권고, 기억 파일 추가.

## 2026-11-03 저녁 — 착수, 단위-가(폴더 트리 + 선택 삭제)
- 브랜치 `feat/update-files-tile`(master 미커밋 `UpdStepsTile.vue` 변경 들고 옴, 확인-4(미커밋 변경)). 조사 뒤 구현 계획을 REF에 쓰고 확인-1(토글 깊이) = 모든 깊이 · 확인-2(steps.json 고쳐 쓰기) = 서버가 자리만 · 확인-3(사라진 폴더) = 배포 메시지에서 뺌 답 받음. 부모·자식 토글 처리 제안 후 사용자 "진행해줘"
- 단위-가 코드: `util/updTree.util.ts`(새) · `hook/updSelect.hook.ts`(새) · `updActions.hook.ts` `useUpdRemove` · `UpdFilesTile.vue` 메인키 탭 트리. vue-tsc 통과
- 확인 `scripts/check/updater-files-ui.mjs` 28/28(격리 5051). 처음 실패 2건은 검사 쪽(Script 타일 키 = `tb…~경로 인코딩`, 머리 줄에 가린 행 = `scrollIntoViewIfNeeded`)

## 2026-11-03 밤 — 단위-가 확인, 요구 추가, 세션 정리
- 사용자가 단위-가(폴더 트리 + 선택 삭제) 화면 확인 끝. 아직 미커밋(브랜치 `feat/update-files-tile` 작업 트리)
- 사용자 요청: VM `nexus-win`을 켜고 updater 하나 실행(`test-win`, test2) -> `REF-windows-vm.md`. 실행 중 `00-stop.ps1`을 시험하려다 나온 질문들이 요구 둘로 이어짐
- **단위-라(실행 권한 이름)**: Files 타일 `실행 켜기`는 실행 비트(111)만 바꾸는데 바로 실행되는 줄 오해됨 -> `실행 권한`으로 바꾸기(사용자 지시, BACKLOG에서 이번 파트 작업으로 올림)
- **시험 실행**: 배포 전 단계별 시험 요구 -> 조사·계획·결정-1~9 확정 -> `REF-update-trial.md` / `history/update-trial.md`로 분리
- 교훈: `배포` 버튼은 대상 고르기 팝업을 열 뿐이지만 팝업의 `배포 N대`는 확인 없이 바로 실행(대상 기본 = 붙어 있는 한가한 인스턴스 전부). 사용자가 먼저 걱정함

## 2026-11-03 저녁 — 단위-나 계획 확정 (코드 전)
- 시험 실행(시험-1~5) 끝난 뒤 단위-나 계획을 코드·시안 ③과 대조해 다시 보임 -> 끌기-1(안내 문구에서 이름 빼기 — 놓기 전엔 이름을 못 앎)·끌기-2(빈 폴더 알림)·끌기-3(폴더 끌기는 사용자 직접 확인) 사용자 "제안대로"
- 코드는 다음 세션(사용자: 세션 재시작 후 진행) -> `REF-update-files-tile.md` "단위-나"

## 2026-11-03 밤 — 단위-나(끌어 놓기 + 폴더 이름 유지) 구현
- 사용자 "진행". 나-1(폴더 펼치기) `util/updDrop.util.ts`(새) · 나-2(올리기 공용화) `updActions.hook.ts` `upload(r, items, have)`, `relPathOf` 제거 -> `pickedItems`(Files·공용 묶음 타일) · 나-3(놓기 화면) `UpdFilesTile.vue` 점선·폴더 행 강조·"여기에 놓기"·안내 말풍선(끌기-1(안내 문구) = 이름 없이)
- 나-4(확인) `updater-files-ui.mjs` +15 = 43/43(폴더 행·파일 행·빈 곳 경로, 겹침 확인, 다른 탭·잠김 무시, 폴더 업로드 이름 유지). 전체 run.sh 처음 실패 1건 = `updater-ui.mjs`가 옛 동작(첫 폴더 이름 뗌)을 기대 -> 기대값 고침 + 업로드 끝날 때까지 기다리게, 다시 돌려 통과
- 결정: 못 받는 때도 `dragover` 기본 동작을 막음(브라우저가 파일을 열어 페이지 이탈 방지). 실제 폴더 끌어 놓기는 사용자 확인 대기(끌기-3(자동 확인 범위))
- 사용자가 test1에서 폴더 끌어 놓기·빈 폴더 확인 "전부 확인" -> 단위-나 끝
- 단위-다(갈아치우기) 구현 이력 → `history/update-replace.md`

## 2026-11-03 밤 — 단위-라(실행 권한 이름)
- 두 안(글자 칸 / 아이콘만 교체) 중 사용자 "추천대로" = 글자 칸. Files·공용 묶음 타일. 확인 스크립트 처음 실패 = 화면이 다시 조회되기 전에 검사(다른 행 칸이 조건을 먼저 만족) -> 그 칸만 기다리게 고침. 전체 run.sh exit 0
- 같은 날 결정: 갈아치우기 폴더를 지워도 replace는 그대로(사용자 "그냥 냅두자") -> `REF-update-replace.md`

## 2026-11-03 밤 — Files 타일 개선 마감
- 사용자 "마감 진행" -> 코드 커밋 `3caff7b`(회원 배치 수정·steps.json 틀 확장자 포함) -> vault 커밋 -> master `--no-ff` 머지. 빌드 확인 중 남은 `apps/core/updater` 실행 파일(Claude가 `go build`로 만든 것)은 지움
- HISTORY 10k 대응: 마감 파트 다섯 개를 INDEX로 접음(file-view·view-tile·video-view·tile-drag·member-drag, 사용자 "제안대로")
- 남은 것 = 집 서버 updater 새 빌드 교체(사용자) · 단위-7(교체 빌드) 정-1~3 답 대기
