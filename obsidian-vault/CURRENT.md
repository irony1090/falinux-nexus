# CURRENT

## 현재 날짜
2026-10-27

> 완료·커밋된 작업의 상세는 `history/*.md`, 설계·재사용 지식은 `REF-*.md`, 마감 파트는 `INDEX-*.md`. 여기는 **현재 상태 + 다음 할 것 + 미해결**만.

---

## ⚠️ 다음 세션 시작 시
- **(2026-10-27 밤) 진행 중 파트 없음** — 파트(찰떡 입력칸) 마감·머지 → `INDEX-ct-field.md`. 격리 환경(5051·3001·`dp-dev`·DB `nexus_dp_tmp`) 정리함. 다음 파트 후보 = `BACKLOG.md`
- **오늘 마감**(2026-10-27): 파트(폴더에서 터미널) `INDEX-dir-shell` → 파트(장비 상태 push) `INDEX-device-push` → 파트(칸 메뉴 장비 디렉토리) `INDEX-row-open-dir` → 파트(찰떡 마무리) `INDEX-ui-chaltteok` → 파트(찰떡 입력칸) `INDEX-ct-field`. push = 사용자
  - test3(8013) = 젤리·셸·장비 push·칸 메뉴·찰떡 마무리·찰떡 입력칸 미반영(supervisor·worker·웹 다시 빌드) / 개발 DB `nexus`에 장비 등록 없음 → 장비 타일 쓰려면 "내 장비"에서 `irony-MAC-ADDress1`·`win-dev` 등록 + worker env `API_KEY` / VM `nexus-win` 꺼짐 / 옛 줄 원문 → `history/current-archive.md`
- **(2026-10-26 밤 vault 색인 정리)** 그때 기준 진행 중 파트 없음
  - 오늘 끝난 것: 파트(회원 관리) 마감(`/api` 정리·열린 장비 주소 일괄 변경·폰 수정 포함) → `INDEX-member.md` · 집 서버 배포 22:19 OK → `REF-home-server.md` · worker 폴더 타일 전체 계획(`REF-worker-dir.md`)도 `INDEX-worker-dir`로 접음
  - MEMORY 슬림화(9.9k -> 5.8k자): 확정 결정은 한 줄 요약, 원문 = `REF-decisions.md`(같은 이름 절)
  - 사용자 쪽 남은 것: 집 서버 PC 옛 파일 정리(기준 = `REF-home-server.md` "서버 PC 정리 기준") · push · 원격 `origin/feat/member` 정리
- **git**: master = 파트(찰떡 입력칸) 머지(10-27, 직전 파트 머지 `d5b4963`), push = 사용자. 로컬 브랜치 `refactor/api-prefix`·`fix/phone-layout`(둘 다 머지됨). worktree는 원래 폴더 하나
- **지금 떠 있는 것**(2026-10-26 밤): test3(8013) = supervisor **`build/api_prefix`**(사용자 빌드, **웹만 10-26 밤 `fix/phone-layout` 내용(= master `8de3e05`)으로 교체** — supervisor 재시작 없음) + worker **`build/take_process/worker`**(이번 변경에 worker 없음, 재시작 안 함) -> **`build/take_process/` 지우지 말 것**. 보관 프로세스 771160 = 실행 파일(`build/supervisor_move/worker`)은 지워졌지만 계속 돎(12:12 터미널, take_process worker가 붙어 있음). 격리 5051·VM `nexus-win` 꺼짐. 남긴 DB = `nexus_api_tmp`·`nexus_member_tmp`(시험 계정, 필요 없으면 DROP)
- **`build/`**(2026-10-26 정리): `api_prefix/`·`take_process/`·`supervisor.env`·`worker.env`·`worker-sqlite.db*`·`home-update/`(ps1 스크립트 + `api_prefix/` 묶음·`nexus-api-prefix.zip`)
- **BACKLOG(10-26 추가)**: Windows 터미널 출력 줄 빠짐(ConPTY 의심, 사용자가 시험 넘김) · 주소 목록 3개 상한이 지금 주소를 밀어냄
- 옛 CURRENT 줄 원문 → `history/current-archive.md`(마지막 = "2026-10-26 밤 vault 색인 정리")
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
