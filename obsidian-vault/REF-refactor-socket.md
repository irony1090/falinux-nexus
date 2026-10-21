# REF: 리팩터링 — supervisor 소켓 알림 규약 정리

> → 이력 `history/refactor-socket.md` / → 현재 `CURRENT.md` / 원래 결정 → `REF-realtime.md`, `REF-node-list-sync.md`, `REF-process-share-reconcile.md`
> 파트(리팩터링)의 R1(목록 조사) 결과 + 1차 정리. 작업 브랜치 = `refactor/socket`(파트 마감 때 master에 `--no-ff` 머지)

## 목적·기준 (2026-10-21 사용자 결정, 파트(리팩터링) 전체)
- **목적 = 사람이 Claude가 짠 코드를 읽기 쉽게.** 성능이 같이 좋아지면 더 좋음
- **기준 = 성능이 떨어지지 않는 선까지만.** 읽기 쉬워져도 성능이 떨어지는 변경은 하지 않는다
- **주석 = 최대한 간단하게**(2026-10-21): 손대는 파일은 주석을 한 줄 기준으로 줄인다. 여러 줄은 순서 의존·동시성·재발 방지만(전역 CLAUDE.md "코드 주석 분량")
  - 범위 = **손대는 파일만**. 주석이 다른 파일을 가리키면(예: "같은 이유(shareAccess.go)") 그 원래 파일에서는 **연관된 주석만 최소한으로** 고친다 — 그 파일 전체 주석 정리는 안 함

## 지금 구조 (2026-10-21 코드 기준)

| 전달 단위 | 키 | 실리는 kind | 받는 쪽을 정하는 곳 |
|---|---|---|---|
| process | `PROCESS:<uid>` (`processTopic`, process.go) | `DATA` · `STATUS` · `PROCESS:UPDATE` | **구독 시점** DB 판정 — 구독 지점 4곳(아래 소켓-6) |
| 계정 | `TILES:<userId>` (`tilesTopic`, tile.go) | `TILES:UPDATE` · `TILES:PINS` · `TILES:NOTICE` · `NODE:CREATE/UPDATE/DELETE` · `ACTIVE:UPDATE` · `SHARE:CHANGE` · `WORKSPACES:UPDATE` (7종 묶음) | 본인 계정이거나 **발행 시점** DB 판정(node·active) |
| 탭 | 토픽 없음, `conn.Emit` 직접 | `TAB:ID`(subscribe.go) · `PROCESS:SIZE_OWNER`(sizeOwner.go `emitSizeOwner`) | 탭 등록부 `tabs.connOf` |
| worker 장비 | push 없음 | — | 프론트 polling 10초(`WORKER_POLL_MS`) |
| presence | push 없음 | — | 프론트 polling 5초(`presence.api.ts REFRESH_MS`) |

- 소켓 연결 때 구독: `handleSubscribeWS` = `subscribeLiveProcesses`(볼 수 있는 살아 있는 process) + `TILES:<userId>` 1개. 끊길 때 `UnsubscribeAll`
- 브라우저 → 서버 소켓 이벤트: `PROCESS:INPUT` · `TAB:VIEW` · `TAB:ACTIVE`

> **부가 설명 — 토픽 키 vs kind 이름** (둘 다 `TILES:`로 시작해 헷갈림, 2026-10-21 사용자 질문)
>
> | | 토픽 키 | kind 이름 |
> |---|---|---|
> | 무엇 | 받는 소켓 묶음의 키(Hub 장부) | 메시지 하나의 종류(`protocol.MsgType`) |
> | 예 | `TILES:7`(계정 7) | `TILES:UPDATE`·`NODE:CREATE`·`ACTIVE:UPDATE` |
> | 브라우저로 가나 | 안 감 — 서버 안에만 | 감 — 프론트 `on('...')` |
> | 바꾸면 | Go만 | Go + 프론트 |
>
> 소켓-1(계정 토픽 이름) = 토픽 키 / 소켓-8(kind 이름 접두어 제각각) = kind 이름

## 정리 후보 목록

| 번호(이름) | 지금 | 정리안 | 영향 |
|---|---|---|---|
| 소켓-1(계정 토픽 이름) | 토픽 이름은 `TILES:<userId>`인데 실리는 건 7종(node·active·공유·보드 목록 포함). 처음 타일 동기화용으로 만든 토픽에 계속 얹어 온 결과 | `USER:<userId>` / `userTopic`으로 이름 바꿈. 토픽 키는 서버 안에만 있어 프론트 무관 | Go만, 작음 |
| 소켓-2(여러 계정에 보내기 중복) | "계정 목록 for -> `Publish(tilesTopic(u))` -> 실패 로그" 루프 3곳(`publishNodeTo`·`publishActive`·`reconcileAccess`) + 한 계정 Publish+로그 6곳(tiles·pins·notice·workspaces·share·active) | `publishUser(userID, kind, payload)` / `publishUsers(users, ...)` 한 쌍으로 모음 | Go만 |
| 소켓-3(볼 수 있는 계정 판정 중복) | 단계 행렬 조회 후 `> PermBlock` 거르는 루프 4벌(`nodeRecipients`, `activeVisibility` process·node 두 갈래, `applyAccess`) | `visibleUsers(users, target)` 하나(target = uid 또는 nodeId) | Go만 |
| 소켓-4(탭 하나로 보내기) | `TAB:ID`·`PROCESS:SIZE_OWNER`는 Hub를 안 거치고 `conn.Emit` 직접. 세 번째 전달 경로인데 이름이 없음. 실제 대상은 `emitSizeOwner` 1곳(`TAB:ID`는 conn을 이미 쥐고 보냄) | `emitTab(tabID, kind, payload)`로 경로 이름을 붙임(토픽화는 안 함). **소켓-2(여러 계정에 보내기 중복) 헬퍼 세트에 합치기 추천**(2026-10-21) | Go만, 작음 |
| 소켓-5(커밋 뒤 발행 규약 섞임) | 함수가 `AfterCommit`을 스스로 거는 것(`publishWorkspacesAfterCommit`) / 호출자가 거는 것(`publishNode`·`publishTiles`·`reconcileAccess`) / 일부러 커밋 전 발행(resize 밀림-가(먼저 발행)) 셋이 섞임 | 규약 하나: "`publish*`는 커밋 뒤 호출자가 `AfterCommit`으로" + 예외만 이름·주석에 표시 | Go만 |
| 소켓-6(process 구독 지점 흩어짐) | `PROCESS:<uid>` 구독·해지가 3파일 4곳(`subscribeLiveProcesses`·`subscribeViewers`=subscribe.go / `applyAccess`=shareAccess.go / `cleanupProcessTopic`=process.go). `accessMu` 잠금은 둘만(의도, 주석 있음) | 한 파일로 모으고 잠금 규칙 주석을 한 곳에 | Go만, 동작 같음 |
| 소켓-7(빈 패키지·낡은 주석) | `internal/supervisor/subscribe`·`internal/supervisor/registry` = `doc.go`만(없는 `*transport.Client`, 안 한 계획). `publishProcess`·`MsgProcessUpdate` 주석 "첫 사용처 resize, worker 확인 뒤에만 발행"이 지금(kill도 씀, 먼저 발행 있음)과 다름. MEMORY 디렉토리 표에도 두 패키지가 남음 | 빈 패키지 삭제 + 주석 고침 + MEMORY 디렉토리 표 갱신 | 작음 |
| 소켓-8(kind 이름 접두어 제각각) | `TILES:*`·`NODE:*`·`ACTIVE:UPDATE`·`SHARE:CHANGE`·`WORKSPACES:UPDATE`(확정 이름은 "보드") / `DATA`·`STATUS`는 worker 프로토콜 이름 그대로 중계 | 바꾸려면 프론트 `on('...')` 문자열도 같이 바뀜. 정리 효과 대비 비용이 커서 **보류 추천** | Go + 프론트 |
| 소켓-9(worker·presence push 없음) | 둘 다 polling | 정리가 아니라 기능 변경 — 파트(리팩터링) 범위 밖, 다음 파트 후보로만 | — |

### 목적·기준으로 다시 본 것
| 번호(이름) | 읽기 효과 | 성능 |
|---|---|---|
| 소켓-1(계정 토픽 이름) | 큼 — 이름이 내용과 다름 | 같음 |
| 소켓-5(커밋 뒤 발행 규약 섞임) | 큼 — 규칙을 모르면 새 발행처에서 롤백 누설 | 같음 |
| 소켓-6(process 구독 지점 흩어짐) | 큼 — 구독이 어디서 바뀌는지 한 파일에서 보임 | 같음 |
| 소켓-2(여러 계정에 보내기 중복) | 중간 | 같거나 조금 나음 — 같은 payload를 여러 계정에 보낼 때 직렬화를 1번만(지금은 계정마다 `Publish` 안에서 다시 직렬화). active는 계정마다 payload가 달라 해당 없음 |
| 소켓-3(볼 수 있는 계정 판정 중복) | 중간 | 같음 — 쿼리 수 그대로 |
| 소켓-7(빈 패키지·낡은 주석) | 중간 — 주석이 동작과 다르면 오해 | 같음 |
| 소켓-4(탭 하나로 보내기) | 작음 | 같음 |
| 소켓-8(kind 이름 접두어 제각각) | 중간 | 같음 — 비용(프론트 동시 수정) 때문에 보류 추천 유지 |

- 성능이 떨어지는 후보는 없음. 소켓-2(여러 계정에 보내기 중복)만 성능이 오를 여지가 있음

- 추천 묶음: 소켓-1(계정 토픽 이름)·소켓-2(여러 계정에 보내기 중복)·소켓-3(볼 수 있는 계정 판정 중복)·소켓-4(탭 하나로 보내기)를 한 묶음(발행 헬퍼), 소켓-5(커밋 뒤 발행 규약)·소켓-6(process 구독 지점)을 한 묶음(규약), 소켓-7(빈 패키지·낡은 주석)은 아무 묶음에 끼움
## 1차 범위 계획 (2026-10-21 R2(범위 결정) — 사용자 "네 제안대로")
범위 = 소켓-1(계정 토픽 이름)·소켓-5(커밋 뒤 발행 규약 섞임)·소켓-6(process 구독 지점 흩어짐). 헬퍼 세트(소켓-2(여러 계정에 보내기 중복)·소켓-3(볼 수 있는 계정 판정 중복)·소켓-4(탭 하나로 보내기))·소켓-7(빈 패키지·낡은 주석)은 다음 묶음. **동작은 하나도 안 바꾼다.**

| 번호(이름) | 파일 | 바꾸는 것 |
|---|---|---|
| 소켓-6(process 구독 지점 흩어짐) | `subscribe.go`로 모음 | `processTopic`(process.go) · `applyAccess`(shareAccess.go) · `cleanupProcessTopic`(process.go) 를 `subscribe.go`로 옮김. `accessMu` 잠금 규칙(누가 잠그고 왜 `subscribeViewers`는 안 잠그는지)을 한 곳에 |
| 소켓-1(계정 토픽 이름) | `subscribe.go` + 호출처 8파일 | `tilesTopic`(tile.go) -> `userTopic`, 키 `TILES:%d` -> `USER:%d`, 정의를 `subscribe.go`의 `processTopic` 옆으로. 호출처: tile·node·active·tileNotice·workspace·shareAccess·subscribe. `messages.go`의 "`TILES:<userId>` 토픽" 주석도 |
| 소켓-5(커밋 뒤 발행 규약 섞임) | `tx.go` + publish 함수들 | 규칙을 `AfterCommit` 주석 한 곳에: "발행은 호출자가 `AfterCommit`으로, 이름이 `…AfterCommit`이면 스스로 건다, 예외 = resize 먼저 발행(밀림-가(먼저 발행))". 함수마다 반복된 "커밋 뒤에만 부를 것"은 지움. `publishWorkspacesAfterCommit`은 이름이 이미 규칙을 말하므로 유지 |

- **코드는 Claude가 고침 · 이름만 바뀌는 파일(tile.go·messages.go 등)도 주석 정리 대상**(2026-10-21 사용자)
- 주석 정리: 손댄 파일 11개(위 + process·processApi·tx·messages) 전체를 한 줄 기준으로. 코드 변경과 **커밋 단계를 나눠** diff를 따로 본다(코드 이동 diff에 주석 diff가 섞이지 않게)
- 확인: `go build`·`go vet` + 격리 실행에서 계정 토픽 7종(`TILES:UPDATE`·`TILES:PINS`·`TILES:NOTICE`·`NODE:*`·`ACTIVE:UPDATE`·`SHARE:CHANGE`·`WORKSPACES:UPDATE`)이 그대로 오는지 + 공유 걸기·풀기 뒤 출력 구독(`applyAccess`)·process 종료 뒤 구독 정리
- vault: `REF-realtime.md`·MEMORY 실시간 push 줄의 `TILES:<userId>` -> `USER:<userId>`

### 1차 범위 결과 (2026-10-21, 커밋 전)
| 단계 | 한 것 |
|---|---|
| 소켓-6(process 구독 지점 흩어짐) | `processTopic`·`cleanupProcessTopic`(process.go)·`applyAccess`(shareAccess.go) -> `subscribe.go`. `accessMu` 잠금 규칙 = `subscribe.go` "process 토픽 구독" 머리 주석 한 곳 |
| 소켓-1(계정 토픽 이름) | `tilesTopic` -> `userTopic`, 키 `USER:%d`, 정의는 `subscribe.go`. `messages.go` 주석 7곳 |
| 소켓-5(커밋 뒤 발행 규약 섞임) | 규칙 = `tx.go` `AfterCommit` 주석. 스스로 감싸던 `afterShareChange` -> `reconcileShareAfterCommit`(이름 규칙에 맞춤). 함수마다 있던 "커밋 뒤에만" 반복 삭제 |
| 주석 정리 | 13파일(active·node·process·processApi·share·shareAccess·subscribe·tile·tileNotice·tx·workspace + messages, 연관 주석만 supervisorRouter.go 1줄). `messages.go`의 중복 Placeholder 설명·"execute는 Linux 전용"·"PROCESS:UPDATE 첫 사용처 resize" 같은 낡은 주석도 이때 정리(소켓-7(빈 패키지·낡은 주석) 중 주석 부분) |
| 확인 | `go build`·`go vet` + 격리 소켓 확인 32/32(변경 전 HEAD 바이너리도 32/32 — 같은 결과). 스크립트 명세 → `REF-test-scripts-1020.md` "2026-10-21 소켓 알림 리팩터" |

- 동작 변경 없음. `git diff --stat`(apps/core) = 13파일 +216 -364. 2026-10-21 재시작 뒤 주석 뺀 diff를 다시 봄: 바뀐 코드 = 이름 2개(`tilesTopic`->`userTopic`·`afterShareChange`->`reconcileShareAfterCommit`) + 함수 3개 본문 그대로 이동, 옛 이름 잔존 없음, 프론트는 kind 이름만 받음
- **커밋 = 브랜치 `refactor/socket` 한 커밋**(코드 이동 + 주석). 계획은 두 커밋으로 나누는 것이었으나 코드만의 patch(scratchpad)가 PC 꺼짐으로 사라져 합침
- 남은 것: 다음 묶음 = 헬퍼 세트(소켓-2(여러 계정에 보내기 중복)·소켓-3(볼 수 있는 계정 판정 중복)·소켓-4(탭 하나로 보내기)) + 소켓-7(빈 패키지·낡은 주석)의 빈 패키지 삭제

- 회귀 확인: 공유·node 동기화·active 배지·크기 우선권 스크립트 명세 재사용(`REF-test-scripts-*`)
