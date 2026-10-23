# REF — worker 폴더 타일: 폴더-6(보드 이동) + Q9(루트 장비 탭)

> → `history/worker-dir-board.md` / 앞 단계 → `REF-worker-dir-exec.md`·`REF-worker-dir-list.md` / 상위 계획 → `REF-worker-dir.md` / → `CURRENT.md`

## 근거 (사용자 결정, 2026-10-22)
- 장비-3(보드 이동): 장비 타일도 새 보드로 분리·다른 보드로 이동 -> 거기서 루트 장비 타일
- 장비-1(루트 장비 타일): 루트면 탭 `[파일] [터미널 N]`, 기능은 폴더 루트와 같음
- Q9(장치파일 process 목록): 루트 장비 타일의 터미널 탭 = 그 장비에서 띄운 내 process 전부
- Q11(옮겨 간 뒤 소속): 옮겨도 연 폴더(nodeId) 연결 유지

## 조사 (2026-10-23)
- 서버 `createWorkspace`·`moveWorkspaceFolder`는 `detachFolder`가 능력-1(루트 가능)로 검사 -> device는 이미 `canRoot: true`라 **서버 변경 없음**(요청 필드 이름만 `folderTileId`)
- 프론트 `useTileBoard`·`detachToBoard`·`moveToBoard`도 타일 id만 받음 -> 장비 타일 메뉴에 붙이면 됨
- 루트를 폴더로 가정한 곳: `RootTerminalList`(루트 nodeId로 조회) · `FolderTile`의 루트 탭·`#unavailable`(기본으로 돌아가기) · `nodeRemove.hook`(지운 폴더를 보던 루트 = `navigate(null)`, 장비 타일에 쓰면 깨짐)
- 루트 타일은 X가 없음(`TileFrame` `isRoot`) — 장비 루트도 닫기 대신 보드 삭제(트리째 기본 보드로) 또는 이동

## 설계안 (2026-10-23 제시·확정, 구현 끝)
| 번호(이름) | 내용 | 손댈 곳 |
|---|---|---|
| 이동-1(메뉴) | 장비 타일 메뉴(루트 아닐 때) = "새 보드로 분리" + "다른 보드로 이동" — 폴더 타일과 같은 `useTileBoard`, 자식 터미널 타일도 같이 감 | `DeviceTile.vue` |
| 이동-2(루트 탭) | 루트 장비 타일 = 탭 `[파일] [터미널 N · 최소화 n]`, `rootTab` 공용 | `DeviceTile.vue`, `RootTerminalList.vue`(루트 종류별 조회) |
| 이동-3(장비 목록 API) | `GET /workers/:mainKey/runs?folderId=` — 인가 = 연 폴더 소유자 + device_key == mainKey(인스턴스 접속 여부는 안 봄). 응답 = 기존 runs 행 | 새 쿼리 `ListVisibleRunsOnDevice`(`device_key LIKE main || '#%'`, 살아 있음 또는 내 타일 있음, `process_permission > 0`) |
| 이동-4(루트 경계) | 연 폴더가 지워졌을 때 루트 장비 타일 처리(Q26) · 폴더 루트 전용 문구가 장비 루트에 안 나오게 | `nodeRemove.hook`, `DeviceTile.vue` |
| 이동-5(확인) | 헤드리스: 분리 -> 새 보드 루트 = 장비 타일·탭 목록·자식 터미널 같이 감 / 다른 보드로 이동 / 새로고침 / 연 폴더 삭제 | scratchpad |

## 질문 (2026-10-23 답)
- Q25(루트 장비 탭 범위) = **가**: 그 장비(main_key, 인스턴스 전부)에서 돈 process 중 내가 볼 수 있는 것 전부 — 다른 타일·다른 폴더의 스크립트 노드로 띄운 것, 공유받은 것 포함. 끝난 process는 내 타일이 남은 것만(이동-3(장비 목록 API) 조건). (기각: 나 = 장비 디렉토리에서 실행한 것만 / 다 = 그 인스턴스만)
- Q26(연 폴더 삭제 때 루트 장비 타일) = **가**: 그대로 두고 "볼 수 없음" 표시(보드 삭제·이동으로 치움). (기각: 나 = 기본 폴더 목록 타일로 바꿈)
- 2026-10-23 "진행" -> 구현(아래)

## 구현 (2026-10-23)
| 번호(이름) | 한 것 |
|---|---|
| 이동-1(메뉴) | `DeviceTile.vue` 메뉴 = 폴더 타일과 같은 `useTileBoard`(루트·볼 수 없음이면 메뉴 없음). 분리 다이얼로그 기본 이름 = 장비 타일 이름 |
| 이동-2(루트 탭) | 탭 마크업을 **`RootTabs.vue`로 빼서** 폴더·장비 루트가 같이 씀. 목록 조회 = `scriptRuns.hook.ts` **`useRootRuns`**(루트 종류별: 폴더 = `/nodes/:id/runs?deep=1`, 장비 = `/workers/:main/runs`) — `RootTabs`(숫자)·`RootTerminalList`(목록)가 같은 쿼리. 소켓 무효화는 `useRunsRefresh`로 빼서 `useScriptRuns`·`useRootRuns`가 각자 등록(전엔 탭 숫자 갱신이 폴더 본문의 `useScriptRuns`에 기대고 있었음) |
| 이동-3(장비 목록 API) | `GET /workers/:mainKey/runs?folderId=`(`nodeRuns.go` `listDeviceRuns`). 인가 = `requireDeviceFolder`(`workerDir.go`, `requireDeviceDir`에서 접속 확인만 뺀 앞부분). 쿼리 `ListVisibleRunsOnDevice`(`shares.sql`) — main_key 비교 = `regexp_replace(device_key, '#[^#]*$', '')`(서브키 = 마지막 `#` 뒤, `LIKE`는 `_`·`%` 때문에 안 씀). `listNodeRuns`의 타일 위치 수집·응답 변환을 `runLocations`·`toRunResponses`로 빼서 같이 씀. 프론트 `listDeviceRuns`·`useDeviceRuns`(쿼리 키 = RUNS_KEY 아래라 `invalidateNodeRuns`가 같이 무효화) |
| 이동-4(루트 경계) | 연 폴더 404 = `unavailable`(루트·비루트 모두). 루트 문구 = "보드를 삭제하면 기본 보드로 옮겨져 닫을 수 있음". `nodeRemove.hook` = 루트가 장비 타일이면 그대로 둠(폴더 루트만 `navigate(null)`) |

## 확인 (2026-10-23, 격리 5051 + `nexus_mv_tmp` + worker `mv-dev` + vite 3001)
- REST 20/20(scratchpad `iso/mv_test.py`, 세션 끝나면 사라짐): 장비 실행 + 다른 폴더 스크립트 + 끝난 것(타일 있음) 포함 · 행 필드(nodeId null·이름·cwd / 스크립트 이름·나 / 타일 위치) · 실행 순서 · 같은 장비 다른 폴더로 조회 = 같은 목록 · 장비 없는 폴더·다른 장비 폴더 403 · 다른 장비 = 빈 목록 · folderId 없음 400 · 다른 계정 404 · 회귀(폴더 터미널 탭 deep) · B = 공유받은 것만(단계 10). 추가: worker 끊긴 뒤에도 200
- 헤드리스 24/24(scratchpad `pw/mv.mjs`, playwright-core 1.64): REST로 기본 보드에 장비 타일 2개(하나는 터미널 자식) 넣고 -> 비루트엔 탭 없음 · 메뉴 "새 보드로 분리" -> 새 보드 루트 = 장비 타일(닫기·메뉴 없음, 탭 `터미널 3`, 자식 터미널 같이 옴, 목록 long.sh·s1·q1) · 서버 트리 루트 device · 새로고침 · 다른 장비 타일을 "다른 보드로 이동" -> 루트 장비의 자식 · 연 폴더 삭제(API) -> 볼 수 없음 + 루트 문구 + 탭 없음, 새로고침 뒤에도 · 보드 삭제 -> 기본 보드에서 비루트·닫기 있음. 회귀: 폴더 루트 탭 숫자·목록
- `go build`·`go test`(router)·`go vet`·`vue-tsc` 통과
- 함정: 격리 worker를 `pgrep -x worker` + cwd로 고르면 **보관 프로세스(`worker -hold`)도 cwd가 같아 같이 잡힘**(이번엔 시험용이라 무해). 정리 = 실행 사본 `/tmp/<nodeId>/<uid>`를 uid로 골라 지움
- **사용자 브라우저 확인 OK(2026-10-23, test3 `build/worker-dir`)**: Linux 장비 + Windows VM 장비(`win-dev`) 둘 다 — 사용자 "매우 만족"
- **못 본 것**: 화면에서 폴더를 지울 때(`nodeRemove.hook`)의 루트 장비 분기(장비 루트 보드 안에서 상위 폴더 타일로 지우는 경우), 모바일, 사용자 브라우저
