# history — CURRENT에서 내려보낸 줄 (원문 보존)

→ `CURRENT.md` · 각 줄의 상세는 그 줄이 가리키는 INDEX/REF
> 2026-10-26 낮까지의 줄 → `history/current-archive-1.md` / 밤 세션 끝 → `history/current-archive-2.md` / 10-26 밤 fix/phone-layout ~ vault 색인 정리 → `history/current-archive-3.md`(분할)

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

## 2026-10-27 밤 파트(찰떡 입력칸) 마감 때 내려보낸 줄
- **(2026-10-27) 진행 중 파트 = 찰떡 입력칸** → `REF-ct-field.md`(단계 F1~F4 · 추가 항목 3개 = 전부 구현) / 브랜치 `feat/ct-field`(master `d5b4963`에서, **미커밋, 마감 확인 대기**)
  - 범위 = 페이지의 input·select **전부**(Vuetify 6개 + 직접 만든 `.mb-input` 16개, 로그인·회원가입 포함). 계기 = 찰떡 마무리 항목 5(select 목록 흰 바탕 흰 글씨)
  - 순서 = 조각 하나(`CtTextField`) 먼저 만들고 높이(32px)·끝 모양 확인 -> 나머지. 브랜치는 착수 때 master에서 새로
  - **격리 떠 있음**(세션 재시작해도 남음): supervisor 5051 + DB `nexus_dp_tmp` + vite 3001 + worker `dp-dev`. 켜고 끄기 `/tmp/claude-1000/-home-jh-bae-irony-nexus/483bf060-2b4f-44c6-86c0-6dd9c05f93c4/scratchpad/dp/on.sh`·`off.sh`(끄기 = exe 경로로 고름), 계정 a1 관리자·o2 주인·v3 공유받음·x4 무관(pw123456). 정리 = supervisor·worker·vite를 exe/cmdline으로 골라 kill(`pgrep -f` 문자열은 내 bash도 잡으니 node exe 확인) + `DROP DATABASE nexus_dp_tmp` + `/tmp/nexus-hold-1000/dp-dev`

## 2026-10-27 밤 test3 = mochi 반영 때 내려보낸 줄
  - test3(8013) = 젤리·셸·장비 push·칸 메뉴·찰떡 마무리·찰떡 입력칸 미반영(supervisor·worker·웹 다시 빌드) / 개발 DB `nexus`에 장비 등록 없음 → 장비 타일 쓰려면 "내 장비"에서 `irony-MAC-ADDress1`·`win-dev` 등록 + worker env `API_KEY` / VM `nexus-win` 꺼짐 / 옛 줄 원문 → `history/current-archive.md`
- **지금 떠 있는 것**(2026-10-26 밤): test3(8013) = supervisor **`build/api_prefix`**(사용자 빌드, **웹만 10-26 밤 `fix/phone-layout` 내용(= master `8de3e05`)으로 교체** — supervisor 재시작 없음) + worker **`build/take_process/worker`**(이번 변경에 worker 없음, 재시작 안 함) -> **`build/take_process/` 지우지 말 것**. 보관 프로세스 771160 = 실행 파일(`build/supervisor_move/worker`)은 지워졌지만 계속 돎(12:12 터미널, take_process worker가 붙어 있음). 격리 5051·VM `nexus-win` 꺼짐. 남긴 DB = `nexus_api_tmp`·`nexus_member_tmp`(시험 계정, 필요 없으면 DROP)

## 2026-10-28 파트(경로 줄 정리) 마감 때 내려보냄
- **(2026-10-27 밤) 진행 중 파트 없음** — 파트(찰떡 입력칸) 마감·머지 → `INDEX-ct-field.md`. 격리 환경(5051·3001·`dp-dev`·DB `nexus_dp_tmp`) 정리함. 다음 파트 후보 = `BACKLOG.md`
  - **집 서버 묶음 `build/home-update/nexus-mochi.zip`**(master `28053b7`, exe 2개 + web, 2026-10-27 밤) = 사용자가 집에서 배포 `OK`(10-27 밤) → `REF-home-deploy.md` "mochi 배포 묶음"
  - **다음 파트(예정, 사용자 2026-10-27) = 타일 ESC로 닫기** → `BACKLOG.md` 첫 줄(착수 때 조율할 것 포함). 기록만, 착수 전
- **(2026-10-28) 진행 중 파트 = 경로 줄 정리**(브랜치 `feat/path-row`, 커밋 전) → `REF-path-row.md`. 구현-1~5 끝, 격리 헤드리스 확인 OK(`vue-tsc` 통과). **사용자 화면 확인 대기** — 격리 띄워 둠: 화면 http://localhost:3001 (alice/pw) · supervisor 5051(`nexus_member_tmp`) · worker `jig-a`(scratchpad `iso/wrk`, TMPDIR `/tmp/claude-1000/pr-t`). 미확인 = 실제 폰 터치 끌기·어두운 테마

## 2026-10-28 파트(패널 창 버튼) 마감 때 내려보냄
- **(2026-10-28) 진행 중 파트 = 패널 창 버튼**(브랜치 `feat/pager-head-btns`, 커밋 전) → `REF-pager-head-btns.md`. 패널-1~3 끝, 격리 헤드리스 확인 OK. **사용자 화면 확인 대기** — 격리 띄워 둠: http://localhost:3001 (alice/pw), 관리 페이지 `/admin`. 격리 worker에 시험 셸 터미널 2개(하나 최소화) 남음 -> 마감 때 정리. "타일 ESC로 닫기"는 `BACKLOG.md` 첫 줄 그대로

## 2026-10-28 파트(닫은 뒤 active) 마감 때 내려보냄
- **(2026-10-28) 진행 중 파트 = 닫은 뒤 active**(브랜치 `feat/close-active`, 커밋 전) → `REF-close-active.md`. 구현·격리 헤드리스 확인 OK. **사용자 화면 확인 대기** — 격리 띄워 둠: http://localhost:3001 (alice/pw), worker TMPDIR `/tmp/claude-1000/pr-t`, 시험 셸 1개(최소화) 남음 -> 마감 때 정리

## 2026-10-28 세션 끝 — CURRENT에서 내려보냄
- **(2026-10-28 오후) 집 서버 직통 전환 진행 중(코드 무관, 운영)** — 원인 = Cloudflare 엣지 LAX 경유. iptime 포트포워딩 + Caddy로 직통. 회사에서 할 것(DNS `direct`·Caddy·방화벽·서비스 등록) 끝, **오늘 밤 집에서(사용자) 1(내부 IP 고정)·2(포트포워딩)·6(시험 실행)** -> `direct` 확인되면 8(주소 전환, 구름만 회색 금지 = 대상도 iptime으로) → 체크리스트 user vault `user/home-server/직통-전환.md` · 근거 `REF-home-server.md` "터널 속도"·"직통 전환 계획(안)"

- (2026-10-28 밤 내려보냄, ② 바이너리 판정 = 사용자 OK로 해결) **(2026-10-28 세션 끝) 다음 = 파트(보기 타일)** — D(더블클릭 열기) 답: ① 해석 맞음 ③ 영상 전엔 아무 일 없음 ④ 보기 타일에 같이. **② 바이너리 판정 제안(앞 8KB NUL, 더블클릭 때만) 확인 대기** → `REF-file-view.md` "다음 파트(보기 타일) 결정". V1~V5 확정, 읽기 전용 텍스트 안 함. 브랜치는 착수 때

- (2026-10-28 밤 마감 때 내려보냄) **(2026-10-28 밤 세션 끝) 진행 중 = 파트(보기 타일)** `feat/file-view` **미커밋**(코드 전부 작업 트리에) — 1단계(사본 없는 뷰어 + D(더블클릭 열기))·2단계(이미지 사본 · 터미널 탭 보기 줄 · 편집 줄 이름) 구현·격리 확인 끝 → `REF-view-tile.md`
  - **다음 세션 첫 일**: 파트(보기 타일) 마감·커밋(+ master 머지) 할지 사용자에게 묻기. 그다음 후보 = 파트(영상 Range, 영상은 계정 전체 한 타일) → `BACKLOG.md`
  - 격리 환경 끔(supervisor 5051·worker `fv-dev`·vite 3001). **남긴 것**: DB `nexus_fv_tmp`(계정 `fv1`/`pw123456`, 장비 `fv-dev`, 폴더 id 1) · `/run/user/1000/fviso` · 이 세션 scratchpad(`iso/`, 재부팅 시 사라짐). 다시 띄우는 법 = `REF-view-tile.md` "격리 확인" + `REF-member-impl.md` "격리 환경 다시 띄우기". 필요 없으면 DROP
  - 커밋 때 빼지 말 것: `apps/core/.gitignore`(view-cache) · 생성 코드 `internal/supervisor/db/gen`. embed `cmd/supervisor/env`에도 `VIEW_CACHE_DIR` 주석 줄 추가함(git 제외 파일)
