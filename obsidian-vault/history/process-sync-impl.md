# HISTORY — process 동기화 1단계 구현 · 실행 확인

> `history/process-sync.md`에서 분할(2026-10-01). 구현·확인 상세 → `REF-process-sync-impl.md` / 설계 이력 → `history/process-sync.md` / 현재 진행 → `CURRENT.md`.

---

## 2026-10-01(6) — sid 충돌 수정 + 커밋 전 사용자 검토로
- 사용자: sid 충돌 "지금 고치고", 커밋은 "내가 한번 살펴보고" → `user.go` `signIn`에 `_nonce` 추가(build/vet 통과). 사용자 재기동 후 확인: 연속 로그인 20쌍 쿠키 동일 0(수정 전 11), 탭 확인 400/404 정상.
- 커밋 전 사용자 검토 목록 제시(대화) — 검토 후 커밋.

## 2026-10-01(5) — 2단계 실행 확인 (미커밋)
- 사용자가 supervisor·worker 직접 재기동(`00005` 적용 확인, `process_subscribers` 없음).
- 1부(탭 id·계정 구독·크기 우선권·kill·목록): 첫 실행에서 2개 실패 → 원인 = 테스트의 두 로그인이 같은 초라 **쿠키(sid)가 동일** → sid 다르게 재실행해 통과. 그 실패 탓에 의도치 않게 실행된 process 1개는 kill로 정리.
- 2부(socat 프록시 + 별도 SQLite 테스트 worker): X·W 통과. Y는 테스트 worker를 못 죽여(pkill 패턴 불일치) 두 개가 된 테스트 결함 → 둘 다 종료 후, 남은 PENDING process로 Y만 다시 확인해 통과.
- 헤드리스 크롬(playwright-core 재설치)으로 1-g 확인 통과.
- 정리: 테스트 worker·socat 종료, 테스트 노드 삭제, live process 0.
- 발견: sid = 쿠키 원본값 충돌(→ `REF-process-sync-impl.md` "발견", 수정 여부 사용자 결정 대기).

## 2026-10-01(4) — 1단계 코드 작성 (미커밋, 실행 확인 전)
- 세션 시작 vault 파악 → ⑭(전체보기 버튼 최종 위치) 닫음, 스크립트 편집 저장 버튼 = 저장하고 계속 편집(1-g 때 같이 작성, 커밋은 따로) → 사용자 "진행해".
- **1-a(DB 정리)**: `00005_drop_process_subscribers.sql` + `processSubscribers.sql`·gen 삭제(쿼리 파일은 `git rm` staged) + `ListLiveProcessesByOwner` → sqlc 재생성.
- **1-b(탭 등록부)**: `router/tabs.go` 신규, `MsgTabID`(`TAB:ID`)·`TabIDEvent`.
- **1-c(계정 구독)**: `subscribeLiveProcesses`(소켓 연결)·`subscribeUser`(exec·Rebind). `ProcessManager` 구독 메서드·`ProcessEntry.subscribers`·`Subscriber` 삭제, `ListLive`·`Detach` 추가.
- **1-d(크기 우선권 탭 단위)**: `sizeOwner.go` 재작성(uid → 탭 id, 후보 = 소유 계정의 연결된 탭 연결 순서, `processOwner`로 계정 조회). `leaveSession` 삭제 → signOut은 `sizeOwnerOnSignOut`만.
- **1-e(목록 API)**: `GET /processes`. 구독/해지 REST 2개·`/subscriptions` 삭제. kill·resize에 소유 계정 검사 추가.
- **1-f(worker 끊김·재접속)**: X = PENDING `PushStatus` + `AgentInteractive.Detach()`(+`Detached()`) + `ProcessManager.Detach`, Detach로 끝난 relay는 토픽 정리 생략. W = Rebind 뒤 `subscribeUser`. Y = `publishLost`.
- **1-g(프론트)**: `tabId.util.ts` 신규, 소켓 URL 함수화, axios `X-Tab-Id` 인터셉터, App.vue `bindTabId`, `listProcesses`, 실행 버튼 `tabReady` 조건, `ScriptEditDialog` 저장 버튼 닫지 않음.
- 검증: `go build ./...`·`go vet`·`npm run type-check` 통과. 구조표에 없던 결정·한계 → `REF-process-sync-impl.md` "1단계 구현 메모".

