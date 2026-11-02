# history — CURRENT에서 내려보낸 줄 (원문 보존, 2026-10-30 세션 끝)

→ `CURRENT.md` · 각 줄의 상세는 그 줄이 가리키는 INDEX/REF
> 그 전 → `history/current-archive.md`(10-28 밤 ~ 10-29) · `-4` · `-3` · `-2` · `-1`

## 2026-10-30 (업데이트 app 설계 세션 끝에 내려보냄)
- **(2026-10-30) 파트(내 장비·관리자 타일 이동) 마감** — `feat/member-drag` 커밋 + master 머지, push = 사용자. → `REF-member-drag.md` · `history/member-drag.md`
  - 열기 규칙(화-9 = 보드 `openTile`과 같게 — 연 곳 바로 뒤, 연 곳 없으면 맨 앞) · 끌기(그 타일만, `parent` 유지, 머리 줄 + 몸통 빈 곳에서 시작, 순서는 URL `t`에만) · 끝 띠 페이지 넘기기 · 끌기 시작하면 팝업 닫기(관리 페이지 + 보드)
  - 사용자 확인 완료(PC·관리자·내 장비·폰) — 실기기에서 찾은 것 2건 고침 = 세로 2칸 타일 위/아래 절반 판정, 반 칸 타일 아래 빈 칸
  - 미확인으로 남긴 것 = 회원→장비 자식 붙기(dev DB에 장비 없음) · 경로 재열림(`openParent`)이 자식 묶음 중간에 낄 수 있음(옛 동작, 안 고침)
  - 그 세션에 띄운 것: test1(8011 vite) · test2(8012 개발 supervisor, DB `nexus`) — 세션 백그라운드 작업. test3(8013 `build/video_view_tmp`, DB `nexus_build`)는 손대지 않음. 계정 `irony`(dev DB admin은 사용자가 넣음, 명령 = `user/dev-admin-grant.md`)
- **(2026-10-29 저녁) 집 서버 묶음 `build/home-update/nexus-drag.zip`**(master `29ae134`, worker 포함, 태그 `DRAG-HOMESERVER-261029`) — **배포 완료(사용자, 10-29 저녁)** = 집 서버도 영상 Range + 타일 끌기 → `REF-home-deploy.md` "drag 배포 묶음". 로컬 `check.ps1`·`prepare.ps1` 지워도 됨(지울지 사용자)
- **(2026-10-29) 파트(타일 끌기) 마감** — `feat/tile-drag` 커밋 + master 머지, push = 사용자. → `REF-tile-drag.md`(계획·결정) · `REF-tile-drag-impl.md`(구현·함정) · `REF-tile-order.md`(순서 구조) · 결정 = MEMORY "터치 제스처"·"끌기"(상세 `REF-decisions-ui.md`)
  - 구조-1~3 · 끌기-1·2 · 제스처-1 · 보드-1 · 칸-1·2 · 끝 띠 + 선택 막대 배치 고침 · 로그인 제목 `NEXUS`
  - **빌드 = 사용자가 직접**: 태그 `# SUPERVISOR-TILE_DRAG-261029`·`# WORKER-TILE_DRAG-261029`(env 첫 줄, 아직 `VIDEO_VIEW-261029`) -> `scripts/build.sh 261029_tile_drag`. **DB 마이그레이션 없음**, 옛 타일 문서(kids)는 supervisor가 처음 읽을 때 order로 변환 저장(`tileOrderMigrate.go`, 실데이터 처음 = test3 `nexus_build`) — 배포 뒤 보드가 그대로인지 확인할 것. 프론트·supervisor 같이, 열린 탭 새로고침
  - 아직 떠 있는 것(정리 여부 사용자 미정):
    | 포트 | 무엇 | 끄기 |
    |---|---|---|
    | 8011(test1) | vite(저장소 `apps/frontend`) | `kill $(cat /tmp/claude-1000/-home-jh-bae-irony-nexus/51cf026c-d38c-4465-8fee-75e8924cd07d/scratchpad/ext/vite.pid)` |
    | 8012(test2) | `socat` -> 5051. **떠 있는 동안 개발 supervisor 8012 못 띄움** | `kill $(cat /tmp/claude-1000/-home-jh-bae-irony-nexus/51cf026c-d38c-4465-8fee-75e8924cd07d/scratchpad/ext/socat.pid)` |
    | 5051 | 격리 supervisor(칸-1 코드, `/tmp/claude-1000/-home-jh-bae-irony-nexus/06a4a87f-5384-4e2f-870b-848ca2e42a62/scratchpad/iso/sup`, DB `nexus_td_tmp`) + worker `td-dev` | `/tmp/claude-1000/-home-jh-bae-irony-nexus/06a4a87f-5384-4e2f-870b-848ca2e42a62/scratchpad/iso/stop.sh /tmp/claude-1000/-home-jh-bae-irony-nexus/06a4a87f-5384-4e2f-870b-848ca2e42a62/scratchpad/iso` |
    | 3001 | 격리 vite(`--config .../scratchpad/iso/vite.iso.mts`) | 그 pid kill |
    - 계정 td1·td2(`pw123456`). td1 보드 = 루트 + X(A) + C(칸시험) — 되돌리기 `python3 /tmp/claude-1000/-home-jh-bae-irony-nexus/1dbb5e34-df83-47a1-b7aa-5d313a551b3b/scratchpad/cell/board.py`, 옛 기본 보드 = `python3 /tmp/claude-1000/-home-jh-bae-irony-nexus/51cf026c-d38c-4465-8fee-75e8924cd07d/scratchpad/ui/setup.py`
    - 다 쓰면: 위 프로세스 끄기 + `DROP DATABASE nexus_td_tmp` + `/run/user/1000/tdiso` 지우기
- **(2026-10-29) 파트(영상 Range) 마감** — `feat/video-view` 커밋 + master 머지, push = 사용자. → `REF-video-view.md`
  - 1단계(영상 재생: Range·서버가 넣는 영상 타일·재생 불가·오프라인 이어 재생) + 2단계(S1(이어보기), 마이그레이션 `00015`)
  - **정식 빌드 = 사용자가 직접**: `apps/core/cmd/*/env` 첫 줄 태그가 지금 `VIDEO_VIEW_TMP-261029`(임시) -> 바꾸고 `scripts/build.sh <이름>` -> test3 교체 뒤 `build/video_view_tmp`(폴더에 `TEMP-BUILD.txt`) 지워도 됨
  - test3(8013) 지금 = **`build/video_view_tmp`**(supervisor·worker, 사용자 실행 10-29, `nexus_build` = goose 15). 사용자 체감 "생각보다 느리지 않음"
  - 집 서버 = 영상 파트 아직 안 보냄 → `BACKLOG.md` "집 서버 영상 배포"
  - 남긴 것: DB `nexus_vv_tmp`(격리 확인용, 계정 `vv1`, 필요 없으면 DROP) · 그 세션 scratchpad `iso/`(픽스처·스크립트, 재부팅 때 사라짐). 격리 프로세스·`/run/user/1000/vviso` 정리함
- **(2026-10-29) 보드 `fa24-PID`(지금 192.168.40.228) worker = `VIEWER_IMAGE-261028`로 교체** → `REF-build.md` "보드 worker 교체"
