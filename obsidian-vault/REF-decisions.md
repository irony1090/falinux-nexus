# REF — 확정 결정 상세 (불변)

> → 이력 `history/decisions.md` · 한 줄 요약 = `MEMORY.md` "확정 결정"(같은 이름) · → `CURRENT.md`
> 2026-10-26 밤 MEMORY 슬림화 때 MEMORY에서 원문 그대로 옮김. 결정을 바꾸면 **이 파일과 MEMORY 한 줄을 같이** 고친다.


## 계층
- **계층: 2단계 고정** (supervisor → worker, 중간 노드 없음. 식별자/라우팅/구독 1홉 평면)

## worker 라우팅·신원
- **worker 라우팅/신원 = 메모리 레지스트리**(`InstanceKey`=`메인키#서브키`) — 살아있는 연결로만 라우팅, **DB가 라우팅 authority 아님**. ※영속 자체는 가능(구 `agents` 제거 = 이름·역할 불명 때문)
  - **관측용 roster는 DB 허용**: `worker_instances(main_key, sub_key, last_seen)` = UI에 main_key별 인스턴스 활성/비활성 표시용(+`first_seen`). online은 저장 말고 레지스트리로 **파생**(roster ≠ authority). 구현 마감 2026-10-12 (→ `REF-node-label.md` "roster 계획")
  - node의 device 결속 = FK 아닌 `device_key TEXT`(**= main_key**, 폴더는 장비 클래스에 묶임 / subkey는 실행 시 런타임 선택) (→ `REF-node-label.md`)

## agent 식별자
- **agent 식별자: 사전 지정 메인키 + supervisor가 접속 시 부여하는 서브키.** 저장/조회는 `메인키#서브키`(`InstanceKey()`). MAC 폐기

## process 영속성
- **process 영속성**(2026-10-22 개정): process는 **worker 밖 보관 프로세스(`worker -hold`, process당 1개)** 가 PTY·화면 ring을 쥐고, worker는 sqlite 목록(살아 있는 것만)으로 다시 붙는다 → worker·supervisor 재시작에도 터미널·화면 유지 / supervisor 영속(PG, 끝난 지 24시간 지나면 주기 삭제) (→ `REF-process-keep.md` K1~K7).

## 재연결
- **재연결 reconciliation (2026-07-01 개정 — 구 `Done(502)` 비관 폐기)**: worker 끊김→관련 process **PENDING**(죽이지 않음) / 재접속→worker가 자기 live 상태 보고로 재동기화(낙관적). **모든 상태전이=process 라우터 `status` 단일 깔때기**(worker 보고 or 끊김 시 supervisor 합성). **frontend 끊김≠종료**(명시적 kill 전까진 실행 유지, 브라우저 종료 무시), **같은 세션 재접속=보던 화면 복원**(ring+세션원장) (→ "종료/재접속 모델")
  - **worker 자체 종료 = process 살려 둠**(2026-10-22 개정, K4(기본 종료 동작)): 종료 신호를 받아도 worker만 끝나고 보관 프로세스·스크립트는 남는다. 옛 "종료 신호 = 스크립트 정리"(`INDEX-worker-orphan.md`) 폐기. 전부 종료 수단은 BACKLOG

## worker OS
- **worker는 Linux·Windows 둘 다**(2026-10-19): OS에 묶인 코드는 `internal/execute/pty`의 `_linux.go`/`_windows.go`(+ worker `*_windows.go`/`*_other.go`)에만 둔다. **스크립트 노드 문법 = 그 OS에서 그대로 실행되게 작성**(Windows는 노드 이름 확장자로 실행기, `#!` 안 씀). worker를 띄우는 방식이 남긴 신호 무시 속성(Linux SIGHUP·Windows Ctrl+C)은 자식에 상속되므로 worker가 되돌린다 (→ `INDEX-windows-worker.md`)

## 에러 처리
- **에러 처리 = panic-style**(2026-06-26 재확정): 핸들러 `panic(web.Err(...))` → `PanicMiddleware` 렌더 (→)

## process 동기화 범위
- **process 동기화 범위 = 계정**(2026-10-01, 1단계 코드·실행 확인·커밋 완료): 같은 계정이면 탭·브라우저·기기 무관 동기화. 인가 = 계정(DB) / 출력 라우팅 = 탭 소켓(메모리) / PTY 크기 우선권 = 서버 발급 탭 id(메모리) → `process_subscribers` 제거(`00005`). 공유는 계정 단위·소유자만 신청 (→ `REF-process-sync.md`, 구현·확인 → `REF-process-sync-impl.md`)

## 공유 판정
- **공유 판정 = DB 함수 한 곳**(2026-10-06, 공유-a): 가까운 공유 단계 계산은 `node_permission`/`process_permission`(`00007_shares.sql`)에만 두고 쿼리·Go는 호출만. 단계 = 0 차단·10 보기·20 조작·30 편집·40 소유자(함수 반환만), 비교는 `>=`. 볼 수 없음 = 404 / 단계 미달 = 403. **공유 단위 = 폴더 + process 단독**(스크립트 노드 공유 금지, 2026-10-07 SH1 개정) (→ `REF-process-share-db.md`, `REF-process-share-authz.md`, `REF-process-share.md`)

## 실시간 push
- **실시간 push = 인가/라우팅 분리**(2026-06-30): 인가=DB(구독 시점 1회) / 라우팅=메모리(`subscribe.Hub` topic Publish, DB 무접촉). 수신자=presence(구독)≠소유자 → 공유·다중 유저 동일 화면 흡수. 토픽 단위: process 출력 = `PROCESS:<uid>` / **node 변경 = 볼 수 있는 접속 계정의 `USER:<userId>`(옛 문서의 `TILES:<userId>` = 같은 것), 발행 시점 DB 판정**(옛 "펼친 폴더 `NODE:`" 폐기, → `REF-node-list-sync.md`). "DB≠라우팅 authority"를 browser 평면으로 확장. **장비 상태**(2026-10-27) = `USER:<userId>`로 `DEVICE:CHANGE {mainKey}` 신호만 → 받은 화면이 기존 API로 재조회, 폴링 60초 안전망, 소켓 connect = `App.vue`(로그인 중 모든 페이지) (→ `REF-device-push.md`). 비대칭: 브라우저 call→서버 Handle / 서버 Emit→브라우저 on (→ `REF-realtime.md`)

## 모바일 키보드
- **모바일 키보드 = `resizes-content` + 열린 동안 배치 영역은 키보드 전 높이(px) 고정**(2026-10-14): 타일 재배치 금지. `dvh`로 배치하는 새 영역도 지킬 것 (→ `REF-mobile-viewport.md` 4). **2026-10-20 개정**: 포커스 터미널 타일만 키보드 위 영역에 맞춤(그 PTY rows만 바뀜) (→ `REF-board-keyboard.md` "맞춤 계획")

## 타일 이동
- **타일 이동 = 페이지 표시의 "타일 위치 규칙"**(2026-10-21): 타일로 가는 경로는 전부 `reveal`(= `revealCol`, 화면 밖일 때만 막 다 보이는 만큼). Grid 단위 이동은 "Grid로 가기"가 목적일 때만 (→ `REF-tile-move.md`)

## 앱 단축키
- **앱 단축키 = 검사 한 곳 `appShortcutsOff`(`tileGrids.store`)를 맨 앞에서 거친다**(2026-10-21): 터미널 타일마다 "단축키 끄기"(타일 트리 `shortcutsOff`) — 그 타일이 active면 앱 단축키 전부 꺼짐. **새 단축키도 반드시 이 검사를 넣는다**. 터미널 자체 동작(Ctrl+C 복사·Ctrl+V 붙여넣기)은 대상 아님 (→ `REF-tile-move.md` 이동-3)

## 타일 종류
- **타일 종류 = 능력으로 묻는다**(2026-10-22, 파트(타일 종류 정리)): 여러 종류가 같이 쓰는 로직은 `type` 비교 대신 능력-1(루트 가능)·2(자식 가능)·3(process 담음) — 서버 `router/tileKind.go` `tileKinds` / 프론트 `widget/util/tileKind.util.ts` `TILE_KINDS`(두 표를 같게). 화면 = 공통 틀 `TileFrame` + 종류 컴포넌트(`tileComponents.ts`). 새 종류 추가 절차 → `REF-tile-kinds.md` "새 종류를 더할 때"

## 전송
- **전송 = supervisor 전송 목록(메모리) + 계정 토픽 `TRANSFER:*` push**(2026-10-23, 파트(파일 전송 확장)): 업로드·내려받기·worker 간 전송 공용. 브라우저 -> supervisor 본문은 **HTTP 조각 512KB**(앞단 nginx `client_max_body_size` 기본 1MB). worker -> supervisor 읽기는 끌어오기(`FILE_OPEN`·`READ`·`CLOSE`). 장비 디렉토리 인가 = `requireDeviceDir` (→ `REF-transfer-ext.md`·`REF-transfer-down.md`)

## 브라우저 API
- **브라우저 API = 전부 `/api` 아래**(2026-10-26): 프론트 페이지 경로와 겹치지 않게 supervisor 코드에서(웹 서버 설정 의존 X). 새 API는 `mount*`에 받은 그룹에만, 프론트 주소는 `getBaseURL()` 뒤 문자열 이어 붙이기(`new URL('/x', base)` 금지). worker 소켓 `/worker`만 밖 (→ `REF-api-prefix.md`, `INDEX-member`)

## 회원·장비 인가
- **회원·장비 인가**(2026-10-26, 파트(회원 관리)): worker 접속 = 메인키 + **API 키(메인키마다 하나)** — 없으면 거부. 키 = 검증용 해시 + 다시 보기용 AES-GCM 암호문(supervisor env **`KEY_SECRET` 필수**, 잃으면 보기만 불가). 장비(메인키) = 회원이 이름 등록, 전체 유일·이름 변경 없음. **실행 권한 = 장비 주인 + 그 장비가 지정된 폴더를 공유받은 사람**. 관리자 = `users.perms`의 `admin`(회원 0명일 때 첫 가입자) (→ `REF-member.md`, `INDEX-member`)

## 배포 전 필수
- **배포 전 필수**(2026-10-26): 회원이 이미 있는 DB에 배포할 때는 **관리자로 지정할 회원을 먼저 사용자에게 묻는다**. worker 실행 파일과 env `API_KEY`는 같이 바꾼다 (→ `REF-member.md` "배포 전 확인", 집 서버 → `REF-home-server.md`)

## 젤리 효과
- 2026-10-27 파트(젤리·관리 기록). 누를 수 있는 것 hover·누름 + 컨트롤 패널 등장·퇴장 = `styles/jelly.scss` 한 곳(세기 중간). **새 버튼 클래스를 만들면 `$box` 목록에 넣는다.** 알약(`FootPill`) 안은 안쪽만, 넓은 줄은 제외
- 가운데 정렬 팝업에 젤리(scale)를 걸 땐 `transform: translate(-50%,-50%)` 금지 -> `translate: -50% -50%` 속성 + 클래스 `jelly-pop`(열림 = `.shown`). 안 지키면 화면상 오른쪽 아래 기준으로 커짐 → `REF-jelly.md` "함정"

## 입력칸
- 2026-10-27 파트(찰떡 입력칸). 화면의 input·select·textarea·체크 = `CtTextField`·`CtSelect`·`CtAutocomplete`·`CtTextarea`·`CtCheckbox`(Vuetify 감싸기, 속성·슬롯 그대로 넘김) + 공용 `size`(sm 32 / md 40)·`wall`(l/r/lr = 곧은 끝). 모양은 `styles/ct-field.scss` 한 곳. **기본 `<select>` 금지**(목록이 OS 흰 바탕 + 앱 글자색 = 흰 바탕 흰 글씨)
- 함정: v-input은 `flex: 1 1 auto` + 칸 줄 `1fr` -> 세로 flex 상자에서 칸이 늘어나 글자가 위로 쏠림. `.ct-field`에서 막아 둠 → `REF-ct-field.md` "함정"
- 예외: 경로 직접 입력칸(`PathInput.vue`, 자체 자동완성) · 숨은 파일 선택

## 패널 창 버튼
- 2026-10-28 파트(패널 창 버튼). Ctrl+Alt 컨트롤 패널(보드 `PagerPopup`·관리 `MemberPagerPopup`) 왼쪽 = active 타일 머리 줄 창 버튼 `[숨기기 | 크기 | 닫기]` 중 그 타일에 있는 것만(비활성으로 자리 차지 X), 동작 = 머리 줄과 같음. 보드는 `TileFrame`이 `node/store/tileHeadActs.store.ts`에 등록 -> 머리 줄 창 버튼 조건·동작을 바꾸면 등록도 같이. 패널엔 위치 글자 없음·도트 자리 최소 3개 폭
- 미해결 1건(사용자 브라우저에서 2번 타일 패널이 빔, 재현 못 함) → `REF-pager-head-btns.md` "미해결"

## 경로 줄 도구
- 2026-10-28 파트(경로 줄 정리). 폴더 타일·장비 타일 경로 줄의 버튼 = 찰떡 묶음 `.ct-grp` > `.ct-seg`(비슷한 기능끼리 한 알약, 묶음 사이 4px, 구분선 없음, 기능 색 = 디렉토리 초록·공유 청록·붙여넣기 파랑). `[뒤로 | 앞으로 | 새로고침]`(`TileNavButtons`)은 경로 왼쪽·타일 왼쪽 벽, 머리 줄엔 두지 않음. 가운데 묶음이 넘치면 `HScroll`(가로 스크롤), 파란 끝 블록은 고정
- 함정: 두 몸통의 `button:where(:not(...))` 초기화 규칙이 몸통 안에 직접 쓴 버튼 클래스의 바탕을 지움 -> 새 버튼 클래스는 제외 목록에 넣는다 → `REF-path-row.md` "함정"

## 모듈 현황 원문 (MEMORY에서 옮김, 2026-10-26)
| 모듈 | 상태 | 문서 |
|------|------|------|
| 통신 인프라·파일 전송·supervisor web | 마감 | `INDEX-infra.md` · `INDEX-transfer-e2e.md` |
| process 실행 (execute/pty/manager/bind) | 배선·터미널 기능·동기화/공유/보드·종료 정책·고아 스크립트·Windows worker 전부 마감. EXEC 실행 방식 = 현상 유지 확정 | `REF-process.md` + 마감 INDEX(`HISTORY.md` "마감 파트") |
| DB/스토어 (sqlc·goose·store) | 구현 완료 | `REF-db.md` |
| Node 카탈로그 | Node 모듈·roster(명부)·폴더 목록 동기화·노드 클립보드·UI 연동 마감, label 보류 | `REF-node-label.md` + `INDEX-node-ops.md` |
| 공용 PATCH 래퍼 (`internal/patch`) | `patch.Field[T]` 3-state(`{valid,value}`), worker도 재사용 예정 | `REF-node-label.md` |
| 프론트엔드 (apps/frontend) | 타일 UI·터미널·⑮(모바일 가상 키보드)·로그인·active 타일·보드 미리보기·회원 관리(내 장비·관리자 페이지) 전부 마감 | `REF-frontend.md` · `REF-node-ui.md`(컨셉) · `REF-node-ui-overview.md`(전체보기·가상 키보드) |
| 실시간 push (socket) | 전송 토대·3모드·ping/pong·node CRUD 발행(`AfterCommit`)·계정 토픽, 마감 | `REF-realtime.md` |
| 위젯 (Skeleton/SkeletonGroup/StickyBox/TileLayout) | 구현만, 실사용 미배선(TileLayout = index.vue 데모만) | `REF-widget.md` |
| 타일 판 공용 (`feature/widget` 판 계약 `PagerBoard`) | 알약 끌기·미니맵·빈 곳 휠·컨트롤 패널 여는 길·Ctrl+Alt+숫자를 보드·관리 페이지가 같이 씀. 새 타일 판 = 계약만 채움. 푸터 액션 버튼 = `FootActions`(공용) | `REF-member-ui.md` "판 계약" · `REF-device-move.md` "푸터 액션 자리" |
| 범용 유틸 (EventInterface/Memoized/LifecycleRegistry/GroupedSet) | 4종 + 공유 리사이즈 관측 그룹(`feature/common`), 반복 주제 | `REF-util.md` · `REF-util-drag.md`(드래그, 보류) |
