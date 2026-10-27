# history — CURRENT에서 내려보낸 줄 (원문 보존)

→ `CURRENT.md` · 각 줄의 상세는 그 줄이 가리키는 INDEX/REF
> 2026-10-26 낮까지의 줄 → `history/current-archive-1.md` / 밤 세션 끝 → `history/current-archive-2.md`(분할)

## 2026-10-26 밤 — fix/phone-layout 머지 때 내려보낸 줄(원문)
  1. **`fix/phone-layout` 브랜치(작업 폴더 = 원래 폴더, 코드 커밋 전)** 사용자 화면 확인 -> 커밋 + master 머지 -> 이 PC supervisor·웹 재빌드. 내용: ① 폰 관리 페이지 첫 타일로 안 넘어가던 버그(`memberUrl.hook.ts`, 원인·수정 → `history/member-ui.md` 10-26 밤) ② 보드 전체보기 폰도 가로 한 줄 + 보드 여럿이면 이름을 줄 위(EV2 개정 → `REF-board-preview.md`). 격리 `phonefix.mjs` 7/7 + 보드 2개 이름 위치 통과
- **새 버그 2개(10-26 밤, 사용자 폰 확인 중 발견) — 수정·`build/api_prefix` 웹 교체 완료, 사용자 확인 대기(`fix/phone-layout`에 같이 커밋)** — 사진 `obsidian-vault/user/Screenshot_20261026_193456_Edge.jpg`(푸터)·`..._193545_Edge.jpg`(자동완성)
  - 버그(빈 페이지 도트): 폰 내 장비 관리에서 마지막 타일을 닫으면 빈 안내 페이지(`MemberPage` `.empty`)는 남는데 `dotCount = lead + grids.length` = 1 -> 지금 위치 표시(pos 1)가 도트 줄 폭 밖에 그려져 잘림. 수정 = `lead + max(1, grids.length)`(`memberBoard.store.ts`)
  - 버그(자동완성 비침): 장비 타일 경로 줄 `.crumbs`가 sticky + `z-index: 1` -> 쌓임 맥락이 1이라 목록 칸 `⋮`(`.dg-btn` 2)·`링크` 라벨(`.dg-r/.dg-l` 1)이 자동완성 위로. 수정 = `.crumbs` z-index 3(`DeviceTileBody.vue`). 폴더 타일은 `.crumbs`가 sticky 아니라 해당 없음
  - 추가(10-26 밤, 사용자 요청 가(화면 이동만)): 상세 타일을 다 닫으면 목록으로 — `memberBoard.store` `close`에서 0개면 `goCol(0)`(폰 = 첫 칸, PC 변화 없음). 관리자 페이지도 같은 store라 같이 적용(도트 수정 포함), 자동완성 비침은 보드 장비 타일 전용이라 관리 쪽 해당 없음. 웹 3번째 교체(직전 웹 = scratchpad `web-prev-2`), 사용자 확인 대기
- **git**: master = `a56b98f`(`refactor/api-prefix` 머지 `c5f9749` 포함, push = 사용자). 로컬 브랜치 `refactor/api-prefix`(머지됨) · `fix/phone-layout`(진행 중). worktree는 원래 폴더 하나(`nexus-member` 삭제, 원격 `origin/feat/member` 정리는 사용자)
- **지금 떠 있는 것**(2026-10-26 밤): test3(8013) = supervisor **`build/api_prefix`**(사용자 빌드, **웹만 10-26 밤 `fix/phone-layout` 작업 트리로 교체(2번째 = 새 버그 2개 포함)** — supervisor 재시작 없음, master 웹은 세션 scratchpad `web-old-1018`) + worker **`build/take_process/worker`**(이번 변경에 worker 없음, 재시작 안 함) -> **`build/take_process/` 지우지 말 것**. 보관 프로세스 771160 = 실행 파일(`build/supervisor_move/worker`)은 지워졌지만 계속 돎(12:12 터미널, take_process worker가 붙어 있음). 격리 5051·VM `nexus-win` 꺼짐. 남긴 DB = `nexus_api_tmp`·`nexus_member_tmp`(시험 계정, 필요 없으면 DROP)

## 2026-10-26 22:19 — 집 서버 배포 끝나 내려보낸 줄(원문)
  2. **집 서버 배포 결과 듣기**: 사용자가 집에서 `build/home-update/nexus-api-prefix.zip`(10-26 밤 다시 묶음: supervisor.exe = `c5f9749` 그대로 + 웹 = master `8de3e05`, **phone-layout 수정 포함**) + `update.ps1`로 배포 예정 → `REF-home-server.md` 끝. 이미 옛 zip으로 배포했어도 새 zip으로 같은 방법 다시 실행하면 됨(exe 같음, 웹만 바뀜)

## 2026-10-26 밤 vault 색인 정리 — 내려보낸 줄(원문)
- **(2026-10-26 밤 세션 끝, 사용자 "vault 정리하고 세션 재시작") 다음 세션 첫 할 일**
  1. ~~`fix/phone-layout`~~ **DONE 10-26 밤** — 사용자 폰 확인 -> 커밋 `77c9901`·`dbc136a` -> master `--no-ff` 머지 `8de3e05`(첫 타일 이동·전체보기 가로 한 줄·빈 페이지 도트·다 닫으면 목록·자동완성 비침 → `history/member-ui.md`·`history/board-preview.md`·`history/path-input.md`). test3 웹은 머지 내용과 같은 작업 트리로 이미 교체됨
  2. ~~집 서버 배포~~ **DONE 10-26 22:19** — 새 `nexus-api-prefix.zip`(웹 `8de3e05`) `update.ps1` OK(8013) → `history/home-server.md`. 남은 것: 사용자 화면 확인 뒤 서버 PC `*.old-20261026-221903`·`C:\Nexus\api_prefix` 삭제(사용자)
  3. **파트(회원 관리) 마감 정리**: MEMORY 불변 결정(판 계약·회원 권한 등) + `INDEX-member` 접기 — MEMORY가 10.3k자로 기준 초과, 이때 줄임. 사용자에게 파트 마감 확인
- **git**: master = `8de3e05`(`fix/phone-layout` 머지, push = 사용자). 로컬 브랜치 `refactor/api-prefix`·`fix/phone-layout`(둘 다 머지됨). worktree는 원래 폴더 하나(`nexus-member` 삭제, 원격 `origin/feat/member` 정리는 사용자)
- **지금 떠 있는 것**(2026-10-26 밤): test3(8013) = supervisor **`build/api_prefix`**(사용자 빌드, **웹만 10-26 밤 `fix/phone-layout` 내용(= master `8de3e05`)으로 교체** — supervisor 재시작 없음) + worker **`build/take_process/worker`**(이번 변경에 worker 없음, 재시작 안 함) -> **`build/take_process/` 지우지 말 것**. 보관 프로세스 771160 = 실행 파일(`build/supervisor_move/worker`)은 지워졌지만 계속 돎(12:12 터미널, take_process worker가 붙어 있음). 격리 5051·VM `nexus-win` 꺼짐. 남긴 DB = `nexus_api_tmp`·`nexus_member_tmp`(시험 계정, 필요 없으면 DROP)
- **`build/`**(2026-10-26 정리): `api_prefix/`·`take_process/`·`supervisor.env`·`worker.env`·`worker-sqlite.db*`·`home-update/`(ps1 스크립트 + `api_prefix/` 묶음·`nexus-api-prefix.zip`)
- **새 BACKLOG(10-26)**: Windows 터미널 출력 줄 빠짐(ConPTY 의심, 사용자가 시험 넘김) · 주소 목록 3개 상한이 지금 주소를 밀어냄
- 오늘 이력: 회원 관리 머지·집 서버 Nexus 터미널 배포 → `history/home-server.md` / `/api` 정리·주소 일괄 변경 → `history/api-prefix.md` / 폰 수정 → `history/member-ui.md`·`history/board-preview.md`. 옛 CURRENT 줄 원문 → `history/current-archive.md` "2026-10-26 밤"
- **최근 마감 파트**(상세는 각 INDEX, 목록 = `HISTORY.md` "마감 파트"): supervisor 주소 실시간 변경 `INDEX-supervisor-move`(10-26) · 빈 곳 길게 누르기 `INDEX-press-pager`(10-25) · 목록 아이콘 보기 `INDEX-dir-grid`(10-24) · 타일 머리 줄 버튼 `INDEX-tile-head`(10-24) · Ctrl+Alt 팝업 버튼·터치 스크롤 양 `INDEX-pager-tools`(10-23) · 글자 크기 규칙 `INDEX-typography`(10-23) · 경로 직접 입력 `INDEX-path-input`(10-23) · 내려받기·worker 간 전송 `INDEX-transfer-down`(10-23) · 파일 전송 확장 `INDEX-transfer-ext`(10-23) · worker 폴더 타일 `INDEX-worker-dir`(10-23) · 타일 종류 정리 `INDEX-tile-kinds`(10-23) · process 살려 두기 `INDEX-process-keep`(10-22) · 새 보드·IUTF8 `INDEX-fix-board-iutf8`(10-22) · 빌드본 분리 `INDEX-build`(10-22) · 리팩터링 `INDEX-refactor`(10-21) · 페이지 이동 `INDEX-tile-move`(10-21) · 한글 입력 중복 `INDEX-term-ime` · 파일 전송 `INDEX-transfer-e2e` · 헤더·푸터 개편 `INDEX-board-chrome` · 페이지 표시 개편 `INDEX-board-pager`(10-20) · URL 바인딩 `INDEX-tile-url` · 터미널 터치 스크롤 `INDEX-term-touch` · Windows worker `INDEX-windows-worker` · PgUp·PgDn 버튼 `INDEX-term-page-keys` · 고아 스크립트 `INDEX-worker-orphan` · 보드 미리보기 `INDEX-board-preview`(10-15~19)

## 2026-10-27 오후 (파트(장비 상태 push) 마감 때 내려보냄)
- **(2026-10-27 오후) 진행 중 파트 = 장비 상태 push** → `REF-device-push.md` / 브랜치 `feat/device-push`(master `4cf6855`에서, **미커밋**)
  - 장비 접속·끊김·라벨·등록/삭제·주인 변경 → `DEVICE:CHANGE` 신호 → 보드·내 장비·관리자 화면이 다시 조회. 폴링 60초 안전망. 소켓 connect를 `App.vue`로 옮김(사용자 승인)
  - 격리 확인 끝(받는 사람 10/10·보드 0.4초·관리 페이지 1.5초 안), 격리 정리 끝. **다음: 사용자 확인 → 마감·커밋·머지**
  - (14:08) 사용자 확인용 격리 다시 띄움: supervisor 5051 + DB `nexus_dp_tmp` + vite 3001 + worker `dp-dev`(scratchpad `dp/on.sh`·`off.sh`). 계정 a1 관리자·o2 주인·v3 공유받음·x4 무관(pw123456). 끝나면 정리

## 2026-10-27 오후 (파트(칸 메뉴 장비 디렉토리) 마감 때 내려보냄)
- **(2026-10-27 오후) 진행 중 파트 없음** — 파트(장비 상태 push) 마감 → `INDEX-device-push.md`. master = `feat/device-push` 머지, push = 사용자. 격리 전부 정리
- **(2026-10-27 오후) 진행 중 파트 = 칸 메뉴 장비 디렉토리** → `REF-row-open-dir.md` / 브랜치 `feat/row-open-dir`(master `e16a519`에서, **미커밋**)
  - 폴더 칸 ⋮·우클릭 메뉴에 "장비 디렉토리 열기"(내 폴더 + 장비(상속 포함)일 때만, 오프라인 = 꺼짐). 격리 헤드리스 확인 OK
  - 격리 떠 있음: supervisor 5051 + DB `nexus_dp_tmp` + vite 3001 + worker `dp-dev`(scratchpad `dp/on.sh`·`off.sh`), 계정 a1·o2·v3·x4(pw123456). **다음: 사용자 확인 → 격리 정리 → 마감·커밋·머지**
  - test3(8013) = 젤리·셸·장비 push 미반영(supervisor·worker·웹 다시 빌드)

## 2026-10-27 저녁 (세션 끝 vault 정리 때 내려보냄)
- **(2026-10-27 오후) 진행 중 파트 없음** — 오늘 마감 = 파트(장비 상태 push) `INDEX-device-push.md` → 파트(칸 메뉴 장비 디렉토리) `INDEX-row-open-dir.md`. 다음 파트는 `BACKLOG.md`에서 사용자와 고른다
  - git: master = `feat/row-open-dir` 머지(그 전 `feat/device-push`), push = 사용자. 격리 전부 정리(5051·3001 닫힘, `nexus_dp_tmp` DROP)
  - test3(8013) = 젤리·셸·장비 push·칸 메뉴 미반영(supervisor·worker·웹 다시 빌드)
- **(2026-10-27 낮) 진행 중 파트 없음** — 파트(폴더에서 터미널) 마감 → `INDEX-dir-shell.md`(그 전 파트(젤리·관리 기록) → `INDEX-jelly.md`). 다음 파트는 `BACKLOG.md`에서 사용자와 고른다
  - git: master = `feat/dir-shell` 머지(10-27), push = 사용자. 개발 서버 8011·8012 꺼짐, VM `nexus-win` 꺼짐, 격리 전부 정리. test3(8013) = 젤리·셸 둘 다 미반영(worker도 새 빌드여야 셸이 됨)
  - 개발 DB `nexus`에 장비 등록 없음(회원 관리 이후) → 개발 환경에서 장비 타일을 쓰려면 아이로니 계정 "내 장비"에서 `irony-MAC-ADDress1`·`win-dev` 등록 + worker env `API_KEY`

## 2026-10-27 파트(찰떡 마무리) 마감 때 내려보낸 줄
- **(2026-10-27 저녁 세션 끝) 진행 중 파트 = 찰떡 마무리** → `REF-ui-chaltteok.md`(항목 표 1~5 + "찰떡 입력칸 구상") / 브랜치 `feat/ui-chaltteok`(master `87ff6f2`에서, **미커밋**)
  - 구현·헤드리스 확인 끝: 1(경로 줄 여백) · 2(관리자 필터 = 시안 A′) · 3(남은 창 5개) · 4(상세 상태 칸 = 안 2, 구분선 없음 + 여백 조정 2건). 사용자가 확인하던 중
  - 5(select 목록 흰 바탕 흰 글씨) = 미해결 → 사용자 제안 "Vuetify 입력칸을 찰떡 공용 컴포넌트로" 논의만. autocomplete = 주인 고르기·공유 대상 찾기
  - **다음 세션 첫 질문**: 지금 파트를 5 빼고 마감(커밋·머지)하고 새 파트(가칭 "찰떡 입력칸")로 갈지(Claude 추천) / 같은 파트에 넣을지 → 정해지면 진행
  - **격리 떠 있음**(세션 재시작해도 남음): supervisor 5051 + DB `nexus_dp_tmp` + vite 3001 + worker `dp-dev`. 켜고 끄기 `/tmp/claude-1000/-home-jh-bae-irony-nexus/483bf060-2b4f-44c6-86c0-6dd9c05f93c4/scratchpad/dp/on.sh`·`off.sh`(끄기 = exe 경로로 고름), 계정 a1 관리자·o2 주인·v3 공유받음·x4 무관(pw123456). 정리 = supervisor·worker·vite를 exe/cmdline으로 골라 kill(`pgrep -f` 문자열은 내 bash도 잡으니 node exe 확인) + `DROP DATABASE nexus_dp_tmp` + `/tmp/nexus-hold-1000/dp-dev`
  - 시안 아티팩트 4개 링크 = REF 항목 표
