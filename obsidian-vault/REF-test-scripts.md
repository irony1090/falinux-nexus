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
- 파트(모바일 가상 키보드) `pw/kb.mjs`(22개) → `REF-mobile-keyboard.md` "키보드-5(확인) 헤드리스 결과"(가짜 visualViewport 방법 포함)
- 같은 파트 `pw/kb3.mjs`(12개, 라(resizes-content 전환): `setViewportSize`로 창 높이를 줄여 흉내, 다이얼로그는 폴더 `button.nm` -> `.row` `button[title="메뉴"]`) → `REF-mobile-keyboard.md` "라(resizes-content 전환)"
- 같은 파트 `pw/kb2.mjs`(10개, 고침-나(커서 기준 이동)) → `REF-mobile-keyboard.md` "고침-나(커서 기준 이동) 헤드리스 결과"(칼럼 이동·소문자 명령 키 함정)

## 2026-10-14 확인 스크립트 (보드 미리보기) — 재부팅으로 사라짐, 재작성 명세 → 결과는 `REF-board-preview.md` "확인 결과"
- 격리: `go build` supervisor·worker → `iso/sup`·`iso/wrk` + `env`(5051, `nexus_bp_tmp`, worker `NAME=bp-dev`), `TMPDIR=<scratchpad>/iso/tmp`. 프론트 = `vite.iso.mts`: `import base from '<repo>/apps/frontend/vite.config.mts'` → `base({mode:'development',command:'serve'})`를 펼치고 `root`·`cacheDir`·`server: {port:3001, strictPort, host:'localhost'}`만 덮음, 실행은 `apps/frontend`에서 `VITE_API_URL=http://localhost:5051 npx vite --config <경로>`
- `pw/bp.mjs`(31개): API로 계정·폴더 F1·F2(`deviceKey: bp-dev`)·F3(장비 없음)·스크립트(`tick-A` 1초 / `tick-B` 0.5초 / `done.sh` 한 줄) → 기본 보드 트리에 폴더 타일 3개 PUT → 페이지에서 `fetch`로 exec(`X-Tab-Id` = `sessionStorage['nexus.tabId']`) → `POST /workspaces`로 F2·F3 분리 → 보드2에서 exec. 선택자: `.OverviewButton` · `.TileStrip.overview.multi` · `.TileStrip .board`(n번째 = `:nth-of-type`) · `.board-cap b` · `.OverviewTile--terminal .xterm-rows` · `.OverviewTile .pill`. 다시 연 직후엔 1초쯤 기다렸다 읽을 것
- 함정: Playwright `:text-is()`를 `:has()` 안에 쓰면 안 잡힘 → `locator(...).filter({ has: locator('.pill', { hasText: /^RUN$/ }) })`
- `pw/bp2.mjs`: 4초마다 출력하는 터미널로 닫힌 동안 쌓인 줄이 다시 열 때 바로 그려지는지(100·400·1000ms)
- 끌 때: supervisor·worker = `/proc/<pid>/cwd`가 `iso/` 아래인 것 / vite = **`ss -ltnp | grep :3001`의 pid**(`pgrep -af vite`는 자기 셸을 잡음, 2026-10-14 또 겪음)

## 2026-10-15 확인 스크립트 (보드 미리보기 중간안) — 재부팅으로 사라짐, 결과는 `REF-board-preview.md` "확인 결과 (중간-3(확인)"
- 격리는 위 "2026-10-14 보드 미리보기"와 같음(`iso/sup`·`iso/wrk` env, `vite.iso.mts` = `base({mode:'development',command:'serve'})` 펼치고 `server.hmr: undefined`)
- `pw/mid.mjs`(15개): API로 계정·폴더 F1·F2(`bp-dev`)·스크립트 `tick-A`(1초)·`tick-B`(0.5초) → 기본 보드에 폴더 타일 `f1`·`f2` PUT → `POST /workspaces {folderTileId:'f2'}`로 보드2 → 페이지에서 `fetch` exec(A = 기본 보드 `f1`, B = 보드2 `f2`) → **새로고침**(이 탭 등록을 비움)
- 판정 = `page.on('request')`로 `/processes/snapshot/<uidB>` 횟수(등록될 때만 요청). 조회 중 닫기 = `page.route`로 스냅샷 응답을 붙잡고 닫은 뒤 놓기
- 함정: 실행이 길어지면 `tick-B 0`이 화면 위로 지나감 → 줄 번호가 연속인지로 판정 / `page.unroute`가 붙잡은 요청을 먼저 처리해 `route.continue` 오류 → 놓고 기다린 뒤 `unroute` / 보드 전환 등록은 `page.goto`(새로고침)가 아니라 `history.pushState` + `popstate`로

## 2026-10-15 확인 스크립트 (고아 스크립트) — 재부팅으로 사라짐 → 결과 `REF-worker-orphan.md` "고아-3(확인) 결과"
- `chk/client.py`(REST 쿠키 + `/subscribe` 소켓에서 `TAB:ID`) · `chk/orphan.py`(시나리오 E·B·A·C·D, 인자로 골라 실행): 계정·폴더(`deviceKey: orphan-dev`)·스크립트 3개 → worker를 `subprocess.Popen`(`nohup` 선택, `TMPDIR=iso/tmp`)으로 띄우고 `GET /workers?nodeId=`의 `instanceKey`로 exec → 스크립트 찾기 = `/proc/*/cmdline`이 `/bin/sh <iso/tmp>/`로 시작 → 그 sid에 속한 프로세스 수로 잔존 판정
- `GET /processes`는 끝난 process를 안 돌려줌 → 최종 상태는 임시 DB `processes`(status·exit_code·finished_at)로 확인
- 수정 전 비교 바이너리 = `git archive HEAD apps/core | tar -x -C <scratchpad>/old` 후 빌드

## 2026-10-19 확인 스크립트 (Windows worker 윈도-5(OS 표시)) → `REF-windows-exec.md`
- `iso5/os_test.py`(14개, scratchpad, 재부팅 시 사라짐): 격리 supervisor(5051, `nexus_os_tmp`) + 실제 worker(`NAME=os-dev`, `EDITOR="cat -n"`) / 계정 생성·로그인 → `/subscribe` 소켓에서 `TAB:ID`(`d.tabId`) 받아 REST에 `X-Tab-Id`(exec 필수, 없으면 400 "탭 id가 없거나") / 가짜 worker = `websockets.sync.client.connect("/worker")`로 `{"k":0,"id":1,"t":"REGISTER","d":{"key":"os-win","os":"windows"}}` 보내고 응답 `d.subKey`로 authKey / 폴더 `deviceKey`로 장비 지정 / exec = `POST /processes/exec {nodeId, authKey, type}` / 가짜 worker 대상 exec는 파일 전송(FileInit)에 응답이 없어 요청이 멈춤 → 타임아웃 3초 = 검사 통과로 판정 / 끝난 process 상태는 DB `processes.status`로

## 2026-10-19 확인 스크립트 (Windows worker 윈도-6(확인), VM) → `REF-windows-conpty.md`·`REF-windows-exec.md`
- `win/vssh`(scratchpad): `SSH_ASKPASS` + `SSH_ASKPASS_REQUIRE=force`로 비밀번호(`admin`) 넣는 ssh 래퍼(sshpass·paramiko 없음), `-p 2222`, known_hosts는 scratchpad 전용. 출력은 CP949(`iconv -f cp949`), worker 로그는 UTF-8
- worker 띄우기 = shared의 `start.ps1 [exe]`(`\\host.lan\Data` -> `C:\nexus` 복사 + `Invoke-CimMethod Win32_Process Create`로 ssh 세션 밖에서 실행, 출력 `C:\nexus\worker.log`). ssh 세션에서 직접 띄우면 세션 종료 때 같이 꺼짐 / 따옴표 섞인 긴 명령은 멈출 수 있어 `.ps1` 파일로
- `iso6/win_test.py [ps bat py noext ctrlc force edit]`(18개, `py` 빼고 돌리면 17개 — VM에 Python 없으면 `py`는 멈춤): 격리 supervisor `IP=172.17.0.1`(VM에서 보이는 docker0 주소) / 출력 확인 = 실행 중에 `GET /processes/snapshot/:uid`(`data` base64) — 스크립트 끝에 `Start-Sleep`을 둬서 끝나기 전에 읽음 / 상태·exit code = DB / 입력 = 구독 소켓 `PROCESS:INPUT` / 손자 확인 = `tasklist /fi "imagename eq PING.EXE"`
