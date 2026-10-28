# REF — 장비 파일 편집·보기 (편집기 · 보기 타일 · 칸 메뉴 내려받기)

> 2026-10-28 신설(사용자와 구상 대화, 미착수). 이력 → `history/file-view.md` / 나중에 할 것 → `BACKLOG.md` / 옛 구상 → `REF-worker-dir.md` Q4(이미지·텍스트 보기)
> 상태: **방향 합의, 세부 기본안 = 사용자 확인 대기**(아래 "정할 것" 표). 착수 때 파트 단위로 CURRENT에 올린다

## 무엇을 만드나 (사용자 요청 원문 요지)
- 장비 디렉토리 타일 안의 **모든 파일**을 각 worker의 진짜 편집기(vi 등)로 열기. 본인 장비라 본인만 읽기·쓰기. 새 파일 작성도
- Windows도 지원(집 서버가 Windows라 중요). 안 되면 브라우저 텍스트 편집기 정도
- 이미지 = 타일 단위로 열리고 확대·축소. 영상도
- 새로고침해도 타일이 남아야 함
- 칸 메뉴(⋮·우클릭)에 **선택하지 않아도** 내려받기 버튼

## 합의한 방향
| 이름 | 내용 |
|---|---|
| 편집 = 터미널 타일 | 칸 메뉴 "편집" -> 새 터미널 타일에서 `<편집기> <파일>`. 기존 "여기서 셸"(`REF-dir-shell.md`)과 같은 경로, 셸 대신 편집기 명령 |
| 권한 = 장비 주인만 | 기존 `requireDeviceDir` 그대로. vi는 `:!명령`으로 셸을 띄우므로 어차피 주인 전용이어야 함. 공유받은 사람은 메뉴 자체가 안 보임 |
| 새 파일 | 경로 줄 "새 파일" -> 이름 입력 -> `<편집기> <새 이름>`(저장할 때 생김) |
| 편집기 찾기 | worker가 접속 때 쓸 수 있는 편집기를 보고(없으면 브라우저 편집기로). 집 서버 = **Windows 10**이라 `edit`(Windows 11 최신 내장)이 기본으로 없음 -> `winget install Microsoft.Edit` 또는 vim 설치 필요 |
| 보기 타일 = 한 종류 | 이미지·영상을 `type = viewer` 하나로, 확장자로 나눔. 필드 = 장비 타일과 같음(`nodeId` 권한 폴더 · `instance` · `path`(**파일** 경로)). 서버 `tileKinds`·프론트 `TILE_KINDS` 두 표에 같이(MEMORY "타일 종류") |
| 새로고침 | 저장된 필드로 `<img>`/`<video>` 주소를 다시 만듦(내용은 저장 안 함). 장비 오프라인 = 안내 + `DEVICE:CHANGE` 오면 다시 불러옴 / 파일 없음 = 안내 + 닫기 |
| 확대·재생 위치 | 브라우저에만(localStorage), DB 아님 |
| 순서 | 파트(편집기) -> 파트(보기 타일: 이미지·사본) -> 파트(영상 Range) |

## 데이터 흐름 (코드 확인 2026-10-28)
- 브라우저 <-> supervisor = **일반 HTTP GET**(브라우저 소켓 아님). 새 화면 표시 전용 API 필요(예: `GET .../files?path=`) — 지금 내려받기는 "등록(`POST .../downloads`) -> `GET /transfers/:id/data` 한 번만" + `attachment`·`octet-stream`·`no-store`라 `<img src>`에 못 씀
- supervisor <-> worker = 이미 붙은 worker 소켓 위 `FILE_OPEN` 1회 + 조각(`sendChunkSize` 256KB)마다 `FILE_READ` 왕복 1회(`cmd/supervisor/router/workerRead.go` `workerReader`)
- supervisor는 **지금 조각 하나만 메모리**, 보내면 버림. 디스크 저장 없음. 요청 끝 = 읽기 닫음
- 조각 왕복이 순서대로라 supervisor<->worker가 멀면 상한(왕복 50ms ≈ 5MB/s) — 영상 재생엔 충분
- **영상 탐색(Range)은 worker 프로토콜 변경 불필요**: `workerReader.SeekTo`(파일 하나일 때 위치 이동, worker 간 전송 이어받기용)가 이미 있음 -> `Range` 헤더 -> `FILE_OPEN` + `SeekTo` + 그 구간만 읽기 -> `206`. (대화 중 "프로토콜에 offset 추가 필요"라 했다가 정정)

## 이미지 사본 (안 1(타일 존재 기준), 영상은 사본 없음)
- 비교했던 것: 안 1(타일 존재) = 타일이 트리에서 빠질 때 삭제 / 안 2(보는 중) = 띄운 브라우저 0개 + 유예 뒤 삭제(presence 집계 필요). **이미지만 안 1, 영상은 사본 없이 그때그때 Range**(사용자 결정)
- 덤: worker가 꺼져 있어도 이미지가 보임("장비 오프라인, 마지막 사본" 표시)
- 타일은 사용자마다 **트리 전체 JSONB 하나**(`tile_trees.tree`) -> "타일 삭제" 사건이 따로 없음. 그래서 삭제 시점:

| 시점 | 하는 일 |
|---|---|
| 이미지 타일 표시 | 사본 없음 또는 원본 크기·수정 시각이 다르면 worker에서 복사 |
| 타일 트리 저장(한 곳) | 새 트리의 보기 타일 목록 vs 그 사용자의 사본 목록 -> 빠진 사본 삭제 |
| supervisor 시작 | 모든 사용자 트리 기준으로 같은 정리(중간에 꺼진 찌꺼기) |

- 사본 키 = 사용자 + `instance` + 경로(같은 파일 타일 둘 = 사본 하나, 둘 다 빠져야 삭제)
- 원본 변경 = 장비 켜져 있으면 표시 때마다 크기·수정 시각 확인(내용 전송 없이 왕복 1회)
- 브라우저 캐시: `ETag`(크기+수정 시각) -> 같으면 `304`. 브라우저마다 첫 표시는 각자 받음(브라우저 캐시는 브라우저별)

## 사본 폴더 env
| 항목 | 기본안 |
|---|---|
| 이름 | `VIEW_CACHE_DIR`(`WEB_DIR`와 같은 형식, `constants/env.go`) |
| 상대 경로 | 실행 파일 폴더 기준 |
| 기본값 | `./view-cache` |
| 비면 | 사본 안 만들고 이미지도 그때그때 전달 |

- **다른 supervisor 둘이 같은 폴더 금지**: 시작 정리가 "내 DB 트리에 없는 사본 삭제"라 개발(`nexus`)·test3(`nexus_build`)이 같은 폴더면 서로 지움 -> 폴더 안에 DB 이름 표시 파일, 다르면 경고 + 사본 기능 끔
- exe 내장 env 견본(`REF-build` E3)에도 새 키 추가

## 칸 메뉴 내려받기 (2026-10-28 추가 요청)
- 지금: 칸 메뉴(`DeviceTileBody.vue` `menuItems`, ⋮·우클릭 공용) = [선택 · 원본 위치로(링크만)]. 내려받기는 **선택한 뒤** 경로 줄 묶음(`selEntries.length` 있을 때)에만
- 요청: 선택 없이 칸 메뉴에서 그 항목 하나 내려받기(파일 = 그대로, 폴더 = zip). `download(spot, [e])` 재사용, `canUpload` 조건 같음
- 같은 메뉴에 "편집"·"보기"도 들어올 자리 -> 메뉴 구성은 파트(편집기)와 같이 정함. 단독으로 먼저 해도 가벼움

## 다음 파트(보기 타일) 결정 (2026-10-28 세션 끝)
- **V1~V5 = 기본안대로 확정**(사용자 "추천대로"). E1·E3~E5는 파트(파일 편집기)에서 정해짐, E2(브라우저 편집기)는 계속 보류
- **읽기 전용 텍스트 보기 = 안 함**(사용자: "없어도 될 것 같다"). 텍스트는 편집기로 봄
- **새 요청 D(더블클릭 열기)**: 실행 파일이 **아닌** 파일을 더블클릭하면 종류로 고름 = 이미지 -> 이미지 보기 / 영상 -> 영상 보기 / 그 밖 -> 편집기(vi·edit). 지금 더블클릭 = 폴더 들어가기 · 실행 파일 실행 · 그 밖 아무 일 없음(`DeviceTileBody.vue` `useItemPress` onDouble)
  - **다음 세션 시작 때 확인할 것**: ① 위 해석이 맞는지 ② 바이너리(zip·so·exe 아닌 실행 비트 없는 것 등)도 편집기로 열지, 아니면 아무 일 없음/내려받기 ③ 영상 파트 전에는 영상 파일 더블클릭 = 무엇(편집기 X → 아무 일 없음 제안) ④ 편집기 부분(그 밖 -> 편집기)은 지금 바로 넣을 수 있음 -> 보기 타일 파트에 같이 넣을지
  - 확장자 판정은 V1(확장자) 목록 한 곳에서(이미지·영상 목록을 더블클릭·칸 메뉴 "보기"가 같이 씀)

## 정할 것 (기본안 — V1~V5는 위 절대로 확정, 표는 원문 보존)
| 이름 | 내용 | 기본안 |
|---|---|---|
| E1(편집기 고르기) | 우선순위 | Linux `vim`->`vi`->`nano` / Windows `vim`->`edit`. 장비별 설정은 나중 |
| E2(브라우저 편집기) | 편집기 없을 때 대안을 이번에? | 나중(읽기·쓰기 API·UTF-8/CP949·CRLF = 파트 크기) |
| E3(편집 대상) | 모든 파일 vs 텍스트만 | 모든 파일(사용자 요청) |
| E4(편집 종료) | vi 닫으면 터미널 타일 | 자동으로 닫음, 오류 종료면 남겨 메시지 |
| E5(새 파일) | 같은 이름이 있으면 | 그 파일을 편집으로 엶 |
| V1(확장자) | 보기 타일 형식 | 이미지 png·jpg·gif·webp·bmp·svg / 영상 mp4·webm·mov |
| V2(SVG 보안) | SVG 스크립트가 Nexus 주소 권한으로 실행될 위험 | **필수**: `Content-Security-Policy: sandbox` |
| V3(사본 상한) | 사본 만드는 이미지 크기 | 20MB, 넘으면 그때그때 |
| V4(확대 조작) | | 휠·핀치 확대, 끌어 이동, 더블클릭 맞춤 |
| V5(재생 불가) | mkv·HEVC 등 | "재생 불가" + 내려받기(ffmpeg 변환은 범위 밖) |

## 파트(파일 편집기) — 착수 2026-10-28 (브랜치 `feat/file-edit`)
결정(사용자 10-28): E4(편집 종료) = 셸과 같게 타일 남김(COMPLETED) · 칸 메뉴 내려받기 이번 파트에 포함 · 이름·브랜치 추천대로. "여러 파일 한 번에 받기"는 이미 있음(선택 -> 경로 줄 `[N개 | ⬇]` = zip) -> 칸 메뉴 내려받기도 끌어 놓기와 같은 규칙(선택된 칸이면 선택 전부, 아니면 그 칸)

E1(편집기 고르기) = **기존 규칙 그대로**: 스크립트 노드 EDIT용 `{WORKER_EDITOR}` -> worker `resolveEditor`(`$VISUAL` > env `EDITOR` > Linux `vi` / Windows PATH `edit`->`nvim`->`vim`->`nano`, 없으면 거절 "worker env에 EDITOR"). 기존 EDIT는 끝날 때 내용을 `nodes.content`로 회수(MsgEditResult)라 장비 파일엔 못 씀 -> "여기서 셸"과 같은 **치환 토큰 + Direct** 방식

| 층 | 파일 | 바꿀 것 |
|---|---|---|
| supervisor | `workerExec.go` | 요청 `Args`(편집 대상 경로 1개) — `Cmd = {WORKER_EDITOR}`일 때만·꼭 1개 |
| supervisor | `process.go`·`process/manager.go` `ExecFile` | args 인자 추가 -> `processes.args`에 저장(재접속 복원 그대로) |
| supervisor | `tileNotice.go` | `fileBase`: 편집 토큰 = "편집" / `procName` = "편집 <파일명>"(args 있음). 실행 목록 SQL은 args 없음 -> "편집"만(SQL 안 고침) |
| worker | `process.go` | Direct + `Cmd`가 편집 토큰이면 `prepareDirectEdit`(지역화 전에 판정) |
| worker | `dir_exec.go` | `prepareDirectEdit`: `prepareEdit` 재사용(편집기 값 펼치기) + **회수 없음** + 실행 비트 검사 없음 + 대상 = 절대 경로·폴더 아님(없으면 새 파일)·상위 폴더 있음 |
| 프론트 | `tileLabel.util.ts`·`TerminalTile.vue` | `EDITOR_CMD` + `fileLabel(cmd, args)` = "편집 <파일명>" |
| 프론트 | `process.api.ts`·`processTerm.store.ts`·`nodeOps.store.ts` | `execFile`에 args |
| 프론트 | `DeviceTileBody.vue` | 칸 메뉴 [선택 · 편집(파일·링크->파일) · 내려받기 · 원본 위치로] + 경로 줄 "새 파일"(`NodeNameDialog` 재사용, 이름에 구분자·`.`·`..` 금지) |

### 구현·확인 (2026-10-28, 미커밋)
- 위 표대로 구현. 추가로 `fileLabel` 인자 = `readonly string[]`(process 응답 `args`가 readonly). protocol·DB·SQL 안 바꿈
- 빌드: Linux `go build ./...` · `GOOS=windows` worker·supervisor · `go test`(router) · `vue-tsc` 통과
- 격리 확인(supervisor 5051 + DB `nexus_fe_tmp` + worker `fe-dev` + vite 3001 헤드리스, 계정 `fe1`) — API 12/12(시험 스크립트 쪽 잘못된 API 1건 제외, DB로 대신 확인):
  - 거절 6종: 편집 토큰 + 인자 없음 400 / 일반 실행 + 인자 400 / 인자 2개 400 / 폴더 편집 "폴더는 편집할 수 없습니다" / 상위 폴더 없음 / 상대 경로 -> worker 거절은 기존처럼 FAILED(126)
  - 기존 파일: vi로 열림 -> 입력(`Goadded` ESC `:wq`) -> 파일 반영, COMPLETED(0), `processes.args` = [파일]
  - 새 파일: 없는 경로 -> `ihello` `:wq` -> 파일 생김
  - 화면: 칸 메뉴 [선택 · 편집 · 내려받기] · 편집 -> vi 터미널 타일 이름 "편집 exist.txt" · 새 파일 버튼(`mdi-file-plus-outline`, 셸 버튼 옆) -> 이름 `a/b.txt` = 경고, `n2.txt` = "편집 n2.txt" 타일 · 칸 메뉴 내려받기(선택 없이) = `exist.txt` 내용 그대로 받음 · 콘솔 오류 0
- 알게 된 것: 새 파일은 편집기가 저장해도 목록이 자동으로 안 바뀜(새로고침 필요, 장비 목록은 원래 수동) / vi 스왑 파일(`.exist.txt.swp`)이 편집 중 목록에 보임 = vi 동작
- 격리 worker 함정: TMPDIR을 scratchpad로 두면 보관 프로세스 소켓 경로가 길어(116 > 85) worker가 못 뜸 -> `TMPDIR=/run/user/1000/feiso`처럼 짧은 경로(exe 경로와 무관해 dev 판정도 피함)
- **Windows(VM `nexus-win`, 2026-10-28) 7/7**: 격리 supervisor를 `IP=0.0.0.0`으로 다시 띄움 + 장비 `win-fe`(폴더 `winfe`) + `shared/worker-fe.exe`·`start-fe.ps1`(env `C:\nexus\env-fe`, 픽스처 `C:\nexus\fx-fe\a.txt`). VM(Windows 11)엔 `edit.exe`만 있음(nvim·vim·nano 없음) -> 폴더 거절 · 상위 폴더 없음 거절 · 기존 파일 `edit`(키 `X` + Ctrl+S + Ctrl+Q) 저장 반영 · 새 파일 `n.txt` 생성
  - **편집기 없음(= 집 서버 Windows 10 상황)**: `start-fe-noeditor.ps1`(PATH = `C:\nexus`만으로 worker) -> 400 "worker가 실행을 거부했습니다: 에디터를 찾을 수 없습니다 — worker env에 EDITOR를 설정하세요" -> 화면엔 오류 창. 집 서버는 `winget install Microsoft.Edit`(또는 vim 설치 + env `EDITOR`)
- **사용자 확인(2026-10-28, 격리 3001 + VM) -> 파트(파일 편집기) 마감**. 정리 끝: 격리 프로세스·DB `nexus_fe_tmp` DROP·`/run/user/1000/feiso`·VM worker·`docker stop nexus-win`(VM `shared/worker-fe.exe`·`start-fe*.ps1`은 남김)
- 집 서버: **편집기를 일부러 아직 안 깖**(사용자 2026-10-28) — 새 빌드로 먼저 "에디터를 찾을 수 없습니다"를 보고, 그다음 `winget install Microsoft.Edit`로 차이를 볼 예정
