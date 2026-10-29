# REF — 파트(영상 Range) 구현

> 2026-10-29 신설(REF-view-tile.md에서 분리). 구상·데이터 흐름 = `REF-file-view.md` "데이터 흐름" / 앞 파트 = `REF-view-tile.md` / 이력 → `history/video-view.md` / 진행 → `CURRENT.md`
> **2026-10-29 파트 마감** — 브랜치 `feat/video-view` 커밋 + master 머지(push = 사용자). 1단계(영상 재생) + 2단계(S1(이어보기)), test3 임시 빌드에서 사용자 확인

## 결정 (2026-10-29, 착수 전 조율)
| 이름 | 사용자 답 |
|---|---|
| Q1(한 타일의 뜻) | **같은 영상 파일(instance+path)은 계정 전체에서 타일 하나**(10-29 재확인). 다른 영상은 새 타일. 다시 열면 그 보드·타일로 이동 + 포커스(active) |
| Q2(타일 위치) | 이미지 보기와 같은 규칙(연 장비 타일의 자식). 다른 점 = 같은 파일 찾는 범위(이미지 = 그 장비 타일 / 영상 = 내 모든 보드) |
| Q3(오프라인) | 끊김 = 안내, 다시 접속 = 다시 불러오고 **보던 자리 유지** |
| Q4(같은 계정 여러 창) | **S1(이어보기)**: 재생 위치를 서버에 저장, 다른 창·기기에서 열거나 새로고침하면 그 자리부터. 열린 창끼리 실시간으로 따라가지는 않음(S2(실시간 따라가기)는 사용자가 원했다가 S1으로 충분하다고 접음). "재생 위치 = 브라우저에만"(REF-file-view 합의 표) 바뀜 |
| P1(위치 지우는 시점) | 타일을 닫으면 재생 위치도 지움(사용자 10-29) = sweep이 보기 타일 없는 파일의 `video_positions` 행 삭제(사본 안 1(타일 존재 기준)과 같은 기준) |
| P2(한 타일 보장 위치) | **C(서버가 넣음)**(10-29). 정정: 보기 타일을 트리에 **새로 넣는 곳은 PUT `/tiles` 하나**(프론트가 만들어 트리째 저장). 나머지 4곳은 터미널 넣기·지우기·보드 간 옮기기라 중복을 안 만듦. 안 = A(프론트만) / B(PUT에서 거절) / C(서버가 넣음, 터미널 WS5(터미널 타일 추가 주체)와 같은 방식) |
| 타일 추가 주체(10-29) | 사용자: "외부 자원에 묶이는 것만 서버에서". 전체 서버 관리는 안 함(배치 = 프론트 PUT 그대로). 범위 = **(가) 하나만 있어야 하는 것(터미널·영상)**, 이미지·장비·폴더는 프론트(사용자 "네 의견에 따르자") → MEMORY·`REF-decisions.md` "타일 추가 주체" |

## 1단계(영상 재생) 파일 (2026-10-29)
| 층 | 파일 | 내용 |
|---|---|---|
| supervisor | `workerFile.go` | `isVideoPath`(viewTypes가 `video/`) · `setViewHeaders`(공통 헤더, 이미지 `viewHeaders`가 304·길이를 더함) · `serveVideo` = 끊김 409(사본 없음) / 접속 = `openWorkerFile` + `http.ServeContent`(Range·If-Range·304·416·여러 구간 전부) · `readSeeker`(workerReader -> io.ReadSeeker, Seek = 위치만 옮김, Read = 256KB 조각 받아 나눠 줌) |
| supervisor | `videoTile.go`(새) · `tile.go` · `tileKind.go` | `POST /tiles/viewers/video` {folderId, instance, path, workspace, parentTileId, size} -> `requireDeviceFolder` -> `lockWorkspaces`(계정 트리 전부) -> 같은 instance+path 보기 타일 있으면 {hash, tileId}, 없으면 요청 보드(삭제됐으면 기본)의 부모(닫혔으면 루트) 아래 `newViewerTile` + `saveLockedTree` -> {hash, tileId, tile, tileVersion} + 커밋 뒤 `publishTiles` |
| supervisor | `tileDoc.go` `checkVideoTiles` + `putTiles` | PUT의 영상 보기 타일은 저장본에 같은 id·instance·path로 있던 것만 통과(새 추가·경로 바꾸기 400). 이미지 보기 타일은 그대로 프론트 |
| 프론트 | `tile.api.ts` `openVideoTile` · `tileTree.store.ts` | `ServerTile = TerminalTile \| ViewerTile` — `applyServerTile`·`upsertOp`가 영상 타일도 받음 |
| 프론트 | `nodeOps.store.ts` `viewFile` | 영상 = `flush` -> API -> 이 보드면 `applyServerTile`(새것)·`activate`, 다른 보드면 `router.push(focusPath)`(터미널 탭 "보드로 이동"과 같은 길) |
| 프론트 | `VideoViewBody.vue`(새) · `ViewerTile.vue` | `<video controls preload=metadata>`, 상태 loading·ok·offline·unplayable·error. 끊김 = 그때 자리(`resume`) 기억 + 정지 + 안내, 다시 접속 = `<video>` 다시 만들어 그 자리로(재생 중이었으면 재생). `<video>` 오류 = probe로 서버 오류 vs V5(재생 불가) 구분, 재생 불가 = [내려받기](`useTransfers().download`, DirEntry 최소 모양). 바탕 = 테마 `terminal`(두 테마 다 어두움) |
| 프론트 | `fileKind.util.ts` `viewIcon` · `DeviceTileBody.vue` · `ViewerRow.vue` · `OverviewTile.vue` | 영상 아이콘 `mdi-movie-outline` / D(더블클릭 열기)·칸 메뉴 "보기"에 영상 |

## 1단계 격리 확인 (2026-10-29)
- 환경: supervisor 5051(`IP=127.0.0.1`, `VIEW_CACHE_DIR=""`) + DB `nexus_vv_tmp` + worker `vv-dev`(TMPDIR `/run/user/1000/vviso`) + vite 3001(HMR 3002, `HMR_CLIENT_PORT=` 비움), 계정 `vv1`/`pw123456`, 폴더 id 1. 픽스처 = scratchpad `iso/fx`(clip.webm 60초 2.3MB · big.webm VP8 120초 89MB · clip.mp4 H.264 · hevc.mp4 · note.txt)
- 헤드리스 함정: 쿠키 호스트를 맞출 것 — API를 `127.0.0.1:5051`로 로그인하면 페이지(`VITE_API_URL=http://localhost:5051`)가 401 -> 둘 다 `localhost`
- curl: 200 전체 = 바이트 동일 · `bytes=0-` 206 · 중간 구간·끝 500바이트 바이트 동일 · 416 · If-Range 같음 206 / 다름 200 · 304 · 여러 구간 multipart · txt 415 · 없음 404
- API 15/15: 새 타일(부모 = 연 장비 타일, version +1) · 같은 파일 다시 = 같은 tileId·추가 없음 · PUT 새 영상 타일 400 · 경로 바꾸기 400 · 크기 바꾸기 + 이미지 추가 200 · 이미지 경로 415 · instance 형식 400 · 남의 장비 403 · 없는 폴더·보드 404 · 부모 닫힘 = 루트 · 보드 떼기 뒤 다른 보드 hash로 답
- 화면 21/21: 더블클릭 webm = 영상 타일(메타 60초·640)·active·머리 영상 아이콘 · 재생 · 탐색 50 · 다른 장비 타일에서 같은 파일 = 타일 하나 + active · 칸 메뉴 "보기"(영상 아이콘) · 새로고침 유지 · big.webm 100초 탐색 = 새 요청 1개 · worker 끔 = 안내 -> 다시 켬 = 102초 자리에서 재생 이어감 · HEVC = "재생할 수 없는 영상" + 내려받기 받아짐 · 폴더 타일을 b2로 뗀 뒤 같은 파일 = `/w/<b2>?focus=<타일>`로 이동 + active · 콘솔 오류 0
- 알게 된 것: 이 헤드리스 Chromium(playwright 1234)은 **H.264 mp4 재생됨**, HEVC는 안 됨 / 2.3MB 파일은 첫 요청에 다 받아 탐색해도 새 요청 없음(큰 파일로 확인)

## 2단계(S1(이어보기)) 파일 (2026-10-29)
| 층 | 파일 | 내용 |
|---|---|---|
| DB | `00015_video_positions.sql` · `query/videoPositions.sql` | `video_positions`(user_id, instance, path, pos, updated_at, PK 앞 셋). `SaveVideoPosition` = **그 파일의 보기 타일이 내 트리에 있을 때만** upsert(닫힌 뒤 늦게 온 저장이 행을 되살리지 않게) · `DeleteVideoPosition` · `DeleteOrphanVideoPositions`(타일 없는 행) |
| supervisor | `videoTile.go` | `GET /tiles/viewers/video/position?instance&path` = {pos \| null} · `PUT` {instance, path, pos} = 200 / 타일 없음 204 / 영상 아님 415 / 음수 400. 인가 = 세션(내 행만) |
| supervisor | `tile.go` `putTiles` + `tileDoc.go` `closedVideos` | P1(위치 지우는 시점): 저장 성공 뒤 같은 트랜잭션에서 빠진 영상 타일의 위치 삭제(PUT에서 빠짐 = 닫기, 보드 간 이동은 서버 동작이라 PUT으로 안 옴) |
| supervisor | `viewCache.go` `runViewSweep` | 사본 끔(`VIEW_CACHE_DIR` 빈 값)이어도 돌고, 시작 + 10분마다 `DeleteOrphanVideoPositions`(저장·닫기 엇갈림 안전망) |
| 프론트 | `tile.api.ts` | `getVideoPosition` · `saveVideoPosition(keepalive)`(keepalive = fetch) |
| 프론트 | `VideoViewBody.vue` | 마운트 때 위치 받기 -> `loadedmetadata`에서 기다렸다가 0초일 때만 옮김(끝 1초 안이면 처음부터). 저장 = 정지·탐색·끝(0)·재생 중 10초·`visibilitychange` hidden·`pagehide`·언마운트(닫기 = 닫는 PUT이 지우거나 서버가 버림), 0.5초 안 차이는 안 보냄 |

## 2단계 격리 확인 (2026-10-29)
- API 11/11: 없음 null · 저장·덮어쓰기 · 타일 없는 파일 204·안 남음 · 415 · 음수 400 · 타일 닫는 PUT = 위치 삭제 · 닫힌 뒤 저장 204 / sweep: 고아 행 넣고 재시작 = "영상 위치 1개 정리"
- 화면: 첫 열기 0 · 탐색 70 + 정지 = 저장 70 · 새로고침 = 70 · 다른 컨텍스트(다른 기기) = 70 · 재생 중 주기 저장(70에서 11.5초 재생 -> 77.97, 기준 +8은 헤드리스 재생이 벽시계보다 느려 너무 빡빡했음) · 끝 = 0 · 닫기 = 삭제 · 다시 열기 = 0 · 오프라인 -> 다시 접속 이어 재생(회귀) · 콘솔 오류 0
- **찾아서 고친 버그(재발 방지)**: 새로고침 직후 명부가 늦게 와 `online` false -> true가 되면 watch가 다시 접속으로 보고 `<video>`를 다시 만듦 -> 위치를 기다리던 옛 요소의 `onMeta`가 옛 요소에 옮기고 값을 지워 새 요소는 0초. 고침 = 다시 만들기는 상태가 offline·error일 때만 + `onMeta`는 기다린 뒤 지금 요소인지 확인
- 관찰: 큰 파일 탐색 때 Chrome이 Range 요청을 몇 ms 간격으로 수십 개 냈다 끊음(요청마다 `FILE_OPEN`) — 동작은 정상. **test3(nginx 뒤) 임시 빌드에서 사용자 체감 = "생각보다 느리지 않음"(10-29)**, nginx 설정 안 바꿈
- 확인 중 사용자 쪽(3001)에서 영상 타일 전부 닫힘(11:02~03) = 사용자 조작으로 추정(테스트 재료만 다시 만듦)
