# CURRENT

## 현재 날짜
2026-10-22

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **파트(빌드본 분리) 2026-10-22 마감·커밋**(→ `INDEX-build.md`). **빌드본 = `build/const`, 8013 실행 중 — 사용자의 Claude 세션이 이 위에서 돎(끄지 말 것)**. 개발 8011·8012는 꺼 둠, 띄울 때 `go run ./cmd/supervisor -env dev.env` 필수
  - 남겨 둔 것: 옛 확인용 `build/_check` 폴더 — 지울지 사용자 판단
  - **MEMORY.md 약 10k자(2026-10-22)**: 긴 줄 = 끝의 REF·history 인덱스 두 줄. 색인 2단화(마감 파트의 REF 이름을 `INDEX-*`로 옮기기) 제안 예정 — 어느 REF가 어느 파트인지 사용자 확인 필요
- **다음: 파트(worker 폴더 타일) 구상·조율 단계(2026-10-22, 코드 없음)** → `REF-worker-dir.md`. Q1~Q11 답 받음(별도 장비 타일·장비-1~3 규칙), 개정 큰 계획 폴더-1(목록 요청)·2(장비 타일)·3(여는 곳)·4(실행)·6(보드 이동)·5(업로드) 확인 대기 + **역할 분담(Claude가 코드 작성 vs 사용자 직접) 답 대기**. 안 나(업로드 기능)는 이 파트의 폴더-5(업로드)로 흡수
- 그 전 파트(리팩터링) 2026-10-21 마감(→ `INDEX-refactor.md`): 브랜치 `refactor/socket`의 커밋 3개(소켓 1차 `8cf32bc`·소켓 2차 `0dc83a7`·소켓 밖 정리 `38a24a7`)를 master에 `--no-ff` 머지, 브랜치는 남겨 둠
  - 리팩터링 뒤 옛 문서의 `tile.go` 함수 위치가 바뀜(→ `REF-refactor-cleanup.md` "결과"), 키보드 맞춤-1(타일 띄우기)은 `hook/tileFloat.hook.ts`
  - 확인 스크립트(scratchpad)는 세션 끝나면 사라짐 → 명세 `REF-test-scripts-1020.md` "2026-10-21 소켓 알림 리팩터"
- **(답 대기) 파트(파일 전송) 이름 정정 제안(2026-10-21)**: 그 파트는 내부 전송 모듈 검증·수정까지였고 **안 나(업로드 기능, UI에서 파일을 worker로 보내기)는 미착수**. 사용자가 "시작도 안 한 걸로 안다"고 해서 이름을 "파일 전송 검증(내부 모듈)"으로 바꾸고 후보에 안 나(업로드 기능)를 적자고 제안함 — 답 받으면 MEMORY·HISTORY·CURRENT·INDEX 반영
- **최근 마감 파트**(상세는 각 INDEX, 목록 = `HISTORY.md` "마감 파트"): 빌드본 분리 `INDEX-build`(10-22) · 리팩터링 `INDEX-refactor`(10-21) · 페이지 이동 `INDEX-tile-move`(10-21) · 한글 입력 중복 `INDEX-term-ime` · 파일 전송 `INDEX-transfer-e2e` · 헤더·푸터 개편 `INDEX-board-chrome` · 페이지 표시 개편 `INDEX-board-pager`(10-20) · URL 바인딩 `INDEX-tile-url` · 터미널 터치 스크롤 `INDEX-term-touch` · Windows worker `INDEX-windows-worker` · PgUp·PgDn 버튼 `INDEX-term-page-keys` · 고아 스크립트 `INDEX-worker-orphan` · 보드 미리보기 `INDEX-board-preview`(10-15~19)
  - 남겨 둔 것: 확인용 노드 irony `test`/`ctrl-alt-test`(225)·`scroll-test`(215)·`scroll-test.ps1`(216) / VM `nexus-win`(`docker start/stop nexus-win`, 화면 `127.0.0.1:8006`) + VM worker `win-dev` → `REF-windows-worker.md` "VM 실행" / Windows 미확인 = `.py`·EDIT 저장 CRLF·TEMP 공백
  - `history/transfer.md` 11.5k자 — 다음에 고칠 때 분할
- **다음 할 것**: ① (미룸) worker PTY `IUTF8`(아래 미해결 줄) ② "다음 파트" 후보에서 사용자가 고름 — 후보: ~~안 나(업로드 기능)~~(→ 파트(worker 폴더 타일) 폴더-5(업로드)로 흡수) / process 단독 공유 UI 여부 / 회원가입 / iPhone 실기기 확인 / worker 화면 송출 + 조작(→ `REF-worker-screen.md`)
- **시연 데이터(2026-10-16) 등록 → `REF-demo.md`**: irony 계정 폴더 `시연-생산지그`(id 200, 안 C(생산 지그 시나리오)) + test1 편집·test2 조작·test3 보기 공유. 코드 변경 없음(REST로 데이터만, 커밋할 코드 없음)
  - **다음 할 것**: 브라우저로 시연 순서 점검(irony + test3 동시) — Nexus 경유 실행은 아직 미확인(xterm.js 한글 정렬·24bit 색 사진, worker 환경 카메라). 시연 순서 → `REF-demo.md` "시연 순서"
  - 사용자가 하려던 기능 추가는 내용이 기억나지 않아 보류(2026-10-16)
- **외부 접속 설정 사용 중**: 프론트 https://test1.falinux.com(8011) · supervisor https://test2.falinux.com(8012) · worker -> localhost:8012 → `REF-external-access.md`. 로컬 전용(5050·3000)으로 되돌리는 법도 거기
- 최근 마감 파트 목록 → `HISTORY.md` "마감 파트" 색인
  - 남은 주의: 개발 DB에 테스트 계정 `active-<시각>` 3개(폴더 11개씩) / 개발용 로그인 기본값 = `apps/frontend/.env.development.local`(git 제외) / worker는 종료 정책 이후 빌드여야 force를 앎
- **다음 플랜**: 위 "다음 파트" 후보에서 사용자가 고름
  - 나중으로 미룸: process 단독 공유 UI를 둘지 결정(지금 서버 API만, 두면 받는 쪽 진입점 = 기본 루트 터미널 탭, `REF-tile-binding-rework.md` R3(경로 없는 process)) / 회원가입
- 5 공유에서 뒤로 미룬 것: ~~숨기기~~(2026-10-13 철회 — 타일 닫기로 충분, → `REF-process-share.md`) / ~~잘라내기~~ → 파트(노드 클립보드)로 마감(2026-10-13) / ~~보드 미리보기~~ → ③ 파트(보드 미리보기)로 마감(2026-10-15, 시안 E = 전체보기 줄) / 단축키 키 조합 / ~~크기 우선권 넘김 실측~~(파트(크기 가져오기)에서 15초로 바꾸고 실측 15.0~15.5초) / ~~`tabGrace` 60초 만료 실측~~(실측-2(탭 보관 60초), 60.1초)
- 커밋 상태(2026-10-21): 마감 파트는 전부 master에 커밋·머지(파트별 코드 + vault). master 최신 = 파트(빌드본 분리) 커밋(2026-10-22, master에 바로), 그 앞이 파트(리팩터링) `--no-ff` 머지(브랜치 `refactor/socket` 남김). **push는 한 번도 안 함**(분기 태그 `checkpoint/before-folder-tile-uid` = `d50f153`, 태그는 push 때 따로 `git push origin <tag>`). 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치). 커밋은 파트 단위(작업 단위마다 안 함, 사용자가 중간 커밋을 요청하면 그때). user vault 전체 git 제외
- 이전 세션(2026-10-12~15)의 scratchpad 스크립트는 이미 사라짐 — 명세는 `REF-test-scripts.md`(~10-13)·`REF-test-scripts-1014.md`(10-14~)·`REF-test-scripts-1020.md`(10-20~) 날짜별 절. 그때 격리 프로세스·임시 DB는 전부 정리됨. PC 재부팅 뒤라면 `docker start postgres15` -> supervisor -> worker -> `npm run dev`(→ MEMORY "로컬 개발 환경")
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
- **정리 잔여(구)**: `register.go` 주석 SendBuffer 테스트. ~~worker `baseDir`·`instanceKey()`~~ → 정돈-4(작은 잔재)로 삭제(2026-10-21)

### 프론트 user/login — 마감(2026-10-12, → `REF-frontend-auth.md`)
- 서버 쪽 잔여(`SESSION_KEY` env화 · checkSession `createdAt=0`) → 파트(서버 정리)에서 해소(2026-10-13)

---

## 미해결 이슈 (이월)
- **(관찰) 2026-10-21 13:03:43 PC 갑자기 재부팅**: 로그가 경고 없이 끊김(journal "uncleanly shut down"). 그 시각 CPU 98% 유휴·메모리 26%·OOM·커널 패닉·하드웨어 오류 기록 없음 -> 앱 버퍼(화면 복원 ring은 process당 256KB 상한 S5(ring 상한)) 원인 아님. 직전 1초에 RustDesk 정시 재시작(매시 :03, 평소 하루 25번 무사)·gnome-shell 커서 drm 오류(다른 날에도 있음)만 있음 -> 전원·하드웨어·GPU 멈춤 쪽 추정, 재발하면 다시 봄
- **(리팩터링 이후) 기본 보드가 아닌 보드에서 새 보드를 만들 수 없음(2026-10-21 사용자 발견)**: 다른 보드로 이동만 됨. 원인 후보 = `TileFrame.vue` ⋮ 메뉴의 "새 보드로 분리"가 `v-if="isDefaultBoard"` — 의도였는지(파트(보드 분리) 결정) 먼저 확인할 것
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
