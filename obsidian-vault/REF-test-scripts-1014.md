# REF — 확인 스크립트 재작성 명세 (2026-10-14 이후)

> 2026-10-19 `REF-test-scripts.md`에서 분할(10k자 기준, 2026-10-14 절부터 통째로). 그 전 날짜 → `REF-test-scripts.md` / 환경·격리 실행·함정 → `REF-test-env.md` / 이력 → `history/test-scripts-1014.md` / 현재 진행 → `CURRENT.md`.
> scratchpad 스크립트는 재부팅하면 사라진다 — 여기 적힌 항목·선택자가 다시 쓸 때의 명세.

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

## 2026-10-19 터치 스크롤 → `REF-term-touch.md`
- scratchpad `touch/`(재부팅 시 사라짐): `npm i playwright-core@1.56` + 캐시 Chromium(`~/.cache/ms-playwright/chromium-1234`) / `page.html` = node_modules의 `xterm.js`·`xterm.css` 복사 + 300줄, `?fix=1`(rule 적용)·`?tall=1`(페이지 길게) / `run.mjs` = `hasTouch`·`isMobile` 컨텍스트, `TAP=1`이면 `touchscreen.tap`으로 스크롤바 먼저 띄움 -> CDP `Input.dispatchTouchEvent` slider 중심에서 10px씩 20번 위로 -> `term.buffer.active.viewportY`·`pointercancel` 여부·`scrollY` 비교
- `touch/swipe.html`·`swipe.mjs`(11개): `touchScroll.util.ts`를 `npx esbuild --format=iife --global-name=TS`로 묶어 씀, `?fixed=1`(상자 overflow auto·40줄)·`?alt=1`(alt 화면) / 상자 400px(16줄 터미널이 상자보다 크면 slider가 잘림)

## 2026-10-19 파일 전송 → `REF-transfer-e2e.md`
- scratchpad `xfer/`(재부팅 시 사라짐): `ov/main.go`(= `cmd/supervisor/main.go` 복사 + `mountXfer(e, r)`) · `ov/xfer_main.go`(`/xfer/file?key=&path=` `SendFile` / `/xfer/buf?key=&name=&size=&seed=` `SendBuffer` / `/xfer/abort?id=` / `/xfer/jobs` / `/xfer/workers`, 라우터는 인터페이스로 받음 — 반환 타입이 비공개라) · `ov/xfer_debug.go`(router 패키지, `XferWorkers`·`XferJobs` = `FindAll`) → `overlay.json` `Replace`로 3파일 끼워 `go build -overlay`
- 격리: `sup/env`(PORT 5051, `IP=localhost`, `DB_NAME=nexus_xfer_tmp`) · `wrk/env`(`WS_HOST="localhost:5152"`, `NAME="xfer-dev"`) · 둘 다 `TMPDIR=xfer/tmp`(worker 수신 루트 = `os.TempDir()`)
- `proxy.py 5152 5051`: asyncio TCP 프록시, `DELAY`(편도 초) 지연, SIGUSR1 = 지금 연결 전부 `transport.abort()`
- (2026-10-20) `proxy.py` 지연 = 받은 시각 + DELAY 큐로 내보냄(읽을 때마다 sleep이면 큰 프레임에서 대역폭까지 깎임) / 새 시나리오 `t4b`(프록시 kill -> 대기 중 abort)·`t10`(프록시 kill 유지 -> 실패 시간 + 재전송 이어받기) / `SendFile` 수신 이름 = 원본 이름이라 `src/<시나리오>.bin` 하드링크 / T8 비교 기준 = `git show HEAD:...transfer.go`를 overlay로 끼운 `sup_old`
- `test.py t1 ... t9`: 시나리오 이름 = REF 표. 키는 `#`가 있어 `urllib.parse.quote(safe='')` 필수 / `.part` 크기 폴링으로 끊는 시점·이어받기(최소 크기 >= 끊긴 크기) 판정 / worker 끄기는 `pgrep -x worker` 중 `/proc/<pid>/cwd` = `xfer/wrk`만

## 2026-10-19 보드 URL 바인딩 → `REF-tile-url.md`
- scratchpad `url/`: `playwright-core@1.56` + `vite.iso.mts`(`base({mode:'development',command:'serve'})` 펼치고 `root`·`cacheDir`(env `ISO_CACHE`)·`server: {port:3001, strictPort, host:'localhost', hmr: undefined}`), 실행은 `apps/frontend`에서 `VITE_API_URL=http://localhost:5051 npx vite --config ...`
- `url.mjs`(20개): `request` 컨텍스트로 계정(`POST /users`·`/users/session`) + 폴더 7개 + `GET /tiles` 트리에 폴더 타일 `f1`~`f7`(size `{w:0.5,h:1}`, 루트도 같은 크기) `PUT /tiles {tree, version}` -> `storageState`를 브라우저에 / 판정 = `location.search`, `.TileFrame--active`, `document.activeElement.dataset.tileId`, `.TileStrip.scrollLeft`, "화면 왼쪽 칼럼 타일" = strip 안에 보이는 `.TileFrame` 중 left·top 최소
- 함정: URL 쓰기는 300ms 지연 + smooth 스크롤 뒤라 `waitForFunction(location.search...)`로 기다릴 것

## 2026-10-20 페이지 표시 → `REF-board-pager.md`
- scratchpad `pager/`: 격리 supervisor(`go build` + 옆 `env` PORT=5051·IP=127.0.0.1·DB_NAME=`nexus_pager_tmp`, `TMPDIR=<scratchpad>/tmp`로 실행) + `vite.iso.mts`(2026-10-19 절과 같음, 3001) + `playwright-core@1.56`
- `pager.mjs`(31개): 계정 + 폴더 f1~f10 + `PUT /tiles`(루트 size `.5x1`, f1~f10 = `[.5,.5] [.5,.5] [1,1] [.5,.5] [.5,.5] [1,.5] [.5,1] [.5,1] [1,.5] [1,.5]`) -> 실제 배치 = Grid 6개 `[루트 f1 f2] [f3] [f4 f5 좁은] [f6] [f7 f8] [f9 f10]`, 칼럼 시작 scrollLeft = 0,636,...,6360(1280x800)
- 판정: `.TileStrip.scrollLeft` vs `.cell` offsetLeft로 만든 칼럼 목록, `.GridPager .count` 글자, `.pill` 폭(정수 > 15 / 반 칸 < 6), `.TileStrip.scrubbing`, `.GridPager__hint.show`, `.TilePeek.open`, `.TileFrame--active` data-tile-id, `document.activeElement`
- 드래그 = `page.mouse` down -> move(steps) — 첫 6px 전엔 포인터를 안 잡음. 스크롤 끝 대기 = scrollLeft가 60ms x 4번 같을 때
- 폰(390x844 isMobile): 표시줄이 `#app-head-nav` 안·가운데인지만. 터치 드래그는 못 봄

## 2026-10-20 한글 입력 중복 → `REF-term-ime.md`
- scratchpad `ime/`: `index.html` + `main.js`(실제 `xtermIme.util.ts`를 절대 경로 import, xterm 두 개 = 수정 있음/없음, `onData` 수집) + `vite.config.mjs`(vite import 없이 객체만, `resolve.alias` `@xterm/xterm` -> 저장소 `node_modules/@xterm/xterm/lib/xterm.mjs`) -> `apps/frontend`에서 `npx vite build --config <ime>/vite.config.mjs` -> `run.mjs`(node http 정적 서버 5199 + `playwright-core@1.56`, 없는 파일 404 처리 — favicon 요청에 죽음)
- 재생 = 입력창에 `CompositionEvent`·`InputEvent`(inputType·isComposing)를 사용자 로그 순서대로 dispatch, 값은 `ta.value` + `setSelectionRange`로 직접. 묶음 안은 동기, 묶음 사이 20ms(타이머 실행)
- 실브라우저 기록 = 콘솔 스크립트(입력창 이벤트 capture + `WebSocket.prototype.send` 가로채 `PROCESS:INPUT` base64 해독). **앱 프레임은 문자열**(`JSON.stringify`) — 바이너리로만 해독하면 SEND 줄이 안 나옴
