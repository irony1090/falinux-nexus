# REF — 확인 스크립트 재작성 명세 (2026-10-22 이후)

> 2026-10-23 `REF-test-scripts-1020.md`에서 분할(10k자 기준, 2026-10-22 절부터 통째로). 그 전 날짜 → `REF-test-scripts-1020.md` / 환경·격리 실행·함정 → `REF-test-env.md` / 이력 → `history/test-scripts-1022.md` / 현재 진행 → `CURRENT.md`.
> scratchpad 스크립트는 재부팅하면 사라진다 — 여기 적힌 항목·선택자가 다시 쓸 때의 명세.

## 2026-10-22 새 보드·IUTF8 (→ `REF-fix-board-iutf8.md` "확인")
- `iutf8_test.go`(package pty): `ExecInteractive("bash","-c", pre+"stty -a | grep -o -- '-\\?iutf8'; head -n1 | od -An -tx1; echo END")` -> 0.3초 뒤 `Write("가나\x7f\n")` -> `Status()`가 완료될 때까지 -> `OutputAll()`. pre = "" / "stty -iutf8; " 두 번. 실행 = `{"Replace":{"<repo>/apps/core/internal/execute/pty/iutf8_scratch_test.go":"<scratchpad>/iutf8_test.go"}}` overlay로 `go test -overlay ... -run TestIUTF8 -v ./internal/execute/pty/`. bash `read`로 하면 `-iutf8` 쪽이 멈춤(bash가 깨진 바이트 뒤를 기다림) -> `head` 사용
- `newboard_test.py`(Python `requests`, 16개): 계정 만들기·로그인 -> `GET /tiles` 루트 아래 A(A1)·B 폴더 타일(`nodeId: null`)로 PUT -> `POST /workspaces {folderTileId:A}`(from 없음) -> `{from:b1, folderTileId:A1}` -> 거부 4개(루트 400·없는 from 404·옛 version 409·다른 계정 404) -> 목록 3개


## 2026-10-22 process 살려 두기 (`keep/keep_test.py`, 40개) → `REF-process-keep.md`
- 격리: `go build -o keep/sup/supervisor`·`keep/wrk/worker`, env는 `-env <절대 경로>`(scratchpad = dev 판정이라 `-env` 필수·경로는 절대로). supervisor `PORT=5051`·`IP=127.0.0.1`·`DB_NAME=nexus_keep_tmp`, worker `WS_HOST=localhost:5051`·`SQLITE=<keep>/wrk/worker.db`·`NAME=keep-dev`. **TMPDIR은 바꾸지 않음** — 소켓 폴더가 `$TMPDIR` 아래라 scratchpad로 바꾸면 unix 소켓 경로 108바이트를 넘음(소켓 = `/tmp/nexus-hold-<uid>/keep-dev`)
- Python(`requests` + `websockets.sync`, `-I` 없이 — 사용자 site-packages) / sqlite는 Python `sqlite3`(CLI 없음) / DB = `docker exec postgres15 psql -U irony -d nexus_keep_tmp`
- 브라우저 흉내 = `/subscribe` 소켓 DATA를 프론트 `writeLive`와 같은 규칙(lastOff)으로 쌓고 `tick N` 번호가 빈 곳 없이 이어지는지로 판정. 입력 = 소켓 `{"k":2,"t":"PROCESS:INPUT","d":{uid,data(base64)}}`, exec 응답 uid = 최상위 `uid`
- worker는 스크립트가 `subprocess.Popen(start_new_session=True)`로 띄우고 끔(TERM·KILL). supervisor 재시작 = cwd·exe가 `keep/sup`인 프로세스만 끄고 다시 띄움 -> 세션이 메모리라 다시 로그인 + 소켓 다시 연결
- 보관 프로세스 찾기 = `/proc/*/cmdline` 끝 두 개가 `-hold`, `.../<uid>.sock`
- 스크립트: `tick.sh`(0.2초마다 `tick N`) · `exit7.sh`(3초 뒤 exit 7) · `io.sh`(`read` 뒤 `stty size` 반복) · `note.txt`(EDIT, 입력 `ggdGinew<Esc>:wq`)
- 정리: cwd가 `keep/` 아래인 프로세스 kill, `DROP DATABASE nexus_keep_tmp`, 소켓 폴더 비었는지 확인

## 2026-10-22 타일 종류 정리 (`pw/tk.mjs`·`pw/tk2.mjs`, playwright-core 1.64) → `REF-tile-kinds.md` "확인"
- 격리 supervisor 5051·worker `tk-dev` 공유, 프론트 2개: 새 = 저장소 `apps/frontend` / 옛 = `git archive HEAD apps/frontend`를 scratchpad에 풀고 `node_modules` symlink. vite 설정 = scratchpad `iso-{new,old}.mts`(그 트리의 `vite.config.mts`를 **정적 import**해 함수 호출 후 `root`·`cacheDir`·`server.port`만 덮음 — 동적 import는 TS 변환이 안 돼 `UserConfig` export 오류). `VITE_API_URL=http://localhost:5051`은 환경 변수로(.env보다 우선)
- 계정마다 새로 만들고 `POST /nodes`로 폴더 F1(`deviceKey=tk-dev`) + 스크립트 `longrun`(sleep 300)·`quick`(바로 끝)
- tk: 루트 -> F1 새 타일 -> 두 스크립트 실행 -> 터미널 ⋮ 메뉴·단축키 끄기 -> 폴더 ⋮ 메뉴 -> RUN 칩 메뉴 -> 최소화(루트 탭) -> 파일 탭 -> F1 닫기(터미널 adopted) -> 끝난 터미널 × title -> 전체보기 라벨. 단계마다 `.TileFrame`별 {클래스, 머리 줄 글자, 자식별 계산 CSS, 버튼 title/disabled, 루트 탭, ro, revoked}
- tk2: F1 새 보드로 분리(이름 다이얼로그 Enter) -> URL·새 보드 타일 -> 기본 보드에서 F1 다시 열고 다른 보드(B2)로 이동 -> `GET /tiles` 기본 트리 종류
- 함정: 행 메뉴도 title "메뉴" -> 머리 줄은 `.TileFrame__head [title="메뉴"]` / 스크립트 행 = `.row`를 `.nm` 글자로 거름 / 옛 트리 symlink node_modules -> 아이콘 글꼴 403(무시) / 정리는 `nexus-hold-1000/tk-dev/` 소켓의 보관 프로세스만, `/tmp/<nodeId>/<uid>` 실행 사본은 uid로 골라 지움(test3도 같은 폴더 씀)

## 2026-10-23 폴더-6(보드 이동) (`iso/mv_test.py`·`pw/mv.mjs`) → `REF-worker-dir-board.md` "확인"
- 격리 supervisor 5051 + `nexus_mv_tmp` + worker `mv-dev` + vite 3001(`iso/iso.mts` = 저장소 설정을 정적 import해 함수 호출 후 root·cacheDir·server만 덮음). 픽스처 `iso/fx/long.sh`
- 장비 타일은 화면 대신 REST(`PUT /tiles`)로 트리에 넣음: `{type:'device', nodeId, instance, path}` + 터미널 타일의 parent를 바꿈. 쿠키는 playwright `storageState`에 `expires: -1` 등 필드를 다 채워야 함(없으면 `expected float`)

## 2026-10-23 전송-2(업로드 REST)+전송-3(전송 목록) (`iso/up_test.py`, 26개) → `REF-transfer-ext.md` "전송-2·3 구현"
- 격리 supervisor 5051 + `nexus_xferext_tmp`(supervisor만 `TMPDIR=<iso>/tmp` -> 업로드 임시 폴더 `<iso>/tmp/nexus-upload-*`) + worker `xe-dev`(`-env <절대 경로>`, TMPDIR 그대로 — 보관 소켓 경로 108바이트). 받을 폴더 = `<iso>/recv`
- Python `requests` + `websockets.sync.client`(쿠키 헤더로 `/subscribe`, `TRANSFER:*` 프레임만 모음). 계정 -> `POST /nodes` 폴더 F(`deviceKey=xe-dev`)·P(장비 없음) -> instanceKey = `GET /workers`
- 올리기 도우미 = `POST /workers/:m/:s/uploads` -> 512KB씩 `PUT /transfers/:id/data?offset=` -> `POST .../finish` -> `GET /transfers`를 0.1초마다 보며 끝날 때까지
- 항목: U1(3조각+α, sha 일치·.part 없음) · U2(`dir/exists`, `../x`는 빠짐) · U3(같은 이름 = failed/exists, 원본 유지) · U4(overwrite) · U5(offset 건너뜀 409·1MB+1 413·덜 받고 finish 409·같은 offset 멱등) · U6(올리는 중 DELETE -> aborted·임시 파일 삭제·뒤 조각 409) · U7(없는 폴더 no_parent·이름 `x/y` 400·같은 이름 폴더 is_dir) · U8(남 계정 404·목록 빔·장비 없는 폴더 403) · U9(빈 파일) · U10(200MB, send 단계 done>0에서 DELETE -> aborted·worker `.part` 없음) · U11(push 순서 upload/running -> send/running -> send/done, tileId)

## 2026-10-23 전송-4(화면: 올리기)·전송-5(화면: 전송 상태) (`pw/up.mjs`, 18개) → `REF-transfer-ext.md` "전송-4·5 구현"
- 위 절의 격리 supervisor·worker + vite 3001(`iso/iso.mts` = 저장소 `vite.config.mts`를 import해 `base({mode:'development',command:'serve'})` 펼치고 root·cacheDir·server(127.0.0.1:3001, strictPort, hmr 없음)만 덮음, `VITE_API_URL=http://localhost:5051`) + `playwright-core@1.56`(scratchpad `pw/`에 npm i)
- **함정: 페이지는 `http://localhost:3001`로 연다** — `127.0.0.1:3001`로 열면 localhost:5051 쿠키가 교차 사이트라 안 실려 로그인 안 됨(장비 타일이 안 뜸)
- 장비 타일 = REST로 `PUT /tiles`(루트 kids에 `{type:'device', size:{w:0.5,h:1}, nodeId, instance, path:<iso>/recv2}`) — `w:1`로 넣으면 화면보다 넓어 막대 오른쪽이 잘려 보임
- 항목: W0(올리기 버튼 켜짐·전송 없으면 푸터 버튼 없음) · W1(`setInputFiles` 2개 -> 목록에 자동으로 뜸·sha·푸터 목록 2줄 완료·받을 곳·끝난 것은 타일 막대에 없음) · W2(같은 이름 + 새 이름 -> 창 "이미 있는 파일 1개: a.bin" -> 건너뛰기 -> 원본 유지·새 파일만) · W3(모두 덮어쓰기) · W4(같은 이름 폴더만 -> 안내 창) · W5(합성 `DataTransfer`로 dragover·drop, 클래스 확인은 `requestAnimationFrame` 한 번 뒤) · W6(150MB -> 타일 막대 진행·푸터 배지 1 -> 막대의 중단 -> 파일·.part 없음·푸터 목록 중단됨)
