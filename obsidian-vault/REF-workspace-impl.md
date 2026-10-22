# REF — 작업 공간 여러 개: 구현 구조안 (작업공간-a~e)

> 2026-10-07 신설. 개념·확정 사항(W1(분리 범위)~W3(이동 범위), URL, 머리 줄) → `REF-workspace.md` / 타일 트리 저장 규칙(⑪) → `REF-node-ui-save.md`(INDEX-node-ui-link) / 공유와의 순서 → `REF-process-share.md` / 서버 구현 메모·확인 결과(작업공간-a·b~) → `REF-workspace-server.md` / 이력 → `history/workspace-impl.md` / 현재 진행 → `CURRENT.md`.

**상태(2026-10-07): WS1(트리 주소)~WS7(보는 작업 공간 전달) 확정. 작업공간-a(DB)·작업공간-b(API)·작업공간-c(exec·안전망·탭) 코드·확인 완료, 작업공간-d(프론트) 코드·type-check 완료, 작업공간-e(확인) 헤드리스 34개·회귀 59개 통과(→ `REF-workspace-front.md` "확인 결과", `REF-workspace-server.md` "작업공간-e"). 커밋 전(파트 마감 때). 작업 공간 여러 개 = 끝, 다음은 공유-d(실시간). 프론트 → `REF-workspace-front.md`. 명칭(열린 질문 5(네이밍))은 화면 글자만 해당 → 작업공간-d 전까지 정하면 됨(코드 이름은 `tile_trees`·`workspace`).**

## 지금 구조 (2026-10-07 코드 확인)
| 자리 | 지금 |
|---|---|
| DB | `tile_trees(user_id PK, tree JSONB, version, updated_at)` — 계정당 1행 |
| API | `GET/PUT /tiles`(version 충돌 409 + 서버 트리), exec 응답에 터미널 타일(`appendTerminalTile`, 별도 트랜잭션) |
| 실시간 | `TILES:<userId>` 토픽, `TILES:UPDATE {tree, version, tabId}` |
| 고아 안전망(3-d) | **프론트**: `terminalRestore.hook` → `listProcesses` → `tileTree.adoptOrphans`(지금 트리에 uid가 없으면 루트에 붙임) |
| 라우트 | `/`·`/login`만. 헤더 `AppHead.vue`가 닉네임 표시 |

**발견(2026-10-07)**: 공유-b 뒤 `listProcesses`가 공유받은 살아 있는 process도 돌려주므로, **받는 쪽 고아 안전망이 그것들을 받는 쪽 `/` 루트에 붙인다**(지금 코드에서 이미 일어남, 공유 행이 없어 실제로는 아직 안 보임). 받는 쪽이 조작 미만이면 kill을 못 해서, 실행 중 터미널 닫기 금지(UI) 때문에 그 타일을 치울 수도 없다 → WS3(고아 안전망 위치)에서 대상을 좁힘.

## 결정 (WS = 작업공간 구현, 2026-10-07 전부 제안대로 확정)
| 결정 | 묻는 것 | 제안 |
|---|---|---|
| WS1(트리 주소) | API가 작업 공간을 무엇으로 가리키나 | **hash.** `tile_trees`에 `id`(PK)·`hash`(URL용, 무작위 12자, UNIQUE)·`name`·`ord`·`is_default`(계정당 하나) 추가. API = `/tiles`(기본, 지금 그대로) · `/tiles/:hash`(그 외) — `/` 화면은 바뀌는 것 없음. 기각: 숫자 id(URL `/w/{hash}`에서 id를 다시 찾아야 함) |
| WS2(변경 방식) | 만들기(분리)·폴더 이동·삭제를 어떻게 저장하나 | **서버 동작 API** 3개: 트랜잭션 안에서 관련 트리를 `FOR UPDATE`로 잡고 타일을 옮긴 뒤 두 트리 version을 올리고 `TILES:UPDATE` 두 번. 요청에 출발 트리 version → 다르면 409(프론트는 저장 대기분을 먼저 보낸 뒤 호출, 409면 다시 불러와 재시도). 도착 트리는 "루트 끝에 붙임"이라 version 검사 없음. 기각: 프론트가 두 트리를 각각 PUT(중간 실패 시 타일이 두 곳에 있거나 사라짐) |
| WS3(고아 안전망 위치) | 트리가 여러 개면 지금 트리만 보고 붙이는 프론트 안전망이 다른 트리에 있는 터미널을 중복으로 붙임 | **서버로 옮김**: `GET /tiles`(기본 작업 공간) 때 계정의 **모든 트리**를 잠그고, 살아 있는 process 중 어느 트리에도 uid가 없는 것을 기본 작업 공간 루트 끝에 붙임. 대상 = **내가 소유하거나 실행한(`started_by`) process만** — 공유받은 것은 공유-d가 SH2(받는 쪽 화면 위치) 규칙으로 넣음(위 "발견" 해소). 프론트 `adoptOrphans`·`terminalRestore`의 순서 의존 1)~2) 제거. 기각: 프론트가 모든 트리를 받아 검사(트리 수만큼 요청, 순서 경합 그대로) |
| WS4(실행 트리 기록) | `processes`에 실행한 작업 공간을 기록하나(REF-workspace에서 "추천"했던 것) | **안 둠.** 실행 중 터미널은 닫을 수 없어 고아는 저장 실패 때만 생김 → 기본 작업 공간에 붙여도 충분. 공유 뒤엔 계정마다 트리가 달라 process 한 행에 트리 하나를 적는 것도 맞지 않음 |
| WS5(터미널 타일 추가 주체) | "uid는 계정 전체 트리에서 하나"를 어떻게 지키나 | **터미널 타일은 서버만 넣음**(exec·고아 안전망·분리·이동·삭제). `PUT /tiles[/:hash]`는 저장본에 없던 uid를 추가할 수 없음(400). 이것만으로 다른 트리와 겹칠 수 없음 — 트리 간 비교 검사 불필요. ※ 별도 "터미널 모음 타일"이 생기는 것이 아님 — 새 터미널 타일 항목을 트리 데이터에 **써 넣는 주체**만 서버로 한정. 이미 있는 터미널 타일의 크기 변경·닫기·폴더 닫기 때 부모로 옮기기는 지금처럼 프론트 PUT(2026-10-07 사용자 질문으로 보충) |
| WS6(TILES 이벤트 구분) | 다른 작업 공간의 저장을 받은 탭은 | 토픽은 `TILES:<userId>` 그대로, 이벤트에 `hash` 추가 → 보는 작업 공간이 아니면 무시(목록 화면용 이름 변경 등은 별도 `WORKSPACES:UPDATE`) |
| WS7(보는 작업 공간 전달) | 크기 우선권 후보 = "그 작업 공간을 보는 탭"을 서버가 어떻게 아나 | **소켓 emit `TAB:VIEW {hash}`**(라우트 바뀔 때마다, 연결 직후 1번) → 탭 등록부(`browserTab`)에 저장. 크기 우선권: 연결 때 가져가기·넘김 후보를 **그 uid의 터미널 타일이 있는 작업 공간을 보는 탭**으로 좁힘(트리 JSON에서 uid 검색). SH6(presence)의 "보는 중"도 이 값을 씀. 기각: 계정 전체 탭 유지(다른 작업 공간을 보는 탭이 크기 소유자가 되면 화면에 그 터미널이 없어 크기를 아무도 안 맞춤) |

## API (작업공간-b·c)
| 메서드 · 경로 | 내용 |
|---|---|
| `GET /workspaces` | 계정의 작업 공간 목록 `[{hash, name, isDefault, ord}]`(기본이 없으면 만듦) |
| `GET/PUT /tiles`, `GET/PUT /tiles/:hash` | 지금 `/tiles`와 같은 규칙(version 409). PUT은 WS5 검사 추가. 기본 GET은 WS3 고아 안전망 |
| `POST /workspaces` | 만들기(분리): `{folderTileId, name?, version}` — 출발 = ~~기본 작업 공간만~~ -> `from`(출발 hash, 비면 기본) 어느 보드든(2026-10-22, `REF-fix-board-iutf8.md`), 루트 폴더 타일 불가. 그 폴더 타일 + 하위가 새 트리의 루트가 됨(W1(분리 범위) = 타일 트리 하위) |
| `POST /workspaces/:hash/move` | 폴더 이동(W3(이동 범위)): `{folderTileId, to: hash, version}` — 도착 작업 공간 루트 끝, 작업 공간의 루트 폴더 타일은 불가 |
| `PATCH /workspaces/:hash` | 이름 변경 |
| `DELETE /workspaces/:hash` | 삭제(W2(삭제 처리)): 루트 폴더 타일째 기본 작업 공간 루트 끝으로, kill 없음. 기본은 삭제 불가 |
| `POST /processes/exec` | 요청에 `workspace`(hash, 없으면 기본) — 부모 타일이 없으면 그 작업 공간 루트 끝 |

## 작업 단위 (REF-workspace 초안을 위 결정으로 갱신)
| 단위 | 내용 | 파일 |
|---|---|---|
| 작업공간-a(DB) | `00008_workspaces.sql`: `tile_trees`에 `id BIGSERIAL PK`·`hash`·`name`·`ord`·`is_default`, 기존 행 = 각 계정의 기본(`is_default = true`, hash 생성), 계정당 기본 하나(부분 UNIQUE). 쿼리 `tileTrees.sql` 갱신 | migrations, query |
| 작업공간-b(API) | `GET /workspaces`, `/tiles[/:hash]`, 만들기·이동·이름·삭제(WS2), WS5 검사, WS6 `hash` | `tile.go`, `workspace.go`(새 파일), `protocol` |
| 작업공간-c(exec·안전망·탭) | exec `workspace` / WS3 서버 고아 안전망 / WS7 `TAB:VIEW` + 크기 우선권 후보 좁힘 | `processApi.go`, `tile.go`, `tabs.go`, `sizeOwner.go` |
| 작업공간-d(프론트) | 라우트 `/w/:hash` + `tileTree.store`가 hash로 불러오기 + `adoptOrphans` 제거 + "새 작업 공간" 버튼·이동·삭제 + 헤더 이름 + `TAB:VIEW` + **터미널 머리 줄 `[실행한 타일 번호] 경로 / 이름`** | `router`, `tileTree.store`, `terminalRestore.hook`, `AppHead.vue`, 타일 컴포넌트 |
| 작업공간-e(확인) | 격리 실행 + 헤드리스: 분리·이동·삭제·URL 직접 접속·탭 2개가 서로 다른 작업 공간·고아 안전망(다른 트리에 있으면 안 붙음)·크기 우선권 후보 | — |

