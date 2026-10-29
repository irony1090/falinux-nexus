# CURRENT

## 현재 날짜
2026-10-29

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **(2026-10-29) 파트(타일 끌기) 마감** — `feat/tile-drag` 커밋 + master 머지, push = 사용자. **진행 중 파트 없음** → `REF-tile-drag.md`(계획·결정) · `REF-tile-drag-impl.md`(구현·함정) · `REF-tile-order.md`(순서 구조) · 결정 = MEMORY "터치 제스처"·"끌기"(상세 `REF-decisions-ui.md`)
  - 구조-1~3 · 끌기-1·2 · 제스처-1 · 보드-1 · 칸-1·2 · 끝 띠 + 선택 막대 배치 고침 · 로그인 제목 `NEXUS`
  - **빌드 = 사용자가 직접**: 태그 `# SUPERVISOR-TILE_DRAG-261029`·`# WORKER-TILE_DRAG-261029`(env 첫 줄, 아직 `VIDEO_VIEW-261029`) -> `scripts/build.sh 261029_tile_drag`. **DB 마이그레이션 없음**, 옛 타일 문서(kids)는 supervisor가 처음 읽을 때 order로 변환 저장(`tileOrderMigrate.go`, 실데이터 처음 = test3 `nexus_build`) — 배포 뒤 보드가 그대로인지 확인할 것. 프론트·supervisor 같이, 열린 탭 새로고침
  - **아직 떠 있는 것(정리 여부 사용자 미정)**:
    | 포트 | 무엇 | 끄기 |
    |---|---|---|
    | 8011(test1) | vite(저장소 `apps/frontend`) | `kill $(cat /tmp/claude-1000/-home-jh-bae-irony-nexus/51cf026c-d38c-4465-8fee-75e8924cd07d/scratchpad/ext/vite.pid)` |
    | 8012(test2) | `socat` -> 5051. **떠 있는 동안 개발 supervisor 8012 못 띄움** | `kill $(cat /tmp/claude-1000/-home-jh-bae-irony-nexus/51cf026c-d38c-4465-8fee-75e8924cd07d/scratchpad/ext/socat.pid)` |
    | 5051 | 격리 supervisor(칸-1 코드, `/tmp/claude-1000/-home-jh-bae-irony-nexus/06a4a87f-5384-4e2f-870b-848ca2e42a62/scratchpad/iso/sup`, DB `nexus_td_tmp`) + worker `td-dev` | `/tmp/claude-1000/-home-jh-bae-irony-nexus/06a4a87f-5384-4e2f-870b-848ca2e42a62/scratchpad/iso/stop.sh /tmp/claude-1000/-home-jh-bae-irony-nexus/06a4a87f-5384-4e2f-870b-848ca2e42a62/scratchpad/iso` |
    | 3001 | 격리 vite(`--config /tmp/claude-1000/-home-jh-bae-irony-nexus/06a4a87f-5384-4e2f-870b-848ca2e42a62/scratchpad/iso/vite.iso.mts`) | 그 pid kill |
    - 계정 td1·td2(`pw123456`). td1 보드 = 루트 + X(A) + C(칸시험) — 되돌리기 `python3 /tmp/claude-1000/-home-jh-bae-irony-nexus/1dbb5e34-df83-47a1-b7aa-5d313a551b3b/scratchpad/cell/board.py`, 옛 기본 보드 = `python3 /tmp/claude-1000/-home-jh-bae-irony-nexus/51cf026c-d38c-4465-8fee-75e8924cd07d/scratchpad/ui/setup.py`
    - 다 쓰면: 위 프로세스 끄기 + `DROP DATABASE nexus_td_tmp` + `/run/user/1000/tdiso` 지우기
- **그다음 파트 후보** = `BACKLOG.md`(사용자 예정 "타일 ESC로 닫기"·회원 관리 판 끌기)
- **(2026-10-29) 파트(영상 Range) 마감** — `feat/video-view` 커밋 + master 머지, push = 사용자. **진행 중 파트 없음** → `REF-video-view.md`
  - 1단계(영상 재생: Range·서버가 넣는 영상 타일·재생 불가·오프라인 이어 재생) + 2단계(S1(이어보기), 마이그레이션 `00015`)
  - **정식 빌드 = 사용자가 직접**: `apps/core/cmd/*/env` 첫 줄 태그가 지금 `VIDEO_VIEW_TMP-261029`(임시) -> 바꾸고 `scripts/build.sh <이름>` -> test3 교체 뒤 `build/video_view_tmp`(폴더에 `TEMP-BUILD.txt`) 지워도 됨
  - test3(8013) 지금 = **`build/video_view_tmp`**(supervisor·worker, 사용자 실행 10-29, `nexus_build` = goose 15). 사용자 체감 "생각보다 느리지 않음"
  - 집 서버 = 영상 파트 아직 안 보냄 → `BACKLOG.md` "집 서버 영상 배포"
  - 남긴 것: DB `nexus_vv_tmp`(격리 확인용, 계정 `vv1`, 필요 없으면 DROP) · 이 세션 scratchpad `iso/`(픽스처·스크립트, 재부팅 때 사라짐). 격리 프로세스·`/run/user/1000/vviso` 정리함
- **(2026-10-29) 보드 `fa24-PID`(지금 192.168.40.228) worker = `VIEWER_IMAGE-261028`로 교체** → `REF-build.md` "보드 worker 교체"
- 파일 편집기·보기 타일·영상·타일 끌기 네 파트 REF는 아직 INDEX로 안 접음(MEMORY·HISTORY가 10k에 닿으면 접기, 파트 묶음은 사용자 확인)
- **git**: master = 파트(타일 끌기) 머지(10-29), push = 사용자. 남은 로컬 브랜치는 아래 "커밋 상태" 줄 + `feat/file-edit`·`feat/file-view`·`feat/video-view`·`feat/tile-drag`(머지됨). worktree는 원래 폴더 하나
- **지금 떠 있는 것**(2026-10-29): test3(8013) = supervisor·worker **`build/video_view_tmp`**. 보관 프로세스 2개(27975·28612)는 실행 파일이 **`build/viewer_image/worker`** -> 그 터미널을 닫기 전엔 `build/viewer_image/` 지우지 말 것. `build/api_prefix/`·`build/mochi/`는 쓰는 프로세스 없음(10-29 확인, 지울지는 사용자)
- **`build/`**(2026-10-29): `video_view_tmp/`(지금 test3, 임시)·`viewer_image/`·`mochi/`·`api_prefix/`·`supervisor.env`·`worker.env`·`worker-sqlite.db*`·`view-cache/`·`home-update/`
- **BACKLOG(10-26 추가)**: Windows 터미널 출력 줄 빠짐(ConPTY 의심, 사용자가 시험 넘김) · 주소 목록 3개 상한이 지금 주소를 밀어냄
- 옛 CURRENT 줄 원문 → `history/current-archive.md`(10-28 밤 ~ 10-29 파트(타일 끌기) 진행 줄), 그 전 = `-4`(10-27 ~ 10-28 경로 줄 정리)·`-3`·`-2`·`-1`
- 빌드 = 태그(`apps/core/cmd/*/env` 첫 줄) 바꾸고 `scripts/build.sh <이름>`(태그 출력·같은 태그 경고) → `REF-build.md` "빌드 태그"
- 남은 흔적(정리 여부 사용자 미정): test1용 DB `nexus_keep`(시험 계정 `smoke<시각>` 1개)·`build/keep`의 env·sqlite(test1 = 재부팅 뒤 안 켬) · `nexus_build`에 `smoke1792672658797`·`smoke1792672663267`(폴더 233·스크립트 234). 보관 프로세스 일괄 종료 `pkill -f 'worker -hold'`는 **이 Claude 세션도 죽으니 금지**
- 두 빌드를 나란히 띄우는 법(env 분리·DB 복사) → `REF-build.md` "두 빌드 나란히"
- **나중에 할 것 = `BACKLOG.md`**(2026-10-22 신설, 세션 시작 시 안 읽음): 다음 파트 후보·알려진 버그·보류 설계·열린 질문·답 대기 / PC 재부팅은 다음 작업이 없을 때 확인
- **최근 마감 파트**(전체 목록·순서 = `HISTORY.md` "마감 파트"): 젤리·관리 기록 `INDEX-jelly`(10-27) · 회원 관리 `INDEX-member`(10-26) · supervisor 주소 실시간 변경 `INDEX-supervisor-move`(10-26) · 빈 곳 길게 누르기 `INDEX-press-pager`(10-25) · 목록 아이콘 보기 `INDEX-dir-grid`(10-24) · 타일 머리 줄 버튼 `INDEX-tile-head`(10-24)
  - 남겨 둔 것: 확인용 노드 irony `test`/`ctrl-alt-test`(225)·`scroll-test`(215)·`scroll-test.ps1`(216) / VM `nexus-win`(`docker start/stop nexus-win`, 화면 `127.0.0.1:8006`) + VM worker `win-dev` → `REF-windows-vm.md` "VM 실행" / Windows 미확인 항목 → `BACKLOG.md`
- **시연 데이터(2026-10-16) 등록 → `REF-demo.md`**(시연 순서 점검은 `BACKLOG.md`): irony 계정 폴더 `시연-생산지그`(id 200, 안 C(생산 지그 시나리오)) + test1 편집·test2 조작·test3 보기 공유. 코드 변경 없음(REST로 데이터만, 커밋할 코드 없음)
- **외부 접속 설정 사용 중**: 프론트 https://test1.falinux.com(8011) · supervisor https://test2.falinux.com(8012) · worker -> localhost:8012 → `REF-external-access.md`. 로컬 전용(5050·3000)으로 되돌리는 법도 거기
- 남은 주의: 개발 DB에 테스트 계정 `active-<시각>` 3개(폴더 11개씩) / 개발용 로그인 기본값 = `apps/frontend/.env.development.local`(git 제외) / worker는 종료 정책 이후 빌드여야 force를 앎
- 커밋 상태(2026-10-26 세션 끝): 위 첫 줄(브랜치). 남겨 둔 브랜치 `feat/dir-grid`·`feat/press-pager`·`feat/chaltteok-ui`·`feat/supervisor-move`(머지됨)·`refactor/socket`·`refactor/tile-kinds`·`feat/worker-dir`·`feat/transfer-ext`·`feat/transfer-down`·`feat/path-input`·`feat/typography`·`feat/pager-tools`·`feat/tile-head`. 커밋 제외 = `apps/core/cmd/irony/`(사용자 스크래치)·worktree `cmd/worker/test.env`. 커밋은 파트 단위(중간 커밋은 사용자가 요청할 때). user vault 전체 git 제외
- 테스트·확인 준비(환경, 헤드리스, 함정, **격리 실행 = 임시 DB + 5051**) → `REF-test-env.md`
- 역할 분담(2026-09-29 개정): node 타일 UI·process 작업은 코드 대부분을 Claude가 작성하고, 로직 구조는 사용자와 먼저 맞춘 뒤 쓴다


---

## 유지 사항 (항상)
- 색 규칙: Vuetify 테마 색 이름 의존, hex 금지 → `REF-frontend.md` "색 규칙" / `position_x/y` 사용 중단(DB 컬럼 유지) / 외부 레이아웃 라이브러리 안 씀 / 트리 드래그 device_key 재상속 이슈 → `REF-node-label.md`
- EXEC 실행 방식 = A(직접 실행, shebang 필수) 현상 유지 확정(2026-10-12) / kill 정책 → `REF-process-kill.md` "확정"
- 마이그레이션 번호는 착수 때 실제 다음 번호로(`00005`는 `process_subscribers` 제거에 씀)
- 모바일 키보드 주의: `appWindown.store.ts`의 `size.inner`(visualViewport)에 판정을 연결하면 안 됨 → `REF-node-ui-overview.md`
