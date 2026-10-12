# REF — 공유-e(프론트 표시)

> 2026-10-08 신설. 결정 SH1~SH10 → `REF-process-share.md` / 서버 응답 → `REF-process-share-api.md` / 머리 줄 → `REF-workspace.md` / 이력 → `history/process-share-display.md` / 현재 → `CURRENT.md`.

**상태(2026-10-08): 공유-e-0~3c 코드 완료 + 사용자 직접 확인 통과(커밋 전). 공유-f(presence) 코드 완료, 확인 전 -> `REF-process-share-presence.md`.**

## 작업 단위 (번호-이름 대응)
| 단위 | 내용 | 상태 |
|---|---|---|
| 공유-e-0(확정) | 제안 표시 항목 확정 | 완료 |
| 공유-e-1(받는 쪽 표시) | ①~④·⑥ + 서버 응답 3곳 추가 | 코드·격리 확인 완료(커밋 전) |
| 공유-e-2(공유 걸기 창) | 아이디 조회 → 닉네임 확인 + 단계 → 완료. O1(경로 줄 공유 버튼) 진입, 폴더만 | 코드·격리 확인 완료(커밋 전) |
| 공유-e-3a(공유 목록 창) | O1 숫자 + O2(적용 공유 목록). 기존 API만, 프론트만 | 코드·격리 확인 완료(커밋 전) |
| 공유-e-3b(행 공유 개수) | O3. 폴더 목록 응답에 `shareCount`(소유자 REST) — 서버 쿼리 1개 + 프론트 | 완료(2026-10-08, 사용자 직접 확인 통과, 커밋 전) |
| 공유-e-3c(내가 공유한 것) | O5. `GET /shares/mine`(노드 + process 단독, 해제만) 신설 + 머리 메뉴 항목 | 완료(2026-10-08, 사용자 직접 확인 통과, 커밋 전) |
| 뒤로 미룸 | 숨기기(+복원 UI "숨긴 공유 N"), O4(접속 표시) = 공유-f(presence) | — |

## 공유-e-0(확정) 결과 (2026-10-08)
| 표시 | 결정 |
|---|---|
| ①(공유 색) | 채택. 타일 왼쪽 띠 + 머리 줄 바탕, 테마 색 이름 |
| ②(공유한 사람 칩) | 채택. 공유 폴더 타일(+그 아래 실행 터미널)에만 |
| ③(단계 아이콘) | 채택. 보기(10) `visibility` + "보기 전용" / 조작(20) `keyboard` / 편집(30) `pencil` |
| ④(아이콘 변형) | 채택. 공유받은 폴더 `mdi-folder-account` |
| ⑥(실행한 사람 표시) | 채택. `startedBy`가 소유자와 다를 때만, **소유자 쪽 화면에도** 같은 규칙 |
| B(역방향 연결) | 제외 — 공유-d 6a(스크립트 행 "실행 중 N")로 구현됨 |

## 지금 코드 (2026-10-08 확인)
| 항목 | 서버 | 프론트 |
|---|---|---|
| 노드 `permission` | 줌 | 안 씀 |
| process `permission` | 줌 | 스크립트 행 종료 버튼만 |
| `ownerNickname` | 공유받은 루트 목록 항목만 | 폴더 행 `← kim`만 |
| `startedBy` | `/nodes/:id/runs`만 (process 응답엔 없음) | 스크립트 행만 |
| 머리 줄 `kim /`(⑤) | `useNodePath` 결과에 소유자 이름이 있는지 미확인 | — |

## 공유-e-1(받는 쪽 표시) 변경안 (계획)
- 서버(REST 응답에만): `processDto.go` `ProcessResponse`에 `startedBy`·`ownerNickname` / 처리 쿼리에 닉네임 JOIN(`shares.sql` runs 쿼리에 같은 JOIN 있음) / `GET /nodes/:id`에 공유받은 노드일 때 `ownerNickname`
- 프론트: `node.api.ts`·`process.api.ts` 타입 / `TileFrame.vue` 머리 줄(①②③⑥) / `FolderTileBody.vue` 폴더 행 아이콘(④) / 판정 함수 한 곳(`shareView.ts` 가칭: 공유받은 것인가·보기 전용인가·단계 이름)

## 공유-e-1 구현 메모 (2026-10-08, 커밋 전)
| 파일 | 변경 |
|---|---|
| `query/users.sql` + gen | `ListUserNicknames(ids[])` 신설 (sqlc generate) |
| `router/processDto.go` | `processResponse`에 `ownerNickname`(소유자가 요청 계정이 아닐 때만)·`startedBy`(소유자와 다를 때만) + `withProcessNames`/`processesOf` |
| `router/processApi.go` | 목록·lookup·exec·resize 응답에 `withProcessNames` 적용 (토픽 발행엔 안 씀) |
| `router/node.go` `getNode` | 소유자가 요청 계정이 아니면 `ownerNickname` |
| `process.api.ts` | 타입에 `permission`·`ownerNickname`·`startedBy` |
| `processTerm.store.ts` | `PROCESS:UPDATE`가 계정 기준 세 값을 덮지 않게 보존(안 하면 이벤트 한 번에 표시가 사라짐) |
| `node/util/shareView.util.ts` (신설) | `PERM_*`·`permInfo`(편집 `mdi-pencil` / 조작 `mdi-keyboard` / 보기 `mdi-eye`)·`isReadOnly` |
| `TileFrame.vue` | ①(`TileFrame--shared`, 테마 색 `info`) ②(공유 시작 지점 = `parentVisible === false`인 폴더 타일에 칩) ③(단계 아이콘+이름, 보기 단계 터미널에 "보기 전용 · 입력 안 됨") ④(`mdi-folder-account`) ⑥(`실행: 이름`) ⑤(터미널 경로 앞에 소유자 닉네임) |
| `FolderTileBody.vue` | 공유받은 루트 폴더 행 아이콘 ④ |

- 정한 것: ②는 공유 시작 폴더 타일에만(터미널은 ⑤의 `kim /` 경로 접두어로 갈음). 공유 폴더 **안의** 하위 폴더 행은 ④ 표시 안 함(응답에 소유자가 없어서)
- 확인할 것: 계정 둘 — 받는 쪽 폴더 타일 ①②③④ / 실행한 터미널 ⑤⑥ / 보기 단계 "보기 전용" / 소켓 이벤트 후에도 표시 유지 / 소유자 쪽 ⑥ / 모바일 머리 줄 넘침

## 공유-e-1 확인 결과 (2026-10-08, 격리 실행: 임시 DB + 5051 + 격리 프론트 3001 + worker, 헤드리스)
| 확인 | 결과 |
|---|---|
| 받는 쪽 폴더 타일 ①②③④ | 통과 (왼쪽 띠·바탕, 소유자 칩, 단계 아이콘, 공유 폴더 아이콘) |
| 받는 쪽 실행한 터미널 ⑤⑥ | 통과 (`[2] own / shareF / run.sh 실행: view`) |
| 소유자 쪽 | 공유 표시 없음, ⑥만 표시 (통과) |
| 보기 단계 "보기 전용 · 입력 안 됨" | 통과 (reload 후) |
| 단계를 바꿀 때 실시간(조작→보기→편집→보기→조작) | **터미널 타일은 안 바뀜 → 수정 후 통과** |
| 공유 풀기(배너)·다시 걸기(복원) | 통과 |
| 모바일 390px 머리 줄 | 넘침 0px (통과) |

- **발견·수정**: `SHARE:CHANGE` 때 폴더 타일은 노드 재조회로 바뀌지만 터미널의 `permission`은 process 스토어에 남아 있었음 → `processTerm.store.ts` `refreshAccount`(계정 기준 세 값만 갱신) + `terminalRestore.hook.ts` `recheck`가 조회 결과마다 호출
- ~~남은 사소한 것: ⑥이 내가 실행한 것에도 `실행: 내 닉네임`으로 나옴~~ → 2026-10-12 "나"로 수정(서버 `startedByMe`) → `REF-process-share-verify.md`
- 확인 스크립트(scratchpad `pw/lib.mjs`·`setup.mjs`·`view1~4.mjs`, `vite.iso.mts`, 재부팅 시 사라짐). 격리 프론트는 `vite`를 import하지 말고 기존 설정 객체를 펼쳐 쓰면 됨(scratchpad에서 `vite`를 못 찾음)

## 공유-e-2(공유 걸기 창) 구조안 (2026-10-08 확정, 코드 없음)
시안: https://claude.ai/artifact/JaeMH1PY8YgcEbJpddDLs8 ("공유 걸기 창 시안", 직접 눌러 보는 데모)

```
FolderTileBody 경로 줄  -> ShareDialog (신설)                    서버
 [O1 공유 버튼] click     step1 아이디 입력 -> 찾기            GET /users/lookup?identification=
                           (404 없는 아이디 / 400 자기 자신 = 같은 창에 문구)
                          step2 닉네임 확인 + 단계(보기/조작/편집, 기본 보기)
                           공유                                PUT /nodes/:id/shares/:userId {permission}
                          step3 완료 한 줄 (previous 있으면 "A -> B로 변경")
```
| 파일 | 내용 |
|---|---|
| `node/api/share.api.ts` (신설) | `lookupUser`·`putNodeShare` + 타입 |
| `node/component/dialog/ShareDialog.vue` (신설) | 2단계 창, `NodeNameDialog` 패턴(`v-model`·`VDialog`·`VCard`), 단계는 `v-radio-group` |
| `FolderTileBody.vue` | 경로 줄 `.tools`에 O1 버튼, **소유자(40)이고 `current` 있을 때만** |
| `shareView.util.ts` | 단계 선택지·설명 재사용 |

| 결정 | 내용 |
|---|---|
| 차단(0) 선택지 | 이번엔 없음(상위 공유 여부는 공유-e-3 목록이 있어야 앎) |
| O1 숫자(적용 인원) | 공유-e-3에서 |
| O1 보이는 조건 | 소유자(40)만 |
| 기본 단계 | 보기(10) |
| 완료 안내 | 창 안 한 줄 후 닫기(기존 토스트 컴포넌트 있는지는 착수 때 확인) |
- 확인 계획: 격리 실행 — 없는 아이디 / 자기 자신 / 걸기 → 받는 쪽 타일 / 같은 사람 단계 변경(`previous`) / 비소유자에게 O1 안 보임 / 모바일 폭

## 공유-e-2 구현 메모·확인 결과 (2026-10-08, 커밋 전)
| 파일 | 변경 |
|---|---|
| `node/api/share.api.ts` (신설) | `lookupUser`·`putNodeShare` + 타입(`NodeShareEntry`·`PutNodeShareResponse`) |
| `node/component/dialog/ShareDialog.vue` (신설) | 3단계(아이디 → 확인 + 단계 → 완료). 404/400은 `status`로 문구 분기, 완료 문구는 `previous` 비교(바뀜 / 이미 같은 단계 / 새로 공유) |
| `node/util/shareView.util.ts` | `PERM_OWNER`·`SHARE_LEVELS`·`levelTitle` 추가 |
| `FolderTileBody.vue` | 경로 줄 O1 버튼(`mdi-account-multiple-plus`, `info` 색) — `current.permission === PERM_OWNER && kind === FOLDER`일 때만 + `ShareDialog` |

- 확인(격리 실행, 임시 DB + 5051 + 격리 프론트 3001, 서버 변경 없음, 11개 통과): 루트 타일에 O1 없음 / 내 폴더 타일에 O1 / 없는 아이디 문구 / 자기 자신 문구 / 닉네임 확인 + 기본 보기 / 조작으로 공유 → 서버 단계 20 / 받는 쪽 타일이 실시간으로 생김 / 같은 사람에게 편집으로 → "조작 → 편집" / 받는 쪽 폴더 타일에 O1 없음 + 모바일 390px 창 정상
- 함정: 계정 동기화로 소유자 보드에는 이미 열린 타일이 있어, 확인 스크립트에서 "새 타일"을 다시 누르면 O1이 2개가 됨(앱 문제 아님)

## 공유 표시 색 = 청록 `share` (2026-10-08 사용자 결정)
- 공유 관련 표시만 청록. 다른 곳의 파랑(`primary`·`info`)은 그대로. `plugins/vuetify.ts`에 `share`(라이트 `#1f8f86` / 다크 `#4fd1c5`)·`on-share` 등록(테마 색 이름 규칙 유지)
- 적용: `TileFrame.vue`(띠·머리 줄 바탕·칩·단계 표시) / `FolderTileBody.vue` O1 버튼 / `ShareDialog.vue` 아바타·라디오. 격리 화면 확인(띠 `rgb(31, 143, 134)`). 시안 두 개도 같은 색으로 갱신. **공유-e-3·공유-f 표시도 `share` 색 사용**

> 공유-e-3(소유자 쪽 목록: 구조안·3a~3c) -> `REF-process-share-owner.md` / 공유-f(presence) -> `REF-process-share-presence.md` (2026-10-08 분할)
