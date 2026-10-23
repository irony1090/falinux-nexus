# BACKLOG — 나중에 할 것 (색인)

> **세션 시작 시 읽지 않는다.** 다음 파트를 고를 때 / 사용자가 요청할 때 연다. → `CURRENT.md` / `MEMORY.md`
> 한 항목 = 한 줄: **이름** · 미룬 날 · 다시 볼 조건 · 상세 위치. 상세는 REF 본문에 두고 여기는 포인터만.
> 착수하면 CURRENT로 올리고 여기서 지운다. 해결되면 지운다(이력은 history에).

## 다음 파트 후보 (기능)
- **업로드 한계 보강** · 2026-10-23 · 필요해지면 · 대기 파일 목록 표시·탭 닫은 뒤 이어 올리기·폴더째 올리기(Q26)·worker 간 폴더 전송(Q36, → `REF-transfer-down.md`) → `REF-transfer-ext.md` "알려진 한계"
- **SFTP 접근** · 2026-10-22 · **진행할 작업이 전혀 없을 때 이야기**(2026-10-23 Q21) · supervisor SFTP 서버로 `/{메인키}/{서브키}/...` 접근, 답 3개 대기 → `REF-worker-dir.md` "나중에"
- **worker 화면 송출 + 조작** · 2026-10-15 아이디어 · 나(VNC 터널 + noVNC) 추천 → `REF-worker-screen.md`
- **이미지·텍스트 보기 타일** · 2026-10-22 · worker -> 브라우저 방향 전송부터 → `REF-worker-dir.md` Q4(이미지·텍스트 보기)
- **process 단독 공유 UI** · 지금 서버 API만, 두면 받는 쪽 진입점 = 기본 루트 터미널 탭 → `REF-tile-binding-rework.md` R3(경로 없는 process)
- **회원가입** · 2026-10-12 사용자 "나중에" → `REF-frontend-auth.md` "회원가입"
- **보드 미리보기 다른 안** · 마감한 시안 E 말고 A(카드 줄)·B(보드 타일)·C(출력 축소본)·D(헤더 메뉴 확장) 보류 → `REF-board-preview.md`
- **사용자가 하려던 기능 추가** · 2026-10-16 · 내용이 기억나지 않아 보류, 떠오르면 다시

- **supervisor 주소 실시간 변경** · 2026-10-23 · 지금 후보(파트(파일 전송 확장) 끝남) · 실행 중인 worker가 접속하는 supervisor 소켓 주소를 재시작 없이 바꾸기. 상세 없음(착수 때 조율: 바꾸는 주체(화면·명령·env 감시)·영속 여부)
- **페이지 표시에 타일 크기 조절** · 2026-10-23 · 지금 후보(파트(파일 전송 확장) 끝남) · Ctrl+Alt로 뜨는 페이지 표시(→ `INDEX-board-pager.md`)에서 타일 크기도 조절. 상세 없음(착수 때 조율)

- **터치 스크롤 양 늘리기** · 2026-10-23 · 파트(글자 크기 규칙) 마감 뒤 = 지금 착수 가능(Q48(터치 스크롤 양), 터미널 14px) · 터치 드래그 때 스크롤되는 양 늘리기(→ `INDEX-term-touch.md`), 글자·줄 높이가 바뀐 뒤 체감 보고

- **worker "전부 종료" 수단** · 2026-10-22 · 파트(process 살려 두기)에서 기본 = 살려 둠(K4), 끌 때 process까지 정리하는 명령(예: `worker -stop-all`)은 필요해지면 → `REF-process-keep.md`

## 확인만 남은 것
- **시연 순서 점검** · 2026-10-16 · 브라우저로 irony + test3 동시, Nexus 경유 실행 미확인(xterm.js 한글 정렬·24bit 색 사진, worker 카메라) → `REF-demo.md` "시연 순서"
- **iPhone 실기기 확인** · 2026-10-01 · iPhone이 생기면(헤드리스로 iOS 키보드 재현 불가) → `REF-node-ui-overview.md` "모바일 가상 키보드"
- **Windows 미확인 항목** · 2026-10-19 · `.py` 실행·EDIT 저장 CRLF·TEMP 경로 공백 → `REF-windows-worker.md`
- **PC 갑작스러운 재부팅** · 2026-10-21 13:03:43 · **다음 진행할 작업이 없을 때 다시 확인**(사용자 2026-10-22). 경고 없이 끊김(journal "uncleanly shut down"), OOM·커널 패닉·하드웨어 오류 기록 없음, 앱 버퍼 원인 아님. 직전 1초 RustDesk 정시 재시작·gnome-shell 커서 drm 오류만 -> 전원·하드웨어·GPU 멈춤 추정

## 알려진 한계 · 보안 (지금 손대지 않음)
- **서브키 위조** · RS6(위조 검증) 보류 · key-subkey 결속 검증 없음, 명부를 인가에 안 쓰는 동안 보류 → `REF-node-label.md`
- **supervisor 레지스트리 영속** · registry 메모리 -> PostgreSQL 미착수
- **Hub 막힘 남은 한계** · 막힘 최대 약 25초, 줄이려면 write deadline / 소켓별 송신 큐 → `REF-realtime.md` "발견"
- **고아 스크립트 남은 한계** · 스크립트가 HUP를 피하고 worker가 crash하면 남음 -> 안 마(시작 시 청소) 후보 → `REF-worker-orphan.md`
- **저장 충돌 검사 없음** · 스크립트 내용 저장 때 나중에 저장한 쪽이 남음, 막으려면 백엔드 버전 검사 → `REF-node-ui-link.md`
- **같은 메인키 worker 2개의 보관 소켓 폴더 공유** · 2026-10-22 · test1·test3이 `/tmp/nexus-hold-1000/<메인키>/`를 같이 씀. 지금은 sqlite 목록이 달라 무해, 폴더를 훑는 정리 코드를 넣거나 두 빌드를 계속 나란히 둘 때 다시 → `REF-process-keep.md` "test3 재배포"
- **PTY 환경에서 `CLAUDE*` 빼기** · 2026-10-22 · Claude가 띄운 worker의 변수가 터미널로 넘어가 그 안의 Claude가 transcript 저장 안 함. worker가 PTY 환경을 만들 때 `CLAUDE*` 제거(코드 위치 미조사). 다음에 worker 실행 코드를 고칠 때 → `REF-build.md` "Claude가 띄운 빌드본"
- **SendFile abort 구분** · 정상 중단 sentinel 에러는 필요해지면 → `REF-transfer.md`

## 보류한 설계 · 정리
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
