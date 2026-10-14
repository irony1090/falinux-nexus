# REF — 확인 스크립트 재작성 명세 (날짜별)

> 2026-10-14 `REF-test-env.md`에서 분할(10k자 기준). 환경·격리 실행·함정 → `REF-test-env.md` / 이력 → `history/test-scripts.md` / 현재 진행 → `CURRENT.md`.
> scratchpad 스크립트는 재부팅하면 사라진다 — 여기 적힌 항목·선택자가 다시 쓸 때의 명세.

## 2026-10-08 확인 스크립트 요령 (공유-f·터미널 빈틈 확인)
- Python `Client` 한 클래스로 REST(쿠키 `http.cookiejar`, `User-Agent` 지정, exec 땐 `X-Tab-Id`) + 소켓(`/subscribe?tabId=`, `Cookie`·`User-Agent` 헤더, 읽기 스레드에서 `TAB:ID`로 탭 id 받기, `TAB:VIEW` 보내기) — 같은 계정의 두 번째 세션 = 같은 아이디로 `POST /users/session`을 한 번 더
- 확인 스크립트끼리 재사용: 앞 파일의 `Client` 정의 부분만 `exec`해서 씀(import 대신 — 파일이 맨 아래에서 바로 실행되므로)
- 헤드리스에서 exec할 탭 id = 페이지 `sessionStorage`에서 이름에 `tab`이 든 키
- "위치 없음"(받는 쪽에 터미널 타일 없음) 만들기 = 받는 쪽 `GET /tiles` → 터미널 타일과 부모 `kids`에서 지움 → `PUT /tiles {tree, version}`(기존 uid를 지우는 건 허용, 추가만 막힘 — WS5(터미널 타일 추가 주체))
- 격리 worker는 `NAME`을 새로 정하고(`presence-dev` 등) 폴더 `deviceKey`를 같게. 다시 띄울 땐 `worker.db`를 지우고
- 끌 때: supervisor·worker는 `/proc/<pid>/cwd`가 격리 폴더인 것만, vite는 명령줄에 `vite.iso.mts`가 든 것만
- 60초 보관(`tabGrace`) 만료·`lastSeen` 시각 정확도는 기다려야 해서 아직 확인 안 함

## 2026-10-12 확인 스크립트 (active 타일·원격 배지) — 재부팅으로 사라짐, 재작성 명세
- playwright: 이전 세션 scratchpad의 `pw/node_modules`를 심볼릭 링크해서 썼음 → 재부팅 뒤엔 scratchpad `pw/`에서 `npm i playwright-core@1.56.1`. 브라우저는 `~/.cache/ms-playwright/chromium-1234/chrome-linux64/chrome`(재부팅에도 남음)
- `active.mjs`(로컬 active 14개): 사용자 5050·3000 위에서 새 계정 + `POST /nodes` 폴더 11개 → 루트 타일 "새 타일" 11번 → 상태 읽기 = `.TileFrame--active .pos` 글자·`.TileFrame--dim` 수·`document.activeElement`·strip 안에 보이는지. 키 = `Control+Alt+Digit<n>`, 두 자리는 `keyboard.down('Control')`·`down('Alt')` 후 `press('Digit1')` `press('Digit2')`. keydown 기록에서 `Alt`·`Control` 키 자체는 빼고 셀 것. 먼 Grid 스크롤은 1초 이상 기다림
- `remote.mjs`(원격 배지 11개): 격리(임시 DB + 5051 + worker `NAME=active-dev` + 프론트 3001) / 계정 A·B·C(닉네임 첫 글자가 배지 글자) / A: 폴더(`deviceKey: active-dev`) + 스크립트 `#!/bin/sh\ncat` + `PUT /nodes/:id/shares/:bId {permission:10}` / UI로 A 실행 → `GET /processes`로 uid → B `POST /processes/:uid/tile` / 배지 읽기 = `.TileFrame__head .ActiveBadges .badge` 글자 / 공유 취소 `DELETE /nodes/:id/shares/:bId`
- 끌 때 함정: `pgrep -f "vite.iso.mts"` 루프로 vite가 안 꺼졌음 → `ss -ltnp | grep :3001`로 pid 보고 `/proc/<pid>/cmdline` 확인 뒤 kill

## 2026-10-12 확인 스크립트 (명부) — 재부팅으로 사라짐, 재작성 명세
- 같은 장비 키 worker 여러 개 = 폴더마다 바이너리 + `env`(`NAME` 같게, `SQLITE`만 다르게). 하나 끄면 "접속 n / 끊김 1" 상태가 만들어짐
- `roster_api.py`(API 10개): 비로그인 401 · 명부 개수·online 수 · 필드 6개 · `?mainKey=` 걸러짐 · 접속 중 삭제 409 · 없는 것 404 · 화면용 폴더(`deviceKey`)·스크립트 생성
- `pw/roster.mjs`(화면 11개): 행 칩 `.dev .cnt` 글자·`title` 줄 수 → 폴더 열기 → 경로 줄 칩 → 장비 지정 다이얼로그(`.dev .cnt`·`.inst` 수·삭제 버튼 수) → 실행 → 고르기 창(`.v-list-item--disabled` 1개) → 삭제 확인 → 삭제 뒤 `.inst` 2개·칩 `2/2`
- **vite 끄기 함정**: `/proc/<pid>/cmdline`은 NUL 구분이라 `grep -q` 루프가 안 잡혔음 → `pgrep -af vite`로 `vite.iso.mts`가 든 pid를 보고 직접 kill

## 2026-10-12 확인 스크립트 (로그인) — 재부팅으로 사라짐, 재작성 명세
- `pw/login.mjs`(12개): 새 컨텍스트 `/` → `/login`(query 없음) · 회원가입 버튼 없음 · `/w/<hash>` → `/login?redirect=/w/<hash>` → 폼 로그인(`input[name="idenfitication"]`·`input[name="password"]`, 버튼 "로그인") → 원래 주소 · 새로고침 유지 · 로그인 상태 `/login` → `/` · `?redirect=//evil` → `/` · 헤더 `button[title="로그아웃"]`(2026-10-13 햄버거 아이콘에서 바꿈, 옛 선택자 `.v-app-bar-nav-icon`) → 확인 창 "로그아웃" → `/login` + `localStorage nexus.user.*` 삭제 · `ctx.clearCookies()` 뒤 폴더 타일 새로고침(`button[title="새로고침"]`) → 401 → `/login` + 설정 삭제 · supervisor kill(pid 인자) 뒤 새로고침 → "서버에 연결할 수 없습니다" + 설정 유지
- `pw/logout_board.mjs`: 보드 행을 DB에 직접 INSERT해 `/w/<hash>`에서 로그아웃 → 알림 없음·redirect 유지 / `pw/board_removed.mjs`: 다른 세션이 `DELETE /workspaces/:hash` → 알림 + `/`
- **함정**: 보드 행을 만들 때 기본 보드 트리를 그대로 복사하면 타일 `id`·`uid`가 겹쳐 보드 삭제가 500 → 복사할 땐 새 id·uid로
- **함정**: 격리 supervisor는 `TMPDIR=<scratchpad>/iso/tmp`(폴더가 있어야 함) — 없는 경로면 `initEnv`에서 panic

## 2026-10-12 확인 스크립트 (종료 정책) → `REF-process-kill.md` "종료-6(확인) 결과"(socat 경유 격리 worker로 끊김 재현 포함)

## 2026-10-12 확인 스크립트 (폴더 목록 동기화) → `REF-node-list-sync.md` "동기화-3(확인) 결과"(worker 없이 supervisor·프론트만, 재연결 = 격리 supervisor 재시작)
- 함정: 같은 계정 두 브라우저는 타일 배치가 동기화돼 한쪽이 폴더로 들어가면 다른 쪽도 따라감

## 2026-10-13 확인 스크립트 (서버 정리) → `history/supervisor-web.md` 2026-10-13(API 11개 + 세션 키 교체, worker 없음)

## 2026-10-13 확인 스크립트 (질의 응답 중복) → `history/term-query.md` 응답-3(확인)(bash 시험 스크립트·대조 실행 방법 포함)

## 2026-10-13 확인 스크립트 (노드 클립보드) → `history/node-clipboard.md` 2026-10-13(2)(선택자 포함, worker 없음)

## 2026-10-13 확인 스크립트 (선택 모드) → `history/node-select.md` 2026-10-13(4)(선택자·드래그 방법 포함, worker 없음)

## 2026-10-13 확인 스크립트 (크기 가져오기) → `REF-size-take.md` "크기-4(확인) 결과"(worker 포함, 같은 계정 3탭 + 다른 계정)

## 2026-10-14 확인 스크립트 (파트(확인·실측))
- 공용 `chk/client.py`: `Client`(쿠키 REST, `X-Tab-Id`) + `Tab`(소켓 하나 = 탭, 읽기 스레드가 `TAB:ID`·`DATA`를 모음, `view`·`input`·`emit`)
- 실측-1(PENDING 입력 버림) `chk/i3.py` → `REF-process-input.md` "실측-1 결과" / 실측-2(탭 보관 60초) `chk/grace.py` → `REF-process-sync-impl.md` "실측-2 결과" / 실측-3(htop 줄 밀림) `pw/htop*.mjs`·`chk/order.py` → `REF-size-take.md` "실측-3 결과"
