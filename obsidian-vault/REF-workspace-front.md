# REF — 보드(작업 공간 여러 개): 프론트 구현 (작업공간-d)

> 2026-10-07 신설. 화면 이름 = **"보드"**(코드 이름은 `workspace`·`tile_trees`, 2026-10-07 사용자 확정). 개념·확정 사항 → `REF-workspace.md` / 구조·결정(WS1(트리 주소)~WS7(보는 작업 공간 전달))·API → `REF-workspace-impl.md` / 서버 구현 → `REF-workspace-server.md` / 이력 → `history/workspace-front.md` / 현재 진행 → `CURRENT.md`.

**상태(2026-10-07): 작업공간-d 코드 작성 + type-check 통과 + 사용자가 실제 화면에서 확인("잘 동작하네") + 작업공간-e(확인) 헤드리스 34개 통과(아래 "확인 결과"). 커밋 전(파트 마감 때).**

## 파일별 변경
| 파일 | 내용 |
|---|---|
| `router/index.ts` | `/`·`/w/:hash` 둘 다 `pages/index.vue`, `meta.board = true` |
| `pages/index.vue` | 라우트 hash(빈 값 = 기본) 지켜봄 → `tileTree.load(hash)`. 계정 바뀜만 `resetTerms`(보드 전환은 xterm 유지). `provideBoards`. `TAB:VIEW {hash}` emit = `[tabReady, hash]` 변할 때(소켓 재연결 뒤 TAB:ID마다 다시). 보던 보드가 목록에서 **있다가 없어지면** `/`로 + 안내(이 탭이 지운 것이면 안내 없음). 트리 404 = 오류 문구 + "기본 보드로" 버튼. `BoardMenu`를 헤더 `#app-head-title`에 Teleport |
| `feature/node/api/tile.api.ts` | `getTiles(hash)`·`putTiles(hash, …)`(빈 값 = `/tiles`), 응답에 `hash` |
| `feature/node/api/board.api.ts` (새) | `listBoards`·`createBoard`·`moveBoardFolder`(409 = 결과)·`renameBoard`·`deleteBoard` |
| `feature/node/store/boards.store.ts` (새) | 목록 + `WORKSPACES:UPDATE` + 재연결 시 다시 불러옴. `boardLabel`(이름 비면 "기본 보드"/"이름 없는 보드")·`boardPath`. `add`(막 만든 보드를 이벤트 전에 넣음)·`rename`·`remove`(`removedHere` 기록) |
| `feature/node/store/tileTree.store.ts` | `hash`(불러온 보드의 실제 hash) 노출, `TILES:UPDATE`는 그 hash만(WS6(TILES 이벤트 구분)), 불러오는 중 push는 hash별 보관. **`adoptOrphans`·`adoptOp` 삭제**(서버 WS3(고아 안전망 위치)). `flush`(저장 대기분 먼저 — 5초·실패 시 중단) + `serverChange`(409면 rebase 뒤 재시도, `CONFLICT_MAX`) + `detachToBoard`·`moveToBoard` |
| `feature/node/hook/terminalRestore.hook.ts` | 복원 조회 = 트리에 있는 uid만(`lookupProcesses`). `listProcesses` + 고아 붙이기 제거 → 순서 의존 2) 주석 갱신 |
| `FolderTileBody.vue` · `process.api` · `processTerm.store` | exec에 `workspace`(지금 보드 hash) |
| `feature/node/component/BoardMenu.vue` (새) | 헤더: 지금 보드 이름 ▼ → 보드 목록(전환) / 이름 변경(`NodeNameDialog` 재사용) / 보드 삭제(기본 제외, 확인창) / "계정 {닉네임}" |
| `AppHead.vue` | 제목 자리에 `#app-head-title`, `meta.board`가 아닌 페이지만 닉네임 |
| `TileFrame.vue` | 폴더 타일(루트 제외) ⋮ 메뉴: "새 보드로 분리"(~~기본 보드에서만~~ 어느 보드에서든 2026-10-22, 이름 기본값 = 폴더 노드 이름) → 만든 보드로 이동 / "다른 보드로 이동" 목록. **터미널 머리 줄 = `[실행한 타일 번호] 경로 / 이름`**(번호 = 부모 타일의 위치 번호, 경로 = `useNodePath(스크립트)`에서 자신 제외, 좁으면 경로부터 줄어듦). 폴더 타일 머리 줄은 그대로(`← opener`) |

## 구현 중 정한 것 (REF-workspace "제안"을 따르거나 새로 정함)
| 항목 | 정한 것 |
|---|---|
| 헤더 | 보드 이름 + 전환 메뉴, 닉네임은 메뉴 안(REF-workspace "제안" 그대로) |
| 분리 뒤 | 만든 보드로 바로 이동 |
| 새 보드 기본 이름 | 폴더 노드 이름(입력창 초기값, 지우면 저장 불가 — `NodeNameDialog`가 빈 이름 거부) |
| 이름 없는 보드 표시 | 기본 = "기본 보드", 그 외 = "이름 없는 보드" |
| 다른 보드 터미널 | 보드를 바꿔도 xterm·등록은 남김(돌아오면 그대로). 새로 등록하는 것은 지금 보드 트리에 있는 uid만 |
| 서버 변경(같이 함) | `tileTreeResponse`에 `hash` 추가 + 사용자 문구 "작업 공간" → "보드" (`tile.go`·`workspace.go`) |

## 확인 결과 (작업공간-e, 2026-10-07)
격리 실행(임시 DB `nexus_wse_tmp` + supervisor 5051 + worker `wse-dev` + 프론트 3001, 확인 뒤 종료·삭제) + 헤드리스 탭 3개(같은 계정). 34개 통과, 두 번 돌려 같은 결과. 스크립트 = 세션 scratchpad `pw/board.mjs`(재부팅 시 사라짐, 격리 프론트 띄우는 법 → `REF-test-env.md` "격리 실행").

| 묶음 | 항목(전부 통과) |
|---|---|
| 머리 줄 | 폴더 타일 안에서 하위 폴더로 들어가 실행 → `[2] F1 / Sub / S1`(번호 = 실행한 폴더 타일 위치, 경로 = 스크립트의 상위 폴더들) / 분리 뒤 같은 터미널이 `[1]`로 바뀜(부모 타일 위치 따라감) |
| 분리 | ⋮ "새 보드로 분리" → `/w/{hash}` 이동 + 헤더 이름 / 새 보드 = 폴더 타일(루트) + 터미널 / API: 터미널·폴더가 새 보드에만 / 터미널 RUN 유지 / 새 보드의 루트 타일엔 ⋮ 없음 |
| URL | 새로고침 → 같은 보드 + 터미널 화면 내용 복원 / 새 탭에서 `/w/{hash}` 직접 접속 / 없는 hash → "보드를 찾을 수 없습니다" + "기본 보드로" → `/` |
| exec | 분리한 보드에서 실행 → 그 보드 트리(기본 보드엔 없음) / 기본 보드에서 실행 → 기본 트리, 다른 보드를 보는 탭에 안 나타남 |
| 탭 2개 다른 보드 | 기본 보드 변경이 다른 보드 탭에 안 보임(TILES 거르기) / 크기 우선권 = 그 보드를 보는 탭(양쪽 모두) |
| 이름 변경 | 이 탭 헤더 + 다른 탭 보드 목록에 반영(`WORKSPACES:UPDATE`) |
| 이동 | 기본 → 다른 보드: 출발에서 빠짐 + 도착을 보는 탭에 나타남(이름·RUN·입력 출력까지) / 옮겨진 터미널 크기 우선권 → 도착을 보는 탭 / 기본 아닌 보드의 ⋮ = "기본 보드"로 이동만("새 보드로 분리" 없음) |
| 삭제(자기 탭) | `/`로 + 안내 없음 / 안의 타일이 기본 보드 끝으로(같은 계정 다른 탭에도) / 목록 = 기본만 / 크기 우선권이 기본을 보는 탭 하나에 |
| 삭제(다른 탭) | 다른 탭이 만든 보드로 메뉴에서 전환 / 다른 탭이 삭제 → 보던 탭은 `/` + "보던 보드가 삭제되어…" 안내, 지운 탭은 안내 없음 |
| 전체 | 계정 전체 트리에서 터미널 uid 중복 없음 |

- 1차 실행에서 이동 직후 받은 탭 머리 줄이 잠깐 `node 9`·상태 `-`로 보였음 → 이름 조회·process 등록 전 표시(코드 의도, "이름을 받기 전에는 node id로 표시"). 다시 돌린 두 번은 처음부터 이름·RUN
- 확인 못 한 것: 화면 크기 변화에 따른 실제 PTY 크기(크기 우선권은 `GET /processes`의 `sizeOwner`로만 봄)
- 희망안(`/` 루트 안 다른 보드 미리보기)은 안 함
