# CURRENT

## 현재 날짜
2026-10-14

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **다음 순서(2026-10-14 사용자 결정)**: ① 파트(확인·실측) → ② 파트(모바일 가상 키보드) → ③ 파트(보드 미리보기 = 루트 보드에서 다른 보드들의 미리보기 타일)
  - ~~① 파트(확인·실측)~~ **마감(2026-10-14)**: 실측-1(PENDING 입력 버림) 7/7 → `REF-process-input.md` / 실측-2(탭 보관 60초) 9/9 → `REF-process-sync-impl.md` / 실측-3(htop 줄 밀림) 원인 확인·밀림-가(먼저 발행)·밀림-나(즉시 반영) 수정 → `REF-size-take.md` "실측-3 결과". 스크립트 명세 → `REF-test-scripts.md` "2026-10-14"
  - ~~② 파트(모바일 가상 키보드)~~ **마감(2026-10-14) → `INDEX-mobile-keyboard.md`**(★ 반복 주제 `REF-mobile-viewport.md`). 특수 키 줄 + 커서 기준 이동 + `resizes-content` 전환(키보드 열린 동안 타일 영역 높이 고정). Fold 8 크롬 정확 / Edge 폰 크기 기본 화면 빈틈 약 34(브라우저 오류, 기록만). iPhone 확인은 나중
    - 폰 사진 올리기 도구는 `vite.config.mts`에서 주석 처리(→ `REF-test-env.md` "폰 사진 올리기")
    - 미룬 별건: 전체 화면을 켜면 같은 칼럼이 `single` -> `row`(빈 칸 추가)로 바뀜(사진 `user/vrKeyboad/photo_2026-10-14_14-21-37.jpg`·`..._14-21-38.jpg`)
  - **다음 = ③ 파트(보드 미리보기)**: `REF-workspace.md` "희망안"에서 설계부터(출력 구독 범위 결정 필요)
  - **색인 2단화 대기**: MEMORY 9.98k·HISTORY 9.9k자로 10k 직전 → 마감 파트(HISTORY "진행 중 파트" 절에 남은 node-label·frontend-auth·node-ui-active·size-take·node-clipboard/select·term-query·process-kill·node-list-sync·tile-binding 등)를 어떤 묶음으로 `INDEX-*`에 접을지 사용자에게 물어볼 것
- **외부 접속 설정 사용 중**: 프론트 https://test1.falinux.com(8011) · supervisor https://test2.falinux.com(8012) · worker -> localhost:8012 → `REF-external-access.md`. 로컬 전용(5050·3000)으로 되돌리는 법도 거기
- **최근 마감 파트**(상세는 HISTORY 색인·각 REF): 2026-10-14 크기 가져오기(→ `REF-size-take.md`) · 확인·실측(→ `REF-size-take.md` "실측-3 결과" 외) / 2026-10-13 노드 클립보드(→ `REF-node-clipboard.md`·`REF-node-select.md`) · 외부 접속·전체 화면 버튼(→ `REF-external-access.md`·`REF-frontend.md`) · 질의 응답 중복(→ `REF-term-query.md`) · 서버 정리(→ `history/supervisor-web.md`·`history/node-label.md` 2026-10-13(2)) · 바인딩 타일(→ `REF-tile-binding-rework.md`·`-impl.md`) / 2026-10-12 5 공유(→ `INDEX-process-share.md`) · active 타일·원격 배지(→ `REF-node-ui-active.md`) · worker 명부(→ `REF-node-label.md` "roster 계획") · user/login 마무리(→ `REF-frontend-auth.md`) · process 종료 정책(→ `REF-process-kill.md`) · 폴더 목록 동기화(→ `REF-node-list-sync.md`)
  - 남은 주의: 개발 DB에 테스트 계정 `active-<시각>` 3개(폴더 11개씩) / 개발용 로그인 기본값 = `apps/frontend/.env.development.local`(git 제외) / worker는 종료 정책 이후 빌드여야 force를 앎
- **다음 플랜**: 위 "다음 순서" ①~③
  - 나중으로 미룸: process 단독 공유 UI를 둘지 결정(지금 서버 API만, 두면 받는 쪽 진입점 = 기본 루트 터미널 탭, `REF-tile-binding-rework.md` R3(경로 없는 process)) / 회원가입
- 5 공유에서 뒤로 미룬 것: ~~숨기기~~(2026-10-13 철회 — 타일 닫기로 충분, → `REF-process-share.md`) / ~~잘라내기~~ → 파트(노드 클립보드)로 마감(2026-10-13) / 보드 미리보기(`/` 루트 타일 안 다른 보드 카드, `REF-workspace.md` "희망안") / 단축키 키 조합 / ~~크기 우선권 넘김 실측~~(파트(크기 가져오기)에서 15초로 바꾸고 실측 15.0~15.5초) / ~~`tabGrace` 60초 만료 실측~~(실측-2(탭 보관 60초), 60.1초)
- 커밋 상태: 2026-10-14 파트(모바일 가상 키보드) 마감까지 전부 커밋(사진 44장은 user vault `user/vrKeyboad/`로 옮김, **user vault 전체 git 제외** 2026-10-14). **push는 한 번도 안 함**(분기 태그 `checkpoint/before-folder-tile-uid` = `d50f153`, 태그는 push 때 따로 `git push origin <tag>`). 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치). 커밋은 파트 단위(작업 단위마다 안 함)
- 재부팅(2026-10-14 세션 끝)으로 사라짐: scratchpad `pw/kb.mjs`·`kb2.mjs`·`kb3.mjs`·`vkchk.mjs`·`unitchk.mjs`·`up.mjs`·`vite.iso.mts`·`iso/`(바이너리·env)·`chk/`·`pw/node_modules` — 명세 → `REF-test-scripts.md`(2026-10-14 확인·실측 + 모바일 가상 키보드 줄). 임시 DB `nexus_kb_tmp` 삭제함. 사용자 DB에 헤드리스 로그인 세션 몇 개(irony, 데이터 변경 없음)
- 이전 재부팅들(2026-10-12~13)로 사라진 스크립트 명세는 `REF-test-scripts.md`·각 파트 REF 확인 절에
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
4. ~~`NODE:<parentId>` 동적 구독/해지~~ — **해소**(2026-10-12 파트(폴더 목록 동기화): 동적 구독 대신 계정 토픽 + 발행 시점 판정, → `REF-node-list-sync.md`)
5. **네이밍**: 전체 개념 "그리드" vs "타일링", `Band`(가칭) — 작업 공간 화면 이름은 **"보드"로 확정**(2026-10-07)
15. **모바일 가상 키보드**: A+B 확정, **② 파트(모바일 가상 키보드) 마감(2026-10-14) → `INDEX-mobile-keyboard.md`**, iPhone 실기기 확인은 나중 → `REF-node-ui-overview.md` "모바일 가상 키보드". **주의**: `appWindown.store.ts`의 `size.inner`(visualViewport)에 판정을 연결하면 안 됨

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
- ~~핸들러 책임 미적용: parentId owner 일치 검증 / 자기 자손으로 Move 사이클 방지~~ — **이미 구현돼 있음**(2026-10-13 확인): 생성 = 부모 편집(30) 이상·소유자 = 부모 소유자 / 이동 = `checkMove`(목적지 편집·같은 소유자·`IsDescendant` 사이클 금지, SB6(이동 사이클 검사)) / 복사 = 폴더만·원본을 다 읽은 뒤 넣음
- ~~남은 빈틈: 생성·이동이 부모 종류를 안 봄~~ → 정리-1(부모 종류 검증)으로 해소(2026-10-13, `requireFolder` 400, DB 제약은 안 넣음 — 생성·이동 경로가 핸들러뿐)
- ※ 마이그레이션 번호는 착수 때 실제 다음 번호로(`00005`는 `process_subscribers` 제거에 씀)

### process — 남은 것
> 배선 → `INDEX-process-wiring.md` / 터미널 기능 → `INDEX-process-terminal.md` / 동기화·공유·보드 → `INDEX-process-share.md`
- EXEC 실행 방식 = **A(직접 실행, shebang 필수) 현상 유지 확정**(2026-10-12 사용자 결정)
- **kill 정책 확정(2026-10-12)** → `REF-process-kill.md` "확정" — 마감·커밋
- (이전 기록, 위로 해소) 끊긴 창 kill 거절 vs 큐잉(입력은 I3(PENDING 중 입력) = 버림 확정) / ~~공유 kill 인가~~(SH3(받은 계정 권한)으로 해소: 조작 이상) / kill 에스컬레이션
- **정리 잔여(구)**: `register.go` 주석 SendBuffer 테스트. worker `baseDir` 필드·`instanceKey()` dead code화(정리 여부 판단)

### 프론트 user/login — 마감(2026-10-12, → `REF-frontend-auth.md`)
- 서버 쪽 잔여(`SESSION_KEY` env화 · checkSession `createdAt=0`) → 파트(서버 정리)에서 해소(2026-10-13)

---

## 미해결 이슈 (이월)
- ~~`NODE:` 동적 구독~~ → 2026-10-12 파트(폴더 목록 동기화)로 해소(계정 토픽, → `REF-node-list-sync.md`)
- ~~같은 터미널을 연 탭이 여러 개면 질의에 탭마다 응답~~ → 파트(질의 응답 중복)로 해결(2026-10-13 마감, → `REF-term-query.md`)
- ~~**input**: I3(PENDING 중 입력 버림) 확인만 남음~~ → 실측-1(PENDING 입력 버림)로 확인 끝(2026-10-14) → `REF-process-input.md`
- ~~**크기 바뀐 직후 htop 줄 밀림**~~ → 실측-3(htop 줄 밀림)에서 원인 확인·수정(2026-10-14 마감, → `REF-size-take.md` "실측-3 결과")
- **Hub 막힘**: ping/pong(`e395433`)으로 해결, 남은 한계 = 막힘 최대 약 25초 → `REF-realtime.md` "발견"
- **파일 전송**: 구현 완료 / e2e 미검증. 잔여: e2e 스모크 / abort sentinel. `history/transfer.md`가 분할 기준 초과(17k자) — 다음에 고칠 때 분할
- **서브키 충돌/위조**: key↔subkey 결속 검증 미구현(명부는 인가에 안 쓰므로 계속 보류 — RS6(위조 검증))
- **supervisor 영속성**: registry 메모리 → PG 미착수

## 잔여 (틈날 때)
- ~~`throwCatch` 상태 코드 / 서버 미기동 시 설정 삭제~~ → 로그인-3(401만 로그아웃)으로 해소(2026-10-12)
- ~~`SESSION_KEY` env화~~ · ~~checkSession createdAt=0~~ → 파트(서버 정리) 정리-2(세션 키 env)·정리-3(createdAt)으로 해소(2026-10-13)
