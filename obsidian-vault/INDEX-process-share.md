# INDEX — process 동기화·공유·보드 (마감)

> 2026-10-12 파트(5 공유) 마감 때 색인 2단화로 MEMORY·HISTORY·CURRENT에서 접은 파트(사용자 "파트 마감 진행"). **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
2026-10-01~10-12. 진행 순서(동기화 계열) 1 탭 id + S(구독 역할 분리) → 4 O(계정 동기화) 다음의 마지막 단계 **5 공유(P·R·T) + presence(접속 상태)** 와 그 토대인 동기화 1단계를 묶은 파트.
- 동기화 1단계: 동기화 범위 = 계정, 서버 발급 탭 id, 계정 구독, 크기 우선권(`REF-process-sync*.md`)
- 공유: 공유-a(DB, 판정 함수) → 공유-b(인가 교체) → 공유-c(공유 API) → 작업공간-a~e(보드 = 작업 공간 여러 개) → 공유-d(실시간, 실시간-0~7) → 공유-e(프론트 표시) → 공유-f(presence) → 공유-g(확인)
- 사용자 확인 중 수정: 루트 타일 막힘 "기본으로 돌아가기" / 공유 전에 실행한 터미널 "내 보드에 열기" / 다른 보드로 가는 터미널(`placeTerminal`) + 폴더 타일 uid(FU1(uid 발급원)~FU4(영속)) / 백엔드 `openFolder` 제거 / ⑥(실행한 사람 표시) "나"
- 커밋: 체크포인트 `3b80310`·`d50f153`(태그 `checkpoint/before-folder-tile-uid`) → 병합 `13d630b` → 공유-g `9cee192` + 마감 vault 커밋. push 안 함

**남은 것(CURRENT로 옮겨 둠)**: 숨기기(`node_shares.hidden_at`, 복원 UI "숨긴 공유 N") / 잘라내기(SH10(노드 복사) 확장) / 보드 미리보기 / 단축키 키 조합 / 60초 보관(`tabGrace`) 만료·크기 우선권 넘김 실측 / 여러 탭 질의 시퀀스 중복 응답(미해결 이슈).

## 이 파트의 알아 둘 동작·규칙 (CURRENT에서 옮김)
- 폴더 **자신**에 직접 공유 행이 있으면 그 폴더를 다른 폴더 아래로 옮겨도 접근 유지(가까운 공유 행 기준) — "범위 밖 이동"은 공유 폴더 **안의 하위 노드**를 밖으로 옮길 때만. 하위 폴더에 따로 건 공유가 상위보다 높으면 높은 쪽이 적용(SH3(받은 계정 권한) "가까운 공유", 2026-10-12 사용자 테스트에서 재확인)
- 루트 스크립트는 장비가 없어 실행 불가
- 시안 읽는 법: 각 시안 맨 위 "결론 반영" 상자가 기준, 본문의 "제안"은 미확정. 시안과 REF가 다르면 REF 우선. 시안 링크 → user vault `user/links.md`
- 공유 표시 색 = 청록 `share`(MEMORY 기술 스택 줄에도 남김)

## 이 파트에만 해당하는 불변 결정
- 없음 — "process 동기화 범위 = 계정", "공유 판정 = DB 함수 한 곳", "실시간 push = 인가/라우팅 분리"는 다른 파트에서도 지켜야 해서 MEMORY에 남김

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- 설계/재사용 지식: `REF-process-sync.md` `REF-process-sync-impl.md` `REF-process-share.md` `REF-process-share-node.md` `REF-process-share-db.md` `REF-process-share-authz.md` `REF-process-share-api.md` `REF-process-share-realtime.md` `REF-process-share-realtime-ui.md` `REF-process-share-reconcile.md` `REF-process-share-display.md` `REF-process-share-terminal-place.md` `REF-process-share-verify.md` `REF-workspace.md` `REF-workspace-impl.md` `REF-workspace-server.md` `REF-workspace-front.md` `REF-process-share-owner.md` `REF-process-share-presence.md`
- 작업 이력: `history/process-sync.md` `history/process-sync-impl.md` `history/process-share.md` `history/process-share-node.md` `history/process-share-db.md` `history/process-share-authz.md` `history/process-share-api.md` `history/process-share-realtime.md` `history/process-share-realtime-ui.md` `history/process-share-reconcile.md` `history/process-share-display.md` `history/process-share-owner.md` `history/process-share-presence.md` `history/process-share-terminal-place.md` `history/process-share-verify.md` `history/workspace.md` `history/workspace-impl.md` `history/workspace-server.md` `history/workspace-front.md`

## 모듈 현황표 행 (MEMORY 원문, 접기 직전)
| 모듈 | 상태 | 문서 |
|------|------|------|
| process 실행 (execute/pty/manager/bind) | 배선(exec/kill/resize/재접속/구독) 마감 → `INDEX-process-wiring.md` / 터미널 기능(입력·SNAPSHOT·EDIT(worker `vi` 편집)) 마감(2026-10-06) → `INDEX-process-terminal.md`. 동기화 1단계(탭 id·계정 구독·크기 우선권) 완료. 남은 것 = 5 공유(P·R·T)(공유-a(DB)·공유-b(인가 교체)·공유-c(공유 API)·작업공간-a(DB)·작업공간-b(API)·작업공간-c(exec·안전망·탭) 코드·확인 완료, 작업공간-d(프론트) 코드 완료, 작업공간-e(확인) 통과(2026-10-07) · 2026-10-08 `master` 체크포인트 커밋(분기 지점) → `feat/folder-tile-uid`(다른 보드로 가는 터미널 수정 + 폴더 타일 uid + `openFolder` 제거) 2026-10-12 `master`에 병합(분기 지점 태그 `checkpoint/before-folder-tile-uid`, `REF-process-share-terminal-place.md`), 공유-d(실시간) — 실시간-0(터미널 UI)~실시간-7(확인) 완료 = 공유-d(실시간) 마감, 공유-e(프론트 표시) 0~3c 완료·사용자 확인(커밋 전, `REF-process-share-display.md` `REF-process-share-owner.md`), 공유-f(presence) 완료(`REF-process-share-presence.md`, 커밋 전), 공유-g(확인) 통과(2026-10-12, `REF-process-share-verify.md`) → 파트 마감 커밋 대기. 화면 이름 = "보드") + presence(접속 상태) | `REF-process.md`(+`-sync`/`-sync-impl`/`-share`/`-share-node`/`-share-db`/`-share-authz`/`-share-api`/`-share-realtime`/`-share-realtime-ui`) + `REF-workspace.md`(작업 공간 여러 개·터미널 머리 줄) + `REF-workspace-impl.md`(구현 구조, WS1(트리 주소)~WS7(보는 작업 공간 전달) 확정) + `REF-workspace-server.md`(서버 구현 메모·확인) + `REF-workspace-front.md`(프론트) |

## CURRENT "다음 세션 시작 시" (원문, 접기 직전)
- **2026-10-12 — 브랜치 `feat/folder-tile-uid` 커밋 → `master`에 `--no-ff` 병합 완료, 분기 지점(`d50f153`)에 태그 `checkpoint/before-folder-tile-uid`(되돌릴 땐 이 태그로)**: ③(다른 보드로 가는 터미널) 수정 + 폴더 타일 uid(FU1~FU4, FU2 개정) **사용자 직접 확인 통과**(2026-10-12) + 백엔드 `openFolder` 제거(폴더 exec = 400, 격리 확인 제거 전 23/23·후 21/21). 변경 = 서버 `tile.go`·`share.go`·`shareAccess.go`·`node.go`·`process.go`·`processApi.go`·`perm.go`·`process/{manager,entry}.go` + 프론트 `tileTree.store.ts`·`ShareDialog.vue`·`FolderTileBody.vue`·`share.api.ts` + vault. **공유-g(확인) 통과(2026-10-12: 통합 70/70·위치 21/21·헤드리스 ⑥, `REF-process-share-verify.md`) + ⑥ "나" 수정(커밋 전). 다음 = 사용자 마감 확인 → 파트(5 공유) 마감 커밋 + 색인 접기(HISTORY 10k자 초과)**
- **2026-10-08 세션 끝 상태(정리 3) — 분기 지점**: 아래 공유-a~f + ①② 수정까지 전부 `master`에 **체크포인트 커밋 `3b80310`(코드)·`d50f153`(vault)**(사용자 요청, 파트 마감 아님 — 공유-g(확인) 남음). 이어서 할 작업(③ 수정 + 폴더 타일 uid)은 위험해 보여 **새 브랜치 `feat/folder-tile-uid`** 에서 진행(사용자 "현 지점을 분기로"). 되돌릴 땐 `master`로
- (정리 2 요약) 공유-d(실시간) 마감 + 공유-e(프론트 표시)·공유-f(presence) 완료(사용자 확인 통과), 지금은 체크포인트 커밋됨 → `REF-process-share-display.md` / `REF-process-share-owner.md` / `REF-process-share-presence.md`
- **사용자 확인 중 발견·수정(2026-10-08)**: ① 루트 타일을 공유 폴더 안으로 열었다가 공유가 풀리면 빠져나올 수 없음 → 배너에 "기본으로 돌아가기"(프론트만, **사용자 확인 통과**) → `REF-process-share-realtime-ui.md` "루트 타일 막힘" ② 터미널을 켜 둔 채 폴더를 공유하면 받는 쪽이 터미널을 열 방법이 없음 → 공유 걸 때 터미널 타일도 넣음 + `POST /processes/:uid/tile` + 스크립트 행 "내 보드에 열기"(격리 확인 통과, **사용자 확인 통과 2026-10-12**) → `REF-process-share-terminal-place.md` "공유 전에 실행한 터미널"
- **③ 다른 보드로 가는 터미널 + 폴더 타일 uid — 코드·격리 확인·사용자 확인 통과(2026-10-12, `master`에 병합)**: 위치 찾기 `placeTerminal`(uid 우선 → nodeId → 기본 루트) → `REF-process-share-terminal-place.md`
- **다음 = 공유-g(확인) → 파트(5 공유) 마감 커밋**. 공유-g에 넣을 것: 공유-b·c 회귀(합본 스크립트 재작성) / 3b(행 공유 개수)·3c(내가 공유한 것)·공유-f 시나리오 / 루트 타일을 공유 폴더 안으로 열기 → 해제 → 돌아가기 / 실행 중 터미널이 있는 폴더 공유 → 받는 쪽 터미널 타일·"내 보드에 열기" / 공유 중 폴더 이동·하위로 들어오기 — ③에서 수정·격리 확인. 시안 링크 → `user/links.md`
- **공유 표시 색 = 청록 `share`**(2026-10-08 사용자: 공유 관련만). 새 공유 표시도 `share` 색 사용. **사소한 미결**: ~~⑥ 내 닉네임~~ → "나"로 수정(2026-10-12) / 숨기기(복원 UI "숨긴 공유 N" + `hidden_at`)는 나중 / 단축키·잘라내기(SH10) 나중
- 알아 둘 동작: 폴더 **자신**에 직접 공유 행이 있으면 그 폴더를 다른 폴더 아래로 옮겨도 접근 유지(가까운 공유 행 기준) — "범위 밖 이동"은 공유 폴더 **안의 하위 노드**를 밖으로 옮길 때만. 루트 스크립트는 장비가 없어 실행 불가. `throwCatch`가 이제 `status`(HTTP 상태)를 던짐(401만 골라 처리하는 수정은 아직 — 아래 잔여)
- 시안: "터미널 최소화 시안"(채택 기준, 메인 보드 + 머리 줄 비교 보드) https://claude.ai/artifact/Q2X7sYsDAHPwmVZmE1icgP · 이전 "터미널 다시 열기 시안"(대체됨) → user vault `user/links.md`. 고칠 땐 Artifact `read`로 `project/*.dc.html`을 받아서
- 커밋은 파트(5 공유) 마감 때 한 번(2026-10-07 사용자, 작업 단위마다 안 함). 예외: 2026-10-08 사용자가 위험 작업 전 체크포인트 커밋을 직접 요청
- 재부팅으로 사라지는 것(세션 scratchpad `/tmp/claude-1000/.../scratchpad/`): (2026-10-12 세션 `12185a69…`) `of_test.py`(openFolder 제거 전·후 + 다른 보드·폴더 uid 회귀 18, Client 클래스 포함 단독 실행) · `iso/{sup,wrk}`(DB `nexus_of_tmp`, 삭제함) / (2026-10-08 정리 4 세션, 재부팅으로 이미 사라짐) `fu_test.py`(다른 보드·폴더 uid 28개, 앞 세션 `presence_test.py`의 `Client`를 `exec`) · `fu2_test.py`(FU2 개정 11개) · `ev_test.py`(보드별 `TILES:UPDATE` 2개) · `reg/`(presence·attach를 worker `fu-dev`용으로 바꾼 복사본) · `iso/{sup,wrk}`(DB `nexus_fu_tmp`, 삭제함) / 이전 세션 것(`312eca0a…/scratchpad/pres/`의 presence·attach·ui 스크립트, 격리 프론트 `vite.iso.mts`). **다시 쓸 땐 `REF-test-env.md` "격리 실행"·"2026-10-08 확인 스크립트 요령" 보고 재작성**. 개발 환경은 영향 없음(임시 DB는 전부 삭제함, 5050·3000은 사용자 것)
- 어디에 무엇이:
  | 파일 | 내용 |
  |---|---|
  | `REF-process-share.md` | SH1(공유 단위)~SH10(노드 복사) 결정표(전부 확정), 받는 쪽·소유자 쪽 표시(확정·제안·미정), **작업 단위·순서** |
  | `REF-process-share-node.md` | SH3(받은 계정 권한) 단계·SH10(노드 복사)·클립보드 |
  | `REF-process-share-db.md` (신설) | 공유-a(DB): `00007_shares.sql`·판정 함수·판정 쿼리, SA1(단계 저장 형식)~SA4(기존 행 채우기) 확정·구현·확인 결과. 판정 쿼리 표의 "쓰는 곳" 열 = 공유-b(인가 교체) 작업 목록 |
  | `REF-process-share-authz.md` | 공유-b(인가 교체): 엔드포인트별 요구 단계, SB1(권한 부족 응답)~SB6(이동 사이클 검사), 작업 단위, **구현 메모·확인 결과(2026-10-07)** |
  | `REF-process-share-realtime.md` (신설) | 공유-d(실시간): 범위·지금 코드·SD1~SD9(전부 확정)·지킬 조건·작업 단위 실시간-0~7 |
  | `REF-process-share-reconcile.md` (신설) | 실시간-1~3: `reconcileAccess` 구조안·쿼리·경합·정할 것 |
  | `REF-process-share-realtime-ui.md` (신설) | 실시간-0(터미널 UI): SD3 확정 전문·지금 코드·구현 계획·확인할 것 |
  | `REF-process-share-api.md` (신설) | 공유-c(공유 API): API 표, SC1(출처 계산 위치)~SC7(복사 제외 개수) 확정, **구현 메모·확인 결과** |
  | `REF-workspace-impl.md` (신설) | 작업공간-a~e 구현 구조: 지금 구조·발견(받는 쪽 고아 안전망)·WS1~WS7(확정)·API·작업 단위 |
  | `REF-workspace-front.md` (신설) | 작업공간-d(프론트): 파일별 변경·구현 중 정한 것·확인할 것 |
  | `REF-workspace-server.md` (신설) | 작업공간-a(DB)·작업공간-b(API)·작업공간-c(exec·안전망·탭) 구현 메모·확인 결과·과도기 위험 |
  | `REF-workspace.md` (신설) | 작업 공간 여러 개(`/`·`/w/{hash}`, `/`에서만 생성, W1(분리 범위)~W3(이동 범위)), 터미널 머리 줄 `[실행한 타일 번호] 경로 / 이름`, 작업공간-a~e |
  | `REF-node-ui-overview.md` "active 타일 표시" | 나(약하게 흐림) + 위치 번호 단축키 확정(구현 시점 미정, 공유와 별개) |
  | user vault `user/links.md` | 오늘 시안 4개 링크·읽는 법·시안별 확정/제안 |
- **작업 순서 = C(끼워 넣기)**: 공유-a(DB) → 공유-b(인가 교체) → 공유-c(공유 API) → 작업공간-a~e → 공유-d(실시간) → 공유-e(프론트) → 공유-f(presence) → 공유-g(확인)
- **뒤에 정해도 되는 것**: 숨기기(SA3(숨기기 시점) = 공유-e(프론트) 때, `node_shares.hidden_at` 컬럼 하나) / 잘라내기(SH10 구현 때) / **보드 미리보기**(`/` 루트 타일 안 다른 보드 카드, REF-workspace "희망안") = 나중에(2026-10-07 사용자) — 덧붙이는 기능이라 막는 작업 없음. 단 "보는 보드만 출력 구독"으로 바꿀 거면 구독 로직을 건드리는 **공유-d(실시간) 착수 때 한 번 확인** / 제안 단계 표시 ①~④·⑥·B·O1~O5·단계 낮출 때 열린 터미널(공유-e 때 시안 보며) / 단축키 키 조합
- **시안 읽는 법**: 각 시안 맨 위 "결론 반영" 상자가 기준, 본문의 "제안"은 미확정. 시안과 REF가 다르면 REF 우선
- 진행 순서(동기화 계열): 1 탭 id + S(구독 역할 분리) → 2 확인·커밋 → 2.5 터미널 입력 → 3 ⑪(타일 트리 서버 저장) → 4 O(계정 동기화) 화면 → 4.5 SNAPSHOT(화면복원) → 4.6 EDIT(worker `vi` 편집) **여기까지 완료** → 5 공유(P·R·T)
- 커밋 상태: **2026-10-08 공유-a~f·작업공간-a~e·앱 이름·로고·①② 수정 코드 + vault 전부 `master`에 체크포인트 커밋함(분기 지점)**. 이후 작업은 `feat/folder-tile-uid` → 2026-10-12 `master`에 병합(태그 `checkpoint/before-folder-tile-uid` = 분기 지점). 태그·병합도 push 안 함. 2026-10-02 커밋들 + `97fb5c4` 모두 push 안 함. vault 변경(2026-10-06 공유 구조 합의분)은 커밋함(push 안 함). 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치)
- 테스트·확인 준비(환경, 헤드리스, 함정, 남은 테스트 데이터, **격리 실행 = 임시 DB + 5051**) → `REF-test-env.md`
- 역할 분담(2026-09-29 개정): node 타일 UI·process 작업은 코드 대부분을 Claude가 작성하고, 로직 구조는 사용자와 먼저 맞춘 뒤 쓴다

## HISTORY 색인 (원문)
### `history/process-sync.md` — process 동기화 범위 · 탭 id · 공유 (2026-10-01 신설, node-ui-terminal에서 분리)
- 2026-10-06 presence(접속 상태) 조회 시점 = 5 공유(P·R·T) 구조 합의 때 함께로 합의, 코드 없음
- 2026-10-01(4)(5) 1단계 코드·실행 확인 → `history/process-sync-impl.md`
- 2026-10-01(3) 1단계 작업 단위 1-a~1-g 제시(착수 승인 대기), 코드 없음
- 2026-10-01(2) Rebind 뒤 출력 끊김 실행 확인(socat 프록시로 worker 연결만 끊기) + 끊김 때 FAILED 502 발행 발견, 코드 수정 없음
- 2026-10-01 설계 개정 논의: O(계정 동기화)~V(탭 id 전달) 결정 + 진행 순서 1~5 + 1단계(탭 id + 구독 역할 분리) 구조 합의, 코드 없음

### `history/process-share.md` — process 공유 (5 공유(P·R·T) + presence, 2026-10-06 신설)
- 2026-10-06(2) SH2(받는 쪽 화면 위치) 확정(받는 쪽 `/` 루트) + 수락 없음 + 공유 사라짐 처리 + 받는 쪽 구분 표시·소유자 쪽 공유 목록 시안(제안), 코드 없음
- 2026-10-06 구조안 제시 + 확정 SH1(공유 단위)·SH3(받은 계정 권한)·SH8(받은 쪽이 만든 노드의 소유)·SH9(받은 쪽이 실행한 process의 소유)·SH10(노드 복사)+클립보드 표시, 코드 없음

### `history/workspace.md` — 작업 공간 여러 개 · 터미널 소유 표시 (2026-10-06 신설, 공유 논의에서 파생)
- 2026-10-06 작업 공간 규칙(`/`에서만 생성·`/w/{hash}`·폴더 단위 이동) + W1(분리 범위)=타일 트리 하위·W2(삭제 처리)=`/`로 되돌림·W3(이동 범위)=제한 없음 + 터미널 머리 줄 `[실행한 타일 번호] 경로` 확정, 코드 없음

### `history/workspace-impl.md` — 작업 공간 여러 개: 구현 구조·결정 (2026-10-07 신설)
- 2026-10-07(2) WS1~WS7 확정(코드 착수는 사용자 지시 뒤) + WS5 의미 보충
- 2026-10-07 구현 구조안 제시(WS1(트리 주소)~WS7(보는 작업 공간 전달)) + 받는 쪽 고아 안전망 문제 발견, 코드 없음


### `history/process-share-realtime.md` — 공유-d(실시간) (2026-10-07 신설)
- 2026-10-08(21) 실시간-7(확인: 조합 시나리오 17·회귀 52·화면 보강) 통과 — 공유-d(실시간) 마감, 코드 변경 없음
- 2026-10-07(9) 구조안 제시 SD1(출력 구독 범위)~SD9(경로 줄 멈춤) + 실시간-1~7, 코드 없음
- 2026-10-07(9) SD3 논의: 닫기 → 최소화 개정, 루트 탭 전환 + 스크립트 행 "실행 중 N", 목록 kill·모든 kill 확인창·종료하고 닫기, 머리 줄 A안 `RUN▾ [최소화|크기|×]` 확정 + 소유자 kill 조건, 시안 v1~v6 + 작업 순서 확정(실시간-0(터미널 UI) 맨 앞, 실시간-6 → 6a·6b 분할)

### `history/process-share-reconcile.md` — 공유-d 실시간-1~3(재계산·연결) (2026-10-07 신설)
- 2026-10-08(18) 실시간-5(노드 응답: 공유 루트 목록·`parentVisible`) 코드 + 격리 실행 12/12(커밋 전)
- 2026-10-08(17) 실시간-4(exec 타일) 코드 + 격리 실행 14/14(커밋 전)
- 2026-10-08(16) 실시간-3(노드 이동·삭제 연결) 코드 + 격리 실행 15/15(커밋 전)
- 2026-10-07(15) SH1(공유 단위) 개정 = 폴더만 공유(스크립트 PUT 400) + 재확인
- 2026-10-07(14) 실시간-2 코드(공유 PUT·DELETE 연결·받는 쪽 타일) + 격리 실행 35/35(커밋 전)
- 2026-10-07(13) 실시간-1 코드(`reconcileAccess`·쿼리 3개·`SHARE:CHANGE`) + 임시 DB 쿼리 확인(커밋 전)
- 2026-10-07(12) 실시간-1(재계산 함수) 구조안 + 정할 것 4개 제시, 코드 없음

### `history/process-share-verify.md` — 공유-g(확인) (2026-10-12 신설)
- 2026-10-12(1) ⑥ "나"(`startedByMe`) 수정 + 통합 확인 70/70 · 위치 회귀 21/21 · 헤드리스 ⑥

### `history/process-share-terminal-place.md` — 터미널 타일 위치 · 폴더 타일 uid (2026-10-08, process-share-realtime에서 분할)
- 2026-10-12(29) 브랜치 커밋 + 분기 지점 태그 `checkpoint/before-folder-tile-uid` + `master` `--no-ff` 병합
- 2026-10-12(28) `openFolder` 제거 + 제거 전 23/23·후 21/21 격리 확인
- 2026-10-12(27) 사용자 직접 확인 통과 + 폴더 process 정리(`openFolder`는 안 쓰는 옛 경로), 코드 없음
- 2026-10-08(26) FU2 개정: 공유한 타일과 같은 uid + uid 우선 찾기(격리 11 통과, 커밋 전)
- 2026-10-08(25) 다른 보드로 가는 터미널 수정(placeTerminal 공용) + 폴더 타일 uid(FU1~FU4) 코드·격리 확인(브랜치 `feat/folder-tile-uid`)
- 2026-10-08(24) 다른 보드로 가는 터미널 조사(빈틈 3곳) + 폴더 타일 uid 지시(FU1~FU4 대기) + 체크포인트 커밋·분기, 코드 없음
- 2026-10-08(23) 공유 전에 실행한 터미널을 받는 쪽이 못 여는 빈틈 수정(공유 시 터미널 타일 + "내 보드에 열기")

### `history/process-share-realtime-ui.md` — 공유-d 실시간-0(터미널 UI) (2026-10-07 신설, process-share-realtime에서 분할)
- 2026-10-08(22) 루트 타일이 공유 폴더 안에서 막힘 수정(배너에 "기본으로 돌아가기") — 사용자 확인 통과
- 2026-10-08(20) 실시간-6b(공유 반영: 볼 수 없는 타일 배너·경로 멈춤·`SHARE:CHANGE` 재조회) 코드 + 헤드리스 확인(커밋 전)
- 2026-10-08(19) 실시간-6a(스크립트 행 "실행 중 N") 코드 + REST 11/11 + 헤드리스 화면 확인(커밋 전)
- 2026-10-07(11) 실시간-0 코드 + 격리 실행 헤드리스 24/24 통과(커밋 전)
- 2026-10-07(10) SD1~SD9 전부 확정 + 실시간-0 구현 계획·확인할 것 4개 제시, 코드 없음

### `history/process-share-display.md` — 공유-e(프론트 표시) (2026-10-08 신설, e-3·f는 분할)
- 2026-10-08(5) 공유-e-2(공유 걸기 창) 코드·격리 확인 11개 통과(커밋 전)
- 2026-10-08(4) 공유-e-2(공유 걸기 창) 구조안·시안 확정, 코드 없음
- 2026-10-08(3) 공유-e-1 격리 실행 확인 통과(+단계 변경 갱신 버그 수정)
- 2026-10-08(2) 공유-e-1(받는 쪽 표시) 코드·빌드 통과(커밋 전)
- 2026-10-08 공유-e-0(제안 항목 확정): ①~④·⑥ 채택·B 제외, 숨기기·O4·O5 분리, 코드 없음

### `history/process-share-owner.md` — 공유-e-3 소유자 쪽 목록(O1~O3·O5) (2026-10-08, display에서 분할)
- 2026-10-08(9) 공유-e-3c(내가 공유한 것) 코드·빌드 통과 → 사용자 확인 통과(커밋 전)
- 2026-10-08(8) 공유-e-3b(행 공유 개수) 코드·빌드 통과 → 3c와 함께 사용자 확인 통과(커밋 전)
- 2026-10-08(7) 공유-e-3a(공유 목록 창) 코드·격리 확인 11개 통과(커밋 전)
- 2026-10-08(6) 공유-e-3 구조안·시안 확정 + 공유 표시 색 청록(`share`) 적용

### `history/process-share-presence.md` — 공유-f 접속 상태(presence) (2026-10-08, display에서 분할)
- 2026-10-08(11) 공유-f-5(확인) 격리 API 25개 + 헤드리스 화면 통과
- 2026-10-08(10) 공유-f(presence) f-1~f-4 코드·빌드 통과, 확인 전(커밋 전)

### `history/workspace-front.md` — 보드(작업 공간 여러 개): 프론트 코드 작성 (2026-10-07 신설)
- 2026-10-07(8) 작업공간-e(확인): 헤드리스 34개 + 공유-b·c 회귀 59개 통과 + 발견(받는 쪽 실행 터미널이 소유자 기본 보드에 고아로 붙음, 공유-d 때 정함), 코드 수정 없음
- 2026-10-07(6) 화면 이름 "보드" 확정 + 작업공간-d(프론트) 코드·type-check 통과(커밋 전, 브라우저 확인은 작업공간-e)

### `history/workspace-server.md` — 작업 공간 여러 개: 서버 코드 작성 (2026-10-07, workspace-impl에서 분할)
- 2026-10-07(5) 작업공간-c(exec·안전망·탭) 코드(exec `workspace`·서버 고아 안전망·`TAB:VIEW`·`resyncSizeOwner`)·worker 포함 격리 실행 20개 통과(커밋 전)
- 2026-10-07(4) 작업공간-b(API) 코드(`/tiles/:hash`·WS5(터미널 타일 추가 주체) 검사·`workspace.go`)·격리 실행 41개 통과(커밋 전) + 과도기 위험(`adoptOrphans` PUT 400)
- 2026-10-07(3) 작업공간-a(DB) 코드(`00008_workspaces.sql`·쿼리)·확인 20개 + down/up 통과(커밋 전) + 실수로 supervisor 실행 → 개발 DB에 00007·00008 적용
### `history/process-share-node.md` — 공유: 권한 단계 · 노드 복사 · 클립보드 (2026-10-06, process-share REF 분할로 신설)
- 2026-10-06 신설(결정 이력 본문은 `history/process-share.md` 2026-10-06)

### `history/process-share-db.md` — 공유: 공유-a(DB) (2026-10-06 신설)
- 2026-10-07 공유-a 검토(커밋 막을 문제 없음) + `ListChildrenVisible` 소유자 40 고정 보완
- 2026-10-06(2) SA1(단계 저장 형식)~SA4(기존 행 채우기) 확정 + 공유-a 코드 작성, build/vet + 임시 DB 확인 통과(커밋 전)
- 2026-10-06 구조안 제시(`00007_shares.sql`·판정 함수·판정 쿼리, SA1~SA4) + sqlc 함수 반환형 인식 확인, 코드 없음

### `history/process-share-authz.md` — 공유: 공유-b(인가 교체) (2026-10-06 신설)
- 2026-10-07 인가-1(판정 헬퍼)~인가-8(확인) 코드 작성 + 격리 실행 53개 항목 통과(커밋 전)
- 2026-10-06(2) SB1(권한 부족 응답)~SB6(이동 사이클 검사) 확정, 코드 없음
- 2026-10-06 구조안 제시(엔드포인트별 요구 단계, SB1~SB6) + `NODE:0` 누설 발견, 코드 없음

### `history/process-share-api.md` — 공유: 공유-c(공유 API) (2026-10-07 신설)
- 2026-10-07(2) SC1~SC7 확정 + 공유API-1(판정 함수 정리)~공유API-7(확인) 코드 작성, 42개 + 회귀 50개 통과(커밋 전)
- 2026-10-07 구조안 제시(API 8개, SC1(출처 계산 위치)~SC7(복사 제외 개수)), 코드 없음

### `history/process-sync-impl.md` — process 동기화 1단계 구현 · 실행 확인 (2026-10-01, process-sync에서 분할)
- 2026-10-01(6) sid 충돌 수정(`signIn` nonce) + 커밋 전 사용자 검토로
- 2026-10-01(5) 2단계 실행 확인: 탭 id·계정 구독·크기 우선권·kill·X·W·Y·프론트(헤드리스) 전부 통과 + sid = 쿠키 원본값 충돌 발견(미해결)
- 2026-10-01(4) 1단계 코드 작성(1-a~1-g), build/vet/type-check 통과
