# REF — 4 O(계정 동기화) 화면: 타일 트리·터미널 실시간 반영

> 2026-10-02 신설. 진행 순서 4단계(`REF-process-sync.md` "진행 순서"). 결정 O(계정 동기화) 자체 → `REF-process-sync.md` / 저장 모델·op 큐·409 재적용(3-c)·터미널 복원(3-d) → `REF-node-ui-save.md`·`REF-node-ui-save-impl.md` / 실시간 push 원칙(인가·라우팅 분리, AfterCommit) → `REF-realtime.md` / 이력 → `history/node-ui-sync.md` / 현재 진행 → `CURRENT.md`.

**상태: 구조안 확정(2026-10-02, 4①~4⑤ 제안대로 — 사용자 "제안대로 기록"). 4-a(서버 발행)·4-b(트리 수신)·4-c(터미널 등록 일원화) 작성·확인 + 사용자 PC 확인(4-d) 완료, 커밋 `4b6bbcd`(2026-10-02, push 안 함).** 4.5 SNAPSHOT(화면복원 ring buffer)과는 따로 진행하기로 함(2026-10-02 사용자 — 위험도·확인 범위 분리). 4에서 "내가 실행하지 않은 터미널 등록"을 한 곳으로 모아 SNAPSHOT이 거기에만 끼면 되게 한다.

## 지금 빈 곳
| 다른 탭에서 한 일 | 지금 내 탭 | 4 뒤 |
|---|---|---|
| 폴더 타일 열기·닫기·이동·크기 | 새로고침해야 보임 | 바로 반영 |
| 스크립트 실행(새 터미널 타일) | 새로고침해야 타일이 생김. 출력은 내 소켓에 이미 오지만 uid를 몰라 버림 | 타일이 생기고 출력도 처음부터 보임(4③) |
| 터미널 타일 닫기 | 새로고침 전까지 남음 | 사라지고 xterm 정리(4④) |

## 흐름
```
server  : socket connect -> subscribe TILES:<userId>
          PUT /tiles ok    -> AfterCommit -> publish TILES:UPDATE {tree, version, tabId}
          exec tile commit -> publish TILES:UPDATE {tree, version, tabId}
browser : TILES:UPDATE -> tabId == mine ? ignore (own response handles it)
                        -> version <= mine ? ignore
                        -> rebase(push)   (same path as 409: server tree + replay pending)
          tree changed -> terminal uids not registered -> lookup + restore  (one entry point, SNAPSHOT hooks here later)
                       -> registered uids gone from tree and not running -> dispose
```

## 결정 4①~4⑤ (2026-10-02 사용자 확정 — 전부 제안대로)
| 결정 | 확정 | 기각 |
|---|---|---|
| 4①(push 내용) | **트리 전체 + version**(⑪-1 저장 형태 그대로, 받는 쪽은 409 재적용 경로 재사용). 크기 상한 256KB라 부담 작음 | 변경분(op)만 — 서버가 op를 알아야 함(⑪-1에서 기각한 구조) |
| 4②(내 변경 되받기) | push에 보낸 탭 id를 실어 **자기 것은 무시** + version 비교 이중 확인. keepalive fetch에도 `X-Tab-Id` 붙임 | version 비교만 |
| 4③(다른 탭 실행 터미널 출력) | **모르는 uid 출력을 잠깐 보관**(uid당 64KB·10초) -> 타일 push로 uid를 알면 등록하며 씀 = 처음부터 보임, 안내 줄 없음. 넘치면 버리고 안내 줄 | 버림(빈 화면 + 안내 줄) — SNAPSHOT 전까지 |
| 4④(다른 탭이 닫은 터미널) | push 반영 뒤 트리에 없고 실행 중 아닌 uid의 xterm 정리 | 둠(새로고침 때 정리) |
| 4⑤(폴더 목록 동기화) | **이번 범위 밖** — 다른 탭의 노드 생성·이름 변경이 폴더 목록에 바로 반영되는 건 `NODE:<parentId>` 동적 구독(CURRENT 열린 질문 4) 몫 | 같이 |

## 세부 규칙
| 항목 | 규칙 |
|---|---|
| 재적용 안전성 | 저장 중(in flight)에 push가 오면 이미 서버에 반영된 내 op가 다시 적용될 수 있다 -> **op를 멱등으로**: `attachOp`는 같은 id가 있으면 throw(버림), close는 대상 없으면 throw(이미 있음), navigate·resize는 같은 값 재적용, adopt·upsert는 이미 uid/id 검사 |
| 저장 응답 처리 | `pending.splice(0, sent)`(개수) -> **보낸 op를 객체로 기억해 그것만 제거**(재적용에서 op가 빠지면 개수가 어긋남). version은 `max(내 것, 응답)` — push가 먼저 와서 더 앞선 경우 되돌리지 않음 |
| 화면 이동 | 원격 변경엔 `reveal` 안 함(보던 위치 유지) |
| 서버 발행 시점 | PUT은 요청 트랜잭션 -> `AfterCommit`(롤백 시 누설 방지, `REF-realtime.md` 주의점 1). exec 타일은 핸들러 안 별도 트랜잭션 커밋 직후 |
| 인가 | 토픽 = 계정 id, 구독은 세션 계정으로만 -> 인가 끝. 공유(5단계) 땐 공유 트리를 따로 둘지 그때 결정 |

## 작업 단위
| 작업 단위 | 내용 | 주요 파일 |
|---|---|---|
| 4-a(서버 발행) | `MsgTilesUpdate = "TILES:UPDATE"` + 연결 시 `TILES:<userId>` 구독 + PUT(AfterCommit)·exec 발행, 보낸 탭 id = `X-Tab-Id` | `protocol/messages.go`, `subscribe.go`, `tile.go`, `processApi.go` |
| 4-b(트리 수신) | `TILES:UPDATE` 수신 -> 자기 것·옛 version 무시 -> `rebase`. op 멱등화, 보낸 op 객체로 제거, version max | `tileTree.store.ts`, `tile.api.ts`(keepalive 탭 id) |
| 4-c(터미널 일원화) | 복원 hook을 "트리의 미등록 터미널 uid -> lookup -> restore"로 일반화(새로고침·push·409 모두 이 길) + 모르는 uid 출력 보관(4③) + 정리(4④) | `terminalRestore.hook.ts`, `processTerm.store.ts` |
| 4-d(확인) | 탭 2개: 폴더 열기·닫기·크기 반영 / 실행 -> 다른 탭에 타일+출력 처음부터 / 닫기 -> 사라짐 / 동시 변경 재적용 / 자기 push 무시 | 헤드리스 |

## 4-a(서버 발행) 구현 (2026-10-02)
| 파일 | 내용 |
|---|---|
| `protocol/messages.go` | `MsgTilesUpdate = "TILES:UPDATE"` + `TilesUpdateEvent{tree(RawMessage), version, tabId}` |
| `subscribe.go` | 연결 시 `tilesTopic(userID)` 구독 |
| `tile.go` | `tilesTopic` = `TILES:<userId>` / `publishTiles`(실패는 로그만) / `putTiles` 성공 시 `AfterCommit`으로 발행, 탭 = `requestTab`(검증된 `X-Tab-Id`, 없거나 남의 세션 탭이면 "") / `appendTerminalTile`이 `(tile, superdb.TileTree 행, err)` 반환 |
| `processApi.go` | exec 타일 커밋 직후 발행(탭 = exec 탭) |

- 확인(supervisor 재시작 후): scratchpad `tiles/push.mjs` — raw WebSocket으로 탭 2개 + 다른 계정 1개. PUT 200 -> 두 탭 각 1번(tree·version·보낸 탭 일치) / 다른 계정 0 / 409·400 발행 없음 / exec -> 터미널 타일 든 트리, 탭 = exec 탭 / `X-Tab-Id` 없는 PUT -> `tabId` "" — 7항목 통과
- 테스트 소켓 프레임: binary, `JSON.parse(TextDecoder)` -> `{t, d}`

## 4-b(트리 수신) 구현 (2026-10-02)
| 파일 | 내용 |
|---|---|
| `tileTree.store.ts` | `TILES:UPDATE` 수신: `tabId`가 내 탭이면 무시 -> 불러오는 중이면 `early`에 보관(받은 트리보다 새것이면 그걸로) -> `version`보다 새것만 `rebase`. `attachOp`는 같은 id가 있으면 throw(멱등). 저장 200 = 보낸 op를 `Set`으로 기억해 그것만 제거 + `version = max`. 409는 응답 version이 더 새것일 때만 rebase. **소켓 재연결 때 `GET /tiles`로 다시 맞춤**(끊긴 동안 놓친 push — 구조안에 없던 것 추가) |
| `tile.api.ts` | keepalive fetch에도 `X-Tab-Id` |

- TS 함정: `let early = null` 대입 직후 `await` 뒤에 읽으면 TS가 `never`로 좁힘(핸들러가 채우는 걸 모름) -> `early as TileTreeResponse | null`로 받아 씀
- 확인(재시작 불필요, 헤드리스 1600×1200 탭 2개): `tiles/sync.mjs` — ① A 열기 -> B 반영, B는 PUT 0 ② 양쪽 동시 열기 -> 4개로 수렴 ③ B 닫기 -> A 반영 ④ A 연속 3번 + B 1번 -> 7개 수렴 ⑤ A exec -> B에 터미널 타일, A는 exec 중 PUT 0 ⑥ B 새로고침 = 서버 — 전부 통과. 회귀: `store.mjs` 4번이 A 3개 -> 4개(push 반영, 기대대로) / `exec.mjs`·`restore.mjs` 같은 결과
- 테스트 함정: API로 만든 노드는 열린 폴더 목록에 안 뜸(4⑤(폴더 목록 동기화) 범위 밖) -> 목록 새로고침 버튼(`[title="새로고침"]`) 누른 뒤 진행. 닫기 버튼 = `[title="닫기"]`

## 4-c(터미널 등록 일원화) 구현 (2026-10-02)
| 파일 | 내용 |
|---|---|
| `processTerm.store.ts` | 요청 중에만 모으던 `early`·`earlyStatus`·`fetchInFlight` -> **항상 켜진 보관함 `held`**(모르는 uid의 출력·최신 상태, uid당 64KB 넘으면 오래된 것부터 버리고 `overflow`, 10초 지나면 버리고 `expired`에 기록 -> 다시 보관해도 overflow). `restore(fetch, { reload })`: 실행 중 + (reload 또는 overflow)면 안내 줄(reload = `reconnected — earlier output not shown`, push = `earlier output not shown`) -> `register`(보관분 쓰기) -> 끝난 process는 **보관분 뒤에** 종료 줄. `prune(keep)` = 트리에 없고 실행 중 아닌 것 dispose(4④) |
| `terminalRestore.hook.ts` | 새로고침 복원(`reload`) + **`sync`**(트리 uid 목록이 바뀔 때마다: `prune` -> 미등록 uid만 `lookupProcesses` -> `restore({reload:false})`). 조회는 한 번에 하나(`syncing`), 끝나면 `next()`로 다시 판단(그 사이 트리 재로드면 복원부터). 조회해도 없던 uid는 `missing`(재조회 안 함). 복원 실패 = `failed` -> tabReady·ready 변화 때만 재시도 |

- **SNAPSHOT 끼울 자리 = `restore` 한 곳**(새로고침·push·409 재적용 전부 이 길)
- 작성 중 잡은 것 2개: ① 조회 중 트리 재로드 시 복원이 영영 안 시작 -> `next()` ② 복원 실패 때 바로 재시도 -> 서버가 죽으면 무한 반복 -> `failed`
- 확인(재시작 불필요): `tiles/sync2.mjs` 탭 2개 — ① A 실행 -> B에 `start-marker`부터 출력, 안내 줄 없음, RUN ② B에서 kill -> 양쪽 `process exited: 137` ③ 바로 끝나는 스크립트 -> B에 출력 + 종료 줄 1번 ④ A에서 닫기 -> B에서도 사라짐 ⑤ B 새로고침 -> 안내 줄 + 새 출력 — 전부 통과. 회귀 `sync`·`store`·`exec`·`restore` 같은 결과
