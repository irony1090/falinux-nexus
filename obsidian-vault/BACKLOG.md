# BACKLOG — 나중에 할 것 (색인)

> **세션 시작 시 읽지 않는다.** 다음 파트를 고를 때 / 사용자가 요청할 때 연다. → `CURRENT.md` / `MEMORY.md`
> 한 항목 = 한 줄: **이름** · 미룬 날 · 다시 볼 조건 · 상세 위치. 상세는 REF 본문에 두고 여기는 포인터만.
> 착수하면 CURRENT로 올리고 여기서 지운다. 해결되면 지운다(이력은 history에).

## 다음 파트 후보 (기능)
- **타일 ESC로 닫기** · 2026-10-27 · **다음 파트로 예정(사용자)** · 타일별로 ESC 등 키를 누르면 그 타일을 끄는 동작. 착수 때 조율: ① 터미널 타일 안 ESC = 프로그램 입력(vim 등)과 충돌 ② "끄다" = 닫기인지 최소화인지 ③ "등" = 다른 키도? ④ 어느 타일이 대상인지(active 타일만?) ⑤ 앱 단축키 규칙(`appShortcutsOff` 검사 맨 앞, MEMORY "앱 단축키") · REF 없음, 착수 때 신설
- **업로드 한계 보강** · 2026-10-23 · 필요해지면 · 대기 파일 목록 표시·탭 닫은 뒤 이어 올리기·폴더째 올리기(Q26)·worker 간 폴더 전송(Q36, → `REF-transfer-down.md`) → `REF-transfer-ext.md` "알려진 한계"
- **SFTP 접근** · 2026-10-22 · **진행할 작업이 전혀 없을 때 이야기**(2026-10-23 Q21) · supervisor SFTP 서버로 `/{메인키}/{서브키}/...` 접근, 답 3개 대기 → `REF-worker-dir.md` "나중에"
- **worker 화면 송출 + 조작** · 2026-10-15 아이디어(2026-10-25 재논의 보류) · 다시 시작하면 SC1(VNC 서버 준비) 방향부터(제안 = 안내 + 내려받기 -> 동봉, 원격 자동 설치 비추천) · 나(VNC 터널 + noVNC) 추천 → `REF-worker-screen.md` "VNC 서버 준비 방식 재논의"
- **이미지·텍스트 보기 타일** · 2026-10-22 · worker -> 브라우저 방향 전송부터 → `REF-worker-dir.md` Q4(이미지·텍스트 보기)
- **process 단독 공유 UI** · 지금 서버 API만, 두면 받는 쪽 진입점 = 기본 루트 터미널 탭 → `REF-tile-binding-rework.md` R3(경로 없는 process)
- **닉네임 따로 받기** · 2026-10-26 · 회원-Q12(닉네임) = 지금은 닉네임 = 아이디(관리자 화면 아이디 확인용 임시 방편), 닉네임 변경·표시 이름이 필요해지면 → `REF-member.md`
- **타일 판 통째 공통화(B안)** · 2026-10-26 · 파트(회원 관리)에서 A안(계약 `PagerBoard`만)으로 함 · 세 번째 타일 판을 만들 때 · 배치(pack)·스크롤·URL·패널까지 공통 모듈로, 보드 고유 동작(전체보기·Grid별 칼럼 수·미리보기)이 걸림 → `REF-member-ui.md` "판 계약"
- **보드 Grid N×M 자동 계산** · 2026-10-26 · 파트(회원 관리) 관리 페이지에서 먼저 써 보고 보드에도 원하면 · 지금 보드 = 최대 2×2(`capacity`), 저장된 타일 크기 0.5·1의 뜻이 "칸 수"로 바뀌는 게 핵심 비용 → `REF-member-impl.md` "시안 2판"
- **보드 미리보기 다른 안** · 마감한 시안 E 말고 A(카드 줄)·B(보드 타일)·C(출력 축소본)·D(헤더 메뉴 확장) 보류 → `REF-board-preview.md`
- **사용자가 하려던 기능 추가** · 2026-10-16 · 내용이 기억나지 않아 보류, 떠오르면 다시



- **worker "전부 종료" 수단** · 2026-10-22 · 파트(process 살려 두기)에서 기본 = 살려 둠(K4), 끌 때 process까지 정리하는 명령(예: `worker -stop-all`)은 필요해지면 → `REF-process-keep.md`

## 확인만 남은 것
- **아이폰 터치 확대 막기** · 2026-10-25 · 아이폰으로 Nexus를 열 일이 생기면(사용자 "괜찮을 것 같다", 미확인) · 핀치·더블탭이 막히는지 → `REF-mobile-viewport.md` 7절
- **시연 순서 점검** · 2026-10-16 · 브라우저로 irony + test3 동시, Nexus 경유 실행 미확인(xterm.js 한글 정렬·24bit 색 사진, worker 카메라) → `REF-demo.md` "시연 순서"
- **iPhone 실기기 확인** · 2026-10-01 · iPhone이 생기면(헤드리스로 iOS 키보드 재현 불가) → `REF-node-ui-overview.md` "모바일 가상 키보드"
- **Windows 미확인 항목** · 2026-10-19 · `.py` 실행·EDIT 저장 CRLF·TEMP 경로 공백 → `REF-windows-worker.md`
- **PC 갑작스러운 재부팅** · 2026-10-21 13:03:43 · **다음 진행할 작업이 없을 때 다시 확인**(사용자 2026-10-22). 경고 없이 끊김(journal "uncleanly shut down"), OOM·커널 패닉·하드웨어 오류 기록 없음, 앱 버퍼 원인 아님. 직전 1초 RustDesk 정시 재시작·gnome-shell 커서 drm 오류만 -> 전원·하드웨어·GPU 멈춤 추정

## 알려진 한계 · 보안 (지금 손대지 않음)
- **패널 창 버튼이 빈 2번 타일** · 2026-10-28 · 다시 보이면(증상 상태를 그대로 두고 그 탭을 Claude in Chrome으로 읽기) · 사용자 브라우저에서만, 새로고침 뒤에도, 헤드리스 재현 못 함 → `REF-pager-head-btns.md` "미해결"
- **패널에서 닫은 뒤 active 없음** · 2026-10-28 · 불편하면 · 이웃 타일 자동 active 제안(다음 번호, 마지막이면 앞 번호, 머리 줄 ×도 같이?) 답 없이 마감 → `REF-pager-head-btns.md`
- **Windows 터미널 출력 줄 빠짐** · 2026-10-26 · 원인 좁히기 중(번호 줄 시험 결과 대기) · 집 서버 `check.ps1` 3번 실행 중 2·3번째에서 5번 항목(`Get-ChildItem` 한꺼번에 약 20줄)이 화면·스크롤에 없음, 빠진 범위가 매번 다름 · 유력 = ConPTY가 한 프레임에 화면 높이 넘게 지나간 줄을 버림(`REF-windows-worker.md` 63행 "ConPTY 출력 = 다시 그린 화면") → `REF-windows-worker.md`
- **서브키 위조** · RS6(위조 검증) 보류 · key-subkey 결속 검증 없음, 명부를 인가에 안 쓰는 동안 보류 → `REF-node-label.md`
- **supervisor 레지스트리 영속** · registry 메모리 -> PostgreSQL 미착수
- **Hub 막힘 남은 한계** · 막힘 최대 약 25초, 줄이려면 write deadline / 소켓별 송신 큐 → `REF-realtime.md` "발견"
- **고아 스크립트 남은 한계** · 스크립트가 HUP를 피하고 worker가 crash하면 남음 -> 안 마(시작 시 청소) 후보 → `REF-worker-orphan.md`
- **저장 충돌 검사 없음** · 스크립트 내용 저장 때 나중에 저장한 쪽이 남음, 막으려면 백엔드 버전 검사 → `REF-node-ui-link.md`
- **같은 메인키 worker 2개의 보관 소켓 폴더 공유** · 2026-10-22 · test1·test3이 `/tmp/nexus-hold-1000/<메인키>/`를 같이 씀. 지금은 sqlite 목록이 달라 무해, 폴더를 훑는 정리 코드를 넣거나 두 빌드를 계속 나란히 둘 때 다시 → `REF-process-keep.md` "test3 재배포"
- **PTY 환경에서 `CLAUDE*` 빼기** · 2026-10-22 · Claude가 띄운 worker의 변수가 터미널로 넘어가 그 안의 Claude가 transcript 저장 안 함. worker가 PTY 환경을 만들 때 `CLAUDE*` 제거(코드 위치 미조사). 다음에 worker 실행 코드를 고칠 때 → `REF-build.md` "Claude가 띄운 빌드본"
- **SendFile abort 구분** · 정상 중단 sentinel 에러는 필요해지면 → `REF-transfer.md`
- **집 서버 Cloudflare Access** · 2026-10-25 미룸(사용자) · (2026-10-26) 파트(회원 관리)의 worker API 키가 배포되면 가짜 worker 등록은 막히므로 필요성 다시 볼 것(`REF-member.md`) · 집 서버를 다른 사람에게 알리기 전 / `/worker` 무인증이라 열어 둔 동안 가짜 worker 등록 가능 · 걸 때 회사 worker용 `/worker` 경로 예외 필요 · → `REF-home-server.md` "코드 쪽 확인 결과"

## 보류한 설계 · 정리
- **집 서버 cloudflared 올리기** · 2026-10-28 · 집에 있을 때(선택, 보안 공지 기준은 10.3으로 이미 충족) · 관리자 PowerShell `winget upgrade Cloudflare.cloudflared` → `REF-home-server.md` "cloudflared 보안 공지"
- **주소 목록 3개 상한이 지금 주소를 밀어냄** · 2026-10-26 · 사용자 판단 대기 · "저장만"을 3번 넘게 하면 지금 붙어 있는 supervisor 주소가 목록에서 지워져, 새 주소가 다 실패하면 돌아올 곳이 없음(주소-1 "3개 넘으면 꼴찌 삭제") · 후보 = 지금 붙은 주소는 지우지 않기 → `REF-device-move.md` "확인"
- **Label 모듈** · 2026-10-12 · 다시 볼 조건 → `REF-node-label.md` "Label 모듈"
- **드래그 인프라(DraggableSession/useDragGhost)** · 2026-10-12 · 드래그 UX(아래 분할 UX 등)를 넣을 때 → `REF-util-drag.md`
- **SNAPSHOT worker 쪽 이전** · 2026-07-16 검토 → `REF-process-snapshot.md` "worker-side 이전 옵션"
- **소켓-8(kind 이름 접두어 제각각)** · 프론트 동시 수정 비용 때문에 보류 / 소켓-9(worker·presence push 없음) = 기능 변경 후보 → `REF-refactor-socket.md`
- **정돈-9(프론트 공용 유틸 미사용)** · 유지 결정 → `REF-refactor-cleanup.md`
- **단축키 키 조합** · 5 공유에서 뒤로 미룸
- **`register.go` 주석 SendBuffer 테스트** · 옛 정리 잔여
- **`history/transfer.md` 분할** · 11.5k자, 다음에 고칠 때

## 열린 질문 (node 타일 UI, 번호 유지)
- **1(분할 UX)** · 버튼 메뉴 vs 드래그 드롭(VSCode류). 시안은 실행/새 타일 버튼
- **3(라우팅 스킴)** · 옛 `/nodes/:parentId?`가 타일 모델에 맞는지
- **5(네이밍)** · 전체 개념 "그리드" vs "타일링", `Band`(가칭). 작업 공간 화면 이름은 "보드"로 확정(2026-10-07)

## 사용자 답 대기
- **`1ef8a97` 커밋 메시지 형식** · 2026-10-22 · 제목·본문 사이 빈 줄이 빠져 `--oneline`에 본문까지 나옴(내용은 정상). 고치려면 master·`refactor/tile-kinds` rebase 필요(push 안 해서 가능) — 고칠지 사용자 답 대기. 이후 커밋은 빈 줄 확인
- **파트(파일 전송) 이름 정정** · 2026-10-21 제안 · 내부 전송 모듈 검증·수정까지였으니 "파일 전송 검증(내부 모듈)"로 바꾸자고 함. 답 받으면 MEMORY·HISTORY·CURRENT·INDEX 반영
