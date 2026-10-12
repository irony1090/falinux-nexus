# REF — 테스트 환경 · 확인용 클라이언트 함정

> 2026-10-06 CURRENT에서 분리(세션마다 읽을 필요 없는 재사용 지식). 이력 → `history/test-env.md` / 현재 진행 → `CURRENT.md`. 각 기능의 확인 항목은 해당 REF의 확인 절(예: `REF-process-snapshot-impl.md` "스냅샷-b/-d", `REF-node-ui-save-impl.md`, `REF-node-ui-sync.md`, `REF-process-input.md` "입력-a/입력-c").

## 환경
- 환경: supervisor·worker는 사용자가 띄움(`docker start postgres15` → supervisor → worker, dev 서버 3000). 테스트 로그인 = `pages/Login.vue` 기본값 계정. 헤드리스 = `~/.cache/ms-playwright/chromium-1234/chrome-linux64/chrome` + `playwright-core`(scratchpad에 `npm i`). worker 연결만 끊는 재현 = `REF-process-reconnect.md` "재현 방법"
- 스크립트는 직접 실행이라 `#!/bin/sh` 등 shebang 필수(없으면 `exec format error`)
- 재부팅으로 사라지는 것: 세션 scratchpad의 playwright-core + 테스트 스크립트(4.5 `snap/server.mjs`·`browser.mjs`·`probe.mjs`, 이전 `tiles/*.mjs`·`input/*.mjs`). 다시 쓸 땐 `REF-process-snapshot-impl.md` "스냅샷-b/-d"·`REF-node-ui-save-impl.md`·`REF-node-ui-sync.md` 각 확인 절(항목·함정), `REF-process-input.md` "입력-a/입력-c" 보고 재작성. 테스트는 계정을 새로 만들고 폴더 `deviceKey=irony-MAC-ADDress1`로 하면 irony 타일 트리를 안 건드림

## 격리 실행(2026-10-07, 개발 DB·5050을 건드리지 않음)
- supervisor·worker 바이너리는 **dev 판정(`IsDev` = 실행 파일 경로에 `os.TempDir()`·`go-build` 포함)이 아니면 실행 파일 옆 `env`로 설정을 덮어씀**(`godotenv.Overload`). scratchpad는 `/tmp` 아래라 dev로 판정됨 → 실행할 때 `TMPDIR=<scratchpad>/tmp`로 바꿔서 dev 판정을 피함
- 절차: `go build -o <dir>/sup/supervisor ./cmd/supervisor` + 옆에 `env`(`PORT=5051`, `DB_NAME=nexus_authz_tmp`, 나머지 기존 값) / worker도 같은 방식(`WS_HOST="localhost:5051"`, `SQLITE=<dir>/wrk/worker.db`, `NAME="authz-dev"`) / `CREATE DATABASE nexus_authz_tmp OWNER irony;`(기본 DB `workspace`로 접속해서) → 실행 → 확인 뒤 프로세스 종료 + `DROP DATABASE`
- 소켓 클라이언트: Python `websockets`(user site에 있음 → `python3 -I`로 돌리면 못 찾음). 프레임 = JSON `{k,id,t,d}`, k: 0 REQ / 1 RES / 2 EVENT, 입력 `PROCESS:INPUT` `{uid, data(base64)}`. 입력 확인용 스크립트 = `#!/bin/sh\ncat`(받은 입력이 출력으로 돌아옴)
- 크기 우선권 넘김 확인은 `tabGrace` 60초를 기다려야 함
- 끌 때는 **실행 위치(`/proc/<pid>/cwd`)가 격리 폴더인 프로세스만** 끈다(`pgrep -f '^\./supervisor$'`만으로 끄면 사용자가 띄운 것까지 끌 수 있음)
- 확인 스크립트(세션 scratchpad, 재부팅 시 사라짐): `authz_test.py`(공유-b) · `share_test.py`(공유-c, 앞 것을 import) · `ws_test.py`(작업공간-b) · `wsc_test.py`(작업공간-c, worker 포함) — 여기까지 지난 세션 것, 사라짐 / 작업공간-e: `pw/lib.mjs`·`pw/board.mjs`(헤드리스 보드) · `reg/share_reg.py`(공유-b·c 회귀 합본 59개, 공유 행은 공유-c API로)
- worker도 격리: `go build -o <dir>/wrk/worker ./cmd/worker` + 옆 `env`(`WS_HOST="localhost:5051"`, `SQLITE=<dir>/wrk/worker.db`, `NAME="ws-dev"`) → 노드 폴더 `deviceKey`를 그 NAME으로, authKey = `GET /workers?nodeId=` 의 `instanceKey`
- 소켓 확인은 `websockets.sync.client.connect`(17.1)가 편함: 스레드에서 프레임 읽기 + `ws.send(json.dumps({"k":2,"t":"TAB:VIEW","d":{...}}).encode())`(브라우저 → 서버 EVENT, binary)
- **DB 쿼리만 확인할 때 저장소에 파일을 두지 않는 법**: scratchpad의 `main.go`를 `go build -overlay overlay.json`으로 `cmd/<임의>/main.go` 자리에 가상으로 끼워 빌드(내부 패키지 import 가능) + goose CLI는 migrations의 `.sql`만 scratchpad로 복사해서 실행(`embed.go`가 있는 폴더를 그대로 넘기면 `.go`를 Go 마이그레이션으로 볼 수 있어 피함 — 직접 확인은 안 함)

- **격리 프론트(헤드리스용, 2026-10-07)**: 소켓 주소가 `websocket.hook.ts`에 `localhost:5050`으로 박혀 있어 env로 못 바꿈 → scratchpad에 `vite.iso.mts`(저장소 `vite.config.mts`를 절대 경로로 import해 펼치고 `root` = `apps/frontend`, `cacheDir` = scratchpad, `server.port` 3001 + `strictPort`, `transform` 플러그인으로 `websocket.hook.ts`의 `localhost:5050` → `5051`)를 두고 `apps/frontend`에서 `VITE_API_URL=http://localhost:5051 npx vite --config <경로>`(이미 있는 env가 `.env`보다 우선). 쿠키는 포트를 안 가리므로 `context.request`로 5051에 로그인하면 3001 페이지에서 그대로 쓰임. 저장소에 파일 안 생김
- **끌 때 함정**: `pgrep -f '<패턴>'`의 패턴이 지금 돌리는 bash 명령줄에도 들어 있으면 자기 셸까지 죽인다(exit 144). vite는 `pgrep -af vite`로 pid를 먼저 보고 끈다
- **supervisor 바이너리는 어떤 인자로 실행해도 서버가 뜬다**(`--help` 없음) + `init()`이 마이그레이션을 적용한다 → 확인 중에 `go run ./cmd/supervisor`를 실행하지 말 것. 빌드 확인은 `go build`만(2026-10-07 사고: 개발 DB에 00007·00008 적용됨)

- **격리 프론트 설정 함정(2026-10-08)**: scratchpad의 `vite.iso.mts`에서 `import { ... } from 'vite'`는 못 찾음(scratchpad에 vite 없음) → `import base from '<repo>/apps/frontend/vite.config.mts'` 후 `{ ...base, root, cacheDir, server: { ..., port: 3001, strictPort: true }, plugins: [...base.plugins, 소켓 주소 교체 플러그인] }`로 객체만 펼쳐 export
- **계정 동기화 함정**: 소유자 보드에는 타일이 이미 열려 있다(타일 트리가 계정 단위로 서버 저장) → 확인 스크립트에서 "새 타일"을 다시 누르면 같은 폴더 타일이 2개가 돼 `.tool.share` 같은 선택자가 2개로 잡힘(strict mode 위반)
- **헤드리스 확인 계정 만들기**: `POST /users {identification, nickname, password}` → `POST /users/session`; playwright `request.newContext().storageState()`를 브라우저 컨텍스트에 넘기면 로그인 상태

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

## 테스트 클라이언트 함정
- 테스트 클라이언트 함정 모음: `PATCH /nodes/:id` 본문은 필드마다 `{valid: true, value: x}`(`internal/patch`, 그냥 값이면 400) / Vuetify 다이얼로그 선택자는 `.v-dialog.v-overlay--active`(두 클래스가 같은 요소) / 소켓 프레임 binary / exec는 `X-Tab-Id` 필수 / STATUS `status` 숫자 / 브라우저 쿠키 도메인 `localhost` / `users.identification`은 해시 저장 / 화면 밖 Grid xterm은 `innerText`에 안 잡힘(`.xterm-rows` `textContent`) / API로 만든 노드는 열린 목록에 안 뜸(`[title="새로고침"]`) / Vite가 heredoc으로 막 만든 파일을 빈 내용으로 캐시할 수 있음(`touch`) / 세션 쿠키 값에 `=` 있을 수 있음(첫 `=`에서 자를 것) / exec 직후 크기 0x0이라 resize 먼저(소유 탭) / vim 테스트는 파일 이름을 매번 다르게(swap 파일)

## DB 상태 (2026-10-02 기준)
- DB에 남은 테스트 데이터: 계정 `input-test-b`(타일 트리 행 있음, 루트만) · `tile-race-<시각>` 1개 · `tile-store-`/`tile-exec-`/`tile-restore-`/`tile-other-`/`dbg-`/`push-`/`push-other-`/`sync-`/`sync2-<시각>` 여러 개(3-c·3-d·4 확인용) + `snap-`/`snap-other-`/`snap-probe-`/`snapb-<이름>-<시각>`(4.5 확인용, 노드는 삭제). irony 타일 트리 = 루트만(version 12)
- DB 노드 현황: 폴더 id 1(`HTOP_TEST_SH_MODI`, `device_key=irony-MAC-ADDress1`) 안에 스크립트 2~4 / 폴더 `test`(14) > `ttt1`(15) > 스크립트 23(사용자 테스트용)

## 2026-10-12 확인 스크립트 (명부) — 재부팅으로 사라짐, 재작성 명세
- 같은 장비 키 worker 여러 개 = 폴더마다 바이너리 + `env`(`NAME` 같게, `SQLITE`만 다르게). 하나 끄면 "접속 n / 끊김 1" 상태가 만들어짐
- `roster_api.py`(API 10개): 비로그인 401 · 명부 개수·online 수 · 필드 6개 · `?mainKey=` 걸러짐 · 접속 중 삭제 409 · 없는 것 404 · 화면용 폴더(`deviceKey`)·스크립트 생성
- `pw/roster.mjs`(화면 11개): 행 칩 `.dev .cnt` 글자·`title` 줄 수 → 폴더 열기 → 경로 줄 칩 → 장비 지정 다이얼로그(`.dev .cnt`·`.inst` 수·삭제 버튼 수) → 실행 → 고르기 창(`.v-list-item--disabled` 1개) → 삭제 확인 → 삭제 뒤 `.inst` 2개·칩 `2/2`
- **vite 끄기 함정**: `/proc/<pid>/cmdline`은 NUL 구분이라 `grep -q` 루프가 안 잡혔음 → `pgrep -af vite`로 `vite.iso.mts`가 든 pid를 보고 직접 kill

## 2026-10-12 확인 스크립트 (로그인) — 재부팅으로 사라짐, 재작성 명세
- `pw/login.mjs`(12개): 새 컨텍스트 `/` → `/login`(query 없음) · 회원가입 버튼 없음 · `/w/<hash>` → `/login?redirect=/w/<hash>` → 폼 로그인(`input[name="idenfitication"]`·`input[name="password"]`, 버튼 "로그인") → 원래 주소 · 새로고침 유지 · 로그인 상태 `/login` → `/` · `?redirect=//evil` → `/` · `.v-app-bar-nav-icon` → 확인 창 "로그아웃" → `/login` + `localStorage nexus.user.*` 삭제 · `ctx.clearCookies()` 뒤 폴더 타일 새로고침(`button[title="새로고침"]`) → 401 → `/login` + 설정 삭제 · supervisor kill(pid 인자) 뒤 새로고침 → "서버에 연결할 수 없습니다" + 설정 유지
- `pw/logout_board.mjs`: 보드 행을 DB에 직접 INSERT해 `/w/<hash>`에서 로그아웃 → 알림 없음·redirect 유지 / `pw/board_removed.mjs`: 다른 세션이 `DELETE /workspaces/:hash` → 알림 + `/`
- **함정**: 보드 행을 만들 때 기본 보드 트리를 그대로 복사하면 타일 `id`·`uid`가 겹쳐 보드 삭제가 500 → 복사할 땐 새 id·uid로
- **함정**: 격리 supervisor는 `TMPDIR=<scratchpad>/iso/tmp`(폴더가 있어야 함) — 없는 경로면 `initEnv`에서 panic

## 2026-10-12 확인 스크립트 (종료 정책) → `REF-process-kill.md` "종료-6(확인) 결과"(socat 경유 격리 worker로 끊김 재현 포함)
