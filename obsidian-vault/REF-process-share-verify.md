# REF — 공유-g(확인): 5 공유 통합 확인

> 2026-10-12 신설. 공유 결정·작업 순서 → `REF-process-share.md` / 단계별 구현·개별 확인 → `REF-process-share-authz.md`(공유-b) · `REF-process-share-api.md`(공유-c) · `REF-process-share-realtime.md`(공유-d) · `REF-process-share-display.md`·`REF-process-share-owner.md`(공유-e) · `REF-process-share-presence.md`(공유-f) · `REF-process-share-terminal-place.md`(터미널 위치·폴더 uid) / 이력 → `history/process-share-verify.md` / 현재 → `CURRENT.md`

**상태(2026-10-12): 통합 API 70/70 + 터미널 위치 회귀 21/21 + 헤드리스 화면(⑥ "나") 통과. ⑥(실행한 사람 표시) "나" 수정 포함. 커밋 `9cee192` → 파트(5 공유) 마감(→ `INDEX-process-share.md`).**

## 방법
- 격리 실행: 임시 DB `nexus_of_tmp` + supervisor 5051 + worker `of-dev` + 격리 프론트 3001(`vite.iso.mts`) → `REF-test-env.md` "격리 실행". 끝나고 전부 삭제
- 스크립트(세션 scratchpad, 재부팅 시 사라짐): `g_test.py`(통합 70, `of_test.py`의 `Client`를 `exec`) · `of_test.py`(터미널 위치·폴더 uid 21) · `ui_prep.py` + `pw/ui.mjs`(헤드리스, playwright-core 1.56.1)
- 계정 3개: A(소유자) · B(받는 쪽) · C(공유 없음). 트리 = A의 F(of-dev) > {s(cat), sub > s2(cat)} / A의 priv > ps / B의 bf

## 항목 (G1 = 공유-b 인가 / G2 = 공유-c API / G3 = 공유-d 실시간 / G4 = 공유-e 표시 / G5 = 공유-f presence)
| 묶음 | 확인한 것 | 개수 |
|---|---|---|
| G1 인가 | 공유 없음 404(노드·목록·exec) / 보기(10): 조회 200·이름 변경·생성·exec·`/workers`·삭제 403, process 목록에 보임(permission 10 + ownerNickname), snapshot 200, kill·resize 403, **입력 버림** / 조작(20): exec 200(소유 = A, startedBy = B), 입력 전달, EDIT 403, 이름 변경 403, 자기 실행 kill 200 / 편집(30): 이름 변경·생성(소유 = A, SH8(받은 쪽이 만든 노드의 소유))·EDIT 200, 다른 소유자 폴더로 이동 403, 루트로 이동 403, 사이클 400, 만든 노드 삭제 200 / 차단(0): 하위 404·목록에서 숨김·아래 차단 있는 상위 삭제 403 / C snapshot 404 | 33 |
| G2 공유 API | lookup 자기 400·없음 404 / 스크립트 공유 400(SH1(공유 단위)) / 자기에게 400 / 단계 15 → 400 / 소유자 아님 403 / 없는 계정 404 / 공유 목록: 직접·상위에서 이어짐(출처 노드) / 소유자 아님 목록 403 / 풀면 상위 공유 행 반환 / process 단독 공유 PUT·목록·DELETE, 끝난 process 400(SC6(process 단독 공유 대상)) / `/shares/mine` NODE·PROCESS / `shareCount` 1(3b(행 공유 개수)) / 복사: 받는 쪽 루트에 소유 40, 차단 하위 제외 | 22 |
| G3 실시간 | 공유·해제 때 `SHARE:CHANGE` / 보기 단계도 출력 수신 / 해제 뒤 snapshot 404 + 출력 수신 멈춤 / 다시 공유 → 출력 수신 재개 | 6 |
| G4 표시 | `startedByMe`: exec 응답·`/processes/lookup`(소유자 쪽은 없음)·`/nodes/:id/runs` 양쪽 | 4 |
| G5 presence | 보기 단계 presence 200(A·B), 탭 id·sid 원문 없음(key 8자), C 404, `/users/presence`는 내가 공유한 계정만, online | 5 |
| 터미널 위치 | `of_test.py` 회귀(폴더 uid FU1·FU2, B2 보드, uid 우선·nodeId 대체, 내 보드에 열기, 다시 공유 중복 없음, 노드 이동, 폴더 exec 400) | 21 |
| 화면 | 헤드리스: B(실행한 사람) 머리 줄 "실행: 나"(title "내가 실행")·실행 목록 "· 나" / A(소유자) 머리 줄·목록 = B 닉네임 | 2 |

- 테스트 쪽 실수 2건(코드 문제 아님): resize는 `X-Tab-Id`가 먼저 필요(없으면 400) / 공유 DELETE 응답은 `{entry}`로 감쌈

## ⑥(실행한 사람 표시) "나" 수정 (2026-10-12, 사용자 "내 터미널에서 내 닉네임이면 나로")
- 닉네임은 유일하지 않음(`users.nickname` 인덱스만) → 프론트에서 닉네임 비교 대신 **서버가 계정 id로 판정한 `startedByMe`** 를 내림
| 자리 | 변경 |
|---|---|
| `query/shares.sql` `ListVisibleLiveProcessesInFolder` | `p.started_by` 열 추가(sqlc 재생성) |
| `nodeRuns.go` `runResponse` | `startedByMe`(omitempty) = `row.StartedBy == 세션` |
| `processDto.go` `processResponse` | `startedByMe`(omitempty) — `withProcessNames`에서 startedBy를 채울 때 같이 |
| 프론트 `process.api.ts`·`node.api.ts` | 타입 `startedByMe?` |
| `processTerm.store.ts` | 이벤트로 덮지 않는 계정 기준 값에 `startedByMe` 추가(PROCESS:UPDATE·`refreshAccount`) |
| `TileFrame.vue` · `ScriptRunList.vue` | `startedByMe`면 "나"(title "내가 실행") |
- 소유자가 자기 스크립트를 실행하면 `startedBy` 자체가 비어 머리 줄에 안 나옴(기존 규칙 그대로). 실행 목록(`/runs`)은 늘 실행한 사람을 보여 주므로 그때도 "· 나"
- 같이 정리: `openFolder` 제거 뒤 남은 "folder" 주석 3곳(`input.go`·`process.go`)

## 확인 못 한 것
- 60초 보관(`tabGrace`) 만료·크기 우선권 넘김(기다려야 함), 여러 탭 질의 시퀀스 중복 응답(미해결 이슈로 이월)
- 루트 타일 막힘 "기본으로 돌아가기"·공유 걸기 창 UI는 2026-10-08 사용자 확인 통과분 그대로(이번엔 다시 안 봄)
