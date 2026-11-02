# history — 업데이트 app P1(보관·화면) 구현

> → `REF-update-p1.md` / 설계 → `REF-update-deploy.md` / 현재 진행 → `CURRENT.md`

## 2026-11-02 (저녁) — P1 코드 + 확인
- 사용자 "진행" -> 브랜치 `feat/updater-p1`
- 단위-1(DB): `00016_updates.sql` 8테이블(차이-1(장비 루트 테이블) `update_devices` 추가) + `query/updates.sql`, `update_uses.ord`(차이-3(uses 순서))
- 단위-2(보관소): `updstore` 패키지 — 폴더 `devices/`·`common/`(차이-2(보관 폴더 접두어)), 하드링크 복사, 폴더 sha, steps.json 위치 기억 파서 + 단위 테스트
- 단위-3(API): `/api/admin/updates/...` — 격리 supervisor(임시 DB `nexus_upd_tmp` + 5051)에서 `updates-api.py` 39개 통과. 처음 1건 실패 = 문법 오류가 있으면 참조 검사를 건너뜀 -> 둘 다 보이게 고침
- 단위-7(CtCode): CodeMirror 6 패키지 추가(`@codemirror/*` · `@lezer/highlight`), `CtCode.vue`, `ScriptEditDialog` 교체(불러온 뒤 포커스)
- 단위-4(페이지 틀): 회원 페이지 틀 공용화 — `MbTileFrame` 분리, `MemberPage` 슬롯, 키 규칙·탭 표 매개변수, 닫기 확인. `run.sh`(회원 회귀 29개) 통과
- 단위-5·6: updater 목록·장비·빌드·Files·steps.json·Script·공용 타일, 이름 입력 창, zip 메뉴
- 단위-8(확인): `updater-ui.mjs` 27개 통과. 고친 것 = 텍스트 판정(NUL만 보던 것 -> 제어 문자 비율 추가), 빌드 타일 800px 높이에서 `열기` 줄 잘림(차이-4(빌드 타일 행)). 스크린샷 확인(PC·폰·다크)
- 마지막: `run.sh` 전체(회원 + API + 화면) 통과, 임시 DB 지움. 커밋 안 함
- (같은 브랜치) `apps/frontend/tsconfig.app.json`의 `baseUrl` 제거 — TypeScript 6 경고 TS5101(baseUrl 폐기, 7.0에서 동작 안 함). `paths`(`./src/*`)는 tsconfig 기준 상대 경로라 그대로 동작. 프로젝트 TS는 5.9.3, 경고는 편집기의 TS 6. vue-tsc 통과 + TS 6.0.3으로 경고 사라짐 확인

## 2026-11-02 (밤) — test1·test2 띄움 · 세션 정리
- 사용자 "테스트 포트 띄워줘" -> test1(8011 vite)·test2(8012 개발 supervisor) 띄움(선택지로 확인). 개발 DB `nexus`에 `00016` 적용
- 보관소가 꺼진 채 뜸 = 첫 `run.sh` 실행 때 남은 기본 폴더 표식(`nexus_check_tmp`) -> 폴더(표식만) 지우고 재시작(→ `REF-update-p1.md` 함정)
- 계정 = 개발 로그인 기본값 계정(관리자, `.env.development.local` 주석 줄) — 비밀번호는 vault에 안 적음
- 사용자 "updater도 켜고 샘플 데이터" -> updater 프로그램은 P2라 아직 없음을 설명, 샘플 데이터 넣기는 답 없이 보류
- 사용자: 페이지는 나옴 -> 다음으로(P1 마감 -> P2). 세션 재시작 전 vault 정리

## 2026-11-02 밤 (세션 재시작 뒤)
- 차이-1(장비 루트 테이블)~차이-5(폴더 업로드) 사용자 확인: 다섯 개 모두 이대로
- test1(8011)·test2(8012) 다시 띄움(세션 재시작으로 꺼져 있었음)
- 샘플 데이터를 개발 DB `nexus`에 넣음(개발 API 경유) → `REF-update-p1.md` "샘플 데이터"
- 아래 줄 넘침·벽에 붙은 입력칸 수정(→ `REF-design-chaltteok.md`) 뒤 P1 마감 커밋(`feat/updater-p1`), master 머지는 사용자 확인 대기
