# REF: 리팩터링 — supervisor 소켓 알림 정리 묶음(계획·결과)

> → 이력 `history/refactor-socket-batches.md` / → 후보 목록·목적·기준 `REF-refactor-socket.md` / → 현재 `CURRENT.md`
> 2026-10-21 `REF-refactor-socket.md`에서 분할(10k자 초과). 번호(이름)는 그 파일 "정리 후보 목록" 기준

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

## 2차 범위 계획 (2026-10-21, 사용자 "A 순서" = 소켓 2차 먼저 -> 다른 영역 R1(목록 조사)) — 사용자 "그대로 진행"
범위 = 소켓-2(여러 계정에 보내기 중복)·소켓-3(볼 수 있는 계정 판정 중복)·소켓-4(탭 하나로 보내기)·소켓-7(빈 패키지·낡은 주석)의 빈 패키지. **동작은 안 바꾼다**(로그 문구만 통일)

| 번호(이름) | 새 것 (위치) | 바꾸는 곳 |
|---|---|---|
| 소켓-2(여러 계정에 보내기 중복) | `publishUser(userID, kind, payload)`·`publishUsers(users, kind, payload)` (`subscribe.go`, 토픽 정의 옆) | 한 계정 Publish+로그 6곳(tiles·pins·notice·workspaces·share·active 루프 안) + 계정 루프 2곳(`publishNodeTo`·`reconcileAccess`). `publishShareChange`는 호출 1곳뿐이라 `publishUsers`로 흡수 |
| 소켓-3(볼 수 있는 계정 판정 중복) | `nodeVisibility(ctx, q, users, ids)`·`processVisibility(ctx, q, users, uids)` = 계정 -> 대상 -> 보기 이상 여부 표 (`perm.go`) | `nodeRecipients`·`activeVisibility` 두 갈래·`listActive`의 표 만들기 2벌. `applyAccess`도 같은 표로(구독/해지 둘 다 필요해 false 칸도 씀) |
| 소켓-4(탭 하나로 보내기) | `emitTab(tabID, kind, payload)` (`subscribe.go`) | `emitSizeOwner` 본문 |
| 소켓-7(빈 패키지·낡은 주석) | — | `internal/supervisor/subscribe`·`registry` 삭제(`doc.go`만, import 0). `bind`는 실제 코드라 유지. MEMORY 디렉토리 표 갱신 |

- 그대로 두는 것: `shareAccess.go` 전후 단계 비교(bool 아닌 단계 값 필요), `tile.go` `visibleAncestors`(계정 1명 × 조상 한 줄씩, 모양이 다름)
- 성능: 쿼리 수 그대로. 소켓-2의 "직렬화 1번" 이득은 Hub(재사용 코어)에 여러 키 발행을 넣어야 생기는데, 실제 이득이 작아(접속 계정 수만큼 json 1회) **Hub는 안 고치는 쪽 추천**
- 로그: `[node]`·`[tiles]`·`[active]`·`[access]`·`[sizeOwner]` 제각각 -> `[push] user=… kind=…` / `[push] tab=… kind=…`로 통일(kind가 출처를 말함)
- 확인: build·vet + 격리 소켓 32/32 재실행(스크립트는 사라져 `REF-test-scripts-1020.md` 명세로 다시 작성) + `GET /tabs/active` 확인 1개 추가(`listActive`가 이번에 바뀜)


### 2차 결과 (2026-10-21, 커밋 전)
| 번호(이름) | 한 것 |
|---|---|
| 소켓-2(여러 계정에 보내기 중복) | `publishUser`·`publishUsers`(`subscribe.go`). `publishNodeTo`·`publishShareChange` 삭제, tiles·pins·notice·workspaces·active 발행이 헬퍼 호출 |
| 소켓-3(볼 수 있는 계정 판정 중복) | `visibility[K]` + `nodeVisibility`·`processVisibility`(`perm.go`). active는 그 위에 `activeSight`·`loadActiveSight`(`active.go`) — `activeVisibility` 삭제, `listActive`의 표 2벌도 같은 것으로. `nodeRecipients`·`applyAccess`도 표 사용 |
| 소켓-4(탭 하나로 보내기) | `emitTab`(`subscribe.go`), `emitSizeOwner` 삭제 |
| 소켓-7(빈 패키지·낡은 주석) | `internal/supervisor/subscribe`·`registry` 삭제, MEMORY 디렉토리 표 갱신 |

- 동작 변경 없음(로그 문구만 `[push] user=… kind=…`/`[push] tab=… kind=…`). `applyAccess`는 행 순회 -> 계정·탭·uid 순회(같은 구독 결과, `connsOfUser` 호출이 계정당 1번으로 줆)
- 확인: build·vet·gofmt + 격리 소켓 확인 **35/35**(32개 + `GET /tabs/active` 보임·안 보임 3개, 변경 전 1차 커밋 `8cf32bc` 바이너리도 35/35). 격리 프로세스·임시 DB 정리 확인
- 소켓 영역 남은 것 = 소켓-8(kind 이름 접두어 제각각, 보류)·소켓-9(worker·presence push 없음, 범위 밖)뿐. 다음 = 다른 영역 R1(목록 조사)