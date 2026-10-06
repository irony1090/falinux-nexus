# REF — 테스트 환경 · 확인용 클라이언트 함정

> 2026-10-06 CURRENT에서 분리(세션마다 읽을 필요 없는 재사용 지식). 이력 → `history/test-env.md` / 현재 진행 → `CURRENT.md`. 각 기능의 확인 항목은 해당 REF의 확인 절(예: `REF-process-snapshot-impl.md` "스냅샷-b/-d", `REF-node-ui-save-impl.md`, `REF-node-ui-sync.md`, `REF-process-input.md` "입력-a/입력-c").

## 환경
- 환경: supervisor·worker는 사용자가 띄움(`docker start postgres15` → supervisor → worker, dev 서버 3000). 테스트 로그인 = `pages/Login.vue` 기본값 계정. 헤드리스 = `~/.cache/ms-playwright/chromium-1234/chrome-linux64/chrome` + `playwright-core`(scratchpad에 `npm i`). worker 연결만 끊는 재현 = `REF-process-reconnect.md` "재현 방법"
- 스크립트는 직접 실행이라 `#!/bin/sh` 등 shebang 필수(없으면 `exec format error`)
- 재부팅으로 사라지는 것: 세션 scratchpad의 playwright-core + 테스트 스크립트(4.5 `snap/server.mjs`·`browser.mjs`·`probe.mjs`, 이전 `tiles/*.mjs`·`input/*.mjs`). 다시 쓸 땐 `REF-process-snapshot-impl.md` "스냅샷-b/-d"·`REF-node-ui-save-impl.md`·`REF-node-ui-sync.md` 각 확인 절(항목·함정), `REF-process-input.md` "입력-a/입력-c" 보고 재작성. 테스트는 계정을 새로 만들고 폴더 `deviceKey=irony-MAC-ADDress1`로 하면 irony 타일 트리를 안 건드림

## 테스트 클라이언트 함정
- 테스트 클라이언트 함정 모음: 소켓 프레임 binary / exec는 `X-Tab-Id` 필수 / STATUS `status` 숫자 / 브라우저 쿠키 도메인 `localhost` / `users.identification`은 해시 저장 / 화면 밖 Grid xterm은 `innerText`에 안 잡힘(`.xterm-rows` `textContent`) / API로 만든 노드는 열린 목록에 안 뜸(`[title="새로고침"]`) / Vite가 heredoc으로 막 만든 파일을 빈 내용으로 캐시할 수 있음(`touch`) / 세션 쿠키 값에 `=` 있을 수 있음(첫 `=`에서 자를 것) / exec 직후 크기 0x0이라 resize 먼저(소유 탭) / vim 테스트는 파일 이름을 매번 다르게(swap 파일)

## DB 상태 (2026-10-02 기준)
- DB에 남은 테스트 데이터: 계정 `input-test-b`(타일 트리 행 있음, 루트만) · `tile-race-<시각>` 1개 · `tile-store-`/`tile-exec-`/`tile-restore-`/`tile-other-`/`dbg-`/`push-`/`push-other-`/`sync-`/`sync2-<시각>` 여러 개(3-c·3-d·4 확인용) + `snap-`/`snap-other-`/`snap-probe-`/`snapb-<이름>-<시각>`(4.5 확인용, 노드는 삭제). irony 타일 트리 = 루트만(version 12)
- DB 노드 현황: 폴더 id 1(`HTOP_TEST_SH_MODI`, `device_key=irony-MAC-ADDress1`) 안에 스크립트 2~4 / 폴더 `test`(14) > `ttt1`(15) > 스크립트 23(사용자 테스트용)
