# REF — 확인 스크립트 재작성 명세 (2026-10-20 이후)

> 2026-10-20 `REF-test-scripts-1014.md`에서 분할(10k자 기준, 2026-10-20 절부터 통째로). 그 전 날짜 → `REF-test-scripts-1014.md` / 환경·격리 실행·함정 → `REF-test-env.md` / 이력 → `history/test-scripts-1020.md` / 현재 진행 → `CURRENT.md`.
> scratchpad 스크립트는 재부팅하면 사라진다 — 여기 적힌 항목·선택자가 다시 쓸 때의 명세.

## 2026-10-20 페이지 표시 → `REF-board-pager.md`
- scratchpad `pager/`: 격리 supervisor(`go build` + 옆 `env` PORT=5051·IP=127.0.0.1·DB_NAME=`nexus_pager_tmp`, `TMPDIR=<scratchpad>/tmp`로 실행) + `vite.iso.mts`(2026-10-19 절과 같음, 3001) + `playwright-core@1.56`
- `pager.mjs`(31개): 계정 + 폴더 f1~f10 + `PUT /tiles`(루트 size `.5x1`, f1~f10 = `[.5,.5] [.5,.5] [1,1] [.5,.5] [.5,.5] [1,.5] [.5,1] [.5,1] [1,.5] [1,.5]`) -> 실제 배치 = Grid 6개 `[루트 f1 f2] [f3] [f4 f5 좁은] [f6] [f7 f8] [f9 f10]`, 칼럼 시작 scrollLeft = 0,636,...,6360(1280x800)
- 판정: `.TileStrip.scrollLeft` vs `.cell` offsetLeft로 만든 칼럼 목록, `.GridPager .count` 글자, `.pill` 폭(정수 > 15 / 반 칸 < 6), `.TileStrip.scrubbing`, `.GridPager__hint.show`, `.TilePeek.open`, `.TileFrame--active` data-tile-id, `document.activeElement`
- 드래그 = `page.mouse` down -> move(steps) — 첫 6px 전엔 포인터를 안 잡음. 스크롤 끝 대기 = scrollLeft가 60ms x 4번 같을 때
- 폰(390x844 isMobile): 인디케이터이 `#app-head-nav` 안·가운데인지만. 터치 드래그는 못 봄

## 2026-10-20 한글 입력 중복 → `REF-term-ime.md`
- scratchpad `ime/`: `index.html` + `main.js`(실제 `xtermIme.util.ts`를 절대 경로 import, xterm 두 개 = 수정 있음/없음, `onData` 수집) + `vite.config.mjs`(vite import 없이 객체만, `resolve.alias` `@xterm/xterm` -> 저장소 `node_modules/@xterm/xterm/lib/xterm.mjs`) -> `apps/frontend`에서 `npx vite build --config <ime>/vite.config.mjs` -> `run.mjs`(node http 정적 서버 5199 + `playwright-core@1.56`, 없는 파일 404 처리 — favicon 요청에 죽음)
- 재생 = 입력창에 `CompositionEvent`·`InputEvent`(inputType·isComposing)를 사용자 로그 순서대로 dispatch, 값은 `ta.value` + `setSelectionRange`로 직접. 묶음 안은 동기, 묶음 사이 20ms(타이머 실행)
- 실브라우저 기록 = 콘솔 스크립트(입력창 이벤트 capture + `WebSocket.prototype.send` 가로채 `PROCESS:INPUT` base64 해독). **앱 프레임은 문자열**(`JSON.stringify`) — 바이너리로만 해독하면 SEND 줄이 안 나옴

## 2026-10-20 헤더·푸터 개편 → `REF-board-chrome.md`
- scratchpad `foot/`: 격리 supervisor(PORT=5051·IP=127.0.0.1·DB_NAME=`nexus_foot_tmp`, `TMPDIR=<scratchpad>/tmp`) + `vite.iso.mts`(2026-10-19 절과 같음, 3001) + `playwright-core@1.56`. 데이터 = 페이지 표시 절과 같은 f1~f10 배치(Grid 6개)
- `foot.mjs`(40개): PC 1280x800 = 헤더 없음 · 푸터 44 맨 아래 · strip 끝 = 푸터 위 · 인디케이터 가운데 · 전체보기 알약 오른쪽 끝 · 버튼이 알약 안 · 스크롤바 0 · 토스트(처음/2.5초/도트 이동/누름/반 칸 `Grid 1-2` = 인디케이터 숫자) · 빈 곳 휠(±반 칸, 작은 delta 8번 = 한 칸, 타일 위 = 그대로) · 위로 당김(안내·`TilePeek`이 막대 위, 다시 내리면 칼럼) · 보드 메뉴 위로 · 알림 드로어 top 0 · 전체보기 알약·돌아가기 / 폰 390x844 = 푸터 56 · 로고만 · 인디케이터 가운데·안 겹침 · 전체 아이콘만 · `...` 위로 펼침 / 로그인 화면 헤더 없음
- 함정: 토스트 `textContent`는 태그·정보가 공백 없이 붙음(`Grid 1column`) -> `\b` 대신 `(?!\d|-)`. Vuetify 메뉴는 열림 애니메이션(scale) 중 재면 위치가 틀림 -> 500ms 기다림

## 2026-10-20 키보드 맞춤 → `REF-board-keyboard.md`
- scratchpad `fit/`: 격리 supervisor(5051·127.0.0.1·`nexus_fit_tmp`) + worker(`NAME=fit-dev`, `SQLITE`=격리 폴더) 둘 다 `TMPDIR=fit/tmp` + `setsid`, `vite.iso.mts`(2026-10-19 절과 같음, 3001) + `playwright-core@1.56`
- `fit.mjs`(28개): 계정·폴더(`deviceKey: fit-dev`)·스크립트(`seq 1 120` + `while read`) -> 페이지에서 `fetch` exec 두 번(`size {w:.5,h:.5}`, 서버가 타일 넣음) -> 새로고침. 폰 390x844 `isMobile`, 열림 = `setViewportSize` 높이 -340. 포커스 = `[data-term-uid] textarea`.focus(). PTY 크기 = `GET /processes`의 rows·cols. 비소유 탭 = 같은 storageState 새 컨텍스트
- 함정: strip을 칸 사이 값(120)으로 옮기면 스냅으로 되돌아감 -> 이동 확인은 Grid 단위(다음 칸 slot left - 8)
