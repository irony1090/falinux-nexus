# REF — 테스트 환경 · 확인용 클라이언트 함정

> 2026-10-06 CURRENT에서 분리(세션마다 읽을 필요 없는 재사용 지식). 이력 → `history/test-env.md` / 현재 진행 → `CURRENT.md`. 각 기능의 확인 항목은 해당 REF의 확인 절(예: `REF-process-snapshot-impl.md` "스냅샷-b/-d", `REF-node-ui-save-impl.md`, `REF-node-ui-sync.md`, `REF-process-input.md` "입력-a/입력-c").
> 2026-10-14 날짜별 확인 스크립트 재작성 명세는 `REF-test-scripts.md`로 분할(이 파일 = 환경·격리 실행·클라이언트 함정·DB 상태).

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

- **(2026-10-13 바뀜)** 소켓 주소가 `VITE_API_URL`에서 만들어지고 `vite.config.mts`가 함수형이 됨 → 아래 transform 플러그인 불필요, iso 설정은 `base({mode:'development', command:'serve'})`를 펼치고 `.env`의 `DEV_*`를 덮을 것(→ `REF-external-access.md`)
- **격리 프론트(헤드리스용, 2026-10-07)**: 소켓 주소가 `websocket.hook.ts`에 `localhost:5050`으로 박혀 있어 env로 못 바꿈 → scratchpad에 `vite.iso.mts`(저장소 `vite.config.mts`를 절대 경로로 import해 펼치고 `root` = `apps/frontend`, `cacheDir` = scratchpad, `server.port` 3001 + `strictPort`, `transform` 플러그인으로 `websocket.hook.ts`의 `localhost:5050` → `5051`)를 두고 `apps/frontend`에서 `VITE_API_URL=http://localhost:5051 npx vite --config <경로>`(이미 있는 env가 `.env`보다 우선). 쿠키는 포트를 안 가리므로 `context.request`로 5051에 로그인하면 3001 페이지에서 그대로 쓰임. 저장소에 파일 안 생김
- **끌 때 함정**: `pgrep -f '<패턴>'`의 패턴이 지금 돌리는 bash 명령줄에도 들어 있으면 자기 셸까지 죽인다(exit 144). vite는 `pgrep -af vite`로 pid를 먼저 보고 끈다
- **supervisor 바이너리는 어떤 인자로 실행해도 서버가 뜬다**(`--help` 없음) + `init()`이 마이그레이션을 적용한다 → 확인 중에 `go run ./cmd/supervisor`를 실행하지 말 것. 빌드 확인은 `go build`만(2026-10-07 사고: 개발 DB에 00007·00008 적용됨)

- **격리 프론트 설정 함정(2026-10-08)**: scratchpad의 `vite.iso.mts`에서 `import { ... } from 'vite'`는 못 찾음(scratchpad에 vite 없음) → `import base from '<repo>/apps/frontend/vite.config.mts'` 후 `{ ...base, root, cacheDir, server: { ..., port: 3001, strictPort: true }, plugins: [...base.plugins, 소켓 주소 교체 플러그인] }`로 객체만 펼쳐 export
- **계정 동기화 함정**: 소유자 보드에는 타일이 이미 열려 있다(타일 트리가 계정 단위로 서버 저장) → 확인 스크립트에서 "새 타일"을 다시 누르면 같은 폴더 타일이 2개가 돼 `.tool.share` 같은 선택자가 2개로 잡힘(strict mode 위반)
- **헤드리스 확인 계정 만들기**: `POST /users {identification, nickname, password}` → `POST /users/session`; playwright `request.newContext().storageState()`를 브라우저 컨텍스트에 넘기면 로그인 상태
- **socat 끊김 재현 함정(2026-10-14)**: `TCP-LISTEN:...,fork`는 연결마다 자식을 만든다 -> 리스너만 끄면 연결이 안 끊긴다. 명령줄 첫 인자가 socat이고 `LISTEN:<포트>`가 든 프로세스를 자식까지 전부 끈다(`/proc/*/cmdline` NUL 분리로 읽음)
- **pkill 자기 셸 함정(2026-10-14)**: heredoc으로 만든 스크립트 안에 `TCP-LISTEN:5152` 같은 문자열이 있으면 `pkill -f 'TCP-LISTEN:515[2]'`가 그 heredoc을 품은 bash 명령줄까지 잡는다(exit 144) -> 포트·`/proc` 기준으로 pid를 골라 끈다
- `GET /users/presence`는 공유 대상 계정만 `?ids=`로 — 자기 계정 `lastSeen`은 `GET /processes/:uid/presence`의 계정 항목으로 본다

## 테스트 클라이언트 함정
- 테스트 클라이언트 함정 모음: `PATCH /nodes/:id` 본문은 필드마다 `{valid: true, value: x}`(`internal/patch`, 그냥 값이면 400) / Vuetify 다이얼로그 선택자는 `.v-dialog.v-overlay--active`(두 클래스가 같은 요소) / 소켓 프레임 binary / exec는 `X-Tab-Id` 필수 / STATUS `status` 숫자 / 브라우저 쿠키 도메인 `localhost` / `users.identification`은 해시 저장 / 화면 밖 Grid xterm은 `innerText`에 안 잡힘(`.xterm-rows` `textContent`) / API로 만든 노드는 열린 목록에 안 뜸(`[title="새로고침"]`) / Vite가 heredoc으로 막 만든 파일을 빈 내용으로 캐시할 수 있음(`touch`) / 세션 쿠키 값에 `=` 있을 수 있음(첫 `=`에서 자를 것) / exec 직후 크기 0x0이라 resize 먼저(소유 탭) / vim 테스트는 파일 이름을 매번 다르게(swap 파일)

## DB 상태 (2026-10-02 기준)
- DB에 남은 테스트 데이터: 계정 `input-test-b`(타일 트리 행 있음, 루트만) · `tile-race-<시각>` 1개 · `tile-store-`/`tile-exec-`/`tile-restore-`/`tile-other-`/`dbg-`/`push-`/`push-other-`/`sync-`/`sync2-<시각>` 여러 개(3-c·3-d·4 확인용) + `snap-`/`snap-other-`/`snap-probe-`/`snapb-<이름>-<시각>`(4.5 확인용, 노드는 삭제). irony 타일 트리 = 루트만(version 12)
- DB 노드 현황: 폴더 id 1(`HTOP_TEST_SH_MODI`, `device_key=irony-MAC-ADDress1`) 안에 스크립트 2~4 / 폴더 `test`(14) > `ttt1`(15) > 스크립트 23(사용자 테스트용)

