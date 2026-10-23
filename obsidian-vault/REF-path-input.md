# REF — 경로 직접 입력 (장비 타일·폴더 타일 경로 줄)

> 2026-10-23 신설. 이력 → `history/path-input.md` / 장비 타일 목록·경로 줄 → `REF-worker-dir-list.md`(`INDEX-worker-dir`) / → `CURRENT.md`

**상태: 2026-10-23 마감 → `INDEX-path-input.md`. 경로-1~7 구현·확인 끝(서버 격리 33/33 + 헤드리스 31/31, 5번 연속 + Windows VM 서버 14/14·화면 12/12), 사용자 브라우저(Linux·Windows) 확인 OK.** 브랜치 `feat/path-input`(master `b17f77e`에서)

사용자 "디렉토리 타일에서 위치가 표시되는 곳, 텍스트로 직접 칠 수 있으면 좋겠어. 경로별로 칠 때마다 자동완성"(지금 당장 필요).

## 착수 결정 (2026-10-23 사용자 답)
| 번호(이름) | 결정 |
|---|---|
| Q39(입력 들어가는 법) | **경로 줄 빈 곳 클릭** = 입력칸(Windows 탐색기식). 이름 클릭은 지금처럼 그 폴더로. 단축키도 둠 |
| Q40(자동완성 대상) | **폴더만**(폴더 링크 포함) |
| Q41(자동완성 조작) | **아래 목록 + Tab** — ↑↓·Enter로 고르고, Tab = 후보들 공통 앞부분까지 채움 |
| Q42(자동완성 원천) | **worker가 거름** — 큰 폴더(500개 넘음)도 빠짐없이. worker 다시 빌드 필요 |
| (범위) | 사용자 "워커 디렉토리 타일도 되야하지만 일반 디렉토리 타일도" -> **폴더 타일(Nexus 노드 폴더)도 이번 파트**. 시작 = `기본` 목록(앞의 `기본/`은 있어도 없어도), 폴더만, 같은 입력칸 |
| Q43(같은 이름 폴더) | **목록 순서대로 첫 번째**(노드 이름은 중복을 막지 않음). 자동완성 목록엔 둘 다 보이고, 목록에서 고르면 고른 그 폴더 |
| Q44(이름 속 /) | **자동완성으로만 들어감** — 친 글자는 `/`마다 단계, `/`가 든 이름은 목록에서 골라야 하고 입력칸엔 `\/`로 보임(폴더 타일만) |

## 큰 계획 (2026-10-23 제안, 폴더 타일 추가 = 경로-6·7 — 번호는 처음 붙인 것 유지, 경로-5(확인)는 맨 끝에)
| 단계 | 내용 | 손댈 곳 |
|---|---|---|
| 경로-1(후보 요청) | `DIR_SUGGEST` {dir, prefix, max} -> {dir, items[{name, path}], more, err}. 폴더만, 앞글자 대소문자 무시(같은 대소문자 먼저), 정렬·상한. `prefix` 없고 `max` 0 = 그 폴더가 있는지만(Enter 검사). Windows `dir` "" = 드라이브 | `protocol`, worker `router/dir_suggest.go` |
| 경로-2(후보 REST) | `GET /workers/:m/:s/dir/suggest?folderId&dir&prefix` — 인가 `requireDeviceDir` | supervisor `workerDir.go` |
| 경로-3(화면: 입력칸) | 빈 곳 클릭·단축키 -> 지금 경로 전체가 선택된 입력칸. Enter = 검사 후 이동(없으면 입력칸에 오류, 타일 경로는 안 바뀜) / Esc·밖 클릭 = 되돌림. `/`·`\` 둘 다 받음 | `DeviceTileBody.vue`(입력칸은 새 컴포넌트) |
| 경로-4(화면: 자동완성) | 마지막 구분자 앞 = 폴더, 뒤 = 앞글자로 나눠 짧게 기다렸다 요청. 고르면 `이름/`까지 채우고 바로 다음 단계 후보 | 같은 컴포넌트 |
| 경로-5(확인) | 격리 + 헤드리스(큰 폴더·Windows식 경로·없는 경로·옛 worker + 폴더 타일 같은 이름·`/` 이름·공유 폴더) | scratchpad |
| 경로-6(폴더 후보 REST) | `GET /nodes/suggest?parentId&prefix`(0 = 기본 목록: 내 것 + 공유받은 것) 폴더만 / `POST /nodes/resolve` {baseId, names[]} -> id(Q43(같은 이름 폴더) 첫 번째) 또는 몇 번째 단계에서 없는지. 권한 = 지금 폴더 보기와 같은 규칙 | supervisor node 라우터, `nodes.sql` |
| 경로-7(화면: 폴더 타일 입력) | 경로-3·4의 입력칸을 폴더 타일 경로 줄에도. 목록에서 고른 단계는 id를 기억(그 앞을 고치면 버림) | `FolderTileBody.vue` |

## 구현 (2026-10-23)
| 단계 | 위치 | 내용 |
|---|---|---|
| 경로-1(후보 요청) | `protocol` `DIR_SUGGEST`, worker `router/dir_suggest.go` | Windows `C:`만 오면 `C:\`로(`C:`는 그 드라이브의 현재 폴더 = 상대 경로라 "절대 경로가 아닙니다"가 났음, VM 확인 때 잡음). 캐시 없이 매번 `f.ReadDir(512)`로 훑으며 앞글자(대소문자 무시)로 거름. 폴더 = `IsDir` 또는 링크·junction이면 `os.Stat`으로 원본이 폴더. 숨은 폴더(`.`)는 앞글자가 `.`일 때만(셸과 같음). 정렬 = 같은 대소문자 앞글자 먼저 -> 이름(대소문자 무시). 상한 200, 넘으면 `more`. `dir` "" = Linux `/`·Windows 드라이브 |
| 경로-2(후보 REST) | supervisor `workerDir.go` `suggestWorkerDir` | `GET /workers/:m/:s/dir/suggest?folderId&dir&prefix&max`, 인가 `requireDeviceDir`. 옛 worker = "핸들러 없음" RES 오류 -> 502 |
| 경로-6(폴더 후보 REST) | supervisor `nodeSuggest.go` | `POST /nodes/suggest {baseId, names, prefix, max}` 한 개로 후보·Enter 둘 다. 단계마다 보이는 자식 폴더를 기존 쿼리(`ListChildren`+`ListSharedRoots` / `ListChildrenVisible`)로 읽어 Go에서 고름(새 SQL 없음). 같은 이름 = 대소문자까지 같은 첫 번째, 없으면 대소문자 무시 첫 번째. 없으면 404 "'x' 폴더가 없습니다" |
| 경로-3·4(화면) | `PathInput.vue`(새) · `hook/pathEdit.hook.ts`(새) | props `value`·`suggest(text)`·`submit(text)`(오류 문구 = 남음). 120ms 뒤 요청, 늦게 온 응답은 순번으로 버림. ↑↓ 강조 -> Enter = 고르기(채우고 바로 다음 단계), 강조 없는 Enter = 이동. Tab = 하나면 끝까지, 여럿이면 공통 앞부분. Esc·blur = 닫힘(목록은 `mousedown.prevent`). 고르는 순간 옛 목록·강조를 비움(빠른 Enter가 옛 후보를 다시 고르던 버그). Ctrl+L = active 타일(window capture) |
| 경로-3·4 장비 | `DeviceTileBody.vue` | 경로 줄 클릭이 버튼·도구가 아니면 입력칸. Windows 판정 = 지금 경로가 드라이브(`C:\`)이거나 드라이브 목록(""), 그때만 `\`도 구분자. Enter = `max` 없이 suggest로 폴더 확인 -> 정리된 경로로 이동, 502(옛 worker)면 확인 없이 이동. 읽기 실패한 타일은 요청한 경로로 시작 |
| 경로-7(폴더) | `FolderTileBody.vue` | 기본값 `기본/이름/…`(이름 속 `\`·`/`는 `\\`·`\/`). 입력칸을 열 때 지금 경로의 각 단계(글자 -> id)를 기억 + 목록에서 고른 단계도 기억 -> 가장 긴 것에서 출발(Enter는 그 단계 글자 그대로도). 그래서 지금 경로 그대로 Enter = 같은 이름이어도 지금 폴더. 맨 앞 빈 칸·`기본` 칸은 뺌(`기본/기본/x` = 이름이 `기본`인 폴더). 같은 이름 후보엔 "같은 이름 n번째" |

## 알려진 한계 (2026-10-23)
- Windows는 2026-10-23 VM에서 확인(드라이브 목록·`C:`·`\`·`/` 구분자·junction·한글). 못 본 것 = 외부 접속(nginx 뒤) · 모바일(입력칸 폭·가상 키보드)
- 사용자 빌드 `build/directory_path`(15:21)는 `C:` 고치기 전 — Linux엔 영향 없음, VM worker(`worker-test3.exe`)는 고친 것
- 후보는 폴더만(Q40) — 파일 경로를 치면 Enter에서 "폴더가 아닙니다" · 장비 타일 후보 상한 50(더 쳐서 좁힘)
- 폴더 타일 이름에 `/`가 든 폴더는 쳐서는 못 들어감(목록에서 골라야 함, Q44) · 공유받은 루트와 내 루트가 같은 이름이면 내 것이 첫 번째
