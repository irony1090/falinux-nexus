# MEMORY

> 매 세션 시작 시 항상 로드. **불변 결정 + 항상 필요한 것만** 간결하게.
> 상세 설계·재사용 지식은 `REF-*.md`, 작업 이력은 `history/*.md`, 현재 진행은 `CURRENT.md`.

## 경로 참조 (중요)
- **현재 프로젝트 (Nexus)**: `/home/jh-bae/irony/nexus`
  - Go 코어: `/home/jh-bae/irony/nexus/apps/core`
- **이전 프로젝트 (PortBridge)**: `/home/jh-bae/irony/test-jig`
  - Go: `.../apps/test-jig-api` / 프론트: `.../apps/test-jig-web` / vault: `.../obsidian-vault`

> 사용자가 **"test-jig" / "예전·구·이전 프로젝트" / "PortBridge"** 로 칭하면 `/home/jh-bae/irony/test-jig`. 해당 코드 참조할 것.

## 프로젝트 개요
- **Nexus** = 2-tier 에이전트 관리 플랫폼
  - **Worker Agent**: 자기 process 관리(실행/모니터링/종료)
  - **Supervisor Agent**: 다수 worker 관리(연결/상태/명령 분배)
- 출발점: PortBridge의 agent↔server 통신 아키텍처 계승. 방향이 "테스트 장비 데스크탑 UI"→"에이전트 계층 관리"로 전환되어 신규 분리. (PortBridge 전용 UI·아이콘·그룹 로직 폐기)

## 확정 결정 (불변 — 장기 유지)
> 한 줄 요약. 상세·경위·포인터 = `REF-decisions.md`의 같은 이름 절. 결정을 바꾸면 두 곳을 같이 고친다
- **계층**: 2단계 고정(supervisor -> worker, 1홉)
- **worker 라우팅·신원**: 메모리 레지스트리(`메인키#서브키`)로만 라우팅, DB는 라우팅 authority 아님. roster `worker_instances` = 관측용(online은 파생). 폴더↔장비 = `device_key`(= 메인키)
- **agent 식별자**: 사전 지정 메인키 + 접속 때 받는 서브키(`InstanceKey()`). MAC 폐기
- **process 영속성**: 보관 프로세스(`worker -hold`)가 PTY·화면을 쥠 -> worker·supervisor 재시작에도 유지. worker가 종료 신호를 받아도 process는 남김
- **재연결**: 끊김 = PENDING(죽이지 않음), 재접속 = worker 보고로 재동기화. 상태 전이는 `status` 한 곳. 브라우저 끊김 ≠ 종료
- **worker OS**: Linux·Windows 둘 다. OS 코드는 `_linux`/`_windows`(`_other`) 파일에만. 스크립트 노드 = 그 OS 문법 그대로
- **에러 처리**: `panic(web.Err(...))` -> `PanicMiddleware`
- **process 동기화 범위**: 계정(탭·기기 무관). 인가 = DB / 출력 라우팅·PTY 크기 우선권 = 메모리
- **공유 판정**: DB 함수 한 곳(`node_permission`/`process_permission`), 쿼리·Go는 호출만. 공유 단위 = 폴더 + process 단독
- **실시간 push**: 인가(DB, 구독 때 1회) / 라우팅(메모리 Hub) 분리. 토픽 `PROCESS:<uid>`·`USER:<userId>`·`TRANSFER:*`. 장비 상태 = `DEVICE:CHANGE` 신호 -> 재조회, 소켓 connect = `App.vue`(로그인 중 모든 페이지)
- **모바일 키보드**: `resizes-content` + 열린 동안 배치 영역 높이 고정(타일 재배치 금지), 포커스 터미널만 키보드 위에 맞춤 (→ 반복 주제 `REF-mobile-viewport.md`)
- **타일 이동**: 타일로 가는 경로는 전부 `reveal`(화면 밖일 때만)
- **앱 단축키**: `appShortcutsOff` 검사를 맨 앞에 — **새 단축키도 반드시**
- **타일 종류**: `type` 비교 말고 능력으로 묻는다. 서버 `tileKinds`·프론트 `TILE_KINDS` 두 표를 같게
- **process 범위**: worker에서 실제로 실행한 것(PTY)만. 보기 타일처럼 실행 없는 것은 타일만, 목록엔 트리 조회로 붙임
- **전송**: supervisor 전송 목록 + `TRANSFER:*` push. 브라우저 본문 = 조각 512KB
- **브라우저 API**: 전부 `/api` 아래. 새 API는 `mount*` 그룹에만, 프론트는 `getBaseURL()` + 문자열(`new URL` 금지)
- **회원·장비 인가**: worker 접속에 API 키(메인키마다 하나, env `KEY_SECRET` 필수). 실행 권한 = 장비 주인 + 지정 폴더 공유받은 사람
- **젤리 효과**: 누를 수 있는 것·컨트롤 패널 = `styles/jelly.scss` 한 곳. 새 버튼 클래스는 `$box` 목록에, 가운데 팝업은 `translate` 속성 + `jelly-pop`
- **닫은 뒤 active**: active 타일이 닫히면 위치 번호 이웃(다음, 마지막이면 앞)이 active(보드·관리 둘 다), 최소화·보드 바꾸기는 해제만
- **패널 창 버튼**: 컨트롤 패널 왼쪽 = active 타일 머리 줄 `[숨기기 | 크기 | 닫기]`(있는 것만). 머리 줄 창 버튼을 바꾸면 `TileFrame`의 `tileHeadActs` 등록도 같이
- **경로 줄 도구**: 폴더·장비 타일 경로 줄 버튼 = `.ct-grp` 묶음(`chaltteok.scss`) + 넘치면 `HScroll`, 이동 묶음은 경로 왼쪽. 몸통의 버튼 초기화 제외 목록에 새 버튼 클래스 넣기
- **입력칸**: input·select·textarea·체크 = `Ct*` 컴포넌트(`feature/widget/component`, 모양 = `styles/ct-field.scss` 한 곳). 기본 `<select>`·직접 그린 입력칸 금지
- **배포 전 필수**: 회원이 있는 DB에 배포하기 전에 관리자로 지정할 회원을 사용자에게 묻는다

## 기술 스택 (확정)
- **레포**: `nexus` / 모노레포 `apps/core`(Go) + `apps/frontend`(프론트, 스캐폴딩 완료 2026-06-30)
- **프론트**: Vue3 + TS + Vuetify4 (scratch preset, Router=standard / **Pinia 미사용**, CSS framework 없음). 상태공유=provide/inject(+컴포넌트 밖 필요 시 reactive 모듈 패턴). 기본폰트=Noto Sans Korean. **색 = Vuetify 테마 색 이름 의존, hex 금지**(2026-09-29), 공유 표시 = 테마 `share`(청록). **글자 크기 = Vuetify(MD3) 단계 이름 `@include vars.font('body-small')`, px 금지**(2026-10-23, 단계 값 한 곳 `styles/_typography.scss`·전체 배율 `$font-scale`, → `INDEX-typography`) (→ `REF-frontend.md`)
- **Go**: 단일 모듈 `nexus`, 멀티 cmd(`cmd/supervisor`, `cmd/worker`), Go 1.26
- **DB**: worker → SQLite(`modernc.org/sqlite`, CGO 없음) / supervisor → PostgreSQL(`pgx/v5`)
- **sqlc**: 두 엔진 한 `sqlc.yaml` → `workerdb`/`superdb` 분리 생성
- **마이그레이션**: goose (스키마 단일 진실 = goose 파일, sqlc가 schema로 읽음)
- **터미널 출력**: xterm.js (프론트 착수 시)

## 로컬 개발 환경 (DB)
- supervisor PostgreSQL = docker 컨테이너 **`postgres15`**(`docker start postgres15` — `postgres`는 이미지명). user `irony` / pass `!Fa1289` / `5432`. 기본 DB `workspace`(구 PortBridge), `nexus` DB는 수동 생성(2026-06-26)
- supervisor 띄우면 `init()`의 mountStore가 goose 마이그레이션 자동 적용(멱등)
- **포트**: supervisor 8012(0.0.0.0) · 프론트 8011 · worker -> localhost:8012(외부 접속 설정, 되돌리는 법 → `REF-external-access.md`)
- **빌드본 = 사용용**: `build/<이름>`(`scripts/build.sh`) = **8013(test3) · DB `nexus_build`**. 사용자 Claude 세션이 그 위에서 돎 — **Claude는 8013을 끄지 않는다**(재기동 = 사용자). 개발 실행 = `go run ./cmd/supervisor -env dev.env` 필수 (→ `REF-build.md`)
- **Windows VM `nexus-win` = 시험할 때만 켠다**(코드 작성 중엔 `docker stop`) (→ `REF-windows-vm.md`) / 포트 점유 함정(`bind: address already in use`) → `REF-test-env.md`

## apps/core 디렉토리
```
cmd/{supervisor,worker}/{main.go, router/, constants/}
internal/
  protocol/ transport/ call/ execute/{,pty}  공유
  subscribe/ manager/ transfer/ util/ web/   재사용 코어
  worker/{store, db/{migrations,query,gen=workerdb}}
  supervisor/{bind, store, db/{...,gen=superdb}}
sqlc.yaml  README.md  .gitignore
```

## 모듈 현황
| 모듈 | 상태 | 문서 |
|------|------|------|
| 통신·전송·process·DB·Node·실시간·프론트(타일 UI·회원 관리) | 마감 | 목록 = `HISTORY.md` "마감 파트" -> 각 INDEX. 주제 REF = `REF-process.md`·`REF-db.md`·`REF-node-label.md`·`REF-frontend.md`·`REF-realtime.md` / 옛 표 = `REF-decisions.md` 끝 |
| 타일 판 공용 `PagerBoard` | 새 타일 판 = 계약만 채움(알약 끌기·미니맵·컨트롤 패널·Ctrl+Alt+숫자). 푸터 액션 = `FootActions` | `REF-member-ui.md` "판 계약" |
| 위젯 · PATCH 래퍼 `internal/patch` | 위젯 = 구현만·미배선 / `patch.Field[T]` 3-state | `REF-widget.md` · `REF-node-label.md` |
| 범용 유틸(반복 주제) | 4종 + 공유 리사이즈 관측 그룹 | `REF-util.md` · `REF-util-drag.md`(보류) |


## reference 인덱스
- **★ 반복 주제(프로젝트 무관, 다른 프로젝트에서도 생김)**: `REF-mobile-viewport.md`(모바일 키보드·뷰포트, Edge 하단 막대 오류 실측) / `REF-util.md`(범용 유틸)
- 설계/재사용 지식(위 모듈 현황표에 있는 것 말고): `REF-decisions.md`(확정 결정 상세) · `REF-test-env.md`(테스트 환경·함정) · `REF-test-scripts*.md`(확인 스크립트 재작성 명세, 날짜별 5개) · `REF-external-access.md`(외부 접속 test1~3 = 8011~8013) · `REF-demo.md`(시연 데이터) · `REF-worker-screen.md`(worker 화면 송출, 아이디어) · `REF-file-view.md`(장비 파일 편집기·보기 타일·칸 메뉴 내려받기, 구상) · `REF-view-tile.md`(파트(보기 타일) 구현)
- **디자인 규칙**: `REF-design-chaltteok.md` 찰떡 규칙(알약 버튼·라벨이 부모 벽에 닿으면 그 끝 = 부모 모양, 다른 작업공간도 직접 참조). Nexus 적용 = 파트(찰떡 UI) 10-25, 묶음 클래스 `styles/chaltteok.scss`
- **진행 중 파트**: 없음 (직전 = 보기 타일(마감, 10-28, `feat/file-view` 머지) → `REF-view-tile.md` · 그 전 파일 편집기 → `REF-file-view.md`. 다음 영상 파트가 두 REF를 이어 써서 INDEX로 안 접음)
- 작업 이력: `HISTORY.md`가 색인(`history/{주제}.md`, REF와 1:1)
- **마감 파트(세션 시작 시 안 읽음, REF가 필요하면 INDEX부터)**: 목록·최근 순서 = `HISTORY.md` "마감 파트"
- 통신/PTY 상세 PLAN: `PLAN-agent-comm.md` / 구독 모델: `PLAN-subscription.md`
- 현재 진행: `CURRENT.md` / 나중에 할 것: `BACKLOG.md`(세션 시작 시 안 읽음)
