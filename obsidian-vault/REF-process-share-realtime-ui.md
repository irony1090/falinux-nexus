# REF — 공유-d 실시간-0(터미널 UI): 최소화 · kill 확인 · 머리 줄

> 2026-10-07 신설(`REF-process-share-realtime.md`에서 분할). 공유-d 전체 결정·작업 단위 → `REF-process-share-realtime.md` / 이력 → `history/process-share-realtime-ui.md` / 시안 → user vault `user/links.md` "터미널 최소화 시안" / 현재 진행 → `CURRENT.md`.

**상태(2026-10-07): 코드 작성·격리 실행 확인 완료(24/24) + 사용자 화면 확인 완료(커밋 전 — 파트 마감 때). 다음 = 실시간-1(재계산 함수).**

## SD3 확정 전문 (2026-10-07 사용자)
**개정(2026-10-07 사용자): 닫기 대신 최소화.** 실행 중 터미널은 내 것·남의 것 모두 **최소화** 가능(닫기는 지금처럼 실행 중엔 막힘). 최소화 = 타일은 트리에 그대로 두고 `minimized: true` 표시만 → 화면 배치에서 빠짐, 다시 열면 원래 자리·크기. 보드 **루트 폴더 타일에 "이 보드의 터미널 N · 최소화 n" 목록**(타일 위치 또는 최소화, 보기 / 열기). uid가 트리에 남으므로 고아 안전망·WS5(터미널 타일 추가 주체) 그대로, 다시 열기 API 불필요(PUT으로 표시만 바꿈). 처음 안(남의 것만 닫기 허용 + 안전망 `started_by` = 나 + 스크립트 행 "실행 중 N"·다시 열기 API)은 대체됨. 루트 타일 = `[파일] [터미널 N · 최소화 n]` 탭(**최소화가 있으면 터미널 탭이 기본**, 터미널 탭은 최소화 먼저) + 스크립트 행 "실행 중 N · 최소화 n"(모든 보드의 그 스크립트 터미널: 보기 / 열기 / 보드로 이동). **두 목록에서도 kill 가능, kill은 어디서든(머리 줄 포함) 확인창 뒤 실행**(2026-10-07 사용자). 확인창 버튼 = 취소 / 종료(타일은 DONE으로 남음) / **종료하고 닫기**(process가 끝났다는 STATUS를 받은 뒤 타일까지 치움 — kill은 신호만 보내고 종료는 비동기라, 끝나지 않으면 타일을 남기고 알림). DONE 터미널은 머리 줄 X + 루트 터미널 탭 행 X로 닫음. **머리 줄 = A(창 버튼 묶음) 채택(2026-10-07 사용자)**: 왼쪽 정보 + 오른쪽 `RUN▾` 칩(눌러서 종료… / 종료하고 닫기…) + 테두리로 묶은 `[최소화 | 크기 | ×]`(이 순서, 크기는 아이콘 → 기존 가로·세로 0.5/1.0 창). X는 항상 눌림 — 실행 중이면 kill 확인창(취소 / 종료 / 종료하고 닫기, 선택지 2번)과 같은 창. kill 버튼을 창 버튼 옆에서 뗌(오클릭 방지). 폴더 타일은 `[크기 | ×]`. 스크립트 행 목록은 다른 보드 위치가 필요해 서버가 스크립트별 실행 중 목록·위치를 내려줌. 시안 → user vault `user/links.md` "터미널 최소화 시안"(이전 "터미널 다시 열기 시안"은 비교용으로 남김)

## 지금 코드 (2026-10-07 확인)
| 자리 | 지금 |
|---|---|
| 머리 줄(`TileFrame.vue`) | 정보 · RUN 칩 · 실행 중이면 kill 버튼(확인 없이 바로) · 크기 버튼 · 보드 메뉴(폴더) · X(실행 중·복원 전이면 비활성) |
| 타일 필드 | `TileNode`(`tile.type.ts`) = id·type·kids·size·adopted. 서버 `tileItem`(`tile.go`)은 모르는 필드를 저장 때 버림 → `minimized`는 서버 구조체에 넣어야 남음 |
| 배치 | `tileGrids.store` `order = flatten(tiles, rootId)` → `pack` → `posOf`·`instanceOf`. 최소화 개념 없음 |
| 확인창 | `useAppDialog().openDialog({ buttons })`가 버튼 여러 개를 받음 → 새 컴포넌트 없이 3버튼 확인창 가능 |
| 실행자 정보 | `ProcessResponse`에 실행한 계정 없음 → "남이 실행한 것" 안내 줄은 지금 못 그림 |
| xterm | `ProcessTerminal.vue` 언마운트 때 detach, 다시 마운트 때 attach(같은 xterm 재사용) → 최소화해도 버퍼 유지 |

## 구현 계획 (2026-10-07 확정)
| 파일 | 변경 |
|---|---|
| `tile.go` | `tileItem`에 `Minimized bool \`json:"minimized,omitempty"\``. `validate`: 폴더에 `minimized`면 400 |
| `tile.type.ts` | `TileNode`에 `minimized?: true` |
| `tileTree.store.ts` | `minimizeOp(tileId, on)`(터미널만, 멱등) + `setMinimized`. `closeOp`는 그대로(넘겨받은 터미널이 `minimized` 유지) |
| `tileGrids.store.ts` | 배치용 `order` = 최소화 제외 / `instanceOf` = 최소화 포함 전체로(번호가 최소화로 흔들리지 않게) / 루트 탭용 `terminals`(보드의 터미널 전부, 위치 번호 또는 최소화) |
| `hook/terminalKill.hook.ts`(새) | `confirmKill(tileId)`: 확인창 취소 / 종료 / 종료하고 닫기. 종료하고 닫기 = kill → 그 uid 상태가 끝남이 될 때까지 기다림 → 닫기+dispose. 제한 시간 넘으면 타일 남기고 알림 |
| `TileFrame.vue` | 머리 줄 A안: 정보 · `RUN▾` 칩(실행 중일 때 메뉴: 종료… / 종료하고 닫기…) · 보드 메뉴(폴더) · 테두리 묶음 `[최소화 \| 크기 \| ×]`(폴더 = `[크기 \| ×]`, 루트 = `[크기]`). 크기 = 아이콘 → 기존 가로·세로 창. X 항상 눌림: 실행 중이면 `confirmKill`, 아니면 닫기. 복원 전(상태 모름)만 비활성 유지 |
| `RootTerminalList.vue`(새) + 루트 타일 | 루트 타일(parent 없음)에만 탭 `[파일] [터미널 N · 최소화 n]`. 터미널 탭 = 최소화 먼저, 그다음 배치된 것(위치 번호). 행 = 이름·경로·상태 + 보기(배치된 것: `reveal`) / 열기(최소화 해제 + `reveal`) / kill(`confirmKill`) / DONE이면 X |

## 확인할 것 (2026-10-07 제안대로 확정)
| 질문 | 확정 |
|---|---|
| 기본 탭 전환 시점 | 처음 그릴 때 최소화가 있으면 터미널 탭 + **이 탭에서 최소화를 누르면 루트 타일을 터미널 탭으로 전환**(타일이 어디로 갔는지 보이게). 다른 탭의 최소화로는 안 바꿈 |
| 종료하고 닫기 제한 시간 | 10초. 넘으면 "종료되지 않아 타일을 남겼습니다" 알림 |
| 크기 우선권 | 최소화로 xterm이 떨어져도 우선권·PTY 크기는 그대로(실시간-0에서 안 건드림) |
| 남이 실행한 것 안내 줄 | 실행자 정보가 응답에 없어 실시간-0에서는 뺌 → 실시간-4(SD2)에서 응답에 실행 계정 추가 때 넣음 |

## 구현 메모 (2026-10-07)
| 자리 | 한 것 |
|---|---|
| `tile.go` | `tileItem.Minimized`(omitempty). 폴더에 있으면 `validate` 400 |
| `processTerm.store.ts` | `STATUS_PILL`(상태 칩 표)을 여기로 옮김 — 머리 줄·목록 행이 같이 씀 |
| `tileGrids.store.ts` | `flat`(전체) / `order`(최소화 제외, 배치·`posOf`) / `instanceOf`는 `flat` 기준 / `terminals`(최소화 먼저) / `rootTab`(보드 load(`ready`)마다 최소화 있으면 `terminals`) |
| `terminalKill.hook.ts` | `killing`(kill 중 uid)은 모듈에 둠 — 머리 줄·목록이 같은 상태를 봄. `confirmKill(tileId, name, choices)`: X·목록 = 종료 + 종료하고 닫기, `RUN▾` 메뉴 "종료…" = 종료만, "종료하고 닫기…" = 그것만. 종료하고 닫기 = kill → `procs[uid].status` watch → 끝나면 닫기, 10초 넘으면 타일 남기고 알림 |
| `TileFrame.vue` | 루트 = `[크기]`만(X 없음). 크기 버튼 = 아이콘(`mdi-resize`), 보이는 크기가 다르면 아이콘 경고색 + 창 맨 위에 "크기 1×.5 (이 화면에서는 1×1)". 루트 탭 = 파일 본문은 `v-show`(목록·스크롤 유지), 터미널 목록은 `v-if`. 탭 라벨의 "· 최소화 n"은 n이 0이면 숨김 |
| `RootTerminalList.vue` + `TerminalListRow.vue` | 행 = 위치 번호(최소화면 점선 아이콘) · 경로 / 이름 (n) · 상태 칩 · 보기/열기 · 실행 중 = stop(확인창) / 끝남 = X(복원 전 막힘) |

## 확인 결과 (2026-10-07, 격리 실행 + 헤드리스 24/24)
- 루트 탭 2개·루트 머리 줄 `[크기]`만 / 터미널 머리 줄 창 버튼 3개
- 최소화 → 배치에서 빠짐 + 루트 터미널 탭으로 전환 + 행 "열기" / 서버 트리에 `minimized: true` 저장 / 새로고침 → 터미널 탭 기본·여전히 배치에 없음 / 열기 → 배치로 돌아옴
- 실행 중 X → 취소·종료·종료하고 닫기 / 취소 → RUN 그대로 / `RUN▾` 종료… → 취소·종료만 → 종료 → 타일 남고 DONE/FAIL / 끝난 X → 확인 없이 닫힘 / 종료하고 닫기 → 타일까지 닫힘
- 목록 stop → 종료 → 행 DONE/FAIL → 행 X → 사라짐 / 폴더 `minimized` PUT → 400
- **주의**: 사용자가 띄운 supervisor(5050)는 옛 바이너리면 `minimized`를 저장 때 버림 → 다시 빌드해야 최소화가 유지됨
- 확인 스크립트 `pw/rt0.mjs`(scratchpad, 재부팅 시 사라짐) — 다시 쓸 땐 위 항목 + `REF-test-env.md` "격리 실행"

## 실시간-6a(스크립트 행 "실행 중 N · 최소화 n") 구현 메모 (2026-10-08)
| 자리 | 한 것 |
|---|---|
| `query/shares.sql` `ListVisibleLiveProcessesInFolder` | 폴더(NULL = 루트)의 직접 자식 스크립트에서 실행 중인 process + 내 단계(`process_permission`) + 실행한 계정 닉네임 |
| `router/nodeRuns.go`(새) | `GET /nodes/:id/runs`(id 0 = 루트 목록, 폴더는 보기(10) 이상). 내 모든 보드 트리(`ListTileTreeDocs`, 잠금 없음)에서 uid -> (hash, tileId, minimized)를 매칭. 트리에 없으면 위치 필드 생략 |
| `node.api.ts` | `RunResponse`·`listRuns`·`useNodeRuns`·`invalidateNodeRuns` |
| `hook/scriptRuns.hook.ts`(새) | `useScriptRuns(folderId)`: 소켓 `TILES:UPDATE`·`STATUS`·`SHARE:CHANGE` 때 쿼리 무효화, 노드 id별 묶음 |
| `hook/terminalKill.hook.ts` | `confirmKillUid(uid, name)` 추가(타일 없이 종료만) — `openKillDialog`로 공통화 |
| `tile/ScriptRunList.vue`(새) | 행 = 위치(현재 보드면 번호·최소화 표시, 다른 보드면 보드 이름, 없으면 "위치 없음") · 실행한 사람 · 상태 · [보기/열기 \| 보드로 이동] · 종료(조작 20 이상). 현재 보드는 로컬 타일이 기준 |
| `tile/FolderTileBody.vue` | 스크립트 행에 "실행 중 N · 최소화 n" 칩(눌러 펼침) + 펼친 `ScriptRunList` |
| 정한 것 | 루트 스크립트는 장비가 없어 실행 불가 → 루트 runs는 항상 빈 목록(API는 지원) / 다른 보드 터미널의 열기 = 보드로 이동(최소화 해제는 그 보드에서) / 다른 보드·위치 없음의 종료 = 종료만(타일 안 닫음) |

## 실시간-6a 확인 결과 (2026-10-08, 격리 실행 임시 DB + 5051 + worker + 격리 프론트 3001)
- REST 11/11: 실행 전 빈 목록 / A·B가 같은 스크립트 실행 → A(40)·B(20)·C(10)가 2개 모두 보고 단계·위치(hash·tileId)·startedBy 닉네임 맞음 / 공유 안 된 D = 404 / 루트 runs = 200 빈 목록 / A 트리에서 한 타일 최소화 → 그 uid만 `minimized` / 한 개 kill 뒤 나머지만 남음
- 헤드리스 화면(B 계정): 공유 폴더 안 스크립트 행에 "실행 중 2" 칩(초록) → 펼치면 행 2개(번호·`이 보드`·실행한 사람·`RUN`·`보기`·종료 버튼). 첫 화면에서 칩 글자색이 `.FolderTileBody button` 규칙에 눌려 `button.runs`로 우선순위를 올림. 다른 보드·종료 확인창은 화면에서 확인하지 않음
- 스크립트(scratchpad `rt3/rt6_test.py`·`rt3/ui_setup.py`·`pw/ui6a.mjs`·`pw/vite.iso.mts`, 재부팅 시 사라짐). 함정: vite 설정이 scratchpad에 있으면 `vite`를 못 찾아 절대 경로 import 필요

## 실시간-6b(공유 반영) 구현 메모 (2026-10-08) — SD6(실효 단계 재조회)·SD9(경로 줄 멈춤)·공유 사라짐 처리
| 단위 | 한 것 |
|---|---|
| 6b-1 신호·status | `api.util.ts` `throwCatch`가 `status`(HTTP 상태)를 같이 던짐(기존 `message` 유지 — 상태 코드가 버려지던 "잔여"를 풀 수 있게 됨, 401 처리는 아직 안 고침) / `hook/shareChange.hook.ts` `useShareChange`: `SHARE:CHANGE`에서 `['NODE']` 쿼리 전부 무효화 / `pages/index.vue`에서 한 번 호출 |
| 6b-2 경로 멈춤·공유 루트 표시 | `NodeResponse`에 `permission`·`ownerNickname`·`parentVisible` / `fetchNodePath`가 `parentVisible === false`면 거기서 멈춤 / 루트 목록의 공유 폴더 행에 `← 소유자 닉네임` |
| 6b-3 터미널 볼 수 없음 | `processTerm`에 `gone`·`markGone`(xterm 비움)·`clearGone`·`pruneGone` / `terminalRestore.hook.ts`: `SHARE:CHANGE`에서 트리 uid 전부 재조회(`recheck`) → 응답에 없으면 `markGone`, 다시 보이면 `clearGone` 뒤 기존 `sync`가 스냅샷으로 새로 등록. 처음 복원·`sync`에서 못 찾은 uid도 `markGone` |
| 6b-4 폴더 볼 수 없음·배너 | `useGetNode` 404는 재시도 안 함 / `TileFrame.vue` `unavailable`(폴더 = 노드 404, 터미널 = `gone`) → 배너 "공유가 취소되었거나 삭제되어 볼 수 없습니다. 닫기만 할 수 있습니다." + 본문 없음 + 머리 줄은 닫기(X)만 |
| 정한 것 | 알림 = 타일 안 배너만(토스트 없음) / 보드 단위 처리 안 함 / 조작 → 보기 단계 낮춤의 "보기 전용" 표시는 공유-e(프론트) / 판정은 HTTP 404만(서버 오류는 기존 오류 표시) |

## 실시간-6b 확인 결과 (2026-10-08, 격리 실행 + 격리 프론트 3001 + 헤드리스, 계정 A(소유자)·B(받는 쪽))
- B 화면(폴더 타일 + 터미널 타일 + 루트): 처음 revoked 0·xterm 1 → A가 공유 풀기 → **revoked 2·xterm 0**(폴더·터미널 모두 배너, 두 타일 머리 줄 버튼 = 닫기 각 1개) → 다시 공유 걸기 → **revoked 0·xterm 1(복원)** → A가 폴더 삭제 → revoked 2·xterm 0. 페이지 오류 없음
- 루트 목록: 공유 폴더 행 `REV_DIR ← rv-a`(소유자 칩), 공유 폴더 타일 경로 줄 `기본 / REV_DIR` — 404 오류 없음, 콘솔 오류 없음
- 확인 안 한 것: 범위 밖 이동 화면, 조작 → 보기 단계 낮춤 표시(공유-e), 모바일
- 스크립트(scratchpad `pw/ui7.mjs`·`pw/ui8.mjs`·`rt3/ui7_setup.py`, 재부팅 시 사라짐)

## 루트 타일 막힘 수정 (2026-10-08, 실시간-6b 보완, 커밋 전)
- 발견(사용자, test1 계정): 루트 타일을 공유받은 폴더 안으로 열면(`navigate` — 루트 타일 `nodeId`가 그 폴더로 바뀌어 서버에 저장) 공유가 풀린 뒤 "볼 수 없음" 배너만 뜨는데, 배너 상태는 닫기만 허용하고 루트 타일엔 닫기가 없어 빠져나올 수 없음. 실시간-6b 확인은 "새 타일로 연 경우"만 봐서 놓침
- 수정(안 A): `TileFrame.vue` 배너에서 **루트 폴더 타일일 때만** "루트 타일은 닫을 수 없어 기본 목록으로 돌아갑니다." + "기본으로 돌아가기"(`navigate(tileId, null)`). 다른 타일은 그대로 "닫기만". 서버·데이터 변경 없음. 사용자 화면 확인 통과
- 안 B(서버 `reconcileAccess`가 루트 `nodeId`를 되돌림)는 타일 트리 쓰기 경합을 건드려서 안 함
