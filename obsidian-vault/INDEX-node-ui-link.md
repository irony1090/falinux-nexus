# INDEX — node UI 연동 (마감)

> 2026-10-06 색인 2단화로 MEMORY·HISTORY에서 접은 파트(사용자 확인). **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
레이아웃(→ `INDEX-node-ui-layout.md`) 위에서 더미를 실제 API·소켓으로 바꾼 작업, 2026-09-30~10-02: 연동 1차(`GET /workers`·폴더 목록·node 관리 UI·스크립트 편집 창) / 6 타일 터미널 임베드 + 크기 우선권(J~N, `ProcessDialog` 삭제) / ⑪(타일 트리 서버 저장, `tile_trees` + `GET/PUT /tiles`, 3-a~3-e) / 4 O(계정 동기화) 화면(`TILES:UPDATE` push, 4-a~4-c). `REF-node-ui.md`(컨셉)·`REF-node-ui-overview.md`(⑮(모바일 가상 키보드) 진행 중)는 MEMORY에 남김.

**남은 것(CURRENT로 옮겨 둠)**: 4⑤(폴더 목록 동기화) = 열린 질문 4(`NODE:` 동적 구독).

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- 설계/재사용 지식: `REF-node-ui-link.md` `REF-node-ui-terminal.md` `REF-node-ui-save.md` `REF-node-ui-save-impl.md` `REF-node-ui-sync.md`
- 작업 이력: `history/node-ui-link.md` `history/node-ui-terminal.md` `history/node-ui-save.md` `history/node-ui-save-impl.md` `history/node-ui-sync.md`
- 참고: `REF-node-ui-terminal.md`의 M(크기 우선권) 규칙은 `REF-process-sync.md` Q(크기 우선권 단위)가 이어 씀

## 모듈 현황표 행 (MEMORY 원문, 접기 직전)
| 모듈 | 상태 | 문서 |
|------|------|------|
| Node 카탈로그 | DB+CRUD+핸들러+PatchNode 커밋(9c9d22e) + **`GET /workers`(접속 인스턴스, `?nodeId=` 상속 해석) + exec 대상 장비 검증**(2026-09-30, HTTP e2e 통과, 커밋 `ba5b857`). **⑪(타일 트리 서버 저장) 서버 쪽 완료**(`tile_trees` + `GET/PUT /tiles` + exec 때 터미널 타일 서버 추가, 커밋 `2aa0920`, 2026-10-02) + 3-c(스토어)·3-d(터미널 복원) 완료(`2f0170f`) — `REF-node-ui-save-impl.md`. **4 O(계정 동기화) 화면 완료**(`TILES:UPDATE` push + 다른 탭 터미널 등록 일원화, `4b6bbcd`) — `REF-node-ui-sync.md`. roster/label 남음 | `REF-node-label.md`(+`REF-node-ui.md` UI 컨셉 / `REF-node-ui-overview.md` 전체보기·가상 키보드 / `REF-node-ui-link.md` 실제 연동 / `REF-node-ui-terminal.md` 타일 터미널·크기 우선권 / `REF-node-ui-save.md`(+`-impl`) 타일 트리 서버 저장 ⑪ / `REF-node-ui-sync.md` 4 O(계정 동기화) 화면 / 레이아웃 설계·구현 1~5 → `INDEX-node-ui-layout.md`) |
| 프론트엔드 (apps/frontend) | 스캐폴딩·socket hook·user/login WIP·전역 다이얼로그 + node 타일 UI(레이아웃 설계·구현 1~5 마감 → `INDEX-node-ui-layout.md`) + 연동 1차·타일 터미널·터미널 입력 완료. 4.5 SNAPSHOT(화면복원) 완료. 남은 것 = 4.6 EDIT(worker `vi` 편집) 프론트 연결·⑮(모바일 가상 키보드) | `REF-frontend.md`(+`REF-node-ui.md` 컨셉 / `REF-node-ui-overview.md` 전체보기·가상 키보드 / `REF-node-ui-link.md` 연동 / `REF-node-ui-terminal.md` 타일 터미널 / `REF-node-ui-save.md` 서버 저장) |

## 이 파트에만 해당하는 불변 결정
- 없음 — 색 규칙(hex 금지)·process 동기화 범위 = 계정 등은 다른 파트에서도 지켜야 해서 MEMORY에 남김

## HISTORY 색인 (원문)
### `history/node-ui-link.md` — node 카탈로그 타일 UI 실제 연동 (더미 → node/process API, 2026-09-30 신설)
- 2026-10-01 스크립트 편집 창 저장 동작 통일 결정(저장 버튼도 저장하고 계속 편집, 코드 미반영)
- 2026-09-30(4) 연동 1차 4(스크립트 편집): `ScriptEditDialog`(브라우저 편집 → `PATCH content`, 새 스크립트 직후 자동 열기), 창 높이는 `VDialog` `height`로
- 2026-09-30(3) 폴더 안(브레드크럼 줄)에서 장비 지정·접속 상태 표시(상속 포함) + 헤드리스 시나리오 확인, 테스트 로그인은 `Login.vue` 기본값 사용 허용
- 2026-09-30(2) 연동 1차 3(node 관리 UI): 생성·이름 변경·장비 지정(직접 입력 + 접속 상태)·삭제(하위를 보던 타일 닫기), `hook/nodeRemove.hook.ts` 신규
- 2026-09-30 연동 1차 2(폴더 타일 목록): 더미 목록·경로·이름을 node API로(`useListChildren`/`useNodePath`/`useGetNode`), 경로 = `getNode` 반복 호출, 실행은 더미 유지

### `history/node-ui-terminal.md` — node 타일 터미널 임베드 + 크기 우선권 (2026-10-01 신설)
- 2026-10-01(3) 설계 개정 논의 → `history/process-sync.md`
- 2026-10-01(2) 구현: 크기 우선권 백엔드(`sizeOwner.go`, signOut 확장) + `processTerm.store`·`ProcessTerminal`·`WorkerPickDialog` + 타일 연결 + `ProcessDialog` 삭제 (미커밋, 실행 확인 전)
- 2026-10-01 구조 합의 J~N(xterm 스토어 소유·상태 단일 구독·실행 흐름·크기 우선권 = 실행한 세션·`ProcessDialog` 삭제), 코드 없음

### `history/node-ui-save.md` — node 타일 트리 서버 저장 ⑪ 설계·3-c 구조안 (2026-10-01 신설)
- 2026-10-02(6) 3-d 구조안 + 3-d①(끝난 process 조회 = 나누기, 반복 파라미터)·②~④ 확정
- 2026-10-02(5) 3-c 열린 항목(닫기 직전 저장) 확정 = `visibilitychange` hidden 즉시 keepalive 저장
- 2026-10-02(4) 서버 쪽 커밋(`2aa0920`) + 3-c(스토어) 구조안 제시(사용자 확인 전) + REF/history `-impl` 분할
- 2026-10-01 구조안 제시 + 결정 ⑪-1(저장 형태)·⑪-2(충돌 처리)·⑪-3(새로고침 뒤 터미널 화면)·⑪-4(타일 없는 실행 중 process) 확정, 코드 없음

### `history/node-ui-save-impl.md` — node 타일 트리 서버 저장 구현 3-a~3-d (2026-10-02, node-ui-save에서 분할)
- 2026-10-02(5) 3-d(터미널 복원) 작성(lookup API·restore·안전망) + 함정 2개(provide 자신 inject 불가·복원 타일 attach 누락) 수정, 확인 통과, 커밋 `2f0170f`
- 2026-10-02(4) 3-c(스토어) 작성(op 큐·409 재적용·keepalive·exec 빠른 경로) + 헤드리스 8항목 통과(커밋 `2f0170f`)
- 2026-10-02(3) 3-b'(exec 연동) 실패 대처 결정(별도 트랜잭션·kill 보상·3-d 안전망) + 작성 + HTTP 6항목 통과
- 2026-10-02(2) 3-b(API) 작성(`router/tile.go` GET/PUT + 트리 검사) + HTTP 확인·첫 GET 경합 통과
- 2026-10-02 3-a(DB) 작성(`tile_trees` + 쿼리 4개), PG 롤백 실행 확인

### `history/node-ui-sync.md` — 4 O(계정 동기화) 화면: 타일 트리·터미널 실시간 반영 (2026-10-02 신설)
- 2026-10-02(5) 사용자 PC 확인 + 커밋 `4b6bbcd`, 세션 종료 정리
- 2026-10-02(4) 4-c(터미널 등록 일원화) 작성(보관함·restore 일원화·prune) + 탭 2개 확인 통과
- 2026-10-02(3) 4-b(트리 수신) 작성(push 반영·op 멱등·재연결 재동기화) + 탭 2개 확인 통과
- 2026-10-02(2) 4-a(서버 발행) 작성 + 소켓 확인 7항목 통과
- 2026-10-02 구조안 제시 + 4①~4⑤ 제안대로 확정(작업 단위 4-a~4-d), SNAPSHOT과 따로 진행 합의, 코드 없음
