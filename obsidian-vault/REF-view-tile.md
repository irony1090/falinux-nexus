# REF — 파트(보기 타일) 구현

> 2026-10-28 신설(REF-file-view.md가 10k에 닿아 구현 기록을 분리). 구상·결정(V1~V5, 사본 안 1, 데이터 흐름) = `REF-file-view.md` / 이력 → `history/view-tile.md` / 진행 → `CURRENT.md`
> 브랜치 `feat/file-view`(미커밋). 단계: 1단계(사본 없는 뷰어 + D(더블클릭 열기)) 사용자 확인 끝 → 2단계(이미지 사본 + 터미널 탭 보기 줄 + 편집 줄 이름) 구현·격리 확인 끝(사용자도 3001에서 조작해 봄) → **2026-10-28 파트 마감, 커밋 + master 머지**(push = 사용자)

## 결정 (사용자 2026-10-28)
| 이름 | 내용 |
|---|---|
| 단계 나누기 | 1단계 = 사본 없는 뷰어 + D(더블클릭 열기) / 2단계 = 이미지 사본(안 1(타일 존재 기준)) |
| ②(바이너리 판정) | 더블클릭 때 그 파일만 앞 8KB NUL 검사(probe). 바이너리 = 안내 창("칸 메뉴의 편집을 쓰세요"), 칸 메뉴 "편집"은 그대로 |
| D(더블클릭 열기) | 폴더 = 들어가기 · 실행 파일 = 실행(기존, 이미지보다 먼저) · 이미지 = 보기 타일 · 영상 = 아무 일 없음 · 그 밖 파일 = probe 후 텍스트면 편집기 |
| 같은 파일 다시 | 이 장비 타일이 이미 연 같은 파일(instance+path) 보기 타일이 있으면 그 타일로(activate) |
| 영상은 한 타일(10-28) | 이미지 = 같은 파일을 여러 타일에 열 수 있음 / **영상 = 계정 전체에서 한 타일**(이미 열려 있으면 그 보드·타일로 이동) — 파트(영상 Range)에서 구현 |
| process 범위 | 보기 타일은 process 아님(MEMORY "process 범위"). 터미널 탭엔 트리 조회로 보기 줄을 붙임 |
| 링크 | 이름에 확장자가 없으면 원본 경로(`targetPath`)로 엶 — 서버가 경로 확장자로 형식을 고르므로 |

## 1단계 파일 (1단계 구현)
| 층 | 파일 | 내용 |
|---|---|---|
| supervisor | `router/workerFile.go`(새) | `GET .../file?folderId&path`: `viewTypes`(V1(확장자)) 아니면 415, `requireDeviceDir`, `openWorkerFile`로 조각 중계. `ETag` = `"크기hex-수정시각hex"`, `If-None-Match` 같으면 304, `Cache-Control: private, no-cache`, `inline`, V2(SVG 보안) `CSP: sandbox` + `nosniff`, 크기 다르면 `abortConn` / `GET .../file/probe` = `{binary}`(앞 `probeSize` 8KB) |
| supervisor | `workerRead.go` | `ModTime()` 추가 |
| supervisor | `workerApi.go` · `tileKind.go` · `tileDoc.go` | 경로 2개 등록 · `viewer` 줄(루트·자식 불가, nodeId·instance·path 필수, uid 없음) · 필드 주석 |
| 프론트 | `tile.type.ts` · `tileKind.util.ts` · `tileComponents.ts` | `viewer`(아이콘 `mdi-image-outline`, 색 `success`) |
| 프론트 | `tileTree.store.ts` | `ViewerTile` 타입 · `openViewer`(장비 타일의 자식) |
| 프론트 | `nodeOps.store.ts` | `viewFile`(같은 파일 재사용 + activate) |
| 프론트 | `ViewerTile.vue` · `ViewerTileBody.vue`(새) | 틀 = 이름(파일명)·인스턴스 라벨, 연 폴더 404 = 볼 수 없음 / 본문 = `<img>` + V4(확대 조작: 휠(커서 기준)·끌기·핀치·더블클릭/두 번 탭 맞춤) + 오른쪽 아래 `[배율 | 맞춤 | 1:1]`. 맞춤 = 작은 이미지는 키우지 않음 |
| 프론트 | `worker.api.ts` | `viewFileUrl`(`BaseAxios.defaults.baseURL` + 문자열) · `probeWorkerFile` · `instanceOnline`(장비 타일 online 계산을 옮김) |
| 프론트 | `fileKind.util.ts`(새, `feature/worker/util`) | V1(확장자) 목록 한 곳 — 서버 `viewTypes`와 같게 |
| 프론트 | `DeviceTileBody.vue` | 칸 메뉴 "보기"(이미지만) · 더블클릭 D(더블클릭 열기) |
| 프론트 | `OverviewTile.vue` · `nodeRemove.hook.ts` | 전체보기 이름 = 파일명 · 폴더 삭제 때 보기 타일도 닫음 |

- 상태 표시: 장비 끊김 = "장비가 접속해 있지 않습니다"(명부 online), 다시 접속 = `<img>` 다시 만듦 / `<img>` 실패 = probe로 서버 메시지(파일 없음 등), 성공이면 "표시할 수 없는 이미지" + [다시]
- 확대 상태 저장 = `userPrefs` 키 `viewer.views`(타일 id -> {s,x,y}, 맞춤이면 지움, 최근 50개), 로그아웃 때 같이 지워짐
- 닫기 규칙: `closeOp`이 장비 타일을 닫으면 보기 타일은 그 부모(폴더 타일)로 올라감(자식 불가 타일이라 kids 없음)

## 격리 확인 (2026-10-28)
- 환경: supervisor 5051(`IP=127.0.0.1`) + DB `nexus_fv_tmp` + worker `fv-dev`(TMPDIR `/run/user/1000/fviso`) + vite 3001, 계정 `fv1`/`pw123456`, 폴더 id 1, 픽스처 = scratchpad `iso/fx`(big.png 1600x1000·small.PNG·x.svg(스크립트 포함)·note.txt·blob.bin·clip.mp4·broken.png·linkimg(확장자 없는 링크 -> big.png)·empty.txt)
- API: 200 + 바이트 동일 · 304 · 대문자 확장자 · SVG 헤더 · txt 415 · 없음 404 · 상대 경로 400 · 폴더 400 · path 없음 400 · 남의 폴더 404 · 세션 없음 401 · probe(txt false / bin·png true / 빈 파일 false / 없음 404) · path 없는 보기 타일 PUT 400
- 화면(헤드리스 1280 + Pixel 7): 더블클릭 이미지 -> 보기 타일 · 새로고침 뒤 타일·배율 유지 · 같은 파일 재사용 · 링크 칸 메뉴 "보기" = 원본 타일 재사용 · txt 메뉴에 보기 없음 · SVG 스크립트 안 돎 · 깨진 png 안내 · mp4 아무 일 없음 · 바이너리 안내 · txt -> "편집 note.txt" 터미널 · 휠 확대(커서 아래 점 고정) · 끌기 · 1:1 · 맞춤 · 핀치 37% -> 110% · 두 번 탭 맞춤 · 폰 더블탭 열기
- 콘솔: `Failed to send error to Vite server`(격리 vite HMR 소켓이 443으로 가는 설정 탓, 페이지 로드 때부터) 말고 오류 없음
- 테스트 함정: 타일을 연 직후 보드가 reveal 스크롤 중이라 boundingBox가 옛 위치 -> 1초 기다린 뒤 잼

## 2단계 (이미지 사본 · 터미널 탭 보기 줄 · 편집 줄 이름) — 2026-10-28
결정(사용자): 정리 = **주기**(시작 + 10분) — REF-file-view의 "트리 저장(한 곳)"은 실제로 트리를 쓰는 곳이 5곳(`tile.go` PUT·`tilePlace.go` 2·`processPurge.go`·`workspace.go`)이라 폐기 / 오프라인 = 마지막 사본 + 표시 / 탭 보기 줄 = process로 만들지 않고 트리 조회 / 편집 줄 이름도 같이

| 층 | 파일 | 내용 |
|---|---|---|
| supervisor | `router/viewCache.go`(새) | `<VIEW_CACHE_DIR>/<userId>/<sha256(instance\0path)>` + `.json`(크기·수정 시각), 받는 중 `.tmp-<임의>`. commit 순서 = 옛 메타 삭제 -> 본문 rename -> 새 메타(사이에 읽으면 사본 없음). V3(사본 상한) 20MB. 표시 파일 `.nexus-db` = `host:port/db`, 다르면 경고 + 끔. `sweep` = `ListViewerTiles`에 없는 사본·남은 임시 파일 삭제(1분 안 것은 남김), 빈 사용자 폴더 삭제 |
| supervisor | `workerFile.go` | `fileQuery` = DB 인가만(`requireDeviceFolder`) + 접속 여부. 끊김 = 사본(없으면 409) / 접속 = `FILE_OPEN` 크기·수정 시각이 사본과 같으면 사본, 다르면 worker에서 받으며 사본 씀. probe는 끊기면 409 |
| supervisor | `constants/env.go`·`dev.env`·embed `env`(git 제외)·`.gitignore`·`supervisorRouter.go` | `VIEW_CACHE_DIR`(키 없음 = `view-cache`, 빈 값 = 끔, 상대 = WEB_DIR과 같음) · 개발 기본 폴더 `cmd/supervisor/view-cache/` git 제외 · `r.view` + `go runViewSweep()` |
| supervisor | `tileTrees.sql` | `ListViewerTiles`(전 계정 보기 타일 user·instance·path, `jsonb_each`) |
| supervisor | `viewerList.go`(새) · `tile.go` | `GET /tiles/viewers?folderId=[&mainKey=]` — 내 모든 보드 트리 너비 우선, 범위 = 터미널 탭과 같음(0 전부 / 폴더 = `IsDescendant` / 장비 = main_key) |
| supervisor | `shares.sql` 실행 목록 3개 · `nodeRuns.go` · `tileNotice.go` | `p.args` 추가(세 행 타입이 같아야 변환됨) · `execName(cmd, args)` = "편집 <파일>" (procName과 공용) |
| 프론트 | `ViewerTileBody.vue` | 끊겨도 요청, 받아지면 왼쪽 위 "오프라인 · 마지막 사본", 실패 + 끊김 = 끊김 안내(probe 안 함) |
| 프론트 | `tile.api.ts` `listViewers` · `scriptRuns.hook.ts` `useRootViewers` | 키에 이 보드 보기 타일 id 목록 포함 + `TILES:UPDATE`로 무효화 |
| 프론트 | `ViewerRow.vue`(새) · `RootTerminalList.vue` · `RootTabs.vue` | 터미널 줄 뒤에 보기 줄([위치 번호] 경로 / 파일 · 보드 [보기 \| 보드로 이동] [×]), 탭 `터미널 N · 보기 n · 최소화 n` |

### 2단계 격리 확인
- 사본: 첫 요청에 생성 · `touch` 뒤 메타 갱신 · 25MB = 사본 없음 · worker 끔 -> 사본으로 200(바이트 동일)·304·사본 없는 파일 409·probe 409 · 시작 정리 = 타일 없는 사본(2시간 전) + 임시 파일 삭제, 타일 있는 것은 남김 · 표시 파일 다른 DB = 경고 + 끔, 되돌리면 다시 켬
- 목록: `/tiles/viewers` 전부·폴더·장비(남의 장비 403) · 실행 목록 이름 "편집 note.txt"(장비·deep)
- 화면: 끊김 = 배지 3개 + 깨진 png는 끊김 안내 / 다시 접속 = 배지 없음 · 터미널 탭 `터미널 2 · 보기 4` + 보기 줄 4개
- 확인 중 보기 타일 3개가 트리에서 사라진 일: 재현 안 됨(저장·재시작 모두 유지). 재시작 직후 headless 아닌 클라이언트의 `GET /api/tiles`가 보여 **사용자 브라우저(3001)에서 닫은 것으로 추정** -> 사용자 확인: 직접 조작·닫음(버그 아님)
