# history — CURRENT에서 내려보낸 줄 (원문 보존)

→ `CURRENT.md` · 각 줄의 상세는 그 줄이 가리키는 INDEX/REF

## 2026-10-26 재부팅 전 정리 때 내림 (이미 마감·INDEX로 접힌 파트들)
- **2026-10-25 파트(찰떡 UI) 마감·master 머지**: 찰떡-1(푸터)·찰떡-2(공유 창)·찰떡-3(내가 공유한 것 창)·찰떡-4(장비 지정 창)·찰떡-5(터미널 탭 목록·칸 실행 팝업) + 공통 `styles/chaltteok.scss` → `REF-design-chaltteok.md` "Nexus에서 쓰는 곳", 경위 `history/design-chaltteok.md`. 사용자 test1 확인 뒤 test3 `build/ctrl_panel/web`만 이 빌드로 교체(백업 안 둠), test1 닫음. 확인 중 실수로 test3 irony 보드에 생긴 터미널·장비 타일은 사용자가 정리함. 브랜치 `feat/chaltteok-ui` 남겨 둠, push 안 함
- **2026-10-25 디자인 규칙 "찰떡 규칙" 정리 끝** → `REF-design-chaltteok.md`(규칙 본문, 다른 작업공간도 직접 참조, 전역 CLAUDE.md엔 안 넣음) · 아티팩트 https://claude.ai/artifact/JzUCKbGWTm4TL8A3GpqeRo · user vault `user/links.md` "디자인 규칙" 절. 코드 변경 없음
- **2026-10-25 터치 확대 막기(master 바로)** → `REF-mobile-viewport.md` 7절: viewport·전역 touch-action·iOS gesturestart. 크롬 헤드리스 확인, test3 `web` 교체(옛 것 = `web.bak` = 파트(빈 곳 길게 누르기), 그 전 = `web.bak.press-pager-전`) — 사용자 확인 OK(2026-10-25, 아이폰은 미확인 → `BACKLOG.md`), test1 닫음
- **2026-10-25 파트(빈 곳 길게 누르기) 마감·master 머지**(→ `INDEX-press-pager.md`): 경로 줄·도구 줄·목록 빈 곳 터치 길게 = 타일 컨트롤 패널 + 고른 칸의 누를 수 있는 라벨 테두리 부드럽게 강조. 사용자 test1 확인 OK(폴더·장비 타일). 마감 뒤 루트 탭 줄 빈 곳도 대상으로 고침(REF "후속"). test3 `build/ctrl_panel/web`을 이 빌드로 교체(옛 화면 = `web.bak` = 파트(목록 아이콘 보기), 그 전 것 = `web.bak.dir-grid-전`, 더 전 = `web.bak.tile-head-전`) — 사용자 test3 확인 OK(2026-10-25, 탭 줄 후속 포함), test1 닫음. 브랜치 `feat/press-pager` 남겨 둠, push 안 함
- **2026-10-24 파트(목록 아이콘 보기) 마감·master 머지**(→ `INDEX-dir-grid.md`): 폴더·장비 타일 목록을 격자 아이콘 칸으로 교체(판 D′) + 누르기 규칙(한 번 = 고름, 더블 = 열기·실행, Ctrl·길게 = 체크) + 칸에서 공유·장비 지정. 사용자 test1 확인 OK. test3 `build/ctrl_panel/web`을 이 빌드로 교체(옛 화면 = `web.bak`, 그 전 것 = `web.bak.tile-head-전`) — 사용자 test3 확인 OK(2026-10-25), test1 닫음. 브랜치 `feat/dir-grid` 남겨 둠, push 안 함
- **2026-10-24 파트(타일 머리 줄 버튼) 마감·master 머지**(→ `INDEX-tile-head.md`): 머리 줄·경로 줄 버튼을 알약·끝 블록으로 통일 + 경로 줄 좁으면 두 줄. 사용자 test1 확인 OK. test3 `build/ctrl_panel/web`만 새 빌드로 교체(옛 것 = `web.bak`), 사용자 확인 OK. 브랜치 `feat/tile-head` 남겨 둠, push 안 함. **vault는 `feat/supervisor-move` 것을 옮겨 와 같이 커밋됨**(supervisor-move REF·history 포함) — 그 브랜치를 master에 머지할 때 vault 충돌(CURRENT·HISTORY·MEMORY)은 master 쪽을 기준으로 정리
- **2026-10-23 파트 2개 마감·master 머지**: 파트(타일 종류 정리) → `INDEX-tile-kinds.md` / 파트(worker 폴더 타일) → `INDEX-worker-dir.md`. 브랜치 `refactor/tile-kinds`·`feat/worker-dir` 남겨 둠, push 안 함
- **2026-10-23 파트(파일 전송 확장)·파트(내려받기·worker 간 전송) 마감·master 머지**(→ `INDEX-transfer-ext.md`·`INDEX-transfer-down.md`): 장비 타일 업로드·전송 상태 + 내려받기(파일·zip)·worker 간 복사(메뉴·끌어 놓기). 사용자 브라우저 확인 OK. 브랜치 `feat/transfer-ext`·`feat/transfer-down`·`feat/path-input` 남겨 둠, push 안 함
- **2026-10-23 파트(Ctrl+Alt 팝업 버튼·터치 스크롤 양) 마감·master 머지**(→ `INDEX-pager-tools.md`): Ctrl+Alt 팝업에 크기·전체보기 버튼 + 터치 길게 누르기 + 터미널 터치 스크롤 2배. 사용자가 이 브랜치로 빌드해 PC·폰 확인 OK(빌드 이름 미확인). 브랜치 `feat/pager-tools` 남겨 둠, push 안 함. 격리 환경(5051·3001·`nexus_pager_tmp`) 정리함. **다음 파트 = 미정**(후보 `BACKLOG.md`)
- **2026-10-23 파트(글자 크기 규칙) 마감·master 머지**(→ `INDEX-typography.md`): 단계 이름 mixin + 단계 값 한 곳 + 배율 손잡이, 터미널 14px. 브랜치 `feat/typography` 남겨 둠
- **2026-10-23 파트(경로 직접 입력) 마감·master 머지**(→ `INDEX-path-input.md`): 장비 타일·폴더 타일 경로 직접 입력 + 단계별 자동완성. 사용자 브라우저(Linux·Windows) 확인 OK. 브랜치 `feat/path-input` 남겨 둠, push 안 함
- **파트(process 살려 두기) 2026-10-22 마감·커밋**(→ `INDEX-process-keep.md`): 보관 프로세스(`worker -hold`)로 worker·supervisor 재시작에도 터미널·화면 유지. 미확인 K7(끝난 뒤 기다림) 10분 경과·K6(Windows)
- **파트(새 보드·IUTF8) 2026-10-22 마감·커밋**(→ `INDEX-fix-board-iutf8.md`): 어느 보드에서든 새 보드로 분리 + worker PTY `IUTF8`. 빌드본 브라우저 확인 OK
- 그 전 파트(리팩터링) 2026-10-21 마감(→ `INDEX-refactor.md`): 브랜치 `refactor/socket`의 커밋 3개(소켓 1차 `8cf32bc`·소켓 2차 `0dc83a7`·소켓 밖 정리 `38a24a7`)를 master에 `--no-ff` 머지, 브랜치는 남겨 둠
  - 리팩터링 뒤 옛 문서의 `tile.go` 함수 위치가 바뀜(→ `REF-refactor-cleanup.md` "결과"), 키보드 맞춤-1(타일 띄우기)은 `hook/tileFloat.hook.ts`
  - 확인 스크립트(scratchpad)는 세션 끝나면 사라짐 → 명세 `REF-test-scripts-1020.md` "2026-10-21 소켓 알림 리팩터"
- 이전 세션(2026-10-12~15)의 scratchpad 스크립트는 이미 사라짐 — 명세는 `REF-test-scripts.md`(~10-13)·`REF-test-scripts-1014.md`(10-14~)·`REF-test-scripts-1020.md`(10-20~)·`REF-test-scripts-1022.md`(10-22~) 날짜별 절. 그때 격리 프로세스·임시 DB는 전부 정리됨. PC 재부팅 뒤라면 `docker start postgres15` -> supervisor -> worker -> `npm run dev`(→ MEMORY "로컬 개발 환경")
