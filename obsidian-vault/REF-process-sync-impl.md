# REF — process 동기화 1단계 구현 · 실행 확인 (탭 id · 계정 구독 · W·X·Y)

> `REF-process-sync.md`에서 10k자 기준 분할(2026-10-01) — 설계(식별 단위·결정 O~V·탭 id 수명·진행 순서)는 그쪽, **1단계 구조표·작업 단위·구현 메모·2단계 확인**은 여기. 이력 → `history/process-sync-impl.md` / 설계 → `REF-process-sync.md` / 재현 방법 → `REF-process-reconnect.md` "재현 방법" / 현재 진행 → `CURRENT.md`.

## 1단계 구조 (합의 2026-10-01, 코드 작성 완료)
| 영역 | 지금 | 1단계 후 |
|---|---|---|
| 탭 id 발급 | 없음 | 소켓 연결 URL `?tabId=`(`sessionStorage` 값) → 서버가 같은 sid이고 지금 다른 소켓이 안 쓰는 id면 재사용, 아니면 새로 발급(`util.RandomKey`) → `TAB:ID` 이벤트(V) → 프론트가 `sessionStorage`에 저장 |
| 탭 등록부(서버 메모리) | `browsers`: 소켓 → sid | 소켓 → `{sid, userId, tabId}` + 탭 id별 연결 시각. 끊긴 탭 id는 60초 보관 |
| REST 탭 식별 | 없음 | axios 공통 헤더 `X-Tab-Id`. 서버는 그 id가 요청 sid의 것인지 확인 |
| 구독(S) | `process_subscribers` 행(sid) | 소켓 연결 시 그 계정의 살아 있는 process 전부 Hub 구독. exec는 그 계정의 모든 소켓을 구독시킴 |
| 제거 | — | `process_subscribers` 테이블(그 시점의 다음 번호 마이그레이션으로 DROP — label 모듈은 미작성이라 먼저 만드는 쪽이 `00005`, 나중 쪽은 그다음 번호. Claude가 작성 시 처리, 사용자 결정 불필요) + 쿼리(`processSubscribers.sql` 전체 — 2026-10-01 추가한 `ListSubscriberSidsByProcess` 포함), 수동 구독/해지 REST 2개, `ProcessEntry.subscribers`·`AddSubscriber`/`RemoveSubscriber`/`IsSubscribed`, `ProcessManager.SubscribeProcess`/`UnsubscribeProcess`/`ListSubscriptions`/`SubscriberSids`, `subscribeSid`, kill 시 자동구독 |
| 크기 우선권(Q) | uid → sid | uid → 탭 id. 다음 후보 = 그 계정의 연결된 탭 중 먼저 연결된 순서. 403·알림 탭 기준 |
| 로그아웃 | sid 구독 해지 + 넘김(`leaveSession`) | 구독 해지 없음. 그 sid의 탭들이 가진 우선권만 즉시 넘김 |
| 목록(U) | `GET /processes/subscriptions` | `GET /processes` (프론트 `processTerm.store`의 `syncOwners`도 이걸로) |
| worker 재접속(W) | `Rebind` 뒤 구독자 0 → 출력 끊김 | `Rebind`·`startRelay` 직후 그 process 소유 계정의 연결된 소켓을 전부 Hub 구독(소켓 연결 시 쓰는 계정 단위 구독 함수 재사용) |
| worker 끊김 표시(X) | `Remove`→`Inter.Done(502)` → 브라우저에 `FAILED 502` | PENDING 분기: `Inter.PushStatus(PENDING)`(relay가 STATUS PENDING 발행) → 종료 STATUS 없이 채널만 닫는 `AgentInteractive.Detach()`(신규, `Done`과 `sync.Once` 공유) → memory 제거. `ProcessManager.Remove`는 그대로(정상 종료·안전망용), 끊김 전용 `Detach(uid)` 추가. 프론트는 PENDING을 이미 실행 중으로 취급(`isRunning`)해 수정 불필요 |
| 재접속 시 소실(Y) | "supervisor만 앎" → DB만 `FAILED 502`, 브라우저엔 아무것도 안 감(entry·relay 없음) | DB 갱신 뒤 Hub로 `STATUS FAILED 502` 직접 발행 + `cleanupProcessTopic` |
| 프론트 | — | 소켓 hook이 연결할 때마다 URL을 다시 만들게(`createWebsocketHook` url을 함수도 받게), 탭 id 유틸 신규, axios 헤더, 탭 id 받기 전 실행 버튼 비활성 |

- 새 쿼리 필요(안): 계정의 살아 있는 process(`owner_user_id` + status PENDING/PROCESS). 기존 `ListProcessesByOwner`는 전부 반환.
- 소켓 연결 핸들러(`handleSubscribeWS`)는 이미 `requireSession`으로 `sess.Data.ID`(user id)를 얻을 수 있음.

## 1단계 작업 단위 (2026-10-01 승인·작성 완료)
| 작업 단위 | 내용 | 주요 파일 |
|---|---|---|
| 1-a(DB 정리) | `process_subscribers` DROP 마이그레이션 + `processSubscribers.sql` 삭제 + 새 쿼리 "계정의 살아 있는 process"(PENDING/PROCESS) → sqlc 재생성 | `migrations/00005_*.sql`, `query/processes.sql` |
| 1-b(탭 등록부) | `browsers` → `{sid, userId, tabId, 연결 시각}` + 탭 id 발급·재사용 + `TAB:ID` + 끊긴 탭 id 60초 보관 | `subscribe.go`, 새 `tabs.go`, `protocol/messages.go` |
| 1-c(계정 구독) | 소켓 연결 시 계정의 살아 있는 process 전부 구독 + exec 때 계정의 모든 소켓 구독 + sid 구독 경로 전부 제거 | `subscribe.go`, `processApi.go`, `process.go`, `process/manager.go`·`entry.go` |
| 1-d(크기 우선권 탭 단위) | uid → 탭 id, 다음 후보 = 계정의 연결된 탭 연결 순서, REST `X-Tab-Id` 확인, 로그아웃 = 우선권만 즉시 넘김 | `sizeOwner.go`, `processApi.go`, `user.go` |
| 1-e(목록 API) | `GET /processes/subscriptions` → `GET /processes` | `processApi.go`, `processDto.go` |
| 1-f(worker 끊김·재접속) | W·X·Y | `process.go`, `process/manager.go`, `execute/agentInteractive.go` |
| 1-g(프론트) | 소켓 URL 함수화(`?tabId=`) + 탭 id 유틸 + axios `X-Tab-Id` + 탭 id 전 실행 비활성 + `syncOwners` → `GET /processes` | `websocket.hook`, 새 `tabId.util.ts`, `api.util`, `process.api.ts`, `processTerm.store.ts` |

- 확인: 1-a~1-f 후 `go build`/`vet`, 1-g 후 `vue-tsc` → 2단계(재기동·실행·출력·kill·크기·탭 2개 우선권 넘김·worker 끊김/재접속 확인 → 커밋). worker 끊김은 socat 프록시(`REF-process-reconnect.md` "재현 방법"), 탭 2개는 헤드리스.

## 1단계 구현 메모 (2026-10-01) — 구조표에 없던 것
| 항목 | 내용 |
|---|---|
| 파일 | 신규 `router/tabs.go`(`tabRegistry`: attach/detach/connOf/belongsTo/tabIDsOfSid/connsOfUser/tabsOfUser) · `migrations/00005_drop_process_subscribers.sql` · 프론트 `common/util/tabId.util.ts`(`readTabId`/`tabReady`/`bindTabId`). `browsers` 필드·`browsersForSid` 삭제 |
| 인가 추가 | kill·resize가 `requireOwnedEntry`로 **소유 계정 검사**(다른 계정이면 404). 전엔 uid만 알면 누구나 kill 가능했음 |
| 탭 id 필수 | exec·resize는 `X-Tab-Id` 없거나 요청 sid의 탭이 아니면 400. `GET /processes`는 없으면 `sizeOwner=false`. CORS `AllowHeaders`에 `X-Tab-Id` 추가 |
| 연결 순서 | 같은 탭 id로 다시 붙으면 **처음 연결 시각 유지**(새로고침해도 다음 후보 순서에서 밀리지 않음) |
| 소켓 연결 순서 | attach → `TAB:ID` Emit → 계정의 살아 있는 process 구독(`ListLive`, PENDING 포함) → `sizeOwnerOnConnect` → `NODE:0` |
| X 재발 방지 | `startRelay` 감시 고루틴이 `inter.Detached()`면 `cleanupProcessTopic`을 건너뜀 — 구독을 남겨야 Rebind 뒤 출력이 오고, 빠른 재접속의 새 구독을 늦은 정리가 지우는 경합도 막음 |
| Y | `publishLost(uid)` = STATUS FAILED 502 직접 발행 + 토픽 정리 |
| 프론트 | 소켓 URL은 함수(`createWebsocketHook(url: string \| () => string)`) — 재연결마다 `?tabId=`를 새로 만듦. `tabReady` = 이번 연결에서 `TAB:ID` 받음(연결이 `CONNECTED`가 아니면 false) → 폴더 타일 "실행" 버튼 비활성 조건 |

- **남은 한계(로그아웃)**: 로그아웃한 sid의 다른 탭 소켓은 서버가 닫지 않는다(그 탭이 다음 세션 확인에서 `auth = null` → `disconnect`할 때까지 계정 토픽 출력을 받음). 로그아웃 순간의 넘김에서만 후보에서 빠지고, 그 뒤 넘김에선 후보가 될 수 있음. 필요하면 signOut에서 그 sid 소켓을 닫는 것으로 해결.
- 같은 이유로 O(계정 동기화)는 서버 쪽만 됨: 다른 탭도 DATA·STATUS를 받지만 그 탭의 `processTerm` store에 uid가 없어 버린다(4단계 화면에서 표시).

## 2단계 실행 확인 (2026-10-01, supervisor·worker = 사용자 재기동 / 클라이언트 = Node 22 `WebSocket` + 헤드리스 크롬)
| 영역 | 확인한 것 | 결과 |
|---|---|---|
| 탭 id | 탭마다 다른 id · `TAB:ID`가 첫 이벤트 · 사용 중인 id 요청 = 새 발급 · 다른 sid가 남의 id 요청 = 새 발급 · 60초 안 같은 id 재연결 = 재사용 · 60초 지나면 새 발급 | 통과 |
| REST 탭 확인 | `X-Tab-Id` 없이 exec = 400 · 다른 sid 탭 id = 400(sid가 실제로 다를 때 — 아래 "발견") | 통과 |
| O(계정 동기화) | 같은 계정 탭 3개(sid 2개) 모두 DATA·STATUS 수신 · 실행 중 새로 연결한 탭도 수신 | 통과 |
| 크기 우선권 | 실행 탭 = 소유자 · `GET /processes` sizeOwner 탭 기준 · 비소유 탭 resize 403 · 새로고침 후 유지 · 탭 닫고 60초 전엔 유지, 60초 뒤 먼저 연결된 탭으로 + `SIZE_OWNER` 알림 · 로그아웃 = 다른 sid 탭으로 즉시(같은 sid 탭 제외) | 통과 |
| kill·목록 | 없는 uid 404 · 종료 STATUS가 계정 탭 전부에 · 종료 후 목록에서 빠짐 | 통과 |
| X | 프록시 끊김 → 브라우저 `PENDING`(FAILED 안 감) · DB PENDING · 끊긴 동안 목록에 PENDING | 통과 |
| W | 재접속 뒤 기존 탭 출력 이어짐(tick 4 → 12) · **끊긴 동안 연결한 탭도 출력 수신** · STATUS PROCESS · DB PROCESS | 통과 |
| Y | worker 재시작(프로세스 소실) → PENDING 동안 연결한 탭에 `FAILED 502` · DB FAILED 502 · 목록에서 빠짐 | 통과 |
| 프론트(헤드리스) | `TAB:ID` → `sessionStorage` · 첫 소켓 `?tabId=` 빈 값 · 이후 REST에 `X-Tab-Id`(CORS 통과) · 새로고침 = 같은 id + 소켓 URL에 이전 id · 같은 브라우저 새 탭 = 다른 id · 페이지 에러 0 | 통과(루트에 스크립트가 없어 실행 버튼 활성 확인은 생략) |

- 테스트 함정(재발 방지): 프록시 경유 테스트 worker는 **별도 SQLite**(빈 파일)로 띄울 것 — 같은 DB면 정체성(subkey)이 같아 사용자 worker와 instanceKey가 겹친다. 종료는 `pkill -f '^\./worker$'`(cwd 확인) — 경로 패턴은 `./worker`로 실행한 프로세스와 안 맞아 안 죽고 두 개가 됨(실제 발생).
- 테스트 스크립트(`lib.mjs`·`part1*.mjs`·`part2.mjs`·`part3.mjs`·`ui.mjs`)는 세션 scratchpad에만 있음.

## 발견
- **sid = 쿠키 원본값 충돌 (2026-10-01 발견 → 같은 날 수정·확인, 미커밋)**: gorilla `CookieStore` 쿠키 = `date(초)|gob(세션 값)|mac`으로 무작위 요소·암호화가 없다. 같은 계정이 **같은 초에 로그인하면 쿠키가 똑같이 나올 수 있다**(측정: 연속 로그인 20쌍 중 11쌍 동일 — gob의 map 순서가 같을 때). 그러면 서로 다른 브라우저가 같은 sid가 된다.
  | 영향 | 무슨 일 |
  |---|---|
  | 탭 확인 | 다른 브라우저 탭 id도 "같은 sid의 탭"으로 통과(같은 계정이라 인가 경계는 안 넘음) |
  | 로그아웃 넘김 | 한쪽 로그아웃 때 다른 브라우저 탭도 "같은 sid"로 후보에서 빠짐(1부 첫 실행에서 실제 발생) |
  | 로그아웃 자체 | CookieStore는 서버 상태가 없어 로그아웃 = 그 브라우저 쿠키 만료뿐. 같은 값을 가진 다른 브라우저 쿠키는 계속 유효 |
  - **수정(사용자 승인)**: `signIn`이 세션 값 `_nonce`(`sessionNonceKey`)에 `util.RandomKey(16)`을 넣고 저장 → 로그인마다 쿠키(sid)가 달라짐. 기존 로그인 쿠키는 그대로 유효(다음 로그인부터 적용). **확인**: 재기동 후 같은 초 연속 로그인 20쌍 중 쿠키 동일 0(수정 전 11), 다른 sid 탭 id 400·자기 탭 404·checkSession 200.
  - 남은 것: 로그아웃 무효화(같은 쿠키 값의 서버 쪽 폐기)는 서버 쪽 세션 저장이 필요한 별도 작업.
- **[실행 확인 2026-10-01] Rebind 뒤 출력 끊김**: worker 끊김 → `applyStatus(PENDING)` → `Remove` → relay 드레인 → `cleanupProcessTopic`이 Hub 구독 전부 해제 → 재접속 `Rebind`가 새 relay를 띄워도 토픽 구독자 0. 열려 있던 소켓은 다시 연결하기 전까지 DATA·STATUS를 하나도 못 받음. → 1단계 W로 해결, 2단계 실행 확인 통과.

  | 시점 | 기존 소켓 A | 새 소켓 B(같은 세션, 재접속 후 연결) |
  |---|---|---|
  | 실행~끊기 전 | DATA 수신(tick 1~5) | — |
  | 끊김 | `STATUS FAILED exit 502` 수신 | — |
  | 재접속 후 13초 | DATA **0건**, `PROCESS` 상태도 못 받음 | 연결 즉시 DATA 수신(tick 24~27) |
  | kill | `STATUS FAILED 137` 수신(kill REST의 `subscribeSid`가 sid 소켓을 다시 구독시켜서) | 수신 |

  재현 방법 → `REF-process-reconnect.md` "재현 방법 — worker 연결만 끊기".
- **[실행 확인 2026-10-01] 끊김 때 브라우저에 FAILED 502가 감**: process는 PENDING(살아 있음, DB도 PENDING)인데 `Remove`의 `Inter.Done(502)` 안전망이 relay로 종료 STATUS를 발행 → 브라우저는 종료로 보고 이후 상태가 갱신되지 않음. 결정 "frontend 끊김 ≠ 종료"와 같은 계열의 모순(worker 끊김 ≠ 종료). → 1단계 X·Y로 해결, 2단계 실행 확인 통과. 구조 잡다가 같은 계열 공백 하나 더 발견: 재접속 때 worker에 없던 process는 DB만 FAILED가 되고 브라우저엔 알림이 없음 → 표 Y로 함께 처리.
