# history — CURRENT에서 내려보낸 줄 (원문 보존)

→ `CURRENT.md` · 각 줄의 상세는 그 줄이 가리키는 INDEX/REF
> 그 전 줄 → `history/current-archive-4.md`(10-27 ~ 10-28 파트(경로 줄 정리)) · `-3` · `-2` · `-1`

## 2026-10-29 (파트(영상 Range) 마감 때 내려보냄)
- **(2026-10-29) 진행 중 파트 = 파트(영상 Range)**, 브랜치 `feat/video-view`(미커밋) → `REF-video-view.md`
  - 1단계(영상 재생) + 2단계(S1(이어보기), 마이그레이션 `00015`) 구현 + 격리 확인 끝, **사용자 확인 대기** → 확인되면 파트 마감(커밋·머지)
  - **임시 빌드 `build/video_view_tmp`**(태그 `VIDEO_VIEW_TMP-261029`, 미커밋 작업 트리, 폴더에 `TEMP-BUILD.txt`) — test3(8013) 실행 = 사용자. 켜질 때 `nexus_build`에 `00015` 적용. worker는 안 바꿔도 됨(프로토콜 그대로). **사용자 test3 확인: "생각보다 느리지 않음, 괜찮음"**(nginx 손댈 것 없음). 마감 뒤 정식 빌드로 교체
  - 격리 환경 켜 둠: supervisor 5051 + DB `nexus_vv_tmp` + worker `vv-dev` + vite **3001**(계정 `vv1`/`pw123456`, 픽스처 scratchpad `iso/fx`) · TMPDIR `/run/user/1000/vviso`
- **(2026-10-29) 보드 `fa24-PID`(지금 192.168.40.228) worker = `VIEWER_IMAGE-261028`로 교체** → `REF-build.md` "보드 worker 교체"
- **(2026-10-28 밤) 파트(보기 타일) 마감** — `feat/file-view` 커밋 + master 머지, push = 사용자. 진행 중 파트 없음 → `REF-view-tile.md`
  - 다음 후보 = 파트(영상 Range, 영상은 계정 전체 한 타일) → `BACKLOG.md`
  - **빌드 `build/viewer_image`**(태그 `VIEWER_IMAGE-261028`, 2026-10-28 밤, master `709095c`) — test3 재기동은 사용자. `build/supervisor.env`에 `VIEW_CACHE_DIR="../view-cache"` 추가(= `build/view-cache`, 빌드 폴더들이 같이 씀, 첫 기동 때 생김). embed env도 같은 값 + 태그
  - test3 재기동 = 사용자(10-28 밤) · 보드 `imx8mp-f24` worker도 `VIEWER_IMAGE-261028`로 교체 — **집 서버에 붙어 있음**(주소 목록 1순위) → `REF-build.md` "보드 worker 교체"
  - 남긴 것: DB `nexus_fv_tmp`(격리 확인용, 필요 없으면 DROP) · `/run/user/1000/fviso`. 집 서버 = 배포 완료(아래 줄), test3(8013) = `build/viewer_image` 재기동은 사용자
- **(10-28 밤) 집 서버 = viewer-image 배포 완료**(사용자) — master `709095c`, 태그 `VIEWER_IMAGE-HOMESERVER-261028`, exe 2개 + web(묶음 `build/home-update/nexus-viewer-image.zip`). 편집기 = 거절 메시지 확인 뒤 `Microsoft.Edit` 설치 + worker 재시작 -> "편집" 정상(`user/home-server/편집기-설치.md`) · 서버 PC 옛 파일 정리 = 사용자 → `REF-home-deploy.md` "viewer-image 배포 묶음"
- **(10-28 밤) 집 서버 직통 전환 끝** — `nexus`·`direct` 둘 다 iptime + Caddy 직통(타자 반응 터널보다 훨씬 좋음), cloudflared = Stopped·Disabled, 내부 IP `192.168.0.2` 고정(사용자). 남은 것 = 내일(10-29)까지 써 보고 cloudflared 삭제 → `BACKLOG.md` "cloudflared 삭제" (보드 `imx8mp-f24` 확인은 안 함, 사용자) → `REF-home-direct.md` · `user/home-server/주소-전환-명령.md`
- **(2026-10-28 저녁) 파트(파일 편집기) 마감** · `feat/file-edit` 커밋 + master 머지, push = 사용자 · 장비 타일 칸 메뉴 "편집"(worker 편집기)·"새 파일"·칸 메뉴 "내려받기", Windows VM 확인 → `REF-file-view.md` · 다음 후보 = 보기 타일(V1~V5 기본안 확인부터) → `BACKLOG.md` · 집 서버엔 편집기 일부러 안 깖(사용자가 차이 볼 예정)
- **(2026-10-28) 진행 중 파트 없음** — 오늘 마감: 파트(경로 줄 정리) `INDEX-path-row` → 파트(패널 창 버튼) `INDEX-pager-head-btns`(미해결 1건 → `BACKLOG.md`) → 파트(닫은 뒤 active) `INDEX-close-active`. 셋 다 master 머지, push = 사용자. 격리 환경 끔, DB `nexus_member_tmp`는 원래대로 남김. 다음 파트 후보 = `BACKLOG.md`(첫 줄 "타일 ESC로 닫기")
  - (낮 기준) 집 서버 = mochi exe + 웹만 3파트 반영본 → 밤에 viewer-image로 전체 교체 · cloudflared 공지 = 조치 불필요(10.3, 밤에 cloudflared 정지) / 보드 worker 2대 mochi → `REF-build.md` "보드 worker 교체" / test3(8013) = `build/mochi` exe(Go 변경 없음) + **웹만 오늘 3파트 반영본으로 교체**(`build/mochi/web`, 옛 웹 `web.bak-*`)
- **오늘 마감**(2026-10-27): 파트(폴더에서 터미널) `INDEX-dir-shell` → 파트(장비 상태 push) `INDEX-device-push` → 파트(칸 메뉴 장비 디렉토리) `INDEX-row-open-dir` → 파트(찰떡 마무리) `INDEX-ui-chaltteok` → 파트(찰떡 입력칸) `INDEX-ct-field`. push = 사용자
  - test3(8013) = **`build/mochi`**(태그 `MOCHI-261027`, 사용자 빌드·재시작 2026-10-27 밤, master `28053b7` = 오늘 6파트 전부 반영) / 개발 DB `nexus`에 장비 등록 없음 → 장비 타일 쓰려면 "내 장비"에서 `irony-MAC-ADDress1`·`win-dev` 등록 + worker env `API_KEY` / VM `nexus-win` 꺼짐 / 옛 줄 원문 → `history/current-archive.md`
- **(2026-10-28) 보드 worker 2대 mochi로 교체 끝** — `fa24-PID`·`imx8mp-f24` test3 등록 OK → `REF-build.md` "보드 worker 교체" / 집 서버 cloudflared 공지 = 조치 불필요 → `REF-home-server.md` "cloudflared 보안 공지"
- **(2026-10-26 밤 vault 색인 정리)** 그때 기준 진행 중 파트 없음
  - 오늘 끝난 것: 파트(회원 관리) 마감(`/api` 정리·열린 장비 주소 일괄 변경·폰 수정 포함) → `INDEX-member.md` · 집 서버 배포 22:19 OK → `REF-home-deploy.md` · worker 폴더 타일 전체 계획(`REF-worker-dir.md`)도 `INDEX-worker-dir`로 접음
  - MEMORY 슬림화(9.9k -> 5.8k자): 확정 결정은 한 줄 요약, 원문 = `REF-decisions.md`(같은 이름 절)
  - 사용자 쪽 남은 것: 집 서버 PC 옛 파일 정리(기준 = `REF-home-deploy.md` "서버 PC 정리 기준") · push · 원격 `origin/feat/member` 정리
