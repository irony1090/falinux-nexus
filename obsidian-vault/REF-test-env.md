# REF — 테스트 환경 · 확인용 클라이언트 함정

> 2026-10-06 CURRENT에서 분리(세션마다 읽을 필요 없는 재사용 지식). 이력 → `history/test-env.md` / 현재 진행 → `CURRENT.md`. 각 기능의 확인 항목은 해당 REF의 확인 절(예: `REF-process-snapshot-impl.md` "스냅샷-b/-d", `REF-node-ui-save-impl.md`, `REF-node-ui-sync.md`, `REF-process-input.md` "입력-a/입력-c").
> 2026-10-14 날짜별 확인 스크립트 재작성 명세는 `REF-test-scripts.md`로 분할(2026-10-14 이후는 `REF-test-scripts-1014.md`, 2026-10-20 이후는 `REF-test-scripts-1020.md`, 2026-10-22 이후는 `REF-test-scripts-1022.md`, 2026-10-23 경로 직접 입력부터 `REF-test-scripts-1023.md`)(이 파일 = 환경·클라이언트 함정·DB 상태, 격리 실행 = `REF-test-isolation.md`).

## 환경
- **끌기·색인 금지·업데이트 app 확인 = `scripts/check/run.sh`**(2026-11-02, 저장소에 있음, 임시 DB `nexus_check_tmp` + 5051 + 임시 `UPDATE_STORE_DIR`, 끝나면 지움). 업데이트 app = `updates-api.py`·`updater-ui.mjs`(→ `REF-update-p1.md`). 격리 supervisor `-env`는 절대 경로
- 환경: supervisor·worker는 사용자가 띄움(`docker start postgres15` → supervisor → worker, dev 서버 3000). 테스트 로그인 = `pages/Login.vue` 기본값 계정. 헤드리스 = `~/.cache/ms-playwright/chromium-1234/chrome-linux64/chrome` + `playwright-core`(scratchpad에 `npm i`). worker 연결만 끊는 재현 = `REF-process-reconnect.md` "재현 방법"
- 스크립트는 직접 실행이라 `#!/bin/sh` 등 shebang 필수(없으면 `exec format error`)
- 재부팅으로 사라지는 것: 세션 scratchpad의 playwright-core + 테스트 스크립트(4.5 `snap/server.mjs`·`browser.mjs`·`probe.mjs`, 이전 `tiles/*.mjs`·`input/*.mjs`). 다시 쓸 땐 `REF-process-snapshot-impl.md` "스냅샷-b/-d"·`REF-node-ui-save-impl.md`·`REF-node-ui-sync.md` 각 확인 절(항목·함정), `REF-process-input.md` "입력-a/입력-c" 보고 재작성. 테스트는 계정을 새로 만들고 폴더 `deviceKey=irony-MAC-ADDress1`로 하면 irony 타일 트리를 안 건드림
- **실제 계정 화면(test1 -> test3 API, irony) 헤드리스 확인 = 보기만**(2026-10-25 찰떡 UI 때 두 번 실수: `button:has(.mdi-play)`가 칸 "실행"에 걸려 스크립트 실행 · 공유 버튼이 늦게 그려져 좌표 클릭이 "장비 디렉토리 열기"에 걸림): 좌표 클릭 금지, 선택자는 클래스·title로 정확히(`button.dg-lbl:has(.mdi-play)`, `button[title="이 폴더 공유"]`), `.RootTabs` 대기 + 2초 뒤 누름. 로그인 = `.env.development.local` 주석 줄 계정으로 test3 `/users/session` -> `storageState`

- **로그인 기본값(2026-10-19)**: `.env.development.local`의 `VITE_DEV_LOGIN_ID/PW` 줄이 주석 처리돼 있어 폼이 빔 -> 헤드리스 확인은 격리 supervisor에 계정을 만들어 씀(→ `REF-test-scripts-1014.md` "2026-10-19 보드 URL 바인딩")
- **8012가 다른 작업공간에 잡혀 있을 때(2026-11-11)**: community 프론트가 8012를 쓰면 dev supervisor가 `bind: address already in use`. 그 프로세스는 건드리지 말고 `dev.env` 사본(`PORT=5051`, 600)으로 `go run . -env <사본 절대 경로>` + vite `VITE_API_URL=http://localhost:5051 npm run dev`(CORS 열림, 소켓은 안 붙음 — 배치·화면 확인은 됨). 이 PC 브라우저 `http://localhost:8011`에서만 보임
- **개발 화면 띄우기(2026-11-05)**: `apps/frontend/.env`의 `VITE_API_URL` 줄이 주석이라 그냥 `npm run dev`면 화면이 같은 출처(8011)의 `/api`를 불러 HTML을 받음(vite에 프록시 없음). `.env`를 안 고치고 `VITE_API_URL="https://test2.falinux.com" npm run dev`(실행 환경변수가 `.env`보다 우선) -> https://test1.falinux.com 접속. 테스트 계정·worker = `POST /api/users` -> `/users/session` -> `POST /api/devices {mainKey}`(응답 `apiKey`) -> worker env(`WS_HOST="localhost:8012"`·`SQLITE`·`NAME`·`API_KEY`, 600) `go run ./cmd/worker -env <절대 경로>` -> 폴더 `POST /nodes {kind: FOLDER, deviceKey}` + 기본 보드 타일

## 격리 실행 → `REF-test-isolation.md`(2026-11-11 분할: 임시 DB + 5051 + 격리 프론트 3001, 끌 때 함정·`pgrep -f` 오판 함정 포함)

## 테스트 클라이언트 함정
- 테스트 클라이언트 함정 모음: `PATCH /nodes/:id` 본문은 필드마다 `{valid: true, value: x}`(`internal/patch`, 그냥 값이면 400) / Vuetify 다이얼로그 선택자는 `.v-dialog.v-overlay--active`(두 클래스가 같은 요소) / 소켓 프레임 binary / exec는 `X-Tab-Id` 필수 / STATUS `status` 숫자 / 브라우저 쿠키 도메인 `localhost` / `users.identification`은 해시 저장 / 화면 밖 Grid xterm은 `innerText`에 안 잡힘(`.xterm-rows` `textContent`) / API로 만든 노드는 열린 목록에 안 뜸(`[title="새로고침"]`) / Vite가 heredoc으로 막 만든 파일을 빈 내용으로 캐시할 수 있음(`touch`) / 세션 쿠키 값에 `=` 있을 수 있음(첫 `=`에서 자를 것) / exec 직후 크기 0x0이라 resize 먼저(소유 탭) / vim 테스트는 파일 이름을 매번 다르게(swap 파일)

## DB 상태 (2026-10-02 기준)
- DB에 남은 테스트 데이터: 계정 `input-test-b`(타일 트리 행 있음, 루트만) · `tile-race-<시각>` 1개 · `tile-store-`/`tile-exec-`/`tile-restore-`/`tile-other-`/`dbg-`/`push-`/`push-other-`/`sync-`/`sync2-<시각>` 여러 개(3-c·3-d·4 확인용) + `snap-`/`snap-other-`/`snap-probe-`/`snapb-<이름>-<시각>`(4.5 확인용, 노드는 삭제). irony 타일 트리 = 루트만(version 12)
- DB 노드 현황: 폴더 id 1(`HTOP_TEST_SH_MODI`, `device_key=irony-MAC-ADDress1`) 안에 스크립트 2~4 / 폴더 `test`(14) > `ttt1`(15) > 스크립트 23(사용자 테스트용)


## 폰 사진 올리기
- 2026-11-02 분할 → `REF-photo-upload.md`

## 포트 점유 함정 (2026-10-26 MEMORY에서 옮김)
- 잔류 supervisor가 포트(옛 5050, 지금 8012) 점유 시 `bind: address already in use`(마이그레이션은 init서 정상, 바인드만 실패). `ss -ltnp | grep :<포트>`로 pid 찾아 kill
