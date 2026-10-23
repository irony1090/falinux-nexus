# history — worker 폴더 타일: 폴더-6(보드 이동) + Q9(루트 장비 탭)

> → `REF-worker-dir-board.md` / 앞 이력 → `history/worker-dir-exec.md` / → `CURRENT.md`

- 2026-10-23 폴더-6 설계 제시(서버 변경 없음 확인, 이동-1~5) + Q25(루트 장비 탭 범위)·Q26(연 폴더 삭제 때) — 답 대기
- 2026-10-23 오전 test3 재기동(사용자 터미널, `CLAUDE*` 변수 0개 확인) -> 이어서 PC 재부팅. 재부팅 뒤 폴더-6 설계를 다시 보여 주기로 함(사용자)
- 2026-10-23 재부팅 뒤 설계 다시 제시 -> Q25(루트 장비 탭 범위) = 가(장비의 볼 수 있는 process 전부, 다른 타일에서 띄운 것·공유받은 것 포함) · Q26(연 폴더 삭제 때) = 가(그대로 두고 "볼 수 없음"). 같은 자리에서 나중 작업 3개를 BACKLOG에 기록(파트(worker 폴더 타일) 끝난 뒤 진행)
- 2026-10-23 "진행" -> 이동-1(메뉴)~4(루트 경계) 구현(`RootTabs.vue`·`useRootRuns` 공용화, `GET /workers/:mainKey/runs`) -> 이동-5(확인) 격리 REST 20/20·헤드리스 24/24, 미커밋 → `REF-worker-dir-board.md` "구현"·"확인"
- 2026-10-23 test3를 `build/worker-dir`(태그 `WORKERDIR-261023`)로 사용자가 다시 빌드·실행 + VM worker `win-dev`를 test3에 붙임 -> **사용자 브라우저 확인 OK(Linux·Windows 장비 기능)**. VM 끔. `build.sh`에 태그 출력·경고 추가(→ `REF-build.md` "빌드 태그")
