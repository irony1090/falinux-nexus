# history — CURRENT에서 내려보낸 줄 (원문 보존)

→ `CURRENT.md` · 각 줄의 상세는 그 줄이 가리키는 INDEX/REF
> 2026-10-26 낮까지의 줄 → `history/current-archive-1.md`(분할)

## 2026-10-26 밤 세션 끝 (CURRENT "다음 세션 시작 시"에서 내림, 원문)
- **(2026-10-26 오후) `feat/member` master 머지 끝(`4ae97a7`, push = 사용자) · 집 서버 회원 관리 배포 진행 중 = Nexus 터미널만으로** → `REF-home-server.md` "회원 관리 배포 — Nexus만으로" **집 서버 배포 완료(16:59 `state=DONE`)**. 파트(회원 관리) 마감 정리는 아직(사용자 확인 대기)
- **(2026-10-26 저녁 세션 끝, 사용자 퇴근 -> 집에선 집 서버 ssh·원격 가능)** 브랜치 `refactor/api-prefix`(작업 폴더 = 원래 폴더, **커밋 전**) = ① 브라우저 API 전부 `/api`(`REF-api-prefix.md`, 격리 14/14) ② 열린 장비 타일 supervisor 주소 일괄 변경 + 푸터 액션 자리(`REF-device-move.md`, 격리 23/24 — 실패 1 = 주소 3개 상한 함정, BACKLOG)
  - 커밋 `9c1651d` + master 머지 `c5f9749`(2026-10-26 저녁, 사용자 요청). push = 사용자
  - worktree `nexus-member` 삭제 + 로컬 `feat/member` 삭제(master에 포함 확인). 원격 `origin/feat/member`는 사용자가 정리. 지우기 전 `KEY_SECRET` 줄을 원래 폴더 `cmd/supervisor/env`(git 밖)로 옮김
  - 이 PC 재빌드(사용자 진행 예정): 빌드명 추천 `api_prefix`(태그 `SUPERVISOR-/WORKER-API_PREFIX-261026`, `cmd/*/env`는 git 밖이라 태그는 손으로 되돌림). **supervisor + web만 교체, worker는 재시작 불필요**(이번 변경에 `cmd/worker`·`internal` 차이 없음, `/worker` 그대로, 끊기면 스스로 재접속) -> worker는 `take_process/worker`로 계속 돌므로 **`build/take_process/` 지우지 말 것**
  - 이 PC = 사용자가 `build/api_prefix` 빌드함 / 집 서버 묶음 = `build/home-update/nexus-api-prefix.zip`(supervisor + web + `update.ps1`, 사용자가 집에서 배포) → `REF-home-server.md` 끝
  - **브랜치 `fix/phone-layout`(커밋 전, 사용자 화면 확인 대기)**: ① 폰 관리 페이지 첫 타일로 안 넘어가던 버그 — 원인 = `memberUrl.hook.ts` 감시 원천에 `grids 있음`이 섞여 첫 타일(push 전) 때 옛 URL을 다시 적용해 타일을 지움 -> 원천 = route만 + `waitGrid`(측정 전이라 미룬 적용이 있을 때만 측정 뒤 `sync`) ② 보드 전체보기 폰도 가로 한 줄(`ovDir` 삭제, EV2 개정 → `REF-board-preview.md`), 보드 여럿이면 폰은 이름을 줄 위·왼쪽 정렬. 격리 `phonefix.mjs` 7/7 + `multi.mjs`(보드 2개 이름 위치) 통과
  - 다음: 사용자 화면 확인 -> 배포 = **supervisor + 웹만**(둘을 같이, worker 그대로): test3 = 새 빌드 / 집 서버 = 집에서 원격으로 exe·web 교체(남겨 둔 `build/home-update/` 스크립트 재사용 가능, env 키는 이미 있음)
  - 파트(회원 관리) 마감 정리(MEMORY 불변 결정·`INDEX-member` 접기 — MEMORY 10.3k자라 이때 줄임)도 같이
  - `build/` 정리함(사용자 요청): 남은 것 = `take_process/`·`supervisor.env`·`worker.env`·`worker-sqlite.db*` + `home-update/`의 ps1 스크립트만. 옛 빌드본 4개·`supervisor_move`·집 서버 옛 zip 2개·`nexus-autostart.ps1`·`nexus-create-user.sql` 삭제(내용은 `REF-home-server.md`에 설명 남음). `supervisor_move/worker -hold`(pid 771160, 12:12 터미널)는 실행 파일만 지워졌고 계속 돎(take_process worker가 붙어 있음)
  - 격리 흔적: DB `nexus_api_tmp`(시험 계정들) 남김 — 필요 없으면 DROP
- **(2026-10-26 세션 끝, 사용자 "여기까지 오케이" -> 세션 재시작)** 다음 세션 첫 할 일 = 사용자가 test3에서 써 본 결과 듣기 -> 문제없으면 파트(회원 관리) 마감
- **(2026-10-26 세션 끝) 파트(회원 관리) = 코드·확인·커밋 끝, 사용자가 test3.falinux.com에 직접 새 배포함** — test3(8013) = **`build/take_process`**(사용자 빌드, `nexus_build` goose v14 = `00014` 적용 확인). 사용자 확인 결과 듣고 이어서
  - (옛 줄) 브랜치: `feat/member`(worktree `/home/jh-bae/irony/nexus-member`, 2026-10-26 저녁 삭제) = `1bd58e7`(회원 관리 전부: 서버·UI 개편·컨트롤 패널·알약 끌기·빈 곳 휠·판 계약) -> `fb2fa56`(master 머지 = supervisor-move 포함) -> `fa50c64`(미등록 process 가져오기). master = `efbe403`(supervisor-move 머지) + vault. **push = 사용자가 함**(분기 태그는 따로 `git push origin <tag>`)
  - (해결된 듯 — 사용자가 `build/take_process`로 배포) 배포 중 막힘: 사용자가 원래 폴더에서 `git checkout feat/member` -> "already checked out at nexus-member". 제안 = ① worktree 폴더에서 바로 빌드(권장, env·`KEY_SECRET` 이미 있음) / ② `git worktree remove --force` 뒤 원래 폴더에서 체크아웃(`test.env`·`node_modules` 링크·env 복사본 지워짐 — 옮긴 뒤)
  - **관리자 지정 끝**: 사용자 "test3는 irony 닉네임 계정" -> `nexus_build`에 닉네임 `irony` 없음, 로그인 아이디 `irony` = 닉네임 `아이로니`(id 3, sha256 일치) -> 사용자 "관리자 지정해줘" -> Claude가 `UPDATE ... WHERE id = 3` 실행, perms = `{worker_register,admin}`
  - **보드 worker 교체 끝(2026-10-26, 사용자 요청)**: `fa24-PID`(transitBoard 보드 `root@192.168.40.203`, `/home/root/nexus-worker`) · `imx8mp-f24`(mirror-info 보드 `falinux@192.168.80.206`, `/opt/nexus-worker`, sudo) — `fa50c64` arm64 정적 + env `API_KEY` + `systemctl restart` -> 둘 다 test3 등록 성공·보관 process 다시 붙음. 기록 = 각 작업공간 vault(transitBoard `REF-board-nexus.md` · mirror-info `REF-board.md` "Nexus worker")
  - 배포 전 확인 나머지 → `REF-member.md` "배포 전 확인"(`KEY_SECRET`·worker 실행 파일 + `API_KEY`·기존 폴더 403). 새 supervisor DB면 장비를 **같은 키로** 등록해야 worker가 붙음(→ `REF-supervisor-move.md` "가져오기 구현" 함정)
  - 배포 뒤 = 파트(회원 관리) 마감: MEMORY 불변 결정 정리(판 계약·회원 권한 등) + `INDEX-member` 접기 + `feat/member` master 머지
  - 이번 세션 경위 → `history/member-ui.md`·`history/member-impl.md`·`history/supervisor-move.md` 2026-10-26 오후. 옛 줄(재부팅 직전·UI 계획·집 서버·supervisor-move 확인 대기)은 `history/current-archive.md` "2026-10-26 세션 끝"
- **지금 떠 있는 것**(2026-10-26 세션 끝): 파트(회원 관리) 격리 **전부 닫음**(5051~5053·3001~3003·worker `jig-a`) — DB `nexus_member_tmp`는 남김(다시 띄우는 법 `REF-member-impl.md`), 임시 `nexus_adopt_b`·`_c` 지움. **test3(8013) = `build/take_process`**(DB `nexus_build`, goose v14, 사용자가 2026-10-26 배포). 8012는 다른 프로젝트(`irony/community`) vite. test1(8011)·개발(8012) 꺼짐. 보관 프로세스 19096·19684·61491 = `build/keep/worker -hold` -> **`build/keep` 지우지 말 것**. VM `nexus-win` 꺼짐(→ `REF-windows-vm.md` "VM 실행"). 웹 백업(`web.bak*`)은 2026-10-25 전부 지움

## 2026-10-26 밤 — fix/phone-layout 머지 때 내려보낸 줄(원문)
  1. **`fix/phone-layout` 브랜치(작업 폴더 = 원래 폴더, 코드 커밋 전)** 사용자 화면 확인 -> 커밋 + master 머지 -> 이 PC supervisor·웹 재빌드. 내용: ① 폰 관리 페이지 첫 타일로 안 넘어가던 버그(`memberUrl.hook.ts`, 원인·수정 → `history/member-ui.md` 10-26 밤) ② 보드 전체보기 폰도 가로 한 줄 + 보드 여럿이면 이름을 줄 위(EV2 개정 → `REF-board-preview.md`). 격리 `phonefix.mjs` 7/7 + 보드 2개 이름 위치 통과
- **새 버그 2개(10-26 밤, 사용자 폰 확인 중 발견) — 수정·`build/api_prefix` 웹 교체 완료, 사용자 확인 대기(`fix/phone-layout`에 같이 커밋)** — 사진 `obsidian-vault/user/Screenshot_20261026_193456_Edge.jpg`(푸터)·`..._193545_Edge.jpg`(자동완성)
  - 버그(빈 페이지 도트): 폰 내 장비 관리에서 마지막 타일을 닫으면 빈 안내 페이지(`MemberPage` `.empty`)는 남는데 `dotCount = lead + grids.length` = 1 -> 지금 위치 표시(pos 1)가 도트 줄 폭 밖에 그려져 잘림. 수정 = `lead + max(1, grids.length)`(`memberBoard.store.ts`)
  - 버그(자동완성 비침): 장비 타일 경로 줄 `.crumbs`가 sticky + `z-index: 1` -> 쌓임 맥락이 1이라 목록 칸 `⋮`(`.dg-btn` 2)·`링크` 라벨(`.dg-r/.dg-l` 1)이 자동완성 위로. 수정 = `.crumbs` z-index 3(`DeviceTileBody.vue`). 폴더 타일은 `.crumbs`가 sticky 아니라 해당 없음
  - 추가(10-26 밤, 사용자 요청 가(화면 이동만)): 상세 타일을 다 닫으면 목록으로 — `memberBoard.store` `close`에서 0개면 `goCol(0)`(폰 = 첫 칸, PC 변화 없음). 관리자 페이지도 같은 store라 같이 적용(도트 수정 포함), 자동완성 비침은 보드 장비 타일 전용이라 관리 쪽 해당 없음. 웹 3번째 교체(직전 웹 = scratchpad `web-prev-2`), 사용자 확인 대기
- **git**: master = `a56b98f`(`refactor/api-prefix` 머지 `c5f9749` 포함, push = 사용자). 로컬 브랜치 `refactor/api-prefix`(머지됨) · `fix/phone-layout`(진행 중). worktree는 원래 폴더 하나(`nexus-member` 삭제, 원격 `origin/feat/member` 정리는 사용자)
- **지금 떠 있는 것**(2026-10-26 밤): test3(8013) = supervisor **`build/api_prefix`**(사용자 빌드, **웹만 10-26 밤 `fix/phone-layout` 작업 트리로 교체(2번째 = 새 버그 2개 포함)** — supervisor 재시작 없음, master 웹은 세션 scratchpad `web-old-1018`) + worker **`build/take_process/worker`**(이번 변경에 worker 없음, 재시작 안 함) -> **`build/take_process/` 지우지 말 것**. 보관 프로세스 771160 = 실행 파일(`build/supervisor_move/worker`)은 지워졌지만 계속 돎(12:12 터미널, take_process worker가 붙어 있음). 격리 5051·VM `nexus-win` 꺼짐. 남긴 DB = `nexus_api_tmp`·`nexus_member_tmp`(시험 계정, 필요 없으면 DROP)
