# history — 경로 직접 입력

> → `REF-path-input.md` / → `CURRENT.md`

## 2026-10-23 착수·조율 (코드 없음)
- 사용자 "지금 디렉토리 타일에서 위치가 표시되는 곳, 텍스트로 직접 칠 수 있으면 좋겠어 / 경로별로 칠 때마다 자동완성" -> 파트(경로 직접 입력) 착수
- 조사: 경로 줄 = `DeviceTileBody.vue` `.crumbs`(이름 버튼 나열), 목록 = `DIR_LIST`(worker 캐시, 500개 페이지)
- Q39(입력 들어가는 법)~Q42(자동완성 원천) 답 -> `REF-path-input.md`. 큰 계획 경로-1(후보 요청)~5(확인) 제안
- 사용자 "노드 디렉토리도 지원되는거지?" -> 폴더 타일로 해석해 확인, "일반 디렉토리 타일도 직접 쳐서 이동 겸 자동완성" -> 범위에 폴더 타일 추가. 같은 이름·이름 속 `/`·공유 폴더 문제 짚음 -> Q43(같은 이름 폴더) 첫 번째 · Q44(이름 속 /) 자동완성으로만. 경로-6(폴더 후보 REST)·7(화면: 폴더 타일 입력) 추가

## 2026-10-23 구현·확인 (경로-1~7)
- 사용자 "진행" -> 브랜치 `feat/path-input`(master `b17f77e`) -> 경로-1(후보 요청)·2(후보 REST)·6(폴더 후보 REST) 서버, 경로-3(화면: 입력칸)·4(화면: 자동완성)·7(화면: 폴더 타일 입력) -> 상세 `REF-path-input.md` "구현"
- 결정(Claude): 폴더 후보·Enter를 한 REST(`POST /nodes/suggest`, max 0 = 찾기만)로 / 숨은 폴더는 `.`을 칠 때만 / 단축키 Ctrl+L은 active 타일만 / 입력칸을 열 때 지금 경로 단계를 기억(지금 경로 그대로 Enter = 같은 이름이어도 지금 폴더)
- 확인: 서버 33/33(`pi_test.py`), 헤드리스 31/31 5번 연속(`pi.mjs`) → `REF-test-scripts-1023.md`. 옛 worker는 master를 worktree로 꺼내 빌드해 확인(502 -> 화면은 후보 없이 이동)
- 잡은 버그: 고른 직후 빠른 Enter가 옛 후보를 다시 고름 -> 고르는 순간 목록·강조 비움. 후보 줄이 기본 버튼 모양 -> 스타일 초기화
- 격리 프로세스(cwd로 고름)·`nexus_pathin_tmp`·픽스처 정리함

## 2026-10-23 Windows VM 확인
- 사용자 "윈도우 vm 켜줄래? 너도 한번 확인해보고, 나도 확인해보기위해 worker까지 최신화" -> `docker start nexus-win`
- 사용자가 이미 `build/directory_path`(태그 `DIRECTORYPATH-261023`, 15:21 = 마지막 화면 수정 15:17 뒤)로 test3를 다시 띄워 둠 -> test3 재기동 불필요
- 격리 supervisor `IP=172.17.0.1`·`nexus_pathwin_tmp` + VM worker `dir-win`(`shared/worker-path.exe`·`start-path.ps1`, 픽스처 `C:\nexus\fx\dirs` 600개 등) -> 서버 `pi/win_test.py` 14/14, 화면 `pw/win.mjs` 12/12(vite도 172.17.0.1:3001)
- 잡은 버그: `C:`만 치고 Enter = "절대 경로가 아닙니다" -> worker가 `C:\`로 받음
- 정리 뒤 VM worker `win-dev`를 test3에 최신 빌드(`shared/worker-test3.exe`, 태그 `DIRECTORYPATH-261023` + `C:` 고침)로 다시 붙임. VM은 사용자 확인 동안 켜 둠

## 2026-10-23 마감
- 사용자 "window 확인했어. VM 종료해도돼. 경로 기능 확인 끝" -> `docker stop nexus-win` -> 파트(경로 직접 입력) 마감, `INDEX-path-input.md`로 접음
- 커밋: `feat/path-input`에 파트 마감 커밋 -> master `--no-ff` 머지. push 안 함
- 마감 뒤 사용자 모바일 확인 "자동완성을 눌러도 키보드 안내려가서 좋더라" -> 반복 주제 `REF-mobile-viewport.md` 5-1절에 기록
