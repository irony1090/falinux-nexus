# REF — 확인 스크립트 재작성 명세 (2026-10-23 경로 직접 입력부터)

> `REF-test-scripts-1022.md`가 10k자에 닿아 새로 엶. 이력 → `history/test-scripts-1023.md` / 그 전 → `REF-test-scripts-1022.md` / 환경·함정 → `REF-test-env.md` / → `CURRENT.md`

## 2026-10-23 경로-1·2·6(서버: 장비·폴더 후보) (`pi/pi_test.py`, 33개) → `REF-path-input.md` "구현"
- 격리 supervisor 5051 + `nexus_pathin_tmp`(supervisor만 `TMPDIR=<pi>/tmp`) + worker 2개: `pi-a`(이 브랜치) · `pi-old`(**master를 `git worktree add`로 꺼내 빌드 = 옛 worker**, worktree엔 git 제외 `cmd/worker/env`가 없어 복사해야 빌드됨, 빌드 뒤 `worktree remove`). worker는 `-env <절대 경로>`만(TMPDIR을 scratchpad로 바꾸면 보관 소켓 경로가 길어 안 뜸)
- 픽스처 `<pi>/fx/big`: 폴더 `d0000`~`d1199`(1200개) + 파일 300개 + `Docs`·`docs2`·`DOCX`·`.hidden`·`한글폴더`·`alpha1`·`alpha2` + 폴더 링크 `lnk`·파일 링크 `flink`·끊긴 링크 `broken`
- 계정 id는 응답에 없음 -> `psql`로 `users.id`(공유 `PUT /nodes/:id/shares/:userId`에 씀)
- 항목: D1(앞글자 `d11` = 100개, Path는 worker가 이음) · D2(앞글자 없음 max 50 = more, 숨은 폴더·파일 빠짐) · D3(`.` = 숨은 폴더) · D4(`doc` -> `docs2, Docs, DOCX` = 같은 대소문자 먼저) · D5(폴더 링크만) · D6(한글) · D7(max 0 = 정리된 경로) · D8(없음 404·파일 400·상대 400) · D9(빈 dir = `/`) · D12(max 상한) · D10(옛 worker 502) · D11(남 계정·장비 없는 폴더·다른 장비 폴더) / N1(기본 목록 폴더만, 같은 이름 둘 다 목록 순서) · N2(공유받은 루트) · N3(같은 이름 = 첫 번째) · N4(baseId) · N5(첫 번째 아래 없음 404 + 이름) · N6(대소문자 무시) · N7(이름 속 `/`) · N8(공유 폴더 아래 막힌(0) 자식 빠짐·못 감) · N9(남 폴더 404·스크립트 400) · N10(빈 요청 = id null) · N11(상한)

## 2026-10-23 경로-3·4·7(화면: 입력칸·자동완성·폴더 타일) (`pw/pi.mjs`, 31개) → `REF-path-input.md` "구현"
- 위 격리 + vite 3001(`pi/iso.mts` = `iso/iso.mts`에서 cacheDir만 바꿈). 타일 3개를 `PUT /tiles`로: 장비 A(`pi-a`, `fx/big`)·장비 O(`pi-old`)·폴더 F(루트 목록). **타일 크기는 0.5·1만 됨**(`validSpan`, 0.25·0.33은 400). 화면 밖 타일은 `scrollIntoViewIfNeeded` 뒤 좌표 클릭
- 경로 줄 빈 곳 = `.crumbs` 상자에서 `.tools` 왼쪽 20px. 타일 찾기 = `[data-tile-id]`, active 만들기 = `.tools .dev` 클릭(pointerdown, 입력칸은 안 열림)
- 항목: E1(빈 곳 클릭 = 지금 경로·전체 선택) · E2(`d119` 10개, `d11` 50개 + 더 있음) · E3(↓↓ 강조 -> Enter = `d1191/` 채움 -> Enter = 이동·닫힘) · E4(Tab 공통 앞부분 `alpha`·포커스 유지·하나면 끝까지 + `/`) · E5(없는 경로 오류·입력칸 남음·Esc 되돌림) · E6(이름 클릭은 그대로) · E7(Ctrl+L) · E8(옛 worker 후보 없음·Enter는 이동) / F1(`기본`) · F2(같은 이름 2번째 표시) · F3(둘째 proj 골라 sub1 -> 그 아래 `mark-p2`) · F4(루트에서 친 경로 = 첫 번째 proj) · F5(지금 경로 그대로 Enter = 지금 폴더 유지) · F6(`a/b` 골라 `/a\/b/` -> 이동) · F7(없는 폴더 오류·`기본` = 루트)
- **잡은 버그**: 고른 직후 다음 단계 후보가 오기 전에 Enter -> 옛 목록·강조가 남아 같은 후보를 다시 고름(4번 중 2번 실패) -> 고르는 순간 목록·강조 비움(`PathInput.vue` `drop`), 이후 5번 연속 31/31

## 2026-10-23 경로-1·3·4 Windows(VM) (`pi/win_test.py` 14개 · `pw/win.mjs` 12개) → `REF-path-input.md`
- 격리 supervisor `IP=172.17.0.1`(VM -> docker0) + `nexus_pathwin_tmp`, VM worker `dir-win` = `shared/worker-path.exe` + `shared/start-path.ps1`(`Invoke-CimMethod`로 ssh 밖 실행, 픽스처 `C:\nexus\fx\dirs` = `d0000`~`d0599`·`Alpha1`·`alpha2`·`ALPHA3`·`.hid`·`한글`·`junc`(mklink /J)·`afile.txt`). ssh 래퍼 `win/vssh`(`SSH_ASKPASS`, `REF-test-scripts-1014.md` "윈도-6"). 다시 띄울 땐 `taskkill /IM worker-path.exe /F`
- **화면도 172.17.0.1로**: vite `host: 172.17.0.1`(`pi/iso-win.mts`) + `VITE_API_URL=http://172.17.0.1:5051`, 페이지 `http://172.17.0.1:3001`(쿠키 같은 사이트). 타일은 `w:0.5`(1이면 도구 자리가 화면 밖)
- 서버 항목: W1(빈 dir = 드라이브, `c` -> `C:\`, Enter = dir "") · W2(`C:\` + `win` -> Windows) · W3(`C:/nexus/...` 슬래시도) · W4(같은 대소문자 먼저) · W5(파일·`.` 빠짐, junction 포함) · W6(한글) · W7(max 0 정리) · W8(`C:`만 = `C:\`) · W9(없음 404·파일 400) · W10(`Documents and Settings`)
- 화면 항목: K1(역슬래시 경로) · K2(`\d059` 10개) · K3(고르면 `\`까지·이동) · K4(`C:` -> 후보 `C:\`·Tab) · K5(Tab `C:\nexus\`) · K6(`c:/nexus/fx` Enter) · K7(`C:` Enter = 드라이브 루트) · K8(빈 칸 = 드라이브 목록) · K9(파일 = 오류)

## 2026-10-23 글자-5(깨지는 곳 맞추기) 전후 스크린샷 (`ty/pw/*.mjs`) → `REF-typography.md`
- 격리 supervisor 5051(`IP=127.0.0.1`) + `nexus_typo_tmp` + worker `ty-dev` / vite 둘: **이 브랜치 3001 · master 3002**(`git worktree add --detach <ty>/master master` + `node_modules` 심볼릭 링크 + `.env` 복사). iso 설정은 `vite.config.mts`가 함수형이라 `await base(e)` 뒤 `root`·`cacheDir`(둘이 따로)·`server.port/host`를 덮음, 실행 env `VITE_API_URL=http://localhost:5051 HMR_CLIENT_PORT=`. 페이지는 `http://localhost:300x`(쿠키 = localhost)
- 데이터(`seed.py`): 계정 `ty1`·`ty2`(pw123456), 폴더 `생산지그-라인A 검사장비`(ty-dev, ty2에 permission 20 공유 — 본문 필드 이름 `permission`) + 하위 폴더·스크립트 3개(`tick.sh` = 1초마다 한글 줄) + 폴더 하나 더. 타일 트리 `PUT /tiles` = 루트 폴더 + 폴더 타일 + 장비 타일(`instance` = `GET /workers?nodeId=`, `path` = 픽스처 `ty/fx`). 터미널은 화면에서 행 우클릭 아닌 `실행` 클릭(`exec.mjs`)
- 찍는 것: `shoot.mjs`(보드·공유 다이얼로그·폰 390×844) · `zoom.mjs`(DPR 2로 폴더 타일·장비 타일·터미널 헤더·푸터 잘라 위아래로 붙임) · `more.mjs`(행 메뉴·스크립트 편집·장비 지정·전체 보기)
