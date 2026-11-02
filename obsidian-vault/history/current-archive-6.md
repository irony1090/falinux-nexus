# history — CURRENT에서 내려보낸 줄 원문 (6, 2026-11-02 저녁)

> 업데이트 app 파트(10-30 설계 상태, 11-02 오후 재검토 중 줄)의 옛 상태. 현재 상태는 `CURRENT.md`, 결정은 `REF-update-redesign.md`. 앞 = `current-archive-5.md`

- **(2026-11-02 오후) 파트(업데이트 app) 설계 재검토 진행 중** — 설계-1(지우고 깔기)~설계-7(trap 신호)·설계-9(개별 실행과 설치본) 합의 끝, 릴리스 구조 개정(장비 루트 > 빌드, `common` = `uses` 의존)·`steps.json` `deploy`/`rollback` + `builtin`·웹 편집(P1 = 원문 편집 먼저). **결정 기록 = `REF-update-redesign.md`**. 남은 것 = 설계-8(supervisor 되돌리기)·설계-10(자기 교체)·설계-11·설계-12 닫음(시안: Claude 아티팩트 + DeepSeek v2, `user/updater-화면-*.html`, 사용자가 두 시안 확인 중) -> **다음 = DB 초안 개정** + 미확정 접속 토큰·공용 묶음 잠금·실행 권한. 다 정하면 `REF-update-deploy.md`·`REF-update-rules.md`·시안 일괄 개정
- **(2026-11-02) 다음 세션 첫 일 = 파트(업데이트 app) 설계 재검토·결정** — 코드 없음, 설계·시안만. `REF-review-1030.md` "설계" 표의 설계-1(지우고 깔기)~설계-6(버전 이름 체계)부터 하나씩 정하고, 설계-7(trap 신호)~설계-12(시안 표기)는 그다음. 정해지면 `REF-update-deploy.md`·`REF-update-rules.md`·시안에 반영
  - 볼 것 = `REF-review-1030.md` + `REF-update-deploy.md` + `REF-update-rules.md` + 시안 `user/배포-화면-찰떡.html`(4탭)
  - 배포 화면은 `MemberPage` 타일 틀을 쓰므로 열기 순서(연 곳 바로 뒤, 없으면 맨 앞)·끌기(`DragSession`)를 따른다 → 설계-11(배포 정보 타일 열기)
- **(2026-10-30) 파트(업데이트 app) — 설계만 끝, 코드 착수 전** → `REF-update-deploy.md`(설계·화면) · `REF-update-rules.md`(확정 규칙) · `history/update-deploy.md` · `history/update-rules.md`
  - 확정: 장비마다 `nexus-update` 상주(Job 밖) · 관리자 웹 화면 버튼 · 장비별+전체 배포 · 릴리스 이력+되돌리기 · **Linux 먼저**
  - 릴리스 = 폴더 트리(`common` + `<os>-<arch>`), 판정은 파일 유무 · 단계 = 스크립트 1개 + `steps.json` · **잠금 = arch 단위(`배포 완료` 확정)** · 되돌리기 = 이전 릴리스 재배포 · 버전 = 빌드에 박힌 env 첫 줄 · 내 장비 = 인스턴스별 `배포 정보` 타일(주인도)
  - **키 = `메인키#서브키`(업데이터 인스턴스)** — 같은 이름 두 대도 각각 타일·배포·이력·잠금
  - 화면 시안 = `user/배포-화면-찰떡.html`(4탭, 렌더 9장 = `user/배포-화면-찰떡/`) · 원판 `user/배포-화면.html` · 설계안 `user/배포-설계.html` → **사용자 확인 대기**
  - **다음 = P1(보관·화면)**. 착수 전 확정 3개 = 콘솔 인터랙션(가/나) · 스크립트를 `common`에 둘지 · 첫 시험 장비(`fa24-PID`/`imx8mp-f24`)
  - 참고: 이 세션 = `ANTHROPIC_API_KEY` 인증이라 아티팩트 게시 불가 → HTML을 `user/`에 둠
