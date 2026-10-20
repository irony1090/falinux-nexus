# CURRENT

## 현재 날짜
2026-10-20

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **파트(파일 전송) 마감(2026-10-20) → `INDEX-transfer-e2e.md`**: e2e 확인 34/34 -> F1-a(재접속 대기)·F1-b(강제 Close 제거)·F2(청크 256KB) + 재확인 21/21. 끊김·worker 재시작이 같은 send 안에서 이어받기, 왕복 50ms 615KB/s -> 4.5MB/s
  - `history/transfer.md` 11.5k자 — 다음에 고칠 때 분할(파트(파일 전송)은 새 쌍 `REF-transfer-e2e`/`history/transfer-e2e`에 써서 안 건드림)
- **다음 할 것**: ① 파트(한글 입력 중복) 사용자 실입력 확인 -> 마감(아래 줄) ② (미룸) worker PTY `IUTF8` ③ 그다음 "다음 파트" 후보에서 사용자가 고름
- **파트(페이지 표시 개편) 마감(2026-10-20) → `INDEX-board-pager.md`**: 헤더 `VPagination` -> `GridPager`(도트·반 칸 작은 도트·드래그 칼럼 이동·아래로 당기면 `TilePeek` + active 타일 이동·창 방식). 격리 헤드리스 31/31 + 사용자 확인. 프로토타입 아티팩트 https://claude.ai/artifact/XLvPvzUkZQrfD1FN6Vaqo5
- **파트(URL 바인딩) 마감(2026-10-19) → `INDEX-tile-url.md`**: `?focus=`(active)·`?view=`(보던 칼럼)로 새로고침 복원. 헤드리스 20/20 + 사용자 브라우저 확인
- **파트(터미널 터치 스크롤) 마감(2026-10-19) → `INDEX-term-touch.md`**: 안 가(스크롤바 `touch-action: none`) + 안 나(쓸어서 스크롤, `touchScroll.util.ts`: 상자 먼저·관성·정상 화면 `scrollLines`·alt 화면 휠). 헤드리스 11/11 + 사용자 실기기 확인. 확인용 노드 irony `scroll-test`(215)·`scroll-test.ps1`(216)은 남겨 둠(사용자 결정)
  - 곁가지 해소: 셸 노드 종료 10초 = 설계대로(대화형 셸이 종료 요청 무시 -> 10초 뒤 SIGKILL), 사용자 "그대로 둬" → `REF-process-kill.md` 메모
- **파트(Windows worker) 마감(2026-10-19) → `INDEX-windows-worker.md`**: `pty` OS별 분리 + ConPTY·Job Object, `.ps1` = powershell / 그 외 `cmd /c` + `SCRIPT_FORMAT`, 에디터 `$VISUAL` > `EDITOR` > PATH 탐색, OS L/W 표시(`00011`). VM 17/17 + Ctrl+C 상속 수정 + 사용자 재접속 확인. 미확인 = `.py`·EDIT 저장 CRLF·TEMP 공백.
  - VM `nexus-win`은 남겨 둠(`docker start/stop nexus-win`, 화면 `127.0.0.1:8006`). VM worker `win-dev`는 사용자 supervisor 8012에 연결된 채(`nexus` 명부에 `win-dev os=windows`) → `REF-windows-worker.md` "VM 실행". 격리 5051 종료·`nexus_win_tmp` 삭제함
- **파트(PgUp·PgDn 버튼) 마감(2026-10-19) → `INDEX-term-page-keys.md`**: active 터미널 타일 오른쪽 아래 버튼 2개, 키 전송, 키보드 열림·전체보기 중 숨김. 사용자 브라우저 확인
- 파트(고아 스크립트) 마감(2026-10-15) → `INDEX-worker-orphan.md`: worker가 SIGHUP·SIGTERM·SIGINT로 스크립트 정리(TERM -> 5초 -> KILL) 후 종료, `nohup`이면 SIGHUP만 무시. 사용자 재실행 확인. 
- (이전 순서) 2026-10-14에 정한 순서 ① 파트(확인·실측) → ② 파트(모바일 가상 키보드) → ③ 파트(보드 미리보기)는 전부 마감
  - ③ 파트(보드 미리보기) **마감(2026-10-15) → `INDEX-board-preview.md`**: 전체보기에 다른 보드 줄 + 다른 보드 터미널 실시간 + 닫을 때 전체보기가 만든 xterm만 정리(중간안). 코드 `52ae8b9`
  - 후보(아래 "다음 플랜"·미해결에서): ~~전체 화면을 켜면 칼럼이 `single` -> `row`~~ 버그 아님(2026-10-15 확인): 전체 화면으로 Grid 영역 높이가 `TILE_THRESHOLD.h` 560 이상 -> `capacity` y 1 -> 2 -> `effectiveSize`가 저장값 h 0.5를 그대로 써서 `single` -> `row`(1칸 축의 0.5를 1로 보는 규칙, `tileSplit.util.ts`) / process 단독 공유 UI 여부 / 회원가입 / ~~파일 전송 e2e~~(2026-10-20 마감 → `INDEX-transfer-e2e.md`) / iPhone 실기기 확인 / ~~Windows worker~~(2026-10-19 마감 → `INDEX-windows-worker.md`) / **worker 화면 송출 + 조작**(2026-10-15 아이디어, VNC 터널 + noVNC, SC1(VNC 서버 준비) = 장비 설치 전제 · SC2(못 쓰는 장비) = 가능 여부 보고 + 버튼 비활성화 결정, → `REF-worker-screen.md`) / ~~EDIT 에디터 소소한 수정~~(윈도-4로 해결)
- **시연 데이터(2026-10-16) 등록 → `REF-demo.md`**: irony 계정 폴더 `시연-생산지그`(id 200, 안 C(생산 지그 시나리오)) + test1 편집·test2 조작·test3 보기 공유. 코드 변경 없음(REST로 데이터만, 커밋할 코드 없음)
  - **다음 할 것**: 브라우저로 시연 순서 점검(irony + test3 동시) — Nexus 경유 실행은 아직 미확인(xterm.js 한글 정렬·24bit 색 사진, worker 환경 카메라). 시연 순서 → `REF-demo.md` "시연 순서"
  - 사용자가 하려던 기능 추가는 내용이 기억나지 않아 보류(2026-10-16)
- **외부 접속 설정 사용 중**: 프론트 https://test1.falinux.com(8011) · supervisor https://test2.falinux.com(8012) · worker -> localhost:8012 → `REF-external-access.md`. 로컬 전용(5050·3000)으로 되돌리는 법도 거기
- 최근 마감 파트 목록 → `HISTORY.md` "마감 파트" 색인
  - 남은 주의: 개발 DB에 테스트 계정 `active-<시각>` 3개(폴더 11개씩) / 개발용 로그인 기본값 = `apps/frontend/.env.development.local`(git 제외) / worker는 종료 정책 이후 빌드여야 force를 앎
- **다음 플랜**: 위 "다음 파트" 후보에서 사용자가 고름
  - 나중으로 미룸: process 단독 공유 UI를 둘지 결정(지금 서버 API만, 두면 받는 쪽 진입점 = 기본 루트 터미널 탭, `REF-tile-binding-rework.md` R3(경로 없는 process)) / 회원가입
- 5 공유에서 뒤로 미룬 것: ~~숨기기~~(2026-10-13 철회 — 타일 닫기로 충분, → `REF-process-share.md`) / ~~잘라내기~~ → 파트(노드 클립보드)로 마감(2026-10-13) / ~~보드 미리보기~~ → ③ 파트(보드 미리보기)로 마감(2026-10-15, 시안 E = 전체보기 줄) / 단축키 키 조합 / ~~크기 우선권 넘김 실측~~(파트(크기 가져오기)에서 15초로 바꾸고 실측 15.0~15.5초) / ~~`tabGrace` 60초 만료 실측~~(실측-2(탭 보관 60초), 60.1초)
- 커밋 상태: 2026-10-19 파트(Windows worker) 마감까지 커밋(`41ff14a`, 코드+vault 한 커밋). 2026-10-19 파트(터미널 터치 스크롤) 마감도 커밋(`23e10a2`). 파트(URL 바인딩) 마감도 커밋(코드 + vault, 파일 전송 e2e 기록 포함). 파트(페이지 표시 개편) 마감도 커밋(코드 + vault, 2026-10-20). 파트(파일 전송) 마감도 커밋(코드 + vault, 2026-10-20). 파트(한글 입력 중복)는 확인 전 상태로 커밋(코드 + vault, 2026-10-20, 사용자 요청) — 마감 때 vault만 따로 커밋. **push는 한 번도 안 함**(분기 태그 `checkpoint/before-folder-tile-uid` = `d50f153`, 태그는 push 때 따로 `git push origin <tag>`). 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치). 커밋은 파트 단위(작업 단위마다 안 함). user vault 전체 git 제외
- 이전 세션(2026-10-12~15)의 scratchpad 스크립트는 이미 사라짐 — 명세는 `REF-test-scripts.md`(~10-13)·`REF-test-scripts-1014.md`(10-14~) 날짜별 절. 그때 격리 프로세스·임시 DB는 전부 정리됨. PC 재부팅 뒤라면 `docker start postgres15` -> supervisor -> worker -> `npm run dev`(→ MEMORY "로컬 개발 환경")
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
- **진행 중 파트 = 한글 입력 중복(2026-10-20)** → `REF-term-ime.md`: 원인 = 리눅스 IME의 "지우고 다시 넣기" 확정에서 xterm이 ①(`_inputEvent`)·②(조합 마무리 타이머)로 두 번 전송. 가) `xtermIme.util.ts` `blockImeRecommit`로 ① 막기 + xterm 정식 6.0.0 복귀, 헤드리스 재생 3/3. **남은 것 = 8011 재시작(vite 캐시가 아직 베타일 수 있음 — `node_modules/.vite/deps/@xterm_xterm.js`에 `_textareaChangeTimer`가 없어야 6.0.0) -> 사용자 실입력 확인(`ime-test` 221, 받침 넘김 말: 아니네·앞에·갑자기) -> 마감**. scratchpad `ime/`(재생 테스트)는 세션 끝나면 사라짐 -> 명세 `REF-test-scripts-1014.md`
- **(미룸) worker PTY에 `IUTF8` 없음(2026-10-20 발견)**: 줄 단위 입력(`read`·`cat`)에서 한글 백스페이스가 1바이트만 지움(`ime-test`에서 화면 `어라 앞앞에` vs 받음 `어라 라 앞앞에`). `internal/execute/pty`에 termios 설정 없음. raw 모드 프로그램(Claude Code·readline)은 영향 없음. 한글 중복 해결 뒤 다룸
- ~~`NODE:` 동적 구독~~ → 2026-10-12 파트(폴더 목록 동기화)로 해소(계정 토픽, → `REF-node-list-sync.md`)
- ~~같은 터미널을 연 탭이 여러 개면 질의에 탭마다 응답~~ → 파트(질의 응답 중복)로 해결(2026-10-13 마감, → `REF-term-query.md`)
- ~~**input**: I3(PENDING 중 입력 버림) 확인만 남음~~ → 실측-1(PENDING 입력 버림)로 확인 끝(2026-10-14) → `REF-process-input.md`
- ~~**크기 바뀐 직후 htop 줄 밀림**~~ → 실측-3(htop 줄 밀림)에서 원인 확인·수정(2026-10-14 마감, → `REF-size-take.md` "실측-3 결과")
- **Hub 막힘**: ping/pong(`e395433`)으로 해결, 남은 한계 = 막힘 최대 약 25초 → `REF-realtime.md` "발견"
- ~~**파일 전송**: e2e 미검증~~ → 파트(파일 전송)로 해결(2026-10-20 마감, → `INDEX-transfer-e2e.md`). `history/transfer.md` 11.5k자 — 다음에 고칠 때 분할
- **서브키 충돌/위조**: key↔subkey 결속 검증 미구현(명부는 인가에 안 쓰므로 계속 보류 — RS6(위조 검증))
- **supervisor 영속성**: registry 메모리 → PG 미착수
- ~~worker 종료 시 실행 중 스크립트가 고아로 남음~~ → 파트(고아 스크립트)로 해결(2026-10-15 마감, → `INDEX-worker-orphan.md`). 남은 한계: 스크립트가 직접 HUP를 피하고 worker가 crash하면 남음 — 안 마(시작 시 청소) 후보

## 잔여 (틈날 때)
- ~~`throwCatch` 상태 코드 / 서버 미기동 시 설정 삭제~~ → 로그인-3(401만 로그아웃)으로 해소(2026-10-12)
- ~~`SESSION_KEY` env화~~ · ~~checkSession createdAt=0~~ → 파트(서버 정리) 정리-2(세션 키 env)·정리-3(createdAt)으로 해소(2026-10-13)
