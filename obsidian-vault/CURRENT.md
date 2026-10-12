# CURRENT

## 현재 날짜
2026-10-12

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **2026-10-12 파트(5 공유) 마감** — 동기화·공유·보드·presence 전부 `master`에 커밋(공유-g(확인) 통합 70/70 + 위치 21/21 + 헤드리스, ⑥ "나" 수정 포함). 상세·원문 색인 → `INDEX-process-share.md`(세션 시작 시 안 읽음, 필요하면 INDEX부터)
- **2026-10-12 파트(active 타일 · 위치 번호 단축키 + 다른 계정 active 배지) 마감** — 로컬 active(`Ctrl+Alt+숫자`) + 원격 배지(터미널·폴더) 사용자 확인 후 커밋 → `REF-node-ui-active.md`. 뺀 것: "위치 없음" 행 열기(`openProcessTile`)는 active 안 됨(응답에 tile id 없음)
  - 개발 DB에 테스트 계정 `active-<시각>` 3개(폴더 11개씩) 남음
- **2026-10-12 파트(worker 명부(roster)) 마감** — 명부-1(기록)~명부-4(확인) 사용자 확인 후 커밋 → `REF-node-label.md` "roster 계획"(구현 메모·한계 포함). label·DraggableSession 리팩터 보류 결정도 같이 커밋
- **2026-10-12 파트(user/login 마무리) 마감** — 로그인-1(상태를 모듈로)~로그인-5(확인) + 보드에서 로그아웃 시 "보드 삭제" 오인 알림 수정, 사용자 확인 후 커밋 → `REF-frontend-auth.md`
  - 개발용 로그인 기본값은 이제 `apps/frontend/.env.development.local`(git 제외, 만들어 둠)
- **회원가입 파트 = 나중에**(2026-10-12 사용자 결정): 가입 정책 + 장비 실행 권한 + bcrypt 전환을 함께 정한 뒤 → `REF-frontend-auth.md` "회원가입"
- **다음 파트 후보**(사용자와 정할 것): process 결정 필요 항목(EXEC 실행 방식·kill 정책, 아래) / Node 핸들러 책임 미적용 2개(아래) / 회원가입
- 5 공유에서 뒤로 미룬 것: 숨기기(`node_shares.hidden_at`, 복원 UI "숨긴 공유 N") / 잘라내기(SH10(노드 복사) 확장) / 보드 미리보기(`/` 루트 타일 안 다른 보드 카드, `REF-workspace.md` "희망안") / 단축키 키 조합 / `tabGrace` 60초 만료·크기 우선권 넘김 실측
- 커밋 상태: 2026-10-12 파트(user/login 마무리) 마감까지 전부 커밋(분기 태그 `checkpoint/before-folder-tile-uid` = `d50f153`). **push는 한 번도 안 함**(태그는 push 때 따로 `git push origin <tag>`). 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치). 커밋은 파트 단위(작업 단위마다 안 함)
- 재부팅으로 사라지는 것: 2026-10-12 두 세션의 scratchpad(`12185a69…`·`f3a527cf…`) 확인 스크립트·격리 설정 전부(`g_test.py`·`of_test.py`·`pw/ui.mjs`·`pw/active.mjs`·`pw/remote.mjs`·`vite.iso.mts`·`pw/node_modules`). 다시 쓸 땐 `REF-test-env.md` "격리 실행"·"2026-10-12 확인 스크립트"·`REF-process-share-verify.md` "방법"(항목표가 곧 재작성 명세). 임시 DB `nexus_active_tmp`는 삭제함, 5050·3000은 사용자 것
- 테스트·확인 준비(환경, 헤드리스, 함정, **격리 실행 = 임시 DB + 5051**) → `REF-test-env.md`
- 역할 분담(2026-09-29 개정): node 타일 UI·process 작업은 코드 대부분을 Claude가 작성하고, 로직 구조는 사용자와 먼저 맞춘 뒤 쓴다

---

## 🎯 node 카탈로그 타일 UI — 남은 것
레이아웃(→ `INDEX-node-ui-layout.md`)·연동(→ `INDEX-node-ui-link.md`) 마감. 컨셉 → `REF-node-ui.md`, 전체보기·가상 키보드 → `REF-node-ui-overview.md`.

- **미리보기**: `apps/frontend`에서 `npm run dev` → `/`. supervisor + 로그인 필요
- **색 규칙(2026-09-29)**: Vuetify 테마 색 이름 의존, hex 금지 → `REF-frontend.md` "색 규칙"
- **기타 유지 사항**: `position_x/y` 사용 중단(DB 컬럼 유지) / 트리 드래그 시 device_key 재상속 이슈(→ `REF-node-label.md`) / 외부 레이아웃 라이브러리 안 씀

**열린 질문 (번호 유지, 해소된 ②⑥~⑭는 REF/history에)**
1. 분할 UX — 버튼메뉴 vs 드래그드롭(VSCode류). 시안은 실행/새 타일 버튼
3. 라우팅 스킴 재검토(옛 `/nodes/:parentId?`가 이 모델에 맞는지)
4. `NODE:<parentId>` 동적 구독/해지 — node도 process처럼 REST로 갈지 미정. 4⑤(폴더 목록 동기화)가 여기 걸림
5. **네이밍**: 전체 개념 "그리드" vs "타일링", `Band`(가칭) — 작업 공간 화면 이름은 **"보드"로 확정**(2026-10-07)
15. **모바일 가상 키보드**: A+B 확정·반영, **iPhone 실기기 확인만 남음** → `REF-node-ui-overview.md` "모바일 가상 키보드". **주의**: `appWindown.store.ts`의 `size.inner`(visualViewport)에 판정을 연결하면 안 됨

- ② `ProcessDialog` 다중 인스턴스 리팩터는 6 타일 임베드(J~N)에서 `ProcessDialog` 삭제로 해소(2026-10-06 정리 때 확인)

---

## ⏸ 보류: DraggableSession / useDragGhost 리팩터 (2026-10-12)
설계는 확정돼 있으나 소비처(캔버스·트리 드래그)가 타일 전환으로 폐기돼 쓸 곳 없음 → 드래그 UX(열린 질문 1 분할 UX 등)를 넣을 때 재검토. 근거·설계 → `REF-util-drag.md`

---

## 진행 중 / 잔여

### Node 모듈 — 남은 단계
> 설계 `REF-node-label.md` / 이력 `history/node-label.md`
4. ~~worker_instances roster~~ — **마감**(2026-10-12, `REF-node-label.md` "roster 계획"). RS6(위조 검증)은 보류 유지
6. ~~label 모듈~~ — **보류**(2026-10-12, 근거·다시 볼 조건 → `REF-node-label.md` "Label 모듈")
- 핸들러 책임 미적용: parentId owner 일치 검증 / 자기 자손으로 Move 사이클 방지
- ※ 마이그레이션 번호는 착수 때 실제 다음 번호로(`00005`는 `process_subscribers` 제거에 씀)

### process — 남은 것
> 배선 → `INDEX-process-wiring.md` / 터미널 기능 → `INDEX-process-terminal.md` / 동기화·공유·보드 → `INDEX-process-share.md`
- EXEC content→실행 세부정책(직접실행 vs `sh -c`) 미정
- **결정 필요**: 끊긴 창 kill 거절 vs 큐잉(입력은 I3(PENDING 중 입력) = 버림 확정) / ~~공유 kill 인가~~(SH3(받은 계정 권한)으로 해소: 조작 이상) / kill 에스컬레이션
- **정리 잔여(구)**: `register.go` 주석 SendBuffer 테스트. worker `baseDir` 필드·`instanceKey()` dead code화(정리 여부 판단)

### 프론트 user/login — 마감(2026-10-12, → `REF-frontend-auth.md`)
- 서버 쪽 잔여는 이번 범위 밖: 아래 "잔여(틈날 때)"의 `SESSION_KEY` env화 · checkSession `createdAt=0`

---

## 미해결 이슈 (이월)
- **`NODE:` 동적 구독**: `NODE:0` 누설은 구독 줄 삭제로 닫음(2026-10-07). 루트·폴더 토픽 설계는 열린 질문 4 → `REF-process-share-authz.md` SB5
- **같은 터미널을 연 탭이 여러 개면 질의 시퀀스(`ESC[6n` 등)에 탭마다 xterm이 응답** → process가 응답을 탭 수만큼 받음. 방안 후보: 크기 소유 탭만 응답 전달 → `REF-process-snapshot-impl.md` "스냅샷-d"
- **input**: I3(PENDING 중 입력 버림) 확인만 남음(socat 재현 → `REF-process-reconnect.md` "재현 방법") → `REF-process-input.md`
- **Hub 막힘**: ping/pong(`e395433`)으로 해결, 남은 한계 = 막힘 최대 약 25초 → `REF-realtime.md` "발견"
- **파일 전송**: 구현 완료 / e2e 미검증. 잔여: e2e 스모크 / abort sentinel. `history/transfer.md`가 분할 기준 초과(17k자) — 다음에 고칠 때 분할
- **서브키 충돌/위조**: key↔subkey 결속 검증 미구현(명부는 인가에 안 쓰므로 계속 보류 — RS6(위조 검증))
- **supervisor 영속성**: registry 메모리 → PG 미착수

## 잔여 (틈날 때)
- ~~`throwCatch` 상태 코드 / 서버 미기동 시 설정 삭제~~ → 로그인-3(401만 로그아웃)으로 해소(2026-10-12)
- `SESSION_KEY` 등 env화(현재 `"irony"` 하드코딩)
- checkSession createdAt=0(pgtype.Timestamptz gob 미직렬화) → sess.Data.ID로 DB 재조회
