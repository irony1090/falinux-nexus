# HISTORY — node 타일 터미널 임베드 + 크기 우선권

> 타일 UI 6(`ProcessDialog` 타일 임베드) + 연동 1차 5(실행) 작업 이력. 설계·결정 → `REF-node-ui-terminal.md` / 현재 진행 → `CURRENT.md`.

---

## 2026-10-01 — 구조 합의 (코드 없음)
- 세션 시작 시 연동 1차 1~4 커밋: `ba5b857` feat(node) / `8223f36` docs(vault). 제외 유지: `apps/core/cmd/irony/`, `provideAppLayout.vue`.
- 구조안 J~N 제시. 코드 확인 중 발견: `TileStrip`이 Grid마다 별도 `v-for`라 Grid 이동 시 타일 재마운트 → xterm을 컴포넌트가 가지면 화면 소실 → J(xterm 소유자) = 스토어.
- 사용자 추가 요구: 크기 동기화는 **처음 실행한 세션이 우선권**, 로그아웃 또는 일정 시간 재접속 없음이면 다음 세션으로, 백엔드가 소유 세션을 들고 있음 → M(크기 우선권)으로 개정.
- 사용자 선택(전부 추천안): 원래 소유자 복귀 시 되찾지 않음 / 대기 60초 / 비소유자 = 소유자 크기 고정 + 스크롤 / 로그아웃 API 이번 범위.
- N 질문("테스트 코드냐")에 대한 정리: `ProcessDialog`는 7월 단일 창 시제품, 테스트 코드는 `index.vue`의 미사용 함수들 — 둘 다 삭제.
- 발견: 로그아웃 API 없음(오판 — (2)에서 정정: `DELETE /users/session` 이미 있음) / `process_subscribers.sid`가 쿠키 원본이라 소유자 표시를 브로드캐스트하면 안 됨.

## 2026-10-01(2) — 구현 (미커밋, 실제 실행 확인 전)
- 사용자 "추천방향으로 진행해줘" → 로그아웃 시 구독 해지도 추천대로.
- **계획과 달라진 것**: ① 로그아웃 API는 이미 있음(`DELETE /users/session`) → 확장만 ② 소유자를 `ProcessEntry`가 아니라 router 맵에 — `Rebind`가 entry를 새로 만들기 때문 ③ `STATUS`도 exec 응답보다 먼저 올 수 있어 출력과 함께 버퍼링.
- 백엔드: `sizeOwner.go` 신규 + subscribe/process/processApi/processDto/user 배선 + sqlc 쿼리 1개 + `MsgProcessSizeOwner`. `go build ./...` + `go vet` 통과.
- 프론트: `processTerm.store.ts`·`ProcessTerminal.vue`·`WorkerPickDialog.vue` 신규, 타일 3종 연결, `ProcessDialog`·`processDialog.store`·`tileDummy.ts` 삭제, `index.vue` 테스트 함수 정리. `vue-tsc` 통과.
- 확인 대기: supervisor 재기동(새 바이너리) 후 실제 실행·출력·kill·크기 동기화.
- 발견(범위 밖): worker 재접속 뒤 재바인딩된 process의 Hub 구독·메모리 구독자 공백 → REF 구현 메모.

## 2026-10-01(3) — 설계 개정 논의 → `history/process-sync.md`
- 크기 우선권 단위(sid → 탭 id)·계정 동기화·공유 설계 논의. 결정 O~V와 진행 순서는 `REF-process-sync.md`.
