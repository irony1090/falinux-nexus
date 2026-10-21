# REF: 리팩터링 — 소켓 밖 영역 정리 후보 (R1(목록 조사) 2회차)

> → 이력 `history/refactor-cleanup.md` / → 소켓 영역 `REF-refactor-socket.md`·`REF-refactor-socket-batches.md` / → 현재 `CURRENT.md`
> 2026-10-21 조사 -> 같은 날 R2(범위 결정) "가 -> 나 -> 다" 승인, 결과는 맨 아래. 목적·기준은 `REF-refactor-socket.md` "목적·기준"(읽기 쉽게, 성능 안 떨어지는 선, 주석은 손대는 파일만 한 줄 기준). 범위는 R2(범위 결정)에서 사용자가 정한다

## 조사 방법
- Go 미사용 함수: `go run golang.org/x/tools/cmd/deadcode@latest -test=false ./cmd/supervisor ./cmd/worker`(v0.38.0은 Go 1.26 문법에서 panic -> latest). `-test`로도 다시 돌려 테스트에서만 쓰는 것(`transport/pipe.go`)은 뺌. 결과 90개(gen 제외)
- 프론트 미사용 export: `src` 전체에서 export 이름을 다른 파일이 한 번도 안 쓰는 것(Python 정규식, 같은 파일 안에서만 쓰는 것은 정상으로 봄)
- 큰 파일: `wc -l` 상위(gen 제외). 주석: 3줄 이상 주석 블록 = 77파일 591줄(따로 항목 안 만듦 — 손대는 파일에서 같이 정리)
- `staticcheck`(~/go/bin)는 go1.25.2로 빌드돼 이 모듈(go1.26)을 못 읽음

## 정리 후보 목록

| 번호(이름) | 지금 | 정리안 | 크기 |
|---|---|---|---|
| 정돈-1(PortBridge 실행 잔재) | `execute/execTmux.go`(190줄, tmux 세션)·`execute/pty/execFIFO_linux.go`(210줄, FIFO)·`execCommand.go`의 `ExecCommand`·`streamOutput` 등 4개·`syncProcess/limitArray.go`·`util/path.util.go`·`util/byte.util.go` = 호출 0. PortBridge 이식 때 가져온 실행 방식(REF-process.md "execute는 스텁 아님") | 삭제. `execCommand.go`의 `CommandStatus` 등 쓰는 것은 남김 | 큼(~500줄) |
| 정돈-2(옛 WS 클라이언트) | `internal/web/wsClient.go`(245줄, `WsClient`) 호출 0 — `transport.Conn`이 대신함 | 파일 삭제 | 중간 |
| 정돈-3(worker 저장소 트랜잭션) | `internal/worker/store/connectManager.go`의 `Transaction` 묶음 + `StorePool.Close`·`Transaction` 12개 호출 0(worker는 `Queries()`만 씀) | 트랜잭션 부분 삭제 | 중간 |
| 정돈-4(작은 잔재) | worker `baseDir` 필드(저장만)·`instanceKey()`(호출 0) / `constants.pathToAbsolutePath` / `transport.Conn.State`·`ConnState.String` / `Interactive.Refresh` | 삭제 | 작음 |
| 정돈-5(실행 인터페이스 안 쓰는 메서드) | `IInteractive.Output`·`ExitCode`를 아무도 안 부름(`OutputAll`·`Status`만 씀) — 구현 3벌(Linux·Windows PTY·AgentInteractive)이 들고 있음 | 인터페이스에서 빼고 구현에서 삭제. 인터페이스 변경이라 정돈-1과 따로 | 작음~중간 |
| 정돈-6(tile.go 세 주제) | `router/tile.go` 820줄 = 타일 문서 모델(타입·validate·marshal·폴더 찾기) + 터미널 타일 넣기·배치(`insertTerminal`·`placeFunc`·`adoptOrphans`) + REST·고정 지도 | 파일 3개로 이동만(`tileDoc.go`·`tilePlace.go`·`tile.go`), 코드 그대로 | 중간(diff는 큼, 내용 변경 0) |
| 정돈-7(주석 "표시줄") | 프론트 4파일 9곳(`tileSplit.util.ts`·`GridPager.vue`·`TileStrip.vue`·`tileGrids.store.ts`) — 지금 이름은 인디케이터 | 주석 낱말만 바꿈 | 작음 |
| 정돈-8(TileFrame.vue 섞인 관심사) | 614줄(script 240 · template 31 · style 223) = 이름·경로 표시 + 공유 표시 + 크기 메뉴 + 키보드 맞춤-1(타일 띄우기, 약 34줄) + 머리 줄 버튼·탭 | 키보드 맞춤 부분을 composable로 빼기 | 중간, 보류 추천했으나 사용자 결정으로 진행(아래 결과) |
| 정돈-9(프론트 공용 유틸 미사용) | `common/util/index.util.ts`(`CircularBuffer`·`decimal`·`deepMap` 등 9개)·`constants.util.ts` 5개·`array.util`·`entity.util`·`lazy.listener`·`draggableLogic` 일부·`useCommitRefs`·`compareSizeGrade` = 이 프로젝트에선 안 씀 | 범용 유틸 모음(→ `REF-util.md`, 프로젝트 사이에 들고 다니는 것)이라 **유지 추천** — 사용자 판단 | — |
| 정돈-10(파일 전송 송신부) | supervisor `SendFile`·`AbortFile`·`transfer.ReadFile`·`SaveFile.Write` 등 호출 0 | 안 나(업로드 기능)에서 쓸 코드 — **유지** | — |

### 목적·기준으로 본 것
| 번호(이름) | 읽기 효과 | 성능 | 위험 |
|---|---|---|---|
| 정돈-1(PortBridge 실행 잔재) | 큼 — `execute`를 읽을 때 "어느 게 진짜 경로냐"가 사라짐 | 같음(빌드 크기만 줄 수 있음) | 없음(호출 0) |
| 정돈-2(옛 WS 클라이언트) | 중간 — `web` 패키지에 클라이언트가 두 개처럼 보임 | 같음 | 없음 |
| 정돈-3(worker 저장소 트랜잭션) | 중간 | 같음 | 없음 |
| 정돈-6(tile.go 세 주제) | 큼 — 820줄에서 주제별로 찾음 | 같음 | 없음(이동만) |
| 정돈-4(작은 잔재)·정돈-7(주석 "표시줄") | 작음 | 같음 | 없음 |
| 정돈-5(실행 인터페이스 안 쓰는 메서드) | 작음~중간 | 같음 | 낮음(Windows 빌드 확인 필요) |

- 추천 묶음: **묶음 가(안 쓰는 코드 삭제)** = 정돈-1·2·3·4 / **묶음 나(파일 나누기)** = 정돈-6 / **묶음 다(작은 것)** = 정돈-5·7. 정돈-8·9·10은 보류·유지
- 확인: 삭제는 `go build`(linux + `GOOS=windows`)·`go vet` + `deadcode` 재실행. tile.go 이동은 build·vet + 격리 소켓 35개(타일·exec 경로를 지남)


## 결과 (2026-10-21, 커밋 전)
| 묶음 | 한 것 |
|---|---|
| 가(안 쓰는 코드 삭제) | 정돈-1: `execTmux.go`·`execFIFO_linux.go`·`limitArray.go`·`byte.util.go`·`path.util.go` 삭제, `execCommand.go` -> `commandStatus.go`(`CommandStatus`만 남김) / 정돈-2: `web/wsClient.go` 삭제 / 정돈-3: worker store `Transaction`·`Close` 삭제 / 정돈-4: worker `baseDir`(필드·`NewWorkerRouter` 인자·main 인자)·`instanceKey()`, supervisor `pathToAbsolutePath`·`separator`·`currentDir`, `transport.ConnState` 전체(쓰기만 하던 `state` 필드 포함), `Interactive.Refresh` |
| 나(파일 나누기) | 정돈-6: `tile.go`(820) -> `tile.go`(267, REST·고정 지도·트리 읽기) + `tileDoc.go`(290, 문서 모델) + `tilePlace.go`(284, 터미널 타일 넣기·고아 입양). 코드 줄 대조 = 같음 |
| 다(작은 것) | 정돈-5: `IInteractive`에서 `Output`·`ExitCode` 빼고 구현 2벌(pty `Interactive`·`AgentInteractive`)에서 삭제 / 정돈-7: "표시줄" -> "인디케이터" 9곳(조사 붙임 고침) |
| 정돈-8(TileFrame.vue 섞인 관심사) | 맞춤-1(타일 띄우기) 34줄 -> `feature/node/hook/tileFloat.hook.ts`의 `useTileFloat(frameEl, termUid, overview)`. 로직 그대로, 판정만 `tile.type === 'terminal'` 대신 `termUid`(터미널일 때만 값). 인자는 `Ref`로 받아 `.value`를 넘기면 type-check가 잡음(반응성 끊김 방지). `TileFrame.vue` +5 -37 |
| 주석 | 손댄 Go 파일의 여러 줄 주석을 한 줄로(`workerRouter.go`·`conn.go`·`connectManager.go`·`doc.go`·`pty` 패키지 주석 등). 재발 방지·동시성 근거(`handleSignals`·`Status` 반환값·`Detach`·`appendTerminalTile`)는 유지 |

- 확인: `go build`(linux·`GOOS=windows`)·`go vet`·`go test ./...` + 프론트 `type-check` + `deadcode` 재실행(남은 것 = 정돈-10(파일 전송 송신부)·테스트 전용 `pipe.go`뿐) + 격리 소켓 35/35(worker 포함)
- 규모: 24파일 +638 -1830(tile 분할 이동 포함, 실제 삭제 약 1,200줄)
- **옛 문서의 `tile.go` 함수 위치**: `tileItem`·`tileDoc`·`validate`·`marshal`·`keepFolderMeta`·`folderTileOf` 등 = `tileDoc.go` / `appendTerminalTile`·`attachViewerTerminal`·`insertTerminal`·`visibleAncestors`·`bindingTile`·`adoptOrphans` = `tilePlace.go` / 나머지(REST·`publishTiles`·`pinsOf`·`newTreeHash`) = `tile.go`
- 정돈-8(TileFrame.vue 섞인 관심사) 확인: 프론트 `type-check` + 사용자 Edge 폰 실기기 5항목(키보드 열기·열린 채 보드 가로 이동·터미널 안 스크롤·닫기·폴더 타일/전체보기에선 안 뜸) 전부 통과
- 남은 것: 정돈-9(프론트 공용 유틸 미사용)·정돈-10(파일 전송 송신부) 유지