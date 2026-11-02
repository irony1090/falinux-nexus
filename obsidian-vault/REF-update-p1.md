# REF — 업데이트 app P1(보관·화면) 구현

> 2026-11-02 신설. 설계·계획 → `REF-update-deploy.md` "P1 착수 계획" / 규칙 → `REF-update-rules.md` / 이력 → `history/update-p1.md` / 현재 진행 → `CURRENT.md`
> 상태(2026-11-02 저녁): **단위-1(DB)~단위-8(확인) 코드 끝, 확인 스크립트 전부 통과. 브랜치 `feat/updater-p1`, 커밋 안 함(파트 마감 때). 차이-1~5 확인 끝 · P1 마감 커밋(11-02 밤), master 머지 대기**

## 파일 지도
| 단위 | 파일 |
|---|---|
| 단위-1(DB) | `apps/core/internal/supervisor/db/migrations/00016_updates.sql` · `query/updates.sql` |
| 단위-2(보관소) | `apps/core/internal/supervisor/updstore/store.go`(폴더·sha·텍스트 판정) · `steps.go`(steps.json 검증, 위치 기억 JSON 파서) · `steps_test.go` |
| 단위-3(API) | `cmd/supervisor/router/update.go`(장비 루트·번들) · `updateFiles.go`(텍스트·검증·읽기·삭제·조각 업로드) · `updateZip.go` · env `UPDATE_STORE_DIR`(`constants/env.go`) |
| 단위-4(페이지 틀) | `feature/member/component/MbTileFrame.vue`(새, 공용 틀) · `MemberTileView.vue`(회원 몸통만) · `MemberPage.vue`(슬롯 `list`·`tile`) · `memberBoard.store.ts`(`PAGE_TABS`·키 규칙·`requestClose`/`guardClose`) · `memberGrid.util.ts`(`GridKeys`·`MEMBER_KEYS`) · `memberUrl.hook.ts` · `pages/Updater.vue` · 라우트 `/updater` · 메뉴(보드·관리) |
| 단위-5·6(타일) | `feature/updater/`: `api/updater.api.ts` · `util/updGrid.util.ts`(키) · `hook/updActions.hook.ts` · `component/Upd{List,TileView,DeviceTile,BuildTile,FilesTile,StepsTile,TextTile,CommonTile,ZipMenu,Prompt}.vue` · `styles/updater.scss` |
| 단위-7(CtCode) | `feature/widget/component/CtCode.vue`(CodeMirror 6) · `ScriptEditDialog.vue` 교체 |
| 단위-8(확인) | `scripts/check/updates-api.py`(API 39개) · `updater-ui.mjs`(화면 27개, `SHOT_DIR`이면 스크린샷) · `run.sh`에 이어 붙임 |

## 계획과 달라진 것 (구현 중 결정, 2026-11-02 사용자 확인 — 다섯 개 모두 이대로)
| 이름 | 무엇 | 이유 |
|---|---|---|
| 차이-1(장비 루트 테이블) | 테이블 7개 + **`update_devices`**(메인키 루트·메모) = 8개 | 빌드·인스턴스가 없는 장비 루트도 목록에 있어야 함(P1은 손으로 만듦) |
| 차이-2(보관 폴더 접두어) | `<STORE>/devices/<메인키>/<버전>/` (설계 = `<STORE>/<메인키>/...`) | 메인키 `common`이 공용 폴더와 부딪힘 |
| 차이-3(uses 순서) | `update_uses.ord` 추가 | 합친 결과(zip B)·깔 순서 = `uses` 배열 순서 |
| 차이-4(빌드 타일 행) | `올린 때` 행 = 시각 · 올린 사람, `고침` 시각은 `상태` 행에 | 800px 높이(폰·흔한 PC)에서 1×1 빌드 타일에 `열기` 줄까지 들어가게 |
| 차이-5(폴더 업로드) | Files 타일에 `폴더 업로드` 버튼(고른 폴더 = 번들 맨 위, 첫 폴더 이름 뗌) | P1 확인 기준 "올린 폴더가 그대로" |

## 동작 규칙 (구현에서 정한 것)
- **쓰기 순서**: 검사 -> 번들 행 `FOR UPDATE`(잠김이면 409) -> DB -> 파일 이름 변경 -> 커밋. 이름 변경을 행 잠금 안에서 해야 같은 파일 동시 저장이 DB와 안 어긋남. 삭제(번들·장비 폴더)만 `AfterCommit`
- **복사로 새 버전** = 하드링크(안 되면 복사). 덮어쓰기는 항상 임시 파일 -> 이름 변경이라 링크된 원본은 안 바뀜(확인 스크립트로 확인)
- **보관소 표식** = `<STORE>/.nexus-db`(호스트:포트/DB), 다른 DB 것이면 보관소 끔(`VIEW_CACHE_DIR`과 같은 방식). 끈 상태면 쓰기 API 503
- **텍스트 판정** = 1MB 이하 · NUL 없음 · 제어 문자(탭·줄바꿈·ESC 등 제외) 1% 이하 · UTF-8. 텍스트만 Script 타일로 열림
- **`.sh` LF** = 웹 편집 저장 때만. 업로드한 파일은 바이트 그대로(CRLF여도), Script 타일에 `CRLF`로 표시되고 저장하면 LF가 됨(CodeMirror가 줄 끝을 LF로 다룸)
- **steps.json 검증** = 서버 한 곳(`updstore.Check`·`CheckRefs`): 문법·구조 오류 + 없는 공용 묶음 = error(저장 거절, 422 + 줄 목록), 아직 없는 스크립트 파일·모르는 키 = warn(저장됨). 문법·구조 오류가 있어도 참조 검사까지 한 번에 보여줌. 저장 성공 때 `update_uses` 다시 만듦, steps.json 삭제 = uses 비움
- **업로드** = 시작(이름 겹침 = 409 `exists`, 화면은 겹치는 이름을 미리 세어 한 번에 확인) -> 512KB 조각 PUT -> 마지막 조각에서 sha·DB·이름 변경. 멈춘 업로드 30분 뒤 정리, 서버 시작 때 `.tmp` 비움
- **zip** = 만들면서 스트리밍. 중간 실패면 끝 목록을 안 써서 깨진 zip으로 알게 함(머리는 이미 보냄)
- **P1 서브키** = 읽기·zip은 다루고 쓰기는 400("자리만")
- **닫기 확인** = 보드 `requestClose`/`guardClose`(공용 틀). 화면 닫기 버튼·컨트롤 패널 닫기가 거침. 저장 안 한 Script·steps.json 타일이 씀

## 타일 키 (URL `t`)
`u<메인키>` 장비 · `b<메인키>~<버전>` 빌드 · `f…` Files · `s…` steps.json · `c<이름>~<버전>` 공용 · `t<b|c><주인>~<버전>~<경로 퍼센트 인코딩>` Script. `~`는 이름 글자에 없음, 경로 인코딩에서 `~` `*`도 바꿈(URL 구분자 `*` `^` `,`와 안 겹치게)

## 함정
- 격리 supervisor `-env`는 **절대 경로**(상대면 dev 판정 때 프로젝트 폴더 기준으로 찾아 panic)
- 확인 스크립트에서 확인 창을 닫은 뒤 `Ctrl+S`는 편집기에 포커스가 없어 브라우저로 감 -> 편집기를 다시 누르고 보냄
- `run.sh`가 `UPDATE_STORE_DIR`을 안 주던 때(11-02 첫 실행) 격리 supervisor가 기본 폴더 `apps/core/cmd/supervisor/update-store`에 `nexus_check_tmp` 표식을 남겨, 뒤에 띄운 개발 supervisor(DB `nexus`)가 보관소를 끔 -> 그 폴더(표식만 있었음)를 지우고 재시작. 지금 `run.sh`는 임시 폴더를 줌

## 샘플 데이터 (개발 DB `nexus`, 2026-11-02)
- 넣은 법 = 개발 API(8012, 계정 irony)로 스크립트 실행. 기존 `fa24-PID`(빌드 `qwe`, 사용자가 만든 것)는 안 건드림
- 공용 `nexus-worker@v1`(worker 바이너리 900KB·update.conf) · `jig-tools@v1`(lib/common.sh·tools/check-port.sh)
- 장비 `sample-jig`: `v1`(스크립트 2·app.bin 1.2MB·config/app.conf·update.conf가 공용을 덮음·steps.json uses 2개, `02-verify.sh` 없어 경고) · `v2`(v1 복사 + 02-verify.sh·CRLF 파일 `win-note.sh`)
- 장비 `sample-board`: 빌드 없는 루트
- 안 넣은 것: 가짜 인스턴스(`updaters`) 행 · 잠금(`update_locks`) 행 — 잠금은 해제가 없어서
